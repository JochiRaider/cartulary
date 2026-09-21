import { requireViewContract } from "@cartulary/view-contracts";
import { useSyncExternalStore } from "react";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookExplicitPatchOwner } from "../runtime/WorkbookExplicitPatchOwner";
import type { WorkbookInspectorAttention } from "./presentation/workbookInspectorPresentationModel";
import {
  type InspectorEditIdentity,
  inspectorEditKey,
  type WorkbookInspectorDraftStore,
} from "./WorkbookInspectorDraftStore";

export function useWorkbookInspectorOrdinaryAttention(
  drafts: WorkbookInspectorDraftStore,
  patches: WorkbookExplicitPatchOwner,
  viewSchemaId: string,
  row: WorkbookQueryRow | null,
  reviewRequiredFor?: (identity: InspectorEditIdentity) => boolean,
): readonly WorkbookInspectorAttention[] {
  useSyncExternalStore(drafts.subscribe, drafts.getSnapshot);
  useSyncExternalStore(patches.subscribe, patches.getSnapshot);
  return workbookInspectorOrdinaryAttention(
    drafts,
    patches,
    viewSchemaId,
    row,
    reviewRequiredFor,
  );
}

/** Immutable projection of existing owners; this adapter holds no work or requests. */
export function workbookInspectorOrdinaryAttention(
  drafts: WorkbookInspectorDraftStore,
  patches: WorkbookExplicitPatchOwner,
  viewSchemaId: string,
  row: WorkbookQueryRow | null,
  reviewRequiredFor?: (identity: InspectorEditIdentity) => boolean,
): readonly WorkbookInspectorAttention[] {
  if (!row) return [];
  const draftSnapshot = drafts.getSnapshot();
  const patchSnapshot = patches.getSnapshot();
  const contract = requireViewContract(viewSchemaId);
  const current = () =>
    drafts.getSnapshot() === draftSnapshot &&
    patches.getSnapshot() === patchSnapshot;
  const destination =
    (fieldKey: string | undefined) => (section: HTMLElement) =>
      [
        ...section.querySelectorAll<HTMLElement>(
          "[data-inspector-saved-field]",
        ),
      ].find((element) => element.dataset.inspectorSavedField === fieldKey) ??
      null;
  const operations = patchSnapshot.entries.filter(
    (entry) =>
      entry.intent.viewSchemaId === viewSchemaId &&
      entry.intent.baseline.record_id === row.record_id &&
      !(entry.receipt && entry.reconciliation === "complete") &&
      entry.intent.owner !== "party_link" &&
      entry.intent.purpose !== "task-lifecycle",
  );
  const operationEntries: WorkbookInspectorAttention[] = operations.map(
    (entry, order) => {
      const category = entry.receipt
        ? "refresh"
        : entry.phase === "uncertain"
          ? "uncertain"
          : entry.phase === "conflict"
            ? "review"
            : entry.phase === "submitting" || entry.phase === "coordinating"
              ? "in_progress"
              : "failure";
      const labels = {
        refresh: "Saved change needs refresh",
        uncertain: "Unconfirmed change",
        review: "Change needs review",
        in_progress: "Saving change",
        failure: "Change was not accepted",
      };
      return {
        workId: `explicit-patch:${entry.id}`,
        viewSchemaId,
        recordId: row.record_id,
        category,
        label: `${labels[category]}: ${entry.intent.changes.map((change) => contract.fieldMap[change.field_key]?.label ?? "original field").join(", ")}`,
        order,
        outcomeIdentity: entry.id,
        isCurrent: () => current() && patchSnapshot.authority !== null,
        destination: destination(entry.intent.changes[0]?.field_key),
      };
    },
  );
  const draftEntries: WorkbookInspectorAttention[] = drafts
    .readRecord(viewSchemaId, row.record_id)
    // A captured authoring revision and its operation are one obligation. Newer
    // authoring has another owner-issued revision even when its text is equal.
    .filter(
      (draft) =>
        !operations.some(
          (operation) => operation.intent.authoringRevision === draft.revision,
        ),
    )
    .map((draft, order) => {
      const reviewRequired =
        reviewRequiredFor?.(draft.identity) ??
        drafts.staleFields(draft.identity, row, [draft.identity.fieldKey])
          .length > 0;
      return {
        workId: `ordinary-draft:${inspectorEditKey(draft.identity)}`,
        viewSchemaId,
        recordId: row.record_id,
        category: reviewRequired ? "review" : "draft",
        label: `${reviewRequired ? "Draft needs review" : "Unsaved draft"}: ${contract.fieldMap[draft.identity.fieldKey]?.label ?? "original field"}`,
        order: operations.length + order,
        isCurrent: () =>
          current() && drafts.read(draft.identity)?.revision === draft.revision,
        destination: destination(draft.identity.fieldKey),
      };
    });
  return [...operationEntries, ...draftEntries];
}
