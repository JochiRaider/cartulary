import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, expect, it, vi } from "vitest";
import { WorkbookRecoveryNavigation } from "../../shared/workbookRecoveryNavigation";
import { WorkbookRecoveryFixture } from "../../testing/WorkbookRecoveryFixture";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import type {
  WorkbookBatchPlan,
  WorkbookBatchReceipt,
  WorkbookBatchTransport,
  WorkbookBatchTransportOutcome,
} from "../runtime/workbookBatchOperation";
import { WorkbookBatchRecovery } from "./WorkbookBatchRecovery";

const timeline = "cartulary.view.timeline.v2";
const hosts = "cartulary.view.hosts.v1";
const field = "timeline.activity_synopsis_text";
const authority = {
  actorId: "analyst",
  incidentId: "incident",
  sessionIdentity: "session",
  role: "editor" as const,
  closed: false,
};
type Scenario = {
  readonly label: string;
  readonly origin: string;
  readonly original: string;
  readonly plan: WorkbookBatchPlan;
  readonly receipt: Omit<WorkbookBatchReceipt, "viewSchemaId">;
};
const cases: readonly Scenario[] = [
  {
    label: "Paste",
    origin: "Timeline",
    original: "first\nsecond",
    plan: {
      operation: "pasteWorkbookClipboard",
      recordIds: ["record"],
      request: {
        view_schema_id: timeline,
        clipboard_text: "first\nsecond",
        start_field_key: field,
        columns: [field],
        targets: [
          { kind: "record", record_id: "record", base_row_version: 1 },
          { kind: "create" },
        ],
      },
    },
    receipt: {
      changeSetId: "change",
      rows: [{ record_id: "record", row_version: 2, cells: {} }],
      conflicts: [],
    },
  },
  {
    label: "Fill",
    origin: "Timeline",
    original: "copied value",
    plan: {
      operation: "applyWorkbookBulkMutation",
      recordIds: ["record"],
      request: {
        view_schema_id: timeline,
        kind: "fill_down_v1",
        field_key: field,
        value: "copied value",
        targets: [{ record_id: "record", base_row_version: 1 }],
      },
    },
    receipt: {
      changeSetId: "change",
      rows: [{ record_id: "record", row_version: 2, cells: {} }],
      conflicts: [
        {
          record_id: "other",
          field_key: field,
          base_row_version: 1,
          current_row_version: 2,
          conflict_token: "conflict",
          conflict_resolution_class: "text_compare_merge" as const,
          base_value: "old",
          client_value: "copied value",
          server_value: "peer",
        },
      ],
    },
  },
  {
    label: "Clear contents",
    origin: "Timeline",
    original: "Clear Activity Synopsis",
    plan: {
      operation: "applyWorkbookBulkMutation",
      recordIds: ["record"],
      request: {
        view_schema_id: timeline,
        kind: "clear_cells_v1",
        field_keys: [field],
        targets: [{ record_id: "record", base_row_version: 1 }],
      },
    },
    receipt: {
      changeSetId: null,
      rows: [],
      conflicts: [
        {
          record_id: "record",
          field_key: field,
          base_row_version: 1,
          current_row_version: 2,
          conflict_token: "conflict",
          conflict_resolution_class: "text_compare_merge" as const,
          base_value: "old",
          client_value: null,
          server_value: "peer",
        },
      ],
    },
  },
  {
    label: "Tag assignment",
    origin: "Timeline",
    original: "triage",
    plan: {
      operation: "applyWorkbookBulkMutation",
      recordIds: ["record"],
      request: {
        view_schema_id: timeline,
        kind: "multi_row_tag_assignment_v1",
        tag_name: "triage",
        targets: [{ record_id: "record", base_row_version: 1 }],
      },
    },
    receipt: { changeSetId: null, rows: [], conflicts: [] },
  },
  {
    label: "Paste",
    origin: "Hosts",
    original: "host.example.test\nother.example.test",
    plan: {
      operation: "pasteWorkbookClipboard",
      entityType: "host",
      recordIds: ["record"],
      request: {
        view_schema_id: hosts,
        clipboard_text: "host.example.test\nother.example.test",
        start_field_key: "host.hostname",
        columns: ["host.hostname"],
        targets: [
          { kind: "record", record_id: "record", base_row_version: 1 },
          { kind: "create" },
        ],
      },
    },
    receipt: {
      changeSetId: "change",
      rows: [{ record_id: "record", row_version: 2, cells: {} }],
      conflicts: [],
    },
  },
];

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (reason: Error) => void;
  const promise = new Promise<T>((done, fail) => {
    resolve = done;
    reject = fail;
  });
  return { promise, resolve, reject };
}

async function flush() {
  for (let i = 0; i < 25; i += 1) await Promise.resolve();
}

function fixture(plan: WorkbookBatchPlan) {
  let sequence = 0;
  const runtime = createWorkbookMutationRuntime(
    { incidentId: "incident", clientInstanceId: "client" },
    { create: () => `batch-${++sequence}` },
    createWorkbookPendingMutationAdapter({
      apiBase: undefined,
      incidentId: "incident",
    }),
  );
  runtime.setAuthority(authority);
  const replay = deferred<WorkbookBatchTransportOutcome>();
  const refresh = deferred<void>();
  const read = vi
    .fn()
    .mockRejectedValueOnce(new Error("Read unavailable"))
    .mockImplementationOnce(() => refresh.promise);
  runtime.registerSurface(plan.request.view_schema_id, read);
  const send = vi
    .fn<WorkbookBatchTransport["send"]>()
    .mockResolvedValueOnce({ kind: "uncertain" })
    .mockImplementationOnce(() => replay.promise);
  runtime.batches.configure({
    capture: (captured, currentAuthority, id) => ({
      id,
      authority: currentAuthority,
      plan: captured,
      apiBase: undefined,
      path: "/captured",
      body: JSON.stringify({ ...captured.request, client_txn_id: id }),
    }),
    send,
  });
  const navigation = new WorkbookRecoveryNavigation();
  render(
    <WorkbookRecoveryFixture navigation={navigation}>
      <WorkbookBatchRecovery runtime={runtime} />
      <button type="button">Newer work</button>
    </WorkbookRecoveryFixture>,
  );
  return { runtime, navigation, send, read, replay, refresh };
}

async function openRetry(
  user: ReturnType<typeof userEvent.setup>,
  label: string,
  origin: string,
) {
  await user.click(screen.getByRole("button", { name: /^Recovery \(\d+\)$/ }));
  await user.click(
    screen.getByRole("button", { name: `${label} · ${origin}` }),
  );
  return screen.getByRole("button", { name: `Retry ${label.toLowerCase()}` });
}

afterEach(cleanup);

it("Batch recovery keeps each shared retry and original input attached through replay and read recovery", async () => {
  for (const scenario of cases) {
    const user = userEvent.setup();
    const plan = scenario.plan;
    const t = fixture(plan);
    const id = t.runtime.batches.admit(plan, { delivery: {} });
    expect(id).toBeTruthy();
    await waitFor(() =>
      expect(t.runtime.batches.getSnapshot().entries[0]?.phase).toBe(
        "uncertain",
      ),
    );
    const retry = await openRetry(user, scenario.label, scenario.origin);
    const original = screen.getByRole("textbox", {
      name: "Original batch input",
    }) as HTMLTextAreaElement;
    expect(original.value).toBe(scenario.original);
    await user.click(retry);
    expect(retry.isConnected).toBe(true);
    expect(document.activeElement).toBe(retry);
    expect(retry.getAttribute("aria-busy")).toBe("true");
    expect(retry.getAttribute("aria-disabled")).toBe("true");
    expect(original.isConnected).toBe(true);
    expect(original.value).toBe(scenario.original);
    fireEvent.click(retry);
    expect(t.send).toHaveBeenCalledTimes(2);
    expect(t.send.mock.calls[1]?.[0]).toBe(t.send.mock.calls[0]?.[0]);
    await act(async () => {
      t.replay.resolve({
        kind: "acknowledged",
        receipt: {
          ...scenario.receipt,
          viewSchemaId: plan.request.view_schema_id,
        } as WorkbookBatchReceipt,
      });
      await flush();
    });
    await waitFor(() =>
      expect(t.runtime.batches.getSnapshot().entries[0]?.reconciliation).toBe(
        "required",
      ),
    );
    expect(
      screen.queryByRole("button", {
        name: `Retry ${scenario.label.toLowerCase()}`,
      }),
    ).toBeNull();
    const outcome = screen.getByRole("status", {
      name: `${scenario.label} outcome`,
    });
    expect(outcome.isConnected).toBe(true);
    expect(document.activeElement).toBe(outcome);
    if (scenario.label === "Fill")
      expect(outcome.textContent).toContain("1 original conflict");
    if (scenario.label === "Clear contents")
      expect(outcome.textContent).toContain("No batch changes saved");
    if (scenario.label === "Tag assignment")
      expect(outcome.textContent).toContain("No changes were needed");
    if (scenario.origin === "Hosts")
      expect(
        screen.queryByRole("button", { name: /Review this change/ }),
      ).toBeNull();
    const retryRefresh = screen.getByRole("button", { name: "Retry refresh" });
    await user.click(retryRefresh);
    expect(retryRefresh.isConnected).toBe(true);
    expect(document.activeElement).toBe(retryRefresh);
    expect(retryRefresh.getAttribute("aria-busy")).toBe("true");
    expect(retryRefresh.getAttribute("aria-disabled")).toBe("true");
    fireEvent.click(retryRefresh);
    expect(t.read).toHaveBeenCalledTimes(2);
    expect(t.send).toHaveBeenCalledTimes(2);
    await act(async () => {
      t.refresh.resolve();
      await flush();
    });
    await waitFor(() =>
      expect(t.runtime.batches.getSnapshot().entries[0]?.reconciliation).toBe(
        "complete",
      ),
    );
    expect(document.activeElement).toBe(outcome);
    expect(t.send).toHaveBeenCalledTimes(2);
    t.navigation.dispose();
    t.runtime.invalidate({ kind: "runtime_disposed" });
    cleanup();
  }
});

it("Batch recovery leaves Retry usable after repeated uncertainty and refresh failure", async () => {
  const scenario = cases[0];
  if (!scenario) throw new Error("Missing paste fixture");
  const user = userEvent.setup();
  const t = fixture(scenario.plan);
  t.runtime.batches.admit(scenario.plan, { delivery: {} });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.phase).toBe("uncertain"),
  );
  const retry = await openRetry(user, scenario.label, scenario.origin);
  await user.click(retry);
  await act(async () => {
    t.replay.resolve({ kind: "uncertain" });
    await flush();
  });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.transportPending).toBe(
      false,
    ),
  );
  expect(document.activeElement).toBe(retry);
  expect(retry.isConnected).toBe(true);
  expect(retry.getAttribute("aria-busy")).toBe("false");
  t.send.mockResolvedValueOnce({
    kind: "acknowledged",
    receipt: { ...scenario.receipt, viewSchemaId: timeline },
  });
  await user.click(retry);
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.reconciliation).toBe(
      "required",
    ),
  );
  expect(t.send).toHaveBeenCalledTimes(3);
  const refresh = screen.getByRole("button", { name: "Retry refresh" });
  await user.click(refresh);
  await act(async () => {
    t.refresh.reject(new Error("Read still unavailable"));
    await flush();
  });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.reconciliation).toBe(
      "required",
    ),
  );
  expect(document.activeElement).toBe(refresh);
  expect(refresh.isConnected).toBe(true);
  expect(refresh.getAttribute("aria-busy")).toBe("false");
  t.read.mockResolvedValueOnce(undefined);
  await user.click(refresh);
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.reconciliation).toBe(
      "complete",
    ),
  );
  expect(t.read).toHaveBeenCalledTimes(3);
  expect(t.send).toHaveBeenCalledTimes(3);
});

it("Batch recovery gives newer focus and scroll intent priority over late settlement", async () => {
  const scenario = cases[0];
  if (!scenario) throw new Error("Missing paste fixture");
  for (const movement of [
    "tab",
    "shift-tab",
    "pointer",
    "scroll",
    "list",
    "close",
  ] as const) {
    const user = userEvent.setup();
    const t = fixture(scenario.plan);
    t.runtime.batches.admit(scenario.plan, { delivery: {} });
    await waitFor(() =>
      expect(t.runtime.batches.getSnapshot().entries[0]?.phase).toBe(
        "uncertain",
      ),
    );
    const retry = await openRetry(user, scenario.label, scenario.origin);
    await user.click(retry);
    const panel = screen.getByRole("region", { name: "Workbook recovery" });
    panel.scrollTop = 23;
    if (movement === "tab") await user.tab();
    else if (movement === "shift-tab") await user.tab({ shift: true });
    else if (movement === "pointer")
      await user.click(screen.getByRole("button", { name: "Newer work" }));
    else if (movement === "scroll") fireEvent.scroll(panel);
    else if (movement === "list")
      await user.click(screen.getByRole("button", { name: "All recovery" }));
    else
      await user.click(screen.getByRole("button", { name: "Close recovery" }));
    const destination = document.activeElement;
    await act(async () => {
      t.replay.resolve({
        kind: "acknowledged",
        receipt: { ...scenario.receipt, viewSchemaId: timeline },
      });
      await flush();
    });
    expect(document.activeElement).toBe(destination);
    if (movement !== "list" && movement !== "close")
      expect(panel.scrollTop).toBe(23);
    if (movement === "scroll") {
      expect(retry.isConnected).toBe(true);
      expect(retry.getAttribute("aria-disabled")).toBe("true");
      await user.tab();
      expect(retry.isConnected).toBe(false);
    }
    if (movement === "list" || movement === "close")
      expect(
        screen.queryByRole("button", { name: "Retry refresh" }),
      ).toBeNull();
    t.navigation.dispose();
    t.runtime.invalidate({ kind: "runtime_disposed" });
    cleanup();
  }
});

it("Batch recovery withdraws protected controls on authority loss without reopening after a late result", async () => {
  const scenario = cases[0];
  if (!scenario) throw new Error("Missing paste fixture");
  const user = userEvent.setup();
  const t = fixture(scenario.plan);
  t.runtime.batches.admit(scenario.plan, { delivery: {} });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.phase).toBe("uncertain"),
  );
  const retry = await openRetry(user, scenario.label, scenario.origin);
  await user.click(retry);
  await act(async () => {
    t.runtime.setAuthority(null);
    await flush();
  });
  expect(retry.isConnected).toBe(false);
  expect(
    screen.queryByRole("textbox", { name: "Original batch input" }),
  ).toBeNull();
  await act(async () => {
    t.replay.resolve({
      kind: "acknowledged",
      receipt: { ...scenario.receipt, viewSchemaId: timeline },
    });
    await flush();
  });
  expect(screen.queryByRole("button", { name: "Retry refresh" })).toBeNull();
  expect(t.navigation.getSnapshot().selected).toBeNull();
});

it("Batch recovery does not reclaim focus after selecting another batch", async () => {
  const scenario = cases[0];
  if (!scenario || scenario.plan.operation !== "pasteWorkbookClipboard")
    throw new Error("Missing paste fixture");
  const user = userEvent.setup();
  const t = fixture(scenario.plan);
  t.runtime.batches.admit(scenario.plan, { delivery: {} });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.phase).toBe("uncertain"),
  );
  const retry = await openRetry(user, scenario.label, scenario.origin);
  await user.click(retry);
  t.send.mockResolvedValueOnce({ kind: "uncertain" });
  const other: WorkbookBatchPlan = {
    ...scenario.plan,
    recordIds: ["other"],
    request: {
      ...scenario.plan.request,
      clipboard_text: "Newer distinct input",
      targets: [{ kind: "record", record_id: "other", base_row_version: 1 }],
    },
  };
  await act(async () => {
    t.runtime.batches.admit(other, { delivery: {} });
    await flush();
  });
  await user.click(screen.getByRole("button", { name: "All recovery" }));
  await user.click(
    screen.getByRole("button", {
      name: "Paste · Timeline",
      description: /Newer distinct input/,
    }),
  );
  const destination = document.activeElement;
  await act(async () => {
    t.replay.resolve({
      kind: "acknowledged",
      receipt: { ...scenario.receipt, viewSchemaId: timeline },
    });
    await flush();
  });
  expect(document.activeElement).toBe(destination);
  expect(
    screen.getByRole("textbox", { name: "Original batch input" }),
  ).toHaveProperty("value", "Newer distinct input");
  expect(t.runtime.batches.getSnapshot().entries).toHaveLength(2);
});

it("Batch recovery lets owner pruning remove an obsolete item and use the shell focus fallback", async () => {
  const scenario = cases[0];
  if (!scenario) throw new Error("Missing paste fixture");
  const user = userEvent.setup();
  const t = fixture(scenario.plan);
  const firstId = t.runtime.batches.admit(scenario.plan, { delivery: {} });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.phase).toBe("uncertain"),
  );
  const retry = await openRetry(user, scenario.label, scenario.origin);
  await user.click(retry);
  await act(async () => {
    t.replay.resolve({
      kind: "acknowledged",
      receipt: { ...scenario.receipt, viewSchemaId: timeline },
    });
    await flush();
  });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.reconciliation).toBe(
      "required",
    ),
  );
  await user.click(screen.getByRole("button", { name: "Retry refresh" }));
  await act(async () => {
    t.refresh.resolve();
    await flush();
  });
  await waitFor(() =>
    expect(t.runtime.batches.getSnapshot().entries[0]?.reconciliation).toBe(
      "complete",
    ),
  );
  const priorOutcome = screen.getByRole("status", { name: "Paste outcome" });
  expect(document.activeElement).toBe(priorOutcome);
  t.send.mockResolvedValueOnce({
    kind: "acknowledged",
    receipt: {
      ...scenario.receipt,
      viewSchemaId: timeline,
      changeSetId: "newer-change",
    },
  });
  t.read.mockResolvedValueOnce(undefined);
  await act(async () => {
    t.runtime.batches.admit(scenario.plan, { delivery: {} });
    await flush();
  });
  await waitFor(() =>
    expect(
      t.runtime.batches.getSnapshot().entries.map((entry) => entry.id),
    ).not.toContain(firstId),
  );
  expect(priorOutcome.isConnected).toBe(false);
  expect(t.navigation.getSnapshot().selected).toBeNull();
  const recoveryHeading = screen.getByRole("heading", { level: 2 });
  await waitFor(() => expect(document.activeElement).toBe(recoveryHeading));
});
