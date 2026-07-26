package rollbackprovider

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

type Provider struct{}

var _ rollbackcontract.RowSourceProvider = Provider{}

func NewProvider() Provider { return Provider{} }

func (Provider) ValidateRollbackValue(value map[string]any) error {
	source, ok := sourceForRollbackValue(value)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if raw, present := source["lifecycle_state"]; present {
		state, valid := raw.(string)
		if !valid || !validLifecycleState(state) {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	if raw, present := source["upload_state"]; present {
		state, valid := raw.(string)
		if !valid || !validUploadState(state) {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	for _, key := range []string{"collector_party_id", "source_party_id", "object_blob_id"} {
		if _, _, err := nullableUUID(source, key); err != nil {
			return rollbackcontract.ErrTargetNotReversible
		}
	}
	return nil
}

func (Provider) RestoreTx(ctx context.Context, tx pgx.Tx, request rollbackcontract.RestoreRequest) error {
	source, ok := sourceForRollbackValue(request.RetainedValue)
	if !ok {
		return rollbackcontract.ErrTargetNotReversible
	}
	if err := (Provider{}).ValidateRollbackValue(request.RetainedValue); err != nil {
		return err
	}
	values := make([]any, 0, 24)
	for _, key := range []string{
		"title", "lifecycle_state", "requested_at", "received_at", "storage_ref", "blob_hash",
		"collector_party_text", "collector_party_id", "source_party_text", "source_party_id",
		"upload_state", "object_blob_id",
	} {
		value, present := source[key]
		if key == "collector_party_id" || key == "source_party_id" || key == "object_blob_id" {
			parsed, _, err := nullableUUID(source, key)
			if err != nil {
				return rollbackcontract.ErrTargetNotReversible
			}
			value = parsed
		}
		values = append(values, present, value)
	}
	_, err := tx.Exec(ctx, `
UPDATE evidence
   SET title = CASE WHEN $2 THEN $3::text ELSE title END,
       lifecycle_state = CASE WHEN $4 THEN $5::text ELSE lifecycle_state END,
       requested_at = CASE WHEN $6 THEN $7::timestamptz ELSE requested_at END,
       received_at = CASE WHEN $8 THEN $9::timestamptz ELSE received_at END,
       storage_ref = CASE WHEN $10 THEN $11::text ELSE storage_ref END,
       blob_hash = CASE WHEN $12 THEN $13::text ELSE blob_hash END,
       collector_party_text = CASE WHEN $14 THEN $15::text ELSE collector_party_text END,
       collector_party_id = CASE WHEN $16 THEN $17::uuid ELSE collector_party_id END,
       source_party_text = CASE WHEN $18 THEN $19::text ELSE source_party_text END,
       source_party_id = CASE WHEN $20 THEN $21::uuid ELSE source_party_id END,
       upload_state = CASE WHEN $22 THEN $23::text ELSE upload_state END,
       object_blob_id = CASE WHEN $24 THEN $25::uuid ELSE object_blob_id END,
       updated_at = $26
 WHERE record_id = $1
`, append([]any{request.RecordID}, append(values, request.Now.UTC())...)...)
	return err
}

func sourceForRollbackValue(value map[string]any) (map[string]any, bool) {
	if source, ok := objectMap(value, "source"); ok {
		return source, len(source) > 0
	}
	if cells, ok := objectMap(value, "cells"); ok {
		source := map[string]any{}
		mapping := map[string]string{
			"evidence.title":                "title",
			"evidence.lifecycle_state":      "lifecycle_state",
			"evidence.requested_at":         "requested_at",
			"evidence.received_at":          "received_at",
			"evidence.storage_ref":          "storage_ref",
			"evidence.blob_hash":            "blob_hash",
			"evidence.collector_party_text": "collector_party_text",
			"evidence.collector_party_id":   "collector_party_id",
			"evidence.source_party_text":    "source_party_text",
			"evidence.source_party_id":      "source_party_id",
			"evidence.upload_state":         "upload_state",
		}
		for fieldKey, sourceKey := range mapping {
			if cell, present := objectMap(cells, fieldKey); present {
				source[sourceKey] = cell["value"]
			}
		}
		return source, len(source) > 0
	}
	if _, ok := value["record_id"]; ok {
		for _, key := range []string{
			"title", "lifecycle_state", "requested_at", "received_at", "storage_ref", "blob_hash",
			"collector_party_text", "collector_party_id", "source_party_text", "source_party_id",
			"upload_state", "object_blob_id",
		} {
			if _, present := value[key]; present {
				return value, true
			}
		}
	}
	return nil, false
}

func objectMap(value map[string]any, key string) (map[string]any, bool) {
	raw, ok := value[key]
	if !ok || raw == nil {
		return nil, false
	}
	typed, ok := raw.(map[string]any)
	return typed, ok
}

func nullableUUID(value map[string]any, key string) (any, bool, error) {
	raw, present := value[key]
	if !present || raw == nil {
		return nil, present, nil
	}
	text, ok := raw.(string)
	if !ok || strings.TrimSpace(text) == "" {
		return nil, true, nil
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return nil, true, err
	}
	return parsed, true, nil
}

func validLifecycleState(value string) bool {
	switch value {
	case "requested", "pending_receipt", "received", "available", "quarantined", "released":
		return true
	default:
		return false
	}
}

func validUploadState(value string) bool {
	switch value {
	case "pending", "available", "failed", "quarantined":
		return true
	default:
		return false
	}
}
