import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { InspectorRecordHistoryAction } from "../../inspector/inspectorCapabilityResolver";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { useWorkbookRecordHistoryController } from "../../inspector/useWorkbookRecordHistoryController";
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
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const summary = useRef<HTMLElement>(null);
  const focusRequested = useRef(false);
  useLayoutEffect(() => {
    if (open && focusRequested.current) {
      focusRequested.current = false;
      summary.current?.focus({ preventScroll: true });
    }
  }, [open]);
  useEffect(() => {
    if (!snapshot.authority) setOpen(false);
  }, [snapshot.authority]);
  if (!snapshot.entries.length) return null;
  const close = () => {
    const restore = summary.current?.parentElement?.contains(
      document.activeElement,
    );
    setOpen(false);
    if (restore) trigger.current?.focus({ preventScroll: true });
  };
  const acknowledged = snapshot.entries.filter(
    (entry) => entry.receipt !== null,
  ).length;
  const unknown = snapshot.entries.filter(
    (entry) => entry.phase === "uncertain",
  ).length;
  const refreshRequired = snapshot.entries.some(
    (entry) => entry.receipt && entry.reconciliation !== "complete",
  );
  return (
    <div style={{ position: "relative" }}>
      <WorkbookInspectorActionButton
        ref={trigger}
        aria-expanded={open}
        onClick={() => {
          if (open) close();
          else {
            focusRequested.current = true;
            setOpen(true);
          }
        }}
      >
        Merge actions ({snapshot.entries.length})
      </WorkbookInspectorActionButton>
      <span
        role="status"
        style={{
          marginInlineStart: "var(--ct-spacing-xs)",
          fontSize: "var(--ct-typography-compact-metadata-fontSize)",
        }}
      >
        {unknown
          ? `${unknown} merge outcome${unknown === 1 ? "" : "s"} unknown.`
          : acknowledged
            ? `${acknowledged} merge${acknowledged === 1 ? "" : "s"} completed.${refreshRequired ? " Refresh still required." : ""}`
            : "Merge in progress."}
      </span>
      {open ? (
        <section
          aria-label="Merge action recovery"
          style={{
            position: "absolute",
            zIndex: 30,
            insetInlineEnd: 0,
            inlineSize: "min(38rem, 90vw)",
            maxBlockSize: "75vh",
            overflow: "auto",
            overflowWrap: "anywhere",
            padding: "var(--ct-spacing-md)",
            background: "var(--ct-colors-surface-1)",
            border: "var(--ct-border-hairline)",
            boxShadow: "var(--ct-elevation-popover)",
          }}
          onKeyDown={(event) => {
            if (
              event.key === "Escape" &&
              !(
                event.target instanceof Element &&
                event.target.closest('[role="alertdialog"]')
              )
            ) {
              event.stopPropagation();
              close();
            }
          }}
        >
          <section
            ref={summary}
            tabIndex={-1}
            aria-label="Merge action recovery summary"
          >
            <strong>Merge actions</strong>
            <p>
              Admitted requests remain here when the inspector closes. A timeout
              or panel closure does not roll back a merge.
            </p>
          </section>
          <WorkbookInspectorActionButton onClick={close}>
            Close merge actions
          </WorkbookInspectorActionButton>
          {snapshot.entries.map((entry) => (
            <MergeRecoveryEntry
              key={entry.attempt.id}
              entry={entry}
              runtime={runtime}
            />
          ))}
        </section>
      ) : null}
    </div>
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
