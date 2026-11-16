import { describe, expect, it } from "vitest";

import {
  assessmentsViewSchemaId,
  commLogViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  handoffViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  indicatorsViewSchemaId,
  investigativeQueriesViewSchemaId,
  lessonViewSchemaId,
  listWorkbookSurfaceContracts,
  notesViewSchemaId,
  optionalStandardizedWorkbookSurfaceIds,
  partiesViewSchemaId,
  requiredBuiltInWorkbookSurfaceIds,
  requiredSystemWorkbookSurfaceIds,
  requireViewContract,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./index";
import { listProjectedWorkbookSurfaceContracts } from "./projection";

const schemaConstantPairs = [
  [timelineViewSchemaId, "cartulary.view.timeline.v2"],
  [hostsViewSchemaId, "cartulary.view.hosts.v1"],
  [identitiesViewSchemaId, "cartulary.view.identities.v1"],
  [evidenceViewSchemaId, "cartulary.view.evidence.v1"],
  [notesViewSchemaId, "cartulary.view.notes.v1"],
  [indicatorsViewSchemaId, "cartulary.view.indicators.v1"],
  [assessmentsViewSchemaId, "cartulary.view.assessments.v1"],
  [taskRequestsViewSchemaId, "cartulary.view.task_requests.v1"],
  [decisionsViewSchemaId, "cartulary.view.decisions.v1"],
  [partiesViewSchemaId, "cartulary.view.parties.v1"],
  [commLogViewSchemaId, "cartulary.view.comm_log.v1"],
  [handoffViewSchemaId, "cartulary.view.handoff.v1"],
  [statusReviewViewSchemaId, "cartulary.view.status_review.v1"],
  [lessonViewSchemaId, "cartulary.view.lesson.v1"],
  [findingsViewSchemaId, "cartulary.view.findings.v1"],
  [investigativeQueriesViewSchemaId, "cartulary.view.investigative_queries.v1"],
  [forensicKeywordsViewSchemaId, "cartulary.view.forensic_keywords.v1"],
] as const;
const schemaConstants = schemaConstantPairs.map(([, literal]) => literal);

describe("workbook surface contracts", () => {
  it("derives every schema constant and status partition in registry order", () => {
    for (const [constant, literal] of schemaConstantPairs)
      expect(constant).toBe(literal);
    expect(
      listWorkbookSurfaceContracts().map((entry) => entry.viewSchemaId),
    ).toEqual(schemaConstants);
    expect(requiredBuiltInWorkbookSurfaceIds).toEqual(
      schemaConstants.slice(0, 5),
    );
    expect(requiredSystemWorkbookSurfaceIds).toEqual(
      schemaConstants.slice(5, 14),
    );
    expect(optionalStandardizedWorkbookSurfaceIds).toEqual(
      schemaConstants.slice(14),
    );
  });

  it("matches generated surface projections by shared identity", () => {
    expect(listWorkbookSurfaceContracts()).toBe(
      listProjectedWorkbookSurfaceContracts(),
    );
    expect(Object.isFrozen(listWorkbookSurfaceContracts())).toBe(true);
  });

  it("joins workbook identity and status metadata in Core 01 Table 7.4-A order", () => {
    const entries = listWorkbookSurfaceContracts();
    expect(entries).toHaveLength(17);
    expect(entries[0]?.viewSchemaId).toBe(timelineViewSchemaId);
    expect(entries[5]?.viewSchemaId).toBe(indicatorsViewSchemaId);
    for (const entry of entries) {
      expect(entry.contract).toBe(requireViewContract(entry.viewSchemaId));
      expect(entry.contract.viewSchemaId).toBe(entry.viewSchemaId);
    }
  });

  it("exposes required reference-pack metadata from view contracts", () => {
    for (const surface of listWorkbookSurfaceContracts()) {
      expect(surface.requiredReferencePackKeys).toEqual(
        surface.contract.requiredReferencePackKeys,
      );
    }
  });
});
