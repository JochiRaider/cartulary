import type { OrdinaryCreateContribution } from "../ordinary/ordinaryCreateContract";
import { prepareOrdinaryCreateFields } from "../ordinary/prepareOrdinaryCreateFields";
/** Metadata creation cannot claim finalized upload inputs or fabricate a locator. */
export const evidenceOrdinaryCreate: OrdinaryCreateContribution = {
  views: ["cartulary.view.evidence.v1"],
  referenceViews: (field) =>
    field.directReferenceContractId === "same_incident_party_ref_v1"
      ? ["cartulary.view.parties.v1"]
      : [],
  prepare: (contract, values, id) =>
    prepareOrdinaryCreateFields(contract, values, id, {
      minimum: (request) =>
        Object.entries(request).some(
          ([key, value]) =>
            key !== "client_txn_id" &&
            !key.endsWith("_party_id") &&
            value != null &&
            value !== "",
        ),
      validate: (request, errors) => {
        if (
          typeof request["evidence.storage_ref"] === "string" &&
          request["evidence.storage_ref"].startsWith("object://")
        )
          errors["evidence.storage_ref"] =
            "Use an external locator, or attach a file through Evidence actions.";
        if (
          ["available", "released"].includes(
            String(request["evidence.lifecycle_state"]),
          )
        )
          errors["evidence.lifecycle_state"] =
            "Use Evidence actions to complete this lifecycle transition.";
      },
    }),
};
