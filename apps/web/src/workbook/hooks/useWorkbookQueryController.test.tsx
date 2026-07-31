import {
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import { afterEach, describe, expect, it } from "vitest";
import { useWorkbookQueryController } from "./useWorkbookQueryController";

afterEach(cleanup);

function QueryControllerHarness({
  instanceId,
}: {
  readonly instanceId: string;
}) {
  const controller = useWorkbookQueryController({
    startupSheetRef: {
      kind: "view_schema",
      id: "cartulary.view.timeline.v2",
    },
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

describe("useWorkbookQueryController", () => {
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
