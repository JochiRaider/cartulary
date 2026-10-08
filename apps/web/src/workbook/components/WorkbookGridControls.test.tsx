import type {
  GridColumnMeasurement,
  GridColumnSizingPort,
} from "@cartulary/grid-adapter";
import {
  gridFilterApplyTestId,
  gridFilterFieldTestId,
  gridFilterValueTestId,
  gridGroupingSelectTestId,
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
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useLayoutEffect, useState } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { useWorkbookColumnLayoutController } from "../layout/useWorkbookColumnLayoutController";
import { defaultWorkbookLayoutState } from "../layout/workbookColumnLayout";
import { workbookOrderedSortLimit } from "../models/workbookGridQueryControls";
import {
  type buildFilterFromDraft,
  clearFilterDraftValue,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  filterDraftForField,
  validateFilterDraft,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { WorkbookCandidateQueryControl } from "./WorkbookCandidateQueryControl";
import { WorkbookGridControls } from "./WorkbookGridControls";

const timelineSurface = "cartulary.view.timeline.v2";
const admitFilter = (draft: FilterDraft) =>
  validateFilterDraft(requireViewContract(timelineSurface), draft);

afterEach(() => {
  cleanup();
  vi.restoreAllMocks();
});

describe("WorkbookGridControls", () => {
  it("keeps literal member correction native input identity and add remove focus", async () => {
    const user = userEvent.setup(),
      applied = vi.fn();
    render(
      <FilterGridControls
        applied={applied}
        initial={{
          ...emptyWorkbookQueryState(),
          filters: [
            {
              fieldKey: "timeline.tags",
              op: "contains_any",
              arg: { values: ["review,priority"] },
            },
          ],
        }}
      />,
    );
    const chip = screen.getByTestId(
      workbookQueryEntryTestId(timelineSurface, "filter", "timeline.tags"),
    );
    await user.click(chip);
    const first = screen.getByRole("textbox", {
      name: "Value 1",
    }) as HTMLTextAreaElement;
    await user.clear(first);
    await user.paste('東京 "quoted",  value');
    expect(screen.getByRole("textbox", { name: "Value 1" })).toBe(first);
    expect(first.value).toBe('東京 "quoted",  value');
    fireEvent.compositionStart(first);
    fireEvent.keyDown(first, { key: "Escape", isComposing: true });
    expect(screen.getByRole("dialog")).toBeTruthy();
    fireEvent.compositionEnd(first);
    for (const key of [
      "Home",
      "End",
      "ArrowLeft",
      "ArrowRight",
      "Delete",
      "Backspace",
      "Tab",
    ]) {
      const event = new KeyboardEvent("keydown", {
        key,
        bubbles: true,
        cancelable: true,
      });
      first.dispatchEvent(event);
      expect(event.defaultPrevented, key).toBe(false);
    }
    await user.click(screen.getByRole("button", { name: "Add value" }));
    const second = screen.getByRole("textbox", {
      name: "Value 2",
    }) as HTMLTextAreaElement;
    expect(document.activeElement).toBe(second);
    expect(
      (screen.getByRole("button", { name: "Apply" }) as HTMLButtonElement)
        .disabled,
    ).toBe(true);
    await user.paste("review\npriority");
    expect(second.value).toBe("review\npriority");
    expect(second.getAttribute("aria-invalid")).toBe("true");
    expect(
      document.getElementById(second.getAttribute("aria-describedby") ?? "")
        ?.textContent,
    ).toContain("Edit or remove");
    expect(applied).not.toHaveBeenCalled();
    expect(chip.textContent).toContain("review,priority");
    await user.clear(second);
    await user.type(second, "priority");
    await user.click(screen.getByRole("button", { name: "Remove value 1" }));
    expect(document.activeElement).toBe(second);
    expect(screen.getByRole("textbox", { name: "Value 1" })).toBe(second);
    await user.click(screen.getByRole("button", { name: "Remove value 1" }));
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Add value" }),
    );
    await user.click(screen.getByRole("button", { name: "Add value" }));
    await user.type(
      screen.getByRole("textbox", { name: "Value 1" }),
      "review,priority",
    );
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied).toHaveBeenCalledWith({
      fieldKey: "timeline.tags",
      op: "contains_any",
      arg: { values: ["review,priority"] },
    });
    expect(document.activeElement).toBe(chip);
  });
  it("reopens literal sets without changing member boundaries in either operator", async () => {
    const user = userEvent.setup();
    for (const [view, fieldKey, op, value] of [
      [timelineSurface, "timeline.tags", "contains_any", "review,priority"],
      [timelineSurface, "timeline.tags", "contains_all", "review,priority"],
      [
        "cartulary.view.parties.v1",
        "party.organization_name",
        "eq",
        "Northwind, Inc.",
      ],
    ] as const) {
      const applied = vi.fn();
      const filter = {
        fieldKey,
        op,
        arg: { values: [value] },
      };
      render(
        <FilterGridControls
          contract={requireViewContract(view)}
          applied={applied}
          initial={{ ...emptyWorkbookQueryState(), filters: [filter] }}
        />,
      );
      await user.click(
        screen.getByTestId(workbookQueryEntryTestId(view, "filter", fieldKey)),
      );
      await user.click(screen.getByRole("button", { name: "Apply" }));
      expect(applied).toHaveBeenCalledWith(filter);
      cleanup();
    }
  });

  it("stages a comma-containing candidate member before explicit query application", async () => {
    const user = userEvent.setup();
    for (const [view, fieldKey, op, value] of [
      [timelineSurface, "timeline.tags", "contains_any", "review,priority"],
      [timelineSurface, "timeline.tags", "contains_all", "review,priority"],
      [
        "cartulary.view.parties.v1",
        "party.organization_name",
        "eq",
        "Northwind, Inc.",
      ],
    ] as const) {
      const applied = vi.fn();
      render(
        <WorkbookCandidateQueryControl
          view={view}
          label="Linked Records"
          query={emptyWorkbookQueryState()}
          onApply={applied}
        />,
      );
      await user.click(screen.getByText("Linked Records ordering and filters"));
      await user.selectOptions(
        screen.getByLabelText("Linked Records filter field"),
        fieldKey,
      );
      await user.selectOptions(
        screen.getByLabelText("Linked Records filter operator"),
        op,
      );
      if (op === "eq")
        await user.selectOptions(
          screen.getByLabelText("Linked Records equality operand"),
          "values",
        );
      await user.type(screen.getByRole("textbox"), value);
      await user.click(screen.getByRole("button", { name: "Add value" }));
      const second = screen.getByRole("textbox", {
        name: "Linked Records filter value 2",
      }) as HTMLTextAreaElement;
      expect(document.activeElement).toBe(second);
      await user.paste("invalid\nmember");
      expect(second.value).toBe("invalid\nmember");
      expect(second.getAttribute("aria-invalid")).toBe("true");
      await user.click(screen.getByRole("button", { name: "Add filter" }));
      expect(
        screen.queryByRole("button", { name: /Remove filter / }),
      ).toBeNull();
      expect(applied).not.toHaveBeenCalled();
      await user.click(
        screen.getByRole("button", {
          name: "Remove linked records filter value 2",
        }),
      );
      await user.click(screen.getByRole("button", { name: "Add filter" }));
      expect(applied).not.toHaveBeenCalled();
      expect(
        screen.getByRole("button", { name: /Remove filter / }).parentElement
          ?.textContent,
      ).toContain(JSON.stringify([value]));
      await user.click(
        screen.getByRole("button", { name: "Apply candidate query" }),
      );
      expect(applied.mock.calls[0]?.[0].filters).toEqual([
        { fieldKey, op, arg: { values: [value] } },
      ]);
      cleanup();
    }
  });
  it("keeps invalid timestamp drafts editable with operand guidance native keys and accepted filters", async () => {
    const user = userEvent.setup(),
      applied = vi.fn();
    const surface = "cartulary.view.notes.v1";
    const contract = requireViewContract(surface);
    const initial: WorkbookQueryState = {
      ...emptyWorkbookQueryState(),
      filters: [
        {
          fieldKey: "note.updated_at",
          op: "eq",
          arg: { value: "2026-04-18T00:00:00Z" },
        },
      ],
    };
    render(
      <FilterGridControls
        contract={contract}
        applied={applied}
        initial={initial}
      />,
    );
    const chip = screen.getByRole("button", { name: /Filter 1, Updated/ });
    await user.click(chip);
    const value = screen.getByTestId(
      gridFilterValueTestId(surface),
    ) as HTMLInputElement;
    await user.clear(value);
    await user.type(value, " tomorrow ");
    const apply = screen.getByRole("button", { name: "Apply" });
    expect((apply as HTMLButtonElement).disabled).toBe(true);
    await user.click(apply);
    expect(applied).not.toHaveBeenCalled();
    expect(value.value).toBe(" tomorrow ");
    expect(value.getAttribute("aria-invalid")).toBe("true");
    expect(
      document.getElementById(value.getAttribute("aria-describedby") ?? "")
        ?.textContent,
    ).toContain("2026-04-18T00:00:00Z");
    expect(screen.getByRole("dialog")).toBeTruthy();
    for (const key of [
      "Home",
      "End",
      "ArrowLeft",
      "ArrowRight",
      "Delete",
      "Backspace",
      "Tab",
    ]) {
      const event = new KeyboardEvent("keydown", {
        key,
        bubbles: true,
        cancelable: true,
      });
      value.dispatchEvent(event);
      expect(event.defaultPrevented, key).toBe(false);
    }
    expect(chip.textContent).toContain("2026-04-18T00:00:00Z");
    await user.clear(value);
    await user.type(value, "2026-04-17T20:00:00.000000001-04:00");
    await user.click(apply);
    expect(applied).toHaveBeenCalledWith({
      fieldKey: "note.updated_at",
      op: "eq",
      arg: { value: "2026-04-17T20:00:00.000000001-04:00" },
    });
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: /Filter 1, Updated/ }),
    );
    await user.click(screen.getByRole("button", { name: /Filter 1, Updated/ }));
    await user.selectOptions(screen.getByLabelText("Operator"), "range");
    const lower = screen.getByLabelText("Lower-bound value"),
      upper = screen.getByLabelText("Upper-bound value");
    await user.type(lower, "2026-04-18T00:00:00.1Z");
    await user.type(upper, "2026-04-18T00:00:00Z");
    expect(lower.getAttribute("aria-invalid")).toBe("true");
    expect(upper.getAttribute("aria-describedby")).toBe(
      lower.getAttribute("aria-describedby"),
    );
    expect(
      (
        screen.getByRole("button", {
          name: "Apply",
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(applied).toHaveBeenCalledTimes(1);
  });

  it("reopens boolean sets and null with typed explicit choices", async () => {
    const user = userEvent.setup();
    const applied = vi.fn();
    render(
      <FilterGridControls
        applied={applied}
        initial={{
          ...emptyWorkbookQueryState(),
          filters: [
            {
              fieldKey: "timeline.has_evidence",
              op: "eq",
              arg: { values: [false] },
            },
          ],
        }}
      />,
    );
    await user.click(
      screen.getByRole("button", {
        name: "Filter 1, Has Evidence, equals [false]",
      }),
    );
    const falseChoice = screen.getByRole("checkbox", { name: "false" });
    expect((falseChoice as HTMLInputElement).checked).toBe(true);
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied.mock.calls[0]?.[0]?.arg).toEqual({ values: [false] });
    await user.click(
      screen.getByRole("button", {
        name: "Filter 1, Has Evidence, equals [false]",
      }),
    );
    await user.selectOptions(
      screen.getByLabelText("Equality operand kind"),
      "null",
    );
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied.mock.calls[1]?.[0]?.arg).toEqual({ value: null });
    await user.click(
      screen.getByRole("button", { name: /Filter 1, Has Evidence/ }),
    );
    await user.selectOptions(
      screen.getByLabelText("Equality operand kind"),
      "value",
    );
    const value = screen.getByRole("combobox", { name: "Value" });
    expect((value as HTMLSelectElement).value).toBe("");
    await user.selectOptions(value, "false");
    expect(applied).toHaveBeenCalledTimes(2);
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied.mock.calls[2]?.[0]?.arg).toEqual({ value: false });
  });
  it("keeps boolean feedback local and reconciles native keyboard focus without applying", async () => {
    const user = userEvent.setup();
    const applied = vi.fn();
    render(
      <FilterGridControls
        applied={applied}
        initial={{
          ...emptyWorkbookQueryState(),
          filters: [
            {
              fieldKey: "timeline.has_evidence",
              op: "eq",
              arg: { values: ["false", true] },
            },
          ],
        }}
      />,
    );
    const chip = screen.getByRole("button", { name: /Filter 1, Has Evidence/ });
    await user.click(chip);
    const apply = screen.getByRole("button", { name: "Apply" });
    const trueChoice = screen.getByRole("checkbox", { name: "true" });
    const falseChoice = screen.getByRole("checkbox", { name: "false" });
    expect((trueChoice as HTMLInputElement).checked).toBe(false);
    expect(apply.hasAttribute("disabled")).toBe(true);
    expect(falseChoice.getAttribute("aria-invalid")).toBe("true");
    expect(
      document.getElementById(
        falseChoice.getAttribute("aria-describedby") ?? "",
      )?.textContent,
    ).toContain("restored filter");
    falseChoice.focus();
    await user.keyboard(" ");
    expect(apply.hasAttribute("disabled")).toBe(false);
    await user.tab({ shift: true });
    expect(document.activeElement).toBe(trueChoice);
    await user.tab();
    expect(document.activeElement).toBe(falseChoice);
    expect(applied).not.toHaveBeenCalled();
    const matching = screen.getByLabelText("Equality operand kind");
    fireEvent.change(matching, { target: { value: "null" } });
    expect(document.activeElement).toBe(matching);
    await user.selectOptions(matching, "value");
    const scalar = screen.getByRole("combobox", { name: "Value" });
    expect((scalar as HTMLSelectElement).value).toBe("");
    expect(apply.hasAttribute("disabled")).toBe(true);
    scalar.focus();
    const arrow = new KeyboardEvent("keydown", {
      bubbles: true,
      cancelable: true,
      key: "ArrowDown",
    });
    scalar.dispatchEvent(arrow);
    expect(arrow.defaultPrevented).toBe(false);
    await user.keyboard("{Escape}");
    expect(document.activeElement).toBe(chip);
    expect(applied).not.toHaveBeenCalled();
    cleanup();
    render(
      <WorkbookCandidateQueryControl
        view={timelineSurface}
        label="Candidates"
        query={emptyWorkbookQueryState()}
        onApply={applied}
      />,
    );
    await user.click(screen.getByText("Candidates ordering and filters"));
    await user.selectOptions(
      screen.getByLabelText("Candidates filter field"),
      "timeline.has_evidence",
    );
    const add = screen.getByRole("button", { name: "Add filter" });
    expect(add.hasAttribute("disabled")).toBe(true);
    await user.selectOptions(
      screen.getByLabelText("Candidates equality operand"),
      "values",
    );
    const choices = screen.getAllByRole("checkbox");
    choices[0]?.focus();
    await user.tab();
    expect(document.activeElement).toBe(choices[1]);
    await user.keyboard(" ");
    await user.tab();
    expect(document.activeElement).toBe(add);
    await user.click(add);
    const candidateMode = screen.getByLabelText("Candidates equality operand");
    choices[1]?.focus();
    fireEvent.change(candidateMode, { target: { value: "null" } });
    expect(document.activeElement).toBe(candidateMode);
    const candidateApply = screen.getByRole("button", {
      name: "Apply candidate query",
    });
    candidateApply.focus();
    fireEvent.change(candidateMode, { target: { value: "value" } });
    expect(document.activeElement).toBe(candidateApply);
    expect(applied).not.toHaveBeenCalled();
    await user.click(
      screen.getByRole("button", { name: "Apply candidate query" }),
    );
    expect(applied.mock.calls[0]?.[0]?.filters).toEqual([
      { fieldKey: "timeline.has_evidence", op: "eq", arg: { values: [false] } },
    ]);
  });

  it("offers ordered enum choices without implicit admission and preserves null and set shapes", async () => {
    const user = userEvent.setup();
    const applied = vi.fn();
    render(<FilterGridControls applied={applied} />);
    const trigger = screen.getByTestId(
      workbookFilterPopoverTriggerTestId(timelineSurface),
    );
    await user.click(trigger);
    await user.selectOptions(
      screen.getByLabelText("Field"),
      "timeline.activity_time_pair_state",
    );
    const value = screen.getByLabelText("Value");
    const choices =
      requireViewContract(timelineSurface).fieldMap[
        "timeline.activity_time_pair_state"
      ]?.enumValues ?? [];
    expect(
      within(value)
        .getAllByRole("option")
        .map((option) => option.textContent),
    ).toEqual(["Choose a value", ...choices]);
    expect((value as HTMLSelectElement).value).toBe("");
    expect(
      screen.getByRole("button", { name: "Apply" }).hasAttribute("disabled"),
    ).toBe(true);
    await user.selectOptions(value, "empty");
    expect(applied).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied.mock.calls[0]?.[0]?.arg).toEqual({ value: "empty" });
    expect(document.activeElement).toBe(trigger);
    await user.click(
      screen.getByRole("button", {
        name: "Filter 1, Activity Time Pair State, equals empty",
      }),
    );
    expect((screen.getByLabelText("Value") as HTMLSelectElement).value).toBe(
      "empty",
    );
    await user.selectOptions(
      screen.getByLabelText("Equality operand kind"),
      "values",
    );
    const checks = screen.getAllByRole("checkbox");
    expect(checks.map((check) => check.parentElement?.textContent)).toEqual(
      choices,
    );
    checks[0]?.focus();
    await user.keyboard(" ");
    await user.click(screen.getByRole("checkbox", { name: "empty" }));
    await user.click(screen.getByRole("checkbox", { name: "empty" }));
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied.mock.calls[1]?.[0]?.arg).toEqual({ values: ["disabled"] });
    await user.click(trigger);
    expect(
      screen
        .getByRole("button", { name: "Custom literals" })
        .getAttribute("aria-expanded"),
    ).toBe("false");
    await user.click(screen.getByRole("checkbox", { name: "empty" }));
    expect(
      (screen.getByRole("button", { name: "Apply" }) as HTMLButtonElement)
        .disabled,
    ).toBe(false);
    await user.click(screen.getByRole("button", { name: "Cancel" }));

    await user.click(
      screen.getByRole("button", {
        name: 'Filter 1, Activity Time Pair State, equals ["disabled"]',
      }),
    );
    expect(
      (screen.getByRole("checkbox", { name: "disabled" }) as HTMLInputElement)
        .checked,
    ).toBe(true);
    await user.click(screen.getByRole("checkbox", { name: "disabled" }));
    expect(
      screen.getByRole("button", { name: "Apply" }).hasAttribute("disabled"),
    ).toBe(true);
    await user.selectOptions(
      screen.getByLabelText("Equality operand kind"),
      "null",
    );
    expect(screen.queryByRole("checkbox")).toBeNull();
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied.mock.calls[2]?.[0]?.arg).toEqual({ value: null });
  });

  it("reopens custom enum sets keeps literal slots and reconciles the complete keyboard order", async () => {
    const user = userEvent.setup();
    const applied = vi.fn();
    render(
      <FilterGridControls
        applied={applied}
        initial={{
          filters: [
            {
              fieldKey: "timeline.activity_time_pair_state",
              op: "eq",
              arg: { values: ["disabled", "Disabled", "custom,token"] },
            },
          ],
          groupBy: null,
          sort: [],
        }}
      />,
    );
    await user.click(
      screen.getByRole("button", {
        name: /Filter 1, Activity Time Pair State/,
      }),
    );
    expect(
      screen
        .getByRole("button", { name: "Custom literals" })
        .getAttribute("aria-expanded"),
    ).toBe("true");
    expect(
      (screen.getByLabelText("Value literal 3") as HTMLInputElement).value,
    ).toBe("custom,token");
    await user.click(screen.getByRole("checkbox", { name: "empty" }));
    expect(
      (screen.getByLabelText("Value literal 2") as HTMLInputElement).value,
    ).toBe("Disabled");
    await user.clear(screen.getByLabelText("Value literal 3"));
    await user.type(screen.getByLabelText("Value literal 3"), "empty");
    expect(screen.getByLabelText("Value literal 3")).toBe(
      document.activeElement,
    );
    await user.click(screen.getByRole("button", { name: "Add literal" }));
    expect(screen.getByLabelText("Value literal 5")).toBeTruthy();
    await user.click(screen.getByRole("button", { name: "Remove literal 5" }));
    expect(document.activeElement?.isConnected).toBe(true);
    const add = screen.getByRole("button", { name: "Add literal" });
    add.focus();
    await user.tab();
    expect(document.activeElement?.getAttribute("aria-label")).toContain(
      "Edit Filter",
    );
    expect(applied).not.toHaveBeenCalled();
    await user.click(screen.getByRole("button", { name: "Apply" }));
    expect(applied.mock.calls[0]?.[0]?.arg.values).toEqual(
      expect.arrayContaining(["disabled", "Disabled", "empty"]),
    );
    expect(applied.mock.calls[0]?.[0]?.arg.values).toHaveLength(3);
    await user.click(
      screen.getByRole("button", {
        name: /Filter 1, Activity Time Pair State/,
      }),
    );
    await user.keyboard("{Escape}");
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(applied).toHaveBeenCalledTimes(1);
  });

  it("scopes enum disclosure by field view and instance and leaves fallback and prefix controls intact", async () => {
    const user = userEvent.setup();
    const apply = vi.fn();
    const contract = requireViewContract(timelineSurface);
    render(
      <>
        <FilterGridControls applied={apply} />
        <WorkbookCandidateQueryControl
          view="cartulary.view.parties.v1"
          label="Other"
          query={emptyWorkbookQueryState()}
          onApply={vi.fn()}
        />
      </>,
    );
    await user.click(
      screen.getByTestId(workbookFilterPopoverTriggerTestId(timelineSurface)),
    );
    await user.selectOptions(
      screen.getByLabelText("Field"),
      "timeline.activity_time_pair_state",
    );
    await user.click(screen.getByRole("button", { name: "Custom literals" }));
    await user.type(screen.getByLabelText("Value literal"), "Disabled");
    await user.selectOptions(
      screen.getByLabelText("Field"),
      "timeline.capture_state",
    );
    expect(
      screen.queryByRole("button", { name: "Custom literals" }),
    ).toBeNull();
    expect(
      screen.getByTestId(gridFilterValueTestId(timelineSurface)).tagName,
    ).toBe("INPUT");
    await user.selectOptions(screen.getByLabelText("Operator"), "prefix");
    expect(
      screen.getByTestId(gridFilterValueTestId(timelineSurface)).tagName,
    ).toBe("INPUT");
    await user.selectOptions(
      screen.getByLabelText("Field"),
      "timeline.activity_time_pair_state",
    );
    expect(
      (
        screen.getByTestId(
          gridFilterValueTestId(timelineSurface),
        ) as HTMLSelectElement
      ).value,
    ).toBe("");
    expect(
      screen
        .getByRole("button", { name: "Custom literals" })
        .getAttribute("aria-expanded"),
    ).toBe("false");
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    const other = screen.getByText("Other ordering and filters");
    await user.click(other);
    await user.selectOptions(
      screen.getByLabelText("Other filter field"),
      "party.party_kind",
    );
    expect(
      within(screen.getByLabelText("Other filter value"))
        .getAllByRole("option")
        .map((o) => o.textContent),
    ).toEqual([
      "Choose a value",
      "person",
      "team",
      "organization",
      "distribution_list",
      "other",
    ]);
    await user.selectOptions(
      screen.getByLabelText("Other filter operator"),
      "prefix",
    );
    expect(screen.getByLabelText("Other filter value").tagName).toBe("INPUT");
    cleanup();
    const field = contract.fieldMap["timeline.activity_time_pair_state"];
    if (!field) throw new Error("Missing field");
    render(
      <FilterGridControls
        applied={apply}
        contract={{
          ...contract,
          fieldMap: {
            ...contract.fieldMap,
            [field.fieldKey]: { ...field, enumValues: [] },
          },
        }}
      />,
    );
    await user.click(
      screen.getByTestId(workbookFilterPopoverTriggerTestId(timelineSurface)),
    );
    await user.selectOptions(screen.getByLabelText("Field"), field.fieldKey);
    expect(
      screen.getByTestId(gridFilterValueTestId(timelineSurface)).tagName,
    ).toBe("INPUT");
    expect(apply).not.toHaveBeenCalled();
  });

  it("stages candidate enum choices only on Add and applies custom mixtures explicitly", async () => {
    const user = userEvent.setup();
    const apply = vi.fn();
    render(
      <WorkbookCandidateQueryControl
        view={timelineSurface}
        label="Candidates"
        query={emptyWorkbookQueryState()}
        onApply={apply}
      />,
    );
    await user.click(screen.getByText("Candidates ordering and filters"));
    await user.selectOptions(
      screen.getByLabelText("Candidates filter field"),
      "timeline.activity_time_pair_state",
    );
    await user.selectOptions(
      screen.getByLabelText("Candidates equality operand"),
      "values",
    );
    await user.click(screen.getByRole("checkbox", { name: "disabled" }));
    await user.click(screen.getByRole("button", { name: "Custom literals" }));
    await user.click(screen.getByRole("button", { name: "Add literal" }));
    await user.type(
      screen.getByLabelText("Candidates filter value literal 2"),
      "Disabled,custom",
    );
    expect(apply).not.toHaveBeenCalled();
    await user.click(
      screen.getByRole("button", { name: "Apply candidate query" }),
    );
    expect(apply.mock.calls[0]?.[0]?.filters).toEqual([]);
    await user.click(screen.getByRole("button", { name: "Add filter" }));
    expect(apply).toHaveBeenCalledTimes(1);
    await user.click(
      screen.getByRole("button", { name: "Apply candidate query" }),
    );
    expect(apply.mock.calls[1]?.[0]?.filters[0]?.arg.values).toEqual(
      expect.arrayContaining(["disabled", "Disabled,custom"]),
    );
  });
  it("keeps impossible date correction in the editor with associated feedback", () => {
    const contract = requireViewContract(timelineSurface);
    const onApplyFilter = vi.fn(admitFilter);
    render(
      <WorkbookGridControls
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        contract={contract}
        surface={timelineSurface}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        queryState={emptyWorkbookQueryState()}
        onApplyFilter={onApplyFilter}
        onFilterDraftChange={vi.fn()}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onGroupByChange={vi.fn()}
        onRemoveFilter={vi.fn()}
        onResetColumns={vi.fn()}
        onSortChange={vi.fn()}
        sizing={sizing}
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
      />,
    );
    const trigger = screen.getByTestId(
      workbookFilterPopoverTriggerTestId(timelineSurface),
    );
    fireEvent.click(trigger);
    fireEvent.change(
      screen.getByTestId(gridFilterFieldTestId(timelineSurface)),
      { target: { value: "timeline.date_entered_sort_day" } },
    );
    const value = screen.getByTestId(gridFilterValueTestId(timelineSurface));
    fireEvent.change(value, { target: { value: " 2026-04-31 " } });
    const apply = screen.getByTestId(gridFilterApplyTestId(timelineSurface));
    expect((apply as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(apply);
    expect(onApplyFilter).not.toHaveBeenCalled();
    expect((value as HTMLInputElement).value).toBe(" 2026-04-31 ");
    expect(value.getAttribute("aria-invalid")).toBe("true");
    expect(
      document.getElementById(value.getAttribute("aria-describedby") ?? "")
        ?.textContent,
    ).toContain("YYYY-MM-DD");
    expect(screen.getByRole("dialog", { name: "Add filter" })).toBeTruthy();
    fireEvent.change(value, { target: { value: "2026-04-18" } });
    expect((apply as HTMLButtonElement).disabled).toBe(false);
    fireEvent.keyDown(value, { key: "Escape" });
    expect(document.activeElement).toBe(trigger);
  });

  it("names range operands associates contradictions and completes correction before reopening", () => {
    render(<StatefulGridControls />);
    const trigger = screen.getByTestId(
      workbookFilterPopoverTriggerTestId(timelineSurface),
    );
    fireEvent.click(trigger);
    fireEvent.change(
      screen.getByTestId(gridFilterFieldTestId(timelineSurface)),
      { target: { value: "timeline.date_entered_sort_day" } },
    );
    fireEvent.change(
      screen.getByTestId(workbookFilterOperatorTestId(timelineSurface)),
      { target: { value: "range" } },
    );
    const lower = screen.getByRole("textbox", {
      name: "Lower-bound value",
    }) as HTMLInputElement;
    const upper = screen.getByRole("textbox", {
      name: "Upper-bound value",
    }) as HTMLInputElement;
    const lowerKind = screen.getByRole("combobox", {
      name: "Lower-bound comparison",
    });
    const upperKind = screen.getByRole("combobox", {
      name: "Upper-bound comparison",
    });
    const apply = screen.getByTestId(
      gridFilterApplyTestId(timelineSurface),
    ) as HTMLButtonElement;
    fireEvent.change(lower, { target: { value: "2026-04-19" } });
    fireEvent.change(upper, { target: { value: "2026-04-18" } });
    const feedbackId = lower.getAttribute("aria-describedby");
    expect(upper.getAttribute("aria-describedby")).toBe(feedbackId);
    expect(document.getElementById(feedbackId ?? "")?.textContent).toContain(
      "on or before",
    );
    expect(apply.disabled).toBe(true);
    lower.focus();
    const tab = new KeyboardEvent("keydown", {
      key: "Tab",
      bubbles: true,
      cancelable: true,
    });
    lower.dispatchEvent(tab);
    expect(tab.defaultPrevented).toBe(false);
    fireEvent.change(lower, { target: { value: "2026-04-18" } });
    fireEvent.change(lowerKind, { target: { value: "gt" } });
    for (const control of [lower, upper, lowerKind, upperKind]) {
      expect(control.getAttribute("aria-invalid")).toBe("true");
      expect(control.getAttribute("aria-describedby")).toBe(feedbackId);
    }
    expect(document.getElementById(feedbackId ?? "")?.textContent).toContain(
      "inclusive",
    );
    fireEvent.change(lowerKind, { target: { value: "gte" } });
    expect(lower.hasAttribute("aria-invalid")).toBe(false);
    expect(apply.disabled).toBe(false);
    fireEvent.click(apply);
    expect(screen.queryByRole("dialog")).toBeNull();
    expect(document.activeElement).toBe(trigger);
    fireEvent.click(trigger);
    expect(
      (
        screen.getByRole("textbox", {
          name: "Lower-bound value",
        }) as HTMLInputElement
      ).value,
    ).toBe("");
    expect(
      (
        screen.getByRole("textbox", {
          name: "Upper-bound value",
        }) as HTMLInputElement
      ).value,
    ).toBe("");
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(document.activeElement).toBe(trigger);
  });

  it("retains a locally valid editor when the admission callback refuses it", () => {
    const contract = requireViewContract(timelineSurface);
    const onApplyFilter = vi.fn(() => ({
      kind: "invalid" as const,
      message: "Subject changed",
      controls: ["field" as const],
    }));
    render(
      <WorkbookGridControls
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        contract={contract}
        surface={timelineSurface}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        queryState={emptyWorkbookQueryState()}
        onApplyFilter={onApplyFilter}
        onFilterDraftChange={vi.fn()}
        onColumnHiddenChange={vi.fn()}
        onColumnMove={vi.fn()}
        onGroupByChange={vi.fn()}
        onRemoveFilter={vi.fn()}
        onResetColumns={vi.fn()}
        onSortChange={vi.fn()}
        sizing={sizing}
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
      />,
    );
    fireEvent.click(
      screen.getByTestId(workbookFilterPopoverTriggerTestId(timelineSurface)),
    );
    fireEvent.change(
      screen.getByTestId(gridFilterFieldTestId(timelineSurface)),
      { target: { value: "timeline.date_entered_sort_day" } },
    );
    const value = screen.getByRole("textbox", {
      name: "Date value",
    }) as HTMLInputElement;
    fireEvent.change(value, { target: { value: " 2026-04-18 " } });
    fireEvent.click(screen.getByTestId(gridFilterApplyTestId(timelineSurface)));
    expect(onApplyFilter).toHaveBeenCalledOnce();
    expect(screen.getByRole("dialog", { name: "Add filter" })).toBeTruthy();
    expect(value.value).toBe(" 2026-04-18 ");
    fireEvent.change(
      screen.getByTestId(gridFilterFieldTestId(timelineSurface)),
      { target: { value: "timeline.has_evidence" } },
    );
    const boolean = screen.getByRole("combobox", {
      name: "Value",
    }) as HTMLSelectElement;
    fireEvent.change(boolean, { target: { value: "false" } });
    fireEvent.click(screen.getByTestId(gridFilterApplyTestId(timelineSurface)));
    expect(onApplyFilter).toHaveBeenCalledTimes(2);
    expect(screen.getByRole("dialog", { name: "Add filter" })).toBeTruthy();
    expect(boolean.value).toBe("false");
  });

  it("excludes invalid date candidates from staging and keeps feedback instance local", () => {
    const onApply = vi.fn();
    render(
      <>
        <WorkbookCandidateQueryControl
          view={timelineSurface}
          label="First"
          query={emptyWorkbookQueryState()}
          onApply={onApply}
        />
        <WorkbookCandidateQueryControl
          view={timelineSurface}
          label="Second"
          query={emptyWorkbookQueryState()}
          onApply={vi.fn()}
        />
      </>,
    );
    for (const summary of screen.getAllByText(/ordering and filters/))
      fireEvent.click(summary);
    fireEvent.change(
      screen.getByRole("combobox", { name: "First filter field" }),
      { target: { value: "timeline.date_entered_sort_day" } },
    );
    const first = screen.getByRole("textbox", {
      name: "First filter value",
    }) as HTMLInputElement;
    const second = screen.getByRole("textbox", { name: "Second filter value" });
    fireEvent.change(first, { target: { value: " 2026-04-31 " } });
    expect(first.getAttribute("aria-describedby")).not.toBe(
      second.getAttribute("aria-describedby"),
    );
    const add = screen.getAllByRole("button", {
      name: "Add filter",
      hidden: true,
    })[0] as HTMLButtonElement;
    const applyCandidate = screen.getAllByRole("button", {
      name: "Apply candidate query",
      hidden: true,
    })[0];
    if (!applyCandidate) throw new Error("Missing candidate query action");
    expect(add.disabled).toBe(true);
    fireEvent.click(add);
    fireEvent.click(applyCandidate);
    expect(onApply).toHaveBeenLastCalledWith(emptyWorkbookQueryState());
    expect(first.value).toBe(" 2026-04-31 ");
    fireEvent.change(first, { target: { value: "2026-04-18" } });
    fireEvent.click(add);
    fireEvent.click(applyCandidate);
    expect(onApply.mock.calls.at(-1)?.[0].filters).toEqual([
      {
        fieldKey: "timeline.date_entered_sort_day",
        op: "eq",
        arg: { value: "2026-04-18" },
      },
    ]);
  });

  it("separates accepted chips from requested filter editing and removal", () => {
    const contract = requireViewContract(timelineSurface);
    const accepted = {
      filters: [
        {
          fieldKey: "timeline.date_entered_sort_day",
          op: "range" as const,
          arg: { gte: "2026-04-01" },
        },
      ],
      groupBy: "timeline.capture_state",
      sort: [
        { fieldKey: "timeline.activity_sort_ts", direction: "desc" as const },
      ],
    };
    const requested = [
      {
        fieldKey: "timeline.date_entered_sort_day",
        op: "range" as const,
        arg: { gte: "2026-01-01" },
      },
    ];
    const onApplyFilter = vi.fn(admitFilter);
    const onClearFilters = vi.fn();
    const onRemoveFilter = vi.fn();
    const props = {
      contract,
      filterDraft: defaultFilterDraft(contract),
      layoutState: defaultWorkbookLayoutState(contract),
      onApplyFilter,
      onClearFilters,
      onColumnHiddenChange: vi.fn(),
      onColumnMove: vi.fn(),
      onFilterDraftChange: vi.fn(),
      onGroupByChange: vi.fn(),
      onRemoveFilter,
      onResetColumns: vi.fn(),
      onSortChange: vi.fn(),
      queryState: accepted,
      surface: timelineSurface,
      freezing: { status: null, onBoundaryChange: vi.fn() },
      sizing,
    };
    const { rerender } = render(
      <WorkbookGridControls
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        {...props}
        requestedFilters={requested}
      />,
    );
    const trigger = screen.getByTestId(
      workbookFilterPopoverTriggerTestId(timelineSurface),
    );
    expect(trigger.textContent).toContain("Unapplied");
    fireEvent.click(trigger);
    fireEvent.click(
      screen.getByRole("button", { name: /Edit unapplied.*Date entered/i }),
    );
    expect(
      screen.getByRole("dialog", { name: "Edit unapplied filter" }),
    ).toBeTruthy();
    const lower = screen.getByTestId(gridFilterValueTestId(timelineSurface));
    expect((lower as HTMLInputElement).value).toBe("2026-01-01");
    fireEvent.change(lower, { target: { value: " 2026-04-31 " } });
    fireEvent.click(screen.getByTestId(gridFilterApplyTestId(timelineSurface)));
    expect((lower as HTMLInputElement).value).toBe(" 2026-04-31 ");
    expect(
      screen.getByRole("dialog", { name: "Edit unapplied filter" }),
    ).toBeTruthy();
    fireEvent.change(lower, { target: { value: "2026-02-01" } });
    fireEvent.keyDown(lower, { key: "Escape" });
    expect(onApplyFilter).not.toHaveBeenCalled();
    const chip = screen.getByTestId(
      workbookQueryEntryTestId(
        timelineSurface,
        "filter",
        "timeline.date_entered_sort_day",
      ),
    );
    fireEvent.click(chip);
    expect(screen.getByRole("dialog", { name: "Edit filter" })).toBeTruthy();
    expect(
      (
        screen.getByTestId(
          gridFilterValueTestId(timelineSurface),
        ) as HTMLInputElement
      ).value,
    ).toBe("2026-04-01");
    fireEvent.keyDown(
      screen.getByTestId(gridFilterValueTestId(timelineSurface)),
      { key: "Escape" },
    );

    rerender(
      <WorkbookGridControls
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        {...props}
        requestedFilters={[]}
      />,
    );
    expect(chip.isConnected).toBe(true);
    fireEvent.click(trigger);
    const restore = screen.getByRole("button", { name: /Restore Date/i });
    expect(restore.parentElement?.textContent).toContain("Unapplied");
    expect(
      (
        screen.getByRole("button", {
          name: "Clear filters",
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    fireEvent.click(restore);
    expect(onApplyFilter).toHaveBeenCalledWith({
      fieldKey: "timeline.date_entered_sort_day",
      op: "range",
      lowerKind: "gte",
      lowerValue: "2026-04-01",
      upperKind: "lte",
      upperValue: "",
    });
    expect(onClearFilters).not.toHaveBeenCalled();

    rerender(
      <WorkbookGridControls
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        {...props}
        requestedFilters={[
          ...accepted.filters,
          {
            fieldKey: "timeline.tags",
            op: "contains_any",
            arg: { values: ["review"] },
          },
        ]}
      />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: /Remove unapplied.*Tags/i }),
    );
    expect(onRemoveFilter).toHaveBeenCalledWith("timeline.tags");
  });

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
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={vi.fn(admitFilter)}
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
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={defaultWorkbookLayoutState(contract)}
        onApplyFilter={vi.fn(admitFilter)}
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

  it("keeps requested Group choices distinct from accepted chips through delay, failure, retry, Revert, and None", () => {
    const acceptedField = "timeline.capture_state";
    const firstChoice = "timeline.has_evidence";
    const latestChoice = "timeline.has_unresolved_mentions";
    const accepted = {
      ...emptyWorkbookQueryState(),
      groupBy: acceptedField,
    };
    const onGroupByChange = vi.fn();
    const renderControls = (
      applied: WorkbookQueryState,
      requestedGroupBy: string | null,
    ) => (
      <ControlledGroupGridControls
        accepted={applied}
        onGroupByChange={onGroupByChange}
        requestedGroupBy={requestedGroupBy}
        subjectKey="incident:timeline:base"
      />
    );
    const { rerender } = render(renderControls(accepted, acceptedField));
    const grouping = screen.getByTestId(
      gridGroupingSelectTestId(timelineSurface),
    ) as HTMLSelectElement;
    const acceptedChipId = workbookQueryEntryTestId(
      timelineSurface,
      "group",
      acceptedField,
    );
    const statusId = `${gridGroupingSelectTestId(timelineSurface)}-unapplied`;
    act(() => grouping.focus());
    fireEvent.change(grouping, { target: { value: firstChoice } });
    expect(onGroupByChange).toHaveBeenLastCalledWith(firstChoice);
    rerender(renderControls(accepted, firstChoice));
    expect(grouping.value).toBe(firstChoice);
    expect(document.activeElement).toBe(grouping);
    expect(screen.getByTestId(acceptedChipId)).toBeTruthy();
    expect(document.getElementById(statusId)?.textContent).toContain(
      "retained results grouped by Capture State",
    );
    expect(grouping.title).toBe("Has Evidence");

    fireEvent.change(grouping, { target: { value: latestChoice } });
    expect(onGroupByChange).toHaveBeenLastCalledWith(latestChoice);
    rerender(renderControls(accepted, latestChoice));
    expect(grouping.value).toBe(latestChoice);
    expect(screen.getByTestId(acceptedChipId)).toBeTruthy();
    rerender(renderControls({ ...accepted }, latestChoice));
    expect(grouping.value).toBe(latestChoice);
    expect(document.activeElement).toBe(grouping);

    const latestAccepted = { ...accepted, groupBy: latestChoice };
    rerender(renderControls(latestAccepted, latestChoice));
    expect(document.getElementById(statusId)).toBeNull();
    expect(screen.queryByTestId(acceptedChipId)).toBeNull();
    expect(
      screen.getByTestId(
        workbookQueryEntryTestId(timelineSurface, "group", latestChoice),
      ),
    ).toBeTruthy();
    expect(document.activeElement).toBe(grouping);

    rerender(renderControls(latestAccepted, null));
    expect(grouping.value).toBe("");
    expect(document.getElementById(statusId)?.textContent).toContain(
      "Requested None; retained results grouped by Has Unresolved Mentions",
    );
    rerender(renderControls({ ...latestAccepted }, null));
    expect(grouping.value).toBe("");
    rerender(renderControls(latestAccepted, latestChoice));
    expect(grouping.value).toBe(latestChoice);
    expect(document.getElementById(statusId)).toBeNull();
    rerender(renderControls({ ...latestAccepted, groupBy: null }, null));
    expect(grouping.value).toBe("");
    expect(document.getElementById(statusId)).toBeNull();
    expect(
      screen.queryByTestId(
        workbookQueryEntryTestId(timelineSurface, "group", latestChoice),
      ),
    ).toBeNull();
    expect(document.activeElement).toBe(grouping);
  });

  it("does not restore a retired Group subject's focus after Escape", async () => {
    const accepted = {
      ...emptyWorkbookQueryState(),
      groupBy: "timeline.capture_state",
    };
    const onGroupByChange = vi.fn();
    const { rerender } = render(
      <>
        <ControlledGroupGridControls
          accepted={accepted}
          onGroupByChange={onGroupByChange}
          requestedGroupBy={accepted.groupBy}
          subjectKey="incident:timeline:saved-one"
        />
        <button type="button">New context destination</button>
      </>,
    );
    const chip = screen.getByTestId(
      workbookQueryEntryTestId(
        timelineSurface,
        "group",
        "timeline.capture_state",
      ),
    );
    fireEvent.click(chip);
    const grouping = screen.getByTestId(
      gridGroupingSelectTestId(timelineSurface),
    );
    expect(document.activeElement).toBe(grouping);
    fireEvent.keyDown(grouping, { key: "Escape" });
    rerender(
      <>
        <ControlledGroupGridControls
          accepted={emptyWorkbookQueryState()}
          onGroupByChange={onGroupByChange}
          requestedGroupBy={null}
          subjectKey="incident:timeline:saved-two"
        />
        <button type="button">New context destination</button>
      </>,
    );
    const outside = screen.getByRole("button", {
      name: "New context destination",
    });
    act(() => outside.focus());
    await act(async () => undefined);
    expect(document.activeElement).toBe(outside);
    expect((grouping as HTMLSelectElement).value).toBe("");
    expect(
      screen.queryByTestId(chip.getAttribute("data-testid") ?? ""),
    ).toBeNull();
  });

  it("keeps invalid drafts visible, excludes them from apply, and resets panels by surface", () => {
    const contract = requireViewContract(timelineSurface);
    const onApplyFilter = vi.fn(admitFilter);
    const invalidDraft: FilterDraft = {
      fieldKey: "Capture State",
      op: "eq",
      operandKind: "value",
      value: "reviewed",
      valueType: "string",
      values: [],
    };
    const common = {
      contract,
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
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
        {...common}
        surface={timelineSurface}
      />,
    );
    fireEvent.click(
      screen.getByTestId(workbookFilterPopoverTriggerTestId(timelineSurface)),
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
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
        {...common}
        surface="cartulary.view.hosts.v1"
      />,
    );
    expect(
      screen.queryByTestId(workbookFilterPopoverTestId(timelineSurface)),
    ).toBeNull();
  });

  it("parses filter controls exactly and restores focus on Escape", () => {
    const contract = requireViewContract(timelineSurface);
    const onApplyFilter = vi.fn(admitFilter);
    const onFilterDraftChange = vi.fn();
    render(
      <WorkbookGridControls
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
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
      booleanOperand: { value: false, values: [] },
      fieldKey: "timeline.has_evidence",
      op: "eq",
      operandKind: "value",
      value: "",
      valueType: "boolean",
      values: [],
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

  it("keeps the fitted width in the open input when Apply follows Fit", async () => {
    const { port, requests } = deferredSizingPort();
    render(<StatefulGridControls sizingPort={port} />);
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const label = requireViewContract(timelineSurface).fields[0]?.label ?? "";
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const input = screen.getByRole("textbox", {
      name: "Width in CSS pixels",
    }) as HTMLInputElement;
    expect(input.value).toBe("240");
    fireEvent.click(
      screen.getByRole("button", { name: "Fit visible content" }),
    );
    await act(async () => {
      requests[0]?.resolve({
        kind: "measured",
        widthPx: 520,
        capped: false,
        cellCount: 2,
      });
    });
    expect(screen.getByText(/Current: 520 px/)).toBeTruthy();
    expect(input.value).toBe("520");
    fireEvent.click(screen.getByRole("button", { name: "Apply width" }));
    expect(screen.getByText(/Current: 520 px/)).toBeTruthy();
  });

  it("reveals the same Width action and retains list context after Escape and Cancel", () => {
    for (const scale of [1, 2]) {
      const mounted = render(<StatefulGridControls />);
      fireEvent.click(screen.getByRole("button", { name: "Columns" }));
      const panel = screen.getByRole("dialog", { name: "Column controls" });
      const geometry = columnsGeometry(panel, scale);
      const label =
        requireViewContract(timelineSurface).fieldMap[
          "timeline.has_unresolved_mentions"
        ]?.label;
      for (const dismiss of ["Escape", "Cancel"]) {
        panel.scrollTop = 1300;
        fireEvent.click(
          screen.getByRole("button", { name: `Width for ${label}` }),
        );
        // jsdom does not clamp scroll when the list is replaced by the short form.
        panel.scrollTop = 0;
        if (dismiss === "Escape")
          fireEvent.keyDown(
            screen.getByRole("textbox", { name: "Width in CSS pixels" }),
            { key: "Escape" },
          );
        else fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
        const target = screen.getByRole("button", {
          name: `Width for ${label}`,
        });
        expect(document.activeElement).toBe(target);
        expect(target.getBoundingClientRect().top).toBeGreaterThanOrEqual(100);
        expect(target.getBoundingClientRect().bottom).toBeLessThanOrEqual(
          100 + 300 * scale,
        );
        expect(panel.scrollTop).toBe(1300);
      }
      geometry.mockRestore();
      mounted.unmount();
    }
  });

  it("reveals moved controls that keep focus and the early Unfreeze destination", () => {
    render(<StatefulGridControls />);
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const panel = screen.getByRole("dialog", { name: "Column controls" });
    const geometry = columnsGeometry(panel);
    const first = requireViewContract(timelineSurface).fields[0]?.label;
    const later = screen.getByRole("button", { name: `Move ${first} later` });
    later.focus();
    for (let step = 0; step < 8; step += 1) {
      fireEvent.click(later);
      expect(document.activeElement).toBe(later);
      expect(later.getBoundingClientRect().bottom).toBeLessThanOrEqual(400);
    }
    const freeze = screen.getByRole("button", {
      name: `Freeze through ${first}`,
    });
    freeze.focus();
    fireEvent.click(freeze);
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: `Width for ${first}` }),
    );
    panel.scrollTop = 1300;
    const unfreeze = screen.getByRole("button", { name: "Unfreeze columns" });
    unfreeze.focus();
    fireEvent.click(unfreeze);
    expect(document.activeElement).toBe(freeze);
    expect(freeze.getBoundingClientRect().top).toBeGreaterThanOrEqual(100);
    geometry.mockRestore();
  });

  it("retires Columns return work on outside focus and same-surface subject replacement", async () => {
    const { port, requests } = deferredSizingPort();
    const view = (subjectKey: string) => (
      <>
        <button type="button">Outside Columns</button>
        <StatefulGridControls subjectKey={subjectKey} sizingPort={port} />
      </>
    );
    const mounted = render(view("first"));
    const label = requireViewContract(timelineSurface).fields[0]?.label;
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const outside = screen.getByRole("button", { name: "Outside Columns" });
    act(() => {
      fireEvent.keyDown(
        screen.getByRole("textbox", { name: "Width in CSS pixels" }),
        { key: "Escape" },
      );
      outside.focus();
    });
    expect(document.activeElement).toBe(outside);
    expect(
      screen.queryByRole("dialog", { name: "Column controls" }),
    ).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const fit = screen.getByRole("button", { name: "Fit visible content" });
    fit.focus();
    fireEvent.click(fit);
    mounted.rerender(view("second"));
    outside.focus();
    expect(requests[0]?.signal.aborted).toBe(true);
    await act(async () =>
      requests[0]?.resolve({
        kind: "measured",
        widthPx: 500,
        capped: false,
        cellCount: 1,
      }),
    );
    expect(document.activeElement).toBe(outside);
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    expect(
      screen.queryByRole("textbox", { name: "Width in CSS pixels" }),
    ).toBeNull();
    expect(document.activeElement).toBe(screen.getAllByRole("checkbox")[0]);
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const nextFit = screen.getByRole("button", { name: "Fit visible content" });
    nextFit.focus();
    fireEvent.click(nextFit);
    mounted.rerender(<button type="button">After Columns unmount</button>);
    const afterUnmount = screen.getByRole("button", {
      name: "After Columns unmount",
    });
    afterUnmount.focus();
    expect(requests[1]?.signal.aborted).toBe(true);
    await act(async () =>
      requests[1]?.resolve({
        kind: "measured",
        widthPx: 600,
        capped: false,
        cellCount: 1,
      }),
    );
    expect(document.activeElement).toBe(afterUnmount);
  });

  it("keeps pending Fit focus and busy state without admitting a duplicate or reclaiming departed focus", async () => {
    const { port, requests } = deferredSizingPort();
    render(
      <>
        <button type="button">Outside Columns</button>
        <StatefulGridControls sizingPort={port} />
      </>,
    );
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const label = requireViewContract(timelineSurface).fields[0]?.label ?? "";
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const fit = screen.getByRole("button", { name: "Fit visible content" });
    fit.focus();
    fireEvent.click(fit);
    expect(fit.getAttribute("aria-busy")).toBe("true");
    expect((fit as HTMLButtonElement).disabled).toBe(false);
    expect(document.activeElement).toBe(fit);
    fireEvent.click(fit);
    expect(requests).toHaveLength(1);
    await act(async () => {
      requests[0]?.resolve({
        kind: "measured",
        widthPx: 520,
        capped: false,
        cellCount: 2,
      });
    });
    expect(document.activeElement).toBe(fit);
    expect(fit.getAttribute("aria-busy")).not.toBe("true");
    fireEvent.click(fit);
    const restore = screen.getByRole("button", { name: "Restore default" });
    restore.focus();
    await act(async () => {
      requests[1]?.resolve({
        kind: "unavailable",
        reason: "Measurement failed.",
      });
    });
    expect(document.activeElement).toBe(restore);
    expect(screen.getByRole("status").textContent).toContain(
      "Measurement failed.",
    );
    fit.focus();
    fireEvent.click(fit);
    const outside = screen.getByRole("button", { name: "Outside Columns" });
    outside.focus();
    fireEvent.pointerDown(outside);
    expect(
      screen.queryByRole("dialog", { name: "Column controls" }),
    ).toBeNull();
    expect(requests[2]?.signal.aborted).toBe(true);
    await act(async () => {
      requests[2]?.resolve({
        kind: "measured",
        widthPx: 700,
        capped: false,
        cellCount: 2,
      });
    });
    expect(document.activeElement).toBe(outside);
  });

  it("keeps pending Fit focused when availability changes and moves to Restore after it becomes unavailable", async () => {
    const { port, requests, setAvailability } = deferredSizingPort();
    render(<StatefulGridControls sizingPort={port} />);
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const label = requireViewContract(timelineSurface).fields[0]?.label ?? "";
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const panel = screen.getByRole("dialog", { name: "Column controls" });
    const geometry = columnsGeometry(panel, 1, 150);
    const fit = screen.getByRole("button", { name: "Fit visible content" });
    fit.focus();
    fireEvent.click(fit);
    act(() => setAvailability("Visible geometry is unavailable."));
    expect(fit.getAttribute("aria-busy")).toBe("true");
    expect((fit as HTMLButtonElement).disabled).toBe(false);
    expect(document.activeElement).toBe(fit);
    await act(async () => {
      requests[0]?.resolve({
        kind: "unavailable",
        reason: "Visible geometry is unavailable.",
      });
    });
    expect((fit as HTMLButtonElement).disabled).toBe(true);
    expect(fit.getAttribute("aria-busy")).not.toBe("true");
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Restore default" }),
    );
    expect(
      screen
        .getByRole("button", { name: "Restore default" })
        .getBoundingClientRect().bottom,
    ).toBeLessThanOrEqual(250);
    expect(
      document.getElementById(fit.getAttribute("aria-describedby") ?? "")
        ?.textContent,
    ).toBe("Visible geometry is unavailable.");
    act(() => setAvailability(null));
    fit.focus();
    fireEvent.click(fit);
    act(() => setAvailability("Visible geometry is unavailable."));
    const restore = screen.getByRole("button", { name: "Restore default" });
    restore.focus();
    await act(async () => {
      requests[1]?.resolve({
        kind: "unavailable",
        reason: "Visible geometry is unavailable.",
      });
    });
    expect(document.activeElement).toBe(restore);
    geometry.mockRestore();
  });

  it("keeps focus on a semantic column action when movement or freezing disables its invoker", () => {
    render(<StatefulGridControls />);
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const fields = requireViewContract(timelineSurface).fields;
    const first = fields[0]?.label;
    const last = fields.at(-1)?.label;
    if (!first || !last)
      throw new Error("Timeline column fixtures are required");
    const earlier = screen.getByRole("button", {
      name: `Move ${first} earlier`,
    });
    const later = screen.getByRole("button", { name: `Move ${first} later` });
    later.focus();
    fireEvent.click(later);
    expect(document.activeElement).toBe(later);
    earlier.focus();
    fireEvent.click(earlier);
    expect((earlier as HTMLButtonElement).disabled).toBe(true);
    expect(document.activeElement).toBe(later);
    const lastEarlier = screen.getByRole("button", {
      name: `Move ${last} earlier`,
    });
    const lastLater = screen.getByRole("button", {
      name: `Move ${last} later`,
    });
    lastEarlier.focus();
    fireEvent.click(lastEarlier);
    lastLater.focus();
    fireEvent.click(lastLater);
    expect((lastLater as HTMLButtonElement).disabled).toBe(true);
    expect(document.activeElement).toBe(lastEarlier);
    const freeze = screen.getByRole("button", {
      name: `Freeze through ${first}`,
    });
    freeze.focus();
    fireEvent.click(freeze);
    expect((freeze as HTMLButtonElement).disabled).toBe(true);
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: `Width for ${first}` }),
    );
    const unfreeze = screen.getByRole("button", { name: "Unfreeze columns" });
    unfreeze.focus();
    fireEvent.click(unfreeze);
    expect((unfreeze as HTMLButtonElement).disabled).toBe(true);
    expect(document.activeElement).toBe(freeze);
  });

  it("keeps newer invalid text and caret while an older Fit completes", async () => {
    const { port, requests } = deferredSizingPort();
    render(<StatefulGridControls sizingPort={port} />);
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const label = requireViewContract(timelineSurface).fields[0]?.label ?? "";
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const input = screen.getByRole("textbox", {
      name: "Width in CSS pixels",
    }) as HTMLInputElement;
    fireEvent.change(input, { target: { value: "410" } });
    fireEvent.click(
      screen.getByRole("button", { name: "Fit visible content" }),
    );
    expect(screen.getByText("Measuring visible content…")).toBeTruthy();
    input.focus();
    fireEvent.change(input, { target: { value: "5x" } });
    input.setSelectionRange(1, 1);
    await act(async () => {
      requests[0]?.resolve({
        kind: "measured",
        widthPx: 520,
        capped: false,
        cellCount: 2,
      });
    });
    expect(screen.getByText(/Current: 520 px/)).toBeTruthy();
    expect(input.value).toBe("5x");
    expect(document.activeElement).toBe(input);
    expect([input.selectionStart, input.selectionEnd]).toEqual([1, 1]);
    fireEvent.click(screen.getByRole("button", { name: "Apply width" }));
    expect(screen.getByRole("alert").textContent).toContain("40 to 4096");
    expect(screen.getByText(/Current: 520 px/)).toBeTruthy();
    fireEvent.change(input, { target: { value: "530" } });
    fireEvent.click(screen.getByRole("button", { name: "Apply width" }));
    expect(screen.getByText(/Current: 530 px/)).toBeTruthy();
    expect(input.value).toBe("530");
    fireEvent.click(
      screen.getByRole("button", { name: "Fit visible content" }),
    );
    fireEvent.change(input, { target: { value: "53x" } });
    fireEvent.click(screen.getByRole("button", { name: "Apply width" }));
    expect(requests[1]?.signal.aborted).toBe(true);
    await act(async () => {
      requests[1]?.resolve({
        kind: "measured",
        widthPx: 700,
        capped: false,
        cellCount: 2,
      });
    });
    expect(screen.getByText(/Current: 530 px/)).toBeTruthy();
    expect(input.value).toBe("53x");
  });

  it("replaces an older draft on Fit and follows passive width only while untouched", async () => {
    const { port, requests } = deferredSizingPort();
    const externalResize = {
      current: null as ((widthPx: number) => Promise<void>) | null,
    };
    render(
      <StatefulGridControls
        externalResize={externalResize}
        sizingPort={port}
      />,
    );
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const label = requireViewContract(timelineSurface).fields[0]?.label ?? "";
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    const input = screen.getByRole("textbox", {
      name: "Width in CSS pixels",
    }) as HTMLInputElement;
    expect(externalResize.current).not.toBeNull();
    await act(async () => {
      await externalResize.current?.(300);
    });
    expect(screen.getByText(/Current: 300 px/)).toBeTruthy();
    expect(input.value).toBe("300");
    fireEvent.change(input, { target: { value: "410" } });
    fireEvent.click(
      screen.getByRole("button", { name: "Fit visible content" }),
    );
    await act(async () => {
      requests[0]?.resolve({
        kind: "measured",
        widthPx: 520,
        capped: false,
        cellCount: 2,
      });
    });
    expect(input.value).toBe("520");
    await act(async () => {
      await externalResize.current?.(550);
    });
    expect(input.value).toBe("550");
    fireEvent.change(input, { target: { value: "55x" } });
    input.setSelectionRange(2, 2);
    await act(async () => {
      await externalResize.current?.(560);
    });
    expect(screen.getByText(/Current: 560 px/)).toBeTruthy();
    expect(input.value).toBe("55x");
    expect([input.selectionStart, input.selectionEnd]).toEqual([2, 2]);
    fireEvent.click(screen.getByRole("button", { name: "Restore default" }));
    expect(screen.getByText(/Current: 240 px/)).toBeTruthy();
    expect(input.value).toBe("240");
  });

  it("cancels pending Fit on panel departure without undoing completed width", async () => {
    const { port, requests } = deferredSizingPort();
    render(<StatefulGridControls sizingPort={port} />);
    fireEvent.click(screen.getByRole("button", { name: "Columns" }));
    const label = requireViewContract(timelineSurface).fields[0]?.label ?? "";
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    fireEvent.click(
      screen.getByRole("button", { name: "Fit visible content" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: `Width for ${label}` }),
    );
    await act(async () => {
      requests[0]?.resolve({
        kind: "measured",
        widthPx: 520,
        capped: false,
        cellCount: 2,
      });
    });
    fireEvent.click(screen.getByRole("button", { name: `Width for ${label}` }));
    expect(screen.getByText(/Current: 240 px/)).toBeTruthy();
    expect(
      (
        screen.getByRole("textbox", {
          name: "Width in CSS pixels",
        }) as HTMLInputElement
      ).value,
    ).toBe("240");
  });

  it("keeps column commands available when every data column is hidden", () => {
    const contract = requireViewContract(timelineSurface);
    const layout = defaultWorkbookLayoutState(contract);
    const onColumnHiddenChange = vi.fn();
    const onColumnMove = vi.fn();
    const onResetColumns = vi.fn();
    render(
      <WorkbookGridControls
        composeControls={({ query, columns }) => (
          <>
            {query}
            {columns}
          </>
        )}
        freezing={{ status: null, onBoundaryChange: vi.fn() }}
        sizing={sizing}
        contract={contract}
        filterDraft={defaultFilterDraft(contract)}
        layoutState={{ ...layout, hiddenFieldKeys: layout.columnOrder }}
        onApplyFilter={vi.fn(admitFilter)}
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

// A clipped scrollport model for jsdom; browser tests supply rendered geometry.
function columnsGeometry(panel: HTMLElement, scale = 1, height = 300) {
  for (const [key, value] of Object.entries({
    clientWidth: 400,
    offsetWidth: 400,
    clientHeight: height,
    offsetHeight: height,
  }))
    Object.defineProperty(panel, key, { configurable: true, value });
  const original = HTMLElement.prototype.getBoundingClientRect;
  return vi
    .spyOn(HTMLElement.prototype, "getBoundingClientRect")
    .mockImplementation(function (this: HTMLElement) {
      if (this === panel)
        return new DOMRect(100, 100, 400 * scale, height * scale);
      if (panel.contains(this) && this.tagName === "BUTTON") {
        const rows = [...panel.querySelectorAll('input[type="checkbox"]')].map(
          (input) => input.parentElement?.parentElement,
        );
        const row = rows.indexOf(this.parentElement);
        const position =
          row >= 0
            ? row
            : [...panel.querySelectorAll("form button")].indexOf(this);
        return new DOMRect(
          150,
          100 + (10 + position * 60 - panel.scrollTop) * scale,
          70 * scale,
          28 * scale,
        );
      }
      return original.call(this);
    });
}

function FilterGridControls({
  applied,
  initial = emptyWorkbookQueryState(),
  contract = requireViewContract(timelineSurface),
}: {
  readonly applied: (filter: ReturnType<typeof buildFilterFromDraft>) => void;
  readonly initial?: WorkbookQueryState;
  readonly contract?: ReturnType<typeof requireViewContract>;
}) {
  const [queryState, setQueryState] = useState(initial);
  const [draft, setDraft] = useState(() =>
    filterDraftForField(
      contract,
      contract.filterFields.includes("timeline.activity_time_pair_state")
        ? "timeline.activity_time_pair_state"
        : (contract.filterFields[0] ?? ""),
      "eq",
    ),
  );
  return (
    <WorkbookGridControls
      composeControls={({ query, columns }) => (
        <>
          {query}
          {columns}
        </>
      )}
      contract={contract}
      surface={contract.viewSchemaId}
      filterDraft={draft}
      queryState={queryState}
      layoutState={defaultWorkbookLayoutState(contract)}
      onApplyFilter={(next) => {
        const validation = validateFilterDraft(contract, next);
        if (validation.kind === "valid") {
          applied(validation.filter);
          setQueryState({ ...queryState, filters: [validation.filter] });
        }
        return validation;
      }}
      onFilterDraftChange={setDraft}
      onRemoveFilter={vi.fn()}
      onGroupByChange={vi.fn()}
      onColumnHiddenChange={vi.fn()}
      onColumnMove={vi.fn()}
      onResetColumns={vi.fn()}
      onSortChange={vi.fn()}
      sizing={sizing}
      freezing={{ status: null, onBoundaryChange: vi.fn() }}
    />
  );
}

function StatefulGridControls({
  initialSort = [],
  sizingPort,
  externalResize,
  subjectKey,
}: {
  readonly initialSort?: WorkbookQueryState["sort"];
  readonly sizingPort?: GridColumnSizingPort;
  readonly subjectKey?: string;
  readonly externalResize?: {
    current: ((widthPx: number) => Promise<void>) | null;
  };
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
      controls.bindColumnSizing({ defaultWidth: () => 240, port: sizingPort }),
    [controls.bindColumnSizing, sizingPort],
  );
  useLayoutEffect(() => {
    if (!externalResize) return;
    const fieldKey = contract.fields[0]?.fieldKey;
    if (!fieldKey) return;
    externalResize.current = (widthPx) =>
      controls.onColumnSizingIntent({ kind: "set_width", fieldKey, widthPx });
    return () => {
      externalResize.current = null;
    };
  }, [contract, controls.onColumnSizingIntent, externalResize]);
  const [filterDraft, setFilterDraft] = useState(() =>
    defaultFilterDraft(contract),
  );
  return (
    <WorkbookGridControls
      composeControls={({ query, columns }) => (
        <>
          {query}
          {columns}
        </>
      )}
      subjectKey={subjectKey}
      freezing={controls.freezing}
      sizing={controls.sizing}
      contract={contract}
      filterDraft={filterDraft}
      layoutState={layoutState}
      onApplyFilter={(draft) => {
        const validation = validateFilterDraft(contract, draft);
        if (validation.kind === "valid")
          setFilterDraft(clearFilterDraftValue(draft, contract));
        return validation;
      }}
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

function deferredSizingPort() {
  let listener: (() => void) | undefined;
  const availability = { current: null as string | null };
  const requests: {
    resolve: (result: GridColumnMeasurement) => void;
    signal: AbortSignal;
  }[] = [];
  const port: GridColumnSizingPort = {
    unavailableReason: () => availability.current,
    subscribe: (next) => {
      listener = next;
      return () => {
        listener = undefined;
      };
    },
    measureVisibleContent: (_field, { signal }) =>
      new Promise((resolve) => {
        requests.push({ resolve, signal });
      }),
  };
  return {
    port,
    requests,
    setAvailability: (reason: string | null) => {
      availability.current = reason;
      listener?.();
    },
  };
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
      composeControls={({ query, columns }) => (
        <>
          {query}
          {columns}
        </>
      )}
      contract={contract}
      filterDraft={defaultFilterDraft(contract)}
      freezing={{ status: null, onBoundaryChange: vi.fn() }}
      layoutState={defaultWorkbookLayoutState(contract)}
      onApplyFilter={vi.fn(admitFilter)}
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

function ControlledGroupGridControls({
  accepted,
  onGroupByChange,
  requestedGroupBy,
  subjectKey,
}: {
  readonly accepted: WorkbookQueryState;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly requestedGroupBy: string | null;
  readonly subjectKey: string;
}) {
  const contract = requireViewContract(timelineSurface);
  return (
    <WorkbookGridControls
      composeControls={({ query, columns }) => (
        <>
          {query}
          {columns}
        </>
      )}
      contract={contract}
      filterDraft={defaultFilterDraft(contract)}
      freezing={{ status: null, onBoundaryChange: vi.fn() }}
      layoutState={defaultWorkbookLayoutState(contract)}
      onApplyFilter={vi.fn(admitFilter)}
      onColumnHiddenChange={vi.fn()}
      onColumnMove={vi.fn()}
      onFilterDraftChange={vi.fn()}
      onGroupByChange={onGroupByChange}
      onRemoveFilter={vi.fn()}
      onResetColumns={vi.fn()}
      onSortChange={vi.fn()}
      queryState={accepted}
      requestedGroupBy={requestedGroupBy}
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
  applyWidth: vi.fn(() => ({ kind: "completed" as const, widthPx: 240 })),
  fitVisible: vi.fn(async () => ({ kind: "cancelled" as const })),
  restoreDefault: vi.fn(() => ({ kind: "completed" as const, widthPx: 240 })),
  cancel: vi.fn(),
  pendingField: null,
  notice: null,
};
