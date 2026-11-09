package packformat

import (
	"context"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
	"io"
	"time"
)

// Test composition uses the production registry and check builders.
func extractWithDiagnostics(ctx context.Context, source io.ReaderAt, size int64, limits ArchiveLimits, destination Destination, scratch DiagnosticScratch) (Inventory, error) {
	return extract(ctx, source, size, limits, destination, scratch, true)
}

// decodeManifestWithDiagnostics uses the same admission rules as every
// retained/built-in reader. Only the execution boundary needs the complete
// winning-check diagnostic stream and its confined external merge workspace.
func decodeManifestWithDiagnostics(ctx context.Context, data []byte, operator bool, scratch DiagnosticScratch) (Manifest, error) {
	program, result := ManifestChecks(data, int64(len(data)), operator)
	if err := runCheckRange(ctx, "manifest_encoding", "manifest_canonical", scratch, program, true); err != nil {
		return Manifest{}, err
	}
	return result(), nil
}

// Hint shape admission is separate from canonical bytes, matching the registry.
func decodeHintWithDiagnostics(ctx context.Context, data []byte, scratch DiagnosticScratch) (string, error) {
	program, result := HintChecks(data, int64(len(data)))
	if err := runCheckRange(ctx, "bundle_shape", "bundle_canonical", scratch, program, true); err != nil {
		return "", err
	}
	return result(), nil
}

// contentSchemaSummary enumerates the schema check across every bounded row.
// Schema evaluation and ordinary admission share shape.walk. No semantic,
// licensing, dependency or compatibility check executes in this pass.
func contentSchemaSummary(ctx context.Context, m Manifest, source ContentSource, scratch DiagnosticScratch) (*ValidationSummary, error) {
	return CheckSummary(ctx, "content_schema", scratch, func(ctx context.Context, emit FindingSink) error {
		return ContentSchemaFindings(ctx, m, source, emit)
	})
}

// verifyDependencies executes all three dependency ranks before registry
// compatibility or indexing. Reachable state is already retained trusted
// content; hostile candidate declarations never supply recursive manifests.
func verifyDependencies(ctx context.Context, candidate Manifest, resolve DependencyResolver, scratch DiagnosticScratch) error {
	program, err := DependencyChecks(candidate, resolve)
	if err != nil {
		return err
	}
	return runCheckRange(ctx, "dependencies", "conflicts", scratch, program, true)
}

// Schema diagnostics enumerate files in canonical path order. Dynamic member
// names are always represented by the declared unknown-member segment; the
// metadata ordinal names the bounded sorted input without echoing hostile paths.
func trustSchemaSummary(ctx context.Context, files map[string][]byte, scratch DiagnosticScratch) (*ValidationSummary, error) {
	return CheckSummary(ctx, "tuf_schema", scratch, func(ctx context.Context, emit FindingSink) error {
		return fixtureTrustSchemaFindings(ctx, files, emit)
	})
}

func verifyTrustWithDiagnostics(ctx context.Context, metadataFiles map[string][]byte, inventory Inventory, repository string, at time.Time, snapshot TrustSnapshot, scratch DiagnosticScratch) (TrustProposal, error) {
	return verifyTrust(ctx, metadataFiles, inventory, repository, at, snapshot, scratch, true)
}

func compileShape(value map[string]any, ancestors map[string]bool) *shape {
	raw, err := canonicaljson.Marshal(value)
	if err != nil {
		panic(err)
	}
	resources := projectionResources()
	resources["inline.schema.json"] = raw
	return &shape{schema.Compile(resources, fixedShapeLimits, "inline.schema.json")}
}
