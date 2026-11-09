package main

import (
	"context"
	producer "github.com/JochiRaider/cartulary/referencepacks/windows"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	os.Exit(producer.RunCommand(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
