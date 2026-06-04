package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const (
	ObjectStoreMigrationRunSchemaID        = "cartulary.object_store_migration_run.v1"
	ObjectStoreMigrationCopyLedgerSchemaID = "cartulary.object_store_migration_copy_ledger.v1"
	ObjectStoreMigrationValidationSchemaID = "cartulary.object_store_migration_validation.v1"
	ObjectStoreMigrationRollbackSchemaID   = "cartulary.object_store_migration_rollback.v1"
	ObjectStoreMigrationProofSchemaID      = "cartulary.object_store_migration_write_quiescence.v1"
	ObjectStoreMigrationProbeSchemaID      = "cartulary.object_store_migration_target_probe.v1"

	ObjectStoreMigrationValidationSchemaVersion = "1.0.0"
	ObjectStoreMigrationToolVersion             = "cartulary-object-store-migration/2026-06-phase-f"

	ObjectStoreBackendMinIOS3 = "minio_s3"
)

type ObjectStoreMigrationState string

const (
	ObjectStoreMigrationStatePlanned             ObjectStoreMigrationState = "planned"
	ObjectStoreMigrationStatePreflighted         ObjectStoreMigrationState = "preflighted"
	ObjectStoreMigrationStateApplicationStopped  ObjectStoreMigrationState = "application_stopped"
	ObjectStoreMigrationStateBackupCaptured      ObjectStoreMigrationState = "backup_captured"
	ObjectStoreMigrationStateTargetPrepared      ObjectStoreMigrationState = "target_prepared"
	ObjectStoreMigrationStateCopying             ObjectStoreMigrationState = "copying"
	ObjectStoreMigrationStateCopied              ObjectStoreMigrationState = "copied"
	ObjectStoreMigrationStateValidating          ObjectStoreMigrationState = "validating"
	ObjectStoreMigrationStateCutoverReady        ObjectStoreMigrationState = "cutover_ready"
	ObjectStoreMigrationStateCutoverCommitted    ObjectStoreMigrationState = "cutover_committed"
	ObjectStoreMigrationStatePostCutoverVerified ObjectStoreMigrationState = "post_cutover_verified"
	ObjectStoreMigrationStateRolledBack          ObjectStoreMigrationState = "rolled_back"
	ObjectStoreMigrationStateFailed              ObjectStoreMigrationState = "failed"
)

type ObjectStoreMigrationEventName string

const (
	ObjectStoreMigrationEventPlanCreated             ObjectStoreMigrationEventName = "plan_created"
	ObjectStoreMigrationEventPreflightPassed         ObjectStoreMigrationEventName = "preflight_passed"
	ObjectStoreMigrationEventWriteQuiescenceVerified ObjectStoreMigrationEventName = "write_quiescence_verified"
	ObjectStoreMigrationEventBackupCaptured          ObjectStoreMigrationEventName = "backup_captured"
	ObjectStoreMigrationEventTargetPrepared          ObjectStoreMigrationEventName = "target_prepared"
	ObjectStoreMigrationEventCopyStarted             ObjectStoreMigrationEventName = "copy_started"
	ObjectStoreMigrationEventCopyCompleted           ObjectStoreMigrationEventName = "copy_completed"
	ObjectStoreMigrationEventValidationStarted       ObjectStoreMigrationEventName = "validation_started"
	ObjectStoreMigrationEventValidationPassed        ObjectStoreMigrationEventName = "validation_passed"
	ObjectStoreMigrationEventCutoverCommitted        ObjectStoreMigrationEventName = "cutover_committed"
	ObjectStoreMigrationEventPostCutoverVerified     ObjectStoreMigrationEventName = "post_cutover_verified"
	ObjectStoreMigrationEventRollbackRequested       ObjectStoreMigrationEventName = "rollback_requested"
	ObjectStoreMigrationEventBlockingFailure         ObjectStoreMigrationEventName = "blocking_failure"
)

type ObjectStoreMigrationCopyStatus string

const (
	ObjectStoreMigrationCopyCopied                   ObjectStoreMigrationCopyStatus = "copied"
	ObjectStoreMigrationCopyAlreadyCopied            ObjectStoreMigrationCopyStatus = "already_copied"
	ObjectStoreMigrationCopyMissingSource            ObjectStoreMigrationCopyStatus = "missing_source"
	ObjectStoreMigrationCopyTargetMismatch           ObjectStoreMigrationCopyStatus = "target_mismatch"
	ObjectStoreMigrationCopyUnsupportedSourceFeature ObjectStoreMigrationCopyStatus = "unsupported_source_feature"
	ObjectStoreMigrationCopyError                    ObjectStoreMigrationCopyStatus = "error"
)

type ObjectStoreMigrationValidationStatus string

const (
	ObjectStoreMigrationValidationPass                     ObjectStoreMigrationValidationStatus = "pass"
	ObjectStoreMigrationValidationMissingSource            ObjectStoreMigrationValidationStatus = "missing_source"
	ObjectStoreMigrationValidationMissingTarget            ObjectStoreMigrationValidationStatus = "missing_target"
	ObjectStoreMigrationValidationSizeMismatch             ObjectStoreMigrationValidationStatus = "size_mismatch"
	ObjectStoreMigrationValidationHashMismatch             ObjectStoreMigrationValidationStatus = "hash_mismatch"
	ObjectStoreMigrationValidationUnsupportedSourceFeature ObjectStoreMigrationValidationStatus = "unsupported_source_feature"
	ObjectStoreMigrationValidationError                    ObjectStoreMigrationValidationStatus = "error"
)

type ObjectStoreMigrationArtifactRef struct {
	Key         string `json:"key"`
	SHA256      string `json:"sha256"`
	SizeBytes   int64  `json:"size_bytes"`
	ContentType string `json:"content_type"`
}

type ObjectStoreMigrationBackupRefs struct {
	BackupSetID                       string                           `json:"backup_set_id"`
	IntegrityManifest                 ObjectStoreMigrationArtifactRef  `json:"integrity_manifest"`
	PostgresArtifact                  ObjectStoreMigrationArtifactRef  `json:"postgres_artifact"`
	ObjectStoreArtifact               ObjectStoreMigrationArtifactRef  `json:"object_store_artifact"`
	ObjectStoreBackupManifestArtifact *ObjectStoreMigrationArtifactRef `json:"object_store_backup_manifest_artifact,omitempty"`
	ObjectStoreBackupSummaryArtifact  *ObjectStoreMigrationArtifactRef `json:"object_store_backup_summary_artifact,omitempty"`
}

type ObjectStoreMigrationRun struct {
	SchemaID          string                           `json:"schema_id"`
	RunID             string                           `json:"run_id"`
	CreatedAt         time.Time                        `json:"created_at"`
	UpdatedAt         time.Time                        `json:"updated_at"`
	CurrentState      ObjectStoreMigrationState        `json:"current_state"`
	StateTimestamps   map[string]time.Time             `json:"state_timestamps"`
	Events            []ObjectStoreMigrationEvent      `json:"events"`
	OperatorIdentity  string                           `json:"operator_identity"`
	SourceEndpointRef RedactionRef                     `json:"source_endpoint_ref"`
	TargetEndpointRef RedactionRef                     `json:"target_endpoint_ref"`
	SourceBucketRef   RedactionRef                     `json:"source_bucket_ref"`
	TargetBucketRef   RedactionRef                     `json:"target_bucket_ref"`
	BackupRefs        []ObjectStoreMigrationBackupRefs `json:"backup_refs"`
	ProbeRef          *ObjectStoreMigrationArtifactRef `json:"probe_ref"`
	CopyLedgerRef     *ObjectStoreMigrationArtifactRef `json:"copy_ledger_ref"`
	ValidationRef     *ObjectStoreMigrationArtifactRef `json:"validation_ref"`
	RollbackRef       *ObjectStoreMigrationArtifactRef `json:"rollback_ref"`
	TerminalResult    *ObjectStoreMigrationState       `json:"terminal_result"`
}

type ObjectStoreMigrationEvent struct {
	Sequence   int                           `json:"sequence"`
	Event      ObjectStoreMigrationEventName `json:"event"`
	FromState  *ObjectStoreMigrationState    `json:"from_state"`
	ToState    ObjectStoreMigrationState     `json:"to_state"`
	OccurredAt time.Time                     `json:"occurred_at"`
	Detail     map[string]string             `json:"detail,omitempty"`
}

type ObjectStoreMigrationWriteQuiescenceProof struct {
	SchemaID                string    `json:"schema_id"`
	ProofKind               string    `json:"proof_kind"`
	CheckedAt               time.Time `json:"checked_at"`
	ProcessState            string    `json:"process_state"`
	HTTPListenerClosed      bool      `json:"http_listener_closed"`
	WebSocketListenerClosed bool      `json:"websocket_listener_closed"`
}

type ObjectStoreMigrationCopyLedger struct {
	SchemaID        string                               `json:"schema_id"`
	RunID           string                               `json:"run_id"`
	SourceBackend   string                               `json:"source_backend"`
	TargetBackend   string                               `json:"target_backend"`
	SourceBucketRef RedactionRef                         `json:"source_bucket_ref"`
	TargetBucketRef RedactionRef                         `json:"target_bucket_ref"`
	ObjectCount     int                                  `json:"object_count"`
	StatusCounts    map[string]int                       `json:"status_counts"`
	Items           []ObjectStoreMigrationCopyLedgerItem `json:"items"`
	Result          string                               `json:"result"`
	ArtifactSHA256  string                               `json:"artifact_sha256"`
}

type ObjectStoreMigrationCopyLedgerItem struct {
	Sequence             int                            `json:"sequence"`
	ObjectBlobID         string                         `json:"object_blob_id,omitempty"`
	SourceBucketRef      RedactionRef                   `json:"source_bucket_ref"`
	SourceKeyRef         RedactionRef                   `json:"source_key_ref"`
	TargetBucketRef      RedactionRef                   `json:"target_bucket_ref"`
	TargetKeyRef         RedactionRef                   `json:"target_key_ref"`
	SourceSizeBytes      int64                          `json:"source_size_bytes"`
	SourceSHA256         string                         `json:"source_sha256"`
	TargetSizeBytes      *int64                         `json:"target_size_bytes,omitempty"`
	TargetSHA256         string                         `json:"target_sha256,omitempty"`
	Status               ObjectStoreMigrationCopyStatus `json:"status"`
	ReasonCode           string                         `json:"reason_code"`
	IdempotencyKeySHA256 string                         `json:"idempotency_key_sha256"`
}

type ObjectStoreMigrationValidation struct {
	SchemaID              string                                   `json:"schema_id"`
	SchemaVersion         string                                   `json:"schema_version"`
	ValidationToolVersion string                                   `json:"validation_tool_version"`
	RunID                 string                                   `json:"run_id"`
	StartedAt             time.Time                                `json:"started_at"`
	CompletedAt           *time.Time                               `json:"completed_at"`
	SourceBackend         string                                   `json:"source_backend"`
	TargetBackend         string                                   `json:"target_backend"`
	SourceSnapshotID      string                                   `json:"source_snapshot_id"`
	TargetSnapshotID      string                                   `json:"target_snapshot_id"`
	SourceBucket          string                                   `json:"source_bucket"`
	TargetBucket          string                                   `json:"target_bucket"`
	IncidentCount         int                                      `json:"incident_count"`
	ObjectBlobCount       int                                      `json:"object_blob_count"`
	ObjectsChecked        []ObjectStoreMigrationValidationObject   `json:"objects_checked"`
	PreviewSampleChecks   []ObjectStoreMigrationPreviewSampleCheck `json:"preview_sample_checks"`
	BlockingDiagnostics   []ObjectStoreMigrationDiagnostic         `json:"blocking_diagnostics"`
	NonblockingWarnings   []ObjectStoreMigrationDiagnostic         `json:"nonblocking_warnings"`
	Result                string                                   `json:"result"`
	ArtifactSHA256        string                                   `json:"artifact_sha256"`
}

type ObjectStoreMigrationValidationObject struct {
	ObjectBlobID     string                               `json:"object_blob_id"`
	IncidentID       string                               `json:"incident_id"`
	StorageRefSHA256 string                               `json:"storage_ref_sha256"`
	SourceSizeBytes  *int64                               `json:"source_size_bytes,omitempty"`
	TargetSizeBytes  *int64                               `json:"target_size_bytes,omitempty"`
	SourceSHA256     string                               `json:"source_sha256,omitempty"`
	TargetSHA256     string                               `json:"target_sha256,omitempty"`
	Status           ObjectStoreMigrationValidationStatus `json:"status"`
	ReasonCode       string                               `json:"reason_code"`
}

type ObjectStoreMigrationPreviewSampleCheck struct {
	ObjectBlobID string `json:"object_blob_id"`
	IncidentID   string `json:"incident_id"`
	RouteClass   string `json:"route_class"`
	Status       string `json:"status"`
	ReasonCode   string `json:"reason_code"`
}

type ObjectStoreMigrationDiagnostic struct {
	DiagnosticID string         `json:"diagnostic_id"`
	Severity     string         `json:"severity"`
	ReasonCode   string         `json:"reason_code"`
	ObjectBlobID *string        `json:"object_blob_id"`
	IncidentID   *string        `json:"incident_id"`
	Message      string         `json:"message"`
	Refs         []RedactionRef `json:"refs"`
}

type ObjectStoreMigrationRollbackEvidence struct {
	SchemaID                       string    `json:"schema_id"`
	RunID                          string    `json:"run_id"`
	CreatedAt                      time.Time `json:"created_at"`
	BeforeCutoverSourceActive      bool      `json:"before_cutover_source_active"`
	BeforeCutoverBackupRetained    bool      `json:"before_cutover_backup_retained"`
	CutoverRollbackProcedure       string    `json:"cutover_rollback_procedure"`
	PostVerificationRollbackClosed bool      `json:"post_verification_rollback_closed"`
}

type ObjectStoreMigrationTargetProbe struct {
	SchemaID        string       `json:"schema_id"`
	RunID           string       `json:"run_id"`
	StartedAt       time.Time    `json:"started_at"`
	CompletedAt     time.Time    `json:"completed_at"`
	TargetBucketRef RedactionRef `json:"target_bucket_ref"`
	ProbeKeyRef     RedactionRef `json:"probe_key_ref"`
	Result          string       `json:"result"`
	SHA256          string       `json:"sha256"`
}

type ObjectStoreMigrationBlob struct {
	ObjectBlobID       uuid.UUID
	IncidentID         uuid.UUID
	StorageKey         string
	EvidenceStorageRef string
	ByteSize           int64
}

type ObjectStoreMigrationCopyParams struct {
	RunID         uuid.UUID
	SourceBackend string
	TargetBackend string
	SourceBucket  string
	TargetBucket  string
	SourceStore   objectstore.Store
	TargetStore   objectstore.Store
	Objects       []ObjectStoreMigrationBlob
}

type ObjectStoreMigrationValidationParams struct {
	RunID         uuid.UUID
	StartedAt     time.Time
	CompletedAt   time.Time
	SourceBackend string
	TargetBackend string
	SourceBucket  string
	TargetBucket  string
	SourceStore   objectstore.Store
	TargetStore   objectstore.Store
	Objects       []ObjectStoreMigrationBlob
}

type ObjectStoreMigrationArtifactSet struct {
	Run             ObjectStoreMigrationRun
	RunBody         []byte
	CopyLedger      ObjectStoreMigrationCopyLedger
	CopyLedgerBody  []byte
	Validation      ObjectStoreMigrationValidation
	ValidationBody  []byte
	Probe           ObjectStoreMigrationTargetProbe
	ProbeBody       []byte
	Rollback        ObjectStoreMigrationRollbackEvidence
	RollbackBody    []byte
	BlockingFailure bool
}

func ValidateObjectStoreMigrationWriteQuiescenceProof(proof ObjectStoreMigrationWriteQuiescenceProof) error {
	if proof.SchemaID != ObjectStoreMigrationProofSchemaID {
		return fmt.Errorf("%w: unsupported migration write-quiescence proof schema %q", ErrInvalidBackupArtifact, proof.SchemaID)
	}
	if proof.ProofKind != "process_stopped" {
		return fmt.Errorf("%w: migration write-quiescence proof_kind must be process_stopped", ErrInvalidBackupArtifact)
	}
	if proof.CheckedAt.IsZero() {
		return fmt.Errorf("%w: migration write-quiescence checked_at is required", ErrInvalidBackupArtifact)
	}
	switch proof.ProcessState {
	case "absent", "stopped_by_supervisor":
	default:
		return fmt.Errorf("%w: migration process_state must prove process absence or supervisor stop", ErrInvalidBackupArtifact)
	}
	if !proof.HTTPListenerClosed || !proof.WebSocketListenerClosed {
		return fmt.Errorf("%w: migration listeners must both be closed", ErrInvalidBackupArtifact)
	}
	return nil
}

func NewObjectStoreMigrationRun(runID uuid.UUID, now time.Time, operatorIdentity string, sourceEndpoint string, targetEndpoint string, sourceBucket string, targetBucket string) (ObjectStoreMigrationRun, error) {
	if runID == uuid.Nil {
		runID = uuid.New()
	}
	now = backupTimestamp(now)
	run := ObjectStoreMigrationRun{
		SchemaID:          ObjectStoreMigrationRunSchemaID,
		RunID:             runID.String(),
		CreatedAt:         now,
		UpdatedAt:         now,
		StateTimestamps:   map[string]time.Time{},
		Events:            []ObjectStoreMigrationEvent{},
		OperatorIdentity:  strings.TrimSpace(operatorIdentity),
		SourceEndpointRef: endpointRedactionRef(sourceEndpoint),
		TargetEndpointRef: endpointRedactionRef(targetEndpoint),
		SourceBucketRef:   hashRedactionRef("bucket", sourceBucket),
		TargetBucketRef:   hashRedactionRef("bucket", targetBucket),
		BackupRefs:        []ObjectStoreMigrationBackupRefs{},
	}
	if run.OperatorIdentity == "" {
		return ObjectStoreMigrationRun{}, fmt.Errorf("%w: migration operator_identity is required", ErrInvalidBackupArtifact)
	}
	if err := ApplyObjectStoreMigrationEvent(&run, ObjectStoreMigrationEventPlanCreated, now, nil); err != nil {
		return ObjectStoreMigrationRun{}, err
	}
	return run, nil
}

func ApplyObjectStoreMigrationEvent(run *ObjectStoreMigrationRun, event ObjectStoreMigrationEventName, at time.Time, detail map[string]string) error {
	if run == nil {
		return fmt.Errorf("%w: migration run is required", ErrInvalidBackupArtifact)
	}
	if isTerminalMigrationState(run.CurrentState) {
		return fmt.Errorf("%w: terminal migration run cannot transition", ErrInvalidBackupArtifact)
	}
	toState, err := migrationDestinationState(run.CurrentState, event)
	if err != nil {
		return err
	}
	at = backupTimestamp(at)
	var from *ObjectStoreMigrationState
	if run.CurrentState != "" {
		current := run.CurrentState
		from = &current
	}
	run.CurrentState = toState
	run.UpdatedAt = at
	if run.StateTimestamps == nil {
		run.StateTimestamps = map[string]time.Time{}
	}
	if _, exists := run.StateTimestamps[string(toState)]; !exists {
		run.StateTimestamps[string(toState)] = at
	}
	normalizedDetail := map[string]string(nil)
	if len(detail) > 0 {
		normalizedDetail = make(map[string]string, len(detail))
		for key, value := range detail {
			key = strings.TrimSpace(key)
			if key == "" {
				continue
			}
			normalizedDetail[key] = strings.TrimSpace(value)
		}
	}
	run.Events = append(run.Events, ObjectStoreMigrationEvent{
		Sequence:   len(run.Events) + 1,
		Event:      event,
		FromState:  from,
		ToState:    toState,
		OccurredAt: at,
		Detail:     normalizedDetail,
	})
	if isTerminalMigrationState(toState) {
		terminal := toState
		run.TerminalResult = &terminal
	}
	return nil
}

func BuildObjectStoreMigrationBackupRefs(backupSet BackupSet, manifest BackupIntegrityManifest) ObjectStoreMigrationBackupRefs {
	ref := ObjectStoreMigrationBackupRefs{
		BackupSetID: backupSet.BackupSetID.String(),
		IntegrityManifest: ObjectStoreMigrationArtifactRef{
			Key:         backupSet.IntegrityManifestKey,
			SHA256:      backupSet.IntegrityManifestSHA256,
			SizeBytes:   backupSet.IntegrityManifestSizeBytes,
			ContentType: "application/json",
		},
		PostgresArtifact:                  migrationArtifactRefFromProof(manifest.PostgresArtifact),
		ObjectStoreArtifact:               migrationArtifactRefFromProof(manifest.ObjectStoreArtifact),
		ObjectStoreBackupManifestArtifact: migrationArtifactRefPtr(manifest.ObjectStoreBackupManifestArtifact),
		ObjectStoreBackupSummaryArtifact:  migrationArtifactRefPtr(manifest.ObjectStoreBackupSummaryArtifact),
	}
	return ref
}

func LoadObjectStoreMigrationBackupRefs(ctx context.Context, storage BackupStorage, backupSet BackupSet) (ObjectStoreMigrationBackupRefs, error) {
	body, err := VerifyArtifactProof(ctx, storage, BackupArtifactProof{
		Key:       backupSet.IntegrityManifestKey,
		SHA256:    backupSet.IntegrityManifestSHA256,
		SizeBytes: backupSet.IntegrityManifestSizeBytes,
	})
	if err != nil {
		return ObjectStoreMigrationBackupRefs{}, err
	}
	manifest, err := DecodeIntegrityManifest(body)
	if err != nil {
		return ObjectStoreMigrationBackupRefs{}, err
	}
	return BuildObjectStoreMigrationBackupRefs(backupSet, manifest), nil
}

func ListObjectStoreMigrationBlobs(ctx context.Context, db postgres.DB) ([]ObjectStoreMigrationBlob, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: postgres DB is required for migration object list", ErrInvalidBackupArtifact)
	}
	rows, err := db.Query(ctx, `
SELECT b.object_blob_id::text,
       b.incident_id::text,
       b.storage_key,
       b.byte_size,
       e.storage_ref
  FROM object_blobs b
  LEFT JOIN evidence e ON e.object_blob_id = b.object_blob_id
 WHERE b.upload_state = 'available'
 ORDER BY b.object_blob_id::text ASC
`)
	if err != nil {
		return nil, fmt.Errorf("list migration object blobs: %w", err)
	}
	defer rows.Close()
	objects := make([]ObjectStoreMigrationBlob, 0)
	for rows.Next() {
		var objectBlobIDRaw string
		var incidentIDRaw string
		var storageKey string
		var byteSize int64
		var storageRef pgtype.Text
		if err := rows.Scan(&objectBlobIDRaw, &incidentIDRaw, &storageKey, &byteSize, &storageRef); err != nil {
			return nil, fmt.Errorf("scan migration object blob: %w", err)
		}
		objectBlobID, err := uuid.Parse(objectBlobIDRaw)
		if err != nil {
			return nil, fmt.Errorf("%w: migration object_blob_id must be UUID", ErrInvalidBackupArtifact)
		}
		incidentID, err := uuid.Parse(incidentIDRaw)
		if err != nil {
			return nil, fmt.Errorf("%w: migration incident_id must be UUID", ErrInvalidBackupArtifact)
		}
		item := ObjectStoreMigrationBlob{
			ObjectBlobID: objectBlobID,
			IncidentID:   incidentID,
			StorageKey:   storageKey,
			ByteSize:     byteSize,
		}
		if storageRef.Valid {
			item.EvidenceStorageRef = storageRef.String
		}
		objects = append(objects, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate migration object blobs: %w", err)
	}
	return objects, nil
}

func CopyObjectStoreMigrationObjects(ctx context.Context, params ObjectStoreMigrationCopyParams) (ObjectStoreMigrationCopyLedger, []byte, error) {
	if params.RunID == uuid.Nil {
		return ObjectStoreMigrationCopyLedger{}, nil, fmt.Errorf("%w: migration run_id is required for copy ledger", ErrInvalidBackupArtifact)
	}
	if strings.TrimSpace(params.SourceBucket) == "" || strings.TrimSpace(params.TargetBucket) == "" {
		return ObjectStoreMigrationCopyLedger{}, nil, fmt.Errorf("%w: migration source and target buckets are required", ErrInvalidBackupArtifact)
	}
	if params.SourceStore == nil || params.TargetStore == nil {
		return ObjectStoreMigrationCopyLedger{}, nil, fmt.Errorf("%w: migration source and target object stores are required", ErrInvalidBackupArtifact)
	}
	ledger := ObjectStoreMigrationCopyLedger{
		SchemaID:        ObjectStoreMigrationCopyLedgerSchemaID,
		RunID:           params.RunID.String(),
		SourceBackend:   normalizeBackendLabel(params.SourceBackend),
		TargetBackend:   normalizeBackendLabel(params.TargetBackend),
		SourceBucketRef: hashRedactionRef("bucket", params.SourceBucket),
		TargetBucketRef: hashRedactionRef("bucket", params.TargetBucket),
		StatusCounts:    map[string]int{},
		Items:           []ObjectStoreMigrationCopyLedgerItem{},
	}
	objects := append([]ObjectStoreMigrationBlob(nil), params.Objects...)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].ObjectBlobID.String() < objects[j].ObjectBlobID.String()
	})
	for index, object := range objects {
		item := copyObjectForMigration(ctx, params, object, index+1)
		ledger.Items = append(ledger.Items, item)
		ledger.StatusCounts[string(item.Status)]++
	}
	ledger.ObjectCount = len(ledger.Items)
	if copyLedgerHasBlockingStatus(ledger) {
		ledger.Result = "fail"
	} else {
		ledger.Result = "pass"
	}
	body, err := EncodeObjectStoreMigrationCopyLedger(ledger)
	if err != nil {
		return ObjectStoreMigrationCopyLedger{}, nil, err
	}
	ledger.ArtifactSHA256 = sha256Hex(canonicalObjectStoreMigrationCopyLedgerBytes(ledger, false))
	return ledger, body, nil
}

func ValidateObjectStoreMigration(ctx context.Context, params ObjectStoreMigrationValidationParams) (ObjectStoreMigrationValidation, []byte, error) {
	if params.RunID == uuid.Nil {
		return ObjectStoreMigrationValidation{}, nil, fmt.Errorf("%w: migration run_id is required for validation", ErrInvalidBackupArtifact)
	}
	completed := backupTimestamp(params.CompletedAt)
	artifact := ObjectStoreMigrationValidation{
		SchemaID:              ObjectStoreMigrationValidationSchemaID,
		SchemaVersion:         ObjectStoreMigrationValidationSchemaVersion,
		ValidationToolVersion: ObjectStoreMigrationToolVersion,
		RunID:                 params.RunID.String(),
		StartedAt:             backupTimestamp(params.StartedAt),
		CompletedAt:           &completed,
		SourceBackend:         normalizeBackendLabel(params.SourceBackend),
		TargetBackend:         normalizeBackendLabel(params.TargetBackend),
		SourceBucket:          strings.TrimSpace(params.SourceBucket),
		TargetBucket:          strings.TrimSpace(params.TargetBucket),
		ObjectsChecked:        []ObjectStoreMigrationValidationObject{},
		PreviewSampleChecks:   []ObjectStoreMigrationPreviewSampleCheck{},
		BlockingDiagnostics:   []ObjectStoreMigrationDiagnostic{},
		NonblockingWarnings:   []ObjectStoreMigrationDiagnostic{},
	}
	objects := append([]ObjectStoreMigrationBlob(nil), params.Objects...)
	sort.Slice(objects, func(i, j int) bool {
		return objects[i].ObjectBlobID.String() < objects[j].ObjectBlobID.String()
	})
	incidentIDs := map[string]struct{}{}
	sourceDigest := sha256.New()
	targetDigest := sha256.New()
	for _, object := range objects {
		incidentIDs[object.IncidentID.String()] = struct{}{}
		item := validateObjectForMigration(ctx, params, object)
		artifact.ObjectsChecked = append(artifact.ObjectsChecked, item)
		_, _ = sourceDigest.Write([]byte(object.ObjectBlobID.String() + ":" + item.SourceSHA256 + "\n"))
		_, _ = targetDigest.Write([]byte(object.ObjectBlobID.String() + ":" + item.TargetSHA256 + "\n"))
		if item.Status != ObjectStoreMigrationValidationPass {
			artifact.BlockingDiagnostics = append(artifact.BlockingDiagnostics, migrationValidationDiagnostic(len(artifact.BlockingDiagnostics)+1, item))
		}
	}
	artifact.IncidentCount = len(incidentIDs)
	artifact.ObjectBlobCount = len(artifact.ObjectsChecked)
	artifact.SourceSnapshotID = hex.EncodeToString(sourceDigest.Sum(nil))
	artifact.TargetSnapshotID = hex.EncodeToString(targetDigest.Sum(nil))
	if len(artifact.PreviewSampleChecks) == 0 {
		artifact.NonblockingWarnings = append(artifact.NonblockingWarnings, ObjectStoreMigrationDiagnostic{
			DiagnosticID: "warning-0001",
			Severity:     "warning",
			ReasonCode:   "zero_preview_candidates",
			ObjectBlobID: nil,
			IncidentID:   nil,
			Message:      "No eligible preview or download candidates were available to the migration validator.",
			Refs:         []RedactionRef{},
		})
	}
	artifact.Result = ComputeObjectStoreMigrationValidationResult(artifact)
	body, err := EncodeObjectStoreMigrationValidation(artifact)
	if err != nil {
		return ObjectStoreMigrationValidation{}, nil, err
	}
	artifact.ArtifactSHA256 = sha256Hex(canonicalObjectStoreMigrationValidationBytes(artifact, false))
	return artifact, body, nil
}

func ProbeObjectStoreMigrationTarget(ctx context.Context, runID uuid.UUID, targetBucket string, targetStore objectstore.Store, startedAt time.Time) (ObjectStoreMigrationTargetProbe, []byte, error) {
	if runID == uuid.Nil {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("%w: migration run_id is required for target probe", ErrInvalidBackupArtifact)
	}
	typed, ok := targetStore.(objectstore.TypedStore)
	if !ok {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("%w: migration target store must implement typed object-store operations", ErrInvalidBackupArtifact)
	}
	payload := []byte("cartulary object-store migration probe\n")
	probeKey := ".cartulary/migrations/" + runID.String() + "/target-probe.bin"
	if _, err := typed.Put(ctx, objectstore.PutObjectRequest{
		Key:         probeKey,
		Body:        bytes.NewReader(payload),
		Size:        int64(len(payload)),
		ContentType: "application/octet-stream",
		Metadata:    objectstore.Metadata{"cartulary-migration-run-id": runID.String()},
		Purpose:     objectstore.PurposeMigrationCopy,
	}); err != nil {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("migration target probe put: %w", err)
	}
	reader, info, err := typed.Get(ctx, objectstore.GetObjectRequest{Key: probeKey, Purpose: objectstore.PurposeMigrationValidation})
	if err != nil {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("migration target probe get: %w", err)
	}
	body, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("migration target probe read: %w", readErr)
	}
	if closeErr != nil {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("migration target probe close: %w", closeErr)
	}
	if info.Size != int64(len(payload)) || !bytes.Equal(body, payload) {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("%w: migration target probe byte mismatch", ErrInvalidBackupArtifact)
	}
	if err := typed.Delete(ctx, objectstore.DeleteObjectRequest{Key: probeKey, Purpose: objectstore.PurposeMigrationValidation}); err != nil {
		return ObjectStoreMigrationTargetProbe{}, nil, fmt.Errorf("migration target probe cleanup: %w", err)
	}
	completedAt := backupTimestamp(time.Now().UTC())
	if !startedAt.IsZero() {
		startedAt = backupTimestamp(startedAt)
	} else {
		startedAt = completedAt
	}
	probe := ObjectStoreMigrationTargetProbe{
		SchemaID:        ObjectStoreMigrationProbeSchemaID,
		RunID:           runID.String(),
		StartedAt:       startedAt,
		CompletedAt:     completedAt,
		TargetBucketRef: hashRedactionRef("bucket", targetBucket),
		ProbeKeyRef:     hashRedactionRef("object_key", probeKey),
		Result:          "pass",
		SHA256:          sha256Hex(payload),
	}
	bodyEncoded := canonicalObjectStoreMigrationProbeBytes(probe)
	return probe, bodyEncoded, nil
}

func BuildObjectStoreMigrationRollbackEvidence(runID uuid.UUID, createdAt time.Time) (ObjectStoreMigrationRollbackEvidence, []byte, error) {
	if runID == uuid.Nil {
		return ObjectStoreMigrationRollbackEvidence{}, nil, fmt.Errorf("%w: migration run_id is required for rollback evidence", ErrInvalidBackupArtifact)
	}
	artifact := ObjectStoreMigrationRollbackEvidence{
		SchemaID:                       ObjectStoreMigrationRollbackSchemaID,
		RunID:                          runID.String(),
		CreatedAt:                      backupTimestamp(createdAt),
		BeforeCutoverSourceActive:      true,
		BeforeCutoverBackupRetained:    true,
		CutoverRollbackProcedure:       "stop_app_restore_backup_if_needed_point_config_to_source_verify_source_path",
		PostVerificationRollbackClosed: true,
	}
	return artifact, canonicalObjectStoreMigrationRollbackBytes(artifact), nil
}

func EncodeObjectStoreMigrationRun(run ObjectStoreMigrationRun) ([]byte, error) {
	if err := ValidateObjectStoreMigrationRun(run); err != nil {
		return nil, err
	}
	return canonicalObjectStoreMigrationRunBytes(run), nil
}

func DecodeObjectStoreMigrationRun(body []byte) (ObjectStoreMigrationRun, error) {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return ObjectStoreMigrationRun{}, fmt.Errorf("%w: migration run JSON object keys must be unique: %v", ErrInvalidBackupArtifact, err)
	}
	var run ObjectStoreMigrationRun
	if err := decodeStrictJSON(body, &run); err != nil {
		return ObjectStoreMigrationRun{}, fmt.Errorf("%w: decode migration run: %v", ErrInvalidBackupArtifact, err)
	}
	if err := ValidateObjectStoreMigrationRun(run); err != nil {
		return ObjectStoreMigrationRun{}, err
	}
	if !bytes.Equal(body, canonicalObjectStoreMigrationRunBytes(run)) {
		return ObjectStoreMigrationRun{}, fmt.Errorf("%w: migration run is not canonical JSON", ErrInvalidBackupArtifact)
	}
	return run, nil
}

func ValidateObjectStoreMigrationRun(run ObjectStoreMigrationRun) error {
	if run.SchemaID != ObjectStoreMigrationRunSchemaID {
		return fmt.Errorf("%w: unsupported migration run schema %q", ErrInvalidBackupArtifact, run.SchemaID)
	}
	if _, err := uuid.Parse(run.RunID); err != nil {
		return fmt.Errorf("%w: migration run_id must be UUID", ErrInvalidBackupArtifact)
	}
	if run.CreatedAt.IsZero() || run.UpdatedAt.IsZero() {
		return fmt.Errorf("%w: migration run timestamps are required", ErrInvalidBackupArtifact)
	}
	if run.OperatorIdentity == "" {
		return fmt.Errorf("%w: migration operator_identity is required", ErrInvalidBackupArtifact)
	}
	if err := validateRedactionRef(run.SourceEndpointRef, "endpoint"); err != nil {
		return err
	}
	if err := validateRedactionRef(run.TargetEndpointRef, "endpoint"); err != nil {
		return err
	}
	if err := validateRedactionRef(run.SourceBucketRef, "bucket"); err != nil {
		return err
	}
	if err := validateRedactionRef(run.TargetBucketRef, "bucket"); err != nil {
		return err
	}
	if len(run.Events) == 0 || run.Events[0].Event != ObjectStoreMigrationEventPlanCreated {
		return fmt.Errorf("%w: migration run must start with plan_created", ErrInvalidBackupArtifact)
	}
	if _, ok := run.StateTimestamps[string(run.CurrentState)]; !ok {
		return fmt.Errorf("%w: migration run missing current state timestamp", ErrInvalidBackupArtifact)
	}
	if isTerminalMigrationState(run.CurrentState) {
		if run.TerminalResult == nil || *run.TerminalResult != run.CurrentState {
			return fmt.Errorf("%w: terminal migration run must record terminal_result", ErrInvalidBackupArtifact)
		}
	}
	return nil
}

func EncodeObjectStoreMigrationCopyLedger(ledger ObjectStoreMigrationCopyLedger) ([]byte, error) {
	if err := ValidateObjectStoreMigrationCopyLedgerWithoutDigest(ledger); err != nil {
		return nil, err
	}
	ledger.ArtifactSHA256 = sha256Hex(canonicalObjectStoreMigrationCopyLedgerBytes(ledger, false))
	if err := ValidateObjectStoreMigrationCopyLedger(ledger); err != nil {
		return nil, err
	}
	return canonicalObjectStoreMigrationCopyLedgerBytes(ledger, true), nil
}

func DecodeObjectStoreMigrationCopyLedger(body []byte) (ObjectStoreMigrationCopyLedger, error) {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return ObjectStoreMigrationCopyLedger{}, fmt.Errorf("%w: migration copy ledger JSON object keys must be unique: %v", ErrInvalidBackupArtifact, err)
	}
	var ledger ObjectStoreMigrationCopyLedger
	if err := decodeStrictJSON(body, &ledger); err != nil {
		return ObjectStoreMigrationCopyLedger{}, fmt.Errorf("%w: decode migration copy ledger: %v", ErrInvalidBackupArtifact, err)
	}
	if err := ValidateObjectStoreMigrationCopyLedger(ledger); err != nil {
		return ObjectStoreMigrationCopyLedger{}, err
	}
	if !bytes.Equal(body, canonicalObjectStoreMigrationCopyLedgerBytes(ledger, true)) {
		return ObjectStoreMigrationCopyLedger{}, fmt.Errorf("%w: migration copy ledger is not canonical JSON", ErrInvalidBackupArtifact)
	}
	return ledger, nil
}

func ValidateObjectStoreMigrationCopyLedger(ledger ObjectStoreMigrationCopyLedger) error {
	if err := ValidateObjectStoreMigrationCopyLedgerWithoutDigest(ledger); err != nil {
		return err
	}
	if !validSHA256Hex(ledger.ArtifactSHA256) {
		return fmt.Errorf("%w: migration copy ledger artifact_sha256 is required", ErrInvalidBackupArtifact)
	}
	if got := sha256Hex(canonicalObjectStoreMigrationCopyLedgerBytes(ledger, false)); got != ledger.ArtifactSHA256 {
		return fmt.Errorf("%w: migration copy ledger artifact_sha256 mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func ValidateObjectStoreMigrationCopyLedgerWithoutDigest(ledger ObjectStoreMigrationCopyLedger) error {
	if ledger.SchemaID != ObjectStoreMigrationCopyLedgerSchemaID {
		return fmt.Errorf("%w: unsupported migration copy ledger schema %q", ErrInvalidBackupArtifact, ledger.SchemaID)
	}
	if _, err := uuid.Parse(ledger.RunID); err != nil {
		return fmt.Errorf("%w: migration copy ledger run_id must be UUID", ErrInvalidBackupArtifact)
	}
	if ledger.ObjectCount != len(ledger.Items) {
		return fmt.Errorf("%w: migration copy ledger object_count mismatch", ErrInvalidBackupArtifact)
	}
	switch ledger.Result {
	case "pass", "fail":
	default:
		return fmt.Errorf("%w: migration copy ledger result must be pass or fail", ErrInvalidBackupArtifact)
	}
	counts := map[string]int{}
	for index, item := range ledger.Items {
		if item.Sequence != index+1 {
			return fmt.Errorf("%w: migration copy ledger sequence mismatch", ErrInvalidBackupArtifact)
		}
		if item.SourceSizeBytes < 0 {
			return fmt.Errorf("%w: migration copy ledger source_size_bytes is negative", ErrInvalidBackupArtifact)
		}
		if item.SourceSHA256 != "" && !validSHA256Hex(item.SourceSHA256) {
			return fmt.Errorf("%w: migration copy ledger source_sha256 must be lowercase hex", ErrInvalidBackupArtifact)
		}
		if item.TargetSHA256 != "" && !validSHA256Hex(item.TargetSHA256) {
			return fmt.Errorf("%w: migration copy ledger target_sha256 must be lowercase hex", ErrInvalidBackupArtifact)
		}
		if item.ReasonCode == "" {
			return fmt.Errorf("%w: migration copy ledger reason_code is required", ErrInvalidBackupArtifact)
		}
		switch item.Status {
		case ObjectStoreMigrationCopyCopied, ObjectStoreMigrationCopyAlreadyCopied, ObjectStoreMigrationCopyMissingSource, ObjectStoreMigrationCopyTargetMismatch, ObjectStoreMigrationCopyUnsupportedSourceFeature, ObjectStoreMigrationCopyError:
		default:
			return fmt.Errorf("%w: migration copy ledger status is outside closed vocabulary", ErrInvalidBackupArtifact)
		}
		counts[string(item.Status)]++
	}
	for status, count := range counts {
		if ledger.StatusCounts[status] != count {
			return fmt.Errorf("%w: migration copy ledger status_counts mismatch", ErrInvalidBackupArtifact)
		}
	}
	return nil
}

func EncodeObjectStoreMigrationValidation(artifact ObjectStoreMigrationValidation) ([]byte, error) {
	if err := ValidateObjectStoreMigrationValidationWithoutDigest(artifact); err != nil {
		return nil, err
	}
	artifact.ArtifactSHA256 = sha256Hex(canonicalObjectStoreMigrationValidationBytes(artifact, false))
	if err := ValidateObjectStoreMigrationValidation(artifact); err != nil {
		return nil, err
	}
	return canonicalObjectStoreMigrationValidationBytes(artifact, true), nil
}

func DecodeObjectStoreMigrationValidation(body []byte) (ObjectStoreMigrationValidation, error) {
	if err := rejectDuplicateJSONKeys(body); err != nil {
		return ObjectStoreMigrationValidation{}, fmt.Errorf("%w: migration validation JSON object keys must be unique: %v", ErrInvalidBackupArtifact, err)
	}
	var artifact ObjectStoreMigrationValidation
	if err := decodeStrictJSON(body, &artifact); err != nil {
		return ObjectStoreMigrationValidation{}, fmt.Errorf("%w: decode migration validation: %v", ErrInvalidBackupArtifact, err)
	}
	if err := ValidateObjectStoreMigrationValidation(artifact); err != nil {
		return ObjectStoreMigrationValidation{}, err
	}
	if !bytes.Equal(body, canonicalObjectStoreMigrationValidationBytes(artifact, true)) {
		return ObjectStoreMigrationValidation{}, fmt.Errorf("%w: migration validation is not canonical JSON", ErrInvalidBackupArtifact)
	}
	return artifact, nil
}

func ValidateObjectStoreMigrationValidation(artifact ObjectStoreMigrationValidation) error {
	if err := ValidateObjectStoreMigrationValidationWithoutDigest(artifact); err != nil {
		return err
	}
	if !validSHA256Hex(artifact.ArtifactSHA256) {
		return fmt.Errorf("%w: migration validation artifact_sha256 is required", ErrInvalidBackupArtifact)
	}
	if got := sha256Hex(canonicalObjectStoreMigrationValidationBytes(artifact, false)); got != artifact.ArtifactSHA256 {
		return fmt.Errorf("%w: migration validation artifact_sha256 mismatch", ErrInvalidBackupArtifact)
	}
	return nil
}

func ValidateObjectStoreMigrationValidationWithoutDigest(artifact ObjectStoreMigrationValidation) error {
	if artifact.SchemaID != ObjectStoreMigrationValidationSchemaID {
		return fmt.Errorf("%w: unsupported migration validation schema %q", ErrInvalidBackupArtifact, artifact.SchemaID)
	}
	if artifact.SchemaVersion != ObjectStoreMigrationValidationSchemaVersion {
		return fmt.Errorf("%w: migration validation schema_version must be %s", ErrInvalidBackupArtifact, ObjectStoreMigrationValidationSchemaVersion)
	}
	if artifact.ValidationToolVersion != ObjectStoreMigrationToolVersion {
		return fmt.Errorf("%w: migration validation_tool_version must be %s", ErrInvalidBackupArtifact, ObjectStoreMigrationToolVersion)
	}
	if _, err := uuid.Parse(artifact.RunID); err != nil {
		return fmt.Errorf("%w: migration validation run_id must be UUID", ErrInvalidBackupArtifact)
	}
	if artifact.StartedAt.IsZero() || artifact.CompletedAt == nil || artifact.CompletedAt.IsZero() {
		return fmt.Errorf("%w: migration validation timestamps are required", ErrInvalidBackupArtifact)
	}
	if artifact.SourceBackend == "" || artifact.TargetBackend == "" || artifact.SourceBucket == "" || artifact.TargetBucket == "" {
		return fmt.Errorf("%w: migration validation source/target backend and bucket fields are required", ErrInvalidBackupArtifact)
	}
	if artifact.ObjectBlobCount != len(artifact.ObjectsChecked) {
		return fmt.Errorf("%w: migration validation object_blob_count mismatch", ErrInvalidBackupArtifact)
	}
	previous := ""
	for _, item := range artifact.ObjectsChecked {
		if item.ObjectBlobID == "" || item.IncidentID == "" {
			return fmt.Errorf("%w: migration validation object IDs are required", ErrInvalidBackupArtifact)
		}
		if previous != "" && previous >= item.ObjectBlobID {
			return fmt.Errorf("%w: migration validation objects_checked not sorted by object_blob_id", ErrInvalidBackupArtifact)
		}
		previous = item.ObjectBlobID
		if _, err := uuid.Parse(item.ObjectBlobID); err != nil {
			return fmt.Errorf("%w: migration validation object_blob_id must be UUID", ErrInvalidBackupArtifact)
		}
		if _, err := uuid.Parse(item.IncidentID); err != nil {
			return fmt.Errorf("%w: migration validation incident_id must be UUID", ErrInvalidBackupArtifact)
		}
		if !validSHA256Hex(item.StorageRefSHA256) {
			return fmt.Errorf("%w: migration validation storage_ref_sha256 is required", ErrInvalidBackupArtifact)
		}
		switch item.Status {
		case ObjectStoreMigrationValidationPass, ObjectStoreMigrationValidationMissingSource, ObjectStoreMigrationValidationMissingTarget, ObjectStoreMigrationValidationSizeMismatch, ObjectStoreMigrationValidationHashMismatch, ObjectStoreMigrationValidationUnsupportedSourceFeature, ObjectStoreMigrationValidationError:
		default:
			return fmt.Errorf("%w: migration validation status is outside closed vocabulary", ErrInvalidBackupArtifact)
		}
		if item.Status == ObjectStoreMigrationValidationPass && item.ReasonCode != "" {
			return fmt.Errorf("%w: passing migration validation object cannot include reason_code", ErrInvalidBackupArtifact)
		}
		if item.Status != ObjectStoreMigrationValidationPass && item.ReasonCode == "" {
			return fmt.Errorf("%w: failing migration validation object requires reason_code", ErrInvalidBackupArtifact)
		}
	}
	for _, diagnostic := range append(append([]ObjectStoreMigrationDiagnostic{}, artifact.BlockingDiagnostics...), artifact.NonblockingWarnings...) {
		if diagnostic.DiagnosticID == "" || diagnostic.Severity == "" || diagnostic.ReasonCode == "" || diagnostic.Message == "" {
			return fmt.Errorf("%w: migration validation diagnostic fields are required", ErrInvalidBackupArtifact)
		}
	}
	if want := ComputeObjectStoreMigrationValidationResult(artifact); artifact.Result != want {
		return fmt.Errorf("%w: migration validation result got %s want %s", ErrInvalidBackupArtifact, artifact.Result, want)
	}
	return nil
}

func ComputeObjectStoreMigrationValidationResult(artifact ObjectStoreMigrationValidation) string {
	for _, item := range artifact.ObjectsChecked {
		if item.Status != ObjectStoreMigrationValidationPass {
			return "fail"
		}
	}
	if len(artifact.BlockingDiagnostics) != 0 {
		return "fail"
	}
	for _, sample := range artifact.PreviewSampleChecks {
		if sample.Status != "pass" {
			return "fail"
		}
	}
	return "pass"
}

func ArtifactRefForBody(key string, body []byte, contentType string) ObjectStoreMigrationArtifactRef {
	if strings.TrimSpace(contentType) == "" {
		contentType = "application/json"
	}
	return ObjectStoreMigrationArtifactRef{
		Key:         key,
		SHA256:      sha256Hex(body),
		SizeBytes:   int64(len(body)),
		ContentType: contentType,
	}
}

func migrationDestinationState(current ObjectStoreMigrationState, event ObjectStoreMigrationEventName) (ObjectStoreMigrationState, error) {
	switch event {
	case ObjectStoreMigrationEventPlanCreated:
		if current == "" {
			return ObjectStoreMigrationStatePlanned, nil
		}
	case ObjectStoreMigrationEventPreflightPassed:
		if current == ObjectStoreMigrationStatePlanned {
			return ObjectStoreMigrationStatePreflighted, nil
		}
	case ObjectStoreMigrationEventWriteQuiescenceVerified:
		if current == ObjectStoreMigrationStatePreflighted {
			return ObjectStoreMigrationStateApplicationStopped, nil
		}
	case ObjectStoreMigrationEventBackupCaptured:
		if current == ObjectStoreMigrationStateApplicationStopped {
			return ObjectStoreMigrationStateBackupCaptured, nil
		}
	case ObjectStoreMigrationEventTargetPrepared:
		if current == ObjectStoreMigrationStateBackupCaptured {
			return ObjectStoreMigrationStateTargetPrepared, nil
		}
	case ObjectStoreMigrationEventCopyStarted:
		if current == ObjectStoreMigrationStateTargetPrepared {
			return ObjectStoreMigrationStateCopying, nil
		}
	case ObjectStoreMigrationEventCopyCompleted:
		if current == ObjectStoreMigrationStateCopying {
			return ObjectStoreMigrationStateCopied, nil
		}
	case ObjectStoreMigrationEventValidationStarted:
		if current == ObjectStoreMigrationStateCopied {
			return ObjectStoreMigrationStateValidating, nil
		}
	case ObjectStoreMigrationEventValidationPassed:
		if current == ObjectStoreMigrationStateValidating {
			return ObjectStoreMigrationStateCutoverReady, nil
		}
	case ObjectStoreMigrationEventCutoverCommitted:
		if current == ObjectStoreMigrationStateCutoverReady {
			return ObjectStoreMigrationStateCutoverCommitted, nil
		}
	case ObjectStoreMigrationEventPostCutoverVerified:
		if current == ObjectStoreMigrationStateCutoverCommitted {
			return ObjectStoreMigrationStatePostCutoverVerified, nil
		}
	case ObjectStoreMigrationEventRollbackRequested:
		switch current {
		case ObjectStoreMigrationStateBackupCaptured, ObjectStoreMigrationStateTargetPrepared, ObjectStoreMigrationStateCopying, ObjectStoreMigrationStateCopied, ObjectStoreMigrationStateValidating, ObjectStoreMigrationStateCutoverReady, ObjectStoreMigrationStateCutoverCommitted:
			return ObjectStoreMigrationStateRolledBack, nil
		}
	case ObjectStoreMigrationEventBlockingFailure:
		if current != "" {
			return ObjectStoreMigrationStateFailed, nil
		}
	}
	return "", fmt.Errorf("%w: event %s is not allowed from state %s", ErrInvalidBackupArtifact, event, current)
}

func isTerminalMigrationState(state ObjectStoreMigrationState) bool {
	switch state {
	case ObjectStoreMigrationStatePostCutoverVerified, ObjectStoreMigrationStateRolledBack, ObjectStoreMigrationStateFailed:
		return true
	default:
		return false
	}
}

func copyObjectForMigration(ctx context.Context, params ObjectStoreMigrationCopyParams, object ObjectStoreMigrationBlob, sequence int) ObjectStoreMigrationCopyLedgerItem {
	item := ObjectStoreMigrationCopyLedgerItem{
		Sequence:        sequence,
		ObjectBlobID:    object.ObjectBlobID.String(),
		SourceBucketRef: hashRedactionRef("bucket", params.SourceBucket),
		SourceKeyRef:    hashRedactionRef("object_key", object.StorageKey),
		TargetBucketRef: hashRedactionRef("bucket", params.TargetBucket),
		TargetKeyRef:    hashRedactionRef("object_key", object.StorageKey),
		Status:          ObjectStoreMigrationCopyError,
		ReasonCode:      "error",
	}
	sourceBody, sourceInfo, sourceSHA, err := readMigrationObject(ctx, params.SourceStore, object.StorageKey, objectstore.PurposeMigrationCopy)
	if err != nil {
		if objectstore.IsObjectNotFound(err) {
			item.Status = ObjectStoreMigrationCopyMissingSource
			item.ReasonCode = "missing_source_object"
			return item
		}
		item.ReasonCode = "source_read_error"
		return item
	}
	item.SourceSizeBytes = int64(len(sourceBody))
	item.SourceSHA256 = sourceSHA
	item.IdempotencyKeySHA256 = migrationCopyIdempotencyKeySHA256(params.SourceBucket, object.StorageKey, params.TargetBucket, object.StorageKey, item.SourceSizeBytes, sourceSHA)
	if sourceInfo.Size != int64(len(sourceBody)) {
		item.Status = ObjectStoreMigrationCopyError
		item.ReasonCode = "source_size_mismatch"
		return item
	}
	targetBody, targetInfo, targetSHA, err := readMigrationObject(ctx, params.TargetStore, object.StorageKey, objectstore.PurposeMigrationValidation)
	if err == nil {
		targetSize := int64(len(targetBody))
		item.TargetSizeBytes = &targetSize
		item.TargetSHA256 = targetSHA
		if targetInfo.Size == item.SourceSizeBytes && targetSize == item.SourceSizeBytes && targetSHA == sourceSHA {
			item.Status = ObjectStoreMigrationCopyAlreadyCopied
			item.ReasonCode = "target_equivalent"
			return item
		}
		item.Status = ObjectStoreMigrationCopyTargetMismatch
		item.ReasonCode = "target_mismatch"
		return item
	}
	if !objectstore.IsObjectNotFound(err) {
		item.Status = ObjectStoreMigrationCopyError
		item.ReasonCode = "target_read_error"
		return item
	}
	typed, ok := params.TargetStore.(objectstore.TypedStore)
	if !ok {
		item.Status = ObjectStoreMigrationCopyError
		item.ReasonCode = "target_store_untyped"
		return item
	}
	_, err = typed.Put(ctx, objectstore.PutObjectRequest{
		Key:         object.StorageKey,
		Body:        bytes.NewReader(sourceBody),
		Size:        int64(len(sourceBody)),
		ContentType: sourceInfo.ContentType,
		Metadata: objectstore.Metadata{
			"cartulary-object-blob-id":   object.ObjectBlobID.String(),
			"cartulary-migration-run-id": params.RunID.String(),
		},
		Purpose: objectstore.PurposeMigrationCopy,
	})
	if err != nil {
		item.Status = ObjectStoreMigrationCopyError
		item.ReasonCode = "target_write_error"
		return item
	}
	targetBody, _, targetSHA, err = readMigrationObject(ctx, params.TargetStore, object.StorageKey, objectstore.PurposeMigrationValidation)
	if err != nil {
		item.Status = ObjectStoreMigrationCopyError
		item.ReasonCode = "target_verify_error"
		return item
	}
	targetSize := int64(len(targetBody))
	item.TargetSizeBytes = &targetSize
	item.TargetSHA256 = targetSHA
	if targetSize != item.SourceSizeBytes || targetSHA != sourceSHA {
		item.Status = ObjectStoreMigrationCopyTargetMismatch
		item.ReasonCode = "target_mismatch"
		return item
	}
	item.Status = ObjectStoreMigrationCopyCopied
	item.ReasonCode = "copied"
	return item
}

func validateObjectForMigration(ctx context.Context, params ObjectStoreMigrationValidationParams, object ObjectStoreMigrationBlob) ObjectStoreMigrationValidationObject {
	item := ObjectStoreMigrationValidationObject{
		ObjectBlobID:     object.ObjectBlobID.String(),
		IncidentID:       object.IncidentID.String(),
		StorageRefSHA256: sha256Hex([]byte(migrationStorageRefForHash(object))),
		Status:           ObjectStoreMigrationValidationError,
		ReasonCode:       "artifact_schema_invalid",
	}
	sourceBody, _, sourceSHA, sourceErr := readMigrationObject(ctx, params.SourceStore, object.StorageKey, objectstore.PurposeMigrationValidation)
	if sourceErr != nil {
		if objectstore.IsObjectNotFound(sourceErr) {
			item.Status = ObjectStoreMigrationValidationMissingSource
			item.ReasonCode = "missing_source_object"
			return item
		}
		item.Status = ObjectStoreMigrationValidationError
		item.ReasonCode = "artifact_schema_invalid"
		return item
	}
	sourceSize := int64(len(sourceBody))
	item.SourceSizeBytes = &sourceSize
	item.SourceSHA256 = sourceSHA
	targetBody, _, targetSHA, targetErr := readMigrationObject(ctx, params.TargetStore, object.StorageKey, objectstore.PurposeMigrationValidation)
	if targetErr != nil {
		if objectstore.IsObjectNotFound(targetErr) {
			item.Status = ObjectStoreMigrationValidationMissingTarget
			item.ReasonCode = "missing_target_object"
			return item
		}
		item.Status = ObjectStoreMigrationValidationError
		item.ReasonCode = "artifact_schema_invalid"
		return item
	}
	targetSize := int64(len(targetBody))
	item.TargetSizeBytes = &targetSize
	item.TargetSHA256 = targetSHA
	switch {
	case sourceSize != targetSize:
		item.Status = ObjectStoreMigrationValidationSizeMismatch
		item.ReasonCode = "size_mismatch"
	case sourceSHA != targetSHA:
		item.Status = ObjectStoreMigrationValidationHashMismatch
		item.ReasonCode = "hash_mismatch"
	default:
		item.Status = ObjectStoreMigrationValidationPass
		item.ReasonCode = ""
	}
	return item
}

func readMigrationObject(ctx context.Context, store objectstore.Store, key string, purpose objectstore.Purpose) ([]byte, objectstore.ObjectInfo, string, error) {
	typed, ok := store.(objectstore.TypedStore)
	var reader io.ReadCloser
	var info objectstore.ObjectInfo
	var err error
	if ok {
		reader, info, err = typed.Get(ctx, objectstore.GetObjectRequest{Key: key, Purpose: purpose})
	} else {
		reader, info, err = store.ReadObject(ctx, key, objectstore.ReadOptions{})
	}
	if err != nil {
		return nil, objectstore.ObjectInfo{}, "", err
	}
	body, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil {
		return nil, objectstore.ObjectInfo{}, "", readErr
	}
	if closeErr != nil {
		return nil, objectstore.ObjectInfo{}, "", closeErr
	}
	return body, info, sha256Hex(body), nil
}

func migrationValidationDiagnostic(sequence int, item ObjectStoreMigrationValidationObject) ObjectStoreMigrationDiagnostic {
	objectBlobID := item.ObjectBlobID
	incidentID := item.IncidentID
	return ObjectStoreMigrationDiagnostic{
		DiagnosticID: fmt.Sprintf("blocking-%04d", sequence),
		Severity:     "blocking",
		ReasonCode:   item.ReasonCode,
		ObjectBlobID: &objectBlobID,
		IncidentID:   &incidentID,
		Message:      "Migration validation detected an object-store byte proof failure.",
		Refs:         []RedactionRef{},
	}
}

func copyLedgerHasBlockingStatus(ledger ObjectStoreMigrationCopyLedger) bool {
	for _, item := range ledger.Items {
		switch item.Status {
		case ObjectStoreMigrationCopyCopied, ObjectStoreMigrationCopyAlreadyCopied:
			continue
		default:
			return true
		}
	}
	return false
}

func migrationStorageRefForHash(object ObjectStoreMigrationBlob) string {
	if strings.TrimSpace(object.EvidenceStorageRef) != "" {
		return object.EvidenceStorageRef
	}
	return object.StorageKey
}

func migrationCopyIdempotencyKeySHA256(sourceBucket string, sourceKey string, targetBucket string, targetKey string, sourceSize int64, sourceSHA string) string {
	value := strings.Join([]string{
		sourceBucket,
		sourceKey,
		targetBucket,
		targetKey,
		fmt.Sprintf("%d", sourceSize),
		sourceSHA,
	}, "\x00")
	return sha256Hex([]byte(value))
}

func migrationArtifactRefFromProof(proof BackupArtifactProof) ObjectStoreMigrationArtifactRef {
	return ObjectStoreMigrationArtifactRef{
		Key:         proof.Key,
		SHA256:      proof.SHA256,
		SizeBytes:   proof.SizeBytes,
		ContentType: proof.ContentType,
	}
}

func migrationArtifactRefPtr(proof *BackupArtifactProof) *ObjectStoreMigrationArtifactRef {
	if proof == nil {
		return nil
	}
	ref := migrationArtifactRefFromProof(*proof)
	return &ref
}

func normalizeBackendLabel(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "s3_compatible"
	}
	return value
}

func endpointRedactionRef(endpoint string) RedactionRef {
	value := strings.TrimSpace(endpoint)
	ref := hashRedactionRef("endpoint", value)
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") {
		parts := strings.SplitN(value, "://", 2)
		ref.Scheme = parts[0]
	}
	return ref
}

func canonicalObjectStoreMigrationRunBytes(run ObjectStoreMigrationRun) []byte {
	events := make([]any, 0, len(run.Events))
	for _, event := range run.Events {
		item := map[string]any{
			"event":       event.Event,
			"from_state":  event.FromState,
			"occurred_at": event.OccurredAt,
			"sequence":    event.Sequence,
			"to_state":    event.ToState,
		}
		if len(event.Detail) > 0 {
			item["detail"] = event.Detail
		}
		events = append(events, item)
	}
	backupRefs := make([]any, 0, len(run.BackupRefs))
	for _, ref := range run.BackupRefs {
		item := map[string]any{
			"backup_set_id":         ref.BackupSetID,
			"integrity_manifest":    migrationArtifactRefCanonicalMap(ref.IntegrityManifest),
			"object_store_artifact": migrationArtifactRefCanonicalMap(ref.ObjectStoreArtifact),
			"postgres_artifact":     migrationArtifactRefCanonicalMap(ref.PostgresArtifact),
		}
		if ref.ObjectStoreBackupManifestArtifact != nil {
			item["object_store_backup_manifest_artifact"] = migrationArtifactRefCanonicalMap(*ref.ObjectStoreBackupManifestArtifact)
		}
		if ref.ObjectStoreBackupSummaryArtifact != nil {
			item["object_store_backup_summary_artifact"] = migrationArtifactRefCanonicalMap(*ref.ObjectStoreBackupSummaryArtifact)
		}
		backupRefs = append(backupRefs, item)
	}
	return marshalCanonical(map[string]any{
		"backup_refs":         backupRefs,
		"copy_ledger_ref":     migrationArtifactRefPtrCanonicalMap(run.CopyLedgerRef),
		"created_at":          run.CreatedAt,
		"current_state":       run.CurrentState,
		"events":              events,
		"operator_identity":   run.OperatorIdentity,
		"probe_ref":           migrationArtifactRefPtrCanonicalMap(run.ProbeRef),
		"rollback_ref":        migrationArtifactRefPtrCanonicalMap(run.RollbackRef),
		"run_id":              run.RunID,
		"schema_id":           run.SchemaID,
		"source_bucket_ref":   redactionRefCanonicalMap(run.SourceBucketRef),
		"source_endpoint_ref": redactionRefCanonicalMap(run.SourceEndpointRef),
		"state_timestamps":    run.StateTimestamps,
		"target_bucket_ref":   redactionRefCanonicalMap(run.TargetBucketRef),
		"target_endpoint_ref": redactionRefCanonicalMap(run.TargetEndpointRef),
		"terminal_result":     run.TerminalResult,
		"updated_at":          run.UpdatedAt,
		"validation_ref":      migrationArtifactRefPtrCanonicalMap(run.ValidationRef),
	})
}

func canonicalObjectStoreMigrationCopyLedgerBytes(ledger ObjectStoreMigrationCopyLedger, includeDigest bool) []byte {
	items := make([]any, 0, len(ledger.Items))
	for _, item := range ledger.Items {
		value := map[string]any{
			"idempotency_key_sha256": item.IdempotencyKeySHA256,
			"object_blob_id":         item.ObjectBlobID,
			"reason_code":            item.ReasonCode,
			"sequence":               item.Sequence,
			"source_bucket_ref":      redactionRefCanonicalMap(item.SourceBucketRef),
			"source_key_ref":         redactionRefCanonicalMap(item.SourceKeyRef),
			"source_sha256":          item.SourceSHA256,
			"source_size_bytes":      item.SourceSizeBytes,
			"status":                 item.Status,
			"target_bucket_ref":      redactionRefCanonicalMap(item.TargetBucketRef),
			"target_key_ref":         redactionRefCanonicalMap(item.TargetKeyRef),
		}
		if item.TargetSizeBytes != nil {
			value["target_size_bytes"] = *item.TargetSizeBytes
		}
		if item.TargetSHA256 != "" {
			value["target_sha256"] = item.TargetSHA256
		}
		items = append(items, value)
	}
	value := map[string]any{
		"items":             items,
		"object_count":      ledger.ObjectCount,
		"result":            ledger.Result,
		"run_id":            ledger.RunID,
		"schema_id":         ledger.SchemaID,
		"source_backend":    ledger.SourceBackend,
		"source_bucket_ref": redactionRefCanonicalMap(ledger.SourceBucketRef),
		"status_counts":     ledger.StatusCounts,
		"target_backend":    ledger.TargetBackend,
		"target_bucket_ref": redactionRefCanonicalMap(ledger.TargetBucketRef),
	}
	if includeDigest {
		value["artifact_sha256"] = ledger.ArtifactSHA256
	}
	return marshalCanonical(value)
}

func canonicalObjectStoreMigrationValidationBytes(artifact ObjectStoreMigrationValidation, includeDigest bool) []byte {
	objects := make([]any, 0, len(artifact.ObjectsChecked))
	for _, item := range artifact.ObjectsChecked {
		value := map[string]any{
			"incident_id":        item.IncidentID,
			"object_blob_id":     item.ObjectBlobID,
			"reason_code":        item.ReasonCode,
			"status":             item.Status,
			"storage_ref_sha256": item.StorageRefSHA256,
			"source_sha256":      item.SourceSHA256,
			"target_sha256":      item.TargetSHA256,
		}
		if item.SourceSizeBytes != nil {
			value["source_size_bytes"] = *item.SourceSizeBytes
		}
		if item.TargetSizeBytes != nil {
			value["target_size_bytes"] = *item.TargetSizeBytes
		}
		objects = append(objects, value)
	}
	samples := make([]any, 0, len(artifact.PreviewSampleChecks))
	for _, sample := range artifact.PreviewSampleChecks {
		samples = append(samples, map[string]any{
			"incident_id":    sample.IncidentID,
			"object_blob_id": sample.ObjectBlobID,
			"reason_code":    sample.ReasonCode,
			"route_class":    sample.RouteClass,
			"status":         sample.Status,
		})
	}
	value := map[string]any{
		"blocking_diagnostics":    migrationDiagnosticsCanonical(artifact.BlockingDiagnostics),
		"completed_at":            artifact.CompletedAt,
		"incident_count":          artifact.IncidentCount,
		"nonblocking_warnings":    migrationDiagnosticsCanonical(artifact.NonblockingWarnings),
		"object_blob_count":       artifact.ObjectBlobCount,
		"objects_checked":         objects,
		"preview_sample_checks":   samples,
		"result":                  artifact.Result,
		"run_id":                  artifact.RunID,
		"schema_id":               artifact.SchemaID,
		"schema_version":          artifact.SchemaVersion,
		"source_backend":          artifact.SourceBackend,
		"source_bucket":           artifact.SourceBucket,
		"source_snapshot_id":      artifact.SourceSnapshotID,
		"started_at":              artifact.StartedAt,
		"target_backend":          artifact.TargetBackend,
		"target_bucket":           artifact.TargetBucket,
		"target_snapshot_id":      artifact.TargetSnapshotID,
		"validation_tool_version": artifact.ValidationToolVersion,
	}
	if includeDigest {
		value["artifact_sha256"] = artifact.ArtifactSHA256
	}
	return marshalCanonical(value)
}

func canonicalObjectStoreMigrationProbeBytes(probe ObjectStoreMigrationTargetProbe) []byte {
	return marshalCanonical(map[string]any{
		"completed_at":      probe.CompletedAt,
		"probe_key_ref":     redactionRefCanonicalMap(probe.ProbeKeyRef),
		"result":            probe.Result,
		"run_id":            probe.RunID,
		"schema_id":         probe.SchemaID,
		"sha256":            probe.SHA256,
		"started_at":        probe.StartedAt,
		"target_bucket_ref": redactionRefCanonicalMap(probe.TargetBucketRef),
	})
}

func canonicalObjectStoreMigrationRollbackBytes(artifact ObjectStoreMigrationRollbackEvidence) []byte {
	return marshalCanonical(map[string]any{
		"before_cutover_backup_retained":    artifact.BeforeCutoverBackupRetained,
		"before_cutover_source_active":      artifact.BeforeCutoverSourceActive,
		"created_at":                        artifact.CreatedAt,
		"cutover_rollback_procedure":        artifact.CutoverRollbackProcedure,
		"post_verification_rollback_closed": artifact.PostVerificationRollbackClosed,
		"run_id":                            artifact.RunID,
		"schema_id":                         artifact.SchemaID,
	})
}

func migrationDiagnosticsCanonical(diagnostics []ObjectStoreMigrationDiagnostic) []any {
	values := make([]any, 0, len(diagnostics))
	for _, diagnostic := range diagnostics {
		refs := make([]any, 0, len(diagnostic.Refs))
		for _, ref := range diagnostic.Refs {
			refs = append(refs, redactionRefCanonicalMap(ref))
		}
		values = append(values, map[string]any{
			"diagnostic_id":  diagnostic.DiagnosticID,
			"incident_id":    diagnostic.IncidentID,
			"message":        diagnostic.Message,
			"object_blob_id": diagnostic.ObjectBlobID,
			"reason_code":    diagnostic.ReasonCode,
			"refs":           refs,
			"severity":       diagnostic.Severity,
		})
	}
	return values
}

func migrationArtifactRefCanonicalMap(ref ObjectStoreMigrationArtifactRef) map[string]any {
	return map[string]any{
		"content_type": ref.ContentType,
		"key":          ref.Key,
		"sha256":       ref.SHA256,
		"size_bytes":   ref.SizeBytes,
	}
}

func migrationArtifactRefPtrCanonicalMap(ref *ObjectStoreMigrationArtifactRef) any {
	if ref == nil {
		return nil
	}
	return migrationArtifactRefCanonicalMap(*ref)
}
