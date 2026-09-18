import type { GridClearIntent } from "@cartulary/grid-adapter";
import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { fullWorkbookViewRow } from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createTimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import {
  normalizeTimelineFullRow,
  rowFromApi,
} from "../models/timelineRowModel";
import {
  planTimelineClear,
  useTimelineClearController,
} from "./useTimelineClearController";

const contract = requireViewContract(timelineViewSchemaId);
const fields = [
  "timeline.activity_synopsis_text",
  "timeline.raw_activity_text",
];
const ids = [
  "11111111-1111-4111-8111-111111111111",
  "22222222-2222-4222-8222-222222222222",
];
const rows = ids.map((id) =>
  rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(contract, id, 2, { [fields[0] as string]: "Source" }),
      "clear fixture",
    ),
  ),
);
function intent(): GridClearIntent {
  const targets = ids.flatMap((recordId) =>
    fields.map((fieldKey) => ({
      surface: {
        kind: "view_schema" as const,
        viewSchemaId: timelineViewSchemaId,
      },
      rowIdentity: { kind: "core_record" as const, recordId },
      fieldKey,
      mutationIdentity: {
        kind: "core_row_version" as const,
        baseRowVersion: 2,
      },
    })),
  );
  const first = targets[0];
  const last = targets.at(-1);
  if (!first || !last) throw new Error("Missing fixture");
  return {
    anchor: first,
    range: { start: last, end: first },
    expandedRange: {
      rowIdentities: ids.map((recordId) => ({ kind: "core_record", recordId })),
      fieldKeys: fields,
    },
    targets,
    delivery: {},
  };
}
function input() {
  return {
    authorized: true,
    contract,
    intent: intent(),
    rows,
    visibleFieldKeys: new Set(fields),
    hasUnsubmittedDraft: () => false,
  };
}

describe("Timeline clear planning", () => {
  it("captures one ordered null command and preserves predecessor ownership", () => {
    expect(planTimelineClear(input())).toEqual({
      kind: "accepted",
      command: {
        fieldKeys: fields,
        targets: ids.map((recordId) => ({ recordId, baseRowVersion: 2 })),
      },
    });
    const clearCells = vi.fn();
    const ready = Promise.resolve();
    const setError = vi.fn();
    const { result } = renderHook(() =>
      useTimelineClearController({
        ...input(),
        port: { clearCells },
        precedingSaves: () => ready,
        rowsRef: { current: rows },
        getVisibleFieldKeys: () => new Set(fields),
        setError,
      }),
    );
    const captured = intent();
    act(() => result.current(captured));
    expect(clearCells).toHaveBeenCalledExactlyOnceWith(
      {
        fieldKeys: fields,
        targets: ids.map((recordId) => ({ recordId, baseRowVersion: 2 })),
      },
      { delivery: captured.delivery, ready },
    );
  });
  it("rejects an entire selection for capability membership authority or unsubmitted drafts", () => {
    const base = input();
    for (const overrides of [
      { authorized: false },
      { rows: [] },
      { visibleFieldKeys: new Set<string>() },
      { hasUnsubmittedDraft: () => true },
      { rows: rows.map((row) => ({ ...row, captureState: "superseded" })) },
      { rows: rows.map((row) => ({ ...row, rowVersion: 3 })) },
    ]) {
      expect(planTimelineClear({ ...base, ...overrides }).kind).toBe(
        "rejected",
      );
    }
    for (const field of [
      "timeline.tags",
      "timeline.capture_state",
      "timeline.recorded_at",
      "missing",
    ]) {
      expect(
        planTimelineClear({
          ...base,
          intent: {
            ...base.intent,
            expandedRange: {
              ...base.intent.expandedRange,
              fieldKeys: [field, ...fields],
            },
          },
        }).kind,
      ).toBe("rejected");
    }
    const firstIdentity = base.intent.expandedRange.rowIdentities[0];
    if (!firstIdentity) throw new Error("Missing fixture identity");
    const duplicates = {
      ...base.intent,
      expandedRange: {
        ...base.intent.expandedRange,
        rowIdentities: [firstIdentity, firstIdentity],
      },
    };
    expect(planTimelineClear({ ...base, intent: duplicates }).kind).toBe(
      "rejected",
    );
  });
  it("distinguishes admitted draft revisions from newer grid and Inspector authoring", () => {
    const registry = createTimelineEditorDraftRegistry();
    const row = rows[0];
    if (!row) throw new Error("fixture");
    const identity = {
      rowKey: row.key,
      field: "activitySynopsisText" as const,
      surface: "grid" as const,
    };
    registry.setDraft(identity, "Older", row);
    expect(
      registry.hasUnsubmittedScalarDraft(row.key, identity.field, []),
    ).toBe(true);
    const admitted = registry.captureRow(row.key, "grid");
    expect(
      registry.hasUnsubmittedScalarDraft(row.key, identity.field, [admitted]),
    ).toBe(false);
    registry.setDraft({ ...identity, surface: "inspector" }, "Inspector", row);
    expect(
      registry.hasUnsubmittedScalarDraft(row.key, identity.field, [admitted]),
    ).toBe(true);
    registry.deleteDraft({ ...identity, surface: "inspector" });
    registry.setDraft(identity, "Newer", row);
    expect(
      registry.hasUnsubmittedScalarDraft(row.key, identity.field, [admitted]),
    ).toBe(true);
    expect(registry.draftValue(identity)).toBe("Newer");
  });
});
