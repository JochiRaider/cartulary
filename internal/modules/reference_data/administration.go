package reference_data

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packstate"
	"github.com/JochiRaider/cartulary/internal/platform/listquery"
	"github.com/jackc/pgx/v5"
)

// AdministrativeVersion projects current administrative state. Successful
// metadata is nullable until a successful envelope exists; failed attempts
// never manufacture digests, signatures or a verification timestamp.
type AdministrativeVersion struct {
	PackKey                  string                     `json:"pack_key"`
	PackKind                 string                     `json:"pack_kind"`
	PackVersion              string                     `json:"pack_version"`
	Condition                string                     `json:"pack_version_state"`
	Health                   string                     `json:"health"`
	AdministrativelyDisabled bool                       `json:"administratively_disabled"`
	DistributionKind         string                     `json:"distribution_kind"`
	Active                   bool                       `json:"active"`
	PendingWork              bool                       `json:"pending_work"`
	ReproducibilityPinned    bool                       `json:"reproducibility_pinned"`
	FallbackFromVersion      *string                    `json:"fallback_from_version"`
	Dependencies             []AdministrativeDependency `json:"dependencies"`
	Removed                  bool                       `json:"removed"`
	MissingReason            *string                    `json:"missing_reason"`
	LastFailureCode          *string                    `json:"last_failure_code"`
	SourceIdentifier         *string                    `json:"source_identifier"`
	ManifestSHA256           *string                    `json:"manifest_sha256"`
	PayloadSHA256            *string                    `json:"payload_sha256"`
	PackContractVersion      *string                    `json:"pack_contract_version"`
	PackReleaseSequence      *int64                     `json:"pack_release_sequence"`
	ContentProfileID         *string                    `json:"content_profile_id"`
	ContentProfileVersion    *string                    `json:"content_profile_version"`
	SourceProfileID          *string                    `json:"source_profile_id"`
	SourceProfileSHA256      *string                    `json:"source_profile_sha256"`
	SourceVersion            *string                    `json:"source_version"`
	SourceAsOf               *string                    `json:"source_as_of"`
	LicenseExpression        *string                    `json:"license_expression"`
	Redistribution           *string                    `json:"redistribution"`
	TrustRepositoryID        *string                    `json:"trust_repository_id"`
	VerificationMethod       *string                    `json:"verification_method"`
	LastVerifiedAt           *time.Time                 `json:"last_verified_at"`
	TrustValidUntil          *time.Time                 `json:"trust_valid_until"`
	VerifiedSignerKeyIDs     []string                   `json:"verified_signer_key_ids"`
	ImportedByUserID         *string                    `json:"imported_by_user_id"`
	ImportedAt               time.Time                  `json:"imported_at"`
	PreviousActiveVersion    *string                    `json:"previous_active_version"`
	ActivatedByUserID        *string                    `json:"activated_by_user_id"`
	ActivatedAt              *time.Time                 `json:"activated_at"`
}

type AdministrativeDependency struct {
	PackKey       string `json:"pack_key"`
	PackVersion   string `json:"pack_version"`
	PayloadSHA256 string `json:"payload_sha256"`
}

const administrativeVersionSelect = `SELECT c.pack_key,c.pack_version,c.health,c.administratively_disabled,
c.distribution_kind,c.removed,c.missing_reason,c.last_failure_code,c.admitted_by_user_id::text,c.admitted_at,
v.manifest_sha256,v.payload_sha256,v.manifest_bytes,e.canonical_envelope,a.previous_version,a.actor_user_id,a.admitted_at,
EXISTS(SELECT 1 FROM reference_pack_current_set s JOIN reference_pack_set_members m USING(pack_set_id)
 WHERE s.singleton AND m.pack_key=c.pack_key AND m.pack_version=c.pack_version),
EXISTS(SELECT 1 FROM reference_pack_operation_keys k JOIN reference_pack_operations o USING(operation_id) WHERE k.pack_key=c.pack_key AND o.terminal_at IS NULL),
(EXISTS(SELECT 1 FROM reference_pack_pins p JOIN reference_pack_set_members m USING(pack_set_id) WHERE m.pack_key=c.pack_key AND m.pack_version=c.pack_version)
 OR EXISTS(SELECT 1 FROM reference_pack_version_pins p WHERE p.pack_key=c.pack_key AND p.pack_version=c.pack_version)),c.fallback_from_version
FROM reference_pack_candidates c LEFT JOIN reference_pack_versions v USING(pack_key,pack_version)
LEFT JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id
LEFT JOIN LATERAL(SELECT convert_from(event.canonical_attestation,'UTF8')::jsonb->>'prior_active_version' previous_version,o.actor_user_id::text,o.admitted_at
FROM reference_pack_events event JOIN reference_pack_operations o USING(operation_id)
WHERE event.pack_key=c.pack_key AND event.pack_version=c.pack_version AND event.event_kind IN ('activation','rollback_activation')
ORDER BY o.admitted_at DESC,o.operation_id DESC LIMIT 1) a ON true`

func scanAdministrativeVersion(row pgx.Row) (AdministrativeVersion, error) {
	v := AdministrativeVersion{VerifiedSignerKeyIDs: []string{}}
	var manifestBytes, envelopeBytes []byte
	err := row.Scan(&v.PackKey, &v.PackVersion, &v.Health, &v.AdministrativelyDisabled, &v.DistributionKind, &v.Removed, &v.MissingReason, &v.LastFailureCode, &v.ImportedByUserID, &v.ImportedAt, &v.ManifestSHA256, &v.PayloadSHA256, &manifestBytes, &envelopeBytes, &v.PreviousActiveVersion, &v.ActivatedByUserID, &v.ActivatedAt, &v.Active, &v.PendingWork, &v.ReproducibilityPinned, &v.FallbackFromVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, ErrNotFound
	}
	if err != nil {
		return v, err
	}
	if !v.Active {
		v.FallbackFromVersion = nil
	}
	v.Condition = packstate.Condition(packstate.Version{Health: packstate.Health(v.Health), Disabled: v.AdministrativelyDisabled})
	for _, p := range packformat.Profiles() {
		if p.Key == v.PackKey {
			v.PackKind = p.Kind
			break
		}
	}
	if len(envelopeBytes) == 0 {
		return v, nil
	}
	envelope, err := decodeSuccessfulEnvelope(envelopeBytes)
	if err != nil {
		return v, err
	}
	m, err := packformat.DecodeManifest(manifestBytes, v.DistributionKind == "operator_imported")
	if err != nil {
		return v, err
	}
	p := provenanceFor("", m, envelope)
	v.Dependencies = make([]AdministrativeDependency, 0, len(m.Dependencies))
	for _, dependency := range m.Dependencies {
		v.Dependencies = append(v.Dependencies, AdministrativeDependency{PackKey: dependency.Key, PackVersion: dependency.Version, PayloadSHA256: dependency.SHA256})
	}
	v.SourceIdentifier = &p.SourceIdentifier
	v.PackContractVersion = &p.PackContractVersion
	v.PackReleaseSequence = &m.Sequence
	v.ContentProfileID = &p.ContentProfileID
	v.ContentProfileVersion = &p.ContentProfileVersion
	v.SourceProfileID = &m.SourceProfileID
	v.SourceProfileSHA256 = &m.SourceProfileSHA256
	v.SourceVersion = &m.SourceVersion
	v.SourceAsOf = m.SourceAsOf
	v.LicenseExpression = &m.License.Expression
	v.Redistribution = &m.License.Redistribution
	v.TrustRepositoryID = m.Repository
	v.VerificationMethod = &p.VerificationMethod
	v.LastVerifiedAt = &p.LastVerifiedAt
	v.TrustValidUntil = p.TrustValidUntil
	v.VerifiedSignerKeyIDs = append([]string{}, p.VerifiedSignerKeyIDs...)
	return v, nil
}

func (c *Coordinator) ListVersions(ctx context.Context) ([]AdministrativeVersion, error) {
	rows, err := c.pool.Query(ctx, administrativeVersionSelect+` ORDER BY c.pack_key COLLATE "C",c.pack_version COLLATE "C"`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	versions := []AdministrativeVersion{}
	for rows.Next() {
		v, err := scanAdministrativeVersion(rows)
		if err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (c *Coordinator) GetVersion(ctx context.Context, key, version string) (AdministrativeVersion, error) {
	return scanAdministrativeVersion(c.pool.QueryRow(ctx, administrativeVersionSelect+` WHERE c.pack_key=$1 AND c.pack_version=$2`, key, version))
}

func (v AdministrativeVersion) Resource() map[string]any {
	// Paging consumes JSON-shaped values, including explicit nulls. This is a
	// projection of the closed typed resource, not a second schema definition.
	raw, _ := json.Marshal(v)
	var resource map[string]any
	_ = json.Unmarshal(raw, &resource)
	return resource
}

func filterAdministrativeVersions(versions []AdministrativeVersion, scope map[string]string) []AdministrativeVersion {
	result := versions[:0]
	for _, v := range versions {
		if state := scope["pack_version_state"]; state != "" && state != v.Condition {
			continue
		}
		if active := scope["active"]; active != "" && v.Active != (active == "true") {
			continue
		}
		if !listquery.MatchSearchTokens(strings.Fields(scope["search"]), v.PackKey, v.PackKind, v.PackVersion, searchableOptionalString(v.SourceIdentifier), searchableOptionalString(v.ManifestSHA256), searchableOptionalString(v.PayloadSHA256)) {
			continue
		}
		result = append(result, v)
	}
	return result
}
