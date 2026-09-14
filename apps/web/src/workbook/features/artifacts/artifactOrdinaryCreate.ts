import { listWorkbookSurfaceContracts } from "@cartulary/view-contracts";
import type { OrdinaryCreateContribution } from "../ordinary/ordinaryCreateContract";
import { prepareOrdinaryCreateFields } from "../ordinary/prepareOrdinaryCreateFields";
export const artifactOrdinaryCreate: OrdinaryCreateContribution = {
  views: ["findings", "investigative_queries", "forensic_keywords"].map(
    (name) => `cartulary.view.${name}.v1`,
  ),
  prepare: (contract, values, id) =>
    prepareOrdinaryCreateFields(contract, values, id),
  referenceViews: (field) => {
    if (field.directReferenceContractId === "incident_member_user_ref_v1")
      return ["incident_members"];
    if (
      ["finding.supporting_refs", "finding.contradictory_refs"].includes(
        field.fieldKey,
      )
    )
      return listWorkbookSurfaceContracts().map((view) => view.viewSchemaId);
    return [];
  },
};
