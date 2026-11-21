package evidence

import (
	"context"
	"errors"
	"iter"
	"maps"
	"reflect"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	evidenceblob "github.com/JochiRaider/cartulary/internal/modules/evidence/internal/blob"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

type portableBlobRewrite struct {
	members *incidentportability.BoundMembers
}

func (p preparedEvidenceImport) bundleBlobRaw(ctx context.Context, d sourceport.Descriptor, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, portableBlobPath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return d.DeclaredFailure("evidence.row_shape_exact")
	}
	return err
}
func (p preparedEvidenceImport) bundleBlobs(ctx context.Context, d sourceport.Descriptor, visit func(evidenceblob.Row) error) error {
	return p.bundleBlobRaw(ctx, d, func(raw map[string]any) error {
		row, ok := evidenceblob.ParseRow(raw)
		if !ok {
			return d.DeclaredFailure("evidence.row_shape_exact")
		}
		return visit(row)
	})
}
func (p preparedEvidenceImport) prepareBlobs(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if err := p.bundleBlobRaw(ctx, d, func(map[string]any) error { return nil }); err != nil {
		return err
	}
	previous := ""
	if err := p.bundleBlobRaw(ctx, d, func(raw map[string]any) error {
		value, ok := raw["object_blob_id"].(string)
		id, err := uuid.Parse(value)
		if !ok || err != nil || id == uuid.Nil || id.String() != value || value <= previous {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		previous = value
		return nil
	}); err != nil {
		return err
	}
	if err := p.bundleBlobRaw(ctx, d, func(raw map[string]any) error {
		if _, present := raw["storage_key"]; present {
			return d.DeclaredFailure("evidence.storage_reference_legal")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := validateIncidentBundleBlobObservations(ctx, p.members, bundle, d); err != nil {
		return err
	}
	for _, phase := range []struct {
		invariant string
		valid     func(evidenceblob.Row) bool
	}{
		{"row_shape_exact", func(evidenceblob.Row) bool { return true }},
		{"envelope_type_scope", func(r evidenceblob.Row) bool { return r.IncidentID == p.incidentID }},
		{"lifecycle_legal", func(r evidenceblob.Row) bool { return r.Coherent() }},
		{"attribution_exact", func(r evidenceblob.Row) bool { _, ok := input.Actors.Lookup(r.CreatedByUserID.String()); return ok }},
	} {
		if err := p.bundleBlobs(ctx, d, func(row evidenceblob.Row) error {
			if !phase.valid(row) {
				return d.DeclaredFailure("evidence." + phase.invariant)
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

// Pull only one prepared source row alongside each rewritten row. The iterator
// is always stopped, including on rejection, so both verified readers close.
func (p preparedEvidenceImport) rewrittenBlobs(ctx context.Context, rewritten incidentportability.MemberReader, d sourceport.Descriptor, visit func(evidenceblob.Row, string) error) error {
	stopped := errors.New("blob comparison stopped")
	next, stop := iter.Pull2(func(yield func(evidenceblob.Row, error) bool) {
		err := p.blobs(ctx, d, func(row evidenceblob.Row) error {
			if !yield(row, nil) {
				return stopped
			}
			return nil
		})
		if err != nil && !errors.Is(err, stopped) {
			yield(evidenceblob.Row{}, err)
		}
	})
	defer stop()
	// Core 01 REQ-01-695: an absent rewritten member is the empty rewrite, and a
	// present member must carry at least one row.
	present, walked := slices.ContainsFunc(rewritten.Members(), func(member incidentportability.Member) bool { return member.Path == portableBlobPath }), 0
	walk := func(visit func(map[string]any) error) error {
		if !present {
			return nil
		}
		return incidentportability.WalkPath(ctx, rewritten, portableBlobPath, visit)
	}
	err := walk(func(raw map[string]any) error {
		walked++
		expected, err, ok := next()
		if err != nil {
			return err
		}
		if !ok {
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		key, ok := raw["storage_key"].(string)
		parts, err := blobref.ParseObjectBlobStorageKey(key)
		if !ok || err != nil || parts.IncidentID != expected.IncidentID || parts.ObjectBlobID != expected.ObjectBlobID {
			return d.DeclaredFailure("evidence.storage_reference_legal")
		}
		source := maps.Clone(raw)
		delete(source, "storage_key")
		actual, ok := evidenceblob.ParseRow(source)
		if !ok || !reflect.DeepEqual(actual, expected) {
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		return visit(expected, key)
	})
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return d.DeclaredFailure("evidence.source_state_exact")
	}
	if err != nil {
		return err
	}
	if present && walked == 0 {
		return d.DeclaredFailure("evidence.source_state_exact")
	}
	_, err, ok := next()
	if err != nil {
		return err
	}
	if ok {
		return d.DeclaredFailure("evidence.source_state_exact")
	}
	return nil
}

// bindRewrittenBlobs binds the rewritten Blob member when staging wrote one.
// Its absence binds nothing: the empty rewrite (Core 01 REQ-01-695).
func bindRewrittenBlobs(source incidentportability.MemberReader) (*incidentportability.BoundMembers, error) {
	if source != nil {
		for _, member := range source.Members() {
			if member.Path == portableBlobPath {
				return incidentportability.BindMembers(source, []string{portableBlobPath})
			}
		}
	}
	return incidentportability.BindMembers(source, nil)
}

func (p preparedEvidenceImport) bindBlobRewrite(ctx context.Context, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if p.blobRewrite == nil {
		return sourceport.ErrPreparedBinding
	}
	bound, err := bindRewrittenBlobs(input.RewrittenObjectBlobs)
	if err != nil {
		return err
	}
	if p.blobRewrite.members != nil && !reflect.DeepEqual(p.blobRewrite.members.Members(), bound.Members()) {
		return sourceport.ErrPreparedBinding
	}
	if err := p.rewrittenBlobs(ctx, bound, d, func(evidenceblob.Row, string) error { return nil }); err != nil {
		return err
	}
	if p.blobRewrite.members == nil {
		p.blobRewrite.members = bound
	}
	return nil
}
