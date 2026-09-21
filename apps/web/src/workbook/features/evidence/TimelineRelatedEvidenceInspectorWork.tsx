import { useEffect, useRef, useSyncExternalStore } from "react";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { TimelineRelatedEvidenceForm } from "./TimelineRelatedEvidenceForm";
import { TimelineRelatedEvidenceOutcome } from "./TimelineRelatedEvidenceOutcome";
import type { WorkbookTimelineRelatedEvidenceOwner } from "./WorkbookTimelineRelatedEvidenceOwner";

/** Local attachment to the existing owner; global recovery retains its own navigation. */
export function TimelineRelatedEvidenceInspectorWork({
  owner,
  recordId,
}: {
  readonly owner: WorkbookTimelineRelatedEvidenceOwner;
  readonly recordId: string;
}) {
  const state = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const token = useRef(Symbol("related-evidence-local-recovery")).current;
  useEffect(() => () => owner.detach(token), [owner, token]);
  if (!state.authority) return null;
  const draft = state.draft?.source.recordId === recordId ? state.draft : null;
  return (
    <>
      {draft && (state.attachment === null || state.attachment === token) ? (
        <div
          tabIndex={-1}
          data-evidence-work-id={`related-evidence-draft:${owner.incidentId}:${draft.id}`}
        >
          {state.attachment === null ? (
            <>
              <p>
                Evidence metadata is retained for this original Timeline record.
              </p>
              <Button onClick={() => owner.resume(token)}>
                Resume Evidence draft
              </Button>
              <Button disabled={owner.busy} onClick={() => owner.discard()}>
                Discard Evidence draft
              </Button>
            </>
          ) : (
            <TimelineRelatedEvidenceForm
              owner={owner}
              attachment={token}
              onSubmit={() => void owner.submit(token)}
              onReview={() => void owner.review()}
            />
          )}
        </div>
      ) : null}
      {state.checkpoints
        .filter(
          (checkpoint) =>
            checkpoint.create.attempt.review.draft.source.recordId === recordId,
        )
        .map((checkpoint) => (
          <TimelineRelatedEvidenceOutcome
            key={checkpoint.id}
            owner={owner}
            checkpoint={checkpoint}
            draftId={draft?.id}
            preparing={state.preparing}
          />
        ))}
    </>
  );
}
