import {
  act,
  cleanup,
  fireEvent,
  screen,
  within,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  acceptedQueryMetadata,
  renderWithWorkbookQueryBrowsing as render,
  renderHookWithWorkbookQueryBrowsing as renderHook,
} from "../../testing/workbookQueryTestSupport";
import type { FilterDraft } from "../models/workbookQuery";
import { filterDraftMembers } from "../models/workbookQuery";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import {
  decisionsViewSchemaId,
  hostsViewSchemaId,
  notesViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import { useWorkbookQueryController } from "./useWorkbookQueryController";
import { useWorkbookSurfaceQueries } from "./useWorkbookSurfaceQueries";

const impossibleDateDraft: FilterDraft = {
  fieldKey: "timeline.date_entered_sort_day",
  op: "eq",
  operandKind: "value",
  value: " 2026-04-31 ",
  valueType: "string",
  values: [],
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
          controller.commands.setQueryStateForSurface(
            "cartulary.view.timeline.v2",
            (current) => ({
              ...current,
              groupBy: "timeline.capture_state",
            }),
          );
        }}
        type="button"
      >
        Update Timeline
      </button>
      <button
        onClick={() => {
          controller.commands.setQueryStateForSurface(
            "cartulary.view.hosts.v1",
            (current) => ({
              ...current,
              groupBy: "host.entity_subtype",
            }),
          );
        }}
        type="button"
      >
        Update Hosts
      </button>
      <output aria-label={`query-controller-state-${instanceId}`}>
        {JSON.stringify({
          hosts: controller.snapshot.queryStateForSurface(
            "cartulary.view.hosts.v1",
          ).groupBy,
          timeline: controller.snapshot.queryStateForSurface(
            "cartulary.view.timeline.v2",
          ).groupBy,
        })}
      </output>
    </section>
  );
}

function FilterQueryHarness({
  draft = impossibleDateDraft,
}: {
  readonly draft?: FilterDraft;
}) {
  const controller = useWorkbookQueryController({
    surface: "cartulary.view.timeline.v2",
  });
  return (
    <section>
      <button
        type="button"
        onClick={() =>
          controller.snapshot.activeQueryControls.onFilterDraftChange(draft)
        }
      >
        Draft impossible date
      </button>
      <button
        type="button"
        onClick={() =>
          controller.snapshot.activeQueryControls.onApplyFilter(draft)
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
            values: filterDraftMembers(["seed"]),
          });
          controller.commands.setQueryStateForSurface(
            "cartulary.view.timeline.v2",
            (current) => ({
              ...current,
              groupBy: "timeline.capture_state",
              sort: [
                {
                  fieldKey: "timeline.date_entered_sort_day",
                  direction: "desc",
                },
              ],
            }),
          );
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
        {JSON.stringify(
          controller.snapshot.queryStateForSurface(
            "cartulary.view.timeline.v2",
          ),
        )}
      </output>
    </section>
  );
}

describe("useWorkbookQueryController", () => {
  it("keeps captured surface setters stable and applies consecutive updates to their original schema after navigation", async () => {
    const query = vi.fn<WorkbookViewQueryPort["query"]>(async (input) => ({
      kind: "accepted",
      value: {
        incidentId: "incident-one",
        rows: [],
        viewSchemaId: input.contract.viewSchemaId,
        ...acceptedQueryMetadata(input.contract.viewSchemaId, input.queryState),
      },
    }));
    const viewQuery = { query };
    const { result, rerender } = renderHook(
      ({ surface }: { readonly surface: string }) => {
        const controller = useWorkbookQueryController({ surface });
        const composition = useWorkbookSurfaceQueries({
          activeContract: workbookContractForViewSchemaId(surface),
          queryStateForSurface: controller.snapshot.queryStateForSurface,
          setQueryStateForSurface: controller.commands.setQueryStateForSurface,
          onAuthorityUncertain: undefined,
          sheetRef: { kind: "view_schema", id: surface },
          surface,
          viewQuery,
        });
        return { controller, composition };
      },
      { initialProps: { surface: notesViewSchemaId as string } },
    );
    const captured = result.current.composition.facadeQueries.generic.setState;
    const update = result.current.controller.commands.setQueryStateForSurface;
    await act(() => result.current.composition.refreshProjection.generic());
    const initialReads = query.mock.calls.length;
    rerender({ surface: notesViewSchemaId });
    act(() =>
      update(hostsViewSchemaId, (current) => ({
        ...current,
        groupBy: "host.entity_subtype",
      })),
    );
    expect(result.current.controller.commands.setQueryStateForSurface).toBe(
      update,
    );
    expect(result.current.composition.facadeQueries.generic.setState).toBe(
      captured,
    );
    expect(query).toHaveBeenCalledTimes(initialReads);
    rerender({ surface: decisionsViewSchemaId });
    act(() => {
      captured((current) => ({
        ...current,
        sort: [...current.sort, { fieldKey: "note.title", direction: "desc" }],
      }));
      captured((current) => ({
        ...current,
        sort: [
          ...current.sort,
          { fieldKey: "note.updated_at", direction: "asc" },
        ],
      }));
    });
    expect(
      result.current.controller.snapshot.queryStateForSurface(notesViewSchemaId)
        .sort,
    ).toEqual([
      { fieldKey: "note.title", direction: "desc" },
      { fieldKey: "note.updated_at", direction: "asc" },
    ]);
    expect(
      result.current.controller.snapshot.queryStateForSurface(
        decisionsViewSchemaId,
      ).sort,
    ).toEqual([]);
    rerender({ surface: notesViewSchemaId });
    rerender({ surface: decisionsViewSchemaId });
    act(() =>
      captured((current) => ({ ...current, sort: current.sort.slice(1) })),
    );
    expect(
      result.current.controller.snapshot.queryStateForSurface(notesViewSchemaId)
        .sort,
    ).toEqual([{ fieldKey: "note.updated_at", direction: "asc" }]);
    expect(
      result.current.controller.snapshot.queryStateForSurface(
        decisionsViewSchemaId,
      ).sort,
    ).toEqual([]);
  });

  it("refuses direct timestamp admission without changing requested state or raw draft", () => {
    const invalid: Extract<FilterDraft, { op: "eq" }> = {
      fieldKey: "note.updated_at",
      op: "eq",
      operandKind: "value",
      valueType: "string",
      value: " tomorrow ",
      values: [],
    };
    function Harness() {
      const { snapshot } = useWorkbookQueryController({
        surface: "cartulary.view.notes.v1",
      });
      const controls = snapshot.activeQueryControls;
      return (
        <>
          <button
            type="button"
            onClick={() =>
              controls.onApplyFilter({
                ...invalid,
                value: "2026-04-18T00:00:00Z",
              })
            }
          >
            Seed timestamp
          </button>
          <button
            type="button"
            onClick={() => controls.onFilterDraftChange(invalid)}
          >
            Draft timestamp
          </button>
          <button type="button" onClick={() => controls.onApplyFilter(invalid)}>
            Apply timestamp
          </button>
          <output aria-label="query">
            {JSON.stringify(
              snapshot.queryStateForSurface("cartulary.view.notes.v1"),
            )}
          </output>
          <output aria-label="draft">
            {JSON.stringify(controls.filterDraft)}
          </output>
        </>
      );
    }
    render(<Harness />);
    fireEvent.click(screen.getByText("Seed timestamp"));
    const accepted = screen.getByLabelText("query").textContent;
    fireEvent.click(screen.getByText("Draft timestamp"));
    expect(screen.getByLabelText("query").textContent).toBe(accepted);
    fireEvent.click(screen.getByText("Apply timestamp"));
    expect(screen.getByLabelText("query").textContent).toBe(accepted);
    expect(
      JSON.parse(screen.getByLabelText("draft").textContent ?? "{}"),
    ).toEqual(invalid);
  });

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

  it("refuses direct unset boolean admission and preserves the authored draft and query", () => {
    const draft: FilterDraft = {
      fieldKey: "timeline.has_evidence",
      op: "eq",
      operandKind: "values",
      valueType: "boolean",
      booleanOperand: { value: undefined, values: [] },
      value: "",
      values: [],
    };
    render(<FilterQueryHarness draft={draft} />);
    fireEvent.click(screen.getByRole("button", { name: "Seed query" }));
    fireEvent.click(
      screen.getByRole("button", { name: "Draft impossible date" }),
    );
    const before = screen.getByLabelText("requested-query").textContent;
    fireEvent.click(
      screen.getByRole("button", { name: "Apply impossible date" }),
    );
    expect(screen.getByLabelText("requested-query").textContent).toBe(before);
    expect(
      JSON.parse(screen.getByLabelText("raw-filter-draft").textContent ?? "{}"),
    ).toEqual(JSON.parse(JSON.stringify(draft)));
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
