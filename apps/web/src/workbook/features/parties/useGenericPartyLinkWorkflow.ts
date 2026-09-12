import type { ViewContract } from "@cartulary/view-contracts";
import { useLayoutEffect, useState, useSyncExternalStore } from "react";
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { type PartyReview, supportedPartyPairs } from "./partyLinkModel";
import type { WorkbookPartyLinkOperationOwner } from "./WorkbookPartyLinkOperationOwner";

/** Presentation binding only. Mutations and accepted results outlive this hook. */
export function useGenericPartyLinkWorkflow({
  owner,
  contract,
  row,
  sourceLabel,
  sheetRef,
  resetKey,
  fieldKey,
  visible,
}: {
  owner: WorkbookPartyLinkOperationOwner;
  contract: ViewContract;
  row: WorkbookQueryRow | null;
  sourceLabel: string;
  sheetRef: SheetRef;
  resetKey: string;
  fieldKey: string;
  visible: boolean;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const pairs = supportedPartyPairs(contract);
  const [selection, setSelection] = useState({ fieldKey, key: "" });
  const pair =
    pairs.find(
      (pair) =>
        pair.key === (selection.fieldKey === fieldKey ? selection.key : ""),
    ) ??
    pairs.find(
      (pair) => pair.textFieldKey === fieldKey || pair.refFieldKey === fieldKey,
    ) ??
    pairs[0] ??
    null;
  const scopeKey = JSON.stringify([
    resetKey,
    visible,
    snapshot.generation,
    sheetRef,
    row?.record_id,
    row?.row_version,
    pair?.key,
    fieldKey,
  ]);
  useLayoutEffect(() => {
    owner.setPresentation(visible && pair && row ? scopeKey : null);
    return () => owner.setPresentation(null);
  }, [owner, scopeKey, visible, pair, row]);
  const review: PartyReview | null =
    row && pair && snapshot.authority && visible
      ? {
          authority: snapshot.authority,
          pair,
          source: row,
          sheetRef,
          sourceLabel,
          presentation: scopeKey,
        }
      : null;
  const focusContext = JSON.stringify([
    visible,
    sheetRef,
    row?.record_id,
    pair?.key,
    fieldKey,
    snapshot.authority,
  ]);
  return {
    owner,
    snapshot,
    pair,
    pairs,
    review,
    scopeKey,
    focusContext,
    selectPair: (key: string) => {
      if (key !== pair?.key) owner.setPresentation(null);
      setSelection({ fieldKey, key });
    },
  };
}
