import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useContext,
  useEffect,
  useState,
  useSyncExternalStore,
} from "react";
import { initialGenericCreateDraft } from "../../models/genericWorkbookModel";
import {
  NoteCreateContext,
  noteSheetAttachment,
} from "../notes/NoteCreateContext";
import { noteCreateView } from "../notes/noteCreateModel";

const noSubscribe = () => () => {};
const noSnapshot = () => null;
/** Keeps the common grid independent of per-artifact draft lifetimes. */
export function useGenericCreateDraft(
  contract: ViewContract,
  actor: string | null,
): [
  Record<string, string>,
  Dispatch<SetStateAction<Record<string, string>>>,
  boolean,
] {
  const context = useContext(NoteCreateContext);
  const [local, setLocal] = useState(() =>
    initialGenericCreateDraft(contract, actor),
  );
  const owner =
    contract.viewSchemaId === noteCreateView ? context?.owner : undefined;
  useEffect(() => {
    if (!owner)
      setLocal((current) => ({
        ...initialGenericCreateDraft(contract, actor),
        ...current,
      }));
  }, [owner, contract, actor]);
  const state = useSyncExternalStore(
    owner?.subscribe ?? noSubscribe,
    owner?.getSnapshot ?? noSnapshot,
  );
  if (!owner || !context) return [local, setLocal, false];
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
