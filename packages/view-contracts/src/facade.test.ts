import { describe, expect, it } from "vitest";

import * as publicFacade from "./index";
import {
  getReferenceFieldContract,
  getViewContract,
  listReferenceFieldContracts,
  listViewContracts,
  listWorkbookSurfaceContracts,
  normalizeViewRowPatchV1,
  normalizeViewRowV1,
  requireViewContract,
  resolveHeaderSortFieldKey,
} from "./index";

describe("view-contracts façade", () => {
  it("exposes the supported public package facade", () => {
    for (const exportedFunction of [
      getReferenceFieldContract,
      getViewContract,
      listReferenceFieldContracts,
      listViewContracts,
      listWorkbookSurfaceContracts,
      normalizeViewRowPatchV1,
      normalizeViewRowV1,
      requireViewContract,
      resolveHeaderSortFieldKey,
    ]) {
      expect(typeof exportedFunction).toBe("function");
    }
    expect(Object.keys(publicFacade).sort()).toEqual(
      [
        "assessmentsViewSchemaId",
        "commLogViewSchemaId",
        "decisionsViewSchemaId",
        "evidenceViewSchemaId",
        "findingsViewSchemaId",
        "forensicKeywordsViewSchemaId",
        "getReferenceFieldContract",
        "getViewContract",
        "handoffViewSchemaId",
        "hostsViewSchemaId",
        "identitiesViewSchemaId",
        "indicatorsViewSchemaId",
        "investigativeQueriesViewSchemaId",
        "lessonViewSchemaId",
        "listReferenceFieldContracts",
        "listViewContracts",
        "listWorkbookSurfaceContracts",
        "normalizeViewRowPatchV1",
        "normalizeViewRowV1",
        "notesViewSchemaId",
        "optionalStandardizedWorkbookSurfaceIds",
        "partiesViewSchemaId",
        "requiredBuiltInWorkbookSurfaceIds",
        "requiredSystemWorkbookSurfaceIds",
        "requireViewContract",
        "resolveHeaderSortFieldKey",
        "statusReviewViewSchemaId",
        "taskRequestsViewSchemaId",
        "timelineViewSchemaId",
      ].sort(),
    );
  });
});
