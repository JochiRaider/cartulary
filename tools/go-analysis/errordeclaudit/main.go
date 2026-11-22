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
	check := flag.Bool("check-writes", false, "check proven operation/status/code writes")
	flag.Parse()
	if *root == "" || (!*check && *out == "") {
		fmt.Fprintln(os.Stderr, "errordeclaudit requires -dir and -out")
		os.Exit(2)
	}
	catalog, err := readCatalog(filepath.Join(*root, *catalogPath))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	config := Config{
		Dir:       *root,
		Patterns:  []string{*pattern},
		APIError:  httpapiPackage + ".APIError",
		CodeField: "Code",
		Writers: map[string]int{
			httpapiPackage + ".WriteError":             3,
			httpapiPackage + ".WriteErrorWithConflict": 3,
			httpapiPackage + ".WriteErrorWithOptions":  3,
		},
		AdmittedPrincipal:  "github.com/JochiRaider/cartulary/internal/platform/httpauth.AdmittedPrincipal",
		AdmittedPathString: httpapiPackage + ".PathString",
		HandlerType:        "net/http.HandlerFunc",
		Catalog:            catalog,
	}
	if *check {
		result, err := checkWrites(config, httpapiPackage+".BindAdmittedOwnerRoutes")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("operations=%d proven_writes=%d violations=%d\n", len(catalog), len(result.Writes), len(result.Violations))
		if len(result.Writes) == 0 {
			fmt.Fprintln(os.Stderr, "error declaration qualification failed: zero proven server writes (TH-HARNESS-REQ-868)")
			fmt.Fprintf(os.Stderr, "reference requests with an admitted session: getViewSchema GET /api/v1/view-schemas/%s?limit=1 and listDeploymentExtensions GET /api/v1/extensions?limit=1\n", samplePathValue)
			fmt.Fprintln(os.Stderr, "reproduce with: make go-error-declaration-check")
			os.Exit(1)
		}
		for _, v := range result.Writes {
			fmt.Printf("proven operation=%s status=%d code=%s writer=%s\n", v.Operation, v.Status, v.Code, v.Writer)
		}
		for _, v := range result.Violations {
			fmt.Fprintf(os.Stderr, "operation=%s status=%d code=%s writer=%s\n", v.Operation, v.Status, v.Code, v.Writer)
		}
		if len(result.Violations) > 0 {
			os.Exit(1)
		}
		return
	}
	report, err := Run(config)
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
