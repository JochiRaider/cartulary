import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
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

    registry.clearSubmittedRow(recordId, submitted);

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
    expect(registry.draftValueForFocusKey(focusKey)).toBeUndefined();
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

  it("invalidates drafts when the schema identity changes", () => {
    const { result, rerender } = renderHook(
      ({ schemaKey }) => useTimelineEditorDraftRegistry(schemaKey),
      { initialProps: { schemaKey: "timeline@1" } },
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

    rerender({ schemaKey: "timeline@2" });

    expect(result.current).not.toBe(originalRegistry);
    expect(
      result.current.draftValue({
        field: "activitySynopsisText",
        rowKey: recordId,
        surface: "grid",
      }),
    ).toBeUndefined();
  });
});
