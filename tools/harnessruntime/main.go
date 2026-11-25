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

	"github.com/JochiRaider/cartulary/internal/testutil/harnessaccount"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesscollect"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessplan"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessreview"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessstress"
	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

func main() { os.Exit(run()) }
func run() int {
	if len(os.Args) < 2 {
		return 2
	}
	switch os.Args[1] {
	case "inspect-go":
		if len(os.Args) != 2 {
			return 2
		}
		if err := sourcecheck.InspectBatch(os.Stdin, os.Stdout); err != nil {
			fmt.Fprintln(os.Stderr, "Go source inspection failed")
			return 1
		}
		return 0
	case "account":
		return harnessaccount.Main(os.Args[2:], os.Stdout, os.Stderr)
	case "recover", "recover-status":
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
		if os.Args[1] == "recover-status" {
			return harnessfixtures.RecoveryStatus(root, environment, os.Stdout)
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
		report, code, err := harnessruntime.Run(ctx, root, binary, os.Args[2:], os.Environ(), os.Stderr)
		if report != nil && report.CleanupDiagnostic != nil {
			fmt.Fprintln(os.Stderr, report.CleanupDiagnostic)
		}
		if err != nil {
			for _, line := range harnessruntime.PreparationDiagnosticLines(err) {
				fmt.Fprintln(os.Stderr, line)
			}
			class, reason := harnessruntime.PublicFailure(err)
			fmt.Fprintf(os.Stderr, "[GRAPH-FAIL] failure_class=%s failure_reason=%s failure_stage=%s\n", class, reason, harnessruntime.PublicFailureStage(err))
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
				// TH-HARNESS-REQ-860: the bounded distinct violation list is part
				// of the summary; repeat it beside the failing status line. A
				// server fatal (TH-HARNESS-REQ-863) names its reason instead of
				// an operation.
				if contract, ok := summary["contract_evidence"].(map[string]any); ok {
					if violations, ok := contract["violations"].([]any); ok {
						for _, entry := range violations {
							if v, ok := entry.(map[string]any); ok {
								if v["kind"] == "server_fatal" {
									fmt.Fprintf(out, "[SERVER-FATAL] reason_code=%v process_kind=%v occurrences=%v\n", v["reason_code"], v["process_kind"], v["occurrences"])
									continue
								}
								fact := fmt.Sprintf("code=%v", v["code"])
								if security, ok := v["security"].(map[string]any); ok {
									fact = fmt.Sprintf("declared=%v observed=%v", security["declared"], security["observed"])
								}
								fmt.Fprintf(out, "[CONTRACT-VIOLATION] %v operation=%v status=%v %s occurrences=%v\n", v["kind"], v["operation_id"], v["status"], fact, v["occurrences"])
							}
						}
					}
				}
			}
		}
		return code
	case "stress":
		// TH-HARNESS-REQ-871: repeat an owner slice beside a declared load.
		if len(os.Args) != 2 {
			return 2
		}
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
		return harnessstress.Main(ctx, root, binary, os.Environ(), os.Stdout, os.Stderr)
	case "stress-load":
		return harnessstress.LoadMain(os.Args[2:], os.Stderr)
	case "review":
		return harnessreview.Main(os.Args[2:], os.Stdin, os.Stdout)
	case "session":
		return harnesssession.Main(os.Args[2:], os.Stdin, os.Stdout)
	case "process":
		return harnesssession.ProcessMain(os.Args[2:], os.Stdin, os.Stderr)
	case "owned-command":
		return harnesssession.OwnedCommandMain(os.Args[2:], os.Stdout, os.Stderr)
	case "admission":
		return harnessadmission.Main(os.Args[2:], os.Stdout)
	case "collect":
		if len(os.Args) != 2 {
			return 2
		}
		return harnesscollect.Main(os.Stdin, os.Stdout)
	case "diagnostics":
		return harnessdiag.Main(os.Args[2:], os.Stdout, os.Stderr)
	case "plan":
		// Diagnostic and maintenance tools read the plan the runtime would
		// execute; nothing is launched or published.
		root, err := os.Getwd()
		if err != nil {
			return 2
		}
		plan, err := harnessruntime.PlanInvocation(root, os.Args[2:], os.Environ())
		if err != nil {
			for _, line := range harnessruntime.PreparationDiagnosticLines(err) {
				fmt.Fprintln(os.Stderr, line)
			}
			fmt.Fprintln(os.Stderr, "graph preparation failed")
			return 2
		}
		if err := json.NewEncoder(os.Stdout).Encode(plan); err != nil {
			return 1
		}
		return 0
	case "capacities":
		root, err := os.Getwd()
		if err != nil {
			return 2
		}
		override := ""
		switch {
		case len(os.Args) == 4 && os.Args[2] == "--override" && os.Args[3] != "":
			override = os.Args[3]
		case len(os.Args) != 2:
			return 2
		}
		capacity, err := harnessplan.CaptureHostCapacity(root, override)
		if err != nil {
			fmt.Fprintln(os.Stderr, "capacity capture failed")
			return 2
		}
		if err := json.NewEncoder(os.Stdout).Encode(capacity); err != nil {
			return 1
		}
		return 0
	default:
		return 2
	}
}
