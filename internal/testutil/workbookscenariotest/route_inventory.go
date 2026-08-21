package workbookscenariotest

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	viewtest "github.com/JochiRaider/cartulary/internal/platform/viewschema/testsupport"
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

var workbookRouteHarnessClasses = []RouteHarnessClass{
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
	RouteMentionResolve         RouteKey = "mention_resolve"
	RouteExplicitMerge          RouteKey = "explicit_merge"
	RouteHostsCreate            RouteKey = "hosts_create"
	RouteHostsQuery             RouteKey = "hosts_query"
	RouteIdentitiesCreate       RouteKey = "identities_create"
	RouteIdentitiesQuery        RouteKey = "identities_query"
	RouteIndicatorsCreate       RouteKey = "indicators_create"
	RouteIndicatorsQuery        RouteKey = "indicators_query"
	RouteTimelineCreate         RouteKey = "timeline_create"
	RouteTimelineQuery          RouteKey = "timeline_query"
	RouteTimelinePatch          RouteKey = "timeline_patch"
	RouteObjectBlobCreate       RouteKey = "object_blob_create"
	RouteEvidenceAttachBlob     RouteKey = "evidence_attach_blob"
	RouteEvidencePreviewHandle  RouteKey = "evidence_preview_handle"
	RouteEvidenceDownloadHandle RouteKey = "evidence_download_handle"
	RoutePartiesCreate          RouteKey = "parties_create"
	RoutePartiesQuery           RouteKey = "parties_query"
	RoutePartiesPatch           RouteKey = "parties_patch"
	RouteAssessmentsCreate      RouteKey = "assessments_create"
	RouteAssessmentsQuery       RouteKey = "assessments_query"
	RouteEvidenceCreate         RouteKey = "evidence_create"
	RouteEvidenceQuery          RouteKey = "evidence_query"
	RouteNotesCreate            RouteKey = "notes_create"
	RouteNotesQuery             RouteKey = "notes_query"
	RouteTaskRequestsCreate     RouteKey = "task_requests_create"
	RouteTaskRequestsQuery      RouteKey = "task_requests_query"
	RouteDecisionsCreate        RouteKey = "decisions_create"
	RouteDecisionsQuery         RouteKey = "decisions_query"
	RouteCommLogCreate          RouteKey = "comm_log_create"
	RouteCommLogQuery           RouteKey = "comm_log_query"
	RouteHandoffCreate          RouteKey = "handoff_create"
	RouteHandoffQuery           RouteKey = "handoff_query"
	RouteStatusReviewCreate     RouteKey = "status_review_create"
	RouteStatusReviewQuery      RouteKey = "status_review_query"
	RouteLessonCreate           RouteKey = "lesson_create"
	RouteLessonQuery            RouteKey = "lesson_query"
)

type RouteSuccessShape string

const (
	RouteSuccessShapeMentionResolution RouteSuccessShape = "mention_resolution"
	RouteSuccessShapeMerge             RouteSuccessShape = "merge_summary"
	RouteSuccessShapeMutationRow       RouteSuccessShape = "mutation_row"
	RouteSuccessShapeQueryRows         RouteSuccessShape = "query_rows"
	RouteSuccessShapeObjectBlob        RouteSuccessShape = "object_blob"
	RouteSuccessShapeEvidenceAttach    RouteSuccessShape = "evidence_attach"
	RouteSuccessShapeEvidenceHandle    RouteSuccessShape = "evidence_handle"
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

const (
	WorkbookAssessmentsViewSchemaID  = "cartulary.view.assessments.v1"
	WorkbookCommLogViewSchemaID      = "cartulary.view.comm_log.v1"
	WorkbookDecisionsViewSchemaID    = "cartulary.view.decisions.v1"
	WorkbookEvidenceViewSchemaID     = "cartulary.view.evidence.v1"
	WorkbookHandoffViewSchemaID      = "cartulary.view.handoff.v1"
	WorkbookLessonViewSchemaID       = "cartulary.view.lesson.v1"
	WorkbookNotesViewSchemaID        = "cartulary.view.notes.v1"
	WorkbookPartiesViewSchemaID      = "cartulary.view.parties.v1"
	WorkbookStatusReviewViewSchemaID = "cartulary.view.status_review.v1"
	WorkbookTaskRequestsViewSchemaID = "cartulary.view.task_requests.v1"
)

type RouteWebSocketExpectation string

const (
	RouteWebSocketNotApplicable RouteWebSocketExpectation = "n/a"
	RouteWebSocketRecordChanged RouteWebSocketExpectation = "record_changed"
)

type RouteWebSocketChangeExpectation struct {
	ViewSchemaID  string
	BuildRecordID func(RouteInventoryContext) string
	RowVersion    int64
	ChangedKeys   []string
}

type RouteInventoryContext struct {
	IncidentID            string
	ActorUserID           string
	TimelineRecordID      string
	MentionID             string
	MergeSurvivorRecordID string
	MergeLoserRecordID    string
	HostRecordID          string
	IdentityRecordID      string
	IndicatorRecordID     string
	PartyRecordID         string
	AssessmentRecordID    string
	EvidenceRecordID      string
	ObjectBlobID          string
	AlternateObjectBlobID string
	NoteRecordID          string
	TaskRequestRecordID   string
	DecisionRecordID      string
	CommLogRecordID       string
	HandoffRecordID       string
	StatusReviewRecordID  string
	LessonRecordID        string
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

	ProjectionTarget           RouteProjectionTarget
	WebSocketExpectation       RouteWebSocketExpectation
	WebSocketViewSchemaID      string
	BuildWebSocketRecordID     func(RouteInventoryContext) string
	WebSocketRowVersion        int64
	AdditionalWebSocketChanges []RouteWebSocketChangeExpectation

	HarnessRequirements map[RouteHarnessClass]RouteHarnessRequirement
}

func ValidateRouteInventory(t testing.TB, routes []RouteInventoryEntry) {
	t.Helper()

	for _, route := range routes {
		if route.Method == "" {
			t.Fatalf("workbook route %s missing method", route.Key)
		}
		if route.BuildPath == nil {
			t.Fatalf("workbook route %s missing path builder", route.Key)
		}
		if route.BuildBody == nil {
			t.Fatalf("workbook route %s missing body builder", route.Key)
		}
		if route.SuccessStatus == 0 {
			t.Fatalf("workbook route %s missing success status", route.Key)
		}
		if route.SuccessShape == "" {
			t.Fatalf("workbook route %s missing success shape", route.Key)
		}
		if route.HarnessRequirements == nil {
			t.Fatalf("workbook route %s missing harness requirements", route.Key)
		}
		for _, harness := range workbookRouteHarnessClasses {
			requirement, ok := route.HarnessRequirements[harness]
			if !ok {
				t.Fatalf("workbook route %s missing harness requirement for %s", route.Key, harness)
			}
			if requirement != RouteHarnessRequired && requirement != RouteHarnessNotApplicable {
				t.Fatalf("workbook route %s has invalid harness requirement %q for %s", route.Key, requirement, harness)
			}
		}
		if route.RequiresCSRF && route.HarnessRequirements[RouteHarnessCSRF] != RouteHarnessRequired {
			t.Fatalf("workbook route %s requires csrf but matrix marks %s", route.Key, route.HarnessRequirements[RouteHarnessCSRF])
		}
		if !route.RequiresCSRF && route.HarnessRequirements[RouteHarnessCSRF] != RouteHarnessNotApplicable {
			t.Fatalf("workbook route %s does not require csrf but matrix marks %s", route.Key, route.HarnessRequirements[RouteHarnessCSRF])
		}
		switch route.ReplayCapability {
		case RouteReplayNotApplicable:
			if route.HarnessRequirements[RouteHarnessReplayDivergent] != RouteHarnessNotApplicable {
				t.Fatalf("workbook route %s marks replay n/a but matrix requires replay", route.Key)
			}
		case RouteReplayStoredPayloadReuse:
			if route.HarnessRequirements[RouteHarnessReplayDivergent] != RouteHarnessRequired {
				t.Fatalf("workbook route %s marks replay reusable but matrix does not require replay", route.Key)
			}
			if route.BuildDivergentBody == nil || route.ReplayStatus == 0 || route.DivergentStatus == 0 || route.DivergentCode == "" {
				t.Fatalf("workbook route %s missing replay metadata", route.Key)
			}
		default:
			t.Fatalf("workbook route %s has invalid replay capability %q", route.Key, route.ReplayCapability)
		}
		switch route.AuthorizationChange {
		case RouteAuthorizationNotApplicable:
			if route.HarnessRequirements[RouteHarnessAuthorization] != RouteHarnessNotApplicable {
				t.Fatalf("workbook route %s marks authorization n/a but matrix requires authorization", route.Key)
			}
		case RouteAuthorizationDemoteViewer, RouteAuthorizationRemoveMember:
			if route.HarnessRequirements[RouteHarnessAuthorization] != RouteHarnessRequired {
				t.Fatalf("workbook route %s must require authorization re-derivation", route.Key)
			}
			if route.AuthorizationStatus == 0 || route.AuthorizationCode == "" {
				t.Fatalf("workbook route %s missing authorization expectation", route.Key)
			}
		default:
			t.Fatalf("workbook route %s has invalid authorization change %q", route.Key, route.AuthorizationChange)
		}
		requiresViewRow := route.HarnessRequirements[RouteHarnessQueryFieldMatrix] == RouteHarnessRequired ||
			route.HarnessRequirements[RouteHarnessEffects] == RouteHarnessRequired
		if requiresViewRow {
			if route.ExpectedViewSchemaID == "" {
				t.Fatalf("workbook route %s missing expected view schema", route.Key)
			}
			if route.AffectedRecordID == nil {
				t.Fatalf("workbook route %s missing affected record selector", route.Key)
			}
		}
		if route.HarnessRequirements[RouteHarnessQueryFieldMatrix] == RouteHarnessRequired && route.AffectedRecordID == nil {
			t.Fatalf("workbook route %s requires query field coverage but has no affected record selector", route.Key)
		}
		if route.ProjectionTarget == RouteProjectionNotApplicable && route.WebSocketExpectation == RouteWebSocketNotApplicable {
			if route.HarnessRequirements[RouteHarnessEffects] != RouteHarnessNotApplicable {
				t.Fatalf("workbook route %s has no projection or websocket expectation but matrix requires effects", route.Key)
			}
		} else {
			if route.HarnessRequirements[RouteHarnessEffects] != RouteHarnessRequired {
				t.Fatalf("workbook route %s must require effects coverage", route.Key)
			}
		}
		if route.WebSocketExpectation == RouteWebSocketRecordChanged {
			if route.WebSocketViewSchemaID == "" || route.BuildWebSocketRecordID == nil || route.WebSocketRowVersion == 0 {
				t.Fatalf("workbook route %s missing websocket expectation metadata", route.Key)
			}
		}
		for index, change := range route.AdditionalWebSocketChanges {
			if change.ViewSchemaID == "" || change.BuildRecordID == nil || change.RowVersion == 0 {
				t.Fatalf("workbook route %s missing additional websocket expectation metadata at index %d", route.Key, index)
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

func WorkbookRouteInventory(ctx RouteInventoryContext) []RouteInventoryEntry {
	return []RouteInventoryEntry{
		{
			Key:    RouteMentionResolve,
			Name:   "mention resolve route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/entity-mentions/" + fixture.MentionID + "/resolve"
			},
			BuildBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return entitytest.MentionResolveRoutePayload(1, clientTxnID, entitytest.MentionActionResolve, uuidPointer(fixture.HostRecordID), nil)
			},
			BuildDivergentBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return entitytest.MentionResolveRoutePayload(1, clientTxnID, entitytest.MentionActionDismiss, nil, nil)
			},
			AffectedRecordID:      func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.TimelineRecordID },
			SuccessStatus:         http.StatusOK,
			SuccessShape:          RouteSuccessShapeMentionResolution,
			RequiresCSRF:          true,
			ExpectedViewSchemaID:  viewtest.TimelineViewSchemaID,
			ReplayCapability:      RouteReplayStoredPayloadReuse,
			ReplayStatus:          http.StatusOK,
			DivergentStatus:       http.StatusConflict,
			DivergentCode:         "client_txn_conflict",
			AuthorizationChange:   RouteAuthorizationDemoteViewer,
			AuthorizationStatus:   http.StatusForbidden,
			AuthorizationCode:     "authorization_denied",
			ProjectionTarget:      RouteProjectionTimeline,
			WebSocketExpectation:  RouteWebSocketRecordChanged,
			WebSocketViewSchemaID: viewtest.TimelineViewSchemaID,
			BuildWebSocketRecordID: func(fixture RouteInventoryContext) string {
				return fixture.TimelineRecordID
			},
			WebSocketRowVersion: 2,
			AdditionalWebSocketChanges: []RouteWebSocketChangeExpectation{
				{
					ViewSchemaID: viewtest.HostsViewSchemaID,
					BuildRecordID: func(fixture RouteInventoryContext) string {
						return fixture.HostRecordID
					},
					RowVersion:  1,
					ChangedKeys: []string{"host.linked_event_count"},
				},
			},
			HarnessRequirements: requiredWorkbookHarnesses(
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
			ExpectedViewSchemaID:  viewtest.TimelineViewSchemaID,
			ReplayCapability:      RouteReplayStoredPayloadReuse,
			ReplayStatus:          http.StatusOK,
			DivergentStatus:       http.StatusConflict,
			DivergentCode:         "client_txn_conflict",
			AuthorizationChange:   RouteAuthorizationDemoteViewer,
			AuthorizationStatus:   http.StatusForbidden,
			AuthorizationCode:     "authorization_denied",
			ProjectionTarget:      RouteProjectionTimeline,
			WebSocketExpectation:  RouteWebSocketRecordChanged,
			WebSocketViewSchemaID: viewtest.TimelineViewSchemaID,
			BuildWebSocketRecordID: func(fixture RouteInventoryContext) string {
				return fixture.TimelineRecordID
			},
			WebSocketRowVersion: 1,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.HostsViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return entitytest.HostCreatePayload(clientTxnID)
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "host.display_name": "VPN Gateway Divergent", "host.hostname": "VPN-GATEWAY-DIVERGENT"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: viewtest.HostsViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionHosts,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.HostsViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.HostRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: viewtest.HostsViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionHosts,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.IdentitiesViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return entitytest.IdentityCreatePayload(clientTxnID)
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "identity.display_name": "VPN User Divergent", "identity.email": "vpn.user.divergent@example.test", "identity.sam_account_name": "VPNDIV"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: viewtest.IdentitiesViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionIdentities,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.IdentitiesViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.IdentityRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: viewtest.IdentitiesViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionIdentities,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.IndicatorsViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return indicatortest.CreatePayload(clientTxnID)
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "indicator.indicator_type": "ipv4_addr", "indicator.value_kind": "atomic", "indicator.display_value": "203.0.113.25"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: viewtest.IndicatorsViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionIndicators,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.IndicatorsViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.IndicatorRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: viewtest.IndicatorsViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionIndicators,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.TimelineViewSchemaID + "/rows"
			},
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "timeline.activity_synopsis_text": "support timeline create"}
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return map[string]any{"client_txn_id": clientTxnID, "timeline.activity_synopsis_text": "support timeline create divergent"}
			},
			AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: viewtest.TimelineViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionTimeline,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewtest.TimelineViewSchemaID + "/query"
			},
			BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.TimelineRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeQueryRows,
			RequiresCSRF:         false,
			ExpectedViewSchemaID: viewtest.TimelineViewSchemaID,
			ReplayCapability:     RouteReplayNotApplicable,
			AuthorizationChange:  RouteAuthorizationRemoveMember,
			AuthorizationStatus:  http.StatusNotFound,
			AuthorizationCode:    "incident_not_found",
			ProjectionTarget:     RouteProjectionTimeline,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
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
				return timelinetest.TimelineCollectionPatchPayload(timelinetest.FieldHostRefs, 1, clientTxnID, timelinetest.CollectionActions(timelinetest.AddResolvedRefAction("WS-023", mustUUID(fixture.HostRecordID))))
			},
			BuildDivergentBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return timelinetest.TimelineCollectionPatchPayload(timelinetest.FieldHostRefs, 1, clientTxnID, timelinetest.CollectionActions(timelinetest.AddResolvedRefAction("WS-024", mustUUID(fixture.MergeLoserRecordID))))
			},
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.TimelineRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: viewtest.TimelineViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionTimeline,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		{
			Key:    RouteObjectBlobCreate,
			Name:   "object blob create route",
			Method: http.MethodPost,
			BuildPath: func(RouteInventoryContext) string {
				return "/api/v1/object-blobs"
			},
			BuildBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return map[string]any{
					"incident_id":       fixture.IncidentID,
					"client_txn_id":     clientTxnID,
					"byte_size":         14,
					"filename_hint":     " support-object.txt ",
					"content_type_hint": "text/plain",
					"sha256_hex":        "9af4c73b2a919f220f4b008e466b52808a1987122d95ff0f2dde00968e36e844",
				}
			},
			BuildDivergentBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return map[string]any{
					"incident_id":   fixture.IncidentID,
					"client_txn_id": clientTxnID,
					"byte_size":     1,
				}
			},
			SuccessStatus:        http.StatusCreated,
			SuccessShape:         RouteSuccessShapeObjectBlob,
			RequiresCSRF:         true,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionNotApplicable,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
			),
		},
		{
			Key:    RouteEvidenceAttachBlob,
			Name:   "evidence attach blob route",
			Method: http.MethodPost,
			BuildPath: func(fixture RouteInventoryContext) string {
				return "/api/v1/evidence-records/" + fixture.EvidenceRecordID + "/attach-blob"
			},
			BuildBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return map[string]any{
					"object_blob_id":   fixture.ObjectBlobID,
					"base_row_version": 1,
					"client_txn_id":    clientTxnID,
				}
			},
			BuildDivergentBody: func(fixture RouteInventoryContext, clientTxnID string) any {
				return map[string]any{
					"object_blob_id":   fixture.AlternateObjectBlobID,
					"base_row_version": 1,
					"client_txn_id":    clientTxnID,
				}
			},
			AffectedRecordID:      func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.EvidenceRecordID },
			SuccessStatus:         http.StatusOK,
			SuccessShape:          RouteSuccessShapeEvidenceAttach,
			RequiresCSRF:          true,
			ExpectedViewSchemaID:  WorkbookEvidenceViewSchemaID,
			ReplayCapability:      RouteReplayStoredPayloadReuse,
			ReplayStatus:          http.StatusOK,
			DivergentStatus:       http.StatusConflict,
			DivergentCode:         "client_txn_conflict",
			AuthorizationChange:   RouteAuthorizationNotApplicable,
			ProjectionTarget:      RouteProjectionNotApplicable,
			WebSocketExpectation:  RouteWebSocketRecordChanged,
			WebSocketViewSchemaID: WorkbookEvidenceViewSchemaID,
			BuildWebSocketRecordID: func(fixture RouteInventoryContext) string {
				return fixture.EvidenceRecordID
			},
			WebSocketRowVersion: 2,
			HarnessRequirements: requiredWorkbookHarnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessQueryFieldMatrix,
				RouteHarnessEffects,
			),
		},
		evidenceHandleRoute(RouteEvidencePreviewHandle, "evidence preview handle route", "preview-handle", "preview"),
		evidenceHandleRoute(RouteEvidenceDownloadHandle, "evidence download handle route", "download-handle", "download"),
		workbookCreateRoute(RoutePartiesCreate, "parties create route", WorkbookPartiesViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "party.display_name": "Support Party", "party.party_kind": "organization"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "party.display_name": "Support Party Divergent", "party.party_kind": "person"}
		}),
		workbookQueryRoute(RoutePartiesQuery, "parties query route", WorkbookPartiesViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.PartyRecordID }),
		{
			Key:       RoutePartiesPatch,
			Name:      "parties patch route",
			Method:    http.MethodPatch,
			BuildPath: func(fixture RouteInventoryContext) string { return "/api/v1/records/" + fixture.PartyRecordID },
			BuildBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return workbookPatchPayload(WorkbookPartiesViewSchemaID, 1, clientTxnID, "party.primary_email", "support@example.test")
			},
			BuildDivergentBody: func(_ RouteInventoryContext, clientTxnID string) any {
				return workbookPatchPayload(WorkbookPartiesViewSchemaID, 1, clientTxnID, "party.primary_email", "support-divergent@example.test")
			},
			AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return fixture.PartyRecordID },
			SuccessStatus:        http.StatusOK,
			SuccessShape:         RouteSuccessShapeMutationRow,
			RequiresCSRF:         true,
			ExpectedViewSchemaID: WorkbookPartiesViewSchemaID,
			ReplayCapability:     RouteReplayStoredPayloadReuse,
			ReplayStatus:         http.StatusOK,
			DivergentStatus:      http.StatusConflict,
			DivergentCode:        "client_txn_conflict",
			AuthorizationChange:  RouteAuthorizationDemoteViewer,
			AuthorizationStatus:  http.StatusForbidden,
			AuthorizationCode:    "authorization_denied",
			ProjectionTarget:     RouteProjectionNotApplicable,
			WebSocketExpectation: RouteWebSocketNotApplicable,
			HarnessRequirements: requiredWorkbookHarnesses(
				RouteHarnessSurfaceEnvelope,
				RouteHarnessCSRF,
				RouteHarnessReplayDivergent,
				RouteHarnessAuthorization,
				RouteHarnessQueryFieldMatrix,
			),
		},
		workbookCreateRoute(RouteAssessmentsCreate, "assessments create route", WorkbookAssessmentsViewSchemaID, func(fixture RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "assessment.subject_ref": fixture.HostRecordID, "assessment.subject_type": "host", "assessment.assessment_state": "confirmed", "assessment.confidence_score": 55, "assessment.rationale": "Support assessment"}
		}, func(fixture RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "assessment.subject_ref": fixture.HostRecordID, "assessment.subject_type": "host", "assessment.assessment_state": "suspected", "assessment.confidence_score": 25, "assessment.rationale": "Support assessment divergent"}
		}),
		workbookQueryRoute(RouteAssessmentsQuery, "assessments query route", WorkbookAssessmentsViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.AssessmentRecordID }),
		workbookCreateRoute(RouteEvidenceCreate, "evidence create route", WorkbookEvidenceViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "evidence.title": "Support evidence"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "evidence.title": "Support evidence divergent"}
		}),
		workbookQueryRoute(RouteEvidenceQuery, "evidence query route", WorkbookEvidenceViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.EvidenceRecordID }),
		workbookCreateRoute(RouteNotesCreate, "notes create route", WorkbookNotesViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "note.title": "Support note"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "note.title": "Support note divergent"}
		}),
		workbookQueryRoute(RouteNotesQuery, "notes query route", WorkbookNotesViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.NoteRecordID }),
		workbookCreateRoute(RouteTaskRequestsCreate, "task requests create route", WorkbookTaskRequestsViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "task.title": "Support task", "task.task_kind": "collection"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "task.title": "Support task divergent", "task.task_kind": "analysis"}
		}),
		workbookQueryRoute(RouteTaskRequestsQuery, "task requests query route", WorkbookTaskRequestsViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.TaskRequestRecordID }),
		workbookCreateRoute(RouteDecisionsCreate, "decisions create route", WorkbookDecisionsViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "decision.summary": "Support decision", "decision.decision_type": "containment", "decision.rationale": "Support rationale"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "decision.summary": "Support decision divergent", "decision.decision_type": "scope", "decision.rationale": "Support divergent rationale"}
		}),
		workbookQueryRoute(RouteDecisionsQuery, "decisions query route", WorkbookDecisionsViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.DecisionRecordID }),
		workbookCreateRoute(RouteCommLogCreate, "communications log create route", WorkbookCommLogViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "comm_log.comm_type": "briefing", "comm_log.audience": "leadership", "comm_log.channel_or_meeting": "Bridge", "comm_log.summary": "Support communication"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "comm_log.comm_type": "briefing", "comm_log.audience": "team", "comm_log.channel_or_meeting": "Chat", "comm_log.summary": "Support communication divergent"}
		}),
		workbookQueryRoute(RouteCommLogQuery, "communications log query route", WorkbookCommLogViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.CommLogRecordID }),
		workbookCreateRoute(RouteHandoffCreate, "handoff create route", WorkbookHandoffViewSchemaID, func(fixture RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "handoff.incoming_owner_user_id": fixture.ActorUserID, "handoff.current_state_summary": "Support handoff"}
		}, func(fixture RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "handoff.incoming_owner_user_id": fixture.ActorUserID, "handoff.current_state_summary": "Support handoff divergent"}
		}),
		workbookQueryRoute(RouteHandoffQuery, "handoff query route", WorkbookHandoffViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.HandoffRecordID }),
		workbookCreateRoute(RouteStatusReviewCreate, "status review create route", WorkbookStatusReviewViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "status_review.current_state_summary": "Support status"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "status_review.current_state_summary": "Support status divergent"}
		}),
		workbookQueryRoute(RouteStatusReviewQuery, "status review query route", WorkbookStatusReviewViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.StatusReviewRecordID }),
		workbookCreateRoute(RouteLessonCreate, "lesson create route", WorkbookLessonViewSchemaID, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "lesson.summary": "Support lesson"}
		}, func(_ RouteInventoryContext, clientTxnID string) any {
			return map[string]any{"client_txn_id": clientTxnID, "lesson.summary": "Support lesson divergent"}
		}),
		workbookQueryRoute(RouteLessonQuery, "lesson query route", WorkbookLessonViewSchemaID, func(fixture RouteInventoryContext) string { return fixture.LessonRecordID }),
	}
}

func requiredWorkbookHarnesses(required ...RouteHarnessClass) map[RouteHarnessClass]RouteHarnessRequirement {
	requirements := make(map[RouteHarnessClass]RouteHarnessRequirement, len(workbookRouteHarnessClasses))
	for _, harness := range workbookRouteHarnessClasses {
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

func workbookCreateRoute(
	key RouteKey,
	name string,
	viewSchemaID string,
	body func(RouteInventoryContext, string) any,
	divergent func(RouteInventoryContext, string) any,
) RouteInventoryEntry {
	return RouteInventoryEntry{
		Key:    key,
		Name:   name,
		Method: http.MethodPost,
		BuildPath: func(fixture RouteInventoryContext) string {
			return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewSchemaID + "/rows"
		},
		BuildBody:            body,
		BuildDivergentBody:   divergent,
		AffectedRecordID:     func(_ RouteInventoryContext, data map[string]any) string { return rowRecordID(data) },
		SuccessStatus:        http.StatusCreated,
		SuccessShape:         RouteSuccessShapeMutationRow,
		RequiresCSRF:         true,
		ExpectedViewSchemaID: viewSchemaID,
		ReplayCapability:     RouteReplayStoredPayloadReuse,
		ReplayStatus:         http.StatusOK,
		DivergentStatus:      http.StatusConflict,
		DivergentCode:        "client_txn_conflict",
		AuthorizationChange:  RouteAuthorizationDemoteViewer,
		AuthorizationStatus:  http.StatusForbidden,
		AuthorizationCode:    "authorization_denied",
		ProjectionTarget:     RouteProjectionNotApplicable,
		WebSocketExpectation: RouteWebSocketNotApplicable,
		HarnessRequirements: requiredWorkbookHarnesses(
			RouteHarnessSurfaceEnvelope,
			RouteHarnessCSRF,
			RouteHarnessReplayDivergent,
			RouteHarnessAuthorization,
			RouteHarnessQueryFieldMatrix,
		),
	}
}

func workbookQueryRoute(key RouteKey, name string, viewSchemaID string, affected func(RouteInventoryContext) string) RouteInventoryEntry {
	return RouteInventoryEntry{
		Key:    key,
		Name:   name,
		Method: http.MethodPost,
		BuildPath: func(fixture RouteInventoryContext) string {
			return "/api/v1/incidents/" + fixture.IncidentID + "/views/" + viewSchemaID + "/query"
		},
		BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
		AffectedRecordID:     func(fixture RouteInventoryContext, _ map[string]any) string { return affected(fixture) },
		SuccessStatus:        http.StatusOK,
		SuccessShape:         RouteSuccessShapeQueryRows,
		RequiresCSRF:         false,
		ExpectedViewSchemaID: viewSchemaID,
		ReplayCapability:     RouteReplayNotApplicable,
		AuthorizationChange:  RouteAuthorizationRemoveMember,
		AuthorizationStatus:  http.StatusNotFound,
		AuthorizationCode:    "incident_not_found",
		ProjectionTarget:     RouteProjectionNotApplicable,
		WebSocketExpectation: RouteWebSocketNotApplicable,
		HarnessRequirements: requiredWorkbookHarnesses(
			RouteHarnessSurfaceEnvelope,
			RouteHarnessAuthorization,
			RouteHarnessQueryFieldMatrix,
		),
	}
}

func evidenceHandleRoute(key RouteKey, name string, pathSuffix string, _ string) RouteInventoryEntry {
	return RouteInventoryEntry{
		Key:    key,
		Name:   name,
		Method: http.MethodPost,
		BuildPath: func(fixture RouteInventoryContext) string {
			return "/api/v1/evidence-records/" + fixture.EvidenceRecordID + "/" + pathSuffix
		},
		BuildBody:            func(RouteInventoryContext, string) any { return map[string]any{} },
		SuccessStatus:        http.StatusOK,
		SuccessShape:         RouteSuccessShapeEvidenceHandle,
		RequiresCSRF:         true,
		ReplayCapability:     RouteReplayNotApplicable,
		AuthorizationChange:  RouteAuthorizationRemoveMember,
		AuthorizationStatus:  http.StatusNotFound,
		AuthorizationCode:    "evidence_record_not_found",
		ProjectionTarget:     RouteProjectionNotApplicable,
		WebSocketExpectation: RouteWebSocketNotApplicable,
		HarnessRequirements: requiredWorkbookHarnesses(
			RouteHarnessSurfaceEnvelope,
			RouteHarnessCSRF,
			RouteHarnessAuthorization,
		),
	}
}

func workbookPatchPayload(viewSchemaID string, baseRowVersion int64, clientTxnID string, fieldKey string, value any) map[string]any {
	return map[string]any{
		"view_schema_id":   viewSchemaID,
		"base_row_version": baseRowVersion,
		"client_txn_id":    clientTxnID,
		"changes": []map[string]any{
			{"field_key": fieldKey, "value": value},
		},
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
