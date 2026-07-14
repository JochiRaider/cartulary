import { fireEvent, render, screen } from "@testing-library/react";
import { useRef } from "react";
import { describe, expect, it } from "vitest";
import { useWorkbookStartupController } from "./useWorkbookStartupController";

function StartupControllerHarness() {
  const selectionVersionRef = useRef(0);
  const controller = useWorkbookStartupController({
    incidentId: "incident-1",
    surfaceSelectionVersionRef: selectionVersionRef,
  });
  return (
    <>
      <button
        onClick={() =>
          controller.commands.selectWorkbookSurface("cartulary.view.hosts.v1", {
            focusFirstGridTarget: true,
          })
        }
        type="button"
      >
        Select Hosts
      </button>
      <output aria-label="startup-controller-state">
        {JSON.stringify({
          focus: controller.snapshot.pendingGridFocusSurface,
          sheetRef: controller.snapshot.startupSheetRef,
          surface: controller.snapshot.surface,
          version: selectionVersionRef.current,
        })}
      </output>
    </>
  );
}

describe("useWorkbookStartupController", () => {
  it("owns selection identity, focus intent, versioning, and URL history", () => {
    window.history.replaceState(
      {},
      "",
      "/?incident_id=incident-1&view_schema_id=cartulary.view.timeline.v2",
    );
    render(<StartupControllerHarness />);
    fireEvent.click(screen.getByRole("button", { name: "Select Hosts" }));

    expect(
      JSON.parse(
        screen.getByLabelText("startup-controller-state").textContent ?? "{}",
      ),
    ).toEqual({
      focus: "cartulary.view.hosts.v1",
      sheetRef: { kind: "view_schema", id: "cartulary.view.hosts.v1" },
      surface: "cartulary.view.hosts.v1",
      version: 1,
    });
    expect(window.location.search).toContain(
      "view_schema_id=cartulary.view.hosts.v1",
    );
  });
});
