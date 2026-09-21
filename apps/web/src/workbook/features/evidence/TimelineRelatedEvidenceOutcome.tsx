import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import type { RelatedEvidenceCheckpoint } from "./timelineRelatedEvidenceOperation";
import type { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";

/** Same retained checkpoint in local and global recovery; neither creates obligations. */
export function TimelineRelatedEvidenceOutcome({
  owner,
  checkpoint,
  draftId,
  preparing,
}: {
  readonly owner: WorkbookTimelineRelatedEvidenceOwner;
  readonly checkpoint: RelatedEvidenceCheckpoint;
  readonly draftId: number | undefined;
  readonly preparing: boolean;
}) {
  return (
    <section
      aria-label="Evidence creation result"
      tabIndex={-1}
      data-evidence-work-id={`related-evidence:${owner.incidentId}:${checkpoint.id}`}
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
      <p role="status" aria-live="off">
        {checkpoint.create.receipt
          ? owner.linkComplete(checkpoint)
            ? "Evidence created and linked to the original Timeline record."
            : "Evidence created; Timeline link incomplete."
          : checkpoint.create.phase === "rejected"
            ? draftId === checkpoint.create.attempt.review.draft.id
              ? "Evidence creation rejected. Your draft is retained."
              : "Evidence creation rejected."
            : checkpoint.create.phase === "uncertain"
              ? "Evidence creation outcome unconfirmed."
              : "Creating Evidence…"}
      </p>
      {checkpoint.message &&
      checkpoint.message !== "Evidence created; Timeline link incomplete." &&
      checkpoint.message !==
        "Evidence created and linked to the original Timeline record." ? (
        <p>{checkpoint.message}</p>
      ) : null}
      {checkpoint.sourceUnavailable ? (
        <p>
          The original Timeline record is unavailable or superseded. The created
          Evidence is retained.
        </p>
      ) : null}
      {checkpoint.evidenceUnavailable ? (
        <p>
          The created Evidence is currently unavailable. Its accepted creation
          is retained.
        </p>
      ) : null}
      {[checkpoint.create, ...checkpoint.links].map((entry) => (
        <div key={entry.attempt.clientTxnId}>
          <p>
            {entry.attempt.stage === "create" ? "Creation" : "Timeline link"}:{" "}
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
                preparing || entry.transportPending || !owner.canReplay()
              }
              onClick={() =>
                void owner.replay(checkpoint.id, entry.attempt.clientTxnId)
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
            Record: {checkpoint.resolution.receipt.data.row.record_id}. Version:{" "}
            {checkpoint.resolution.receipt.data.row.row_version}. Request:{" "}
            {checkpoint.resolution.receipt.meta.request_id}.
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
            preparing ||
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
  );
}
