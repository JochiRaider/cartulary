import {
  gridFilterFieldTestId,
  gridSavedRowsSelector,
  gridShellTestId,
  rowCellTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import {
  cleanup,
  fireEvent,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { expect, vi } from "vitest";

import { requireJSONBodyAt } from "./fetchMockTestSupport";
import type { RecordChangedPayload } from "./workbookShellPhase4";

export const timelineViewSchemaId = "cartulary.view.timeline.v1";

type WebSocketLike = {
  onmessage: ((event: MessageEvent) => void) | null;
};

export type TimelineWorkbookFetchMock = ReturnType<typeof vi.fn>;
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
  return {
    record_id: recordId,
    row_version: rowVersion,
    cells: {
      "timeline.occurred_at": { value: occurredAt },
      "timeline.summary": { value: summary },
      "timeline.details": { value: details },
      "timeline.source_text": { value: sourceText },
      "timeline.host_refs": { value: collectionValue(true, hostRefs) },
      "timeline.identity_refs": { value: collectionValue(true, identityRefs) },
      "timeline.evidence_count": { value: evidenceCount },
      "timeline.tags": { value: collectionValue(false, tags.map(tagItem)) },
      "timeline.edited_at": { value: editedAt },
      "timeline.recorded_at": { value: "" },
      "timeline.sort_ts": { value: occurredAt },
      "timeline.capture_state": { value: captureState },
      "timeline.replacement_record_id": { value: null },
      "timeline.occurred_day": { value: occurredAt.slice(0, 10) },
      "timeline.recorded_day": { value: "" },
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
  return `[data-testid="${CSS.escape(testId)}"]`;
}

function positiveIntegerEnv(name: string, fallback: number) {
  const raw = process.env[name];
  if (raw === undefined || raw.trim() === "") {
    return fallback;
  }
  const parsed = Number(raw);
  return Number.isInteger(parsed) && parsed > 0 ? parsed : fallback;
}

const workbookAsyncTimeoutMs = positiveIntegerEnv(
  "CARTULARY_TEST_ASYNC_TIMEOUT_MS",
  3000,
);

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
  surface: WorkbookSurface = "timeline",
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
  await screen.findByTestId(gridFilterFieldTestId(surface), undefined, {
    timeout: workbookAsyncTimeoutMs,
  });
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
    surface: "timeline",
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
  await new Promise((resolve) => window.setTimeout(resolve, 0));
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
  await new Promise((resolve) => window.setTimeout(resolve, 0));
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
  return {
    kind: "collection_value_v1",
    ordered,
    items,
  };
}

function tagItem(value: string) {
  return {
    item_ref: `record_tag:${value}`,
    item_kind: "tag",
    display_text: value,
    raw_text: value,
  };
}
