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

const schemaConstants = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
  indicatorsViewSchemaId,
  assessmentsViewSchemaId,
  taskRequestsViewSchemaId,
  decisionsViewSchemaId,
  partiesViewSchemaId,
  commLogViewSchemaId,
  handoffViewSchemaId,
  statusReviewViewSchemaId,
  lessonViewSchemaId,
  findingsViewSchemaId,
  investigativeQueriesViewSchemaId,
  forensicKeywordsViewSchemaId,
];

describe("workbook surface contracts", () => {
  it("derives every schema constant and status partition in registry order", () => {
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
      expect(entry.requiredReferencePackKeys).toEqual(
        entry.contract.requiredReferencePackKeys,
      );
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
