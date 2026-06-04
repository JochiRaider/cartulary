package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const RestoreVerificationInterval = 7 * 24 * time.Hour

type RestoreVerificationProbe interface {
	ProbeRestoredBackup(ctx context.Context, result RestoreResult) error
}

type RestoreVerificationTarget struct {
	RestoreTarget
	Probe RestoreVerificationProbe
}

type RestoreVerificationService struct {
	store  *Store
	runner *RestoreRunner
	now    func() time.Time
}

type RestoreVerificationResult struct {
	BackupSet     BackupSet
	Run           RestoreVerificationRun
	RestoreResult RestoreResult
	Artifact      RestoreVerificationArtifact
	ArtifactProof BackupArtifactProof
}

func NewRestoreVerificationService(store *Store, runner *RestoreRunner) *RestoreVerificationService {
	return &RestoreVerificationService{
		store:  store,
		runner: runner,
		now: func() time.Time {
			return time.Now().UTC()
		},
	}
}

func RestoreVerificationBasisSHA256(parts map[string]string) (string, error) {
	if len(parts) == 0 {
		return "", ErrInvalidVerificationBasis
	}
	normalized := make(map[string]string, len(parts))
	keys := make([]string, 0, len(parts))
	for key, value := range parts {
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" || value == "" {
			return "", ErrInvalidVerificationBasis
		}
		keys = append(keys, key)
		normalized[key] = value
	}
	sort.Strings(keys)
	canonical := make([]struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}, 0, len(keys))
	for _, key := range keys {
		canonical = append(canonical, struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}{Key: key, Value: normalized[key]})
	}
	body, err := json.Marshal(canonical)
	if err != nil {
		return "", fmt.Errorf("encode restore verification basis: %w", err)
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}

func (service *RestoreVerificationService) VerifyLatestSuccessfulRetained(ctx context.Context, target RestoreVerificationTarget, asOf time.Time, verificationBasisSHA256 string) (RestoreVerificationResult, error) {
	if service == nil || service.store == nil || service.runner == nil {
		return RestoreVerificationResult{}, fmt.Errorf("%w: restore verification requires store and runner", ErrInvalidBackupMetadata)
	}
	if !validSHA256Hex(verificationBasisSHA256) {
		return RestoreVerificationResult{}, ErrInvalidVerificationBasis
	}
	if asOf.IsZero() {
		asOf = service.now()
	}
	asOf = asOf.UTC()
	backupSet, err := NewBackupCatalog(service.store, service.runner.storage).RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return RestoreVerificationResult{}, err
	}
	return service.VerifyBackupSet(ctx, target, backupSet, verificationBasisSHA256)
}

func (service *RestoreVerificationService) VerifyBackupSet(ctx context.Context, target RestoreVerificationTarget, backupSet BackupSet, verificationBasisSHA256 string) (RestoreVerificationResult, error) {
	if service == nil || service.store == nil || service.runner == nil {
		return RestoreVerificationResult{}, fmt.Errorf("%w: restore verification requires store and runner", ErrInvalidBackupMetadata)
	}
	if !validSHA256Hex(verificationBasisSHA256) {
		return RestoreVerificationResult{}, ErrInvalidVerificationBasis
	}
	startedAt := service.now().UTC()
	runID := uuid.New()

	restoreTarget := target.RestoreTarget
	restoreTarget.Readiness = nil
	recorder := &restoreVerificationStepRecorder{}
	restoreTarget.Observer = restoreVerificationObserver{
		outer:    restoreTarget.Observer,
		recorder: recorder,
	}
	restoreResult, restoreErr := service.runner.RestoreBackupSet(ctx, restoreTarget, backupSet)
	selectedIncidentID, incidentErr := selectRestoreVerificationIncidentID(ctx, restoreTarget.Postgres)
	var probeErr error
	if restoreErr == nil {
		if incidentErr != nil {
			restoreErr = incidentErr
		} else if selectedIncidentID != nil {
			if target.Probe == nil {
				probeErr = fmt.Errorf("%w: restore verification probe is required when incidents exist", ErrInvalidBackupArtifact)
			} else {
				probeErr = target.Probe.ProbeRestoredBackup(ctx, restoreResult)
			}
			if probeErr != nil {
				restoreErr = probeErr
			}
		}
	}

	completedAt := service.now().UTC()
	artifact, artifactProof, artifactErr := service.writeRestoreVerificationArtifact(ctx, runID, backupSet, restoreTarget, restoreResult, restoreErr, incidentErr, probeErr, selectedIncidentID, recorder)
	if artifactErr != nil {
		restoreErr = artifactErr
		completedAt = service.now().UTC()
	}
	if restoreErr != nil {
		runParams := CreateRestoreVerificationRunParams{
			RestoreVerificationRunID: runID,
			BackupSetID:              backupSet.BackupSetID,
			StartedAt:                startedAt,
			CompletedAt:              completedAt,
			VerificationState:        VerificationFailed,
			VerificationBasisSHA256:  verificationBasisSHA256,
			FailureReason:            restoreVerificationFailureReason(restoreErr),
			FailureMessage:           "restore verification failed; inspect protected run logs for details",
			ConsistencyReport:        restoreResult.ConsistencyReport,
		}
		updated, run, recordErr := service.store.RecordRestoreVerificationCompletion(ctx, runParams)
		if recordErr != nil {
			return RestoreVerificationResult{}, fmt.Errorf("%w; additionally failed to record restore verification failure: %v", restoreErr, recordErr)
		}
		return RestoreVerificationResult{
			BackupSet:     updated,
			Run:           run,
			RestoreResult: restoreResult,
			Artifact:      artifact,
			ArtifactProof: artifactProof,
		}, restoreErr
	}

	runParams := CreateRestoreVerificationRunParams{
		RestoreVerificationRunID: runID,
		BackupSetID:              restoreResult.BackupSet.BackupSetID,
		StartedAt:                startedAt,
		CompletedAt:              completedAt,
		VerificationState:        VerificationVerified,
		VerificationBasisSHA256:  verificationBasisSHA256,
		ConsistencyReport:        restoreResult.ConsistencyReport,
	}
	updated, run, err := service.store.RecordRestoreVerificationCompletion(ctx, runParams)
	if err != nil {
		return RestoreVerificationResult{}, err
	}
	return RestoreVerificationResult{
		BackupSet:     updated,
		Run:           run,
		RestoreResult: restoreResult,
		Artifact:      artifact,
		ArtifactProof: artifactProof,
	}, nil
}

func (service *RestoreVerificationService) writeRestoreVerificationArtifact(
	ctx context.Context,
	runID uuid.UUID,
	backupSet BackupSet,
	target RestoreTarget,
	restoreResult RestoreResult,
	restoreErr error,
	incidentErr error,
	probeErr error,
	selectedIncidentID *string,
	recorder *restoreVerificationStepRecorder,
) (RestoreVerificationArtifact, BackupArtifactProof, error) {
	if service == nil || service.runner == nil || service.runner.storage == nil {
		return RestoreVerificationArtifact{}, BackupArtifactProof{}, fmt.Errorf("%w: restore verification artifact requires backup storage", ErrInvalidBackupMetadata)
	}
	artifact := buildRestoreVerificationArtifact(ctx, backupSet, target, restoreResult, restoreErr, incidentErr, probeErr, selectedIncidentID, recorder)
	body, err := EncodeRestoreVerificationArtifact(artifact)
	if err != nil {
		return RestoreVerificationArtifact{}, BackupArtifactProof{}, fmt.Errorf("encode restore verification artifact: %w", err)
	}
	decoded, err := DecodeRestoreVerificationArtifact(body)
	if err != nil {
		return RestoreVerificationArtifact{}, BackupArtifactProof{}, err
	}
	key := fmt.Sprintf("backup_sets/%s/restore-verification-%s.json", backupSet.BackupSetID.String(), runID.String())
	proof, err := service.runner.storage.WriteArtifact(ctx, key, body, "application/json")
	if err != nil {
		return decoded, BackupArtifactProof{}, fmt.Errorf("write restore verification artifact: %w", err)
	}
	return decoded, proof, nil
}

func buildRestoreVerificationArtifact(
	ctx context.Context,
	backupSet BackupSet,
	target RestoreTarget,
	restoreResult RestoreResult,
	restoreErr error,
	incidentErr error,
	probeErr error,
	selectedIncidentID *string,
	recorder *restoreVerificationStepRecorder,
) RestoreVerificationArtifact {
	manifestResult := "fail"
	if ValidateObjectStoreBackupManifestForBackup(backupSet, restoreResult.ObjectStoreBackupManifest) == nil {
		manifestResult = "pass"
	}
	projectionResult := "fail"
	if recorder != nil && recorder.Contains(RestoreStepConsistencyCheck) {
		projectionResult = "pass"
	}
	if restoreErr == nil {
		projectionResult = "pass"
	}

	blobCounts := restoreVerificationBlobCounts(ctx, target)
	incidentStatus := "skipped_no_incidents"
	queryViewSchemaID := ""
	if selectedIncidentID != nil {
		queryViewSchemaID = RestoreVerificationTimelineViewID
		if probeErr == nil && incidentErr == nil && restoreErr == nil {
			incidentStatus = "pass"
		} else {
			incidentStatus = "fail"
		}
	}

	result := "pass"
	reasons := make([]string, 0)
	if restoreErr != nil || incidentErr != nil || probeErr != nil ||
		manifestResult != "pass" || projectionResult != "pass" || blobCounts.Failed != 0 || incidentStatus == "fail" {
		result = "fail"
		reasons = restoreVerificationArtifactFailureReasons(restoreErr, incidentErr, probeErr, manifestResult, projectionResult, blobCounts, incidentStatus)
	}
	return RestoreVerificationArtifact{
		SchemaID:                RestoreVerificationArtifactSchemaID,
		BackupSetID:             backupSet.BackupSetID.String(),
		SelectedIncidentID:      selectedIncidentID,
		IncidentOpenCheck:       RestoreVerificationIncidentOpenCheck{Status: incidentStatus},
		QueryViewSchemaID:       queryViewSchemaID,
		BlobCheckCounts:         blobCounts,
		ManifestCheckResult:     manifestResult,
		ProjectionRebuildResult: projectionResult,
		Result:                  result,
		FailureReasons:          reasons,
	}
}

func restoreVerificationBlobCounts(ctx context.Context, target RestoreTarget) RestoreVerificationBlobCheckCounts {
	if target.Postgres == nil || target.ObjectStore == nil {
		return RestoreVerificationBlobCheckCounts{Failed: 1, Total: 1}
	}
	_, count, err := verifyRestoredBlobRowsDetailed(ctx, target.Postgres, target.ObjectStore)
	if err != nil {
		return RestoreVerificationBlobCheckCounts{Total: 1, Failed: 1}
	}
	return RestoreVerificationBlobCheckCounts{Total: count, Passed: count}
}

func restoreVerificationArtifactFailureReasons(restoreErr error, incidentErr error, probeErr error, manifestResult string, projectionResult string, blobCounts RestoreVerificationBlobCheckCounts, incidentStatus string) []string {
	reasons := make(map[string]struct{})
	add := func(reason string) {
		reasons[reason] = struct{}{}
	}
	if restoreErr != nil {
		switch {
		case errors.Is(restoreErr, ErrInvalidBackupArtifact):
			add("backup_artifact_invalid")
		case errors.Is(restoreErr, ErrNoSuccessfulRetainedBackup):
			add("backup_selection_failed")
		case errors.Is(restoreErr, ErrLatestSuccessfulBackupStale):
			add("backup_selection_failed")
		default:
			add("restore_failed")
		}
	}
	if manifestResult != "pass" {
		add("manifest_check_failed")
	}
	if projectionResult != "pass" {
		add("projection_rebuild_failed")
	}
	if blobCounts.Failed != 0 {
		add("blob_lifecycle_inconsistent")
	}
	if incidentErr != nil {
		add("incident_open_failed")
	}
	if probeErr != nil || incidentStatus == "fail" {
		add("workbook_query_failed")
	}
	if len(reasons) == 0 {
		add("restore_verification_failed")
	}
	ordered := make([]string, 0, len(reasons))
	for reason := range reasons {
		ordered = append(ordered, reason)
	}
	sort.Strings(ordered)
	return ordered
}

func selectRestoreVerificationIncidentID(ctx context.Context, target postgresQueryer) (*string, error) {
	if target == nil {
		return nil, fmt.Errorf("%w: restore verification target postgres is required", ErrInvalidBackupArtifact)
	}
	var incidentID string
	if err := target.QueryRow(ctx, `
SELECT id::text
FROM incidents
ORDER BY id::text ASC
LIMIT 1
`).Scan(&incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select restore verification incident: %w", err)
	}
	return &incidentID, nil
}

type postgresQueryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type restoreVerificationStepRecorder struct {
	Steps []RestoreStep
}

func (recorder *restoreVerificationStepRecorder) RecordRestoreStep(step RestoreStep) {
	recorder.Steps = append(recorder.Steps, step)
}

func (recorder *restoreVerificationStepRecorder) Contains(step RestoreStep) bool {
	if recorder == nil {
		return false
	}
	for _, got := range recorder.Steps {
		if got == step {
			return true
		}
	}
	return false
}

type restoreVerificationObserver struct {
	outer    RestoreStepObserver
	recorder *restoreVerificationStepRecorder
}

func (observer restoreVerificationObserver) RecordRestoreStep(step RestoreStep) {
	if observer.outer != nil {
		observer.outer.RecordRestoreStep(step)
	}
	if observer.recorder != nil {
		observer.recorder.RecordRestoreStep(step)
	}
}

func restoreVerificationFailureReason(err error) string {
	switch {
	case errors.Is(err, ErrInvalidBackupArtifact):
		return "invalid_backup_artifact"
	case errors.Is(err, ErrNoSuccessfulRetainedBackup):
		return "no_successful_retained_backup"
	case errors.Is(err, ErrAmbiguousBackupSelection):
		return "ambiguous_backup_selection"
	case errors.Is(err, ErrLatestSuccessfulBackupStale):
		return "latest_backup_stale"
	default:
		return "restore_verification_failed"
	}
}
