package server

import (
	"io"
	"log"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/savannahar68/echo-server/config"
	"github.com/savannahar68/echo-server/core"
)

var conClients = 0
var cronFrequency = 1 * time.Second
var lastCronExecutionTime = time.Now()

const EngineStatus_WAITING int32 = 1 << 1
const EngineStatus_BUSY int32 = 1 << 2
const EngineStatus_SHUTTING_DOWN int32 = 1 << 3
const EngineStatus_TRANSACTION int32 = 1 << 4

var eStatus int32 = EngineStatus_WAITING

var connectedClients map[int]*core.Client

func init() {
	connectedClients = make(map[int]*core.Client)
}

func RunAsyncTCPServer(wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)
	}()

	log.Println("Starting an asynchronous server on ", config.Host, config.Port)
	maxClients := 20000

	events := make([]syscall.Kevent_t, maxClients)

	// create a socket
	serverFd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)

	if err != nil {
		println("Err", err)
		panic(err)
	}

	defer func(fd int) {
		err := syscall.Close(fd)
		if err != nil {

		}
	}(serverFd)

	// set the server socket as non-blocking
	err = syscall.SetNonblock(serverFd, true)
	if err != nil {
		panic(err)
	}

	// Bind the IP and Port
	ip4 := net.ParseIP(config.Host)
	err = syscall.Bind(serverFd, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	})
	if err != nil {
		panic(err)
	}

	// start listening on socket
	err = syscall.Listen(serverFd, maxClients)
	if err != nil {
		panic(err)
	}

	// Async io - event loop start

	// create KQueue
	kqFd, err := syscall.Kqueue()
	if err != nil {
		log.Println("Error creating Kqueue descriptor!")
		panic(err)
	}
	defer syscall.Close(kqFd)

	socketServerEvent := syscall.Kevent_t{
		Ident:  uint64(serverFd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
	}

	// Listen to read events on the Server itself
	if changeEventRegistered, err := syscall.Kevent(kqFd, []syscall.Kevent_t{socketServerEvent}, nil, nil); err != nil || changeEventRegistered == -1 {
		panic(err)
	}

	for atomic.LoadInt32(&eStatus) != EngineStatus_SHUTTING_DOWN {

		// Active delete of expired keys
		if time.Now().After(lastCronExecutionTime.Add(cronFrequency)) {
			core.DeleteExpiredKeys()
			lastCronExecutionTime = time.Now()
		}

		nevents, err := syscall.Kevent(kqFd, nil, events, nil)
		if err != nil {
			continue
		}

		if !atomic.CompareAndSwapInt32(&eStatus, EngineStatus_WAITING, EngineStatus_BUSY) {
			switch eStatus {
			case EngineStatus_SHUTTING_DOWN:
				return
			}
		}

		for i := 0; i < nevents; i++ {

			// accept incoming events
			if events[i].Ident == uint64(serverFd) {
				fd, _, err := syscall.Accept(serverFd)
				if err != nil {
					panic(err)
				}

				connectedClients[fd] = core.NewClient(fd)

				conClients++
				err = syscall.SetNonblock(fd, true)
				if err != nil {
					return
				}

				// add this TCP connection to be monitored
				socketClientEvent := syscall.Kevent_t{
					Ident:  uint64(fd),
					Filter: syscall.EVFILT_READ,
					Flags:  syscall.EV_ADD | syscall.EV_ENABLE,
				}

				if changeEventRegistered, err := syscall.Kevent(kqFd, []syscall.Kevent_t{socketClientEvent}, nil, nil); err != nil || changeEventRegistered == -1 {
					panic(err)
				}
			} else {
				comm := connectedClients[int(events[i].Ident)]
				if comm == nil {
					return
				}
				cmds, err := readCommands(comm)
				if err != nil {
					err := syscall.Close(int(events[i].Ident))
					delete(connectedClients, int(events[i].Ident))

					if err != nil {
						return
					}
					conClients -= 1
					if err == io.EOF {
						break
					}
					log.Println("err", err)
				} else {
					log.Println("command", cmds)
					respond(cmds, *comm)
				}
			}

		}

		// Waiting for next event
		atomic.StoreInt32(&eStatus, EngineStatus_WAITING)
	}

}

func WaitForSignal(wg *sync.WaitGroup, sigs chan os.Signal) {
	defer wg.Done()
	<-sigs

	// Wait for existing command and then shut down

	// If server is busy wait
	for atomic.LoadInt32(&eStatus) == EngineStatus_BUSY {
	}

	atomic.StoreInt32(&eStatus, EngineStatus_SHUTTING_DOWN)

	core.Shutdown()
	os.Exit(0)
}
