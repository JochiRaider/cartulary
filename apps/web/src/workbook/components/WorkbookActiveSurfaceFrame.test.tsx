import {
  workbookConflictControlTestId,
  workbookConflictResolverTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  screen,
  waitFor,
} from "@testing-library/react";
import { createRef } from "react";
import { afterEach, expect, it, vi } from "vitest";
import {
  WorkbookRecoveryNavigation,
  workbookRecoveryKey,
} from "../../shared/workbookRecoveryNavigation";
import {
  errorEnvelope,
  successEnvelope,
  timelineRow,
} from "../../testing/timelineWorkbookTestSupport";
import { WorkbookRecoveryFixture } from "../../testing/WorkbookRecoveryFixture";
import { renderWithWorkbookQueryBrowsing as render } from "../../testing/workbookQueryTestSupport";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { createWorkbookMutationRuntime } from "../runtime/createWorkbookMutationRuntime";
import { WorkbookActiveSurfaceFrame } from "./WorkbookActiveSurfaceFrame";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

it("keeps the original editor accessible while conflict recovery opens and closes", () => {
  const incidentId = "10000000-0000-4000-8000-000000000001";
  const runtime = createWorkbookMutationRuntime(
    { incidentId, clientInstanceId: "client" },
    { create: () => "unused-transaction" },
    createWorkbookPendingMutationAdapter({ apiBase: undefined, incidentId }),
  );
  const focus = {
    resolverActivation: null,
    focusSameFieldSummary: vi.fn(),
    sameFieldSummaryRef: createRef<HTMLDivElement>(),
  };
  const props = {
    activeContent: (
      <textarea
        aria-label="Original editor"
        defaultValue="  Exact local draft  "
      />
    ),
    activeSurfaceRef: createRef<HTMLElement>(),
    apiBase: undefined,
    focus,
    mutationRuntime: runtime,
    onActivateOrigin: vi.fn(),
  };
  const navigation = new WorkbookRecoveryNavigation();
  const frame = () => (
    <WorkbookRecoveryFixture navigation={navigation}>
      <WorkbookActiveSurfaceFrame
        {...props}
        sheetRef={{ kind: "view_schema", id: timelineViewSchemaId }}
      />
    </WorkbookRecoveryFixture>
  );
  const { rerender } = render(frame());
  const editor = screen.getByRole("textbox", { name: "Original editor" });
  editor.focus();
  const conflict = runtime.registerConflict({
    conflict: {
      base_row_version: 1,
      current_row_version: 2,
      record_id: "20000000-0000-4000-8000-000000000001",
      field_key: "timeline.activity_synopsis_text",
      conflict_token: "conflict-token",
      conflict_resolution_class: "text_compare_merge",
      base_value: "Base",
      server_value: "Saved",
      client_value: "  Exact local draft  ",
    },
    rowLabel: "Conflict row",
    surfaceLabel: "Timeline",
    viewSchemaId: timelineViewSchemaId,
    sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
  });
  rerender(frame());
  expect(
    screen.queryByRole("region", { name: "Workbook conflict recovery" }),
  ).toBeNull();
  act(() =>
    navigation.activate(
      workbookRecoveryKey("core", `conflict:${conflict.key}`),
    ),
  );
  expect(
    screen.getByRole("region", { name: "Workbook conflict recovery" }),
  ).toBeTruthy();
  expect(screen.getByRole("textbox", { name: "Original editor" })).toBe(editor);
  expect(editor.closest('[inert], [aria-hidden="true"]')).toBeNull();
  expect(document.activeElement).not.toBe(document.body);
  expect(editor).toHaveProperty("value", "  Exact local draft  ");
  expect(focus.focusSameFieldSummary).not.toHaveBeenCalled();
  act(() => navigation.close());
  rerender(frame());
  expect(
    screen.queryByRole("region", { name: "Workbook conflict recovery" }),
  ).toBeNull();
  expect(document.activeElement).not.toBe(document.body);
  expect(runtime.getSnapshot().conflicts).toHaveLength(1);
  runtime.invalidate({ kind: "runtime_disposed" });
});

it.each([
  "resolved",
  "failed",
  "refreshed",
  "during refresh",
  "returned",
  "newer submission",
  "replacement",
  "blocked admission",
] as const)("keeps the selected conflict draft and focus when an earlier resolution is %s", async (settlement) => {
  const incidentId = "10000000-0000-4000-8000-000000000001";
  const runtime = createWorkbookMutationRuntime(
    { incidentId, clientInstanceId: "client" },
    { create: () => crypto.randomUUID() },
    createWorkbookPendingMutationAdapter({ apiBase: undefined, incidentId }),
  );
  let releaseRefresh: () => void = () => {};
  const refreshGate = new Promise<void>((resolve) => {
    releaseRefresh = resolve;
  });
  let refreshStarted: () => void = () => {};
  const refreshHit = new Promise<void>((resolve) => {
    refreshStarted = resolve;
  });
  runtime.registerSurface(timelineViewSchemaId, async () => {
    refreshStarted();
    if (settlement === "during refresh") await refreshGate;
  });
  const firstRecordId = "20000000-0000-4000-8000-000000000001";
  const secondRecordId = "20000000-0000-4000-8000-000000000002";
  const register = (recordId: string, token: string) =>
    runtime.registerConflict({
      batchOperationId: "paste-1",
      conflict: {
        base_row_version: 1,
        current_row_version: 2,
        record_id: recordId,
        field_key: "timeline.activity_synopsis_text",
        conflict_token: token,
        conflict_resolution_class: "text_compare_merge",
        base_value: "Base",
        server_value: `Saved ${token}`,
        client_value: `Local ${token}`,
      },
      rowLabel: token,
      surfaceLabel: "Timeline",
      viewSchemaId: timelineViewSchemaId,
    });
  const first = register(firstRecordId, "first-token");
  const second = register(secondRecordId, "second-token");
  let release: (response: Response) => void = () => {};
  const held = new Promise<Response>((resolve) => {
    release = resolve;
  });
  let releaseSecond: (response: Response) => void = () => {};
  const secondHeld = new Promise<Response>((resolve) => {
    releaseSecond = resolve;
  });
  const fetchMock = vi
    .fn()
    .mockImplementationOnce(() => held)
    .mockImplementationOnce(() => secondHeld);
  vi.stubGlobal("fetch", fetchMock);
  const navigation = new WorkbookRecoveryNavigation();
  const batch = navigation.register("batch", {});
  batch.update([
    {
      id: "paste-1",
      label: "Paste conflicts",
      summary: "Review both conflicts",
      origin: "Timeline",
      sheetRef: { kind: "view_schema", id: timelineViewSchemaId },
      attention: "attention",
      order: 0,
      conflictKeys: [first.key, second.key],
    },
  ]);
  const focus = {
    resolverActivation: null,
    sameFieldSummaryRef: createRef<HTMLDivElement>(),
  };
  render(
    <WorkbookRecoveryFixture navigation={navigation}>
      <WorkbookActiveSurfaceFrame
        activeContent={<p>Timeline grid</p>}
        activeSurfaceRef={createRef<HTMLElement>()}
        apiBase={undefined}
        focus={focus}
        mutationRuntime={runtime}
        onActivateOrigin={vi.fn()}
        sheetRef={{ kind: "view_schema", id: timelineViewSchemaId }}
      />
    </WorkbookRecoveryFixture>,
  );
  act(() => navigation.activate(workbookRecoveryKey("batch", "paste-1")));
  fireEvent.change(
    screen.getByTestId(workbookConflictControlTestId("merged-value")),
    {
      target: { value: "  Exact A merge  " },
    },
  );
  fireEvent.click(
    screen.getByTestId(workbookConflictControlTestId("keep-saved")),
  );
  await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(1));
  expect(
    screen
      .getByTestId(workbookConflictControlTestId("keep-saved"))
      .getAttribute("aria-disabled"),
  ).toBe("true");
  fireEvent.click(
    screen.getByTestId(workbookConflictControlTestId("keep-saved")),
  );
  expect(fetchMock).toHaveBeenCalledTimes(1);
  if (settlement === "replacement") {
    act(() =>
      runtime.registerConflict({
        batchOperationId: "paste-1",
        conflict: {
          ...first.conflict,
          conflict_token: "replacement-token",
          server_value: "Replacement saved",
        },
        rowLabel: "first-token",
        surfaceLabel: "Timeline",
        viewSchemaId: timelineViewSchemaId,
      }),
    );
    expect(
      screen
        .getByTestId(workbookConflictControlTestId("use-merged"))
        .getAttribute("aria-busy"),
    ).toBe("false");
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("use-merged")),
    );
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(
      screen
        .getByTestId(workbookConflictControlTestId("use-merged"))
        .getAttribute("aria-busy"),
    ).toBe("true");
  }
  const resolvedResponse = () =>
    successEnvelope({
      view_schema_id: timelineViewSchemaId,
      change_set_id: "30000000-0000-4000-8000-000000000001",
      row: timelineRow({
        recordId: firstRecordId,
        rowVersion: 2,
        captureState: "rough",
      }),
    });
  if (settlement === "during refresh") {
    release(resolvedResponse());
    await refreshHit;
  }
  fireEvent.click(
    screen.getByTestId(workbookConflictControlTestId("paste-next")),
  );
  expect(
    screen
      .getByTestId(workbookConflictResolverTestId())
      .getAttribute("data-conflict-record-id"),
  ).toBe(secondRecordId);
  const secondDraft = screen.getByTestId(
    workbookConflictControlTestId("merged-value"),
  ) as HTMLTextAreaElement;
  fireEvent.change(secondDraft, { target: { value: "  Exact B merge  " } });
  secondDraft.focus();
  if (settlement === "blocked admission") {
    const blocked = vi.spyOn(runtime, "beginEntityWrite").mockReturnValue(null);
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("use-merged")),
    );
    expect(
      screen.getByText(
        "Finish the earlier batch or merge before resolving this edit.",
      ),
    ).toBeTruthy();
    expect(
      screen
        .getByTestId(workbookConflictControlTestId("use-merged"))
        .getAttribute("aria-busy"),
    ).toBe("false");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    blocked.mockRestore();
  }
  if (settlement === "newer submission") {
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("use-merged")),
    );
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
  }
  if (settlement === "replacement")
    expect(runtime.getSnapshot().conflicts[0]?.conflict.conflict_token).toBe(
      "replacement-token",
    );
  if (settlement === "returned")
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("paste-previous")),
    );
  await act(async () => {
    if (settlement === "during refresh") releaseRefresh();
    else
      release(
        settlement === "failed"
          ? errorEnvelope("resolution_refused", 409)
          : settlement === "refreshed"
            ? errorEnvelope("same_field_conflict", 409, {
                ...first.conflict,
                conflict_token: "first-refreshed-token",
                server_value: "Saved again",
                current_row_version: 3,
              })
            : resolvedResponse(),
      );
  });
  await waitFor(() =>
    expect(runtime.getSnapshot().conflicts.map((entry) => entry.key)).toEqual(
      settlement === "resolved" ||
        settlement === "during refresh" ||
        settlement === "returned" ||
        settlement === "newer submission" ||
        settlement === "blocked admission"
        ? [second.key]
        : [first.key, second.key],
    ),
  );
  if (settlement === "returned") {
    await waitFor(() => expect(navigation.getSnapshot().selected).toBeNull());
    act(() => navigation.activate(workbookRecoveryKey("batch", "paste-1")));
    expect(
      screen.getByTestId(workbookConflictControlTestId("merged-value")),
    ).toHaveProperty("value", "  Exact B merge  ");
  } else {
    expect(
      screen.getByRole("region", { name: "Workbook conflict recovery" }),
    ).toBeTruthy();
    expect(
      runtime.getSnapshot().conflicts.find((entry) => entry.key === second.key)
        ?.mergedDraft,
    ).toBe("  Exact B merge  ");
    expect(secondDraft.isConnected).toBe(true);
    expect(
      screen
        .getByTestId(workbookConflictResolverTestId())
        .getAttribute("data-conflict-record-id"),
    ).toBe(secondRecordId);
    expect(secondDraft.value).toBe("  Exact B merge  ");
    expect(document.activeElement).toBe(secondDraft);
  }
  expect(screen.queryByText("resolution_refused")).toBeNull();
  if (settlement === "blocked admission")
    expect(
      screen.getByText(
        "Finish the earlier batch or merge before resolving this edit.",
      ),
    ).toBeTruthy();
  if (settlement === "newer submission") {
    expect(
      screen
        .getByTestId(workbookConflictControlTestId("use-merged"))
        .getAttribute("aria-busy"),
    ).toBe("true");
    await act(async () => releaseSecond(errorEnvelope("second_refused", 409)));
    expect(screen.getByText("second_refused")).toBeTruthy();
    expect(
      screen.getByTestId(workbookConflictControlTestId("merged-value")),
    ).toHaveProperty("value", "  Exact B merge  ");
  }
  if (settlement === "replacement") {
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("paste-previous")),
    );
    expect(
      screen
        .getByTestId(workbookConflictControlTestId("use-merged"))
        .getAttribute("aria-busy"),
    ).toBe("true");
    await act(async () =>
      releaseSecond(errorEnvelope("replacement_refused", 409)),
    );
    expect(screen.getByText("replacement_refused")).toBeTruthy();
  }
  if (settlement === "refreshed") {
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("paste-previous")),
    );
    expect(
      screen.getByText(
        "The saved value changed again. Review the refreshed conflict.",
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(workbookConflictControlTestId("merged-value")),
    ).toHaveProperty("value", "  Exact A merge  ");
  }
  if (settlement === "failed") {
    fireEvent.click(
      screen.getByTestId(workbookConflictControlTestId("paste-previous")),
    );
    expect(screen.getByText("resolution_refused")).toBeTruthy();
  }
  runtime.invalidate({ kind: "runtime_disposed" });
  batch.unregister();
  navigation.dispose();
});
