package config

import (
	"strconv"
	"strings"
	"testing"
)

type testEnterpriseConfiguration struct {
	Claimed              bool   `toml:"claimed"`
	ProviderManifestPath string `toml:"provider_manifest_path"`
}

type testNetworkFlowConfiguration struct {
	Claimed             bool   `toml:"claimed"`
	KeyRingManifestPath string `toml:"key_ring_manifest_path"`
}

type testRevisionsConfiguration struct {
	ConflictTokenKeyRingManifestPath string `toml:"conflict_token_key_ring_manifest_path"`
}

func loadWithTestCatalog(t testing.TB, options LoadOptions) (document, error) {
	t.Helper()
	if options.ExtensionPolicy == nil {
		options.ExtensionPolicy = testExtensionPolicyForCharacterization(t)
	}
	catalog := testOwnerNamespaceCatalog(t)
	return loadWithOptionsAndCatalog(options, catalog)
}

func testOwnerNamespaceCatalog(t testing.TB) Catalog {
	t.Helper()
	builder := &CatalogBuilder{}
	registerTestOwnerNamespace(t, builder, "enterprise", "enterprise_authentication", []string{
		"enterprise_authentication.claimed",
		"enterprise_authentication.provider_manifest_path",
	}, func(decoder NamespaceDecoder) (testEnterpriseConfiguration, []Diagnostic) {
		return decodeTestNamespace[testEnterpriseConfiguration](decoder, "enterprise_authentication")
	}, func(value testEnterpriseConfiguration, path []string, raw string) (testEnterpriseConfiguration, *Diagnostic) {
		joined := strings.Join(path, ".")
		switch joined {
		case "enterprise_authentication.claimed":
			claimed, err := strconv.ParseBool(raw)
			if err != nil {
				return value, &Diagnostic{Path: joined, ReasonCode: "type_mismatch", Message: err.Error()}
			}
			value.Claimed = claimed
		case "enterprise_authentication.provider_manifest_path":
			value.ProviderManifestPath = raw
		default:
			return value, &Diagnostic{Path: joined, ReasonCode: "unknown_key", Message: "unknown fixture-owner overlay"}
		}
		return value, nil
	})
	registerTestOwnerNamespace(t, builder, "networkflow", "network_flow_activity", []string{
		"network_flow_activity.claimed",
		"network_flow_activity.key_ring_manifest_path",
	}, func(decoder NamespaceDecoder) (testNetworkFlowConfiguration, []Diagnostic) {
		return decodeTestNamespace[testNetworkFlowConfiguration](decoder, "network_flow_activity")
	}, func(value testNetworkFlowConfiguration, path []string, raw string) (testNetworkFlowConfiguration, *Diagnostic) {
		joined := strings.Join(path, ".")
		switch joined {
		case "network_flow_activity.claimed":
			claimed, err := strconv.ParseBool(raw)
			if err != nil {
				return value, &Diagnostic{Path: joined, ReasonCode: "type_mismatch", Message: err.Error()}
			}
			value.Claimed = claimed
		case "network_flow_activity.key_ring_manifest_path":
			value.KeyRingManifestPath = raw
		default:
			return value, &Diagnostic{Path: joined, ReasonCode: "unknown_key", Message: "unknown fixture-owner overlay"}
		}
		return value, nil
	})
	registerTestOwnerNamespace(t, builder, "revisions", "revisions", []string{
		"revisions.conflict_token_key_ring_manifest_path",
	}, func(decoder NamespaceDecoder) (testRevisionsConfiguration, []Diagnostic) {
		return decodeTestNamespace[testRevisionsConfiguration](decoder, "revisions")
	}, func(value testRevisionsConfiguration, path []string, raw string) (testRevisionsConfiguration, *Diagnostic) {
		joined := strings.Join(path, ".")
		if joined != "revisions.conflict_token_key_ring_manifest_path" {
			return value, &Diagnostic{Path: joined, ReasonCode: "unknown_key", Message: "unknown fixture-owner overlay"}
		}
		value.ConflictTokenKeyRingManifestPath = raw
		return value, nil
	})
	catalog, err := builder.Build()
	if err != nil {
		t.Fatalf("build test owner namespace catalog: %v", err)
	}
	return catalog
}

func registerTestOwnerNamespace[T any](
	t testing.TB,
	builder *CatalogBuilder,
	id string,
	namespace string,
	paths []string,
	decode func(NamespaceDecoder) (T, []Diagnostic),
	overlay func(T, []string, string) (T, *Diagnostic),
) {
	t.Helper()
	key, err := NewKey[T]("test." + id)
	if err != nil {
		t.Fatalf("create %s test namespace key: %v", namespace, err)
	}
	err = Register(builder, Definition[T]{
		Key:          key,
		Namespace:    namespace,
		Paths:        paths,
		Decode:       decode,
		ApplyOverlay: overlay,
		Project:      func(value T, _ NamespacePresence) (T, []Diagnostic) { return value, nil },
		Clone:        func(value T) T { return value },
	})
	if err != nil {
		t.Fatalf("register %s test namespace: %v", namespace, err)
	}
}

func testOwnerValue[T any](t testing.TB, cfg document, namespace string) T {
	t.Helper()
	value, present := cfg.namespaces[namespace].(*T)
	if !present || value == nil {
		t.Fatalf("test owner namespace %q is unavailable", namespace)
	}
	return *value
}
