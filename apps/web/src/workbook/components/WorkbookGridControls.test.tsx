import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  workbookColumnsMenuTriggerTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookQueryEntryTestId,
  workbookSortAppliedEntryTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import { useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  defaultWorkbookLayoutState,
  type WorkbookResolvedLayoutState,
} from "../layout/workbookColumnLayout";
import { workbookOrderedSortLimit } from "../models/workbookGridQueryControls";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { WorkbookGridControls } from "./WorkbookGridControls";

const timelineSurface = "cartulary.view.timeline.v2";

afterEach(() => {
  cleanup();
});

describe("WorkbookGridControls", () => {
  it("opens a focused filter chip for editing instead of removing it", () => {
    const contract = requireViewContract(timelineSurface);
    const onRemoveFilter = vi.fn();
    render(
      <WorkbookGridControls
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={vi.fn()}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onFilterDraftChange={vi.fn()}
        onGroupByChange={vi.fn()}
        onRemoveFilter={onRemoveFilter}
        onResetColumns={vi.fn()}
        onSortChange={vi.fn()}
        queryState={{
          filters: [
            {
              arg: { values: ["alpha"] },
              fieldKey: "timeline.tags",
              op: "contains_any",
            },
          ],
          groupBy: null,
          sort: [],
        }}
        surface={timelineSurface}
      />,
    );

    const chip = screen.getByTestId(
      workbookQueryEntryTestId(timelineSurface, "filter", "timeline.tags"),
    );
    fireEvent.click(chip);
    expect(onRemoveFilter).not.toHaveBeenCalled();
    expect(
      screen.getByTestId(workbookFilterPopoverTestId(timelineSurface)),
    ).toBeInstanceOf(HTMLElement);
    expect(
      (
        screen.getByTestId(
          gridFilterFieldTestId(timelineSurface),
        ) as HTMLSelectElement
      ).value,
    ).toBe("timeline.tags");
    fireEvent.keyDown(
      screen.getByTestId(workbookFilterPopoverTestId(timelineSurface)),
      { key: "Escape" },
    );
    expect(document.activeElement).toBe(chip);
  });

  it("gives the active-query region one Tab stop with roving navigation", () => {
    const contract = requireViewContract(timelineSurface);
    const onRemoveFilter = vi.fn();
    render(
      <WorkbookGridControls
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={vi.fn()}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onFilterDraftChange={vi.fn()}
        onGroupByChange={vi.fn()}
        onRemoveFilter={onRemoveFilter}
        onResetColumns={vi.fn()}
        onSortChange={vi.fn()}
        queryState={{
          filters: [
            {
              arg: { values: ["alpha"] },
              fieldKey: "timeline.tags",
              op: "contains_any",
            },
          ],
          groupBy: "timeline.capture_state",
          sort: [{ direction: "desc", fieldKey: "timeline.activity_sort_ts" }],
        }}
        surface={timelineSurface}
      />,
    );
    const chips = Array.from(
      screen
        .getByRole("toolbar", { name: "Active query chips" })
        .querySelectorAll<HTMLButtonElement>("button[data-query-entry-key]"),
    );
    expect(chips.map((chip) => chip.tabIndex)).toEqual([0, -1, -1]);
    chips[0]?.focus();
    fireEvent.keyDown(chips[0] as HTMLButtonElement, { key: "ArrowRight" });
    expect(document.activeElement).toBe(chips[1]);
    fireEvent.keyDown(chips[1] as HTMLButtonElement, { key: "End" });
    expect(document.activeElement).toBe(chips[2]);
    fireEvent.keyDown(chips[2] as HTMLButtonElement, { key: "Delete" });
    expect(onRemoveFilter).toHaveBeenCalledWith("timeline.tags");
  });

  it("exposes add, direction, priority, boundary, removal, and limit recovery actions", () => {
    const contract = requireViewContract(timelineSurface);
    const sortable = contract.fields
      .map((field) => field.fieldKey)
      .filter((fieldKey) => contract.sortableFieldMap[fieldKey]);
    expect(sortable.length).toBeGreaterThan(workbookOrderedSortLimit);
    render(<StatefulGridControls />);
    fireEvent.click(
      screen.getByTestId(workbookSortMenuTriggerTestId(timelineSurface)),
    );
    expect(
      screen.getByTestId(workbookSortMenuTestId(timelineSurface)),
    ).toBeInstanceOf(HTMLElement);

    for (const fieldKey of sortable.slice(0, workbookOrderedSortLimit)) {
      fireEvent.click(
        screen.getByTestId(workbookSortOptionTestId(timelineSurface, fieldKey)),
      );
    }
    const first = sortable[0];
    const second = sortable[1];
    const ninth = sortable[workbookOrderedSortLimit];
    expect(first).toBeDefined();
    expect(second).toBeDefined();
    expect(ninth).toBeDefined();
    if (first === undefined || second === undefined || ninth === undefined) {
      return;
    }
    const firstLabel = contract.fieldMap[first]?.label ?? first;
    const secondLabel = contract.fieldMap[second]?.label ?? second;
    const ninthOption = screen.getByTestId(
      workbookSortOptionTestId(timelineSurface, ninth),
    );
    expect((ninthOption as HTMLButtonElement).disabled).toBe(true);
    expect(
      screen.getByText(/Remove a sort before adding another/),
    ).toBeInstanceOf(HTMLElement);

    const firstOption = screen.getByTestId(
      workbookSortAppliedEntryTestId(timelineSurface, first),
    );
    expect(firstOption.textContent).toContain(`1. ${firstLabel}: asc`);
    fireEvent.click(
      screen.getByRole("menuitemcheckbox", {
        name: `Set ${firstLabel} descending`,
      }),
    );
    expect(
      screen.getByTestId(workbookSortAppliedEntryTestId(timelineSurface, first))
        .textContent,
    ).toContain(`1. ${firstLabel}: desc`);
    expect(
      (
        screen.getByRole("menuitem", {
          name: `Move ${firstLabel} earlier`,
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    fireEvent.click(
      screen.getByRole("menuitem", { name: `Move ${firstLabel} later` }),
    );
    expect(
      screen.getByTestId(
        workbookSortAppliedEntryTestId(timelineSurface, second),
      ).textContent,
    ).toContain(`1. ${secondLabel}: asc`);
    fireEvent.click(
      screen.getByRole("menuitem", { name: `Remove ${firstLabel} sort` }),
    );
    const recoveredNinthOption = screen.getByTestId(
      workbookSortOptionTestId(timelineSurface, ninth),
    );
    expect((recoveredNinthOption as HTMLButtonElement).disabled).toBe(false);
    fireEvent.click(recoveredNinthOption);
    expect(
      screen.getByTestId(
        workbookSortAppliedEntryTestId(timelineSurface, ninth),
      ),
    ).toBeInstanceOf(HTMLElement);
  });

  it("keeps invalid drafts visible, excludes them from apply, and resets panels by surface", () => {
    const contract = requireViewContract(timelineSurface);
    const onApplyFilter = vi.fn();
    const invalidDraft: FilterDraft = {
      booleanValue: "",
      fieldKey: "Capture State",
      op: "eq",
      operandKind: "value",
      value: "reviewed",
      valueType: "string",
      values: "",
    };
    const common = {
      contract,
      defaultFilterPopoverOpen: true,
      filterDraft: invalidDraft,
      layoutState: defaultWorkbookLayoutState(contract),
      onApplyFilter,
      onColumnHiddenChange: vi.fn(),
      onColumnMove: vi.fn(),
      onFilterDraftChange: vi.fn(),
      onGroupByChange: vi.fn(),
      onRemoveFilter: vi.fn(),
      onResetColumns: vi.fn(),
      onSortChange: vi.fn(),
      queryState: emptyWorkbookQueryState(),
    };
    const { rerender } = render(
      <WorkbookGridControls {...common} surface={timelineSurface} />,
    );
    expect(screen.getByText("Select a supported filter field.")).toBeInstanceOf(
      HTMLElement,
    );
    const apply = screen.getByTestId(gridFilterApplyTestId(timelineSurface));
    expect((apply as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(apply);
    expect(onApplyFilter).not.toHaveBeenCalled();

    rerender(
      <WorkbookGridControls
        {...common}
        defaultFilterPopoverOpen={false}
        surface="cartulary.view.hosts.v1"
      />,
    );
    expect(
      screen.queryByTestId(workbookFilterPopoverTestId(timelineSurface)),
    ).toBeNull();
  });

  it("parses filter controls exactly and restores focus on Escape", () => {
    const contract = requireViewContract(timelineSurface);
    const onApplyFilter = vi.fn();
    const onFilterDraftChange = vi.fn();
    render(
      <WorkbookGridControls
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={onApplyFilter}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onFilterDraftChange={onFilterDraftChange}
        onGroupByChange={vi.fn()}
        onRemoveFilter={vi.fn()}
        onResetColumns={vi.fn()}
        onSortChange={vi.fn()}
        queryState={emptyWorkbookQueryState()}
        surface={timelineSurface}
      />,
    );
    const trigger = screen.getByTestId(
      workbookFilterPopoverTriggerTestId(timelineSurface),
    );
    fireEvent.click(trigger);
    const field = screen.getByTestId(gridFilterFieldTestId(timelineSurface));
    expect(document.activeElement).toBe(field);
    fireEvent.change(field, { target: { value: "timeline.has_evidence" } });
    fireEvent.change(
      screen.getByTestId(gridFilterValueTestId(timelineSurface)),
      {
        target: { value: "false" },
      },
    );
    fireEvent.click(screen.getByTestId(gridFilterApplyTestId(timelineSurface)));
    expect(onApplyFilter).toHaveBeenCalledWith({
      booleanValue: "false",
      fieldKey: "timeline.has_evidence",
      op: "eq",
      operandKind: "value",
      value: "",
      valueType: "boolean",
      values: "",
    });

    fireEvent.click(trigger);
    fireEvent.keyDown(
      screen.getByTestId(workbookFilterPopoverTestId(timelineSurface)),
      { key: "Escape" },
    );
    expect(document.activeElement).toBe(trigger);
  });

  it("keeps column commands available when every data column is hidden", () => {
    const contract = requireViewContract(timelineSurface);
    const layout = defaultWorkbookLayoutState(contract);
    const onColumnHiddenChange = vi.fn();
    const onColumnMove = vi.fn();
    const onResetColumns = vi.fn();
    render(
      <WorkbookGridControls
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={{ ...layout, hiddenFieldKeys: layout.columnOrder }}
        onApplyFilter={vi.fn()}
        onColumnHiddenChange={onColumnHiddenChange}
        onColumnMove={onColumnMove}
        onFilterDraftChange={vi.fn()}
        onGroupByChange={vi.fn()}
        onRemoveFilter={vi.fn()}
        onResetColumns={onResetColumns}
        onSortChange={vi.fn()}
        queryState={emptyWorkbookQueryState()}
        surface={timelineSurface}
      />,
    );
    fireEvent.click(
      screen.getByTestId(workbookColumnsMenuTriggerTestId(timelineSurface)),
    );
    const visibilityItems = screen.getAllByRole("menuitemcheckbox");
    expect(visibilityItems).toHaveLength(layout.columnOrder.length);
    expect(
      visibilityItems.every(
        (item) => item.getAttribute("aria-checked") === "false",
      ),
    ).toBe(true);
    const first = layout.columnOrder[0];
    const last = layout.columnOrder.at(-1);
    expect(first).toBeDefined();
    expect(last).toBeDefined();
    if (first === undefined || last === undefined) return;
    const firstLabel = contract.fieldMap[first]?.label ?? first;
    const lastLabel = contract.fieldMap[last]?.label ?? last;
    expect(
      (
        screen.getByRole("menuitem", {
          name: `Move ${firstLabel} earlier`,
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      (
        screen.getByRole("menuitem", {
          name: `Move ${lastLabel} later`,
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    const firstVisibilityItem = visibilityItems[0];
    expect(firstVisibilityItem).toBeDefined();
    if (firstVisibilityItem === undefined) return;
    fireEvent.click(firstVisibilityItem);
    expect(onColumnHiddenChange).toHaveBeenCalledWith(first, false);
    fireEvent.click(screen.getByRole("menuitem", { name: "Reset columns" }));
    expect(onResetColumns).toHaveBeenCalledOnce();
  });
});

function StatefulGridControls() {
  const contract = requireViewContract(timelineSurface);
  const [queryState, setQueryState] = useState<WorkbookQueryState>(
    emptyWorkbookQueryState(),
  );
  const [layoutState] = useState<WorkbookResolvedLayoutState>(() =>
    defaultWorkbookLayoutState(contract),
  );
  const [filterDraft, setFilterDraft] = useState(() =>
    defaultFilterDraft(contract),
  );
  return (
    <WorkbookGridControls
      contract={contract}
      filterDraft={filterDraft}
      layoutState={layoutState}
      onApplyFilter={() => undefined}
      onClearFilters={() => {
        setQueryState((current) => ({ ...current, filters: [] }));
      }}
      onColumnHiddenChange={() => undefined}
      onColumnMove={() => undefined}
      onFilterDraftChange={setFilterDraft}
      onGroupByChange={(groupBy) => {
        setQueryState((current) => ({ ...current, groupBy }));
      }}
      onRemoveFilter={(fieldKey) => {
        setQueryState((current) => ({
          ...current,
          filters: current.filters.filter(
            (filter) => filter.fieldKey !== fieldKey,
          ),
        }));
      }}
      onResetColumns={() => undefined}
      onSortChange={(sort) => {
        setQueryState((current) => ({ ...current, sort }));
      }}
      queryState={queryState}
      surface={timelineSurface}
    />
  );
}
