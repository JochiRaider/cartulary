import {
  getViewContract,
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";

const timelineContract = requireViewContract(timelineViewSchemaId);

const createRelatedTargetContracts = new Map<string, ViewContract>();
for (const featureGroup of timelineContract.inspectorConfig.featureGroups) {
  if (
    featureGroup.routeBinding.kind !== "view_row_create" ||
    featureGroup.routeBinding.owner !== "view_row_create_route" ||
    featureGroup.routeBinding.targetViewSchemaId === undefined
  ) {
    continue;
  }
  const targetContract = getViewContract(
    featureGroup.routeBinding.targetViewSchemaId,
  );
  if (targetContract !== undefined) {
    createRelatedTargetContracts.set(
      targetContract.viewSchemaId,
      targetContract,
    );
  }
}

export const timelineCreateRelatedTargetContracts: ReadonlyMap<
  string,
  ViewContract
> = createRelatedTargetContracts;
