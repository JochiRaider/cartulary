import { requireViewContract } from "@cartulary/view-contracts";
import { act, cleanup, renderHook } from "@testing-library/react";
import { useSyncExternalStore } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { deferred } from "../../testing/fetchMockTestSupport";
import { observationSource } from "../../testing/observationTestSupport";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { useTimelineObservationSource } from "./hooks/useTimelineObservationSource";
import type { TimelineCommittedRecordIdleResult } from "./models/timelineControllerPorts";
import { rowFromApi } from "./models/timelineRowModel";

afterEach(cleanup);
it("Observation source preparation preserves raw committed strings and rejects saved or local source changes", async () => {
  const raw = {
    ...fullWorkbookViewRow(
      requireViewContract(observationSource.viewSchemaId),
      observationSource.recordId,
      4,
      { [observationSource.fieldKey]: observationSource.text },
    ),
    view_schema_id: observationSource.viewSchemaId,
  };
  const row = rowFromApi(raw),
    rowsRef = { current: [row] },
    drafts = createTimelineEditorDraftRegistry();
  const runtime = new WorkbookMutationRuntime(
    { clientInstanceId: "test", incidentId: observationSource.incidentId },
    { create: () => "secure" },
    createWorkbookPendingMutationAdapter({
      apiBase: undefined,
      incidentId: observationSource.incidentId,
    }),
  );
  runtime.inspectorDrafts.setAuthority({
    incidentId: observationSource.incidentId,
    actorId: "actor",
    sessionIdentity: "session",
    role: "editor",
    closed: false,
  });
  const inspectorIdentity = {
    viewSchemaId: observationSource.viewSchemaId,
    recordId: raw.record_id,
    fieldKey: observationSource.fieldKey,
    action: "value",
  };
  const waitForIdle = vi.fn(
    async (): Promise<TimelineCommittedRecordIdleResult> => ({
      row,
      rowVersion: 4,
    }),
  );
  const options = {
    runtime,
    selectedRow: row,
    available: true,
    rowsRef,
    drafts,
    waitForIdle,
  };
  let renderedReady = false;
  const { result, rerender } = renderHook(
    (props) => {
      const port = useTimelineObservationSource(props);
      renderedReady = useSyncExternalStore(port.subscribe, port.ready);
      return port;
    },
    { initialProps: options },
  );
  expect(result.current.source(observationSource.fieldKey)).toEqual(
    observationSource,
  );
  expect(
    await result.current.prepare(
      observationSource,
      new AbortController().signal,
    ),
  ).toBe(true);
  act(() =>
    runtime.inspectorDrafts.update(
      inspectorIdentity,
      raw,
      "unsaved text",
      "test-attachment",
    ),
  );
  expect(renderedReady).toBe(false);
  expect(result.current.ready()).toBe(false);
  expect(
    await result.current.prepare(
      observationSource,
      new AbortController().signal,
    ),
  ).toBe(false);
  expect(waitForIdle).toHaveBeenCalledTimes(1);
  expect(raw.cells[observationSource.fieldKey]?.value).toBe(
    observationSource.text,
  );
  act(() => runtime.inspectorDrafts.discard(inspectorIdentity));
  expect(renderedReady).toBe(true);
  act(() =>
    drafts.setDraft(
      { rowKey: row.key, field: "rawActivityText", surface: "grid" },
      "independent grid text",
    ),
  );
  expect(renderedReady).toBe(false);
  expect(
    await result.current.prepare(
      observationSource,
      new AbortController().signal,
    ),
  ).toBe(false);
  expect(waitForIdle).toHaveBeenCalledTimes(1);
  act(() => drafts.clearAll());
  expect(renderedReady).toBe(true);
  act(() =>
    drafts.setDraft(
      { rowKey: row.key, field: "hostRefs", surface: "inspector" },
      "raw independent host?",
    ),
  );
  expect(renderedReady).toBe(false);
  expect(
    await result.current.prepare(
      observationSource,
      new AbortController().signal,
    ),
  ).toBe(false);
  act(() => drafts.clearAll());
  expect(renderedReady).toBe(true);
  const pending = deferred<TimelineCommittedRecordIdleResult>();
  waitForIdle.mockReturnValue(pending.promise);
  const run = result.current.prepare(
    observationSource,
    new AbortController().signal,
  );
  const saved = rowFromApi({
    ...raw,
    row_version: 5,
    cells: {
      ...raw.cells,
      [observationSource.fieldKey]: { value: "new saved text" },
    },
  });
  rowsRef.current = [saved];
  rerender({ ...options, selectedRow: saved });
  pending.resolve({ row: saved, rowVersion: 5 });
  expect(await run).toBe(false);
  expect(result.current.source(observationSource.fieldKey)?.text).toBe(
    "new saved text",
  );
  rerender({ ...options, available: false });
  expect(result.current.source(observationSource.fieldKey)).toBeNull();
});
