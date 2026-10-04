package reference_data

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type attestationVersions struct {
	Root      *int64 `json:"root"`
	Targets   *int64 `json:"targets"`
	Snapshot  *int64 `json:"snapshot"`
	Timestamp *int64 `json:"timestamp"`
}

// This is the closed historical evidence object. Operation correlation lives
// in its relational binding, not an extra member of the signed identity.
type packAttestation struct {
	SchemaID      string              `json:"schema_id"`
	ID            string              `json:"attestation_id"`
	Kind          string              `json:"event_kind"`
	Key           string              `json:"pack_key"`
	Version       string              `json:"pack_version"`
	Distribution  string              `json:"distribution_kind"`
	Method        string              `json:"verification_method"`
	Result        string              `json:"result"`
	At            time.Time           `json:"occurred_at"`
	ActorKind     string              `json:"actor_kind"`
	Actor         *uuid.UUID          `json:"actor_user_id"`
	Operator      *uuid.UUID          `json:"operator_operation_id"`
	Container     *string             `json:"container_sha256"`
	Manifest      *string             `json:"manifest_sha256"`
	Payload       *string             `json:"payload_sha256"`
	SourceProfile *string             `json:"source_profile_id"`
	SourceDigest  *string             `json:"source_profile_sha256"`
	Repository    *string             `json:"trust_repository_id"`
	Versions      attestationVersions `json:"trusted_metadata_versions"`
	Signers       []string            `json:"verified_signer_key_ids"`
	ValidUntil    *time.Time          `json:"trust_valid_until"`
	Summary       *string             `json:"validation_summary_ref"`
	Previous      *string             `json:"prior_active_version"`
	Set           *string             `json:"resulting_pack_set_id"`
}

type attestationInput struct {
	Operation                  uuid.UUID
	Key, Version, Kind, Result string
	At                         time.Time
	PreviousSet                string
	ResultingSet               *string
	Container                  *string
	Verification               bool
	Summary                    *string
}

func appendPackAttestationTx(ctx context.Context, tx pgx.Tx, input attestationInput) ([]byte, error) {
	if !slices.Contains([]string{"import_verification", "reverification", "refresh_verification", "activation", "rollback_activation", "safety_fallback", "profile_reconciliation", "dependency_invalidation", "disablement", "removal", "exact_reimport", "payload_invalidation", "trust_root_update"}, input.Kind) || !slices.Contains([]string{"succeeded", "failed", "rejected"}, input.Result) || input.At.IsZero() {
		return nil, errors.New("reference pack: invalid attestation event")
	}
	a := packAttestation{SchemaID: "cartulary.reference_pack_attestation.v1", Kind: input.Kind, Key: input.Key, Version: input.Version, Result: input.Result, At: input.At.UTC(), Set: input.ResultingSet, Container: input.Container, Signers: []string{}, Summary: input.Summary}
	var manifest, encoded []byte
	if err := tx.QueryRow(ctx, `SELECT o.actor_kind,o.actor_user_id,c.distribution_kind,v.manifest_bytes,e.canonical_envelope
 FROM reference_pack_operations o CROSS JOIN reference_pack_candidates c LEFT JOIN reference_pack_versions v USING(pack_key,pack_version)
 LEFT JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id
 WHERE o.operation_id=$1 AND c.pack_key=$2 AND c.pack_version=$3`, input.Operation, input.Key, input.Version).Scan(&a.ActorKind, &a.Actor, &a.Distribution, &manifest, &encoded); err != nil {
		return nil, err
	}
	if a.ActorKind == "local_operator" {
		a.Operator = &input.Operation
	}
	a.Method = "packaged_release_manifest_v1"
	if a.Distribution == "operator_imported" {
		a.Method = "tuf_1_0_35_offline_bundle_v1"
	}
	if input.PreviousSet != "" {
		if err := tx.QueryRow(ctx, `SELECT (SELECT pack_version FROM reference_pack_set_members WHERE pack_set_id=$1 AND pack_key=$2)`, input.PreviousSet, input.Key).Scan(&a.Previous); err != nil {
			return nil, err
		}
	}
	if manifest != nil {
		m, err := packformat.DecodeManifest(manifest, a.Distribution == "operator_imported")
		if err != nil {
			return nil, err
		}
		a.Repository = m.Repository
		a.SourceProfile = &m.SourceProfileID
		a.SourceDigest = &m.SourceProfileSHA256
		digest := packformat.Digest(manifest)
		a.Manifest = &digest
		payload, err := packformat.PayloadDigest(m.Files)
		if err != nil {
			return nil, err
		}
		a.Payload = &payload
		if encoded != nil && (!input.Verification || input.Result == "succeeded") {
			envelope, err := decodeSuccessfulEnvelope(encoded)
			if err != nil {
				return nil, err
			}
			p := provenanceFor("", m, envelope)
			a.Signers = p.VerifiedSignerKeyIDs
			a.ValidUntil = p.TrustValidUntil
			if proposal := envelope.TrustProposal; proposal != nil {
				root, err := packformat.RootVersion(proposal.Root)
				if err != nil {
					return nil, err
				}
				a.Versions.Root = &root
				targets, snapshot, timestamp := proposal.Metadata["targets"].Version, proposal.Metadata["snapshot"].Version, proposal.Metadata["timestamp"].Version
				a.Versions.Targets = &targets
				a.Versions.Snapshot = &snapshot
				a.Versions.Timestamp = &timestamp
			}
		}
	}
	preimage, err := canonicaljson.Marshal(a)
	if err != nil {
		return nil, err
	}
	var object map[string]any
	if err := json.Unmarshal(preimage, &object); err != nil {
		return nil, err
	}
	delete(object, "attestation_id")
	preimage, err = canonicaljson.Marshal(object)
	if err != nil {
		return nil, err
	}
	a.ID = "rpa_" + packformat.Digest(preimage)
	encoded, err = canonicaljson.Marshal(a)
	if err != nil {
		return nil, err
	}
	if err := packformat.ValidateAttestation(encoded); err != nil {
		return nil, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO reference_pack_events(attestation_id,operation_id,event_kind,pack_key,pack_version,canonical_attestation) VALUES($1,$2,$3,$4,$5,$6)`, a.ID, input.Operation, input.Kind, input.Key, input.Version, encoded)
	if err != nil {
		return nil, err
	}
	return encoded, appendPackAuditTx(ctx, tx, input.Operation, input.Kind, input.Result, input.At, input.Key, input.Version, nil)
}
