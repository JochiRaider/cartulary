package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packstate"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type builtinRelease struct {
	SchemaID            string        `json:"schema_id"`
	ReleaseID           string        `json:"application_release_id"`
	SourceProfileSHA256 string        `json:"source_profile_sha256"`
	Packs               []builtinPack `json:"packs"`
}
type builtinPack struct {
	Binding packformat.BuiltinReleaseBinding `json:"binding"`
	Members map[string]string                `json:"members"`
}
type builtinSource map[string]string

func (s builtinSource) Open(_ context.Context, path string) (io.ReadCloser, error) {
	value, ok := s[path]
	if !ok {
		return nil, errors.New("missing built-in release member")
	}
	return io.NopCloser(strings.NewReader(value)), nil
}

type builtinRows []packformat.ContentRow

func (s *builtinRows) Append(_ context.Context, row packformat.ContentRow) error {
	*s = append(*s, row)
	return nil
}

func releaseArtifact(name string) ([]byte, error) {
	for _, artifact := range contractreferencepacks.Artifacts {
		if artifact.Path == "contracts/reference-packs/builtins/"+name {
			return []byte(artifact.JSON), nil
		}
	}
	return nil, errors.New("reference pack: application release binding absent")
}
func loadBuiltinRelease(ctx context.Context) (builtinRelease, []preparedVersion, error) {
	data, err := releaseArtifact("release.v1.json")
	if err != nil {
		return builtinRelease{}, nil, err
	}
	var release builtinRelease
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return release, nil, err
	}
	object, ok := value.(map[string]any)
	if !ok || len(object) != 4 || object["schema_id"] == nil || object["application_release_id"] == nil || object["source_profile_sha256"] == nil || object["packs"] == nil {
		return release, nil, errors.New("reference pack: invalid Base package shape")
	}
	packages, ok := object["packs"].([]any)
	if !ok {
		return release, nil, errors.New("reference pack: invalid Base package list")
	}
	for _, value := range packages {
		p, ok := value.(map[string]any)
		if !ok || len(p) != 2 || p["binding"] == nil || p["members"] == nil {
			return release, nil, errors.New("reference pack: invalid Base package member")
		}
		binding, err := canonicaljson.Marshal(p["binding"])
		if err != nil {
			return release, nil, err
		}
		if _, err := packformat.DecodeBuiltinReleaseBinding(binding); err != nil {
			return release, nil, err
		}
	}
	if err := json.Unmarshal(data, &release); err != nil {
		return release, nil, err
	}
	if release.SchemaID != "cartulary.reference_pack_builtin_release.v1" || release.ReleaseID == "" || len(release.Packs) != 3 {
		return release, nil, errors.New("reference pack: incomplete Base release")
	}
	source, err := releaseArtifact("source-profile.v1.json")
	if err != nil || packformat.Digest(source) != release.SourceProfileSHA256 {
		return release, nil, errors.New("reference pack: source-profile binding mismatch")
	}
	prepared := []preparedVersion{}
	seen := map[string]bool{}
	previous := ""
	for _, pack := range release.Packs {
		bindingBytes, err := canonicaljson.Marshal(pack.Binding)
		if err != nil {
			return release, nil, err
		}
		binding, err := packformat.DecodeBuiltinReleaseBinding(bindingBytes)
		if err != nil {
			return release, nil, err
		}
		if seen[binding.Key] || binding.Key <= previous {
			return release, nil, errors.New("reference pack: unexpected Base registry")
		}
		seen[binding.Key] = true
		previous = binding.Key
		manifestBytes := []byte(pack.Members["manifest.json"])
		manifest, err := packformat.DecodeManifest(manifestBytes, false)
		if err != nil {
			return release, nil, err
		}
		payloadSHA, err := packformat.PayloadDigest(manifest.Files)
		if err != nil || !binding.Matches(manifest, packformat.Digest(manifestBytes), payloadSHA) || manifest.SourceProfileSHA256 != release.SourceProfileSHA256 {
			return release, nil, errors.New("reference pack: Base release digest mismatch")
		}
		inventory := packformat.Inventory{}
		for path, data := range pack.Members {
			inventory[path] = packformat.Member{Path: path, Size: int64(len(data)), SHA256: packformat.Digest([]byte(data))}
		}
		if err := packformat.ValidateInventory(manifest, inventory, false); err != nil {
			return release, nil, err
		}
		for _, file := range manifest.Files {
			if file.Role == "notice" {
				if err := packformat.ValidateNotice([]byte(pack.Members[file.Path])); err != nil {
					return release, nil, err
				}
			}
		}
		var rows builtinRows
		if err := packformat.ValidateContent(ctx, manifest, builtinSource(pack.Members), &rows); err != nil {
			return release, nil, err
		}
		if len(rows) == 0 {
			return release, nil, errors.New("reference pack: empty Base registry")
		}
		prepared = append(prepared, preparedVersion{Content: &VerifiedContent{Manifest: manifest, ManifestBytes: manifestBytes, ManifestSHA256: binding.ManifestSHA256, PayloadSHA256: binding.PayloadSHA256, Inventory: inventory}, Rows: rows})
	}
	return release, prepared, nil
}

// ReconcileBaseRelease runs before readiness in every deployment profile. Its
// input is release-bound canonical content, never runtime-generated empty packs.
type BaseReleaseOptions struct {
	Observer       OperationObserver
	ProfileClaimed bool
	ClockTrusted   bool
	Limits         Limits
}

func ReconcileBaseRelease(ctx context.Context, pool *pgxpool.Pool, storage ArtifactStorage, options BaseReleaseOptions, at time.Time) (resultErr error) {
	ctx, end := observeReferenceOperation(ctx, options.Observer, "reference_pack.reconcile")
	defer func() { end(referenceOutcome(resultErr)) }()
	if pool == nil || storage == nil || at.IsZero() {
		return errors.New("reference pack: incomplete Base startup dependencies")
	}
	if err := validateCoordinatorLimits(options.Limits); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(options.Limits.ReferencePacks.MaxVerificationSeconds)*time.Second)
	defer cancel()
	release, prepared, err := loadBuiltinRelease(ctx)
	if err != nil {
		return err
	}
	operationID := uuid.New()
	claimed := options.ProfileClaimed
	archive := options.Limits.verificationArchiveLimits()
	if archive.ContainerBytes < 1 || archive.ExtractedBytes < 1 || archive.CompressionRatio < 1 || archive.CompressionRatio > 1000 || archive.Members < 1 || options.Limits.ReferencePacks.MaxVerificationSeconds < 60 || options.Limits.ReferencePacks.MaxVerificationSeconds > 86400 {
		return errors.New("reference pack: invalid runtime policy")
	}
	capture, err := inspectReconciliation(ctx, pool, storage)
	if err != nil {
		return err
	}
	frozen, _ := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_reconciliation.v1", "application_release_id": release.ReleaseID, "profile_claimed": claimed, "clock_trusted": options.ClockTrusted, "max_verification_seconds": options.Limits.ReferencePacks.MaxVerificationSeconds, "max_container_bytes": archive.ContainerBytes, "max_extracted_bytes": archive.ExtractedBytes, "max_compression_ratio": archive.CompressionRatio, "max_members": archive.Members})
	if _, err := pool.Exec(ctx, `INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,admitted_at,frozen_input) VALUES($1,'reconcile','application_release',$2,$3)`, operationID, at, frozen); err != nil {
		return err
	}
	created := []StorageRef{}
	leases := []io.Closer{}
	defer func() {
		for _, lease := range leases {
			_ = lease.Close()
		}
	}()
	commitStarted := false
	defer func() {
		if resultErr != nil && !commitStarted {
			for _, ref := range created {
				resultErr = errors.Join(resultErr, storage.RemovePublished(ref))
			}
		}
	}()
	for i := range prepared {
		p := &prepared[i]
		p.Envelope = successfulEnvelope{SchemaID: "cartulary.reference_pack_successful_envelope.v1", OperationID: operationID, PackKey: p.Content.Manifest.Key, PackVersion: p.Content.Manifest.Version, DistributionKind: "packaged_builtin", ManifestSHA256: p.Content.ManifestSHA256, PayloadSHA256: p.Content.PayloadSHA256, VerifiedAt: at.UTC()}
		var exists bool
		if err := pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2)`, p.Content.Manifest.Key, p.Content.Manifest.Version).Scan(&exists); err != nil {
			return err
		}
		if exists {
			repository := &canonicalRepository{pool: pool, storage: storage}
			if err := repository.requireAvailable(ctx, "", provenanceFor("", p.Content.Manifest, p.Envelope)); err != nil {
				return err
			}
			continue
		}
		builder, err := newIndexBuilder(ctx, pool, operationID, p.Content.Manifest, p.Content.ManifestSHA256, p.Content.PayloadSHA256)
		if err != nil {
			return err
		}
		for _, row := range p.Rows {
			if err := builder.Append(ctx, row); err != nil {
				return err
			}
		}
		p.IndexID, err = builder.Complete(ctx)
		if err != nil {
			return err
		}
		pack := release.Packs[i]
		paths := make([]string, 0, len(pack.Members))
		for path := range pack.Members {
			paths = append(paths, path)
		}
		slices.Sort(paths)
		for _, path := range paths {
			data := pack.Members[path]
			digest := packformat.Digest([]byte(data))
			ref, lease, err := storage.PublishStream(ctx, digest, int64(len(data)), strings.NewReader(data))
			if err != nil {
				return err
			}
			leases = append(leases, lease)
			created = append(created, ref)
			p.Objects = append(p.Objects, preparedObject{ID: uuid.New(), Path: path, Digest: digest, Size: int64(len(data)), Reference: ref})
		}
	}
	// Objects survive an uncertain commit. Unreferenced objects are reclaimed
	// only by owner collection after checking all retained reference families.
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := lockByteRetentionTx(ctx, tx, false); err != nil {
		return err
	}
	for _, profile := range packformat.Profiles() {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_key_state(pack_key) VALUES($1) ON CONFLICT DO NOTHING`, profile.Key); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `SELECT 1 FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, profile.Key); err != nil {
			return err
		}
	}
	var currentID *string
	// Establish the singleton in the same transaction as the first usable Base
	// set. A freshly migrated or reset recovery target has no retained pack state.
	inserted, err := tx.Exec(ctx, `INSERT INTO reference_pack_current_set(singleton,revision,configuration_sha256) VALUES(true,1,repeat('0',64)) ON CONFLICT DO NOTHING`)
	if err != nil {
		return err
	}
	var currentRevision int64
	if err := tx.QueryRow(ctx, `SELECT pack_set_id,revision FROM reference_pack_current_set WHERE singleton FOR UPDATE`).Scan(&currentID, &currentRevision); err != nil {
		return err
	}
	if err := capture.guard(ctx, tx, inserted.RowsAffected() == 1, currentRevision); err != nil {
		return err
	}
	for _, key := range mandatoryRegistryKeys {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_registry_usage(pack_key,revision) VALUES($1,1) ON CONFLICT DO NOTHING`, key); err != nil {
			return err
		}
	}
	base := []packstate.Version{}
	for i, p := range prepared {
		var previous *string
		err := tx.QueryRow(ctx, `SELECT current_envelope_id FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, p.Content.Manifest.Key, p.Content.Manifest.Version).Scan(&previous)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if previous == nil {
			id, err := publishVersionTx(ctx, tx, p, operationID, nil)
			if err != nil {
				return err
			}
			previous = &id
		} else {
			var manifest, payload string
			if err := tx.QueryRow(ctx, `SELECT manifest_sha256,payload_sha256 FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2`, p.Content.Manifest.Key, p.Content.Manifest.Version).Scan(&manifest, &payload); err != nil {
				return err
			}
			if manifest != p.Content.ManifestSHA256 || payload != p.Content.PayloadSHA256 {
				return errors.New("reference pack: retained Base binding collision")
			}
		}
		m := p.Content.Manifest
		base = append(base, packstate.Version{Member: memberFor(p.Content), Manifest: m, Health: packstate.Available, Builtin: true, Envelope: &packstate.Envelope{ID: *previous, VerifiedAt: at}})
		binding, err := canonicaljson.Marshal(release.Packs[i].Binding)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_release_bindings(application_release_id,pack_key,pack_version,canonical_binding) VALUES($1,$2,$3,$4) ON CONFLICT DO NOTHING`, release.ReleaseID, m.Key, m.Version, binding); err != nil {
			return err
		}
		var retained []byte
		if err := tx.QueryRow(ctx, `SELECT canonical_binding FROM reference_pack_release_bindings WHERE application_release_id=$1 AND pack_key=$2 AND pack_version=$3`, release.ReleaseID, m.Key, m.Version).Scan(&retained); err != nil {
			return err
		}
		if !bytes.Equal(binding, retained) {
			return errors.New("reference pack: retained release binding collision")
		}
	}
	selected := []packstate.Version{}
	previous := make([]packstate.Version, 0, len(capture.versions))
	for _, inspected := range capture.versions {
		version := inspected.state
		previous = append(previous, version)
		if inspected.code != "" {
			if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET health=$3,last_failure_code=$4,missing_reason=$5 WHERE pack_key=$1 AND pack_version=$2`, version.Member.Key, version.Member.Version, version.Health, inspected.code, version.MissingReason); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, version.Member.Key); err != nil {
				return err
			}
		}
		// Current-release built-ins replace old built-ins. Imported healthy
		// required registries remain effective when the profile is claimed.
		if claimed && !version.Builtin {
			selected = append(selected, version)
		}
	}
	effective, _, err := packstate.ResolveEffectiveSet(selected, base)
	if err != nil {
		return err
	}
	members := make([]PackSetMember, 0, len(effective))
	for _, version := range effective {
		members = append(members, version.Member)
	}
	set, err := publishSetTx(ctx, tx, members, operationID)
	if err != nil {
		return err
	}
	if !claimed {
		if _, err := tx.Exec(ctx, `UPDATE reference_pack_candidates SET fallback_from_version=NULL WHERE fallback_from_version IS NOT NULL`); err != nil {
			return err
		}
	}
	configurationSHA := packformat.Digest(frozen)
	previousSet := ""
	if currentID != nil {
		previousSet = *currentID
	}
	if err := publishSelectionConsequencesTx(ctx, tx, operationID, previousSet, set.ID, previous, effective, at, claimed); err != nil {
		return err
	}
	for _, inspected := range capture.versions {
		if inspected.code != "" {
			member := inspected.state.Member
			if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: operationID, Key: member.Key, Version: member.Version, Kind: "payload_invalidation", Result: "failed", At: at, PreviousSet: previousSet, ResultingSet: &set.ID}); err != nil {
				return err
			}
		}
		if !inspected.state.Builtin && !slices.ContainsFunc(effective, func(v packstate.Version) bool { return v.Member == inspected.state.Member }) {
			member := inspected.state.Member
			if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: operationID, Key: member.Key, Version: member.Version, Kind: "profile_reconciliation", Result: "succeeded", At: at, PreviousSet: previousSet, ResultingSet: &set.ID}); err != nil {
				return err
			}
		}
	}
	for _, p := range prepared {
		if _, err := appendPackAttestationTx(ctx, tx, attestationInput{Operation: operationID, Key: p.Content.Manifest.Key, Version: p.Content.Manifest.Version, Kind: "profile_reconciliation", Result: "succeeded", At: at, PreviousSet: previousSet, ResultingSet: &set.ID}); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_current_set SET application_release_id=$1,profile_claimed=$2,configuration_sha256=$3,revision=revision+CASE WHEN configuration_sha256<>$3 THEN 1 ELSE 0 END WHERE singleton`, release.ReleaseID, claimed, configurationSHA); err != nil {
		return err
	}
	outcome, _ := canonicaljson.Marshal(map[string]any{"pack_set_id": set.ID})
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_operations SET terminal_at=$2,final_outcome=$3 WHERE operation_id=$1`, operationID, at, outcome); err != nil {
		return err
	}
	commitStarted = true
	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func memberFor(content *VerifiedContent) PackSetMember {
	m := content.Manifest
	return PackSetMember{Key: m.Key, Version: m.Version, ManifestSHA256: content.ManifestSHA256, PayloadSHA256: content.PayloadSHA256, Contract: m.Contract, ProfileID: m.ProfileID, ProfileVersion: m.ProfileVersion}
}

func loadStateVersionTx(ctx context.Context, tx pgx.Tx, key, version string) (packstate.Version, error) {
	var state packstate.Version
	var manifest, encoded []byte
	var distribution string
	var envelopeID, missingReason *string
	err := tx.QueryRow(ctx, `SELECT c.health,c.administratively_disabled,c.removed,c.missing_reason,c.distribution_kind,c.current_envelope_id,v.manifest_bytes,v.manifest_sha256,v.payload_sha256,e.canonical_envelope FROM reference_pack_candidates c JOIN reference_pack_versions v USING(pack_key,pack_version) LEFT JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id WHERE c.pack_key=$1 AND c.pack_version=$2`, key, version).Scan(&state.Health, &state.Disabled, &state.Removed, &missingReason, &distribution, &envelopeID, &manifest, &state.Member.ManifestSHA256, &state.Member.PayloadSHA256, &encoded)
	if err != nil {
		return state, err
	}
	state.Builtin = distribution == "packaged_builtin"
	state.MissingReason = missingReason
	state.Manifest, err = packformat.DecodeManifest(manifest, !state.Builtin)
	if err != nil {
		return state, err
	}
	m := state.Manifest
	state.Member.Key = m.Key
	state.Member.Version = m.Version
	state.Member.Contract = m.Contract
	state.Member.ProfileID = m.ProfileID
	state.Member.ProfileVersion = m.ProfileVersion
	if envelopeID != nil {
		envelope, err := decodeSuccessfulEnvelope(encoded)
		if err != nil {
			return state, err
		}
		state.Envelope = &packstate.Envelope{ID: *envelopeID, VerifiedAt: envelope.VerifiedAt, ValidUntil: envelope.TrustValidUntil}
		if envelope.ContainerSHA256 != nil {
			state.Envelope.ContainerSHA256 = *envelope.ContainerSHA256
		}
	}
	return state, nil
}
