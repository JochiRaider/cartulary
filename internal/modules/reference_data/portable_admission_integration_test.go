package reference_data

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testPortablePreparationAdmission(t *testing.T, pool *pgxpool.Pool, storage *coordinatorMemoryStorage, actor uuid.UUID, at time.Time) {
	t.Helper()
	ctx := context.Background()
	definitions := []jobs.Definition{{JobKind: "test.portable", ProgressUnitID: "reference_pack.import.request.v1", HandlerName: "test_portable"}}
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		t.Fatal(err)
	}
	transactions := collaborationsupport.NewJobTransactionsWithTerminalEffects(catalog, JobTerminalEffects{}, collaborationsupport.TestWorkerRuntimeContracts(definitions))
	policy := jobs.ProductionRuntimePolicy()
	policy.MaximumFailures = 1
	policy.RetryDelays = nil
	manager, err := jobs.NewManager(jobs.ManagerOptions{Postgres: pool, Transactions: transactions, Catalog: catalog, Policy: policy, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	r := &incidentReferences{pool: pool, storage: storage, executions: transactions, verifier: &Coordinator{pool: pool, storage: storage, configuration: Configuration{ClockTrusted: true}, limits: DefaultLimits(), now: time.Now, operations: manager}}
	raw, err := os.ReadFile("../../../contracts/reference-packs/fixtures/portable-input.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Allowed portableInputVector }
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	vector := fixture.Allowed
	var trustVector struct {
		Bootstrap  string `json:"bootstrap"`
		Repository string `json:"repository_id"`
	}
	trustBytes, err := os.ReadFile("../../../contracts/reference-packs/fixtures/signed-container.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(trustBytes, &trustVector); err != nil {
		t.Fatal(err)
	}
	root, err := packformat.AdmitBootstrap([]byte(trustVector.Bootstrap))
	if err != nil {
		t.Fatal(err)
	}
	if err := ReconcileTrustBootstrap(ctx, pool, TrustBootstrap{repositories: map[string]packformat.TrustSnapshot{trustVector.Repository: root}}, at); err != nil {
		t.Fatal(err)
	}
	r.verifier.now = func() time.Time { return at }
	body, err := base64.StdEncoding.DecodeString(vector.Container)
	if err != nil {
		t.Fatal(err)
	}
	// Build the source catalog from the independent signed tuple and local Base.
	repository := canonicalRepository{pool: pool, storage: storage}
	current, err := repository.CurrentSet(ctx)
	if err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	retained := retention{repository: &repository}
	data, err := retained.ExportReferencesTx(ctx, tx, []string{current.ID})
	_ = tx.Rollback(ctx)
	if err != nil {
		t.Fatal(err)
	}
	refs, err := DecodeIncidentBundleReferences(data)
	if err != nil {
		t.Fatal(err)
	}
	for index := range refs.Versions {
		if refs.Versions[index].Key == vector.Reference.Key {
			refs.Versions[index] = vector.Reference
		}
	}
	for index := range refs.Sets[0].Members {
		if refs.Sets[0].Members[index].Key == vector.Reference.Key {
			refs.Sets[0].Members[index] = vector.Reference.SetMember
		}
	}
	refs.Sets[0], err = packformat.BuildSet(refs.Sets[0].Members)
	if err != nil {
		t.Fatal(err)
	}
	data, err = EncodeIncidentBundleReferences(refs.Sets, refs.Versions)
	if err != nil {
		t.Fatal(err)
	}
	refs, err = DecodeIncidentBundleReferences(data)
	if err != nil {
		t.Fatal(err)
	}
	data, err = canonicaljson.Marshal(refs)
	if err != nil {
		t.Fatal(err)
	}
	content := packformat.EmptyPortableContent()
	content.Containers = []packformat.PortableContainer{vector.Descriptor}
	newRequest := func() (IncidentReferenceImportRequest, *PreparedReferenceImport, int) {
		t.Helper()
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		job, err := transactions.CreateQueuedTx(ctx, tx, jobs.EnqueueParams{JobKind: "test.portable", Scope: jobs.Scope{Kind: jobs.ScopeKindDeployment}, SubmittedByUserID: actor, Cancelable: true, Progress: jobs.Progress{Completed: 0, Total: intPtr(1)}}, time.Now())
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		jobID := uuid.MustParse(job.JobID)
		execution, claimed, err := manager.Claim(ctx, jobID)
		if err != nil || !claimed {
			t.Fatal(claimed, err)
		}
		request := IncidentReferenceImportRequest{References: data, IncidentID: uuid.New(), OperationID: jobID, ActorID: actor, At: at, Execution: execution, ExecutionStarted: time.Now(), OpenContainer: func(context.Context, string) (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(body)), nil }}
		input, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_portability_input.v1", "incident_id": request.IncidentID.String(), "source_operation_id": jobID.String(), "catalog_sha256": packformat.Digest(data), "content_manifest": content})
		if err != nil {
			t.Fatal(err)
		}
		p := &PreparedReferenceImport{owner: r, incident: request.IncidentID, operation: jobID, actor: actor, at: at, refs: refs, canonical: data, input: input}
		fresh := -1
		for index, reference := range refs.Versions {
			v, err := r.prepareVersion(ctx, reference)
			if err != nil {
				t.Fatal(err)
			}
			p.versions = append(p.versions, v)
			if reference.Key == vector.Reference.Key {
				fresh = index
			}
		}
		return request, p, fresh
	}
	prepare := func(request IncidentReferenceImportRequest) portableContainerInput {
		t.Helper()
		input, err := r.preparePortableContainer(ctx, request, vector.Descriptor, vector.Reference)
		if err != nil || input == nil {
			t.Fatal("prepare transport", err)
		}
		return *input
	}
	reject := func(request IncidentReferenceImportRequest, p *PreparedReferenceImport, index int, reason string) {
		t.Helper()
		input := prepare(request)
		_, err := r.admitPortablePreparation(ctx, request, p, map[int]portableContainerInput{index: input})
		var rejected *OperationRejection
		if !errors.As(err, &rejected) || rejected.Reason != reason {
			t.Fatal("admission rejection", reason, err)
		}
		if _, exists := storage.objects[input.object.Reference.String()]; exists {
			t.Fatal("uncommitted input retained")
		}
		var count int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operations WHERE job_id=$1`, p.operation).Scan(&count); err != nil || count != 0 {
			t.Fatal("rejection mutated operation", count, err)
		}
		if err := manager.RecordExecutionFailure(ctx, request.Execution, false); err != nil {
			t.Fatal(err)
		}
	}
	request, p, index := newRequest()
	r.verifier.configuration.ClockTrusted = false
	reject(request, p, index, "clock_untrusted")
	r.verifier.configuration.ClockTrusted = true
	request, p, index = newRequest()
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, vector.Reference.Key); err != nil {
		t.Fatal(err)
	}
	reject(request, p, index, "stale_admission_state")
	request, p, index = newRequest()
	input := prepare(request)
	cohort, err := r.admitPortablePreparation(ctx, request, p, map[int]portableContainerInput{index: input})
	if err != nil {
		t.Fatal("admit private cohort", err)
	}
	var available, attempts int
	if err := pool.QueryRow(ctx, `SELECT (SELECT count(*) FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2),(SELECT count(*) FROM reference_pack_attempts WHERE operation_id=$3)`, vector.Reference.Key, vector.Reference.Version, cohort.operation).Scan(&available, &attempts); err != nil || available != 0 || attempts != 0 {
		t.Fatal("admission created successful facts", available, attempts, err)
	}
	firstAttempt, err := r.verifyPortablePreparation(ctx, request, p, cohort)
	if err != nil {
		t.Fatal("private destination verification", err)
	}
	var verdict string
	var envelope []byte
	if err := pool.QueryRow(ctx, `SELECT verdict,canonical_prepared FROM reference_pack_attempt_members WHERE attempt_id=$1`, firstAttempt.ID).Scan(&verdict, &envelope); err != nil || verdict != "succeeded" {
		t.Fatal("destination verdict", verdict, err)
	}
	private, err := decodePrepared(envelope)
	if err != nil || private.Envelope.OperationID != cohort.operation || !private.Envelope.VerifiedAt.Equal(at) {
		t.Fatal("private successful envelope", err)
	}
	r.verifier.now = func() time.Time { return at.Add(time.Second) }
	retriedAttempt, err := r.verifyPortablePreparation(ctx, request, p, cohort)
	if err != nil || retriedAttempt.ID == firstAttempt.ID || !retriedAttempt.Start.Equal(at.Add(time.Second)) {
		t.Fatal("restart freshness", err)
	}
	var prior string
	if err := pool.QueryRow(ctx, `SELECT outcome FROM reference_pack_attempts WHERE attempt_id=$1`, firstAttempt.ID).Scan(&prior); err != nil || prior != "interrupted" {
		t.Fatal("previous execution history", prior, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2`, vector.Reference.Key, vector.Reference.Version).Scan(&available); err != nil || available != 0 {
		t.Fatal("private verification published content", available, err)
	}
	publication := *p
	publication.versions = append([]importedReferenceVersion{}, p.versions...)
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.publishPortableAttemptTx(ctx, tx, &publication, cohort, retriedAttempt, at.Add(2*time.Second)); err != nil {
		_ = tx.Rollback(ctx)
		t.Fatal("private publication rollback", err)
	}
	if !publication.versions[index].available || publication.versions[index].envelope == "" {
		t.Fatal("publication did not retain envelope")
	}
	_ = tx.Rollback(ctx)
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM reference_pack_versions WHERE pack_key=$1 AND pack_version=$2`, vector.Reference.Key, vector.Reference.Version).Scan(&available); err != nil || available != 0 {
		t.Fatal("rolled-back incident published content", available, err)
	}
	conflicting, other, otherIndex := newRequest()
	reject(conflicting, other, otherIndex, "verification_pending")
	// Changing current state cannot rebase retry selections. Publication guards
	// subsequently reject the stale revision; replay still returns frozen facts.
	if _, err := pool.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, vector.Reference.Key); err != nil {
		t.Fatal(err)
	}
	retry := *p
	retry.versions = nil
	request.OpenContainer = func(context.Context, string) (io.ReadCloser, error) {
		t.Fatal("retry reopened source")
		return nil, nil
	}
	loaded, found, err := r.loadPortablePreparation(ctx, request, &retry)
	// Namespace leases are process-local; only durable byte identity is replayed.
	input.object.publication = nil
	if err != nil || !found || loaded.operation != cohort.operation || loaded.inputs[index].object != input.object || retry.versions[index].revision != p.versions[index].revision {
		t.Fatal("frozen replay", found, err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = lockPublicationDependenciesTx(ctx, tx, cohort.operation)
	_ = tx.Rollback(ctx)
	var stale *OperationRejection
	if !errors.As(err, &stale) || stale.Reason != "stale_admission_state" {
		t.Fatal("stale publication", err)
	}
	altered := *p
	altered.actor = uuid.New()
	if _, _, err := r.loadPortablePreparation(ctx, request, &altered); err == nil {
		t.Fatal("replay accepted different actor")
	}
	if err := manager.RecordExecutionFailure(ctx, request.Execution, false); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.loadPortablePreparation(ctx, request, p); err == nil {
		t.Fatal("terminal cohort resumed")
	}
	var health string
	if err := pool.QueryRow(ctx, `SELECT health FROM reference_pack_candidates WHERE pack_key=$1 AND pack_version=$2`, vector.Reference.Key, vector.Reference.Version).Scan(&health); err != nil || health != "failed" {
		t.Fatal("parent abort left staged candidate", health, err)
	}
	if err := validateHistoricalAttempts(ctx, pool); err != nil {
		t.Fatal("aborted admission history", err)
	}
	t.Run("required rejection retains verdict without publishing catalog", func(t *testing.T) {
		originalBody, originalContent := body, content
		defer func() { body, content = originalBody, originalContent }()
		invalid := &engineStorage{container: body}
		rewriteEngineContainer(t, invalid, func(files map[string][]byte) {
			var targets map[string]any
			if err := json.Unmarshal(files["metadata/targets.json"], &targets); err != nil {
				t.Fatal(err)
			}
			signature := targets["signatures"].([]any)[0].(map[string]any)
			text := signature["sig"].(string)
			replacement := "0"
			if text[0] == '0' {
				replacement = "1"
			}
			signature["sig"] = replacement + text[1:]
			encoded, err := canonicaljson.Marshal(targets)
			if err != nil {
				t.Fatal(err)
			}
			files["metadata/targets.json"] = encoded
		})
		body = invalid.container
		content.Containers = []packformat.PortableContainer{{ManifestSHA256: vector.Descriptor.ManifestSHA256, ContainerSHA256: packformat.Digest(body), SizeBytes: int64(len(body))}}
		content.RequiredMembers = []packformat.PortableRequiredMember{{SetID: refs.Sets[0].ID, Key: vector.Reference.Key}}
		request, _, _ := newRequest()
		encoded, err := packformat.EncodePortableContent(content, refs)
		if err != nil {
			t.Fatal(err)
		}
		request.ContentManifest = &encoded
		path, err := packformat.PortableContainerPath(vector.Descriptor.ManifestSHA256)
		if err != nil {
			t.Fatal(err)
		}
		request.EmbeddedPaths = []string{path}
		scope, err := r.BeginImportExecution(ctx, request.Execution, request.ExecutionStarted)
		if err != nil {
			t.Fatal(err)
		}
		defer scope.Close()
		request.ExecutionScope = scope
		_, err = r.PrepareImport(scope.Context(), request)
		var rejected *IncidentBundleReferenceValidationError
		if !errors.As(err, &rejected) || rejected.InvariantID != IncidentBundleReferenceDegradationInvariant {
			t.Fatal("required content rejection", err)
		}
		if outcome, _, _ := scope.ClassifyAbsent(ctx, err); outcome != "content_rejected" {
			t.Fatal("content classified as abort", outcome)
		}
		tx, err := pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if err := transactions.ValidateExecutionTx(ctx, tx, request.Execution); err != nil {
			t.Fatal(err)
		}
		if err := scope.AbortTx(ctx, tx); err != nil {
			t.Fatal(err)
		}
		if _, err := transactions.CompleteFailedTx(ctx, tx, request.Execution, jobs.FailureCompletion{Progress: jobs.Progress{Completed: 0, Total: intPtr(1)}, ErrorSummary: jobs.ErrorSummary{Code: "incident_bundle_import_rejected", Message: "Rejected"}}, time.Now()); err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatal(err)
		}
		var outcome string
		var versions, catalogs int
		if err := pool.QueryRow(ctx, `SELECT a.outcome,(SELECT count(*) FROM reference_pack_versions WHERE pack_key=$2 AND pack_version=$3),(SELECT count(*) FROM reference_pack_portable_catalogs WHERE operation_id=a.operation_id) FROM reference_pack_attempts a WHERE a.attempt_id=$1`, scope.requiredRejection.attempt.ID, vector.Reference.Key, vector.Reference.Version).Scan(&outcome, &versions, &catalogs); err != nil || outcome != "content_rejected" || versions != 0 || catalogs != 0 {
			t.Fatal("required rejection leaked success", outcome, versions, catalogs, err)
		}
		if err := validateHistoricalAttempts(ctx, pool); err != nil {
			t.Fatal("rejected attempt history", err)
		}
		if err := validatePortablePreparations(ctx, pool); err != nil {
			t.Fatal("rejected preparation history", err)
		}
		tx, err = pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `ALTER TABLE reference_pack_operations DISABLE TRIGGER USER`); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE reference_pack_operations SET final_outcome=convert_to('{"failed_count":0,"member_count":1,"outcome":"succeeded","schema_id":"cartulary.reference_pack_attempt_result.v1"}','UTF8') WHERE operation_id=$1`, scope.operation); err != nil {
			t.Fatal(err)
		}
		if err := validatePortablePreparations(ctx, tx); !errors.Is(err, errHistoricalIntegrity) {
			t.Fatal("success without catalog restored", err)
		}
	})
	request, _, _ = newRequest()
	contentBytes, err := packformat.EncodePortableContent(content, refs)
	if err != nil {
		t.Fatal(err)
	}
	request.ContentManifest = &contentBytes
	path, err := packformat.PortableContainerPath(vector.Descriptor.ManifestSHA256)
	if err != nil {
		t.Fatal(err)
	}
	request.EmbeddedPaths = []string{path}
	scope, err := r.BeginImportExecution(ctx, request.Execution, request.ExecutionStarted)
	if err != nil {
		t.Fatal(err)
	}
	defer scope.Close()
	request.ExecutionScope = scope
	p, err = r.PrepareImport(scope.Context(), request)
	if err != nil || p.cohort == nil || p.attempt == nil {
		t.Fatal("live destination admission and verification", err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err := transactions.ValidateExecutionTx(ctx, tx, request.Execution); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Portable destination fixture','active',$3,$3)`, p.incident, p.incident.String(), actor); err != nil {
		t.Fatal(err)
	}
	if err := r.ApplyImportTx(ctx, tx, p); err != nil {
		t.Fatal("parent publication", err)
	}
	if _, err := transactions.CompleteSucceededTx(ctx, tx, request.Execution, jobs.SuccessCompletion{Progress: jobs.Progress{Completed: 1, Total: intPtr(1)}, ResultSummary: jobs.ResultSummary{Code: "incident_bundle_imported", Message: "Imported"}}, time.Now()); err != nil {
		t.Fatal("parent terminal result", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	if err := validateHistoricalAttempts(ctx, pool); err != nil {
		t.Fatal("published attempt history", err)
	}
	if err := validatePortableCatalogs(ctx, pool); err != nil {
		t.Fatal("published destination catalog", err)
	}
	after, err := repository.CurrentSet(ctx)
	if err != nil || after.ID != current.ID {
		t.Fatal("destination publication changed activation", err)
	}
	tx, err = pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	if err := r.ApplyImportTx(ctx, tx, p); err != nil {
		t.Fatal("exact parent replay", err)
	}
	_ = tx.Rollback(ctx)

	if err := validatePortablePreparations(ctx, pool); err != nil {
		t.Fatal("retained preparation integrity", err)
	}
	for _, corruption := range []struct{ name, disable, mutation string }{
		{"closed context", "reference_pack_portable_preparations", `UPDATE reference_pack_portable_preparations SET canonical_context='{"schema_id":"retired"}'::bytea WHERE operation_id=$1`},
		{"canonical source catalog", "reference_pack_portable_preparations", `UPDATE reference_pack_portable_preparations SET canonical_references=canonical_references||decode('0a','hex') WHERE operation_id=$1`},
		{"complete selection inventory", "reference_pack_portable_selections", `DELETE FROM reference_pack_portable_selections WHERE operation_id=$1 AND ordinal=1`},
		{"exact transport descriptor", "reference_pack_portable_selections", `UPDATE reference_pack_portable_selections SET expected_container_bytes=expected_container_bytes+1 WHERE operation_id=$1 AND verification_ordinal IS NOT NULL`},
		{"retained input ownership", "", `DELETE FROM reference_pack_object_refs WHERE object_id IN (SELECT input_object_id FROM reference_pack_portable_selections WHERE operation_id=$1 AND input_object_id IS NOT NULL)`},
	} {
		t.Run(corruption.name, func(t *testing.T) {
			tx, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback(ctx)
			if corruption.disable != "" {
				if _, err := tx.Exec(ctx, "ALTER TABLE "+corruption.disable+" DISABLE TRIGGER USER"); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := tx.Exec(ctx, corruption.mutation, p.cohort.operation); err != nil {
				t.Fatal(err)
			}
			if err := validatePortablePreparations(ctx, tx); !errors.Is(err, errHistoricalIntegrity) {
				t.Fatal("corrupt preparation restored", err)
			}
		})
	}

}
