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
    // Independent expectations qualify representative browser coverage of each
    // renderer. Registry-derived expectations would conceal a wrong assignment.
    expect(
      Object.fromEntries(
        registrations.map((entry) => [entry.viewSchemaId, entry.renderer]),
      ),
    ).toEqual({
      "cartulary.view.assessments.v1": "assessment",
      "cartulary.view.comm_log.v1": "contract",
      "cartulary.view.decisions.v1": "contract",
      "cartulary.view.evidence.v1": "contract",
      "cartulary.view.findings.v1": "contract",
      "cartulary.view.forensic_keywords.v1": "contract",
      "cartulary.view.handoff.v1": "contract",
      "cartulary.view.hosts.v1": "entity_hosts",
      "cartulary.view.identities.v1": "entity_identities",
      "cartulary.view.indicators.v1": "contract",
      "cartulary.view.investigative_queries.v1": "contract",
      "cartulary.view.lesson.v1": "contract",
      "cartulary.view.notes.v1": "contract",
      "cartulary.view.parties.v1": "contract",
      "cartulary.view.status_review.v1": "contract",
      "cartulary.view.task_requests.v1": "contract",
      "cartulary.view.timeline.v2": "timeline",
    });
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
