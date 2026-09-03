import { fireEvent, render, screen } from "@testing-library/react";
import { useRef } from "react";
import { describe, expect, it } from "vitest";
import type { WorkbookPreferencePort } from "../ports/WorkbookPreferencePort";
import { useWorkbookStartupController } from "./useWorkbookStartupController";

const preferencePort: WorkbookPreferencePort = {
  setDefaultSheet: async () => ({ kind: "accepted", value: undefined }),
  setHomeSheet: async () => ({ kind: "accepted", value: undefined }),
};

function StartupControllerHarness() {
  const selectionVersionRef = useRef(0);
  const controller = useWorkbookStartupController({
    incidentId: "incident-1",
    preferencePort,
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
      <button
        onClick={() =>
          controller.commands.selectWorkbookSurface(
            "cartulary.view.timeline.v2",
          )
        }
        type="button"
      >
        Select Timeline Without Focus
      </button>
      <button
        onClick={() =>
          controller.commands.selectExtensionWorkspace({
            extension_profile_id: "cartulary.extension.network_flow.v1",
            kind: "extension_workspace",
            workspace_key: "network-analysis",
          })
        }
        type="button"
      >
        Select Extension
      </button>
      <button
        onClick={() => {
          const request = controller.snapshot.gridEntryFocusRequest;
          if (request.kind === "pending") {
            controller.commands.acknowledgeGridEntryFocus({
              generation: request.generation,
              viewSchemaId: request.viewSchemaId,
            });
          }
        }}
        type="button"
      >
        Acknowledge Current
      </button>
      <button
        onClick={() =>
          controller.commands.acknowledgeGridEntryFocus({
            generation: 1,
            viewSchemaId: "cartulary.view.hosts.v1",
          })
        }
        type="button"
      >
        Acknowledge Generation One
      </button>
      <output aria-label="startup-controller-state">
        {JSON.stringify({
          focus: controller.snapshot.gridEntryFocusRequest,
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
      focus: {
        generation: 1,
        kind: "pending",
        viewSchemaId: "cartulary.view.hosts.v1",
      },
      sheetRef: { kind: "view_schema", id: "cartulary.view.hosts.v1" },
      surface: "cartulary.view.hosts.v1",
      version: 1,
    });
    expect(window.location.search).toContain(
      "view_schema_id=cartulary.view.hosts.v1",
    );
  });

  it("uses monotonic generations and exact acknowledgement for repeated selections", () => {
    render(<StartupControllerHarness />);
    const selectHosts = screen.getByRole("button", { name: "Select Hosts" });

    fireEvent.click(selectHosts);
    fireEvent.click(selectHosts);
    expect(
      JSON.parse(
        screen.getByLabelText("startup-controller-state").textContent ?? "{}",
      ).focus,
    ).toEqual({
      generation: 2,
      kind: "pending",
      viewSchemaId: "cartulary.view.hosts.v1",
    });

    fireEvent.click(
      screen.getByRole("button", { name: "Acknowledge Generation One" }),
    );
    expect(
      JSON.parse(
        screen.getByLabelText("startup-controller-state").textContent ?? "{}",
      ).focus,
    ).toEqual({
      generation: 2,
      kind: "pending",
      viewSchemaId: "cartulary.view.hosts.v1",
    });

    fireEvent.click(
      screen.getByRole("button", { name: "Acknowledge Current" }),
    );
    expect(
      JSON.parse(
        screen.getByLabelText("startup-controller-state").textContent ?? "{}",
      ).focus,
    ).toEqual({ kind: "idle" });
  });

  it("cancels an older request before every base or extension selection", () => {
    render(<StartupControllerHarness />);
    const selectHosts = screen.getByRole("button", { name: "Select Hosts" });

    fireEvent.click(selectHosts);
    fireEvent.click(
      screen.getByRole("button", { name: "Select Timeline Without Focus" }),
    );
    expect(
      JSON.parse(
        screen.getByLabelText("startup-controller-state").textContent ?? "{}",
      ).focus,
    ).toEqual({ kind: "idle" });

    fireEvent.click(selectHosts);
    fireEvent.click(screen.getByRole("button", { name: "Select Extension" }));
    expect(
      JSON.parse(
        screen.getByLabelText("startup-controller-state").textContent ?? "{}",
      ).focus,
    ).toEqual({ kind: "idle" });
  });
});
