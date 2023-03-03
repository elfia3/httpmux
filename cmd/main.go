package main

import (
	"context"
	transport "httpmux/transport/http"

	"log"
	"os/signal"
	"syscall"
)

const (
	ServerHost = "localhost:5002"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	service := transport.CreateService(4, ServerHost)
	go service.Run()

	<-ctx.Done()
	log.Println("gracefully shutting down...")

	if err := service.ShutdownGracefully(); err != nil {
		log.Printf("Error with gracefull shutdown : %s\n", err.Error())
	}
}
