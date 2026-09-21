import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { taskAuthority } from "../../testing/taskWorkbookTestSupport";
import { useWorkbookInspectorEditDraft } from "./useWorkbookInspectorEditDraft";
import { useWorkbookInspectorFieldFeedback } from "./useWorkbookInspectorFieldFeedback";
import { WorkbookInspectorDraftStore } from "./WorkbookInspectorDraftStore";
import { inspectorFieldFeedback } from "./workbookInspectorFieldFeedback";

describe("Inspector field feedback", () => {
  it("binds field failures only to the captured canonical edit and revision", () => {
    const captured = {
      viewSchemaId: "cartulary.view.notes.v1",
      recordId: "note-a",
      fieldKey: "note.title",
      action: "replace",
      revision: 4,
    };
    const failure = {
      kind: "validation" as const,
      message: "The update was rejected.",
      fields: [{ field: "note.title", message: "Enter a title." }],
    };
    expect(inspectorFieldFeedback(failure, captured, captured)).toBe(
      "Enter a title.",
    );
    for (const current of [
      { ...captured, revision: 5 },
      { ...captured, recordId: "note-b" },
      { ...captured, fieldKey: "note.body" },
      { ...captured, viewSchemaId: "cartulary.view.timeline.v2" },
      { ...captured, action: "clear" },
    ])
      expect(inspectorFieldFeedback(failure, captured, current)).toBeNull();
    expect(
      inspectorFieldFeedback({ ...failure, fields: [] }, captured, captured),
    ).toBeNull();
    expect(inspectorFieldFeedback(failure, captured, null)).toBeNull();
  });
});

it("retains exact authoring and keeps newer feedback when an older rejection arrives", () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const contract = requireViewContract("cartulary.view.notes.v1"),
    field = contract.fieldMap["note.title"];
  if (!field) throw new Error("Missing Note title");
  const initial = { recordId: "note-a", active: true };
  const hook = renderHook(
    ({ recordId, active }) => {
      const edit = useWorkbookInspectorEditDraft({
        store,
        row: {
          record_id: recordId,
          row_version: 1,
          cells: { "note.title": { value: "Saved" } },
        },
        field,
        viewSchemaId: contract.viewSchemaId,
        active,
        presentation: "sheet",
      });
      return { edit, feedback: useWorkbookInspectorFieldFeedback(edit) };
    },
    { initialProps: initial },
  );
  act(() => hook.result.current.edit.update("  first raw  "));
  const older = hook.result.current.feedback.capture();
  act(() => hook.result.current.edit.update("  newer raw  "));
  const newer = hook.result.current.feedback.capture();
  act(() =>
    newer({
      kind: "validation",
      message: "Rejected",
      fields: [{ field: "note.title", message: "Newer feedback" }],
    }),
  );
  act(() =>
    older({
      kind: "validation",
      message: "Rejected",
      fields: [{ field: "note.title", message: "Older feedback" }],
    }),
  );
  expect(hook.result.current.feedback.message).toBe("Newer feedback");
  expect(hook.result.current.edit.value).toBe("  newer raw  ");
  act(() => hook.result.current.edit.update(null));
  expect(hook.result.current.edit.value).toBeNull();
  expect(hook.result.current.feedback.message).toBeNull();
  hook.rerender({ recordId: "note-b", active: true });
  act(() => newer({ kind: "validation", message: "Late rejection" }));
  expect(hook.result.current.feedback.actionError).toBeNull();
  hook.unmount();
});
