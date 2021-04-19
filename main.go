package main

import (
	"context"
	"flag"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	configLocation := flag.String("config", "config.json", "configuration file")
	flag.Parse()

	config, err := configuration.Load(*configLocation)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	_, err = lib.New(config, ctx)
	if err != nil {
		log.Println(err)
		cancel()
	}

	go func() {
		shutdown := make(chan os.Signal, 1)
		signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		sig := <-shutdown
		log.Println("received shutdown signal", sig)
		cancel()
	}()

	<-ctx.Done()                //waiting for context end; may happen by shutdown signal
	time.Sleep(1 * time.Second) //give go routines time for cleanup
}
