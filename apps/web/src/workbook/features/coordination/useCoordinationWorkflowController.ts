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
  const generationRef = useRef(0);

  useEffect(() => {
    void resetKey;
    generationRef.current += 1;
    setLifecycleRecordId("");
    setLifecycleStatus("blocked");
    setLifecycleBlockedReason("");
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
    const finish = mutation.beginMutation();
    try {
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
    } finally {
      finish();
    }
  }, [
    lifecycleBlockedReason,
    lifecycleRecordId,
    lifecycleStatus,
    mutation,
    mutationCommands,
    rows,
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
  };
}
