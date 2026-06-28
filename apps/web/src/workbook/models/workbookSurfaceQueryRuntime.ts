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
const fallbackWorkbookContract =
  allWorkbookContracts.find(
    (contract) => contract.viewSchemaId === timelineViewSchemaId,
  ) ?? allWorkbookContracts[0];

export function workbookContractForViewSchemaId(
  viewSchemaId: string,
): ViewContract {
  return (
    allWorkbookContracts.find(
      (contract) => contract.viewSchemaId === viewSchemaId,
    ) ??
    fallbackWorkbookContract ??
    (() => {
      throw new Error("Workbook surface registry has no contracts.");
    })()
  );
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
