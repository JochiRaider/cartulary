import { gridSavedRowsSelector, rowCellTestId } from "@cartulary/ui-contracts";
import { cleanup, fireEvent, waitFor, within } from "@testing-library/react";
import { vi } from "vitest";

import { requireJSONBodyAt } from "./fetchMockTestSupport";
import type { RecordChangedPayload } from "./workbookShellPhase4";

export const timelineViewSchemaId = "cartulary.view.timeline.v1";

type WebSocketLike = {
  onmessage: ((event: MessageEvent) => void) | null;
};

type FetchMock = ReturnType<typeof vi.fn>;

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

export function installTimelineWorkbookTestGlobals() {
  const fetchMock = vi.fn();
  vi.stubGlobal("fetch", fetchMock);
  vi.stubGlobal(
    "WebSocket",
    class {
      onmessage: ((event: MessageEvent) => void) | null = null;

      close() {}
    } as unknown as typeof WebSocket,
  );
  return fetchMock;
}

export function cleanupTimelineWorkbookTestGlobals() {
  cleanup();
  vi.unstubAllGlobals();
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

export function gridScalarInput(
  container: HTMLElement,
  recordId: string,
  fieldKey: string,
) {
  const row = Array.from(
    container.querySelectorAll<HTMLElement>(gridSavedRowsSelector()),
  ).find(
    (candidate) => candidate.getAttribute("data-grid-record-id") === recordId,
  );
  if (!row) {
    throw new Error(`Expected visible grid row for record ${recordId}.`);
  }
  return within(row).getByTestId(rowCellTestId(recordId, fieldKey));
}

export async function changeInputValue(
  input: HTMLInputElement | HTMLTextAreaElement,
  value: string,
) {
  setInputValueWithoutEvent(input, value);
  fireEvent.input(input, { bubbles: true });
  await waitFor(() => {
    if (input.value !== value) {
      throw new Error(`Expected input value ${value}, got ${input.value}.`);
    }
  });
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

export function extractTimelinePatchBody(fetchSpy: FetchMock, index: number) {
  return requireJSONBodyAt<{
    base_row_version: number;
    changes: Array<{ field_key: string; value: string | null }>;
  }>(fetchSpy, index, `timeline request body at index ${index}`);
}

export function extractTimelineJSONBody(fetchSpy: FetchMock, index: number) {
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
