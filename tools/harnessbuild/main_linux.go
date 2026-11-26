//go:build linux

// Command harnessbuild runs registered producers through the harness artifact
// store (P5-3). Make and the operation runner start it with `go run`, which
// needs no prebuilt binary, so it also builds the harness runtime itself.
//
//	harnessbuild produce [--store DIR] PRODUCER...
//
// Each producer is declared in tools/harness_producers.json and publishes at
// its default path in the workspace layout. A caller that sets a producer's
// binding to another path provides that executable, which is checked and
// kept. The store defaults to the layout's artifact_store directory.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/artifactstore"
)

// collection bounds: records unused for 30 days go, and one invocation removes
// at most 16 entries.
const (
	collectAge   = 30 * 24 * time.Hour
	collectLimit = 16
)

const usage = "usage: harnessbuild produce [--store DIR] PRODUCER..."

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 || args[0] != "produce" {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	flags := flag.NewFlagSet("produce", flag.ContinueOnError)
	store := flags.String("store", "", "artifact store directory")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	if err := produce(*store, flags.Args()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		reportFailure()
		return 1
	}
	return 0
}

func produce(store string, ids []string) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	registry, err := artifactstore.ReadRegistry(root)
	if err != nil {
		return err
	}
	workspace, err := artifactstore.OpenWorkspace(root, os.Environ())
	if err != nil {
		return err
	}
	if store == "" {
		if store, err = workspace.Directory("artifact_store"); err != nil {
			return err
		}
	}
	declarations := make([]artifactstore.Declaration, 0, len(ids))
	for _, id := range ids {
		d, ok := registry.Lookup(id)
		if !ok {
			return fmt.Errorf("no registered producer %q", id)
		}
		// A caller that provides the executable at its own path keeps it.
		if _, provided, err := workspace.Provided(d); err != nil {
			return fmt.Errorf("producer %s: %w", id, err)
		} else if !provided {
			declarations = append(declarations, d)
		}
	}
	if slices.ContainsFunc(declarations, artifactstore.Declaration.IsGo) {
		if err := ensureGoToolchain(); err != nil {
			return err
		}
	}
	s := artifactstore.Store{Root: store}
	controls := artifactstore.Controls{
		Disable: os.Getenv("CARTULARY_BUILD_CACHE_DISABLE") == "1",
		Force:   os.Getenv("CARTULARY_FORCE_REBUILD") == "1",
	}
	for _, d := range declarations {
		producer, views, err := workspace.Plan(d)
		if err != nil {
			return fmt.Errorf("producer %s: %w", d.ID, err)
		}
		if _, err := s.Produce(producer, views, controls); err != nil {
			return fmt.Errorf("producer %s: %w", d.ID, err)
		}
	}
	_, err = s.Collect(time.Now(), collectAge, collectLimit)
	return err
}

// ensureGoToolchain applies the Go toolchain readiness owner before a Go
// build: machine-state paths, cache corruption markers and the pinned
// toolchain.
func ensureGoToolchain() error {
	command := exec.Command("bash", "tools/harness/readiness/go-toolchain-readiness.sh", "ensure")
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("go toolchain readiness: %w", err)
	}
	return nil
}

// reportFailure records a build failure in the command's failure context.
func reportFailure() {
	if os.Getenv("CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT") == "" || os.Getenv("NODE_BIN") == "" {
		return
	}
	command := exec.Command(os.Getenv("NODE_BIN"), "tools/harness/runtime/command-failure-cli.mjs", "harness", "tool_diagnostic_failure")
	command.Stdout, command.Stderr = os.Stderr, os.Stderr
	_ = command.Run()
}
