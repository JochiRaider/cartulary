import type {
  ExplicitPatchContribution,
  ExplicitPatchIntent,
} from "../../runtime/WorkbookExplicitPatchOwner";
import {
  type TaskLifecycleDraftStore,
  taskGuardFields,
  taskPatchErrors,
  taskViewId,
} from "./taskLifecycleModel";

export function taskExplicitPatchContribution(
  intent: ExplicitPatchIntent,
  drafts: TaskLifecycleDraftStore,
): readonly ExplicitPatchContribution[] {
  if (intent.viewSchemaId !== taskViewId) return [];
  const captured = drafts.capture(intent.baseline.record_id);
  return [
    {
      dependencies: intent.changes.some((change) =>
        taskGuardFields.some((field) => field === change.field_key),
      )
        ? taskGuardFields
        : [],
      validate: (row, request) => {
        const errors = taskPatchErrors(row, request.changes);
        return errors.length
          ? {
              kind: "validation",
              message: errors[0]?.message ?? "Invalid Task state",
              fields: errors,
            }
          : null;
      },
      acknowledged: () => {
        if (intent.purpose === "task-lifecycle")
          drafts.acknowledge(intent.baseline.record_id, captured);
      },
    },
  ];
}
