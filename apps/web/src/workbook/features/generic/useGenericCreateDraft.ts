import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useContext,
  useSyncExternalStore,
} from "react";
import {
  NoteCreateContext,
  noteSheetAttachment,
} from "../notes/NoteCreateContext";
import { noteCreateView } from "../notes/noteCreateModel";
import { useOrdinaryCreateDraft } from "../ordinary/useOrdinaryCreateDraft";
import type { WorkbookOrdinaryCreateOwner } from "../ordinary/WorkbookOrdinaryCreateOwner";

const noSubscribe = () => () => {};
const noSnapshot = () => null;
/** Keeps the common grid independent of per-artifact draft lifetimes. */
export function useGenericCreateDraft(
  contract: ViewContract,
  _actor: string | null,
  ordinary?: WorkbookOrdinaryCreateOwner,
): [
  Record<string, string>,
  Dispatch<SetStateAction<Record<string, string>>>,
  boolean,
] {
  const context = useContext(NoteCreateContext);
  const retained = useOrdinaryCreateDraft(ordinary, contract);
  const owner =
    contract.viewSchemaId === noteCreateView ? context?.owner : undefined;
  const state = useSyncExternalStore(
    owner?.subscribe ?? noSubscribe,
    owner?.getSnapshot ?? noSnapshot,
  );
  if (!owner || !context) return retained;
  return [
    { ...state?.draft?.values },
    (value) => {
      if (!owner.getSnapshot().draft)
        owner.beginSheet(context.sheetRef, noteSheetAttachment);
      const current = { ...owner.getSnapshot().draft?.values };
      const next = typeof value === "function" ? value(current) : value;
      for (const [key, text] of Object.entries(next))
        if (!Object.hasOwn(current, key) || text !== current[key])
          owner.update(key, text);
    },
    owner.busy || !owner.canSubmit(),
  ];
}
