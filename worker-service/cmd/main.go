package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx,cancel := context.WithCancel(context.Background())
	Init(ctx)
	log.Println("Job to initialize cluster operations started successfully")
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-sigc
	log.Println("Request to close service received, sig: ",sig)
	cancel()
}
