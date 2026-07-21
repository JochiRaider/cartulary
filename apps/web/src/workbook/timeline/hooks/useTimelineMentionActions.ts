import { type Dispatch, type SetStateAction, useCallback } from "react";
import { apiPath } from "../../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../../services/workbookApi";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import {
  beginTimelineEntityRefreshBarrier,
  type TimelineEntityCatalogInput,
  type TimelineEntityRefreshSettleState,
  type TimelineViewportContinuityBarrier,
} from "../models/timelineViewportContinuityModel";
import type {
  DismissedMention,
  InspectorMention,
} from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import {
  buildMentionActionPayload,
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

type EntityCreateEnvelope = {
  data: {
    row: {
      record_id: string;
    };
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
  apiBase,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  enqueueSaveWork,
  entityCatalogInput,
  finishSave,
  incidentId,
  loadRows,
  nextClientTxnId,
  onRefreshEntities,
  resolvePendingSocketTxn,
  rowsRef,
  setDismissedMentionsByRow,
  setInspectorMessage,
  settleViewportContinuityBarrier,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly apiBase?: string | undefined;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target: TimelineMentionViewportContinuityTarget,
    options?: { barrier?: TimelineViewportContinuityBarrier },
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly enqueueSaveWork: (work: () => Promise<void>) => void;
  readonly entityCatalogInput: TimelineEntityCatalogInput;
  readonly finishSave: (nextState: "Syncing" | "Saved" | "Conflict") => void;
  readonly incidentId: string;
  readonly loadRows: (options: TimelineMentionLoadRowsOptions) => Promise<void>;
  readonly nextClientTxnId: () => string;
  readonly onRefreshEntities?: (() => Promise<void> | void) | undefined;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
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
      const entityRefreshBarrier =
        beginTimelineEntityRefreshBarrier(entityCatalogInput);
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
        const createIntent = buildMentionEntityCreateIntent(
          mention,
          createClientTxnId,
        );
        if (createIntent === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Cannot create an entity from this mention.");
          finishSave("Conflict");
          return;
        }

        const createResult = await fetchWorkbookJSON<EntityCreateEnvelope>(
          apiPath(
            apiBase,
            `/api/v1/incidents/${incidentId}/views/${createIntent.viewSchemaId}/rows`,
          ),
          {
            method: "POST",
            body: JSON.stringify(createIntent.payload),
          },
        );
        if (!createResult.ok) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(parseErrorMessage(createResult.payload));
          finishSave("Conflict");
          return;
        }
        const createEnvelope = readEnvelope<EntityCreateEnvelope>(
          createResult.payload,
        );
        const createdRecordId = createEnvelope.data.row.record_id;
        const currentMention = currentMentionFromRow(currentRow, mention);
        const mentionID = entityMentionID(currentMention);
        if (mentionID === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing entity mention identifier.");
          finishSave("Conflict");
          return;
        }
        const payload = buildMentionActionPayload(
          currentMention,
          "resolve_item",
          resolveClientTxnId,
          createdRecordId,
        );
        if (payload === null) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Missing mention row version.");
          finishSave("Conflict");
          return;
        }

        trackPendingSocketTxn(resolveClientTxnId);
        const result = await fetchWorkbookJSON<MentionActionEnvelope>(
          apiPath(apiBase, `/api/v1/entity-mentions/${mentionID}/resolve`),
          {
            method: "POST",
            body: JSON.stringify(payload),
          },
        );
        if (!result.ok) {
          resolvePendingSocketTxn(resolveClientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(parseErrorMessage(result.payload));
          finishSave("Conflict");
          return;
        }

        await loadRows({
          showLoading: false,
          viewportContinuityToken,
        });
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
      });
    },
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      entityCatalogInput,
      finishSave,
      incidentId,
      loadRows,
      nextClientTxnId,
      onRefreshEntities,
      resolvePendingSocketTxn,
      rowsRef,
      setInspectorMessage,
      settleViewportContinuityBarrier,
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
        const result = await fetchWorkbookJSON<MentionActionEnvelope>(
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
        finishSave("Saved");
      });
    },
    [
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      finishSave,
      loadRows,
      nextClientTxnId,
      resolvePendingSocketTxn,
      rowsRef,
      setDismissedMentionsByRow,
      setInspectorMessage,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  return { createEntityFromMention, submitMentionAction };
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
    (mention.itemRef.startsWith("entity_mention:")
      ? mention.itemRef.slice("entity_mention:".length) || null
      : null)
  );
}

function buildMentionEntityCreateIntent(
  mention: InspectorMention,
  clientTxnId: string,
): {
  readonly payload: Record<string, unknown>;
  readonly viewSchemaId: string;
} | null {
  const rawText = mention.rawText.trim();
  if (rawText === "") {
    return null;
  }
  if (mention.entityType === "host") {
    const payload: Record<string, unknown> = {
      client_txn_id: clientTxnId,
      "host.display_name": rawText,
    };
    if (rawText.includes(".")) {
      payload["host.fqdn"] = rawText;
    } else {
      payload["host.hostname"] = rawText;
    }
    return { payload, viewSchemaId: hostsViewSchemaId };
  }
  if (mention.entityType === "identity") {
    const payload: Record<string, unknown> = {
      client_txn_id: clientTxnId,
      "identity.display_name": rawText,
    };
    if (rawText.includes("@")) {
      payload["identity.upn"] = rawText;
      payload["identity.email"] = rawText;
    } else {
      payload["identity.sam_account_name"] = rawText;
    }
    return { payload, viewSchemaId: identitiesViewSchemaId };
  }
  return null;
}
