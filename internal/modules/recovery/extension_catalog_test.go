package recovery_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func TestGeneratedLogicalCodecCatalog_Unit(t *testing.T) {
	catalog := testExtensionBackupCatalog(t)
	bindings := catalog.Bindings()
	if len(bindings) != 5 {
		t.Fatalf("binding count = %d", len(bindings))
	}
	for i, binding := range bindings {
		codec := binding.CurrentCodec.Postgres
		if codec == nil {
			t.Fatalf("binding %d discarded its logical codec", i)
		}
		identity, mapping := codec.Binding(), codec.Mapping()
		if identity.BindingID != binding.BindingID || identity.LogicalFamilyID != binding.LogicalFamilyID ||
			identity.CodecContractRef != "cartulary.network_flow_activity.postgres_rows.v3" ||
			identity.ContentEncodingAlgorithmID != "extension_registry_canonical_json_v1" ||
			identity.LogicalIdentityAlgorithmID != binding.LogicalFamilyID+"_identity_v1" ||
			identity.Limits.MaxItems != int64(binding.CurrentCodec.MaxItems) ||
			identity.Limits.MaxEntryBytes != binding.CurrentCodec.MaxEntryBytes ||
			identity.Limits.MaxBindingBytes != binding.CurrentCodec.MaxTotalBytes ||
			mapping.SchemaName != "public" {
			t.Fatalf("binding %d lost admitted codec/mapping identity", i)
		}
		mapping.Columns[0].ColumnName = "mutated"
		if catalog.Bindings()[i].CurrentCodec.Postgres.Mapping().Columns[0].ColumnName == "mutated" {
			t.Fatal("catalog exposes mutable codec data")
		}
	}
}

func testExtensionBackupCatalog(t testing.TB) *recovery.ExtensionBackupCatalog {
	t.Helper()
	catalog, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		t.Fatalf("construct generated extension recovery catalog: %v", err)
	}
	return catalog
}
