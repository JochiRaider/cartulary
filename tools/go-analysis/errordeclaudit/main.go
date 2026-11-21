// Command errordeclaudit is the advisory static error-declaration audit
// (Testing Harness TH-HARNESS-REQ-867). It reports, for each operation of the
// generated operation catalog, the error codes its handler can produce but
// does not declare. Findings never change its exit status.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

const httpapiPackage = "github.com/JochiRaider/cartulary/internal/platform/httpapi"

func main() {
	root := flag.String("dir", "", "repository root")
	pattern := flag.String("pattern", "./cmd/server", "server program package pattern")
	catalogPath := flag.String("catalog", "internal/gen/openapioperations/catalog_gen.go", "generated operation catalog, relative to -dir")
	out := flag.String("out", "", "report path")
	flag.Parse()
	if *root == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "errordeclaudit requires -dir and -out")
		os.Exit(2)
	}
	catalog, err := readCatalog(filepath.Join(*root, *catalogPath))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	report, err := Run(Config{
		Dir:       *root,
		Patterns:  []string{*pattern},
		APIError:  httpapiPackage + ".APIError",
		CodeField: "Code",
		Writers: map[string]int{
			httpapiPackage + ".WriteError":             3,
			httpapiPackage + ".WriteErrorWithConflict": 3,
			httpapiPackage + ".WriteErrorWithOptions":  3,
		},
		HandlerType: "net/http.HandlerFunc",
		Catalog:     catalog,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	encoded, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*out, append(encoded, '\n'), 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("error-declaration-audit operations=%d rooted=%d unrooted=%d findings=%d\n",
		report.OperationCount, report.RootedCount, len(report.UnrootedOperations), report.FindingCount)
}
