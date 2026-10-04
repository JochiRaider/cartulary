package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"io"
	"slices"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// PortableReferenceSource supplies only this owner's retained artifact
// bindings. Incident Bundles never reads Reporting tables, and neither owner
// resolves the deployment's current Reference Pack selection during export.
type PortableReferenceSource struct {
	Outputs interface {
		Get(context.Context, objectstore.GetObjectRequest) (io.ReadCloser, objectstore.ObjectInfo, error)
	}
}

func (PortableReferenceSource) ReferenceBindingsTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID) ([]reference_data.SetBinding, error) {
	if tx == nil || incident == uuid.Nil {
		return nil, errors.New("reporting: invalid reference export")
	}
	rows, err := tx.Query(ctx, `SELECT snapshot_id,derivation_version,export_model_sha256,export_model_json FROM reporting_snapshots WHERE incident_id=$1 ORDER BY snapshot_id`, incident)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bindings := map[string]reference_data.SetBinding{}
	encoded := map[string][]byte{}
	for rows.Next() {
		var id uuid.UUID
		var derivation, digest string
		var raw []byte
		if err := rows.Scan(&id, &derivation, &digest, &raw); err != nil {
			return nil, err
		}
		canonical, err := canonicalReportingBytes(raw)
		if err != nil {
			return nil, errors.New("reporting: malformed portable snapshot")
		}
		var model SnapshotModel
		if err := json.Unmarshal(canonical, &model); err != nil {
			return nil, err
		}
		encodedModel, err := canonicalJSON(model)
		if err != nil || reportingObjectDigest(model.SchemaID, encodedModel) != digest {
			return nil, errors.New("reporting: altered portable snapshot")
		}
		admitted, err := canonicalReportingBytes(encodedModel)
		if err != nil || !bytes.Equal(admitted, canonical) {
			return nil, errors.New("reporting: unknown portable snapshot members")
		}
		if derivation != DerivationVersion || model.DerivationVersion != derivation || validateSnapshotModelIdentity(model) != nil || model.SnapshotID != id.String() || model.IncidentID != incident.String() {
			return nil, errors.New("reporting: invalid portable snapshot identity")
		}
		binding := model.ReferencePacks
		data, err := canonicaljson.Marshal(binding)
		if err != nil {
			return nil, err
		}
		if previous, ok := encoded[binding.SetID]; ok {
			if !bytes.Equal(previous, data) {
				return nil, errors.New("reporting: conflicting portable reference binding")
			}
			continue
		}
		if len(bindings) >= reference_data.PortableReferenceSetLimit {
			return nil, errors.New("reporting: portable reference set limit exceeded")
		}
		bindings[binding.SetID] = binding
		encoded[binding.SetID] = data
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(bindings))
	for id := range bindings {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	result := make([]reference_data.SetBinding, 0, len(ids))
	for _, id := range ids {
		result = append(result, bindings[id])
	}
	return result, nil
}
