package recovery

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

var (
	ErrRestoreObjectNamespaceNotFresh = errors.New("recovery: restore object namespace is not pristine")
	ErrRestoreTargetNotEmpty          = errors.New("recovery: restore target is not empty")
	ErrExtensionCodecUnsupported      = errors.New("recovery: extension backup codec unsupported")
	ErrExtensionBindingInvalid        = errors.New("recovery: extension backup binding invalid")
)

// ExtensionBackupCodec is the exact current packaged codec identity. An older
// ID or digest never selects a compatibility reader.
type ExtensionBackupCodec struct {
	CodecID       string
	CodecSHA256   string
	MaxItems      int
	MaxEntryBytes int64
	MaxTotalBytes int64
	Postgres      *extensioncodec.Postgres
}

// ExtensionBackupBinding is Recovery's immutable physical catalog view.
// Application composition translates the Extensions registry and owner-supplied
// physical table bindings into this shape.
type ExtensionBackupBinding struct {
	ProfileID                      string
	ImplementationBindingSHA256    string
	PhysicalStateBindingSHA256     string
	BindingID                      string
	LogicalFamilyID                string
	StorageKind                    string
	RestoreOrderGroup              int
	PostRestoreValidationAlgorithm string
	CurrentCodec                   ExtensionBackupCodec
}

type ExtensionBackupCatalog struct {
	bindings             []ExtensionBackupBinding
	pristineMetadataRows map[string]ExtensionPristineMetadata
}

// ExtensionPristineMetadata declares an exact schema-migration-owned metadata
// seed that is present before any application process starts. It is not a
// compatibility allowance for initialized deployment state.
type ExtensionPristineMetadata struct {
	ProfileID          string
	MigrationLineageID string
	StateVersion       int
	MetadataVersion    int
	LastMigrationID    string
}

func NewExtensionBackupCatalog(source []ExtensionBackupBinding, pristineMetadata []ExtensionPristineMetadata) (*ExtensionBackupCatalog, error) {
	if len(source) > 4096 {
		return nil, ErrExtensionBindingInvalid
	}
	bindings := make([]ExtensionBackupBinding, len(source))
	for index, binding := range source {
		if binding.ProfileID == "" ||
			!validSHA256Hex(binding.ImplementationBindingSHA256) ||
			!validSHA256Hex(binding.PhysicalStateBindingSHA256) ||
			binding.BindingID == "" ||
			binding.LogicalFamilyID == "" ||
			binding.StorageKind != "postgres" ||
			binding.RestoreOrderGroup < 0 || binding.RestoreOrderGroup > 1024 ||
			binding.PostRestoreValidationAlgorithm != extensioncodec.PostgresRestoreAlgorithmID ||
			!validExtensionBackupCodec(binding.CurrentCodec) {
			return nil, fmt.Errorf("%w: %s", ErrExtensionBindingInvalid, binding.BindingID)
		}
		logical := binding.CurrentCodec.Postgres.Binding()
		if logical.BindingID != binding.BindingID || logical.LogicalFamilyID != binding.LogicalFamilyID ||
			!strings.HasPrefix(binding.BindingID, binding.ProfileID+".") || binding.CurrentCodec.Postgres.Mapping().SchemaName != "public" {
			return nil, ErrExtensionBindingInvalid
		}
		bindings[index] = binding
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].RestoreOrderGroup != bindings[j].RestoreOrderGroup {
			return bindings[i].RestoreOrderGroup < bindings[j].RestoreOrderGroup
		}
		if bindings[i].LogicalFamilyID != bindings[j].LogicalFamilyID {
			return bindings[i].LogicalFamilyID < bindings[j].LogicalFamilyID
		}
		return bindings[i].BindingID < bindings[j].BindingID
	})
	seenBindings := make(map[string]struct{}, len(bindings))
	seenTables := make(map[string]struct{})
	for _, binding := range bindings {
		identity := binding.ProfileID + "\x1f" + binding.BindingID
		if _, duplicate := seenBindings[identity]; duplicate {
			return nil, fmt.Errorf("%w: duplicate %s", ErrExtensionBindingInvalid, binding.BindingID)
		}
		seenBindings[identity] = struct{}{}
		for _, table := range []string{binding.CurrentCodec.Postgres.Mapping().TableName} {
			if _, duplicate := seenTables[table]; duplicate {
				return nil, fmt.Errorf("%w: table %s has multiple bindings", ErrExtensionBindingInvalid, table)
			}
			seenTables[table] = struct{}{}
		}
	}
	pristine := make(map[string]ExtensionPristineMetadata, len(pristineMetadata))
	for _, row := range pristineMetadata {
		if row.ProfileID == "" || row.MigrationLineageID == "" ||
			row.StateVersion < 1 || row.MetadataVersion < 1 {
			return nil, fmt.Errorf("%w: invalid pristine metadata row for %s", ErrExtensionBindingInvalid, row.ProfileID)
		}
		if _, duplicate := pristine[row.ProfileID]; duplicate {
			return nil, fmt.Errorf("%w: duplicate pristine metadata row for %s", ErrExtensionBindingInvalid, row.ProfileID)
		}
		pristine[row.ProfileID] = row
	}
	return &ExtensionBackupCatalog{bindings: bindings, pristineMetadataRows: pristine}, nil
}

func (c *ExtensionBackupCatalog) Bindings() []ExtensionBackupBinding {
	if c == nil {
		return nil
	}
	return append([]ExtensionBackupBinding(nil), c.bindings...)
}

func validExtensionBackupCodec(codec ExtensionBackupCodec) bool {
	if codec.Postgres == nil {
		return false
	}
	limits := codec.Postgres.Binding().Limits
	return limits.MaxItems == int64(codec.MaxItems) && limits.MaxEntryBytes == codec.MaxEntryBytes && limits.MaxBindingBytes == codec.MaxTotalBytes && codec.CodecID != "" && validSHA256Hex(codec.CodecSHA256) &&
		codec.MaxItems > 0 && codec.MaxEntryBytes > 0 &&
		codec.MaxTotalBytes >= codec.MaxEntryBytes
}

func (catalog *ExtensionBackupCatalog) validateStateCatalog(state *recoverystate.Catalog) error {
	if catalog == nil || state == nil || len(catalog.bindings) > 4096 {
		return ErrExtensionBindingInvalid
	}
	required := make(map[string]bool)
	for _, table := range state.RequiredTableNames() {
		required[table] = true
	}
	for _, binding := range catalog.bindings {
		if binding.CurrentCodec.Postgres == nil {
			return ErrExtensionBindingInvalid
		}
		for _, table := range []string{binding.CurrentCodec.Postgres.Mapping().TableName} {
			if !required[table] {
				return fmt.Errorf("%w: binding table is not in the required catalog", ErrExtensionBindingInvalid)
			}
		}
	}
	return nil
}
