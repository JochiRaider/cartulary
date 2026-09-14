import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { WorkbookLocalDraftStore } from "../../models/WorkbookLocalDraftStore";
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
    registry.beginCapture("draft-1");
    registry.setDraft(
      { field: "activitySynopsisText", rowKey: "draft-1", surface: "grid" },
      "typing after submission",
    );
    registry.acceptCapture("draft-1", refreshed);
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
    registry.clearSubmittedRow("draft-1", refreshed.values);
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
    registry.clearSubmittedRow(recordId, submitted.values);
    expect(registry.draftValue(identity)).toBe("  pending Ω  ");
    registry.clearSubmittedRow(recordId, submitted.values, { hostRefs: "" });
    expect(registry.draftValue(identity)).toBe("  pending Ω  ");
    registry.setDraft(identity, "newer text");
    registry.clearSubmittedRow(
      recordId,
      submitted.values,
      submitted.collectionDrafts,
    );
    expect(registry.materializeRow(committedRow()).collectionDrafts.tags).toBe(
      "newer text",
    );
    const accepted = registry.materializeRow(committedRow());
    registry.clearSubmittedRow(
      recordId,
      accepted.values,
      accepted.collectionDrafts,
    );
    expect(registry.draftValue(identity)).toBeUndefined();
  });

  it("invalidates drafts when the runtime lifetime changes", () => {
    const { result, rerender } = renderHook(
      ({ store }) => useTimelineEditorDraftRegistry(store),
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
