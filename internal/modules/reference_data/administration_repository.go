package reference_data

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packstate"
	"github.com/jackc/pgx/v5"
)

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

func (c *referenceDependencies) ListVersions(ctx context.Context) ([]AdministrativeVersion, error) {
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

func (c *referenceDependencies) GetVersion(ctx context.Context, key, version string) (AdministrativeVersion, error) {
	return scanAdministrativeVersion(c.pool.QueryRow(ctx, administrativeVersionSelect+` WHERE c.pack_key=$1 AND c.pack_version=$2`, key, version))
}
