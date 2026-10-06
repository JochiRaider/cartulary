package recovery

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

var (
	ErrRestoreTargetNotEmpty     = errors.New("recovery: restore target is not empty")
	ErrExtensionCodecUnsupported = errors.New("recovery: extension backup codec unsupported")
	ErrExtensionBindingInvalid   = errors.New("recovery: extension backup binding invalid")
)

// ExtensionBackupCodec is one exact packaged codec identity. Historical
// compatibility exists only as another explicit row; current codecs are never
// used as best-effort readers for historical bytes.
type ExtensionBackupCodec struct {
	CodecID       string
	CodecSHA256   string
	MaxItems      int
	MaxEntryBytes int64
	MaxTotalBytes int64
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
	HistoricalCodecs               []ExtensionBackupCodec
	PostgresTables                 []string
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
	bindings := make([]ExtensionBackupBinding, len(source))
	for index, binding := range source {
		binding.PostgresTables = canonicalRecoveryStrings(binding.PostgresTables)
		binding.HistoricalCodecs = append([]ExtensionBackupCodec(nil), binding.HistoricalCodecs...)
		sort.Slice(binding.HistoricalCodecs, func(i, j int) bool {
			if binding.HistoricalCodecs[i].CodecID != binding.HistoricalCodecs[j].CodecID {
				return binding.HistoricalCodecs[i].CodecID < binding.HistoricalCodecs[j].CodecID
			}
			return binding.HistoricalCodecs[i].CodecSHA256 < binding.HistoricalCodecs[j].CodecSHA256
		})
		if binding.ProfileID == "" ||
			!validSHA256Hex(binding.ImplementationBindingSHA256) ||
			!validSHA256Hex(binding.PhysicalStateBindingSHA256) ||
			binding.BindingID == "" ||
			binding.LogicalFamilyID == "" ||
			binding.StorageKind != "postgres" ||
			binding.RestoreOrderGroup < 1 ||
			binding.PostRestoreValidationAlgorithm == "" ||
			len(binding.PostgresTables) == 0 ||
			!validExtensionBackupCodec(binding.CurrentCodec) {
			return nil, fmt.Errorf("%w: %s", ErrExtensionBindingInvalid, binding.BindingID)
		}
		seenCodecIDs := map[string]struct{}{binding.CurrentCodec.CodecID: {}}
		seenCodecDigests := map[string]struct{}{binding.CurrentCodec.CodecSHA256: {}}
		for _, codec := range binding.HistoricalCodecs {
			if !validExtensionBackupCodec(codec) {
				return nil, fmt.Errorf("%w: historical codec for %s", ErrExtensionBindingInvalid, binding.BindingID)
			}
			if _, duplicate := seenCodecIDs[codec.CodecID]; duplicate {
				return nil, fmt.Errorf("%w: duplicate codec ID for %s", ErrExtensionBindingInvalid, binding.BindingID)
			}
			if _, duplicate := seenCodecDigests[codec.CodecSHA256]; duplicate {
				return nil, fmt.Errorf("%w: duplicate codec digest for %s", ErrExtensionBindingInvalid, binding.BindingID)
			}
			seenCodecIDs[codec.CodecID] = struct{}{}
			seenCodecDigests[codec.CodecSHA256] = struct{}{}
		}
		bindings[index] = binding
	}
	sort.Slice(bindings, func(i, j int) bool {
		if bindings[i].RestoreOrderGroup != bindings[j].RestoreOrderGroup {
			return bindings[i].RestoreOrderGroup < bindings[j].RestoreOrderGroup
		}
		if bindings[i].ProfileID != bindings[j].ProfileID {
			return bindings[i].ProfileID < bindings[j].ProfileID
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
		for _, table := range binding.PostgresTables {
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
	result := make([]ExtensionBackupBinding, len(c.bindings))
	for index, binding := range c.bindings {
		binding.PostgresTables = append([]string(nil), binding.PostgresTables...)
		binding.HistoricalCodecs = append([]ExtensionBackupCodec(nil), binding.HistoricalCodecs...)
		result[index] = binding
	}
	return result
}

func validExtensionBackupCodec(codec ExtensionBackupCodec) bool {
	return codec.CodecID != "" && validSHA256Hex(codec.CodecSHA256) &&
		codec.MaxItems > 0 && codec.MaxEntryBytes > 0 &&
		codec.MaxTotalBytes >= codec.MaxEntryBytes
}

func canonicalRecoveryStrings(source []string) []string {
	result := append([]string(nil), source...)
	sort.Strings(result)
	write := 0
	for _, value := range result {
		value = strings.TrimSpace(value)
		if value == "" || (write > 0 && result[write-1] == value) {
			continue
		}
		result[write] = value
		write++
	}
	return result[:write]
}
