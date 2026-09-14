import {
  decisionsViewSchemaId,
  taskRequestsViewSchemaId,
} from "@cartulary/view-contracts";
/** Initial Task/Decision lifecycle guards shared by ordinary and contextual creation. */
export function initialCoordinationCreateErrors(
  view: string,
  values: Readonly<Record<string, unknown>>,
) {
  const errors: Record<string, string> = {};
  const text = (key: string) =>
    typeof values[key] === "string" ? values[key].trim() : "";
  if (view === taskRequestsViewSchemaId) {
    const status = text("task.status") || "open";
    if (status === "blocked" && !text("task.blocked_reason"))
      errors["task.blocked_reason"] = "Blocked Tasks need a reason.";
    if (status !== "blocked" && text("task.blocked_reason"))
      errors["task.blocked_reason"] =
        "A reason is only saved while the Task is blocked.";
    if (status !== "done" && text("task.completed_at"))
      errors["task.completed_at"] =
        "Completion time is only saved for a done Task.";
  } else if (
    view === decisionsViewSchemaId &&
    values["decision.status"] === "superseded"
  )
    errors["decision.status"] =
      "Create a proposed, approved, rejected, or executed Decision.";
  return errors;
}
