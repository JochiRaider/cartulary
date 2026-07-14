package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/app/operator"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Exit(operator.RunOperatorCLIContext(ctx, os.Args[1:], os.Stdout, os.Stderr))
}
