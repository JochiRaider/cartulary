import { useCallback } from "react";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import type { TimelineEvidenceAttachmentPort } from "../ports/TimelineEvidenceAttachmentPort";

type TimelineEvidenceViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

export function useTimelineEvidenceAttach({
  applyAcceptedRowMutation,
  beginSave,
  beginViewportContinuity,
  clearViewportContinuity,
  enqueueSaveWork,
  evidenceAttachmentPort,
  finishSave,
  resolvePendingSocketTxn,
  rowsRef,
  setInspectorMessage,
  trackPendingSocketTxn,
  waitForCommittedRecordIdle,
}: {
  readonly applyAcceptedRowMutation: (
    rowKey: string,
    mutation: {
      readonly row: import("../models/workbookTimelineModel").TimelineApiRow;
      readonly viewSchemaId: string;
    },
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
  readonly evidenceAttachmentPort: TimelineEvidenceAttachmentPort;
  readonly finishSave: (nextState: "Syncing" | "Saved" | "Conflict") => void;
  readonly resolvePendingSocketTxn: (clientTxnId: string) => void;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly setInspectorMessage: (message: string | null) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<{ row: WorkbookRow | null; rowVersion: number } | null>;
}) {
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
        const effectiveSnapshot =
          snapshot.recordId === null
            ? snapshot
            : (await waitForCommittedRecordIdle(snapshot.recordId))?.row;
        if (effectiveSnapshot === null || effectiveSnapshot === undefined) {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(
            "The selected Timeline row is no longer available.",
          );
          finishSave("Conflict");
          return;
        }
        const result = await evidenceAttachmentPort.attach({
          file,
          onTimelineClientTxnId: trackPendingSocketTxn,
          target: effectiveSnapshot,
        });
        if (result.clientTxnId !== null) {
          resolvePendingSocketTxn(result.clientTxnId);
        }
        if (result.outcome.kind === "rejected") {
          clearViewportContinuity(viewportContinuityToken);
          setInspectorMessage(result.outcome.failure.message);
          finishSave("Conflict");
          return;
        }
        applyAcceptedRowMutation(effectiveSnapshot.key, result.outcome.value, {
          continueOnFreshDraft: effectiveSnapshot.recordId === null,
          promoteToCommittedRowInspect: effectiveSnapshot.recordId === null,
          detectAutoResolution: false,
          viewportContinuityToken,
        });
        setInspectorMessage("Evidence attached.");
        finishSave("Saved");
      });
    },
    [
      applyAcceptedRowMutation,
      beginSave,
      beginViewportContinuity,
      clearViewportContinuity,
      enqueueSaveWork,
      evidenceAttachmentPort,
      finishSave,
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
