package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ad/anonstickerbot/app"
)

var (
	version = "dev"
)

func main() {
	fmt.Printf("starting version %s\n", version)

	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	done := make(chan bool, 1)

	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	if err := app.Run(ctx, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	go func() {
		sig := <-sigs
		fmt.Println(sig)
		cancel()
		done <- true
	}()

	<-done
	fmt.Println("exiting")
}
