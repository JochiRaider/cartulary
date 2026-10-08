import type { ViewContract } from "@cartulary/view-contracts";
import { listWorkbookSurfaceRegistryEntries } from "./workbookSurfaceRegistry";

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
