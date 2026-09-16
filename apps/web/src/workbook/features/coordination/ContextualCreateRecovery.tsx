import { useRef, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { ContextualCreateForm } from "./ContextualCreateForm";
import type { WorkbookContextualTaskDecisionCreateOwner } from "./WorkbookContextualTaskDecisionCreateOwner";
export function ContextualCreateRecovery({
  owner,
}: {
  readonly owner: WorkbookContextualTaskDecisionCreateOwner;
}) {
  const current = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const token = useRef(Symbol("contextual-create-recovery")).current;
  const items = new Map<string, WorkbookRecoveryItem>();
  if (current.authority) {
    if (current.draft) {
      const draft = current.draft;
      items.set(String(draft.id), {
        id: String(draft.id),
        label: `${draft.target.title} draft`,
        summary: current.preparing ? "Preparing submission" : "Draft retained",
        origin: draft.presentation.label,
        sheetRef: draft.presentation.sheetRef,
        order: draft.id,
        attention: current.preparing ? "progress" : "draft",
      });
    }
    for (const entry of current.entries) {
      const draft = entry.attempt.review.draft;
      items.set(String(draft.id), {
        id: String(draft.id),
        label: `${draft.target.title} creation`,
        refreshViews:
          entry.receipt && entry.refresh !== "complete"
            ? [
                draft.source.viewSchemaId,
                draft.target.viewSchemaId,
                ...entry.observations.flatMap((event) =>
                  event.payload.affected_views.map(
                    (view) => view.view_schema_id,
                  ),
                ),
              ]
            : [],
        origin: draft.presentation.label,
        sheetRef: draft.presentation.sheetRef,
        order: draft.id,
        summary: entry.receipt
          ? entry.refresh === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : entry.phase === "uncertain"
            ? "Outcome unconfirmed"
            : entry.phase === "rejected"
              ? "Review retained draft"
              : "Creating",
        attention:
          entry.receipt && entry.refresh === "complete"
            ? "completed"
            : entry.receipt ||
                entry.phase === "uncertain" ||
                entry.phase === "rejected"
              ? "attention"
              : "progress",
      });
    }
  }
  const selected = useWorkbookRecoverySource(
    "contextual-create",
    [...items.values()],
    { detach: () => owner.detach(token) },
  );
  const snapshot = {
    ...current,
    draft:
      current.draft && String(current.draft.id) === selected
        ? current.draft
        : null,
    entries: current.entries.filter(
      (entry) => String(entry.attempt.review.draft.id) === selected,
    ),
  };
  return (
    <WorkbookRecoveryDetail source="contextual-create" item={selected}>
      <section aria-label="Retained contextual creation">
        {snapshot.draft && snapshot.attachment !== token ? (
          <div>
            <p>
              {snapshot.draft.target.title} draft retained from{" "}
              {snapshot.draft.presentation.surfaceLabel}.
            </p>
            <WorkbookInspectorActionButton
              tone="secondary"
              type="button"
              onClick={() => owner.resume(token)}
            >
              Resume contextual draft
            </WorkbookInspectorActionButton>{" "}
            <WorkbookInspectorActionButton
              tone="secondary"
              type="button"
              disabled={owner.busy}
              onClick={() => owner.discard()}
            >
              Discard draft
            </WorkbookInspectorActionButton>
          </div>
        ) : null}
        {snapshot.draft ? (
          <ContextualCreateForm
            owner={owner}
            attachment={token}
            disabled={owner.busy}
            onSubmit={() => void owner.submit(token)}
          />
        ) : null}
        <ul>
          {snapshot.entries.map((entry) => (
            <li key={entry.attempt.clientTxnId}>
              <p>
                {entry.receipt
                  ? `${entry.attempt.review.draft.target.title} created.`
                  : entry.phase === "uncertain"
                    ? "Creation outcome unconfirmed."
                    : entry.phase === "rejected"
                      ? "Creation rejected; the draft is retained."
                      : "Creating…"}{" "}
                {entry.message}
              </p>
              {entry.receipt ? (
                <details>
                  <summary>Creation receipt</summary>
                  <p>
                    Record: {entry.receipt.data.row.record_id}. Version:{" "}
                    {entry.receipt.data.row.row_version}. Change set:{" "}
                    {entry.receipt.data.change_set_id}. Request:{" "}
                    {entry.receipt.meta.request_id}.
                  </p>
                </details>
              ) : null}
              {entry.phase === "uncertain" ? (
                <WorkbookInspectorActionButton
                  tone="secondary"
                  type="button"
                  disabled={
                    entry.transportPending ||
                    snapshot.preparing ||
                    !owner.canReplay()
                  }
                  onClick={() => void owner.replay(entry.attempt.clientTxnId)}
                >
                  Recover original creation
                </WorkbookInspectorActionButton>
              ) : null}
              {entry.receipt && entry.refresh !== "complete" ? (
                <WorkbookInspectorActionButton
                  tone="secondary"
                  type="button"
                  disabled={entry.refresh === "refreshing"}
                  onClick={() =>
                    void owner.retryRefresh(entry.attempt.clientTxnId)
                  }
                >
                  Retry creation refresh
                </WorkbookInspectorActionButton>
              ) : null}
            </li>
          ))}
        </ul>
      </section>
    </WorkbookRecoveryDetail>
  );
}
