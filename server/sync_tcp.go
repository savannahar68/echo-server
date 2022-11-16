package server

import (
	"fmt"
	"io"

	"github.com/savannahar68/echo-server/core"
)

func readCommand(c io.ReadWriter) (*core.RedisCmd, error) {
	// TODO: Max read in one shot is 512 bytes
	// To allow input > 512 bytes, then repeated read until
	// we get EOF or designated delimiter
	var buf []byte = make([]byte, 512)
	_, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}
	tokens, err := core.DecodeArrayString(buf)

	if err != nil {
		return nil, err
	}

	return &core.RedisCmd{
		Cmd:  tokens[0],
		Args: tokens[1:],
	}, nil
}

func readCommands(c io.ReadWriter) (core.RedisCmds, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf[:])
	if err != nil {
		return nil, err
	}
	values, err := core.Decode(buf[:n])

	if err != nil {
		return nil, err
	}

	var cmds = make([]*core.RedisCmd, 0)

	for _, value := range values {
		tokens, err := toArrayString(value.([]interface{}))
		if err != nil {
			return nil, err
		}
		cmds = append(cmds, &core.RedisCmd{
			Cmd:  tokens[0],
			Args: tokens[1:],
		})
	}
	return cmds, nil
}

func toArrayString(value []interface{}) ([]string, error) {
	as := make([]string, len(value))
	for i := range as {
		as[i] = value[i].(string)
	}
	return as, nil
}

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmds core.RedisCmds, c core.Client) {
	core.EvalAndRespond(cmds, &c)
}

//
//func RunSyncTCPServer() {
//	log.Println("starting a synchronous TCP server on", config.Host, config.Port)
//
//	var con_clients int = 0
//
//	// listening to the configured host:port
//	lsnr, err := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))
//	if err != nil {
//		panic(err)
//	}
//
//	for {
//		// blocking call: waiting for the new client to connect
//		c, err := lsnr.Accept()
//		if err != nil {
//			panic(err)
//		}
//
//		// increment the number of concurrent clients
//		con_clients += 1
//		log.Println("client connected with address:", c.RemoteAddr(), "concurrent clients", con_clients)
//
//		for {
//			// over the socket, continuously read the command and print it out
//			cmd, err := readCommand(c)
//			if err != nil {
//				c.Close()
//				con_clients -= 1
//				log.Println("client disconnected", c.RemoteAddr(), "concurrent clients", con_clients)
//				if err == io.EOF {
//					break
//				}
//				log.Println("err", err)
//			}
//			log.Println("command", cmd)
//			respond(core.RedisCmds{cmd}, *core.NewClient(0)) // TODO: to run async server made this as default 0 fd
//		}
//	}
//}
