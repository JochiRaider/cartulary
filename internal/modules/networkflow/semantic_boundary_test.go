package networkflow

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

func assertSemanticBoundary(t *testing.T) {
	allowedImports := map[string]bool{
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

	for _, path := range []string{"semantic_failure.go", "query.go", "query_filter.go", "mapping.go", "graph.go", "graph_source_composer.go", "saved_graph_application.go", "graph_temporal.go", "graph_projection_adapter.go", "graph_response_v2.go", "graph_telemetry.go"} {
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
