import type { GridColumn } from "@cartulary/grid-adapter";
import {
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  applyWorkbookLayoutToColumns,
  defaultWorkbookLayoutState,
  moveWorkbookColumn,
  reorderWorkbookColumns,
  setWorkbookColumnHidden,
  setWorkbookColumnWidth,
} from "../layout/workbookColumnLayout";
import {
  enumFilterChoices,
  toggleEnumFilterChoice,
} from "./workbookEnumFilterOperand";
import {
  applyFilterDraft,
  buildFilterFromDraft,
  buildQueryRequest,
  buildSavedViewQueryJson,
  changeFilterDraftOperandKind,
  compareWorkbookGroupValues,
  cycleWorkbookSortField,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  filterChipLabel,
  filterDraftForField,
  filterDraftFromFilter,
  toggleSortField,
  updateGroupBy,
  validateFilterDraft,
  workbookGroupValue,
  workbookQueryStateFromSavedViewQueryJson,
} from "./workbookQuery";
import {
  workbookContractForViewSchemaId,
  workbookQuerySurfaceSlot,
} from "./workbookSurfaceQueryRuntime";

describe("workbookQuery", () => {
  it("round trips declared boolean scalar set and null operands without changing JSON types", () => {
    const consumers = listViewContracts().flatMap((contract) =>
      contract.fields
        .filter(
          (field) =>
            field.readKind === "boolean" &&
            contract.filterFields.includes(field.fieldKey) &&
            field.filterOps.includes("eq"),
        )
        .map((field) => ({ contract, field })),
    );
    expect(consumers.length).toBeGreaterThan(0);
    for (const { contract, field } of consumers) {
      const initial = filterDraftForField(contract, field.fieldKey, "eq");
      expect(initial).toMatchObject({ valueType: "boolean" });
      expect(validateFilterDraft(contract, initial).kind).toBe("invalid");
      for (const arg of [
        { value: true },
        { value: false },
        { values: [true] },
        { values: [false] },
        { values: [false, true] },
        { value: null },
      ]) {
        const filter = { fieldKey: field.fieldKey, op: "eq" as const, arg };
        const state = { ...emptyWorkbookQueryState(), filters: [filter] };
        const saved = buildSavedViewQueryJson(contract, state);
        const restored = workbookQueryStateFromSavedViewQueryJson(
          contract,
          saved,
        );
        expect(restored.filters).toEqual([filter]);
        const draft = filterDraftFromFilter(contract, filter);
        expect(draft).toMatchObject({ valueType: "boolean" });
        expect(validateFilterDraft(contract, draft)).toEqual({
          kind: "valid",
          filter,
        });
        expect(
          buildQueryRequest(
            contract,
            applyFilterDraft(contract, emptyWorkbookQueryState(), draft),
          ).filters,
        ).toEqual(saved.filters);
        if (draft.op !== "eq") throw new Error("Expected equality");
        if (arg.value === null) {
          expect(
            validateFilterDraft(contract, { ...draft, operandKind: "value" })
              .kind,
          ).toBe("invalid");
          expect(
            validateFilterDraft(contract, { ...draft, operandKind: "values" })
              .kind,
          ).toBe("invalid");
        }
      }
    }
  });
  it("rejects malformed boolean operands without coercion and preserves contract overrides", () => {
    const contract = requireViewContract("cartulary.view.task_requests.v1");
    const fieldKey = "task.no_owner";
    const field = contract.fieldMap[fieldKey];
    if (!field) throw new Error("Missing declared Task Requests boolean");
    const draft = filterDraftForField(contract, fieldKey, "eq");
    if (draft.op !== "eq") throw new Error("Expected equality");
    const state = {
      ...emptyWorkbookQueryState(),
      filters: [{ fieldKey, op: "eq" as const, arg: { value: false } }],
    };
    for (const arg of [
      { value: "true" },
      { value: "false" },
      { value: 0 },
      { value: 1 },
      { values: [] },
      { values: Array(1) },
      { values: ["true"] },
      { values: [false, "false"] },
      { values: [true, null] },
      { values: [false, 0] },
      { values: "false" },
      { value: false, values: [true] },
      { value: null, extra: true },
      {},
    ]) {
      const filter = { fieldKey, op: "eq" as const, arg };
      const saved = buildSavedViewQueryJson(contract, {
        ...state,
        filters: [filter],
      });
      const restored = workbookQueryStateFromSavedViewQueryJson(
        contract,
        saved,
      );
      expect(restored.filters).toEqual([filter]);
      const invalid = filterDraftFromFilter(contract, filter);
      expect(validateFilterDraft(contract, invalid).kind).toBe("invalid");
      expect(applyFilterDraft(contract, state, invalid)).toBe(state);
      expect(buildFilterFromDraft(invalid)).toBeNull();
      if (invalid.op !== "eq") throw new Error("Expected equality");
      expect(invalid.booleanOperand?.invalidRestoredArg).toEqual(arg);
      expect(
        validateFilterDraft(
          contract,
          changeFilterDraftOperandKind(invalid, "null"),
        ),
      ).toEqual({
        kind: "valid",
        filter: { fieldKey, op: "eq", arg: { value: null } },
      });
    }
    for (const operandKind of ["value", "values", "null"] as const) {
      for (const valueType of ["string", "number"] as const) {
        const invalid = {
          ...draft,
          operandKind,
          valueType,
          value: "false",
          values: ["true"],
          booleanOperand: { value: false, values: [true] },
        };
        expect(validateFilterDraft(contract, invalid).kind).toBe("invalid");
        expect(applyFilterDraft(contract, state, invalid)).toBe(state);
      }
    }
    for (const booleanOperand of [
      undefined,
      { value: "false", values: [] },
      { value: false, values: [false, "true"] },
      { value: 0, values: [true, null] },
    ]) {
      for (const operandKind of ["value", "values"] as const) {
        const invalid = {
          ...draft,
          operandKind,
          booleanOperand,
        } as FilterDraft;
        if (operandKind === "value" && booleanOperand?.value === false)
          continue;
        expect(validateFilterDraft(contract, invalid).kind).toBe("invalid");
        expect(applyFilterDraft(contract, state, invalid)).toBe(state);
      }
    }
    const duplicate = {
      ...draft,
      operandKind: "values" as const,
      booleanOperand: { value: undefined, values: [true, false, true, false] },
    };
    expect(validateFilterDraft(contract, duplicate)).toMatchObject({
      kind: "valid",
      filter: { arg: { values: [true, false, true, false] } },
    });
    expect(
      buildQueryRequest(contract, applyFilterDraft(contract, state, duplicate))
        .filters?.[0]?.arg,
    ).toEqual({ values: [true, false, true, false] });
    for (const overridden of [
      {
        ...contract,
        filterFields: contract.filterFields.filter((key) => key !== fieldKey),
      },
      {
        ...contract,
        fieldMap: {
          ...contract.fieldMap,
          [fieldKey]: { ...field, filterOps: [] },
        },
      },
      {
        ...contract,
        fieldMap: {
          ...contract.fieldMap,
          [fieldKey]: {
            ...field,
            readKind: "string" as const,
          },
        },
      },
    ]) {
      expect(validateFilterDraft(overridden, duplicate).kind).toBe("invalid");
      expect(applyFilterDraft(overridden, state, duplicate)).toBe(state);
      expect(filterDraftForField(overridden, fieldKey, "eq")).not.toMatchObject(
        { valueType: "boolean" },
      );
    }
  });

  it("projects active enum metadata and preserves exact set literals on reopening", () => {
    for (const contract of listViewContracts()) {
      for (const field of contract.fields.filter(
        (field) => field.readKind === "enum" && field.filterOps.includes("eq"),
      )) {
        const draft = filterDraftForField(contract, field.fieldKey, "eq");
        if (draft.op !== "eq") throw new Error("Expected equality");
        expect(enumFilterChoices(contract, draft)).toEqual(field.enumValues);
        const first = field.enumValues?.[0];
        if (!first) continue;
        const custom = ["Custom,with,commas", "CUSTOM", first];
        const filter = {
          fieldKey: field.fieldKey,
          op: "eq" as const,
          arg: { values: custom },
        };
        const reopened = filterDraftFromFilter(contract, filter);
        if (reopened.op !== "eq") throw new Error("Expected equality");
        expect(reopened.values).toEqual(custom);
        const saved = buildSavedViewQueryJson(contract, {
          ...emptyWorkbookQueryState(),
          filters: [filter],
        });
        const restored = workbookQueryStateFromSavedViewQueryJson(
          contract,
          saved,
        ).filters[0];
        expect(restored).toEqual(filter);
        if (!restored) throw new Error("Missing restored enum filter");
        expect(filterDraftFromFilter(contract, restored)).toEqual(reopened);
        const twice = toggleEnumFilterChoice(
          toggleEnumFilterChoice(reopened, first, true),
          first,
          true,
        );
        expect(twice.values).toEqual(custom);
        const removed = toggleEnumFilterChoice(twice, first, false);
        expect(removed.values).toEqual(custom.slice(0, 2));
        expect(buildFilterFromDraft(reopened)?.arg.values).toEqual(
          expect.arrayContaining(custom),
        );
        expect(
          buildFilterFromDraft({ ...reopened, values: [first] })?.arg,
        ).toEqual({ values: [first] });
        expect(
          buildFilterFromDraft({ ...reopened, values: ["", " "] }),
        ).toBeNull();
        expect(
          enumFilterChoices(contract, {
            fieldKey: field.fieldKey,
            op: "prefix",
            value: "pre",
          }),
        ).toBeNull();
        const missing = {
          ...contract,
          fieldMap: {
            ...contract.fieldMap,
            [field.fieldKey]: { ...field, enumValues: null },
          },
        };
        expect(enumFilterChoices(missing, draft)).toBeNull();
        expect(
          enumFilterChoices(
            {
              ...missing,
              fieldMap: {
                ...missing.fieldMap,
                [field.fieldKey]: { ...field, enumValues: [] },
              },
            },
            draft,
          ),
        ).toBeNull();
      }
    }
    const timeline = requireViewContract("cartulary.view.timeline.v2");
    expect(
      enumFilterChoices(
        timeline,
        filterDraftForField(timeline, "timeline.capture_state", "eq"),
      ),
    ).toBeNull();
  });
  it("retains enum custom literals case variants and mixed equality shapes", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const draft = filterDraftForField(
      contract,
      "timeline.activity_time_pair_state",
      "eq",
    );
    if (draft.op !== "eq") throw new Error("Expected equality");
    for (const value of ["disable", "Disabled", "disabled", "empty", "unset"]) {
      const filter = buildFilterFromDraft({ ...draft, value });
      expect(filter?.arg).toEqual({ value });
      expect(validateFilterDraft(contract, { ...draft, value }).kind).toBe(
        "valid",
      );
      if (!filter) throw new Error("Expected filter");
      expect(
        buildFilterFromDraft(
          filterDraftFromFilter(
            requireViewContract("cartulary.view.timeline.v2"),
            filter,
          ),
        ),
      ).toEqual(filter);
    }
    const set = {
      ...draft,
      operandKind: "values" as const,
      values: "disabled, disable, Disabled, disabled",
    };
    const filter = buildFilterFromDraft(set);
    expect(filter?.arg.values).toEqual(
      expect.arrayContaining(["Disabled", "disable", "disabled"]),
    );
    expect(filter?.arg.values).toHaveLength(3);
    if (!filter) throw new Error("Expected set");
    expect(
      buildFilterFromDraft(
        filterDraftFromFilter(
          requireViewContract("cartulary.view.timeline.v2"),
          filter,
        ),
      ),
    ).toEqual(filter);
    expect(buildFilterFromDraft({ ...set, values: " , , " })).toBeNull();
    expect(
      buildFilterFromDraft({ ...draft, operandKind: "null" })?.arg,
    ).toEqual({ value: null });
  });

  it("validates calendar dates for every declared date filter", () => {
    const consumers = listViewContracts().flatMap((contract) =>
      contract.fields
        .filter(
          (field) =>
            field.readKind === "date" && field.filterOps.includes("eq"),
        )
        .map((field) => ({ contract, field })),
    );
    expect(consumers.length).toBeGreaterThan(0);
    for (const { contract, field } of consumers) {
      const draft = filterDraftForField(contract, field.fieldKey, "eq");
      if (draft.op !== "eq") throw new Error("Date equality unavailable");
      for (const value of [
        "0000-02-29",
        "0001-01-01",
        "0099-12-31",
        "2000-02-29",
        "2024-02-29",
        "2026-04-18",
        "9999-12-31",
        " 2026-04-18 \n",
      ]) {
        expect(
          validateFilterDraft(contract, { ...draft, value }).kind,
          `${field.fieldKey}: ${value}`,
        ).toBe("valid");
      }
      for (const value of [
        "",
        "1900-02-29",
        "2026-02-29",
        "2026-04-31",
        "2026-00-01",
        "2026-13-01",
        "2026-04-00",
        "2026-4-18",
        "26-04-18",
        "2026/04/18",
        "2026-04-18T00:00:00Z",
        "2026-04-18 extra",
        "10000-01-01",
      ]) {
        expect(
          validateFilterDraft(contract, { ...draft, value }).kind,
          `${field.fieldKey}: ${value}`,
        ).toBe("invalid");
      }
    }
  });

  it("preserves date equality sets null encoding and exact raw input", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const draft = filterDraftForField(
      contract,
      "timeline.date_entered_sort_day",
      "eq",
    );
    if (draft.op !== "eq") throw new Error("Expected equality");
    const values = " 2026-04-19, 2026-04-18\n2026-04-19, , ";
    const set = { ...draft, operandKind: "values" as const, values };
    expect(
      applyFilterDraft(contract, emptyWorkbookQueryState(), set).filters[0]
        ?.arg,
    ).toEqual({ values: ["2026-04-18", "2026-04-19"] });
    expect(set.values).toBe(values);
    for (const values of [
      ", \n ,",
      "2026-04-31,2026-04-18",
      "2026-04-18,2026-04-31,2026-04-19",
      "2026-04-18,null",
      "2026-04-18,2026-04-19T00:00:00Z",
    ]) {
      const invalid = { ...set, values };
      const state = emptyWorkbookQueryState();
      expect(applyFilterDraft(contract, state, invalid)).toBe(state);
      expect(invalid.values).toBe(values);
    }
    expect(
      applyFilterDraft(contract, emptyWorkbookQueryState(), {
        ...draft,
        operandKind: "null",
        value: "2026-04-31",
      }).filters[0]?.arg,
    ).toEqual({ value: null });
    expect(
      validateFilterDraft(contract, {
        ...draft,
        valueType: "number",
        value: "20260418",
      }).kind,
    ).toBe("invalid");
    expect(
      validateFilterDraft(contract, {
        ...draft,
        valueType: "boolean",
        booleanOperand: { value: true, values: [] },
      }).kind,
    ).toBe("invalid");
  });

  it("validates one-sided date ranges and inclusive boundaries", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const draft = filterDraftForField(
      contract,
      "timeline.date_entered_sort_day",
      "range",
    );
    if (draft.op !== "range") throw new Error("Expected range");
    for (const [lowerValue, upperValue, kind] of [
      ["", "", "invalid"],
      [" ", "\n", "invalid"],
      ["2026-04-18", "", "valid"],
      ["", "2026-04-18", "valid"],
      [" 2026-04-18 ", "2026-04-19", "valid"],
      ["2026-04-19", "2026-04-18", "invalid"],
      ["2026-04-31", "", "invalid"],
      ["", "2026-02-29", "invalid"],
      ["2026-4-18", "2026-04-19", "invalid"],
    ])
      expect(
        validateFilterDraft(contract, {
          ...draft,
          lowerValue: lowerValue ?? "",
          upperValue: upperValue ?? "",
        }).kind,
      ).toBe(kind);
    for (const lowerKind of ["gte", "gt"] as const) {
      for (const upperKind of ["lte", "lt"] as const) {
        expect(
          validateFilterDraft(contract, {
            ...draft,
            lowerKind,
            upperKind,
            lowerValue: "2026-04-18",
            upperValue: "2026-04-18",
          }).kind,
        ).toBe(
          lowerKind === "gte" && upperKind === "lte" ? "valid" : "invalid",
        );
        expect(
          validateFilterDraft(contract, {
            ...draft,
            lowerKind,
            upperKind,
            lowerValue: "2026-04-18",
            upperValue: "2026-04-19",
          }).kind,
        ).toBe("valid");
      }
    }
    expect(
      applyFilterDraft(contract, emptyWorkbookQueryState(), {
        ...draft,
        lowerKind: "gt",
        lowerValue: " 2026-04-18 ",
        upperValue: "",
      }).filters[0]?.arg,
    ).toEqual({ gt: "2026-04-18" });
  });

  it("keeps non-date and timestamp operands unchanged and rejects wrong-schema dates", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const text: FilterDraft = {
      fieldKey: "timeline.tags",
      op: "contains_any",
      values: "  query words  ",
    };
    expect(
      applyFilterDraft(contract, emptyWorkbookQueryState(), text).filters[0]
        ?.arg,
    ).toEqual({ values: ["query words"] });
    const timestampContract = requireViewContract(
      "cartulary.view.task_requests.v1",
    );
    const timestamp = filterDraftForField(
      timestampContract,
      "task.due_at",
      "range",
    );
    if (timestamp.op !== "range") throw new Error("Expected timestamp range");
    expect(
      validateFilterDraft(timestampContract, {
        ...timestamp,
        lowerValue: "unchanged timestamp input",
        upperValue: "2026-04-18T00:00:00+04:00",
      }).kind,
    ).toBe("valid");
    const state = emptyWorkbookQueryState();
    expect(
      applyFilterDraft(requireViewContract("cartulary.view.hosts.v1"), state, {
        ...timestamp,
        fieldKey: "timeline.date_entered_sort_day",
        lowerValue: "2026-04-18",
      }),
    ).toBe(state);
  });

  it("refuses impossible date drafts without replacing query state", () => {
    const state = emptyWorkbookQueryState();
    expect(
      applyFilterDraft(
        requireViewContract("cartulary.view.timeline.v2"),
        state,
        {
          fieldKey: "timeline.date_entered_sort_day",
          op: "eq",
          operandKind: "value",
          value: " 2026-04-31 ",
          valueType: "string",
          values: "",
        },
      ),
    ).toBe(state);
  });

  it("builds every declared filter operator without losing argument shape", () => {
    const filters = [
      buildFilterFromDraft({
        fieldKey: "eq",
        op: "eq",
        operandKind: "null",
        value: "",
        valueType: "string",
        values: "",
      }),
      buildFilterFromDraft({
        fieldKey: "range",
        lowerKind: "gt",
        lowerValue: "1",
        op: "range",
        upperKind: "lte",
        upperValue: "9",
      }),
      buildFilterFromDraft({
        fieldKey: "any",
        op: "contains_any",
        values: "b, a",
      }),
      buildFilterFromDraft({
        fieldKey: "all",
        op: "contains_all",
        values: "b, a",
      }),
      buildFilterFromDraft({ fieldKey: "prefix", op: "prefix", value: "pre" }),
      buildFilterFromDraft({
        fieldKey: "text",
        op: "full_text",
        query: "one  two",
      }),
    ];
    expect(filters).toEqual([
      { arg: { value: null }, fieldKey: "eq", op: "eq" },
      { arg: { gt: "1", lte: "9" }, fieldKey: "range", op: "range" },
      { arg: { values: ["a", "b"] }, fieldKey: "any", op: "contains_any" },
      { arg: { values: ["a", "b"] }, fieldKey: "all", op: "contains_all" },
      { arg: { value: "pre" }, fieldKey: "prefix", op: "prefix" },
      { arg: { query: "one  two" }, fieldKey: "text", op: "full_text" },
    ]);
    expect(
      filters.map(
        (filter) =>
          filter &&
          filterDraftFromFilter(
            requireViewContract("cartulary.view.timeline.v2"),
            filter,
          ).op,
      ),
    ).toEqual([
      "eq",
      "range",
      "contains_any",
      "contains_all",
      "prefix",
      "full_text",
    ]);
  });

  it("builds tag and boolean filters from the client-local draft state", () => {
    const state = applyFilterDraft(
      requireViewContract("cartulary.view.timeline.v2"),
      emptyWorkbookQueryState(),
      {
        fieldKey: "timeline.tags",
        op: "contains_any",
        values: "phish, c2",
      },
    );

    expect(state.filters).toEqual([
      {
        fieldKey: "timeline.tags",
        op: "contains_any",
        arg: {
          values: ["c2", "phish"],
        },
      },
    ]);

    expect(
      applyFilterDraft(
        requireViewContract("cartulary.view.timeline.v2"),
        emptyWorkbookQueryState(),
        {
          booleanOperand: { value: true, values: [] },
          fieldKey: "timeline.has_evidence",
          op: "eq",
          operandKind: "value",
          value: "",
          valueType: "boolean",
          values: "",
        },
      ).filters,
    ).toEqual([
      {
        fieldKey: "timeline.has_evidence",
        op: "eq",
        arg: {
          value: true,
        },
      },
    ]);
  });

  it("keeps grouping separate from authored sorting and cleared overrides", () => {
    expect([null, "z", "a"].sort(compareWorkbookGroupValues)).toEqual([
      "a",
      "z",
      null,
    ]);
    expect([true, null, false].sort(compareWorkbookGroupValues)).toEqual([
      false,
      true,
      null,
    ]);
    expect([10, null, 2].sort(compareWorkbookGroupValues)).toEqual([
      2,
      10,
      null,
    ]);
    expect(
      workbookGroupValue({ group_values: { state: false } }, "state"),
    ).toBe(false);
    expect(
      workbookGroupValue({ group_values: { state: null } }, "state"),
    ).toBeNull();

    const contract = requireViewContract("cartulary.view.timeline.v2");
    const next = updateGroupBy(
      contract,
      toggleSortField(
        contract,
        emptyWorkbookQueryState(),
        "timeline.activity_synopsis_text",
      ),
      "timeline.capture_state",
    );

    expect(buildQueryRequest(contract, next)).toEqual({
      group_by: "timeline.capture_state",
      sort: [
        { field_key: "timeline.activity_synopsis_text", direction: "asc" },
      ],
    });
    expect(buildQueryRequest(contract, { ...next, sort: [] })).toEqual({
      group_by: "timeline.capture_state",
    });
  });

  it("supports ordered additive sort cycles and enforces the owner limit", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const sortable = contract.fields
      .map((field) => field.fieldKey)
      .filter((fieldKey) => contract.sortableFieldMap[fieldKey])
      .slice(0, 9);
    let state = emptyWorkbookQueryState();
    for (const fieldKey of sortable) {
      state = cycleWorkbookSortField(contract, state, fieldKey, true);
    }
    expect(state.sort).toHaveLength(Math.min(8, sortable.length));
    const first = state.sort[0];
    expect(first).toBeDefined();
    if (first === undefined) return;
    const descending = cycleWorkbookSortField(
      contract,
      state,
      first.fieldKey,
      true,
    );
    expect(descending.sort[0]).toEqual({ ...first, direction: "desc" });
    const removed = cycleWorkbookSortField(
      contract,
      descending,
      first.fieldKey,
      true,
    );
    expect(
      removed.sort.some((entry) => entry.fieldKey === first.fieldKey),
    ).toBe(false);
  });

  it("describes filter chip labels from contract metadata", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    expect(
      filterChipLabel(contract, {
        fieldKey: "timeline.capture_state",
        op: "eq",
        arg: { value: "reviewed" },
      }),
    ).toBe("Capture State: reviewed");
  });

  it("initializes filter drafts from the declared contract operators", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    expect(defaultFilterDraft(contract).fieldKey).toBe(
      "timeline.date_entered_sort_day",
    );
    expect(workbookQuerySurfaceSlot("cartulary.view.timeline.v2")).toBe(
      "timeline",
    );
    expect(workbookQuerySurfaceSlot("cartulary.view.hosts.v1")).toBe("hosts");
    expect(workbookQuerySurfaceSlot("cartulary.view.assessments.v1")).toBe(
      "assessments",
    );
    expect(workbookQuerySurfaceSlot("cartulary.view.notes.v1")).toBe("generic");
    expect(
      workbookContractForViewSchemaId("cartulary.view.notes.v1").viewSchemaId,
    ).toBe("cartulary.view.notes.v1");
    expect(() =>
      workbookContractForViewSchemaId("cartulary.view.unknown.v1"),
    ).toThrow("Unknown workbook view schema");
  });

  it("keeps a full semantic layout permutation while compiling visible columns", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const initial = defaultWorkbookLayoutState(contract);
    const [first, second] = initial.columnOrder;
    expect(first).toBeDefined();
    expect(second).toBeDefined();
    if (first === undefined || second === undefined) return;
    const reordered = reorderWorkbookColumns(contract, initial, second, first);
    const hidden = setWorkbookColumnHidden(contract, reordered, first, true);
    const sized = setWorkbookColumnWidth(contract, hidden, second, 480);
    const columns = contract.fields.map(
      (field): GridColumn<Record<string, never>> => ({
        fieldKey: field.fieldKey,
        label: field.label,
        renderCell: () => null,
      }),
    );

    expect(sized.columnOrder).toHaveLength(contract.fields.length);
    expect(sized.columnOrder.slice(0, 2)).toEqual([second, first]);
    expect(
      applyWorkbookLayoutToColumns(contract, columns, sized)
        .slice(0, 1)
        .map((column) => ({ key: column.fieldKey, width: column.width })),
    ).toEqual([{ key: second, width: 480 }]);
    expect(moveWorkbookColumn(contract, sized, second, "earlier")).toEqual(
      sized,
    );
  });

  it("rejects structural layout keys and out-of-range widths", () => {
    const contract = requireViewContract("cartulary.view.timeline.v2");
    const initial = defaultWorkbookLayoutState(contract);
    expect(
      setWorkbookColumnWidth(contract, initial, "__selection__", 96),
    ).toEqual(initial);
    const fieldKey = initial.columnOrder[0] ?? "";
    expect(setWorkbookColumnWidth(contract, initial, fieldKey, 39)).toEqual(
      initial,
    );
    expect(setWorkbookColumnWidth(contract, initial, fieldKey, 4097)).toEqual(
      initial,
    );
  });
});
