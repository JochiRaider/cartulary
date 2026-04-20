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
	seenTests := make(map[string]map[string]struct{})

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

		if entry.Action != "run" || entry.Package == "" || entry.Test == "" {
			continue
		}
		if strings.Contains(entry.Test, "/") {
			continue
		}
		if !strings.HasPrefix(entry.Test, "Test") && !strings.HasPrefix(entry.Test, "Benchmark") && !strings.HasPrefix(entry.Test, "Fuzz") {
			continue
		}

		if _, ok := seenTests[entry.Package]; !ok {
			seenTests[entry.Package] = make(map[string]struct{})
			order = append(order, entry.Package)
		}
		if _, ok := seenTests[entry.Package][entry.Test]; ok {
			continue
		}

		seenTests[entry.Package][entry.Test] = struct{}{}
		testsByPackage[entry.Package] = append(testsByPackage[entry.Package], entry.Test)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("phase inventory parse failed: read go test json: %w", err)
	}

	matchedTests := 0
	for _, pkg := range order {
		matchedTests += len(testsByPackage[pkg])
	}
	if matchedTests == 0 {
		return fmt.Errorf("phase matched zero tests")
	}

	fmt.Fprintf(stdout, "matched go tests: %d across %d packages\n", matchedTests, len(order))
	for _, pkg := range order {
		fmt.Fprintf(stdout, "  %s\n", pkg)
		for _, testName := range testsByPackage[pkg] {
			fmt.Fprintf(stdout, "    %s\n", testName)
		}
	}

	_ = stderr
	return nil
}

func openInput(path string) (io.Reader, func(), error) {
	if path == "-" {
		return os.Stdin, func() {}, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil, fmt.Errorf("open go test json log %q: %w", path, err)
	}

	return file, func() {
		_ = file.Close()
	}, nil
}
