package recovery

import (
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
)

func logicalTestBinding(t testing.TB, id, family, table string, group int) ExtensionBackupBinding {
	t.Helper()
	id = "test_profile." + id
	family = "test_profile." + family
	definition := extensioncodec.Definition{SchemaID: "cartulary.test_profile.rows.v1", ProfileID: "test_profile", Families: []extensioncodec.Family{{LogicalFamilyID: family, IdentityFieldIDs: []string{"item_id"}, Fields: []extensioncodec.Field{{FieldID: "item_id", ValueKind: "text"}, {FieldID: "value", ValueKind: "text"}}}}}
	mapping := extensioncodec.Mapping{SchemaID: extensioncodec.PostgresMappingSchemaID, BindingID: id, CodecContractRef: definition.SchemaID, SchemaName: "public", TableName: table, Columns: []extensioncodec.Column{{FieldID: "item_id", ColumnName: "id"}, {FieldID: "value", ColumnName: "value"}}}
	codec, err := extensioncodec.NewPostgres(definition, mapping, extensioncodec.Binding{BindingID: id, LogicalFamilyID: family, CodecContractRef: definition.SchemaID, LogicalIdentityAlgorithmID: family + "_identity_v1", ContentEncodingAlgorithmID: "extension_registry_canonical_json_v1", Limits: extensioncodec.Limits{MaxItems: 10, MaxEntryBytes: 128, MaxBindingBytes: 1024}})
	if err != nil {
		t.Fatal(err)
	}
	return ExtensionBackupBinding{ProfileID: "test_profile", ImplementationBindingSHA256: strings.Repeat("3", 64), PhysicalStateBindingSHA256: strings.Repeat("4", 64), BindingID: id, LogicalFamilyID: family, StorageKind: "postgres", RestoreOrderGroup: group, PostRestoreValidationAlgorithm: extensioncodec.PostgresRestoreAlgorithmID, CurrentCodec: ExtensionBackupCodec{CodecID: id + ".codec", CodecSHA256: strings.Repeat("1", 64), MaxItems: 10, MaxEntryBytes: 128, MaxTotalBytes: 1024, Postgres: codec}}
}

func currentTestProof(binding ExtensionBackupBinding) PostgresBindingProof {
	return PostgresBindingProof{SchemaID: PostgresBindingProofV1SchemaID, ProfileID: binding.ProfileID, BindingID: binding.BindingID, PhysicalStateBindingSHA256: binding.PhysicalStateBindingSHA256, ImplementationBindingSHA256: binding.ImplementationBindingSHA256, BackupCodecID: binding.CurrentCodec.CodecID, BackupCodecSHA256: binding.CurrentCodec.CodecSHA256, StorageKind: "postgres", BindingDigest: strings.Repeat("5", 64)}
}

func TestExtensionBackupCatalogUsesOnlyExactCurrentCodec(t *testing.T) {
	binding := logicalTestBinding(t, "binding", "family", "test_table", 1)
	if _, err := NewExtensionBackupCatalog([]ExtensionBackupBinding{binding}, nil); err != nil {
		t.Fatal(err)
	}
	proof := currentTestProof(binding)
	if err := admitPostgresBindingProof(proof, binding); err != nil {
		t.Fatalf("current codec rejected: %v", err)
	}
	for _, identity := range [][2]string{{"codec.retired", strings.Repeat("2", 64)}, {binding.CurrentCodec.CodecID, strings.Repeat("2", 64)}, {"codec.retired", binding.CurrentCodec.CodecSHA256}} {
		changed := proof
		changed.BackupCodecID, changed.BackupCodecSHA256 = identity[0], identity[1]
		if err := admitPostgresBindingProof(changed, binding); !errors.Is(err, ErrExtensionCodecUnsupported) {
			t.Fatalf("noncurrent pair: %v", err)
		}
	}
	for _, algorithm := range []string{"", "test_profile.validate_state_v1", "extension_validate_filesystem_binding_restore_v1"} {
		changed := binding
		changed.PostRestoreValidationAlgorithm = algorithm
		if _, err := NewExtensionBackupCatalog([]ExtensionBackupBinding{changed}, nil); !errors.Is(err, ErrExtensionBindingInvalid) {
			t.Fatalf("unsupported restore algorithm %q: %v", algorithm, err)
		}
	}
}

func TestExtensionBackupCatalogCanonicalizesBindingsAndRejectsSharedTables(t *testing.T) {
	first := logicalTestBinding(t, "first", "z_family", "z_table", 1)
	second := logicalTestBinding(t, "second", "a_family", "a_table", 1)
	last := logicalTestBinding(t, "last", "a_family", "last_table", 2)
	catalog, err := NewExtensionBackupCatalog([]ExtensionBackupBinding{last, first, second}, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := catalog.Bindings()
	if len(got) != 3 || got[0].BindingID != second.BindingID || got[1].BindingID != first.BindingID || got[2].BindingID != last.BindingID {
		t.Fatalf("binding order: %#v", got)
	}
	duplicate := logicalTestBinding(t, "duplicate", "duplicate", "a_table", 3)
	if _, err := NewExtensionBackupCatalog(append(got, duplicate), nil); !errors.Is(err, ErrExtensionBindingInvalid) {
		t.Fatalf("shared table admitted: %v", err)
	}
	got[0].ProfileID = "mutated"
	if catalog.Bindings()[0].ProfileID != "test_profile" {
		t.Fatal("catalog exposed mutable records")
	}
}
