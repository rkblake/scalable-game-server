package main

import (
	"context"
	"log"
	// "net"
	"os"
	"os/signal"
	"syscall"
)

var ctx = context.Background()
var proxy = Proxy{}

func main() {
	log.Println("[Server] Starting...")

	// connect to docker
	err := ConnectDocker()
	if err != nil {
		panic(err)
	}

	// forward packets
	log.Println("[Server] Starting reverse proxy")
	// proxy := Proxy{}
	// l, err := net.Listen("tcp", "127.0.0.1:9000")
	// if err != nil {
	// panic(err)
	// }
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	go proxy.Start(cancelCtx, "0.0.0.0:9000")

	log.Println("[Server] starting container health check")
	ContainerHealthCheck()

	// start http server
	log.Println("[Server] starting HTTP listener")
	go HandleEndpoints()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("[Server] Received SIGINT. cleaning up")
		CleanupDocker()
		cancelFunc()
		log.Println("[Server] Finished cleanup. shutting down")
		os.Exit(0)
	}()

	log.Println("[Server] Started")

	select {}
}
