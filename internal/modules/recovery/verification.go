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

type RestoreVerificationProbe interface {
	ProbeRestoredBackup(ctx context.Context, result *RestoreResult) error
}

type RestoreVerificationTarget struct {
	RestoreTarget
	Probe RestoreVerificationProbe
}

type RestoreVerificationService struct {
	backups       backupRepository
	verifications verificationRepository
	runner        *RestoreRunner
	now           func() time.Time
}

type RestoreVerificationResult struct {
	BackupSet     BackupSet
	Run           RestoreVerificationRun
	RestoreResult RestoreResult
	Artifact      RestoreVerificationArtifact
	ArtifactProof BackupArtifactProof
}

func NewRestoreVerificationService(repository recoveryRepository, runner *RestoreRunner) *RestoreVerificationService {
	return &RestoreVerificationService{
		backups:       repository,
		verifications: repository,
		runner:        runner,
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

func (service *RestoreVerificationService) VerifyLatestSuccessfulRetained(ctx context.Context, target RestoreVerificationTarget, asOf time.Time, verificationBasis RestoreVerificationBasis) (RestoreVerificationResult, error) {
	if service == nil || service.backups == nil || service.verifications == nil || service.runner == nil {
		return RestoreVerificationResult{}, fmt.Errorf("%w: restore verification requires store and runner", ErrInvalidBackupMetadata)
	}
	if err := verificationBasis.Validate(); err != nil {
		return RestoreVerificationResult{}, err
	}
	if asOf.IsZero() {
		asOf = service.now()
	}
	asOf = asOf.UTC()
	backupSet, err := NewBackupCatalog(
		service.backups,
		service.runner.storage,
		service.runner.extensionBackups,
		service.runner.stateCatalog,
	).RestoreCandidateBackup(ctx, asOf)
	if err != nil {
		return RestoreVerificationResult{}, err
	}
	return service.VerifyBackupSet(ctx, target, backupSet, verificationBasis)
}

func (service *RestoreVerificationService) VerifyBackupSet(ctx context.Context, target RestoreVerificationTarget, backupSet BackupSet, verificationBasis RestoreVerificationBasis) (RestoreVerificationResult, error) {
	return service.VerifyBackupSetAttempt(ctx, target, backupSet, verificationBasis, uuid.New())
}

func (service *RestoreVerificationService) VerifyBackupSetAttempt(
	ctx context.Context,
	target RestoreVerificationTarget,
	backupSet BackupSet,
	verificationBasis RestoreVerificationBasis,
	runID uuid.UUID,
) (RestoreVerificationResult, error) {
	if service == nil || service.backups == nil || service.verifications == nil || service.runner == nil {
		return RestoreVerificationResult{}, fmt.Errorf("%w: restore verification requires store and runner", ErrInvalidBackupMetadata)
	}
	if runID == uuid.Nil {
		return RestoreVerificationResult{}, fmt.Errorf("%w: restore verification attempt ID is required", ErrInvalidBackupMetadata)
	}
	verificationBasisSHA256, err := verificationBasis.SHA256()
	if err != nil {
		return RestoreVerificationResult{}, err
	}
	startedAt := service.now().UTC()

	restoreTarget := target.RestoreTarget
	restoreTarget.Readiness = nil
	restoreResult, restoreErr := service.runner.RestoreBackupSet(ctx, restoreTarget, backupSet)
	var selectedIncidentID *string
	var incidentErr error
	var probeErr error
	if restoreErr == nil {
		selectedIncidentID, incidentErr = selectRestoreVerificationIncidentID(ctx, restoreTarget.Postgres)
		restoreResult.SelectedIncidentID = selectedIncidentID
		if incidentErr != nil {
			restoreErr = incidentErr
		} else if selectedIncidentID != nil {
			if target.Probe == nil {
				probeErr = fmt.Errorf("%w: restore verification probe is required when incidents exist", ErrInvalidBackupArtifact)
			} else {
				probeErr = target.Probe.ProbeRestoredBackup(ctx, &restoreResult)
			}
			if probeErr != nil {
				restoreErr = probeErr
			}
		}
	}

	completedAt := service.now().UTC()
	artifact, artifactProof, artifactErr := service.writeRestoreVerificationArtifact(
		ctx,
		runID,
		backupSet,
		restoreResult,
		restoreErr,
		incidentErr,
		probeErr,
		selectedIncidentID,
		verificationBasis,
		verificationBasisSHA256,
		completedAt,
	)
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
		updated, run, recordErr := service.verifications.RecordRestoreVerificationCompletion(ctx, runParams)
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
	updated, run, err := service.verifications.RecordRestoreVerificationCompletion(ctx, runParams)
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
	restoreResult RestoreResult,
	restoreErr error,
	incidentErr error,
	probeErr error,
	selectedIncidentID *string,
	verificationBasis RestoreVerificationBasis,
	verificationBasisSHA256 string,
	completedAt time.Time,
) (RestoreVerificationArtifact, BackupArtifactProof, error) {
	if service == nil || service.runner == nil || service.runner.storage == nil {
		return RestoreVerificationArtifact{}, BackupArtifactProof{}, fmt.Errorf("%w: restore verification artifact requires backup storage", ErrInvalidBackupMetadata)
	}
	artifact := buildRestoreVerificationArtifact(
		runID,
		backupSet,
		restoreResult,
		restoreErr,
		incidentErr,
		probeErr,
		selectedIncidentID,
		verificationBasis,
		verificationBasisSHA256,
		completedAt,
	)
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
	runID uuid.UUID,
	backupSet BackupSet,
	restoreResult RestoreResult,
	restoreErr error,
	incidentErr error,
	probeErr error,
	selectedIncidentID *string,
	verificationBasis RestoreVerificationBasis,
	verificationBasisSHA256 string,
	completedAt time.Time,
) RestoreVerificationArtifact {
	workbookProbe := RestoreVerificationWorkbookProbeArtifact{
		Status: "skipped",
		Reason: "no_incidents",
	}
	if restoreErr != nil && selectedIncidentID == nil {
		workbookProbe.Reason = "verification_failed_before_probe"
	}
	if selectedIncidentID != nil && restoreResult.WorkbookProbe == nil {
		workbookProbe.Reason = "verification_failed_before_probe"
	}
	if selectedIncidentID != nil && restoreResult.WorkbookProbe != nil {
		workbookProbe = RestoreVerificationWorkbookProbeArtifact{Status: "executed"}
		executed := restoreResult.WorkbookProbe
		rowCount := executed.RowCount
		workbookProbe.RegistrationID = executed.RegistrationID
		workbookProbe.ViewSchemaID = executed.ViewSchemaID
		workbookProbe.RowCount = &rowCount
	}

	result := "pass"
	if restoreErr != nil || incidentErr != nil || probeErr != nil {
		result = "fail"
	}
	return RestoreVerificationArtifact{
		SchemaID:                   RestoreVerificationArtifactSchemaID,
		VerificationAttemptID:      runID.String(),
		BackupSetID:                backupSet.BackupSetID.String(),
		ConsistencyPointAt:         backupSet.ConsistencyPointAt,
		VerificationBasis:          verificationBasis,
		VerificationBasisSHA256:    verificationBasisSHA256,
		RecoveryStateCatalogSHA256: verificationBasis.RecoveryStateCatalogSHA256,
		CodecRegistrySHA256:        verificationBasis.CodecRegistrySHA256,
		ManifestSHA256:             restoreVerificationManifestSHA256(backupSet, restoreResult),
		RestoredObjectCount:        restoreResult.RestoredObjectCount,
		SelectedIncidentID:         selectedIncidentID,
		WorkbookProbe:              workbookProbe,
		Result:                     result,
		CompletedAt:                completedAt,
	}
}

func restoreVerificationManifestSHA256(backupSet BackupSet, result RestoreResult) string {
	if validSHA256Hex(result.IntegrityManifestSHA256) {
		return result.IntegrityManifestSHA256
	}
	return backupSet.IntegrityManifestSHA256
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
