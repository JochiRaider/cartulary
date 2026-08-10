package revisions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/gen/contractrevisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/deleterestorecontract"
)

var (
	ErrMissingSnapshotCapture     = errors.New("revisions: missing snapshot capture")
	ErrDuplicateSnapshotCapture   = errors.New("revisions: duplicate snapshot capture")
	ErrInvalidCapturedSnapshot    = errors.New("revisions: invalid captured snapshot")
	ErrSnapshotRecordTypeMismatch = errors.New("revisions: snapshot record type mismatch")
)

const SnapshotSchemaRegistrySchemaID = "cartulary.revisions_snapshot_schema_registry.v1"

type SnapshotSchemaRequirement struct {
	RecordType       string
	SourceOwner      SourceOwnerModule
	SnapshotSchemaID string
}

type snapshotSchemaRegistry struct {
	SchemaID        string                        `json:"schema_id"`
	RegistryVersion int                           `json:"registry_version"`
	Schemas         []snapshotSchemaRegistryEntry `json:"schemas"`
}

type snapshotSchemaRegistryEntry struct {
	RecordType       string            `json:"record_type"`
	SourceOwner      SourceOwnerModule `json:"source_owner"`
	SnapshotSchemaID string            `json:"snapshot_schema_id"`
}

func ParseSnapshotSchemaRequirements(data []byte) ([]SnapshotSchemaRequirement, error) {
	var registry snapshotSchemaRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("%w: decode snapshot registry: %v", ErrInvalidCapturedSnapshot, err)
	}
	if registry.SchemaID != SnapshotSchemaRegistrySchemaID || registry.RegistryVersion != 1 || len(registry.Schemas) == 0 {
		return nil, fmt.Errorf("%w: snapshot registry identity", ErrInvalidCapturedSnapshot)
	}
	requirements := make([]SnapshotSchemaRequirement, 0, len(registry.Schemas))
	for _, entry := range registry.Schemas {
		requirements = append(requirements, SnapshotSchemaRequirement(entry))
	}
	return requirements, nil
}

func CurrentSnapshotSchemaRequirements() ([]SnapshotSchemaRequirement, error) {
	artifact, ok := contractrevisions.Index["contracts/revisions/snapshot-schema-registry.v1.json"]
	if !ok {
		return nil, ErrMissingSnapshotCapture
	}
	return ParseSnapshotSchemaRequirements([]byte(artifact.JSON))
}

// CapturedRecordSnapshot is an opaque, validated canonical history value.
// Only Appender capture can create a present value.
type CapturedRecordSnapshot struct {
	recordID         uuid.UUID
	recordType       string
	snapshotSchemaID string
	value            map[string]any
}

func (snapshot CapturedRecordSnapshot) SnapshotSchemaID() string {
	return snapshot.snapshotSchemaID
}

func (snapshot CapturedRecordSnapshot) RecordID() uuid.UUID {
	return snapshot.recordID
}

type recordSnapshotCapture struct {
	sourceOwner      SourceOwnerModule
	recordType       string
	snapshotSchemaID string
	source           deleterestorecontract.DeleteRestoreSource
}

type RecordSnapshotCaptureCatalog struct {
	byRecordType map[string]recordSnapshotCapture
}

// NewRecordSnapshotCaptureCatalog requires exact source-owner closure for all
// admitted record families before any revision writer can be constructed.
func NewRecordSnapshotCaptureCatalog(contributions []ProviderContribution) (*RecordSnapshotCaptureCatalog, error) {
	requirements, err := CurrentSnapshotSchemaRequirements()
	if err != nil {
		return nil, err
	}
	return NewRecordSnapshotCaptureCatalogForRequirements(contributions, requirements)
}

// NewRecordSnapshotCaptureCatalogForRequirements compiles an exact catalog for
// an explicitly supplied contract projection. Production composition uses
// NewRecordSnapshotCaptureCatalog; candidate and isolated owner tests can use
// this constructor without weakening current-catalog closure.
func NewRecordSnapshotCaptureCatalogForRequirements(contributions []ProviderContribution, requirements []SnapshotSchemaRequirement) (*RecordSnapshotCaptureCatalog, error) {
	required := make(map[string]SnapshotSchemaRequirement, len(requirements))
	for _, requirement := range requirements {
		requirement.RecordType = strings.TrimSpace(requirement.RecordType)
		requirement.SnapshotSchemaID = strings.TrimSpace(requirement.SnapshotSchemaID)
		if requirement.RecordType == "" || requirement.SourceOwner == "" || requirement.SnapshotSchemaID == "" ||
			!strings.HasPrefix(requirement.SnapshotSchemaID, "cartulary.revisions.snapshot.") {
			return nil, fmt.Errorf("%w: invalid snapshot requirement for record type %q", ErrInvalidCapturedSnapshot, requirement.RecordType)
		}
		if _, duplicate := required[requirement.RecordType]; duplicate {
			return nil, fmt.Errorf("%w: record type %q", ErrDuplicateSnapshotCapture, requirement.RecordType)
		}
		required[requirement.RecordType] = requirement
	}
	catalog := &RecordSnapshotCaptureCatalog{byRecordType: map[string]recordSnapshotCapture{}}
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			schemaID := strings.TrimSpace(record.SnapshotSchemaID)
			requirement, admitted := required[record.RecordType]
			if !admitted || schemaID == "" {
				return nil, fmt.Errorf("%w: record type %q snapshot schema", ErrMissingSnapshotCapture, record.RecordType)
			}
			if record.SourceOwnerModule != contribution.SourceOwnerModule ||
				record.SourceOwnerModule != requirement.SourceOwner ||
				strings.TrimSpace(record.RecordType) == "" ||
				nilDeleteRestoreSource(record.DeleteRestoreSource) ||
				schemaID != requirement.SnapshotSchemaID {
				return nil, fmt.Errorf("%w: record type %q schema %q", ErrInvalidCapturedSnapshot, record.RecordType, schemaID)
			}
			if _, duplicate := catalog.byRecordType[record.RecordType]; duplicate {
				return nil, fmt.Errorf("%w: record type %q", ErrDuplicateSnapshotCapture, record.RecordType)
			}
			catalog.byRecordType[record.RecordType] = recordSnapshotCapture{
				sourceOwner:      contribution.SourceOwnerModule,
				recordType:       record.RecordType,
				snapshotSchemaID: schemaID,
				source:           record.DeleteRestoreSource,
			}
		}
	}
	for recordType := range required {
		if _, ok := catalog.byRecordType[recordType]; !ok {
			return nil, fmt.Errorf("%w: record type %q", ErrMissingSnapshotCapture, recordType)
		}
	}
	if len(catalog.byRecordType) != len(required) {
		return nil, fmt.Errorf("%w: unexpected record snapshot capture", ErrInvalidCapturedSnapshot)
	}
	return catalog, nil
}

func (catalog *RecordSnapshotCaptureCatalog) DeclaredRecordTypes() []string {
	if catalog == nil {
		return nil
	}
	result := make([]string, 0, len(catalog.byRecordType))
	for recordType := range catalog.byRecordType {
		result = append(result, recordType)
	}
	sort.Strings(result)
	return result
}

func (catalog *RecordSnapshotCaptureCatalog) captureTx(
	ctx context.Context,
	tx pgx.Tx,
	envelope RecordEnvelope,
) (CapturedRecordSnapshot, error) {
	if catalog == nil {
		return CapturedRecordSnapshot{}, fmt.Errorf("%w: record type %q", ErrMissingSnapshotCapture, envelope.RecordType)
	}
	capture, ok := catalog.byRecordType[envelope.RecordType]
	if !ok {
		return CapturedRecordSnapshot{}, fmt.Errorf("%w: record type %q", ErrMissingSnapshotCapture, envelope.RecordType)
	}
	raw, err := capture.source.SnapshotTx(ctx, tx, envelope.RecordID)
	if err != nil {
		return CapturedRecordSnapshot{}, err
	}
	value, err := validateCanonicalSnapshotValue(capture.snapshotSchemaID, envelope, raw)
	if err != nil {
		return CapturedRecordSnapshot{}, err
	}
	return CapturedRecordSnapshot{
		recordID:         envelope.RecordID,
		recordType:       envelope.RecordType,
		snapshotSchemaID: capture.snapshotSchemaID,
		value:            value,
	}, nil
}

func validateCanonicalSnapshotValue(schemaID string, envelope RecordEnvelope, raw map[string]any) (map[string]any, error) {
	if len(raw) != 2 {
		return nil, fmt.Errorf("%w: record type %q source snapshot must contain exactly record and source", ErrInvalidCapturedSnapshot, envelope.RecordType)
	}
	record, recordOK := raw["record"].(map[string]any)
	source, sourceOK := raw["source"].(map[string]any)
	if !recordOK || !sourceOK || record == nil || source == nil {
		return nil, fmt.Errorf("%w: record type %q source snapshot members", ErrInvalidCapturedSnapshot, envelope.RecordType)
	}
	if got, ok := record["record_id"].(string); !ok || got != envelope.RecordID.String() {
		return nil, fmt.Errorf("%w: record id", ErrInvalidCapturedSnapshot)
	}
	if got, ok := record["record_type"].(string); !ok || got != envelope.RecordType {
		return nil, fmt.Errorf("%w: got %v want %q", ErrSnapshotRecordTypeMismatch, record["record_type"], envelope.RecordType)
	}
	value := map[string]any{
		"snapshot_schema_id": schemaID,
		"record":             record,
		"source":             source,
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%w: encode: %v", ErrInvalidCapturedSnapshot, err)
	}
	var cloned map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(encoded)))
	decoder.UseNumber()
	if err := decoder.Decode(&cloned); err != nil {
		return nil, fmt.Errorf("%w: clone: %v", ErrInvalidCapturedSnapshot, err)
	}
	return cloned, nil
}

func capturedSnapshotValue(snapshot *CapturedRecordSnapshot, recordID uuid.UUID) (map[string]any, error) {
	if snapshot == nil {
		return nil, nil
	}
	if snapshot.recordID != recordID || snapshot.recordID == uuid.Nil ||
		strings.TrimSpace(snapshot.snapshotSchemaID) == "" || len(snapshot.value) != 3 {
		return nil, fmt.Errorf("%w: record %s", ErrInvalidCapturedSnapshot, recordID)
	}
	copyValue, err := cloneJSONMap(snapshot.value)
	if err != nil {
		return nil, err
	}
	return copyValue, nil
}

func validatePersistedSnapshotValue(expectedSchemaID string, recordID uuid.UUID, recordType string, value map[string]any) error {
	if value == nil {
		return nil
	}
	if len(value) != 3 || expectedSchemaID == "" {
		return fmt.Errorf("%w: record %s canonical envelope", ErrInvalidCapturedSnapshot, recordID)
	}
	if schemaID, ok := value["snapshot_schema_id"].(string); !ok || schemaID != expectedSchemaID {
		return fmt.Errorf("%w: record %s snapshot schema", ErrInvalidCapturedSnapshot, recordID)
	}
	record, recordOK := value["record"].(map[string]any)
	source, sourceOK := value["source"].(map[string]any)
	if !recordOK || !sourceOK || record == nil || source == nil {
		return fmt.Errorf("%w: record %s canonical members", ErrInvalidCapturedSnapshot, recordID)
	}
	if got, ok := record["record_id"].(string); !ok || got != recordID.String() {
		return fmt.Errorf("%w: record %s identity", ErrInvalidCapturedSnapshot, recordID)
	}
	if got, ok := record["record_type"].(string); !ok || got != recordType {
		return fmt.Errorf("%w: record %s type", ErrInvalidCapturedSnapshot, recordID)
	}
	return nil
}

func (catalog *RecordSnapshotCaptureCatalog) validatePersisted(recordID uuid.UUID, recordType string, value map[string]any) error {
	if catalog == nil {
		return fmt.Errorf("%w: record type %q", ErrMissingSnapshotCapture, recordType)
	}
	capture, ok := catalog.byRecordType[recordType]
	if !ok {
		return fmt.Errorf("%w: record type %q", ErrMissingSnapshotCapture, recordType)
	}
	return validatePersistedSnapshotValue(capture.snapshotSchemaID, recordID, recordType, value)
}

func cloneJSONMap(value map[string]any) (map[string]any, error) {
	if value == nil {
		return nil, nil
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("%w: encode JSON value: %v", ErrInvalidCapturedSnapshot, err)
	}
	var cloned map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(encoded)))
	decoder.UseNumber()
	if err := decoder.Decode(&cloned); err != nil {
		return nil, fmt.Errorf("%w: decode JSON value: %v", ErrInvalidCapturedSnapshot, err)
	}
	return cloned, nil
}

func nilHistoryTargetSemantics(value HistoryTargetSemantics) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}
