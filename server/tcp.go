package server

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/savannahar68/echo-server/config"
)

func Start(serverConfig config.Server) error {
	listen, err := net.Listen(serverConfig.Type, serverConfig.IncomingRequest+":"+strconv.Itoa(serverConfig.Port))
	if err != nil {
		return errors.Unwrap(err)
	}

	for {
		conn, err := listen.Accept()
		if err != nil {
			return errors.Unwrap(err)
		}

		for {

			b := make([]byte, 1024)
			read, err := conn.Read(b)
			if err != nil {
				err := conn.Close()
				fmt.Println("Connection closed!")
				if err != nil {
					return err
				}
				return errors.Unwrap(err)
			}

			fmt.Printf("Read command %s", string(rune(read)))

			write, err := conn.Write(b)
			if err != nil {
				return err
			}

			fmt.Printf("Command written to pipeline %s", write)
		}
	}
}
