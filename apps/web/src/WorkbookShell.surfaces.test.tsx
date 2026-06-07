import {
  dataTestIdSelector,
  evidenceAccessMessageTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
  evidencePreviewFrameTestId,
  genericEditRecordSelectTestId,
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  rowCellTestId,
  rowInspectButtonTestId,
  savedViewOptionTestId,
  savedViewSelectorTestId,
  surfaceTabTestId,
  systemViewSwitcherGroupTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { deferred } from "./fetchMockTestSupport";
import { errorEnvelope, successEnvelope } from "./timelineWorkbookTestSupport";
import {
  buildGenericCreatePayload,
  buildGenericPatchChange,
  WorkbookShell,
} from "./WorkbookShell";
import {
  commLogViewSchemaId,
  evidenceViewSchemaId,
  indicatorsViewSchemaId,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

const savedViewId = "11111111-1111-4111-8111-111111111111";

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

describe("WorkbookShell surface selection", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  let evidenceRows: Array<{
    record_id: string;
    row_version: number;
    cells: Record<string, { value: unknown }>;
  }>;
  let timelineRows: Array<{
    record_id: string;
    row_version: number;
    cells: Record<string, { value: unknown }>;
  }>;
  let genericRowsByView: Record<
    string,
    Array<{
      record_id: string;
      row_version: number;
      cells: Record<string, { value: unknown }>;
    }>
  >;
  let savedViews: Array<{
    saved_view_id: string;
    view_schema_id: string;
    display_name: string;
  }>;
  let queryResponseOverride:
    | ((
        viewSchemaId: string,
        init: RequestInit | undefined,
      ) => Promise<Response> | Response | null)
    | null;
  let startupResponseOverride:
    | (() => Promise<Response> | Response | null)
    | null;
  let recordPatchResponseOverride:
    | ((
        recordId: string,
        init: RequestInit | undefined,
      ) => Promise<Response> | Response | null)
    | null;
  let attachErrorByRecordID: Record<string, Response>;
  let handleErrorByRecordID: Record<string, Response>;
  let handleHrefByRecordID: Record<
    string,
    { download?: string; preview?: string }
  >;
  let startupSelection: {
    selected_sheet_ref: { kind: "saved_view" | "view_schema"; id: string };
    selected_view_schema_id: string;
    selected_saved_view: unknown | null;
    source: "default" | "explicit" | "home" | "timeline";
  };
  let uploadShouldFail: boolean;

  beforeEach(() => {
    window.history.replaceState({}, "", "/");
    evidenceRows = [];
    timelineRows = [];
    genericRowsByView = {};
    savedViews = [];
    queryResponseOverride = null;
    startupResponseOverride = null;
    recordPatchResponseOverride = null;
    attachErrorByRecordID = {};
    handleErrorByRecordID = {};
    handleHrefByRecordID = {};
    startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
      selected_view_schema_id: timelineViewSchemaId,
      selected_saved_view: null,
      source: "timeline",
    };
    uploadShouldFail = false;
    fetchMock = vi.fn(async (input: RequestInfo | URL, init?: RequestInit) => {
      const url = String(input);
      const method = (init?.method ?? "GET").toUpperCase();
      if (url.endsWith("/api/v1/auth/session")) {
        return successEnvelope({
          user_id: "user-1",
          memberships: [{ incident_id: "incident-1", role: "admin" }],
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1")) {
        return successEnvelope({
          incident_id: "incident-1",
          incident_key: "IR-1",
          title: "Incident 1",
          description: null,
          severity: null,
          tlp: null,
          current_phase: null,
          primary_external_case_ref: null,
          incident_version: 1,
        });
      }
      if (url.endsWith("/api/v1/incidents/incident-1/memberships")) {
        return successEnvelope({
          memberships: [
            {
              incident_id: "incident-1",
              user_id: "user-1",
              display_name: "Admin User",
              role: "admin",
              membership_version: 1,
            },
          ],
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-preferences/")) {
        return successEnvelope({
          default_sheet_ref: null,
          home_sheet_ref: null,
        });
      }
      if (url.includes("/api/v1/incidents/incident-1/workbook-startup")) {
        const override = startupResponseOverride?.();
        if (override) {
          return override;
        }
        return successEnvelope({
          incident_id: "incident-1",
          ...startupSelection,
          cleared_pointers: [],
          home_sheet_ref: null,
          default_sheet_ref: null,
        });
      }
      if (
        method === "GET" &&
        url.includes("/api/v1/incidents/incident-1/saved-views")
      ) {
        return successEnvelope({
          saved_views: savedViews,
        });
      }
      const evidenceHandleMatch = url.match(
        /\/api\/v1\/evidence-records\/([^/]+)\/(preview|download)-handle$/,
      );
      if (evidenceHandleMatch) {
        const recordId = decodeURIComponent(evidenceHandleMatch[1] ?? "");
        const kind = evidenceHandleMatch[2] as "download" | "preview";
        const error = handleErrorByRecordID[recordId];
        if (error) {
          return error;
        }
        const href =
          handleHrefByRecordID[recordId]?.[kind] ??
          `/api/v1/evidence-handles/${kind}-token`;
        return successEnvelope({
          href,
          method: "GET",
          filename: "evidence.txt",
          ...(kind === "preview" ? { preview_kind: "text_inline" } : {}),
          content_type: "text/plain",
        });
      }
      if (
        method === "POST" &&
        url.endsWith(
          `/api/v1/incidents/incident-1/views/${evidenceViewSchemaId}/rows`,
        )
      ) {
        return successEnvelope(
          {
            view_schema_id: evidenceViewSchemaId,
            change_set_id: "change-evidence",
            row: evidenceRow("evidence-created", 1, "Attached screenshot"),
          },
          201,
        );
      }
      if (method === "POST" && url.endsWith("/api/v1/object-blobs")) {
        return successEnvelope(
          {
            object_blob_id: "blob-created",
            upload_target: {
              href: "/api/v1/object-uploads/test-token",
              method: "PUT",
              headers: {
                "X-Upload-Contract": "phase5",
              },
            },
          },
          201,
        );
      }
      if (
        method === "PUT" &&
        url.endsWith("/api/v1/object-uploads/test-token")
      ) {
        return new Response(null, { status: uploadShouldFail ? 500 : 200 });
      }
      if (
        method === "POST" &&
        url.endsWith("/api/v1/evidence-records/evidence-created/attach-blob")
      ) {
        return successEnvelope({
          record_id: "evidence-created",
          row_version: 2,
        });
      }
      const attachMatch = url.match(
        /\/api\/v1\/evidence-records\/([^/]+)\/attach-blob$/,
      );
      if (method === "POST" && attachMatch) {
        const recordId = decodeURIComponent(attachMatch[1] ?? "");
        const error = attachErrorByRecordID[recordId];
        if (error) {
          return error;
        }
        const row = evidenceStateRow(recordId, 2, "Attached evidence", {
          lifecycleState: "available",
          uploadState: "available",
        });
        evidenceRows = [
          row,
          ...evidenceRows.filter((candidate) => candidate.record_id !== recordId),
        ];
        return successEnvelope({
          view_schema_id: evidenceViewSchemaId,
          change_set_id: "change-evidence-attach",
          row,
          object_blob_id: "blob-created",
        });
      }
      if (method === "PATCH" && url.includes("/api/v1/records/")) {
        const recordPatchMatch = url.match(
          /\/api\/v1\/records\/([^/?]+)(?:\?.*)?$/,
        );
        if (recordPatchMatch) {
          const override = recordPatchResponseOverride?.(
            recordPatchMatch[1] ?? "",
            init,
          );
          if (override) {
            return override;
          }
        }
      }
      if (method === "PATCH" && url.endsWith("/api/v1/records/timeline-1")) {
        const row = timelineRow("timeline-1", 2, "Selected row", 1);
        timelineRows = [row];
        return successEnvelope({
          view_schema_id: timelineViewSchemaId,
          change_set_id: "change-timeline",
          row,
        });
      }
      const viewQueryMatch = url.match(
        /\/api\/v1\/incidents\/incident-1\/views\/([^/]+)\/query(?:\?.*)?$/,
      );
      if (viewQueryMatch) {
        const viewSchemaId = decodeURIComponent(viewQueryMatch[1] ?? "");
        const override = queryResponseOverride?.(viewSchemaId, init);
        if (override) {
          return override;
        }
        return successEnvelope({
          incident_id: "incident-1",
          view_schema_id: viewSchemaId,
          rows:
            viewSchemaId === evidenceViewSchemaId
              ? evidenceRows
              : viewSchemaId === timelineViewSchemaId
                ? timelineRows
                : (genericRowsByView[viewSchemaId] ?? []),
        });
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

  afterEach(() => {
    cleanup();
    vi.restoreAllMocks();
    vi.unstubAllGlobals();
  });

  it("selects required built-in and system view surfaces by view_schema_id", async () => {
    render(<WorkbookShell incidentId="incident-1" />);

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
          .getByTestId(systemViewSwitcherGroupTestId("scope-assessment"))
          .querySelectorAll("[data-view-schema-id]"),
      ).map((option) => option.getAttribute("data-view-schema-id")),
    ).toEqual([
      "cartulary.view.indicators.v1",
      "cartulary.view.assessments.v1",
      "cartulary.view.parties.v1",
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
      "cartulary.view.comm_log.v1",
      "cartulary.view.handoff.v1",
    ]);
    const systemViewOptions = [
      ...Array.from(
        screen
          .getByTestId(systemViewSwitcherGroupTestId("scope-assessment"))
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
      "cartulary.view.parties.v1",
      "cartulary.view.task_requests.v1",
      "cartulary.view.decisions.v1",
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
            "scope-assessment",
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

    fireEvent.click(screen.getByTestId(systemViewSwitcherTriggerTestId()));
    fireEvent.click(
      screen.getByTestId(
        systemViewSwitcherOptionTestId(
          "scope-assessment",
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
    startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: evidenceViewSchemaId },
      selected_view_schema_id: evidenceViewSchemaId,
      selected_saved_view: null,
      source: "default",
    };

    render(<WorkbookShell incidentId="incident-1" />);

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
    startupResponseOverride = () => delayedStartup.promise;

    render(<WorkbookShell incidentId="incident-1" />);

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
        incident_id: "incident-1",
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
    startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: statusReviewViewSchemaId },
      selected_view_schema_id: statusReviewViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    genericRowsByView[statusReviewViewSchemaId] = [
      statusReviewRow(
        "status-review-1",
        1,
        "Direct Status Review surface load",
      ),
    ];

    render(<WorkbookShell incidentId="incident-1" />);

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
            "status-review-1",
            "status_review.current_state_summary",
          ),
        )
      ).textContent,
    ).toBe("Direct Status Review surface load");
  });

  it("rejects generic query envelopes from a different view_schema_id", async () => {
    startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: statusReviewViewSchemaId },
      selected_view_schema_id: statusReviewViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    queryResponseOverride = (viewSchemaId) => {
      if (viewSchemaId !== statusReviewViewSchemaId) {
        return null;
      }
      return successEnvelope({
        incident_id: "incident-1",
        view_schema_id: evidenceViewSchemaId,
        rows: [
          statusReviewRow(
            "status-review-1",
            1,
            "Mismatched Status Review result",
          ),
        ],
      });
    };

    render(<WorkbookShell incidentId="incident-1" />);

    expect(
      (await screen.findByTestId("generic-surface-load-error")).textContent,
    ).toContain(
      `Surface load returned ${evidenceViewSchemaId} for ${statusReviewViewSchemaId}.`,
    );
    expect(
      screen.queryByTestId(
        rowCellTestId("status-review-1", "status_review.current_state_summary"),
      ),
    ).toBeNull();
  });

  it("keeps a selected saved-view sheet ref distinct from its base view_schema", async () => {
    startupSelection = {
      selected_sheet_ref: { kind: "saved_view", id: savedViewId },
      selected_view_schema_id: evidenceViewSchemaId,
      selected_saved_view: {
        saved_view_id: savedViewId,
        incident_id: "incident-1",
        view_schema_id: evidenceViewSchemaId,
      },
      source: "home",
    };

    render(<WorkbookShell incidentId="incident-1" />);

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
    savedViews = [
      {
        saved_view_id: savedViewId,
        view_schema_id: timelineViewSchemaId,
        display_name: "Timeline saved view",
      },
      {
        saved_view_id: evidenceSavedViewId,
        view_schema_id: evidenceViewSchemaId,
        display_name: "Evidence saved view",
      },
    ];

    render(<WorkbookShell incidentId="incident-1" />);

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

  it("passes invalid explicit base surfaces to backend startup fallback", async () => {
    window.history.replaceState(
      {},
      "",
      "/?view_schema_id=cartulary.view.unknown.v1",
    );
    startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
      selected_view_schema_id: timelineViewSchemaId,
      selected_saved_view: null,
      source: "timeline",
    };

    render(<WorkbookShell incidentId="incident-1" />);

    await waitFor(() => {
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining(
          `/workbook-startup?view_schema_id=${encodeURIComponent(
            "cartulary.view.unknown.v1",
          )}`,
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
    evidenceRows = [evidenceRow("evidence-initial", 1, "initial")];
    queryResponseOverride = (viewSchemaId, init) => {
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
          incident_id: "incident-1",
          view_schema_id: evidenceViewSchemaId,
          rows: [evidenceRow("evidence-newer", 1, "newer")],
        });
      }
      return null;
    };

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );
    await expectRecordIds(evidenceViewSchemaId, ["evidence-initial"]);

    applyGenericFilter(evidenceViewSchemaId, "evidence.storage_ref", "older");
    await waitFor(() => {
      expect(staleEvidenceQueryStarted).toBe(true);
    });
    applyGenericFilter(evidenceViewSchemaId, "evidence.storage_ref", "newer");
    await expectRecordIds(evidenceViewSchemaId, ["evidence-newer"]);

    staleEvidenceQuery.resolve(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: evidenceViewSchemaId,
        rows: [evidenceRow("evidence-older", 1, "older")],
      }),
    );
    await flushMicrotasks();

    expect(currentRecordIds(evidenceViewSchemaId)).toEqual(["evidence-newer"]);
  });

  it("keeps party-link mutations syncing until workbook and references refresh", async () => {
    const linkedTask = taskRequestRow(
      "task-1",
      4,
      "Task requester link",
      "Requester raw",
      "party-1",
    );
    const clearedTask = taskRequestRow(
      "task-1",
      5,
      "Task requester link",
      "Requester raw",
      null,
    );
    startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: taskRequestsViewSchemaId },
      selected_view_schema_id: taskRequestsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    genericRowsByView[taskRequestsViewSchemaId] = [linkedTask];
    genericRowsByView[partiesViewSchemaId] = [
      partyRow("party-1", "Requester Party"),
    ];
    const patchResponse = deferred<Response>();
    const refreshResponse = deferred<Response>();
    let patchAccepted = false;
    let refreshStarted = false;
    let clearPatchBody: Record<string, unknown> | null = null;
    recordPatchResponseOverride = (recordId, init) => {
      if (recordId !== "task-1") {
        return null;
      }
      clearPatchBody = JSON.parse(String(init?.body ?? "{}")) as Record<
        string,
        unknown
      >;
      return patchResponse.promise;
    };
    queryResponseOverride = (viewSchemaId) => {
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

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.change(
      await screen.findByTestId(
        genericEditRecordSelectTestId(taskRequestsViewSchemaId),
      ),
      { target: { value: "task-1" } },
    );
    const clearButton = await screen.findByTestId("party-link-clear-link");
    fireEvent.click(clearButton);

    await waitFor(() => {
      expect(clearPatchBody).toMatchObject({
        view_schema_id: taskRequestsViewSchemaId,
        base_row_version: 4,
        changes: [{ field_key: "task.requester_party_id", value: null }],
      });
    });
    expect(screen.getByTestId("generic-mutation-state").textContent).toBe(
      "Syncing",
    );
    expect((clearButton as HTMLButtonElement).disabled).toBe(true);

    patchAccepted = true;
    genericRowsByView[taskRequestsViewSchemaId] = [clearedTask];
    patchResponse.resolve(
      successEnvelope({
        view_schema_id: taskRequestsViewSchemaId,
        change_set_id: "change-task-clear-link",
        row: clearedTask,
      }),
    );
    await waitFor(() => {
      expect(refreshStarted).toBe(true);
    });
    expect(screen.getByTestId("generic-mutation-state").textContent).toBe(
      "Syncing",
    );
    expect((clearButton as HTMLButtonElement).disabled).toBe(true);

    refreshResponse.resolve(
      successEnvelope({
        incident_id: "incident-1",
        view_schema_id: taskRequestsViewSchemaId,
        rows: [clearedTask],
      }),
    );

    await waitFor(() => {
      expect(screen.getByTestId("generic-mutation-state").textContent).toBe(
        "Saved",
      );
    });
    expect((clearButton as HTMLButtonElement).disabled).toBe(false);
    expect(currentRecordIds(taskRequestsViewSchemaId)).toEqual(["task-1"]);
  });

  it("keeps failed generic party-link mutations in Conflict", async () => {
    startupSelection = {
      selected_sheet_ref: { kind: "view_schema", id: taskRequestsViewSchemaId },
      selected_view_schema_id: taskRequestsViewSchemaId,
      selected_saved_view: null,
      source: "explicit",
    };
    genericRowsByView[taskRequestsViewSchemaId] = [
      taskRequestRow(
        "task-1",
        4,
        "Task requester conflict",
        "Requester raw",
        "party-1",
      ),
    ];
    let clearPatchBody: Record<string, unknown> | null = null;
    recordPatchResponseOverride = (recordId, init) => {
      if (recordId !== "task-1") {
        return null;
      }
      clearPatchBody = JSON.parse(String(init?.body ?? "{}")) as Record<
        string,
        unknown
      >;
      return errorEnvelope("row_version_conflict", 409);
    };

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.change(
      await screen.findByTestId(
        genericEditRecordSelectTestId(taskRequestsViewSchemaId),
      ),
      { target: { value: "task-1" } },
    );
    fireEvent.click(await screen.findByTestId("party-link-clear-link"));

    await waitFor(() => {
      expect(clearPatchBody).toMatchObject({
        view_schema_id: taskRequestsViewSchemaId,
        base_row_version: 4,
        changes: [{ field_key: "task.requester_party_id", value: null }],
      });
      expect(screen.getByTestId("generic-mutation-state").textContent).toBe(
        "Conflict",
      );
    });
  });

  it("Phase 4 U-4-WB-03 issues opaque evidence preview and download handles from the evidence surface", async () => {
    evidenceRows = [
      {
        record_id: "evidence-1",
        row_version: 4,
        cells: {
          "evidence.title": { value: "EDR package" },
          "evidence.lifecycle_state": { value: "available" },
          "evidence.requested_at": { value: null },
          "evidence.received_at": { value: null },
          "evidence.storage_ref": { value: "slot" },
          "evidence.blob_hash": { value: "sha" },
          "evidence.collector_party_text": { value: "IR" },
          "evidence.source_party_text": { value: "Endpoint" },
          "evidence.upload_state": { value: "available" },
          "evidence.linked_record_count": { value: 0 },
          "evidence.edited_at": { value: null },
        },
      },
    ];
    const anchorClick = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => undefined);

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );
    fireEvent.click(
      await screen.findByTestId(evidencePreviewButtonTestId("evidence-1")),
    );

    const frame = await screen.findByTestId(
      evidencePreviewFrameTestId("evidence-1"),
    );
    expect(frame.getAttribute("src")).toBe(
      "/api/v1/evidence-handles/preview-token",
    );
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining(
        "/api/v1/evidence-records/evidence-1/preview-handle",
      ),
      expect.objectContaining({ method: "POST", body: "{}" }),
    );

    fireEvent.click(screen.getByTestId("evidence-download-evidence-1"));

    await waitFor(() => {
      expect(anchorClick).toHaveBeenCalled();
    });
    expect(fetchMock).toHaveBeenCalledWith(
      expect.stringContaining(
        "/api/v1/evidence-records/evidence-1/download-handle",
      ),
      expect.objectContaining({ method: "POST", body: "{}" }),
    );
  });

  it("FE-I-P6-01 Verify attach flow uses generated protocol types, public error envelopes, and stable evidence selectors without raw object URLs or paths.", async () => {
    evidenceRows = [
      evidenceStateRow("evidence-attach", 4, "Attach target", {
        lifecycleState: "available",
        uploadState: "available",
      }),
      evidenceStateRow("evidence-blocked", 5, "Blocked target", {
        lifecycleState: "quarantined",
        uploadState: "available",
      }),
      evidenceStateRow("evidence-failed", 6, "Failed target", {
        lifecycleState: "available",
        uploadState: "failed",
      }),
      evidenceStateRow("evidence-inconsistent", 7, "Inconsistent target", {
        lifecycleState: "available",
        uploadState: "storage-backend-mismatch",
      }),
      evidenceStateRow("evidence-raw-handle", 8, "Raw handle target", {
        lifecycleState: "available",
        uploadState: "available",
      }),
      evidenceStateRow("evidence-public-error", 9, "Public error target", {
        lifecycleState: "available",
        uploadState: "available",
      }),
    ];
    handleHrefByRecordID["evidence-raw-handle"] = {
      preview:
        "https://minio.internal/cartulary-evidence-bucket/object_blob_storage_key_v1",
    };
    handleErrorByRecordID["evidence-public-error"] =
      rawStorageErrorEnvelope();
    attachErrorByRecordID["evidence-public-error"] =
      rawStorageErrorEnvelope();
    const anchorClick = vi
      .spyOn(HTMLAnchorElement.prototype, "click")
      .mockImplementation(() => undefined);

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(evidenceViewSchemaId)),
    );

    for (const recordId of [
      "evidence-attach",
      "evidence-blocked",
      "evidence-failed",
      "evidence-inconsistent",
      "evidence-raw-handle",
      "evidence-public-error",
    ]) {
      const attachInput = await screen.findByTestId(
        evidenceAttachFileInputTestId(recordId),
      );
      expect(
        attachInput.getAttribute("data-testid"),
      ).toBe(evidenceAttachFileInputTestId(recordId));
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
          evidencePreviewButtonTestId("evidence-blocked"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      (
        screen.getByTestId(
          evidenceDownloadButtonTestId("evidence-failed"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      (
        screen.getByTestId(
          evidencePreviewButtonTestId("evidence-inconsistent"),
        ) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      screen.getByTestId(evidenceAccessMessageTestId("evidence-blocked"))
        .textContent,
    ).toContain("Blocked:");
    expect(
      screen.getByTestId(evidenceAccessMessageTestId("evidence-failed"))
        .textContent,
    ).toContain("Failed:");
    expect(
      screen.getByTestId(evidenceAccessMessageTestId("evidence-inconsistent"))
        .textContent,
    ).toContain("Inconsistent:");

    fireEvent.change(
      screen.getByTestId(evidenceAttachFileInputTestId("evidence-attach")),
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
        screen.getByTestId(evidenceAccessMessageTestId("evidence-attach"))
          .textContent,
      ).toBe("Evidence attached.");
    });
    const createBlobCall = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith("/api/v1/object-blobs"),
    );
    expect(createBlobCall).toBeDefined();
    expect(JSON.parse(String((createBlobCall?.[1] as RequestInit).body))).toEqual(
      {
        incident_id: "incident-1",
        client_txn_id: expect.stringMatching(/^evidence-blob-/u),
        byte_size: 18,
        filename_hint: "safe-evidence.txt",
        content_type_hint: "text/plain",
      },
    );
    const attachCall = fetchMock.mock.calls.find(([input]) =>
      String(input).endsWith(
        "/api/v1/evidence-records/evidence-attach/attach-blob",
      ),
    );
    expect(attachCall).toBeDefined();
    expect(JSON.parse(String((attachCall?.[1] as RequestInit).body))).toEqual({
      object_blob_id: "blob-created",
      base_row_version: 4,
      client_txn_id: expect.stringMatching(/^evidence-attach-/u),
    });

    fireEvent.change(
      screen.getByTestId(evidenceAttachFileInputTestId("evidence-public-error")),
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
        screen.getByTestId(evidenceAccessMessageTestId("evidence-public-error"))
          .textContent,
      ).toBe("Conflict.");
    });

    fireEvent.click(
      screen.getByTestId(evidencePreviewButtonTestId("evidence-raw-handle")),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(evidenceAccessMessageTestId("evidence-raw-handle"))
          .textContent,
      ).toBe("Evidence handle is unavailable.");
    });
    expect(
      screen.queryByTestId(evidencePreviewFrameTestId("evidence-raw-handle")),
    ).toBeNull();

    fireEvent.click(
      screen.getByTestId(evidencePreviewButtonTestId("evidence-public-error")),
    );
    await waitFor(() => {
      expect(
        screen.getByTestId(evidenceAccessMessageTestId("evidence-public-error"))
          .textContent,
      ).toBe("Conflict.");
    });

    fireEvent.click(
      screen.getByTestId(evidencePreviewButtonTestId("evidence-attach")),
    );
    const frame = await screen.findByTestId(
      evidencePreviewFrameTestId("evidence-attach"),
    );
    expect(frame.getAttribute("src")).toBe(
      "/api/v1/evidence-handles/preview-token",
    );
    fireEvent.click(
      screen.getByTestId(evidenceDownloadButtonTestId("evidence-attach")),
    );
    await waitFor(() => {
      expect(anchorClick).toHaveBeenCalled();
    });
    expectNoRawStorageDetails(document.body);
  });

  it("Phase 5 E-5-01 orchestrates selected Timeline evidence attachment inline", async () => {
    timelineRows = [timelineRow("timeline-1", 1, "Selected row", 0)];

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(timelineViewSchemaId)),
    );
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId("timeline-1")),
    );
    const input = await screen.findByTestId(
      "timeline-evidence-file-timeline-1",
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
          "/api/v1/evidence-records/evidence-created/attach-blob",
        ),
        expect.objectContaining({ method: "POST" }),
      );
      expect(fetchMock).toHaveBeenCalledWith(
        expect.stringContaining("/api/v1/records/timeline-1"),
        expect.objectContaining({ method: "PATCH" }),
      );
    });
    const uploadCall = fetchMock.mock.calls.find(([url]) =>
      String(url).includes("/api/v1/object-uploads/test-token"),
    );
    expect(uploadCall).toBeDefined();
    const uploadInit = uploadCall?.[1] as RequestInit;
    expect(uploadInit.credentials).toBe("omit");
    const uploadHeaders = uploadInit.headers as Headers;
    expect(uploadHeaders.get("X-Upload-Contract")).toBe("phase5");
    expect(uploadHeaders.get("Content-Type")).toBe("text/plain");
    expect(
      (await screen.findByTestId("timeline-inspector-message")).textContent,
    ).toBe("Evidence attached.");
  });

  it("Phase 5 E-5-01 surfaces upload failures inline without issuing Timeline patches", async () => {
    uploadShouldFail = true;
    timelineRows = [timelineRow("timeline-1", 1, "Selected row", 0)];

    render(<WorkbookShell incidentId="incident-1" />);

    fireEvent.click(
      await screen.findByTestId(surfaceTabTestId(timelineViewSchemaId)),
    );
    fireEvent.click(
      await screen.findByTestId(rowInspectButtonTestId("timeline-1")),
    );
    fireEvent.change(
      await screen.findByTestId("timeline-evidence-file-timeline-1"),
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
      expect(screen.getByTestId("timeline-inspector-message").textContent).toBe(
        "upload_failed_500",
      );
    });
    expect(
      fetchMock.mock.calls.some(([input, init]) => {
        return (
          String(input).endsWith("/api/v1/records/timeline-1") &&
          ((init as RequestInit | undefined)?.method ?? "GET") === "PATCH"
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
) {
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.occurred_at": { value: "" },
      "timeline.summary": { value: summary },
      "timeline.details": { value: "" },
      "timeline.source_text": { value: "" },
      "timeline.host_refs": {
        value: { kind: "collection_value_v1", ordered: true, items: [] },
      },
      "timeline.identity_refs": {
        value: { kind: "collection_value_v1", ordered: true, items: [] },
      },
      "timeline.evidence_count": { value: evidenceCount },
      "timeline.tags": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "timeline.edited_at": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.recorded_at": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.sort_ts": { value: "2026-04-24T10:00:00.000Z" },
      "timeline.capture_state": { value: "rough" },
      "timeline.replacement_record_id": { value: null },
      "timeline.occurred_day": { value: null },
      "timeline.recorded_day": { value: "2026-04-24" },
      "timeline.has_evidence": { value: evidenceCount > 0 },
      "timeline.attached_evidence_ids": {
        value: { kind: "collection_value_v1", ordered: false, items: [] },
      },
      "timeline.has_unresolved_mentions": { value: false },
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
      "evidence.source_party_text": { value: "" },
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
          raw_path: "/var/lib/cartulary/object-blobs/object_blob_storage_key_v1",
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
      "status_review.timestamp_day": { value: "2026-04-24" },
      "status_review.next_report_day": { value: null },
      "status_review.edited_at": { value: "2026-04-24T15:00:00.000Z" },
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
      "task.status": { value: "open" },
      "task.requester_party_text": { value: requesterText },
      "task.requester_party_id": { value: requesterPartyId },
      "task.owner_user_id": { value: null },
      "task.decision_record_id": { value: null },
      "task.due_at": { value: null },
      "task.priority": { value: "normal" },
      "task.external_ticket_ref": { value: null },
      "task.blocked_reason": { value: null },
      "task.edited_at": { value: "2026-04-24T15:00:00.000Z" },
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
      "party.primary_email": { value: null },
      "party.primary_phone": { value: null },
      "party.external_ref": { value: null },
      "party.notes": { value: null },
      "party.edited_at": { value: "2026-04-24T15:00:00.000Z" },
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
  fireEvent.change(screen.getByTestId(gridFilterFieldTestId(surface)), {
    target: { value: fieldKey },
  });
  fireEvent.change(screen.getByTestId(gridFilterValueTestId(surface)), {
    target: { value },
  });
  fireEvent.click(screen.getByTestId(gridFilterApplyTestId(surface)));
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

describe("generic workbook mutation payloads", () => {
  it("Phase 4 U-4-WB-04 builds required creates with direct values, timestamps, and explicit clears", () => {
    const evidence = requireViewContract(evidenceViewSchemaId);

    expect(
      buildGenericCreatePayload(evidence, {}, "txn-evidence-missing"),
    ).toBeNull();
    expect(
      buildGenericCreatePayload(
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

  it("Phase 4 U-4-WB-05 builds direct clears and typed collection actions", () => {
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
        "party-1",
      ),
    ).toEqual({
      field_key: "comm_log.audience_party_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_party_ref", party_id: "party-1" }],
      },
    });
    expect(
      buildGenericPatchChange(
        requireField(decisions, "decision.support_refs"),
        "record-1",
      ),
    ).toEqual({
      field_key: "decision.support_refs",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [{ op: "add_record_ref", linked_record_id: "record-1" }],
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
