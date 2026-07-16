package viewschema

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
)

func TestBaseRegistryPublicResources(t *testing.T) {
	resources := ListPublicResources()
	gotIDs := make([]string, 0, len(resources))
	for _, resource := range resources {
		gotIDs = append(gotIDs, resource.ViewSchemaID)
	}
	wantIDs := []string{
		"cartulary.view.assessments.v1",
		"cartulary.view.comm_log.v1",
		"cartulary.view.decisions.v1",
		"cartulary.view.evidence.v1",
		"cartulary.view.findings.v1",
		"cartulary.view.forensic_keywords.v1",
		"cartulary.view.handoff.v1",
		"cartulary.view.hosts.v1",
		"cartulary.view.identities.v1",
		"cartulary.view.indicators.v1",
		"cartulary.view.investigative_queries.v1",
		"cartulary.view.lesson.v1",
		"cartulary.view.notes.v1",
		"cartulary.view.parties.v1",
		"cartulary.view.status_review.v1",
		"cartulary.view.task_requests.v1",
		"cartulary.view.timeline.v2",
	}
	if !reflect.DeepEqual(gotIDs, wantIDs) {
		t.Fatalf("unexpected Base view schema ids:\ngot  %v\nwant %v", gotIDs, wantIDs)
	}
	if !slices.IsSorted(gotIDs) {
		t.Fatalf("view schema ids must be sorted ascending: %v", gotIDs)
	}

	for _, resource := range resources {
		requirePublicResourceShape(t, resource)
		requireFieldOrderPreserved(t, resource)
		requireNoInternalMembers(t, resource)
	}
}

func TestSupportPhase9TaskDecisionInternalProjectionBindings(t *testing.T) {
	tests := []struct {
		viewSchemaID   string
		baseProjection string
	}{
		{viewSchemaID: "cartulary.view.decisions.v1", baseProjection: "decision_grid_projection"},
		{viewSchemaID: "cartulary.view.task_requests.v1", baseProjection: "task_request_grid_projection"},
	}
	for _, tc := range tests {
		t.Run(tc.viewSchemaID, func(t *testing.T) {
			schema, ok := Lookup(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing schema %s", tc.viewSchemaID)
			}
			if schema.BaseProjection != tc.baseProjection {
				t.Fatalf("%s base projection: got %q want %q", tc.viewSchemaID, schema.BaseProjection, tc.baseProjection)
			}

			resource, ok := LookupPublicResource(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing public resource %s", tc.viewSchemaID)
			}
			requireNoInternalMembers(t, resource)
		})
	}
}

func TestSupportPhase9CoordinationInternalProjectionAndFilterBindings(t *testing.T) {
	tests := []struct {
		viewSchemaID   string
		artifactType   string
		baseProjection string
	}{
		{viewSchemaID: "cartulary.view.comm_log.v1", artifactType: "comm_log", baseProjection: "artifact_grid_projection"},
		{viewSchemaID: "cartulary.view.handoff.v1", artifactType: "handoff", baseProjection: "artifact_grid_projection"},
		{viewSchemaID: "cartulary.view.lesson.v1", artifactType: "lesson", baseProjection: "artifact_grid_projection"},
		{viewSchemaID: "cartulary.view.status_review.v1", artifactType: "status_review", baseProjection: "artifact_grid_projection"},
	}
	for _, tc := range tests {
		t.Run(tc.viewSchemaID, func(t *testing.T) {
			schema, ok := Lookup(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing schema %s", tc.viewSchemaID)
			}
			if schema.BaseProjection != tc.baseProjection {
				t.Fatalf("%s base projection: got %q want %q", tc.viewSchemaID, schema.BaseProjection, tc.baseProjection)
			}
			filter, ok := schema.CanonicalSourceFilter()
			if !ok {
				t.Fatalf("%s missing canonical source filter", tc.viewSchemaID)
			}
			want := CanonicalSourceFilter{Kind: "artifact_type", Field: "artifact_type", Value: tc.artifactType}
			if filter != want {
				t.Fatalf("%s canonical source filter: got %#v want %#v", tc.viewSchemaID, filter, want)
			}

			resource, ok := LookupPublicResource(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing public resource %s", tc.viewSchemaID)
			}
			requireNoInternalMembers(t, resource)
		})
	}
}

func TestSupportPhase9HandoffAckStateExposesDeclaredEnumOrder(t *testing.T) {
	resource, ok := LookupPublicResource("cartulary.view.handoff.v1")
	if !ok {
		t.Fatal("missing handoff public resource")
	}
	for _, field := range resource.Fields {
		if field.FieldKey != "handoff.ack_state" {
			continue
		}
		if field.ReadKind != "enum" {
			t.Fatalf("handoff.ack_state read_kind: got %q want enum", field.ReadKind)
		}
		want := []string{"pending", "acknowledged"}
		if !reflect.DeepEqual(field.EnumValues, want) {
			t.Fatalf("handoff.ack_state enum values: got %#v want %#v", field.EnumValues, want)
		}
		return
	}
	t.Fatal("handoff.ack_state field not exposed")
}

func TestSupportPhase9OptionalStandardizedSurfacesExposedAsAdditiveResources(t *testing.T) {
	tests := []struct {
		viewSchemaID string
		artifactType string
		wantFields   []string
	}{
		{
			viewSchemaID: "cartulary.view.findings.v1",
			artifactType: "finding",
			wantFields: []string{
				"finding.statement",
				"finding.kind",
				"finding.state",
				"finding.owner_user_id",
				"finding.confidence_score",
				"finding.closed_at",
				"finding.updated_at",
				"finding.supporting_refs",
				"finding.contradictory_refs",
				"finding.confidence_band",
			},
		},
		{
			viewSchemaID: "cartulary.view.investigative_queries.v1",
			artifactType: "investigative_query",
			wantFields: []string{
				"investigative_query.platform",
				"investigative_query.purpose",
				"investigative_query.query_text",
				"investigative_query.created_by_user_id",
				"investigative_query.created_at",
				"investigative_query.query_id",
				"investigative_query.created_day",
			},
		},
		{
			viewSchemaID: "cartulary.view.forensic_keywords.v1",
			artifactType: "forensic_keyword",
			wantFields: []string{
				"forensic_keyword.pattern",
				"forensic_keyword.reason",
				"forensic_keyword.match_mode",
				"forensic_keyword.case_sensitive",
				"forensic_keyword.created_at",
				"forensic_keyword.keyword_id",
				"forensic_keyword.created_day",
			},
		},
	}
	resources := ListPublicResources()
	if len(resources) != 17 {
		t.Fatalf("current build discovery must expose fourteen required plus three optional surfaces, got %d", len(resources))
	}
	if _, ok := LookupPublicResource("cartulary.view.hypotheses.v1"); ok {
		t.Fatal("cartulary.view.hypotheses.v1 must not be exposed")
	}
	for _, resource := range resources {
		if strings.Contains(resource.ViewSchemaID, "hypothes") {
			t.Fatalf("hypothesis surface must not be exposed in discovery: %s", resource.ViewSchemaID)
		}
	}
	for _, tc := range tests {
		t.Run(tc.viewSchemaID, func(t *testing.T) {
			schema, ok := Lookup(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing optional schema %s", tc.viewSchemaID)
			}
			if schema.BaseProjection != "artifact_grid_projection" {
				t.Fatalf("%s base projection: got %q", tc.viewSchemaID, schema.BaseProjection)
			}
			filter, ok := schema.CanonicalSourceFilter()
			if !ok || filter != (CanonicalSourceFilter{Kind: "artifact_type", Field: "artifact_type", Value: tc.artifactType}) {
				t.Fatalf("%s filter: got %#v ok=%v want artifact_type=%s", tc.viewSchemaID, filter, ok, tc.artifactType)
			}
			resource, ok := LookupPublicResource(tc.viewSchemaID)
			if !ok {
				t.Fatalf("missing public resource %s", tc.viewSchemaID)
			}
			gotFields := make([]string, 0, len(resource.Fields))
			for _, field := range resource.Fields {
				gotFields = append(gotFields, field.FieldKey)
			}
			if !reflect.DeepEqual(gotFields, tc.wantFields) {
				t.Fatalf("%s fields:\ngot  %v\nwant %v", tc.viewSchemaID, gotFields, tc.wantFields)
			}
			requirePublicResourceShape(t, resource)
			requireNoInternalMembers(t, resource)
		})
	}
}

func TestPhase9_U_9_13_ArtifactVariantRegistryClosure(t *testing.T) {
	want := []ArtifactVariant{
		{
			ArtifactVariantID:         "note",
			DurableDiscriminatorKind:  "artifact_type",
			DurableDiscriminatorField: "artifact_type",
			DurableDiscriminatorValue: "note",
			PublicSurfaceRef:          "cartulary.view.notes.v1",
			SurfaceStatus:             "required built-in sheet",
			IdentifierField:           "record_id",
			RequiredStructuredState:   []string{"title", "body"},
			OptionalStructuredState:   []string{"tags via record_tags"},
			LifecycleNotes:            "shared artifact lifecycle only; Notes-sheet create and add-linked-note create use the same underlying artifact shape",
		},
		{
			ArtifactVariantID:         "comm_log",
			DurableDiscriminatorKind:  "artifact_type",
			DurableDiscriminatorField: "artifact_type",
			DurableDiscriminatorValue: "comm_log",
			PublicSurfaceRef:          "cartulary.view.comm_log.v1",
			SurfaceStatus:             "required workbook-native coordination surface",
			IdentifierField:           "comm_id",
			RequiredStructuredState:   []string{"comm_id", "comm_type", "timestamp_utc", "audience", "channel_or_meeting", "summary"},
			OptionalStructuredState:   []string{"decision_ids[]", "action_task_ids[]", "next_report_at", "privilege_tag", "audience_party_ids[]", "attendee_party_ids[]"},
			LifecycleNotes:            "audience text remains required source-preserving text even when supplemental party refs are present",
		},
		{
			ArtifactVariantID:         "handoff",
			DurableDiscriminatorKind:  "artifact_type",
			DurableDiscriminatorField: "artifact_type",
			DurableDiscriminatorValue: "handoff",
			PublicSurfaceRef:          "cartulary.view.handoff.v1",
			SurfaceStatus:             "required workbook-native coordination surface",
			IdentifierField:           "handoff_id",
			RequiredStructuredState:   []string{"handoff_id", "timestamp_utc", "outgoing_owner_user_id", "incoming_owner_user_id", "current_state_summary"},
			OptionalStructuredState:   []string{"open_task_ids[]", "open_decision_ids[]", "open_risk_refs[]", "next_checks", "acknowledged_at"},
			LifecycleNotes:            "derived ack_state uses exact tokens pending|acknowledged; risk refs are child rows rather than first-class risk records",
		},
		{
			ArtifactVariantID:         "status_review",
			DurableDiscriminatorKind:  "artifact_type",
			DurableDiscriminatorField: "artifact_type",
			DurableDiscriminatorValue: "status_review",
			PublicSurfaceRef:          "cartulary.view.status_review.v1",
			SurfaceStatus:             "required workbook-native coordination surface",
			IdentifierField:           "status_review_id",
			RequiredStructuredState:   []string{"status_review_id", "timestamp_utc", "review_owner_user_id", "current_state_summary"},
			OptionalStructuredState:   []string{"blocked_task_ids[]", "pending_evidence_ids[]", "open_decision_ids[]", "active_risks_summary", "next_report_at"},
			LifecycleNotes:            "coordination artifact only; no separate subtype lifecycle machine beyond ordinary artifact lifecycle",
		},
		{
			ArtifactVariantID:         "lesson",
			DurableDiscriminatorKind:  "artifact_type",
			DurableDiscriminatorField: "artifact_type",
			DurableDiscriminatorValue: "lesson",
			PublicSurfaceRef:          "cartulary.view.lesson.v1",
			SurfaceStatus:             "required workbook-native coordination surface",
			IdentifierField:           "lesson_id",
			RequiredStructuredState:   []string{"lesson_id", "timestamp_utc", "summary", "owner_user_id"},
			OptionalStructuredState:   []string{"follow_up_task_ids[]", "closure_state", "evidence_refs[]"},
			LifecycleNotes:            "closure_state uses the exact closed vocabulary defined in Core 02 section 18; lessons remain artifact-backed and reuse shared history and links",
		},
		{
			ArtifactVariantID:         "finding",
			DurableDiscriminatorKind:  "artifact_type",
			DurableDiscriminatorField: "artifact_type",
			DurableDiscriminatorValue: "finding",
			SubkindDimension:          "finding.kind",
			PublicSurfaceRef:          "cartulary.view.findings.v1",
			SurfaceStatus:             "standardized optional workbook surface",
			IdentifierField:           "record_id",
			RequiredStructuredState:   []string{"finding.kind", "statement", "state", "confidence_score", "owner_user_id"},
			OptionalStructuredState:   []string{"closed_at", "supporting_refs[]", "contradictory_refs[]"},
			LifecycleNotes:            "this is the only current-profile row that covers both findings and hypotheses; finding.kind is required structured state; finding.state uses the exact open|closed vocabulary defined in Core 02 section 18; closed_at is server-managed",
		},
	}
	got := ListArtifactVariants()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected artifact variant registry:\ngot  %#v\nwant %#v", got, want)
	}

	requiredSurfaceKinds := map[string]string{
		"cartulary.view.notes.v1":         "built_in_sheet",
		"cartulary.view.comm_log.v1":      "system_view",
		"cartulary.view.handoff.v1":       "system_view",
		"cartulary.view.status_review.v1": "system_view",
		"cartulary.view.lesson.v1":        "system_view",
	}
	requiredCoordinationSurfaces := map[string]string{
		"cartulary.view.comm_log.v1":      "comm_log",
		"cartulary.view.handoff.v1":       "handoff",
		"cartulary.view.status_review.v1": "status_review",
		"cartulary.view.lesson.v1":        "lesson",
		"cartulary.view.findings.v1":      "finding",
	}
	for _, variant := range want {
		byID, ok := LookupArtifactVariant(variant.ArtifactVariantID)
		if !ok || !reflect.DeepEqual(byID, variant) {
			t.Fatalf("lookup by variant id %s: got %#v ok=%v want %#v", variant.ArtifactVariantID, byID, ok, variant)
		}
		byArtifactType, ok := LookupArtifactVariantByArtifactType(variant.DurableDiscriminatorValue)
		if !ok || !reflect.DeepEqual(byArtifactType, variant) {
			t.Fatalf("lookup by artifact type %s: got %#v ok=%v want %#v", variant.DurableDiscriminatorValue, byArtifactType, ok, variant)
		}
		if wantKind, required := requiredSurfaceKinds[variant.PublicSurfaceRef]; required {
			resource, ok := LookupPublicResource(variant.PublicSurfaceRef)
			if !ok {
				t.Fatalf("required artifact variant surface %s is not exposed", variant.PublicSurfaceRef)
			}
			if resource.SurfaceKind != wantKind {
				t.Fatalf("%s surface kind: got %q want %q", variant.PublicSurfaceRef, resource.SurfaceKind, wantKind)
			}
		}
		if wantArtifactType, required := requiredCoordinationSurfaces[variant.PublicSurfaceRef]; required {
			schema, ok := Lookup(variant.PublicSurfaceRef)
			if !ok {
				t.Fatalf("missing schema for %s", variant.PublicSurfaceRef)
			}
			filter, ok := schema.CanonicalSourceFilter()
			if !ok || filter != (CanonicalSourceFilter{Kind: "artifact_type", Field: "artifact_type", Value: wantArtifactType}) {
				t.Fatalf("%s filter: got %#v ok=%v want artifact_type=%s", variant.PublicSurfaceRef, filter, ok, wantArtifactType)
			}
		}
	}
	if _, ok := LookupArtifactVariant("hypothesis"); ok {
		t.Fatal("hypothesis must not be a separate artifact variant")
	}
	if _, ok := LookupArtifactVariantByArtifactType("hypothesis"); ok {
		t.Fatal("artifact_type='hypothesis' must not be a supported variant discriminator")
	}
	if _, ok := LookupArtifactVariantByArtifactType("investigative_query"); ok {
		t.Fatal("investigative_query must remain outside the closed artifact variant registry")
	}
	if _, ok := LookupArtifactVariantByArtifactType("forensic_keyword"); ok {
		t.Fatal("forensic_keyword must remain outside the closed artifact variant registry")
	}
	if _, ok := LookupPublicResource("cartulary.view.hypotheses.v1"); ok {
		t.Fatal("cartulary.view.hypotheses.v1 must not be exposed")
	}
}

func requirePublicResourceShape(t testing.TB, resource ViewSchemaResource) {
	t.Helper()

	if resource.ViewSchemaID == "" || resource.SurfaceKind == "" || resource.Title == "" {
		t.Fatalf("resource has missing identity members: %#v", resource)
	}
	if !reflect.DeepEqual(resource.TechnicalFields, []string{"record_id", "row_version"}) {
		t.Fatalf("%s has unexpected technical fields: %v", resource.ViewSchemaID, resource.TechnicalFields)
	}
	if len(resource.DefaultSort) == 0 || resource.DefaultSort[len(resource.DefaultSort)-1] != (SortEntry{FieldKey: "record_id", Direction: "asc"}) {
		t.Fatalf("%s default_sort must end with record_id asc: %#v", resource.ViewSchemaID, resource.DefaultSort)
	}
	if resource.SortNullOrder != "last" {
		t.Fatalf("%s has unexpected sort_null_order: %q", resource.ViewSchemaID, resource.SortNullOrder)
	}
	if len(resource.Fields) == 0 {
		t.Fatalf("%s must expose fields", resource.ViewSchemaID)
	}
	requireInspectorConfigShape(t, resource)

	fieldKeys := make(map[string]struct{}, len(resource.Fields))
	for _, field := range resource.Fields {
		if field.FieldKey == "" || field.Label == "" || field.ReadKind == "" || field.WriteKind == "" || field.FilterOps == nil {
			t.Fatalf("%s exposes incomplete field entry: %#v", resource.ViewSchemaID, field)
		}
		if _, exists := fieldKeys[field.FieldKey]; exists {
			t.Fatalf("%s exposes duplicate field key %s", resource.ViewSchemaID, field.FieldKey)
		}
		fieldKeys[field.FieldKey] = struct{}{}
		if field.GridEditable {
			internalField, ok := LookupField(resource.ViewSchemaID, field.FieldKey)
			if !ok || internalField.WriteKind != "direct_value" || !internalField.Writable {
				t.Fatalf("%s exposes invalid grid-editable field %s", resource.ViewSchemaID, field.FieldKey)
			}
		}
		if field.Sortable {
			sortFieldKey := field.FieldKey
			if field.HeaderSortFieldKey != nil {
				sortFieldKey = *field.HeaderSortFieldKey
			}
			if !slices.Contains(resource.SortFields, sortFieldKey) {
				t.Fatalf("%s sortable field %s is not backed by sort_fields key %s", resource.ViewSchemaID, field.FieldKey, sortFieldKey)
			}
		}
	}
	for _, fieldKey := range resource.FilterFields {
		if _, ok := fieldKeys[fieldKey]; !ok {
			t.Fatalf("%s filter field %s is not in fields[]", resource.ViewSchemaID, fieldKey)
		}
	}
	for _, predicate := range resource.SyntheticFilterPredicates {
		if _, ok := fieldKeys[predicate.FieldKey]; ok {
			t.Fatalf("%s synthetic predicate %s must not also be a field", resource.ViewSchemaID, predicate.FieldKey)
		}
	}
}

func requireInspectorConfigShape(t testing.TB, resource ViewSchemaResource) {
	t.Helper()

	config := resource.InspectorConfig
	if config.InspectorConfigSchemaID != "cartulary.inspector_config.v1" {
		t.Fatalf("%s inspector_config schema id: got %q", resource.ViewSchemaID, config.InspectorConfigSchemaID)
	}
	if config.ViewSchemaID != resource.ViewSchemaID {
		t.Fatalf("%s inspector_config view_schema_id mismatch: %#v", resource.ViewSchemaID, config)
	}
	if config.DefaultOpen {
		t.Fatalf("%s inspector_config.default_open must be false", resource.ViewSchemaID)
	}
	if config.SubjectBinding.Kind != "selected_record" {
		t.Fatalf("%s inspector_config subject: %#v", resource.ViewSchemaID, config.SubjectBinding)
	}
	if config.NoRowState != "no_row_selected" {
		t.Fatalf("%s inspector_config no_row_state: got %q", resource.ViewSchemaID, config.NoRowState)
	}
	if config.UnsupportedFeatureBehavior != "omit_feature" {
		t.Fatalf("%s inspector_config unsupported behavior: got %q", resource.ViewSchemaID, config.UnsupportedFeatureBehavior)
	}
	allowedPanels := map[string]struct{}{
		"details":       {},
		"relationships": {},
		"evidence":      {},
		"history":       {},
		"workflow":      {},
	}
	if len(config.Panels) == 0 || len(config.Panels) > 5 {
		t.Fatalf("%s inspector panels bound: got %d", resource.ViewSchemaID, len(config.Panels))
	}
	declaredPanels := map[string]struct{}{}
	for _, panel := range config.Panels {
		if panel.PanelID == "" || panel.Label == "" {
			t.Fatalf("%s inspector panel incomplete: %#v", resource.ViewSchemaID, panel)
		}
		if _, ok := allowedPanels[panel.PanelID]; !ok {
			t.Fatalf("%s inspector panel id not closed vocabulary: %#v", resource.ViewSchemaID, panel)
		}
		if _, exists := declaredPanels[panel.PanelID]; exists {
			t.Fatalf("%s duplicate inspector panel_id %s", resource.ViewSchemaID, panel.PanelID)
		}
		declaredPanels[panel.PanelID] = struct{}{}
	}
	if len(config.FeatureGroups) > 64 {
		t.Fatalf("%s inspector feature group bound: got %d", resource.ViewSchemaID, len(config.FeatureGroups))
	}
	allowedRoutes := map[string]struct{}{
		"panel_read":            {},
		"view_row_create":       {},
		"record_patch":          {},
		"record_action":         {},
		"entity_mention_action": {},
		"evidence_access":       {},
		"surface_pivot":         {},
	}
	allowedOwners := map[string]struct{}{
		"current_row_projection":         {},
		"view_query_route":               {},
		"view_row_create_route":          {},
		"record_patch_route":             {},
		"record_mark_reviewed_route":     {},
		"record_supersede_route":         {},
		"record_delete_route":            {},
		"record_restore_route":           {},
		"record_history_route":           {},
		"record_rollback_route":          {},
		"record_merge_route":             {},
		"entity_mention_resolve_route":   {},
		"evidence_attach_blob_route":     {},
		"evidence_preview_handle_route":  {},
		"evidence_download_handle_route": {},
	}
	allowedConditions := map[string]struct{}{
		"no_row_selected":              {},
		"incident_closed":              {},
		"authorization_lost":           {},
		"row_version_changed":          {},
		"record_deleted":               {},
		"record_merged":                {},
		"evidence_preview_unavailable": {},
		"merge_target_unavailable":     {},
		"record_not_deleted":           {},
		"rollback_target_unavailable":  {},
		"party_text_unavailable":       {},
		"pivot_target_unavailable":     {},
	}
	allowedSuccessBehaviors := map[string]struct{}{
		"preserve_selected_row":    {},
		"retarget_selected_row":    {},
		"clear_to_no_row_selected": {},
		"surface_pivot":            {},
	}
	allowedFailureBehaviors := map[string]struct{}{
		"show_same_shell_error_preserve_selection":        {},
		"show_same_shell_error_invalidate_pending_action": {},
		"show_same_shell_error_clear_subject":             {},
	}
	allowedRoles := map[string]struct{}{"viewer": {}, "editor": {}, "reviewer": {}, "admin": {}}
	featureKeys := map[string]struct{}{}
	gotFeatureKeys := make([]string, 0, len(config.FeatureGroups))
	for _, group := range config.FeatureGroups {
		if group.FeatureGroupKey == "" || group.PanelID == "" || group.Label == "" {
			t.Fatalf("%s inspector feature group incomplete: %#v", resource.ViewSchemaID, group)
		}
		if _, exists := featureKeys[group.FeatureGroupKey]; exists {
			t.Fatalf("%s duplicate inspector feature_group_key %s", resource.ViewSchemaID, group.FeatureGroupKey)
		}
		featureKeys[group.FeatureGroupKey] = struct{}{}
		gotFeatureKeys = append(gotFeatureKeys, group.FeatureGroupKey)
		if _, ok := declaredPanels[group.PanelID]; !ok {
			t.Fatalf("%s inspector feature group references unknown panel %s", resource.ViewSchemaID, group.PanelID)
		}
		if group.MinimumIncidentRole != nil {
			if _, ok := allowedRoles[*group.MinimumIncidentRole]; !ok {
				t.Fatalf("%s inspector feature group has unknown minimum role: %#v", resource.ViewSchemaID, group)
			}
		}
		if _, ok := allowedRoutes[group.RouteBinding.Kind]; !ok {
			t.Fatalf("%s inspector feature group has unknown route kind: %#v", resource.ViewSchemaID, group.RouteBinding)
		}
		if _, ok := allowedOwners[group.RouteBinding.Owner]; !ok {
			t.Fatalf("%s inspector feature group has unknown route owner: %#v", resource.ViewSchemaID, group.RouteBinding)
		}
		if len(group.SeedBindings) > 16 {
			t.Fatalf("%s inspector seed binding bound: %#v", resource.ViewSchemaID, group)
		}
		for _, binding := range group.SeedBindings {
			if binding.TargetFieldKey == "" {
				t.Fatalf("%s inspector seed binding missing target field: %#v", resource.ViewSchemaID, binding)
			}
			switch binding.Source.Kind {
			case "selected_record_id", "selected_field_value", "literal":
			default:
				t.Fatalf("%s inspector seed binding source kind: %#v", resource.ViewSchemaID, binding.Source)
			}
		}
		if len(group.DisabledWhen) > 16 {
			t.Fatalf("%s inspector disabled_when bound: %#v", resource.ViewSchemaID, group)
		}
		for _, condition := range group.DisabledWhen {
			if _, ok := allowedConditions[condition]; !ok {
				t.Fatalf("%s inspector disabled_when condition: %#v", resource.ViewSchemaID, group.DisabledWhen)
			}
		}
		if _, ok := allowedSuccessBehaviors[group.SuccessResultBehavior]; !ok {
			t.Fatalf("%s inspector success_result_behavior: %#v", resource.ViewSchemaID, group)
		}
		if _, ok := allowedFailureBehaviors[group.FailureResultBehavior]; !ok {
			t.Fatalf("%s inspector failure_result_behavior: %#v", resource.ViewSchemaID, group)
		}
	}
	if expected, ok := expectedInspectorFeatureRegistry()[resource.ViewSchemaID]; ok {
		if !reflect.DeepEqual(gotFeatureKeys, expected) {
			t.Fatalf("%s inspector feature registry:\ngot  %#v\nwant %#v", resource.ViewSchemaID, gotFeatureKeys, expected)
		}
	}
}

func expectedInspectorFeatureRegistry() map[string][]string {
	return map[string][]string{
		"cartulary.view.assessments.v1":           {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "assessment.subject_pivot", "assessment.prior_history", "assessment.support_refs.manage", "evidence.refs.manage", "create_related.task_request", "create_related.decision"},
		"cartulary.view.comm_log.v1":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "comm.decisions.link", "comm.action_tasks.link", "comm.parties.manage", "comm.next_report.manage", "create_related.task_request", "create_related.status_review"},
		"cartulary.view.decisions.v1":             {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "decision.support_refs.manage", "decision.affected_records.manage", "decision.status.transition", "decision.supersede", "create_related.task_request", "create_related.comm_log", "create_related.status_review"},
		"cartulary.view.evidence.v1":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "evidence.preview_handle", "evidence.download_handle", "evidence.attach_blob", "party.collector.link", "party.source.link", "party.reference.clear", "relationships.manage", "surface_pivot.linked_records", "surface_pivot.timeline", "create_related.note", "create_related.task_request", "create_related.decision"},
		"cartulary.view.findings.v1":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "finding.support_refs.manage", "finding.contradictory_refs.manage", "finding.evidence_refs.manage", "finding.owner.manage", "finding.close_or_reopen", "create_related.task_request", "create_related.decision"},
		"cartulary.view.forensic_keywords.v1":     {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "keyword.evidence_refs.manage", "keyword.timeline_rows.link", "keyword.findings.link", "create_related.task_request"},
		"cartulary.view.handoff.v1":               {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "handoff.acknowledge", "handoff.open_tasks.review", "handoff.open_decisions.review", "handoff.risks.review", "handoff.next_checks.manage", "create_related.task_request", "create_related.status_review"},
		"cartulary.view.hosts.v1":                 {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "entity.aliases.read", "entity.relationships.manage", "entity.merge", "surface_pivot.timeline", "surface_pivot.evidence", "surface_pivot.assessments", "create_related.note", "create_related.task_request", "create_related.decision"},
		"cartulary.view.identities.v1":            {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "entity.aliases.read", "entity.relationships.manage", "entity.merge", "surface_pivot.timeline", "surface_pivot.evidence", "surface_pivot.assessments", "create_related.note", "create_related.task_request", "create_related.decision"},
		"cartulary.view.indicators.v1":            {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "indicator.observations.pivot", "indicator.lifecycle.read", "relationships.manage", "create_related.task_request", "create_related.decision"},
		"cartulary.view.investigative_queries.v1": {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "query.source.link", "query.result.link", "query.evidence_refs.manage", "query.findings.link", "create_related.task_request"},
		"cartulary.view.lesson.v1":                {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "lesson.followup_tasks.manage", "lesson.evidence_refs.manage", "lesson.owner.manage", "lesson.close_or_reopen", "create_related.task_request"},
		"cartulary.view.notes.v1":                 {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "artifact.source_links.manage", "artifact.evidence_refs.manage", "artifact.tags.manage", "artifact.related_notes.manage", "surface_pivot.source_records", "create_related.task_request", "create_related.decision"},
		"cartulary.view.parties.v1":               {"details.read", "relationships.read", "history.read", "record.delete", "record.restore", "history.rollback", "party.usage_pivot.requester", "party.usage_pivot.collector_source", "party.usage_pivot.audience_attendee", "party.usage_pivot.owner_stakeholder", "party.reference.link", "party.reference.clear", "party.reference.clear_both"},
		"cartulary.view.status_review.v1":         {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "status_review.blocked_tasks.review", "status_review.pending_evidence.review", "status_review.open_decisions.review", "status_review.risks.review", "status_review.next_report.manage", "create_related.task_request", "create_related.comm_log"},
		"cartulary.view.task_requests.v1":         {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "task.links.manage", "task.requester_party.link", "task.requester_party.clear", "task.decision.link", "task.decision.clear", "task.status.transition", "create_related.comm_log", "create_related.status_review", "create_related.lesson"},
		"cartulary.view.timeline.v2":              {"details.read", "relationships.read", "evidence.read", "history.read", "record.delete", "record.restore", "history.rollback", "entity_mentions.resolve", "entity_mentions.create_host", "entity_mentions.create_identity", "entity_mentions.dismiss", "entity_mentions.restore", "indicator.observations.manage", "relationships.manage", "evidence.attach_blob", "evidence.preview_handle", "evidence.download_handle", "timeline.mark_reviewed", "timeline.supersede", "create_related.note", "create_related.task_request", "create_related.decision", "create_related.evidence", "create_related.comm_log", "create_related.handoff", "create_related.status_review", "create_related.lesson"},
	}
}

func requireFieldOrderPreserved(t testing.TB, resource ViewSchemaResource) {
	t.Helper()

	var artifactJSON string
	for _, path := range contracttest.ViewSchemaArtifactPaths(t) {
		if strings.HasSuffix(path, "/index.json") {
			continue
		}
		var document schemaDocument
		candidateJSON := contracttest.ContractArtifactJSON(t, path)
		if err := json.Unmarshal([]byte(candidateJSON), &document); err != nil {
			t.Fatalf("unmarshal %s: %v", path, err)
		}
		if document.ViewSchemaID == resource.ViewSchemaID {
			artifactJSON = candidateJSON
			break
		}
	}
	if artifactJSON == "" {
		t.Fatalf("missing generated artifact for %s", resource.ViewSchemaID)
	}

	var document schemaDocument
	if err := json.Unmarshal([]byte(artifactJSON), &document); err != nil {
		t.Fatalf("unmarshal %s: %v", resource.ViewSchemaID, err)
	}
	want := make([]string, 0, len(document.Fields))
	for _, field := range document.Fields {
		want = append(want, field.FieldKey)
	}
	got := make([]string, 0, len(resource.Fields))
	for _, field := range resource.Fields {
		got = append(got, field.FieldKey)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s field order changed:\ngot  %v\nwant %v", resource.ViewSchemaID, got, want)
	}
}

func requireNoInternalMembers(t testing.TB, resource ViewSchemaResource) {
	t.Helper()

	payload, err := json.Marshal(resource)
	if err != nil {
		t.Fatalf("marshal resource: %v", err)
	}
	for _, forbidden := range []string{"write_target", "write_action", "base_projection", "canonical_source_filter", "read_model", "create_writable", "writable"} {
		if strings.Contains(string(payload), `"`+forbidden+`"`) {
			t.Fatalf("%s public resource leaked %s: %s", resource.ViewSchemaID, forbidden, payload)
		}
	}
}
