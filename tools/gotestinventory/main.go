package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

type event struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

type testState struct {
	Status string
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: gotestinventory <go-test-json-log-file|-")
		os.Exit(2)
	}

	if err := run(os.Args[1], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(path string, stdout io.Writer, stderr io.Writer) error {
	reader, closeFn, err := openInput(path)
	if err != nil {
		return err
	}
	defer closeFn()

	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	order := make([]string, 0)
	testsByPackage := make(map[string][]string)
	testStates := make(map[string]map[string]*testState)

	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		var entry event
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return fmt.Errorf("phase inventory parse failed: invalid go test json on line %d: %w", lineNumber, err)
		}

		if entry.Package == "" || entry.Test == "" {
			continue
		}
		if strings.Contains(entry.Test, "/") {
			continue
		}
		if !strings.HasPrefix(entry.Test, "Test") && !strings.HasPrefix(entry.Test, "Benchmark") && !strings.HasPrefix(entry.Test, "Fuzz") {
			continue
		}

		if _, ok := testStates[entry.Package]; !ok {
			testStates[entry.Package] = make(map[string]*testState)
			order = append(order, entry.Package)
		}
		if _, ok := testStates[entry.Package][entry.Test]; !ok {
			testStates[entry.Package][entry.Test] = &testState{}
			testsByPackage[entry.Package] = append(testsByPackage[entry.Package], entry.Test)
		}
		switch entry.Action {
		case "run":
			if testStates[entry.Package][entry.Test].Status == "" {
				testStates[entry.Package][entry.Test].Status = "run"
			}
		case "pass", "fail", "skip":
			testStates[entry.Package][entry.Test].Status = entry.Action
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("phase inventory parse failed: read go test json: %w", err)
	}

	matchedTests := 0
	skipped := make([]string, 0)
	failed := make([]string, 0)
	incomplete := make([]string, 0)
	for _, pkg := range order {
		for _, testName := range testsByPackage[pkg] {
			state := testStates[pkg][testName]
			switch state.Status {
			case "pass":
				matchedTests++
			case "skip":
				skipped = append(skipped, fmt.Sprintf("%s/%s", pkg, testName))
			case "fail":
				failed = append(failed, fmt.Sprintf("%s/%s", pkg, testName))
			default:
				incomplete = append(incomplete, fmt.Sprintf("%s/%s", pkg, testName))
			}
		}
	}
	if matchedTests == 0 {
		if len(skipped) > 0 || len(failed) > 0 || len(incomplete) > 0 {
			return fmt.Errorf(
				"go test inventory requires top-level pass: skipped=%s failed=%s incomplete=%s",
				joinValues(skipped),
				joinValues(failed),
				joinValues(incomplete),
			)
		}
		return fmt.Errorf("phase matched zero tests")
	}
	if len(skipped) > 0 || len(failed) > 0 || len(incomplete) > 0 {
		return fmt.Errorf(
			"go test inventory requires top-level pass: skipped=%s failed=%s incomplete=%s",
			joinValues(skipped),
			joinValues(failed),
			joinValues(incomplete),
		)
	}

	fmt.Fprintf(stdout, "matched go tests: %d across %d packages\n", matchedTests, len(order))
	for _, pkg := range order {
		printedPackage := false
		for _, testName := range testsByPackage[pkg] {
			if testStates[pkg][testName].Status != "pass" {
				continue
			}
			if !printedPackage {
				fmt.Fprintf(stdout, "  %s\n", pkg)
				printedPackage = true
			}
			fmt.Fprintf(stdout, "    %s\n", testName)
		}
	}

	_ = stderr
	return nil
}

func joinValues(values []string) string {
	if len(values) == 0 {
		return "none"
	}
	return strings.Join(values, ", ")
}

func openInput(path string) (io.Reader, func(), error) {
	if path == "-" {
		return os.Stdin, func() {}, nil
	}

	file, err := os.Open(path) // #nosec G304 -- gotestinventory intentionally accepts an explicit go test JSON log path or stdin.
	if err != nil {
		return nil, nil, fmt.Errorf("open go test json log %q: %w", path, err)
	}

	return file, func() {
		_ = file.Close()
	}, nil
}
