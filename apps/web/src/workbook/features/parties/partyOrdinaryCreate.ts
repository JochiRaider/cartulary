import type { OrdinaryCreateContribution } from "../ordinary/ordinaryCreateContract";
import { prepareOrdinaryCreateFields } from "../ordinary/prepareOrdinaryCreateFields";
/** Exact email/external-reference reuse and cross-key conflicts remain server owned. */
export const partyOrdinaryCreate: OrdinaryCreateContribution = {
  views: ["cartulary.view.parties.v1"],
  referenceViews: () => [],
  prepare: (contract, values, id) =>
    prepareOrdinaryCreateFields(contract, values, id, {
      validate: (request, errors) => {
        const timezone = request["party.timezone_name"];
        if (typeof timezone === "string" && !isWorkbookTimezoneName(timezone))
          errors["party.timezone_name"] =
            "Enter an available IANA timezone name.";
      },
    }),
};

import { isWorkbookTimezoneName } from "../../adapters/workbookStringContracts";
