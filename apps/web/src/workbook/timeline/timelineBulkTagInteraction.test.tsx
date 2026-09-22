import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import { WorkbookBatchOperationOwner } from "../runtime/WorkbookBatchOperationOwner";
import { createTimelineBulkTagCommandAdapter } from "./adapters/createTimelineBulkTagCommandAdapter";
import { createTimelineBulkTagReadiness } from "./bulk/createTimelineBulkTagReadiness";
import { useTimelineBulkTagController } from "./bulk/useTimelineBulkTagController";
import { TimelineBulkTagControl } from "./components/TimelineBulkTagControl";
import { createTimelineEditorDraftRegistry } from "./editing/useTimelineEditorDraftRegistry";
import { timelinePendingSavesRefsFor } from "./models/timelinePendingSaves";
import {
  createDraftRow,
  normalizeTimelineFullRow,
  rowFromApi,
  type WorkbookRow,
} from "./models/timelineRowModel";

const contract = requireViewContract(timelineViewSchemaId);
function row(id: string, version = 3): WorkbookRow {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(contract, id, version, {
        "timeline.activity_synopsis_text": id,
      }),
      "bulk fixture",
    ),
  );
}
function fixture() {
  let id = 0;
  const owner = new WorkbookBatchOperationOwner(
    "incident",
    { create: () => `batch-${++id}` },
    {
      pending: () => [],
      sealPending: () => {},
      available: () => true,
      reserve: () => () => {},
      conflicts: () => false,
      accepted: () => {},
      refresh: async () => {},
    },
  );
  const send = vi.fn(async () => ({
    kind: "acknowledged" as const,
    receipt: {
      viewSchemaId: timelineViewSchemaId,
      changeSetId: "change",
      rows: [],
      conflicts: [],
    },
  }));
  owner.configure({
    capture: (plan, authority, id) => ({
      plan,
      authority,
      id,
      apiBase: undefined,
      path: "/batch",
      body: JSON.stringify(plan.request),
    }),
    send,
  });
  owner.setAuthority({
    actorId: "actor",
    incidentId: "incident",
    sessionIdentity: "session",
    role: "editor",
    closed: false,
  });
  const port = createTimelineBulkTagCommandAdapter(owner);
  const rowsRef = {
    current: [row("first"), row("second", 4), createDraftRow(1)],
  };
  const readiness = {
    subscribe: () => () => {},
    blockingReason: vi.fn<() => string | null>(() => null),
  };
  return { owner, send, port, rowsRef, readiness };
}

describe("Timeline bulk tag interaction", () => {
  it("retains pending members and explicit deselection while pruning departed and unauthorized records", async () => {
    const f = fixture();
    const { result, rerender } = renderHook(
      ({ authorized, rows }) =>
        useTimelineBulkTagController({
          context: { authorized, capabilityAvailable: true },
          port: f.port,
          readiness: f.readiness,
          precedingSaves: async () => {},
          rows,
          rowsRef: f.rowsRef,
        }),
      { initialProps: { authorized: true, rows: f.rowsRef.current } },
    );
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first", "second"]),
      ),
    );
    f.rowsRef.current = [
      { ...row("first"), pendingSignature: "save" },
      row("second"),
      row("appended"),
    ];
    rerender({ authorized: true, rows: f.rowsRef.current });
    expect([...result.current.controls.selectedRecordIds]).toEqual([
      "first",
      "second",
    ]);
    expect(
      result.current.snapshot.gridSelection.isRecordSelectable?.({
        kind: "data",
        data: f.rowsRef.current[0] as WorkbookRow,
        rowIdentity: { kind: "core_record", recordId: "first" },
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 3 },
      }),
    ).toBe(true);
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["second"]),
      ),
    );
    f.rowsRef.current = [row("first"), row("second", 5)];
    rerender({ authorized: true, rows: f.rowsRef.current });
    expect([...result.current.controls.selectedRecordIds]).toEqual(["second"]);
    f.rowsRef.current = [row("first")];
    rerender({ authorized: true, rows: f.rowsRef.current });
    await waitFor(() =>
      expect(result.current.controls.selectedRecordIds.size).toBe(0),
    );
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first"]),
      ),
    );
    rerender({ authorized: false, rows: f.rowsRef.current });
    await waitFor(() =>
      expect(result.current.controls.selectedRecordIds.size).toBe(0),
    );
  });

  it("captures the complete set once per delivery and waits for prerequisites without retargeting", async () => {
    const f = fixture();
    let resolve = () => {};
    const ready = new Promise<void>((done) => {
      resolve = done;
    });
    const { result } = renderHook(() =>
      useTimelineBulkTagController({
        context: { authorized: true, capabilityAvailable: true },
        port: f.port,
        readiness: f.readiness,
        precedingSaves: () => ready,
        rows: f.rowsRef.current,
        rowsRef: f.rowsRef,
      }),
    );
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first", "second"]),
      ),
    );
    const delivery = {};
    act(() => {
      result.current.controls.assignTag("  triaged  ", delivery);
      result.current.controls.assignTag("  triaged  ", delivery);
    });
    expect(f.owner.getSnapshot().entries).toHaveLength(1);
    expect(f.owner.getSnapshot().entries[0]?.plan.request).toMatchObject({
      tag_name: "triaged",
      targets: [
        { record_id: "first", base_row_version: 3 },
        { record_id: "second", base_row_version: 4 },
      ],
    });
    expect(f.send).not.toHaveBeenCalled();
    act(() => result.current.controls.assignTag("  triaged  ", {}));
    expect(f.owner.getSnapshot().entries).toHaveLength(2);
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["second"]),
      ),
    );
    await act(async () => {
      resolve();
    });
    await waitFor(() => expect(f.send).toHaveBeenCalledTimes(2));
    expect(f.send.mock.calls[0]).toBeDefined();
  });

  it("surfaces real admission refusal and rejects incomplete or blocked intended targets", () => {
    const f = fixture();
    const { result } = renderHook(() =>
      useTimelineBulkTagController({
        context: { authorized: true, capabilityAvailable: true },
        port: f.port,
        readiness: f.readiness,
        precedingSaves: async () => {},
        rows: f.rowsRef.current,
        rowsRef: f.rowsRef,
      }),
    );
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first", "missing"]),
      ),
    );
    expect(result.current.controls.assignTag("raw", {})).toMatchObject({
      kind: "rejected",
      message: expect.stringContaining("Selection changed"),
    });
    expect(f.owner.getSnapshot().entries).toHaveLength(0);
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first"]),
      ),
    );
    f.readiness.blockingReason.mockReturnValue("Review the failed edit.");
    expect(result.current.controls.assignTag("raw", {})).toEqual({
      kind: "rejected",
      message: "Review the failed edit.",
    });
    f.readiness.blockingReason.mockReturnValue(null);
    f.owner.closeIncident();
    expect(result.current.controls.assignTag("raw", {})).toMatchObject({
      kind: "rejected",
      message: expect.stringContaining("editing access"),
    });
    expect(f.send).not.toHaveBeenCalled();
    // Read actual draft revisions, including a newer edit after queue admission.
    const execute = vi.fn();
    const runtime = createWorkbookMutationRuntime(
      { incidentId: "incident", clientInstanceId: "test" },
      { create: () => "id" },
      { execute },
    );
    const drafts = createTimelineEditorDraftRegistry();
    const pending = timelinePendingSavesRefsFor(
      runtime,
      runtime.pendingQueue(),
    );
    const selectedRow = row("first");
    const readiness = createTimelineBulkTagReadiness({
      runtime,
      drafts,
      pending,
      rows: { current: [selectedRow, row("second")] },
    });
    const selection = new Set(["first"]);
    const identity = {
      rowKey: selectedRow.key,
      field: "activitySynopsisText",
      surface: "grid",
    } as const;
    drafts.setDraft(identity, "captured scalar");
    expect(readiness.blockingReason(selection)).toContain("unsaved edit");
    const revisions = drafts.captureRow(selectedRow.key, "grid");
    const admitted = pending.pendingQueueRef.current.model.admit({
      id: "prerequisite",
      kind: "patch",
      source: "autosave",
      incidentId: "incident",
      clientInstanceId: "test",
      viewSchemaId: timelineViewSchemaId,
      rowKey: selectedRow.key,
      recordId: "first",
      clientTxnId: "txn",
      coalesceKey: "first",
      enqueueOrder: 1,
      payloadIntent: {
        view_schema_id: timelineViewSchemaId,
        base_row_version: 3,
        client_txn_id: "txn",
        changes: [
          {
            field_key: "timeline.activity_synopsis_text",
            value: "captured scalar",
          },
        ],
      },
    });
    expect(admitted.accepted).toBe(true);
    pending.replayContextByUnitId.set("prerequisite", {
      draftRevisions: revisions,
      sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
      focusField: identity.field,
      focusKey: "first",
      surface: "grid",
      rowSnapshot: selectedRow,
      continueOnFreshDraft: false,
      detectAutoResolution: false,
      promoteToCommittedRowInspect: false,
      viewportContinuityToken: undefined,
    });
    expect(readiness.blockingReason(selection)).toBeNull();
    drafts.setDraft(identity, "newer unsubmitted scalar");
    expect(readiness.blockingReason(selection)).toContain("unsaved edit");
    drafts.clearRow(selectedRow.key);
    drafts.setDraft(
      { ...identity, rowKey: row("second").key },
      "unselected work",
    );
    expect(readiness.blockingReason(selection)).toBeNull();
    drafts.setDraft(
      { ...identity, surface: "inspector", field: "tags" },
      "unsubmitted collection",
    );
    expect(readiness.blockingReason(selection)).toContain("unsaved edit");
    expect(execute).not.toHaveBeenCalled();
  });

  it("owns raw text locally and retains its native input through zero selection and late outcomes", async () => {
    const f = fixture();
    let renders = 0;
    let controller: ReturnType<typeof useTimelineBulkTagController> | undefined;
    function Harness() {
      renders++;
      controller = useTimelineBulkTagController({
        context: { authorized: true, capabilityAvailable: true },
        port: f.port,
        readiness: f.readiness,
        precedingSaves: async () => {},
        rows: f.rowsRef.current,
        rowsRef: f.rowsRef,
      });
      return <TimelineBulkTagControl binding={controller.controls} />;
    }
    render(<Harness />);
    act(() =>
      controller?.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first"]),
      ),
    );
    const input = screen.getByRole("textbox", {
      name: "Tag for selected Timeline records",
    }) as HTMLInputElement;
    const count = renders;
    act(() => input.focus());
    fireEvent.change(input, { target: { value: "  newer Ω  " } });
    input.setSelectionRange(2, 6, "backward");
    expect(renders).toBe(count);
    act(() =>
      controller?.snapshot.gridSelection.onSelectedRecordIdsChange(new Set()),
    );
    expect(screen.getByRole("textbox")).toBe(input);
    expect(input.value).toBe("  newer Ω  ");
    expect(document.activeElement).toBe(input);
    expect([
      input.selectionStart,
      input.selectionEnd,
      input.selectionDirection,
    ]).toEqual([2, 6, "backward"]);
    expect(
      screen
        .getByRole("button", { name: "Assign tag" })
        .hasAttribute("disabled"),
    ).toBe(true);
    act(() =>
      controller?.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first"]),
      ),
    );
    fireEvent.submit(
      screen.getByRole("form", { name: "Timeline bulk record actions" }),
    );
    fireEvent.change(input, { target: { value: "next draft" } });
    await act(async () => {
      await Promise.resolve();
    });
    expect(input.value).toBe("next draft");
    expect(document.activeElement).toBe(input);
    expect(screen.queryByText(/Assignment accepted/)).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Clear tag draft" }));
    expect(input.value).toBe("");
    expect(controller?.controls.selectedRecordIds.size).toBe(1);
  });
});
