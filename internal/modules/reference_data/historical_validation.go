package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var errHistoricalIntegrity = errors.New("reference pack: retained historical integrity check failed")

// ValidateRequiredState runs before serving traffic. It authenticates retained
// success at its recorded verification instant, including expired historical
// pins, without advancing trust or changing health or administrative intent.
func ValidateRequiredState(ctx context.Context, db postgres.DB, storage ArtifactStorage, limits Limits) error {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(limits.ReferencePacks.MaxVerificationSeconds)*time.Second)
	defer cancel()
	return validateHistoricalState(ctx, db, storage, limits, true, false)
}

// RestoreHistoricalState is called under Recovery's exclusive target lease,
// after authoritative rows and bytes have been restored. Indexes are derived
// afresh; signatures and freshness are checked at each envelope's successful
// instant, never at the restore clock. No success or provenance is invented.
func RestoreHistoricalState(ctx context.Context, db postgres.DB, storage ArtifactStorage, limits Limits) error {
	return validateHistoricalState(ctx, db, storage, limits, false, true)
}

func validateHistoricalState(ctx context.Context, db postgres.DB, storage ArtifactStorage, limits Limits, requiredOnly, rebuild bool) error {
	if db == nil || storage == nil {
		return errors.New("reference pack: historical validation dependencies required")
	}
	if err := validateRootHistory(ctx, db); err != nil {
		return err
	}
	if err := validateHistoricalAttestations(ctx, db); err != nil {
		return err
	}
	if err := validateHistoricalAttempts(ctx, db); err != nil {
		return err
	}
	if err := validateHistoricalDiagnostics(ctx, db); err != nil {
		return err
	}
	if err := validatePortablePreparations(ctx, db); err != nil {
		return err
	}
	if err := validatePortableCatalogs(ctx, db); err != nil {
		return err
	}
	// Keyset iteration retains one bounded envelope at a time. Restore checks
	// every successful envelope, including disabled and failed pinned versions.
	after := ""
	for {
		var id string
		var data []byte
		var current bool
		err := db.QueryRow(ctx, `SELECT e.envelope_id,e.canonical_envelope,c.current_envelope_id=e.envelope_id
 FROM reference_pack_envelopes e JOIN reference_pack_candidates c USING(pack_key,pack_version)
 WHERE e.envelope_id COLLATE "C">$1 AND (NOT $2 OR EXISTS (
   SELECT 1 FROM reference_pack_set_members m WHERE m.pack_key=e.pack_key AND m.pack_version=e.pack_version
   AND (m.provenance_envelope_id=e.envelope_id OR c.current_envelope_id=e.envelope_id)
   AND (EXISTS(SELECT 1 FROM reference_pack_current_set s WHERE s.pack_set_id=m.pack_set_id)
        OR EXISTS(SELECT 1 FROM reference_pack_pins p WHERE p.pack_set_id=m.pack_set_id)))
 OR EXISTS(SELECT 1 FROM reference_pack_version_pins p WHERE p.pack_key=e.pack_key AND p.pack_version=e.pack_version
 AND (p.envelope_id=e.envelope_id OR c.current_envelope_id=e.envelope_id)))
 ORDER BY e.envelope_id COLLATE "C" LIMIT 1`, after, requiredOnly).Scan(&id, &data, &current)
		if errors.Is(err, pgx.ErrNoRows) {
			break
		}
		if err != nil {
			return err
		}
		if err := validateHistoricalEnvelope(ctx, db, storage, limits, id, data, rebuild && current); err != nil {
			return err
		}
		after = id
	}
	// Foreign keys do not prove that every required version has a complete
	// derived index. The serving boundary fails closed on absent generations.
	var missing bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_set_members m
 JOIN reference_pack_candidates c USING(pack_key,pack_version)
 LEFT JOIN reference_pack_index_generations i ON i.index_id=c.current_index_id
 WHERE (EXISTS(SELECT 1 FROM reference_pack_current_set s WHERE s.pack_set_id=m.pack_set_id)
 OR EXISTS(SELECT 1 FROM reference_pack_pins p WHERE p.pack_set_id=m.pack_set_id))
 AND (c.current_envelope_id IS NULL OR i.index_id IS NULL OR NOT i.complete
 OR i.manifest_sha256<>m.manifest_sha256 OR i.payload_sha256<>m.payload_sha256))`).Scan(&missing); err != nil {
		return err
	}
	if missing {
		return errHistoricalIntegrity
	}
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_version_pins p
 JOIN reference_pack_candidates c USING(pack_key,pack_version)
 JOIN reference_pack_versions v USING(pack_key,pack_version)
 LEFT JOIN reference_pack_index_generations i ON i.index_id=c.current_index_id
 WHERE c.removed OR c.current_envelope_id IS NULL OR i.index_id IS NULL OR NOT i.complete
 OR i.manifest_sha256<>v.manifest_sha256 OR i.payload_sha256<>v.payload_sha256)`).Scan(&missing); err != nil {
		return err
	}
	if missing {
		return errHistoricalIntegrity
	}
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_current_set s
 JOIN reference_pack_set_members m USING(pack_set_id) JOIN reference_pack_candidates c USING(pack_key,pack_version)
 WHERE c.health<>'verified_available' OR c.administratively_disabled OR c.removed)`).Scan(&missing); err != nil {
		return err
	}
	if missing {
		return errHistoricalIntegrity
	}
	if err := CheckReadyState(ctx, db); err != nil {
		return err
	}
	return validateHistoricalSets(ctx, db, requiredOnly)
}

func validateHistoricalAttestations(ctx context.Context, db postgres.DB) error {
	rows, err := db.Query(ctx, `SELECT e.attestation_id,e.event_kind,e.pack_key,e.pack_version,e.canonical_attestation,o.operation_id,o.actor_kind,o.actor_user_id FROM reference_pack_events e JOIN reference_pack_operations o USING(operation_id) ORDER BY operation_id,attestation_id COLLATE "C"`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id, kind, key, version string
		var operation uuid.UUID
		var actorKind string
		var actor *uuid.UUID
		var data []byte
		if err := rows.Scan(&id, &kind, &key, &version, &data, &operation, &actorKind, &actor); err != nil {
			return err
		}
		if err := packformat.ValidateAttestation(data); err != nil {
			return errHistoricalIntegrity
		}
		var a packAttestation
		if err := json.Unmarshal(data, &a); err != nil {
			return err
		}
		if a.ID != id || a.Kind != kind || a.Key != key || a.Version != version || a.ActorKind != actorKind || !reflect.DeepEqual(a.Actor, actor) || (actorKind == "local_operator" && (a.Operator == nil || *a.Operator != operation)) {
			return errHistoricalIntegrity
		}
	}
	return rows.Err()
}

func validateHistoricalSets(ctx context.Context, db postgres.DB, requiredOnly bool) error {
	after := ""
	for {
		var id string
		var data, provenance []byte
		err := db.QueryRow(ctx, `SELECT pack_set_id,canonical_set,canonical_provenance FROM reference_pack_sets s
 WHERE pack_set_id COLLATE "C">$1 AND (NOT $2 OR EXISTS(SELECT 1 FROM reference_pack_current_set c WHERE c.pack_set_id=s.pack_set_id) OR EXISTS(SELECT 1 FROM reference_pack_pins p WHERE p.pack_set_id=s.pack_set_id))
 ORDER BY pack_set_id COLLATE "C" LIMIT 1`, after, requiredOnly).Scan(&id, &data, &provenance)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		set, err := decodeRetainedSet(data)
		if err != nil || set.ID != id {
			return errHistoricalIntegrity
		}
		if _, err := decodeRetainedProvenance(provenance, set); err != nil {
			return errHistoricalIntegrity
		}
		anchors := make([]PackProvenance, 0, len(set.Members))
		var memberCount int
		if err := db.QueryRow(ctx, `SELECT count(*) FROM reference_pack_set_members WHERE pack_set_id=$1`, id).Scan(&memberCount); err != nil {
			return err
		}
		if memberCount != len(set.Members) {
			return errHistoricalIntegrity
		}
		for _, member := range set.Members {
			var manifestBytes, envelopeBytes []byte
			if err := db.QueryRow(ctx, `SELECT v.manifest_bytes,e.canonical_envelope FROM reference_pack_set_members m
 JOIN reference_pack_versions v USING(pack_key,pack_version) JOIN reference_pack_envelopes e ON e.envelope_id=m.provenance_envelope_id
 WHERE m.pack_set_id=$1 AND m.pack_key=$2 AND m.pack_version=$3 AND m.manifest_sha256=$4 AND m.payload_sha256=$5 AND e.pack_key=m.pack_key AND e.pack_version=m.pack_version`, id, member.Key, member.Version, member.ManifestSHA256, member.PayloadSHA256).Scan(&manifestBytes, &envelopeBytes); err != nil {
				return errHistoricalIntegrity
			}
			envelope, err := decodeSuccessfulEnvelope(envelopeBytes)
			if err != nil {
				return errHistoricalIntegrity
			}
			manifest, err := packformat.DecodeManifest(manifestBytes, envelope.DistributionKind == "operator_imported")
			if err != nil {
				return errHistoricalIntegrity
			}
			anchors = append(anchors, provenanceFor(id, manifest, envelope))
		}
		expected, err := canonicaljson.Marshal(anchors)
		if err != nil || !bytes.Equal(expected, provenance) {
			return errHistoricalIntegrity
		}
		after = id
	}
}

func validateRootHistory(ctx context.Context, db postgres.DB) error {
	var missing bool
	if err := db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_repositories p WHERE NOT EXISTS(SELECT 1 FROM reference_pack_roots r WHERE r.repository_id=p.repository_id AND r.root_version=p.root_version))`).Scan(&missing); err != nil {
		return err
	}
	if missing {
		return errHistoricalIntegrity
	}
	rows, err := db.Query(ctx, `SELECT r.repository_id,r.root_version,r.canonical_bytes,r.sha256,r.predecessor_version,r.transition_evidence,p.root_version
 FROM reference_pack_roots r JOIN reference_pack_repositories p USING(repository_id)
 ORDER BY r.repository_id COLLATE "C",r.root_version`)
	if err != nil {
		return err
	}
	defer rows.Close()
	var repository string
	var prior []byte
	var version, current int64
	for rows.Next() {
		var id, digest string
		var next, expectedCurrent int64
		var data, evidence []byte
		var predecessor *int64
		if err := rows.Scan(&id, &next, &data, &digest, &predecessor, &evidence, &expectedCurrent); err != nil {
			return err
		}
		if repository != id {
			if repository != "" && version != current {
				return errHistoricalIntegrity
			}
			repository, prior, version = id, nil, 0
		}
		if digest != packformat.Digest(data) || (prior == nil && predecessor != nil) || (prior != nil && (predecessor == nil || *predecessor != version)) {
			return errHistoricalIntegrity
		}
		transition, err := packformat.VerifyHistoricalRoot(prior, data, repository)
		if err != nil || transition.Version != next {
			return errHistoricalIntegrity
		}
		if prior == nil {
			if string(evidence) != "{}" {
				return errHistoricalIntegrity
			}
		} else {
			expected, err := canonicaljson.Marshal(transition)
			// transition_evidence is JSONB: PostgreSQL renders equivalent JSON
			// with its own whitespace and key order. Authenticate its value,
			// while keeping exact-byte checks for signed root material above.
			retained, retainedErr := canonicaljson.Canonicalize(evidence)
			if err != nil || retainedErr != nil || !bytes.Equal(retained, expected) {
				return errHistoricalIntegrity
			}
		}
		prior, version, current = data, next, expectedCurrent
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if repository != "" && version != current {
		return errHistoricalIntegrity
	}
	return nil
}

type discardContentRows struct{}

func (discardContentRows) Append(ctx context.Context, _ packformat.ContentRow) error {
	return ctx.Err()
}

func validateHistoricalEnvelope(ctx context.Context, db postgres.DB, storage ArtifactStorage, limits Limits, id string, data []byte, rebuild bool) error {
	envelope, err := decodeSuccessfulEnvelope(data)
	if err != nil {
		return errHistoricalIntegrity
	}
	if id != "rpenv_"+packformat.Digest(data) {
		return errHistoricalIntegrity
	}
	var repository *string
	var rootVersion *int64
	if envelope.TrustProposal != nil {
		id := envelope.TrustProposal.Binding["trust_repository_id"].(string)
		repository = &id
		version, err := packformat.RootVersion(envelope.TrustProposal.Root)
		if err != nil {
			return errHistoricalIntegrity
		}
		rootVersion = &version
	}
	var matches bool
	// PostgreSQL timestamps project the envelope instant at microsecond
	// precision. The canonical envelope remains the exact historical value.
	if err := db.QueryRow(ctx, `SELECT operation_id=$2 AND pack_key=$3 AND pack_version=$4 AND verified_at=$5
 AND trust_valid_until IS NOT DISTINCT FROM $6::timestamptz AND container_sha256 IS NOT DISTINCT FROM $7::text
 AND container_ref IS NOT DISTINCT FROM $8::text AND repository_id IS NOT DISTINCT FROM $9::text
 AND root_version IS NOT DISTINCT FROM $10::bigint
 AND (container_ref IS NULL OR EXISTS(SELECT 1 FROM reference_pack_objects o JOIN reference_pack_object_refs r USING(object_id)
 WHERE o.storage_ref=reference_pack_envelopes.container_ref AND o.sha256=reference_pack_envelopes.container_sha256 AND o.available
 AND r.owner_kind='envelope' AND r.owner_id=reference_pack_envelopes.envelope_id AND r.logical_path='container'))
 FROM reference_pack_envelopes WHERE envelope_id=$1`,
		id, envelope.OperationID, envelope.PackKey, envelope.PackVersion, envelope.VerifiedAt, envelope.TrustValidUntil,
		envelope.ContainerSHA256, envelope.ContainerRef, repository, rootVersion).Scan(&matches); err != nil || !matches {
		return errHistoricalIntegrity
	}
	var manifestBytes []byte
	var manifestSHA, payloadSHA string
	if err := db.QueryRow(ctx, `SELECT manifest_bytes,manifest_sha256,payload_sha256 FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2`, envelope.PackKey, envelope.PackVersion).Scan(&manifestBytes, &manifestSHA, &payloadSHA); err != nil {
		return err
	}
	if manifestSHA != envelope.ManifestSHA256 || payloadSHA != envelope.PayloadSHA256 || packformat.Digest(manifestBytes) != manifestSHA {
		return errHistoricalIntegrity
	}
	manifest, err := packformat.DecodeManifest(manifestBytes, envelope.DistributionKind == "operator_imported")
	if err != nil {
		return errHistoricalIntegrity
	}
	var removed bool
	if err := db.QueryRow(ctx, `SELECT removed FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, envelope.PackKey, envelope.PackVersion).Scan(&removed); err != nil {
		return err
	}
	if removed && envelope.DistributionKind != "operator_imported" {
		return errHistoricalIntegrity
	}
	if !removed {
		if err := checkRetainedVersion(ctx, db, storage, envelope.PackKey, envelope.PackVersion); err != nil {
			return err
		}
	} else {
		rebuild = false
	}
	var tx pgx.Tx
	var builder *indexBuilder
	if rebuild {
		tx, err = db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		builder, err = newIndexBuilderTx(ctx, tx, envelope.OperationID, manifest, manifestSHA, payloadSHA)
		if err != nil {
			return err
		}
	}
	var sink packformat.ContentSink = discardContentRows{}
	if builder != nil {
		sink = builder
	}
	switch envelope.DistributionKind {
	case "packaged_builtin":
		if envelope.TrustSnapshot != nil || envelope.TrustProposal != nil || envelope.TrustValidUntil != nil || envelope.ContainerRef != nil || envelope.ContainerSHA256 != nil {
			return errHistoricalIntegrity
		}
		var binding []byte
		if err := db.QueryRow(ctx, `SELECT canonical_binding FROM reference_pack_release_bindings WHERE pack_key=$1 AND pack_version=$2 ORDER BY application_release_id COLLATE "C" LIMIT 1`, envelope.PackKey, envelope.PackVersion).Scan(&binding); err != nil {
			return err
		}
		releaseBinding, err := packformat.DecodeBuiltinReleaseBinding(binding)
		if err != nil || !releaseBinding.Matches(manifest, manifestSHA, payloadSHA) {
			return errHistoricalIntegrity
		}
		if err := packformat.ValidateContent(ctx, manifest, retainedContentSource{db, storage, envelope.PackKey, envelope.PackVersion}, sink); err != nil {
			return err
		}
	case "operator_imported":
		if manifest.Repository == nil || envelope.TrustSnapshot == nil || envelope.TrustProposal == nil || envelope.ContainerRef == nil || envelope.ContainerSHA256 == nil || envelope.TrustValidUntil == nil || !envelope.VerifiedAt.Before(*envelope.TrustValidUntil) {
			return errHistoricalIntegrity
		}
		for version, root := range envelope.TrustSnapshot.RootHistory {
			var retained []byte
			if err := db.QueryRow(ctx, `SELECT canonical_bytes FROM reference_pack_roots WHERE repository_id=$1 AND root_version=$2`, *manifest.Repository, version).Scan(&retained); err != nil || !bytes.Equal(retained, root) {
				return errHistoricalIntegrity
			}
		}
		ref, err := ParseStorageRef(*envelope.ContainerRef)
		if err != nil {
			return errHistoricalIntegrity
		}
		verified, err := verifyCanonicalContainer(ctx, storage, verificationAttempt{Retained: &ref, ContainerSHA256: *envelope.ContainerSHA256, Start: envelope.VerifiedAt, ClockTrusted: true, Limits: limits.verificationArchiveLimits(), Identity: retainedVerificationIdentity{db: db, storage: storage, manifest: manifest, manifestSHA: manifestSHA, payloadSHA: payloadSHA}, Repositories: map[string]packformat.TrustSnapshot{*manifest.Repository: *envelope.TrustSnapshot}}, func(context.Context, packformat.Manifest, string, string) (packformat.ContentSink, error) {
			return sink, nil
		})
		if err != nil {
			return err
		}
		if err := verified.Close(); err != nil {
			return err
		}
		actualProposal, actualErr := canonicaljson.Marshal(verified.Trust)
		retainedProposal, retainedErr := canonicaljson.Marshal(envelope.TrustProposal)
		if verified.ManifestSHA256 != manifestSHA || verified.PayloadSHA256 != payloadSHA || actualErr != nil || retainedErr != nil || !bytes.Equal(actualProposal, retainedProposal) || !verified.Trust.ValidUntil.Equal(*envelope.TrustValidUntil) {
			return errHistoricalIntegrity
		}
	default:
		return errHistoricalIntegrity
	}
	if builder != nil {
		index, err := builder.Complete(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET current_index_id=$3 WHERE pack_key=$1 AND pack_version=$2 AND current_envelope_id=$4`, envelope.PackKey, envelope.PackVersion, index, id); err != nil {
			return err
		}
		return tx.Commit(ctx)
	}
	return nil
}

type retainedContentSource struct {
	db           postgres.DB
	storage      VerificationStorage
	key, version string
}

func (s retainedContentSource) Open(ctx context.Context, path string) (io.ReadCloser, error) {
	var raw string
	if err := s.db.QueryRow(ctx, `SELECT o.storage_ref FROM reference_pack_object_refs r JOIN reference_pack_objects o USING(object_id) WHERE r.owner_kind='version' AND r.owner_id=$1 AND r.logical_path=$2 AND o.available`, versionObjectID(s.key, s.version), path).Scan(&raw); err != nil {
		return nil, err
	}
	ref, err := ParseStorageRef(raw)
	if err != nil {
		return nil, err
	}
	reader, size, err := s.storage.OpenPublished(ctx, ref)
	if err != nil {
		return nil, err
	}
	return &retainedMemberReader{Reader: io.NewSectionReader(reader, 0, size), Closer: reader}, nil
}

type retainedMemberReader struct {
	io.Reader
	io.Closer
}
