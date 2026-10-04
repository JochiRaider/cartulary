package packformat

import (
	"context"
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var metadataShapes = map[string]*shape{
	"root":      compileProjection("tuf_root_metadata.v1.schema.json"),
	"timestamp": compileProjection("tuf_timestamp_metadata.v1.schema.json"),
	"snapshot":  compileProjection("tuf_snapshot_metadata.v1.schema.json"),
	"targets":   compileProjection("tuf_targets_metadata.v1.schema.json"),
}

func metadataSchemaFindings(data []byte, kind string, maxSignatures int, path string, emit FindingSink) (map[string]any, error) {
	invalid := false
	report := func(f Finding) error { invalid = true; return emit(f) }
	if len(data) > 2097152 {
		if err := report(LimitFinding(path, "max_metadata_file_bytes", 2097152, int64(len(data)))); err != nil {
			return nil, err
		}
		return nil, fail("tuf_metadata_invalid")
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		if err := report(Finding{Path: path}); err != nil {
			return nil, err
		}
		return nil, fail("tuf_metadata_invalid")
	}
	rule := metadataShapes[kind]
	if rule == nil {
		return nil, errors.New("reference pack: unknown metadata role")
	}
	if err := rule.walk(value, path, report); err != nil {
		return nil, err
	}
	if !validPackNumbers(value) {
		if err := report(Finding{Path: path}); err != nil {
			return nil, err
		}
	}
	if invalid {
		return nil, fail("tuf_metadata_invalid")
	}
	m := value.(map[string]any)
	signatures := m["signatures"].([]any)
	if len(signatures) > maxSignatures {
		limit := "max_signatures"
		if maxSignatures == 128 {
			limit = "max_root_update_signatures"
		}
		if err := report(LimitFinding(path+".signatures", limit, int64(maxSignatures), int64(len(signatures)))); err != nil {
			return nil, err
		}
	}
	previous := ""
	for i, item := range signatures {
		key := item.(map[string]any)["keyid"].(string)
		if key <= previous {
			if err := report(Finding{Path: path + ".signatures[" + strconv.Itoa(i) + "].keyid"}); err != nil {
				return nil, err
			}
		}
		previous = key
	}
	if invalid {
		return nil, fail("tuf_metadata_invalid")
	}
	return m, nil
}

// Schema diagnostics enumerate files in canonical path order. Dynamic member
// names are always represented by the declared unknown-member segment; the
// metadata ordinal names the bounded sorted input without echoing hostile paths.
func TrustSchemaSummary(ctx context.Context, files map[string][]byte, scratch DiagnosticScratch) (*ValidationSummary, error) {
	return CheckSummary(ctx, "tuf_schema", scratch, func(ctx context.Context, emit FindingSink) error {
		return trustSchemaFindings(ctx, files, emit, nil)
	})
}

func trustSchemaFindings(ctx context.Context, files map[string][]byte, emit FindingSink, admitted map[string]metadata) error {
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	for _, role := range []string{"timestamp", "snapshot", "targets"} {
		path := "metadata/" + role + ".json"
		if _, exists := files[path]; !exists {
			if err := emit(Finding{Path: path}); err != nil {
				return err
			}
		}
	}
	total := 0
	for i, name := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := "$.metadata[" + strconv.Itoa(i) + "]"
		data := files[name]
		if len(data) > 8388608-total {
			return emit(LimitFinding(path, "max_metadata_total_bytes", 8388608, int64(total)+int64(len(data))))
		}
		total += len(data)
		kind, maxSignatures := "", 64
		switch name {
		case "metadata/targets.json":
			kind = "targets"
		case "metadata/snapshot.json":
			kind = "snapshot"
		case "metadata/timestamp.json":
			kind = "timestamp"
		default:
			if strings.HasPrefix(name, "metadata/") && strings.HasSuffix(name, ".root.json") {
				kind, maxSignatures = "root", 128
			}
		}
		if kind == "" {
			if err := emit(Finding{Path: path}); err != nil {
				return err
			}
			continue
		}
		fields, err := metadataSchemaFindings(data, kind, maxSignatures, path, emit)
		var failure *Failure
		if err != nil && !errors.As(err, &failure) {
			return err
		}
		if err == nil && admitted != nil {
			m, err := metadataFromFields(data, fields)
			if err != nil {
				return err
			}
			admitted[name] = m
		}
	}
	return nil
}
