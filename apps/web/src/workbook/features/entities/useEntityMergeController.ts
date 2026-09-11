import {
  useCallback,
  useEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorLocalErrorFeedback,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  buildMergePlan,
  type EntityMergePlan,
} from "../../models/entityMergePlan";
import type { EntityRow } from "../../models/entityWorkbookModel";
import type { WorkbookOperationFieldFailure } from "../../mutations/workbookOperationOutcome";

export type EntityMergePreconditionDetail = {
  readonly key: string;
  readonly label: string;
  readonly value: string;
};

export type EntityMergeController = {
  readonly commands: {
    readonly clearPlan: () => void;
    readonly confirm: () => Promise<void>;
    readonly reset: () => void;
    readonly review: () => void;
    readonly invalidateReview: () => void;
    readonly discardDrafts: () => void;
    readonly selectCandidate: (recordId: string) => void;
    readonly setReason: (reason: string) => void;
    readonly start: () => void;
  };
  readonly snapshot: {
    readonly candidateId: string;
    readonly loser: EntityRow | null;
    readonly feedback: WorkbookInspectorFeedback | null;
    readonly plan: EntityMergePlan | null;
    readonly preconditionDetails: readonly EntityMergePreconditionDetail[];
    readonly reason: string;
    readonly reviewed: EntityMergeReview | null;
    readonly hasAffectedDraft: boolean;
  };
};

import type { EntityMergeReview } from "./entityMergeReview";
import type { WorkbookEntityMergeOwner } from "./WorkbookEntityMergeOwner";

const defaultMergeReason = "Merge duplicate entity";

const mergeDetailLabels: Readonly<Record<string, string>> = {
  reason_code: "Reason",
  record_type: "Record type",
  identifier_class: "Identifier class",
  normalized_value: "Normalized value",
  blocking_record_id: "Blocking record",
  survivor_record_id: "Survivor record",
  loser_record_id: "Loser record",
  survivor_base_row_version: "Survivor supplied version",
  loser_base_row_version: "Loser supplied version",
  survivor_current_row_version: "Survivor current version",
  loser_current_row_version: "Loser current version",
};

function preconditionDetails(
  fields: readonly WorkbookOperationFieldFailure[] | undefined,
): readonly EntityMergePreconditionDetail[] {
  if (fields === undefined) return [];
  return fields.flatMap((field) => {
    const label = mergeDetailLabels[field.field];
    return label === undefined
      ? []
      : [{ key: field.field, label, value: field.message }];
  });
}

export function useEntityMergeController({
  canMerge,
  hasAffectedDraft,
  discardAffectedDrafts,
  lifecycleResetKey,
  loadSurvivorPreview,
  originSurface,
  owner,
  rows,
  selectedEntity,
}: {
  readonly canMerge: boolean;
  readonly hasAffectedDraft: (recordIds: readonly string[]) => boolean;
  readonly discardAffectedDrafts: (recordIds: readonly string[]) => void;
  readonly lifecycleResetKey: string;
  readonly loadSurvivorPreview: (recordId: string) => Promise<void>;
  readonly originSurface: string;
  readonly owner: WorkbookEntityMergeOwner;
  readonly rows: readonly EntityRow[];
  readonly selectedEntity: EntityRow | null;
}): EntityMergeController {
  const [candidateId, setCandidateId] = useState("");
  const [reason, setReasonState] = useState(defaultMergeReason);
  const [feedback, setFeedback] = useState<WorkbookInspectorFeedback | null>(
    null,
  );
  const [details, setDetails] = useState<
    readonly EntityMergePreconditionDetail[]
  >([]);
  const [, renderReview] = useState(0);
  const authority = useSyncExternalStore(
    owner.subscribe,
    owner.getSnapshot,
    owner.getSnapshot,
  );
  const reviewedRef = useRef<{ review: EntityMergeReview; key: string } | null>(
    null,
  );
  const generationRef = useRef(0);
  const inFlight = useRef(false);
  const loser = rows.find((row) => row.recordId === candidateId) ?? null;
  const plan =
    selectedEntity === null || loser === null
      ? null
      : buildMergePlan(selectedEntity, loser);
  const affectedDraft =
    selectedEntity !== null &&
    loser !== null &&
    hasAffectedDraft([selectedEntity.recordId, loser.recordId]);
  const presentationKey = JSON.stringify([
    lifecycleResetKey,
    originSurface,
    authority.generation,
    selectedEntity?.recordId,
  ]);
  const previousPresentationKey = useRef(presentationKey);
  if (previousPresentationKey.current !== presentationKey) {
    previousPresentationKey.current = presentationKey;
    generationRef.current++;
    reviewedRef.current = null;
  }
  const reviewKey = JSON.stringify([
    presentationKey,
    canMerge,
    candidateId,
    reason,
    affectedDraft,
    selectedEntity?.rawRow,
    loser?.rawRow,
    selectedEntity?.state,
    loser?.state,
    selectedEntity?.rowVersion,
    loser?.rowVersion,
    plan,
  ]);
  if (reviewedRef.current?.key !== reviewKey) reviewedRef.current = null;
  const latest = useRef({
    reviewKey,
    candidateId,
    plan,
    loser,
    selectedEntity,
    affectedDraft,
    reason,
    canMerge,
  });
  latest.current = {
    reviewKey,
    candidateId,
    plan,
    loser,
    selectedEntity,
    affectedDraft,
    reason,
    canMerge,
  };

  const invalidateReview = useCallback(() => {
    reviewedRef.current = null;
    renderReview((value) => value + 1);
  }, []);
  const discardDrafts = useCallback(() => {
    invalidateReview();
    const { selectedEntity, loser } = latest.current;
    if (selectedEntity !== null && loser !== null)
      discardAffectedDrafts([selectedEntity.recordId, loser.recordId]);
  }, [discardAffectedDrafts, invalidateReview]);
  const clearPlan = useCallback(() => {
    generationRef.current++;
    invalidateReview();
    setCandidateId("");
    setDetails([]);
  }, [invalidateReview]);
  const reset = useCallback(() => {
    clearPlan();
    setReasonState(defaultMergeReason);
    setFeedback(null);
  }, [clearPlan]);
  useEffect(
    () => () => {
      generationRef.current++;
      reviewedRef.current = null;
    },
    [],
  );
  const selectCandidate = useCallback(
    (id: string) => {
      generationRef.current++;
      invalidateReview();
      setCandidateId(id);
      setFeedback(null);
      setDetails([]);
    },
    [invalidateReview],
  );
  const setReason = useCallback(
    (value: string) => {
      invalidateReview();
      setReasonState(value);
    },
    [invalidateReview],
  );
  const start = useCallback(() => {
    setFeedback(
      workbookInspectorLocalErrorFeedback(
        "Select a loser to review the merge plan.",
      ),
    );
  }, []);
  const review = useCallback(() => {
    const current = latest.current;
    const authority = owner.getSnapshot();
    if (
      inFlight.current ||
      !current.canMerge ||
      !owner.canSubmit() ||
      authority.authority === null ||
      current.selectedEntity === null ||
      current.loser === null ||
      current.plan === null
    )
      return;
    if (current.affectedDraft) {
      setFeedback(
        workbookInspectorLocalErrorFeedback(
          "Finish or explicitly discard changes to both merge participants before reviewing.",
        ),
      );
      return;
    }
    if (!current.plan.valid || current.reason.trim() === "") {
      setFeedback(
        workbookInspectorLocalErrorFeedback(
          current.plan.issues[0] ?? "Provide a merge reason before reviewing.",
        ),
      );
      return;
    }
    const record = (row: EntityRow) =>
      Object.freeze({
        recordId: row.recordId,
        label: row.label,
        baseRowVersion: row.rowVersion,
      });
    const captured: EntityMergeReview = Object.freeze({
      survivor: record(current.selectedEntity),
      loser: record(current.loser),
      entityType: current.selectedEntity.entityType,
      authority: authority.authority,
      authorityGeneration: authority.generation,
      originSurface,
      presentationGeneration: generationRef.current,
      lifecycleKey: lifecycleResetKey,
      reason: current.reason,
      plan: current.plan,
    });
    reviewedRef.current = { key: current.reviewKey, review: captured };
    setFeedback(null);
    renderReview((value) => value + 1);
  }, [lifecycleResetKey, originSurface, owner]);
  const confirm = useCallback(async () => {
    const captured = reviewedRef.current;
    if (
      inFlight.current ||
      captured === null ||
      captured.key !== latest.current.reviewKey ||
      !latest.current.canMerge ||
      latest.current.affectedDraft ||
      !owner.canSubmit() ||
      owner.getSnapshot().generation !== captured.review.authorityGeneration
    ) {
      invalidateReview();
      return;
    }
    inFlight.current = true;
    invalidateReview();
    const reviewed = captured.review;
    const generation = generationRef.current;
    const currentPresentation = () =>
      generationRef.current === generation &&
      owner.getSnapshot().generation === reviewed.authorityGeneration;
    setFeedback(null);
    setDetails([]);
    const attempt = owner.admit(reviewed, {
      isCurrent: currentPresentation,
      matchesReview: () =>
        captured.key === latest.current.reviewKey &&
        !hasAffectedDraft([
          reviewed.survivor.recordId,
          reviewed.loser.recordId,
        ]),
      acknowledged: (receipt) => {
        if (currentPresentation())
          setFeedback(
            workbookInspectorMessageFeedback(
              `Merged ${reviewed.loser.label} (${receipt.loser_record_id}) into ${reviewed.survivor.label} (${receipt.survivor_record_id}). Change set ${receipt.change_set_id}.`,
              "polite",
            ),
          );
      },
      reconcile: async (receipt) => {
        if (!currentPresentation()) return;
        await loadSurvivorPreview(receipt.survivor_record_id);
        if (
          currentPresentation() &&
          latest.current.candidateId === reviewed.loser.recordId &&
          latest.current.reason === reviewed.reason
        )
          setCandidateId("");
      },
    });
    if (attempt === null) {
      inFlight.current = false;
      setFeedback(
        workbookInspectorLocalErrorFeedback(
          "This merge could not be admitted. Finish pending actions, refresh both records and review again.",
        ),
      );
      return;
    }
    try {
      await owner.execute(attempt);
      if (!currentPresentation()) return;
      const entry = owner
        .getSnapshot()
        .entries.find((entry) => entry.attempt.id === attempt.id);
      if (entry?.failure) {
        setFeedback(workbookInspectorOperationFailureFeedback(entry.failure));
        setDetails(
          entry.failure.kind === "validation"
            ? preconditionDetails(entry.failure.fields)
            : [],
        );
      } else if (entry?.phase === "uncertain") {
        setFeedback(
          workbookInspectorMessageFeedback(
            "The merge outcome is unknown. Open Merge actions to replay this exact request. Closing the inspector does not roll it back.",
            "polite",
          ),
        );
      }
    } finally {
      inFlight.current = false;
    }
  }, [hasAffectedDraft, invalidateReview, loadSurvivorPreview, owner]);

  return {
    commands: {
      clearPlan,
      confirm,
      reset,
      selectCandidate,
      setReason,
      start,
      review,
      invalidateReview,
      discardDrafts,
    },
    snapshot: {
      candidateId: authority.authority?.role ? candidateId : "",
      feedback: authority.authority?.role ? feedback : null,
      loser: authority.authority?.role ? loser : null,
      plan: authority.authority?.role ? plan : null,
      preconditionDetails: authority.authority?.role ? details : [],
      reason: authority.authority?.role ? reason : "",
      reviewed: reviewedRef.current?.review ?? null,
      hasAffectedDraft: affectedDraft,
    },
  };
}
