import {
  coordinationWorkflowTestId,
  currentIncidentRoleTestId,
  dataTestIdSelector,
  entityInspectButtonTestId,
  entityInspectorTestId,
  entityMergeControlTestId,
  entityMergePreconditionDetailsTestId,
  entityReusableIdentifierItemTestId,
  entityReusableIdentifiersSectionTestId,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  genericCreateFieldTestId,
  genericEditRecordSelectTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  incidentAdministrationTestId,
  incidentControlsCloseButtonTestId,
  incidentControlsMenuItemTestId,
  incidentControlsMenuTestId,
  incidentControlsPanelTestId,
  incidentControlsTriggerTestId,
  incidentLandingTestId,
  incidentMembershipCreateButtonTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  savedViewActionMenuTestId,
  savedViewActionMenuTriggerTestId,
  savedViewCreateButtonTestId,
  savedViewDeleteButtonTestId,
  savedViewDuplicateButtonTestId,
  savedViewNameInputTestId,
  savedViewOptionTestId,
  savedViewScopeSelectTestId,
  savedViewSelectorTestId,
  savedViewSetDefaultButtonTestId,
  savedViewSetHomeButtonTestId,
  savedViewUpdateButtonTestId,
  saveStateTestId,
  surfaceTabTestId,
  systemViewSwitcherGroupTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
  timelineEvidenceFileInputTestId,
  timelineInspectorMessageTestId,
  timelineInspectorTestId,
  timelinePreviewRowTestId,
  workbookAddRowButtonTestId,
  workbookFilterPopoverTriggerTestId,
  workbookImportAssistantTestId,
  workbookInspectorToggleTestId,
  workbookShellReadyTestId,
  workbookShellSlotTestId,
  workbookViewBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  createEvent,
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { useRef, useState } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { createAppAuthorizationRecoveryPort } from "../app/api/appShellClient";
import { fetchJSON } from "../services/browserApi";
import { deferred } from "../testing/fetchMockTestSupport";
import {
  errorEnvelope,
  flushWorkbookAsync,
  successEnvelope,
  waitForWorkbookRows,
} from "../testing/timelineWorkbookTestSupport";
import { waitForEntityInspectorReady } from "../testing/workbookInspectorTestSupport";
import { buildGenericCreateRequest } from "./features/generic/genericCreateRequestBuilder";
import { useGenericPartyLinkWorkflow } from "./features/parties/useGenericPartyLinkWorkflow";
import { buildGenericPatchChange } from "./models/genericWorkbookModel";
import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  evidenceViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./models/workbookSurfaceRegistry";
import type { GenericMutationCommandPort } from "./mutations/workbookMutationCommandPorts";
import {
  type WorkbookAccountApplicationMenuProps,
  type WorkbookIncidentControlsRendererProps,
  WorkbookShell as WorkbookShellImpl,
} from "./WorkbookShell";

vi.mock(
  "@cartulary/grid-adapter",
  async () => import("@cartulary/grid-adapter/test-support"),
);

const authorizationRecovery = createAppAuthorizationRecoveryPort({
  loadCurrentSession: (signal) => fetchJSON("/api/v1/auth/session", { signal }),
});

function WorkbookShell(
  props: Omit<Parameters<typeof WorkbookShellImpl>[0], "authorizationRecovery">,
) {
  return (
    <WorkbookShellImpl
      {...props}
      authorizationRecovery={authorizationRecovery}
    />
  );
}

const savedViewId = "11111111-1111-4111-8111-111111111111";
const savedViewCopyId = "33333333-3333-4333-8333-333333333333";
const testUserId = "99999999-9999-4999-8999-999999999999";
const testTimestamp = "2026-07-31T20:00:00Z";
const testAccountMenuTriggerTestId = "test-account-menu-trigger";

function TestAccountApplicationMenu({
  currentIncidentRole,
  incidentControls,
}: WorkbookAccountApplicationMenuProps) {
  const [open, setOpen] = useState(false);
  const [controlsOpen, setControlsOpen] = useState(false);
  const triggerRef = useRef<HTMLButtonElement | null>(null);

  return (
    <div>
      <button
        ref={triggerRef}
        aria-expanded={open}
        aria-haspopup="menu"
        data-testid={testAccountMenuTriggerTestId}
        type="button"
        onClick={() => {
          setOpen((current) => !current);
        }}
      >
        Account
      </button>
      {open ? (
        <div role="menu">
          <div
            aria-disabled="true"
            data-testid={currentIncidentRoleTestId()}
            role="menuitem"
            tabIndex={0}
          >
            Current incident role: {currentIncidentRole || "viewer"}
          </div>
          <button
            data-testid={incidentLandingTestId("return")}
            role="menuitem"
            type="button"
          >
            Incidents
          </button>
          <button
            aria-controls={
              controlsOpen ? incidentControlsMenuTestId() : undefined
            }
            aria-expanded={controlsOpen}
            aria-haspopup="menu"
            data-active-section={incidentControls.activeSection}
            data-testid={incidentControlsTriggerTestId()}
            role="menuitem"
            type="button"
            onClick={() => {
              setControlsOpen((current) => !current);
            }}
          >
            Controls
          </button>
          {controlsOpen ? (
            <div data-testid={incidentControlsMenuTestId()} role="menu">
              {incidentControls.items.map((item) => (
                <button
                  key={item.section}
                  data-testid={incidentControlsMenuItemTestId(item.section)}
                  role="menuitem"
                  type="button"
                  onClick={() => {
                    setOpen(false);
                    setControlsOpen(false);
                    incidentControls.onSelectSection(
                      item.section,
                      triggerRef.current,
                    );
                  }}
                >
                  {item.label}
                </button>
              ))}
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function TestIncidentControls({
  activeSection,
}: WorkbookIncidentControlsRendererProps) {
  if (activeSection === "summary") {
    return (
      <>
        <div data-testid={incidentAdministrationTestId("summary-key")}>
          IR-1
        </div>
        <div
          data-testid={incidentAdministrationTestId("pref-default-sheet-ref")}
        >
          View schema: Timeline (cartulary.view.timeline.v2)
        </div>
        <div data-testid={incidentAdministrationTestId("pref-home-sheet-ref")}>
          Saved view: {savedViewId}
        </div>
      </>
    );
  }
  if (activeSection === "incident-fields") {
    return (
      <button
        data-testid={incidentAdministrationTestId("patch-button")}
        type="button"
      >
        Patch
      </button>
    );
  }
  if (activeSection === "memberships") {
    return (
      <button
        data-testid={incidentMembershipCreateButtonTestId()}
        type="button"
      >
        Create membership
      </button>
    );
  }
  return null;
}

async function openTimelineInspectorFromContext(recordId: string) {
  fireEvent.contextMenu(
    await screen.findByTestId(
      rowCellTestId(recordId, "timeline.activity_synopsis_text"),
    ),
    { clientX: 32, clientY: 48 },
  );
  fireEvent.click(await screen.findByTestId(rowInspectButtonTestId(recordId)));
  await screen.findByTestId(timelineInspectorTestId());
}

type TestSavedViewResource = {
  created_at: string;
  saved_view_id: string;
  incident_id: string;
  view_schema_id: string;
  display_name: string;
  scope: "private" | "shared" | "system";
  query_json: unknown;
  layout_json: unknown;
  owner_user_id: string | null;
  saved_view_version: number;
  updated_at: string;
};

function testSavedViewResource(
  overrides: Partial<TestSavedViewResource> &
    Pick<TestSavedViewResource, "saved_view_id" | "view_schema_id">,
): TestSavedViewResource {
  return {
    created_at: testTimestamp,
    display_name: "Saved view",
    incident_id: "10000000-0000-4000-8000-000000000001",
    layout_json: {
      column_order: [],
      column_widths: [],
      hidden_field_keys: [],
      layout_schema_id: "cartulary.layout.v1",
    },
    owner_user_id: testUserId,
    query_json: { filters: [], sort: [] },
    saved_view_version: 1,
    scope: "private",
    updated_at: testTimestamp,
    ...overrides,
  };
}

function requireField(
  contract: ViewContract,
  fieldKey: string,
): ViewFieldContract {
  const field = contract.fieldMap[fieldKey];
  if (!field) {
    throw new Error(`Missing field ${fieldKey} on ${contract.viewSchemaId}`);
  }
  return field;
}

type SurfaceTestRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
};

type SurfaceStartupSelection = {
  selected_sheet_ref: { kind: "saved_view" | "view_schema"; id: string };
  selected_view_schema_id: string;
  selected_saved_view: unknown | null;
  source: "default" | "explicit" | "home" | "timeline";
};

type SurfaceQueryResponseOverride = (
  viewSchemaId: string,
  init: RequestInit | undefined,
) => Promise<Response> | Response | null;

type SurfaceTestScenario = {
  attachErrorByRecordID: Record<string, Response>;
  evidenceRows: SurfaceTestRow[];
  genericRowsByView: Record<string, SurfaceTestRow[]>;
  handleErrorByRecordID: Record<string, Response>;
  handleHrefByRecordID: Record<string, { download?: string; preview?: string }>;
  incidentStatus: "active" | "closed";
  mergeResponseOverride:
    | ((
        survivorRecordId: string,
        init: RequestInit | undefined,
      ) => Promise<Response> | Response | null)
    | null;
  pendingQueryCount: number;
  pendingDeferredQueryIds: Set<string>;
  queryResponseOverride: SurfaceQueryResponseOverride | null;
  queryTrace: string[];
  recordPatchResponseOverride:
    | ((
        recordId: string,
        init: RequestInit | undefined,
      ) => Promise<Response> | Response | null)
    | null;
  savedViews: TestSavedViewResource[];
  startupResponseOverride: (() => Promise<Response> | Response | null) | null;
  startupSelection: SurfaceStartupSelection;
  timelineRows: SurfaceTestRow[];
  uploadShouldFail: boolean;
  workbookDefaultSheetRef: {
    kind: "saved_view" | "view_schema";
    id: string;
  } | null;
  workbookHomeSheetRef: {
    kind: "saved_view" | "view_schema";
    id: string;
  } | null;
  deferQuery: (viewSchemaId: string) => {
    abort: () => void;
    reject: (reason?: unknown) => void;
    resolve: (response: Response) => void;
  };
};

function appendSurfaceQueryTrace(
  scenario: SurfaceTestScenario,
  entry: string,
): void {
  scenario.queryTrace.push(entry);
  if (scenario.queryTrace.length > 64) {
    scenario.queryTrace.shift();
  }
}

function createSurfaceTestScenario(): SurfaceTestScenario {
  const scenario: SurfaceTestScenario = {
    attachErrorByRecordID: {},
    evidenceRows: [],
    genericRowsByView: {},
    handleErrorByRecordID: {},
    handleHrefByRecordID: {},
    incidentStatus: "active",
    mergeResponseOverride: null,
    pendingQueryCount: 0,
    pendingDeferredQueryIds: new Set(),
    queryResponseOverride: null,
    queryTrace: [],
    recordPatchResponseOverride: null,
    savedViews: [],
    startupResponseOverride: null,
    startupSelection: {
      selected_sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
      selected_view_schema_id: timelineViewSchemaId,
      selected_saved_view: null,
      source: "timeline",
    },
    timelineRows: [],
    uploadShouldFail: false,
    workbookDefaultSheetRef: null,
    workbookHomeSheetRef: null,
    deferQuery: () => {
      throw new Error("Surface query deferral is not initialized.");
    },
  };
  scenario.deferQuery = (viewSchemaId) => {
    if (scenario.pendingDeferredQueryIds.has(viewSchemaId)) {
      throw new Error(`Query ${viewSchemaId} is already deferred.`);
    }
    const pending = deferred<Response>();
    const previousOverride = scenario.queryResponseOverride;
    let settled = false;
    let removeAbortListener = () => {};
    scenario.pendingDeferredQueryIds.add(viewSchemaId);
    const settle = (
      complete: () => void,
      alreadySettledMessage: string,
    ): void => {
      if (settled) {
        throw new Error(alreadySettledMessage);
      }
      settled = true;
      removeAbortListener();
      scenario.pendingDeferredQueryIds.delete(viewSchemaId);
      complete();
    };
    scenario.queryResponseOverride = (candidateViewSchemaId, init) => {
      if (candidateViewSchemaId === viewSchemaId) {
        scenario.queryResponseOverride = previousOverride;
        const signal = init?.signal;
        if (signal) {
          const abort = () => {
            if (settled) {
              return;
            }
            settle(
              () => pending.reject(new DOMException("Aborted", "AbortError")),
              `Deferred query ${viewSchemaId} already settled.`,
            );
          };
          signal.addEventListener("abort", abort, { once: true });
          removeAbortListener = () => {
            signal.removeEventListener("abort", abort);
          };
          if (signal.aborted) {
            abort();
          }
        }
        return pending.promise;
      }
      return previousOverride?.(candidateViewSchemaId, init) ?? null;
    };
    return {
      abort: () => {
        settle(
          () => pending.reject(new DOMException("Aborted", "AbortError")),
          `Deferred query ${viewSchemaId} already settled.`,
        );
      },
      reject: (reason?: unknown) => {
        settle(
          () => pending.reject(reason),
          `Deferred query ${viewSchemaId} already settled.`,
        );
      },
      resolve: (response: Response) => {
        settle(
          () => pending.resolve(response),
          `Deferred query ${viewSchemaId} already settled.`,
        );
      },
    };
  };
  return scenario;
}

describe("WorkbookShell surface selection", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let scenario: SurfaceTestScenario;

  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    vi.spyOn(document, "cookie", "get").mockReturnValue(
      "cartulary_csrf=evidence-shell-csrf",
    );
    const currentScenario = createSurfaceTestScenario();
    scenario = currentScenario;
    fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: testUserId,
          memberships: [
            {
              incident_id: "10000000-0000-4000-8000-000000000001",
              role: "admin",
            },
          ],
        });
      }
      if (
        url.endsWith("/api/v1/incidents/10000000-0000-4000-8000-000000000001")
      ) {
        return successEnvelope({
          closed_at:
            currentScenario.incidentStatus === "closed" ? testTimestamp : null,
          created_at: testTimestamp,
          created_by_user_id: testUserId,
          incident_id: "10000000-0000-4000-8000-000000000001",
          incident_key: "IR-1",
          title: "Incident 1",
          description: null,
          severity: null,
          tlp: null,
          current_phase: null,
          primary_external_case_ref: null,
          incident_version: 1,
          status: currentScenario.incidentStatus,
          updated_at: testTimestamp,
          updated_by_user_id: testUserId,
        });
      }
      if (
        url.endsWith(
          "/api/v1/incidents/10000000-0000-4000-8000-000000000001/memberships",
        )
      ) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "10000000-0000-4000-8000-000000000001",
              user_id: testUserId,
              display_name: "Admin User",
              added_by_user_id: testUserId,
              joined_at: testTimestamp,
              role: "admin",
              membership_version: 1,
              updated_at: testTimestamp,
              updated_by_user_id: testUserId,
            },
          ],
        });
      }
      if (
        url.endsWith(
          "/api/v1/incidents/10000000-0000-4000-8000-000000000001/workbook-preferences/default",
        )
      ) {
        const request = JSON.parse(String(init?.body ?? "{}")) as {
          default_sheet_ref?: SurfaceTestScenario["workbookDefaultSheetRef"];
        };
        if (method === "PUT" && request.default_sheet_ref !== undefined) {
          currentScenario.workbookDefaultSheetRef = request.default_sheet_ref;
        }
        return successEnvelope({
          default_sheet_ref: currentScenario.workbookDefaultSheetRef,
          incident_id: "10000000-0000-4000-8000-000000000001",
          created_at: testTimestamp,
          updated_at: testTimestamp,
          updated_by_user_id: testUserId,
        });
      }
      if (
        url.endsWith(
          "/api/v1/incidents/10000000-0000-4000-8000-000000000001/workbook-preferences/me",
        )
      ) {
        const request = JSON.parse(String(init?.body ?? "{}")) as {
          home_sheet_ref?: SurfaceTestScenario["workbookHomeSheetRef"];
        };
        if (method === "PUT" && request.home_sheet_ref !== undefined) {
          currentScenario.workbookHomeSheetRef = request.home_sheet_ref;
        }
        return successEnvelope({
          home_sheet_ref: currentScenario.workbookHomeSheetRef,
          incident_id: "10000000-0000-4000-8000-000000000001",
          created_at: testTimestamp,
          updated_at: testTimestamp,
          user_id: testUserId,
        });
      }
      if (
        url.includes(
          "/api/v1/incidents/10000000-0000-4000-8000-000000000001/workbook-startup",
        )
      ) {
        const override = await currentScenario.startupResponseOverride?.();
        if (override) {
          return override;
        }
        return successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          ...currentScenario.startupSelection,
          extension_workspace_availability: {
            schema_id: "cartulary.extension_workspace_availability.v1",
            incident_id: "10000000-0000-4000-8000-000000000001",
            workspaces: [],
          },
          cleared_pointers: [],
          home_sheet_ref: null,
          default_sheet_ref: null,
        });
      }
      if (
        method === "POST" &&
        url.includes(
          "/api/v1/incidents/10000000-0000-4000-8000-000000000001/saved-views",
        )
      ) {
        const body = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        const created = testSavedViewResource({
          saved_view_id:
            typeof body.display_name === "string" &&
            body.display_name.endsWith(" Copy")
              ? savedViewCopyId
              : "44444444-4444-4444-8444-444444444444",
          view_schema_id: String(body.view_schema_id),
          display_name: String(body.display_name ?? "Saved view"),
          scope: body.scope === "shared" ? "shared" : "private",
          query_json: body.query_json ?? {},
          layout_json: body.layout_json ?? {},
          saved_view_version: 1,
        });
        currentScenario.savedViews = [
          ...currentScenario.savedViews.filter(
            (savedView) => savedView.saved_view_id !== created.saved_view_id,
          ),
          created,
        ];
        return successEnvelope(created, 201);
      }
      const savedViewMutationMatch = url.match(
        /\/api\/v1\/incidents\/10000000-0000-4000-8000-000000000001\/saved-views\/([^/?]+)$/,
      );
      if (savedViewMutationMatch && method === "PATCH") {
        const savedViewID = decodeURIComponent(savedViewMutationMatch[1] ?? "");
        const existing = currentScenario.savedViews.find(
          (savedView) => savedView.saved_view_id === savedViewID,
        );
        if (!existing) {
          return errorEnvelope("not_found", 404);
        }
        const body = JSON.parse(String(init?.body ?? "{}")) as Record<
          string,
          unknown
        >;
        const updated: TestSavedViewResource = {
          ...existing,
          display_name: String(body.display_name ?? existing.display_name),
          scope: body.scope === "shared" ? "shared" : "private",
          query_json: body.query_json ?? existing.query_json,
          layout_json: body.layout_json ?? existing.layout_json,
          saved_view_version: existing.saved_view_version + 1,
        };
        currentScenario.savedViews = currentScenario.savedViews.map(
          (savedView) =>
            savedView.saved_view_id === savedViewID ? updated : savedView,
        );
        return successEnvelope(updated);
      }
      if (savedViewMutationMatch && method === "DELETE") {
        const savedViewID = decodeURIComponent(savedViewMutationMatch[1] ?? "");
        currentScenario.savedViews = currentScenario.savedViews.filter(
          (savedView) => savedView.saved_view_id !== savedViewID,
        );
        return successEnvelope({ deleted: true, saved_view_id: savedViewID });
      }
      if (
        method === "GET" &&
        url.includes(
          "/api/v1/incidents/10000000-0000-4000-8000-000000000001/saved-views",
        )
      ) {
        return new Response(
          JSON.stringify({
            data: { saved_views: currentScenario.savedViews },
            meta: {
              paging: { has_more: false, limit: 100, next_cursor: null },
              request_id: "req-saved-view-list",
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        );
      }
      const evidenceHandleMatch = url.match(
        /\/api\/v1\/evidence-records\/([^/]+)\/(preview|download)-handle$/,
      );
      if (evidenceHandleMatch) {
        const recordId = decodeURIComponent(evidenceHandleMatch[1] ?? "");
        const kind = evidenceHandleMatch[2] as "download" | "preview";
        const error = currentScenario.handleErrorByRecordID[recordId];
        if (error) {
          return error;
        }
        const href =
          currentScenario.handleHrefByRecordID[recordId]?.[kind] ??
          `/api/v1/evidence-handles/${kind}-token`;
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000001001",
          record_id: recordId,
          object_blob_id: "00000000-0000-4000-8000-000000003001",
          handle_kind: kind,
          href,
          method: "GET",
          expires_at: "2026-07-26T12:05:00Z",
          single_use: true,
          media_class: "text",
          filename: "evidence.txt",
          ...(kind === "preview" ? { preview_kind: "text_inline" } : {}),
          disposition: kind === "preview" ? "inline" : "attachment",
          content_type: "text/plain",
          size_bytes: 18,
          sha256: null,
          evidence_lifecycle_state: "available",
          upload_state: "available",
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${evidenceViewSchemaId}/rows`,
        )
      ) {
        return successEnvelope(
          {
            view_schema_id: evidenceViewSchemaId,
            change_set_id: "30000000-0000-4000-8000-000000000001",
            row: evidenceRow(
              "00000000-0000-4000-8000-000000004001",
              1,
              "Attached screenshot",
            ),
          },
          201,
        );
      }
      if (method === "POST" && url.endsWith("/api/v1/object-blobs")) {
        const request = JSON.parse(String(init?.body)) as {
          readonly byte_size: number;
          readonly content_type_hint: string | null;
          readonly filename_hint: string | null;
          readonly incident_id: string;
        };
        return successEnvelope(
          {
            incident_id: request.incident_id,
            object_blob_id: "00000000-0000-4000-8000-000000003001",
            upload_state: "pending",
            target_expires_at: "2026-07-26T12:05:00Z",
            pending_expires_at: "2026-07-26T12:10:00Z",
            upload_target: {
              href: "/api/v1/object-uploads/test-token",
              method: "PUT",
              expires_at: "2026-07-26T12:05:00Z",
              headers: {
                "Content-Type":
                  request.content_type_hint ?? "application/octet-stream",
                "X-Upload-Contract": "evidence_lifecycle",
              },
            },
            accepted_contract: {
              incident_id: request.incident_id,
              byte_size: request.byte_size,
              filename_hint: request.filename_hint,
              content_type_hint: request.content_type_hint,
              sha256_hex: null,
            },
          },
          201,
        );
      }
      if (
        method === "PUT" &&
        url.endsWith("/api/v1/object-uploads/test-token")
      ) {
        return new Response(null, {
          status: currentScenario.uploadShouldFail ? 500 : 200,
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          "/api/v1/evidence-records/00000000-0000-4000-8000-000000004001/attach-blob",
        )
      ) {
        const row = evidenceStateRow(
          "00000000-0000-4000-8000-000000004001",
          2,
          "Attached screenshot",
          { lifecycleState: "requested", uploadState: "available" },
        );
        currentScenario.evidenceRows = [
          row,
          ...currentScenario.evidenceRows.filter(
            (candidate) => candidate.record_id !== row.record_id,
          ),
        ];
        return successEnvelope({
          view_schema_id: evidenceViewSchemaId,
          change_set_id: "30000000-0000-4000-8000-000000000001",
          row,
          object_blob_id: "00000000-0000-4000-8000-000000003001",
        });
      }
      const attachMatch = url.match(
        /\/api\/v1\/evidence-records\/([^/]+)\/attach-blob$/,
      );
      if (method === "POST" && attachMatch) {
        const recordId = decodeURIComponent(attachMatch[1] ?? "");
        const error = currentScenario.attachErrorByRecordID[recordId];
        if (error) {
          return error;
        }
        const row = evidenceStateRow(recordId, 2, "Attached evidence", {
          lifecycleState: "requested",
          uploadState: "available",
        });
        currentScenario.evidenceRows = [
          row,
          ...currentScenario.evidenceRows.filter(
            (candidate) => candidate.record_id !== recordId,
          ),
        ];
        return successEnvelope({
          view_schema_id: evidenceViewSchemaId,
          change_set_id: "30000000-0000-4000-8000-000000000001",
          row,
          object_blob_id: "00000000-0000-4000-8000-000000003001",
        });
      }
      if (method === "PATCH" && url.includes("/api/v1/records/")) {
        const recordPatchMatch = url.match(
          /\/api\/v1\/records\/([^/?]+)(?:\?.*)?$/,
        );
        if (recordPatchMatch) {
          const override = await currentScenario.recordPatchResponseOverride?.(
            recordPatchMatch[1] ?? "",
            init,
          );
          if (override) {
            return override;
          }
          const body = JSON.parse(String(init?.body ?? "{}")) as {
            changes?: Array<{ field_key?: unknown; value?: unknown }>;
            view_schema_id?: unknown;
          };
          if (body.view_schema_id === evidenceViewSchemaId) {
            const current = currentScenario.evidenceRows.find(
              (candidate) => candidate.record_id === recordPatchMatch[1],
            );
            const lifecycle = body.changes?.find(
              (change) =>
                change.field_key === "evidence.lifecycle_state" &&
                typeof change.value === "string",
            )?.value;
            if (current && typeof lifecycle === "string") {
              const row = evidenceStateRow(
                current.record_id,
                current.row_version + 1,
                String(current.cells["evidence.title"]?.value ?? "Evidence"),
                {
                  lifecycleState: lifecycle,
                  uploadState: String(
                    current.cells["evidence.upload_state"]?.value ?? "pending",
                  ),
                },
              );
              currentScenario.evidenceRows = [
                row,
                ...currentScenario.evidenceRows.filter(
                  (candidate) => candidate.record_id !== row.record_id,
                ),
              ];
              return successEnvelope({
                view_schema_id: evidenceViewSchemaId,
                change_set_id: "30000000-0000-4000-8000-000000000002",
                row,
              });
            }
          }
        }
      }
      const entityPasteMatch = url.match(
        /\/api\/v1\/incidents\/10000000-0000-4000-8000-000000000001\/views\/([^/]+)\/clipboard-paste$/,
      );
      if (method === "POST" && entityPasteMatch) {
        const viewSchemaId = decodeURIComponent(entityPasteMatch[1] ?? "");
        if (
          viewSchemaId === hostsViewSchemaId ||
          viewSchemaId === identitiesViewSchemaId
        ) {
          const body = JSON.parse(String(init?.body ?? "{}")) as {
            clipboard_text?: unknown;
          };
          const nextRows = applyEntityClipboardPaste(
            viewSchemaId,
            currentScenario.genericRowsByView[viewSchemaId] ?? [],
            typeof body.clipboard_text === "string" ? body.clipboard_text : "",
          );
          currentScenario.genericRowsByView[viewSchemaId] = nextRows.allRows;
          return successEnvelope({
            view_schema_id: viewSchemaId,
            change_set_id: "30000000-0000-4000-8000-000000000001",
            rows: nextRows.changedRows,
            conflicts: [],
          });
        }
      }
      const mergeMatch = url.match(/\/api\/v1\/records\/([^/]+)\/merge$/);
      if (method === "POST" && mergeMatch) {
        const survivorRecordId = decodeURIComponent(mergeMatch[1] ?? "");
        const override = await currentScenario.mergeResponseOverride?.(
          survivorRecordId,
          init,
        );
        if (override) {
          return override;
        }
        const body = JSON.parse(String(init?.body ?? "{}")) as {
          loser_record_id?: unknown;
        };
        const loserRecordId =
          typeof body.loser_record_id === "string" ? body.loser_record_id : "";
        const viewSchemaId =
          [hostsViewSchemaId, identitiesViewSchemaId].find((candidate) =>
            (currentScenario.genericRowsByView[candidate] ?? []).some(
              (row) => row.record_id === survivorRecordId,
            ),
          ) ?? hostsViewSchemaId;
        const rows = currentScenario.genericRowsByView[viewSchemaId] ?? [];
        const survivor = rows.find((row) => row.record_id === survivorRecordId);
        currentScenario.genericRowsByView[viewSchemaId] = rows
          .filter((row) => row.record_id !== loserRecordId)
          .map((row) =>
            row.record_id === survivorRecordId
              ? { ...row, row_version: row.row_version + 1 }
              : row,
          );
        return successEnvelope({
          incident_id: "00000000-0000-4000-8000-000000000001",
          record_type: viewSchemaId === hostsViewSchemaId ? "host" : "identity",
          survivor_record_id: survivorRecordId,
          loser_record_id: loserRecordId,
          survivor_row_version: (survivor?.row_version ?? 1) + 1,
          loser_row_version: 2,
          change_set_id: "30000000-0000-4000-8000-000000000001",
          merged_into_record_id: survivorRecordId,
          merge_summary: {
            record_type:
              viewSchemaId === hostsViewSchemaId ? "host" : "identity",
            repointed_mention_resolution_count: 1,
            repointed_link_count: 0,
            deduped_link_count: 0,
            repointed_tag_count: 0,
            deduped_tag_count: 0,
            repointed_assessment_count: 0,
            exact_match_classes: [],
            suggestion_aliases_copied_count: 0,
            suggestion_alias_duplicate_noop_count: 0,
            provenance_only_retained_count: 0,
          },
        });
      }
      if (
        method === "PATCH" &&
        url.endsWith("/api/v1/records/21000000-0000-4000-8000-000000000001")
      ) {
        const row = timelineRow(
          "21000000-0000-4000-8000-000000000001",
          2,
          "Selected row",
          1,
        );
        currentScenario.timelineRows = [row];
        return successEnvelope({
          view_schema_id: timelineViewSchemaId,
          change_set_id: "30000000-0000-4000-8000-000000000001",
          row,
        });
      }
      const viewQueryMatch = url.match(
        /\/api\/v1\/incidents\/10000000-0000-4000-8000-000000000001\/views\/([^/]+)\/query(?:\?.*)?$/,
      );
      if (viewQueryMatch) {
        const viewSchemaId = decodeURIComponent(viewQueryMatch[1] ?? "");
        appendSurfaceQueryTrace(currentScenario, `requested:${viewSchemaId}`);
        currentScenario.pendingQueryCount += 1;
        try {
          const override = currentScenario.queryResponseOverride?.(
            viewSchemaId,
            init,
          );
          const response =
            override ??
            successEnvelope({
              incident_id: "10000000-0000-4000-8000-000000000001",
              view_schema_id: viewSchemaId,
              rows:
                viewSchemaId === evidenceViewSchemaId
                  ? currentScenario.evidenceRows
                  : viewSchemaId === timelineViewSchemaId
                    ? currentScenario.timelineRows
                    : (currentScenario.genericRowsByView[viewSchemaId] ?? []),
            });
          const resolvedResponse = await response;
          appendSurfaceQueryTrace(currentScenario, `resolved:${viewSchemaId}`);
          return resolvedResponse;
        } catch (error) {
          appendSurfaceQueryTrace(
            currentScenario,
            `${error instanceof DOMException && error.name === "AbortError" ? "aborted" : "rejected"}:${viewSchemaId}`,
          );
          throw error;
        } finally {
          currentScenario.pendingQueryCount -= 1;
        }
      }
      return successEnvelope({});
    });
    vi.stubGlobal("fetch", fetchMock);
    vi.stubGlobal(
      "WebSocket",
      class {
        onmessage: ((event: MessageEvent) => void) | null = null;

        close() {}
      } as unknown as typeof WebSocket,
    );
  });

  afterEach(async () => {
    cleanup();
    await Promise.resolve();
    await Promise.resolve();
    const pendingQueryError =
      scenario.pendingQueryCount === 0 &&
      scenario.pendingDeferredQueryIds.size === 0
        ? null
        : new Error(
            `Surface test left ${scenario.pendingQueryCount} query response(s) and deferred view_schema_ids=${
              [...scenario.pendingDeferredQueryIds].join(",") || "(none)"
            } pending. Trace: ${scenario.queryTrace.join(" ") || "(empty)"}`,
          );
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
    if (pendingQueryError) {
      throw pendingQueryError;
    }
  });

  it("keeps late surface fixture requests bound to the originating test scenario", async () => {
    const originatingScenario = scenario;
    originatingScenario.genericRowsByView[hostsViewSchemaId] = [
      hostRow({
        displayName: "Origin host",
        hostname: "origin.example.test",
        recordId: "00000000-0000-4000-8000-000000000702",
        rowVersion: 1,
      }),
    ];
    const lateRequest = Promise.resolve().then(() =>
      fetch(
        `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${hostsViewSchemaId}/query`,
      ),
    );
    const replacementScenario = createSurfaceTestScenario();
    replacementScenario.genericRowsByView[hostsViewSchemaId] = [
      hostRow({
        displayName: "Replacement host",
        hostname: "replacement.example.test",
        recordId: "00000000-0000-4000-8000-000000000703",
        rowVersion: 1,
      }),
    ];
    scenario = replacementScenario;

    try {
      const response = await lateRequest;
      const envelope = (await response.json()) as {
        data: { rows: SurfaceTestRow[] };
      };
      expect(envelope.data.rows.map((row) => row.record_id)).toEqual([
        "00000000-0000-4000-8000-000000000702",
      ]);
      expect(originatingScenario.queryTrace).toEqual([
        `requested:${hostsViewSchemaId}`,
        `resolved:${hostsViewSchemaId}`,
      ]);
      expect(replacementScenario.queryTrace).toEqual([]);
    } finally {
      scenario = originatingScenario;
    }
  });

  it("keeps closed incidents readable while disabling grid mutation entry points", async () => {
    scenario.incidentStatus = "closed";

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    expect(await screen.findByText("Closed, read-only")).toBeTruthy();
    const addRow = await screen.findByTestId(
      workbookAddRowButtonTestId(timelineViewSchemaId),
    );
    expect((addRow as HTMLButtonElement).disabled).toBe(true);
    await waitFor(() => {
      expect(screen.getByRole("grid").getAttribute("aria-readonly")).toBe(
        "true",
      );
    });
  });

  it("opens incident controls from a menu into a bounded drawer without mounting all controls from the trigger", async () => {
    scenario.workbookDefaultSheetRef = {
      kind: "view_schema",
      id: timelineViewSchemaId,
    };
    scenario.workbookHomeSheetRef = {
      kind: "saved_view",
      id: savedViewId,
    };

    render(
      <WorkbookShell
        accountApplicationMenu={(props) => (
          <TestAccountApplicationMenu {...props} />
        )}
        incidentId="10000000-0000-4000-8000-000000000001"
        renderIncidentControls={(props) => <TestIncidentControls {...props} />}
      />,
    );

    const accountTrigger = screen.getByTestId(testAccountMenuTriggerTestId);
    expect(screen.queryByTestId(currentIncidentRoleTestId())).toBeNull();
    expect(screen.queryByTestId(incidentControlsTriggerTestId())).toBeNull();
    expect(screen.queryByTestId(incidentLandingTestId("return"))).toBeNull();
    expect(screen.queryByTestId(incidentControlsPanelTestId())).toBeNull();

    fireEvent.click(accountTrigger);
    expect(screen.getByTestId(currentIncidentRoleTestId()).textContent).toBe(
      "Current incident role: viewer",
    );
    expect(screen.getByTestId(incidentLandingTestId("return"))).toBeTruthy();
    const trigger = screen.getByTestId(incidentControlsTriggerTestId());
    expect(trigger.getAttribute("aria-haspopup")).toBe("menu");
    expect(trigger.getAttribute("data-active-section")).toBe("summary");
    fireEvent.click(trigger);
    const menu = await screen.findByTestId(incidentControlsMenuTestId());
    expect(menu.getAttribute("role")).toBe("menu");
    expect(screen.queryByTestId(incidentControlsPanelTestId())).toBeNull();
    expect(
      screen
        .getByTestId(incidentControlsMenuItemTestId("incident-fields"))
        .getAttribute("role"),
    ).toBe("menuitem");

    fireEvent.click(
      screen.getByTestId(incidentControlsMenuItemTestId("summary")),
    );
    const panel = await screen.findByTestId(incidentControlsPanelTestId());
    expect(panel.getAttribute("role")).toBe("dialog");
    expect(panel.textContent).toContain("Summary and preferences");
    expect(screen.queryByTestId(incidentControlsMenuTestId())).toBeNull();
    await waitFor(() => {
      expect(document.activeElement).toBe(
        screen.getByTestId(incidentControlsCloseButtonTestId()),
      );
    });
    expect(
      (await screen.findByTestId(incidentAdministrationTestId("summary-key")))
        .textContent,
    ).toBe("IR-1");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-default-sheet-ref"))
        .textContent,
    ).toBe("View schema: Timeline (cartulary.view.timeline.v2)");
    expect(
      screen.getByTestId(incidentAdministrationTestId("pref-home-sheet-ref"))
        .textContent,
    ).toBe(`Saved view: ${savedViewId}`);

    fireEvent.keyDown(panel, { key: "Escape" });
    await waitFor(() => {
      expect(screen.queryByTestId(incidentControlsPanelTestId())).toBeNull();
    });
    await waitFor(() => {
      expect(document.activeElement).toBe(accountTrigger);
    });

    fireEvent.click(accountTrigger);
    expect(
      screen
        .getByTestId(incidentControlsTriggerTestId())
        .getAttribute("data-active-section"),
    ).toBe("summary");
    fireEvent.click(screen.getByTestId(incidentControlsTriggerTestId()));
    fireEvent.click(
      await screen.findByTestId(
        incidentControlsMenuItemTestId("incident-fields"),
      ),
    );
    expect(
      await screen.findByTestId(incidentAdministrationTestId("patch-button")),
    ).toBeTruthy();
    expect(
      screen.queryByTestId(incidentAdministrationTestId("summary-key")),
    ).toBeNull();
    fireEvent.click(screen.getByTestId(incidentControlsCloseButtonTestId()));
    await waitFor(() => {
      expect(screen.queryByTestId(incidentControlsPanelTestId())).toBeNull();
    });
    await waitFor(() => {
      expect(document.activeElement).toBe(accountTrigger);
    });

    fireEvent.click(accountTrigger);
    expect(
      screen
        .getByTestId(incidentControlsTriggerTestId())
        .getAttribute("data-active-section"),
    ).toBe("incident-fields");
    fireEvent.click(screen.getByTestId(incidentControlsTriggerTestId()));
    fireEvent.click(
      await screen.findByTestId(incidentControlsMenuItemTestId("memberships")),
    );
    expect(
      await screen.findByTestId(incidentMembershipCreateButtonTestId()),
    ).toBeTruthy();
    fireEvent.click(screen.getByTestId(incidentControlsCloseButtonTestId()));
    await waitFor(() => {
      expect(screen.queryByTestId(incidentControlsPanelTestId())).toBeNull();
    });
  });

  it("lazy-loads the Workbook Import Assistant only for the claimed Import profile", async () => {
    render(
      <WorkbookShell
        accountApplicationMenu={(props) => (
          <TestAccountApplicationMenu {...props} />
        )}
        extensionProfiles={[
          {
            profile_id: "import",
            claimable: true,
            claimed: true,
            contract_major: 1,
            route_families: ["/api/v1/import-sessions"],
            workspace_keys: [],
            capabilities: [],
          },
        ]}
        incidentId="10000000-0000-4000-8000-000000000001"
        renderIncidentControls={(props) => <TestIncidentControls {...props} />}
      />,
    );

    fireEvent.click(screen.getByTestId(testAccountMenuTriggerTestId));
    fireEvent.click(screen.getByTestId(incidentControlsTriggerTestId()));
    const importItem = await screen.findByTestId(
      incidentControlsMenuItemTestId("import-assistant"),
    );
    fireEvent.click(importItem);
    expect(
      await screen.findByTestId(workbookImportAssistantTestId()),
    ).toBeTruthy();
    expect(screen.getByLabelText("Source workbook")).toBeTruthy();
  });

  it("selects required built-in and system view surfaces by view_schema_id", async () => {
    scenario.genericRowsByView[indicatorsViewSchemaId] = [
      indicatorRow(
        "24000000-0000-4000-8000-000000000001",
        1,
        "ipv4_addr",
        "203.0.113.42",
      ),
    ];

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    const workbookShell = screen.getByTestId(workbookShellReadyTestId());
    expect(workbookShell.style.display).toBe("grid");
    expect(workbookShell.style.gridTemplateRows).toBe("auto minmax(0, 1fr)");
    expect(workbookShell.style.blockSize).toBe("100%");
    expect(["0", "0px"]).toContain(workbookShell.style.minBlockSize);
    expect(workbookShell.style.overflow).toBe("hidden");
    const shellContentRegion = workbookShell.children[1];
    expect(shellContentRegion).toBeInstanceOf(HTMLElement);
    if (!(shellContentRegion instanceof HTMLElement)) {
      throw new Error("Expected workbook shell content region to exist");
    }
    expect(shellContentRegion.style.blockSize).toBe("100%");
    const shellActiveSurface = shellContentRegion.children[0];
    expect(shellActiveSurface).toBeInstanceOf(HTMLElement);
    expect((shellActiveSurface as HTMLElement).style.gridTemplateRows).toBe(
      "minmax(0, 1fr)",
    );
    expect((shellActiveSurface as HTMLElement).style.blockSize).toBe("100%");

    const topBar = document.querySelector(
      dataTestIdSelector(workbookShellSlotTestId("top-bar")),
    );
    expect(topBar).toBeInstanceOf(HTMLElement);
    expect(topBar?.textContent?.match(/Timeline/g) ?? []).toHaveLength(1);
    expect(
      topBar?.querySelector("[data-workbook-query-surface-title='true']"),
    ).toBeNull();
    expect(
      topBar?.querySelector(
        dataTestIdSelector(gridGroupingSelectTestId(timelineViewSchemaId)),
      ),
    ).toBeNull();
    expect(
      topBar?.querySelector(
        dataTestIdSelector(
          workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
        ),
      ),
    ).toBeNull();
    const viewBarQueryControls = screen.getByTestId(
      workbookViewBarQueryControlsTestId(timelineViewSchemaId),
    );
    expect(viewBarQueryControls.style.overflow).toBe("visible");

    const viewBar = await screen.findByTestId(
      workbookShellSlotTestId("view-bar"),
    );
    expect(viewBar).toBeInstanceOf(HTMLElement);
    expect(viewBar.contains(viewBarQueryControls)).toBe(true);
    expect(
      viewBar.querySelector(
        dataTestIdSelector(gridGroupingSelectTestId(timelineViewSchemaId)),
      ),
    ).toBeInstanceOf(HTMLElement);
    expect(
      viewBar.querySelector(
        dataTestIdSelector(
          workbookFilterPopoverTriggerTestId(timelineViewSchemaId),
        ),
      ),
    ).toBeInstanceOf(HTMLElement);
    expect(
      viewBar.querySelector(
        dataTestIdSelector(gridFilterFieldTestId(timelineViewSchemaId)),
      ),
    ).toBeNull();
    expect(
      viewBar.querySelector(
        dataTestIdSelector(savedViewSelectorTestId(timelineViewSchemaId)),
      ),
    ).toBeInstanceOf(HTMLElement);
    expect(
      viewBar.querySelector(
        dataTestIdSelector(
          savedViewActionMenuTriggerTestId(timelineViewSchemaId),
        ),
      ),
    ).toBeInstanceOf(HTMLElement);
    expect(viewBar.style.overflow).toBe("visible");
    fireEvent.click(
      screen.getByTestId(
        savedViewActionMenuTriggerTestId(timelineViewSchemaId),
      ),
    );
    const savedViewActionMenu = screen.getByTestId(
      savedViewActionMenuTestId(timelineViewSchemaId),
    );
    expect(savedViewActionMenu).toBeInstanceOf(HTMLElement);
    const savedViewControlGroup =
      savedViewActionMenu.parentElement?.parentElement;
    expect(savedViewControlGroup).toBeInstanceOf(HTMLElement);
    expect((savedViewControlGroup as HTMLElement).style.overflow).toBe(
      "visible",
    );
    const savedViewToolbarLeftRail = savedViewControlGroup?.parentElement;
    expect(savedViewToolbarLeftRail).toBeInstanceOf(HTMLElement);
    expect((savedViewToolbarLeftRail as HTMLElement).style.overflow).toBe(
      "visible",
    );
    const savedViewToolbar = savedViewToolbarLeftRail?.parentElement;
    expect(savedViewToolbar).toBeInstanceOf(HTMLElement);
    expect((savedViewToolbar as HTMLElement).style.overflow).toBe("visible");
    expect(
      viewBar.querySelector(
        dataTestIdSelector(workbookInspectorToggleTestId(timelineViewSchemaId)),
      ),
    ).toBeInstanceOf(HTMLElement);
    expect(
      viewBar.querySelector(
        dataTestIdSelector(workbookAddRowButtonTestId(timelineViewSchemaId)),
      ),
    ).toBeInstanceOf(HTMLElement);
    const inspectorButton = screen.getByTestId(
      workbookInspectorToggleTestId(timelineViewSchemaId),
    );
    const addRowButton = screen.getByTestId(
      workbookAddRowButtonTestId(timelineViewSchemaId),
    );
    expect(inspectorButton.compareDocumentPosition(addRowButton)).toBe(
      Node.DOCUMENT_POSITION_FOLLOWING,
    );

    const builtInTabIds = screen
      .getAllByRole("button")
      .map((button) => button.getAttribute("data-testid") ?? "")
      .filter((testId) => testId.startsWith("surface-tab-"));
    expect(builtInTabIds).toEqual(
      requiredBuiltInWorkbookSurfaceIds.map((viewSchemaId) =>
        surfaceTabTestId(viewSchemaId),
      ),
    );

    fireEvent.click(screen.getByTestId(systemViewSwitcherTriggerTestId()));
    expect(
      Array.from(
        screen
          .getByTestId(systemViewSwitcherGroupTestId("scope-indicators"))
          .querySelectorAll("[data-view-schema-id]"),
      ).map((option) => option.getAttribute("data-view-schema-id")),
    ).toEqual([
      "cartulary.view.indicators.v1",
      "cartulary.view.assessments.v1",
    ]);
    expect(
      Array.from(
        screen
          .getByTestId(systemViewSwitcherGroupTestId("coordination"))
          .querySelectorAll("[data-view-schema-id]"),
      ).map((option) => option.getAttribute("data-view-schema-id")),
    ).toEqual([
      "cartulary.view.task_requests.v1",
      "cartulary.view.decisions.v1",
      "cartulary.view.parties.v1",
      "cartulary.view.comm_log.v1",
      "cartulary.view.handoff.v1",
    ]);
    const systemViewOptions = [
      ...Array.from(
        screen
          .getByTestId(systemViewSwitcherGroupTestId("scope-indicators"))
          .querySelectorAll("[data-view-schema-id]"),
      ),
      ...Array.from(
        screen
          .getByTestId(systemViewSwitcherGroupTestId("coordination"))
          .querySelectorAll("[data-view-schema-id]"),
      ),
      ...Array.from(
        screen
          .getByTestId(systemViewSwitcherGroupTestId("review-learning"))
          .querySelectorAll("[data-view-schema-id]"),
      ),
      ...Array.from(
        screen
          .getByTestId(
            systemViewSwitcherGroupTestId("optional-artifact-surfaces"),
          )
          .querySelectorAll("[data-view-schema-id]"),
      ),
    ].map((option) => option.getAttribute("data-view-schema-id"));
    expect(systemViewOptions).toEqual([
      "cartulary.view.indicators.v1",
      "cartulary.view.assessments.v1",
      "cartulary.view.task_requests.v1",
      "cartulary.view.decisions.v1",
      "cartulary.view.parties.v1",
      "cartulary.view.comm_log.v1",
      "cartulary.view.handoff.v1",
      "cartulary.view.status_review.v1",
      "cartulary.view.lesson.v1",
      ...optionalStandardizedWorkbookSurfaceIds,
    ]);
    expect(
      screen
        .getByTestId(
          systemViewSwitcherOptionTestId(
            "scope-indicators",
            indicatorsViewSchemaId,
          ),
        )
        .getAttribute("data-view-schema-id"),
    ).toBe(indicatorsViewSchemaId);
    fireEvent.click(screen.getByTestId(systemViewSwitcherTriggerTestId()));

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${evidenceViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(
      screen.getByTestId(gridShellTestId(evidenceViewSchemaId)),
    ).toBeTruthy();
    expect(topBar?.textContent?.match(/Evidence/g) ?? []).toHaveLength(1);
    expect(
      topBar?.querySelector("[data-workbook-query-surface-title='true']"),
    ).toBeNull();
    expect(
      topBar?.querySelector(
        dataTestIdSelector(gridFilterFieldTestId(evidenceViewSchemaId)),
      ),
    ).toBeNull();

    fireEvent.click(screen.getByTestId(systemViewSwitcherTriggerTestId()));
    fireEvent.click(
      screen.getByTestId(
        systemViewSwitcherOptionTestId(
          "scope-indicators",
          indicatorsViewSchemaId,
        ),
      ),
    );

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${indicatorsViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(window.location.search).toContain(
      `view_schema_id=${encodeURIComponent(indicatorsViewSchemaId)}`,
    );
    await waitFor(() => {
      expect(document.activeElement).toBe(
        screen.getByTestId(
          genericCreateFieldTestId("indicator.indicator_type"),
        ),
      );
    });
    expect(
      topBar?.querySelector("[data-workbook-query-surface-title='true']"),
    ).toBeNull();
    expect(
      screen.getByTestId(systemViewSwitcherTriggerTestId()).textContent,
    ).toBe("System views");
    expect(topBar?.textContent).toContain("Indicators");

    fireEvent.click(screen.getByTestId(systemViewSwitcherTriggerTestId()));
    const commLogOption = screen.getByTestId(
      systemViewSwitcherOptionTestId("coordination", commLogViewSchemaId),
    );
    fireEvent.mouseDown(commLogOption);
    fireEvent.click(commLogOption);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${commLogViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(window.location.search).toContain(
      `view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`,
    );
    expect(
      screen.getByTestId(gridShellTestId(commLogViewSchemaId)),
    ).toBeTruthy();
  });

  it("uses backend startup selection for the initial workbook grid surface", async () => {
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: evidenceViewSchemaId },
      selected_view_schema_id: evidenceViewSchemaId,
      selected_saved_view: null,
      source: "default",
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/workbook-startup"),
        expect.objectContaining({ credentials: "include" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${evidenceViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(window.location.search).toContain(
      `view_schema_id=${encodeURIComponent(evidenceViewSchemaId)}`,
    );
  });

  it("keeps a user-selected system view when a delayed startup response resolves", async () => {
    const delayedStartup = deferred<Response>();
    scenario.startupResponseOverride = () => delayedStartup.promise;

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/workbook-startup"),
        expect.objectContaining({ credentials: "include" }),
      );
    });

    fireEvent.click(screen.getByTestId(systemViewSwitcherTriggerTestId()));
    fireEvent.click(
      screen.getByTestId(
        systemViewSwitcherOptionTestId("coordination", commLogViewSchemaId),
      ),
    );

    await waitFor(() => {
      expect(window.location.search).toContain(
        `view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`,
      );
    });

    delayedStartup.resolve(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        extension_workspace_availability: {
          schema_id: "cartulary.extension_workspace_availability.v1",
          incident_id: "10000000-0000-4000-8000-000000000001",
          workspaces: [],
        },
        selected_sheet_ref: { kind: "view_schema", id: evidenceViewSchemaId },
        selected_view_schema_id: evidenceViewSchemaId,
        selected_saved_view: null,
        source: "default",
        cleared_pointers: [],
        home_sheet_ref: null,
        default_sheet_ref: null,
      }),
    );

    await waitFor(() => {
      expect(window.location.search).toContain(
        `view_schema_id=${encodeURIComponent(commLogViewSchemaId)}`,
      );
      expect(
        screen.getByTestId(gridShellTestId(commLogViewSchemaId)),
      ).toBeTruthy();
    });
    expect(window.location.search).not.toContain(
      `view_schema_id=${encodeURIComponent(evidenceViewSchemaId)}`,
    );
  });

  it("loads a direct Status Review URL with the active surface query result", async () => {
    window.history.replaceState(
      {},
      "",
      `/?view_schema_id=${encodeURIComponent(statusReviewViewSchemaId)}`,
    );
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: statusReviewViewSchemaId },
      selected_view_schema_id: statusReviewViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.genericRowsByView[statusReviewViewSchemaId] = [
      statusReviewRow(
        "00000000-0000-4000-8000-000000000501",
        1,
        "Direct Status Review surface load",
      ),
    ];
    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${statusReviewViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    expect(
      (
        await screen.findByTestId(
          rowCellTestId(
            "00000000-0000-4000-8000-000000000501",
            "status_review.current_state_summary",
          ),
        )
      ).textContent,
    ).toBe("Direct Status Review surface load");
  });

  it("rejects generic query envelopes from a different view_schema_id", async () => {
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: statusReviewViewSchemaId },
      selected_view_schema_id: statusReviewViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.queryResponseOverride = (viewSchemaId) => {
      if (viewSchemaId !== statusReviewViewSchemaId) {
        return null;
      }
      return successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: evidenceViewSchemaId,
        rows: [
          statusReviewRow(
            "00000000-0000-4000-8000-000000000501",
            1,
            "Mismatched Status Review result",
          ),
        ],
      });
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    expect(await screen.findByText("Workbook view load failed.")).toBeTruthy();
    expect(
      screen.queryByTestId(
        rowCellTestId(
          "00000000-0000-4000-8000-000000000501",
          "status_review.current_state_summary",
        ),
      ),
    ).toBeNull();
  });

  it("keeps a selected saved-view sheet ref distinct from its base view_schema", async () => {
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "saved_view", id: savedViewId },
      selected_view_schema_id: evidenceViewSchemaId,
      selected_saved_view: testSavedViewResource({
        saved_view_id: savedViewId,
        view_schema_id: evidenceViewSchemaId,
      }),
      source: "home",
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${evidenceViewSchemaId}/query`),
        expect.objectContaining({ method: "POST" }),
      );
      expect(window.location.search).toContain("sheet_ref_kind=saved_view");
    });
    expect(window.location.search).toContain(`sheet_ref_id=${savedViewId}`);
    expect(window.location.search).not.toContain("view_schema_id=");
  });

  it("renders saved views only in the active surface selector and preserves selected saved-view identity", async () => {
    const evidenceSavedViewId = "22222222-2222-4222-8222-222222222222";
    scenario.savedViews = [
      testSavedViewResource({
        saved_view_id: savedViewId,
        view_schema_id: timelineViewSchemaId,
        display_name: "Timeline saved view",
      }),
      testSavedViewResource({
        saved_view_id: evidenceSavedViewId,
        view_schema_id: evidenceViewSchemaId,
        display_name: "Evidence saved view",
      }),
    ];

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    const timelineSelector = await screen.findByTestId(
      savedViewSelectorTestId(timelineViewSchemaId),
    );
    expect(
      timelineSelector.querySelector(
        dataTestIdSelector(
          savedViewOptionTestId(timelineViewSchemaId, savedViewId),
        ),
      ),
    ).not.toBeNull();
    expect(
      timelineSelector.querySelector(
        dataTestIdSelector(
          savedViewOptionTestId(timelineViewSchemaId, evidenceSavedViewId),
        ),
      ),
    ).toBeNull();
    const timelineQueryCallCount = () =>
      fetchMock.mock.calls.filter(
        ([input, init]) =>
          String(input).includes(`/views/${timelineViewSchemaId}/query`) &&
          ((init as RequestInit | undefined)?.method ?? "GET") === "POST",
      ).length;
    await waitFor(() => {
      expect(timelineQueryCallCount()).toBeGreaterThan(0);
    });
    const timelineQueryCountBeforeSavedViewSelect = timelineQueryCallCount();

    fireEvent.change(timelineSelector, {
      target: { value: savedViewId },
    });

    await waitFor(() => {
      expect(window.location.search).toContain("sheet_ref_kind=saved_view");
      expect(timelineQueryCallCount()).toBeGreaterThan(
        timelineQueryCountBeforeSavedViewSelect,
      );
    });
    expect(window.location.search).toContain(`sheet_ref_id=${savedViewId}`);
    expect(window.location.search).not.toContain("view_schema_id=");
    expect(timelineSelector.getAttribute("data-selected-sheet-ref-kind")).toBe(
      "saved_view",
    );
    expect(timelineSelector.getAttribute("data-selected-saved-view-id")).toBe(
      savedViewId,
    );

    fireEvent.click(screen.getByTestId(surfaceTabTestId(evidenceViewSchemaId)));
    const evidenceSelector = await screen.findByTestId(
      savedViewSelectorTestId(evidenceViewSchemaId),
    );
    expect(
      evidenceSelector.querySelector(
        dataTestIdSelector(
          savedViewOptionTestId(evidenceViewSchemaId, evidenceSavedViewId),
        ),
      ),
    ).not.toBeNull();
    expect(
      evidenceSelector.querySelector(
        dataTestIdSelector(
          savedViewOptionTestId(evidenceViewSchemaId, savedViewId),
        ),
      ),
    ).toBeNull();
    expect(window.location.search).toContain(
      `view_schema_id=${encodeURIComponent(evidenceViewSchemaId)}`,
    );
  });

  it("Verify saved-view create/update/select/default UI uses active surface scope and public saved-view/workbook-preference contracts.", async () => {
    const systemSavedViewId = "22222222-2222-4222-8222-222222222222";
    scenario.timelineRows = [
      timelineRow("21000000-0000-4000-8000-000000000001", 1, "Selected row", 0),
    ];
    scenario.savedViews = [
      testSavedViewResource({
        saved_view_id: savedViewId,
        view_schema_id: timelineViewSchemaId,
        display_name: "Timeline saved view",
        scope: "private",
        query_json: {
          filters: [
            {
              field_key: "timeline.capture_state",
              op: "eq",
              arg: { value: "reviewed" },
            },
          ],
          group_by: "timeline.capture_state",
          sort: [{ field_key: "timeline.activity_sort_ts", direction: "desc" }],
        },
        layout_json: {
          layout_schema_id: "cartulary.layout.v1",
          column_order: [
            "timeline.activity_synopsis_text",
            "timeline.activity_utc_text",
          ],
          hidden_field_keys: ["timeline.raw_activity_text"],
          column_widths: [
            { field_key: "timeline.activity_synopsis_text", width_px: 360 },
          ],
        },
      }),
      testSavedViewResource({
        saved_view_id: systemSavedViewId,
        view_schema_id: timelineViewSchemaId,
        display_name: "System timeline",
        owner_user_id: null,
        scope: "system",
        query_json: {
          filters: [
            {
              field_key: "timeline.tags",
              op: "contains_any",
              arg: { values: ["alpha", "zeta"] },
            },
          ],
          sort: [],
        },
      }),
    ];

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    const selector = await screen.findByTestId(
      savedViewSelectorTestId(timelineViewSchemaId),
    );
    const timelineQueryCallCount = () =>
      fetchMock.mock.calls.filter(
        ([input, init]) =>
          String(input).includes(`/views/${timelineViewSchemaId}/query`) &&
          ((init as RequestInit | undefined)?.method ?? "GET") === "POST",
      ).length;
    const savedViewMutationBodies = (method: string) =>
      fetchMock.mock.calls
        .filter(
          ([input, init]) =>
            String(input).includes(
              "/api/v1/incidents/10000000-0000-4000-8000-000000000001/saved-views",
            ) &&
            (
              (init as RequestInit | undefined)?.method ?? "GET"
            ).toUpperCase() === method,
        )
        .map(([, init]) =>
          JSON.parse(String((init as RequestInit | undefined)?.body ?? "{}")),
        ) as Array<Record<string, unknown>>;
    await waitFor(() => {
      expect(timelineQueryCallCount()).toBeGreaterThan(0);
    });
    const queryCountBeforeSelect = timelineQueryCallCount();

    fireEvent.change(selector, { target: { value: savedViewId } });

    await waitFor(() => {
      expect(window.location.search).toContain("sheet_ref_kind=saved_view");
      expect(timelineQueryCallCount()).toBeGreaterThan(queryCountBeforeSelect);
    });
    const savedViewQueryBody = JSON.parse(
      String(
        (
          fetchMock.mock.calls
            .filter(
              ([input, init]) =>
                String(input).includes(
                  `/views/${timelineViewSchemaId}/query`,
                ) &&
                ((init as RequestInit | undefined)?.method ?? "GET") === "POST",
            )
            .at(-1)?.[1] as RequestInit | undefined
        )?.body ?? "{}",
      ),
    );
    expect(savedViewQueryBody).toEqual({
      filters: [
        {
          arg: { value: "reviewed" },
          field_key: "timeline.capture_state",
          op: "eq",
        },
      ],
      group_by: "timeline.capture_state",
      sort: [
        { direction: "asc", field_key: "timeline.capture_state" },
        { direction: "desc", field_key: "timeline.activity_sort_ts" },
      ],
    });
    expect(window.location.search).toContain(`sheet_ref_id=${savedViewId}`);
    expect(window.location.search).not.toContain("view_schema_id=");

    openSavedViewActions(timelineViewSchemaId);
    fireEvent.change(
      screen.getByTestId(savedViewNameInputTestId(timelineViewSchemaId)),
      { target: { value: "Updated shared view" } },
    );
    fireEvent.change(
      screen.getByTestId(savedViewScopeSelectTestId(timelineViewSchemaId)),
      { target: { value: "shared" } },
    );
    fireEvent.click(
      await screen.findByTestId(
        savedViewUpdateButtonTestId(timelineViewSchemaId, savedViewId),
      ),
    );

    await waitFor(() => {
      expect(savedViewMutationBodies("PATCH")).toHaveLength(1);
    });
    const patchBody = savedViewMutationBodies("PATCH")[0] ?? {};
    expect(Object.keys(patchBody).sort()).toEqual([
      "base_saved_view_version",
      "display_name",
      "layout_json",
      "query_json",
      "scope",
    ]);
    expect(patchBody).toMatchObject({
      base_saved_view_version: 1,
      display_name: "Updated shared view",
      scope: "shared",
      query_json: {
        filters: [
          {
            arg: { value: "reviewed" },
            field_key: "timeline.capture_state",
            op: "eq",
          },
        ],
        group_by: "timeline.capture_state",
        sort: [{ direction: "desc", field_key: "timeline.activity_sort_ts" }],
      },
      layout_json: {
        layout_schema_id: "cartulary.layout.v1",
      },
    });
    expect(patchBody).not.toHaveProperty("view_schema_id");
    expect(patchBody).not.toHaveProperty("saved_view_id");
    expect(patchBody).not.toHaveProperty("owner_user_id");

    openSavedViewActions(timelineViewSchemaId);
    const createButton = screen.getByTestId(
      savedViewCreateButtonTestId(timelineViewSchemaId),
    );
    const homeButton = screen.getByTestId(
      savedViewSetHomeButtonTestId(timelineViewSchemaId),
    );
    const defaultButton = screen.getByTestId(
      savedViewSetDefaultButtonTestId(timelineViewSchemaId),
    );
    expect(createButton.parentElement).not.toBe(homeButton.parentElement);
    expect(defaultButton.parentElement).toBe(homeButton.parentElement);

    fireEvent.click(homeButton);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/workbook-preferences/me"),
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify({
            home_sheet_ref: { kind: "saved_view", id: savedViewId },
          }),
        }),
      );
    });
    openSavedViewActions(timelineViewSchemaId);
    fireEvent.click(
      screen.getByTestId(savedViewSetDefaultButtonTestId(timelineViewSchemaId)),
    );
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/workbook-preferences/default"),
        expect.objectContaining({
          method: "PUT",
          body: JSON.stringify({
            default_sheet_ref: { kind: "saved_view", id: savedViewId },
          }),
        }),
      );
    });

    fireEvent.change(selector, { target: { value: systemSavedViewId } });
    openSavedViewActions(timelineViewSchemaId);
    const systemUpdateButton = await screen.findByTestId(
      savedViewUpdateButtonTestId(timelineViewSchemaId, systemSavedViewId),
    );
    const systemDeleteButton = await screen.findByTestId(
      savedViewDeleteButtonTestId(timelineViewSchemaId, systemSavedViewId),
    );
    expect((systemUpdateButton as HTMLButtonElement).disabled).toBe(true);
    expect((systemDeleteButton as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(
      screen.getByTestId(
        savedViewDuplicateButtonTestId(timelineViewSchemaId, systemSavedViewId),
      ),
    );

    await waitFor(() => {
      expect(savedViewMutationBodies("POST")).toHaveLength(1);
    });
    expect(savedViewMutationBodies("POST")[0]).toMatchObject({
      display_name: "System timeline Copy",
      scope: "private",
      view_schema_id: timelineViewSchemaId,
      query_json: {
        filters: [
          {
            arg: { values: ["alpha", "zeta"] },
            field_key: "timeline.tags",
            op: "contains_any",
          },
        ],
        sort: [],
      },
    });

    openSavedViewActions(timelineViewSchemaId);
    const copyDeleteButton = await screen.findByTestId(
      savedViewDeleteButtonTestId(timelineViewSchemaId, savedViewCopyId),
    );
    fireEvent.click(copyDeleteButton);
    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/saved-views/${savedViewCopyId}`),
        expect.objectContaining({ method: "DELETE" }),
      );
      expect(window.location.search).toContain(
        `view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
      );
    });
    expect(
      screen.getByTestId(
        rowCellTestId(
          "21000000-0000-4000-8000-000000000001",
          "timeline.activity_synopsis_text",
        ),
      ),
    ).not.toBeNull();

    openSavedViewActions(timelineViewSchemaId);
    fireEvent.change(
      screen.getByTestId(savedViewNameInputTestId(timelineViewSchemaId)),
      { target: { value: "Created from current state" } },
    );
    fireEvent.click(
      screen.getByTestId(savedViewCreateButtonTestId(timelineViewSchemaId)),
    );
    await waitFor(() => {
      expect(savedViewMutationBodies("POST")).toHaveLength(2);
    });
    expect(savedViewMutationBodies("POST")[1]).toMatchObject({
      display_name: "Created from current state",
      scope: "private",
      view_schema_id: timelineViewSchemaId,
    });
  });

  it("passes invalid explicit base surfaces to backend startup fallback", async () => {
    window.history.replaceState(
      {},
      "",
      "/?view_schema_id=cartulary.view.unknown.v1",
    );
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
      selected_view_schema_id: timelineViewSchemaId,
      selected_saved_view: null,
      source: "timeline",
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(
          `/workbook-startup?sheet_ref_id=${encodeURIComponent(
            "cartulary.view.unknown.v1",
          )}&sheet_ref_kind=view_schema`,
        ),
        expect.objectContaining({ credentials: "include" }),
      );
    });
    expect(window.location.search).toContain(
      `view_schema_id=${encodeURIComponent(timelineViewSchemaId)}`,
    );
  });

  it("ignores superseded generic surface query responses after rapid filters", async () => {
    const staleEvidenceQuery = deferred<Response>();
    let staleEvidenceQueryStarted = false;
    scenario.evidenceRows = [
      evidenceRow("00000000-0000-4000-8000-000000000601", 1, "initial"),
    ];
    scenario.queryResponseOverride = (viewSchemaId, init) => {
      if (viewSchemaId !== evidenceViewSchemaId) {
        return null;
      }
      const value = stringFilterValue(parseRequestBody(init));
      if (value === "older") {
        staleEvidenceQueryStarted = true;
        return staleEvidenceQuery.promise;
      }
      if (value === "newer") {
        return successEnvelope({
          incident_id: "10000000-0000-4000-8000-000000000001",
          view_schema_id: evidenceViewSchemaId,
          rows: [
            evidenceRow("00000000-0000-4000-8000-000000000602", 1, "newer"),
          ],
        });
      }
      return null;
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );
    await expectRecordIds(evidenceViewSchemaId, [
      "00000000-0000-4000-8000-000000000601",
    ]);

    applyGenericFilter(evidenceViewSchemaId, "evidence.storage_ref", "older");
    await waitFor(() => {
      expect(staleEvidenceQueryStarted).toBe(true);
    });
    applyGenericFilter(evidenceViewSchemaId, "evidence.storage_ref", "newer");
    await expectRecordIds(evidenceViewSchemaId, [
      "00000000-0000-4000-8000-000000000602",
    ]);

    staleEvidenceQuery.resolve(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: evidenceViewSchemaId,
        rows: [evidenceRow("00000000-0000-4000-8000-000000000603", 1, "older")],
      }),
    );
    await flushMicrotasks();

    expect(currentRecordIds(evidenceViewSchemaId)).toEqual([
      "00000000-0000-4000-8000-000000000602",
    ]);
  });

  it("retains accepted generic rows and marks them stale when refresh fails", async () => {
    scenario.evidenceRows = [
      evidenceRow("00000000-0000-4000-8000-000000000605", 3, "retained"),
    ];
    scenario.queryResponseOverride = (viewSchemaId, init) => {
      if (viewSchemaId !== evidenceViewSchemaId) return null;
      return stringFilterValue(parseRequestBody(init)) === "failure"
        ? errorEnvelope("projection_failed", 500)
        : null;
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);
    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );
    await expectRecordIds(evidenceViewSchemaId, [
      "00000000-0000-4000-8000-000000000605",
    ]);

    applyGenericFilter(evidenceViewSchemaId, "evidence.storage_ref", "failure");

    expect(
      await screen.findByText(
        "projection_failed Previously loaded rows may be stale.",
      ),
    ).toBeTruthy();
    expect(currentRecordIds(evidenceViewSchemaId)).toEqual([
      "00000000-0000-4000-8000-000000000605",
    ]);
  });

  it("keeps generic surface access loss routed through the shell access-lost callback", async () => {
    const onIncidentAccessLost = vi.fn();
    scenario.evidenceRows = [
      evidenceRow("00000000-0000-4000-8000-000000000604", 2, "protected"),
    ];
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: evidenceViewSchemaId },
      selected_view_schema_id: evidenceViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.queryResponseOverride = (viewSchemaId, init) =>
      viewSchemaId === evidenceViewSchemaId &&
      stringFilterValue(parseRequestBody(init)) === "denied"
        ? errorEnvelope("authorization_denied", 403)
        : null;

    render(
      <WorkbookShell
        incidentId="10000000-0000-4000-8000-000000000001"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );

    await expectRecordIds(evidenceViewSchemaId, [
      "00000000-0000-4000-8000-000000000604",
    ]);
    applyGenericFilter(evidenceViewSchemaId, "evidence.storage_ref", "denied");

    await waitFor(() => {
      expect(onIncidentAccessLost).toHaveBeenCalledTimes(1);
    });
    expect(await screen.findByText("authorization_denied")).toBeTruthy();
    expect(currentRecordIds(evidenceViewSchemaId)).toEqual([]);
  });

  it("keeps entity surface access loss routed through the shell access-lost callback", async () => {
    const onIncidentAccessLost = vi.fn();
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: hostsViewSchemaId },
      selected_view_schema_id: hostsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.queryResponseOverride = (viewSchemaId) =>
      viewSchemaId === hostsViewSchemaId
        ? errorEnvelope("authorization_denied", 403)
        : null;

    render(
      <WorkbookShell
        incidentId="10000000-0000-4000-8000-000000000001"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );

    await waitFor(() => {
      expect(onIncidentAccessLost).toHaveBeenCalledTimes(1);
    });
    expect(await screen.findByText("authorization_denied")).toBeTruthy();
  });

  it("keeps assessment surface access loss routed through the shell access-lost callback", async () => {
    const onIncidentAccessLost = vi.fn();
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: assessmentsViewSchemaId },
      selected_view_schema_id: assessmentsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.queryResponseOverride = (viewSchemaId) =>
      viewSchemaId === assessmentsViewSchemaId
        ? errorEnvelope("authorization_denied", 403)
        : null;

    render(
      <WorkbookShell
        incidentId="10000000-0000-4000-8000-000000000001"
        onIncidentAccessLost={onIncidentAccessLost}
      />,
    );

    await waitFor(() => {
      expect(onIncidentAccessLost).toHaveBeenCalledTimes(1);
    });
    expect(await screen.findByText("authorization_denied")).toBeTruthy();
  });

  it("dispatches entity-origin paste through create targets while preserving exact-match reuse results", async () => {
    const existingRecordId = "00000000-0000-4000-8000-000000006100";
    const createdRecordId = "00000000-0000-4000-8000-000000006101";
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: hostsViewSchemaId },
      selected_view_schema_id: hostsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.genericRowsByView[hostsViewSchemaId] = [
      hostRow({
        displayName: "Reusable host",
        hostname: "reuse.example.test",
        recordId: existingRecordId,
        rowVersion: 4,
      }),
    ];

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    const displayNameCell = await screen.findByTestId(
      rowCellTestId(existingRecordId, "host.display_name"),
    );
    const displayNameGridCell = displayNameCell.closest('[role="gridcell"]');
    expect(displayNameGridCell).toBeTruthy();
    fireEvent.mouseDown(displayNameGridCell as HTMLElement);
    const pasteEvent = createEvent.paste(displayNameCell, {
      clipboardData: {
        getData: () =>
          [
            "Pasted host reuse\treuse.example.test",
            "Pasted host create\tcreate.example.test",
          ].join("\n"),
      },
    });
    fireEvent(displayNameCell, pasteEvent);
    expect(pasteEvent.defaultPrevented).toBe(true);

    await waitFor(() => {
      expect(
        fetchMock.mock.calls.some(
          ([input, init]) =>
            String(input).endsWith(
              `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${hostsViewSchemaId}/clipboard-paste`,
            ) &&
            ((init as RequestInit | undefined)?.method ?? "GET") === "POST",
        ),
      ).toBe(true);
    });
    const pasteCall = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith(
        `/api/v1/incidents/10000000-0000-4000-8000-000000000001/views/${hostsViewSchemaId}/clipboard-paste`,
      ),
    );
    expect(pasteCall).toBeDefined();
    const pasteBody = JSON.parse(
      String((pasteCall?.[1] as RequestInit | undefined)?.body ?? "{}"),
    );
    expect(pasteBody).toMatchObject({
      view_schema_id: hostsViewSchemaId,
      clipboard_text:
        "Pasted host reuse\treuse.example.test\nPasted host create\tcreate.example.test",
      format: "tsv",
      start_field_key: "host.display_name",
      columns: ["host.display_name", "host.hostname"],
      targets: [{ kind: "create" }, { kind: "create" }],
    });
    expect(window.location.href).not.toContain("/imports");
    await expectRecordIds(hostsViewSchemaId, [
      existingRecordId,
      createdRecordId,
    ]);
    expect(
      screen.getByTestId(rowCellTestId(existingRecordId, "host.display_name"))
        .textContent,
    ).toContain("Pasted host reuse");
    expect(
      screen.getByTestId(rowCellTestId(createdRecordId, "host.display_name"))
        .textContent,
    ).toContain("Pasted host create");
  });

  it("keeps entity merge review, confirmation, and dependent Timeline preview bound to stable record ids", async () => {
    const survivorId = "00000000-0000-4000-8000-000000006300";
    const loserId = "00000000-0000-4000-8000-000000006301";
    const unrelatedId = "00000000-0000-4000-8000-000000006302";
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: hostsViewSchemaId },
      selected_view_schema_id: hostsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.genericRowsByView[hostsViewSchemaId] = [
      hostRow({
        aliases: ["survivor-alias", "shared-alias"],
        displayName: "Survivor host",
        fqdn: "survivor.example.test",
        hostname: "SURVIVOR.example.test",
        linkedEventCount: 2,
        recordId: survivorId,
        reusableIdentifiers: [
          {
            identifierClass: "fqdn",
            itemRef: "entity_preserved_identifier:survivor-legacy-fqdn",
            rawValue: "legacy-survivor.example.test",
          },
        ],
        rowVersion: 7,
      }),
      hostRow({
        aliases: ["Shared-Alias", "loser-alias"],
        displayName: "Loser host",
        fqdn: "loser.example.test",
        hostname: "survivor.example.test",
        linkedEventCount: 1,
        recordId: loserId,
        reusableIdentifiers: [
          {
            identifierClass: "hostname",
            itemRef: "entity_preserved_identifier:loser-legacy-hostname",
            rawValue: "old-loser-host",
          },
        ],
        rowVersion: 3,
      }),
      hostRow({
        displayName: "Unrelated host",
        hostname: "unrelated.example.test",
        recordId: unrelatedId,
        rowVersion: 1,
      }),
    ];
    scenario.timelineRows = [
      timelineRow(
        "21000000-0000-4000-8000-000000000002",
        5,
        "Dependent row",
        0,
        {
          hostRefs: [
            timelineEntityRef(survivorId, "Survivor host", "host-ref-1"),
          ],
        },
      ),
    ];

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(entityInspectButtonTestId("host", survivorId)),
    );
    await screen.findByTestId(entityInspectorTestId("host"));
    expect(
      screen.getByTestId(
        entityReusableIdentifiersSectionTestId("host", survivorId),
      ).textContent,
    ).toContain("Reusable identifiers");
    expect(
      screen.getByTestId(
        entityReusableIdentifierItemTestId(
          "host",
          survivorId,
          "entity_preserved_identifier:survivor-legacy-fqdn",
        ),
      ).textContent,
    ).toBe("FQDN: legacy-survivor.example.test");
    expect(
      (
        await screen.findByTestId(
          timelinePreviewRowTestId("21000000-0000-4000-8000-000000000002"),
        )
      ).textContent,
    ).toContain("Dependent row");

    fireEvent.change(
      screen.getByTestId(entityMergeControlTestId("loser-record")),
      {
        target: { value: loserId },
      },
    );
    const mergePlan = await screen.findByTestId(
      entityMergeControlTestId("plan"),
    );
    expect(mergePlan.textContent).toContain(
      "Survivor Survivor host absorbs loser Loser host",
    );
    expect(mergePlan.textContent).toContain(
      "FQDN: Carry as reusable loser.example.test",
    );
    expect(mergePlan.textContent).toContain(
      "Hostname: Duplicate no-op survivor.example.test",
    );
    expect(mergePlan.textContent).toContain(
      "Hostname: Carry as reusable old-loser-host",
    );
    expect(mergePlan.textContent).toContain("Aliases to copy: loser-alias");
    expect(mergePlan.textContent).toContain(
      "Alias duplicate no-op: Shared-Alias",
    );
    expect(mergePlan.textContent).toContain(
      "Provenance-only values: Merge lineage and source provenance are retained server-side; no editable cell value is copied for them.",
    );
    expect(mergePlan.textContent).toContain(
      "Linked events visible on surface: survivor=2, loser=1.",
    );

    fireEvent.click(screen.getByTestId(entityMergeControlTestId("confirm")));

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/api/v1/records/${survivorId}/merge`),
        expect.objectContaining({ method: "POST" }),
      );
    });
    const mergeCall = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith(`/api/v1/records/${survivorId}/merge`),
    );
    expect(
      JSON.parse(String((mergeCall?.[1] as RequestInit | undefined)?.body)),
    ).toMatchObject({
      loser_record_id: loserId,
      survivor_base_row_version: 7,
      loser_base_row_version: 3,
      reason: "Merge duplicate entity",
    });
    await waitFor(() => {
      expect(
        screen.getByTestId(entityMergeControlTestId("message")).textContent,
      ).toContain("Merged Loser host into Survivor host (host).");
    });
    await expectRecordIds(hostsViewSchemaId, [survivorId, unrelatedId]);
    expect(
      screen.queryByTestId(rowCellTestId(loserId, "host.display_name")),
    ).toBeNull();

    fireEvent.click(
      screen.getByTestId(entityInspectButtonTestId("host", unrelatedId)),
    );
    await waitFor(() => {
      expect(
        screen.queryByTestId(
          timelinePreviewRowTestId("21000000-0000-4000-8000-000000000002"),
        ),
      ).toBeNull();
    });
  });

  it("mounts the matching entity inspector subject after deferred query hydration", async () => {
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: hostsViewSchemaId },
      selected_view_schema_id: hostsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    const host = hostRow({
      displayName: "Deferred host",
      hostname: "deferred.example.test",
      recordId: "00000000-0000-4000-8000-000000000701",
      rowVersion: 7,
    });
    scenario.genericRowsByView[hostsViewSchemaId] = [host];
    const deferredHostQuery = scenario.deferQuery(hostsViewSchemaId);
    const { container } = render(
      <WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />,
    );

    await waitFor(() => {
      expect(scenario.queryTrace).toContain(`requested:${hostsViewSchemaId}`);
    });
    deferredHostQuery.resolve(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: hostsViewSchemaId,
        rows: [host],
      }),
    );
    await waitForWorkbookRows({
      container,
      expectedRecordIds: ["00000000-0000-4000-8000-000000000701"],
      surface: hostsViewSchemaId,
    });

    fireEvent.click(
      screen.getByTestId(
        entityInspectButtonTestId(
          "host",
          "00000000-0000-4000-8000-000000000701",
        ),
      ),
    );
    await waitForEntityInspectorReady(container, {
      entityType: "host",
      recordId: "00000000-0000-4000-8000-000000000701",
      rowVersion: 7,
      viewSchemaId: hostsViewSchemaId,
    });
    await flushWorkbookAsync();
    await flushWorkbookAsync();

    expect(
      screen
        .getByTestId(entityInspectorTestId("host"))
        .getAttribute("data-record-id"),
    ).toBe("00000000-0000-4000-8000-000000000701");
    await waitForEntityInspectorReady(container, {
      entityType: "host",
      recordId: "00000000-0000-4000-8000-000000000701",
      rowVersion: 7,
      viewSchemaId: hostsViewSchemaId,
    });
  });

  it("submits physical alias add/remove actions and restores inspector focus after authoritative refresh", async () => {
    const recordId = "00000000-0000-4000-8000-000000006200";
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: hostsViewSchemaId },
      selected_view_schema_id: hostsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.genericRowsByView[hostsViewSchemaId] = [
      hostRow({
        aliases: ["Existing alias"],
        displayName: "Alias host",
        hostname: "alias-host",
        recordId,
        rowVersion: 1,
      }),
    ];
    const submittedActions: Array<Record<string, unknown>> = [];
    scenario.recordPatchResponseOverride = (patchedRecordId, init) => {
      if (patchedRecordId !== recordId) {
        return null;
      }
      const body = JSON.parse(String(init?.body ?? "{}")) as {
        changes?: Array<{
          action_payload?: { actions?: Array<Record<string, unknown>> };
        }>;
      };
      const action = body.changes?.[0]?.action_payload?.actions?.[0] ?? {};
      submittedActions.push(action);
      const aliases =
        action.op === "add_alias"
          ? ["Existing alias", String(action.alias_text)]
          : ["Added alias"];
      const row = hostRow({
        aliases,
        displayName: "Alias host",
        hostname: "alias-host",
        recordId: patchedRecordId,
        rowVersion: submittedActions.length + 1,
      });
      scenario.genericRowsByView[hostsViewSchemaId] = [row];
      return successEnvelope({
        view_schema_id: hostsViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row,
      });
    };

    const { container } = render(
      <WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await waitForWorkbookRows({
      container,
      expectedRecordIds: [recordId],
      surface: hostsViewSchemaId,
    });
    fireEvent.click(
      screen.getByTestId(entityInspectButtonTestId("host", recordId)),
    );
    await waitForEntityInspectorReady(container, {
      entityType: "host",
      recordId,
      rowVersion: 1,
      viewSchemaId: hostsViewSchemaId,
    });
    const aliasInput = screen.getByRole("textbox", { name: "Alias text" });
    fireEvent.change(aliasInput, { target: { value: "Added alias" } });
    fireEvent.click(screen.getByRole("button", { name: "Add alias" }));

    await waitFor(() => {
      expect(submittedActions).toHaveLength(1);
      expect(
        screen.getByRole("button", { name: "Remove alias Added alias" }),
      ).not.toBeNull();
      expect(document.activeElement).toBe(
        screen.getByRole("textbox", { name: "Alias text" }),
      );
    });
    expect(submittedActions[0]).toEqual({
      op: "add_alias",
      alias_text: "Added alias",
    });
    await waitForEntityInspectorReady(container, {
      entityType: "host",
      recordId,
      rowVersion: 2,
      viewSchemaId: hostsViewSchemaId,
    });

    fireEvent.click(
      screen.getByRole("button", { name: "Remove alias Existing alias" }),
    );
    await waitFor(() => {
      expect(submittedActions).toHaveLength(2);
      expect(
        screen.queryByRole("button", { name: "Remove alias Existing alias" }),
      ).toBeNull();
      expect(document.activeElement).toBe(
        screen.getByRole("textbox", { name: "Alias text" }),
      );
    });
    expect(submittedActions[1]).toEqual({
      op: "remove_alias",
      item_ref: "entity_alias:00000000-0000-0000-0000-000000000001",
    });
    await waitForEntityInspectorReady(container, {
      entityType: "host",
      recordId,
      rowVersion: 3,
      viewSchemaId: hostsViewSchemaId,
    });
  });

  it("surfaces entity merge precondition failures without submitting a second plan", async () => {
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: identitiesViewSchemaId },
      selected_view_schema_id: identitiesViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.genericRowsByView[identitiesViewSchemaId] = [
      identityRow({
        displayName: "Survivor identity",
        recordId: "00000000-0000-4000-8000-000000000803",
        reusableIdentifiers: [
          {
            identifierClass: "email",
            itemRef: "entity_preserved_identifier:survivor-legacy-email",
            rawValue: "legacy@example.test",
          },
        ],
        rowVersion: 11,
        upn: "survivor@example.test",
      }),
      identityRow({
        displayName: "Loser identity",
        email: "collision@example.test",
        recordId: "00000000-0000-4000-8000-000000000802",
        rowVersion: 2,
        upn: "loser@example.test",
      }),
    ];
    let mergeSubmissions = 0;
    scenario.mergeResponseOverride = () => {
      mergeSubmissions += 1;
      return new Response(
        JSON.stringify({
          error: {
            status: 409,
            code: "merge_precondition_failed",
            message: "merge_precondition_failed",
            request_id: "req-merge-precondition",
            retryable: false,
            details: {
              reason_code: "carry_forward_identifier_collision",
              identifier_class: "email",
              normalized_value: "collision@example.test",
              blocking_record_id: "00000000-0000-4000-8000-000000000801",
            },
          },
        }),
        { status: 409, headers: { "Content-Type": "application/json" } },
      );
    };

    const { container } = render(
      <WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />,
    );
    await waitForWorkbookRows({
      container,
      expectedRecordIds: [
        "00000000-0000-4000-8000-000000000803",
        "00000000-0000-4000-8000-000000000802",
      ],
      surface: identitiesViewSchemaId,
    });

    fireEvent.click(
      screen.getByTestId(
        entityInspectButtonTestId(
          "identity",
          "00000000-0000-4000-8000-000000000803",
        ),
      ),
    );
    await waitForEntityInspectorReady(container, {
      entityType: "identity",
      recordId: "00000000-0000-4000-8000-000000000803",
      rowVersion: 11,
      viewSchemaId: identitiesViewSchemaId,
    });
    expect(
      screen.getByTestId(
        entityReusableIdentifierItemTestId(
          "identity",
          "00000000-0000-4000-8000-000000000803",
          "entity_preserved_identifier:survivor-legacy-email",
        ),
      ).textContent,
    ).toBe("Email: legacy@example.test");
    fireEvent.change(
      await screen.findByTestId(entityMergeControlTestId("loser-record")),
      {
        target: { value: "00000000-0000-4000-8000-000000000802" },
      },
    );
    fireEvent.click(
      await screen.findByTestId(entityMergeControlTestId("confirm")),
    );

    await waitFor(() => {
      expect(
        screen.getByTestId(entityMergeControlTestId("message")).textContent,
      ).toContain(
        "merge_precondition_failed: carry_forward_identifier_collision",
      );
    });
    expect(
      screen.getByTestId(
        entityMergePreconditionDetailsTestId(
          "identity",
          "00000000-0000-4000-8000-000000000803",
        ),
      ).textContent,
    ).toContain("Blocking record: 00000000-0000-4000-8000-000000000801");
    expect(
      screen.getByTestId(
        entityMergePreconditionDetailsTestId(
          "identity",
          "00000000-0000-4000-8000-000000000803",
        ),
      ).textContent,
    ).toContain("Normalized value: collision@example.test");
    expect(currentRecordIds(identitiesViewSchemaId)).toEqual([
      "00000000-0000-4000-8000-000000000803",
      "00000000-0000-4000-8000-000000000802",
    ]);
    expect(mergeSubmissions).toBe(1);
  });

  it("keeps party-link mutations syncing until workbook and references refresh", async () => {
    const linkedTask = taskRequestRow(
      "00000000-0000-4000-8000-000000000901",
      4,
      "Task requester link",
      "Requester raw",
      "00000000-0000-4000-8000-000000000911",
    );
    const clearedTask = taskRequestRow(
      "00000000-0000-4000-8000-000000000901",
      5,
      "Task requester link",
      "Requester raw",
      null,
    );
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: taskRequestsViewSchemaId },
      selected_view_schema_id: taskRequestsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.genericRowsByView[taskRequestsViewSchemaId] = [linkedTask];
    scenario.genericRowsByView[partiesViewSchemaId] = [
      partyRow("00000000-0000-4000-8000-000000000911", "Requester Party"),
    ];
    const patchResponse = deferred<Response>();
    const refreshResponse = deferred<Response>();
    let patchAccepted = false;
    let refreshStarted = false;
    let clearPatchBody: Record<string, unknown> | null = null;
    scenario.recordPatchResponseOverride = (recordId, init) => {
      if (recordId !== "00000000-0000-4000-8000-000000000901") {
        return null;
      }
      clearPatchBody = JSON.parse(String(init?.body ?? "{}")) as Record<
        string,
        unknown
      >;
      return patchResponse.promise;
    };
    scenario.queryResponseOverride = (viewSchemaId) => {
      if (
        viewSchemaId === taskRequestsViewSchemaId &&
        patchAccepted &&
        !refreshStarted
      ) {
        refreshStarted = true;
        return refreshResponse.promise;
      }
      return null;
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    await expectRecordIds(taskRequestsViewSchemaId, [
      "00000000-0000-4000-8000-000000000901",
    ]);
    fireEvent.click(
      screen.getByTestId(
        rowCellTestId("00000000-0000-4000-8000-000000000901", "task.title"),
      ),
    );
    fireEvent.click(
      await screen.findByTestId(
        workbookInspectorToggleTestId(taskRequestsViewSchemaId),
      ),
    );
    fireEvent.change(
      await screen.findByTestId(
        genericEditRecordSelectTestId(taskRequestsViewSchemaId),
      ),
      { target: { value: "00000000-0000-4000-8000-000000000901" } },
    );
    const clearButton = await screen.findByTestId(
      coordinationWorkflowTestId("party-clear-link"),
    );
    fireEvent.click(clearButton);

    await waitFor(() => {
      expect(clearPatchBody).toMatchObject({
        view_schema_id: taskRequestsViewSchemaId,
        base_row_version: 4,
        changes: [{ field_key: "task.requester_party_id", value: null }],
      });
    });
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Syncing");
    expect((clearButton as HTMLButtonElement).disabled).toBe(true);

    patchAccepted = true;
    scenario.genericRowsByView[taskRequestsViewSchemaId] = [clearedTask];
    patchResponse.resolve(
      successEnvelope({
        view_schema_id: taskRequestsViewSchemaId,
        change_set_id: "30000000-0000-4000-8000-000000000001",
        row: {
          ...clearedTask,
          record_id: "00000000-0000-4000-8000-000000004101",
        },
      }),
    );
    await waitFor(() => {
      expect(refreshStarted).toBe(true);
    });
    expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Syncing");
    expect((clearButton as HTMLButtonElement).disabled).toBe(true);

    refreshResponse.resolve(
      successEnvelope({
        incident_id: "10000000-0000-4000-8000-000000000001",
        view_schema_id: taskRequestsViewSchemaId,
        rows: [clearedTask],
      }),
    );

    await waitFor(() => {
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe("Saved");
    });
    expect((clearButton as HTMLButtonElement).disabled).toBe(false);
    expect(currentRecordIds(taskRequestsViewSchemaId)).toEqual([
      "00000000-0000-4000-8000-000000000901",
    ]);
  });

  it("keeps failed generic party-link mutations in Conflict", async () => {
    scenario.startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: taskRequestsViewSchemaId },
      selected_view_schema_id: taskRequestsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    scenario.genericRowsByView[taskRequestsViewSchemaId] = [
      taskRequestRow(
        "00000000-0000-4000-8000-000000000901",
        4,
        "Task requester conflict",
        "Requester raw",
        "00000000-0000-4000-8000-000000000911",
      ),
    ];
    let clearPatchBody: Record<string, unknown> | null = null;
    scenario.recordPatchResponseOverride = (recordId, init) => {
      if (recordId !== "00000000-0000-4000-8000-000000000901") {
        return null;
      }
      clearPatchBody = JSON.parse(String(init?.body ?? "{}")) as Record<
        string,
        unknown
      >;
      return errorEnvelope("row_version_conflict", 409);
    };

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    await expectRecordIds(taskRequestsViewSchemaId, [
      "00000000-0000-4000-8000-000000000901",
    ]);
    fireEvent.click(
      screen.getByTestId(
        rowCellTestId("00000000-0000-4000-8000-000000000901", "task.title"),
      ),
    );
    fireEvent.click(
      await screen.findByTestId(
        workbookInspectorToggleTestId(taskRequestsViewSchemaId),
      ),
    );
    fireEvent.change(
      await screen.findByTestId(
        genericEditRecordSelectTestId(taskRequestsViewSchemaId),
      ),
      { target: { value: "00000000-0000-4000-8000-000000000901" } },
    );
    fireEvent.click(
      await screen.findByTestId(coordinationWorkflowTestId("party-clear-link")),
    );

    await waitFor(() => {
      expect(clearPatchBody).toMatchObject({
        view_schema_id: taskRequestsViewSchemaId,
        base_row_version: 4,
        changes: [{ field_key: "task.requester_party_id", value: null }],
      });
      expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
        "Conflict",
      );
    });
    expect(
      screen.getAllByText("This row changed; refresh it before retrying."),
    ).toHaveLength(2);
    expect(screen.getByText("Public error code")).not.toBeNull();
    expect(screen.getAllByText("row_version_conflict")).toHaveLength(2);
  });

  it("retains a created party for explicit link retry after partial completion", async () => {
    const createdPartyId = "00000000-0000-4000-8000-000000004201";
    const createPartyFromText = vi.fn(async () => ({
      kind: "accepted" as const,
      value: {
        changeSetId: "00000000-0000-4000-8000-000000005201",
        row: partyRow(createdPartyId, "Created Party"),
        viewSchemaId: partiesViewSchemaId,
      },
    }));
    const mutationCommands: GenericMutationCommandPort = {
      canCreateRecord: () => true,
      createRecord: async () => ({
        kind: "rejected",
        failure: { kind: "terminal", message: "not used" },
      }),
      createPartyFromText,
      patchRecord: async () => ({
        kind: "rejected",
        failure: { kind: "terminal", message: "not used" },
      }),
    };
    const submitLinkPatch = vi
      .fn<() => Promise<boolean>>()
      .mockResolvedValueOnce(false)
      .mockResolvedValueOnce(true);
    const rejectMutationFailure = vi.fn();
    const setValidationError = vi.fn();
    const selectedRow = taskRequestRow(
      "00000000-0000-4000-8000-000000000901",
      4,
      "Task requester link",
      "Created Party",
      null,
    );
    const { result } = renderHook(() =>
      useGenericPartyLinkWorkflow({
        mutation: {
          beginMutation: vi.fn(),
          rejectMutationFailure,
          setValidationError,
        },
        mutationCommands,
        originViewSchemaId: taskRequestsViewSchemaId,
        partyLinkPairs: [
          {
            key: "requester",
            label: "Requester",
            refFieldKey: "task.requester_party_id",
            textFieldKey: "task.requester_party_text",
          },
        ],
        resetKey: "reset-1",
        selectedRow,
        selectedSubject: {
          kind: "live",
          label: "Task requester link",
          recordId: selectedRow.record_id,
          rowVersion: selectedRow.row_version,
          surfaceLabel: "Task Requests",
          viewSchemaId: taskRequestsViewSchemaId,
        },
        submitLinkPatch,
      }),
    );

    await act(async () => result.current.createPartyFromText());
    expect(createPartyFromText).toHaveBeenCalledTimes(1);
    expect(submitLinkPatch).toHaveBeenNthCalledWith(
      1,
      [
        {
          field_key: "task.requester_party_id",
          value: createdPartyId,
        },
      ],
      "party-link-created",
    );
    expect(result.current.partialCompletionMessage).toContain(
      "party was created",
    );

    await act(async () => result.current.retryCreatedPartyLink());
    expect(createPartyFromText).toHaveBeenCalledTimes(1);
    expect(submitLinkPatch).toHaveBeenCalledTimes(2);
    expect(result.current.partialCompletionMessage).toBeNull();
    expect(rejectMutationFailure).not.toHaveBeenCalled();
    expect(setValidationError).not.toHaveBeenCalled();
  });

  it("issues opaque evidence preview and download handles from the evidence surface", async () => {
    scenario.evidenceRows = [
      evidenceStateRow(
        "00000000-0000-4000-8000-000000004002",
        4,
        "EDR package",
        { lifecycleState: "available", uploadState: "available" },
      ),
    ];
    const anchorClick = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => undefined);

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );
    fireEvent.click(
      await screen.findByTestId(
        evidencePreviewButtonTestId("00000000-0000-4000-8000-000000004002"),
      ),
    );

    const frame = await screen.findByTestId(
      evidencePreviewFrameTestId("00000000-0000-4000-8000-000000004002"),
    );
    expect(frame.getAttribute("src")).toBe(
      "/api/v1/evidence-handles/preview-token",
    );
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining(
        "/api/v1/evidence-records/00000000-0000-4000-8000-000000004002/preview-handle",
      ),
      expect.objectContaining({ method: "POST", body: "{}" }),
    );

    fireEvent.click(
      screen.getByTestId(
        evidenceDownloadButtonTestId("00000000-0000-4000-8000-000000004002"),
      ),
    );

    await waitFor(() => {
      expect(anchorClick).toHaveBeenCalled();
    });
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining(
        "/api/v1/evidence-records/00000000-0000-4000-8000-000000004002/download-handle",
      ),
      expect.objectContaining({ method: "POST", body: "{}" }),
    );
  });

  it("Verify attach flow uses generated protocol types, public error envelopes, and stable evidence selectors without raw object URLs or paths.", async () => {
    scenario.evidenceRows = [
      evidenceStateRow(
        "00000000-0000-4000-8000-000000004010",
        4,
        "Attach target",
        {
          lifecycleState: "available",
          uploadState: "available",
        },
      ),
      evidenceStateRow(
        "00000000-0000-4000-8000-000000004011",
        5,
        "Blocked target",
        {
          lifecycleState: "quarantined",
          uploadState: "available",
        },
      ),
      evidenceStateRow(
        "00000000-0000-4000-8000-000000004012",
        6,
        "Failed target",
        {
          lifecycleState: "available",
          uploadState: "failed",
        },
      ),
      evidenceStateRow(
        "00000000-0000-4000-8000-000000004013",
        7,
        "Inconsistent target",
        {
          lifecycleState: "available",
          uploadState: "storage-backend-mismatch",
        },
      ),
      evidenceStateRow(
        "00000000-0000-4000-8000-000000004014",
        8,
        "Raw handle target",
        {
          lifecycleState: "available",
          uploadState: "available",
        },
      ),
      evidenceStateRow(
        "00000000-0000-4000-8000-000000004015",
        9,
        "Public error target",
        {
          lifecycleState: "available",
          uploadState: "available",
        },
      ),
    ];
    scenario.handleHrefByRecordID["00000000-0000-4000-8000-000000004014"] = {
      preview:
        "https://minio.internal/cartulary-evidence-bucket/object_blob_storage_key_v1",
    };
    scenario.handleErrorByRecordID["00000000-0000-4000-8000-000000004015"] =
      rawStorageErrorEnvelope();
    scenario.attachErrorByRecordID["00000000-0000-4000-8000-000000004015"] =
      rawStorageErrorEnvelope();
    const anchorClick = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => undefined);

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );

    for (const recordId of [
      "00000000-0000-4000-8000-000000004010",
      "00000000-0000-4000-8000-000000004011",
      "00000000-0000-4000-8000-000000004012",
      "00000000-0000-4000-8000-000000004013",
      "00000000-0000-4000-8000-000000004014",
      "00000000-0000-4000-8000-000000004015",
    ]) {
      const attachInput = await screen.findByTestId(
        evidenceAttachFileInputTestId(recordId),
      );
      expect(attachInput.getAttribute("data-testid")).toBe(
        evidenceAttachFileInputTestId(recordId),
      );
      expect(
        screen
          .getByTestId(evidencePreviewButtonTestId(recordId))
          .getAttribute("data-testid"),
      ).toBe(evidencePreviewButtonTestId(recordId));
      expect(
        screen
          .getByTestId(evidenceDownloadButtonTestId(recordId))
          .getAttribute("data-testid"),
      ).toBe(evidenceDownloadButtonTestId(recordId));
    }
    expect(
      (
        screen.getByTestId(
          evidencePreviewButtonTestId("00000000-0000-4000-8000-000000004011"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      (
        screen.getByTestId(
          evidenceDownloadButtonTestId("00000000-0000-4000-8000-000000004012"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      (
        screen.getByTestId(
          evidencePreviewButtonTestId("00000000-0000-4000-8000-000000004013"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      screen.getByTestId(
        evidenceAccessMessageTestId("00000000-0000-4000-8000-000000004011"),
      ).textContent,
    ).toContain("Blocked:");
    expect(
      screen.getByTestId(
        evidenceAccessMessageTestId("00000000-0000-4000-8000-000000004012"),
      ).textContent,
    ).toContain("Failed:");
    expect(
      screen.getByTestId(
        evidenceAccessMessageTestId("00000000-0000-4000-8000-000000004013"),
      ).textContent,
    ).toContain("Inconsistent:");

    fireEvent.change(
      screen.getByTestId(
        evidenceAttachFileInputTestId("00000000-0000-4000-8000-000000004010"),
      ),
      {
        target: {
          files: [
            new File(["safe evidence body"], "safe-evidence.txt", {
              type: "text/plain",
            }),
          ],
        },
      },
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(
          evidenceAccessMessageTestId("00000000-0000-4000-8000-000000004010"),
        ).textContent,
      ).toBe("Evidence attached.");
    });
    const createBlobCall = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith("/api/v1/object-blobs"),
    );
    expect(createBlobCall).toBeDefined();
    expect(
      JSON.parse(String((createBlobCall?.[1] as RequestInit).body)),
    ).toEqual({
      incident_id: "10000000-0000-4000-8000-000000000001",
      client_txn_id: expect.stringMatching(/^evidence-blob-/u),
      byte_size: 18,
      filename_hint: "safe-evidence.txt",
      content_type_hint: "text/plain",
    });
    const attachCall = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith(
        "/api/v1/evidence-records/00000000-0000-4000-8000-000000004010/attach-blob",
      ),
    );
    expect(attachCall).toBeDefined();
    expect(JSON.parse(String((attachCall?.[1] as RequestInit).body))).toEqual({
      object_blob_id: "00000000-0000-4000-8000-000000003001",
      base_row_version: 4,
      client_txn_id: expect.stringMatching(/^evidence-attach-/u),
    });
    const lifecyclePatchCall = fetchMock.mock.calls.find(
      ([input, init]) =>
        String(input).endsWith(
          "/api/v1/records/00000000-0000-4000-8000-000000004010",
        ) &&
        (init as RequestInit | undefined)?.method === "PATCH" &&
        String((init as RequestInit | undefined)?.body).includes(
          "evidence.lifecycle_state",
        ),
    );
    expect(lifecyclePatchCall).toBeDefined();
    expect(
      JSON.parse(String((lifecyclePatchCall?.[1] as RequestInit).body)),
    ).toEqual({
      view_schema_id: evidenceViewSchemaId,
      base_row_version: 2,
      client_txn_id: expect.stringMatching(/^evidence-available-/u),
      changes: [{ field_key: "evidence.lifecycle_state", value: "available" }],
    });

    fireEvent.change(
      screen.getByTestId(
        evidenceAttachFileInputTestId("00000000-0000-4000-8000-000000004015"),
      ),
      {
        target: {
          files: [
            new File(["unsafe evidence body"], "unsafe-evidence.txt", {
              type: "text/plain",
            }),
          ],
        },
      },
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(
          evidenceAccessMessageTestId("00000000-0000-4000-8000-000000004015"),
        ).textContent,
      ).toBe("Conflict.");
    });

    fireEvent.click(
      screen.getByTestId(
        evidencePreviewButtonTestId("00000000-0000-4000-8000-000000004014"),
      ),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(
          evidenceAccessMessageTestId("00000000-0000-4000-8000-000000004014"),
        ).textContent,
      ).toBe("Evidence handle is unavailable.");
    });
    expect(
      screen.queryByTestId(
        evidencePreviewFrameTestId("00000000-0000-4000-8000-000000004014"),
      ),
    ).toBeNull();

    fireEvent.click(
      screen.getByTestId(
        evidencePreviewButtonTestId("00000000-0000-4000-8000-000000004015"),
      ),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(
          evidenceAccessMessageTestId("00000000-0000-4000-8000-000000004015"),
        ).textContent,
      ).toBe("Conflict.");
    });

    fireEvent.click(
      screen.getByTestId(
        evidencePreviewButtonTestId("00000000-0000-4000-8000-000000004010"),
      ),
    );
    const frame = await screen.findByTestId(
      evidencePreviewFrameTestId("00000000-0000-4000-8000-000000004010"),
    );
    expect(frame.getAttribute("src")).toBe(
      "/api/v1/evidence-handles/preview-token",
    );
    fireEvent.click(
      screen.getByTestId(
        evidenceDownloadButtonTestId("00000000-0000-4000-8000-000000004010"),
      ),
    );
    await waitFor(() => {
      expect(anchorClick).toHaveBeenCalled();
    });
    expectNoRawStorageDetails(document.body);
  });

  it("orchestrates selected Timeline evidence attachment inline", async () => {
    scenario.timelineRows = [
      timelineRow("21000000-0000-4000-8000-000000000001", 1, "Selected row", 0),
    ];

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(timelineViewSchemaId)),
    );
    await openTimelineInspectorFromContext(
      "21000000-0000-4000-8000-000000000001",
    );
    const input = await screen.findByTestId(
      timelineEvidenceFileInputTestId("21000000-0000-4000-8000-000000000001"),
    );
    fireEvent.change(input, {
      target: {
        files: [
          new File(["screenshot body"], "screenshot.txt", {
            type: "text/plain",
          }),
        ],
      },
    });

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(`/views/${evidenceViewSchemaId}/rows`),
        expect.objectContaining({ method: "POST" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/object-blobs"),
        expect.objectContaining({ method: "POST" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/object-uploads/test-token"),
        expect.objectContaining({ method: "PUT" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(
          "/api/v1/records/21000000-0000-4000-8000-000000000001",
        ),
        expect.objectContaining({ method: "PATCH" }),
      );
    });
    const evidenceCreateCall = fetchMock.mock.calls.find(
      ([url, init]) =>
        String(url).includes(`/views/${evidenceViewSchemaId}/rows`) &&
        (init as RequestInit).method === "POST",
    );
    expect(evidenceCreateCall).toBeDefined();
    expect(
      JSON.parse(String((evidenceCreateCall?.[1] as RequestInit).body)),
    ).toEqual({
      client_txn_id: expect.stringMatching(/^timeline-client-/u),
      "evidence.title": "screenshot.txt",
      "evidence.collector_party_text": "Workbook upload",
      "evidence.lifecycle_state": "available",
      "evidence.initial_object_blob_id": "00000000-0000-4000-8000-000000003001",
    });
    expect(
      fetchMock.mock.calls.some(([url]) =>
        String(url).includes("/api/v1/evidence-records/"),
      ),
    ).toBe(false);
    const uploadCall = fetchMock.mock.calls.find(([url]) =>
      String(url).includes("/api/v1/object-uploads/test-token"),
    );
    expect(uploadCall).toBeDefined();
    const uploadInit = uploadCall?.[1] as RequestInit;
    expect(uploadInit.credentials).toBe("include");
    const uploadHeaders = uploadInit.headers as Headers;
    expect(uploadHeaders.get("X-CSRF-Token")).toBe("evidence-shell-csrf");
    expect(uploadHeaders.get("X-Upload-Contract")).toBe("evidence_lifecycle");
    expect(uploadHeaders.get("Content-Type")).toBe("text/plain");
  });

  it("surfaces upload failures inline without issuing Timeline patches", async () => {
    scenario.uploadShouldFail = true;
    scenario.timelineRows = [
      timelineRow("21000000-0000-4000-8000-000000000001", 1, "Selected row", 0),
    ];

    render(<WorkbookShell incidentId="10000000-0000-4000-8000-000000000001" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(timelineViewSchemaId)),
    );
    await openTimelineInspectorFromContext(
      "21000000-0000-4000-8000-000000000001",
    );
    fireEvent.change(
      await screen.findByTestId(
        timelineEvidenceFileInputTestId("21000000-0000-4000-8000-000000000001"),
      ),
      {
        target: {
          files: [
            new File(["screenshot body"], "screenshot.txt", {
              type: "text/plain",
            }),
          ],
        },
      },
    );

    await waitFor(() => {
      expect(
        screen.getByTestId(timelineInspectorMessageTestId()).textContent,
      ).toBe("upload_failed_500");
    });
    expect(
      fetchMock.mock.calls.some(([input, init]) => {
        return (
          String(input).endsWith(
            "/api/v1/records/21000000-0000-4000-8000-000000000001",
          ) && ((init as RequestInit | undefined)?.method ?? "GET") === "PATCH"
        );
      }),
    ).toBe(false);
  });
});

function timelineRow(
  recordId: string,
  rowVersion: number,
  summary: string,
  evidenceCount: number,
  relationships: {
    hostRefs?: Array<Record<string, unknown>>;
    identityRefs?: Array<Record<string, unknown>>;
  } = {},
) {
  const hasEvidence = evidenceCount > 0;
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.date_entered_text": { value: "2026-04-24" },
      "timeline.analyst_text": { value: "Analyst" },
      "timeline.mitre_stage_text": { value: "" },
      "timeline.device_object_text": { value: "" },
      "timeline.ip_address_text": { value: "" },
      "timeline.activity_utc_text": { value: "" },
      "timeline.activity_local_text": { value: "" },
      "timeline.activity_synopsis_text": { value: summary },
      "timeline.raw_activity_text": { value: "" },
      "timeline.data_source_text": { value: "" },
      "timeline.host_refs": {
        value: {
          ...collectionValue(relationships.hostRefs ?? []),
          ordered: true,
        },
      },
      "timeline.identity_refs": {
        value: {
          ...collectionValue(relationships.identityRefs ?? []),
          ordered: true,
        },
      },
      "timeline.evidence_count": { value: evidenceCount },
      "timeline.tags": {
        value: collectionValue([]),
      },
      "timeline.edited_at": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.recorded_at": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.activity_sort_ts": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.capture_state": { value: "rough" },
      "timeline.replacement_record_id": { value: null },
      "timeline.date_entered_sort_day": { value: "2026-04-24" },
      "timeline.activity_time_pair_state": { value: "disabled" },
      "timeline.has_evidence": { value: hasEvidence },
      "timeline.attached_evidence_ids": {
        value: collectionValue([]),
      },
      "timeline.has_unresolved_mentions": { value: false },
    },
    group_values: {
      "timeline.date_entered_sort_day": "2026-04-24",
      "timeline.activity_time_pair_state": "disabled",
      "timeline.capture_state": "rough",
      "timeline.has_evidence": hasEvidence,
      "timeline.has_unresolved_mentions": false,
    },
  };
}

function indicatorRow(
  recordId: string,
  rowVersion: number,
  indicatorType: string,
  displayValue: string,
) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "indicator.indicator_type": { value: indicatorType },
      "indicator.value_kind": { value: "atomic" },
      "indicator.display_value": { value: displayValue },
      "indicator.normalized_value": { value: displayValue },
      "indicator.defanged_value": { value: displayValue },
      "indicator.hash_algorithm": { value: null },
      "indicator.hash_value": { value: null },
      "indicator.stix_pattern": { value: null },
      "indicator.first_observed_at": { value: "2026-04-24T10:00:00.000Z" },
      "indicator.last_observed_at": { value: "2026-04-24T10:00:00.000Z" },
      "indicator.observation_count": { value: 1 },
      "indicator.lifecycle_summary": { value: "active" },
      "indicator.supporting_link_count": { value: 1 },
    },
    group_values: {
      "indicator.indicator_type": indicatorType,
      "indicator.value_kind": "atomic",
      "indicator.lifecycle_summary": "active",
    },
  };
}

function evidenceRow(recordId: string, rowVersion: number, title: string) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "evidence.title": { value: title },
      "evidence.lifecycle_state": { value: "requested" },
      "evidence.requested_at": { value: null },
      "evidence.received_at": { value: null },
      "evidence.storage_ref": { value: "" },
      "evidence.blob_hash": { value: "" },
      "evidence.collector_party_text": { value: "Workbook upload" },
      "evidence.collector_party_id": { value: null },
      "evidence.source_party_text": { value: "" },
      "evidence.source_party_id": { value: null },
      "evidence.upload_state": { value: "pending" },
      "evidence.linked_record_count": { value: 0 },
      "evidence.edited_at": { value: null },
    },
  };
}

function evidenceStateRow(
  recordId: string,
  rowVersion: number,
  title: string,
  state: { lifecycleState: string; uploadState: string },
) {
  const row = evidenceRow(recordId, rowVersion, title);
  return {
    ...row,
    cells: {
      ...row.cells,
      "evidence.lifecycle_state": { value: state.lifecycleState },
      "evidence.upload_state": { value: state.uploadState },
    },
  };
}

const rawStorageLeakSentinels = [
  "https://minio.internal",
  "object://object-blob-storage-ref",
  "cartulary-evidence-bucket",
  "object_blob_storage_key_v1",
  "/var/lib/cartulary/object-blobs",
  "seaweedfs",
  "s3_backend",
  "object-store implementation",
] as const;

function rawStorageErrorEnvelope(): Response {
  return new Response(
    JSON.stringify({
      error: {
        status: 409,
        code: "object_store_unavailable",
        message:
          "https://minio.internal/cartulary-evidence-bucket object_blob_storage_key_v1 /var/lib/cartulary/object-blobs seaweedfs s3_backend object-store implementation",
        request_id: "req-raw-storage",
        retryable: false,
        details: {
          reason_code: "object_blob_storage_key_malformed",
          raw_object_url:
            "https://minio.internal/cartulary-evidence-bucket/object_blob_storage_key_v1",
          raw_object_ref: "object://object-blob-storage-ref",
          raw_path:
            "/var/lib/cartulary/object-blobs/object_blob_storage_key_v1",
          raw_object_key: "object_blob_storage_key_v1",
          bucket_name: "cartulary-evidence-bucket",
          backend_path: "/var/lib/cartulary/object-blobs",
          storage_backend: "s3_backend",
          object_store_detail: "seaweedfs object-store implementation",
        },
      },
    }),
    {
      headers: { "Content-Type": "application/json" },
      status: 409,
    },
  );
}

function expectNoRawStorageDetails(root: ParentNode) {
  const text = root.textContent ?? "";
  for (const sentinel of rawStorageLeakSentinels) {
    expect(text).not.toContain(sentinel);
  }
}

function statusReviewRow(
  recordId: string,
  rowVersion: number,
  summary: string,
) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "status_review.timestamp_utc": { value: "2026-04-24T15:00:00.000Z" },
      "status_review.review_owner_user_id": { value: "user-1" },
      "status_review.current_state_summary": { value: summary },
      "status_review.active_risks_summary": { value: null },
      "status_review.next_report_at": { value: null },
      "status_review.blocked_task_ids": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "status_review.pending_evidence_ids": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "status_review.open_decision_ids": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "status_review.status_review_id": { value: recordId },
      "status_review.timestamp_day": { value: "2026-04-24" },
      "status_review.next_report_day": { value: null },
      "status_review.updated_at": { value: "2026-04-24T15:00:00.000Z" },
    },
    group_values: {
      "status_review.timestamp_day": "2026-04-24",
      "status_review.review_owner_user_id": "user-1",
      "status_review.next_report_day": null,
    },
  };
}

function taskRequestRow(
  recordId: string,
  rowVersion: number,
  title: string,
  requesterText: string | null,
  requesterPartyId: string | null,
) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "task.title": { value: title },
      "task.task_kind": { value: "request" },
      "task.workstream": { value: null },
      "task.status": { value: "open" },
      "task.requester_party_text": { value: requesterText },
      "task.requester_party_id": { value: requesterPartyId },
      "task.owner_user_id": { value: null },
      "task.decision_record_id": { value: null },
      "task.due_at": { value: null },
      "task.priority": { value: "normal" },
      "task.external_ticket_ref": { value: null },
      "task.blocked_reason": { value: null },
      "task.completed_at": { value: null },
      "task.closure_summary": { value: null },
      "task.linked_record_ids": { value: collectionValue([]) },
      "task.linked_record_count": { value: 0 },
      "task.updated_at": { value: "2026-04-24T15:00:00.000Z" },
      "task.no_owner": { value: true },
    },
  };
}

function partyRow(recordId: string, displayName: string) {
  return {
    record_id: recordId,
    row_version: 1,
    cells: {
      "party.display_name": { value: displayName },
      "party.party_kind": { value: "person" },
      "party.organization_name": { value: null },
      "party.role_title": { value: null },
      "party.primary_email": { value: null },
      "party.timezone_name": { value: null },
      "party.external_ref": { value: null },
      "party.notes": { value: null },
      "party.updated_at": { value: "2026-04-24T15:00:00.000Z" },
    },
  };
}

function parseRequestBody(init: RequestInit | undefined) {
  return JSON.parse(String(init?.body ?? "{}")) as {
    filters?: Array<{
      arg?: { value?: unknown };
    }>;
  };
}

function stringFilterValue(body: ReturnType<typeof parseRequestBody>) {
  const [filter] = body.filters ?? [];
  return typeof filter?.arg?.value === "string" ? filter.arg.value : null;
}

function applyGenericFilter(
  surface: Parameters<typeof gridFilterFieldTestId>[0],
  fieldKey: string,
  value: string,
) {
  fireEvent.click(
    screen.getByTestId(workbookFilterPopoverTriggerTestId(surface)),
  );
  fireEvent.change(screen.getByTestId(gridFilterFieldTestId(surface)), {
    target: { value: fieldKey },
  });
  fireEvent.change(screen.getByTestId(gridFilterValueTestId(surface)), {
    target: { value },
  });
  fireEvent.click(screen.getByTestId(gridFilterApplyTestId(surface)));
}

function openSavedViewActions(
  surface: Parameters<typeof savedViewSelectorTestId>[0],
) {
  fireEvent.click(
    screen.getByTestId(savedViewActionMenuTriggerTestId(surface)),
  );
}

async function expectRecordIds(
  surface: Parameters<typeof gridShellTestId>[0],
  expected: string[],
) {
  await waitFor(() => {
    expect(currentRecordIds(surface)).toEqual(expected);
  });
}

function currentRecordIds(surface: Parameters<typeof gridShellTestId>[0]) {
  const grid = screen.getByTestId(gridShellTestId(surface));
  return Array.from(grid.querySelectorAll(gridSavedRowsSelector())).map(
    (row) => row.getAttribute("data-grid-record-id") ?? "",
  );
}

async function flushMicrotasks() {
  await Promise.resolve();
  await Promise.resolve();
}

type TestViewRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  group_values?: Record<string, unknown>;
};

function collectionValue(items: Array<Record<string, unknown>>) {
  return {
    kind: "collection_value_v1",
    ordered: false,
    items,
  };
}

function timelineEntityRef(
  recordId: string,
  displayText: string,
  itemRef: string,
) {
  return {
    item_ref: itemRef,
    item_kind: "resolved_ref",
    entity_type: "host",
    display_text: displayText,
    raw_text: displayText,
    resolved_record_id: recordId,
    mention_row_version: 1,
    resolution_method: "manual",
    auto_resolved: false,
  };
}

function hostRow({
  aliases = [],
  displayName,
  fqdn = null,
  hostname,
  linkedEventCount = 0,
  recordId,
  reusableIdentifiers = [],
  rowVersion,
}: {
  aliases?: string[];
  displayName: string;
  fqdn?: string | null;
  hostname: string;
  linkedEventCount?: number;
  recordId: string;
  reusableIdentifiers?: Array<{
    identifierClass: string;
    itemRef: string;
    rawValue: string;
    normalizedValue?: string;
    displayText?: string;
  }>;
  rowVersion: number;
}): TestViewRow {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "host.display_name": { value: displayName },
      "host.hostname": { value: hostname },
      "host.aad_device_id": { value: "" },
      "host.fqdn": { value: fqdn ?? "" },
      "host.reusable_identifiers": {
        value: collectionValue(
          reusableIdentifiers.map((identifier) => ({
            item_ref: identifier.itemRef,
            item_kind: "reusable_identifier",
            identifier_class: identifier.identifierClass,
            raw_value: identifier.rawValue,
            normalized_value:
              identifier.normalizedValue ?? identifier.rawValue.toLowerCase(),
            display_text: identifier.displayText ?? identifier.rawValue,
          })),
        ),
      },
      "host.aliases": {
        value: collectionValue(
          aliases.map((alias, index) => ({
            item_ref: `entity_alias:00000000-0000-0000-0000-${String(index + 1).padStart(12, "0")}`,
            item_kind: "alias",
            alias_text: alias,
            display_text: alias,
          })),
        ),
      },
      "host.host_state": { value: "canonical" },
      "host.linked_event_count": { value: linkedEventCount },
      "host.evidence_count": { value: 0 },
      "host.location": { value: "" },
      "host.os_platform": { value: "" },
      "host.business_owner": { value: "" },
      "host.criticality": { value: "" },
      "host.containment_status": { value: "" },
      "host.edited_at": { value: "2026-04-24T15:00:00.000Z" },
    },
    group_values: {
      "host.host_state": "canonical",
      "host.criticality": "",
      "host.containment_status": "",
    },
  };
}

function identityRow({
  aliases = [],
  displayName,
  email = "",
  recordId,
  reusableIdentifiers = [],
  rowVersion,
  upn,
}: {
  aliases?: string[];
  displayName: string;
  email?: string;
  recordId: string;
  reusableIdentifiers?: Array<{
    identifierClass: string;
    itemRef: string;
    rawValue: string;
    normalizedValue?: string;
    displayText?: string;
  }>;
  rowVersion: number;
  upn: string;
}): TestViewRow {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "identity.display_name": { value: displayName },
      "identity.aad_object_id": { value: "" },
      "identity.sid": { value: "" },
      "identity.upn": { value: upn },
      "identity.email": { value: email },
      "identity.sam_account_name": { value: "" },
      "identity.reusable_identifiers": {
        value: collectionValue(
          reusableIdentifiers.map((identifier) => ({
            item_ref: identifier.itemRef,
            item_kind: "reusable_identifier",
            identifier_class: identifier.identifierClass,
            raw_value: identifier.rawValue,
            normalized_value:
              identifier.normalizedValue ?? identifier.rawValue.toLowerCase(),
            display_text: identifier.displayText ?? identifier.rawValue,
          })),
        ),
      },
      "identity.aliases": {
        value: collectionValue(
          aliases.map((alias, index) => ({
            item_ref: `entity_alias:10000000-0000-0000-0000-${String(index + 1).padStart(12, "0")}`,
            item_kind: "alias",
            alias_text: alias,
            display_text: alias,
          })),
        ),
      },
      "identity.identity_state": { value: "canonical" },
      "identity.linked_event_count": { value: 0 },
      "identity.evidence_count": { value: 0 },
      "identity.privilege_level": { value: "" },
      "identity.mfa_state": { value: "" },
      "identity.reset_status": { value: "" },
      "identity.edited_at": { value: "2026-04-24T15:00:00.000Z" },
    },
    group_values: {
      "identity.identity_state": "canonical",
      "identity.privilege_level": "",
      "identity.mfa_state": "",
      "identity.reset_status": "",
    },
  };
}

function applyEntityClipboardPaste(
  viewSchemaId: string,
  currentRows: TestViewRow[],
  clipboardText: string,
): { allRows: TestViewRow[]; changedRows: TestViewRow[] } {
  const entityType = viewSchemaId === hostsViewSchemaId ? "host" : "identity";
  const displayField =
    entityType === "host" ? "host.display_name" : "identity.display_name";
  const primaryField = entityType === "host" ? "host.hostname" : "identity.upn";
  const changedRows: TestViewRow[] = [];
  const nextRows = [...currentRows];

  for (const [index, line] of clipboardText.split(/\r?\n/u).entries()) {
    if (line.trim() === "") {
      continue;
    }
    const [displayName = "", primaryValue = ""] = line.split("\t");
    const existingIndex = nextRows.findIndex(
      (row) => row.cells[primaryField]?.value === primaryValue,
    );
    if (existingIndex >= 0) {
      const current = nextRows[existingIndex];
      if (current === undefined) {
        continue;
      }
      const updated = {
        ...current,
        row_version: current.row_version + 1,
        cells: {
          ...current.cells,
          [displayField]: { value: displayName },
        },
      };
      nextRows[existingIndex] = updated;
      changedRows.push(updated);
      continue;
    }

    const created =
      entityType === "host"
        ? hostRow({
            displayName,
            hostname: primaryValue,
            recordId: `00000000-0000-4000-8000-00000000610${index}`,
            rowVersion: 1,
          })
        : identityRow({
            displayName,
            recordId: `00000000-0000-4000-8000-00000000620${index}`,
            rowVersion: 1,
            upn: primaryValue,
          });
    nextRows.push(created);
    changedRows.push(created);
  }

  return { allRows: nextRows, changedRows };
}

describe("generic workbook mutation payloads", () => {
  it("builds required creates with direct values, timestamps, and explicit clears", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);

    expect(
      buildGenericCreateRequest(evidence, {}, "txn-evidence-missing"),
    ).toBeNull();
    expect(
      buildGenericCreateRequest(
        evidence,
        {
          "evidence.title": " Endpoint package ",
          "evidence.requested_at": "2026-04-24T12:00:00Z",
          "evidence.collector_party_id": "",
        },
        "txn-evidence-create",
      ),
    ).toMatchObject({
      client_txn_id: "txn-evidence-create",
      "evidence.title": "Endpoint package",
      "evidence.requested_at": "2026-04-24T12:00:00Z",
      "evidence.collector_party_id": null,
    });
  });

  it("builds direct clears and typed collection actions", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);
    const notes = requireViewContract("cartulary.view.notes.v1");
    const commLog = requireViewContract("cartulary.view.comm_log.v1");
    const handoff = requireViewContract("cartulary.view.handoff.v1");
    const decisions = requireViewContract("cartulary.view.decisions.v1");

    expect(
      buildGenericPatchChange(
        requireField(evidence, "evidence.source_party_id"),
        "",
      ),
    ).toEqual({ field_key: "evidence.source_party_id", value: null });
    expect(
      buildGenericPatchChange(requireField(notes, "note.tags"), " urgent "),
    ).toEqual({
      field_key: "note.tags",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_tag", tag_name: "urgent" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(commLog, "comm_log.audience_party_ids"),
        "00000000-0000-4000-8000-000000000911",
      ),
    ).toEqual({
      field_key: "comm_log.audience_party_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [
          {
            op: "add_party_ref",
            party_id: "00000000-0000-4000-8000-000000000911",
          },
        ],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(commLog, "comm_log.audience_party_ids"),
        "party_ref:00000000-0000-4000-8000-000000000911",
        "remove",
      ),
    ).toEqual({
      field_key: "comm_log.audience_party_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [
          {
            op: "remove_party_ref",
            item_ref: "party_ref:00000000-0000-4000-8000-000000000911",
          },
        ],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(commLog, "comm_log.attendee_party_ids"),
        "party_ref:party-2",
        "remove",
      ),
    ).toEqual({
      field_key: "comm_log.attendee_party_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "remove_party_ref", item_ref: "party_ref:party-2" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(decisions, "decision.support_refs"),
        "20000000-0000-4000-8000-000000000001",
      ),
    ).toEqual({
      field_key: "decision.support_refs",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [
          {
            op: "add_record_ref",
            linked_record_id: "20000000-0000-4000-8000-000000000001",
          },
        ],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(handoff, "handoff.open_risk_refs"),
        "risk_ref:abc",
        "remove",
      ),
    ).toEqual({
      field_key: "handoff.open_risk_refs",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "remove_risk_ref", item_ref: "risk_ref:abc" }],
      },
    });
  });
});
