import { useEffect, useState, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import type { InspectorRecordHistoryAction } from "../../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { useWorkbookRecordHistoryController } from "../../inspector/useWorkbookRecordHistoryController";
import { useWorkbookRecordHistoryState } from "../../inspector/useWorkbookRecordHistoryState";
import { WorkbookRecordHistoryPanel } from "../../inspector/WorkbookInspectorRecordHistory";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { EntityMergeOperation } from "./entityMergeOperation";

export function WorkbookEntityMergeRecovery({
  runtime,
}: {
  readonly runtime: WorkbookMutationRuntime;
}) {
  const owner = runtime.entityMerge;
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: readonly WorkbookRecoveryItem[] = snapshot.authority
    ? snapshot.entries.map((entry, order) => ({
        id: entry.attempt.id,
        label: "Entity merge",
        origin: `${entry.attempt.review.loser.label} → ${entry.attempt.review.survivor.label}`,
        sheetRef: {
          kind: "view_schema",
          id:
            entry.attempt.review.entityType === "host"
              ? hostsViewSchemaId
              : identitiesViewSchemaId,
        },
        refreshViews:
          entry.receipt && entry.reconciliation !== "complete"
            ? [
                entry.attempt.review.entityType === "host"
                  ? hostsViewSchemaId
                  : identitiesViewSchemaId,
              ]
            : [],
        order,
        summary: entry.receipt
          ? entry.reconciliation === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : entry.phase === "uncertain"
            ? "Outcome unconfirmed"
            : entry.phase === "rejected"
              ? "Review required"
              : "In progress",
        attention:
          entry.receipt && entry.reconciliation === "complete"
            ? "completed"
            : entry.receipt ||
                entry.phase === "uncertain" ||
                entry.phase === "rejected"
              ? "attention"
              : "progress",
      }))
    : [];
  const selected = useWorkbookRecoverySource("entity-merge", items);
  return (
    <WorkbookRecoveryDetail source="entity-merge" item={selected}>
      <section aria-label="Merge action recovery">
        {snapshot.entries
          .filter((entry) => entry.attempt.id === selected)
          .map((entry) => (
            <MergeRecoveryEntry
              key={entry.attempt.id}
              entry={entry}
              runtime={runtime}
            />
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}

function MergeRecoveryEntry({
  entry,
  runtime,
}: {
  readonly entry: EntityMergeOperation;
  readonly runtime: WorkbookMutationRuntime;
}) {
  const [historyOpen, setHistoryOpen] = useState(false);
  const owner = runtime.entityMerge;
  const { review } = entry.attempt;
  const receipt = entry.receipt;
  const pending =
    entry.transportPending ||
    entry.phase === "preparing" ||
    entry.phase === "submitting";
  return (
    <article
      aria-label={`Merge ${review.entityType}: ${review.loser.label} into ${review.survivor.label}`}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-sm)",
        borderBlockStart: "var(--ct-border-hairline)",
        paddingBlock: "var(--ct-spacing-md)",
      }}
    >
      <strong>
        {review.entityType === "host" ? "Host" : "Identity"} merge
      </strong>
      <p style={{ margin: 0 }}>
        Survivor: {review.survivor.label} ({review.survivor.recordId})<br />
        Historical loser: {review.loser.label} ({review.loser.recordId})
      </p>
      <p style={{ margin: 0 }}>Reviewed reason: {review.reason}</p>
      <p role="status" style={{ margin: 0 }}>
        {mergeOperationStatus(entry)}
      </p>
      {entry.failure ? (
        <p style={{ margin: 0 }}>{entry.failure.message}</p>
      ) : null}
      {entry.failure?.kind === "validation" && entry.failure.fields?.length ? (
        <dl>
          {entry.failure.fields.map((field) => (
            <div key={field.field}>
              <dt>{field.field.replaceAll("_", " ")}</dt>
              <dd style={{ marginInlineStart: 0, whiteSpace: "pre-wrap" }}>
                {field.message}
              </dd>
            </div>
          ))}
        </dl>
      ) : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          disabled={pending || !owner.canSubmit()}
          onClick={() => void owner.replay(entry.attempt.id)}
        >
          Replay exact merge request
        </WorkbookInspectorActionButton>
      ) : null}
      {receipt ? (
        <>
          <p style={{ margin: 0 }}>
            Merge completed. Change set: {receipt.change_set_id}. The loser
            remains in history.
          </p>
          <details>
            <summary>Merge receipt</summary>
            <p style={{ margin: 0 }}>
              Incident: {receipt.incident_id}
              <br />
              Survivor version: {receipt.survivor_row_version}; loser version:{" "}
              {receipt.loser_row_version}
            </p>
            <table
              style={{
                inlineSize: "100%",
                textAlign: "start",
                borderSpacing: "var(--ct-spacing-xs)",
              }}
            >
              <caption>Reusable identifiers by class</caption>
              <thead>
                <tr>
                  <th scope="col">Class</th>
                  <th scope="col">Promoted</th>
                  <th scope="col">Carried</th>
                  <th scope="col">Duplicates</th>
                </tr>
              </thead>
              <tbody>
                {receipt.merge_summary.exact_match_classes.map((row) => (
                  <tr key={row.identifier_class}>
                    <th scope="row">
                      {row.identifier_class.replaceAll("_", " ")}
                    </th>
                    <td>{row.promoted_count}</td>
                    <td>{row.carried_count}</td>
                    <td>{row.duplicate_noop_count}</td>
                  </tr>
                ))}
              </tbody>
            </table>
            <p style={{ margin: 0 }}>
              Mentions repointed:{" "}
              {receipt.merge_summary.repointed_mention_resolution_count}. Links
              repointed: {receipt.merge_summary.repointed_link_count}; duplicate
              links: {receipt.merge_summary.deduped_link_count}. Tags repointed:{" "}
              {receipt.merge_summary.repointed_tag_count}; duplicate tags:{" "}
              {receipt.merge_summary.deduped_tag_count}. Assessments repointed:{" "}
              {receipt.merge_summary.repointed_assessment_count}.
            </p>
            <p style={{ margin: 0 }}>
              Suggestion aliases copied:{" "}
              {receipt.merge_summary.suggestion_aliases_copied_count}; duplicate
              aliases:{" "}
              {receipt.merge_summary.suggestion_alias_duplicate_noop_count}.
              Provenance-only identifiers retained:{" "}
              {receipt.merge_summary.provenance_only_retained_count}.
            </p>
          </details>
          {entry.reconciliation !== "complete" ? (
            <WorkbookInspectorActionButton
              disabled={pending || entry.reconciliation === "refreshing"}
              onClick={() => void owner.refresh(entry.attempt.id)}
            >
              Refresh completed merge
            </WorkbookInspectorActionButton>
          ) : null}
          <WorkbookInspectorActionButton
            onClick={() => setHistoryOpen((value) => !value)}
            aria-expanded={historyOpen}
          >
            {historyOpen ? "Close merge history" : "Review merge history"}
          </WorkbookInspectorActionButton>
          {historyOpen ? (
            <MergeHistoryReview entry={entry} runtime={runtime} />
          ) : null}
        </>
      ) : null}
      {!pending &&
      (entry.phase === "rejected" ||
        (receipt && entry.reconciliation === "complete")) ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          onClick={() => owner.dismiss(entry.attempt.id)}
        >
          Dismiss merge action
        </WorkbookInspectorActionButton>
      ) : null}
    </article>
  );
}
function mergeOperationStatus(entry: EntityMergeOperation): string {
  if (entry.phase === "preparing")
    return "Waiting for earlier writes and checking the captured review.";
  if (entry.phase === "submitting")
    return "Merge submitted; waiting for acknowledgement.";
  if (entry.phase === "uncertain")
    return entry.transportPending
      ? "The merge outcome is unknown. Transport has not settled; the server may have committed."
      : "The merge outcome is unknown. Explicit replay sends the same transaction and exact request, even if the loser is no longer visible.";
  if (entry.phase === "rejected")
    return "Merge rejected. Refresh both records and review again before a new attempt.";
  return entry.reconciliation === "complete"
    ? "Merge acknowledged; current projections refreshed."
    : entry.reconciliation === "refreshing"
      ? "Merge acknowledged; refreshing current projections."
      : "Merge acknowledged. Refresh is still required; refreshing will not send another merge.";
}
const rollbackActions: ReadonlySet<InspectorRecordHistoryAction> = new Set([
  "rollback",
]);
function MergeHistoryReview({
  entry,
  runtime,
}: {
  readonly entry: EntityMergeOperation;
  readonly runtime: WorkbookMutationRuntime;
}) {
  const review = entry.attempt.review;
  const receipt = entry.receipt;
  const recordId = review.survivor.recordId;
  const controller = useWorkbookRecordHistoryController({
    presentation: useWorkbookRecordHistoryState(),
    owner: runtime.history,
    subject: {
      kind: "live",
      recordId,
      label: review.survivor.label,
      rowVersion: Math.max(
        receipt?.survivor_row_version ?? review.survivor.baseRowVersion,
        runtime.history.latestVersion(recordId) ?? 0,
      ),
      surfaceLabel: review.entityType === "host" ? "Hosts" : "Identities",
      viewSchemaId:
        review.entityType === "host"
          ? hostsViewSchemaId
          : identitiesViewSchemaId,
    },
    coordinate: runtime.coordinateHistory.bind(runtime),
    canMutate: runtime.entityMerge.canSubmit(),
    ownerEffects: {
      deleteAccepted: () => {},
      restoreAccepted: () => {},
      rollbackAccepted: () => {},
      refresh: () =>
        runtime.entityMerge.refresh(entry.attempt.id, {
          requireAcceptance: true,
        }),
    },
  });
  const open = controller.commands.open;
  useEffect(() => {
    open();
  }, [open]);
  return (
    <>
      <p>
        Review the merge change set in existing History. Reversal uses the
        existing change-set rollback confirmation.
      </p>
      <WorkbookRecordHistoryPanel
        browsingControls={controller.commands}
        actions={rollbackActions}
        canMutate={runtime.entityMerge.canSubmit()}
        idleRecordId={recordId}
        state={controller.snapshot}
        onOpenHistory={open}
        onCancelPendingAction={controller.commands.cancel}
        onConfirmPendingAction={() => void controller.commands.confirm()}
        onPreviewDeleteRestore={controller.commands.previewDeleteRestore}
        onPreviewRollback={controller.commands.previewRollback}
      />
    </>
  );
}
