import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useWorkbookQueryController } from "./useWorkbookQueryController";

function QueryControllerHarness() {
  const controller = useWorkbookQueryController({
    startupSheetRef: {
      kind: "view_schema",
      id: "cartulary.view.timeline.v2",
    },
    surface: "cartulary.view.timeline.v2",
  });
  return (
    <>
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
      <output aria-label="query-controller-state">
        {JSON.stringify({
          hosts: controller.snapshot.hostQueryState.groupBy,
          timeline: controller.snapshot.timelineQueryState.groupBy,
        })}
      </output>
    </>
  );
}

describe("useWorkbookQueryController", () => {
  it("keeps query state isolated by exact view_schema_id", () => {
    render(<QueryControllerHarness />);
    fireEvent.click(screen.getByRole("button", { name: "Update Timeline" }));
    fireEvent.click(screen.getByRole("button", { name: "Update Hosts" }));

    expect(
      JSON.parse(
        screen.getByLabelText("query-controller-state").textContent ?? "{}",
      ),
    ).toEqual({
      hosts: "host.entity_subtype",
      timeline: "timeline.capture_state",
    });
  });
});
