package evidence

import (
	"context"
	"reflect"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/jackc/pgx/v5"
)

func (p preparedEvidenceImport) insertBlobs(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if err := p.blobs(ctx, d, func(row portableBlobRow) error {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM object_blobs WHERE object_blob_id=$1)`, row.ObjectBlobID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		return nil
	}); err != nil {
		return err
	}
	return p.rewrittenBlobs(ctx, p.blobRewrite.members, d, func(row portableBlobRow, key string) error {
		tag, err := tx.Exec(ctx, `INSERT INTO object_blobs(object_blob_id,incident_id,created_by_user_id,storage_key,upload_state,byte_size,filename_hint,content_type_hint,expected_sha256_hex,observed_size,observed_content_type,observed_sha256_hex,target_expires_at,pending_expires_at,finalized_at,terminal_reason,failed_at,cleanup_due_at,cleaned_up_at,created_at,updated_at)
VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`, row.ObjectBlobID, row.IncidentID, p.actorID, key, row.UploadState, row.ByteSize, row.FilenameHint, row.ContentTypeHint, row.ExpectedSHA256Hex, row.ObservedSize, row.ObservedContentType, row.ObservedSHA256Hex, row.TargetExpiresAt, row.PendingExpiresAt, row.FinalizedAt, row.TerminalReason, row.FailedAt, row.CleanupDueAt, row.CleanedUpAt, row.CreatedAt, row.UpdatedAt)
		if err != nil {
			return err
		}
		if tag.RowsAffected() != 1 {
			return d.DeclaredFailure("evidence.source_identity_admitted")
		}
		return input.Attributions.RecordImportedAttribution("object_blobs", row.ObjectBlobID.String(), "created_by_user_id", row.CreatedByUserID.String())
	})
}

func (p preparedEvidenceImport) compareBlobs(ctx context.Context, tx pgx.Tx, input sourceport.ImportContext, d sourceport.Descriptor) error {
	if p.blobRewrite == nil || p.blobRewrite.members == nil {
		return sourceport.ErrPreparedBinding
	}
	bound, err := incidentportability.BindMembers(input.RewrittenObjectBlobs, []string{portableBlobPath})
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(bound.Members(), p.blobRewrite.members.Members()) {
		return sourceport.ErrPreparedBinding
	}
	rows, err := tx.Query(ctx, `SELECT object_blob_id,incident_id,created_by_user_id,storage_key,upload_state,byte_size,filename_hint,content_type_hint,expected_sha256_hex,observed_size,observed_content_type,observed_sha256_hex,target_expires_at,pending_expires_at,finalized_at,terminal_reason,failed_at,cleanup_due_at,cleaned_up_at,created_at,updated_at FROM object_blobs WHERE incident_id=$1 ORDER BY object_blob_id`, p.incidentID)
	if err != nil {
		return err
	}
	defer rows.Close()
	err = p.rewrittenBlobs(ctx, p.blobRewrite.members, d, func(expected portableBlobRow, key string) error {
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return err
			}
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		var actual portableBlobRow
		var actualKey string
		if err := rows.Scan(&actual.ObjectBlobID, &actual.IncidentID, &actual.CreatedByUserID, &actualKey, &actual.UploadState, &actual.ByteSize, &actual.FilenameHint, &actual.ContentTypeHint, &actual.ExpectedSHA256Hex, &actual.ObservedSize, &actual.ObservedContentType, &actual.ObservedSHA256Hex, &actual.TargetExpiresAt, &actual.PendingExpiresAt, &actual.FinalizedAt, &actual.TerminalReason, &actual.FailedAt, &actual.CleanupDueAt, &actual.CleanedUpAt, &actual.CreatedAt, &actual.UpdatedAt); err != nil {
			return err
		}
		actual.utc()
		expected.CreatedByUserID = p.actorID
		if actualKey != key || !reflect.DeepEqual(actual, expected) {
			return d.DeclaredFailure("evidence.source_state_exact")
		}
		return nil
	})
	if err != nil {
		return err
	}
	if rows.Next() {
		return d.DeclaredFailure("evidence.source_state_exact")
	}
	return rows.Err()
}

func (p preparedEvidenceImport) validateBlobAttributions(ctx context.Context, recorder incidentportability.AttributionRecorder, d sourceport.Descriptor) error {
	var expectedCount, actualCount uint64
	if err := p.blobs(ctx, d, func(row portableBlobRow) error {
		expectedCount++
		seen := false
		if err := recorder.VisitImportedAttributions("object_blobs", row.ObjectBlobID.String(), func(a incidentportability.ImportedAttribution) error {
			if seen || a.SourceTable != "object_blobs" || a.SourceRowID != row.ObjectBlobID.String() || a.SourceColumn != "created_by_user_id" || a.SourceActorID != row.CreatedByUserID.String() || a.LocalUserID != p.actorID {
				return d.DeclaredFailure("evidence.attribution_exact")
			}
			seen = true
			return ctx.Err()
		}); err != nil {
			return err
		}
		if !seen {
			return d.DeclaredFailure("evidence.attribution_exact")
		}
		return nil
	}); err != nil {
		return err
	}
	if err := recorder.VisitImportedAttributions("object_blobs", "", func(incidentportability.ImportedAttribution) error { actualCount++; return ctx.Err() }); err != nil {
		return err
	}
	if actualCount != expectedCount {
		return d.DeclaredFailure("evidence.attribution_exact")
	}
	return nil
}
