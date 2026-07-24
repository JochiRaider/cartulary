package recovery

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestExtensionBackupCatalogUsesOnlyExactCurrentOrHistoricalCodec(t *testing.T) {
	current := ExtensionBackupCodec{
		CodecID:       "codec.current",
		CodecSHA256:   strings.Repeat("1", 64),
		MaxItems:      10,
		MaxEntryBytes: 1,
		MaxTotalBytes: 10,
	}
	historical := ExtensionBackupCodec{
		CodecID:       "codec.historical",
		CodecSHA256:   strings.Repeat("2", 64),
		MaxItems:      10,
		MaxEntryBytes: 128,
		MaxTotalBytes: 1024,
	}
	binding := ExtensionBackupBinding{
		ProfileID:                      "test_profile",
		ImplementationBindingSHA256:    strings.Repeat("3", 64),
		PhysicalStateBindingSHA256:     strings.Repeat("4", 64),
		BindingID:                      "test.binding",
		LogicalFamilyID:                "test.family",
		StorageKind:                    "postgres",
		RestoreOrderGroup:              1,
		PostRestoreValidationAlgorithm: "test.validate.v1",
		CurrentCodec:                   current,
		HistoricalCodecs:               []ExtensionBackupCodec{historical},
		PostgresTables:                 []string{"test_table"},
	}
	catalog, err := NewExtensionBackupCatalog([]ExtensionBackupBinding{binding}, nil)
	if err != nil {
		t.Fatalf("construct extension backup catalog: %v", err)
	}
	snapshot := PostgresSnapshotArtifact{
		SchemaID: PostgresSnapshotArtifactSchemaID,
		Tables: []PostgresSnapshotTable{{
			TableName: "test_table",
			RowCount:  1,
			Rows:      []json.RawMessage{json.RawMessage(`{"value":"historical"}`)},
		}},
	}
	count, byteLength, digest, err := extensionBindingSnapshotProof(snapshot, binding, historical)
	if err != nil {
		t.Fatalf("construct historical proof: %v", err)
	}
	proof := ExtensionBindingProof{
		ProfileID:                   binding.ProfileID,
		ImplementationBindingSHA256: binding.ImplementationBindingSHA256,
		PhysicalBindingSHA256:       binding.PhysicalStateBindingSHA256,
		BindingID:                   binding.BindingID,
		CodecID:                     historical.CodecID,
		CodecSHA256:                 historical.CodecSHA256,
		ItemCount:                   count,
		ContentByteLength:           byteLength,
		ContentSHA256:               digest,
	}
	if err := validateExtensionBindingProofs(catalog, []ExtensionBindingProof{proof}, snapshot); err != nil {
		t.Fatalf("exact packaged historical codec was rejected: %v", err)
	}

	proof.CodecSHA256 = strings.Repeat("5", 64)
	if err := validateExtensionBindingProofs(catalog, []ExtensionBindingProof{proof}, snapshot); !errors.Is(err, ErrExtensionCodecUnsupported) {
		t.Fatalf("unpackaged historical codec error got %v want %v", err, ErrExtensionCodecUnsupported)
	}
	proof.CodecSHA256 = historical.CodecSHA256
	proof.CodecID = current.CodecID
	if err := validateExtensionBindingProofs(catalog, []ExtensionBindingProof{proof}, snapshot); !errors.Is(err, ErrExtensionCodecUnsupported) {
		t.Fatalf("mixed codec identity error got %v want %v", err, ErrExtensionCodecUnsupported)
	}
}

func TestExtensionBackupCatalogCanonicalizesBindingsAndRejectsSharedTables(t *testing.T) {
	codec := func(id, digest string) ExtensionBackupCodec {
		return ExtensionBackupCodec{
			CodecID: id, CodecSHA256: digest,
			MaxItems: 10, MaxEntryBytes: 10, MaxTotalBytes: 100,
		}
	}
	binding := func(id, table string, group int, digit string) ExtensionBackupBinding {
		return ExtensionBackupBinding{
			ProfileID: "test_profile", ImplementationBindingSHA256: strings.Repeat("a", 64),
			PhysicalStateBindingSHA256: strings.Repeat("b", 64),
			BindingID:                  id, LogicalFamilyID: id + ".family", StorageKind: "postgres",
			RestoreOrderGroup: group, PostRestoreValidationAlgorithm: id + ".validate",
			CurrentCodec:   codec(id+".codec", strings.Repeat(digit, 64)),
			PostgresTables: []string{table},
		}
	}
	catalog, err := NewExtensionBackupCatalog([]ExtensionBackupBinding{
		binding("second", "z_table", 2, "2"),
		binding("first", "a_table", 1, "1"),
	}, nil)
	if err != nil {
		t.Fatalf("construct canonical catalog: %v", err)
	}
	got := catalog.Bindings()
	if len(got) != 2 || got[0].BindingID != "first" || got[1].BindingID != "second" {
		t.Fatalf("binding order got %#v", got)
	}

	duplicate := binding("duplicate", "a_table", 3, "3")
	if _, err := NewExtensionBackupCatalog(append(got, duplicate), nil); !errors.Is(err, ErrExtensionBindingInvalid) {
		t.Fatalf("shared physical table error got %v want %v", err, ErrExtensionBindingInvalid)
	}
}
