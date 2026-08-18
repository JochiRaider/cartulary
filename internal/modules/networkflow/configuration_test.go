package networkflow

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestNetworkFlowConfigurationContribution_Unit(t *testing.T) {
	t.Run("owner parses only its closed overlay paths", func(t *testing.T) {
		configuration, finding := ApplyConfigurationOverlay(Configuration{}, []string{"network_flow_activity", "claimed"}, "true")
		if finding != nil || !configuration.Claimed {
			t.Fatalf("claim overlay = %#v, finding %#v", configuration, finding)
		}
		configuration, finding = ApplyConfigurationOverlay(configuration, []string{"network_flow_activity", "key_ring_manifest_path"}, "/etc/cartulary/network-flow-key-rings.json")
		if finding != nil || configuration.KeyRingManifestPath != "/etc/cartulary/network-flow-key-rings.json" {
			t.Fatalf("manifest overlay = %#v, finding %#v", configuration, finding)
		}
		if _, finding := ApplyConfigurationOverlay(configuration, []string{"network_flow_activity", "unknown"}, "value"); finding == nil || finding.ReasonCode != "unknown_key" {
			t.Fatalf("unknown overlay finding = %#v", finding)
		}
	})

	t.Run("unclaimed is inert and clears owner-local path material", func(t *testing.T) {
		configuration, findings := NormalizeAndValidateConfiguration(Configuration{
			KeyRingManifestPath: "/must/not/be/read",
		})
		if configuration.Claimed || configuration.KeyRingManifestPath != "" || len(findings) != 0 {
			t.Fatalf("unclaimed configuration = %#v, %#v", configuration, findings)
		}
	})

	t.Run("claimed requires both key-ring purposes", func(t *testing.T) {
		_, findings := NormalizeAndValidateConfiguration(Configuration{Claimed: true})
		if len(findings) != 2 ||
			findings[0].ReasonCode != "network_flow_cursor_key_missing" ||
			findings[1].ReasonCode != "network_flow_safe_digest_key_missing" {
			t.Fatalf("missing manifest findings = %#v", findings)
		}
	})

	t.Run("claimed path validation preserves diagnostics and normalization", func(t *testing.T) {
		for name, path := range map[string]string{
			"relative": "relative/manifest.json",
			"NUL":      "/etc/cartulary/\x00manifest.json",
			"variable": "/etc/$CARTULARY/manifest.json",
			"parent":   "/etc/cartulary/../manifest.json",
		} {
			t.Run(name, func(t *testing.T) {
				_, findings := NormalizeAndValidateConfiguration(Configuration{
					Claimed:             true,
					KeyRingManifestPath: path,
				})
				if len(findings) != 1 || findings[0].Path != keyRingConfigPath {
					t.Fatalf("path findings = %#v", findings)
				}
			})
		}
		configuration, findings := NormalizeAndValidateConfiguration(Configuration{
			Claimed:             true,
			KeyRingManifestPath: "/etc//cartulary/network-flow-key-rings.json",
		})
		if len(findings) != 0 ||
			configuration.KeyRingManifestPath != "/etc/cartulary/network-flow-key-rings.json" {
			t.Fatalf("normalized configuration = %#v, %#v", configuration, findings)
		}
	})

	t.Run("manifest read failures remain owner-defined and redacted", func(t *testing.T) {
		for failure, reason := range map[ManifestReadFailure]string{
			ManifestUnreadable: "network_flow_cursor_key_missing",
			ManifestUnsafe:     "network_flow_cursor_key_invalid",
			ManifestTooLarge:   "network_flow_cursor_key_invalid",
		} {
			err := KeyRingManifestReadError(failure)
			if err == nil || !strings.Contains(err.Error(), reason) {
				t.Fatalf("failure %q error = %v; want %q", failure, err, reason)
			}
		}
	})

	t.Run("incident portability binding is closed and inert before query", func(t *testing.T) {
		binding := NewPortabilityStateBinding()
		if _, err := binding.RetainedAuthoritativeStatePresentTx(
			context.Background(),
			nil,
			uuid.MustParse("11111111-1111-1111-1111-111111111111"),
			[]string{ExtensionFamilyTables},
		); err == nil {
			t.Fatal("missing transaction query capability was admitted")
		}
	})
}

func TestNetworkFlowEffectiveResourceLimits_Unit(t *testing.T) {
	t.Parallel()

	type limitCase struct {
		configKey    string
		discoveryKey string
		minimum      int64
		maximum      int64
	}
	cases := []limitCase{
		{"max_active_tables_per_incident", "network_flow.max_active_tables_per_incident", 1, 128},
		{"max_retained_tables_per_incident", "network_flow.max_retained_tables_per_incident", 1, 512},
		{"max_selected_tables_per_query", "network_flow.max_selected_tables_per_query", 1, 64},
		{"max_columns_per_csv", "network_flow.max_columns_per_csv", 1, 512},
		{"max_header_scalar_length", "network_flow.max_header_scalar_length", 1, 256},
		{"max_raw_cell_scalar_length", "network_flow.max_raw_cell_scalar_length", 1, 16384},
		{"max_rows_per_csv", "network_flow.max_rows_per_csv", 1, 5000000},
		{"max_accepted_rows_per_table", "network_flow.max_accepted_rows_per_table", 1, 5000000},
		{"max_rejected_row_diagnostics", "network_flow.max_rejected_row_diagnostics", 0, 100000},
		{"max_filters_per_query", "network_flow.max_filters_per_query", 0, 16},
		{"max_sorts_per_query", "network_flow.max_sorts_per_query", 0, 8},
		{"max_query_limit", "network_flow.max_query_limit", 1, 1000},
		{"max_graph_vertices", "network_flow.max_graph_vertices", 1, 100000},
		{"max_graph_edges", "network_flow.max_graph_edges", 0, 250000},
		{"max_active_graph_views_per_incident", "network_flow.max_active_graph_views_per_incident", 1, 32},
		{"max_retained_graph_views_per_incident", "network_flow.max_retained_graph_views_per_incident", 1, 128},
		{"max_nonterminal_graph_jobs_per_incident", "network_flow.max_nonterminal_graph_jobs_per_incident", 1, 4},
		{"max_example_row_refs_per_edge", "network_flow.max_example_row_refs_per_edge", 0, 100},
		{"max_binding_source_row_refs", "network_flow.max_binding_source_row_refs", 1, 1000},
		{"max_aggregate_counter_digits", "network_flow.max_aggregate_counter_digits", 1, 128},
		{"max_contributing_rows_per_graph", "network_flow.max_contributing_rows_per_graph", 1, 5000000},
		{"max_time_buckets_per_graph", "network_flow.max_time_buckets_per_graph", 1, 1024},
		{"graph_materialization_timeout_seconds", "network_flow.graph_materialization_timeout_seconds", 30, 3600},
	}
	base := Configuration{Claimed: true, KeyRingManifestPath: "/etc/cartulary/network-flow-key-rings.json"}

	normalized, findings := NormalizeAndValidateConfiguration(base)
	if len(findings) != 0 {
		t.Fatalf("default effective limits findings = %#v", findings)
	}
	defaults := effectiveLimitsResource(normalized.EffectiveResourceLimits())
	if len(defaults) != len(cases) || defaults["network_flow.max_header_scalar_length"] != int64(256) {
		t.Fatalf("default effective limits = %#v", defaults)
	}

	for _, testCase := range cases {
		t.Run(testCase.configKey, func(t *testing.T) {
			for label, value := range map[string]int64{"minimum": testCase.minimum, "maximum": testCase.maximum} {
				configuration, finding := ApplyConfigurationOverlay(base, []string{"network_flow_activity", "resource_limits", testCase.configKey}, fmt.Sprint(value))
				if finding != nil {
					t.Fatalf("%s overlay finding = %#v", label, finding)
				}
				// A retained limit at its minimum also requires its active
				// counterpart at the same value for the pair to remain valid.
				switch testCase.configKey {
				case "max_retained_tables_per_incident":
					configuration, finding = ApplyConfigurationOverlay(configuration, []string{"network_flow_activity", "resource_limits", "max_active_tables_per_incident"}, "1")
				case "max_retained_graph_views_per_incident":
					configuration, finding = ApplyConfigurationOverlay(configuration, []string{"network_flow_activity", "resource_limits", "max_active_graph_views_per_incident"}, "1")
				}
				if finding != nil {
					t.Fatalf("%s relationship overlay finding = %#v", label, finding)
				}
				resolved, findings := NormalizeAndValidateConfiguration(configuration)
				if len(findings) != 0 {
					t.Fatalf("%s findings = %#v", label, findings)
				}
				if got := effectiveLimitsResource(resolved.EffectiveResourceLimits())[testCase.discoveryKey]; got != value {
					t.Fatalf("%s effective value = %v; want %d", label, got, value)
				}
			}

			for reason, value := range map[string]int64{"below_minimum": testCase.minimum - 1, "above_maximum": testCase.maximum + 1} {
				configuration, finding := ApplyConfigurationOverlay(base, []string{"network_flow_activity", "resource_limits", testCase.configKey}, fmt.Sprint(value))
				if finding != nil {
					t.Fatalf("invalid overlay parse finding = %#v", finding)
				}
				_, findings := NormalizeAndValidateConfiguration(configuration)
				if !hasConfigurationFinding(findings, testCase.configKey, reason) {
					t.Fatalf("%s findings = %#v", reason, findings)
				}
			}
		})
	}

	for name, raw := range map[string]string{
		"explicit null": "null",
		"non-object":    `"invalid"`,
		"non-integer":   `{"max_graph_vertices":1.5}`,
		"member null":   `{"max_graph_vertices":null}`,
		"unknown":       `{"unknown":1}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, finding := ApplyConfigurationOverlay(base, []string{"network_flow_activity", "resource_limits"}, raw); finding == nil {
				t.Fatalf("invalid resource-limit overlay %q was accepted", raw)
			}
		})
	}

	t.Run("cross-limit relationships", func(t *testing.T) {
		for _, pair := range [][2]string{
			{"max_active_tables_per_incident", "max_retained_tables_per_incident"},
			{"max_active_graph_views_per_incident", "max_retained_graph_views_per_incident"},
		} {
			configuration, finding := ApplyConfigurationOverlay(base, []string{"network_flow_activity", "resource_limits"}, fmt.Sprintf(`{"%s":2,"%s":1}`, pair[0], pair[1]))
			if finding != nil {
				t.Fatalf("cross-limit overlay finding = %#v", finding)
			}
			_, findings := NormalizeAndValidateConfiguration(configuration)
			if len(findings) != 1 || findings[0].ReasonCode != "invalid_limit_relationship" {
				t.Fatalf("cross-limit findings = %#v", findings)
			}
		}
	})

	t.Run("clones cannot mutate retained configuration", func(t *testing.T) {
		configuration, finding := ApplyConfigurationOverlay(base, []string{"network_flow_activity", "resource_limits", "max_graph_vertices"}, "123")
		if finding != nil {
			t.Fatal(finding)
		}
		cloned := CloneConfiguration(configuration)
		*cloned.ResourceLimits.MaxGraphVertices = 456
		if *configuration.ResourceLimits.MaxGraphVertices != 123 {
			t.Fatal("resource-limit override pointer was shared across configuration clones")
		}
	})

	if err := ValidateEffectiveLimits(EffectiveLimits{}); err == nil {
		t.Fatal("zero EffectiveLimits was silently defaulted")
	}
}

func hasConfigurationFinding(findings []ConfigurationFinding, key, reason string) bool {
	path := "network_flow_activity.resource_limits." + key
	for _, finding := range findings {
		if finding.Path == path && finding.ReasonCode == reason {
			return true
		}
	}
	return false
}
