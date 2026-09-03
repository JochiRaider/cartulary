import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import { defaultWorkbookLayoutState } from "../layout/workbookColumnLayout";
import {
  applyWorkbookSortCommand,
  createWorkbookGridControlsTransientState,
  parseDeclaredFieldKey,
  parseDeclaredGroupField,
  parseWorkbookBooleanDraftValue,
  projectWorkbookGridQueryControls,
  reduceWorkbookGridControlsTransientState,
  validateFilterDraft,
  workbookGridSurfaceTransientState,
  workbookOrderedSortLimit,
} from "./workbookGridQueryControls";
import { defaultFilterDraft, type WorkbookSortEntry } from "./workbookQuery";

const surface = "cartulary.view.timeline.v2";

describe("workbookGridQueryControls", () => {
  it("executes the complete ordered-sort lifecycle with reference-preserving no-ops", () => {
    const contract = requireViewContract(surface);
    const sortable = contract.fields
      .map((field) => field.fieldKey)
      .filter((fieldKey) => contract.sortableFieldMap[fieldKey]);
    expect(sortable.length).toBeGreaterThan(workbookOrderedSortLimit);
    const first = sortable[0];
    const second = sortable[1];
    expect(first).toBeDefined();
    expect(second).toBeDefined();
    if (first === undefined || second === undefined) return;

    const one = applyWorkbookSortCommand(contract, [], {
      kind: "sort_add",
      fieldKey: first,
    });
    expect(one).toEqual([{ direction: "asc", fieldKey: first }]);
    expect(
      applyWorkbookSortCommand(contract, one, {
        kind: "sort_add",
        fieldKey: first,
      }),
    ).toBe(one);
    expect(
      applyWorkbookSortCommand(contract, one, {
        kind: "sort_move",
        direction: "earlier",
        fieldKey: first,
      }),
    ).toBe(one);

    const two = applyWorkbookSortCommand(contract, one, {
      kind: "sort_add",
      fieldKey: second,
    });
    const descending = applyWorkbookSortCommand(contract, two, {
      kind: "sort_set_direction",
      direction: "desc",
      fieldKey: first,
    });
    expect(descending[0]).toEqual({ direction: "desc", fieldKey: first });
    const moved = applyWorkbookSortCommand(contract, descending, {
      kind: "sort_move",
      direction: "later",
      fieldKey: first,
    });
    expect(moved.map((entry) => entry.fieldKey)).toEqual([second, first]);
    expect(
      applyWorkbookSortCommand(contract, moved, {
        kind: "sort_move",
        direction: "later",
        fieldKey: first,
      }),
    ).toBe(moved);
    expect(
      applyWorkbookSortCommand(contract, moved, {
        kind: "sort_remove",
        fieldKey: first,
      }),
    ).toEqual([{ direction: "asc", fieldKey: second }]);
  });

  it("rejects a ninth or duplicate sort and recovers admission after removal", () => {
    const contract = requireViewContract(surface);
    const sortable = contract.fields
      .map((field) => field.fieldKey)
      .filter((fieldKey) => contract.sortableFieldMap[fieldKey]);
    let sort: readonly WorkbookSortEntry[] = sortable
      .slice(0, workbookOrderedSortLimit)
      .map((fieldKey) => ({ direction: "asc", fieldKey }));
    const ninth = sortable[workbookOrderedSortLimit];
    const first = sortable[0];
    expect(ninth).toBeDefined();
    expect(first).toBeDefined();
    if (ninth === undefined || first === undefined) return;
    expect(
      applyWorkbookSortCommand(contract, sort, {
        kind: "sort_add",
        fieldKey: ninth,
      }),
    ).toBe(sort);
    const shortened = applyWorkbookSortCommand(contract, sort, {
      kind: "sort_remove",
      fieldKey: first,
    });
    sort = applyWorkbookSortCommand(contract, shortened, {
      kind: "sort_add",
      fieldKey: ninth,
    });
    expect(sort).toHaveLength(workbookOrderedSortLimit);
    expect(sort.at(-1)?.fieldKey).toBe(ninth);
  });

  it("projects closure-free semantic chips in canonical responsive order", () => {
    const contract = requireViewContract(surface);
    const layoutState = defaultWorkbookLayoutState(contract);
    const projection = projectWorkbookGridQueryControls({
      chromeMode: "narrow_desktop",
      contract,
      filterChipTestId: (fieldKey) => `filter:${fieldKey}`,
      layoutState,
      queryState: {
        filters: [
          {
            arg: { values: ["alpha"] },
            fieldKey: "timeline.tags",
            op: "contains_any",
          },
        ],
        groupBy: "timeline.capture_state",
        sort: [{ direction: "desc", fieldKey: "timeline.activity_sort_ts" }],
      },
    });
    expect(projection.chips.map((chip) => chip.key)).toEqual([
      "group:timeline.capture_state",
      "sort:timeline.activity_sort_ts",
      "filter:timeline.tags",
    ]);
    expect(projection.chips.map((chip) => chip.command.kind)).toEqual([
      "group_set",
      "sort_remove",
      "filter_remove",
    ]);
    expect(JSON.stringify(projection.chips)).not.toContain("function");
    expect(projection.visibleChipCapacity).toBe(6);
    expect(projection.columns).toHaveLength(layoutState.columnOrder.length);
  });

  it("keeps transient drafts surface-keyed and closes the inactive surface", () => {
    const contract = requireViewContract(surface);
    const firstDraft = defaultFilterDraft(contract);
    const secondDraft = { ...firstDraft, value: "second" };
    let state = createWorkbookGridControlsTransientState(
      surface,
      firstDraft,
      false,
    );
    state = reduceWorkbookGridControlsTransientState(state, {
      type: "toggle_panel",
      panel: "filters",
      surface,
    });
    state = reduceWorkbookGridControlsTransientState(state, {
      type: "change_filter_draft",
      filterDraft: { ...firstDraft, value: "local" },
      surface,
    });
    state = reduceWorkbookGridControlsTransientState(state, {
      type: "sync_filter_draft",
      filterDraft: secondDraft,
      surface,
    });
    expect(
      workbookGridSurfaceTransientState(state, surface, firstDraft).filterDraft
        .value,
    ).toBe("local");

    state = reduceWorkbookGridControlsTransientState(state, {
      type: "activate_surface",
      filterDraft: secondDraft,
      surface: "cartulary.view.hosts.v1",
    });
    expect(
      workbookGridSurfaceTransientState(state, surface, firstDraft).openPanel,
    ).toBeNull();
    expect(
      workbookGridSurfaceTransientState(
        state,
        "cartulary.view.hosts.v1",
        secondDraft,
      ),
    ).toEqual({ filterDraft: secondDraft, openPanel: null });
  });

  it("parses controlled values exactly and rejects malformed filter drafts", () => {
    const contract = requireViewContract(surface);
    expect(parseWorkbookBooleanDraftValue("true")).toBe("true");
    expect(parseWorkbookBooleanDraftValue("truthy")).toBeNull();
    expect(
      parseDeclaredFieldKey("timeline.capture_state", contract.groupingFields),
    ).toBe("timeline.capture_state");
    expect(
      parseDeclaredFieldKey("Capture State", contract.groupingFields),
    ).toBeNull();
    expect(parseDeclaredGroupField("", contract.groupingFields)).toEqual({
      kind: "none",
    });
    expect(
      parseDeclaredGroupField("Capture State", contract.groupingFields),
    ).toBeNull();
    expect(
      validateFilterDraft(contract, {
        booleanValue: "",
        fieldKey: "Capture State",
        value: "reviewed",
      }),
    ).toEqual({ kind: "invalid", message: "Select a supported filter field." });
  });
});
