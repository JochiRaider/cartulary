import { useRef, useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { WorkbookInspectorActionButton as Button } from "../../inspector/presentation/WorkbookInspectorActions";
import { workbookConflictQueueKey } from "../../runtime/workbookConflictModel";
import { TimelineRelatedEvidenceForm } from "./TimelineRelatedEvidenceForm";
import { TimelineRelatedEvidenceOutcome } from "./TimelineRelatedEvidenceOutcome";
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
          <TimelineRelatedEvidenceOutcome
            key={checkpoint.id}
            owner={owner}
            checkpoint={checkpoint}
            draftId={state.draft?.id}
            preparing={state.preparing}
          />
        ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
