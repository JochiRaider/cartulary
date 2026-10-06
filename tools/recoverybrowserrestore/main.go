package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/testutil/recoverybrowsertest"
)

func main() {
	root := flag.String("runtime-root", "", "private browser fixture runtime root")
	flag.Parse()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := recoverybrowsertest.Run(ctx, *root, os.Stdin, os.Stdout)
	stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "restore browser restore target: %v\n", err)
		os.Exit(1)
	}
}
