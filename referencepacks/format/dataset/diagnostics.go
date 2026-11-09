package dataset

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

// A resource guard has already emitted its bounded finding. Further traversal
// is inapplicable, and must not be attempted merely to increase an issue count.
var stop = errors.New("reference dataset: resource guard reached")

func reportFailure(path string, err error, emit schema.FindingSink) error {
	var failure *Failure
	if !errors.As(err, &failure) {
		return err
	}
	f := schema.Finding{Path: path}
	if failure.Limit != "" {
		f = schema.LimitFinding(path, failure.Limit, int64(failure.Maximum), int64(failure.Observed))
	}
	if err := emit(f); err != nil {
		return err
	}
	if failure.Limit != "" {
		return stop
	}
	return nil
}

func guardResult(err error) error {
	if errors.Is(err, stop) || schema.IsLimit(err) {
		return nil
	}
	return err
}

func descriptorRules(d Descriptor, semantics bool, emit schema.FindingSink) error {
	previous := ""
	for i, lookup := range d.Lookups {
		path := "$.descriptor.lookups[" + strconv.Itoa(i) + "]"
		if semantics && lookup.Name <= previous {
			if err := emit(schema.Finding{Path: path + ".name"}); err != nil {
				return err
			}
		}
		previous = lookup.Name
		last := ""
		if !semantics && lookup.InputType == "scalar" && (len(lookup.Components) != 1 || lookup.Components[0].Name != "value") {
			if err := emit(schema.Finding{Path: path + ".components"}); err != nil {
				return err
			}
		}
		for j, component := range lookup.Components {
			p := path + ".components[" + strconv.Itoa(j) + "]"
			if semantics && component.Name <= last {
				if err := emit(schema.Finding{Path: p + ".name"}); err != nil {
					return err
				}
			}
			last = component.Name
			if !semantics && (component.Type == "integer") != (component.Normalization == "safe_integer") {
				if err := emit(schema.Finding{Path: p + ".normalization"}); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func documentFindings(ctx context.Context, raw []byte, kind, path string, terminated bool, emit schema.FindingSink) (valid bool, err error) {
	doc, err := canonicaljson.ScanTokens(ctx, raw, MaxDepth)
	if ctx.Err() != nil {
		return false, ctx.Err()
	}
	if err != nil {
		failure := invalid()
		if errors.Is(err, canonicaljson.ErrDepth) {
			failure = excess("max_dataset_json_depth", MaxDepth, MaxDepth+1, true)
		}
		return false, reportFailure(path, failure, emit)
	}
	value, err := canonicaljson.DecodeStrictBounded(raw, MaxDepth)
	if err != nil {
		return false, reportFailure(path, invalid(), emit)
	}
	valid = true
	var entryID *string
	if kind != "descriptor" {
		if m, ok := value.(map[string]any); ok {
			if id, ok := m[kind+"_id"].(string); ok && len(id) <= 512 && line(id) {
				entryID = &id
			}
		}
	}
	report := func(f schema.Finding) error { valid = false; f.EntryID = entryID; return emit(f) }
	if !doc.SafeIntegers {
		if err := report(schema.Finding{Path: path}); err != nil {
			return false, err
		}
	}
	if err := datasetShapes[kind].Walk(value, path, report); err != nil {
		return false, err
	}
	if !terminated || !doc.Canonical {
		want, actual := "canonical", "noncanonical"
		if err := report(schema.Finding{Path: path, Details: schema.SafeDetails{ExpectedToken: &want, ActualToken: &actual}}); err != nil {
			return false, err
		}
	}
	return valid, nil
}

// StructureFindings enumerates all bounded structural defects. Paths and
// details contain only engine-owned tokens and ordinals, never payload values.
// No sink writes or semantic traversal occur in this phase.
func StructureFindings(ctx context.Context, source Source, expected Expected, emit schema.FindingSink) (err error) {
	if source == nil || emit == nil {
		return errors.New("reference dataset: incomplete structural check")
	}
	defer func() { err = guardResult(err) }()
	if expected.Objects < 0 || expected.Objects > MaxObjects || expected.Relationships < 0 || expected.Relationships > MaxRelationships || expected.SourceArtifacts < 1 || expected.SourceArtifacts > 64 {
		return reportFailure("$", invalid(), emit)
	}
	raw, err := descriptorBytes(ctx, source)
	if err != nil {
		return reportFailure("$.descriptor", err, emit)
	}
	valid, err := documentFindings(ctx, raw, "descriptor", "$.descriptor", true, emit)
	if err != nil {
		return err
	}
	if valid {
		var d Descriptor
		if err := json.Unmarshal(raw, &d); err != nil {
			return err
		}
		if err := descriptorRules(d, false, emit); err != nil {
			return err
		}
	}
	for _, member := range []struct {
		path, kind     string
		count, maximum int64
	}{{ObjectsPath, "object", expected.Objects, MaxObjects}, {RelationshipsPath, "relationship", expected.Relationships, MaxRelationships}} {
		base := "$." + member.kind + "s"
		path := func(i int64) string {
			if i < 0 {
				return base
			}
			return base + "[" + strconv.FormatInt(i, 10) + "]"
		}
		if err := scanRows(ctx, source, member.path, member.count, member.maximum, func(i int64, raw []byte, terminated bool) error {
			_, err := projectRow(ctx, raw, member.kind, path(i), terminated, false, emit)
			return err
		}, func(i int64, err error) error { return reportFailure(path(i), err, emit) }); err != nil {
			return err
		}
	}
	return ctx.Err()
}
