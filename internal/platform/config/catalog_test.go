package config

import (
	"reflect"
	"testing"
)

func TestCatalogSnapshotLifecycle_Unit(t *testing.T) {
	type ownerSettings struct {
		Enabled bool
		Labels  map[string]string
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
	builder := &CatalogBuilder{}
	if err := Register(builder, Definition[string]{
		Key:       keyB,
		Namespace: "beta",
		Paths:     []string{"beta.value"},
		Project: func(Source) (string, []Diagnostic) {
			order = append(order, "owner.beta")
			return "beta", nil
		},
		Clone: func(value string) string { return value },
	}); err != nil {
		t.Fatalf("register beta: %v", err)
	}
	if err := Register(builder, Definition[ownerSettings]{
		Key:       keyA,
		Namespace: "alpha",
		Paths:     []string{"alpha.enabled", "alpha.claimed"},
		ClaimPath: "alpha.claimed",
		Project: func(Source) (ownerSettings, []Diagnostic) {
			order = append(order, "owner.alpha")
			return ownerSettings{Enabled: true, Labels: map[string]string{"mode": "strict"}}, nil
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
	if got, want := catalog.IDs(), []string{"owner.alpha", "owner.beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("catalog order: got %v want %v", got, want)
	}
	snapshot, err := catalog.materialize(document{})
	if err != nil {
		t.Fatalf("materialize snapshot: %v", err)
	}
	if got, want := order, []string{"owner.alpha", "owner.beta"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("projection order: got %v want %v", got, want)
	}
	settings, err := Value(snapshot, keyA)
	if err != nil {
		t.Fatalf("read typed value: %v", err)
	}
	if !settings.Enabled {
		t.Fatal("typed value was not retained")
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
				Project:   func(Source) (ownerSettings, []Diagnostic) { return ownerSettings{}, nil },
				Clone:     func(value ownerSettings) ownerSettings { return value },
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
			Project:   func(Source) (ownerSettings, []Diagnostic) { return ownerSettings{}, nil },
			Clone:     func(value ownerSettings) ownerSettings { return value },
		}); err != nil {
			t.Fatalf("register parent: %v", err)
		}
		if err := Register(overlapBuilder, Definition[string]{
			Key:       childKey,
			Namespace: "alpha.child",
			Paths:     []string{"alpha.child.value"},
			Project:   func(Source) (string, []Diagnostic) { return "", nil },
			Clone:     func(value string) string { return value },
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
			Project: func(Source) (string, []Diagnostic) {
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
		_, err = diagnosticCatalog.materialize(document{})
		diagnostics, ok := DiagnosticsFromError(err)
		if !ok {
			t.Fatalf("materialize did not return diagnostics: %v", err)
		}
		if diagnostics[0].Path != "a" || diagnostics[1].Path != "z" {
			t.Fatalf("diagnostics are not sorted: %+v", diagnostics)
		}
	})

	if got, want := ValidationPhases(), validationPhases; !reflect.DeepEqual(got, want) {
		t.Fatalf("validation phase copy: got %v want %v", got, want)
	}
}
