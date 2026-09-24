import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  workbookColumnsMenuTriggerTestId,
  workbookFilterOperatorTestId,
  workbookFilterPopoverTestId,
  workbookFilterPopoverTriggerTestId,
  workbookQueryEntryTestId,
  workbookQueryOverflowEntryTestId,
  workbookSortAppliedEntryTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { useLayoutEffect, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useWorkbookColumnLayoutController } from "../layout/useWorkbookColumnLayoutController";
import { defaultWorkbookLayoutState } from "../layout/workbookColumnLayout";
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
  it("leaves native filter control keys alone and contains the complete range Tab order after pointer focus", () => {
    render(<StatefulGridControls />);
    fireEvent.click(
      screen.getByTestId(workbookFilterPopoverTriggerTestId(timelineSurface)),
    );
    const field = screen.getByTestId(gridFilterFieldTestId(timelineSurface));
    const arrow = new KeyboardEvent("keydown", {
      bubbles: true,
      cancelable: true,
      key: "ArrowDown",
    });
    field.dispatchEvent(arrow);
    expect(arrow.defaultPrevented).toBe(false);
    expect(document.activeElement).toBe(field);

    fireEvent.change(field, {
      target: { value: "timeline.date_entered_sort_day" },
    });
    fireEvent.change(
      screen.getByTestId(workbookFilterOperatorTestId(timelineSurface)),
      {
        target: { value: "range" },
      },
    );
    const lowerComparison = screen.getByRole("combobox", {
      name: "Lower-bound comparison",
    });
    const lowerValue = screen.getByTestId(
      gridFilterValueTestId(timelineSurface),
    );
    const upperComparison = screen.getByRole("combobox", {
      name: "Upper-bound comparison",
    });
    const upperValue = screen.getByRole("textbox", {
      name: "Upper-bound value",
    });
    upperValue.focus();
    for (const control of [
      lowerComparison,
      lowerValue,
      upperComparison,
      upperValue,
    ]) {
      control.focus();
      const tab = new KeyboardEvent("keydown", {
        bubbles: true,
        cancelable: true,
        key: "Tab",
      });
      control.dispatchEvent(tab);
      expect(tab.defaultPrevented).toBe(false);
      expect(document.activeElement).toBe(control);
    }
    fireEvent.change(lowerValue, { target: { value: "2026-01-01" } });
    fireEvent.change(upperValue, { target: { value: "2026-12-31" } });
    const apply = screen.getByTestId(gridFilterApplyTestId(timelineSurface));
    expect((apply as HTMLButtonElement).disabled).toBe(false);
    apply.focus();
    fireEvent.keyDown(apply, { key: "Tab" });
    expect(document.activeElement).toBe(field);
    fireEvent.keyDown(field, { key: "Tab", shiftKey: true });
    expect(document.activeElement).toBe(apply);
  });

  it("returns filter dismissal focus and reconciles controls removed by draft changes", () => {
    render(
      <>
        <StatefulGridControls />
        <button type="button">Outside destination</button>
      </>,
    );
    const trigger = screen.getByTestId(
      workbookFilterPopoverTriggerTestId(timelineSurface),
    );
    fireEvent.click(trigger);
    const field = screen.getByTestId(gridFilterFieldTestId(timelineSurface));
    fireEvent.change(field, { target: { value: "timeline.capture_state" } });
    const value = screen.getByTestId(gridFilterValueTestId(timelineSurface));
    value.focus();
    fireEvent.change(
      screen.getByRole("combobox", { name: "Equality operand kind" }),
      {
        target: { value: "null" },
      },
    );
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Cancel" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(document.activeElement).toBe(trigger);

    fireEvent.click(trigger);
    const outside = screen.getByRole("button", { name: "Outside destination" });
    fireEvent.blur(screen.getByTestId(gridFilterFieldTestId(timelineSurface)), {
      relatedTarget: outside,
    });
    outside.focus();
    expect(screen.queryByRole("dialog", { name: "Add filter" })).toBeNull();
    expect(document.activeElement).toBe(outside);
  });
  it("opens a focused filter chip for editing instead of removing it", () => {
    const contract = requireViewContract(timelineSurface);
    const onRemoveFilter = vi.fn();
    render(
      <WorkbookGridControls
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
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
    fireEvent.click(
      screen.getByTestId(
        workbookQueryOverflowEntryTestId(
          timelineSurface,
          "filter",
          "timeline.tags",
        ),
      ),
    );
    expect(document.activeElement).toBe(
      screen.getByTestId(workbookFilterOperatorTestId(timelineSurface)),
    );
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
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
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

  it("keeps focus in Sort when Add replaces its control and a boundary move disables", () => {
    const contract = requireViewContract(timelineSurface);
    const [first, second] = contract.sortFields;
    if (first === undefined || second === undefined)
      throw new Error("Timeline needs two sortable fields");
    const firstLabel = contract.fieldMap[first]?.label ?? first;
    const secondLabel = contract.fieldMap[second]?.label ?? second;
    render(<StatefulGridControls />);
    fireEvent.click(
      screen.getByTestId(workbookSortMenuTriggerTestId(timelineSurface)),
    );
    const firstAdd = screen.getByRole("menuitem", {
      name: `Add sort ${firstLabel}`,
    });
    firstAdd.focus();
    fireEvent.click(firstAdd);
    const firstDirection = screen.getByRole("menuitemcheckbox", {
      name: `Set ${firstLabel} descending`,
    });
    expect(document.activeElement).toBe(firstDirection);
    const secondAdd = screen.getByRole("menuitem", {
      name: `Add sort ${secondLabel}`,
    });
    secondAdd.focus();
    fireEvent.click(secondAdd);
    const secondDirection = screen.getByRole("menuitemcheckbox", {
      name: `Set ${secondLabel} descending`,
    });
    expect(document.activeElement).toBe(secondDirection);
    const moveEarlier = screen.getByRole("menuitem", {
      name: `Move ${secondLabel} earlier`,
    });
    moveEarlier.focus();
    fireEvent.click(moveEarlier);
    expect(document.activeElement).toBe(secondDirection);
    expect((moveEarlier as HTMLButtonElement).disabled).toBe(true);
  });

  it("keeps the roving Sort item aligned with pointer focus and removes toward a neighbor", () => {
    const contract = requireViewContract(timelineSurface);
    const [first, second] = contract.sortFields;
    if (first === undefined || second === undefined)
      throw new Error("Timeline needs two sortable fields");
    const firstLabel = contract.fieldMap[first]?.label ?? first;
    const secondLabel = contract.fieldMap[second]?.label ?? second;
    render(
      <StatefulGridControls
        initialSort={[
          { fieldKey: first, direction: "asc" },
          { fieldKey: second, direction: "asc" },
        ]}
      />,
    );
    fireEvent.click(
      screen.getByTestId(workbookSortMenuTriggerTestId(timelineSurface)),
    );
    const removeFirst = screen.getByRole("menuitem", {
      name: `Remove ${firstLabel} sort`,
    });
    act(() => removeFirst.focus());
    expect(removeFirst.tabIndex).toBe(0);
    expect(
      screen
        .getByTestId(workbookSortMenuTestId(timelineSurface))
        .querySelectorAll('[role^="menuitem"][tabindex="0"]'),
    ).toHaveLength(1);
    fireEvent.keyDown(removeFirst, { key: "ArrowDown" });
    expect(document.activeElement).toBe(
      screen.getByRole("menuitemcheckbox", {
        name: `Set ${secondLabel} descending`,
      }),
    );
    act(() => removeFirst.focus());
    fireEvent.keyDown(removeFirst, { key: "Home" });
    expect(document.activeElement).toBe(
      screen.getByRole("menuitemcheckbox", {
        name: `Set ${firstLabel} descending`,
      }),
    );
    removeFirst.focus();
    fireEvent.click(removeFirst);
    expect(document.activeElement).toBe(
      screen.getByRole("menuitemcheckbox", {
        name: `Set ${secondLabel} descending`,
      }),
    );
  });

  it("returns from the final sort to Add and from a removed invoking chip to Sort", () => {
    const contract = requireViewContract(timelineSurface);
    const first = contract.sortFields[0];
    if (first === undefined) throw new Error("Timeline needs a sortable field");
    const label = contract.fieldMap[first]?.label ?? first;
    render(
      <StatefulGridControls
        initialSort={[{ fieldKey: first, direction: "asc" }]}
      />,
    );
    const trigger = screen.getByTestId(
      workbookSortMenuTriggerTestId(timelineSurface),
    );
    const chip = screen.getByTestId(
      workbookQueryEntryTestId(timelineSurface, "sort", first),
    );
    fireEvent.click(chip);
    const direction = screen.getByRole("menuitemcheckbox", {
      name: `Set ${label} descending`,
    });
    expect(document.activeElement).toBe(direction);
    const remove = screen.getByRole("menuitem", {
      name: `Remove ${label} sort`,
    });
    act(() => remove.focus());
    fireEvent.click(remove);
    const add = screen.getByRole("menuitem", { name: `Add sort ${label}` });
    expect(document.activeElement).toBe(add);
    fireEvent.keyDown(add, { key: "Escape" });
    expect(document.activeElement).toBe(trigger);
  });

  it("shows requested Sort controls before acceptance and leaves late updates outside focus", () => {
    const contract = requireViewContract(timelineSurface);
    const [first, second] = contract.sortFields;
    if (first === undefined || second === undefined)
      throw new Error("Timeline needs two sortable fields");
    const firstLabel = contract.fieldMap[first]?.label ?? first;
    const secondLabel = contract.fieldMap[second]?.label ?? second;
    const empty = emptyWorkbookQueryState();
    const one = [{ fieldKey: first, direction: "asc" as const }];
    const two = [...one, { fieldKey: second, direction: "asc" as const }];
    const onSortChange = vi.fn();
    const { rerender } = render(
      <>
        <ControlledSortGridControls
          accepted={empty}
          requestedSort={[]}
          onSortChange={onSortChange}
          subjectKey="timeline"
        />
        <button type="button">Outside destination</button>
      </>,
    );
    const trigger = screen.getByTestId(
      workbookSortMenuTriggerTestId(timelineSurface),
    );
    fireEvent.click(trigger);
    const addFirst = screen.getByRole("menuitem", {
      name: `Add sort ${firstLabel}`,
    });
    act(() => addFirst.focus());
    fireEvent.click(addFirst);
    expect(onSortChange).toHaveBeenLastCalledWith(one);
    rerender(
      <>
        <ControlledSortGridControls
          accepted={empty}
          requestedSort={one}
          onSortChange={onSortChange}
          subjectKey="timeline"
        />
        <button type="button">Outside destination</button>
      </>,
    );
    const firstDirection = screen.getByRole("menuitemcheckbox", {
      name: `Set ${firstLabel} descending`,
    });
    expect(document.activeElement).toBe(firstDirection);
    expect(
      screen.queryByTestId(
        workbookQueryEntryTestId(timelineSurface, "sort", first),
      ),
    ).toBeNull();
    expect(
      screen.getByText(
        "Sort changes are unapplied until the query is accepted.",
      ),
    ).toBeTruthy();
    const addSecond = screen.getByRole("menuitem", {
      name: `Add sort ${secondLabel}`,
    });
    act(() => addSecond.focus());
    fireEvent.click(addSecond);
    expect(onSortChange).toHaveBeenLastCalledWith(two);
    rerender(
      <>
        <ControlledSortGridControls
          accepted={empty}
          requestedSort={two}
          onSortChange={onSortChange}
          subjectKey="timeline"
        />
        <button type="button">Outside destination</button>
      </>,
    );
    const secondDirection = screen.getByRole("menuitemcheckbox", {
      name: `Set ${secondLabel} descending`,
    });
    expect(document.activeElement).toBe(secondDirection);
    rerender(
      <>
        <ControlledSortGridControls
          accepted={{ ...empty }}
          requestedSort={two}
          onSortChange={onSortChange}
          subjectKey="timeline"
        />
        <button type="button">Outside destination</button>
      </>,
    );
    expect(document.activeElement).toBe(secondDirection);
    expect(
      screen.getByText(
        "Sort changes are unapplied until the query is accepted.",
      ),
    ).toBeTruthy();
    const outside = screen.getByRole("button", { name: "Outside destination" });
    act(() => outside.focus());
    expect(
      screen.queryByTestId(workbookSortMenuTestId(timelineSurface)),
    ).toBeNull();
    // A failed/retained replacement can be reverted after the user has left.
    rerender(
      <>
        <ControlledSortGridControls
          accepted={empty}
          requestedSort={[]}
          onSortChange={onSortChange}
          subjectKey="timeline"
        />
        <button type="button">Outside destination</button>
      </>,
    );
    expect(document.activeElement).toBe(outside);
    rerender(
      <>
        <ControlledSortGridControls
          accepted={{ ...empty, sort: two }}
          requestedSort={two}
          onSortChange={onSortChange}
          subjectKey="timeline"
        />
        <button type="button">Outside destination</button>
      </>,
    );
    expect(document.activeElement).toBe(outside);
    rerender(
      <>
        <ControlledSortGridControls
          accepted={{ ...empty, sort: two }}
          requestedSort={two}
          onSortChange={onSortChange}
          subjectKey="next-surface"
        />
        <button type="button">Outside destination</button>
      </>,
    );
    expect(document.activeElement).toBe(outside);
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
      <WorkbookGridControls
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
        {...common}
        surface={timelineSurface}
      />,
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
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
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
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
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

  it("validates explicit sizing and restores one default with predictable cancellation focus", () => {
    render(<StatefulGridControls />);
    const trigger = screen.getByRole("button", { name: "Columns" });
    fireEvent.click(trigger);
    const firstLabel =
      requireViewContract(timelineSurface).fields[0]?.label ?? "";
    const widthLabel = `Width for ${firstLabel}`;
    fireEvent.click(screen.getByRole("button", { name: widthLabel }));
    const input = screen.getByRole("textbox", { name: "Width in CSS pixels" });
    expect(document.activeElement).toBe(input);
    for (const value of ["", "39", "4097", "40.5", "invalid"]) {
      fireEvent.change(input, { target: { value } });
      fireEvent.click(screen.getByRole("button", { name: "Apply width" }));
      expect(screen.getByRole("alert").textContent).toContain("40 to 4096");
      expect(screen.getByText(/Current: 240 px/)).toBeTruthy();
      expect((input as HTMLInputElement).value).toBe(value);
    }
    for (const value of ["40", "4096", "240"]) {
      fireEvent.change(input, { target: { value } });
      fireEvent.click(screen.getByRole("button", { name: "Apply width" }));
      expect(screen.queryByRole("alert")).toBeNull();
      expect(
        screen.getByText(new RegExp(`Current: ${value} px.*Custom width`)),
      ).toBeTruthy();
    }
    fireEvent.click(screen.getByRole("button", { name: "Restore default" }));
    expect(screen.queryByText(/Custom width/)).toBeNull();
    expect(screen.getByText(/Current: 240 px/)).toBeTruthy();
    fireEvent.change(input, { target: { value: "unapplied" } });
    fireEvent.keyDown(input, { key: "Escape" });
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: widthLabel }),
    );
    fireEvent.click(screen.getByRole("button", { name: widthLabel }));
    expect(
      (
        screen.getByRole("textbox", {
          name: "Width in CSS pixels",
        }) as HTMLInputElement
      ).value,
    ).toBe("240");
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    fireEvent.keyDown(screen.getByRole("dialog", { name: "Column controls" }), {
      key: "Escape",
    });
    expect(document.activeElement).toBe(trigger);
    expect(
      screen.queryByRole("dialog", { name: "Column controls" }),
    ).toBeNull();
  });

  it("keeps column commands available when every data column is hidden", () => {
    const contract = requireViewContract(timelineSurface);
    const layout = defaultWorkbookLayoutState(contract);
    const onColumnHiddenChange = vi.fn();
    const onColumnMove = vi.fn();
    const onResetColumns = vi.fn();
    render(
      <WorkbookGridControls
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
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
    const visibilityItems = screen.getAllByRole("checkbox");
    expect(visibilityItems).toHaveLength(layout.columnOrder.length);
    expect(
      visibilityItems.every((item) => !(item as HTMLInputElement).checked),
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
        screen.getByRole("button", {
          name: `Move ${firstLabel} earlier`,
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(
      (
        screen.getByRole("button", {
          name: `Move ${lastLabel} later`,
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    const firstVisibilityItem = visibilityItems[0];
    expect(firstVisibilityItem).toBeDefined();
    if (firstVisibilityItem === undefined) return;
    fireEvent.click(firstVisibilityItem);
    expect(onColumnHiddenChange).toHaveBeenCalledWith(first, false);
    fireEvent.click(screen.getByRole("button", { name: "Reset columns" }));
    expect(onResetColumns).toHaveBeenCalledOnce();
  });
});

function StatefulGridControls({
  initialSort = [],
}: {
  readonly initialSort?: WorkbookQueryState["sort"];
}) {
  const contract = requireViewContract(timelineSurface);
  const [queryState, setQueryState] = useState<WorkbookQueryState>({
    ...emptyWorkbookQueryState(),
    sort: initialSort,
  });
  const owner = useWorkbookColumnLayoutController({ activeContract: contract });
  const controls = owner.snapshot.activeLayoutControls;
  const layoutState = controls.layoutState;
  useLayoutEffect(
    () =>
      controls.bindColumnSizing({ defaultWidth: () => 240, port: undefined }),
    [controls.bindColumnSizing],
  );
  const [filterDraft, setFilterDraft] = useState(() =>
    defaultFilterDraft(contract),
  );
  return (
    <WorkbookGridControls
      freezing={controls.freezing}
      sizing={controls.sizing}
      contract={contract}
      filterDraft={filterDraft}
      layoutState={layoutState}
      onApplyFilter={() => undefined}
      onClearFilters={() => {
        setQueryState((current) => ({ ...current, filters: [] }));
      }}
      onColumnHiddenChange={controls.onColumnHiddenChange}
      onColumnMove={controls.onColumnMove}
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
      onResetColumns={controls.onResetColumns}
      onSortChange={(sort) => {
        setQueryState((current) => ({ ...current, sort }));
      }}
      queryState={queryState}
      surface={timelineSurface}
    />
  );
}

function ControlledSortGridControls({
  accepted,
  onSortChange,
  requestedSort,
  subjectKey,
}: {
  readonly accepted: WorkbookQueryState;
  readonly onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  readonly requestedSort: WorkbookQueryState["sort"];
  readonly subjectKey: string;
}) {
  const contract = requireViewContract(timelineSurface);
  return (
    <WorkbookGridControls
      contract={contract}
      filterDraft={defaultFilterDraft(contract)}
      freezing={{ status: null, onBoundaryChange: vi.fn() }}
      layoutState={defaultWorkbookLayoutState(contract)}
      onApplyFilter={vi.fn()}
      onColumnHiddenChange={vi.fn()}
      onColumnMove={vi.fn()}
      onFilterDraftChange={vi.fn()}
      onGroupByChange={vi.fn()}
      onRemoveFilter={vi.fn()}
      onResetColumns={vi.fn()}
      onSortChange={onSortChange}
      queryState={accepted}
      requestedSort={requestedSort}
      sizing={sizing}
      subjectKey={subjectKey}
      surface={timelineSurface}
    />
  );
}

const sizing = {
  read: () => ({
    defaultWidth: 240,
    width: 240,
    overridden: false,
    unavailableReason: "Column measurement is unavailable.",
  }),
  onIntent: vi.fn(),
  restoreDefault: vi.fn(),
  cancel: vi.fn(),
  pendingField: null,
  notice: null,
};
