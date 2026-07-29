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

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
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

type ExtensionBindingProof struct {
	ProfileID                   string `json:"profile_id"`
	ImplementationBindingSHA256 string `json:"implementation_binding_sha256"`
	PhysicalBindingSHA256       string `json:"physical_binding_sha256"`
	BindingID                   string `json:"binding_id"`
	CodecID                     string `json:"codec_id"`
	CodecSHA256                 string `json:"codec_sha256"`
	ItemCount                   int    `json:"item_count"`
	ContentByteLength           int64  `json:"content_byte_length"`
	ContentSHA256               string `json:"content_sha256"`
}

func captureExtensionBindingProofs(catalog *ExtensionBackupCatalog, snapshot PostgresSnapshotArtifact) ([]ExtensionBindingProof, error) {
	if catalog == nil {
		return nil, fmt.Errorf("%w: catalog is required", ErrExtensionBindingInvalid)
	}
	proofs := make([]ExtensionBindingProof, 0, len(catalog.bindings))
	for _, binding := range catalog.bindings {
		codec := binding.CurrentCodec
		count, byteLength, digest, err := extensionBindingSnapshotProof(snapshot, binding, codec)
		if err != nil {
			return nil, err
		}
		if count > codec.MaxItems || byteLength > codec.MaxTotalBytes {
			return nil, fmt.Errorf("%w: %s exceeds codec bounds", ErrExtensionBindingInvalid, binding.BindingID)
		}
		proofs = append(proofs, ExtensionBindingProof{
			ProfileID: binding.ProfileID, ImplementationBindingSHA256: binding.ImplementationBindingSHA256,
			PhysicalBindingSHA256: binding.PhysicalStateBindingSHA256, BindingID: binding.BindingID,
			CodecID: codec.CodecID, CodecSHA256: codec.CodecSHA256,
			ItemCount: count, ContentByteLength: byteLength, ContentSHA256: digest,
		})
	}
	return proofs, nil
}

func validateExtensionBindingProofs(catalog *ExtensionBackupCatalog, proofs []ExtensionBindingProof, snapshot PostgresSnapshotArtifact) error {
	if catalog == nil || len(proofs) != len(catalog.bindings) {
		return fmt.Errorf("%w: proof/catalog cardinality", ErrExtensionBindingInvalid)
	}
	for index, binding := range catalog.bindings {
		proof := proofs[index]
		if proof.ProfileID != binding.ProfileID ||
			proof.ImplementationBindingSHA256 != binding.ImplementationBindingSHA256 ||
			proof.PhysicalBindingSHA256 != binding.PhysicalStateBindingSHA256 ||
			proof.BindingID != binding.BindingID {
			return fmt.Errorf("%w: identity mismatch at %s", ErrExtensionBindingInvalid, binding.BindingID)
		}
		codec, ok := selectExtensionBackupCodec(binding, proof.CodecID, proof.CodecSHA256)
		if !ok {
			return fmt.Errorf("%w: %s", ErrExtensionCodecUnsupported, binding.BindingID)
		}
		count, byteLength, digest, err := extensionBindingSnapshotProof(snapshot, binding, codec)
		if err != nil {
			return err
		}
		if proof.ItemCount != count || proof.ContentByteLength != byteLength ||
			proof.ContentSHA256 != digest || count > codec.MaxItems || byteLength > codec.MaxTotalBytes {
			return fmt.Errorf("%w: content proof mismatch at %s", ErrExtensionBindingInvalid, binding.BindingID)
		}
	}
	return nil
}

func validateRestoredExtensionBindings(ctx context.Context, catalog *ExtensionBackupCatalog, proofs []ExtensionBindingProof, db postgres.DB) error {
	body, err := CapturePostgresSnapshotArtifact(ctx, db)
	if err != nil {
		return err
	}
	snapshot, err := DecodePostgresSnapshotArtifact(body)
	if err != nil {
		return err
	}
	return validateExtensionBindingProofs(catalog, proofs, snapshot)
}

func extensionBindingSnapshotProof(snapshot PostgresSnapshotArtifact, binding ExtensionBackupBinding, codec ExtensionBackupCodec) (int, int64, string, error) {
	byName := make(map[string]PostgresSnapshotTable, len(snapshot.Tables))
	for _, table := range snapshot.Tables {
		byName[table.TableName] = table
	}
	digest := sha256.New()
	count := 0
	var byteLength int64
	for _, tableName := range binding.PostgresTables {
		table, present := byName[tableName]
		if !present {
			return 0, 0, "", fmt.Errorf("%w: binding %s is missing table %s", ErrExtensionBindingInvalid, binding.BindingID, tableName)
		}
		_, _ = digest.Write([]byte("table:" + tableName + "\n"))
		rows := make([]string, 0, len(table.Rows))
		for _, raw := range table.Rows {
			var value any
			if err := json.Unmarshal(raw, &value); err != nil {
				return 0, 0, "", fmt.Errorf("%w: decode %s row", ErrExtensionBindingInvalid, tableName)
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				return 0, 0, "", err
			}
			if int64(len(encoded)) > codec.MaxEntryBytes {
				return 0, 0, "", fmt.Errorf("%w: %s entry exceeds codec bound", ErrExtensionBindingInvalid, binding.BindingID)
			}
			rows = append(rows, string(encoded))
		}
		sort.Strings(rows)
		for _, row := range rows {
			_, _ = digest.Write([]byte(row))
			_, _ = digest.Write([]byte("\n"))
			count++
			byteLength += int64(len(row))
		}
	}
	return count, byteLength, hex.EncodeToString(digest.Sum(nil)), nil
}

func selectExtensionBackupCodec(binding ExtensionBackupBinding, codecID, digest string) (ExtensionBackupCodec, bool) {
	if binding.CurrentCodec.CodecID == codecID && binding.CurrentCodec.CodecSHA256 == digest {
		return binding.CurrentCodec, true
	}
	for _, codec := range binding.HistoricalCodecs {
		if codec.CodecID == codecID && codec.CodecSHA256 == digest {
			return codec, true
		}
	}
	return ExtensionBackupCodec{}, false
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
