import type { ViewContract } from "@cartulary/view-contracts";
import { useSyncExternalStore } from "react";
import type { WorkbookGridDraftStore } from "../models/WorkbookGridDraftStore";

export type ParkedGridDraft = {
  readonly key: string;
  readonly recordId: string;
  readonly label: string;
  readonly value: string | null;
  readonly reason: string;
  readonly discard: () => void;
};

/** Readable local work whose original cell is currently unavailable for editing. */
export function WorkbookParkedGridDrafts({
  drafts,
}: {
  readonly drafts: readonly ParkedGridDraft[];
}) {
  if (!drafts.length) return null;
  return (
    <details data-grid-editor-external-action="true">
      <summary>Unsaved cells ({drafts.length})</summary>
      <div style={{ maxBlockSize: "12rem", overflow: "auto" }}>
        {drafts.map((draft) => (
          <section key={draft.key} aria-label={`Unsaved ${draft.label}`}>
            <span>
              {draft.label} · {draft.recordId} · {draft.reason}
            </span>
            <textarea
              aria-label={`Retained ${draft.label}`}
              readOnly
              value={draft.value ?? ""}
              rows={2}
              style={{
                boxSizing: "border-box",
                display: "block",
                width: "100%",
              }}
            />
            {draft.value === null ? <span>Explicit clear</span> : null}
            <button type="button" onClick={draft.discard}>
              Discard {draft.label} draft
            </button>
          </section>
        ))}
      </div>
    </details>
  );
}

export function WorkbookUnavailableGridDrafts({
  store,
  contract,
  recordIds,
  fieldKeys,
}: {
  readonly store: WorkbookGridDraftStore;
  readonly contract: ViewContract;
  readonly recordIds: readonly string[];
  readonly fieldKeys: readonly string[];
}) {
  useSyncExternalStore(store.subscribe, store.getSnapshot);
  const drafts = store.list(contract.viewSchemaId).flatMap((draft) => {
    const { identity } = draft;
    const field = contract.fieldMap[identity.fieldKey];
    const reason = !store.canAuthor()
      ? "Editing is unavailable."
      : !recordIds.includes(identity.recordId)
        ? "Original row is outside this result."
        : !fieldKeys.includes(identity.fieldKey) || !field?.gridEditable
          ? "Original field is unavailable."
          : null;
    return reason
      ? [
          {
            key: `${identity.recordId}:${identity.fieldKey}`,
            recordId: identity.recordId,
            label: field?.label ?? identity.fieldKey,
            value: draft.value,
            reason,
            discard: () => store.discard(identity),
          },
        ]
      : [];
  });
  return <WorkbookParkedGridDrafts drafts={drafts} />;
}
