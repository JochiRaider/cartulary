package reference_data

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type executionAttempt struct {
	ID, OperationID uuid.UUID
	Actor           *uuid.UUID
	Frozen          frozenOperation
	Start           time.Time
	Count           int64
}
type frozenMember struct {
	Ordinal      int64
	Key, Version string
	Envelope     *successfulEnvelope
}
type preparedObjectRecord struct {
	ID        uuid.UUID `json:"id"`
	Path      string    `json:"path"`
	Digest    string    `json:"sha256"`
	Size      int64     `json:"size"`
	Reference string    `json:"reference"`
}
type preparedRecord struct {
	SchemaID string                 `json:"schema_id"`
	Manifest []byte                 `json:"manifest"`
	Envelope successfulEnvelope     `json:"envelope"`
	Objects  []preparedObjectRecord `json:"objects"`
	IndexID  uuid.UUID              `json:"index_id"`
}

func encodePrepared(p preparedVersion) ([]byte, error) {
	r := preparedRecord{SchemaID: "cartulary.reference_pack_prepared.v1", Manifest: p.Content.ManifestBytes, Envelope: p.Envelope, IndexID: p.IndexID, Objects: []preparedObjectRecord{}}
	for _, o := range p.Objects {
		r.Objects = append(r.Objects, preparedObjectRecord{ID: o.ID, Path: o.Path, Digest: o.Digest, Size: o.Size, Reference: o.Reference.String()})
	}
	slices.SortFunc(r.Objects, func(a, b preparedObjectRecord) int { return strings.Compare(a.Path, b.Path) })
	data, err := canonicaljson.Marshal(r)
	if err != nil {
		return nil, err
	}
	if _, err := decodePrepared(data); err != nil {
		return nil, err
	}
	return data, nil
}
func decodePrepared(data []byte) (preparedVersion, error) {
	var r preparedRecord
	if err := packformat.ValidatePrepared(data); err != nil {
		return preparedVersion{}, err
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return preparedVersion{}, err
	}
	canonical, err := canonicaljson.Marshal(r)
	if err != nil || !bytes.Equal(data, canonical) || r.SchemaID != "cartulary.reference_pack_prepared.v1" || r.IndexID == uuid.Nil || r.Envelope.TrustProposal == nil {
		return preparedVersion{}, errors.New("reference pack: corrupt prepared result")
	}
	envelope, err := canonicaljson.Marshal(r.Envelope)
	if err != nil {
		return preparedVersion{}, err
	}
	if r.Envelope, err = decodeSuccessfulEnvelope(envelope); err != nil {
		return preparedVersion{}, err
	}
	m, err := packformat.DecodeManifest(r.Manifest, true)
	if err != nil {
		return preparedVersion{}, err
	}
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil {
		return preparedVersion{}, err
	}
	if packformat.Digest(r.Manifest) != r.Envelope.ManifestSHA256 || payload != r.Envelope.PayloadSHA256 || m.Key != r.Envelope.PackKey || m.Version != r.Envelope.PackVersion {
		return preparedVersion{}, errors.New("reference pack: prepared identity mismatch")
	}
	p := preparedVersion{IndexID: r.IndexID, Envelope: r.Envelope, Content: &verifiedContent{Manifest: m, ManifestBytes: r.Manifest, ManifestSHA256: r.Envelope.ManifestSHA256, PayloadSHA256: payload, VerifiedAt: r.Envelope.VerifiedAt, Trust: *r.Envelope.TrustProposal}}
	for _, o := range r.Objects {
		ref, err := ParseStorageRef(o.Reference)
		if err != nil {
			return p, err
		}
		p.Objects = append(p.Objects, preparedObject{ID: o.ID, Path: o.Path, Digest: o.Digest, Size: o.Size, Reference: ref})
	}
	return p, nil
}

func (c *Coordinator) beginAttempt(ctx context.Context, execution jobs.Execution, operationID uuid.UUID, start time.Time) (executionAttempt, error) {
	tx, err := c.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return executionAttempt{}, err
	}
	defer tx.Rollback(ctx)
	if _, err := c.executionGuard.ExtensionCancellationContextTx(ctx, tx, execution); err != nil {
		return executionAttempt{}, err
	}
	a, err := beginAttemptTx(ctx, tx, operationID, start)
	if err != nil {
		return a, err
	}
	return a, tx.Commit(ctx)
}

func beginAttemptTx(ctx context.Context, tx pgx.Tx, operationID uuid.UUID, start time.Time) (executionAttempt, error) {
	a := executionAttempt{ID: uuid.New(), OperationID: operationID, Start: start.UTC()}
	var encoded []byte
	var terminal *time.Time
	if err := tx.QueryRow(ctx, `SELECT actor_user_id,frozen_input,terminal_at FROM reference_pack_operations WHERE operation_id=$1 FOR UPDATE`, operationID).Scan(&a.Actor, &encoded, &terminal); err != nil {
		return a, err
	}
	if terminal != nil {
		return a, errors.New("reference pack: terminal operation cannot start another attempt")
	}
	frozen, err := decodeFrozenOperation(encoded)
	if err != nil {
		return a, err
	}
	a.Frozen = frozen
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operation_members WHERE operation_id=$1`, operationID).Scan(&a.Count); err != nil {
		return a, err
	}
	if a.Frozen.Kind == "import" && a.Count == 0 {
		a.Count = 1
	}
	return a, retainAttemptStartTx(ctx, tx, a)
}

func retainAttemptStartTx(ctx context.Context, tx pgx.Tx, a executionAttempt) error {
	// The Jobs execution lease authorizes this new attempt. Prior incomplete
	// preparation is evidence of interruption, never a retained content verdict.
	interrupted, err := packformat.EncodeAttemptResult("interrupted", 0, 0)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_attempts SET completed_at=$2,outcome='interrupted',canonical_result=$3 WHERE operation_id=$1 AND completed_at IS NULL`, a.OperationID, a.Start, interrupted); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `INSERT INTO reference_pack_attempts(attempt_id,operation_id,started_at) VALUES($1,$2,$3)`, a.ID, a.OperationID, a.Start); err != nil {
		return err
	}
	return nil
}

func (c *verificationService) prepareMember(ctx context.Context, a executionAttempt, m frozenMember) error {
	identity := operationVerificationIdentity{referenceDependencies: c.referenceDependencies, operationID: a.OperationID}
	if a.Frozen.Kind != "import" {
		identity.retained = &m
	}
	input := verificationAttempt{Observer: c.observer, Start: a.Start, ClockTrusted: a.Frozen.ClockTrusted, Limits: c.limits.verificationArchiveLimits(), Identity: identity}
	input.ResolveTrust = func(ctx context.Context, id string, versions []int64) (packformat.TrustSnapshot, bool, error) {
		return c.frozenTrust(ctx, a, m, id, versions)
	}
	input.IdentityAdmitted = func(_ context.Context, key, version string) error {
		if key != m.Key || version != m.Version {
			return &OperationRejection{Reason: "stale_admission_state"}
		}
		return nil
	}
	if a.Frozen.Kind == "import" {
		ref, err := ParseStorageRef(*a.Frozen.ContainerReference)
		if err != nil {
			return err
		}
		input.Retained = &ref
		input.ContainerSHA256 = *a.Frozen.ContainerSHA256
	} else {
		if m.Envelope == nil || m.Envelope.ContainerRef == nil || m.Envelope.ContainerSHA256 == nil {
			return errors.New("reference pack: incomplete frozen envelope")
		}
		ref, err := ParseStorageRef(*m.Envelope.ContainerRef)
		if err != nil {
			return err
		}
		input.Retained = &ref
		input.ContainerSHA256 = *m.Envelope.ContainerSHA256
	}
	return c.prepareVerificationMember(ctx, a, m, input)
}

// The coordinator owns the one verification/index/artifact preparation path.
// Lifecycle Jobs and destination cohorts supply their frozen input bindings;
// neither transport can publish these private results.
func (c *verificationService) prepareVerificationMember(ctx context.Context, a executionAttempt, m frozenMember, input verificationAttempt) error {
	var index *indexBuilder
	content, err := verifyCanonicalContainer(ctx, c.storage, input, func(ctx context.Context, manifest packformat.Manifest, manifestSHA, payloadSHA string) (packformat.ContentSink, error) {
		var err error
		index, err = newIndexBuilder(ctx, c.pool, a.OperationID, manifest, manifestSHA, payloadSHA)
		return index, err
	})
	if err != nil {
		return c.retainMemberFailure(ctx, a, m, err)
	}
	defer content.Close()
	indexID, err := index.Complete(ctx)
	if err != nil {
		return err
	}
	prepared, err := prepareVerifiedObjects(ctx, c.pool, c.storage, a.OperationID, a.ID, input, content, indexID)
	if err != nil {
		return err
	}
	if err := content.Close(); err != nil {
		return err
	}
	content.workspace = nil
	data, err := encodePrepared(prepared)
	if err != nil {
		return err
	}
	_, err = c.pool.Exec(ctx, `INSERT INTO reference_pack_attempt_members(attempt_id,ordinal,pack_key,pack_version,verdict,canonical_prepared) VALUES($1,$2,$3,$4,'succeeded',$5)`, a.ID, m.Ordinal, m.Key, m.Version, data)
	return err
}

func (c *verificationService) retainMemberFailure(ctx context.Context, a executionAttempt, m frozenMember, err error) error {
	var rejection *ContentRejection
	if !errors.As(err, &rejection) {
		return err
	}
	summary := rejection.Summary
	if summary == nil {
		rawSummary, err := packformat.CheckSummary(ctx, rejection.CheckID, nil, func(_ context.Context, emit packformat.FindingSink) error { return emit(packformat.Finding{Path: "$"}) })
		if err != nil {
			return err
		}
		summary = summaryFromFormat(rawSummary)
	}
	encoded, err := canonicaljson.Marshal(summary)
	if err != nil {
		return err
	}
	validated, err := packformat.DecodeValidationSummary(encoded)
	if err != nil {
		return err
	}
	if validated.Result != "failed" || validated.Issues[0].CheckID != rejection.CheckID || validated.Issues[0].Code != rejection.Code {
		return errors.New("reference pack: inconsistent validation result")
	}
	var key, version *string
	if m.Key != "" {
		key = &m.Key
		version = &m.Version
	}
	_, err = c.pool.Exec(ctx, `INSERT INTO reference_pack_attempt_members(attempt_id,ordinal,pack_key,pack_version,verdict,failure_code,check_id,canonical_validation_summary) VALUES($1,$2,$3,$4,'content_rejected',$5,$6,$7)`, a.ID, m.Ordinal, key, version, rejection.Code, rejection.CheckID, encoded)
	return err
}
