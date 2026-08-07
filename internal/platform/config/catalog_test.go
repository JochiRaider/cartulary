package config

import (
	"errors"
	"reflect"
	"testing"
)

func TestCatalogSnapshotLifecycle_Unit(t *testing.T) {
	if err := ValidateSnapshotForStartup(Snapshot{}); !errors.Is(err, errSnapshotNotAdmitted) {
		t.Fatalf("unadmitted snapshot startup error = %v", err)
	}

	type ownerSettings struct {
		Enabled bool              `toml:"enabled"`
		Labels  map[string]string `toml:"labels"`
	}

	keyA, err := NewKey[ownerSettings]("owner.alpha")
	if err != nil {
		t.Fatalf("create alpha key: %v", err)
	}
	keyB, err := NewKey[string]("owner.beta")
	if err != nil {
		t.Fatalf("create beta key: %v", err)
	}

	var order []string
	var ownedPresenceObserved bool
	var foreignPresenceObserved bool
	builder := &CatalogBuilder{}
	if err := Register(builder, Definition[string]{
		Key:       keyB,
		Namespace: "beta",
		Paths:     []string{"beta.value"},
		Decode: func(decoder NamespaceDecoder) (string, []Diagnostic) {
			var wire struct {
				Value string `toml:"value"`
			}
			if err := decoder.Decode(&wire); err != nil {
				return "", []Diagnostic{{Path: "beta", ReasonCode: "type_mismatch", Message: err.Error()}}
			}
			return wire.Value, nil
		},
		Project: func(value string, _ NamespacePresence) (string, []Diagnostic) {
			order = append(order, "owner.beta")
			return value, nil
		},
		Clone: func(value string) string { return value },
	}); err != nil {
		t.Fatalf("register beta: %v", err)
	}
	if err := Register(builder, Definition[ownerSettings]{
		Key:       keyA,
		Namespace: "alpha",
		Paths:     []string{"alpha.enabled", "alpha.claimed"},
		Decode: func(decoder NamespaceDecoder) (ownerSettings, []Diagnostic) {
			return decodeTestNamespace[ownerSettings](decoder, "alpha")
		},
		Project: func(value ownerSettings, presence NamespacePresence) (ownerSettings, []Diagnostic) {
			order = append(order, "owner.alpha")
			ownedPresenceObserved = presence.Defined("alpha", "enabled")
			foreignPresenceObserved = presence.Defined("beta", "value")
			return value, nil
		},
		Clone: func(value ownerSettings) ownerSettings {
			cloned := value
			cloned.Labels = make(map[string]string, len(value.Labels))
			for name, label := range value.Labels {
				cloned.Labels[name] = label
			}
			return cloned
		},
	}); err != nil {
		t.Fatalf("register alpha: %v", err)
	}
	catalog, err := builder.Build()
	if err != nil {
		t.Fatalf("build catalog: %v", err)
	}
	cfg := document{
		Application: ApplicationConfig{PublicOrigin: "https://cartulary.example"},
		presence: configPresence{overlayPaths: map[string]struct{}{
			"alpha.enabled": {},
			"beta.value":    {},
		}},
	}
	unknown, findings := catalog.decodeNamespaces(&cfg, map[string]any{
		"alpha": map[string]any{"enabled": true, "labels": map[string]any{"mode": "strict"}},
		"beta":  map[string]any{"value": "beta"},
	})
	if len(unknown) > 0 || len(findings) > 0 {
		t.Fatalf("decode fixture owner namespaces: unknown=%v findings=%v", unknown, findings)
	}
	unknownDocument := document{}
	unknown, findings = catalog.decodeNamespaces(&unknownDocument, map[string]any{
		"alpha": map[string]any{"unexpected": true},
	})
	if !reflect.DeepEqual(unknown, []string{"alpha.unexpected"}) || len(findings) != 0 {
		t.Fatalf("fixture-owner namespace was not closed: unknown=%v findings=%v", unknown, findings)
	}
	snapshot, err := catalog.materialize(cfg)
	if err != nil {
		t.Fatalf("materialize snapshot: %v", err)
	}
	if got, want := order, []string{"owner.alpha", "owner.beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("projection order: got %v want %v", got, want)
	}
	if !ownedPresenceObserved || foreignPresenceObserved {
		t.Fatalf("namespace presence isolation: owned=%t foreign=%t", ownedPresenceObserved, foreignPresenceObserved)
	}
	core := snapshot.Core()
	if core.Application.PublicOrigin != "https://cartulary.example" {
		t.Fatalf("core projection = %#v", core)
	}
	core.Application.PublicOrigin = "https://mutated.example"
	if again := snapshot.Core(); again.Application.PublicOrigin != "https://cartulary.example" {
		t.Fatalf("snapshot returned mutable Core state: %#v", again)
	}
	settings, err := Value(snapshot, keyA)
	if err != nil {
		t.Fatalf("read typed value: %v", err)
	}
	if !settings.Enabled {
		t.Fatal("typed value was not retained")
	}
	if requested := RequestedClaimRegistrationIDs(snapshot); len(requested) != 0 {
		t.Fatalf("unregistered owner Boolean leaked into claim requests: %#v", requested)
	}
	settings.Labels["mode"] = "mutated"
	again, err := Value(snapshot, keyA)
	if err != nil {
		t.Fatalf("read typed value again: %v", err)
	}
	if again.Labels["mode"] != "strict" {
		t.Fatal("snapshot returned mutable owner state")
	}
	if _, err := Value(snapshot, Key[int]{id: keyA.id, valueType: reflect.TypeOf(0)}); err == nil {
		t.Fatal("type-confused lookup succeeded")
	}

	t.Run("duplicate identity", func(t *testing.T) {
		duplicateBuilder := &CatalogBuilder{}
		for range 2 {
			if err := Register(duplicateBuilder, Definition[ownerSettings]{
				Key:       keyA,
				Namespace: "alpha",
				Paths:     []string{"alpha.enabled"},
				Decode: func(decoder NamespaceDecoder) (ownerSettings, []Diagnostic) {
					return decodeTestNamespace[ownerSettings](decoder, "alpha")
				},
				Project: func(value ownerSettings, _ NamespacePresence) (ownerSettings, []Diagnostic) { return value, nil },
				Clone:   func(value ownerSettings) ownerSettings { return value },
			}); err != nil {
				t.Fatalf("register duplicate candidate: %v", err)
			}
		}
		if _, err := duplicateBuilder.Build(); err == nil {
			t.Fatal("duplicate contribution ID was accepted")
		}
	})

	t.Run("overlapping namespace", func(t *testing.T) {
		childKey, err := NewKey[string]("owner.child")
		if err != nil {
			t.Fatalf("create child key: %v", err)
		}
		overlapBuilder := &CatalogBuilder{}
		if err := Register(overlapBuilder, Definition[ownerSettings]{
			Key:       keyA,
			Namespace: "alpha",
			Paths:     []string{"alpha.enabled"},
			Decode: func(decoder NamespaceDecoder) (ownerSettings, []Diagnostic) {
				return decodeTestNamespace[ownerSettings](decoder, "alpha")
			},
			Project: func(value ownerSettings, _ NamespacePresence) (ownerSettings, []Diagnostic) { return value, nil },
			Clone:   func(value ownerSettings) ownerSettings { return value },
		}); err != nil {
			t.Fatalf("register parent: %v", err)
		}
		if err := Register(overlapBuilder, Definition[string]{
			Key:       childKey,
			Namespace: "alpha.child",
			Paths:     []string{"alpha.child.value"},
			Decode: func(decoder NamespaceDecoder) (string, []Diagnostic) {
				var value string
				if err := decoder.Decode(&value); err != nil {
					return "", []Diagnostic{{Path: "alpha.child", ReasonCode: "type_mismatch", Message: err.Error()}}
				}
				return value, nil
			},
			Project: func(value string, _ NamespacePresence) (string, []Diagnostic) { return value, nil },
			Clone:   func(value string) string { return value },
		}); err != nil {
			t.Fatalf("register child: %v", err)
		}
		if _, err := overlapBuilder.Build(); err == nil {
			t.Fatal("overlapping namespaces were accepted")
		}
	})

	t.Run("diagnostic ordering", func(t *testing.T) {
		diagnosticKey, err := NewKey[string]("owner.diagnostics")
		if err != nil {
			t.Fatalf("create diagnostic key: %v", err)
		}
		diagnosticBuilder := &CatalogBuilder{}
		if err := Register(diagnosticBuilder, Definition[string]{
			Key:       diagnosticKey,
			Namespace: "diagnostics",
			Paths:     []string{"diagnostics.value"},
			Decode: func(decoder NamespaceDecoder) (string, []Diagnostic) {
				var wire struct {
					Value string `toml:"value"`
				}
				if err := decoder.Decode(&wire); err != nil {
					return "", []Diagnostic{{Path: "diagnostics", ReasonCode: "type_mismatch", Message: err.Error()}}
				}
				return wire.Value, nil
			},
			Project: func(_ string, _ NamespacePresence) (string, []Diagnostic) {
				return "", []Diagnostic{
					{Path: "z", ReasonCode: "b", Message: "second"},
					{Path: "a", ReasonCode: "a", Message: "first"},
				}
			},
			Clone: func(value string) string { return value },
		}); err != nil {
			t.Fatalf("register diagnostics: %v", err)
		}
		diagnosticCatalog, err := diagnosticBuilder.Build()
		if err != nil {
			t.Fatalf("build diagnostic catalog: %v", err)
		}
		diagnosticDocument := document{}
		_, _ = diagnosticCatalog.decodeNamespaces(&diagnosticDocument, map[string]any{"diagnostics": map[string]any{"value": ""}})
		_, err = diagnosticCatalog.materialize(diagnosticDocument)
		diagnostics, ok := DiagnosticsFromError(err)
		if !ok {
			t.Fatalf("materialize did not return diagnostics: %v", err)
		}
		if diagnostics[0].Path != "a" || diagnostics[1].Path != "z" {
			t.Fatalf("diagnostics are not sorted: %+v", diagnostics)
		}
	})
}

func decodeTestNamespace[T any](decoder NamespaceDecoder, path string) (T, []Diagnostic) {
	var value T
	if err := decoder.Decode(&value); err != nil {
		return value, []Diagnostic{{Path: path, ReasonCode: "type_mismatch", Message: err.Error()}}
	}
	return value, nil
}
