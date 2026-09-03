import { type Dispatch, type SetStateAction, useCallback, useRef } from "react";
import type { MentionResolutionAction } from "../../collaboration/workbookCollaborationMessages";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  planTimelineMentionEntityCreation,
  planTimelineMentionResolution,
  type TimelineMentionActionContext,
  type TimelineMentionResolutionPlan,
  type TimelineMentionSubject,
  timelineMentionSubject,
} from "../models/timelineMentionActionPlan";
import type { WorkbookRow } from "../models/timelineRowModel";
import type {
  TimelineContinuityRequirementName,
  TimelineSourceRecordRequirement,
} from "../models/timelineViewportContinuityModel";
import type {
  AutoResolutionNotice,
  DismissedMention,
  InspectorMention,
} from "../models/workbookMentionChips";
import type { TimelineMentionPorts } from "../ports/TimelineMentionPort";

type TimelineMentionViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type TimelineMentionLoadRowsOptions = {
  afterProjectionCommit?: () => void;
  showLoading: boolean;
  freshnessRetryDepth?: number;
  sourceRecordRequirement?: TimelineSourceRecordRequirement;
  viewportContinuityToken?: number;
};

type TimelineMentionActionsInput = {
  readonly actionContext: TimelineMentionActionContext;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target: TimelineMentionViewportContinuityTarget,
    options?: {
      requirements?: readonly TimelineContinuityRequirementName[];
    },
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly enqueueSaveWork: (work: () => Promise<void>) => void;
  readonly finishSave: (nextState: "Syncing" | "Saved" | "Conflict") => void;
  readonly knownEntityTypes: ReadonlyMap<string, "host" | "identity">;
  readonly loadRows: (options: TimelineMentionLoadRowsOptions) => Promise<void>;
  readonly mentionPorts: TimelineMentionPorts;
  readonly nextClientTxnId: () => string;
  readonly onRefreshEntities?: (() => Promise<void> | void) | undefined;
  readonly requireViewportContinuitySourceRecord: (
    token: number,
    requirement: TimelineSourceRecordRequirement,
  ) => void;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly setDismissedMentionsByRow: Dispatch<
    SetStateAction<Record<string, DismissedMention[]>>
  >;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly settleViewportContinuityFollowUp: (
    token: number,
    requirement: TimelineContinuityRequirementName,
    state: "settled" | "terminal",
  ) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<{ row: WorkbookRow | null; rowVersion: number } | null>;
};

type TimelineMentionAccepted = Extract<
  Awaited<ReturnType<TimelineMentionPorts["resolution"]["resolve"]>>,
  { readonly kind: "accepted" }
>["value"];

export function useTimelineMentionActions(input: TimelineMentionActionsInput) {
  const inputRef = useRef(input);
  inputRef.current = input;

  const createEntityFromMention = useCallback((mention: InspectorMention) => {
    const current = inputRef.current;
    const subject = timelineMentionSubject(
      mention,
      current.actionContext.surfaceKey,
    );
    const plan = planTimelineMentionEntityCreation({
      context: current.actionContext,
      mention,
      rows: current.rowsRef.current,
      subject,
    });
    if (plan.kind === "reject") {
      publishMentionPlanRejection(current, plan.reason);
      return;
    }
    const token = current.beginViewportContinuity(
      { kind: "row-inspect", recordId: subject.rowRecordId },
      { requirements: ["entity-refresh"] },
    );
    current.beginSave();
    current.setInspectorMessage(null);
    current.enqueueSaveWork(() =>
      executeMentionEntityCreation({ inputRef, mention, subject, token }),
    );
  }, []);

  const submitMentionAction = useCallback(
    (
      mention: InspectorMention,
      action: MentionResolutionAction,
      resolvedRecordId?: string,
    ) => {
      const current = inputRef.current;
      const subject = timelineMentionSubject(
        mention,
        current.actionContext.surfaceKey,
      );
      const plan = planTimelineMentionResolution({
        action,
        context: current.actionContext,
        knownEntityTypes: current.knownEntityTypes,
        mention,
        resolvedRecordId,
        rows: current.rowsRef.current,
        subject,
      });
      if (plan.kind === "reject") {
        publishMentionPlanRejection(current, plan.reason);
        return;
      }
      const token = current.beginViewportContinuity({
        kind: "row-inspect",
        recordId: subject.rowRecordId,
      });
      current.beginSave();
      current.setInspectorMessage(null);
      current.enqueueSaveWork(() =>
        executeMentionResolution({
          action,
          inputRef,
          mention,
          resolvedRecordId,
          subject,
          token,
        }),
      );
    },
    [],
  );

  const handleUndoAutoResolutionNotice = useCallback(
    (notice: AutoResolutionNotice) => {
      const mention = timelineMentionForAutoResolutionNotice(
        inputRef.current.rowsRef.current,
        notice,
      );
      if (mention !== null) {
        submitMentionAction(mention, "revert_to_unresolved");
      }
    },
    [submitMentionAction],
  );

  return {
    createEntityFromMention,
    handleUndoAutoResolutionNotice,
    submitMentionAction,
  };
}

async function executeMentionEntityCreation(options: {
  readonly inputRef: { readonly current: TimelineMentionActionsInput };
  readonly mention: InspectorMention;
  readonly subject: TimelineMentionSubject;
  readonly token: number;
}): Promise<void> {
  const currentPlan = await currentMentionPlan({
    action: "resolve_item",
    forEntityCreation: true,
    inputRef: options.inputRef,
    mention: options.mention,
    subject: options.subject,
  });
  if (currentPlan.kind === "reject") {
    failMentionPlan(
      options.inputRef.current,
      options.token,
      currentPlan.reason,
    );
    return;
  }
  const current = options.inputRef.current;
  const createResult = await current.mentionPorts.entityCreation.createEntity({
    clientTxnId: current.nextClientTxnId(),
    entityType: currentPlan.mention.entityType,
    rawText: currentPlan.mention.rawText,
  });
  if (createResult.kind === "rejected") {
    failMentionOperation(current, options.token, createResult.failure);
    return;
  }
  const resolutionPlan = await currentMentionPlan({
    action: "resolve_item",
    allowUnknownResolvedTarget: true,
    inputRef: options.inputRef,
    mention: options.mention,
    resolvedRecordId: createResult.value.recordId,
    subject: options.subject,
  });
  if (resolutionPlan.kind === "reject") {
    failMentionPlan(
      options.inputRef.current,
      options.token,
      resolutionPlan.reason,
    );
    return;
  }
  const accepted = await dispatchMentionResolution(
    options.inputRef,
    resolutionPlan,
    options.subject,
  );
  if (accepted === null) {
    options.inputRef.current.clearViewportContinuity(options.token);
    options.inputRef.current.finishSave("Conflict");
    return;
  }
  await settleMentionProjection(
    options.inputRef.current,
    accepted,
    options.token,
  );
  await settleEntityRefresh(options.inputRef.current, options.token);
}

async function executeMentionResolution(options: {
  readonly action: MentionResolutionAction;
  readonly inputRef: { readonly current: TimelineMentionActionsInput };
  readonly mention: InspectorMention;
  readonly resolvedRecordId?: string | undefined;
  readonly subject: TimelineMentionSubject;
  readonly token: number;
}): Promise<void> {
  const plan = await currentMentionPlan({
    action: options.action,
    inputRef: options.inputRef,
    mention: options.mention,
    resolvedRecordId: options.resolvedRecordId,
    subject: options.subject,
  });
  if (plan.kind === "reject") {
    failMentionPlan(options.inputRef.current, options.token, plan.reason);
    return;
  }
  const accepted = await dispatchMentionResolution(
    options.inputRef,
    plan,
    options.subject,
  );
  if (accepted === null) {
    options.inputRef.current.clearViewportContinuity(options.token);
    options.inputRef.current.finishSave("Conflict");
    return;
  }
  await settleMentionProjection(
    options.inputRef.current,
    accepted,
    options.token,
    mentionFollowUp(
      options.inputRef.current,
      options.action,
      plan.mention,
      accepted,
    ),
  );
}

async function currentMentionPlan(options: {
  readonly action: MentionResolutionAction;
  readonly allowUnknownResolvedTarget?: boolean;
  readonly forEntityCreation?: boolean;
  readonly inputRef: { readonly current: TimelineMentionActionsInput };
  readonly mention: InspectorMention;
  readonly resolvedRecordId?: string | undefined;
  readonly subject: TimelineMentionSubject;
}): Promise<TimelineMentionResolutionPlan> {
  const idle = await options.inputRef.current.waitForCommittedRecordIdle(
    options.subject.rowRecordId,
  );
  const current = options.inputRef.current;
  const rows = idle?.row === null || idle === null ? [] : [idle.row];
  return options.forEntityCreation
    ? planTimelineMentionEntityCreation({
        context: current.actionContext,
        mention: options.mention,
        rows,
        subject: options.subject,
      })
    : planTimelineMentionResolution({
        action: options.action,
        ...(options.allowUnknownResolvedTarget === undefined
          ? {}
          : {
              allowUnknownResolvedTarget: options.allowUnknownResolvedTarget,
            }),
        context: current.actionContext,
        knownEntityTypes: current.knownEntityTypes,
        mention: options.mention,
        resolvedRecordId: options.resolvedRecordId,
        rows,
        subject: options.subject,
      });
}

async function dispatchMentionResolution(
  inputRef: { readonly current: TimelineMentionActionsInput },
  plan: Extract<TimelineMentionResolutionPlan, { kind: "dispatch" }>,
  subject: TimelineMentionSubject,
) {
  const current = inputRef.current;
  const clientTxnId = current.nextClientTxnId();
  current.trackPendingSocketTxn(clientTxnId);
  const result = await current.mentionPorts.resolution.resolve({
    ...plan.request,
    clientTxnId,
  });
  if (result.kind === "accepted") {
    const latest = inputRef.current;
    const settlementPlan = planTimelineMentionResolution({
      action: plan.request.action,
      allowUnknownResolvedTarget: true,
      context: latest.actionContext,
      knownEntityTypes: latest.knownEntityTypes,
      mention: plan.mention,
      resolvedRecordId: plan.request.resolvedRecordId,
      rows: latest.rowsRef.current,
      subject,
    });
    if (settlementPlan.kind === "dispatch") return result.value;
    latest.resolvePendingSocketTxn(clientTxnId);
    return null;
  }
  inputRef.current.resolvePendingSocketTxn(clientTxnId);
  inputRef.current.setInspectorMessage(
    workbookInspectorOperationFailureFeedback(result.failure),
  );
  return null;
}

async function settleMentionProjection(
  input: TimelineMentionActionsInput,
  accepted: TimelineMentionAccepted,
  token: number,
  followUp?: (() => void) | undefined,
): Promise<void> {
  const requirement: TimelineSourceRecordRequirement = {
    recordId: accepted.sourceRecord.recordId,
    minimumRowVersion: accepted.sourceRecord.rowVersion,
  };
  input.requireViewportContinuitySourceRecord(token, requirement);
  let projectionCommitted = false;
  await input.loadRows({
    afterProjectionCommit: () => {
      followUp?.();
      projectionCommitted = true;
      input.finishSave("Saved");
    },
    showLoading: false,
    sourceRecordRequirement: requirement,
    viewportContinuityToken: token,
  });
  if (!projectionCommitted) input.finishSave("Conflict");
}

async function settleEntityRefresh(
  input: TimelineMentionActionsInput,
  token: number,
): Promise<void> {
  let state: "settled" | "terminal" = "settled";
  try {
    if (input.onRefreshEntities === undefined) state = "terminal";
    else await input.onRefreshEntities();
  } catch (error) {
    state = "terminal";
    throw error;
  } finally {
    input.settleViewportContinuityFollowUp(token, "entity-refresh", state);
  }
}

function mentionFollowUp(
  input: TimelineMentionActionsInput,
  action: MentionResolutionAction,
  mention: InspectorMention,
  accepted: TimelineMentionAccepted,
): (() => void) | undefined {
  if (action === "dismiss_item") {
    return () => {
      input.setDismissedMentionsByRow((current) => ({
        ...current,
        [mention.rowRecordId]: [
          ...(current[mention.rowRecordId] ?? []).filter(
            (item) => item.itemRef !== mention.itemRef,
          ),
          dismissedMention(mention, accepted),
        ],
      }));
    };
  }
  if (action !== "revert_to_unresolved") return undefined;
  return () => {
    input.setDismissedMentionsByRow((current) =>
      withoutDismissedMention(current, mention),
    );
  };
}

function dismissedMention(
  mention: InspectorMention,
  accepted: TimelineMentionAccepted,
): DismissedMention {
  const entityMention = accepted.entityMention;
  return {
    autoResolved: mention.autoResolved,
    confidence: mention.confidence,
    displayText: mention.displayText,
    entityType:
      entityMention.entityType === "identity" ? "identity" : mention.entityType,
    fieldKey:
      entityMention.sourceFieldKey === "timeline.identity_refs"
        ? "timeline.identity_refs"
        : mention.fieldKey,
    itemRef: mention.itemRef,
    matchedAliasText: mention.matchedAliasText,
    mentionRowVersion: entityMention.rowVersion,
    priorTargetEntityRecordId:
      mention.anchor.targetEntityRecordId ??
      mention.priorTargetEntityRecordId ??
      mention.resolvedRecordId,
    provenance: mention.provenance,
    rawText: entityMention.rawText ?? mention.rawText,
    resolutionMethod:
      entityMention.resolutionMethod ?? mention.resolutionMethod,
    resolvedRecordId: mention.resolvedRecordId,
    rowRecordId: mention.rowRecordId,
  };
}

function withoutDismissedMention(
  current: Record<string, DismissedMention[]>,
  mention: InspectorMention,
): Record<string, DismissedMention[]> {
  const retained = (current[mention.rowRecordId] ?? []).filter(
    (item) => item.itemRef !== mention.itemRef,
  );
  if (retained.length > 0) {
    return { ...current, [mention.rowRecordId]: retained };
  }
  const next = { ...current };
  delete next[mention.rowRecordId];
  return next;
}

function failMentionPlan(
  input: TimelineMentionActionsInput,
  token: number,
  reason: Extract<TimelineMentionResolutionPlan, { kind: "reject" }>["reason"],
): void {
  input.clearViewportContinuity(token);
  publishMentionPlanRejection(input, reason);
  input.finishSave("Conflict");
}

function publishMentionPlanRejection(
  input: TimelineMentionActionsInput,
  reason: Extract<TimelineMentionResolutionPlan, { kind: "reject" }>["reason"],
): void {
  const message =
    reason === "target_missing"
      ? "Select a target first."
      : reason === "mention_version_missing"
        ? "Missing mention row version."
        : reason === "mention_missing"
          ? "Missing entity mention identifier."
          : "The selected Timeline mention is no longer available.";
  input.setInspectorMessage(workbookInspectorMessageFeedback(message, "none"));
}

function failMentionOperation(
  input: TimelineMentionActionsInput,
  token: number,
  failure: Parameters<typeof workbookInspectorOperationFailureFeedback>[0],
): void {
  input.clearViewportContinuity(token);
  input.setInspectorMessage(workbookInspectorOperationFailureFeedback(failure));
  input.finishSave("Conflict");
}

export function timelineMentionForAutoResolutionNotice(
  rows: readonly WorkbookRow[],
  notice: AutoResolutionNotice,
): InspectorMention | null {
  const row = rows.find(
    (candidate) => candidate.recordId === notice.rowRecordId,
  );
  if (row?.recordId === null || row === undefined) return null;
  const activeItems =
    notice.fieldKey === "timeline.identity_refs"
      ? row.collectionValues.identityRefs
      : row.collectionValues.hostRefs;
  const activeItem = activeItems.find(
    (item) =>
      item.itemRef === notice.itemRef && item.itemKind === "resolved_ref",
  );
  if (activeItem === undefined) return null;
  return {
    anchor: {
      entityMentionId: entityMentionIdFromItemRef(activeItem.itemRef),
      fieldKey: notice.fieldKey,
      itemRef: activeItem.itemRef,
      recordId: row.recordId,
      targetEntityRecordId: activeItem.resolvedRecordId,
    },
    autoResolved: activeItem.autoResolved,
    chipState: activeItem.autoResolved ? "auto_resolved" : "resolved",
    confidence: activeItem.confidence,
    displayText: activeItem.displayText,
    entityType: activeItem.entityType,
    fieldKey: notice.fieldKey,
    isActiveRelationshipValue: true,
    itemRef: activeItem.itemRef,
    matchedAliasText: activeItem.matchedAliasText,
    mentionRowVersion: activeItem.mentionRowVersion,
    priorTargetEntityRecordId: null,
    provenance: activeItem.provenance,
    rawText: activeItem.rawText,
    resolutionMethod: activeItem.resolutionMethod,
    resolvedRecordId: activeItem.resolvedRecordId,
    rowRecordId: row.recordId,
    sourceKind: "entity_mention",
    status: "resolved",
  };
}

function entityMentionIdFromItemRef(itemRef: string): string | null {
  return itemRef.startsWith("entity_mention:")
    ? itemRef.slice("entity_mention:".length) || null
    : null;
}
