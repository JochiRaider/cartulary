package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesscollect"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessruntime"
)

func main() { os.Exit(run()) }
func run() int {
	if len(os.Args) < 2 {
		return 2
	}
	switch os.Args[1] {
	case "fixtures":
		return harnessfixtures.Main(os.Args[2:])
	case "admission":
		return harnessadmission.Main(os.Args[2:], os.Stdout)
	case "collect":
		if len(os.Args) != 2 {
			return 2
		}
		return harnesscollect.Main(os.Stdin, os.Stdout)
	case "diagnostics":
		return harnessdiag.Main(os.Args[2:], os.Stdout, os.Stderr)
	case "prepare":
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		root, err := os.Getwd()
		if err != nil {
			return 2
		}
		plan, err := harnessruntime.Prepare(ctx, root, os.Args[2:], os.Environ())
		if err != nil {
			fmt.Fprintln(os.Stderr, "graph preparation failed")
			return 2
		}
		fmt.Fprintf(os.Stdout, "graph=%s source=%s units=%d\n", plan.GraphDigest(), plan.SourceDigest(), plan.UnitCount())
		return 0
	default:
		return 2
	}
}
