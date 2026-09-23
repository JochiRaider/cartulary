import { useRef, useState, useSyncExternalStore } from "react";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type TaskLifecycleDraftStore,
  taskDraftStaleFields,
  taskLifecycleChanges,
  taskPatchErrors,
  taskValue,
  taskViewId,
} from "./taskLifecycleModel";

export type CoordinationWorkflowMutationPorts = Pick<
  GenericSurfaceMutationController,
  "submitPatchMutation"
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
  const [feedback, setFeedback] = useState<{
    recordId: string;
    captured: ReturnType<TaskLifecycleDraftStore["capture"]>;
    failure: WorkbookOperationFailure;
  } | null>(null);
  const current = useRef({
    recordId: row.record_id,
    captured: drafts.capture(row.record_id),
    disabled,
  });
  current.current = {
    recordId: row.record_id,
    captured: drafts.capture(row.record_id),
    disabled,
  };
  const draft = drafts.read(row);
  const changes = taskLifecycleChanges(draft, row);
  const visibleFailure =
    feedback?.recordId === row.record_id &&
    feedback.captured === current.current.captured &&
    !disabled
      ? feedback.failure
      : null;
  const fieldFailures =
    visibleFailure?.kind === "validation"
      ? (visibleFailure.fields ?? []).filter((item) =>
          changes.some((change) => change.field_key === item.field),
        )
      : [];
  const errors = [...taskPatchErrors(row, changes), ...fieldFailures];
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
    const captured = drafts.capture(row.record_id);
    setFeedback(null);
    try {
      const accepted = await mutation.submitPatchMutation({
        changes,
        purpose: "task-lifecycle",
        viewSchemaId: taskViewId,
        baseline: row,
        onFailure: (failure) => {
          if (
            current.current.recordId === row.record_id &&
            current.current.captured === captured
          )
            setFeedback({ recordId: row.record_id, captured, failure });
        },
      });
      if (accepted) drafts.acknowledge(row.record_id, captured);
    } finally {
      submitting.current = false;
    }
  };
  return {
    value,
    changes,
    errors,
    actionFailure: fieldFailures.length ? null : visibleFailure,
    staleFields,
    submit,
    update: (field: string, next: string) => drafts.update(row, field, next),
    review: (field: string, keepDraft: boolean) =>
      drafts.review(row, field, keepDraft),
  };
}
