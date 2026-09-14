import type { EntityRecordWriteBoundary } from "../../mutations/entityRecordWriteBoundary";
import type { OrdinaryCreateContribution } from "../ordinary/ordinaryCreateContract";
import { prepareOrdinaryCreateFields } from "../ordinary/prepareOrdinaryCreateFields";
export const entityOrdinaryCreate: OrdinaryCreateContribution = {
  views: ["cartulary.view.hosts.v1", "cartulary.view.identities.v1"],
  referenceViews: () => [],
  prepare: (contract, values, id) =>
    prepareOrdinaryCreateFields(contract, values, id, {
      scalar: (_field, raw) => {
        const value = raw
          .normalize("NFC")
          .replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
        return /[\p{Cc}\p{Cf}\p{Cs}]/u.test(value)
          ? { value, error: "Remove unsupported control characters." }
          : { value };
      },
    }),
};

export function createEntityOrdinaryCreate(
  writes: EntityRecordWriteBoundary,
): OrdinaryCreateContribution {
  return {
    ...entityOrdinaryCreate,
    accepted: (row) => writes.acceptVersion(row.record_id, row.row_version),
    reserve: (contract) =>
      writes.begin({
        recordIds: [],
        unknownEntityType:
          contract.viewSchemaId === "cartulary.view.hosts.v1"
            ? "host"
            : "identity",
      }),
  };
}
