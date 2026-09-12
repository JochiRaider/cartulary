import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { WorkbookSurfaceRefreshError } from "../../collaboration/workbookSurfacePort";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import type { TimelineMentionCandidatePort } from "../actions/TimelineMentionCandidatePort";
import {
  initialMentionCreateDraft,
  type MentionCreateReview,
} from "../actions/timelineMentionCreationModel";
import {
  type MentionAction,
  type MentionBinding,
  type MentionSubject,
  sameMentionIntent,
} from "../actions/timelineMentionOperationModel";
import { useTimelineMentionCandidates } from "../actions/useTimelineMentionCandidates";
import type { WorkbookTimelineMentionOperationOwner } from "../actions/WorkbookTimelineMentionOperationOwner";
import type { TimelineCommittedRecordIdleResult } from "../models/timelineControllerPorts";
import { timelineMentionSubject } from "../models/timelineMentionActionPlan";
import type { WorkbookRow } from "../models/timelineRowModel";
import {
  type AutoResolutionNotice,
  buildInspectorMentions,
  type DismissedMention,
  type InspectorMention,
} from "../models/workbookMentionChips";

type Input = {
  readonly owner: WorkbookTimelineMentionOperationOwner;
  readonly candidatePort: TimelineMentionCandidatePort;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly earlierSaves: { readonly current: Promise<void> };
  readonly selectedMention: InspectorMention | null;
  readonly selectedTargetId: string;
  readonly setSelectedTargetId: (id: string) => void;
  readonly presentationKey: string;
  readonly presentationActive: boolean;
  readonly restoreActionFocus?: (sourceRecordId: string) => void;
  readonly refreshProjection?: () => Promise<void>;
  readonly waitForCommittedRecordIdle: (
    id: string,
    options: { signal: AbortSignal; refreshIfMissing: boolean },
  ) => Promise<TimelineCommittedRecordIdleResult | null>;
  readonly setInspectorMessage: (
    feedback: WorkbookInspectorFeedback | null,
  ) => void;
};
export function useTimelineMentionActions(input: Input) {
  const { owner } = input;
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const current = useRef(input);
  current.current = input;
  useLayoutEffect(
    () =>
      owner.registerPresentationRefresh(async (recordId, version) => {
        try {
          await current.current.refreshProjection?.();
        } catch (error) {
          // A socket-triggered query can supersede this read. A committed visible
          // source at the required version already satisfies the projection obligation.
          if (
            !(error instanceof WorkbookSurfaceRefreshError) ||
            error.recovery.kind !== "cancelled" ||
            (current.current.rowsRef.current.find(
              (row) => row.recordId === recordId,
            )?.rowVersion ?? 0) < version
          )
            throw error;
        }
      }),
    [owner],
  );
  const alive = useRef(true);
  const completionFocus = useRef<{
    key: number;
    creation: boolean;
    presentationKey: string;
    presentationActive: boolean;
    selectedMentionId: string | null;
    authorityGeneration: number;
    invoker: Element | null;
    sourceRecordId: string;
  } | null>(null);
  const rememberCompletionFocus = useCallback(
    (creation = false) => {
      const retained = owner.getSnapshot();
      const entry = creation
        ? retained.creations.at(-1)
        : retained.entries.at(-1);
      if (!entry) return;
      completionFocus.current = {
        key: entry.key,
        creation,
        presentationKey: current.current.presentationKey,
        presentationActive: current.current.presentationActive,
        selectedMentionId:
          current.current.selectedMention?.entityMentionId ?? null,
        authorityGeneration: retained.generation,
        invoker: document.activeElement,
        sourceRecordId: entry.attempt.review.subject.sourceRecordId,
      };
    },
    [owner],
  );
  useLayoutEffect(() => {
    const focus = completionFocus.current;
    if (!focus) return;
    if (
      focus.presentationKey !== input.presentationKey ||
      focus.presentationActive !== input.presentationActive ||
      focus.selectedMentionId !==
        (input.selectedMention?.entityMentionId ?? null) ||
      focus.authorityGeneration !== snapshot.generation
    ) {
      completionFocus.current = null;
      return;
    }
    const creation = focus.creation
      ? snapshot.creations.find((entry) => entry.key === focus.key)
      : null;
    if (creation && ["pending", "refreshing"].includes(creation.refresh))
      return;
    const key = creation ? creation.linkKey : focus.key;
    const entry = snapshot.entries.find((entry) => entry.key === key);
    if (entry?.refresh !== "complete") return;
    completionFocus.current = null;
    if (
      document.activeElement === focus.invoker ||
      document.activeElement === document.body
    )
      input.restoreActionFocus?.(focus.sourceRecordId);
  }, [
    input.presentationKey,
    input.presentationActive,
    input.selectedMention?.entityMentionId,
    input.restoreActionFocus,
    snapshot,
  ]);
  const subject = input.selectedMention
    ? timelineMentionSubject(
        input.selectedMention,
        input.rowsRef.current.find(
          (row) => row.recordId === input.selectedMention?.rowRecordId,
        ) ?? null,
        owner.incidentId,
      )
    : null;
  const presentationKey = `${input.presentationKey}:${input.presentationActive}:${subject?.mentionId ?? ""}:${subject?.mentionRowVersion ?? ""}`;
  const presentation = useRef({ key: presentationKey, generation: 0 });
  if (presentation.current.key !== presentationKey)
    presentation.current = {
      key: presentationKey,
      generation: presentation.current.generation + 1,
    };
  const [createReview, setCreateReview] = useState<MentionCreateReview | null>(
    null,
  );
  const draftRef = useRef(createReview);
  draftRef.current = createReview;
  const candidates = useTimelineMentionCandidates(
    input.candidatePort,
    subject?.entityType ?? "host",
    `${presentationKey}:${snapshot.generation}`,
    input.presentationActive &&
      !!snapshot.authority &&
      subject !== null &&
      subject.state !== "dismissed",
  );
  useEffect(() => {
    alive.current = true;
    return () => {
      alive.current = false;
    };
  }, []);
  useEffect(() => {
    if (
      presentation.current.key !== presentationKey ||
      owner.getSnapshot().generation !== snapshot.generation
    )
      return;
    setCreateReview(null);
    current.current.setSelectedTargetId("");
  }, [owner, presentationKey, snapshot.generation]);
  const binding = useCallback(
    (
      expected: MentionSubject,
      reviewedIntent: () => boolean,
      notice = false,
    ): MentionBinding => {
      const generation = presentation.current.generation;
      const origin = current.current.presentationKey;
      const earlier = current.current.earlierSaves.current;
      const isCurrent = () =>
        alive.current &&
        (notice || current.current.presentationActive) &&
        current.current.presentationKey === origin &&
        (notice || presentation.current.generation === generation) &&
        reviewedIntent();
      return {
        isCurrent,
        prepare: async (signal) => {
          await earlier;
          if (signal.aborted || !isCurrent()) return null;
          const idle = await current.current.waitForCommittedRecordIdle(
            expected.sourceRecordId,
            { signal, refreshIfMissing: false },
          );
          if (!idle?.row || signal.aborted || !isCurrent()) return null;
          return currentSubject(idle.row, expected.mentionId, owner);
        },
      };
    },
    [owner],
  );
  const feedback = (message: string) =>
    input.setInspectorMessage(
      workbookInspectorMessageFeedback(message, "none"),
    );
  function act(intent: MentionAction) {
    if (!subject || !snapshot.authority) {
      feedback("Mention identity is unavailable. Refresh before acting.");
      return;
    }
    const target =
      intent.action === "resolve_item"
        ? candidates.candidates.find(
            (candidate) => candidate.recordId === intent.resolvedRecordId,
          )
        : null;
    if (
      intent.action === "resolve_item" &&
      (!target || target.entityType !== subject.entityType)
    ) {
      feedback("Choose a loaded eligible target.");
      return;
    }
    const reviewedIntent = () =>
      intent.action !== "resolve_item" ||
      current.current.selectedTargetId === intent.resolvedRecordId;
    if (
      !owner.submit(
        { subject, intent, authority: snapshot.authority },
        binding(subject, reviewedIntent),
      )
    )
      feedback(
        "Review the current mention and recover any earlier operation before acting.",
      );
    else {
      rememberCompletionFocus();
      input.setInspectorMessage(null);
    }
  }
  function startCreate() {
    if (
      subject?.state !== "unresolved" ||
      !snapshot.authority ||
      !owner.canCreate(subject.entityType) ||
      owner.creationForMention(subject.mentionId)?.receipt
    )
      return;
    setCreateReview({
      subject,
      authority: snapshot.authority,
      draft: initialMentionCreateDraft(subject),
    });
  }
  function submitCreate() {
    const review = createReview;
    if (!review || !subject || !sameMentionIntent(review.subject, subject)) {
      feedback("Review the current mention before creating.");
      return;
    }
    const accepted = owner.createAndResolve(
      review,
      binding(review.subject, () => draftRef.current === review),
    );
    if (!accepted)
      feedback(
        "Review the entity fields and recover any earlier creation before submitting.",
      );
    else {
      rememberCompletionFocus(true);
      input.setInspectorMessage(null);
    }
  }
  const handleUndoAutoResolutionNotice = useCallback(
    (notice: AutoResolutionNotice) => {
      const row = current.current.rowsRef.current.find(
        (row) => row.recordId === notice.rowRecordId,
      );
      const subject = row
        ? currentSubject(row, notice.entityMentionId, owner)
        : null;
      const authority = owner.getSnapshot().authority;
      const matches = () => {
        const row = current.current.rowsRef.current.find(
          (row) => row.recordId === notice.rowRecordId,
        );
        const latest = row
          ? currentSubject(row, notice.entityMentionId, owner)
          : null;
        return (
          latest?.state === "resolved" &&
          latest.resolutionMethod === "auto_match" &&
          latest.mentionRowVersion === notice.mentionRowVersion &&
          latest.resolvedRecordId === notice.resolvedRecordId &&
          latest.itemRef === notice.itemRef
        );
      };
      if (!subject || !authority || !matches()) return;
      if (
        owner.submit(
          { subject, authority, intent: { action: "revert_to_unresolved" } },
          binding(subject, matches, true),
        )
      )
        rememberCompletionFocus();
    },
    [owner, binding, rememberCompletionFocus],
  );
  return {
    owner,
    snapshot,
    subject,
    candidates,
    createReview,
    selectedTargetId: input.selectedTargetId,
    changeTarget: input.setSelectedTargetId,
    act,
    startCreate,
    submitCreate,
    cancelCreate: () => setCreateReview(null),
    updateCreateDraft: (key: string, value: string) =>
      setCreateReview((review) =>
        review ? { ...review, draft: { ...review.draft, [key]: value } } : null,
      ),
    linkCreated: (key: number) => {
      if (
        subject &&
        owner.linkCreated(
          key,
          subject,
          binding(subject, () => true),
        )
      )
        rememberCompletionFocus();
    },
    handleUndoAutoResolutionNotice,
  };
}
export function currentSubject(
  row: WorkbookRow,
  id: string | null,
  owner: WorkbookTimelineMentionOperationOwner,
): MentionSubject | null {
  const observed = owner.getSnapshot().mentions;
  const dismissed: DismissedMention[] = observed
    .filter(
      (subject) =>
        subject.sourceRecordId === row.recordId &&
        subject.state === "dismissed",
    )
    .map((subject) => ({
      entityMentionId: subject.mentionId,
      rowRecordId: subject.sourceRecordId,
      fieldKey: subject.sourceFieldKey,
      entityType: subject.entityType,
      itemRef: subject.itemRef,
      rawText: subject.rawText,
      resolvedRecordId: null,
      mentionRowVersion: subject.mentionRowVersion,
      resolutionMethod: null,
      autoResolved: false,
    }));
  const mention = buildInspectorMentions(row, dismissed, observed).find(
    (mention) => mention.entityMentionId === id,
  );
  return mention
    ? timelineMentionSubject(mention, row, owner.incidentId)
    : null;
}
