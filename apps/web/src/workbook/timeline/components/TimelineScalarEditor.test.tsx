import { timelineScalarEditorTestId } from "@cartulary/ui-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { TimelineScalarEditor } from "./TimelineScalarEditor";

it("TimelineScalarEditor preserves controlled draft read-only presence and commit behavior", () => {
  const onBlurCommit = vi.fn();
  const onDraftChange = vi.fn();
  const onEditModeChange = vi.fn();
  const onPasteCommit = vi.fn();
  const registerInput = vi.fn();
  const dataTestId = timelineScalarEditorTestId({
    fieldKey: "timeline.activity_synopsis_text",
    recordId: "record-1",
    surface: "grid",
  });
  const props = {
    committedValue: "Committed",
    controlId: "timeline-editor-test",
    dataTestId,
    field: "activitySynopsisText" as const,
    onBlurCommit,
    onDraftChange,
    onEditModeChange,
    onFocusAnchor: vi.fn(),
    onFocusRecord: vi.fn(),
    onKeyCommit: vi.fn(),
    onPasteCommit,
    presenceFieldKey: "timeline.activity_synopsis_text",
    registerInput,
    rowKey: "record-1",
    rowRecordId: "record-1",
    surface: "grid" as const,
  };
  const { rerender } = render(<TimelineScalarEditor {...props} />);
  const input = screen.getByTestId(dataTestId) as HTMLInputElement;

  fireEvent.focus(input);
  fireEvent.change(input, { target: { value: "Draft" } });
  fireEvent.blur(input);
  expect(onDraftChange).toHaveBeenCalledWith(
    "record-1",
    "activitySynopsisText",
    "grid",
    "Draft",
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
  expect(pasteEvent.defaultPrevented).toBe(true);
  expect(input.value).toBe("Native paste");
  expect(onPasteCommit).toHaveBeenCalledWith(
    "record-1",
    "activitySynopsisText",
    "grid",
    "Native paste",
  );
  fireEvent.blur(input);

  rerender(
    <TimelineScalarEditor {...props} draftValue="Controlled" readOnly />,
  );
  expect(input.value).toBe("Controlled");
  fireEvent.change(input, { target: { value: "Ignored" } });
  expect(input.value).toBe("Controlled");
});
