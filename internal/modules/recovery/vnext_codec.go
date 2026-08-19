package recovery

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

const (
	PostgresSnapshotArtifactV2SchemaID                    = "cartulary.postgres_snapshot_artifact.v2"
	PostgresSnapshotUnitV1SchemaID                        = "cartulary.postgres_snapshot_unit.v1"
	ObjectStoreBackupManifestV2SchemaID                   = "cartulary.object_store_backup_manifest.v2"
	ObjectStoreBackupSummaryV2SchemaID                    = "cartulary.object_store_backup_summary.v2"
	BackupIntegrityManifestV3SchemaID                     = "cartulary.backup_integrity_manifest.v3"
	GraphProjectionRestoreSourceRegistryV2SchemaID        = "cartulary.graph_projection_restore_source_registry.v2"
	GraphProjectionRestoreImplementationBindingV2SchemaID = "cartulary.graph_projection_restore_implementation_binding.v2"
	GraphProjectionRestoreSourceRegistryV3SchemaID        = "cartulary.graph_projection_restore_source_registry.v3"
	GraphProjectionRestoreImplementationBindingV3SchemaID = "cartulary.graph_projection_restore_implementation_binding.v3"
	VNextTransactionIsolation                             = "repeatable_read_read_only"
	vNextCodecRegistryDomain                              = "CARTULARY-RECOVERY-CODEC-REGISTRY-VNEXT\n"
	vNextPostgresSnapshotDigestDomain                     = "CARTULARY-POSTGRES-SNAPSHOT-ARTIFACT-V2\n"
	vNextObjectManifestDigestDomain                       = "CARTULARY-OBJECT-STORE-BACKUP-MANIFEST-V2\n"
	vNextIntegrityManifestDigestDomain                    = "CARTULARY-BACKUP-INTEGRITY-MANIFEST-V3\n"
	vNextNDJSONContentType                                = "application/x-ndjson"
	vNextJSONContentType                                  = "application/json"
	VNextMetadataArtifactScheme                           = "backup-stream-v2://"
)

var ErrVNextBackup = errors.New("recovery: invalid vNext backup")

// VNextSnapshot is a capability scoped to one repository-owned, read-only,
// repeatable-read transaction. Implementations must invoke visit once per row
// in canonical row order and must not retain the capability after the callback.
type VNextSnapshot interface {
	StreamCanonicalTableRows(
		ctx context.Context,
		tableName string,
		visit func(json.RawMessage) error,
	) error
	QueryRows(context.Context, string, ...any) (VNextRows, error)
}

type VNextRows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}

// VNextSnapshotRepository owns the concrete transaction and guarantees that
// every table and source-owner inventory observes the same database snapshot.
type VNextSnapshotRepository interface {
	WithinRepeatableReadReadOnly(
		ctx context.Context,
		run func(VNextSnapshot) error,
	) error
}

type VNextObjectMember struct {
	LogicalObjectID string
	StorageKey      string
	ContentType     string
	PlaintextBytes  int64
	PlaintextSHA256 string
	Open            func(context.Context) (io.ReadCloser, error)
}

type VNextObjectInventoryProvider interface {
	OwnerID() string
	ObjectFamilyID() string
	InventoryAlgorithmID() string
	Inventory(context.Context, VNextSnapshot) ([]VNextObjectMember, error)
}

type vNextObjectInventoryProvider struct {
	ownerID     string
	familyID    string
	algorithmID string
	inventory   func(context.Context, VNextSnapshot) ([]VNextObjectMember, error)
}

func NewVNextObjectInventoryProvider(
	ownerID string,
	familyID string,
	algorithmID string,
	inventory func(context.Context, VNextSnapshot) ([]VNextObjectMember, error),
) VNextObjectInventoryProvider {
	return vNextObjectInventoryProvider{
		ownerID: ownerID, familyID: familyID, algorithmID: algorithmID, inventory: inventory,
	}
}

func (provider vNextObjectInventoryProvider) OwnerID() string {
	return provider.ownerID
}

func (provider vNextObjectInventoryProvider) ObjectFamilyID() string {
	return provider.familyID
}

func (provider vNextObjectInventoryProvider) InventoryAlgorithmID() string {
	return provider.algorithmID
}

func (provider vNextObjectInventoryProvider) Inventory(
	ctx context.Context,
	snapshot VNextSnapshot,
) ([]VNextObjectMember, error) {
	if provider.inventory == nil {
		return nil, fmt.Errorf("%w: inventory function is missing", ErrVNextBackup)
	}
	return provider.inventory(ctx, snapshot)
}

type VNextObjectSource interface {
	OpenRecoveryObject(context.Context, string) (io.ReadCloser, error)
	StatRecoveryObject(context.Context, string) (VNextObjectSourceInfo, error)
}

type VNextObjectSourceInfo struct {
	PlaintextBytes int64
	ContentType    string
}

func VNextStoredObjectMember(
	source VNextObjectSource,
	logicalObjectID string,
	storageKey string,
	contentType string,
	plaintextBytes int64,
	plaintextSHA256 string,
) VNextObjectMember {
	return VNextObjectMember{
		LogicalObjectID: logicalObjectID,
		StorageKey:      storageKey,
		ContentType:     contentType,
		PlaintextBytes:  plaintextBytes,
		PlaintextSHA256: plaintextSHA256,
		Open: func(ctx context.Context) (io.ReadCloser, error) {
			if source == nil {
				return nil, fmt.Errorf("%w: object source is unavailable", ErrVNextBackup)
			}
			return source.OpenRecoveryObject(ctx, storageKey)
		},
	}
}

func VNextLogicalObjectID(prefix string, values ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(values, "\x00")))
	return prefix + ":" + hex.EncodeToString(sum[:])
}

type VNextObjectInventoryCatalog struct {
	providers map[string]VNextObjectInventoryProvider
}

func NewVNextObjectInventoryCatalog(
	stateCatalog *recoverystate.Catalog,
	providers ...VNextObjectInventoryProvider,
) (*VNextObjectInventoryCatalog, error) {
	if stateCatalog == nil {
		return nil, fmt.Errorf("%w: recovery-state catalog is required", ErrVNextBackup)
	}
	if err := stateCatalog.ValidateFrozen(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVNextBackup, err)
	}
	expected := make(map[string]recoverystate.CatalogObjectFamily)
	for _, family := range stateCatalog.Document().ObjectFamilies {
		if family.BackupInclusion == recoverystate.InclusionRequired {
			expected[family.ObjectFamilyID] = family
		}
	}
	registered := make(map[string]VNextObjectInventoryProvider, len(providers))
	for _, provider := range providers {
		if provider == nil {
			return nil, fmt.Errorf("%w: nil object inventory provider", ErrVNextBackup)
		}
		familyID := strings.TrimSpace(provider.ObjectFamilyID())
		family, admitted := expected[familyID]
		if !admitted {
			return nil, fmt.Errorf("%w: unclassified object family %q", ErrVNextBackup, familyID)
		}
		if _, duplicate := registered[familyID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate inventory provider for %q", ErrVNextBackup, familyID)
		}
		if provider.OwnerID() != family.OwnerID ||
			family.InventoryAlgorithmID == nil ||
			provider.InventoryAlgorithmID() != *family.InventoryAlgorithmID {
			return nil, fmt.Errorf("%w: inventory provider does not match owner facts for %q", ErrVNextBackup, familyID)
		}
		registered[familyID] = provider
	}
	if len(registered) != len(expected) {
		missing := make([]string, 0)
		for familyID := range expected {
			if registered[familyID] == nil {
				missing = append(missing, familyID)
			}
		}
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: missing object inventory providers: %s", ErrVNextBackup, strings.Join(missing, ", "))
	}
	return &VNextObjectInventoryCatalog{providers: registered}, nil
}

type VNextPostgresUnitDescriptor struct {
	TableName       string `json:"table_name"`
	CodecID         string `json:"codec_id"`
	LogicalRef      string `json:"logical_ref"`
	RowCount        int64  `json:"row_count"`
	PlaintextBytes  int64  `json:"plaintext_bytes"`
	PlaintextSHA256 string `json:"plaintext_sha256"`
}

type VNextPostgresSnapshotArtifact struct {
	SchemaID                   string                        `json:"schema_id"`
	BackupSetID                string                        `json:"backup_set_id"`
	ConsistencyPointAt         time.Time                     `json:"consistency_point_at"`
	RecoveryStateCatalogSHA256 string                        `json:"recovery_state_catalog_sha256"`
	TransactionIsolation       string                        `json:"transaction_isolation"`
	Units                      []VNextPostgresUnitDescriptor `json:"units"`
	SnapshotSHA256             string                        `json:"snapshot_sha256"`
}

type VNextObjectManifestEntry struct {
	OwnerID         string `json:"owner_id"`
	ObjectFamilyID  string `json:"object_family_id"`
	LogicalObjectID string `json:"logical_object_id"`
	StorageKey      string `json:"storage_key"`
	ContentType     string `json:"content_type"`
	PlaintextBytes  int64  `json:"plaintext_bytes"`
	PlaintextSHA256 string `json:"plaintext_sha256"`
	ArtifactRef     string `json:"artifact_ref"`
}

type VNextObjectStoreBackupManifest struct {
	SchemaID                   string                     `json:"schema_id"`
	BackupSetID                string                     `json:"backup_set_id"`
	ConsistencyPointAt         time.Time                  `json:"consistency_point_at"`
	RecoveryStateCatalogSHA256 string                     `json:"recovery_state_catalog_sha256"`
	Objects                    []VNextObjectManifestEntry `json:"objects"`
	ManifestSHA256             string                     `json:"manifest_sha256"`
}

type VNextArtifactCount struct {
	Kind  string `json:"kind"`
	Count int64  `json:"count"`
}

type VNextObjectStoreBackupSummary struct {
	SchemaID            string               `json:"schema_id"`
	BackupSetID         string               `json:"backup_set_id"`
	ConsistencyPointAt  time.Time            `json:"consistency_point_at"`
	ObjectCount         int64                `json:"object_count"`
	TotalPlaintextBytes int64                `json:"total_plaintext_bytes"`
	FamilyCounts        []VNextArtifactCount `json:"family_counts"`
	ManifestSHA256      string               `json:"manifest_sha256"`
}

type VNextArtifactProof struct {
	Kind            string `json:"kind"`
	SchemaID        string `json:"schema_id"`
	LogicalRef      string `json:"logical_ref"`
	ContentType     string `json:"content_type"`
	PlaintextBytes  int64  `json:"plaintext_bytes"`
	PlaintextSHA256 string `json:"plaintext_sha256"`
	EnvelopeRef     string `json:"envelope_ref"`
	EnvelopeSHA256  string `json:"envelope_sha256"`
}

type VNextBackupIntegrityManifest struct {
	SchemaID                   string               `json:"schema_id"`
	BackupSetID                string               `json:"backup_set_id"`
	ConsistencyPointAt         time.Time            `json:"consistency_point_at"`
	CreatedAt                  time.Time            `json:"created_at"`
	RetainedUntil              time.Time            `json:"retained_until"`
	RecoveryStateCatalogSHA256 string               `json:"recovery_state_catalog_sha256"`
	CodecRegistrySHA256        string               `json:"codec_registry_sha256"`
	PostgresSnapshotRef        string               `json:"postgres_snapshot_ref"`
	ObjectStoreManifestRef     string               `json:"object_store_manifest_ref"`
	Artifacts                  []VNextArtifactProof `json:"artifacts"`
	ManifestSHA256             string               `json:"manifest_sha256"`
}

type VNextCapturedBackup struct {
	BackupSetID         uuid.UUID
	PostgresProof       BackupArtifactStreamProof
	ObjectManifestProof BackupArtifactStreamProof
	IntegrityProof      BackupArtifactStreamProof
	IntegrityManifest   VNextBackupIntegrityManifest
}

type VNextGraphProjectionRestoreArtifacts struct {
	AlgorithmID                 string
	RecoveryStateCatalogSHA256  string
	SourceRegistryJSON          []byte
	SourceRegistrySHA256        string
	ImplementationBindingJSON   []byte
	ImplementationBindingSHA256 string
}

type VNextCaptureParams struct {
	BackupSetID        uuid.UUID
	ConsistencyPointAt time.Time
	CreatedAt          time.Time
	RetainedUntil      time.Time
}

type VNextCaptureService struct {
	snapshots   VNextSnapshotRepository
	storage     StreamingBackupStorage
	state       *recoverystate.Catalog
	inventories *VNextObjectInventoryCatalog
	generation  *vNextRecoveryGeneration
}

func NewVNextCaptureService(
	snapshots VNextSnapshotRepository,
	storage StreamingBackupStorage,
	stateCatalog *recoverystate.Catalog,
	inventories *VNextObjectInventoryCatalog,
) (*VNextCaptureService, error) {
	if snapshots == nil || storage == nil || stateCatalog == nil || inventories == nil {
		return nil, fmt.Errorf("%w: capture dependencies are required", ErrVNextBackup)
	}
	if err := stateCatalog.ValidateFrozen(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVNextBackup, err)
	}
	if len(stateCatalog.RequiredTableNames()) != recoverystate.RequiredTableCount {
		return nil, fmt.Errorf("%w: required table count drift", ErrVNextBackup)
	}
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		return nil, err
	}
	if stateCatalog.DigestSHA256() != registry.current.stateCatalog.DigestSHA256() {
		return nil, fmt.Errorf("%w: capture requires the current Recovery generation catalog", ErrVNextBackup)
	}
	return &VNextCaptureService{
		snapshots: snapshots, storage: storage, state: stateCatalog, inventories: inventories,
		generation: registry.current,
	}, nil
}

func (service *VNextCaptureService) Capture(
	ctx context.Context,
	params VNextCaptureParams,
) (VNextCapturedBackup, error) {
	if service == nil {
		return VNextCapturedBackup{}, fmt.Errorf("%w: capture service is unavailable", ErrVNextBackup)
	}
	if params.BackupSetID == uuid.Nil {
		params.BackupSetID = uuid.New()
	}
	params.ConsistencyPointAt = backupTimestamp(params.ConsistencyPointAt.UTC())
	params.CreatedAt = backupTimestamp(params.CreatedAt.UTC())
	params.RetainedUntil = backupTimestamp(params.RetainedUntil.UTC())
	if params.ConsistencyPointAt.IsZero() || params.CreatedAt.IsZero() ||
		!params.RetainedUntil.After(params.CreatedAt) {
		return VNextCapturedBackup{}, fmt.Errorf("%w: capture timestamps are invalid", ErrVNextBackup)
	}

	var unitProofs []VNextArtifactProof
	var units []VNextPostgresUnitDescriptor
	var objectProofs map[string]BackupArtifactStreamProof
	var objectEntries []VNextObjectManifestEntry
	err := service.snapshots.WithinRepeatableReadReadOnly(ctx, func(snapshot VNextSnapshot) error {
		if snapshot == nil {
			return fmt.Errorf("%w: snapshot capability is nil", ErrVNextBackup)
		}
		for _, tableName := range service.state.RequiredTableNames() {
			proof, rowCount, err := service.captureTableUnit(ctx, snapshot, params.BackupSetID, tableName)
			if err != nil {
				return err
			}
			units = append(units, VNextPostgresUnitDescriptor{
				TableName: tableName, CodecID: PostgresSnapshotUnitV1SchemaID,
				LogicalRef: proof.LogicalRef, RowCount: rowCount,
				PlaintextBytes: proof.PlaintextBytes, PlaintextSHA256: proof.PlaintextSHA256,
			})
			unitProofs = append(unitProofs, vNextArtifactProof("postgres_snapshot_unit", PostgresSnapshotUnitV1SchemaID, proof))
		}
		var err error
		objectEntries, objectProofs, err = service.captureObjectInventories(ctx, snapshot, params.BackupSetID)
		return err
	})
	if err != nil {
		return VNextCapturedBackup{}, err
	}

	postgresArtifact := VNextPostgresSnapshotArtifact{
		SchemaID:                   PostgresSnapshotArtifactV2SchemaID,
		BackupSetID:                params.BackupSetID.String(),
		ConsistencyPointAt:         params.ConsistencyPointAt,
		RecoveryStateCatalogSHA256: service.state.DigestSHA256(),
		TransactionIsolation:       VNextTransactionIsolation,
		Units:                      units,
	}
	if err := assignSelfDigest(vNextPostgresSnapshotDigestDomain, &postgresArtifact.SnapshotSHA256, postgresArtifact); err != nil {
		return VNextCapturedBackup{}, err
	}
	postgresProof, err := service.writeJSONArtifact(
		ctx, params.BackupSetID, "postgres_snapshot", PostgresSnapshotArtifactV2SchemaID, postgresArtifact,
	)
	if err != nil {
		return VNextCapturedBackup{}, err
	}

	objectManifest := VNextObjectStoreBackupManifest{
		SchemaID:                   ObjectStoreBackupManifestV2SchemaID,
		BackupSetID:                params.BackupSetID.String(),
		ConsistencyPointAt:         params.ConsistencyPointAt,
		RecoveryStateCatalogSHA256: service.state.DigestSHA256(),
		Objects:                    objectEntries,
	}
	if err := assignSelfDigest(vNextObjectManifestDigestDomain, &objectManifest.ManifestSHA256, objectManifest); err != nil {
		return VNextCapturedBackup{}, err
	}
	objectManifestProof, err := service.writeJSONArtifact(
		ctx, params.BackupSetID, "object_store_manifest", ObjectStoreBackupManifestV2SchemaID, objectManifest,
	)
	if err != nil {
		return VNextCapturedBackup{}, err
	}
	summary := buildVNextObjectSummary(params, objectManifest)
	objectSummaryProof, err := service.writeJSONArtifact(
		ctx, params.BackupSetID, "object_store_summary", ObjectStoreBackupSummaryV2SchemaID, summary,
	)
	if err != nil {
		return VNextCapturedBackup{}, err
	}
	graphSourceRegistryProof, err := service.writeRawJSONArtifact(
		ctx,
		params.BackupSetID,
		"graph_projection_restore_source_registry",
		service.generation.graph.SourceRegistryJSON,
	)
	if err != nil {
		return VNextCapturedBackup{}, err
	}
	graphImplementationBindingProof, err := service.writeRawJSONArtifact(
		ctx,
		params.BackupSetID,
		"graph_projection_restore_implementation_binding",
		service.generation.graph.ImplementationBindingJSON,
	)
	if err != nil {
		return VNextCapturedBackup{}, err
	}

	artifacts := append([]VNextArtifactProof(nil), unitProofs...)
	artifacts = append(artifacts,
		vNextArtifactProof("graph_projection_restore_implementation_binding", service.generation.bindingSchemaID, graphImplementationBindingProof),
		vNextArtifactProof("graph_projection_restore_source_registry", service.generation.registrySchemaID, graphSourceRegistryProof),
		vNextArtifactProof("postgres_snapshot", PostgresSnapshotArtifactV2SchemaID, postgresProof),
		vNextArtifactProof("object_store_manifest", ObjectStoreBackupManifestV2SchemaID, objectManifestProof),
		vNextArtifactProof("object_store_summary", ObjectStoreBackupSummaryV2SchemaID, objectSummaryProof),
	)
	sort.Slice(artifacts, func(left, right int) bool {
		if artifacts[left].Kind != artifacts[right].Kind {
			return artifacts[left].Kind < artifacts[right].Kind
		}
		return artifacts[left].LogicalRef < artifacts[right].LogicalRef
	})
	integrity := VNextBackupIntegrityManifest{
		SchemaID:                   BackupIntegrityManifestV3SchemaID,
		BackupSetID:                params.BackupSetID.String(),
		ConsistencyPointAt:         params.ConsistencyPointAt,
		CreatedAt:                  params.CreatedAt,
		RetainedUntil:              params.RetainedUntil,
		RecoveryStateCatalogSHA256: service.state.DigestSHA256(),
		CodecRegistrySHA256:        service.generation.codecRegistrySHA256,
		PostgresSnapshotRef:        postgresProof.LogicalRef,
		ObjectStoreManifestRef:     objectManifestProof.LogicalRef,
		Artifacts:                  artifacts,
	}
	if err := assignSelfDigest(vNextIntegrityManifestDigestDomain, &integrity.ManifestSHA256, integrity); err != nil {
		return VNextCapturedBackup{}, err
	}
	integrityProof, err := service.writeJSONArtifact(
		ctx, params.BackupSetID, "integrity_manifest", BackupIntegrityManifestV3SchemaID, integrity,
	)
	if err != nil {
		return VNextCapturedBackup{}, err
	}
	_ = objectProofs // object proofs are bound by the private object manifest.
	return VNextCapturedBackup{
		BackupSetID: params.BackupSetID, PostgresProof: postgresProof,
		ObjectManifestProof: objectManifestProof,
		IntegrityProof:      integrityProof, IntegrityManifest: integrity,
	}, nil
}

func VNextMetadataArtifactKey(logicalRef string) string {
	return VNextMetadataArtifactScheme + logicalRef
}

func VNextLogicalRefFromMetadataKey(key string) (string, bool) {
	if !strings.HasPrefix(key, VNextMetadataArtifactScheme) {
		return "", false
	}
	logicalRef := strings.TrimPrefix(key, VNextMetadataArtifactScheme)
	if _, err := validateBackupLogicalRef(logicalRef); err != nil {
		return "", false
	}
	return logicalRef, true
}

func VNextProofFromMetadata(
	ctx context.Context,
	storage StreamingBackupStorage,
	key string,
	contentType string,
	plaintextBytes int64,
	plaintextSHA256 string,
) (BackupArtifactStreamProof, error) {
	logicalRef, ok := VNextLogicalRefFromMetadataKey(key)
	if !ok {
		return BackupArtifactStreamProof{}, fmt.Errorf("%w: invalid vNext metadata key", ErrVNextBackup)
	}
	resolver, ok := storage.(VNextObjectProofResolver)
	if !ok {
		return BackupArtifactStreamProof{}, fmt.Errorf("%w: vNext proof resolver is required", ErrVNextBackup)
	}
	return resolver.ResolveObjectProof(ctx, VNextObjectManifestEntry{
		LogicalObjectID: "metadata",
		StorageKey:      key,
		ContentType:     contentType,
		PlaintextBytes:  plaintextBytes,
		PlaintextSHA256: plaintextSHA256,
		ArtifactRef:     logicalRef,
	})
}

func (service *VNextCaptureService) captureTableUnit(
	ctx context.Context,
	snapshot VNextSnapshot,
	backupSetID uuid.UUID,
	tableName string,
) (BackupArtifactStreamProof, int64, error) {
	logicalRef := fmt.Sprintf("backup_sets/%s/vnext/postgres/%s.ndjson", backupSetID, tableName)
	envelopeRef := logicalRef + ".envelope.json"
	reader, writer := io.Pipe()
	rowCountResult := make(chan int64, 1)
	go func() {
		rowCount, err := writeVNextPostgresUnit(ctx, writer, snapshot, tableName)
		rowCountResult <- rowCount
		_ = writer.CloseWithError(err)
	}()
	proof, writeErr := service.storage.WriteArtifactStream(ctx, BackupArtifactStreamWriteRequest{
		LogicalRef: logicalRef, EnvelopeRef: envelopeRef,
		ContentType: vNextNDJSONContentType, Plaintext: reader,
	})
	_ = reader.Close()
	rowCount := <-rowCountResult
	if writeErr != nil {
		return BackupArtifactStreamProof{}, 0, fmt.Errorf("capture vNext table %s: %w", tableName, writeErr)
	}
	return proof, rowCount, nil
}

func writeVNextPostgresUnit(
	ctx context.Context,
	writer io.Writer,
	snapshot VNextSnapshot,
	tableName string,
) (int64, error) {
	rowsHasher := sha256.New()
	if err := writeCanonicalJSONLine(writer, struct {
		SchemaID   string `json:"schema_id"`
		RecordKind string `json:"record_kind"`
		TableName  string `json:"table_name"`
	}{PostgresSnapshotUnitV1SchemaID, "header", tableName}); err != nil {
		return 0, err
	}
	var rowCount int64
	err := snapshot.StreamCanonicalTableRows(ctx, tableName, func(raw json.RawMessage) error {
		canonical, err := canonicalJSONObject(raw)
		if err != nil {
			return fmt.Errorf("%w: table %s row is not canonicalizable: %v", ErrVNextBackup, tableName, err)
		}
		record := struct {
			SchemaID   string          `json:"schema_id"`
			RecordKind string          `json:"record_kind"`
			Row        json.RawMessage `json:"row"`
		}{PostgresSnapshotUnitV1SchemaID, "row", canonical}
		line, err := json.Marshal(record)
		if err != nil {
			return err
		}
		line = append(line, '\n')
		if _, err := rowsHasher.Write(line); err != nil {
			return err
		}
		if _, err := writer.Write(line); err != nil {
			return err
		}
		rowCount++
		return nil
	})
	if err != nil {
		return 0, err
	}
	if err := writeCanonicalJSONLine(writer, struct {
		SchemaID   string `json:"schema_id"`
		RecordKind string `json:"record_kind"`
		RowCount   int64  `json:"row_count"`
		RowsSHA256 string `json:"rows_sha256"`
	}{
		PostgresSnapshotUnitV1SchemaID, "trailer", rowCount,
		hex.EncodeToString(rowsHasher.Sum(nil)),
	}); err != nil {
		return 0, err
	}
	return rowCount, nil
}

func (service *VNextCaptureService) captureObjectInventories(
	ctx context.Context,
	snapshot VNextSnapshot,
	backupSetID uuid.UUID,
) ([]VNextObjectManifestEntry, map[string]BackupArtifactStreamProof, error) {
	familyIDs := make([]string, 0, len(service.inventories.providers))
	for familyID := range service.inventories.providers {
		familyIDs = append(familyIDs, familyID)
	}
	sort.Strings(familyIDs)
	entries := make([]VNextObjectManifestEntry, 0)
	proofs := make(map[string]BackupArtifactStreamProof)
	seenStorageKeys := make(map[string]struct{})
	for _, familyID := range familyIDs {
		provider := service.inventories.providers[familyID]
		members, err := provider.Inventory(ctx, snapshot)
		if err != nil {
			return nil, nil, fmt.Errorf("inventory vNext object family %s: %w", familyID, err)
		}
		sort.Slice(members, func(left, right int) bool {
			if members[left].LogicalObjectID != members[right].LogicalObjectID {
				return members[left].LogicalObjectID < members[right].LogicalObjectID
			}
			return members[left].StorageKey < members[right].StorageKey
		})
		seenIDs := make(map[string]struct{}, len(members))
		for _, member := range members {
			if err := validateVNextObjectMember(member); err != nil {
				return nil, nil, fmt.Errorf("inventory vNext object family %s: %w", familyID, err)
			}
			if _, duplicate := seenIDs[member.LogicalObjectID]; duplicate {
				return nil, nil, fmt.Errorf("%w: duplicate logical object %s/%s", ErrVNextBackup, familyID, member.LogicalObjectID)
			}
			seenIDs[member.LogicalObjectID] = struct{}{}
			if _, duplicate := seenStorageKeys[member.StorageKey]; duplicate {
				return nil, nil, fmt.Errorf("%w: storage key %q is owned more than once", ErrVNextBackup, member.StorageKey)
			}
			seenStorageKeys[member.StorageKey] = struct{}{}
			body, err := member.Open(ctx)
			if err != nil {
				return nil, nil, fmt.Errorf("open vNext object %s/%s: %w", familyID, member.LogicalObjectID, err)
			}
			logicalRef := fmt.Sprintf(
				"backup_sets/%s/vnext/objects/%s/%s",
				backupSetID, familyID, member.LogicalObjectID,
			)
			proof, writeErr := service.storage.WriteArtifactStream(ctx, BackupArtifactStreamWriteRequest{
				LogicalRef: logicalRef, EnvelopeRef: logicalRef + ".envelope.json",
				ContentType: member.ContentType, Plaintext: body,
			})
			closeErr := body.Close()
			if writeErr != nil {
				return nil, nil, fmt.Errorf("capture vNext object %s/%s: %w", familyID, member.LogicalObjectID, writeErr)
			}
			if closeErr != nil {
				return nil, nil, fmt.Errorf("close vNext object %s/%s: %w", familyID, member.LogicalObjectID, closeErr)
			}
			if proof.PlaintextBytes != member.PlaintextBytes ||
				proof.PlaintextSHA256 != member.PlaintextSHA256 {
				return nil, nil, fmt.Errorf("%w: object changed during capture: %s/%s", ErrVNextBackup, familyID, member.LogicalObjectID)
			}
			proofs[logicalRef] = proof
			entries = append(entries, VNextObjectManifestEntry{
				OwnerID: provider.OwnerID(), ObjectFamilyID: familyID,
				LogicalObjectID: member.LogicalObjectID, StorageKey: member.StorageKey,
				ContentType: member.ContentType, PlaintextBytes: member.PlaintextBytes,
				PlaintextSHA256: member.PlaintextSHA256, ArtifactRef: logicalRef,
			})
		}
	}
	sort.Slice(entries, func(left, right int) bool {
		if entries[left].ObjectFamilyID != entries[right].ObjectFamilyID {
			return entries[left].ObjectFamilyID < entries[right].ObjectFamilyID
		}
		return entries[left].LogicalObjectID < entries[right].LogicalObjectID
	})
	return entries, proofs, nil
}

func (service *VNextCaptureService) writeJSONArtifact(
	ctx context.Context,
	backupSetID uuid.UUID,
	kind string,
	schemaID string,
	value any,
) (BackupArtifactStreamProof, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return BackupArtifactStreamProof{}, fmt.Errorf("encode vNext %s: %w", kind, err)
	}
	logicalRef := fmt.Sprintf("backup_sets/%s/vnext/%s.json", backupSetID, kind)
	return service.storage.WriteArtifactStream(ctx, BackupArtifactStreamWriteRequest{
		LogicalRef: logicalRef, EnvelopeRef: logicalRef + ".envelope.json",
		ContentType: vNextJSONContentType, Plaintext: bytes.NewReader(body),
	})
}

func (service *VNextCaptureService) writeRawJSONArtifact(
	ctx context.Context,
	backupSetID uuid.UUID,
	kind string,
	body []byte,
) (BackupArtifactStreamProof, error) {
	logicalRef := fmt.Sprintf("backup_sets/%s/vnext/%s.json", backupSetID, kind)
	return service.storage.WriteArtifactStream(ctx, BackupArtifactStreamWriteRequest{
		LogicalRef: logicalRef, EnvelopeRef: logicalRef + ".envelope.json",
		ContentType: vNextJSONContentType, Plaintext: bytes.NewReader(body),
	})
}

func buildVNextObjectSummary(
	params VNextCaptureParams,
	manifest VNextObjectStoreBackupManifest,
) VNextObjectStoreBackupSummary {
	counts := make(map[string]int64)
	var total int64
	for _, object := range manifest.Objects {
		counts[object.ObjectFamilyID]++
		total += object.PlaintextBytes
	}
	families := make([]string, 0, len(counts))
	for family := range counts {
		families = append(families, family)
	}
	sort.Strings(families)
	familyCounts := make([]VNextArtifactCount, 0, len(families))
	for _, family := range families {
		familyCounts = append(familyCounts, VNextArtifactCount{Kind: family, Count: counts[family]})
	}
	return VNextObjectStoreBackupSummary{
		SchemaID:           ObjectStoreBackupSummaryV2SchemaID,
		BackupSetID:        params.BackupSetID.String(),
		ConsistencyPointAt: params.ConsistencyPointAt,
		ObjectCount:        int64(len(manifest.Objects)), TotalPlaintextBytes: total,
		FamilyCounts: familyCounts, ManifestSHA256: manifest.ManifestSHA256,
	}
}

func VNextCodecRegistrySHA256() string {
	return contractrecovery.RecoveryGenerations[0].CodecRegistrySHA256
}

func vNextArtifactProof(kind string, schemaID string, proof BackupArtifactStreamProof) VNextArtifactProof {
	return VNextArtifactProof{
		Kind: kind, SchemaID: schemaID, LogicalRef: proof.LogicalRef,
		ContentType: proof.ContentType, PlaintextBytes: proof.PlaintextBytes,
		PlaintextSHA256: proof.PlaintextSHA256, EnvelopeRef: proof.EnvelopeRef,
		EnvelopeSHA256: proof.EnvelopeSHA256,
	}
}

func streamProof(proof VNextArtifactProof) BackupArtifactStreamProof {
	return BackupArtifactStreamProof{
		LogicalRef: proof.LogicalRef, ContentType: proof.ContentType,
		PlaintextBytes: proof.PlaintextBytes, PlaintextSHA256: proof.PlaintextSHA256,
		EnvelopeRef: proof.EnvelopeRef, EnvelopeSHA256: proof.EnvelopeSHA256,
	}
}

func validateVNextObjectMember(member VNextObjectMember) error {
	if strings.TrimSpace(member.LogicalObjectID) == "" ||
		strings.TrimSpace(member.StorageKey) == "" ||
		strings.ContainsAny(member.StorageKey, "\x00\r\n") ||
		member.PlaintextBytes < 0 ||
		!validSHA256Hex(member.PlaintextSHA256) ||
		member.Open == nil {
		return fmt.Errorf("%w: invalid object inventory member", ErrVNextBackup)
	}
	if _, err := validateBackupContentType(member.ContentType); err != nil {
		return err
	}
	return nil
}

func canonicalJSONObject(raw json.RawMessage) (json.RawMessage, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value map[string]any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	if len(value) == 0 {
		return nil, fmt.Errorf("row object is empty")
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return nil, fmt.Errorf("row has trailing data")
	}
	return json.Marshal(value)
}

func writeCanonicalJSONLine(writer io.Writer, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	body = append(body, '\n')
	_, err = writer.Write(body)
	return err
}

func assignSelfDigest(domain string, destination *string, value any) error {
	*destination = ""
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: encode digest preimage: %v", ErrVNextBackup, err)
	}
	sum := sha256.Sum256(append([]byte(domain), body...))
	*destination = hex.EncodeToString(sum[:])
	return nil
}

func validateSelfDigest(domain string, actual string, value any) error {
	if !validSHA256Hex(actual) {
		return fmt.Errorf("%w: invalid self digest", ErrVNextBackup)
	}
	body, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: encode digest preimage: %v", ErrVNextBackup, err)
	}
	sum := sha256.Sum256(append([]byte(domain), body...))
	if hex.EncodeToString(sum[:]) != actual {
		return fmt.Errorf("%w: self digest mismatch", ErrVNextBackup)
	}
	return nil
}

type VNextRestoreMutation interface {
	PreparePostgresTables(context.Context, []string) error
	InsertPostgresRow(context.Context, string, json.RawMessage) error
	FinishPostgresTable(context.Context, string) error
	RestoreObject(context.Context, VNextObjectManifestEntry, io.Reader) error
	RunCatalogAlgorithm(context.Context, string) error
}

type VNextRestoreTarget interface {
	WithAtomicRestore(context.Context, *recoverystate.Catalog, func(VNextRestoreMutation) error) error
}

type VNextRestoreAlgorithmCatalog struct {
	algorithms map[string]struct{}
}

func NewVNextRestoreAlgorithmCatalog(
	stateCatalog *recoverystate.Catalog,
	algorithmIDs ...string,
) (*VNextRestoreAlgorithmCatalog, error) {
	if stateCatalog == nil {
		return nil, fmt.Errorf("%w: recovery-state catalog is required", ErrVNextBackup)
	}
	expected := make(map[string]struct{})
	for _, table := range stateCatalog.Document().Tables {
		if table.RestoreAction == recoverystate.RebuildState ||
			table.RestoreAction == recoverystate.InvalidateState {
			if table.AlgorithmID == nil {
				return nil, fmt.Errorf("%w: table %s lacks restore algorithm", ErrVNextBackup, table.TableName)
			}
			expected[*table.AlgorithmID] = struct{}{}
		}
	}
	for _, family := range stateCatalog.Document().ObjectFamilies {
		if family.ValidationAlgorithmID != nil {
			expected[*family.ValidationAlgorithmID] = struct{}{}
		}
		if family.RestoreAlgorithmID != nil {
			expected[*family.RestoreAlgorithmID] = struct{}{}
		}
	}
	actual := make(map[string]struct{}, len(algorithmIDs))
	for _, algorithmID := range algorithmIDs {
		if _, admitted := expected[algorithmID]; !admitted {
			return nil, fmt.Errorf("%w: unowned restore algorithm %q", ErrVNextBackup, algorithmID)
		}
		actual[algorithmID] = struct{}{}
	}
	missing := make([]string, 0)
	for algorithmID := range expected {
		if _, exists := actual[algorithmID]; !exists {
			missing = append(missing, algorithmID)
		}
	}
	if len(missing) != 0 {
		sort.Strings(missing)
		return nil, fmt.Errorf("%w: missing restore algorithms: %s", ErrVNextBackup, strings.Join(missing, ", "))
	}
	return &VNextRestoreAlgorithmCatalog{algorithms: actual}, nil
}

func RequiredVNextRestoreAlgorithmIDs(stateCatalog *recoverystate.Catalog) []string {
	if stateCatalog == nil {
		return nil
	}
	seen := make(map[string]struct{})
	for _, table := range stateCatalog.Document().Tables {
		if table.AlgorithmID != nil {
			seen[*table.AlgorithmID] = struct{}{}
		}
	}
	for _, family := range stateCatalog.Document().ObjectFamilies {
		if family.ValidationAlgorithmID != nil {
			seen[*family.ValidationAlgorithmID] = struct{}{}
		}
		if family.RestoreAlgorithmID != nil {
			seen[*family.RestoreAlgorithmID] = struct{}{}
		}
	}
	result := make([]string, 0, len(seen))
	for algorithmID := range seen {
		result = append(result, algorithmID)
	}
	sort.Strings(result)
	return result
}

type VNextRestoreService struct {
	storage     StreamingBackupStorage
	state       *recoverystate.Catalog
	algorithms  *VNextRestoreAlgorithmCatalog
	generations *vNextRecoveryGenerationRegistry
}

type VNextRestoreVerificationEvidence struct {
	ManifestSHA256        string
	RestoredObjectCount   int64
	GraphRestoreArtifacts VNextGraphProjectionRestoreArtifacts
	stateCatalog          *recoverystate.Catalog
	codecRegistrySHA256   string
}

func NewVNextRestoreService(
	storage StreamingBackupStorage,
	stateCatalog *recoverystate.Catalog,
	algorithms *VNextRestoreAlgorithmCatalog,
) (*VNextRestoreService, error) {
	if storage == nil || stateCatalog == nil || algorithms == nil {
		return nil, fmt.Errorf("%w: restore dependencies are required", ErrVNextBackup)
	}
	if err := stateCatalog.ValidateFrozen(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrVNextBackup, err)
	}
	generations, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		return nil, err
	}
	if stateCatalog.DigestSHA256() != generations.current.stateCatalog.DigestSHA256() {
		return nil, fmt.Errorf("%w: restore composition requires the current Recovery catalog", ErrVNextBackup)
	}
	return &VNextRestoreService{
		storage: storage, state: stateCatalog, algorithms: algorithms, generations: generations,
	}, nil
}

func (service *VNextRestoreService) ReadVerificationEvidence(
	ctx context.Context,
	integrityProof BackupArtifactStreamProof,
) (VNextRestoreVerificationEvidence, error) {
	if service == nil {
		return VNextRestoreVerificationEvidence{}, fmt.Errorf("%w: restore service is required", ErrVNextBackup)
	}
	integrityBody, err := service.readJSONArtifact(ctx, integrityProof, 32<<20)
	if err != nil {
		return VNextRestoreVerificationEvidence{}, err
	}
	var integrity VNextBackupIntegrityManifest
	if err := strictDecodeJSON(integrityBody, &integrity); err != nil {
		return VNextRestoreVerificationEvidence{}, fmt.Errorf("%w: decode integrity manifest: %v", ErrVNextBackup, err)
	}
	generation, err := service.validateIntegrityManifest(integrity)
	if err != nil {
		return VNextRestoreVerificationEvidence{}, err
	}
	proofs := make(map[string]VNextArtifactProof, len(integrity.Artifacts))
	var objectManifestProof VNextArtifactProof
	for _, proof := range integrity.Artifacts {
		if _, duplicate := proofs[proof.LogicalRef]; duplicate {
			return VNextRestoreVerificationEvidence{}, fmt.Errorf("%w: duplicate artifact proof", ErrVNextBackup)
		}
		proofs[proof.LogicalRef] = proof
		if proof.LogicalRef == integrity.ObjectStoreManifestRef {
			objectManifestProof = proof
		}
	}
	graphArtifacts, err := service.resolveGraphProjectionRestoreArtifacts(ctx, generation, proofs)
	if err != nil {
		return VNextRestoreVerificationEvidence{}, err
	}
	if objectManifestProof.SchemaID != ObjectStoreBackupManifestV2SchemaID {
		return VNextRestoreVerificationEvidence{}, fmt.Errorf("%w: object manifest proof is missing", ErrVNextBackup)
	}
	objectBody, err := service.readJSONArtifact(ctx, streamProof(objectManifestProof), 1<<30)
	if err != nil {
		return VNextRestoreVerificationEvidence{}, err
	}
	var objectManifest VNextObjectStoreBackupManifest
	if err := strictDecodeJSON(objectBody, &objectManifest); err != nil {
		return VNextRestoreVerificationEvidence{}, fmt.Errorf("%w: decode object manifest: %v", ErrVNextBackup, err)
	}
	if err := service.validateObjectManifest(integrity, generation, objectManifest); err != nil {
		return VNextRestoreVerificationEvidence{}, err
	}
	return VNextRestoreVerificationEvidence{
		ManifestSHA256:        integrity.ManifestSHA256,
		RestoredObjectCount:   int64(len(objectManifest.Objects)),
		GraphRestoreArtifacts: graphArtifacts,
		stateCatalog:          generation.stateCatalog,
		codecRegistrySHA256:   generation.codecRegistrySHA256,
	}, nil
}

func (service *VNextRestoreService) Restore(
	ctx context.Context,
	target VNextRestoreTarget,
	integrityProof BackupArtifactStreamProof,
) error {
	if service == nil || target == nil {
		return fmt.Errorf("%w: restore service and target are required", ErrVNextBackup)
	}
	integrityBody, err := service.readJSONArtifact(ctx, integrityProof, 32<<20)
	if err != nil {
		return err
	}
	var integrity VNextBackupIntegrityManifest
	if err := strictDecodeJSON(integrityBody, &integrity); err != nil {
		return fmt.Errorf("%w: decode integrity manifest: %v", ErrVNextBackup, err)
	}
	generation, err := service.validateIntegrityManifest(integrity)
	if err != nil {
		return err
	}
	proofs := make(map[string]VNextArtifactProof, len(integrity.Artifacts))
	for _, proof := range integrity.Artifacts {
		if _, duplicate := proofs[proof.LogicalRef]; duplicate {
			return fmt.Errorf("%w: duplicate artifact proof %q", ErrVNextBackup, proof.LogicalRef)
		}
		proofs[proof.LogicalRef] = proof
		if err := service.storage.ReadArtifactStream(ctx, streamProof(proof), io.Discard); err != nil {
			return fmt.Errorf("preflight vNext artifact %s: %w", proof.LogicalRef, err)
		}
	}
	if _, err := service.resolveGraphProjectionRestoreArtifacts(ctx, generation, proofs); err != nil {
		return err
	}
	postgresProof, admitted := proofs[integrity.PostgresSnapshotRef]
	if !admitted || postgresProof.SchemaID != PostgresSnapshotArtifactV2SchemaID {
		return fmt.Errorf("%w: postgres snapshot proof is missing", ErrVNextBackup)
	}
	objectManifestProof, admitted := proofs[integrity.ObjectStoreManifestRef]
	if !admitted || objectManifestProof.SchemaID != ObjectStoreBackupManifestV2SchemaID {
		return fmt.Errorf("%w: object manifest proof is missing", ErrVNextBackup)
	}
	postgresBody, err := service.readJSONArtifact(ctx, streamProof(postgresProof), 32<<20)
	if err != nil {
		return err
	}
	var postgresArtifact VNextPostgresSnapshotArtifact
	if err := strictDecodeJSON(postgresBody, &postgresArtifact); err != nil {
		return fmt.Errorf("%w: decode postgres snapshot: %v", ErrVNextBackup, err)
	}
	if err := service.validatePostgresArtifact(integrity, generation, postgresArtifact, proofs); err != nil {
		return err
	}
	objectBody, err := service.readJSONArtifact(ctx, streamProof(objectManifestProof), 1<<30)
	if err != nil {
		return err
	}
	var objectManifest VNextObjectStoreBackupManifest
	if err := strictDecodeJSON(objectBody, &objectManifest); err != nil {
		return fmt.Errorf("%w: decode object manifest: %v", ErrVNextBackup, err)
	}
	if err := service.validateObjectManifest(integrity, generation, objectManifest); err != nil {
		return err
	}
	objectProofs := make(map[string]BackupArtifactStreamProof, len(objectManifest.Objects))
	// Object envelopes are not top-level integrity artifacts, so restore needs
	// their exact envelope digests. V2 object entries deliberately omit that
	// field; resolve it through the storage proof resolver capability.
	resolver, ok := service.storage.(VNextObjectProofResolver)
	if len(objectManifest.Objects) != 0 && !ok {
		return fmt.Errorf("%w: object proof resolver is required", ErrVNextBackup)
	}
	for _, object := range objectManifest.Objects {
		proof, err := resolver.ResolveObjectProof(ctx, object)
		if err != nil {
			return err
		}
		if proof.LogicalRef != object.ArtifactRef ||
			proof.PlaintextBytes != object.PlaintextBytes ||
			proof.PlaintextSHA256 != object.PlaintextSHA256 ||
			proof.ContentType != object.ContentType {
			return fmt.Errorf("%w: resolved object proof mismatch", ErrVNextBackup)
		}
		if err := service.storage.ReadArtifactStream(ctx, proof, io.Discard); err != nil {
			return fmt.Errorf("preflight vNext object %s: %w", object.ArtifactRef, err)
		}
		objectProofs[object.ArtifactRef] = proof
	}

	return target.WithAtomicRestore(ctx, generation.stateCatalog, func(mutation VNextRestoreMutation) error {
		tableNames := generation.stateCatalog.RequiredTableNames()
		if err := mutation.PreparePostgresTables(ctx, tableNames); err != nil {
			return err
		}
		for _, unit := range postgresArtifact.Units {
			proof := proofs[unit.LogicalRef]
			if err := service.restoreUnit(ctx, mutation, unit, proof); err != nil {
				return err
			}
		}
		for _, object := range objectManifest.Objects {
			proof := objectProofs[object.ArtifactRef]
			reader, writer := io.Pipe()
			streamErr := make(chan error, 1)
			go func() {
				err := service.storage.ReadArtifactStream(ctx, proof, writer)
				streamErr <- err
				_ = writer.CloseWithError(err)
			}()
			restoreErr := mutation.RestoreObject(ctx, object, reader)
			_ = reader.Close()
			readErr := <-streamErr
			if restoreErr != nil {
				return restoreErr
			}
			if readErr != nil {
				return readErr
			}
		}
		for _, algorithmID := range RequiredVNextRestoreAlgorithmIDs(generation.stateCatalog) {
			if err := mutation.RunCatalogAlgorithm(ctx, algorithmID); err != nil {
				return fmt.Errorf("run catalog restore algorithm %s: %w", algorithmID, err)
			}
		}
		return nil
	})
}

// VNextObjectProofResolver binds object-manifest entries to immutable envelope
// metadata without expanding the manifest's adopted v2 wire shape.
type VNextObjectProofResolver interface {
	ResolveObjectProof(context.Context, VNextObjectManifestEntry) (BackupArtifactStreamProof, error)
}

func (service *VNextRestoreService) restoreUnit(
	ctx context.Context,
	mutation VNextRestoreMutation,
	unit VNextPostgresUnitDescriptor,
	proof VNextArtifactProof,
) error {
	reader, writer := io.Pipe()
	streamErr := make(chan error, 1)
	go func() {
		err := service.storage.ReadArtifactStream(ctx, streamProof(proof), writer)
		streamErr <- err
		_ = writer.CloseWithError(err)
	}()
	parseErr := decodeVNextPostgresUnit(reader, unit, func(row json.RawMessage) error {
		return mutation.InsertPostgresRow(ctx, unit.TableName, row)
	})
	_ = reader.Close()
	readErr := <-streamErr
	if parseErr != nil {
		return parseErr
	}
	if readErr != nil {
		return readErr
	}
	return mutation.FinishPostgresTable(ctx, unit.TableName)
}

func decodeVNextPostgresUnit(
	reader io.Reader,
	unit VNextPostgresUnitDescriptor,
	insert func(json.RawMessage) error,
) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 64*1024*1024)
	line := int64(0)
	rowCount := int64(0)
	rowsHasher := sha256.New()
	trailerSeen := false
	for scanner.Scan() {
		line++
		raw := append([]byte(nil), scanner.Bytes()...)
		var discriminator struct {
			SchemaID   string `json:"schema_id"`
			RecordKind string `json:"record_kind"`
		}
		if err := json.Unmarshal(raw, &discriminator); err != nil {
			return fmt.Errorf("%w: decode unit line %d: %v", ErrVNextBackup, line, err)
		}
		if discriminator.SchemaID != PostgresSnapshotUnitV1SchemaID {
			return fmt.Errorf("%w: wrong unit schema", ErrVNextBackup)
		}
		switch discriminator.RecordKind {
		case "header":
			if line != 1 {
				return fmt.Errorf("%w: misplaced unit header", ErrVNextBackup)
			}
			var header struct {
				SchemaID   string `json:"schema_id"`
				RecordKind string `json:"record_kind"`
				TableName  string `json:"table_name"`
			}
			if err := strictDecodeJSON(raw, &header); err != nil || header.TableName != unit.TableName {
				return fmt.Errorf("%w: unit header mismatch", ErrVNextBackup)
			}
		case "row":
			if line == 1 || trailerSeen {
				return fmt.Errorf("%w: misplaced unit row", ErrVNextBackup)
			}
			var record struct {
				SchemaID   string          `json:"schema_id"`
				RecordKind string          `json:"record_kind"`
				Row        json.RawMessage `json:"row"`
			}
			if err := strictDecodeJSON(raw, &record); err != nil {
				return fmt.Errorf("%w: invalid unit row: %v", ErrVNextBackup, err)
			}
			canonical, err := canonicalJSONObject(record.Row)
			if err != nil || !bytes.Equal(canonical, record.Row) {
				return fmt.Errorf("%w: non-canonical unit row", ErrVNextBackup)
			}
			_, _ = rowsHasher.Write(append(raw, '\n'))
			if err := insert(record.Row); err != nil {
				return err
			}
			rowCount++
		case "trailer":
			if line == 1 || trailerSeen {
				return fmt.Errorf("%w: misplaced unit trailer", ErrVNextBackup)
			}
			var trailer struct {
				SchemaID   string `json:"schema_id"`
				RecordKind string `json:"record_kind"`
				RowCount   int64  `json:"row_count"`
				RowsSHA256 string `json:"rows_sha256"`
			}
			if err := strictDecodeJSON(raw, &trailer); err != nil ||
				trailer.RowCount != rowCount ||
				trailer.RowCount != unit.RowCount ||
				trailer.RowsSHA256 != hex.EncodeToString(rowsHasher.Sum(nil)) {
				return fmt.Errorf("%w: unit trailer mismatch", ErrVNextBackup)
			}
			trailerSeen = true
		default:
			return fmt.Errorf("%w: invalid unit record kind", ErrVNextBackup)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if !trailerSeen {
		return fmt.Errorf("%w: unit trailer is missing", ErrVNextBackup)
	}
	return nil
}

func (service *VNextRestoreService) readJSONArtifact(
	ctx context.Context,
	proof BackupArtifactStreamProof,
	maxBytes int64,
) ([]byte, error) {
	if proof.PlaintextBytes > maxBytes {
		return nil, fmt.Errorf("%w: JSON artifact exceeds bound", ErrVNextBackup)
	}
	var body bytes.Buffer
	if err := service.storage.ReadArtifactStream(ctx, proof, &body); err != nil {
		return nil, err
	}
	return body.Bytes(), nil
}

func (service *VNextRestoreService) validateIntegrityManifest(
	manifest VNextBackupIntegrityManifest,
) (*vNextRecoveryGeneration, error) {
	generation, admitted := service.generations.lookup(
		manifest.RecoveryStateCatalogSHA256,
		manifest.CodecRegistrySHA256,
	)
	if manifest.SchemaID != BackupIntegrityManifestV3SchemaID ||
		!admitted ||
		len(manifest.Artifacts) < 3 || len(manifest.Artifacts) > 4096 {
		return nil, fmt.Errorf("%w: integrity manifest facts mismatch", ErrVNextBackup)
	}
	actual := manifest.ManifestSHA256
	copyValue := manifest
	copyValue.ManifestSHA256 = ""
	if err := validateSelfDigest(vNextIntegrityManifestDigestDomain, actual, copyValue); err != nil {
		return nil, err
	}
	return generation, nil
}

func (service *VNextRestoreService) resolveGraphProjectionRestoreArtifacts(
	ctx context.Context,
	generation *vNextRecoveryGeneration,
	proofs map[string]VNextArtifactProof,
) (VNextGraphProjectionRestoreArtifacts, error) {
	var registryProofs []VNextArtifactProof
	var bindingProofs []VNextArtifactProof
	for _, proof := range proofs {
		switch {
		case proof.Kind == "graph_projection_restore_source_registry" ||
			proof.SchemaID == GraphProjectionRestoreSourceRegistryV2SchemaID ||
			proof.SchemaID == GraphProjectionRestoreSourceRegistryV3SchemaID:
			registryProofs = append(registryProofs, proof)
		case proof.Kind == "graph_projection_restore_implementation_binding" ||
			proof.SchemaID == GraphProjectionRestoreImplementationBindingV2SchemaID ||
			proof.SchemaID == GraphProjectionRestoreImplementationBindingV3SchemaID:
			bindingProofs = append(bindingProofs, proof)
		}
	}
	if generation == nil || len(registryProofs) != 1 || len(bindingProofs) != 1 {
		return VNextGraphProjectionRestoreArtifacts{}, fmt.Errorf("%w: Graph Projection restore artifacts are missing or duplicated", ErrVNextBackup)
	}
	registryProof := registryProofs[0]
	bindingProof := bindingProofs[0]
	if registryProof.Kind != "graph_projection_restore_source_registry" ||
		bindingProof.Kind != "graph_projection_restore_implementation_binding" {
		return VNextGraphProjectionRestoreArtifacts{}, fmt.Errorf("%w: Graph Projection restore artifact binding mismatch", ErrVNextBackup)
	}
	if registryProof.SchemaID != generation.registrySchemaID ||
		registryProof.PlaintextSHA256 != generation.graph.SourceRegistrySHA256 ||
		bindingProof.SchemaID != generation.bindingSchemaID ||
		bindingProof.PlaintextSHA256 != generation.graph.ImplementationBindingSHA256 {
		return VNextGraphProjectionRestoreArtifacts{}, fmt.Errorf("%w: Graph Projection restore artifact binding mismatch", ErrVNextBackup)
	}
	registryBody, err := service.readJSONArtifact(ctx, streamProof(registryProof), 1<<20)
	if err != nil {
		return VNextGraphProjectionRestoreArtifacts{}, err
	}
	bindingBody, err := service.readJSONArtifact(ctx, streamProof(bindingProof), 1<<20)
	if err != nil {
		return VNextGraphProjectionRestoreArtifacts{}, err
	}
	if !bytes.Equal(registryBody, generation.graph.SourceRegistryJSON) ||
		!bytes.Equal(bindingBody, generation.graph.ImplementationBindingJSON) {
		return VNextGraphProjectionRestoreArtifacts{}, fmt.Errorf("%w: Graph Projection restore artifact canonical bytes mismatch", ErrVNextBackup)
	}
	return VNextGraphProjectionRestoreArtifacts{
		AlgorithmID:                 generation.graph.AlgorithmID,
		RecoveryStateCatalogSHA256:  generation.stateCatalog.DigestSHA256(),
		SourceRegistryJSON:          append([]byte(nil), registryBody...),
		SourceRegistrySHA256:        registryProof.PlaintextSHA256,
		ImplementationBindingJSON:   append([]byte(nil), bindingBody...),
		ImplementationBindingSHA256: bindingProof.PlaintextSHA256,
	}, nil
}

func (service *VNextRestoreService) validatePostgresArtifact(
	integrity VNextBackupIntegrityManifest,
	generation *vNextRecoveryGeneration,
	artifact VNextPostgresSnapshotArtifact,
	proofs map[string]VNextArtifactProof,
) error {
	if artifact.SchemaID != PostgresSnapshotArtifactV2SchemaID ||
		artifact.BackupSetID != integrity.BackupSetID ||
		!artifact.ConsistencyPointAt.Equal(integrity.ConsistencyPointAt) ||
		artifact.RecoveryStateCatalogSHA256 != integrity.RecoveryStateCatalogSHA256 ||
		artifact.TransactionIsolation != VNextTransactionIsolation ||
		generation == nil ||
		len(artifact.Units) != len(generation.stateCatalog.RequiredTableNames()) {
		return fmt.Errorf("%w: postgres snapshot facts mismatch", ErrVNextBackup)
	}
	actual := artifact.SnapshotSHA256
	copyValue := artifact
	copyValue.SnapshotSHA256 = ""
	if err := validateSelfDigest(vNextPostgresSnapshotDigestDomain, actual, copyValue); err != nil {
		return err
	}
	expectedTables := generation.stateCatalog.RequiredTableNames()
	for index, unit := range artifact.Units {
		if unit.TableName != expectedTables[index] ||
			unit.CodecID != PostgresSnapshotUnitV1SchemaID ||
			unit.RowCount < 0 ||
			!validSHA256Hex(unit.PlaintextSHA256) {
			return fmt.Errorf("%w: postgres unit facts mismatch", ErrVNextBackup)
		}
		proof, ok := proofs[unit.LogicalRef]
		if !ok || proof.SchemaID != PostgresSnapshotUnitV1SchemaID ||
			proof.PlaintextBytes != unit.PlaintextBytes ||
			proof.PlaintextSHA256 != unit.PlaintextSHA256 {
			return fmt.Errorf("%w: postgres unit proof mismatch", ErrVNextBackup)
		}
	}
	return nil
}

func (service *VNextRestoreService) validateObjectManifest(
	integrity VNextBackupIntegrityManifest,
	generation *vNextRecoveryGeneration,
	manifest VNextObjectStoreBackupManifest,
) error {
	if manifest.SchemaID != ObjectStoreBackupManifestV2SchemaID ||
		manifest.BackupSetID != integrity.BackupSetID ||
		!manifest.ConsistencyPointAt.Equal(integrity.ConsistencyPointAt) ||
		generation == nil ||
		manifest.RecoveryStateCatalogSHA256 != generation.stateCatalog.DigestSHA256() {
		return fmt.Errorf("%w: object manifest facts mismatch", ErrVNextBackup)
	}
	actual := manifest.ManifestSHA256
	copyValue := manifest
	copyValue.ManifestSHA256 = ""
	if err := validateSelfDigest(vNextObjectManifestDigestDomain, actual, copyValue); err != nil {
		return err
	}
	families := make(map[string]recoverystate.CatalogObjectFamily)
	for _, family := range generation.stateCatalog.Document().ObjectFamilies {
		families[family.ObjectFamilyID] = family
	}
	seen := make(map[string]struct{}, len(manifest.Objects))
	previous := ""
	for _, object := range manifest.Objects {
		family, admitted := families[object.ObjectFamilyID]
		identity := object.ObjectFamilyID + "\x00" + object.LogicalObjectID
		if !admitted || family.OwnerID != object.OwnerID ||
			(previous != "" && identity <= previous) ||
			!validSHA256Hex(object.PlaintextSHA256) ||
			object.PlaintextBytes < 0 {
			return fmt.Errorf("%w: object manifest entry mismatch", ErrVNextBackup)
		}
		if _, duplicate := seen[identity]; duplicate {
			return fmt.Errorf("%w: duplicate object manifest entry", ErrVNextBackup)
		}
		seen[identity] = struct{}{}
		previous = identity
	}
	return nil
}

func strictDecodeJSON(body []byte, destination any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return fmt.Errorf("trailing JSON value")
		}
		return err
	}
	return nil
}

func verifyVNextBackupSetDurability(
	ctx context.Context,
	storage BackupStorage,
	stateCatalog *recoverystate.Catalog,
	backupSet BackupSet,
) error {
	_, err := selectVNextBackupSetGeneration(ctx, storage, stateCatalog, backupSet)
	return err
}

func selectVNextBackupSetGeneration(
	ctx context.Context,
	storage BackupStorage,
	stateCatalog *recoverystate.Catalog,
	backupSet BackupSet,
) (*vNextRecoveryGeneration, error) {
	selection, err := readVNextBackupSetGeneration(ctx, storage, stateCatalog, backupSet)
	if err != nil {
		return nil, err
	}
	restore := selection.restore
	integrity := selection.integrity
	generation := selection.generation
	proofs := selection.proofs
	for _, proof := range proofs {
		if err := restore.storage.ReadArtifactStream(ctx, streamProof(proof), io.Discard); err != nil {
			return nil, err
		}
	}
	postgresLogicalRef, ok := VNextLogicalRefFromMetadataKey(backupSet.PostgresArtifactKey)
	if !ok || postgresLogicalRef != integrity.PostgresSnapshotRef {
		return nil, fmt.Errorf("%w: postgres metadata selector mismatch", ErrVNextBackup)
	}
	postgresProof, ok := proofs[postgresLogicalRef]
	if !ok ||
		postgresProof.PlaintextBytes != backupSet.PostgresArtifactSizeBytes ||
		postgresProof.PlaintextSHA256 != backupSet.PostgresArtifactSHA256 {
		return nil, fmt.Errorf("%w: postgres metadata proof mismatch", ErrVNextBackup)
	}
	postgresBody, err := restore.readJSONArtifact(ctx, streamProof(postgresProof), 32<<20)
	if err != nil {
		return nil, err
	}
	var postgresArtifact VNextPostgresSnapshotArtifact
	if err := strictDecodeJSON(postgresBody, &postgresArtifact); err != nil {
		return nil, fmt.Errorf("%w: decode postgres snapshot: %v", ErrVNextBackup, err)
	}
	if err := restore.validatePostgresArtifact(integrity, generation, postgresArtifact, proofs); err != nil {
		return nil, err
	}
	objectLogicalRef, ok := VNextLogicalRefFromMetadataKey(backupSet.ObjectStoreArtifactKey)
	if !ok || objectLogicalRef != integrity.ObjectStoreManifestRef {
		return nil, fmt.Errorf("%w: object metadata selector mismatch", ErrVNextBackup)
	}
	objectProof, ok := proofs[objectLogicalRef]
	if !ok ||
		objectProof.PlaintextBytes != backupSet.ObjectStoreArtifactSizeBytes ||
		objectProof.PlaintextSHA256 != backupSet.ObjectStoreArtifactSHA256 {
		return nil, fmt.Errorf("%w: object metadata proof mismatch", ErrVNextBackup)
	}
	objectBody, err := restore.readJSONArtifact(ctx, streamProof(objectProof), 1<<30)
	if err != nil {
		return nil, err
	}
	var objectManifest VNextObjectStoreBackupManifest
	if err := strictDecodeJSON(objectBody, &objectManifest); err != nil {
		return nil, fmt.Errorf("%w: decode object manifest: %v", ErrVNextBackup, err)
	}
	if err := restore.validateObjectManifest(integrity, generation, objectManifest); err != nil {
		return nil, err
	}
	return generation, nil
}

type vNextBackupGenerationSelection struct {
	restore    *VNextRestoreService
	integrity  VNextBackupIntegrityManifest
	generation *vNextRecoveryGeneration
	proofs     map[string]VNextArtifactProof
}

func readVNextBackupSetGeneration(
	ctx context.Context,
	storage BackupStorage,
	stateCatalog *recoverystate.Catalog,
	backupSet BackupSet,
) (vNextBackupGenerationSelection, error) {
	if stateCatalog == nil {
		return vNextBackupGenerationSelection{}, fmt.Errorf("%w: current recovery-state catalog is required", ErrVNextBackup)
	}
	streaming, err := RequireStreamingBackupStorage(storage)
	if err != nil {
		return vNextBackupGenerationSelection{}, err
	}
	integrityProof, err := VNextProofFromMetadata(
		ctx,
		streaming,
		backupSet.IntegrityManifestKey,
		vNextJSONContentType,
		backupSet.IntegrityManifestSizeBytes,
		backupSet.IntegrityManifestSHA256,
	)
	if err != nil {
		return vNextBackupGenerationSelection{}, err
	}
	if integrityProof.PlaintextBytes > 32<<20 {
		return vNextBackupGenerationSelection{}, fmt.Errorf("%w: integrity manifest exceeds bound", ErrVNextBackup)
	}
	var integrityBody bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, integrityProof, &integrityBody); err != nil {
		return vNextBackupGenerationSelection{}, err
	}
	var integrity VNextBackupIntegrityManifest
	if err := strictDecodeJSON(integrityBody.Bytes(), &integrity); err != nil {
		return vNextBackupGenerationSelection{}, fmt.Errorf("%w: decode integrity manifest: %v", ErrVNextBackup, err)
	}
	algorithms, err := NewVNextRestoreAlgorithmCatalog(
		stateCatalog,
		RequiredVNextRestoreAlgorithmIDs(stateCatalog)...,
	)
	if err != nil {
		return vNextBackupGenerationSelection{}, err
	}
	restore, err := NewVNextRestoreService(streaming, stateCatalog, algorithms)
	if err != nil {
		return vNextBackupGenerationSelection{}, err
	}
	generation, err := restore.validateIntegrityManifest(integrity)
	if err != nil {
		return vNextBackupGenerationSelection{}, err
	}
	if integrity.BackupSetID != backupSet.BackupSetID.String() ||
		!integrity.ConsistencyPointAt.Equal(backupSet.ConsistencyPointAt) {
		return vNextBackupGenerationSelection{}, fmt.Errorf("%w: integrity manifest does not match backup metadata", ErrVNextBackup)
	}
	proofs := make(map[string]VNextArtifactProof, len(integrity.Artifacts))
	for _, proof := range integrity.Artifacts {
		if _, duplicate := proofs[proof.LogicalRef]; duplicate {
			return vNextBackupGenerationSelection{}, fmt.Errorf("%w: duplicate artifact proof", ErrVNextBackup)
		}
		proofs[proof.LogicalRef] = proof
	}
	if _, err := restore.resolveGraphProjectionRestoreArtifacts(ctx, generation, proofs); err != nil {
		return vNextBackupGenerationSelection{}, err
	}
	return vNextBackupGenerationSelection{
		restore: restore, integrity: integrity, generation: generation, proofs: proofs,
	}, nil
}
