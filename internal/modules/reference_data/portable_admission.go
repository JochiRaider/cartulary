package reference_data

import (
	"bytes"
	"container/heap"
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type portableVerificationCohort struct {
	operation uuid.UUID
	context   packformat.PortableVerificationContext
	// Keys are canonical source-catalog ordinals, starting at zero. The order
	// records verification ordinals and is immutable across execution attempts.
	inputs map[int]portableContainerInput
	order  []int
}

// Schedule only lexical dependencies on exact source tuples. This is not
// dependency validation: only fully verified predecessors may later satisfy a
// dependency. Stable source order resolves ties and orders residual cycles.
type portableOrdinalHeap []int

func (h portableOrdinalHeap) Len() int           { return len(h) }
func (h portableOrdinalHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h portableOrdinalHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *portableOrdinalHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *portableOrdinalHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

func portableVerificationOrder(versions []importedReferenceVersion, inputs map[int]portableContainerInput) []int {
	tuples := make(map[string]int, len(inputs))
	for index := range inputs {
		v := versions[index].reference
		tuples[v.Key+"\x00"+v.Version+"\x00"+v.PayloadSHA256] = index
	}
	incoming := make(map[int]int, len(inputs))
	successors := make(map[int][]int, len(inputs))
	ready := &portableOrdinalHeap{}
	for index, input := range inputs {
		incoming[index] = 0
		for _, d := range input.identity.Dependencies {
			if predecessor, ok := tuples[d.Key+"\x00"+d.Version+"\x00"+d.SHA256]; ok {
				incoming[index]++
				successors[predecessor] = append(successors[predecessor], index)
			}
		}
		if incoming[index] == 0 {
			heap.Push(ready, index)
		}
	}
	order := make([]int, 0, len(inputs))
	for ready.Len() > 0 {
		index := heap.Pop(ready).(int)
		order = append(order, index)
		delete(incoming, index)
		for _, successor := range successors[index] {
			incoming[successor]--
			if incoming[successor] == 0 {
				heap.Push(ready, successor)
			}
		}
	}
	residual := make([]int, 0, len(incoming))
	for index := range incoming {
		residual = append(residual, index)
	}
	slices.Sort(residual)
	return append(order, residual...)
}

// Admission retains private inputs, never content or trust. Its Jobs lease,
// repository/key/configuration guards and immutable rows make restarts replay
// exactly the same cohort. A commit acknowledgement failure keeps input bytes.
func (r *incidentReferences) admitPortablePreparation(ctx context.Context, request IncidentReferenceImportRequest, p *PreparedReferenceImport, inputs map[int]portableContainerInput) (cohort *portableVerificationCohort, resultErr error) {
	keep := false
	defer func() {
		for index, input := range inputs {
			_ = input.object.releasePublication()
			input.object.publication = nil
			inputs[index] = input
		}
	}()
	defer func() {
		if !keep {
			var cleanup error
			for _, input := range inputs {
				cleanup = errors.Join(cleanup, r.storage.RemovePublished(input.object.Reference), input.object.releasePublication())
			}
			if cleanup != nil {
				resultErr = cleanup
				cohort = nil
			}
		}
	}()
	if len(inputs) == 0 || request.Execution.JobID() != p.operation {
		return nil, errors.New("reference pack: portable verification requires its parent execution")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if err := lockByteRetentionTx(ctx, tx, false); err != nil {
		return nil, err
	}
	if err := r.executions.ValidateExecutionTx(ctx, tx, request.Execution); err != nil {
		return nil, err
	}
	cohort = &portableVerificationCohort{operation: uuid.NewSHA1(p.operation, []byte("reference_pack:incident_retention")), inputs: inputs, order: portableVerificationOrder(p.versions, inputs)}
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_operations WHERE operation_id=$1)`, cohort.operation).Scan(&exists); err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("reference pack: existing portable preparation must be resumed")
	}
	type rootRevision struct {
		id                string
		revision, version int64
	}
	repositories := []string{}
	dependencies := []packformat.Dependency{}
	freshKeys := []string{}
	for _, input := range inputs {
		if input.identity.Repository != "" {
			repositories = append(repositories, input.identity.Repository)
		}
		dependencies = append(dependencies, input.identity.Dependencies...)
		freshKeys = append(freshKeys, input.identity.Key)
	}
	slices.Sort(repositories)
	roots := []rootRevision{}
	for _, id := range slices.Compact(repositories) {
		root := rootRevision{id: id}
		err := tx.QueryRow(ctx, `SELECT revision,root_version FROM reference_pack_repositories WHERE repository_id=$1 FOR UPDATE`, id).Scan(&root.revision, &root.version)
		if errors.Is(err, pgx.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		roots = append(roots, root)
	}
	// Reuse the ordinary import's transitive retained-dependency inventory.
	admission := operationAdmission{Kind: "import", Import: &importIdentity{Dependencies: dependencies}}
	captured, err := admissionDependencyRevisions(ctx, tx, admission, nil)
	if err != nil {
		return nil, err
	}
	selected := []string{}
	for _, v := range p.versions {
		selected = append(selected, v.reference.Key)
	}
	revisions := map[string]int64{}
	for _, key := range admissionGuardKeys(selected, captured) {
		inserted, err := tx.Exec(ctx, `INSERT INTO reference_pack_key_state(pack_key) VALUES($1) ON CONFLICT DO NOTHING`, key)
		if err != nil {
			return nil, err
		}
		var revision int64
		if err := tx.QueryRow(ctx, `SELECT revision FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, key).Scan(&revision); err != nil {
			return nil, err
		}
		revisions[key] = revision
		for _, v := range p.versions {
			if v.reference.Key == key && (v.revision == 0 && inserted.RowsAffected() != 1 || v.revision != 0 && v.revision != revision) {
				return nil, &OperationRejection{Reason: "stale_admission_state"}
			}
		}
	}
	confirmed, err := admissionDependencyRevisions(ctx, tx, admission, nil)
	if err != nil {
		return nil, err
	}
	if !slices.Equal(captured, confirmed) {
		return nil, &OperationRejection{Reason: "stale_admission_state"}
	}
	for _, d := range captured {
		if revisions[d.key] != d.revision {
			return nil, &OperationRejection{Reason: "stale_admission_state"}
		}
	}
	var pending bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reference_pack_operation_keys k JOIN reference_pack_operations o USING(operation_id) WHERE k.pack_key=ANY($1) AND o.terminal_at IS NULL)`, freshKeys).Scan(&pending); err != nil {
		return nil, err
	}
	if pending {
		return nil, &OperationRejection{Reason: "verification_pending"}
	}
	cohort.context = packformat.PortableVerificationContext{SchemaID: "cartulary.reference_pack_portable_verification_context.v1", ClockTrusted: r.verifier.configuration.ClockTrusted, TimeoutSeconds: r.verifier.limits.ReferencePacks.MaxVerificationSeconds}
	if err := tx.QueryRow(ctx, `SELECT configuration_sha256 FROM reference_pack_current_set WHERE singleton FOR UPDATE`).Scan(&cohort.context.ConfigurationSHA256); err != nil {
		return nil, err
	}
	if !cohort.context.ClockTrusted {
		return nil, &OperationRejection{Reason: "clock_untrusted"}
	}
	contextBytes, err := canonicaljson.Marshal(cohort.context)
	if err != nil {
		return nil, err
	}
	if _, err := packformat.DecodePortableVerificationContext(contextBytes); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operations(operation_id,job_id,kind,actor_kind,actor_user_id,admitted_at,frozen_input) VALUES($1,$2,'portable_retention','user',$3,$4,$5)`, cohort.operation, p.operation, p.actor, p.at, p.input); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_portable_preparations(operation_id,canonical_references,canonical_context) VALUES($1,$2,$3)`, cohort.operation, p.canonical, contextBytes); err != nil {
		return nil, err
	}
	slices.Sort(selected)
	for _, key := range slices.Compact(selected) {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_keys(operation_id,pack_key,admitted_revision) VALUES($1,$2,$3)`, cohort.operation, key, revisions[key]); err != nil {
			return nil, err
		}
	}
	for _, d := range captured {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_dependency_keys(operation_id,pack_key,admitted_revision) VALUES($1,$2,$3)`, cohort.operation, d.key, d.revision); err != nil {
			return nil, err
		}
	}
	for _, root := range roots {
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_repositories(operation_id,repository_id,admitted_revision,root_version) VALUES($1,$2,$3,$4)`, cohort.operation, root.id, root.revision, root.version); err != nil {
			return nil, err
		}
	}
	ordinals := map[int]int64{}
	for ordinal, index := range cohort.order {
		input := inputs[index]
		ordinals[index] = int64(ordinal + 1)
		if err := insertCandidateTx(ctx, tx, input.identity.Key, input.identity.Version, "operator_imported", &p.actor, p.at); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_operation_members(operation_id,ordinal,pack_key,pack_version,envelope_id) SELECT $1,$2,pack_key,pack_version,current_envelope_id FROM reference_pack_candidates WHERE pack_key=$3 AND pack_version=$4`, cohort.operation, ordinal+1, input.identity.Key, input.identity.Version); err != nil {
			return nil, err
		}
		object := input.object
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_objects(object_id,sha256,storage_ref,size_bytes,generation) VALUES($1,$2,$3,$4,1)`, object.ID, object.Digest, object.Reference.String(), object.Size); err != nil {
			return nil, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_object_refs(owner_kind,owner_id,logical_path,object_id) VALUES('operation',$1,$2,$3)`, cohort.operation.String(), fmt.Sprintf("input/%d/container", ordinal+1), object.ID); err != nil {
			return nil, err
		}
	}
	for index, v := range p.versions {
		var envelope *string
		if v.envelope != "" {
			envelope = &v.envelope
		}
		var ordinal *int64
		var object *uuid.UUID
		var digest *string
		var size *int64
		if input, ok := inputs[index]; ok {
			n := ordinals[index]
			ordinal = &n
			object = &input.object.ID
			digest = &input.object.Digest
			size = &input.object.Size
		}
		if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_portable_selections(operation_id,ordinal,pack_key,pack_version,envelope_id,available,reason_code,verification_ordinal,input_object_id,expected_container_sha256,expected_container_bytes) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, cohort.operation, index+1, v.reference.Key, v.reference.Version, envelope, v.available, v.reason, ordinal, object, digest, size); err != nil {
			return nil, err
		}
	}
	keep = true // Commit may be authoritative even when acknowledgement is lost.
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return cohort, nil
}

// Resume consults no current availability, source reader, or newly admitted
// versions. Frozen bytes and source selections survive a new Jobs lease.
func (r *incidentReferences) loadPortablePreparation(ctx context.Context, request IncidentReferenceImportRequest, p *PreparedReferenceImport) (*portableVerificationCohort, bool, error) {
	operation := uuid.NewSHA1(p.operation, []byte("reference_pack:incident_retention"))
	var input, references, contextBytes []byte
	var actor, job uuid.UUID
	var terminal *time.Time
	err := r.pool.QueryRow(ctx, `SELECT o.frozen_input,o.actor_user_id,o.job_id,o.terminal_at,p.canonical_references,p.canonical_context FROM reference_pack_operations o JOIN reference_pack_portable_preparations p USING(operation_id) WHERE operation_id=$1`, operation).Scan(&input, &actor, &job, &terminal, &references, &contextBytes)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if actor != p.actor || job != p.operation || !bytes.Equal(input, p.input) || !bytes.Equal(references, p.canonical) || terminal != nil || request.Execution.JobID() != job {
		return nil, true, errors.New("reference pack: portable preparation replay mismatch")
	}
	frozen, err := packformat.DecodePortableVerificationContext(contextBytes)
	if err != nil {
		return nil, true, err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, true, err
	}
	defer tx.Rollback(ctx)
	if err := r.executions.ValidateExecutionTx(ctx, tx, request.Execution); err != nil {
		return nil, true, err
	}
	cohort := &portableVerificationCohort{operation: operation, context: frozen, inputs: map[int]portableContainerInput{}}
	rows, err := tx.Query(ctx, `SELECT s.ordinal,s.pack_key,s.pack_version,s.envelope_id,s.available,s.reason_code,k.admitted_revision,s.verification_ordinal,o.object_id,o.sha256,o.storage_ref,o.size_bytes,s.expected_container_sha256,s.expected_container_bytes FROM reference_pack_portable_selections s JOIN reference_pack_operation_keys k USING(operation_id,pack_key) LEFT JOIN reference_pack_objects o ON o.object_id=s.input_object_id WHERE s.operation_id=$1 ORDER BY s.ordinal`, operation)
	if err != nil {
		return nil, true, err
	}
	defer rows.Close()
	order := map[int64]int{}
	p.versions = nil
	for rows.Next() {
		var ordinal int
		var key, version string
		var envelope, reference, digest, expectedDigest *string
		var verificationOrdinal, size, expectedSize *int64
		var object *uuid.UUID
		v := importedReferenceVersion{}
		if err := rows.Scan(&ordinal, &key, &version, &envelope, &v.available, &v.reason, &v.revision, &verificationOrdinal, &object, &digest, &reference, &size, &expectedDigest, &expectedSize); err != nil {
			return nil, true, err
		}
		if ordinal != len(p.versions)+1 || ordinal > len(p.refs.Versions) {
			return nil, true, errHistoricalIntegrity
		}
		v.reference = p.refs.Versions[ordinal-1]
		if key != v.reference.Key || version != v.reference.Version {
			return nil, true, errHistoricalIntegrity
		}
		if envelope != nil {
			v.envelope = *envelope
		}
		if verificationOrdinal != nil {
			if object == nil || digest == nil || reference == nil || size == nil || expectedDigest == nil || expectedSize == nil || *digest != *expectedDigest || *size != *expectedSize {
				return nil, true, errHistoricalIntegrity
			}
			ref, err := ParseStorageRef(*reference)
			if err != nil {
				return nil, true, errHistoricalIntegrity
			}
			cohort.inputs[ordinal-1] = portableContainerInput{descriptor: packformat.PortableContainer{ManifestSHA256: v.reference.ManifestSHA256, ContainerSHA256: *digest, SizeBytes: *size}, object: preparedObject{ID: *object, Path: "input/container", Digest: *digest, Reference: ref, Size: *size}}
			order[*verificationOrdinal] = ordinal - 1
		}
		p.versions = append(p.versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, true, err
	}
	rows.Close()
	if len(p.versions) != len(p.refs.Versions) || len(order) == 0 {
		return nil, true, errHistoricalIntegrity
	}
	for ordinal := 1; ordinal <= len(order); ordinal++ {
		index, ok := order[int64(ordinal)]
		if !ok {
			return nil, true, errHistoricalIntegrity
		}
		cohort.order = append(cohort.order, index)
	}
	return cohort, true, nil
}
