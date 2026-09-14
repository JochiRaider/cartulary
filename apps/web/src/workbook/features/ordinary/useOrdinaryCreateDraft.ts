import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useLayoutEffect,
  useMemo,
  useSyncExternalStore,
} from "react";
import type { WorkbookOrdinaryCreateOwner } from "./WorkbookOrdinaryCreateOwner";

const empty = Object.freeze({});
const noSubscribe = () => () => {};
const noSnapshot = () => null;
export function useOrdinaryCreateDraft(
  owner: WorkbookOrdinaryCreateOwner | undefined,
  contract: ViewContract,
): [
  Record<string, string>,
  Dispatch<SetStateAction<Record<string, string>>>,
  boolean,
] {
  const snapshot = useSyncExternalStore(
    owner?.subscribe ?? noSubscribe,
    owner?.getSnapshot ?? noSnapshot,
  );
  const attachment = useMemo(
    () => Symbol("ordinary authoring presentation"),
    [],
  );
  useLayoutEffect(
    () => owner?.attach(contract.viewSchemaId, attachment),
    [owner, contract.viewSchemaId, attachment],
  );
  return [
    snapshot?.schemas[contract.viewSchemaId]?.values ?? empty,
    (value) => {
      if (!owner) return;
      const current =
        owner.getSnapshot().schemas[contract.viewSchemaId]?.values ?? {};
      const next = typeof value === "function" ? value(current) : value;
      for (const key of new Set([
        ...Object.keys(current),
        ...Object.keys(next),
      ]))
        if (current[key] !== next[key])
          owner.update(contract.viewSchemaId, key, next[key]);
    },
    !owner?.canAuthor(),
  ];
}
