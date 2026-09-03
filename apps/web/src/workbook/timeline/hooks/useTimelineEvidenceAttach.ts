import { useCallback, useRef } from "react";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorMessageFeedback,
  workbookInspectorOperationFailureFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  planTimelineEvidenceTarget,
  type TimelineEvidenceActionContext,
  type TimelineEvidenceTargetIdentity,
  timelineEvidenceTargetIdentity,
} from "../models/timelineEvidenceAttachmentPlan";
import type { TimelineApiRow, WorkbookRow } from "../models/timelineRowModel";
import type { TimelineEvidenceAttachmentPort } from "../ports/TimelineEvidenceAttachmentPort";

type TimelineEvidenceViewportContinuityTarget =
  | { kind: "row-inspect"; recordId: string }
  | { kind: "input"; focusKey: string }
  | { kind: "scroll-only" };

type TimelineEvidenceAttachInput = {
  readonly actionContext: TimelineEvidenceActionContext;
  readonly applyAcceptedRowMutation: (
    rowKey: string,
    mutation: {
      readonly row: TimelineApiRow;
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
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly trackPendingSocketTxn: (clientTxnId: string) => void;
  readonly waitForCommittedRecordIdle: (
    recordId: string,
  ) => Promise<{ row: WorkbookRow | null; rowVersion: number } | null>;
};

export function useTimelineEvidenceAttach(input: TimelineEvidenceAttachInput) {
  const inputRef = useRef(input);
  inputRef.current = input;

  const attachEvidenceFileToTimeline = useCallback(
    (target: WorkbookRow, file: File) => {
      const current = inputRef.current;
      const identity = timelineEvidenceTargetIdentity(
        target,
        current.actionContext.surfaceKey,
      );
      const initial = planTimelineEvidenceTarget({
        context: current.actionContext,
        identity,
        rows: current.rowsRef.current,
      });
      if (initial.kind === "reject") {
        publishUnavailableEvidenceTarget(current, false);
        return;
      }
      const viewportContinuityToken = current.beginViewportContinuity(
        initial.target.recordId === null
          ? { kind: "scroll-only" }
          : { kind: "row-inspect", recordId: initial.target.recordId },
      );
      current.beginSave();
      current.setInspectorMessage(
        workbookInspectorMessageFeedback("Uploading evidence.", "none"),
      );
      current.enqueueSaveWork(() =>
        executeTimelineEvidenceAttachment({
          file,
          identity,
          inputRef,
          viewportContinuityToken,
        }),
      );
    },
    [],
  );

  const handleTimelineEvidenceFiles = useCallback(
    (target: WorkbookRow, files: FileList | File[]) => {
      const [file] = Array.from(files);
      if (file !== undefined) attachEvidenceFileToTimeline(target, file);
    },
    [attachEvidenceFileToTimeline],
  );

  return { attachEvidenceFileToTimeline, handleTimelineEvidenceFiles };
}

async function executeTimelineEvidenceAttachment(options: {
  readonly file: File;
  readonly identity: TimelineEvidenceTargetIdentity;
  readonly inputRef: { readonly current: TimelineEvidenceAttachInput };
  readonly viewportContinuityToken: number;
}): Promise<void> {
  const beforeCreate = await currentEvidenceTarget(options);
  if (beforeCreate === null) return;
  const created =
    await options.inputRef.current.evidenceAttachmentPort.createEvidence({
      file: options.file,
    });
  if (created.kind === "rejected") {
    failEvidenceAttachment(
      options.inputRef.current,
      options.viewportContinuityToken,
      created.failure,
    );
    return;
  }
  const beforeAttach = await currentEvidenceTarget(options);
  if (beforeAttach === null) return;
  const current = options.inputRef.current;
  const result = await current.evidenceAttachmentPort.attachEvidence({
    evidenceRecordId: created.value.evidenceRecordId,
    onTimelineClientTxnId: current.trackPendingSocketTxn,
    target: beforeAttach,
  });
  if (result.clientTxnId !== null) {
    options.inputRef.current.resolvePendingSocketTxn(result.clientTxnId);
  }
  const settled = planTimelineEvidenceTarget({
    context: options.inputRef.current.actionContext,
    identity: options.identity,
    rows: options.inputRef.current.rowsRef.current,
  });
  if (settled.kind === "reject") {
    publishUnavailableEvidenceTarget(
      options.inputRef.current,
      true,
      options.viewportContinuityToken,
    );
    return;
  }
  if (result.outcome.kind === "rejected") {
    failEvidenceAttachment(
      options.inputRef.current,
      options.viewportContinuityToken,
      result.outcome.failure,
    );
    return;
  }
  options.inputRef.current.applyAcceptedRowMutation(
    beforeAttach.key,
    result.outcome.value,
    {
      continueOnFreshDraft: beforeAttach.recordId === null,
      promoteToCommittedRowInspect: beforeAttach.recordId === null,
      detectAutoResolution: false,
      viewportContinuityToken: options.viewportContinuityToken,
    },
  );
  options.inputRef.current.setInspectorMessage(
    workbookInspectorMessageFeedback("Evidence attached.", "none"),
  );
  options.inputRef.current.finishSave("Saved");
}

async function currentEvidenceTarget(options: {
  readonly identity: TimelineEvidenceTargetIdentity;
  readonly inputRef: { readonly current: TimelineEvidenceAttachInput };
  readonly viewportContinuityToken: number;
}): Promise<WorkbookRow | null> {
  const input = options.inputRef.current;
  const currentPlan = planTimelineEvidenceTarget({
    context: input.actionContext,
    identity: options.identity,
    rows: input.rowsRef.current,
  });
  if (currentPlan.kind === "reject") {
    publishUnavailableEvidenceTarget(
      input,
      true,
      options.viewportContinuityToken,
    );
    return null;
  }
  if (currentPlan.target.recordId === null) return currentPlan.target;
  const idle = await input.waitForCommittedRecordIdle(
    currentPlan.target.recordId,
  );
  const latest = options.inputRef.current;
  const idlePlan = planTimelineEvidenceTarget({
    context: latest.actionContext,
    identity: options.identity,
    rows: idle?.row === null || idle === null ? [] : [idle.row],
  });
  if (idlePlan.kind === "dispatch") return idlePlan.target;
  publishUnavailableEvidenceTarget(
    latest,
    true,
    options.viewportContinuityToken,
  );
  return null;
}

function publishUnavailableEvidenceTarget(
  input: TimelineEvidenceAttachInput,
  finish: boolean,
  viewportContinuityToken?: number,
): void {
  if (viewportContinuityToken !== undefined) {
    input.clearViewportContinuity(viewportContinuityToken);
  }
  input.setInspectorMessage(
    workbookInspectorMessageFeedback(
      "The selected Timeline row is no longer available.",
      "none",
    ),
  );
  if (finish) input.finishSave("Conflict");
}

function failEvidenceAttachment(
  input: TimelineEvidenceAttachInput,
  viewportContinuityToken: number,
  failure: Parameters<typeof workbookInspectorOperationFailureFeedback>[0],
): void {
  input.clearViewportContinuity(viewportContinuityToken);
  input.setInspectorMessage(workbookInspectorOperationFailureFeedback(failure));
  input.finishSave("Conflict");
}
