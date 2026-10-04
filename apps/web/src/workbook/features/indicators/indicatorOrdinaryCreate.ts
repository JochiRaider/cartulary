import { validateIndicatorCreateReceipt } from "../../adapters/createIndicatorCreateTransport";
import type { OrdinaryCreateContribution } from "../ordinary/ordinaryCreateContract";
import { prepareOrdinaryCreateFields } from "../ordinary/prepareOrdinaryCreateFields";
import { indicatorCreateErrors } from "./indicatorCreateModel";
/** The source owner evaluates raw values against the effective registry. */
export const indicatorOrdinaryCreate: OrdinaryCreateContribution = {
  views: ["cartulary.view.indicators.v1"],
  referenceViews: () => [],
  validateReceipt: (contract, body, receipt, status) =>
    !!validateIndicatorCreateReceipt({ contract, body }, receipt, status),
  prepare: (contract, values, id) =>
    prepareOrdinaryCreateFields(contract, values, id, {
      text: (field, raw) =>
        [
          "indicator.display_value",
          "indicator.normalized_value",
          "indicator.defanged_value",
        ].includes(field.fieldKey)
          ? { value: raw }
          : undefined,
      scalar: (_field, raw) => ({ value: raw }),
      minimum: (request) =>
        [
          "indicator.indicator_type",
          "indicator.value_kind",
          "indicator.display_value",
        ].every((key) => !!request[key]),
      validate: (request, errors) => {
        const text = Object.fromEntries(
          Object.entries(request)
            .filter(([key]) => key !== "client_txn_id")
            .map(([key, value]) => [key, value == null ? "" : String(value)]),
        );
        Object.assign(errors, indicatorCreateErrors(contract, text));
      },
    }),
};
