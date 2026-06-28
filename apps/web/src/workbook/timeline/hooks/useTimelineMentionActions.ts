import { type Dispatch, type SetStateAction, useCallback } from "react";
import { apiPath } from "../../../services/browserApi";
import {
  fetchJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../../services/workbookApi";
import {
  beginTimelineEntityRefreshBarrier,
  type TimelineEntityCatalogInput,
  type TimelineEntityRefreshSettleState,
  type TimelineViewportContinuityBarrier,
  timelineEntityRefreshExpectationForMention,
  withTimelineEntityRefreshExpectation,
} from "../models/timelineViewportContinuityModel";
import type {
  DismissedMention,
  InspectorMention,
} from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import type { TimelineMutationEnvelope } from "../services/timelineMutationRequests";
import {
  buildMentionActionPayload,
  buildMentionPatchPayload,
  type MentionResolutionAction,
} from "../services/workbookCollaborationMessages";

type MentionActionEnvelope = {
  data: {
    incident_id: string;
    entity_mention: {
      entity_mention_id: string;
      source_record_id: string;
      source_field_key: string;
      entity_type: "host" | "identity" | string;
      raw_text: string;
      resolution_status: "unresolved" | "resolved" | "dismissed" | string;
      resolved_record_id: string | null;
      row_version: number;
      resolution_method: string | null;
    };
    source_record: {
      record_id: string;
      row_version: number;
    };
    change_set_id: string;
  };
};

type TimelineMentionViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type TimelineMentionLoadRowsOptions = {
  showLoading: boolean;
  freshnessRetryDepth?: number;
  viewportContinuityToken?: number;
};

export function useTimelineMentionActions({
  advanceViewportContinuity,
  apiBase,
  applyRowMutation,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  enqueueSaveWork,
  entityCatalogInput,
  finishSave,
  loadRows,
  nextClientTxnId,
  onRefreshEntities,
  resolvePendingSocketTxn,
  resolveViewportContinuityElement,
  rowsRef,
  setDismissedMentionsByRow,
  setInspectorMessage,
  settleViewportContinuityBarrier,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly advanceViewportContinuity: (
    token: number | undefined,
    options?: {
      barrier?: TimelineViewportContinuityBarrier;
      target?: TimelineMentionViewportContinuityTarget | null;
    },
  ) => void;
  readonly apiBase?: string | undefined;
  readonly applyRowMutation: (
    rowKey: string,
    envelope: TimelineMutationEnvelope,
    options?: {
      detectAutoResolution?: boolean;
      viewportContinuityToken?: number;
    },
  ) => WorkbookRow;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target: TimelineMentionViewportContinuityTarget,
    options?: { barrier?: TimelineViewportContinuityBarrier },
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly enqueueSaveWork: (work: () => Promise<void>) => void;
  readonly entityCatalogInput: TimelineEntityCatalogInput;
  readonly finishSave: (nextState: "Syncing" | "Saved" | "Conflict") => void;
  readonly loadRows: (options: TimelineMentionLoadRowsOptions) => Promise<void>;
  readonly nextClientTxnId: () => string;
  readonly onRefreshEntities?: (() => Promise<void> | void) | undefined;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly resolveViewportContinuityElement: (
    target: TimelineMentionViewportContinuityTarget,
  ) => HTMLElement | null;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly setDismissedMentionsByRow: Dispatch<
    SetStateAction<Record<string, DismissedMention[]>>
  >;
  readonly setInspectorMessage: (message: string | null) => void;
  readonly settleViewportContinuityBarrier: (
    token: number,
    refreshState: TimelineEntityRefreshSettleState,
  ) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<{ row: WorkbookRow | null; rowVersion: number } | null>;
}) {
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
      const requiresEntityRefresh =
        action === "resolve_item" && resolvedRecordId === undefined;
      const entityRefreshBarrier = requiresEntityRefresh
        ? beginTimelineEntityRefreshBarrier(entityCatalogInput)
        : null;
      const viewportContinuityToken = beginViewportContinuity(
        {
          kind: "row-inspect",
          recordId,
        },
        {
          barrier: entityRefreshBarrier,
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
        if (requiresEntityRefresh) {
          const payload = buildMentionPatchPayload(
            currentRow,
            mention,
            action,
            clientTxnId,
            resolvedRecordId,
          );
          if (payload === null) {
            clearViewportContinuity(viewportContinuityToken);
            finishSave("Conflict");
            return;
          }
          trackPendingSocketTxn(clientTxnId);
          const result = await fetchJSON<TimelineMutationEnvelope>(
            apiPath(apiBase, `/api/v1/records/${recordId}`),
            {
              method: "PATCH",
              body: JSON.stringify(payload),
            },
          );
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            clearViewportContinuity(viewportContinuityToken);
            setInspectorMessage(parseErrorMessage(result.payload));
            finishSave("Conflict");
            return;
          }

          const envelope = readEnvelope<TimelineMutationEnvelope>(
            result.payload,
          );
          const committed = applyRowMutation(currentRow.key, envelope, {
            detectAutoResolution: false,
            viewportContinuityToken,
          });
          const expectedEntity = timelineEntityRefreshExpectationForMention(
            committed,
            mention.itemRef,
          );
          if (expectedEntity !== null) {
            advanceViewportContinuity(viewportContinuityToken, {
              barrier: withTimelineEntityRefreshExpectation(
                entityRefreshBarrier,
                expectedEntity,
              ),
            });
          }
          finishSave("Saved");
          let refreshState: TimelineEntityRefreshSettleState = "complete";
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
            settleViewportContinuityBarrier(
              viewportContinuityToken,
              refreshState,
            );
          }
          return;
        }

        const activeItem = [
          ...currentRow.collectionValues.hostRefs,
          ...currentRow.collectionValues.identityRefs,
        ].find((item) => item.itemRef === mention.itemRef);
        const currentMention =
          activeItem === undefined
            ? mention
            : {
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
        const mentionID =
          currentMention.anchor.entityMentionId ??
          (currentMention.itemRef.startsWith("entity_mention:")
            ? currentMention.itemRef.slice("entity_mention:".length) || null
            : null);
        if (mentionID === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing entity mention identifier.");
          finishSave("Conflict");
          return;
        }
        const payload = buildMentionActionPayload(
          currentMention,
          action,
          clientTxnId,
          resolvedRecordId,
        );
        if (payload === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing mention row version.");
          finishSave("Conflict");
          return;
        }
        trackPendingSocketTxn(clientTxnId);
        const result = await fetchJSON<MentionActionEnvelope>(
          apiPath(apiBase, `/api/v1/entity-mentions/${mentionID}/resolve`),
          {
            method: "POST",
            body: JSON.stringify(payload),
          },
        );
        if (!result.ok) {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(parseErrorMessage(result.payload));
          finishSave("Conflict");
          return;
        }

        const envelope = readEnvelope<MentionActionEnvelope>(result.payload);
        const entityMention = envelope.data.entity_mention;
        if (action === "dismiss_item") {
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
                    entityMention.source_field_key === "timeline.identity_refs"
                      ? "timeline.identity_refs"
                      : currentMention.fieldKey,
                  entityType:
                    entityMention.entity_type === "identity"
                      ? "identity"
                      : currentMention.entityType,
                  itemRef: currentMention.itemRef,
                  rawText: entityMention.raw_text || currentMention.rawText,
                  resolvedRecordId: currentMention.resolvedRecordId,
                  mentionRowVersion: entityMention.row_version,
                  resolutionMethod:
                    entityMention.resolution_method ??
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
        if (action === "revert_to_unresolved") {
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

        await loadRows({
          showLoading: false,
          viewportContinuityToken,
        });
        const restoreMentionActionFocus = () => {
          resolveViewportContinuityElement({
            kind: "row-inspect",
            recordId,
          })?.focus({ preventScroll: true });
        };
        restoreMentionActionFocus();
        window.requestAnimationFrame(restoreMentionActionFocus);
        window.setTimeout(restoreMentionActionFocus, 0);
        finishSave("Saved");
      });
    },
    [
      advanceViewportContinuity,
      apiBase,
      applyRowMutation,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      entityCatalogInput,
      finishSave,
      loadRows,
      nextClientTxnId,
      onRefreshEntities,
      resolvePendingSocketTxn,
      resolveViewportContinuityElement,
      rowsRef,
      setDismissedMentionsByRow,
      setInspectorMessage,
      settleViewportContinuityBarrier,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  return { submitMentionAction };
}
