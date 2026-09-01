import { type Dispatch, type SetStateAction, useCallback } from "react";
import type { MentionResolutionAction } from "../../collaboration/workbookCollaborationMessages";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
} from "../../inspector/workbookInspectorErrorModel";
import type {
  TimelineContinuityRequirementName,
  TimelineSourceRecordRequirement,
} from "../models/timelineViewportContinuityModel";
import type {
  AutoResolutionNotice,
  DismissedMention,
  InspectorMention,
} from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import type { TimelineMentionPort } from "../ports/TimelineMentionPort";

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

export function useTimelineMentionActions({
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  enqueueSaveWork,
  finishSave,
  loadRows,
  mentionPort,
  nextClientTxnId,
  onRefreshEntities,
  requireViewportContinuitySourceRecord,
  resolvePendingSocketTxn,
  rowsRef,
  setDismissedMentionsByRow,
  setInspectorMessage,
  settleViewportContinuityFollowUp,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
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
  readonly loadRows: (options: TimelineMentionLoadRowsOptions) => Promise<void>;
  readonly mentionPort: TimelineMentionPort;
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
}) {
  const createEntityFromMention = useCallback(
    (mention: InspectorMention) => {
      const snapshot = rowsRef.current.find(
        (candidate) => candidate.recordId === mention.rowRecordId,
      );
      if (!snapshot || snapshot.recordId === null) {
        return;
      }

      const recordId = snapshot.recordId;
      const createClientTxnId = nextClientTxnId();
      const resolveClientTxnId = nextClientTxnId();
      const viewportContinuityToken = beginViewportContinuity(
        {
          kind: "row-inspect",
          recordId,
        },
        {
          requirements: ["entity-refresh"],
        },
      );
      beginSave();
      setInspectorMessage(null);
      enqueueSaveWork(async () => {
        const idleRecord = await waitForCommittedRecordIdle(recordId);
        if (idleRecord === null || idleRecord.row === null) {
          clearViewportContinuity(viewportContinuityToken);
          finishSave("Conflict");
          return;
        }
        const currentRow = idleRecord.row;
        const createResult = await mentionPort.createEntity({
          clientTxnId: createClientTxnId,
          entityType: mention.entityType,
          rawText: mention.rawText,
        });
        if (createResult.kind === "rejected") {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(
            workbookInspectorErrorPresentation(createResult.failure),
          );
          finishSave("Conflict");
          return;
        }
        const createdRecordId = createResult.value.recordId;
        const currentMention = currentMentionFromRow(currentRow, mention);
        const mentionID = entityMentionID(currentMention);
        if (mentionID === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing entity mention identifier.");
          finishSave("Conflict");
          return;
        }
        if (currentMention.mentionRowVersion === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing mention row version.");
          finishSave("Conflict");
          return;
        }

        trackPendingSocketTxn(resolveClientTxnId);
        const result = await mentionPort.resolve({
          action: "resolve_item",
          baseMentionRowVersion: currentMention.mentionRowVersion,
          clientTxnId: resolveClientTxnId,
          expectedSourceRecordId: recordId,
          mentionId: mentionID,
          resolvedRecordId: createdRecordId,
        });
        if (result.kind === "rejected") {
          resolvePendingSocketTxn(resolveClientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(
            workbookInspectorErrorPresentation(result.failure),
          );
          finishSave("Conflict");
          return;
        }

        const sourceRecordRequirement: TimelineSourceRecordRequirement = {
          recordId: result.value.sourceRecord.recordId,
          minimumRowVersion: result.value.sourceRecord.rowVersion,
        };
        requireViewportContinuitySourceRecord(
          viewportContinuityToken,
          sourceRecordRequirement,
        );
        let projectionCommitted = false;
        await loadRows({
          afterProjectionCommit: () => {
            projectionCommitted = true;
            finishSave("Saved");
          },
          showLoading: false,
          sourceRecordRequirement,
          viewportContinuityToken,
        });
        if (!projectionCommitted) {
          finishSave("Conflict");
        }

        let refreshState: "settled" | "terminal" = "settled";
        try {
          if (onRefreshEntities === undefined) {
            refreshState = "terminal";
          } else {
            await onRefreshEntities();
          }
        } catch (error) {
          refreshState = "terminal";
          throw error;
        } finally {
          settleViewportContinuityFollowUp(
            viewportContinuityToken,
            "entity-refresh",
            refreshState,
          );
        }
      });
    },
    [
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      finishSave,
      loadRows,
      mentionPort,
      nextClientTxnId,
      onRefreshEntities,
      requireViewportContinuitySourceRecord,
      resolvePendingSocketTxn,
      rowsRef,
      setInspectorMessage,
      settleViewportContinuityFollowUp,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const submitMentionAction = useCallback(
    (
      mention: InspectorMention,
      action: MentionResolutionAction,
      resolvedRecordId?: string,
    ) => {
      const snapshot = rowsRef.current.find(
        (candidate) => candidate.recordId === mention.rowRecordId,
      );
      if (!snapshot || snapshot.recordId === null) {
        return;
      }

      const recordId = snapshot.recordId;
      const clientTxnId = nextClientTxnId();
      if (action === "resolve_item" && resolvedRecordId === undefined) {
        setInspectorMessage("Select a target first.");
        return;
      }
      const viewportContinuityToken = beginViewportContinuity({
        kind: "row-inspect",
        recordId,
      });
      beginSave();
      setInspectorMessage(null);
      enqueueSaveWork(async () => {
        const idleRecord = await waitForCommittedRecordIdle(recordId);
        if (idleRecord === null || idleRecord.row === null) {
          clearViewportContinuity(viewportContinuityToken);
          finishSave("Conflict");
          return;
        }
        const currentRow = idleRecord.row;
        const currentMention = currentMentionFromRow(currentRow, mention);
        const mentionID = entityMentionID(currentMention);
        if (mentionID === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing entity mention identifier.");
          finishSave("Conflict");
          return;
        }
        if (currentMention.mentionRowVersion === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing mention row version.");
          finishSave("Conflict");
          return;
        }
        trackPendingSocketTxn(clientTxnId);
        const result = await mentionPort.resolve({
          action,
          baseMentionRowVersion: currentMention.mentionRowVersion,
          clientTxnId,
          expectedSourceRecordId: recordId,
          mentionId: mentionID,
          ...(resolvedRecordId === undefined ? {} : { resolvedRecordId }),
        });
        if (result.kind === "rejected") {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(
            workbookInspectorErrorPresentation(result.failure),
          );
          finishSave("Conflict");
          return;
        }

        const sourceRecordRequirement: TimelineSourceRecordRequirement = {
          recordId: result.value.sourceRecord.recordId,
          minimumRowVersion: result.value.sourceRecord.rowVersion,
        };
        requireViewportContinuitySourceRecord(
          viewportContinuityToken,
          sourceRecordRequirement,
        );
        const entityMention = result.value.entityMention;
        const applyMentionFollowUp =
          action === "dismiss_item"
            ? () => {
                setDismissedMentionsByRow((current) => {
                  const rowMentions = current[recordId] ?? [];
                  return {
                    ...current,
                    [recordId]: [
                      ...rowMentions.filter(
                        (item) => item.itemRef !== currentMention.itemRef,
                      ),
                      {
                        rowRecordId: recordId,
                        fieldKey:
                          entityMention.sourceFieldKey ===
                          "timeline.identity_refs"
                            ? "timeline.identity_refs"
                            : currentMention.fieldKey,
                        entityType:
                          entityMention.entityType === "identity"
                            ? "identity"
                            : currentMention.entityType,
                        itemRef: currentMention.itemRef,
                        rawText:
                          entityMention.rawText ?? currentMention.rawText,
                        resolvedRecordId: currentMention.resolvedRecordId,
                        mentionRowVersion: entityMention.rowVersion,
                        resolutionMethod:
                          entityMention.resolutionMethod ??
                          currentMention.resolutionMethod,
                        autoResolved: currentMention.autoResolved,
                        displayText: currentMention.displayText,
                        priorTargetEntityRecordId:
                          currentMention.anchor.targetEntityRecordId ??
                          currentMention.priorTargetEntityRecordId ??
                          currentMention.resolvedRecordId,
                        provenance: currentMention.provenance,
                        confidence: currentMention.confidence,
                        matchedAliasText: currentMention.matchedAliasText,
                      },
                    ],
                  };
                });
              }
            : action === "revert_to_unresolved"
              ? () => {
                  setDismissedMentionsByRow((current) => {
                    const rowMentions = (current[recordId] ?? []).filter(
                      (item) => item.itemRef !== currentMention.itemRef,
                    );
                    if (rowMentions.length < 1) {
                      const next = { ...current };
                      delete next[recordId];
                      return next;
                    }
                    return {
                      ...current,
                      [recordId]: rowMentions,
                    };
                  });
                }
              : undefined;

        let projectionCommitted = false;
        await loadRows({
          afterProjectionCommit: () => {
            applyMentionFollowUp?.();
            projectionCommitted = true;
            finishSave("Saved");
          },
          showLoading: false,
          sourceRecordRequirement,
          viewportContinuityToken,
        });
        if (!projectionCommitted) {
          finishSave("Conflict");
        }
      });
    },
    [
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      finishSave,
      loadRows,
      mentionPort,
      nextClientTxnId,
      requireViewportContinuitySourceRecord,
      resolvePendingSocketTxn,
      rowsRef,
      setDismissedMentionsByRow,
      setInspectorMessage,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const handleUndoAutoResolutionNotice = useCallback(
    (notice: AutoResolutionNotice) => {
      const mention = timelineMentionForAutoResolutionNotice(
        rowsRef.current,
        notice,
      );
      if (mention !== null) {
        submitMentionAction(mention, "revert_to_unresolved");
      }
    },
    [rowsRef, submitMentionAction],
  );

  return {
    createEntityFromMention,
    handleUndoAutoResolutionNotice,
    submitMentionAction,
  };
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
    rowRecordId: row.recordId,
    fieldKey: notice.fieldKey,
    entityType: activeItem.entityType,
    itemRef: activeItem.itemRef,
    rawText: activeItem.rawText,
    resolvedRecordId: activeItem.resolvedRecordId,
    mentionRowVersion: activeItem.mentionRowVersion,
    resolutionMethod: activeItem.resolutionMethod,
    autoResolved: activeItem.autoResolved,
    status: "resolved",
    chipState: activeItem.autoResolved ? "auto_resolved" : "resolved",
    anchor: {
      recordId: row.recordId,
      fieldKey: notice.fieldKey,
      itemRef: activeItem.itemRef,
      entityMentionId: entityMentionIDFromItemRef(activeItem.itemRef),
      targetEntityRecordId: activeItem.resolvedRecordId,
    },
    sourceKind: "entity_mention",
    isActiveRelationshipValue: true,
    priorTargetEntityRecordId: null,
    displayText: activeItem.displayText,
    provenance: activeItem.provenance,
    confidence: activeItem.confidence,
    matchedAliasText: activeItem.matchedAliasText,
  };
}

function currentMentionFromRow(
  row: WorkbookRow,
  mention: InspectorMention,
): InspectorMention {
  const activeItem = [
    ...row.collectionValues.hostRefs,
    ...row.collectionValues.identityRefs,
  ].find((item) => item.itemRef === mention.itemRef);
  if (activeItem === undefined) {
    return mention;
  }
  return {
    ...mention,
    entityType: activeItem.entityType,
    rawText: activeItem.rawText,
    resolvedRecordId: activeItem.resolvedRecordId,
    mentionRowVersion: activeItem.mentionRowVersion,
    resolutionMethod: activeItem.resolutionMethod,
    autoResolved: activeItem.autoResolved,
    displayText: activeItem.displayText,
    provenance: activeItem.provenance,
    confidence: activeItem.confidence,
    matchedAliasText: activeItem.matchedAliasText,
    anchor: {
      ...mention.anchor,
      targetEntityRecordId: activeItem.resolvedRecordId,
    },
  };
}

function entityMentionID(mention: InspectorMention): string | null {
  return (
    mention.anchor.entityMentionId ??
    entityMentionIDFromItemRef(mention.itemRef)
  );
}

function entityMentionIDFromItemRef(itemRef: string): string | null {
  return itemRef.startsWith("entity_mention:")
    ? itemRef.slice("entity_mention:".length) || null
    : null;
}
