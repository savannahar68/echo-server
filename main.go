package main

import (
	"flag"
	"fmt"

	"github.com/savannahar68/echo-server/config"
	"github.com/savannahar68/echo-server/server"
)

func main() {
	fmt.Println("Starting server!")
	port := flag.Int("port", 7379, "Port on which server needs to be started!")
	incomingRequest := flag.String("incoming-connection", "0.0.0.0", "Accept incoming connections form!")
	serverType := flag.String("type", "tcp", "Type of server!")

	flag.Parse()

	serverConfig := config.Server{
		Port:            *port,
		IncomingRequest: *incomingRequest,
		Type:            *serverType,
	}

	fmt.Println("Starting server!")
	err := server.Start(serverConfig)

	if err != nil {
		_ = fmt.Errorf("unable to start server on port %v with error %+v", port, err)
	}
	fmt.Println("Shutting server!")
}
