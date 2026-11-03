package extensioncodec_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/extensioncodec"
)

func TestLogicalCodecAndBindingProof_Unit(t *testing.T) {
	t.Run("logical values and physical renaming", func(t *testing.T) {
		definition, mapping, binding := codecFixture()
		codec, err := extensioncodec.NewPostgres(definition, mapping, binding)
		if err != nil {
			t.Fatal(err)
		}
		values := []any{[]string{"é", "<"}, "z\"\\\u2028", int64(9223372036854775807), json.RawMessage(`{"x":"<>&","empty":[],"big":9223372036854775807}`), nil,
			time.Date(2026, 10, 31, 10, 11, 12, 123456000, time.FixedZone("offset", 7200)), [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}}
		want := `{"array":["é","<"],"id":"z\"\\` + "\u2028" + `","integer":9223372036854775807,"json":{"big":9223372036854775807,"empty":[],"x":"<>&"},"nullable":null,"time":"2026-10-31T08:11:12.123456Z","uuid":"00010203-0405-0607-0809-0a0b0c0d0e0f"}` + "\n"
		wantIdentity := `["z\"\\` + "\u2028" + `"]` + "\n"
		entry, err := codec.EncodeRow(values)
		if err != nil || string(entry.Content) != want || string(entry.Identity) != wantIdentity {
			t.Fatalf("logical entry = %q / %q, %v", entry.Identity, entry.Content, err)
		}
		mapping.SchemaName, mapping.TableName = "other_schema", "other_table"
		for i := range mapping.Columns {
			mapping.Columns[i].ColumnName = "physical_" + mapping.Columns[i].FieldID
		}
		renamed, err := extensioncodec.NewPostgres(definition, mapping, binding)
		if err != nil {
			t.Fatal(err)
		}
		renamedEntry, err := renamed.EncodeRow(values)
		if err != nil || !bytes.Equal(entry.Content, renamedEntry.Content) || !bytes.Equal(entry.Identity, renamedEntry.Identity) {
			t.Fatalf("physical renaming changed logical bytes: %v", err)
		}
		if left, right := proofFor(t, []extensioncodec.Entry{entry}), proofFor(t, []extensioncodec.Entry{renamedEntry}); left != right {
			t.Fatal("physical renaming changed binding proof")
		}
		definition.Families[0].Fields[0].FieldID = "mutated"
		mapping.Columns[0].ColumnName = "mutated"
		copyFamily, copyMapping := codec.Family(), codec.Mapping()
		copyFamily.IdentityFieldIDs[0], copyFamily.Fields[0].FieldID, copyMapping.Columns[0].FieldID = "mutated", "mutated", "mutated"
		if got, err := codec.EncodeRow(values); err != nil || !bytes.Equal(got.Content, entry.Content) || !bytes.Equal(got.Identity, entry.Identity) {
			t.Fatal("codec aliases mutable input/accessor data")
		}
	})
	t.Run("declared composite identity order and nullable Boolean", func(t *testing.T) {
		definition, mapping, binding := codecFixture()
		definition.Families[0].Fields = []extensioncodec.Field{{FieldID: "flag", ValueKind: "boolean", Nullable: true}, {FieldID: "id", ValueKind: "text"}, {FieldID: "parent", ValueKind: "text"}}
		definition.Families[0].IdentityFieldIDs = []string{"parent", "id"}
		mapping.Columns = []extensioncodec.Column{{FieldID: "flag", ColumnName: "flag"}, {FieldID: "id", ColumnName: "id"}, {FieldID: "parent", ColumnName: "parent"}}
		codec, err := extensioncodec.NewPostgres(definition, mapping, binding)
		if err != nil {
			t.Fatal(err)
		}
		for _, flag := range []struct {
			value   any
			encoded string
		}{{nil, "null"}, {false, "false"}, {true, "true"}} {
			entry, err := codec.EncodeRow([]any{flag.value, "child", "parent"})
			want := `{"flag":` + flag.encoded + `,"id":"child","parent":"parent"}` + "\n"
			if err != nil || string(entry.Content) != want || string(entry.Identity) != "[\"parent\",\"child\"]\n" {
				t.Fatalf("composite entry = %q / %q, %v", entry.Identity, entry.Content, err)
			}
		}
		if _, err := codec.EncodeRow([]any{int64(1), "child", "parent"}); !errors.Is(err, extensioncodec.ErrInvalidEntry) {
			t.Fatal("integer accepted as Boolean")
		}
	})
	t.Run("unrepresentable cells", func(t *testing.T) {
		definition, mapping, binding := codecFixture()
		codec, err := extensioncodec.NewPostgres(definition, mapping, binding)
		if err != nil {
			t.Fatal(err)
		}
		base := []any{[]string{}, "a", int64(-9223372036854775808), json.RawMessage(`{}`), nil, time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), "00000000-0000-0000-0000-000000000000"}
		if _, err := codec.EncodeRow(base); err != nil {
			t.Fatal(err)
		}
		for _, change := range []struct {
			index int
			value any
		}{
			{0, []any{"valid", nil}}, {1, nil}, {1, "\xff"}, {2, float64(1)}, {2, uint64(9223372036854775808)},
			{3, json.RawMessage(`{"n":1.5}`)}, {3, json.RawMessage(`{"n":1,"n":2}`)}, {3, map[string]any{"n": float64(1)}},
			{5, time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)}, {5, time.Date(10000, 1, 1, 0, 0, 0, 0, time.UTC)},
			{5, time.Date(2026, 1, 1, 0, 0, 0, 1, time.UTC)}, {5, "infinity"}, {6, "invalid"},
		} {
			values := append([]any(nil), base...)
			values[change.index] = change.value
			if _, err := codec.EncodeRow(values); !errors.Is(err, extensioncodec.ErrInvalidEntry) {
				t.Errorf("invalid cell %d (%T) error = %v", change.index, change.value, err)
			}
		}
		if _, err := codec.EncodeRow(base[:len(base)-1]); !errors.Is(err, extensioncodec.ErrInvalidEntry) {
			t.Fatal("missing field accepted")
		}
		base[1] = strings.Repeat("a", 4091)
		if entry, err := codec.EncodeRow(base); err != nil || len(entry.Identity) != 4096 {
			t.Fatal("maximum-size identity rejected")
		}
		base[1] = strings.Repeat("a", 4092)
		if _, err := codec.EncodeRow(base); !errors.Is(err, extensioncodec.ErrLimitExceeded) {
			t.Fatal("oversize identity accepted")
		}
	})
	t.Run("closed contracts", func(t *testing.T) {
		definition, mapping, binding := codecFixture()
		raw, _ := canonicaljson.MarshalExtension(definition)
		if _, err := extensioncodec.ParseDefinition(raw); err != nil {
			t.Fatal(err)
		}
		for _, mutate := range []func(map[string]any){
			func(v map[string]any) { v["unknown"] = true },
			func(v map[string]any) {
				delete(v["families"].([]any)[0].(map[string]any)["fields"].([]any)[0].(map[string]any), "nullable")
			},
			func(v map[string]any) {
				v["families"].([]any)[0].(map[string]any)["identity_field_ids"] = []any{"missing"}
			},
		} {
			var value map[string]any
			if err := json.Unmarshal(raw, &value); err != nil {
				t.Fatal(err)
			}
			mutate(value)
			changed, _ := canonicaljson.MarshalExtension(value)
			if _, err := extensioncodec.ParseDefinition(changed); !errors.Is(err, extensioncodec.ErrInvalidContract) {
				t.Fatal("malformed definition accepted")
			}
		}
		mappingJSON, _ := canonicaljson.MarshalExtension(mapping)
		if _, err := extensioncodec.ParseMapping(mappingJSON); err != nil {
			t.Fatal(err)
		}
		mapping.Columns[0].ColumnName = "name; DELETE FROM state"
		if _, err := extensioncodec.NewPostgres(definition, mapping, binding); !errors.Is(err, extensioncodec.ErrInvalidContract) {
			t.Fatal("SQL expression accepted as identifier")
		}
		_, mapping, _ = codecFixture()
		mapping.Columns[0].ColumnName = mapping.Columns[1].ColumnName
		if _, err := extensioncodec.NewPostgres(definition, mapping, binding); !errors.Is(err, extensioncodec.ErrInvalidContract) {
			t.Fatal("duplicate physical column accepted")
		}
		_, mapping, _ = codecFixture()
		binding.LogicalIdentityAlgorithmID = "test.unsupported"
		if _, err := extensioncodec.NewPostgres(definition, mapping, binding); !errors.Is(err, extensioncodec.ErrInvalidContract) {
			t.Fatal("unsupported identity algorithm accepted")
		}
	})
	t.Run("fixed proof vectors", func(t *testing.T) {
		for _, vector := range []struct {
			name    string
			entries []extensioncodec.Entry
			digest  string
			bytes   int64
		}{
			{"empty", nil, "07b348aeadf48c4f9e50a53abbea7879435ded8760dd88aa6ed8d605ffcacb4e", 0},
			{"single", []extensioncodec.Entry{{Identity: []byte("[\"a\"]\n"), Content: []byte("{\"id\":\"a\",\"n\":1}\n")}}, "b836ff2dc5e5b51fd6f02cf30c8e769df317961e96a37b40bf7f3dc2de3254ca", 17},
			{"composite unordered", []extensioncodec.Entry{{Identity: []byte("[\"z\",2]\n"), Content: []byte("{\"id\":\"z\",\"n\":2}\n")}, {Identity: []byte("[\"a\",10]\n"), Content: []byte("{\"id\":\"a\",\"n\":10}\n")}}, "ec9934617e009b28ece0dc99af6b2e125b6b7ede6bd7aee2fab73bbb3f640217", 35},
		} {
			t.Run(vector.name, func(t *testing.T) {
				got := proofFor(t, vector.entries)
				if got.BindingDigest != vector.digest || got.ItemCount != int64(len(vector.entries)) || got.ContentBytes != vector.bytes {
					t.Fatalf("proof = %+v", got)
				}
			})
		}
	})
	t.Run("bounds duplicates interruption and cleanup", func(t *testing.T) {
		for _, test := range []struct {
			name    string
			limits  extensioncodec.Limits
			entries []extensioncodec.Entry
			want    error
		}{
			{"count", extensioncodec.Limits{MaxItems: 1, MaxEntryBytes: 1, MaxBindingBytes: 2}, []extensioncodec.Entry{{Identity: []byte("a")}, {Identity: []byte("b")}}, extensioncodec.ErrLimitExceeded},
			{"entry bytes", extensioncodec.Limits{MaxItems: 2, MaxEntryBytes: 1, MaxBindingBytes: 2}, []extensioncodec.Entry{{Identity: []byte("a"), Content: []byte("xx")}}, extensioncodec.ErrLimitExceeded},
			{"aggregate bytes", extensioncodec.Limits{MaxItems: 2, MaxEntryBytes: 2, MaxBindingBytes: 3}, []extensioncodec.Entry{{Identity: []byte("a"), Content: []byte("xx")}, {Identity: []byte("b"), Content: []byte("xx")}}, extensioncodec.ErrLimitExceeded},
			{"duplicate", extensioncodec.Limits{MaxItems: 2, MaxEntryBytes: 1, MaxBindingBytes: 2}, []extensioncodec.Entry{{Identity: []byte("a")}, {Identity: []byte("a")}}, extensioncodec.ErrDuplicateIdentity},
		} {
			t.Run(test.name, func(t *testing.T) {
				path, root := proofScratch(t)
				proof, err := extensioncodec.ComputeProof(context.Background(), "test.rows", "postgres", test.limits, root, func(yield func(extensioncodec.Entry) error) error {
					for _, entry := range test.entries {
						_ = yield(entry)
					}
					return nil
				})
				if !errors.Is(err, test.want) || proof != (extensioncodec.Proof{}) {
					t.Fatalf("proof=%+v error=%v", proof, err)
				}
				requireEmptyScratch(t, path)
			})
		}
		path, root := proofScratch(t)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := extensioncodec.ComputeProof(ctx, "test.rows", "postgres", proofLimits(), root, func(yield func(extensioncodec.Entry) error) error {
			if err := yield(extensioncodec.Entry{Identity: []byte("a")}); err != nil {
				return err
			}
			cancel()
			return nil
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation error=%v", err)
		}
		requireEmptyScratch(t, path)
	})
	t.Run("bounded merge across many chunks", func(t *testing.T) {
		const count = 131073
		path, root := proofScratch(t)
		stream := func(reverse, duplicate bool) extensioncodec.Entries {
			return func(yield func(extensioncodec.Entry) error) error {
				for i := 0; i < count; i++ {
					ordinal := i
					if reverse {
						ordinal = count - 1 - i
					}
					if duplicate && i == count-1 {
						ordinal = 0
					}
					if err := yield(extensioncodec.Entry{Identity: []byte(fmt.Sprintf("%06d", ordinal)), Content: []byte("x")}); err != nil {
						return err
					}
				}
				return nil
			}
		}
		first, err := extensioncodec.ComputeProof(context.Background(), "test.rows", "postgres", proofLimits(), root, stream(false, false))
		if err != nil {
			t.Fatal(err)
		}
		second, err := extensioncodec.ComputeProof(context.Background(), "test.rows", "postgres", proofLimits(), root, stream(true, false))
		if err != nil {
			t.Fatal(err)
		}
		if first != second || first.ItemCount != count || first.ContentBytes != count {
			t.Fatal("multi-pass merge is not order independent")
		}
		if _, err := extensioncodec.ComputeProof(context.Background(), "test.rows", "postgres", proofLimits(), root, stream(false, true)); !errors.Is(err, extensioncodec.ErrDuplicateIdentity) {
			t.Fatalf("cross-chunk duplicate error=%v", err)
		}
		requireEmptyScratch(t, path)
	})
}

func codecFixture() (extensioncodec.Definition, extensioncodec.Mapping, extensioncodec.Binding) {
	fields := []extensioncodec.Field{{FieldID: "array", ValueKind: "text_array"}, {FieldID: "id", ValueKind: "text"}, {FieldID: "integer", ValueKind: "int64"}, {FieldID: "json", ValueKind: "json"}, {FieldID: "nullable", ValueKind: "text", Nullable: true}, {FieldID: "time", ValueKind: "timestamp"}, {FieldID: "uuid", ValueKind: "uuid"}}
	definition := extensioncodec.Definition{SchemaID: "test.rows_contract.v1", ProfileID: "test", Families: []extensioncodec.Family{{LogicalFamilyID: "test.rows", IdentityFieldIDs: []string{"id"}, Fields: fields}}}
	mapping := extensioncodec.Mapping{SchemaID: extensioncodec.PostgresMappingSchemaID, BindingID: "test.rows", CodecContractRef: definition.SchemaID, SchemaName: "public", TableName: "rows", Columns: make([]extensioncodec.Column, len(fields))}
	for i, field := range fields {
		mapping.Columns[i] = extensioncodec.Column{FieldID: field.FieldID, ColumnName: field.FieldID}
	}
	binding := extensioncodec.Binding{BindingID: "test.rows", LogicalFamilyID: "test.rows", CodecContractRef: definition.SchemaID, LogicalIdentityAlgorithmID: "test.rows_identity_v1", ContentEncodingAlgorithmID: "extension_registry_canonical_json_v1", Limits: proofLimits()}
	return definition, mapping, binding
}

func proofLimits() extensioncodec.Limits {
	return extensioncodec.Limits{MaxItems: 1048576, MaxEntryBytes: 67108864, MaxBindingBytes: 9223372036854775807}
}
func proofScratch(t *testing.T) (string, *os.Root) {
	t.Helper()
	path := t.TempDir()
	root, err := os.OpenRoot(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	return path, root
}
func requireEmptyScratch(t *testing.T, path string) {
	t.Helper()
	files, err := os.ReadDir(path)
	if err != nil || len(files) != 0 {
		t.Fatalf("proof scratch not empty: %v", err)
	}
}
func proofFor(t *testing.T, entries []extensioncodec.Entry) extensioncodec.Proof {
	t.Helper()
	path, root := proofScratch(t)
	proof, err := extensioncodec.ComputeProof(context.Background(), "test.rows", "postgres", proofLimits(), root, func(yield func(extensioncodec.Entry) error) error {
		for _, entry := range entries {
			if err := yield(entry); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	requireEmptyScratch(t, path)
	return proof
}
