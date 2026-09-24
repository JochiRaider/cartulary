import { timelineScalarEditorTestId } from "@cartulary/ui-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { createTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import { TimelineScalarEditor } from "./TimelineScalarEditor";

it("TimelineScalarEditor preserves controlled draft read-only presence and commit behavior", () => {
  const onBlurCommit = vi.fn();
  const registry = createTimelineEditorDraftRegistry();
  const onDraftChange = vi.fn((rowKey, field, surface, value, input) =>
    registry.setDraft(
      { rowKey, field, surface },
      value,
      undefined,
      input !== undefined,
    ),
  );
  const onEditModeChange = vi.fn();
  const registerInput = vi.fn();
  const dataTestId = timelineScalarEditorTestId({
    fieldKey: "timeline.activity_synopsis_text",
    recordId: "record-1",
    surface: "grid",
  });
  const props = {
    committedValue: "Committed",
    editorDraftRegistry: registry,
    controlId: "timeline-editor-test",
    dataTestId,
    field: "activitySynopsisText" as const,
    onBlurCommit,
    onDraftChange,
    onEditModeChange,
    onFocusAnchor: vi.fn(),
    onFocusRecord: vi.fn(),
    onKeyCommit: vi.fn(),
    presenceFieldKey: "timeline.activity_synopsis_text",
    registerInput,
    rowKey: "record-1",
    rowRecordId: "record-1",
    surface: "grid" as const,
  };
  const { rerender } = render(<TimelineScalarEditor {...props} />);
  const input = screen.getByTestId(dataTestId) as HTMLInputElement;

  fireEvent.focus(input);
  fireEvent.input(input, { target: { value: "Draft" } });
  fireEvent.blur(input);
  expect(onDraftChange).toHaveBeenCalledWith(
    "record-1",
    "activitySynopsisText",
    "grid",
    "Draft",
    { composing: false, pasteCompleted: false },
  );
  expect(onBlurCommit).toHaveBeenCalledWith(
    "record-1",
    "activitySynopsisText",
    "grid",
    "Draft",
  );
  expect(onEditModeChange).toHaveBeenCalledWith(
    "record-1",
    "timeline.activity_synopsis_text",
    true,
  );
  expect(onEditModeChange).toHaveBeenCalledWith(
    "record-1",
    "timeline.activity_synopsis_text",
    false,
  );

  const findInput = document.createElement("input");
  findInput.dataset.gridEditorExternalAction = "true";
  document.body.append(findInput);
  onBlurCommit.mockClear();
  fireEvent.focus(input);
  fireEvent.input(input, { target: { value: "Unfinished exact draft" } });
  fireEvent.blur(input, { relatedTarget: findInput });
  expect(onBlurCommit).not.toHaveBeenCalled();
  expect(input.value).toBe("Unfinished exact draft");
  findInput.remove();
  fireEvent.input(input, { target: { value: "Draft" } });
  fireEvent.focus(input);
  input.setSelectionRange(0, input.value.length);
  const pasteEvent = new Event("paste", {
    bubbles: true,
    cancelable: true,
  });
  Object.defineProperty(pasteEvent, "clipboardData", {
    value: { getData: () => "Native paste" },
  });
  fireEvent(input, pasteEvent);
  expect(pasteEvent.defaultPrevented).toBe(false);
  expect(input.value).toBe("Draft");
  // This only verifies the notification boundary. Real insertion, caret and
  // browser history are asserted by the production clipboard regression.
  fireEvent.input(input, {
    target: { value: "Native paste" },
    inputType: "insertFromPaste",
  });
  expect(onDraftChange).toHaveBeenLastCalledWith(
    "record-1",
    "activitySynopsisText",
    "grid",
    "Native paste",
    { composing: false, pasteCompleted: true },
  );
  const copy = new Event("copy", { bubbles: true, cancelable: true });
  fireEvent(input, copy);
  expect(copy.defaultPrevented).toBe(false);
  const cut = new Event("cut", { bubbles: true, cancelable: true });
  fireEvent(input, cut);
  expect(cut.defaultPrevented).toBe(false);
  const oversized = new Event("paste", { bubbles: true, cancelable: true });
  Object.defineProperty(oversized, "clipboardData", {
    value: { getData: () => "x".repeat(8 * 1024 * 1024 + 1) },
  });
  fireEvent(input, oversized);
  expect(oversized.defaultPrevented).toBe(true);
  expect(screen.getByRole("alert").textContent).toContain("8 MiB");
  fireEvent.blur(input);

  registry.setDraft(
    { rowKey: "record-1", field: "activitySynopsisText", surface: "grid" },
    "Controlled",
  );
  rerender(<TimelineScalarEditor {...props} readOnly />);
  expect(input.value).toBe("Controlled");
  onDraftChange.mockClear();
  fireEvent.input(input, { target: { value: "Ignored" } });
  expect(onDraftChange).not.toHaveBeenCalled();
  expect(input.readOnly).toBe(true);
});

it("TimelineScalarEditor leaves managed departure to the grid owner", () => {
  const onCloseGridEditor = vi.fn();
  const onKeyCommit = vi.fn();
  const registry = createTimelineEditorDraftRegistry();
  render(
    <TimelineScalarEditor
      committedValue="Saved"
      controlId="managed-editor"
      dataTestId="managed-editor"
      editorDraftRegistry={registry}
      field="activitySynopsisText"
      onBlurCommit={vi.fn()}
      onCloseGridEditor={onCloseGridEditor}
      onDraftChange={vi.fn()}
      onEditModeChange={vi.fn()}
      onFocusAnchor={vi.fn()}
      onFocusRecord={vi.fn()}
      onKeyCommit={onKeyCommit}
      presenceFieldKey="timeline.activity_synopsis_text"
      registerInput={vi.fn()}
      rowKey="row-1"
      rowRecordId="record-1"
      surface="grid"
    />,
  );
  const input = screen.getByTestId("managed-editor") as HTMLInputElement;
  for (const key of ["Enter", "Tab"] as const)
    for (const modifier of ["ctrlKey", "metaKey", "altKey"] as const)
      fireEvent.keyDown(input, { key, [modifier]: true });
  expect(onCloseGridEditor).not.toHaveBeenCalled();
  expect(onKeyCommit).toHaveBeenCalledTimes(6);
  expect(input.value).toBe("Saved");
});
