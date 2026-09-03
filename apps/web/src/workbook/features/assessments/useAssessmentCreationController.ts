import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorLocalErrorFeedback,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type {
  AssessmentCreateDraft,
  AssessmentSubjectType,
} from "../../models/assessmentWorkbookModel";
import {
  followOnAssessmentDraft,
  initialAssessmentDraft,
} from "../../models/assessmentWorkbookModel";
import { assessmentsViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { AssessmentMutationCommandPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

type AssessmentCreationController = {
  readonly commands: {
    readonly cancel: () => void;
    readonly openFollowOn: (selected: WorkbookQueryRow | null) => boolean;
    readonly openStandalone: (defaultSubjectRecordId: string) => void;
    readonly rejectFailure: (failure: WorkbookOperationFailure) => void;
    readonly rejectStart: (message: string) => void;
    readonly reset: () => void;
    readonly submit: (canCreate: boolean) => Promise<void>;
    readonly updateDraft: (
      update: (current: AssessmentCreateDraft) => AssessmentCreateDraft,
    ) => void;
  };
  readonly snapshot: {
    readonly draft: AssessmentCreateDraft;
    readonly draftMode: "follow_on" | "standalone";
    readonly isSubmitting: boolean;
    readonly feedback: WorkbookInspectorFeedback | null;
  };
};

export function useAssessmentCreationController({
  beginMutation,
  lifecycleResetKey,
  mutationCommands,
  onRefreshAssessmentRows,
  subjectRecordIds,
}: {
  readonly beginMutation: () => () => void;
  readonly lifecycleResetKey: string;
  readonly mutationCommands: AssessmentMutationCommandPort;
  readonly onRefreshAssessmentRows: () => Promise<void>;
  readonly subjectRecordIds: Readonly<
    Record<AssessmentSubjectType, readonly string[]>
  >;
}): AssessmentCreationController {
  const [draft, setDraft] = useState<AssessmentCreateDraft>(() =>
    initialAssessmentDraft(assessmentsContract),
  );
  const [draftMode, setDraftMode] = useState<"follow_on" | "standalone">(
    "standalone",
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [feedback, setFeedback] = useState<WorkbookInspectorFeedback | null>(
    null,
  );
  const generationRef = useRef(0);
  const previousLifecycleKeyRef = useRef(lifecycleResetKey);

  const resetState = useCallback(() => {
    setDraft(initialAssessmentDraft(assessmentsContract));
    setDraftMode("standalone");
    setIsSubmitting(false);
    setFeedback(null);
  }, []);

  const reset = useCallback(() => {
    generationRef.current += 1;
    resetState();
  }, [resetState]);

  useEffect(() => {
    if (previousLifecycleKeyRef.current === lifecycleResetKey) return;
    previousLifecycleKeyRef.current = lifecycleResetKey;
    generationRef.current += 1;
    resetState();
  }, [lifecycleResetKey, resetState]);

  useEffect(
    () => () => {
      generationRef.current += 1;
    },
    [],
  );

  useEffect(() => {
    if (draftMode === "follow_on") return;
    setDraft((current) => {
      const available = subjectRecordIds[current.subjectType];
      if (
        current.subjectRecordId !== "" &&
        available.includes(current.subjectRecordId)
      ) {
        return current;
      }
      return { ...current, subjectRecordId: available[0] ?? "" };
    });
  }, [draftMode, subjectRecordIds]);

  const openStandalone = useCallback((defaultSubjectRecordId: string) => {
    setDraft(
      initialAssessmentDraft(assessmentsContract, {
        subjectRecordId: defaultSubjectRecordId,
        subjectType: "host",
      }),
    );
    setDraftMode("standalone");
    setFeedback(null);
  }, []);

  const openFollowOn = useCallback((selected: WorkbookQueryRow | null) => {
    if (selected === null) {
      setFeedback(
        workbookInspectorLocalErrorFeedback(
          "Select an assessment before creating a follow-on.",
        ),
      );
      return false;
    }
    const followOnDraft = followOnAssessmentDraft(
      assessmentsContract,
      selected,
    );
    if (followOnDraft === null) {
      setFeedback(
        workbookInspectorLocalErrorFeedback(
          "The selected assessment has no valid subject.",
        ),
      );
      return false;
    }
    setDraft(followOnDraft);
    setDraftMode("follow_on");
    setFeedback(null);
    return true;
  }, []);

  const rejectStart = useCallback((nextMessage: string) => {
    setFeedback(workbookInspectorLocalErrorFeedback(nextMessage));
  }, []);

  const rejectFailure = useCallback((failure: WorkbookOperationFailure) => {
    setFeedback(workbookInspectorOperationFailureFeedback(failure));
  }, []);

  const updateDraft = useCallback(
    (update: (current: AssessmentCreateDraft) => AssessmentCreateDraft) => {
      setDraft(update);
    },
    [],
  );

  const cancel = useCallback(() => {
    generationRef.current += 1;
    resetState();
  }, [resetState]);

  const submit = useCallback(
    async (canCreate: boolean) => {
      if (!canCreate) {
        setFeedback(
          workbookInspectorLocalErrorFeedback(
            "Assessment creation requires an active editor role.",
          ),
        );
        return;
      }
      if (!mutationCommands.canCreate({ draft })) {
        setFeedback(
          workbookInspectorLocalErrorFeedback(
            "Complete the required assessment fields.",
          ),
        );
        return;
      }
      const generation = generationRef.current;
      const submittedDraft = draft;
      setIsSubmitting(true);
      setFeedback(null);
      const finishMutation = beginMutation();
      try {
        const outcome = await mutationCommands.create({
          draft: submittedDraft,
        });
        if (generationRef.current !== generation) return;
        if (outcome.kind === "rejected") {
          setFeedback(
            workbookInspectorOperationFailureFeedback(outcome.failure),
          );
          return;
        }
        await onRefreshAssessmentRows();
        if (generationRef.current !== generation) return;
        setDraft(
          initialAssessmentDraft(assessmentsContract, {
            subjectType: submittedDraft.subjectType,
            subjectRecordId: submittedDraft.subjectRecordId,
          }),
        );
        setFeedback(
          workbookInspectorMessageFeedback("Assessment created.", "polite"),
        );
      } finally {
        finishMutation();
        if (generationRef.current === generation) setIsSubmitting(false);
      }
    },
    [beginMutation, draft, mutationCommands, onRefreshAssessmentRows],
  );

  return {
    commands: {
      cancel,
      openFollowOn,
      openStandalone,
      rejectFailure,
      rejectStart,
      reset,
      submit,
      updateDraft,
    },
    snapshot: { draft, draftMode, feedback, isSubmitting },
  };
}
