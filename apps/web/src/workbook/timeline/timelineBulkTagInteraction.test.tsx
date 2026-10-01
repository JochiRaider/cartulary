import { requireViewContract } from "@cartulary/view-contracts";
import { act, fireEvent, screen, waitFor } from "@testing-library/react";
import { useEffect } from "react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../testing/timelineWorkbookTestSupport";
import {
  acceptedQueryMetadata,
  renderWithWorkbookQueryBrowsing as render,
  renderHookWithWorkbookQueryBrowsing as renderHook,
} from "../../testing/workbookQueryTestSupport";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useWorkbookQueryBrowser } from "../query/WorkbookQueryBrowsingContext";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import { WorkbookBatchOperationOwner } from "../runtime/WorkbookBatchOperationOwner";
import { WorkbookMutationRuntimeRegistry } from "../runtime/WorkbookMutationRuntimeRegistry";
import type {
  WorkbookBatchTransport,
  WorkbookBatchTransportOutcome,
} from "../runtime/workbookBatchOperation";
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
import { timelineRecordSelectionPresentation } from "./models/timelineRowsModel";
import { timelineMutationOwnerFor } from "./mutations/WorkbookTimelineMutationOwner";

function requirePresent<T>(value: T | null | undefined): T {
  if (value === null || value === undefined)
    throw new Error("Expected current test state");
  return value;
}

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
  const send = vi.fn<WorkbookBatchTransport["send"]>(async () => ({
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
  const runtime = createWorkbookMutationRuntime(
    { incidentId: "incident", clientInstanceId: "test" },
    { create: () => "runtime-id" },
    { execute: vi.fn() },
  );
  runtime.setAuthority(owner.getSnapshot().authority);
  const authoring = timelineMutationOwnerFor(runtime).bulkTagAuthoring;
  const rowsRef = {
    current: [row("first"), row("second", 4), createDraftRow(1)],
  };
  const readiness = {
    subscribe: () => () => {},
    blockingReason: vi.fn<() => string | null>(() => null),
  };
  const query: WorkbookViewQueryPort["query"] = async (input) => ({
    kind: "accepted",
    value: {
      ...acceptedQueryMetadata(timelineViewSchemaId, input.queryState),
      incidentId: "incident",
      viewSchemaId: timelineViewSchemaId,
      rows: rowsRef.current.flatMap((row) =>
        row.rawRow === null ? [] : [row.rawRow],
      ),
    },
  });
  return { owner, send, port, rowsRef, readiness, query, authoring, runtime };
}

function useBulkTagFixture(
  input: Parameters<typeof useTimelineBulkTagController>[0],
  query: WorkbookViewQueryPort["query"],
) {
  const { binding, snapshot } = useWorkbookQueryBrowser(
    { query },
    timelineViewSchemaId,
  );
  useEffect(() => {
    const browser = binding.currentBrowser();
    if (!browser) return;
    let current = true;
    const controller = new AbortController();
    browser.observeRows(
      input.rows.flatMap((row) => (row.rawRow === null ? [] : [row.rawRow])),
    );
    void browser
      .query({
        contract,
        queryState: emptyWorkbookQueryState(),
        signal: controller.signal,
      })
      .then((result) => {
        if (
          current &&
          binding.currentBrowser() === browser &&
          result.kind === "accepted"
        )
          browser.accept(result.value);
      });
    return () => {
      current = false;
      controller.abort();
    };
  }, [binding, input.rows]);
  return {
    ...useTimelineBulkTagController(input),
    queryReady: snapshot.accepted !== null,
  };
}

describe("Timeline bulk tag interaction", () => {
  it("normalizes committed readable selection context without changing source or draft values", () => {
    const saved = row("first");
    const cases = [
      {
        synopsis: " Initial\n triage ",
        utc: " 2026-04-10T10:00:00Z ",
        local: "ignored",
        label: "Initial triage — 2026-04-10T10:00:00Z",
      },
      {
        synopsis: " ",
        utc: "",
        local: "",
        label: "No synopsis or activity time",
      },
      {
        synopsis: "",
        utc: "2026-04-10T10:00:00Z",
        local: "",
        label: "No synopsis — 2026-04-10T10:00:00Z",
      },
      {
        synopsis: "Synopsis only",
        utc: "\t",
        local: "",
        label: "Synopsis only",
      },
      {
        synopsis: "Local",
        utc: "",
        local: "2026-04-10 08:00",
        label: "Local — 2026-04-10 08:00 (local time)",
      },
      {
        synopsis: "🚀".repeat(121),
        utc: "",
        local: "",
        label: `${"🚀".repeat(119)}…`,
        full: "🚀".repeat(121),
      },
      { synopsis: "Ω".repeat(120), utc: "", local: "", label: "Ω".repeat(120) },
    ];
    for (const entry of cases) {
      const record = {
        ...saved,
        committedValues: {
          ...saved.committedValues,
          activitySynopsisText: entry.synopsis,
          activityUTCText: entry.utc,
          activityLocalText: entry.local,
        },
        values: {
          ...saved.values,
          activitySynopsisText: "Unsubmitted draft",
          activityUTCText: "Draft time",
        },
      };
      const before = JSON.stringify(record);
      expect(timelineRecordSelectionPresentation(record)).toEqual({
        label: `Select Timeline record: ${entry.label}`,
        description: `${entry.full ? `${entry.full}. ` : ""}Record ID: first`,
      });
      expect(JSON.stringify(record)).toBe(before);
      expect(
        timelineRecordSelectionPresentation({ ...record, recordId: "second" }),
      ).toEqual({
        label: `Select Timeline record: ${entry.label}`,
        description: `${entry.full ? `${entry.full}. ` : ""}Record ID: second`,
      });
    }
  });

  it("keeps selection labels committed through retained drafts and updates selected identity after acceptance", async () => {
    const f = fixture();
    const { result, rerender } = renderHook(
      ({ rows }) =>
        useBulkTagFixture(
          {
            context: { authorized: true, capabilityAvailable: true },
            port: f.port,
            authoring: f.authoring,
            readiness: f.readiness,
            precedingSaves: async () => {},
            rows,
            rowsRef: f.rowsRef,
          },
          f.query,
        ),
      { initialProps: { rows: f.rowsRef.current } },
    );
    await waitFor(() => expect(result.current.queryReady).toBe(true));
    act(() =>
      result.current.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first"]),
      ),
    );
    const present = () =>
      result.current.snapshot.gridSelection.getRecordSelectionPresentation({
        kind: "data",
        data: f.rowsRef.current.find(
          (r) => r.recordId === "first",
        ) as WorkbookRow,
        rowIdentity: { kind: "core_record", recordId: "first" },
        mutationIdentity: { kind: "core_row_version", baseRowVersion: 3 },
      });
    const original = row("first");
    f.rowsRef.current = [
      {
        ...original,
        values: {
          ...original.values,
          activitySynopsisText: "Rejected retained draft",
        },
        pendingSignature: "pending",
      },
      row("second"),
    ];
    rerender({ rows: f.rowsRef.current });
    expect(present()).toEqual({
      label: "Select Timeline record: first",
      description: "Record ID: first",
    });
    const accepted = row("first", 4);
    f.rowsRef.current = [
      row("second"),
      {
        ...accepted,
        committedValues: {
          ...accepted.committedValues,
          activitySynopsisText: "Committed update",
        },
        values: {
          ...accepted.values,
          activitySynopsisText: "Newer retained draft",
        },
      },
    ];
    rerender({ rows: f.rowsRef.current });
    expect(present()).toEqual({
      label: "Select Timeline record: Committed update",
      description: "Record ID: first",
    });
    expect([...result.current.controls.selectedRecordIds]).toEqual(["first"]);
    expect(f.send).not.toHaveBeenCalled();
  });
  it("retains pending members and explicit deselection while pruning departed and unauthorized records", async () => {
    const f = fixture();
    const { result, rerender } = renderHook(
      ({ authorized, rows }) =>
        useBulkTagFixture(
          {
            context: { authorized, capabilityAvailable: true },
            port: f.port,
            authoring: f.authoring,
            readiness: f.readiness,
            precedingSaves: async () => {},
            rows,
            rowsRef: f.rowsRef,
          },
          f.query,
        ),
      { initialProps: { authorized: true, rows: f.rowsRef.current } },
    );
    await waitFor(() => expect(result.current.queryReady).toBe(true));
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
      useBulkTagFixture(
        {
          context: { authorized: true, capabilityAvailable: true },
          port: f.port,
          authoring: f.authoring,
          readiness: f.readiness,
          precedingSaves: () => ready,
          rows: f.rowsRef.current,
          rowsRef: f.rowsRef,
        },
        f.query,
      ),
    );
    await waitFor(() => expect(result.current.queryReady).toBe(true));
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

  it("surfaces real admission refusal and rejects incomplete or blocked intended targets", async () => {
    const f = fixture();
    const { result } = renderHook(() =>
      useBulkTagFixture(
        {
          context: { authorized: true, capabilityAvailable: true },
          port: f.port,
          authoring: f.authoring,
          readiness: f.readiness,
          precedingSaves: async () => {},
          rows: f.rowsRef.current,
          rowsRef: f.rowsRef,
        },
        f.query,
      ),
    );
    await waitFor(() => expect(result.current.queryReady).toBe(true));
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

  it("retains raw text across detachment and its native input through zero selection and late outcomes", async () => {
    const f = fixture();
    let renders = 0;
    let controller: ReturnType<typeof useBulkTagFixture> | undefined;
    function Harness() {
      renders++;
      controller = useBulkTagFixture(
        {
          context: { authorized: true, capabilityAvailable: true },
          port: f.port,
          authoring: f.authoring,
          readiness: f.readiness,
          precedingSaves: async () => {},
          rows: f.rowsRef.current,
          rowsRef: f.rowsRef,
        },
        f.query,
      );
      return <TimelineBulkTagControl binding={controller.controls} />;
    }
    const view = render(<Harness />);
    await waitFor(() => expect(controller?.queryReady).toBe(true));
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
    fireEvent.change(input, { target: { value: "  retained Ω 東京 é 🚀  " } });
    const sent = f.send.mock.calls.length;
    view.rerender(<></>);
    view.rerender(<Harness />);
    await waitFor(() => expect(controller?.queryReady).toBe(true));
    const returned = screen.getByRole("textbox", {
      name: "Tag for selected Timeline records",
    }) as HTMLInputElement;
    expect(returned.value).toBe("  retained Ω 東京 é 🚀  ");
    const authority = requirePresent(f.runtime.batches.getSnapshot().authority);
    act(() => f.runtime.invalidate({ kind: "session_unavailable" }));
    expect(
      screen.queryByRole("textbox", {
        name: "Tag for selected Timeline records",
      }),
    ).toBeNull();
    act(() =>
      f.runtime.setAuthority({ ...authority, sessionIdentity: "restored" }),
    );
    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe(
      "  retained Ω 東京 é 🚀  ",
    );
    expect(document.activeElement).not.toBe(screen.getByRole("textbox"));
    expect(controller?.controls.selectedRecordIds.size).toBe(0);
    expect(document.activeElement).not.toBe(returned);
    expect(f.send.mock.calls.length).toBe(sent);
    expect(screen.getByText("Select records to assign this tag.")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Clear tag draft" }));
    expect(document.activeElement).toBe(screen.getByRole("textbox"));
    view.rerender(<></>);
    view.rerender(<Harness />);
    await waitFor(() => expect(controller?.queryReady).toBe(true));
    expect(
      screen.queryByRole("textbox", {
        name: "Tag for selected Timeline records",
      }),
    ).toBeNull();
    act(() =>
      controller?.snapshot.gridSelection.onSelectedRecordIdsChange(
        new Set(["first"]),
      ),
    );
    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe("");
    expect(f.send.mock.calls.length).toBe(sent);
    f.runtime.invalidate({ kind: "runtime_disposed" });
  });
  it("conceals same-account tag authoring and permanently retires it at security boundaries without status work", () => {
    for (const retirement of [
      "incident",
      "account",
      "disposal",
      "actor",
    ] as const) {
      const registry = new WorkbookMutationRuntimeRegistry();
      const scope = { incidentId: "incident", clientInstanceId: "client" };
      const create = () =>
        createWorkbookMutationRuntime(
          scope,
          { create: () => "id" },
          { execute: vi.fn() },
        );
      const runtime = registry.acquire(scope, create);
      const authority = {
        actorId: "actor",
        incidentId: "incident",
        sessionIdentity: "first",
        role: "editor" as const,
        closed: false,
      };
      // Lazy owner initialization must inherit already accepted authority.
      runtime.setAuthority(authority);
      const slot = timelineMutationOwnerFor(runtime).bulkTagAuthoring;
      const binding = slot.bind();
      const detach = binding.attach();
      const status = runtime.getSnapshot();
      const changed = vi.fn();
      const unsubscribe = runtime.subscribe(changed);
      const first = requirePresent(binding.getSnapshot());
      expect(binding.update("  Ω 東京 é 🚀  ", first.generation)).toBe(true);
      const typed = requirePresent(binding.getSnapshot());
      expect(binding.getSnapshot()).toBe(typed);
      expect(typed.revision).toBe(first.revision + 1);
      expect(runtime.getSnapshot()).toBe(status);
      expect(changed).not.toHaveBeenCalled();
      expect(runtime.pendingQueue().model.snapshot().units).toEqual([]);
      expect(runtime.batches.getSnapshot().entries).toEqual([]);
      expect(runtime.getSnapshot().conflicts).toEqual([]);
      registry.sessionUnavailable();
      expect(binding.getSnapshot()).toBeNull();
      expect(binding.update("obsolete hidden edit", typed.generation)).toBe(
        false,
      );
      runtime.applyAuthorizationRecoveryState("resumed");
      expect(binding.getSnapshot()).toBeNull();
      runtime.setAuthority({ ...authority, sessionIdentity: "restored" });
      expect(binding.getSnapshot()).toMatchObject({
        raw: typed.raw,
        revision: typed.revision,
      });
      expect(binding.update("obsolete authorized edit", typed.generation)).toBe(
        false,
      );
      runtime.setAuthority({ ...authority, role: "viewer" });
      const viewer = requirePresent(binding.getSnapshot());
      expect(viewer).toMatchObject({ raw: typed.raw, canEdit: false });
      expect(binding.update("viewer edit", viewer.generation)).toBe(false);
      runtime.setAuthority(authority);
      runtime.invalidate({ kind: "incident_closed" });
      const closed = requirePresent(binding.getSnapshot());
      expect(closed).toMatchObject({ raw: typed.raw, canEdit: false });
      expect(binding.update("closed edit", closed.generation)).toBe(false);
      expect(binding.clear(closed.generation)).toBe(true);
      expect(binding.getSnapshot()?.revision).toBe(typed.revision + 1);
      runtime.setAuthority(authority);
      binding.update(
        typed.raw,
        requirePresent(binding.getSnapshot()).generation,
      );
      detach();
      expect(
        binding.update(
          "detached edit",
          requirePresent(binding.getSnapshot()).generation,
        ),
      ).toBe(false);
      const next = slot.bind();
      next.attach();
      detach(); // Old cleanup cannot detach the successor.
      expect(next.isCurrent(next.getSnapshot()?.generation)).toBe(true);
      const separate = createWorkbookMutationRuntime(
        { ...scope, clientInstanceId: "other" },
        { create: () => "id" },
        { execute: vi.fn() },
      );
      separate.setAuthority(authority);
      expect(
        timelineMutationOwnerFor(separate).bulkTagAuthoring.getSnapshot()?.raw,
      ).toBe("");
      separate.invalidate({ kind: "runtime_disposed" });
      const beforeRetirement = requirePresent(next.getSnapshot());
      if (retirement === "account") registry.replaceAccount();
      else if (retirement === "disposal") registry.dispose();
      else if (retirement === "actor")
        runtime.setAuthority({ ...authority, actorId: "other" });
      else
        registry.acquire({ ...scope, incidentId: "other" }, () =>
          createWorkbookMutationRuntime(
            { ...scope, incidentId: "other" },
            { create: () => "id" },
            { execute: vi.fn() },
          ),
        );
      runtime.setAuthority(authority);
      next.attach();
      expect(next.getSnapshot()).toBeNull();
      expect(next.update("retired edit", beforeRetirement.generation)).toBe(
        false,
      );
      expect(next.clear(beforeRetirement.generation)).toBe(false);
      unsubscribe();
      registry.dispose();
    }
  });

  it("retains newer or cleared authoring through detached late acceptance and rejection without focus restoration", async () => {
    for (const outcome of ["acknowledged", "rejected"] as const) {
      const f = fixture();
      let settle!: (value: WorkbookBatchTransportOutcome) => void;
      f.send.mockImplementation(
        () =>
          new Promise((resolve) => {
            settle = resolve;
          }),
      );
      let controller: ReturnType<typeof useBulkTagFixture> | undefined;
      function Harness() {
        controller = useBulkTagFixture(
          {
            context: { authorized: true, capabilityAvailable: true },
            port: f.port,
            authoring: f.authoring,
            readiness: f.readiness,
            precedingSaves: async () => {},
            rows: f.rowsRef.current,
            rowsRef: f.rowsRef,
          },
          f.query,
        );
        return <TimelineBulkTagControl binding={controller.controls} />;
      }
      const view = render(<Harness />);
      await waitFor(() => expect(controller?.queryReady).toBe(true));
      act(() =>
        controller?.snapshot.gridSelection.onSelectedRecordIdsChange(
          new Set(["first", "second"]),
        ),
      );
      const input = screen.getByRole("textbox") as HTMLInputElement;
      fireEvent.change(input, { target: { value: "  draft A Ω  " } });
      fireEvent.submit(screen.getByRole("form"));
      await waitFor(() => expect(f.send).toHaveBeenCalledOnce());
      const attempt = requirePresent(f.send.mock.calls[0])[0];
      expect(attempt.plan.request).toMatchObject({
        tag_name: "draft A Ω",
        targets: [
          { record_id: "first", base_row_version: 3 },
          { record_id: "second", base_row_version: 4 },
        ],
      });
      fireEvent.change(input, { target: { value: "  draft B 東京  " } });
      const cellDrafts = createTimelineEditorDraftRegistry(
        f.runtime.localDraftsForSurface(timelineViewSchemaId),
        timelineMutationOwnerFor(f.runtime).capture,
      );
      const cellIdentity = {
        rowKey: row("second").key,
        field: "activitySynopsisText",
        surface: "grid",
      } as const;
      cellDrafts.setDraft(cellIdentity, "independent cell authoring");
      const cellRevisions = cellDrafts.captureRow(cellIdentity.rowKey, "grid");
      if (outcome === "rejected")
        fireEvent.click(
          screen.getByRole("button", { name: "Clear tag draft" }),
        );
      expect(cellDrafts.captureRow(cellIdentity.rowKey, "grid")).toEqual(
        cellRevisions,
      );
      const revision = requirePresent(f.authoring.getSnapshot()).revision;
      const old = requirePresent(controller);
      view.rerender(<></>);
      expect(old.controls.assignTag("obsolete", {})).toMatchObject({
        kind: "rejected",
      });
      await act(async () =>
        settle(
          outcome === "acknowledged"
            ? {
                kind: "acknowledged",
                receipt: {
                  viewSchemaId: timelineViewSchemaId,
                  changeSetId: "A",
                  rows: [],
                  conflicts: [],
                },
              }
            : {
                kind: "rejected",
                failure: { kind: "validation", message: "A rejected" },
              },
        ),
      );
      expect(f.authoring.getSnapshot()).toMatchObject({
        raw: outcome === "acknowledged" ? "  draft B 東京  " : "",
        revision,
      });
      expect(f.owner.getSnapshot().entries[0]?.attempt).toEqual(attempt);
      view.rerender(<Harness />);
      await waitFor(() => expect(controller?.queryReady).toBe(true));
      expect(controller?.controls.selectedRecordIds.size).toBe(0);
      if (outcome === "acknowledged") {
        const returned = screen.getByRole("textbox") as HTMLInputElement;
        expect(returned.value).toBe("  draft B 東京  ");
        expect(document.activeElement).not.toBe(returned);
        expect(screen.queryByText(/Assignment accepted/)).toBeNull();
      } else expect(screen.queryByRole("textbox")).toBeNull();
      expect(f.send).toHaveBeenCalledOnce();
      view.unmount();
      f.runtime.invalidate({ kind: "runtime_disposed" });
    }
  });
});
