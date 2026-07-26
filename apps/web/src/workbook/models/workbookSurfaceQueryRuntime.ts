import type { ViewContract } from "@cartulary/view-contracts";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  listWorkbookSurfaceRegistryEntries,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";

export type WorkbookQuerySurfaceSlot =
  | "timeline"
  | "hosts"
  | "identities"
  | "assessments"
  | "generic";

const allWorkbookContracts = listWorkbookSurfaceRegistryEntries().map(
  (entry) => entry.contract,
);

export function workbookContractForViewSchemaId(
  viewSchemaId: string,
): ViewContract {
  const contract = allWorkbookContracts.find(
    (candidate) => candidate.viewSchemaId === viewSchemaId,
  );
  if (contract === undefined) {
    throw new Error(
      `Unknown workbook view schema: ${viewSchemaId || "<empty>"}`,
    );
  }
  return contract;
}

export function workbookQuerySurfaceSlot(
  viewSchemaId: string,
): WorkbookQuerySurfaceSlot {
  switch (viewSchemaId) {
    case timelineViewSchemaId:
      return "timeline";
    case hostsViewSchemaId:
      return "hosts";
    case identitiesViewSchemaId:
      return "identities";
    case assessmentsViewSchemaId:
      return "assessments";
    default:
      return "generic";
  }
}
