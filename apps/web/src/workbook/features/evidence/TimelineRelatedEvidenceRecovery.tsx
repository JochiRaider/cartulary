import { useRef, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { workbookConflictQueueKey } from "../../runtime/workbookConflictModel";
import { TimelineRelatedEvidenceForm } from "./TimelineRelatedEvidenceForm";
import type { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";
export function TimelineRelatedEvidenceRecovery({
  owner,
}: {
  readonly owner: WorkbookTimelineRelatedEvidenceOwner;
}) {
  const current = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const token = useRef(Symbol("related-evidence-recovery")).current;
  const items = new Map<string, WorkbookRecoveryItem>();
  if (current.authority) {
    if (current.draft) {
      const draft = current.draft;
      items.set(String(draft.id), {
        id: String(draft.id),
        label: "Timeline Evidence draft",
        summary: current.preparing ? "Preparing submission" : "Draft retained",
        origin: draft.presentation.label,
        sheetRef: draft.presentation.sheetRef,
        order: draft.id,
        attention: current.preparing ? "progress" : "draft",
      });
    }
    for (const checkpoint of current.checkpoints) {
      const draft = checkpoint.create.attempt.review.draft;
      const stages = [checkpoint.create, ...checkpoint.links];
      const conflictKeys = stages.flatMap((stage) =>
        stage.failure?.kind === "same_field_conflict"
          ? [workbookConflictQueueKey(stage.failure.conflict)]
          : [],
      );
      const completed =
        owner.linkComplete(checkpoint) &&
        stages.every((stage) => !stage.receipt || stage.refresh === "complete");
      const active = stages.some((stage) => stage.phase === "submitting");
      items.set(String(draft.id), {
        id: String(draft.id),
        label: "Timeline Evidence creation",
        conflictKeys,
        priority: conflictKeys.length ? 2 : 10,
        refreshViews: stages.some(
          (stage) => stage.receipt && stage.refresh !== "complete",
        )
          ? [
              "cartulary.view.timeline.v2",
              "cartulary.view.evidence.v1",
              "cartulary.view.parties.v1",
            ]
          : [],
        origin: draft.presentation.label,
        sheetRef: draft.presentation.sheetRef,
        order: draft.id,
        summary: completed
          ? "Created and linked"
          : active
            ? "Creating or linking"
            : "Review creation, link or refresh",
        attention: completed ? "completed" : active ? "progress" : "attention",
      });
    }
  }
  const selected = useWorkbookRecoverySource(
    "related-evidence",
    [...items.values()],
    { detach: () => owner.detach(token) },
  );
  const state = {
    ...current,
    draft:
      current.draft && String(current.draft.id) === selected
        ? current.draft
        : null,
    checkpoints: current.checkpoints.filter(
      (checkpoint) =>
        String(checkpoint.create.attempt.review.draft.id) === selected,
    ),
  };
  return (
    <WorkbookRecoveryDetail source="related-evidence" item={selected}>
      <section aria-label="Retained Timeline Evidence creation">
        {state.draft && state.attachment !== token ? (
          <div>
            <p>
              Evidence metadata is retained for its original Timeline record.
            </p>
            <Button
              type="button"
              tone="secondary"
              onClick={() => owner.resume(token)}
            >
              Resume Evidence draft
            </Button>
            <Button
              type="button"
              tone="secondary"
              disabled={owner.busy}
              onClick={() => owner.discard()}
            >
              Discard Evidence draft
            </Button>
          </div>
        ) : null}
        {state.draft ? (
          <TimelineRelatedEvidenceForm
            owner={owner}
            attachment={token}
            onSubmit={() => void owner.submit(token)}
            onReview={() => void owner.review()}
          />
        ) : null}
        {state.checkpoints.map((checkpoint) => (
          <section
            key={checkpoint.id}
            aria-label="Evidence creation result"
            style={{
              borderBlockStart: "var(--ct-border-hairline)",
              paddingBlock: "var(--ct-spacing-sm)",
            }}
          >
            <p>
              Timeline source:{" "}
              {checkpoint.create.attempt.review.draft.presentation.label}
            </p>
            <details>
              <summary>Original Timeline record</summary>
              <p>{checkpoint.create.attempt.review.draft.source.recordId}</p>
            </details>
            <p role="status">
              {checkpoint.create.receipt
                ? owner.linkComplete(checkpoint)
                  ? "Evidence created and linked to the original Timeline record."
                  : "Evidence created; Timeline link incomplete."
                : checkpoint.create.phase === "rejected"
                  ? state.draft?.id ===
                    checkpoint.create.attempt.review.draft.id
                    ? "Evidence creation rejected. Your draft is retained."
                    : "Evidence creation rejected."
                  : checkpoint.create.phase === "uncertain"
                    ? "Evidence creation outcome unconfirmed."
                    : "Creating Evidence…"}
            </p>
            {checkpoint.message &&
            checkpoint.message !==
              "Evidence created; Timeline link incomplete." &&
            checkpoint.message !==
              "Evidence created and linked to the original Timeline record." ? (
              <p>{checkpoint.message}</p>
            ) : null}
            {checkpoint.sourceUnavailable ? (
              <p>
                The original Timeline record is unavailable or superseded. The
                created Evidence is retained.
              </p>
            ) : null}
            {checkpoint.evidenceUnavailable ? (
              <p>
                The created Evidence is currently unavailable. Its accepted
                creation is retained.
              </p>
            ) : null}
            {[checkpoint.create, ...checkpoint.links].map((entry) => (
              <div key={entry.attempt.clientTxnId}>
                <p>
                  {entry.attempt.stage === "create"
                    ? "Creation"
                    : "Timeline link"}
                  :{" "}
                  {entry.phase === "accepted"
                    ? `saved; ${entry.refresh === "complete" ? "views refreshed" : "views need refresh"}`
                    : entry.phase === "uncertain"
                      ? "outcome unconfirmed"
                      : entry.phase === "rejected"
                        ? "rejected"
                        : "submitting"}
                  .
                </p>
                {entry.failure ? <p>{entry.failure.message}</p> : null}
                {entry.receipt ? (
                  <details>
                    <summary>
                      {entry.attempt.stage === "create"
                        ? "Evidence creation receipt"
                        : "Timeline link receipt"}
                    </summary>
                    <p>
                      Record: {entry.receipt.data.row.record_id}. Version:{" "}
                      {entry.receipt.data.row.row_version}. Change set:{" "}
                      {entry.receipt.data.change_set_id}. Request:{" "}
                      {entry.receipt.meta.request_id}.
                    </p>
                  </details>
                ) : null}
                {entry.phase === "uncertain" ? (
                  <Button
                    type="button"
                    tone="primary"
                    disabled={
                      state.preparing ||
                      entry.transportPending ||
                      !owner.canReplay()
                    }
                    onClick={() =>
                      void owner.replay(
                        checkpoint.id,
                        entry.attempt.clientTxnId,
                      )
                    }
                  >
                    {entry.attempt.stage === "create"
                      ? "Recover Evidence creation"
                      : "Recover Timeline link"}
                  </Button>
                ) : null}
                {entry.phase === "rejected" &&
                entry.failure?.kind === "client_txn_conflict" ? (
                  <Button
                    type="button"
                    tone="secondary"
                    disabled={owner.busy || !owner.canSubmit()}
                    onClick={() =>
                      owner.reviewNewRequestId(
                        checkpoint.id,
                        entry.attempt.clientTxnId,
                      )
                    }
                  >
                    Retry with a new request ID
                  </Button>
                ) : null}
              </div>
            ))}
            {checkpoint.resolution ? (
              <details>
                <summary>Collection review receipt</summary>
                <p>
                  Record: {checkpoint.resolution.receipt.data.row.record_id}.
                  Version: {checkpoint.resolution.receipt.data.row.row_version}.
                  Request: {checkpoint.resolution.receipt.meta.request_id}.
                </p>
              </details>
            ) : null}
            {checkpoint.create.receipt && !owner.linkComplete(checkpoint) ? (
              <div>
                <Button
                  type="button"
                  tone="secondary"
                  disabled={!owner.canSubmit()}
                  aria-disabled={owner.busy || !owner.canSubmit()}
                  onClick={() => {
                    if (!owner.busy) void owner.reviewLink(checkpoint.id);
                  }}
                >
                  Review original Timeline link
                </Button>
                {checkpoint.review ? (
                  <Button
                    type="button"
                    tone="primary"
                    disabled={owner.busy || !owner.canSubmit()}
                    onClick={() => void owner.link(checkpoint.id)}
                  >
                    Link created Evidence
                  </Button>
                ) : null}
              </div>
            ) : null}
            {checkpoint.create.receipt ? (
              <Button
                type="button"
                tone="secondary"
                disabled={
                  state.preparing ||
                  [checkpoint.create, ...checkpoint.links].some(
                    (entry) => entry.refresh === "refreshing",
                  )
                }
                onClick={() => void owner.retryRefresh(checkpoint.id)}
              >
                Refresh Evidence result
              </Button>
            ) : null}
          </section>
        ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
