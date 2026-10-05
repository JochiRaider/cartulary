package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// IncidentReferences retains source references as historical evidence. It does
// not make source trust, activation or provenance authoritative locally.
type IncidentReferences interface {
	BeginImportExecution(context.Context, jobs.Execution, time.Time) (*IncidentReferenceExecution, error)
	PrepareImport(context.Context, IncidentReferenceImportRequest) (*PreparedReferenceImport, error)
	ApplyImportTx(context.Context, pgx.Tx, *PreparedReferenceImport) error
	ExportTx(context.Context, pgx.Tx, uuid.UUID, []SetBinding) ([]byte, error)
	ExportContentTx(context.Context, pgx.Tx, IncidentReferenceExportRequest) ([]byte, error)
}

// A nil content pointer means the transport member is absent. A present empty
// document is invalid; it is never repaired into the absence default.
type IncidentReferenceImportRequest struct {
	References                       []byte
	ContentManifest                  *[]byte
	EmbeddedPaths                    []string
	IncidentID, OperationID, ActorID uuid.UUID
	At                               time.Time
	Execution                        jobs.Execution
	ExecutionStarted                 time.Time
	ExecutionScope                   *IncidentReferenceExecution
	OpenContainer                    func(context.Context, string) (io.ReadCloser, error)
}

type portableExecutionGuard interface {
	ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error
}

type IncidentReferenceOptions struct {
	Postgres      postgres.DB
	Storage       ArtifactStorage
	Configuration Configuration
	Limits        Limits
	JobExecutions portableExecutionGuard
	JobOperations referenceJobOperations
	Observer      OperationObserver
	Now           func() time.Time
}

type incidentReferences struct {
	pool       postgres.DB
	storage    ArtifactStorage
	verifier   *verificationService
	executions portableExecutionGuard
}

func NewIncidentReferences(options IncidentReferenceOptions) (IncidentReferences, error) {
	if options.Postgres == nil || options.Storage == nil || options.JobExecutions == nil || options.JobOperations == nil || options.Now == nil {
		return nil, errors.New("reference pack: incident retention dependencies required")
	}
	if err := validateCoordinatorLimits(options.Limits); err != nil {
		return nil, err
	}
	c := &verificationService{referenceDependencies: &referenceDependencies{pool: options.Postgres, storage: options.Storage, configuration: options.Configuration, limits: options.Limits, observer: options.Observer, now: options.Now}, operations: options.JobOperations}
	return &incidentReferences{pool: options.Postgres, storage: options.Storage, verifier: c, executions: options.JobExecutions}, nil
}

// PreparedReferenceImport is deliberately opaque to Incident Bundles. Reuse
// is prepared outside publication locks and rechecked under ordered key guards.
type PreparedReferenceImport struct {
	owner                      *incidentReferences
	incident, operation, actor uuid.UUID
	at                         time.Time
	refs                       IncidentBundleReferences
	canonical                  []byte
	input                      []byte
	replay                     bool
	versions                   []importedReferenceVersion
	cohort                     *portableVerificationCohort
	attempt                    *executionAttempt
}

type importedReferenceVersion struct {
	reference IncidentBundleVersionReference
	revision  int64
	envelope  string
	available bool
	reason    *string
}

type referenceImportResolution struct {
	SchemaID string                             `json:"schema_id"`
	Versions []referenceImportVersionResolution `json:"versions"`
}
type referenceImportVersionResolution struct {
	Key       string  `json:"pack_key"`
	Version   string  `json:"pack_version"`
	Available bool    `json:"available"`
	Reason    *string `json:"reason_code"`
}

func (r *incidentReferences) PrepareImport(ctx context.Context, request IncidentReferenceImportRequest) (result *PreparedReferenceImport, resultErr error) {
	payload, incident, operation, actor, at := request.References, request.IncidentID, request.OperationID, request.ActorID, request.At
	if incident == uuid.Nil || operation == uuid.Nil || actor == uuid.Nil || at.IsZero() {
		return nil, errors.New("reference pack: invalid incident retention attribution")
	}
	refs, err := DecodeIncidentBundleReferences(payload)
	if err != nil {
		return nil, err
	}
	canonical, err := canonicaljson.Marshal(refs)
	if err != nil {
		return nil, err
	}
	content := packformat.EmptyPortableContent()
	if request.ContentManifest != nil {
		content, err = packformat.DecodePortableContent(*request.ContentManifest, refs.format())
		if err != nil {
			var failure *packformat.Failure
			if errors.As(err, &failure) {
				return nil, &IncidentBundleReferenceValidationError{InvariantID: failure.Code}
			}
			return nil, err
		}
	}
	if err := packformat.ValidatePortableContainerInventory(content, request.EmbeddedPaths); err != nil {
		return nil, &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceIdentityInvariant}
	}
	input, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_portability_input.v1", "incident_id": incident.String(), "source_operation_id": operation.String(), "catalog_sha256": packformat.Digest(canonical), "content_manifest": content})
	if err != nil {
		return nil, err
	}
	required := packformat.RequiredPortableVersions(content, refs.format())
	p := &PreparedReferenceImport{owner: r, incident: incident, operation: operation, actor: actor, at: at.UTC(), refs: refs, canonical: canonical, input: input}
	// Proven exact replay precedes fresh availability and revision checks.
	var previous, previousInput []byte
	var previousOperation, previousActor uuid.UUID
	err = r.pool.QueryRow(ctx, `SELECT c.canonical_references,c.operation_id,o.actor_user_id,o.frozen_input FROM reference_pack_portable_catalogs c JOIN reference_pack_operations o USING(operation_id) WHERE c.incident_id=$1`, incident).Scan(&previous, &previousOperation, &previousActor, &previousInput)
	if err == nil {
		if previousOperation != uuid.NewSHA1(operation, []byte("reference_pack:incident_retention")) || previousActor != actor || !bytes.Equal(previous, canonical) || !bytes.Equal(previousInput, input) {
			return nil, errors.New("reference pack: incident reference replay mismatch")
		}
		p.replay = true
		return p, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	scope := request.ExecutionScope
	if scope != nil && (scope.owner != r || scope.execution != request.Execution || scope.started != request.ExecutionStarted) {
		return nil, errors.New("reference pack: foreign parent execution scope")
	}
	if scope != nil {
		ctx = scope.Context()
	}
	cohort, found, err := r.loadPortablePreparation(ctx, request, p)
	if err != nil {
		return nil, err
	}
	if found {
		if scope == nil {
			return nil, errors.New("reference pack: fresh import execution scope required")
		}
		scope.operation = cohort.operation
		if err := scope.activate(cohort.context.TimeoutSeconds); err != nil {
			return nil, err
		}
	} else {
		inputs := map[int]portableContainerInput{}
		transferred := false
		defer func() {
			if !transferred {
				var cleanup error
				for _, input := range inputs {
					cleanup = errors.Join(cleanup, r.storage.RemovePublished(input.object.Reference), input.object.releasePublication())
				}
				if cleanup != nil {
					result = nil
					resultErr = cleanup
				}
			}
		}()
		paths := map[string]bool{}
		for _, path := range request.EmbeddedPaths {
			paths[path] = true
		}
		descriptors := map[string]packformat.PortableContainer{}
		for _, descriptor := range content.Containers {
			path, err := packformat.PortableContainerPath(descriptor.ManifestSHA256)
			if err != nil {
				return nil, err
			}
			if paths[path] {
				descriptors[descriptor.ManifestSHA256] = descriptor
			}
		}
		for index, reference := range refs.Versions {
			version, err := r.prepareVersion(ctx, reference)
			if err != nil {
				return nil, err
			}
			p.versions = append(p.versions, version)
			if version.available {
				continue
			}
			if descriptor, present := descriptors[reference.ManifestSHA256]; present {
				if scope == nil {
					return nil, errors.New("reference pack: fresh import execution scope required")
				}
				if err := scope.activate(r.verifier.limits.ReferencePacks.MaxVerificationSeconds); err != nil {
					return nil, err
				}
				input, err := r.preparePortableContainer(ctx, request, descriptor, reference)
				if err != nil {
					return nil, err
				}
				if input != nil {
					inputs[index] = *input
					continue
				}
			}
			if required[reference.Key+"\x00"+reference.Version] {
				return nil, &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceDegradationInvariant}
			}
		}
		if len(inputs) == 0 {
			return p, nil
		}
		transferred = true // admission owns cleanup, including uncertain-commit retention
		cohort, err = r.admitPortablePreparation(ctx, request, p, inputs)
		if err != nil {
			return nil, err
		}
		scope.operation = cohort.operation
	}
	attempt, err := r.verifyPortablePreparation(ctx, request, p, cohort)
	if err != nil {
		return nil, err
	}
	p.cohort = cohort
	p.attempt = &attempt
	for ordinal, sourceIndex := range cohort.order {
		v := p.versions[sourceIndex].reference
		if !required[v.Key+"\x00"+v.Version] {
			continue
		}
		var rejected bool
		if err := r.pool.QueryRow(ctx, `SELECT verdict='content_rejected' FROM reference_pack_attempt_members WHERE attempt_id=$1 AND ordinal=$2`, attempt.ID, ordinal+1).Scan(&rejected); err != nil {
			return nil, err
		}
		if rejected {
			scope.requiredRejection = p
			return nil, &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceDegradationInvariant}
		}
	}
	return p, nil
}

func (r *incidentReferences) prepareVersion(ctx context.Context, reference IncidentBundleVersionReference) (importedReferenceVersion, error) {
	v := importedReferenceVersion{reference: reference}
	var manifestBytes, envelopeBytes []byte
	var manifestDigest, payloadDigest, health string
	var removed, indexed bool
	err := r.pool.QueryRow(ctx, `SELECT k.revision,v.manifest_sha256,v.payload_sha256,v.manifest_bytes,
 c.health,c.removed,e.envelope_id,e.canonical_envelope,
 EXISTS(SELECT 1 FROM reference_pack_index_generations i WHERE i.index_id=c.current_index_id AND i.complete
 AND i.manifest_sha256=v.manifest_sha256 AND i.payload_sha256=v.payload_sha256)
 FROM reference_pack_versions v JOIN reference_pack_candidates c USING(pack_key,pack_version)
 JOIN reference_pack_key_state k USING(pack_key)
 JOIN reference_pack_envelopes e ON e.envelope_id=c.current_envelope_id
 WHERE v.pack_key=$1 AND v.pack_version=$2`, reference.Key, reference.Version).Scan(&v.revision, &manifestDigest, &payloadDigest, &manifestBytes, &health, &removed, &v.envelope, &envelopeBytes, &indexed)
	if errors.Is(err, pgx.ErrNoRows) {
		err = r.pool.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1`, reference.Key).Scan(&v.revision)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return v, err
		}
		return unavailableReference(v, "not_retained"), nil
	}
	if err != nil {
		return v, err
	}
	if manifestDigest != reference.ManifestSHA256 || payloadDigest != reference.PayloadSHA256 {
		return v, &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceIdentityInvariant}
	}
	envelope, err := decodeSuccessfulEnvelope(envelopeBytes)
	if err != nil {
		return v, err
	}
	manifest, err := packformat.DecodeManifest(manifestBytes, envelope.DistributionKind == "operator_imported")
	if err != nil {
		return v, err
	}
	method := "packaged_release_manifest_v1"
	if envelope.DistributionKind == "operator_imported" {
		method = "tuf_1_0_35_offline_bundle_v1"
	}
	if reference.Contract != manifest.Contract || reference.ProfileID != manifest.ProfileID || reference.ProfileVersion != manifest.ProfileVersion ||
		reference.DistributionKind != envelope.DistributionKind || reference.VerificationMethod != method ||
		reference.SourceProfileID != manifest.SourceProfileID || reference.SourceProfileSHA256 != manifest.SourceProfileSHA256 {
		return v, &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceIdentityInvariant}
	}
	if removed {
		return unavailableReference(v, "removed"), nil
	}
	if health != "verified_available" || !indexed {
		return unavailableReference(v, "not_usable"), nil
	}
	if err := checkRetainedVersion(ctx, r.pool, r.storage, reference.Key, reference.Version); err != nil {
		var rejected *ContentRejection
		if errors.As(err, &rejected) {
			return unavailableReference(v, "content_unavailable"), nil
		}
		return v, err
	}
	v.available = true
	return v, nil
}

func unavailableReference(v importedReferenceVersion, reason string) importedReferenceVersion {
	v.reason = &reason
	return v
}

func (r *incidentReferences) ApplyImportTx(ctx context.Context, tx pgx.Tx, p *PreparedReferenceImport) error {
	if tx == nil || p == nil || p.owner != r {
		return errors.New("reference pack: foreign incident preparation")
	}
	operation := uuid.NewSHA1(p.operation, []byte("reference_pack:incident_retention"))
	var previous, previousInput []byte
	var previousOperation, previousActor uuid.UUID
	err := tx.QueryRow(ctx, `SELECT c.canonical_references,c.operation_id,o.actor_user_id,o.frozen_input FROM reference_pack_portable_catalogs c JOIN reference_pack_operations o USING(operation_id) WHERE c.incident_id=$1`, p.incident).Scan(&previous, &previousOperation, &previousActor, &previousInput)
	if err == nil {
		if previousOperation != operation || previousActor != p.actor || !bytes.Equal(previous, p.canonical) || !bytes.Equal(previousInput, p.input) {
			return errors.New("reference pack: incident reference replay mismatch")
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if p.replay {
		return errors.New("reference pack: committed reference catalog disappeared")
	}
	if (p.cohort == nil) != (p.attempt == nil) {
		return errors.New("reference pack: incomplete portable execution")
	}
	if p.cohort != nil {
		copied := *p
		copied.versions = slices.Clone(p.versions)
		p = &copied
		if err := r.publishPortableAttemptTx(ctx, tx, p, p.cohort, *p.attempt, r.verifier.now().UTC()); err != nil {
			return err
		}
	} else {
		// Every selected key is serialized with removal, reimport and pinning. No
		// current-set or trust mutation is needed to retain historical references.
		keys := []string{}
		for _, v := range p.versions {
			keys = append(keys, v.reference.Key)
		}
		slices.Sort(keys)
		for _, key := range slices.Compact(keys) {
			inserted, err := tx.Exec(ctx, `INSERT INTO reference_pack_key_state(pack_key,revision) VALUES($1,1) ON CONFLICT DO NOTHING`, key)
			if err != nil {
				return err
			}
			var revision int64
			if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, key).Scan(&revision); err != nil {
				return err
			}
			for _, v := range p.versions {
				if v.reference.Key == key && ((v.revision == 0 && inserted.RowsAffected() != 1) || (v.revision != 0 && v.revision != revision)) {
					return &OperationRejection{Reason: "stale_admission_state"}
				}
			}
		}
	}
	var declared struct {
		Content packformat.PortableContent `json:"content_manifest"`
	}
	if err := json.Unmarshal(p.input, &declared); err != nil {
		return err
	}
	required := packformat.RequiredPortableVersions(declared.Content, p.refs.format())
	for _, version := range p.versions {
		if required[version.reference.Key+"\x00"+version.reference.Version] && !version.available {
			return &IncidentBundleReferenceValidationError{InvariantID: IncidentBundleReferenceDegradationInvariant}
		}
	}
	// The caller's operation identity is retained, with a separate owner UUID so
	// it cannot collide with an ordinary Reference Data lifecycle operation.
	input := p.input
	resolution := referenceImportResolution{SchemaID: "cartulary.reference_pack_portability_resolution.v1", Versions: []referenceImportVersionResolution{}}
	for _, v := range p.versions {
		resolution.Versions = append(resolution.Versions, referenceImportVersionResolution{v.reference.Key, v.reference.Version, v.available, v.reason})
	}
	result, err := canonicaljson.Marshal(resolution)
	if err != nil {
		return err
	}
	if err := packformat.ValidatePortabilityRetention(input, result, p.refs.format()); err != nil {
		return err
	}
	if p.cohort != nil {
		changed, err := tx.Exec(ctx, `UPDATE reference_pack_operations SET terminal_at=$2,final_outcome=$3 WHERE operation_id=$1 AND terminal_at IS NULL`, operation, r.verifier.now().UTC(), result)
		if err != nil {
			return err
		}
		if changed.RowsAffected() != 1 {
			return errors.New("reference pack: obsolete portable publication")
		}
	} else {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operations(operation_id,kind,actor_kind,actor_user_id,admitted_at,terminal_at,frozen_input,final_outcome) VALUES($1,'portable_retention','user',$2,$3,$3,$4,$5)`, operation, p.actor, p.at, input, result); err != nil {
			return err
		}
	}
	for _, v := range p.versions {
		if !v.available {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_version_pins(owner_kind,owner_id,pack_key,pack_version,envelope_id,operation_id) VALUES('incident_reference_catalog',$1,$2,$3,$4,$5)`, p.incident.String(), v.reference.Key, v.reference.Version, v.envelope, operation); err != nil {
			return err
		}
	}
	for _, set := range p.refs.Sets {
		complete := true
		for _, member := range set.Members {
			if !slices.ContainsFunc(p.versions, func(v importedReferenceVersion) bool { return v.available && v.reference.PackSetMember == member }) {
				complete = false
				break
			}
		}
		if !complete {
			continue
		}
		retained, err := retainSetTx(ctx, tx, set.Members, operation)
		if err != nil {
			return err
		}
		if retained.ID != set.ID {
			return errors.New("reference pack: incident set identity mismatch")
		}
		if err := insertPinTx(ctx, tx, set.ID, "incident_reference_catalog", p.incident.String(), operation); err != nil {
			return err
		}
	}
	_, err = tx.Exec(ctx, `INSERT INTO reference_pack_portable_catalogs(incident_id,operation_id,catalog_sha256,canonical_references,canonical_resolution) VALUES($1,$2,$3,$4,$5)`, p.incident, operation, packformat.Digest(p.canonical), p.canonical, result)
	return err
}

func (r *incidentReferences) ExportTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID, bindings []SetBinding) ([]byte, error) {
	if tx == nil || incident == uuid.Nil {
		return nil, errors.New("reference pack: invalid incident export")
	}
	var data []byte
	var digest string
	err := tx.QueryRow(ctx, `SELECT catalog_sha256,canonical_references FROM reference_pack_portable_catalogs WHERE incident_id=$1`, incident).Scan(&digest, &data)
	if errors.Is(err, pgx.ErrNoRows) {
		data, err = EncodeIncidentBundleReferences(nil, nil)
		digest = packformat.Digest(data)
	}
	if err != nil {
		return nil, err
	}
	if packformat.Digest(data) != digest {
		return nil, errHistoricalIntegrity
	}
	refs, err := DecodeIncidentBundleReferences(data)
	if err != nil {
		return nil, err
	}
	if len(bindings) > PortableReferenceSetLimit {
		return nil, consumerError("invalid_pack_request")
	}
	retained := &retention{repository: &canonicalRepository{pool: tx, storage: r.storage}}
	ids := make([]string, 0, len(bindings))
	for _, binding := range bindings {
		if err := retained.ValidateBinding(ctx, binding); err != nil {
			return nil, err
		}
		ids = append(ids, binding.SetID)
	}
	nativeBytes, err := retained.ExportReferencesTx(ctx, tx, ids)
	if err != nil {
		return nil, err
	}
	native, err := DecodeIncidentBundleReferences(nativeBytes)
	if err != nil {
		return nil, err
	}
	return mergeIncidentReferences(refs, native)
}

// The union preserves unavailable imported tuples and rejects conflicting
// historical facts; native references never replace source evidence.
func mergeIncidentReferences(imported, native IncidentBundleReferences) ([]byte, error) {
	sets := map[string]PackSet{}
	versions := map[string]IncidentBundleVersionReference{}
	for _, catalog := range []IncidentBundleReferences{imported, native} {
		for _, set := range catalog.Sets {
			if previous, ok := sets[set.ID]; ok && !reflect.DeepEqual(previous, set) {
				return nil, errHistoricalIntegrity
			}
			sets[set.ID] = set
		}
		for _, version := range catalog.Versions {
			key := version.Key + "\x00" + version.Version
			if previous, ok := versions[key]; ok && previous != version {
				return nil, errHistoricalIntegrity
			}
			versions[key] = version
		}
	}
	if len(sets) > PortableReferenceSetLimit || len(versions) > 16384 {
		return nil, consumerError("invalid_pack_request")
	}
	selected := make([]PackSet, 0, len(sets))
	for _, set := range sets {
		selected = append(selected, set)
	}
	referenced := make([]IncidentBundleVersionReference, 0, len(versions))
	for _, version := range versions {
		referenced = append(referenced, version)
	}
	return EncodeIncidentBundleReferences(selected, referenced)
}

func (r *incidentReferences) retainedContentTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID) (packformat.PortableContent, error) {
	if tx == nil || incident == uuid.Nil {
		return packformat.PortableContent{}, errors.New("reference pack: invalid content export")
	}
	var input, result, references []byte
	err := tx.QueryRow(ctx, `SELECT o.frozen_input,c.canonical_resolution,c.canonical_references FROM reference_pack_portable_catalogs c JOIN reference_pack_operations o USING(operation_id) WHERE c.incident_id=$1`, incident).Scan(&input, &result, &references)
	if errors.Is(err, pgx.ErrNoRows) {
		return packformat.EmptyPortableContent(), nil
	}
	if err != nil {
		return packformat.PortableContent{}, err
	}
	refs, err := DecodeIncidentBundleReferences(references)
	if err != nil || packformat.ValidatePortabilityRetention(input, result, refs.format()) != nil {
		return packformat.PortableContent{}, errHistoricalIntegrity
	}
	var frozen struct {
		Content packformat.PortableContent `json:"content_manifest"`
	}
	if err := json.Unmarshal(input, &frozen); err != nil {
		return packformat.PortableContent{}, errHistoricalIntegrity
	}
	if len(frozen.Content.RequiredMembers) == 0 {
		return packformat.EmptyPortableContent(), nil
	}
	frozen.Content.Containers = []packformat.PortableContainer{}
	return frozen.Content, nil
}
