package networkflow

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strconv"
	"testing"
)

func assertSemanticBoundary(t *testing.T) {
	allowedImports := map[string]bool{
		"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction": true,
		"github.com/JochiRaider/cartulary/internal/modules/indicators":            true,
		"github.com/JochiRaider/cartulary/internal/modules/imports":               true,
		"github.com/JochiRaider/cartulary/internal/platform/extensionstore":       true,
		"regexp":          true,
		"bytes":           true,
		"context":         true,
		"crypto/sha256":   true,
		"encoding/binary": true,
		"encoding/hex":    true,
		"encoding/json":   true,
		"errors":          true,
		"fmt":             true,
		"github.com/JochiRaider/cartulary/internal/gen/contractnetworkflow":     true,
		"github.com/JochiRaider/cartulary/internal/gen/networkflowmapping":      true,
		"github.com/JochiRaider/cartulary/internal/modules/graphprojection":     true,
		"github.com/JochiRaider/cartulary/internal/modules/incidents/admission": true,
		"github.com/JochiRaider/cartulary/internal/platform/authn":              true,
		"github.com/JochiRaider/cartulary/internal/platform/jobs":               true,
		"github.com/JochiRaider/cartulary/internal/platform/strictjson":         true,
		"github.com/google/uuid":  true,
		"github.com/jackc/pgx/v5": true,
		"io":                      true,
		"math":                    true,
		"math/big":                true,
		"net/netip":               true,
		"sort":                    true,
		"strconv":                 true,
		"strings":                 true,
		"time":                    true,
		"unicode":                 true,
		"unicode/utf8":            true,
	}

	roots := []string{"graph_materialization_payload.go", "application.go", "table_receipts.go", "indicator_link.go", "indicator_link_admission.go", "indicator_link_graph_admission.go", "indicator_link_receipts.go", "transaction_participants.go", "semantic_failure.go", "query.go", "query_filter.go", "mapping.go", "graph.go", "graph_source_composer.go", "saved_graph_application.go", "graph_temporal.go", "graph_projection_adapter.go", "graph_response_v2.go", "graph_telemetry.go", "table_query_application.go", "graph_query_application.go", "saved_graph_read_application.go", "request_decoding.go"}
	applicationFiles, err := filepath.Glob("*_application.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range applicationFiles {
		if !slices.Contains(roots, path) {
			roots = append(roots, path)
		}
	}
	assertSemanticReferences(t, roots)
	for _, path := range roots {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, dependency := range file.Imports {
			name, err := strconv.Unquote(dependency.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !allowedImports[name] {
				t.Fatalf("semantic file %s imports an unreviewed dependency %s", path, name)
			}
		}
	}
	for _, cause := range []error{context.Canceled, context.DeadlineExceeded} {
		failure := graphProjectionFailedForContext(cause)
		if !errors.Is(failure, cause) {
			t.Fatalf("lost cancellation cause: %v", failure)
		}
	}
	unsafe := &semanticFailure{kind: "unrecognized", reason: "private SQL", cause: errors.New("credential sentinel")}
	public := semanticHTTPError(unsafe)
	if public.Status != 500 || public.Code != "internal_error" || len(public.Details) != 0 {
		t.Fatalf("unknown semantic error did not fail closed: %#v", public)
	}
}
