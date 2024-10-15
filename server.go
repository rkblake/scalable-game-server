package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var ctx = context.Background()

func main() {
	log.Println("[Server] Starting...")

	// connect to docker
	err := ConnectDocker()
	if err != nil {
		panic(err)
	}

	// forward packets
	log.Println("[Server] Starting reverse proxy")
	go func() {
		err := ForwardTCP()
		if err != nil {
			log.Println(err)
		}
	}()

	go func() {
		err := ForwardUDP()
		if err != nil {
			log.Println(err)
		}
	}()

	log.Println("[Server] starting container health check")
	ContainerHealthCheck()

	log.Println("[Server] Started")

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("[Server] Received SIGINT. cleaning up")
		CleanupDocker()
		log.Println("[Server] Finished cleanup. shutting down")
		os.Exit(0)
	}()

	// start http server
	HandleEndpoints()

}
