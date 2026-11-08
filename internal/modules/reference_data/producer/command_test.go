package producer

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	attackinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/attack"
)

const producerCLIChildMode = "CARTULARY_ATTACK_PRODUCER_CLI_TEST_CHILD"

func TestMain(m *testing.M) {
	// Only the test executable has a child mode. The production entrypoint has
	// no environment switch or public audit override.
	switch os.Getenv(producerCLIChildMode) {
	case "synthetic":
		os.Exit(runCommand(context.Background(), os.Args[1:], os.Stdout, os.Stderr, syntheticCommandAudit))
	case "production":
		os.Exit(RunCommand(context.Background(), os.Args[1:], os.Stdout, os.Stderr))
	case "":
		os.Exit(m.Run())
	default:
		os.Exit(2)
	}
}

func syntheticCommandAudit(ctx context.Context, root *os.Root, lock, profile, policy []byte, builtAt string, builder Builder) (AuditReport, error) {
	inputs, err := readSyntheticATTACKInputs(lock)
	if err != nil {
		return AuditReport{}, err
	}
	// The dispatcher must still select the current embedded profile and policy.
	currentProfile, err := attackinputs.Files.ReadFile("source-profile.native.v2.r2.json")
	if err != nil {
		return AuditReport{}, err
	}
	if !bytes.Equal(profile, currentProfile) || !bytes.Equal(policy, inputs.policyBytes) {
		return AuditReport{}, errors.New("command did not select current source inputs")
	}
	return auditATTACKSources(ctx, root, inputs, builtAt, builder)
}

func runProducerCLIChild(t *testing.T, mode string, args []string, wantStatus int) (string, string) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Env = append(os.Environ(), producerCLIChildMode+"="+mode)
	cmd.WaitDelay = 5 * time.Second
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run() // Synchronous Run joins the bounded child before temp cleanup.
	if ctx.Err() != nil {
		t.Fatalf("producer child exceeded its deadline: %v", ctx.Err())
	}
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatal(err)
	}
	if cmd.ProcessState == nil || cmd.ProcessState.ExitCode() != wantStatus {
		t.Fatalf("producer child want exit %d: %v; stdout=%s stderr=%s", wantStatus, err, &stdout, &stderr)
	}
	t.Logf("producer child mode=%s native_exit=%d", mode, cmd.ProcessState.ExitCode())
	return stdout.String(), stderr.String()
}

func TestATTACKProducerCLI(t *testing.T) {
	t.Run("dispatch rejects removed flag and arguments", func(t *testing.T) {
		for _, args := range [][]string{{"--candidate", "proposal.json"}, {"unexpected"}, {"--action", "unsupported"}} {
			stdout, stderr := runProducerCLIChild(t, "production", args, 1)
			if stdout != "" || !strings.Contains(stderr, "reference-pack-source:") {
				t.Fatal("invalid command diagnostic lost", stdout, stderr)
			}
		}
	})
	t.Run("retired Base reissue input produces no output", func(t *testing.T) {
		builder, err := BuilderIdentity()
		if err != nil {
			t.Fatal(err)
		}
		directory := t.TempDir()
		inputPath := filepath.Join(directory, "input.json")
		outputDirectory := filepath.Join(directory, "output")
		outputPath := filepath.Join(outputDirectory, "release.json")
		reportPath := filepath.Join(directory, "report.json")
		// The parent and production-dispatch child use the same executable.
		// Valid legacy fields plus an absent source distinguish dispatch rejection
		// from a later input-validation or previous-release read failure.
		input := encoded(t, map[string]any{
			"schema_id":        "cartulary.base_registry_reissue_input.v1",
			"previous_release": filepath.Join(directory, "missing-previous-release.json"),
			"output_file":      outputPath,
			"built_at":         "2026-11-06T00:00:00Z",
			"builder_sha256":   builder.SHA256,
		})
		if err := os.WriteFile(inputPath, input, 0600); err != nil {
			t.Fatal(err)
		}
		stdout, stderr := runProducerCLIChild(t, "production", []string{"--action", "assemble", "--input", inputPath, "--report-file", reportPath}, 1)
		if stdout != "" || stderr != "reference-pack-source: unsupported package input schema\n" {
			t.Fatal("retired Base input reached package execution", stdout, stderr)
		}
		for _, path := range []string{outputPath, outputDirectory, reportPath} {
			if _, err := os.Lstat(path); !os.IsNotExist(err) {
				t.Fatal("retired Base input created output", path, err)
			}
		}
	})
	t.Run("help succeeds", func(t *testing.T) {
		stdout, stderr := runProducerCLIChild(t, "production", []string{"--help"}, 0)
		if stdout != "" || !strings.Contains(stderr, "Usage of reference-pack-producer:") || strings.Contains(stderr, "reference-pack-source:") {
			t.Fatal("help behavior changed", stdout, stderr)
		}
	})
	for _, scenario := range []string{"unbound", "synthetic"} {
		t.Run(scenario+" input produces no report", func(t *testing.T) {
			root := testRoot(t)
			lock := append(bytes.Clone(fixtureBytes(t, "source-lock.v20.1.json")), '\n')
			if scenario == "synthetic" {
				var fetch *fixtureFetcher
				lock, fetch = sourceFixture(t, "valid")
				if _, err := Acquire(t.Context(), root, lock, fetch); err != nil {
					t.Fatal(err)
				}
			}
			args, destination := producerAuditArgs(t, root, lock)
			stdout, stderr := runProducerCLIChild(t, "production", args, 1)
			if stdout != "" || !strings.Contains(stderr, "outside frozen profile") {
				t.Fatal("unbound lock accepted", stdout, stderr)
			}
			for _, path := range []string{destination, filepath.Join(root.Name(), "audits")} {
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Fatal("fabricated report", path, err)
				}
			}
		})
	}
	for _, tc := range []struct {
		name, scenario, status string
		exit                   int
		issues                 []string
	}{
		{"success", "native_created_valid", "source_audit_passed", 0, nil},
		{"attributable failure", "native_created_invalid", "source_audit_failed", 1, []string{nativeToolIssue}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := testRoot(t)
			lock, fetch := sourceFixture(t, tc.scenario)
			if _, err := Acquire(t.Context(), root, lock, fetch); err != nil {
				t.Fatal(err)
			}
			args, destination := producerAuditArgs(t, root, lock)
			inputs := syntheticATTACKInputs(t, lock)
			builder, err := BuilderIdentity()
			if err != nil {
				t.Fatal(err)
			}
			var first []byte
			for range 2 {
				stdout, stderr := runProducerCLIChild(t, "synthetic", args, tc.exit)
				if !strings.Contains(stdout, "source audit: "+tc.status) || (tc.exit == 0 && stderr != "") || (tc.exit == 1 && !strings.Contains(stderr, "reference-pack-source: source audit failed")) {
					t.Fatal("audit diagnostic lost", stdout, stderr)
				}
				got, err := os.ReadFile(destination)
				if err != nil {
					t.Fatal(err)
				}
				var report AuditReport
				if err := DecodeClosed(got, &report); err != nil {
					t.Fatal(err)
				}
				assertAuditReportContract(t, report, inputs)
				if report.Status != tc.status || !slices.Equal(report.Issues, tc.issues) || report.UniqueRecords != 6 || report.LockSHA256 != Digest(lock) || report.PolicySHA256 != Digest(inputs.policyBytes) || report.BuiltAt != "2026-11-06T00:00:00Z" || report.Builder != builder {
					t.Fatalf("report outcome or provenance changed: %+v", report)
				}
				if !bytes.Equal(got, encoded(t, report)) || (first != nil && !bytes.Equal(first, got)) {
					t.Fatal("requested report is not canonical or replay changed its identity")
				}
				assertRetainedCommandReport(t, root, got)
				if !strings.Contains(stdout, "audits/"+Digest(got)+".json") {
					t.Fatal("audit output does not identify retained report", stdout)
				}
				first = got
			}
			t.Run("invalid report destination retains audit", func(t *testing.T) {
				// A fresh retained root proves retention precedes the delivery error;
				// an earlier successful delivery must not hide a missing write.
				failedRoot := testRoot(t)
				if _, err := Acquire(t.Context(), failedRoot, lock, fetch); err != nil {
					t.Fatal(err)
				}
				badArgs, _ := producerAuditArgs(t, failedRoot, lock)
				badArgs[len(badArgs)-1] = failedRoot.Name()
				stdout, stderr := runProducerCLIChild(t, "synthetic", badArgs, 1)
				if stdout != "" || !strings.Contains(stderr, failedRoot.Name()) || !strings.Contains(stderr, "reference-pack-source:") || (tc.exit == 1 && !strings.Contains(stderr, "source audit failed")) {
					t.Fatal("delivery error or original audit error lost", stdout, stderr)
				}
				assertRetainedCommandReport(t, failedRoot, first)
			})
		})
	}
}

func producerAuditArgs(t *testing.T, root *os.Root, lock []byte) ([]string, string) {
	t.Helper()
	dir := t.TempDir()
	lockPath, reportPath := filepath.Join(dir, "lock.json"), filepath.Join(dir, "report.json")
	if err := os.WriteFile(lockPath, lock, 0600); err != nil {
		t.Fatal(err)
	}
	return []string{"--action", "audit", "--lock", lockPath, "--directory", root.Name(), "--built-at", "2026-11-06T00:00:00Z", "--report-file", reportPath}, reportPath
}

func assertRetainedCommandReport(t *testing.T, root *os.Root, want []byte) {
	t.Helper()
	names, err := filepath.Glob(filepath.Join(root.Name(), "audits", "*.json"))
	if err != nil || len(names) != 1 || filepath.Base(names[0]) != Digest(want)+".json" {
		t.Fatal("retained report inventory or identity changed", names, err)
	}
	retained, err := root.ReadFile("audits/" + Digest(want) + ".json")
	if err != nil || !bytes.Equal(retained, want) {
		t.Fatal("retained and requested report bytes differ", err)
	}
}
