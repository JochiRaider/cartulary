package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/JochiRaider/cartulary/internal/platform/openapicompat"
)

func main() {
	root := flag.String("root", ".", "repository root")
	reportPath := flag.String("report", "", "optional report output path")
	reportOnly := flag.Bool("report-only", false, "emit an integrity-checked diff without enforcing version or change-set policy")
	quiet := flag.Bool("quiet", false, "emit only a concise success summary")
	flag.Parse()

	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "openapi compatibility failed: resolve root: %v\n", err)
		os.Exit(1)
	}
	var report openapicompat.Report
	if *reportOnly {
		report, err = openapicompat.ReportRepository(absoluteRoot)
	} else {
		report, err = openapicompat.CheckRepository(absoluteRoot)
	}
	if err != nil {
		if !*reportOnly {
			if diagnosticReport, reportErr := openapicompat.ReportRepository(absoluteRoot); reportErr == nil {
				if payload, encodeErr := json.MarshalIndent(diagnosticReport, "", "  "); encodeErr == nil {
					payload = append(payload, '\n')
					if *reportPath != "" {
						_ = os.MkdirAll(filepath.Dir(*reportPath), 0o700)
						_ = writeReportAtomically(*reportPath, payload)
					} else {
						_, _ = os.Stdout.Write(payload)
					}
				}
			}
		}
		fmt.Fprintf(os.Stderr, "openapi compatibility failed: %v\n", err)
		os.Exit(1)
	}
	payload, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "openapi compatibility failed: encode report: %v\n", err)
		os.Exit(1)
	}
	payload = append(payload, '\n')
	if *reportPath != "" {
		if err := os.MkdirAll(filepath.Dir(*reportPath), 0o700); err != nil {
			fmt.Fprintf(os.Stderr, "openapi compatibility failed: create report directory: %v\n", err)
			os.Exit(1)
		}
		if err := writeReportAtomically(*reportPath, payload); err != nil {
			fmt.Fprintf(os.Stderr, "openapi compatibility failed: write report: %v\n", err)
			os.Exit(1)
		}
	}
	if *quiet {
		fmt.Printf(
			"OpenAPI compatibility passed: %s -> %s (%d reviewed changes)\n",
			report.BaselineVersion,
			report.TargetVersion,
			len(report.Changes),
		)
		return
	}
	if _, err := os.Stdout.Write(payload); err != nil {
		fmt.Fprintf(os.Stderr, "openapi compatibility failed: write stdout: %v\n", err)
		os.Exit(1)
	}
}

func writeReportAtomically(path string, payload []byte) error {
	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer func() {
		_ = os.Remove(temporaryName)
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(payload); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}
