//go:build linux

// Command harnessbuild runs registered producers through the harness artifact
// store (P5-3). It is itself the registered producer "harnessbuild" (P6-5):
// launchers run its published executable, which first checks that the store
// root for its current key holds its own bytes and otherwise hands the same
// command to `go run ./tools/harnessbuild`. A run that is not the published
// executable publishes it afterwards, so only the first run after a change to
// its sources compiles it. It also builds the harness runtime itself.
//
//	harnessbuild produce [--store DIR] [--roots FILE] PRODUCER...
//
// Each producer is declared in tools/harness_producers.json and publishes at
// its default path in the workspace layout. The producers a selected one
// requires run first. A caller that sets a producer's binding to another path
// provides that file, which is checked and kept. Installations (the Node
// runtime, the frontend dependencies, the Playwright browsers and the test
// service images) install in place under a receipt; under the installed-only
// preparation policy they install nothing and fail when missing or stale.
// The report renderer keeps its own content-addressed installations, whose
// layout the Reporting NLSpec owns, and the release inventory regenerates
// each time it is produced. The store defaults to the
// layout's artifact_store directory. With --roots, the sealed root of each
// store producer is written to FILE as a JSON object, for a caller that
// copies a root, such as the run's frontend artifact.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"sync"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/artifactstore"
)

// launcherID is this command's own producer.
const launcherID = "harnessbuild"

// collection bounds: records unused for 30 days go, and one invocation removes
// at most 16 entries.
const (
	collectAge   = 30 * 24 * time.Hour
	collectLimit = 16
)

const usage = "usage: harnessbuild produce [--store DIR] [--roots FILE] [--no-cache] [--force] PRODUCER..."

func main() { os.Exit(run(os.Args[1:])) }

func run(args []string) int {
	if len(args) == 0 || args[0] != "produce" {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	flags := flag.NewFlagSet("produce", flag.ContinueOnError)
	store := flags.String("store", "", "artifact store directory")
	roots := flags.String("roots", "", "file to write each producer's sealed root to")
	noCache := flags.Bool("no-cache", false, "neither read nor record store keys")
	force := flags.Bool("force", false, "rebuild and record, ignoring a recorded key")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() == 0 {
		fmt.Fprintln(os.Stderr, usage)
		return 2
	}
	controls := artifactstore.Controls{Disable: *noCache, Force: *force}
	if err := produce(*store, *roots, controls, flags.Args(), args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		reportFailure()
		return 1
	}
	return 0
}

func produce(store, rootsFile string, controls artifactstore.Controls, ids, args []string) error {
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
	// The Go toolchain is readied once, before the first Go command; a Go
	// binary whose key and root the store answers runs none.
	workspace.GoReady = sync.OnceValue(ensureGoToolchain)
	if store == "" {
		if store, err = workspace.Directory("artifact_store"); err != nil {
			return err
		}
	}
	s := artifactstore.Store{Root: store}
	launcher, resident, err := residentLauncher(workspace, registry)
	if err != nil {
		return err
	}
	if resident {
		producer, _, err := workspace.Plan(launcher)
		if err != nil {
			return err
		}
		if executable, err := os.Executable(); err != nil || !s.Holds(producer, launcher.Name, executable) {
			return compileAndRun(workspace, args)
		}
	}
	selected, err := registry.Closure(ids)
	if err != nil {
		return err
	}
	declarations := make([]artifactstore.Declaration, 0, len(selected))
	for _, d := range selected {
		// A caller that provides the file at its own path keeps it.
		if _, provided, err := workspace.Provided(d); err != nil {
			return fmt.Errorf("producer %s: %w", d.ID, err)
		} else if !provided {
			declarations = append(declarations, d)
		}
	}
	if slices.ContainsFunc(declarations, func(d artifactstore.Declaration) bool { return d.IsGo() && d.Kind != "go_binary" }) {
		if err := workspace.GoReady(); err != nil {
			return err
		}
	}
	installedOnly := os.Getenv("CARTULARY_PREPARATION_POLICY") == "installed_only"
	produced := map[string]string{}
	for _, d := range declarations {
		switch {
		case d.IsInstallation():
			installation, lock, err := workspace.PlanInstallation(d)
			if err != nil {
				return fmt.Errorf("producer %s: %w", d.ID, err)
			}
			if _, err := artifactstore.Install(installation, lock, controls, installedOnly); err != nil {
				return fmt.Errorf("producer %s: %w", d.ID, err)
			}
			continue
		case d.Kind == "report_renderer":
			renderer, err := workspace.PlanRenderer(d, s, controls)
			if err == nil {
				err = renderer.Produce()
			}
			if err != nil {
				return fmt.Errorf("producer %s: %w", d.ID, err)
			}
			continue
		case d.Kind == "release_inventory":
			inventory, err := workspace.PlanInventory(d)
			if err == nil {
				err = inventory.Generate()
			}
			if err != nil {
				return fmt.Errorf("producer %s: %w", d.ID, err)
			}
			continue
		}
		producer, views, err := workspace.Plan(d)
		if err != nil {
			return fmt.Errorf("producer %s: %w", d.ID, err)
		}
		result, err := s.Produce(producer, views, controls)
		if err != nil {
			return fmt.Errorf("producer %s: %w", d.ID, err)
		}
		produced[d.ID] = s.RootPath(result.Root)
	}
	if rootsFile != "" {
		data, err := json.Marshal(produced)
		if err != nil {
			return err
		}
		if err := os.WriteFile(rootsFile, append(data, '\n'), 0o600); err != nil {
			return err
		}
	}
	// A compiled run publishes the launcher for the next one. It needs the
	// Node runtime for the binary's receipt, so the run that installs Node
	// leaves it to a later run.
	if !resident && !controls.Disable && workspace.Getenv("NODE_BIN") != "" && !slices.Contains(ids, launcherID) {
		producer, views, err := workspace.Plan(launcher)
		if err != nil {
			return fmt.Errorf("producer %s: %w", launcherID, err)
		}
		if _, err := s.Produce(producer, views, artifactstore.Controls{}); err != nil {
			return fmt.Errorf("producer %s: %w", launcherID, err)
		}
	}
	_, err = s.Collect(time.Now(), collectAge, collectLimit)
	return err
}

// residentLauncher returns the launcher's declaration and whether this
// process is its published executable.
func residentLauncher(workspace *artifactstore.Workspace, registry artifactstore.Registry) (artifactstore.Declaration, bool, error) {
	launcher, ok := registry.Lookup(launcherID)
	if !ok {
		return launcher, false, fmt.Errorf("the producer registry declares no %s producer", launcherID)
	}
	view, err := workspace.View(launcher)
	if err != nil {
		return launcher, false, err
	}
	executable, err := os.Executable()
	if err != nil {
		return launcher, false, nil
	}
	published, errPublished := os.Stat(view)
	running, errRunning := os.Stat(executable)
	return launcher, errPublished == nil && errRunning == nil && os.SameFile(published, running), nil
}

// compileAndRun replaces this stale launcher with `go run` of the current
// sources and the same arguments.
func compileAndRun(workspace *artifactstore.Workspace, args []string) error {
	goExecutable := workspace.Getenv("GO")
	if goExecutable == "" {
		goExecutable = "go"
	}
	resolved, err := exec.LookPath(goExecutable)
	if err != nil {
		return err
	}
	return syscall.Exec(resolved, append([]string{goExecutable, "run", "./tools/harnessbuild"}, args...), os.Environ())
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
