package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Preparations remain part of retained history after abort and after successful
// publication. Restoration validates their original inputs and selections, not
// today's availability or trust freshness.
type historicalPreparationQuery interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func validatePortablePreparations(ctx context.Context, db historicalPreparationQuery) error {
	after := uuid.Nil
	for {
		var operation, job uuid.UUID
		var references, input, contextBytes, result []byte
		var kind string
		var terminal, hasCatalog bool
		var memberCount int64
		err := db.QueryRow(ctx, `SELECT p.operation_id,o.job_id,o.kind,o.frozen_input,o.final_outcome,o.terminal_at IS NOT NULL,p.canonical_references,p.canonical_context,
 EXISTS(SELECT 1 FROM reference_pack_portable_catalogs c WHERE c.operation_id=p.operation_id),
 (SELECT count(*) FROM reference_pack_operation_members m WHERE m.operation_id=p.operation_id)
 FROM reference_pack_portable_preparations p JOIN reference_pack_operations o USING(operation_id) WHERE p.operation_id>$1 ORDER BY p.operation_id LIMIT 1`, after).Scan(&operation, &job, &kind, &input, &result, &terminal, &references, &contextBytes, &hasCatalog, &memberCount)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		refs, err := DecodeIncidentBundleReferences(references)
		if err != nil {
			return errHistoricalIntegrity
		}
		canonical, err := canonicaljson.Marshal(refs)
		if err != nil || !bytes.Equal(canonical, references) {
			return errHistoricalIntegrity
		}
		frozen, err := packformat.DecodePortabilityInput(input, refs.format())
		if err != nil || kind != "portable_retention" || job != frozen.SourceOperationID || operation != uuid.NewSHA1(job, []byte("reference_pack:incident_retention")) || terminal != (result != nil) || hasCatalog && !terminal {
			return errHistoricalIntegrity
		}
		if _, err := packformat.DecodePortableVerificationContext(contextBytes); err != nil {
			return errHistoricalIntegrity
		}
		if memberCount < 1 || memberCount > int64(len(refs.Versions)) {
			return errHistoricalIntegrity
		}
		descriptors := map[string]packformat.PortableContainer{}
		for _, descriptor := range frozen.Content.Containers {
			descriptors[descriptor.ManifestSHA256] = descriptor
		}
		rows, err := db.Query(ctx, `SELECT s.ordinal,s.pack_key,s.pack_version,s.available,s.reason_code,s.envelope_id,s.verification_ordinal,
 s.expected_container_sha256,s.expected_container_bytes,o.sha256,o.size_bytes,
 (EXISTS(SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=s.input_object_id AND r.owner_kind='operation' AND r.owner_id=s.operation_id::text AND r.logical_path='input/'||s.verification_ordinal::text||'/container')
 OR ($2 AND ((o.available IS FALSE AND o.generation>1 AND NOT EXISTS(SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=o.object_id))
 OR EXISTS(SELECT 1 FROM reference_pack_object_refs r WHERE r.object_id=o.object_id AND r.owner_kind IN ('envelope','version','release','backup'))))),
 EXISTS(SELECT 1 FROM reference_pack_operation_keys k WHERE k.operation_id=s.operation_id AND k.pack_key=s.pack_key)
 FROM reference_pack_portable_selections s LEFT JOIN reference_pack_objects o ON o.object_id=s.input_object_id WHERE s.operation_id=$1 ORDER BY s.ordinal`, operation, terminal)
		if err != nil {
			return err
		}
		selected := 0
		ordinals := map[int64]bool{}
		valid := true
		for rows.Next() {
			var ordinal int
			var key, version string
			var available, retained, guarded bool
			var reason, envelope, expectedDigest, digest *string
			var verificationOrdinal, expectedSize, size *int64
			if err := rows.Scan(&ordinal, &key, &version, &available, &reason, &envelope, &verificationOrdinal, &expectedDigest, &expectedSize, &digest, &size, &retained, &guarded); err != nil {
				rows.Close()
				return err
			}
			if ordinal != selected+1 || ordinal > len(refs.Versions) {
				valid = false
				break
			}
			reference := refs.Versions[ordinal-1]
			if key != reference.Key || version != reference.Version || !guarded || available != (reason == nil) || available && envelope == nil {
				valid = false
				break
			}
			if verificationOrdinal != nil {
				descriptor, present := descriptors[reference.ManifestSHA256]
				if !present || available || ordinals[*verificationOrdinal] || *verificationOrdinal < 1 || *verificationOrdinal > memberCount || !retained || expectedDigest == nil || digest == nil || expectedSize == nil || size == nil || *expectedDigest != descriptor.ContainerSHA256 || *digest != *expectedDigest || *expectedSize != descriptor.SizeBytes || *size != *expectedSize {
					valid = false
					break
				}
				ordinals[*verificationOrdinal] = true
			}
			selected++
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		if !valid || selected != len(refs.Versions) || int64(len(ordinals)) != memberCount {
			return errHistoricalIntegrity
		}
		if terminal && !hasCatalog {
			if packformat.ValidateAttemptResult(result) != nil {
				return errHistoricalIntegrity
			}
			var outcome struct {
				Outcome string `json:"outcome"`
				Members int64  `json:"member_count"`
			}
			if json.Unmarshal(result, &outcome) != nil || outcome.Members != memberCount || outcome.Outcome == "interrupted" || outcome.Outcome == "succeeded" {
				return errHistoricalIntegrity
			}
			if outcome.Outcome == "content_rejected" {
				var completed bool
				if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_attempts WHERE operation_id=$1 AND completed_at IS NOT NULL AND outcome='content_rejected' AND canonical_result=$2)`, operation, result).Scan(&completed); err != nil {
					return err
				}
				if !completed {
					return errHistoricalIntegrity
				}
			}
		}
		if hasCatalog {
			var completed bool
			if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_attempts WHERE operation_id=$1 AND completed_at IS NOT NULL AND outcome IN ('succeeded','content_rejected'))`, operation).Scan(&completed); err != nil {
				return err
			}
			if !completed {
				return errHistoricalIntegrity
			}
		}
		after = operation
	}
}
