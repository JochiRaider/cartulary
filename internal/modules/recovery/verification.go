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

	restoreTarget := target.RestoreTarget
	restoreTarget.Readiness = nil
	restoreResult, restoreErr := service.runner.RestoreBackupSet(ctx, restoreTarget, backupSet)
	if restoreErr == nil && target.Probe != nil {
		restoreErr = target.Probe.ProbeRestoredBackup(ctx, restoreResult)
	}

	completedAt := service.now().UTC()
	if restoreErr != nil {
		runParams := CreateRestoreVerificationRunParams{
			BackupSetID:             backupSet.BackupSetID,
			StartedAt:               startedAt,
			CompletedAt:             completedAt,
			VerificationState:       VerificationFailed,
			VerificationBasisSHA256: verificationBasisSHA256,
			FailureReason:           restoreVerificationFailureReason(restoreErr),
			FailureMessage:          "restore verification failed; inspect protected run logs for details",
			ConsistencyReport:       restoreResult.ConsistencyReport,
		}
		updated, run, recordErr := service.store.RecordRestoreVerificationCompletion(ctx, runParams)
		if recordErr != nil {
			return RestoreVerificationResult{}, fmt.Errorf("%w; additionally failed to record restore verification failure: %v", restoreErr, recordErr)
		}
		return RestoreVerificationResult{
			BackupSet:     updated,
			Run:           run,
			RestoreResult: restoreResult,
		}, restoreErr
	}

	runParams := CreateRestoreVerificationRunParams{
		BackupSetID:             restoreResult.BackupSet.BackupSetID,
		StartedAt:               startedAt,
		CompletedAt:             completedAt,
		VerificationState:       VerificationVerified,
		VerificationBasisSHA256: verificationBasisSHA256,
		ConsistencyReport:       restoreResult.ConsistencyReport,
	}
	updated, run, err := service.store.RecordRestoreVerificationCompletion(ctx, runParams)
	if err != nil {
		return RestoreVerificationResult{}, err
	}
	return RestoreVerificationResult{
		BackupSet:     updated,
		Run:           run,
		RestoreResult: restoreResult,
	}, nil
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
