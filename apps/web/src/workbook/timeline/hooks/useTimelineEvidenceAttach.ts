import { useCallback } from "react";
import { apiPath } from "../../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../../services/workbookApi";
import {
  createAndAttachEvidenceBlob,
  evidenceAttachPublicErrorMessage,
  evidencePublicErrorMessage,
} from "../../../services/workbookEvidence";
import {
  evidenceViewSchemaId,
  timelineViewSchemaId,
} from "../../models/workbookSurfaceRegistry";
import {
  buildAttachedEvidenceCreatePayload,
  buildAttachedEvidencePatchPayload,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type { TimelineMutationEnvelope } from "../services/timelineMutationRequests";

type EvidenceRowCreateEnvelope = {
  data: {
    row: {
      record_id: string;
      row_version: number;
    };
  };
};

type TimelineEvidenceViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

function evidenceTitleFromFile(file: File): string {
  return file.name.trim() || "Workbook attachment";
}

export function useTimelineEvidenceAttach({
  apiBase,
  applyRowMutation,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  enqueueSaveWork,
  finishSave,
  incidentId,
  nextClientTxnId,
  resolvePendingSocketTxn,
  rowsRef,
  setInspectorMessage,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly apiBase?: string | undefined;
  readonly applyRowMutation: (
    rowKey: string,
    envelope: TimelineMutationEnvelope,
    options?: {
      continueOnFreshDraft?: boolean;
      promoteToCommittedRowInspect?: boolean;
      detectAutoResolution?: boolean;
      viewportContinuityToken?: number;
    },
  ) => WorkbookRow;
  readonly beginSave: () => void;
  readonly beginViewportContinuity: (
    target: TimelineEvidenceViewportContinuityTarget,
  ) => number;
  readonly clearViewportContinuity: (token: number) => void;
  readonly enqueueSaveWork: (work: () => Promise<void>) => void;
  readonly finishSave: (nextState: "Syncing" | "Saved" | "Conflict") => void;
  readonly incidentId: string;
  readonly nextClientTxnId: () => string;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly setInspectorMessage: (message: string | null) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<{ row: WorkbookRow | null; rowVersion: number } | null>;
}) {
  const createAndAttachEvidenceFile = useCallback(
    async (file: File): Promise<string> => {
      const createEvidence = await fetchWorkbookJSON<EvidenceRowCreateEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${evidenceViewSchemaId}/rows`,
        ),
        {
          method: "POST",
          body: JSON.stringify({
            client_txn_id: nextClientTxnId(),
            "evidence.title": evidenceTitleFromFile(file),
            "evidence.collector_party_text": "Workbook upload",
          }),
        },
      );
      if (!createEvidence.ok) {
        throw new Error(evidencePublicErrorMessage(createEvidence.payload));
      }
      const evidenceEnvelope = readEnvelope<EvidenceRowCreateEnvelope>(
        createEvidence.payload,
      );
      const evidenceRecord = evidenceEnvelope.data.row;

      await createAndAttachEvidenceBlob({
        apiBase,
        attachClientTxnId: nextClientTxnId,
        baseRowVersion: evidenceRecord.row_version,
        createClientTxnId: nextClientTxnId,
        evidenceRecordId: evidenceRecord.record_id,
        file,
        incidentId,
      });
      return evidenceRecord.record_id;
    },
    [apiBase, incidentId, nextClientTxnId],
  );

  const attachEvidenceFileToTimeline = useCallback(
    (target: WorkbookRow, file: File) => {
      const snapshot =
        rowsRef.current.find((candidate) => candidate.key === target.key) ??
        target;
      const viewportContinuityToken = beginViewportContinuity(
        snapshot.recordId === null
          ? { kind: "scroll-only" }
          : { kind: "row-inspect", recordId: snapshot.recordId },
      );
      beginSave();
      setInspectorMessage("Uploading evidence.");

      enqueueSaveWork(async () => {
        try {
          const effectiveSnapshot =
            snapshot.recordId === null
              ? snapshot
              : (await waitForCommittedRecordIdle(snapshot.recordId))?.row;
          if (effectiveSnapshot === null || effectiveSnapshot === undefined) {
            throw new Error("invalid_timeline_row");
          }
          const evidenceRecordId = await createAndAttachEvidenceFile(file);
          const clientTxnId = nextClientTxnId();
          const payload =
            effectiveSnapshot.recordId === null
              ? buildAttachedEvidenceCreatePayload(
                  evidenceRecordId,
                  clientTxnId,
                )
              : buildAttachedEvidencePatchPayload(
                  effectiveSnapshot,
                  evidenceRecordId,
                  clientTxnId,
                );
          if (payload === null) {
            throw new Error("invalid_timeline_row");
          }

          trackPendingSocketTxn(clientTxnId);
          const targetPath =
            effectiveSnapshot.recordId === null
              ? apiPath(
                  apiBase,
                  `/api/v1/incidents/${incidentId}/views/${timelineViewSchemaId}/rows`,
                )
              : apiPath(
                  apiBase,
                  `/api/v1/records/${effectiveSnapshot.recordId}`,
                );
          const result = await fetchWorkbookJSON<TimelineMutationEnvelope>(
            targetPath,
            {
              method: effectiveSnapshot.recordId === null ? "POST" : "PATCH",
              body: JSON.stringify(payload),
            },
          );
          if (!result.ok) {
            resolvePendingSocketTxn(clientTxnId);
            throw new Error(parseErrorMessage(result.payload));
          }

          const envelope = readEnvelope<TimelineMutationEnvelope>(
            result.payload,
          );
          applyRowMutation(effectiveSnapshot.key, envelope, {
            continueOnFreshDraft: effectiveSnapshot.recordId === null,
            promoteToCommittedRowInspect: effectiveSnapshot.recordId === null,
            detectAutoResolution: false,
            viewportContinuityToken,
          });
          setInspectorMessage("Evidence attached.");
          finishSave("Saved");
        } catch (error) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(evidenceAttachPublicErrorMessage(error));
          finishSave("Conflict");
        }
      });
    },
    [
      apiBase,
      applyRowMutation,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      createAndAttachEvidenceFile,
      enqueueSaveWork,
      finishSave,
      incidentId,
      nextClientTxnId,
      resolvePendingSocketTxn,
      rowsRef,
      setInspectorMessage,
      trackPendingSocketTxn,
      waitForCommittedRecordIdle,
    ],
  );

  const handleTimelineEvidenceFiles = useCallback(
    (target: WorkbookRow, files: FileList | File[]) => {
      const [file] = Array.from(files);
      if (!file) {
        return;
      }
      attachEvidenceFileToTimeline(target, file);
    },
    [attachEvidenceFileToTimeline],
  );

  return {
    attachEvidenceFileToTimeline,
    handleTimelineEvidenceFiles,
  };
}
