package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(app.RunServerContext(ctx, os.Stdout, os.Stderr))
}
