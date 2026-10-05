package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func validatePortableCatalogs(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}) error {
	after := uuid.Nil
	for {
		var incident, operation uuid.UUID
		var digest string
		var kind string
		var terminal bool
		var data, resolution, input, outcome []byte
		err := db.QueryRow(ctx, `SELECT c.incident_id,c.operation_id,c.catalog_sha256,c.canonical_references,c.canonical_resolution,o.frozen_input,o.final_outcome,o.kind,o.terminal_at IS NOT NULL
  FROM reference_pack_portable_catalogs c JOIN reference_pack_operations o USING(operation_id)
  WHERE c.incident_id>$1 ORDER BY c.incident_id LIMIT 1`, after).Scan(&incident, &operation, &digest, &data, &resolution, &input, &outcome, &kind, &terminal)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		refs, err := DecodeIncidentBundleReferences(data)
		if err != nil || kind != "portable_retention" || !terminal || packformat.Digest(data) != digest || !bytes.Equal(resolution, outcome) || packformat.ValidatePortabilityRetention(input, resolution, refs.format()) != nil {
			return errHistoricalIntegrity
		}
		var identity struct {
			IncidentID      string `json:"incident_id"`
			SourceOperation string `json:"source_operation_id"`
		}
		if json.Unmarshal(input, &identity) != nil || identity.IncidentID != incident.String() {
			return errHistoricalIntegrity
		}
		source, err := uuid.Parse(identity.SourceOperation)
		if err != nil || uuid.NewSHA1(source, []byte("reference_pack:incident_retention")) != operation {
			return errHistoricalIntegrity
		}
		var result referenceImportResolution
		if json.Unmarshal(resolution, &result) != nil {
			return errHistoricalIntegrity
		}
		expected := 0
		for index, version := range result.Versions {
			if !version.Available {
				continue
			}
			expected++
			var present bool
			member := refs.Versions[index]
			if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_version_pins p JOIN reference_pack_versions v USING(pack_key,pack_version)
   WHERE p.owner_kind='incident_reference_catalog' AND p.owner_id=$1 AND p.operation_id=$2
   AND p.pack_key=$3 AND p.pack_version=$4 AND v.manifest_sha256=$5 AND v.payload_sha256=$6)`, incident.String(), operation, member.Key, member.Version, member.ManifestSHA256, member.PayloadSHA256).Scan(&present); err != nil {
				return err
			}
			if !present {
				return errHistoricalIntegrity
			}
		}
		var count int
		if err := db.QueryRow(ctx, `SELECT count(*) FROM reference_pack_version_pins WHERE owner_kind='incident_reference_catalog' AND owner_id=$1`, incident.String()).Scan(&count); err != nil {
			return err
		}
		if count != expected {
			return errHistoricalIntegrity
		}
		for _, set := range refs.Sets {
			complete := true
			for _, member := range set.Members {
				available := false
				for _, v := range result.Versions {
					if v.Key == member.Key && v.Version == member.Version {
						available = v.Available
						break
					}
				}
				if !available {
					complete = false
					break
				}
			}
			if !complete {
				continue
			}
			var pinned bool
			if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_pins p JOIN reference_pack_sets s USING(pack_set_id)
WHERE p.owner_kind='incident_reference_catalog' AND p.owner_id=$1 AND p.operation_id=$2 AND s.pack_set_id=$3 AND s.pack_set_sha256=$4)`, incident.String(), operation, set.ID, set.SHA256).Scan(&pinned); err != nil {
				return err
			}
			if !pinned {
				return errHistoricalIntegrity
			}
		}
		after = incident
	}
}
