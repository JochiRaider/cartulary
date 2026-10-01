import { cleanup, fireEvent, screen, within } from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { renderWithWorkbookQueryBrowsing as render } from "../../testing/workbookQueryTestSupport";
import type { FilterDraft } from "../models/workbookQuery";
import { useWorkbookQueryController } from "./useWorkbookQueryController";

const impossibleDateDraft: FilterDraft = {
  booleanValue: "",
  fieldKey: "timeline.date_entered_sort_day",
  op: "eq",
  operandKind: "value",
  value: " 2026-04-31 ",
  valueType: "string",
  values: "",
};

afterEach(cleanup);

function QueryControllerHarness({
  instanceId,
}: {
  readonly instanceId: string;
}) {
  const controller = useWorkbookQueryController({
    surface: "cartulary.view.timeline.v2",
  });
  return (
    <section aria-label={`Workbook ${instanceId}`}>
      <button
        onClick={() => {
          controller.commands.setTimelineQueryState((current) => ({
            ...current,
            groupBy: "timeline.capture_state",
          }));
        }}
        type="button"
      >
        Update Timeline
      </button>
      <button
        onClick={() => {
          controller.commands.setHostQueryState((current) => ({
            ...current,
            groupBy: "host.entity_subtype",
          }));
        }}
        type="button"
      >
        Update Hosts
      </button>
      <output aria-label={`query-controller-state-${instanceId}`}>
        {JSON.stringify({
          hosts: controller.snapshot.hostQueryState.groupBy,
          timeline: controller.snapshot.timelineQueryState.groupBy,
        })}
      </output>
    </section>
  );
}

function FilterQueryHarness() {
  const controller = useWorkbookQueryController({
    surface: "cartulary.view.timeline.v2",
  });
  return (
    <section>
      <button
        type="button"
        onClick={() =>
          controller.snapshot.activeQueryControls.onFilterDraftChange(
            impossibleDateDraft,
          )
        }
      >
        Draft impossible date
      </button>
      <button
        type="button"
        onClick={() =>
          controller.snapshot.activeQueryControls.onApplyFilter(
            impossibleDateDraft,
          )
        }
      >
        Apply impossible date
      </button>
      <output aria-label="raw-filter-draft">
        {JSON.stringify(controller.snapshot.activeQueryControls.filterDraft)}
      </output>
      <button
        onClick={() => {
          controller.snapshot.activeQueryControls.onApplyFilter({
            fieldKey: "timeline.tags",
            op: "contains_any",
            values: "seed",
          });
          controller.commands.setTimelineQueryState((current) => ({
            ...current,
            groupBy: "timeline.capture_state",
            sort: [
              { fieldKey: "timeline.date_entered_sort_day", direction: "desc" },
            ],
          }));
        }}
        type="button"
      >
        Seed query
      </button>
      {(["2026-04-01", "2026-09-01"] as const).map((lowerValue) => (
        <button
          key={lowerValue}
          onClick={() =>
            controller.snapshot.activeQueryControls.onApplyFilter({
              fieldKey: "timeline.date_entered_sort_day",
              lowerKind: "gte",
              lowerValue,
              op: "range",
              upperKind: "lte",
              upperValue: "",
            })
          }
          type="button"
        >
          Apply {lowerValue}
        </button>
      ))}
      <output aria-label="requested-query">
        {JSON.stringify(controller.snapshot.timelineQueryState)}
      </output>
    </section>
  );
}

describe("useWorkbookQueryController", () => {
  it("refuses direct invalid date admission and preserves raw input and requested intent", () => {
    render(<FilterQueryHarness />);
    fireEvent.click(screen.getByRole("button", { name: "Seed query" }));
    fireEvent.click(
      screen.getByRole("button", { name: "Draft impossible date" }),
    );
    const requested = screen.getByLabelText("requested-query").textContent;
    fireEvent.click(
      screen.getByRole("button", { name: "Apply impossible date" }),
    );
    expect(screen.getByLabelText("requested-query").textContent).toBe(
      requested,
    );
    expect(
      JSON.parse(screen.getByLabelText("raw-filter-draft").textContent ?? "{}"),
    ).toEqual(impossibleDateDraft);
  });

  it("applies consecutive filter edits to the latest requested query", () => {
    render(<FilterQueryHarness />);
    fireEvent.click(screen.getByRole("button", { name: "Seed query" }));
    fireEvent.click(screen.getByRole("button", { name: "Apply 2026-04-01" }));
    fireEvent.click(screen.getByRole("button", { name: "Apply 2026-09-01" }));

    expect(
      JSON.parse(screen.getByLabelText("requested-query").textContent ?? "{}"),
    ).toEqual({
      filters: [
        {
          arg: { gte: "2026-09-01" },
          fieldKey: "timeline.date_entered_sort_day",
          op: "range",
        },
        {
          arg: { values: ["seed"] },
          fieldKey: "timeline.tags",
          op: "contains_any",
        },
      ],
      groupBy: "timeline.capture_state",
      sort: [{ fieldKey: "timeline.date_entered_sort_day", direction: "desc" }],
    });
  });

  it("keeps query state isolated by exact view_schema_id", () => {
    render(<QueryControllerHarness instanceId="one" />);
    fireEvent.click(screen.getByRole("button", { name: "Update Timeline" }));
    fireEvent.click(screen.getByRole("button", { name: "Update Hosts" }));

    expect(
      JSON.parse(
        screen.getByLabelText("query-controller-state-one").textContent ?? "{}",
      ),
    ).toEqual({
      hosts: "host.entity_subtype",
      timeline: "timeline.capture_state",
    });
  });

  it("keeps query defaults and updates isolated between Workbook instances", () => {
    render(
      <>
        <QueryControllerHarness instanceId="one" />
        <QueryControllerHarness instanceId="two" />
      </>,
    );
    const first = within(screen.getByRole("region", { name: "Workbook one" }));
    fireEvent.click(first.getByRole("button", { name: "Update Timeline" }));

    expect(
      JSON.parse(
        screen.getByLabelText("query-controller-state-one").textContent ?? "{}",
      ),
    ).toEqual({
      hosts: null,
      timeline: "timeline.capture_state",
    });
    expect(
      JSON.parse(
        screen.getByLabelText("query-controller-state-two").textContent ?? "{}",
      ),
    ).toEqual({
      hosts: null,
      timeline: null,
    });
  });
});
