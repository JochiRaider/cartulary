import { useCallback, useEffect, useRef, useState } from "react";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import { normalizeGenericTextValue } from "../../models/genericWorkbookModel";
import type {
  CoordinationMutationCommandPort,
  TaskLifecycleStatus,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

export type CoordinationWorkflowMutationPorts = Pick<
  GenericSurfaceMutationController,
  | "beginMutation"
  | "completeGenericMutation"
  | "rejectMutationFailure"
  | "setValidationError"
>;

export function useCoordinationWorkflowController({
  mutation,
  mutationCommands,
  resetKey,
  rows,
}: {
  readonly mutation: CoordinationWorkflowMutationPorts;
  readonly mutationCommands: CoordinationMutationCommandPort;
  readonly resetKey: string;
  readonly rows: readonly WorkbookQueryRow[];
}) {
  const [lifecycleRecordId, setLifecycleRecordId] = useState("");
  const [lifecycleStatus, setLifecycleStatus] =
    useState<TaskLifecycleStatus>("blocked");
  const [lifecycleBlockedReason, setLifecycleBlockedReason] = useState("");
  const [supersedeTargetId, setSupersedeTargetId] = useState("");
  const [supersedeReplacementId, setSupersedeReplacementId] = useState("");
  const [supersedeReason, setSupersedeReason] = useState("");
  const generationRef = useRef(0);

  useEffect(() => {
    void resetKey;
    generationRef.current += 1;
    setLifecycleRecordId("");
    setLifecycleStatus("blocked");
    setLifecycleBlockedReason("");
    setSupersedeTargetId("");
    setSupersedeReplacementId("");
    setSupersedeReason("");
  }, [resetKey]);

  useEffect(
    () => () => {
      generationRef.current += 1;
    },
    [],
  );

  const submitLifecyclePatch = useCallback(async () => {
    const target = rows.find((row) => row.record_id === lifecycleRecordId);
    if (!target) {
      mutation.setValidationError("Select a task row.");
      return;
    }
    let blockedReason: string | undefined;
    if (lifecycleStatus === "blocked") {
      const reason = normalizeGenericTextValue(lifecycleBlockedReason);
      if (reason === "") {
        mutation.setValidationError("Blocked tasks need a reason.");
        return;
      }
      blockedReason = reason;
    }
    const generation = generationRef.current;
    mutation.beginMutation();
    const result = await mutationCommands.updateTaskLifecycle({
      baseRowVersion: target.row_version,
      blockedReason,
      recordId: target.record_id,
      status: lifecycleStatus,
    });
    if (generationRef.current !== generation) return;
    if (result.kind === "rejected") {
      mutation.rejectMutationFailure(result.failure);
      return;
    }
    if (lifecycleStatus !== "blocked") {
      setLifecycleBlockedReason("");
    }
    await mutation.completeGenericMutation();
  }, [
    lifecycleBlockedReason,
    lifecycleRecordId,
    lifecycleStatus,
    mutation,
    mutationCommands,
    rows,
  ]);

  const submitSupersede = useCallback(async () => {
    const target = rows.find((row) => row.record_id === supersedeTargetId);
    if (!target || supersedeReplacementId === "") {
      mutation.setValidationError("Select target and superseding decisions.");
      return;
    }
    if (target.record_id === supersedeReplacementId) {
      mutation.setValidationError("Select a different superseding decision.");
      return;
    }
    const reason = normalizeGenericTextValue(supersedeReason);
    if (reason === "") {
      mutation.setValidationError("Reason is required.");
      return;
    }
    const generation = generationRef.current;
    mutation.beginMutation();
    const result = await mutationCommands.supersedeDecision({
      baseRowVersion: target.row_version,
      reason,
      replacementRecordId: supersedeReplacementId,
      targetRecordId: target.record_id,
    });
    if (generationRef.current !== generation) return;
    if (result.kind === "rejected") {
      mutation.rejectMutationFailure(result.failure);
      return;
    }
    setSupersedeReason("");
    await mutation.completeGenericMutation();
  }, [
    mutation,
    mutationCommands,
    rows,
    supersedeReason,
    supersedeReplacementId,
    supersedeTargetId,
  ]);

  return {
    lifecycle: {
      blockedReason: lifecycleBlockedReason,
      recordId: lifecycleRecordId,
      setBlockedReason: setLifecycleBlockedReason,
      setRecordId: setLifecycleRecordId,
      setStatus: setLifecycleStatus,
      status: lifecycleStatus,
      submit: submitLifecyclePatch,
    },
    supersede: {
      reason: supersedeReason,
      replacementId: supersedeReplacementId,
      setReason: setSupersedeReason,
      setReplacementId: setSupersedeReplacementId,
      setTargetId: setSupersedeTargetId,
      submit: submitSupersede,
      targetId: supersedeTargetId,
    },
  };
}
