import { useRef, useSyncExternalStore } from "react";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type TaskLifecycleDraftStore,
  taskDraftStaleFields,
  taskLifecycleChanges,
  taskPatchErrors,
  taskValue,
  taskViewId,
} from "./taskLifecycleModel";

export type CoordinationWorkflowMutationPorts = Partial<
  Pick<GenericSurfaceMutationController, "explicitPatches">
> &
  Pick<
    GenericSurfaceMutationController,
    "beginMutation" | "completeGenericMutation" | "submitPatchMutation"
  >;

export function useCoordinationWorkflowController({
  mutation,
  drafts,
  row,
  disabled,
}: {
  readonly mutation: CoordinationWorkflowMutationPorts;
  readonly drafts: TaskLifecycleDraftStore;
  readonly row: WorkbookQueryRow;
  readonly disabled: boolean;
}) {
  useSyncExternalStore(drafts.subscribe, drafts.getSnapshot);
  const submitting = useRef(false);
  const draft = drafts.read(row);
  const changes = taskLifecycleChanges(draft, row);
  const errors = taskPatchErrors(row, changes);
  const staleFields = taskDraftStaleFields(draft, row);
  const value = (field: string) => draft.values[field] ?? taskValue(row, field);
  const submit = async () => {
    if (
      disabled ||
      submitting.current ||
      errors.length > 0 ||
      staleFields.length > 0
    )
      return;
    submitting.current = true;
    const finish = mutation.beginMutation();
    try {
      const accepted = await mutation.submitPatchMutation({
        baseRowVersion: row.row_version,
        changes,
        purpose: "task-lifecycle",
        recordId: row.record_id,
        viewSchemaId: taskViewId,
        baseline: row,
      });
      if (accepted) drafts.clear(row.record_id);
    } finally {
      submitting.current = false;
      finish();
    }
  };
  return {
    value,
    changes,
    errors,
    staleFields,
    submit,
    update: (field: string, next: string) => drafts.update(row, field, next),
    review: (field: string, keepDraft: boolean) =>
      drafts.review(row, field, keepDraft),
  };
}
