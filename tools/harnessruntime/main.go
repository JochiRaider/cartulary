package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesscollect"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessreview"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func main() { os.Exit(run()) }
func run() int {
	if len(os.Args) < 2 {
		return 2
	}
	switch os.Args[1] {
	case "recover":
		if len(os.Args) != 2 {
			return 2
		}
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		root, err := os.Getwd()
		if err != nil {
			return 2
		}
		environment := map[string]string{}
		for _, entry := range os.Environ() {
			if key, value, ok := strings.Cut(entry, "="); ok {
				environment[key] = value
			}
		}
		return harnessfixtures.RecoverRun(ctx, root, environment, os.Stdout)
	case "graph":
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		root, err := os.Getwd()
		if err != nil {
			return 2
		}
		binary, err := os.Executable()
		if err != nil {
			return 2
		}
		report, code, err := harnessruntime.Run(ctx, root, binary, os.Args[2:], os.Environ())
		if report != nil && report.CleanupDiagnostic != nil {
			fmt.Fprintln(os.Stderr, report.CleanupDiagnostic)
		}
		if err != nil {
			class, reason := harnessruntime.PublicFailure(err)
			fmt.Fprintf(os.Stderr, "[GRAPH-FAIL] failure_class=%s failure_reason=%s\n", class, reason)
			return code
		}
		if report != nil {
			summary := report.Summary
			if report.JSONOutput || os.Getenv("CARTULARY_OUTPUT_MODE") == "machine" {
				_ = json.NewEncoder(os.Stdout).Encode(summary)
			} else if os.Getenv("CARTULARY_OUTPUT_MODE") != "quiet" || code != 0 {
				out := os.Stdout
				if code != 0 {
					out = os.Stderr
				}
				results := os.Getenv("CARTULARY_TEST_RESULTS_DIR")
				if results == "" {
					results = ".cartulary/test-results"
				}
				fmt.Fprintf(out, "[GRAPH] target=%v status=%v duration_ms=%v run_root=%s\n", summary["target"], summary["status"], summary["wall_duration_ms"], filepath.Join(results, fmt.Sprint(summary["run_id"])))
			}
		}
		return code
	case "review":
		return harnessreview.Main(os.Args[2:], os.Stdin, os.Stdout)
	case "session":
		return harnesssession.Main(os.Args[2:], os.Stdin, os.Stdout)
	case "process":
		return harnesssession.ProcessMain(os.Args[2:], os.Stdin, os.Stderr)
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
