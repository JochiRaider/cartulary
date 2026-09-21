import { useSyncExternalStore } from "react";
import type { WorkbookInspectorAttention } from "../../inspector/presentation/workbookInspectorPresentationModel";
import type { EvidenceWorkAttention } from "./evidenceWorkAttention";

type Entry = Readonly<{
  recordId: string | null;
  attention: EvidenceWorkAttention | null;
}>;
type Owner = {
  subscribe: (listener: () => void) => () => void;
  getSnapshot: () => readonly Entry[];
};
const empty: readonly Entry[] = [];
const noSnapshot = () => empty;
const noSubscription = () => () => {};

/** Subscription/admission only; the source owner decides which stage remains unresolved. */
export function useEvidenceInspectorAttention(
  owner: Owner | null,
  viewSchemaId: string,
  recordId: string | null,
): readonly WorkbookInspectorAttention[] {
  const entries = useSyncExternalStore(
    owner?.subscribe ?? noSubscription,
    owner?.getSnapshot ?? noSnapshot,
  );
  return entries.flatMap((entry, order) => {
    const attention = entry.attention;
    if (!attention || !recordId || entry.recordId !== recordId) return [];
    return [
      {
        ...attention,
        viewSchemaId,
        recordId,
        order,
        isCurrent: () => owner?.getSnapshot() === entries,
        destination: (section: HTMLElement) =>
          section.querySelector<HTMLElement>(
            `[data-evidence-work-id="${CSS.escape(attention.workId)}"]`,
          ) ?? section,
      },
    ];
  });
}
