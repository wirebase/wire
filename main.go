package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/wirebase/wire/project"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get working dir: %v", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go handleInterrupt(cancel)

	prj := project.New(wd, time.Millisecond*500)
	err = prj.Run(ctx)
	if err != nil {
		log.Fatalf("failed to run development server: %v", err)
	}

	println("shutting down")
}

// handleInterrupt will watch for signals and call cancel if an interrupt signal was
// received
func handleInterrupt(cancel context.CancelFunc) {
	sigs := make(chan os.Signal, 0)
	signal.Notify(sigs, os.Interrupt)
	<-sigs
	cancel()
}
