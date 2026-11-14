package evidence

import (
	"context"
	"errors"
	"iter"
	"maps"
	"reflect"

	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/google/uuid"
)

type portableBlobRewrite struct {
	members *incidentportability.BoundMembers
}

func (p preparedEvidenceImport) blobRaw(ctx context.Context, d sourceport.Descriptor, visit func(map[string]any) error) error {
	err := incidentportability.WalkPath(ctx, p.members, portableBlobPath, visit)
	var malformed *incidentportability.StrictNDJSONDecodeError
	if errors.As(err, &malformed) {
		return d.DeclaredFailure("evidence.row_shape_exact")
	}
	return err
}
func (p preparedEvidenceImport) blobs(ctx context.Context, d sourceport.Descriptor, visit func(portableBlobRow) error) error {
	return p.blobRaw(ctx, d, func(raw map[string]any) error {
		row, ok := parsePortableBlobRow(raw)
		if !ok {
			return d.DeclaredFailure("evidence.row_shape_exact")
		}
		return visit(row)
	})
}
func (p preparedEvidenceImport) prepareBlobs(ctx context.Context, bundle sourceport.Bundle, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if err := p.blobRaw(ctx, d, func(map[string]any) error { return nil }); err != nil {
		return err
	}
	previous := ""
	if err := p.blobRaw(ctx, d, func(raw map[string]any) error {
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
	if err := p.blobRaw(ctx, d, func(raw map[string]any) error {
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
		valid     func(portableBlobRow) bool
	}{
		{"row_shape_exact", func(portableBlobRow) bool { return true }},
		{"envelope_type_scope", func(r portableBlobRow) bool { return r.IncidentID == p.incidentID }},
		{"lifecycle_legal", func(r portableBlobRow) bool { return r.lifecycleLegal() }},
		{"attribution_exact", func(r portableBlobRow) bool { _, ok := input.Actors.Lookup(r.CreatedByUserID.String()); return ok }},
	} {
		if err := p.blobs(ctx, d, func(row portableBlobRow) error {
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
func (p preparedEvidenceImport) rewrittenBlobs(ctx context.Context, rewritten incidentportability.MemberReader, d sourceport.Descriptor, visit func(portableBlobRow, string) error) error {
	stopped := errors.New("blob comparison stopped")
	next, stop := iter.Pull2(func(yield func(portableBlobRow, error) bool) {
		err := p.blobs(ctx, d, func(row portableBlobRow) error {
			if !yield(row, nil) {
				return stopped
			}
			return nil
		})
		if err != nil && !errors.Is(err, stopped) {
			yield(portableBlobRow{}, err)
		}
	})
	defer stop()
	err := incidentportability.WalkPath(ctx, rewritten, portableBlobPath, func(raw map[string]any) error {
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
		actual, ok := parsePortableBlobRow(source)
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
	_, err, ok := next()
	if err != nil {
		return err
	}
	if ok {
		return d.DeclaredFailure("evidence.source_state_exact")
	}
	return nil
}

func (p preparedEvidenceImport) bindBlobRewrite(ctx context.Context, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if p.blobRewrite == nil {
		return sourceport.ErrPreparedBinding
	}
	bound, err := incidentportability.BindMembers(input.RewrittenObjectBlobs, []string{portableBlobPath})
	if err != nil {
		return err
	}
	if p.blobRewrite.members != nil && !reflect.DeepEqual(p.blobRewrite.members.Members(), bound.Members()) {
		return sourceport.ErrPreparedBinding
	}
	if err := p.rewrittenBlobs(ctx, bound, d, func(portableBlobRow, string) error { return nil }); err != nil {
		return err
	}
	if p.blobRewrite.members == nil {
		p.blobRewrite.members = bound
	}
	return nil
}
