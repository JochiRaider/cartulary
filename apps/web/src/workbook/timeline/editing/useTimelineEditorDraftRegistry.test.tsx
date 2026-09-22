import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { WorkbookLocalDraftStore } from "../../models/WorkbookLocalDraftStore";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { useTimelineMutationCommands } from "../hooks/useTimelineMutationCommands";
import { TimelineCaptureLifecycle } from "../models/TimelineCaptureLifecycle";
import { timelinePendingSavesRefsFor } from "../models/timelinePendingSaves";
import {
  normalizeTimelineFullRow,
  rowFromApi,
} from "../models/timelineRowModel";
import {
  createTimelineEditorDraftRegistry,
  useTimelineEditorDraftRegistry,
} from "./useTimelineEditorDraftRegistry";

const timelineContract = requireViewContract(timelineViewSchemaId);
const recordId = "11111111-1111-4111-8111-111111111111";

function committedRow() {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timelineContract, recordId, 3, {
        "timeline.activity_synopsis_text": "authoritative value",
      }),
      "editor draft registry fixture",
    ),
  );
}

describe("Timeline editor draft registry", () => {
  it("exposes retained grid collection work for unavailable targets without empty cancellations or Inspector drafts", () => {
    const registry = createTimelineEditorDraftRegistry();
    const identity = {
      rowKey: recordId,
      field: "hostRefs" as const,
      surface: "grid" as const,
    };
    registry.setDraft(identity, "  retained Ω  ");
    registry.setDraft({ ...identity, field: "tags" }, "");
    registry.setDraft({ ...identity, surface: "inspector" }, "Inspector work");
    const drafts = registry.retainedGridDrafts();
    expect(drafts).toHaveLength(1);
    expect(drafts[0]).toMatchObject({
      rowKey: recordId,
      fieldKey: "timeline.host_refs",
      value: "  retained Ω  ",
    });
    drafts[0]?.discard();
    expect(registry.retainedGridDrafts()).toEqual([]);
    expect(registry.draftValue({ ...identity, surface: "inspector" })).toBe(
      "Inspector work",
    );
  });

  it("settles repeated collection callers once per surface revision and admits later identical tokens", () => {
    const row = committedRow();
    const registry = createTimelineEditorDraftRegistry();
    const runtime = new WorkbookMutationRuntime(
      { incidentId: "incident", clientInstanceId: "test" },
      { create: () => "id" },
      { execute: vi.fn() },
    );
    const enqueue =
      vi.fn<
        Parameters<
          typeof useTimelineMutationCommands
        >[0]["enqueuePendingReplayUnit"]
      >();
    const rowsRef = { current: [row] };
    let txn = 0;
    const hook = renderHook(() =>
      useTimelineMutationCommands({
        captureActionBlocksRecord: () => false,
        beginViewportContinuity: () => 1,
        clearViewportContinuity: vi.fn(),
        clientInstanceId: "test",
        conflictQueueRef: { current: {} },
        editorDraftRegistry: registry,
        enqueuePendingReplayUnit: enqueue,
        incidentId: "incident",
        latestCommittedTimelineRow: (id) => (id === row.recordId ? row : null),
        nextClientTxnId: () => String(++txn),
        pendingSavesRefs: timelinePendingSavesRefsFor(
          runtime,
          runtime.pendingQueue(),
        ),
        rowsRef,
        rowStoreCommands: {
          replaceRows: (rows) => {
            rowsRef.current = rows;
          },
          updateRows: (update) => {
            rowsRef.current = update(rowsRef.current);
          },
        },
      }),
    );
    const grid = {
      rowKey: row.key,
      field: "hostRefs" as const,
      surface: "grid" as const,
    };
    const inspector = { ...grid, surface: "inspector" as const };
    const first = vi.fn(),
      second = vi.fn(),
      other = vi.fn();
    registry.setDraft(grid, "raw Ω");
    const captured = registry.captureRow(
      row.key,
      "grid",
      new Set(["timeline.host_refs"]),
    );
    hook.result.current.commands.queueCollectionSave(
      row.key,
      "timeline.host_refs",
      "hostRefs",
      "raw Ω",
      "grid",
      first,
    );
    hook.result.current.commands.queueCollectionSave(
      row.key,
      "timeline.host_refs",
      "hostRefs",
      "raw Ω",
      "grid",
      second,
    );
    expect(enqueue).toHaveBeenCalledTimes(1);
    expect(first).not.toHaveBeenCalled();
    registry.setDraft(inspector, "raw Ω");
    hook.result.current.commands.queueCollectionSave(
      row.key,
      "timeline.host_refs",
      "hostRefs",
      "raw Ω",
      "inspector",
      other,
    );
    expect(enqueue).toHaveBeenCalledTimes(2);
    expect(enqueue.mock.calls[0]?.[0].mutationSignature).not.toBe(
      enqueue.mock.calls[1]?.[0].mutationSignature,
    );
    registry.setDraft(grid, "newer text");
    registry.settleRevisions(row.key, captured ?? new Map());
    enqueue.mock.calls[0]?.[1]?.({ kind: "accepted" });
    expect(first).toHaveBeenCalledExactlyOnceWith({ kind: "accepted" });
    expect(second).toHaveBeenCalledExactlyOnceWith({ kind: "accepted" });
    expect(other).not.toHaveBeenCalled();
    expect(registry.draftValue(grid)).toBe("newer text");
    expect(registry.draftValue(inspector)).toBe("raw Ω");
    registry.setDraft(grid, "raw Ω");
    hook.result.current.commands.queueCollectionSave(
      row.key,
      "timeline.host_refs",
      "hostRefs",
      "raw Ω",
      "grid",
      first,
    );
    expect(enqueue).toHaveBeenCalledTimes(3);
    expect(enqueue.mock.calls[0]?.[0].mutationSignature).not.toBe(
      enqueue.mock.calls[2]?.[0].mutationSignature,
    );
    enqueue.mock.calls[2]?.[1]?.({
      kind: "validation_error",
      message: "Rejected",
    });
    hook.result.current.commands.queueCollectionSave(
      row.key,
      "timeline.host_refs",
      "hostRefs",
      "raw Ω",
      "grid",
      second,
    );
    expect(enqueue).toHaveBeenCalledTimes(3);
    expect(second).toHaveBeenLastCalledWith({
      kind: "validation_error",
      message: "Rejected",
    });
    expect(registry.draftValue(grid)).toBe("raw Ω");
    const stale = {
      rowKey: "draft-retired",
      field: "activitySynopsisText" as const,
      surface: "grid" as const,
    };
    registry.setDraft(stale, "Unknown retained work");
    const staleOutcome = vi.fn();
    hook.result.current.commands.queueScalarSave(
      stale.rowKey,
      stale.field,
      {
        surface: "grid",
        continueOnFreshDraft: false,
        preserveInputFocus: false,
      },
      "Unknown retained work",
      staleOutcome,
    );
    expect(staleOutcome).toHaveBeenCalledWith(
      expect.objectContaining({ kind: "stale_target" }),
    );
    expect(enqueue).toHaveBeenCalledTimes(3);
    expect(registry.draftValue(stale)).toBe("Unknown retained work");
    registry.setDraft(grid, "refused tokens");
    const retryCollection = () =>
      hook.result.current.commands.queueCollectionSave(
        row.key,
        "timeline.host_refs",
        "hostRefs",
        "refused tokens",
        "grid",
        first,
      );
    retryCollection();
    expect(enqueue).toHaveBeenCalledTimes(4);
    enqueue.mock.calls[3]?.[2]?.();
    enqueue.mock.calls[3]?.[1]?.({
      kind: "rejected_mutation",
      message: "Queue capacity",
    });
    expect(registry.draftValue(grid)).toBe("refused tokens");
    retryCollection();
    expect(enqueue).toHaveBeenCalledTimes(5);
    expect(enqueue.mock.calls[4]?.[0].clientTxnId).not.toBe(
      enqueue.mock.calls[3]?.[0].clientTxnId,
    );
    hook.unmount();
  });

  it("joins scalar paste and departure by authoring revision while fencing identical newer edits", () => {
    const row = committedRow();
    const registry = createTimelineEditorDraftRegistry();
    const runtime = new WorkbookMutationRuntime(
      { incidentId: "incident", clientInstanceId: "test" },
      { create: () => "id" },
      { execute: vi.fn() },
    );
    const enqueue =
      vi.fn<
        Parameters<
          typeof useTimelineMutationCommands
        >[0]["enqueuePendingReplayUnit"]
      >();
    const rowsRef = { current: [row] };
    let txn = 0;
    const hook = renderHook(() =>
      useTimelineMutationCommands({
        captureActionBlocksRecord: () => false,
        beginViewportContinuity: () => 1,
        clearViewportContinuity: vi.fn(),
        clientInstanceId: "test",
        conflictQueueRef: { current: {} },
        editorDraftRegistry: registry,
        enqueuePendingReplayUnit: enqueue,
        incidentId: "incident",
        latestCommittedTimelineRow: (id) => (id === row.recordId ? row : null),
        nextClientTxnId: () => String(++txn),
        pendingSavesRefs: timelinePendingSavesRefsFor(
          runtime,
          runtime.pendingQueue(),
        ),
        rowsRef,
        rowStoreCommands: {
          replaceRows: (rows) => {
            rowsRef.current = rows;
          },
          updateRows: (update) => {
            rowsRef.current = update(rowsRef.current);
          },
        },
      }),
    );
    const grid = {
      rowKey: row.key,
      field: "activitySynopsisText" as const,
      surface: "grid" as const,
    };
    const inspector = { ...grid, surface: "inspector" as const };
    const first = vi.fn(),
      second = vi.fn(),
      other = vi.fn();
    registry.setDraft(grid, "raw Ω");
    const captured = registry.captureRow(
      row.key,
      "grid",
      new Set(["timeline.activity_synopsis_text"]),
    );
    hook.result.current.commands.queueScalarSave(
      row.key,
      "activitySynopsisText",
      {
        continueOnFreshDraft: false,
        preserveInputFocus: false,
        surface: "grid",
      },
      "raw Ω",
      first,
    );
    // An unsubmitted collection draft is not part of this scalar operation.
    registry.setDraft(
      { ...grid, field: "hostRefs" },
      "separate collection input",
    );
    hook.result.current.commands.queueScalarSave(
      row.key,
      "activitySynopsisText",
      {
        continueOnFreshDraft: false,
        preserveInputFocus: false,
        surface: "grid",
      },
      "raw Ω",
      second,
    );
    expect(enqueue).toHaveBeenCalledTimes(1);
    expect(first).not.toHaveBeenCalled();
    registry.setDraft(inspector, "raw Ω");
    hook.result.current.commands.queueScalarSave(
      row.key,
      "activitySynopsisText",
      {
        continueOnFreshDraft: false,
        preserveInputFocus: false,
        surface: "inspector",
      },
      "raw Ω",
      other,
    );
    expect(enqueue).toHaveBeenCalledTimes(2);
    expect(enqueue.mock.calls[0]?.[0].mutationSignature).not.toBe(
      enqueue.mock.calls[1]?.[0].mutationSignature,
    );
    registry.setDraft(grid, "raw Ω", row, true);
    registry.settleRevisions(row.key, captured ?? new Map());
    enqueue.mock.calls[0]?.[1]?.({ kind: "accepted" });
    expect(first).toHaveBeenCalledExactlyOnceWith({ kind: "accepted" });
    expect(second).toHaveBeenCalledExactlyOnceWith({ kind: "accepted" });
    expect(other).not.toHaveBeenCalled();
    expect(registry.draftValue(grid)).toBe("raw Ω");
    expect(registry.draftValue(inspector)).toBe("raw Ω");
    hook.result.current.commands.queueScalarSave(
      row.key,
      "activitySynopsisText",
      {
        continueOnFreshDraft: false,
        preserveInputFocus: false,
        surface: "grid",
      },
      "raw Ω",
      first,
    );
    expect(enqueue).toHaveBeenCalledTimes(3);
    expect(enqueue.mock.calls[0]?.[0].mutationSignature).not.toBe(
      enqueue.mock.calls[2]?.[0].mutationSignature,
    );
    enqueue.mock.calls[2]?.[1]?.({
      kind: "validation_error",
      message: "Rejected",
    });
    hook.result.current.commands.queueScalarSave(
      row.key,
      "activitySynopsisText",
      {
        continueOnFreshDraft: false,
        preserveInputFocus: false,
        surface: "grid",
      },
      "raw Ω",
      second,
    );
    expect(enqueue).toHaveBeenCalledTimes(3);
    expect(second).toHaveBeenLastCalledWith({
      kind: "validation_error",
      message: "Rejected",
    });
    expect(registry.draftValue(grid)).toBe("raw Ω");
    registry.setDraft(
      grid,
      row.committedValues.activitySynopsisText,
      row,
      true,
    );
    hook.result.current.commands.queueScalarSave(
      row.key,
      "activitySynopsisText",
      {
        continueOnFreshDraft: false,
        preserveInputFocus: false,
        surface: "grid",
      },
      row.committedValues.activitySynopsisText,
      second,
    );
    expect(enqueue).toHaveBeenCalledTimes(3);
    expect(second).toHaveBeenLastCalledWith({ kind: "accepted" });
    expect(registry.draftValue(grid)).toBeUndefined();
    expect(registry.draftValue({ ...grid, field: "hostRefs" })).toBe(
      "separate collection input",
    );
    registry.setDraft(grid, "refused scalar", row, true);
    const retryScalar = () =>
      hook.result.current.commands.queueScalarSave(
        row.key,
        "activitySynopsisText",
        {
          continueOnFreshDraft: false,
          preserveInputFocus: false,
          surface: "grid",
        },
        "refused scalar",
        first,
      );
    retryScalar();
    expect(enqueue).toHaveBeenCalledTimes(4);
    enqueue.mock.calls[3]?.[2]?.();
    enqueue.mock.calls[3]?.[1]?.({
      kind: "rejected_mutation",
      message: "Queue capacity",
    });
    expect(registry.draftValue(grid)).toBe("refused scalar");
    retryScalar();
    expect(enqueue).toHaveBeenCalledTimes(5);
    expect(enqueue.mock.calls[4]?.[0].clientTxnId).not.toBe(
      enqueue.mock.calls[3]?.[0].clientTxnId,
    );
    hook.unmount();
  });

  it("retains authoring baselines across remount and requires review only for relevant changes", () => {
    const store = new WorkbookLocalDraftStore();
    const original = createTimelineEditorDraftRegistry(store);
    const row = committedRow();
    const identity = {
      rowKey: recordId,
      field: "activitySynopsisText" as const,
      surface: "grid" as const,
    };
    original.setDraft(identity, "Local", row);
    const registry = createTimelineEditorDraftRegistry(store);
    const remote = {
      ...row,
      rowVersion: 4,
      committedValues: {
        ...row.committedValues,
        activitySynopsisText: "Remote",
      },
      values: { ...row.values, activitySynopsisText: "Remote" },
    };
    expect(registry.needsReview(identity, remote)).toBe(true);
    expect(
      registry.authoringRow(registry.materializeRow(remote), "grid")
        .committedValues.activitySynopsisText,
    ).toBe("authoritative value");
    const capture = registry.captureRow(recordId, "grid");
    registry.review(identity, remote);
    expect(registry.needsReview(identity, remote)).toBe(false);
    expect(registry.captureRow(recordId, "grid")).not.toEqual(capture);
    registry.acceptPredecessor(
      row,
      ["timeline.activity_synopsis_text"],
      row.committedValues,
    );
    expect(registry.needsReview(identity, remote)).toBe(false);
    expect(registry.needsReview(identity, row)).toBe(true);
    const ownAccepted = {
      ...remote,
      rowVersion: 5,
      committedValues: {
        ...remote.committedValues,
        activitySynopsisText: "Accepted predecessor",
      },
    };
    registry.acceptPredecessor(
      ownAccepted,
      ["timeline.activity_synopsis_text"],
      remote.committedValues,
    );
    expect(registry.needsReview(identity, ownAccepted)).toBe(false);
    expect(registry.draftValue(identity)).toBe("Local");
  });

  it("materializes inspector submissions without unrelated grid authoring", () => {
    const registry = createTimelineEditorDraftRegistry();
    const row = committedRow();
    registry.setDraft(
      { rowKey: recordId, field: "activitySynopsisText", surface: "grid" },
      "Grid only",
      row,
    );
    registry.setDraft(
      { rowKey: recordId, field: "analystText", surface: "inspector" },
      "Inspector only",
      row,
    );
    const grid = registry.materializeRow(row);
    const inspector = registry.materializeRow(grid, {
      field: "analystText",
      value: "Inspector only",
      surface: "inspector",
    });
    expect(inspector.values.activitySynopsisText).toBe(
      row.committedValues.activitySynopsisText,
    );
    expect(inspector.values.analystText).toBe("Inspector only");
    expect(grid.values.analystText).toBe(row.committedValues.analystText);
  });
  it("materializes invalid scalar text across authoritative row replacement", () => {
    const registry = createTimelineEditorDraftRegistry();
    registry.setDraft(
      {
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      },
      "invalid local text",
    );

    const refreshed = committedRow();
    expect(registry.materializeRow(refreshed).values.activitySynopsisText).toBe(
      "invalid local text",
    );
    expect(refreshed.values.activitySynopsisText).toBe("authoritative value");
    registry.beginCapture("draft-1");
    registry.setDraft(
      { field: "activitySynopsisText", rowKey: "draft-1", surface: "grid" },
      "typing after submission",
    );
    registry.capture.promote("draft-1", refreshed);
    expect(registry.resolveRowKey("draft-1")).toBe(recordId);
    const input = document.createElement("textarea");
    document.body.append(input);
    registry.registerInput(
      { field: "activitySynopsisText", rowKey: recordId, surface: "grid" },
      input,
    );
    expect(
      registry.inputElementForFocusKey("draft-1:activitySynopsisText:grid"),
    ).toBe(input);
    input.remove();
    expect(
      registry.inputElementForFocusKey("draft-1:activitySynopsisText:grid"),
    ).toBeNull();
    registry.retainRows(new Set([recordId]));
    expect(registry.materializeRow(refreshed).values.activitySynopsisText).toBe(
      "typing after submission",
    );
    registry.settleRevisions("draft-1", new Map());
    expect(registry.materializeRow(refreshed).values.activitySynopsisText).toBe(
      "typing after submission",
    );
    registry.clearAll();
    expect(registry.resolveRowKey("draft-1")).toBe("draft-1");
  });

  it("prefers an explicit commit value over a stale registered draft", () => {
    const registry = createTimelineEditorDraftRegistry();
    const row = committedRow();
    registry.setDraft(
      {
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      },
      "stale registered value",
    );

    const committed = registry.materializeRow(row, {
      field: "activitySynopsisText",
      value: "value accepted by the input event",
    });

    expect(committed.values.activitySynopsisText).toBe(
      "value accepted by the input event",
    );
  });

  it("keeps grid and inspector drafts distinct and clears only submitted values", () => {
    const registry = createTimelineEditorDraftRegistry();
    const submitted = committedRow().values;
    registry.setDraft(
      {
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      },
      submitted.activitySynopsisText,
    );
    registry.setDraft(
      {
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "inspector",
      },
      "newer inspector edit",
    );

    registry.settleRevisions(
      recordId,
      registry.captureRow(recordId, "grid") ?? new Map(),
    );

    expect(
      registry.draftValue({
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      }),
    ).toBeUndefined();
    expect(
      registry.draftValue({
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "inspector",
      }),
    ).toBe("newer inspector edit");
  });

  it("owns semantic editor registration, unregistration, and row removal", () => {
    const registry = createTimelineEditorDraftRegistry();
    const input = document.createElement("input");
    document.body.append(input);
    registry.registerInput(
      {
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      },
      input,
    );
    registry.setDraft(
      {
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      },
      "invalid local text",
    );
    const focusKey = `${recordId}:activitySynopsisText:grid`;

    expect(registry.inputElementForFocusKey(focusKey)).toBe(input);

    registry.registerInput(
      {
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      },
      null,
    );
    expect(registry.inputElementForFocusKey(focusKey)).toBeNull();
    expect(registry.draftValueForFocusKey(focusKey)).toBe("invalid local text");

    registry.retainRows(new Set());

    expect(registry.inputElementForFocusKey(focusKey)).toBeNull();
    expect(registry.draftValueForFocusKey(focusKey)).toBe("invalid local text");
  });

  it("retires only captured revisions in the originating editor context", () => {
    const registry = createTimelineEditorDraftRegistry();
    const identity = {
      rowKey: recordId,
      field: "activitySynopsisText" as const,
      surface: "grid" as const,
    };
    registry.setDraft(identity, "A");
    registry.setDraft({ ...identity, surface: "inspector" }, "A");
    const captured = registry.captureRow(recordId, "grid");
    registry.setDraft(identity, "B");
    registry.setDraft(identity, "A");
    registry.settleRevisions(recordId, captured ?? new Map());
    expect(
      registry.clearCapturedScalarField(recordId, identity.field, captured),
    ).toBe(false);
    expect(registry.draftValue(identity)).toBe("A");
    expect(registry.draftValue({ ...identity, surface: "inspector" })).toBe(
      "A",
    );
    registry.settleRevisions(
      recordId,
      registry.captureRow(recordId, "grid") ?? new Map(),
    );
    expect(registry.draftValue(identity)).toBeUndefined();
    expect(registry.draftValue({ ...identity, surface: "inspector" })).toBe(
      "A",
    );
  });

  it("rejects stale, hidden, disabled, and disconnected editor elements", () => {
    const registry = createTimelineEditorDraftRegistry();
    const identity = {
      field: "activitySynopsisText" as const,
      rowKey: recordId,
      surface: "grid" as const,
    };
    const focusKey = `${recordId}:activitySynopsisText:grid`;
    const input = document.createElement("input");
    document.body.append(input);
    registry.registerInput(identity, input);
    input.hidden = true;
    expect(registry.inputElementForFocusKey(focusKey)).toBeNull();

    input.hidden = false;
    input.disabled = true;
    registry.registerInput(identity, input);
    expect(registry.inputElementForFocusKey(focusKey)).toBeNull();

    input.disabled = false;
    registry.registerInput(identity, input);
    input.remove();
    expect(registry.inputElementForFocusKey(focusKey)).toBeNull();
  });

  it("retains collection text across refresh and clears only the submitted version", () => {
    const registry = createTimelineEditorDraftRegistry();
    const identity = {
      field: "tags" as const,
      rowKey: recordId,
      surface: "grid" as const,
    };
    registry.setDraft(identity, "  pending Ω  ");
    const submitted = registry.materializeRow(committedRow());
    expect(submitted.collectionDrafts.tags).toBe("  pending Ω  ");
    registry.settleRevisions(recordId, new Map());
    expect(registry.draftValue(identity)).toBe("  pending Ω  ");
    registry.settleRevisions(recordId, new Map());
    expect(registry.draftValue(identity)).toBe("  pending Ω  ");
    registry.setDraft(identity, "newer text");
    registry.settleRevisions(recordId, new Map());
    expect(registry.materializeRow(committedRow()).collectionDrafts.tags).toBe(
      "newer text",
    );
    const inspector = { ...identity, surface: "inspector" as const };
    registry.setDraft(inspector, "independent Inspector text");
    const revisions = registry.captureRow(
      recordId,
      "grid",
      new Set(["timeline.tags"]),
    );
    expect(
      registry.materializeRow(committedRow(), { surface: "inspector" })
        .collectionDrafts.tags,
    ).toBe("independent Inspector text");
    registry.settleRevisions(recordId, revisions ?? new Map());
    expect(registry.draftValue(identity)).toBeUndefined();
    expect(registry.draftValue(inspector)).toBe("independent Inspector text");
    registry.setDraft(identity, "new token");
    const captured = registry.captureRow(
      recordId,
      "grid",
      new Set(["timeline.tags"]),
    );
    registry.setDraft(identity, "new token", undefined, true);
    registry.settleRevisions(recordId, captured ?? new Map());
    expect(registry.draftValue(identity)).toBe("new token");
    expect(
      registry.captureRow(recordId, "grid", new Set(["timeline.analyst_text"]))
        .size,
    ).toBe(0);
  });

  it("invalidates drafts when the runtime lifetime changes", () => {
    const { result, rerender } = renderHook(
      ({ store }) =>
        useTimelineEditorDraftRegistry(
          store,
          new TimelineCaptureLifecycle(store),
        ),
      { initialProps: { store: new WorkbookLocalDraftStore() } },
    );
    act(() => {
      result.current.setDraft(
        {
          field: "activitySynopsisText",
          rowKey: recordId,
          surface: "grid",
        },
        "invalid local text",
      );
    });
    const originalRegistry = result.current;

    rerender({ store: new WorkbookLocalDraftStore() });

    expect(result.current).not.toBe(originalRegistry);
    expect(
      result.current.draftValue({
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      }),
    ).toBeUndefined();
  });
  it("retains refused local values across surface detachment without retaining input references", () => {
    const store = new WorkbookLocalDraftStore();
    const identity = {
      field: "activitySynopsisText",
      rowKey: recordId,
      surface: "grid",
    } as const;
    const first = createTimelineEditorDraftRegistry(store);
    first.setDraft(identity, "Refused local draft");
    const input = document.createElement("textarea");
    document.body.append(input);
    first.registerInput(identity, input);
    first.registerInput(identity, null);
    input.remove();
    const remounted = createTimelineEditorDraftRegistry(store);
    expect(remounted.draftValue(identity)).toBe("Refused local draft");
    expect(
      remounted.inputElementForFocusKey(
        `${recordId}:activitySynopsisText:grid`,
      ),
    ).toBeNull();
    remounted.deleteDraft(identity);
    expect(first.draftValue(identity)).toBeUndefined();
    first.setDraft(identity, "Another draft");
    store.clear();
    expect(remounted.draftValue(identity)).toBeUndefined();
  });
});
