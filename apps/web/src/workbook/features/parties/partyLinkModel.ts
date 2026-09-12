import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

export const partyViewId = "cartulary.view.parties.v1";
export const partyPairs = [
  {
    key: "evidence.collector_party_text:evidence.collector_party_id",
    label: "Collector",
    viewSchemaId: "cartulary.view.evidence.v1",
    recordType: "evidence",
    textFieldKey: "evidence.collector_party_text",
    refFieldKey: "evidence.collector_party_id",
    linkBinding: "party.collector.link",
    clearBinding: "party.reference.clear",
  },
  {
    key: "evidence.source_party_text:evidence.source_party_id",
    label: "Source",
    viewSchemaId: "cartulary.view.evidence.v1",
    recordType: "evidence",
    textFieldKey: "evidence.source_party_text",
    refFieldKey: "evidence.source_party_id",
    linkBinding: "party.source.link",
    clearBinding: "party.reference.clear",
  },
  {
    key: "task.requester_party_text:task.requester_party_id",
    label: "Requester",
    viewSchemaId: "cartulary.view.task_requests.v1",
    recordType: "task_request",
    textFieldKey: "task.requester_party_text",
    refFieldKey: "task.requester_party_id",
    linkBinding: "task.requester_party.link",
    clearBinding: "task.requester_party.clear",
  },
] as const;
export type PartyPair = (typeof partyPairs)[number];
export type PartyAction = "link" | "clear_link" | "clear_text" | "clear_both";
export type PartyReview = Readonly<{
  authority: WorkbookMutationAuthority;
  pair: PartyPair;
  source: WorkbookQueryRow;
  sheetRef: SheetRef;
  sourceLabel: string;
  presentation: string;
}>;
export function supportedPartyPairs(
  contract: ViewContract,
): readonly PartyPair[] {
  return partyPairs.filter(
    (pair) =>
      pair.viewSchemaId === contract.viewSchemaId &&
      contract.fieldMap[pair.textFieldKey]?.writeKind === "direct_value" &&
      contract.fieldMap[pair.refFieldKey]?.directReferenceContractId ===
        "same_incident_party_ref_v1" &&
      [pair.linkBinding, pair.clearBinding].every((key) =>
        contract.inspectorConfig.featureGroups.some(
          (group) =>
            group.featureGroupKey === key &&
            group.routeBinding.kind === "record_patch" &&
            group.routeBinding.owner === "record_patch_route",
        ),
      ),
  );
}
export function partyValues(row: WorkbookQueryRow, pair: PartyPair) {
  return {
    text: String(row.cells[pair.textFieldKey]?.value ?? ""),
    reference: String(row.cells[pair.refFieldKey]?.value ?? ""),
  };
}
export function partyState(row: WorkbookQueryRow, pair: PartyPair) {
  const { text, reference } = partyValues(row, pair);
  return text
    ? reference
      ? "Linked with source wording"
      : "Source wording only"
    : reference
      ? "Party link only"
      : "No source wording or Party link";
}
export function partyChanges(
  pair: PartyPair,
  action: PartyAction,
  target = "",
) {
  switch (action) {
    case "link":
      return [{ field_key: pair.refFieldKey, value: target }];
    case "clear_link":
      return [{ field_key: pair.refFieldKey, value: null }];
    case "clear_text":
      return [{ field_key: pair.textFieldKey, value: null }];
    case "clear_both":
      return [
        { field_key: pair.textFieldKey, value: null },
        { field_key: pair.refFieldKey, value: null },
      ];
  }
}
export function partyCreateDraft(text: string): Record<string, string> {
  return Object.fromEntries(
    requireViewContract(partyViewId)
      .fields.filter((field) => field.createWritable)
      .map((field) => [
        field.fieldKey,
        field.fieldKey === "party.display_name" ? text : "",
      ]),
  );
}

export type PartyPage = Readonly<{
  rows: readonly WorkbookQueryRow[];
  nextCursor: string | null;
  hasMore: boolean;
}>;
export interface PartyLinkReadPort {
  page(cursor: string | null, signal: AbortSignal): Promise<PartyPage>;
  source(
    viewSchemaId: string,
    recordId: string,
    signal: AbortSignal,
  ): Promise<WorkbookQueryRow>;
}
