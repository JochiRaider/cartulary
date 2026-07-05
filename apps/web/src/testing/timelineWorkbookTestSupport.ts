import {
  dataTestIdSelector,
  gridSavedRowsSelector,
  gridShellTestId,
  pendingQueueCountTestId,
  pendingQueueNoticeTestId,
  rowCellTestId,
  saveStateTestId,
  type WorkbookSurface,
  workbookTopBarQueryControlsTestId,
} from "@cartulary/ui-contracts";
import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  cleanup,
  fireEvent,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, vi } from "vitest";
import type { RecordChangedPayload } from "../workbook/timeline/services/workbookCollaborationMessages";
import { requireJSONBodyAt } from "./fetchMockTestSupport";

export const timelineViewSchemaId = "cartulary.view.timeline.v2";

type WebSocketLike = {
  onmessage: ((event: MessageEvent) => void) | null;
};

export type TimelineWorkbookFetchMock = ReturnType<typeof vi.fn>;
type TimelineFetchResponse = Response | Promise<Response>;
export type TimelineRecordActionName = "mark-reviewed" | "supersede";
type TimelineRouteName =
  | "authSession"
  | "conflictResolution"
  | "recordAction"
  | "recordPatch"
  | "rowQuery"
  | "fallback";

type TimelineRoutedFetchQueues = Record<
  TimelineRouteName,
  TimelineFetchResponse[]
>;

export type TimelineWorkbookRouteFetch = {
  mockAuthSessionOnce: (response: TimelineFetchResponse) => void;
  mockConflictResolutionOnce: (response: TimelineFetchResponse) => void;
  mockRecordActionOnce: (response: TimelineFetchResponse) => void;
  mockRecordPatchOnce: (response: TimelineFetchResponse) => void;
  mockRowQueryOnce: (response: TimelineFetchResponse) => void;
  mockFallbackOnce: (response: TimelineFetchResponse) => void;
};

export type TimelineWebSocketMock = {
  close: () => void;
  emit: (message: Record<string, unknown>) => void;
  onclose: ((event: CloseEvent) => void) | null;
  onerror: ((event: Event) => void) | null;
  onmessage: ((event: MessageEvent) => void) | null;
  onopen: ((event: Event) => void) | null;
  readyState: number;
  send: (message: string) => void;
  sentMessages: string[];
};

const timelineWebSockets: TimelineWebSocketMock[] = [];

type TimelineRowOptions = {
  recordId: string;
  rowVersion: number;
  dateEnteredText?: string;
  analystText?: string;
  mitreStageText?: string;
  deviceObjectText?: string;
  ipAddressText?: string;
  activityUTCText?: string;
  activityLocalText?: string;
  rawActivityText?: string;
  activitySynopsisText?: string;
  dataSourceText?: string;
  occurredAt?: string;
  summary?: string;
  details?: string;
  sourceText?: string;
  captureState: string;
  evidenceCount?: number;
  tags?: string[];
  editedAt?: string;
  hasEvidence?: boolean;
  hostRefs?: Array<Record<string, unknown>>;
  identityRefs?: Array<Record<string, unknown>>;
};

export type WorkbookViewApiRow = {
  record_id: string;
  row_version: number;
  cells: Record<string, { value: unknown }>;
  view_schema_id?: string;
};

export type WorkbookViewRowOverrides = Record<string, unknown>;

type RecordChangedPayloadOptions = {
  recordId: string;
  rowVersion: number;
  clientTxnId: string;
  changeSetId?: string;
  actorUserId?: string;
  changedFieldKeys?: string[];
  affectedViews?: Array<{
    patch_cells?: {
      record_id: string;
      row_version: number;
      cells: Record<string, { value: unknown }>;
      group_values?: Record<string, unknown>;
    };
    view_schema_id: string;
    change_kind: string;
  }>;
};

export function deferred<T>() {
  let resolve!: (value: T | PromiseLike<T>) => void;
  let reject!: (reason?: unknown) => void;

  const promise = new Promise<T>((innerResolve, innerReject) => {
    resolve = innerResolve;
    reject = innerReject;
  });

  return { promise, resolve, reject };
}

export function installTimelineWorkbookTestGlobals(): TimelineWorkbookFetchMock {
  const fetchMock = vi.fn();
  timelineWebSockets.splice(0);
  vi.stubGlobal("fetch", fetchMock);
  vi.stubGlobal(
    "WebSocket",
    class {
      static readonly OPEN = 1;
      onclose: ((event: CloseEvent) => void) | null = null;
      onerror: ((event: Event) => void) | null = null;
      onmessage: ((event: MessageEvent) => void) | null = null;
      onopen: ((event: Event) => void) | null = null;
      readyState = 1;
      sentMessages: string[] = [];

      constructor() {
        timelineWebSockets.push(this as TimelineWebSocketMock);
        window.setTimeout(() => {
          this.onopen?.(new Event("open"));
        }, 0);
      }

      close() {
        this.readyState = 3;
        this.onclose?.(new CloseEvent("close"));
      }

      emit(message: Record<string, unknown>) {
        this.onmessage?.(
          new MessageEvent("message", {
            data: JSON.stringify(message),
          }),
        );
      }

      send(message: string) {
        this.sentMessages.push(message);
      }
    } as unknown as typeof WebSocket,
  );
  return fetchMock;
}

function fetchMethod(init: RequestInit | undefined) {
  return String(init?.method ?? "GET").toUpperCase();
}

function fetchURLText(input: RequestInfo | URL) {
  return input instanceof Request ? input.url : String(input);
}

function timelineRecordActionFromURL(
  urlText: string,
): TimelineRecordActionName | null {
  if (urlText.endsWith("/mark-reviewed")) {
    return "mark-reviewed";
  }
  if (urlText.endsWith("/supersede")) {
    return "supersede";
  }
  return null;
}

function timelineFetchRoute(
  input: RequestInfo | URL,
  init: RequestInit | undefined,
): TimelineRouteName {
  const url = fetchURLText(input);
  const method = fetchMethod(init);
  if (method === "GET" && url.endsWith("/api/v1/auth/session")) {
    return "authSession";
  }
  if (
    method === "POST" &&
    url.includes("/api/v1/incidents/") &&
    url.includes(`/views/${timelineViewSchemaId}/query`)
  ) {
    return "rowQuery";
  }
  if (
    method === "POST" &&
    url.includes("/api/v1/records/") &&
    url.includes("/conflicts/") &&
    url.endsWith("/resolve")
  ) {
    return "conflictResolution";
  }
  if (
    method === "POST" &&
    url.includes("/api/v1/records/") &&
    timelineRecordActionFromURL(url) !== null
  ) {
    return "recordAction";
  }
  if (method === "PATCH" && url.includes("/api/v1/records/")) {
    return "recordPatch";
  }
  return "fallback";
}

function queuedResponseDiagnostic(queues: TimelineRoutedFetchQueues) {
  return Object.entries(queues)
    .map(([name, queue]) => `${name}=${queue.length}`)
    .join(" ");
}

export function routeTimelineWorkbookFetchMock(
  fetchMock: TimelineWorkbookFetchMock,
): TimelineWorkbookRouteFetch {
  const queues: TimelineRoutedFetchQueues = {
    authSession: [],
    conflictResolution: [],
    recordAction: [],
    recordPatch: [],
    rowQuery: [],
    fallback: [],
  };

  fetchMock.mockImplementation(
    (input: RequestInfo | URL, init?: RequestInit) => {
      const route = timelineFetchRoute(input, init);
      const response = queues[route].shift();
      if (response !== undefined) {
        return response;
      }
      if (route === "authSession") {
        return errorEnvelope("session_required", 401);
      }
      const method = fetchMethod(init);
      throw new Error(
        `missing routed Timeline workbook fetch response route=${route} method=${method} url=${fetchURLText(
          input,
        )} queues=${queuedResponseDiagnostic(queues)}`,
      );
    },
  );

  return {
    mockAuthSessionOnce: (response) => {
      queues.authSession.push(response);
    },
    mockConflictResolutionOnce: (response) => {
      queues.conflictResolution.push(response);
    },
    mockRecordActionOnce: (response) => {
      queues.recordAction.push(response);
    },
    mockRecordPatchOnce: (response) => {
      queues.recordPatch.push(response);
    },
    mockRowQueryOnce: (response) => {
      queues.rowQuery.push(response);
    },
    mockFallbackOnce: (response) => {
      queues.fallback.push(response);
    },
  };
}

export function cleanupTimelineWorkbookTestGlobals() {
  cleanup();
  timelineWebSockets.splice(0);
  vi.unstubAllGlobals();
}

export function latestTimelineWebSocket(): TimelineWebSocketMock | null {
  return timelineWebSockets[timelineWebSockets.length - 1] ?? null;
}

export function successEnvelope(data: unknown, status = 200) {
  return new Response(
    JSON.stringify({
      data,
      meta: { request_id: `req-${status}` },
    }),
    {
      status,
      headers: { "Content-Type": "application/json" },
    },
  );
}

export function workbookCollectionValue(
  ordered: boolean,
  items: readonly Record<string, unknown>[] = [],
) {
  return {
    kind: "collection_value_v1",
    ordered,
    items: [...items],
  };
}

export function defaultWorkbookViewCellValue(
  field: ViewFieldContract,
): unknown {
  if (field.readKind === "collection") {
    return workbookCollectionValue(false);
  }
  if (field.readKind === "number") {
    return 0;
  }
  if (field.readKind === "boolean") {
    return false;
  }
  return null;
}

export function fullWorkbookViewRow(
  contract: ViewContract,
  recordId: string,
  rowVersion: number,
  overrides: WorkbookViewRowOverrides,
): WorkbookViewApiRow {
  return {
    record_id: recordId,
    row_version: rowVersion,
    view_schema_id: contract.viewSchemaId,
    cells: Object.fromEntries(
      contract.fields.map((field) => [
        field.fieldKey,
        {
          value:
            field.fieldKey in overrides
              ? overrides[field.fieldKey]
              : defaultWorkbookViewCellValue(field),
        },
      ]),
    ),
  };
}

export function viewRowsEnvelope(
  viewSchemaId: string,
  rows: readonly WorkbookViewApiRow[],
  incidentId = "incident-1",
) {
  return successEnvelope({
    incident_id: incidentId,
    view_schema_id: viewSchemaId,
    rows: [...rows],
  });
}

export function viewRowsEnvelopeForView(
  viewSchemaId: string,
  rowsByView: Record<string, readonly WorkbookViewApiRow[]>,
  incidentId = "incident-1",
) {
  return viewRowsEnvelope(
    viewSchemaId,
    rowsByView[viewSchemaId] ?? [],
    incidentId,
  );
}

export function timelineRowsEnvelope(
  rows: readonly ReturnType<typeof timelineRow>[],
  incidentId = "incident-1",
) {
  return successEnvelope({
    incident_id: incidentId,
    view_schema_id: timelineViewSchemaId,
    rows: [...rows],
  });
}

export function timelineMutationEnvelope(
  row: ReturnType<typeof timelineRow>,
  changeSetId: string,
) {
  return successEnvelope({
    view_schema_id: timelineViewSchemaId,
    change_set_id: changeSetId,
    row,
  });
}

export function errorEnvelope(
  code: string,
  status: number,
  conflict?: unknown,
) {
  return new Response(
    JSON.stringify({
      error: {
        status,
        code,
        message: code,
        request_id: "req-error",
        retryable: false,
        details: {},
        ...(conflict === undefined ? {} : { conflict }),
      },
    }),
    {
      status,
      headers: { "Content-Type": "application/json" },
    },
  );
}

export function buildRecordChangedPayload({
  recordId,
  rowVersion,
  clientTxnId,
  changeSetId = `change-set-${rowVersion}`,
  actorUserId = "user-1",
  changedFieldKeys = ["timeline.host_refs"],
  affectedViews = [
    {
      view_schema_id: timelineViewSchemaId,
      change_kind: "invalidate",
    },
  ],
}: RecordChangedPayloadOptions): RecordChangedPayload {
  return {
    record_id: recordId,
    row_version: rowVersion,
    change_set_id: changeSetId,
    client_txn_id: clientTxnId,
    actor_user_id: actorUserId,
    changed_field_keys: changedFieldKeys,
    affected_views: affectedViews,
  };
}

export function emitRecordChanged(
  socket: WebSocketLike | null | undefined,
  payload: RecordChangedPayload,
) {
  socket?.onmessage?.(
    new MessageEvent("message", {
      data: JSON.stringify({
        type: "record_changed",
        payload,
      }),
    }),
  );
}

export function timelineRow({
  recordId,
  rowVersion,
  dateEnteredText = "",
  analystText = "",
  mitreStageText = "",
  deviceObjectText = "",
  ipAddressText = "",
  activityUTCText,
  activityLocalText = "",
  rawActivityText,
  activitySynopsisText,
  dataSourceText = "",
  occurredAt = "",
  summary = "",
  details = "",
  sourceText = "",
  captureState,
  evidenceCount = 0,
  tags = [],
  editedAt = "",
  hasEvidence = false,
  hostRefs = [],
  identityRefs = [],
}: TimelineRowOptions) {
  const utcText = activityUTCText ?? occurredAt;
  const rawText = rawActivityText ?? (sourceText !== "" ? sourceText : details);
  const synopsisText = activitySynopsisText ?? summary;
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.date_entered_text": { value: dateEnteredText },
      "timeline.analyst_text": { value: analystText },
      "timeline.mitre_stage_text": { value: mitreStageText },
      "timeline.device_object_text": { value: deviceObjectText },
      "timeline.ip_address_text": { value: ipAddressText },
      "timeline.activity_utc_text": { value: utcText },
      "timeline.activity_local_text": { value: activityLocalText },
      "timeline.raw_activity_text": { value: rawText },
      "timeline.activity_synopsis_text": { value: synopsisText },
      "timeline.data_source_text": { value: dataSourceText },
      "timeline.host_refs": { value: collectionValue(true, hostRefs) },
      "timeline.identity_refs": { value: collectionValue(true, identityRefs) },
      "timeline.evidence_count": { value: evidenceCount },
      "timeline.tags": { value: collectionValue(false, tags.map(tagItem)) },
      "timeline.attached_evidence_ids": { value: collectionValue(false, []) },
      "timeline.edited_at": { value: editedAt },
      "timeline.recorded_at": { value: "" },
      "timeline.activity_sort_ts": { value: utcText },
      "timeline.date_entered_sort_day": {
        value: dateEnteredText === "" ? null : dateEnteredText.slice(0, 10),
      },
      "timeline.activity_time_pair_state": { value: "disabled" },
      "timeline.capture_state": { value: captureState },
      "timeline.replacement_record_id": { value: null },
      "timeline.has_evidence": { value: hasEvidence },
      "timeline.has_unresolved_mentions": { value: false },
    },
  };
}

export function visibleGridRows(container: HTMLElement): HTMLDivElement[] {
  return Array.from(
    container.querySelectorAll<HTMLDivElement>(gridSavedRowsSelector()),
  );
}

function testIdSelector(testId: string) {
  return dataTestIdSelector(testId);
}

function positiveIntegerEnv(name: string, fallback: number) {
  const raw = process.env[name];
  if (raw === undefined || raw.trim() === "") {
    return fallback;
  }
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

export const workbookAsyncTimeoutMs = positiveIntegerEnv(
  "CARTULARY_TEST_ASYNC_TIMEOUT_MS",
  3000,
);

export async function flushWorkbookAsync() {
  await new Promise((resolve) => window.setTimeout(resolve, 0));
}

function workbookGridScope(
  container: HTMLElement,
  surface: WorkbookSurface,
): HTMLElement {
  return (
    container.querySelector<HTMLElement>(
      testIdSelector(gridShellTestId(surface)),
    ) ?? container
  );
}

function workbookRowDiagnostic(options: {
  container: HTMLElement;
  expectedRecordIds?: string[] | undefined;
  expectedVisibleRows?: number | undefined;
  surface: WorkbookSurface;
}) {
  const grid = workbookGridScope(options.container, options.surface);
  const mountedRecordIds = visibleGridRows(grid)
    .map((row) => row.getAttribute("data-grid-record-id") ?? "")
    .filter(Boolean);
  const expected =
    options.expectedRecordIds === undefined
      ? `count=${options.expectedVisibleRows ?? "(any)"}`
      : `record_ids=${options.expectedRecordIds.join(",") || "(none)"}`;
  return [
    `Expected workbook rows for surface ${options.surface}: ${expected}.`,
    `Mounted record_ids=${mountedRecordIds.join(",") || "(none)"}.`,
    `Grid selector=${gridShellTestId(options.surface)} row_selector=${gridSavedRowsSelector()}.`,
  ].join(" ");
}

export function visibleGridRowRecordIds(
  container: HTMLElement,
  surface?: WorkbookSurface,
): string[] {
  const scope = surface ? workbookGridScope(container, surface) : container;
  return visibleGridRows(scope).map(
    (row) => row.getAttribute("data-grid-record-id") ?? "",
  );
}

export function requiredGridRow(
  container: HTMLElement,
  index: number,
): HTMLDivElement {
  const rows = visibleGridRows(container);
  const row = rows[index];
  if (!row) {
    throw new Error(
      `Expected visible grid row at index ${index}, but found ${rows.length}.`,
    );
  }
  return row;
}

export async function waitForVisibleGridRowRecordIds(
  container: HTMLElement,
  expectedRecordIds: string[],
  surface: WorkbookSurface = timelineViewSchemaId,
) {
  await waitFor(
    () => {
      expect(visibleGridRowRecordIds(container, surface)).toEqual(
        expectedRecordIds,
      );
    },
    {
      onTimeout: (error) =>
        new Error(
          `${error.message}\n${workbookRowDiagnostic({
            container,
            expectedRecordIds,
            surface,
          })}`,
        ),
      timeout: workbookAsyncTimeoutMs,
    },
  );
}

export async function waitForWorkbookRows({
  container,
  expectedRecordIds,
  expectedVisibleRows,
  surface,
}: {
  container: HTMLElement;
  expectedRecordIds?: string[] | undefined;
  expectedVisibleRows?: number | undefined;
  surface: WorkbookSurface;
}) {
  if (expectedRecordIds === undefined && expectedVisibleRows === undefined) {
    throw new Error(
      "waitForWorkbookRows requires expectedRecordIds or expectedVisibleRows.",
    );
  }
  const grid = await screen.findByTestId(gridShellTestId(surface), undefined, {
    timeout: workbookAsyncTimeoutMs,
  });
  await screen.findByTestId(
    workbookTopBarQueryControlsTestId(surface),
    undefined,
    {
      timeout: workbookAsyncTimeoutMs,
    },
  );
  await waitFor(
    () => {
      if (expectedRecordIds !== undefined) {
        expect(visibleGridRowRecordIds(container, surface)).toEqual(
          expectedRecordIds,
        );
        return;
      }
      expect(
        visibleGridRows(workbookGridScope(container, surface)),
      ).toHaveLength(expectedVisibleRows ?? 0);
    },
    {
      onTimeout: (error) =>
        new Error(
          `${error.message}\n${workbookRowDiagnostic({
            container,
            expectedRecordIds,
            expectedVisibleRows,
            surface,
          })}`,
        ),
      timeout: workbookAsyncTimeoutMs,
    },
  );
  return grid;
}

// Initial workbook readiness only covers mounted controls and row count.
// Use row-identity waits for refreshes that preserve the same row count.
export async function waitForWorkbookReady({
  container,
  expectedVisibleRows,
  surface,
}: {
  container: HTMLElement;
  expectedVisibleRows: number;
  surface: WorkbookSurface;
}) {
  return waitForWorkbookRows({
    container,
    expectedVisibleRows,
    surface,
  });
}

export async function waitForTimelineWorkbookReady(
  container: HTMLElement,
  expectedVisibleRows: number,
) {
  return waitForWorkbookReady({
    container,
    expectedVisibleRows,
    surface: timelineViewSchemaId,
  });
}

export function gridScalarInput(
  container: HTMLElement,
  recordId: string,
  fieldKey: string,
  surface?: WorkbookSurface,
) {
  const row = Array.from(
    (surface
      ? workbookGridScope(container, surface)
      : container
    ).querySelectorAll<HTMLElement>(gridSavedRowsSelector()),
  ).find(
    (candidate) => candidate.getAttribute("data-grid-record-id") === recordId,
  );
  if (!row) {
    throw new Error(`Expected visible grid row for record ${recordId}.`);
  }
  return within(row).getByTestId(rowCellTestId(recordId, fieldKey));
}

function elementDiagnostic(element: Element | null) {
  if (!(element instanceof HTMLElement)) {
    return "(none)";
  }
  const parts = [element.tagName.toLowerCase()];
  const testId = element.getAttribute("data-testid");
  const role = element.getAttribute("role");
  if (element.id) {
    parts.push(`#${element.id}`);
  }
  if (testId) {
    parts.push(`data-testid=${testId}`);
  }
  if (role) {
    parts.push(`role=${role}`);
  }
  return parts.join(" ");
}

export async function focusReadyGridScalarInput({
  container,
  fieldKey,
  recordId,
  surface = timelineViewSchemaId,
}: {
  readonly container: HTMLElement;
  readonly fieldKey: string;
  readonly recordId: string;
  readonly surface?: WorkbookSurface | undefined;
}) {
  const input = gridScalarInput(container, recordId, fieldKey, surface);
  const expectedAnchor = `${surface}:${recordId}:${fieldKey}`;

  input.focus();
  await waitFor(
    () => {
      expect(document.activeElement).toBe(input);
      expect(screen.getByTestId("workbook-focus-anchor").textContent).toBe(
        expectedAnchor,
      );
    },
    {
      onTimeout: (error) => {
        const focusAnchor =
          screen.queryByTestId("workbook-focus-anchor")?.textContent ??
          "(missing)";
        return new Error(
          `${error.message}\nExpected ready focus anchor ${expectedAnchor}.\n` +
            `Actual focus anchor=${focusAnchor}.\n` +
            `Active element=${elementDiagnostic(document.activeElement)}.\n` +
            workbookRowDiagnostic({
              container,
              expectedRecordIds: [recordId],
              surface,
            }),
        );
      },
      timeout: workbookAsyncTimeoutMs,
    },
  );

  return input;
}

export async function findWorkbookCell(
  container: HTMLElement,
  surface: WorkbookSurface,
  recordId: string,
  fieldKey: string,
) {
  let cell: HTMLElement | null = null;
  await waitFor(
    () => {
      cell = gridScalarInput(container, recordId, fieldKey, surface);
    },
    {
      onTimeout: (error) =>
        new Error(
          `${error.message}\n${workbookRowDiagnostic({
            container,
            expectedRecordIds: [recordId],
            surface,
          })}`,
        ),
      timeout: workbookAsyncTimeoutMs,
    },
  );
  return cell ?? gridScalarInput(container, recordId, fieldKey, surface);
}

export async function typeInputValue(
  input: HTMLInputElement | HTMLTextAreaElement,
  value: string,
) {
  const user = userEvent.setup();
  await user.click(input);
  if (input.ownerDocument.activeElement === input) {
    input.setSelectionRange(0, input.value.length);
    await user.keyboard(value === "" ? "{Backspace}" : value);
  } else {
    fireEvent.change(input, { target: { value } });
  }
  await waitFor(() => {
    if (input.value !== value) {
      throw new Error(`Expected input value ${value}, got ${input.value}.`);
    }
  });
  await flushWorkbookAsync();
}

export async function changeInputValue(
  input: HTMLInputElement | HTMLTextAreaElement,
  value: string,
) {
  const user = userEvent.setup();
  await user.click(input);
  fireEvent.change(input, { target: { value } });
  await waitFor(
    () => {
      if (input.value !== value) {
        throw new Error(`Expected input value ${value}, got ${input.value}.`);
      }
    },
    {
      onTimeout: (error) =>
        new Error(
          `${error.message}\ncontrolled_input_replacement_mismatch expected=${JSON.stringify(
            value,
          )} actual=${JSON.stringify(input.value)}`,
        ),
      timeout: workbookAsyncTimeoutMs,
    },
  );
  await flushWorkbookAsync();
}

function controlledInputValueDiagnostic(
  input: HTMLInputElement | HTMLTextAreaElement,
  value: string,
  testId: string | null,
) {
  const currentInput =
    testId === null
      ? input
      : (screen.queryByTestId(testId) as
          | HTMLInputElement
          | HTMLTextAreaElement
          | null);
  const actual = currentInput?.value ?? "(missing)";
  return [
    "controlled_input_replacement_mismatch",
    `test_id=${testId ?? "(none)"}`,
    `expected=${JSON.stringify(value)}`,
    `actual=${JSON.stringify(actual)}`,
  ].join(" ");
}

export async function changeQueuedCellValue(
  input: HTMLInputElement | HTMLTextAreaElement,
  value: string,
) {
  const testId = input.getAttribute("data-testid");
  fireEvent.focus(input);
  fireEvent.change(input, { target: { value } });
  await waitFor(
    () => {
      const currentInput =
        testId === null
          ? input
          : (screen.getByTestId(testId) as
              | HTMLInputElement
              | HTMLTextAreaElement);
      if (currentInput.value !== value) {
        throw new Error(
          `Expected controlled input value ${value}, got ${currentInput.value}.`,
        );
      }
    },
    {
      onTimeout: (error) =>
        new Error(
          `${error.message}\n${controlledInputValueDiagnostic(
            input,
            value,
            testId,
          )}`,
        ),
      timeout: workbookAsyncTimeoutMs,
    },
  );
}

function pendingQueueDiagnostic(options: {
  expectedPendingUnits?: number;
  expectedSaveState?: "Conflict" | "Saved" | "Syncing";
  noticeIncludes?: string;
}) {
  const saveState =
    screen.queryByTestId(saveStateTestId())?.textContent ?? "(missing)";
  const notice =
    screen.queryByTestId(pendingQueueNoticeTestId())?.textContent ??
    "(missing)";
  const count =
    screen.queryByTestId(pendingQueueCountTestId())?.textContent ?? "(missing)";
  return [
    "Expected pending queue state.",
    `expected_save_state=${options.expectedSaveState ?? "(any)"}`,
    `expected_pending_units=${options.expectedPendingUnits ?? "(any)"}`,
    `expected_notice=${JSON.stringify(options.noticeIncludes ?? "(any)")}`,
    `actual_save_state=${JSON.stringify(saveState)}`,
    `actual_count=${JSON.stringify(count)}`,
    `actual_notice=${JSON.stringify(notice)}`,
  ].join(" ");
}

export async function waitForPendingQueueState(options: {
  expectedPendingUnits?: number;
  expectedSaveState?: "Conflict" | "Saved" | "Syncing";
  noticeIncludes?: string;
}) {
  await waitFor(
    () => {
      if (options.expectedSaveState !== undefined) {
        expect(screen.getByTestId(saveStateTestId()).textContent).toBe(
          options.expectedSaveState,
        );
      }
      if (options.noticeIncludes !== undefined) {
        expect(
          screen.getByTestId(pendingQueueNoticeTestId()).textContent,
        ).toContain(options.noticeIncludes);
      }
      if (options.expectedPendingUnits !== undefined) {
        expect(
          screen.getByTestId(pendingQueueCountTestId()).textContent,
        ).toContain(String(options.expectedPendingUnits));
      }
    },
    {
      onTimeout: (error) =>
        new Error(`${error.message}\n${pendingQueueDiagnostic(options)}`),
      timeout: workbookAsyncTimeoutMs,
    },
  );
}

export function timelineRecordPatchCalls(fetchSpy: TimelineWorkbookFetchMock) {
  return fetchSpy.mock.calls.filter(([url, init]) => {
    const method =
      init && typeof init === "object" && "method" in init
        ? String((init as { method?: unknown }).method ?? "GET").toUpperCase()
        : "GET";
    return (
      fetchURLText(url as RequestInfo | URL).includes("/api/v1/records/") &&
      method === "PATCH"
    );
  });
}

export function timelineRecordActionCalls(
  fetchSpy: TimelineWorkbookFetchMock,
  action?: TimelineRecordActionName,
) {
  return fetchSpy.mock.calls.filter(([url, init]) => {
    const method =
      init && typeof init === "object" && "method" in init
        ? String((init as { method?: unknown }).method ?? "GET").toUpperCase()
        : "GET";
    const urlText = fetchURLText(url as RequestInfo | URL);
    const callAction = timelineRecordActionFromURL(urlText);
    return (
      method === "POST" &&
      urlText.includes("/api/v1/records/") &&
      callAction !== null &&
      (action === undefined || callAction === action)
    );
  });
}

function timelineConflictResolutionCalls(fetchSpy: TimelineWorkbookFetchMock) {
  return fetchSpy.mock.calls.filter(([url, init]) => {
    const method =
      init && typeof init === "object" && "method" in init
        ? String((init as { method?: unknown }).method ?? "GET").toUpperCase()
        : "GET";
    const urlText = fetchURLText(url as RequestInfo | URL);
    return (
      method === "POST" &&
      urlText.includes("/api/v1/records/") &&
      urlText.includes("/conflicts/") &&
      urlText.endsWith("/resolve")
    );
  });
}

function timelineFetchCallDiagnostic(fetchSpy: TimelineWorkbookFetchMock) {
  return fetchSpy.mock.calls
    .map(([url, init], index) => {
      const method =
        init && typeof init === "object" && "method" in init
          ? String((init as { method?: unknown }).method ?? "")
          : "(none)";
      return `${index}:${method}:${String(url)}`;
    })
    .join(" | ");
}

export function timelineRecordPatchCallURLs(
  fetchSpy: TimelineWorkbookFetchMock,
) {
  return timelineRecordPatchCalls(fetchSpy).map(([url]) => {
    const urlText = fetchURLText(url as RequestInfo | URL);
    return new URL(urlText, "http://cartulary.test").pathname;
  });
}

export async function waitForTimelineRecordPatchCalls(
  fetchSpy: TimelineWorkbookFetchMock,
  expectedCount: number,
) {
  await waitFor(
    () => {
      expect(timelineRecordPatchCalls(fetchSpy)).toHaveLength(expectedCount);
    },
    {
      onTimeout: (error) => {
        return new Error(
          `${error.message}\nExpected ${expectedCount} Timeline record PATCH calls. ` +
            `Actual=${timelineRecordPatchCalls(fetchSpy).length}. Calls=${timelineFetchCallDiagnostic(
              fetchSpy,
            )}`,
        );
      },
      timeout: workbookAsyncTimeoutMs,
    },
  );
}

export async function waitForTimelineConflictResolutionCalls(
  fetchSpy: TimelineWorkbookFetchMock,
  expectedCount: number,
) {
  await waitFor(
    () => {
      expect(timelineConflictResolutionCalls(fetchSpy)).toHaveLength(
        expectedCount,
      );
    },
    {
      onTimeout: (error) => {
        return new Error(
          `${error.message}\nExpected ${expectedCount} Timeline conflict-resolution POST calls. ` +
            `Actual=${timelineConflictResolutionCalls(fetchSpy).length}. Calls=${timelineFetchCallDiagnostic(
              fetchSpy,
            )}`,
        );
      },
      timeout: workbookAsyncTimeoutMs,
    },
  );
}

export async function waitForTimelineRecordActionCalls(
  fetchSpy: TimelineWorkbookFetchMock,
  expectedCount: number,
): Promise<void>;
export async function waitForTimelineRecordActionCalls(
  fetchSpy: TimelineWorkbookFetchMock,
  action: TimelineRecordActionName,
  expectedCount: number,
): Promise<void>;
export async function waitForTimelineRecordActionCalls(
  fetchSpy: TimelineWorkbookFetchMock,
  actionOrExpectedCount: TimelineRecordActionName | number,
  maybeExpectedCount?: number,
) {
  const action =
    typeof actionOrExpectedCount === "number"
      ? undefined
      : actionOrExpectedCount;
  const expectedCount =
    typeof actionOrExpectedCount === "number"
      ? actionOrExpectedCount
      : maybeExpectedCount;
  if (expectedCount === undefined) {
    throw new Error("expected Timeline record action count is required");
  }

  await waitFor(
    () => {
      expect(timelineRecordActionCalls(fetchSpy, action)).toHaveLength(
        expectedCount,
      );
    },
    {
      onTimeout: (error) => {
        const routeLabel = action === undefined ? "action" : `action:${action}`;
        return new Error(
          `${error.message}\nExpected ${expectedCount} Timeline record ${routeLabel} calls. ` +
            `Actual=${timelineRecordActionCalls(fetchSpy, action).length}. Calls=${timelineFetchCallDiagnostic(
              fetchSpy,
            )}`,
        );
      },
      timeout: workbookAsyncTimeoutMs,
    },
  );
}

export function setInputValueWithoutEvent(
  input: HTMLInputElement | HTMLTextAreaElement,
  value: string,
) {
  const prototype =
    input instanceof HTMLTextAreaElement
      ? HTMLTextAreaElement.prototype
      : HTMLInputElement.prototype;
  const valueSetter = Object.getOwnPropertyDescriptor(prototype, "value")?.set;
  valueSetter?.call(input, value);
}

export function extractTimelinePatchBody(
  fetchSpy: TimelineWorkbookFetchMock,
  index: number,
) {
  return requireJSONBodyAt<{
    base_row_version: number;
    changes: Array<{ field_key: string; value: string | null }>;
  }>(fetchSpy, index, `timeline request body at index ${index}`);
}

export function extractTimelineRecordPatchBody(
  fetchSpy: TimelineWorkbookFetchMock,
  index: number,
) {
  const call = timelineRecordPatchCalls(fetchSpy)[index];
  if (!call) {
    throw new Error(
      `missing Timeline record PATCH request at index ${index}; found ${
        timelineRecordPatchCalls(fetchSpy).length
      }`,
    );
  }
  return requireJSONBodyAt<{
    base_row_version: number;
    changes: Array<{ field_key: string; value: string | null }>;
  }>(
    { mock: { calls: [call] } },
    0,
    `timeline record PATCH request at index ${index}`,
  );
}

export function extractTimelineRecordActionBody(
  fetchSpy: TimelineWorkbookFetchMock,
  index?: number,
): Record<string, unknown>;
export function extractTimelineRecordActionBody(
  fetchSpy: TimelineWorkbookFetchMock,
  action: TimelineRecordActionName,
  index?: number,
): Record<string, unknown>;
export function extractTimelineRecordActionBody(
  fetchSpy: TimelineWorkbookFetchMock,
  actionOrIndex: TimelineRecordActionName | number = 0,
  maybeIndex = 0,
) {
  const action = typeof actionOrIndex === "number" ? undefined : actionOrIndex;
  const index = typeof actionOrIndex === "number" ? actionOrIndex : maybeIndex;
  const calls = timelineRecordActionCalls(fetchSpy, action);
  const call = calls[index];
  if (!call) {
    const routeLabel = action === undefined ? "action" : `action ${action}`;
    throw new Error(
      `missing Timeline record ${routeLabel} request at index ${index}; found ${calls.length}`,
    );
  }
  return requireJSONBodyAt<Record<string, unknown>>(
    { mock: { calls: [call] } },
    0,
    `timeline record action request at index ${index}`,
  );
}

export function extractTimelineConflictResolutionBody(
  fetchSpy: TimelineWorkbookFetchMock,
  index: number,
) {
  const call = timelineConflictResolutionCalls(fetchSpy)[index];
  if (!call) {
    throw new Error(
      `missing Timeline conflict resolution request at index ${index}; found ${
        timelineConflictResolutionCalls(fetchSpy).length
      }`,
    );
  }
  return requireJSONBodyAt<Record<string, unknown>>(
    { mock: { calls: [call] } },
    0,
    `timeline conflict resolution request at index ${index}`,
  );
}

export function extractTimelineJSONBody(
  fetchSpy: TimelineWorkbookFetchMock,
  index: number,
) {
  return requireJSONBodyAt<Record<string, unknown>>(
    fetchSpy,
    index,
    `timeline JSON request body at index ${index}`,
  );
}

function collectionValue(
  ordered: boolean,
  items: Array<Record<string, unknown>>,
) {
  return workbookCollectionValue(ordered, items);
}

function tagItem(value: string) {
  return {
    item_ref: `record_tag:${value}`,
    item_kind: "tag",
    display_text: value,
    raw_text: value,
  };
}
