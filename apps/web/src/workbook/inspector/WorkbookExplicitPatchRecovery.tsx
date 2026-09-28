import { requireViewContract } from "@cartulary/view-contracts";
import { useId, useLayoutEffect, useMemo, useSyncExternalStore } from "react";
import { genericInspectorRowLabel } from "../models/genericWorkbookModel";
import type { WorkbookExplicitPatchOwner } from "../runtime/WorkbookExplicitPatchOwner";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
import { WorkbookInspectorNoticeView } from "./presentation/WorkbookInspectorFeedback";
import { workbookInspectorErrorPresentation } from "./workbookInspectorErrorModel";

/** Surface-local recovery remains reachable after the originating inspector closes. */
export function WorkbookExplicitPatchRecovery({
  owner,
  viewSchemaId,
  recordId,
  fieldKey,
  includeInitialFailures = false,
}: {
  owner: WorkbookExplicitPatchOwner;
  viewSchemaId: string;
  recordId?: string;
  fieldKey?: string;
  includeInitialFailures?: boolean;
}) {
  const read = useMemo(() => {
    let previous:
      | {
          authority: ReturnType<typeof owner.getSnapshot>["authority"];
          entries: ReturnType<typeof owner.getSnapshot>["entries"];
        }
      | undefined;
    return () => {
      const snapshot = owner.getSnapshot();
      const entries = snapshot.entries.filter(
        (entry) =>
          entry.intent.viewSchemaId === viewSchemaId &&
          (recordId !== undefined || !owner.resultIsAttached(entry)) &&
          (recordId === undefined ||
            includeInitialFailures ||
            owner.hasRecoveryAttempt(entry.id) ||
            (entry.phase !== "rejected" &&
              entry.phase !== "preparation_failed")) &&
          (recordId === undefined ||
            entry.intent.baseline.record_id === recordId) &&
          (fieldKey === undefined ||
            entry.intent.changes.some(
              (change) => change.field_key === fieldKey,
            )) &&
          entry.intent.owner !== "party_link" &&
          entry.intent.purpose !== "task-lifecycle",
      );
      if (
        previous?.authority === snapshot.authority &&
        previous.entries.length === entries.length &&
        entries.every((entry, index) => entry === previous?.entries[index])
      )
        return previous;
      previous = { authority: snapshot.authority, entries };
      return previous;
    };
  }, [owner, viewSchemaId, recordId, fieldKey, includeInitialFailures]);
  const snapshot = useSyncExternalStore(owner.subscribe, read);
  const attachment = useId();
  useLayoutEffect(
    () =>
      recordId && fieldKey
        ? owner.attachInspectorResult(
            attachment,
            viewSchemaId,
            recordId,
            fieldKey,
          )
        : undefined,
    [owner, attachment, viewSchemaId, recordId, fieldKey],
  );
  const entries = snapshot.entries;
  if (!snapshot.authority || !entries.length) return null;
  const contract = requireViewContract(viewSchemaId);
  return (
    <section
      aria-label="Inspector changes"
      style={{
        display: "grid",
        gap: "var(--ct-spacing-xs)",
        maxBlockSize: "calc(var(--ct-layout-statusStripHeight) * 5)",
        overflowY: "auto",
        padding: "var(--ct-spacing-xs)",
      }}
    >
      {entries.map((entry) => (
        <div key={entry.id}>
          <p style={{ margin: 0 }}>
            {genericInspectorRowLabel(contract, entry.intent.baseline)} —{" "}
            {entry.intent.operationLabel ??
              entry.intent.changes
                .map(
                  (change) =>
                    contract.fieldMap[change.field_key]?.label ?? "Field",
                )
                .join(", ")}
            :{" "}
            {entry.receipt
              ? `Saved, version ${entry.receipt.row.row_version}.`
              : entry.phase === "uncertain"
                ? "The outcome could not be confirmed. The original request is retained."
                : entry.phase === "coordinating"
                  ? "Waiting for earlier writes and current saved values."
                  : entry.phase === "submitting"
                    ? "Saving changes."
                    : entry.phase === "conflict"
                      ? "Review the saved-field conflict. Your draft is retained."
                      : "The change was not accepted. Your draft is retained."}
          </p>
          <WorkbookInspectorNoticeView
            consume={owner.inspectorNotices.consume}
            notice={owner.inspectorNotice(
              entry,
              entry.failure
                ? {
                    kind: "error",
                    error: workbookInspectorErrorPresentation(entry.failure),
                  }
                : {
                    kind: "message",
                    announcement: "none",
                    message: entry.receipt
                      ? entry.reconciliation === "complete"
                        ? "Change saved."
                        : "The change was saved. The view still needs a refresh."
                      : entry.phase === "uncertain"
                        ? "Outcome unconfirmed. Recover the original request."
                        : "Saving inspector changes…",
                  },
            )}
          />
          {entry.phase === "uncertain" ? (
            <Button
              type="button"
              disabled={!owner.canSubmit()}
              onClick={() => void owner.replay(entry.id)}
            >
              Retry original change
            </Button>
          ) : null}
          {entry.receipt && entry.reconciliation === "required" ? (
            <Button type="button" onClick={() => void owner.refresh(entry.id)}>
              Refresh saved change
            </Button>
          ) : null}
          {entry.phase === "preparation_failed" ||
          entry.phase === "rejected" ||
          (entry.receipt && entry.reconciliation === "complete") ? (
            <Button type="button" onClick={() => owner.dismiss(entry.id)}>
              Dismiss change notice
            </Button>
          ) : null}
        </div>
      ))}
    </section>
  );
}
