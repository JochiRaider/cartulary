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
  targetWorkbookInspectorFeedback,
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
import type {
  AutoResolutionDisclosure,
  WorkbookTimelineMentionOperationOwner,
} from "../actions/WorkbookTimelineMentionOperationOwner";
import type {
  DisclosureReviewNavigationScope,
  TimelineCommittedRecordIdleResult,
} from "../models/timelineControllerPorts";
import { timelineMentionSubject } from "../models/timelineMentionActionPlan";
import type { WorkbookRow } from "../models/timelineRowModel";
import {
  type AutoResolutionNotice,
  buildInspectorMentions,
  type DisclosureReviewFeedback,
  type DismissedMention,
  disclosureReviewKey,
  type InspectorMention,
} from "../models/workbookMentionChips";

type Input = {
  readonly currentCommittedRow?: (recordId: string) => WorkbookRow | null;
  readonly acceptDisclosureSource?: (row: WorkbookRow) => {
    readonly accepted: boolean;
    readonly row: WorkbookRow;
    readonly stale: boolean;
  };
  readonly owner: WorkbookTimelineMentionOperationOwner;
  readonly candidatePort: TimelineMentionCandidatePort;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly earlierSaves: { readonly current: Promise<void> };
  readonly selectedMention: InspectorMention | null;
  readonly selectedMentionRef: string | null;
  readonly selectedRowId: string | null;
  readonly inspectorInvalidationGeneration: number;
  readonly reviewSurfaceKey: string;
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

type DisclosureReviewRequest = {
  readonly id: number;
  readonly key: string;
  readonly notice: AutoResolutionDisclosure;
  readonly controller: AbortController;
  readonly authorityGeneration: number;
  readonly surfaceKey: string;
  readonly presentationKey: string;
  readonly presentationActive: boolean;
  readonly selectedRowId: string | null;
  readonly selectedMentionRef: string | null;
  readonly invalidationGeneration: number;
  phase: "reading" | "navigating";
  ownedFocus: boolean;
  cancelProgress: (() => void) | null;
  cancelNavigationDeadline: (() => void) | null;
};
export function useTimelineMentionActions(input: Input) {
  const { owner } = input;
  const snapshot = useSyncExternalStore(
    owner.subscribe,
    owner.getActionSnapshot,
  );
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
    notice: boolean;
    presentationKey: string;
    presentationActive: boolean;
    selectedMentionId: string | null;
    authorityGeneration: number;
    invoker: Element | null;
    sourceRecordId: string;
  } | null>(null);
  const rememberCompletionFocus = useCallback(
    (creation = false, notice = false) => {
      const retained = owner.getSnapshot();
      const entry = creation
        ? retained.creations.at(-1)
        : retained.entries.at(-1);
      if (!entry) return;
      completionFocus.current = {
        key: entry.key,
        creation,
        notice,
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
    if (entry?.refresh !== "complete" && !(focus.notice && entry?.receipt))
      return;
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
        input.currentCommittedRow?.(input.selectedMention.rowRecordId) ??
          input.rowsRef.current.find(
            (row) => row.recordId === input.selectedMention?.rowRecordId,
          ) ??
          null,
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
  const reviewRequest = useRef<DisclosureReviewRequest | null>(null);
  const reviewSequence = useRef(0);
  const [reviewFeedback, setReviewFeedback] =
    useState<DisclosureReviewFeedback | null>(null);
  const isReviewCurrent = useCallback(
    (request: DisclosureReviewRequest) => {
      const live = current.current;
      if (
        !alive.current ||
        reviewRequest.current !== request ||
        reviewRequest.current.id !== request.id ||
        request.controller.signal.aborted ||
        request.authorityGeneration !== owner.getSnapshot().generation ||
        request.surfaceKey !== live.reviewSurfaceKey ||
        !owner
          .getDisclosureSnapshot()
          .some((notice) => disclosureReviewKey(notice) === request.key)
      )
        return false;
      if (request.phase === "reading")
        return (
          live.presentationKey === request.presentationKey &&
          live.presentationActive === request.presentationActive &&
          live.selectedRowId === request.selectedRowId &&
          live.selectedMentionRef === request.selectedMentionRef &&
          live.inspectorInvalidationGeneration ===
            request.invalidationGeneration
        );
      return (
        live.selectedRowId === request.notice.rowRecordId &&
        live.selectedMentionRef === request.notice.itemRef
      );
    },
    [owner],
  );
  const cancelDisclosureReview = useCallback((deferFeedback = false) => {
    const request = reviewRequest.current;
    if (!request) return;
    reviewRequest.current = null;
    request.controller.abort();
    request.cancelProgress?.();
    request.cancelNavigationDeadline?.();
    const clear = () => {
      if (alive.current && reviewRequest.current === null)
        setReviewFeedback((feedback) =>
          feedback?.key === request.key ? null : feedback,
        );
    };
    if (deferFeedback) setTimeout(clear, 0);
    else clear();
  }, []);
  useEffect(() => {
    const interrupt = (event: Event) => {
      const request = reviewRequest.current;
      if (!request || request.ownedFocus) return;
      const control =
        event.target instanceof Element
          ? event.target.closest("[data-disclosure-review-key]")
          : null;
      if (
        control?.getAttribute("data-disclosure-review-key") === request.key &&
        (event.type === "pointerdown" ||
          event.type === "focusin" ||
          (event instanceof KeyboardEvent &&
            (event.key === "Enter" || event.key === " ")))
      )
        return;
      // Fence the request now; publish notice state after native focus/input work.
      cancelDisclosureReview(true);
    };
    const nativeInput = () => cancelDisclosureReview(true);
    for (const type of [
      "pointerdown",
      "keydown",
      "wheel",
      "focusin",
      "compositionstart",
    ])
      document.addEventListener(type, interrupt, true);
    document.addEventListener("input", nativeInput);
    return () => {
      for (const type of [
        "pointerdown",
        "keydown",
        "wheel",
        "focusin",
        "compositionstart",
      ])
        document.removeEventListener(type, interrupt, true);
      document.removeEventListener("input", nativeInput);
      const request = reviewRequest.current;
      reviewRequest.current = null;
      request?.controller.abort();
      request?.cancelProgress?.();
      request?.cancelNavigationDeadline?.();
    };
  }, [cancelDisclosureReview]);
  useEffect(() => {
    const request = reviewRequest.current;
    const contextChanged =
      request !== null &&
      (request.surfaceKey !== input.reviewSurfaceKey ||
        request.authorityGeneration !== snapshot.generation ||
        (request.phase === "reading" &&
          (request.presentationKey !== input.presentationKey ||
            request.presentationActive !== input.presentationActive ||
            request.selectedRowId !== input.selectedRowId ||
            request.selectedMentionRef !== input.selectedMentionRef ||
            request.invalidationGeneration !==
              input.inspectorInvalidationGeneration)));
    if (request && (contextChanged || !isReviewCurrent(request)))
      cancelDisclosureReview();
    if (
      reviewFeedback &&
      !owner
        .getDisclosureSnapshot()
        .some((notice) => disclosureReviewKey(notice) === reviewFeedback.key)
    )
      setReviewFeedback(null);
  }, [
    cancelDisclosureReview,
    input.inspectorInvalidationGeneration,
    input.presentationActive,
    input.presentationKey,
    input.reviewSurfaceKey,
    input.selectedMentionRef,
    input.selectedRowId,
    isReviewCurrent,
    owner,
    reviewFeedback,
    snapshot,
  ]);
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
          if (!idle || signal.aborted || !isCurrent()) return null;
          const row =
            idle.row ??
            (notice
              ? await owner.readSource(expected.sourceRecordId, signal)
              : null);
          if (!row || signal.aborted || !isCurrent()) return null;
          return currentSubject(row, expected.mentionId, owner);
        },
      };
    },
    [owner],
  );
  const feedback = (message: string) =>
    input.setInspectorMessage(
      input.selectedMention
        ? targetWorkbookInspectorFeedback(
            workbookInspectorMessageFeedback(message, "none"),
            input.selectedMention.rowRecordId,
            {
              kind: "relationship_item",
              panel: "relationships",
              fieldKey: input.selectedMention.fieldKey,
              itemRef: input.selectedMention.itemRef,
            },
          )
        : {
            ...workbookInspectorMessageFeedback(message, "none"),
            destination: { kind: "panel", panel: "relationships" },
          },
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
      const row =
        current.current.currentCommittedRow?.(notice.rowRecordId) ??
        current.current.rowsRef.current.find(
          (row) => row.recordId === notice.rowRecordId,
        );
      const subject = row
        ? currentSubject(row, notice.entityMentionId, owner)
        : owner.latestMention(notice.entityMentionId ?? "");
      const authority = owner.getSnapshot().authority;
      const matches = () => {
        const row =
          current.current.currentCommittedRow?.(notice.rowRecordId) ??
          current.current.rowsRef.current.find(
            (row) => row.recordId === notice.rowRecordId,
          );
        const latest = row
          ? currentSubject(row, notice.entityMentionId, owner)
          : owner.latestMention(notice.entityMentionId ?? "");
        return (
          latest?.sourceRecordId === notice.rowRecordId &&
          latest.sourceFieldKey === notice.fieldKey &&
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
        rememberCompletionFocus(false, true);
    },
    [owner, binding, rememberCompletionFocus],
  );
  const startDisclosureReview = useCallback(
    (
      notice: AutoResolutionDisclosure,
      navigate: (
        recordId: string,
        itemRef: string,
        scope: DisclosureReviewNavigationScope,
      ) => void,
    ) => {
      const key = disclosureReviewKey(notice);
      if (
        reviewRequest.current?.key === key &&
        isReviewCurrent(reviewRequest.current)
      )
        return;
      cancelDisclosureReview();
      if (
        !owner
          .getDisclosureSnapshot()
          .some((item) => disclosureReviewKey(item) === key)
      )
        return;
      const live = current.current;
      const request: DisclosureReviewRequest = {
        id: ++reviewSequence.current,
        key,
        notice,
        controller: new AbortController(),
        authorityGeneration: owner.getSnapshot().generation,
        surfaceKey: live.reviewSurfaceKey,
        presentationKey: live.presentationKey,
        presentationActive: live.presentationActive,
        selectedRowId: live.selectedRowId,
        selectedMentionRef: live.selectedMentionRef,
        invalidationGeneration: live.inspectorInvalidationGeneration,
        phase: "reading",
        ownedFocus: false,
        cancelProgress: null,
        cancelNavigationDeadline: null,
      };
      reviewRequest.current = request;
      setReviewFeedback(null);
      const progress = setTimeout(() => {
        if (isReviewCurrent(request))
          setReviewFeedback({ key, phase: "pending" });
      }, 120);
      request.cancelProgress = () => clearTimeout(progress);
      const fail = () => {
        if (!isReviewCurrent(request)) return;
        reviewRequest.current = null;
        request.controller.abort();
        request.cancelProgress?.();
        request.cancelNavigationDeadline?.();
        setReviewFeedback({
          key,
          phase: "failure",
          message: "Could not open the source. Review again to retry.",
        });
      };
      const run = async () => {
        try {
          const row =
            current.current.rowsRef.current.find(
              (candidate) => candidate.recordId === notice.rowRecordId,
            ) ??
            (await owner.readSource(
              notice.rowRecordId,
              request.controller.signal,
            ));
          if (!isReviewCurrent(request)) return;
          const subject = currentSubject(row, notice.entityMentionId, owner);
          if (
            !subject ||
            subject.sourceRecordId !== notice.rowRecordId ||
            subject.sourceFieldKey !== notice.fieldKey ||
            subject.itemRef !== notice.itemRef ||
            subject.mentionRowVersion !== notice.mentionRowVersion ||
            subject.resolvedRecordId !== notice.resolvedRecordId ||
            subject.state !== "resolved" ||
            subject.resolutionMethod !== "auto_match"
          ) {
            fail();
            return;
          }
          const sourceAcceptance =
            current.current.acceptDisclosureSource?.(row);
          if (sourceAcceptance && !sourceAcceptance.accepted) {
            fail();
            return;
          }
          if (!isReviewCurrent(request)) return;
          const accepted =
            current.current.currentCommittedRow?.(notice.rowRecordId) ??
            sourceAcceptance?.row ??
            row;
          const currentMention = currentSubject(
            accepted,
            notice.entityMentionId,
            owner,
          );
          if (!currentMention || !sameMentionIntent(subject, currentMention)) {
            fail();
            return;
          }
          request.phase = "navigating";
          const deadline = setTimeout(fail, 5000);
          request.cancelNavigationDeadline = () => clearTimeout(deadline);
          navigate(notice.rowRecordId, notice.itemRef, {
            signal: request.controller.signal,
            isCurrent: () => isReviewCurrent(request),
            runOwnedFocus: (focus) => {
              if (!isReviewCurrent(request)) return false;
              request.ownedFocus = true;
              try {
                return focus();
              } finally {
                request.ownedFocus = false;
              }
            },
            settle: (focused) => {
              if (!isReviewCurrent(request)) return;
              if (!focused) {
                fail();
                return;
              }
              reviewRequest.current = null;
              request.cancelProgress?.();
              request.cancelNavigationDeadline?.();
              request.controller.abort();
              setReviewFeedback(null);
            },
          });
        } catch {
          fail();
        }
      };
      void run();
    },
    [cancelDisclosureReview, isReviewCurrent, owner],
  );
  return {
    startDisclosureReview,
    reviewFeedback,
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
function currentSubject(
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
