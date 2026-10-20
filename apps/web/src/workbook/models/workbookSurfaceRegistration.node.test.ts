import { listViewContracts } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildWorkbookSurfaceRegistrations,
  listWorkbookSurfaceRegistrations,
} from "./workbookSurfaceRegistration";

describe("WorkbookSurfaceRegistration", () => {
  it("registers every exposed schema exactly once and excludes extension workspaces", () => {
    const registrations = listWorkbookSurfaceRegistrations();
    expect(registrations.map((entry) => entry.viewSchemaId).sort()).toEqual(
      listViewContracts()
        .map((contract) => contract.viewSchemaId)
        .sort(),
    );
    expect(
      registrations.some(
        (entry) => entry.viewSchemaId === "network_flow_activity",
      ),
    ).toBe(false);
  });

  it("fails registry construction for duplicate and missing owners", () => {
    const [first, ...rest] = listWorkbookSurfaceRegistrations();
    expect(first).toBeDefined();
    if (!first) {
      return;
    }
    const definition = {
      viewSchemaId: first.viewSchemaId,
      ownerId: first.ownerId,
      renderer: first.renderer,
      policy: first.policy,
    };
    expect(() =>
      buildWorkbookSurfaceRegistrations(listViewContracts(), [
        definition,
        definition,
        ...rest.map(({ contract: _contract, ...entry }) => entry),
      ]),
    ).toThrow(/Duplicate workbook surface registration/u);
    expect(() =>
      buildWorkbookSurfaceRegistrations(
        listViewContracts(),
        rest.map(({ contract: _contract, ...entry }) => entry),
      ),
    ).toThrow(/Missing workbook surface registration/u);
  });
});
