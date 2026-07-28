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
  buildMentionActionPayload,
  type MentionResolutionAction,
} from "../../runtime/workbookCollaborationMessages";
import type {
  TimelineContinuityRequirementName,
  TimelineSourceRecordRequirement,
} from "../models/timelineViewportContinuityModel";
import type {
  DismissedMention,
  InspectorMention,
} from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";

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
  afterProjectionCommit?: () => void;
  showLoading: boolean;
  freshnessRetryDepth?: number;
  sourceRecordRequirement?: TimelineSourceRecordRequirement;
  viewportContinuityToken?: number;
};

export function useTimelineMentionActions({
  apiBase,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  enqueueSaveWork,
  finishSave,
  incidentId,
  loadRows,
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
  readonly apiBase?: string | undefined;
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
  readonly incidentId: string;
  readonly loadRows: (options: TimelineMentionLoadRowsOptions) => Promise<void>;
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
  readonly setInspectorMessage: (message: string | null) => void;
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

        let sourceRecordRequirement: TimelineSourceRecordRequirement;
        try {
          const envelope = readEnvelope<MentionActionEnvelope>(result.payload);
          sourceRecordRequirement = requireMentionActionSourceRecord(
            envelope,
            recordId,
          );
        } catch {
          resolvePendingSocketTxn(resolveClientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Mention action source record was invalid.");
          finishSave("Conflict");
          return;
        }
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
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      finishSave,
      incidentId,
      loadRows,
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

        let envelope: MentionActionEnvelope;
        let sourceRecordRequirement: TimelineSourceRecordRequirement;
        try {
          envelope = readEnvelope<MentionActionEnvelope>(result.payload);
          sourceRecordRequirement = requireMentionActionSourceRecord(
            envelope,
            recordId,
          );
        } catch {
          resolvePendingSocketTxn(clientTxnId);
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage("Mention action source record was invalid.");
          finishSave("Conflict");
          return;
        }
        requireViewportContinuitySourceRecord(
          viewportContinuityToken,
          sourceRecordRequirement,
        );
        const entityMention = envelope.data.entity_mention;
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
                          entityMention.source_field_key ===
                          "timeline.identity_refs"
                            ? "timeline.identity_refs"
                            : currentMention.fieldKey,
                        entityType:
                          entityMention.entity_type === "identity"
                            ? "identity"
                            : currentMention.entityType,
                        itemRef: currentMention.itemRef,
                        rawText:
                          entityMention.raw_text || currentMention.rawText,
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
      apiBase,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      finishSave,
      loadRows,
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

  return { createEntityFromMention, submitMentionAction };
}

function requireMentionActionSourceRecord(
  envelope: MentionActionEnvelope,
  expectedRecordId: string,
): TimelineSourceRecordRequirement {
  const sourceRecord = envelope.data.source_record;
  if (
    sourceRecord.record_id !== expectedRecordId ||
    !Number.isSafeInteger(sourceRecord.row_version) ||
    sourceRecord.row_version < 1
  ) {
    throw new Error(
      `Mention action source record ${sourceRecord.record_id}@${sourceRecord.row_version} does not match ${expectedRecordId}.`,
    );
  }
  return {
    recordId: sourceRecord.record_id,
    minimumRowVersion: sourceRecord.row_version,
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
