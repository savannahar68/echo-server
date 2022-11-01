package main

import (
	"flag"
	"log"

	"github.com/savannahar68/echo-server/config"
	"github.com/savannahar68/echo-server/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for the dice server")
	flag.IntVar(&config.Port, "port", 7379, "port for the dice server")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("rolling the dice 🎲")
	server.RunAsyncTCPServer()
}
