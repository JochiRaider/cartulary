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
import {
  projectWorkbookQueryEntries,
  projectWorkbookViewBarWorkingSet,
} from "./workbookViewBarWorkingSet";

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
    expect(projection.chips.map((chip) => chip.removeCommand.kind)).toEqual([
      "group_set",
      "sort_remove",
      "filter_remove",
    ]);
    expect(JSON.stringify(projection.chips)).not.toContain("function");
    expect(projection.visibleChipCapacity).toBe(2);
    expect(projection.columns).toHaveLength(layoutState.columnOrder.length);
  });

  it("projects deterministic semantic slots at maximum query pressure", () => {
    const contract = requireViewContract(surface);
    const filterTemplate = contract.fieldMap[contract.filterFields[0] ?? ""];
    expect(filterTemplate).toBeDefined();
    if (filterTemplate === undefined) return;
    const filterFields = Array.from(
      { length: 16 },
      (_, index) => `fixture.filter_${index + 1}`,
    );
    const maximumContract = {
      ...contract,
      fieldMap: {
        ...contract.fieldMap,
        ...Object.fromEntries(
          filterFields.map((fieldKey, index) => [
            fieldKey,
            {
              ...filterTemplate,
              fieldKey,
              label: `Fixture field ${index + 1}`,
            },
          ]),
        ),
      },
    };
    const queryState = {
      filters: filterFields.map((fieldKey) => ({
        arg: { value: "the-same-unbroken-value" },
        fieldKey,
        op: "eq" as const,
      })),
      groupBy: contract.groupingFields[0] ?? null,
      sort: contract.sortFields.slice(0, 8).map((fieldKey) => ({
        direction: "asc" as const,
        fieldKey,
      })),
    };
    const entries = projectWorkbookQueryEntries(maximumContract, queryState);
    expect(entries).toHaveLength(25);
    expect(new Set(entries.map((entry) => entry.token)).size).toBe(25);
    expect(entries.slice(0, 3).map((entry) => entry.token)).toEqual([
      "G",
      "S1",
      "S2",
    ]);
    expect(entries.length - 3).toBe(22);
    expect(entries.length - 2).toBe(23);
    expect(entries.length).toBe(25);
  });

  it("projects one closure-free working set across saved-view, query, layout, and focus state", () => {
    const contract = requireViewContract(surface);
    const layoutState = defaultWorkbookLayoutState(contract);
    const groupField = contract.groupingFields[0];
    const sortField = contract.sortFields[0];
    const filterField = contract.filterFields[0];
    expect(groupField).toBeDefined();
    expect(sortField).toBeDefined();
    expect(filterField).toBeDefined();
    if (
      groupField === undefined ||
      sortField === undefined ||
      filterField === undefined
    ) {
      return;
    }
    const selectedSavedView = {
      display_name: "Priority investigations",
      layout_json: {},
      owner_user_id: "user-1",
      query_json: {},
      saved_view_id: "saved-1",
      saved_view_version: 4,
      scope: "private" as const,
      view_schema_id: surface,
    };
    const workingSet = projectWorkbookViewBarWorkingSet({
      availableActions: {
        createSavedView: true,
        deleteSavedView: true,
        duplicateSavedView: true,
        resetSavedView: true,
        setDefault: false,
        setHome: true,
        updateSavedView: true,
      },
      chromeMode: "base",
      contract,
      createAvailable: true,
      incidentId: "incident-1",
      inspectorAvailable: true,
      isSavedViewDirty: true,
      layoutState,
      queryState: {
        filters: [
          {
            arg: { value: "open" },
            fieldKey: filterField,
            op: "eq",
          },
        ],
        groupBy: groupField,
        sort: [{ direction: "desc", fieldKey: sortField }],
      },
      savedViewResourceKind: "ready",
      selectedSavedView,
      transient: {
        activeEntryKey: `filter:${filterField}`,
        openPanel: "filters",
        subjectKey: `incident-1:${surface}:saved_view:saved-1:4`,
      },
      viewSchemaId: surface,
    });

    expect(workingSet.controlOrder).toEqual([
      "saved_view",
      "sort",
      "group",
      "filters",
      "columns",
      "applied_query",
      "inspector",
      "create",
    ]);
    expect(workingSet.savedView).toMatchObject({
      accessibleName: "Priority investigations, Modified",
      displayName: "Priority investigations",
      mutable: true,
      primaryAction: "update",
      savedViewId: "saved-1",
      version: 4,
    });
    expect(workingSet.savedView.actionGroups).toEqual([
      { actions: ["create"], id: "create" },
      { actions: ["update", "reset"], id: "selected" },
      { actions: ["duplicate"], id: "duplicate" },
      { actions: ["set_home"], id: "startup" },
      { actions: ["delete"], id: "delete" },
    ]);
    expect(workingSet.filterEditor).toEqual({
      fieldKey: filterField,
      mode: "edit",
    });
    expect(workingSet.focusTargets).toMatchObject({
      invokingEntryKey: `filter:${filterField}`,
      ownerControlByKind: {
        filter: "filters",
        group: "group",
        sort: "sort",
      },
    });
    expect(workingSet.visibleQueryEntries.map((entry) => entry.token)).toEqual([
      "G",
      "S1",
      "F1",
    ]);
    expect(JSON.stringify(workingSet)).not.toContain("function");
  });

  it("resets the pure presentation seam when its semantic subject changes", () => {
    const contract = requireViewContract(surface);
    const filterDraft = defaultFilterDraft(contract);
    let state = createWorkbookGridControlsTransientState(
      "incident:timeline:base",
      filterDraft,
      false,
    );
    state = reduceWorkbookGridControlsTransientState(state, {
      panel: "filters",
      subjectKey: "incident:timeline:base",
      type: "toggle_panel",
    });
    expect(state.value.openPanel).toBe("filters");
    state = reduceWorkbookGridControlsTransientState(state, {
      filterDraft,
      subjectKey: "incident:timeline:saved_view:one:2",
      type: "activate_subject",
    });
    expect(state).toEqual(
      createWorkbookGridControlsTransientState(
        "incident:timeline:saved_view:one:2",
        filterDraft,
        false,
      ),
    );
  });

  it("scopes transient drafts to a subject and closes stale panels", () => {
    const contract = requireViewContract(surface);
    const firstDraft = defaultFilterDraft(contract);
    const secondDraft =
      firstDraft.op === "eq" ? { ...firstDraft, value: "second" } : firstDraft;
    let state = createWorkbookGridControlsTransientState(
      surface,
      firstDraft,
      false,
    );
    state = reduceWorkbookGridControlsTransientState(state, {
      type: "toggle_panel",
      panel: "filters",
      subjectKey: surface,
    });
    state = reduceWorkbookGridControlsTransientState(state, {
      type: "change_filter_draft",
      filterDraft:
        firstDraft.op === "eq" ? { ...firstDraft, value: "local" } : firstDraft,
      subjectKey: surface,
    });
    state = reduceWorkbookGridControlsTransientState(state, {
      type: "sync_filter_draft",
      filterDraft: secondDraft,
      subjectKey: surface,
    });
    expect(
      workbookGridSurfaceTransientState(state, surface, firstDraft).filterDraft,
    ).toMatchObject({ value: "local" });

    state = reduceWorkbookGridControlsTransientState(state, {
      type: "activate_subject",
      filterDraft: secondDraft,
      subjectKey: "cartulary.view.hosts.v1",
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
    ).toMatchObject({
      activeEntryKey: null,
      editingFilterFieldKey: null,
      filterDraft: secondDraft,
      openPanel: null,
    });
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
        op: "eq",
        operandKind: "value",
        value: "reviewed",
        valueType: "string",
        values: "",
      }),
    ).toEqual({ kind: "invalid", message: "Select a supported filter field." });
  });
});
