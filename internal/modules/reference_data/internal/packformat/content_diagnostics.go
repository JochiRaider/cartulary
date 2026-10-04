package packformat

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// ContentSemanticsFindings visits every admitted row even after a semantic
// failure. It emits no source values and never starts index construction.
func ContentSemanticsFindings(ctx context.Context, m Manifest, source ContentSource, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: missing semantic findings sink")
	}
	return validateContentPhase(ctx, m, source, nil, contentSemantics, emit)
}

func RegistryCompatibilityFindings(ctx context.Context, m Manifest, source ContentSource, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: missing registry findings sink")
	}
	if profiles[m.Key].Kind != "type_registry" {
		return nil
	}
	return validateContentPhase(ctx, m, source, nil, contentRegistry, emit)
}

// ContentSchemaSummary enumerates the schema check across every bounded row.
// Schema evaluation and ordinary admission share shape.walk. No semantic,
// licensing, dependency or compatibility check executes in this pass.
func ContentSchemaSummary(ctx context.Context, m Manifest, source ContentSource, scratch DiagnosticScratch) (*ValidationSummary, error) {
	return CheckSummary(ctx, "content_schema", scratch, func(ctx context.Context, emit FindingSink) error {
		return ContentSchemaFindings(ctx, m, source, emit)
	})
}

func ContentSchemaFindings(ctx context.Context, m Manifest, source ContentSource, emit FindingSink) error {
	if source == nil || emit == nil {
		return errors.New("reference pack: incomplete content schema check")
	}
	profile, ok := profiles[m.Key]
	if !ok {
		return errors.New("reference pack: content schema lacks admitted profile")
	}
	kinds := []string{"entry"}
	if profile.Shape == "objects_relationships" {
		kinds = []string{"object", "relationship"}
	}
	for _, kind := range kinds {
		path, diagnosticPath, expected := "payload/entries.ndjson", "$.entries", m.Summary.Entries
		if kind == "object" {
			path, diagnosticPath, expected = "payload/objects.ndjson", "$.objects", m.Summary.Objects
		}
		if kind == "relationship" {
			path, diagnosticPath, expected = "payload/relationships.ndjson", "$.relationships", m.Summary.Relationships
		}
		if expected == nil || !admittedContentCount(kind, *expected) {
			return errors.New("reference pack: content schema lacks admitted count")
		}
		reader, err := source.Open(ctx, path)
		if err != nil {
			return ErrStorage
		}
		boundedStop, scanErr := enumerateContentRows(ctx, reader, contentShapes[m.Key+"/"+kind], diagnosticPath, kind, *expected, emit)
		closeErr := reader.Close()
		if scanErr != nil || closeErr != nil {
			return errors.Join(scanErr, closeErr)
		}
		if boundedStop {
			break
		}
	}
	return nil
}

func enumerateContentRows(ctx context.Context, source io.Reader, shape *shape, path, kind string, expected int64, emit FindingSink) (bool, error) {
	reader := bufio.NewReaderSize(contextReader{ctx, source}, 1048576)
	count := int64(0)
	for {
		line, err := reader.ReadSlice('\n')
		if ctx.Err() != nil {
			return false, ctx.Err()
		}
		if errors.Is(err, io.EOF) && len(line) == 0 {
			break
		}
		rowPath := path + "[" + strconv.FormatInt(count, 10) + "]"
		if errors.Is(err, bufio.ErrBufferFull) {
			// The full buffer has no LF: even an immediately following LF would
			// exceed the inclusive line bound. Stop without reading that byte.
			return true, emit(LimitFinding(rowPath, "max_line_bytes", 1048576, 1048577))
		}
		if err != nil && !errors.Is(err, io.EOF) {
			return false, ErrStorage
		}
		count++
		if count > expected {
			want, actual := strconv.FormatInt(expected, 10), strconv.FormatInt(count, 10)
			return true, emit(Finding{Path: path, Details: SafeDetails{ExpectedToken: &want, ActualToken: &actual}})
		}
		terminated := err == nil
		data := line
		if terminated {
			data = line[:len(line)-1]
		}
		value, decodeErr := canonicaljson.DecodeStrict(data)
		if decodeErr != nil {
			if err := emit(Finding{Path: rowPath}); err != nil {
				return false, err
			}
		} else {
			var entryID *string
			if object, ok := value.(map[string]any); ok {
				name := "entry_id"
				if kind == "object" {
					name = "object_id"
				}
				if id, ok := object[name].(string); ok && len(id) <= 512 && singleLine(id, 512) {
					if rule := shape.properties[name]; rule != nil && rule.matches(id) {
						entryID = &id
					}
				}
			}
			rowEmit := func(f Finding) error { f.EntryID = entryID; return emit(f) }
			if !validPackNumbers(value) {
				if err := rowEmit(Finding{Path: rowPath}); err != nil {
					return false, err
				}
			}
			if err := shape.walk(value, rowPath, rowEmit); err != nil {
				return false, err
			}
			canonical, canonicalErr := canonicaljson.Marshal(value)
			if !terminated || canonicalErr != nil || !bytes.Equal(canonical, data) {
				want, actual := "canonical", "noncanonical"
				if err := rowEmit(Finding{Path: rowPath, Details: SafeDetails{ExpectedToken: &want, ActualToken: &actual}}); err != nil {
					return false, err
				}
			}
		}
		if !terminated {
			break
		}
	}
	if count != expected {
		want, actual := strconv.FormatInt(expected, 10), strconv.FormatInt(count, 10)
		return false, emit(Finding{Path: path, Details: SafeDetails{ExpectedToken: &want, ActualToken: &actual}})
	}
	return false, nil
}
