import { act, cleanup, render, screen } from "@testing-library/react";
import { createRef } from "react";
import { afterEach, expect, it, vi } from "vitest";
import {
  WorkbookRecoveryNavigation,
  workbookRecoveryKey,
} from "../../shared/workbookRecoveryNavigation";
import { WorkbookRecoveryFixture } from "../../testing/WorkbookRecoveryFixture";
import { createWorkbookPendingMutationAdapter } from "../adapters/createWorkbookPendingMutationAdapter";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { WorkbookActiveSurfaceFrame } from "./WorkbookActiveSurfaceFrame";

afterEach(cleanup);

it("keeps the original editor accessible while conflict recovery opens and closes", () => {
  const incidentId = "10000000-0000-4000-8000-000000000001";
  const runtime = new WorkbookMutationRuntime(
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
