package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/recoverybrowsertest"
)

func main() {
	root := flag.String("runtime-root", "", "private browser fixture runtime root")
	attempt := flag.String("attempt-id", "", "owning browser fixture attempt")
	cleanupMS := flag.Int64("cleanup-budget-ms", 0, "owning browser process cleanup allowance")
	flag.Parse()
	if *attempt == "" || *cleanupMS <= 0 || *cleanupMS > 86400000 {
		fmt.Fprintln(os.Stderr, "attempt-id and a bounded cleanup-budget-ms are required")
		os.Exit(1)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	err := recoverybrowsertest.Run(ctx, *root, os.Stdin, os.Stdout, recoverybrowsertest.NewLifecycle(*attempt, os.Stderr, time.Duration(*cleanupMS)*time.Millisecond))
	stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "restore browser restore target: %v\n", err)
		os.Exit(1)
	}
}
