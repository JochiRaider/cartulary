package imports

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type unitApplyOutcome struct {
	ImportSessionID    uuid.UUID
	ImportUnitID       uuid.UUID
	ApplyJobID         uuid.UUID
	DiscoverySequence  int
	UnitCommitID       string
	Status             string
	ActorUserID        uuid.UUID
	TargetKind         string
	TargetViewSchemaID *string
	ExtensionProfileID *string
	OwnerBindingID     string
	SourceDigest       string
	MappingFingerprint string
	OwnerResult        any
	ResourceRefs       []jobs.ResourceRef
	ChangeSetID        *uuid.UUID
	ErrorCode          *string
	ReasonCode         *string
	CommittedAt        time.Time
}

type appliedUnitCommit struct {
	OwnerResult  any
	ResourceRefs []jobs.ResourceRef
	ChangeSetID  *uuid.UUID
}

type applyFinalization struct {
	SessionStatus string
	JobStatus     string
	ResultCode    string
	ErrorCode     string
	ResourceRefs  []jobs.ResourceRef
	Applied       int
	Failed        int
	Canceled      int
	OutcomeDigest string
}

type unitOutcomeQuerier interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func unitCommitID(sessionID uuid.UUID, unitID uuid.UUID) string {
	return "import-unit:" + sessionID.String() + ":" + unitID.String()
}

func (s *Store) getUnitOutcome(
	ctx context.Context,
	sessionID uuid.UUID,
	unitID uuid.UUID,
) (unitApplyOutcome, error) {
	return findUnitOutcome(ctx, s.pool, sessionID, unitID)
}

func findUnitOutcome(
	ctx context.Context,
	querier unitOutcomeQuerier,
	sessionID uuid.UUID,
	unitID uuid.UUID,
) (unitApplyOutcome, error) {
	return scanUnitOutcome(querier.QueryRow(ctx, `
SELECT import_session_id, import_unit_id, apply_job_id, discovery_sequence,
       unit_commit_id, outcome_status, actor_user_id, target_kind,
       target_view_schema_id, extension_profile_id, owner_binding_id,
       source_content_sha256, mapping_fingerprint, owner_result_json,
       resource_refs_json, change_set_id, error_code, reason_code, committed_at
  FROM import_unit_apply_outcomes
 WHERE import_session_id = $1
   AND import_unit_id = $2
`, sessionID, unitID))
}

func scanUnitOutcome(row pgx.Row) (unitApplyOutcome, error) {
	var outcome unitApplyOutcome
	var ownerResultJSON []byte
	var resourceRefsJSON []byte
	err := row.Scan(
		&outcome.ImportSessionID,
		&outcome.ImportUnitID,
		&outcome.ApplyJobID,
		&outcome.DiscoverySequence,
		&outcome.UnitCommitID,
		&outcome.Status,
		&outcome.ActorUserID,
		&outcome.TargetKind,
		&outcome.TargetViewSchemaID,
		&outcome.ExtensionProfileID,
		&outcome.OwnerBindingID,
		&outcome.SourceDigest,
		&outcome.MappingFingerprint,
		&ownerResultJSON,
		&resourceRefsJSON,
		&outcome.ChangeSetID,
		&outcome.ErrorCode,
		&outcome.ReasonCode,
		&outcome.CommittedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return unitApplyOutcome{}, ErrNotFound
	}
	if err != nil {
		return unitApplyOutcome{}, err
	}
	if err := json.Unmarshal(ownerResultJSON, &outcome.OwnerResult); err != nil {
		return unitApplyOutcome{}, err
	}
	if err := json.Unmarshal(resourceRefsJSON, &outcome.ResourceRefs); err != nil {
		return unitApplyOutcome{}, err
	}
	return outcome, nil
}

func (s *Store) insertApplyUnitPlansTx(
	ctx context.Context,
	tx pgx.Tx,
	sessionID uuid.UUID,
	jobID uuid.UUID,
	selected []uuid.UUID,
	admittedAt time.Time,
) error {
	type applyUnitPlan struct {
		unitID             uuid.UUID
		discoverySequence  int
		sourceRowsJSON     []byte
		mappingJSON        []byte
		mappingFingerprint string
		sourceFileKind     string
		sourceDigest       string
		sourceStreamRef    string
		parserProfileID    string
		parserVersion      string
		locatorKind        string
		locator            string
		sourceRectA1       string
		targetKind         string
		targetViewSchemaID *string
		extensionProfileID *string
		ownerBindingID     string
	}
	rows, err := tx.Query(ctx, `
SELECT u.import_unit_id,
       u.discovery_sequence,
       u.source_rows_json,
       u.approved_mapping_json,
       u.mapping_fingerprint,
       s.source_file_kind,
       s.source_content_sha256,
       COALESCE(u.source_stream_ref, ''),
       s.parser_profile_id,
       s.parser_version,
       u.locator_kind,
       u.locator,
       u.source_rect_a1
  FROM import_sessions s
  JOIN import_units u
    ON u.import_session_id = s.import_session_id
 WHERE s.import_session_id = $1
   AND u.import_unit_id = ANY($2)
   AND u.unit_status = 'ready'
   AND u.mapping_fingerprint IS NOT NULL
   AND u.approved_mapping_json IS NOT NULL
 ORDER BY u.discovery_sequence ASC, u.import_unit_id ASC
 FOR UPDATE OF s, u
`, sessionID, selected)
	if err != nil {
		return err
	}
	plans := make([]applyUnitPlan, 0, len(selected))
	for rows.Next() {
		var plan applyUnitPlan
		if err := rows.Scan(
			&plan.unitID,
			&plan.discoverySequence,
			&plan.sourceRowsJSON,
			&plan.mappingJSON,
			&plan.mappingFingerprint,
			&plan.sourceFileKind,
			&plan.sourceDigest,
			&plan.sourceStreamRef,
			&plan.parserProfileID,
			&plan.parserVersion,
			&plan.locatorKind,
			&plan.locator,
			&plan.sourceRectA1,
		); err != nil {
			rows.Close()
			return err
		}
		var mapping ApprovedMapping
		if err := json.Unmarshal(plan.mappingJSON, &mapping); err != nil {
			rows.Close()
			return err
		}
		target, ok := lookupApprovedImportTarget(mapping)
		if !ok || !target.readyCheckImportable() {
			rows.Close()
			return importApplyBlockedError("target_kind_not_importable")
		}
		plan.targetKind = target.TargetKind
		plan.ownerBindingID = target.FacadeBindingID
		if target.TargetKind == ImportTargetKindViewSchema {
			value := target.ViewSchemaID
			plan.targetViewSchemaID = &value
		} else {
			value := target.ExtensionProfileID
			plan.extensionProfileID = &value
		}
		if plan.ownerBindingID == "" {
			rows.Close()
			if target.TargetKind == ImportTargetKindViewSchema {
				return importApplyBlockedError("owner_create_contract_unavailable")
			}
			return importApplyBlockedError("owner_apply_contract_unavailable")
		}
		plans = append(plans, plan)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	if len(plans) != len(selected) {
		return importApplyBlockedError("unit_not_ready")
	}
	for _, plan := range plans {
		if _, err := tx.Exec(ctx, `
INSERT INTO import_apply_unit_plans (
    import_session_id, import_unit_id, apply_job_id, discovery_sequence,
    source_file_kind, source_content_sha256, source_stream_ref,
    source_rows_sha256, parser_profile_id, parser_version, locator_kind,
    locator, source_rect_a1, mapping_fingerprint, approved_mapping_json,
    approved_mapping_sha256, target_kind, target_view_schema_id,
    extension_profile_id, owner_binding_id, target_registry_sha256,
    admitted_at
) VALUES (
	    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11,
	    $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22
)
`, sessionID, plan.unitID, jobID, plan.discoverySequence, plan.sourceFileKind,
			plan.sourceDigest, plan.sourceStreamRef, sha256Hex(plan.sourceRowsJSON),
			plan.parserProfileID, plan.parserVersion, plan.locatorKind, plan.locator,
			plan.sourceRectA1, plan.mappingFingerprint, plan.mappingJSON,
			sha256Hex(plan.mappingJSON), plan.targetKind, plan.targetViewSchemaID,
			plan.extensionProfileID, plan.ownerBindingID, importtargetregistry.RegistrySHA256,
			admittedAt.UTC()); err != nil {
			return err
		}
	}
	return nil
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func (s *Store) lockApplyUnitTx(
	ctx context.Context,
	tx pgx.Tx,
	start ApplyStartResult,
	unitID uuid.UUID,
) (ApplyUnitData, error) {
	var unit ApplyUnitData
	var sourceRowsJSON []byte
	var mappingJSON []byte
	var admittedSourceFileKind string
	var admittedSourceDigest string
	var admittedSourceStreamRef string
	var admittedSourceRowsDigest string
	var admittedParserProfileID string
	var admittedParserVersion string
	var admittedLocatorKind string
	var admittedLocator string
	var admittedSourceRectA1 string
	var admittedMappingFingerprint string
	var admittedMappingDigest string
	var admittedTargetRegistryDigest string
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return ApplyUnitData{}, err
	}
	err = tx.QueryRow(ctx, `
SELECT u.import_unit_id,
       u.discovery_sequence,
       u.source_rows_json,
       u.approved_mapping_json,
       COALESCE(u.mapping_fingerprint, ''),
       s.source_file_kind,
       s.source_content_sha256,
       COALESCE(u.source_stream_ref, ''),
       s.parser_profile_id,
       s.parser_version,
       u.locator_kind,
       u.locator,
       u.source_rect_a1,
       p.source_file_kind,
       p.source_content_sha256,
       p.source_stream_ref,
       p.source_rows_sha256,
       p.parser_profile_id,
       p.parser_version,
       p.locator_kind,
       p.locator,
       p.source_rect_a1,
       p.mapping_fingerprint,
       p.approved_mapping_sha256,
       p.target_registry_sha256
  FROM import_sessions s
  JOIN import_units u
    ON u.import_session_id = s.import_session_id
  JOIN import_apply_unit_plans p
    ON p.import_session_id = s.import_session_id
   AND p.import_unit_id = u.import_unit_id
   AND p.apply_job_id = s.apply_job_id
 WHERE s.import_session_id = $1
   AND s.incident_id = $2
   AND s.apply_job_id = $3
   AND s.session_status = 'applying'
   AND u.import_unit_id = $4
   AND u.unit_status = 'applying'
   AND u.import_unit_id = ANY(s.selected_unit_ids)
 FOR UPDATE OF s, u, p
`, start.ImportSessionID, start.IncidentID, jobID, unitID).Scan(
		&unit.UnitID,
		&unit.DiscoverySequence,
		&sourceRowsJSON,
		&mappingJSON,
		&unit.MappingFingerprint,
		&unit.SourceFileKind,
		&unit.SourceContentSHA256,
		&unit.SourceStreamRef,
		&unit.ParserProfileID,
		&unit.ParserVersion,
		&unit.LocatorKind,
		&unit.Locator,
		&unit.SourceRectA1,
		&admittedSourceFileKind,
		&admittedSourceDigest,
		&admittedSourceStreamRef,
		&admittedSourceRowsDigest,
		&admittedParserProfileID,
		&admittedParserVersion,
		&admittedLocatorKind,
		&admittedLocator,
		&admittedSourceRectA1,
		&admittedMappingFingerprint,
		&admittedMappingDigest,
		&admittedTargetRegistryDigest,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ApplyUnitData{}, importApplyBlockedError("unit_not_ready")
	}
	if err != nil {
		return ApplyUnitData{}, err
	}
	if unit.SourceFileKind != admittedSourceFileKind ||
		unit.SourceContentSHA256 != admittedSourceDigest ||
		unit.SourceStreamRef != admittedSourceStreamRef ||
		sha256Hex(sourceRowsJSON) != admittedSourceRowsDigest ||
		unit.ParserProfileID != admittedParserProfileID ||
		unit.ParserVersion != admittedParserVersion ||
		unit.LocatorKind != admittedLocatorKind ||
		unit.Locator != admittedLocator ||
		unit.SourceRectA1 != admittedSourceRectA1 {
		return ApplyUnitData{}, importApplyBlockedError("source_changed")
	}
	if unit.MappingFingerprint != admittedMappingFingerprint ||
		sha256Hex(mappingJSON) != admittedMappingDigest {
		return ApplyUnitData{}, importApplyBlockedError("source_changed")
	}
	if admittedTargetRegistryDigest != importtargetregistry.RegistrySHA256 {
		return ApplyUnitData{}, importApplyBlockedError("target_kind_not_importable")
	}
	if err := json.Unmarshal(sourceRowsJSON, &unit.SourceRows); err != nil {
		return ApplyUnitData{}, err
	}
	if err := json.Unmarshal(mappingJSON, &unit.ApprovedMapping); err != nil {
		return ApplyUnitData{}, err
	}
	return unit, nil
}

func (s *Store) ensureApplyJobRunnableTx(
	ctx context.Context,
	tx pgx.Tx,
	start ApplyStartResult,
) error {
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return err
	}
	var status string
	if err := tx.QueryRow(ctx, `
SELECT status
  FROM jobs
 WHERE job_id = $1
 FOR UPDATE
`, jobID).Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return jobs.ErrNotFound
		}
		return err
	}
	switch status {
	case jobs.StatusRunning:
		return nil
	case jobs.StatusCancelRequested:
		return errImportUnitCanceled
	default:
		return fmt.Errorf("import apply job is not runnable in status %q", status)
	}
}

func (s *Store) lockUnitOutcomePlanTx(
	ctx context.Context,
	tx pgx.Tx,
	start ApplyStartResult,
	unitID uuid.UUID,
) (ApplyUnitData, importTarget, error) {
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return ApplyUnitData{}, importTarget{}, err
	}
	var unit ApplyUnitData
	var mappingJSON []byte
	var targetKind string
	var targetViewSchemaID *string
	var extensionProfileID *string
	var ownerBindingID string
	err = tx.QueryRow(ctx, `
SELECT p.import_unit_id,
       p.discovery_sequence,
       p.approved_mapping_json,
       p.mapping_fingerprint,
       p.source_file_kind,
       p.source_content_sha256,
       p.source_stream_ref,
       p.parser_profile_id,
       p.parser_version,
       p.locator_kind,
       p.locator,
       p.source_rect_a1,
       p.target_kind,
       p.target_view_schema_id,
       p.extension_profile_id,
       p.owner_binding_id
  FROM import_sessions s
  JOIN import_units u
    ON u.import_session_id = s.import_session_id
  JOIN import_apply_unit_plans p
    ON p.import_session_id = s.import_session_id
   AND p.import_unit_id = u.import_unit_id
   AND p.apply_job_id = s.apply_job_id
 WHERE s.import_session_id = $1
   AND s.incident_id = $2
   AND s.apply_job_id = $3
   AND s.session_status = 'applying'
   AND u.import_unit_id = $4
   AND u.unit_status = 'applying'
   AND u.import_unit_id = ANY(s.selected_unit_ids)
 FOR UPDATE OF s, u, p
`, start.ImportSessionID, start.IncidentID, jobID, unitID).Scan(
		&unit.UnitID,
		&unit.DiscoverySequence,
		&mappingJSON,
		&unit.MappingFingerprint,
		&unit.SourceFileKind,
		&unit.SourceContentSHA256,
		&unit.SourceStreamRef,
		&unit.ParserProfileID,
		&unit.ParserVersion,
		&unit.LocatorKind,
		&unit.Locator,
		&unit.SourceRectA1,
		&targetKind,
		&targetViewSchemaID,
		&extensionProfileID,
		&ownerBindingID,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return ApplyUnitData{}, importTarget{}, importApplyBlockedError("unit_not_ready")
	}
	if err != nil {
		return ApplyUnitData{}, importTarget{}, err
	}
	if err := json.Unmarshal(mappingJSON, &unit.ApprovedMapping); err != nil {
		return ApplyUnitData{}, importTarget{}, err
	}
	target := importTarget{TargetKind: targetKind}
	switch targetKind {
	case ImportTargetKindViewSchema:
		if targetViewSchemaID == nil {
			return ApplyUnitData{}, importTarget{}, fmt.Errorf("view import plan has no target schema")
		}
		target.ViewSchemaID = *targetViewSchemaID
		target.FacadeBindingID = ownerBindingID
	case ImportTargetKindNetworkFlowTable:
		if extensionProfileID == nil {
			return ApplyUnitData{}, importTarget{}, fmt.Errorf("analytical import plan has no profile")
		}
		target.ExtensionProfileID = *extensionProfileID
		target.FacadeBindingID = ownerBindingID
	default:
		return ApplyUnitData{}, importTarget{}, fmt.Errorf(
			"import plan has unsupported target kind %q",
			targetKind,
		)
	}
	return unit, target, nil
}

func (s *Store) insertAppliedUnitOutcomeTx(
	ctx context.Context,
	tx pgx.Tx,
	start ApplyStartResult,
	unit ApplyUnitData,
	target importTarget,
	actorUserID uuid.UUID,
	commit appliedUnitCommit,
	now time.Time,
) (unitApplyOutcome, error) {
	outcome, err := buildUnitOutcome(start, unit, target, actorUserID, "applied", commit, "", "", now)
	if err != nil {
		return unitApplyOutcome{}, err
	}
	if err := insertUnitOutcomeTx(ctx, tx, outcome); err != nil {
		return unitApplyOutcome{}, err
	}
	tag, err := tx.Exec(ctx, `
UPDATE import_units
   SET unit_status = 'applied',
       updated_at = $3
 WHERE import_session_id = $1
   AND import_unit_id = $2
   AND unit_status = 'applying'
`, start.ImportSessionID, unit.UnitID, now.UTC())
	if err != nil {
		return unitApplyOutcome{}, err
	}
	if tag.RowsAffected() != 1 {
		return unitApplyOutcome{}, fmt.Errorf("import unit applied outcome state precondition failed")
	}
	return outcome, nil
}

func (s *Store) recordTerminalUnitOutcome(
	ctx context.Context,
	start ApplyStartResult,
	unitID uuid.UUID,
	actorUserID uuid.UUID,
	status string,
	errorCode string,
	reasonCode string,
	now time.Time,
) (unitApplyOutcome, error) {
	if status != "failed" && status != "canceled" {
		return unitApplyOutcome{}, fmt.Errorf("unsupported import unit outcome status %q", status)
	}
	if existing, err := s.getUnitOutcome(ctx, start.ImportSessionID, unitID); err == nil {
		return existing, nil
	} else if !errors.Is(err, ErrNotFound) {
		return unitApplyOutcome{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return unitApplyOutcome{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if existing, findErr := findUnitOutcome(ctx, tx, start.ImportSessionID, unitID); findErr == nil {
		if err := tx.Commit(ctx); err != nil {
			return unitApplyOutcome{}, err
		}
		return existing, nil
	} else if !errors.Is(findErr, ErrNotFound) {
		return unitApplyOutcome{}, findErr
	}
	unit, target, err := s.lockUnitOutcomePlanTx(ctx, tx, start, unitID)
	if err != nil {
		if existing, findErr := findUnitOutcome(ctx, tx, start.ImportSessionID, unitID); findErr == nil {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				return unitApplyOutcome{}, commitErr
			}
			return existing, nil
		}
		return unitApplyOutcome{}, err
	}
	outcome, err := buildUnitOutcome(
		start,
		unit,
		target,
		actorUserID,
		status,
		appliedUnitCommit{OwnerResult: map[string]any{}, ResourceRefs: []jobs.ResourceRef{}},
		errorCode,
		reasonCode,
		now,
	)
	if err != nil {
		return unitApplyOutcome{}, err
	}
	if err := insertUnitOutcomeTx(ctx, tx, outcome); err != nil {
		return unitApplyOutcome{}, err
	}
	tag, err := tx.Exec(ctx, `
UPDATE import_units
   SET unit_status = $3,
       updated_at = $4
 WHERE import_session_id = $1
   AND import_unit_id = $2
   AND unit_status = 'applying'
`, start.ImportSessionID, unitID, status, now.UTC())
	if err != nil {
		return unitApplyOutcome{}, err
	}
	if tag.RowsAffected() != 1 {
		return unitApplyOutcome{}, fmt.Errorf("import unit terminal outcome state precondition failed")
	}
	if err := tx.Commit(ctx); err != nil {
		return unitApplyOutcome{}, err
	}
	return outcome, nil
}

func buildUnitOutcome(
	start ApplyStartResult,
	unit ApplyUnitData,
	target importTarget,
	actorUserID uuid.UUID,
	status string,
	commit appliedUnitCommit,
	errorCode string,
	reasonCode string,
	now time.Time,
) (unitApplyOutcome, error) {
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return unitApplyOutcome{}, err
	}
	outcome := unitApplyOutcome{
		ImportSessionID:    start.ImportSessionID,
		ImportUnitID:       unit.UnitID,
		ApplyJobID:         jobID,
		DiscoverySequence:  unit.DiscoverySequence,
		UnitCommitID:       unitCommitID(start.ImportSessionID, unit.UnitID),
		Status:             status,
		ActorUserID:        actorUserID,
		TargetKind:         target.TargetKind,
		OwnerBindingID:     target.FacadeBindingID,
		SourceDigest:       unit.SourceContentSHA256,
		MappingFingerprint: unit.MappingFingerprint,
		OwnerResult:        commit.OwnerResult,
		ResourceRefs:       append([]jobs.ResourceRef{}, commit.ResourceRefs...),
		ChangeSetID:        commit.ChangeSetID,
		CommittedAt:        now.UTC(),
	}
	if target.TargetKind == ImportTargetKindViewSchema {
		value := target.ViewSchemaID
		outcome.TargetViewSchemaID = &value
	} else {
		value := target.ExtensionProfileID
		outcome.ExtensionProfileID = &value
	}
	if errorCode != "" {
		outcome.ErrorCode = &errorCode
	}
	if reasonCode != "" {
		outcome.ReasonCode = &reasonCode
	}
	return outcome, nil
}

func insertUnitOutcomeTx(ctx context.Context, tx pgx.Tx, outcome unitApplyOutcome) error {
	ownerResult, err := json.Marshal(outcome.OwnerResult)
	if err != nil {
		return err
	}
	resourceRefs, err := json.Marshal(outcome.ResourceRefs)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO import_unit_apply_outcomes (
    import_session_id, import_unit_id, apply_job_id, discovery_sequence,
    unit_commit_id, outcome_status, actor_user_id, target_kind,
    target_view_schema_id, extension_profile_id, owner_binding_id,
    source_content_sha256, mapping_fingerprint, owner_result_json,
    resource_refs_json, change_set_id, error_code, reason_code, committed_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19
)
`, outcome.ImportSessionID, outcome.ImportUnitID, outcome.ApplyJobID, outcome.DiscoverySequence,
		outcome.UnitCommitID, outcome.Status, outcome.ActorUserID, outcome.TargetKind,
		outcome.TargetViewSchemaID, outcome.ExtensionProfileID, outcome.OwnerBindingID,
		outcome.SourceDigest, outcome.MappingFingerprint, ownerResult, resourceRefs,
		outcome.ChangeSetID, outcome.ErrorCode, outcome.ReasonCode, outcome.CommittedAt.UTC())
	return err
}

func (s *Store) prepareApplyFinalization(
	ctx context.Context,
	start ApplyStartResult,
) (applyFinalization, error) {
	return deriveApplyFinalization(ctx, s.pool, start)
}

func (s *Store) finalizeApplyFromOutcomesTx(
	ctx context.Context,
	tx pgx.Tx,
	start ApplyStartResult,
	expected applyFinalization,
	now time.Time,
) error {
	actual, err := deriveApplyFinalization(ctx, tx, start)
	if err != nil {
		return err
	}
	if actual.OutcomeDigest != expected.OutcomeDigest ||
		actual.SessionStatus != expected.SessionStatus ||
		actual.JobStatus != expected.JobStatus {
		return fmt.Errorf("import apply outcomes changed before finalization")
	}
	tag, err := tx.Exec(ctx, `
UPDATE import_sessions
   SET session_status = $2,
       updated_at = $3
 WHERE import_session_id = $1
   AND session_status = 'applying'
`, start.ImportSessionID, actual.SessionStatus, now.UTC())
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		var status string
		if scanErr := tx.QueryRow(ctx, `
SELECT session_status
  FROM import_sessions
 WHERE import_session_id = $1
 FOR UPDATE
`, start.ImportSessionID).Scan(&status); scanErr == nil && status == actual.SessionStatus {
			return nil
		}
		return fmt.Errorf("import apply finalization state precondition failed")
	}
	return nil
}

func deriveApplyFinalization(
	ctx context.Context,
	querier unitOutcomeQuerier,
	start ApplyStartResult,
) (applyFinalization, error) {
	jobID, err := uuid.Parse(start.Job.JobID)
	if err != nil {
		return applyFinalization{}, err
	}
	var sessionStatus string
	var applyJobID *uuid.UUID
	var selectedJSON []byte
	if err := querier.QueryRow(ctx, `
SELECT session_status, apply_job_id, to_jsonb(selected_unit_ids)
  FROM import_sessions
 WHERE import_session_id = $1
`, start.ImportSessionID).Scan(&sessionStatus, &applyJobID, &selectedJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return applyFinalization{}, ErrNotFound
		}
		return applyFinalization{}, err
	}
	if applyJobID == nil || *applyJobID != jobID ||
		(sessionStatus != "applying" &&
			sessionStatus != "applied" &&
			sessionStatus != "partially_applied" &&
			sessionStatus != "failed" &&
			sessionStatus != "canceled") {
		return applyFinalization{}, fmt.Errorf("import apply finalization session mismatch")
	}
	var persistedSelected []uuid.UUID
	if err := json.Unmarshal(selectedJSON, &persistedSelected); err != nil {
		return applyFinalization{}, err
	}
	if !uuidSlicesEqual(persistedSelected, start.SelectedUnitIDs) {
		return applyFinalization{}, fmt.Errorf("import apply finalization selection mismatch")
	}
	rows, err := querier.Query(ctx, `
SELECT import_session_id, import_unit_id, apply_job_id, discovery_sequence,
       unit_commit_id, outcome_status, actor_user_id, target_kind,
       target_view_schema_id, extension_profile_id, owner_binding_id,
       source_content_sha256, mapping_fingerprint, owner_result_json,
       resource_refs_json, change_set_id, error_code, reason_code, committed_at
  FROM import_unit_apply_outcomes
 WHERE import_session_id = $1
   AND import_unit_id = ANY($2)
 ORDER BY discovery_sequence ASC, import_unit_id ASC
`, start.ImportSessionID, start.SelectedUnitIDs)
	if err != nil {
		return applyFinalization{}, err
	}
	defer rows.Close()
	outcomes := make([]unitApplyOutcome, 0, len(start.SelectedUnitIDs))
	for rows.Next() {
		outcome, err := scanUnitOutcome(rows)
		if err != nil {
			return applyFinalization{}, err
		}
		if outcome.ApplyJobID != jobID {
			return applyFinalization{}, fmt.Errorf("import unit outcome job mismatch")
		}
		outcomes = append(outcomes, outcome)
	}
	if err := rows.Err(); err != nil {
		return applyFinalization{}, err
	}
	if len(outcomes) != len(start.SelectedUnitIDs) {
		return applyFinalization{}, fmt.Errorf("import unit outcomes incomplete")
	}
	seen := make(map[uuid.UUID]struct{}, len(outcomes))
	finalization := applyFinalization{
		ResourceRefs: []jobs.ResourceRef{{
			Kind:  "import_session",
			ID:    start.ImportSessionID.String(),
			Route: "/api/v1/import-sessions/" + start.ImportSessionID.String(),
		}},
	}
	digestRows := make([]map[string]any, 0, len(outcomes))
	for _, outcome := range outcomes {
		if _, duplicate := seen[outcome.ImportUnitID]; duplicate {
			return applyFinalization{}, fmt.Errorf("duplicate import unit outcome")
		}
		seen[outcome.ImportUnitID] = struct{}{}
		switch outcome.Status {
		case "applied":
			finalization.Applied++
			finalization.ResourceRefs = append(finalization.ResourceRefs, outcome.ResourceRefs...)
		case "failed":
			finalization.Failed++
		case "canceled":
			finalization.Canceled++
		default:
			return applyFinalization{}, fmt.Errorf("unsupported import unit outcome %q", outcome.Status)
		}
		digestRows = append(digestRows, map[string]any{
			"import_unit_id":      outcome.ImportUnitID.String(),
			"unit_commit_id":      outcome.UnitCommitID,
			"outcome_status":      outcome.Status,
			"source_digest":       outcome.SourceDigest,
			"mapping_fingerprint": outcome.MappingFingerprint,
			"resource_refs":       outcome.ResourceRefs,
		})
	}
	switch {
	case finalization.Applied == len(outcomes):
		finalization.SessionStatus = "applied"
		finalization.JobStatus = jobs.StatusSucceeded
		finalization.ResultCode = "import_session_applied"
	case finalization.Applied > 0 && finalization.Canceled > 0:
		finalization.SessionStatus = "partially_applied"
		finalization.JobStatus = jobs.StatusCanceled
		finalization.ResultCode = "import_session_partially_applied"
	case finalization.Applied > 0:
		finalization.SessionStatus = "partially_applied"
		finalization.JobStatus = jobs.StatusSucceeded
		finalization.ResultCode = "import_session_partially_applied"
	case finalization.Canceled > 0:
		finalization.SessionStatus = "canceled"
		finalization.JobStatus = jobs.StatusCanceled
		finalization.ResultCode = "import_session_canceled"
	default:
		finalization.SessionStatus = "failed"
		finalization.JobStatus = jobs.StatusFailed
		finalization.ErrorCode = "import_apply_failed"
	}
	canonical, err := json.Marshal(map[string]any{
		"import_session_id": start.ImportSessionID.String(),
		"apply_job_id":      jobID.String(),
		"selected_unit_ids": uuidStrings(start.SelectedUnitIDs),
		"session_status":    finalization.SessionStatus,
		"job_status":        finalization.JobStatus,
		"outcomes":          digestRows,
	})
	if err != nil {
		return applyFinalization{}, err
	}
	sum := sha256.Sum256(canonical)
	finalization.OutcomeDigest = hex.EncodeToString(sum[:])
	return finalization, nil
}

func uuidSlicesEqual(left []uuid.UUID, right []uuid.UUID) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if !bytes.Equal(left[index][:], right[index][:]) {
			return false
		}
	}
	return true
}
