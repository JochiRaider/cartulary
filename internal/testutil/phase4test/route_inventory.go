package phase4test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
)

type RouteHarnessClass string

const (
	RouteHarnessSurfaceEnvelope  RouteHarnessClass = "surface_envelope"
	RouteHarnessCSRF             RouteHarnessClass = "csrf"
	RouteHarnessReplayDivergent  RouteHarnessClass = "replay_divergent_conflict"
	RouteHarnessAuthorization    RouteHarnessClass = "authorization_rederivation"
	RouteHarnessQueryFieldMatrix RouteHarnessClass = "default_query_meta_field_keys"
	RouteHarnessEffects          RouteHarnessClass = "projection_websocket_consequences"
)

var phase4RouteHarnessClasses = []RouteHarnessClass{
	RouteHarnessSurfaceEnvelope,
	RouteHarnessCSRF,
	RouteHarnessReplayDivergent,
	RouteHarnessAuthorization,
	RouteHarnessQueryFieldMatrix,
	RouteHarnessEffects,
}

type RouteHarnessRequirement string

const (
	RouteHarnessRequired      RouteHarnessRequirement = "required"
	RouteHarnessNotApplicable RouteHarnessRequirement = "n/a"
)

type RouteKey string

const (
	RouteMentionResolve   RouteKey = "mention_resolve"
	RouteExplicitMerge    RouteKey = "explicit_merge"
	RouteHostsCreate      RouteKey = "hosts_create"
	RouteHostsQuery       RouteKey = "hosts_query"
	RouteIdentitiesCreate RouteKey = "identities_create"
	RouteIdentitiesQuery  RouteKey = "identities_query"
	RouteIndicatorsCreate RouteKey = "indicators_create"
	RouteIndicatorsQuery  RouteKey = "indicators_query"
	RouteTimelineCreate   RouteKey = "timeline_create"
	RouteTimelineQuery    RouteKey = "timeline_query"
	RouteTimelinePatch    RouteKey = "timeline_patch"
)

type RouteSuccessShape string

const (
	RouteSuccessShapeMentionResolution RouteSuccessShape = "mention_resolution"
	RouteSuccessShapeMerge             RouteSuccessShape = "merge_summary"
	RouteSuccessShapeMutationRow       RouteSuccessShape = "mutation_row"
	RouteSuccessShapeQueryRows         RouteSuccessShape = "query_rows"
)

type RouteReplayCapability string

const (
	RouteReplayNotApplicable      RouteReplayCapability = "n/a"
	RouteReplayStoredPayloadReuse RouteReplayCapability = "stored_payload_reuse"
)

type RouteAuthorizationChange string

const (
	RouteAuthorizationNotApplicable RouteAuthorizationChange = "n/a"
	RouteAuthorizationDemoteViewer  RouteAuthorizationChange = "demote_to_viewer"
	RouteAuthorizationRemoveMember  RouteAuthorizationChange = "remove_membership"
)

type RouteProjectionTarget string

const (
	RouteProjectionNotApplicable RouteProjectionTarget = "n/a"
	RouteProjectionHosts         RouteProjectionTarget = "hosts"
	RouteProjectionIdentities    RouteProjectionTarget = "identities"
	RouteProjectionIndicators    RouteProjectionTarget = "indicators"
	RouteProjectionTimeline      RouteProjectionTarget = "timeline"
)

type RouteWebSocketExpectation string

const (
	RouteWebSocketNotApplicable RouteWebSocketExpectation = "n/a"
	RouteWebSocketRecordChanged RouteWebSocketExpectation = "record_changed"
)

type RouteInventoryContext struct {
	IncidentID            string
	TimelineRecordID      string
	MentionID             string
	MergeSurvivorRecordID string
	MergeLoserRecordID    string
	HostRecordID          string
	IdentityRecordID      string
	IndicatorRecordID     string
}

type RouteInventoryEntry struct {
	Key                  RouteKey
	Name                 string
	Method               string
	BuildPath            func(RouteInventoryContext) string
	BuildBody            func(RouteInventoryContext, string) any
	BuildDivergentBody   func(RouteInventoryContext, string) any
	AffectedRecordID     func(RouteInventoryContext, map[string]any) string
	SuccessStatus        int
	SuccessShape         RouteSuccessShape
	RequiresCSRF         bool
	ExpectedViewSchemaID string

	ReplayCapability RouteReplayCapability
	ReplayStatus     int
	DivergentStatus  int
	DivergentCode    string

	AuthorizationChange RouteAuthorizationChange
	AuthorizationStatus int
	AuthorizationCode   string

	ProjectionTarget       RouteProjectionTarget
	WebSocketExpectation   RouteWebSocketExpectation
	WebSocketViewSchemaID  string
	BuildWebSocketRecordID func(RouteInventoryContext) string
	WebSocketRowVersion    int64

	HarnessRequirements map[RouteHarnessClass]RouteHarnessRequirement
}

func Phase4RouteHarnessClasses() []RouteHarnessClass {
	return append([]RouteHarnessClass(nil), phase4RouteHarnessClasses...)
}

func ValidateRouteInventory(t testing.TB, routes []RouteInventoryEntry) {
	t.Helper()

	for _, route := range routes {
		if route.Method == "" {
			t.Fatalf("phase4 route %s missing method", route.Key)
		}
		if route.BuildPath == nil {
			t.Fatalf("phase4 route %s missing path builder", route.Key)
		}
		if route.BuildBody == nil {
			t.Fatalf("phase4 route %s missing body builder", route.Key)
		}
		if route.SuccessStatus == 0 {
			t.Fatalf("phase4 route %s missing success status", route.Key)
		}
		if route.SuccessShape == "" {
			t.Fatalf("phase4 route %s missing success shape", route.Key)
		}
		if route.AffectedRecordID == nil {
			t.Fatalf("phase4 route %s missing affected record selector", route.Key)
		}
		if route.HarnessRequirements == nil {
			t.Fatalf("phase4 route %s missing harness requirements", route.Key)
		}
		for _, harness := range phase4RouteHarnessClasses {
			requirement, ok := route.HarnessRequirements[harness]
			if !ok {
				t.Fatalf("phase4 route %s missing harness requirement for %s", route.Key, harness)
			}
			if requirement != RouteHarnessRequired && requirement != RouteHarnessNotApplicable {
				t.Fatalf("phase4 route %s has invalid harness requirement %q for %s", route.Key, requirement, harness)
			}
		}
		if route.RequiresCSRF && route.HarnessRequirements[RouteHarnessCSRF] != RouteHarnessRequired {
			t.Fatalf("phase4 route %s requires csrf but matrix marks %s", route.Key, route.HarnessRequirements[RouteHarnessCSRF])
		}
		if !route.RequiresCSRF && route.HarnessRequirements[RouteHarnessCSRF] != RouteHarnessNotApplicable {
			t.Fatalf("phase4 route %s does not require csrf but matrix marks %s", route.Key, route.HarnessRequirements[RouteHarnessCSRF])
		}
		switch route.ReplayCapability {
		case RouteReplayNotApplicable:
			if route.HarnessRequirements[RouteHarnessReplayDivergent] != RouteHarnessNotApplicable {
				t.Fatalf("phase4 route %s marks replay n/a but matrix requires replay", route.Key)
			}
		case RouteReplayStoredPayloadReuse:
			if route.HarnessRequirements[RouteHarnessReplayDivergent] != RouteHarnessRequired {
				t.Fatalf("phase4 route %s marks replay reusable but matrix does not require replay", route.Key)
			}
			if route.BuildDivergentBody == nil || route.ReplayStatus == 0 || route.DivergentStatus == 0 || route.DivergentCode == "" {
				t.Fatalf("phase4 route %s missing replay metadata", route.Key)
			}
		default:
			t.Fatalf("phase4 route %s has invalid replay capability %q", route.Key, route.ReplayCapability)
		}
		switch route.AuthorizationChange {
		case RouteAuthorizationNotApplicable:
			if route.HarnessRequirements[RouteHarnessAuthorization] != RouteHarnessNotApplicable {
				t.Fatalf("phase4 route %s marks authorization n/a but matrix requires authorization", route.Key)
			}
		case RouteAuthorizationDemoteViewer, RouteAuthorizationRemoveMember:
			if route.HarnessRequirements[RouteHarnessAuthorization] != RouteHarnessRequired {
				t.Fatalf("phase4 route %s must require authorization re-derivation", route.Key)
			}
			if route.AuthorizationStatus == 0 || route.AuthorizationCode == "" {
				t.Fatalf("phase4 route %s missing authorization expectation", route.Key)
			}
		default:
			t.Fatalf("phase4 route %s has invalid authorization change %q", route.Key, route.AuthorizationChange)
		}
		if route.ExpectedViewSchemaID == "" {
			t.Fatalf("phase4 route %s missing expected view schema", route.Key)
		}
		if route.HarnessRequirements[RouteHarnessQueryFieldMatrix] != RouteHarnessRequired {
			t.Fatalf("phase4 route %s must require query field matrix coverage", route.Key)
		}
		if route.ProjectionTarget == RouteProjectionNotApplicable && route.WebSocketExpectation == RouteWebSocketNotApplicable {
			if route.HarnessRequirements[RouteHarnessEffects] != RouteHarnessNotApplicable {
				t.Fatalf("phase4 route %s has no projection or websocket expectation but matrix requires effects", route.Key)
			}
		} else {
			if route.HarnessRequirements[RouteHarnessEffects] != RouteHarnessRequired {
				t.Fatalf("phase4 route %s must require effects coverage", route.Key)
			}
		}
		if route.WebSocketExpectation == RouteWebSocketRecordChanged {
			if route.WebSocketViewSchemaID == "" || route.BuildWebSocketRecordID == nil || route.WebSocketRowVersion == 0 {
				t.Fatalf("phase4 route %s missing websocket expectation metadata", route.Key)
			}
		}
	}
}

func RoutesForHarness(t testing.TB, routes []RouteInventoryEntry, harness RouteHarnessClass) []RouteInventoryEntry {
	t.Helper()

	ValidateRouteInventory(t, routes)

	filtered := make([]RouteInventoryEntry, 0, len(routes))
	for _, route := range routes {
		if route.HarnessRequirements[harness] == RouteHarnessRequired {
			filtered = append(filtered, route)
		}
	}
	return filtered
}

func Phase4RouteInventory(ctx RouteInventoryContext) []RouteInventoryEntry {
	return []RouteInventoryEntry{
		{
			Key:    RouteMentionResolve,
			Name:   "mention resolve route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/entity-mentions/" + fixture.MentionID + "/resolve"
			},
			BuildBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return fixtures.MentionResolveRoutePayload(1, clientTxnID, golden.Phase4MentionActionResolve, uuidPointer(fixture.HostRecordID), nil)
			},
			BuildDivergentBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return fixtures.MentionResolveRoutePayload(1, clientTxnID, golden.Phase4MentionActionDismiss, nil, nil)
			},
			AffectedRecordID:      func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.TimelineRecordID },
			SuccessStatus:         http.StatusOK,
			SuccessShape:          RouteSuccessShapeMentionResolution,
			RequiresCSRF:          true,
			ExpectedViewSchemaID:  golden.Phase4TimelineViewSchemaID,
			ReplayCapability:      RouteReplayStoredPayloadReuse,
			ReplayStatus:          http.StatusOK,
			DivergentStatus:       http.StatusConflict,
			DivergentCode:         "client_txn_conflict",
			AuthorizationChange:   RouteAuthorizationDemoteViewer,
			AuthorizationStatus:   http.StatusForbidden,
			AuthorizationCode:     "authorization_denied",
			ProjectionTarget:      RouteProjectionTimeline,
			WebSocketExpectation:  RouteWebSocketRecordChanged,
			WebSocketViewSchemaID: golden.Phase4TimelineViewSchemaID,
			BuildWebSocketRecordID: func(fixture RouteInventoryContext) string {
				return fixture.TimelineRecordID
			},
			WebSocketRowVersion: 2,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteExplicitMerge,
			Name:   "explicit merge route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/records/" + fixture.MergeSurvivorRecordID + "/merge"
			},
			BuildBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return explicitMergePayload(fixture, clientTxnID, "support merge duplicate host")
			},
			BuildDivergentBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return explicitMergePayload(fixture, clientTxnID, "support merge divergent")
			},
			AffectedRecordID:      func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.TimelineRecordID },
			SuccessStatus:         http.StatusOK,
			SuccessShape:          RouteSuccessShapeMerge,
			RequiresCSRF:          true,
			ExpectedViewSchemaID:  golden.Phase4TimelineViewSchemaID,
			ReplayCapability:      RouteReplayStoredPayloadReuse,
			ReplayStatus:          http.StatusOK,
			DivergentStatus:       http.StatusConflict,
			DivergentCode:         "client_txn_conflict",
			AuthorizationChange:   RouteAuthorizationDemoteViewer,
			AuthorizationStatus:   http.StatusForbidden,
			AuthorizationCode:     "authorization_denied",
			ProjectionTarget:      RouteProjectionTimeline,
			WebSocketExpectation:  RouteWebSocketRecordChanged,
			WebSocketViewSchemaID: golden.Phase4TimelineViewSchemaID,
			BuildWebSocketRecordID: func(fixture RouteInventoryContext) string {
				return fixture.TimelineRecordID
			},
			WebSocketRowVersion: 1,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteHostsCreate,
			Name:   "hosts create route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4HostsViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any { return fixtures.HostCreatePayload(clientTxnID) },
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "host.display_name": "VPN Gateway Divergent", "host.hostname": "VPN-GATEWAY-DIVERGENT"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: golden.Phase4HostsViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionHosts,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteHostsQuery,
			Name:   "hosts query route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4HostsViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.HostRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: golden.Phase4HostsViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionHosts,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteIdentitiesCreate,
			Name:   "identities create route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4IdentitiesViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return fixtures.IdentityCreatePayload(clientTxnID)
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "identity.display_name": "VPN User Divergent", "identity.email": "vpn.user.divergent@example.test", "identity.sam_account_name": "VPNDIV"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: golden.Phase4IdentitiesViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionIdentities,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteIdentitiesQuery,
			Name:   "identities query route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4IdentitiesViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.IdentityRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: golden.Phase4IdentitiesViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionIdentities,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteIndicatorsCreate,
			Name:   "indicators create route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4IndicatorsViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return fixtures.IndicatorCreatePayload(clientTxnID)
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "indicator.indicator_type": golden.Phase4IndicatorExamples[0].IndicatorType, "indicator.value_kind": golden.Phase4IndicatorExamples[0].ValueKind, "indicator.display_value": "203.0.113.25"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: golden.Phase4IndicatorsViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionIndicators,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteIndicatorsQuery,
			Name:   "indicators query route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4IndicatorsViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.IndicatorRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: golden.Phase4IndicatorsViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionIndicators,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteTimelineCreate,
			Name:   "timeline create route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4TimelineViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "timeline.summary": "support timeline create"}
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "timeline.summary": "support timeline create divergent"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: golden.Phase4TimelineViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionTimeline,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteTimelineQuery,
			Name:   "timeline query route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + golden.Phase4TimelineViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.TimelineRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: golden.Phase4TimelineViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionTimeline,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:       RouteTimelinePatch,
			Name:      "timeline patch route",
			Method:    http.MethodPatch,
			BuildPath: func(fixture RouteInventoryContext) string { return "/api/v1/records/" + fixture.TimelineRecordID },
			BuildBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return fixtures.TimelineCollectionPatchPayload(golden.Phase4FieldTimelineHostRefs, 1, clientTxnID, fixtures.CollectionActions(fixtures.AddResolvedRefAction("WS-023", mustUUID(fixture.HostRecordID))))
			},
			BuildDivergentBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return fixtures.TimelineCollectionPatchPayload(golden.Phase4FieldTimelineHostRefs, 1, clientTxnID, fixtures.CollectionActions(fixtures.AddResolvedRefAction("WS-024", mustUUID(fixture.MergeLoserRecordID))))
			},
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.TimelineRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: golden.Phase4TimelineViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionTimeline,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredPhase4Harnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
	}
}

func requiredPhase4Harnesses(required ...RouteHarnessClass) map[RouteHarnessClass]RouteHarnessRequirement {
	requirements := make(map[RouteHarnessClass]RouteHarnessRequirement, len(phase4RouteHarnessClasses))
	for _, harness := range phase4RouteHarnessClasses {
		requirements[harness] = RouteHarnessNotApplicable
	}
	for _, harness := range required {
		requirements[harness] = RouteHarnessRequired
	}
	return requirements
}

func explicitMergePayload(fixture RouteInventoryContext, clientTxnID string, reason string) map[string]any {
	return map[string]any{
		"loser_record_id":           fixture.MergeLoserRecordID,
		"survivor_base_row_version": 1,
		"loser_base_row_version":    1,
		"client_txn_id":             clientTxnID,
		"reason":                    reason,
	}
}

func rowRecordID(data map[string]any) string {
	row, ok := data["row"].(map[string]any)
	if !ok {
		return ""
	}
	recordID, _ := row["record_id"].(string)
	return recordID
}

func mustUUID(value string) uuid.UUID {
	return uuid.MustParse(value)
}

func uuidPointer(value string) *uuid.UUID {
	parsed := mustUUID(value)
	return &parsed
}
