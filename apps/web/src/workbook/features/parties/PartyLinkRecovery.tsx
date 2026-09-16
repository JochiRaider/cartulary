import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "../../inspector/presentation/WorkbookInspectorFeedback";
import { workbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import type { ExplicitPatchOperation } from "../../runtime/WorkbookExplicitPatchOwner";
import type { WorkbookPartyLinkOperationOwner } from "./WorkbookPartyLinkOperationOwner";

export function PartyPatchFeedback({
  owner,
  entry,
}: {
  owner: WorkbookPartyLinkOperationOwner;
  entry: ExplicitPatchOperation;
}) {
  return (
    <section aria-label="Party source operation">
      <p role="status">
        {entry.receipt
          ? entry.reconciliation === "complete"
            ? "Source change saved."
            : "Source change saved; refresh is still required."
          : entry.phase === "uncertain"
            ? "Source change outcome is uncertain. The original request is retained."
            : entry.phase === "conflict"
              ? "The source field changed. Resolve its conflict, then review the complete Party action again."
              : entry.phase === "preparation_failed"
                ? "The source change was not sent. Review the current source and try again."
                : entry.phase === "rejected"
                  ? "The source change was rejected. Review current source values before another attempt."
                  : "Saving source change…"}
      </p>
      {entry.failure ? (
        <WorkbookInspectorPublicError
          error={workbookInspectorErrorPresentation(entry.failure)}
        />
      ) : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          disabled={!owner.canSubmit()}
          onClick={() => void owner.patches.replay(entry.id)}
        >
          Replay original source change
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.receipt && entry.reconciliation === "required" ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          onClick={() => void owner.patches.refresh(entry.id)}
        >
          Refresh source result
        </WorkbookInspectorActionButton>
      ) : null}
    </section>
  );
}
export function PartyLinkRecovery({
  owner,
}: {
  owner: WorkbookPartyLinkOperationOwner;
}) {
  const current = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: WorkbookRecoveryItem[] = [];
  if (current.authority) {
    for (const [order, entry] of current.creations.entries()) {
      const link = current.patches.find((patch) => patch.id === entry.linkId);
      const completed =
        entry.receipt &&
        entry.refresh === "complete" &&
        link?.receipt &&
        link.reconciliation === "complete";
      const progress =
        entry.phase === "preparing" ||
        entry.phase === "submitting" ||
        link?.phase === "coordinating" ||
        link?.phase === "submitting";
      items.push({
        id: entry.id,
        operationIds: entry.linkId ? [entry.linkId] : [],
        label: "Party creation and link",
        priority: link?.phase === "conflict" ? 2 : 10,
        refreshViews: [
          ...(entry.receipt && entry.refresh !== "complete"
            ? ["cartulary.view.parties.v1"]
            : []),
          ...(link?.receipt && link.reconciliation !== "complete"
            ? [entry.attempt.review.pair.viewSchemaId]
            : []),
        ],
        origin: entry.attempt.review.sourceLabel,
        sheetRef: entry.attempt.review.sheetRef,
        order,
        summary: completed
          ? "Created and linked"
          : progress
            ? "Creating or linking"
            : "Review creation, link or refresh",
        attention: completed
          ? "completed"
          : progress
            ? "progress"
            : "attention",
      });
    }
    for (const [order, entry] of current.patches.entries()) {
      if (current.creations.some((creation) => creation.linkId === entry.id))
        continue;
      items.push({
        id: entry.id,
        operationIds: [entry.id],
        label: "Party source change",
        priority: entry.phase === "conflict" ? 2 : 10,
        refreshViews:
          entry.receipt && entry.reconciliation !== "complete"
            ? [entry.intent.review.pair.viewSchemaId]
            : [],
        origin: entry.intent.review.sourceLabel,
        sheetRef: entry.intent.review.sheetRef,
        order,
        summary: entry.receipt
          ? entry.reconciliation === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : "Source change pending or needs review",
        attention:
          entry.receipt && entry.reconciliation === "complete"
            ? "completed"
            : entry.phase === "coordinating" || entry.phase === "submitting"
              ? "progress"
              : "attention",
      });
    }
  }
  const selected = useWorkbookRecoverySource("party-link", items);
  const snapshot = {
    ...current,
    creations: current.creations.filter((entry) => entry.id === selected),
    patches: current.patches.filter(
      (entry) =>
        entry.id === selected ||
        current.creations.some(
          (creation) =>
            creation.id === selected && creation.linkId === entry.id,
        ),
    ),
  };
  return (
    <WorkbookRecoveryDetail source="party-link" item={selected}>
      <section aria-label="Retained Party operations">
        {snapshot.creations.map((entry) => (
          <section key={entry.id}>
            <p>
              {entry.attempt.review.sourceLabel} —{" "}
              {entry.attempt.review.pair.label}
            </p>
            <p role="status">
              {entry.receipt
                ? `Party saved: ${String(entry.receipt.data.row.cells["party.display_name"]?.value ?? "Party")}. ${entry.linkId ? "Source linking has its own result." : "Return to this source and pair to review linking the saved Party."}`
                : entry.phase === "uncertain"
                  ? "Party creation outcome is uncertain."
                  : (entry.failure?.message ?? "Saving Party…")}
            </p>
            {entry.failure ? (
              <WorkbookInspectorPublicError
                error={workbookInspectorErrorPresentation(entry.failure)}
              />
            ) : null}
            {entry.phase === "uncertain" ? (
              <WorkbookInspectorActionButton
                tone="secondary"
                disabled={!owner.canSubmit()}
                onClick={() => void owner.replayCreation(entry.id)}
              >
                Replay Party creation
              </WorkbookInspectorActionButton>
            ) : null}
            {entry.receipt && entry.refresh === "required" ? (
              <>
                <p>Party saved; its view still needs refreshing.</p>
                <WorkbookInspectorActionButton
                  tone="secondary"
                  onClick={() => void owner.refreshCreation(entry.id)}
                >
                  Refresh saved Party
                </WorkbookInspectorActionButton>
              </>
            ) : null}
          </section>
        ))}
        {snapshot.patches.map((entry) => (
          <section key={entry.id}>
            <p>
              {entry.intent.review?.sourceLabel} —{" "}
              {entry.intent.review?.pair.label}
            </p>
            <PartyPatchFeedback owner={owner} entry={entry} />
          </section>
        ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
