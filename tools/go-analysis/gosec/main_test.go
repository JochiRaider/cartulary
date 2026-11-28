package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

type reportedIssue struct {
	RuleID string `json:"rule_id"`
	File   string `json:"file"`
	Line   string `json:"line"`
}

type jsonReport struct {
	Issues []reportedIssue  `json:"Issues"`
	Stats  map[string]int   `json:"Stats"`
	Errors map[string][]any `json:"Golang errors"`
}

// fixtureDir is resolved before any test changes directory.
var fixtureDir = func() string {
	dir, err := filepath.Abs(filepath.Join("testdata", "fixture"))
	if err != nil {
		panic(err)
	}
	return dir
}()

// scanFixture runs the driver in the fixture module, as the wrappers run it
// at the repository root.
func scanFixture(t *testing.T, args ...string) (int, jsonReport, string) {
	t.Helper()
	t.Chdir(fixtureDir)
	var stdout, stderr bytes.Buffer
	code := run(args, &stdout, &stderr)
	var report jsonReport
	if strings.Contains(strings.Join(args, " "), "-fmt=json") && stdout.Len() > 0 {
		if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
			t.Fatalf("report %q: %v", stdout.String(), err)
		}
	}
	return code, report, stderr.String()
}

func files(issues []reportedIssue) []string {
	var names []string
	for _, found := range issues {
		names = append(names, found.RuleID+" "+filepath.ToSlash(found.File[strings.Index(found.File, "fixture"):])+":"+found.Line)
	}
	slices.Sort(names)
	return names
}

// The seeded finding proves the driver loads and checks packages: suppressed,
// build-excluded and generated code report nothing, a nested module loads from
// its own root, and an unsuppressed finding fails the scan.
func TestDriverReportsTheSeededFinding(t *testing.T) {
	code, report, stderr := scanFixture(t, "-include=G401", "-exclude-generated", "-fmt=json", "-quiet", "./...")
	if code != exitFailure {
		t.Fatalf("exit %d with findings: %s", code, stderr)
	}
	want := []string{"G401 fixture/nested/inner/hash.go:6", "G401 fixture/seeded/hash.go:7"}
	if got := files(report.Issues); !slices.Equal(got, want) {
		t.Fatalf("issues %v, want %v", got, want)
	}
	if report.Stats["nosec"] != 1 || report.Stats["found"] != 2 || report.Stats["files"] != 2 || len(report.Errors) != 0 {
		t.Fatalf("stats %v errors %v", report.Stats, report.Errors)
	}
}

func TestDriverScopesAndNoFail(t *testing.T) {
	code, report, _ := scanFixture(t, "-include=G401", "-fmt=json", "-no-fail", "-exclude-dir=nested", "./seeded", "./generated")
	if code != exitSuccess {
		t.Fatalf("-no-fail exit %d", code)
	}
	want := []string{"G401 fixture/generated/hash.go:8", "G401 fixture/seeded/hash.go:7"}
	if got := files(report.Issues); !slices.Equal(got, want) {
		t.Fatalf("issues %v, want %v", got, want)
	}
}

func TestDriverQuietAndCleanScan(t *testing.T) {
	code, _, stderr := scanFixture(t, "-include=G101", "-quiet", "./...")
	if code != exitSuccess || stderr != "" {
		t.Fatalf("a clean quiet scan exited %d with %q", code, stderr)
	}
}

// Only the flags the wrappers use are accepted, so no CLI flag the driver
// does not implement can be silently ignored.
func TestDriverRefusesUnsupportedFlags(t *testing.T) {
	for _, args := range [][]string{{"-exclude=G401", "./..."}, {"-tags=x", "./..."}, {"-include=G401"}, {"-concurrency=0", "./..."}} {
		if code, _, _ := scanFixture(t, args...); code != exitFailure {
			t.Fatalf("%v exited %d", args, code)
		}
	}
}
