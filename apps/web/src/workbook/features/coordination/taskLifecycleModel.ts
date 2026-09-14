import { requireViewContract } from "@cartulary/view-contracts";
import type { WorkbookProtocolPatchRecordRequest } from "../../adapters/workbookProtocolTypes";
import { buildGenericPatchChange } from "../../models/genericWorkbookModel";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

export const taskViewId = "cartulary.view.task_requests.v1";
export const taskStatuses = [
  "open",
  "in_progress",
  "blocked",
  "done",
  "canceled",
] as const;
export type TaskLifecycleStatus = (typeof taskStatuses)[number];
export const taskGuardFields = [
  "task.status",
  "task.owner_user_id",
  "task.blocked_reason",
  "task.completed_at",
] as const;
export type TaskGuardField = (typeof taskGuardFields)[number];
export type RecordPatchChange =
  WorkbookProtocolPatchRecordRequest["changes"][number];
export type TaskFieldError = {
  readonly field: string;
  readonly message: string;
};
export type TaskDraft = {
  readonly baseline: WorkbookQueryRow;
  readonly values: Readonly<Record<string, string>>;
};

export function taskValue(row: WorkbookQueryRow, field: string): string {
  const value = row.cells[field]?.value;
  return value === null || value === undefined ? "" : String(value);
}
export function taskFieldEqual(
  left: WorkbookQueryRow,
  right: WorkbookQueryRow,
  field: string,
): boolean {
  return (
    JSON.stringify(left.cells[field]?.value ?? null) ===
    JSON.stringify(right.cells[field]?.value ?? null)
  );
}
export function taskStatus(value: string): TaskLifecycleStatus | null {
  return taskStatuses.find((status) => status === value) ?? null;
}
export function taskTransitionAllowed(from: string, to: string): boolean {
  return (
    taskStatus(from) !== null &&
    taskStatus(to) !== null &&
    !(
      (from === "done" && to === "canceled") ||
      (from === "canceled" && to === "done")
    )
  );
}
export function taskPatchErrors(
  row: WorkbookQueryRow,
  changes: readonly RecordPatchChange[],
): readonly TaskFieldError[] {
  if (
    !changes.some((change) =>
      taskGuardFields.some((field) => field === change.field_key),
    )
  )
    return [];
  const next = Object.fromEntries(
    taskGuardFields.map((field) => [field, taskValue(row, field)]),
  );
  for (const change of changes)
    if ("value" in change)
      next[change.field_key] =
        change.value === null ? "" : String(change.value);
  const from = taskValue(row, "task.status"),
    to = next["task.status"] ?? "";
  const errors: TaskFieldError[] = [];
  if (!taskTransitionAllowed(from, to))
    errors.push({
      field: "task.status",
      message: `A ${from} Task cannot become ${to}. Reopen it first.`,
    });
  if (to === "blocked" && !next["task.blocked_reason"]?.trim())
    errors.push({
      field: "task.blocked_reason",
      message:
        "Blocked Tasks need a reason. Supply it with the status in Workflow.",
    });
  if (
    to !== "blocked" &&
    from !== "blocked" &&
    next["task.blocked_reason"]?.trim()
  )
    errors.push({
      field: "task.blocked_reason",
      message: "A reason is only saved while the Task is blocked.",
    });
  if (!["done", "canceled"].includes(to) && !next["task.owner_user_id"])
    errors.push({
      field: "task.owner_user_id",
      message:
        "Active Tasks need an owner. Supply one with the status in Workflow.",
    });
  const explicitTime = changes.find(
    (change) => change.field_key === "task.completed_at",
  );
  if (to === "done" && explicitTime && !next["task.completed_at"])
    errors.push({
      field: "task.completed_at",
      message: "Completed Tasks need a completion time.",
    });
  if (to !== "done" && from !== "done" && next["task.completed_at"])
    errors.push({
      field: "task.completed_at",
      message: "Completion time is only saved for a done Task.",
    });
  const completionField =
    requireViewContract(taskViewId).fieldMap["task.completed_at"];
  if (
    explicitTime &&
    completionField &&
    buildGenericPatchChange(
      completionField,
      next["task.completed_at"] ?? "",
      "add",
      taskViewId,
    ) === null
  )
    errors.push({
      field: "task.completed_at",
      message: "Enter a valid RFC 3339 completion time with a timezone.",
    });
  for (const change of changes) {
    const field = requireViewContract(taskViewId).fieldMap[change.field_key];
    if (
      field &&
      "value" in change &&
      change.field_key !== "task.completed_at" &&
      buildGenericPatchChange(
        field,
        change.value === null ? "" : String(change.value),
        "add",
        taskViewId,
      ) === null &&
      !errors.some((error) => error.field === change.field_key)
    ) {
      errors.push({
        field: change.field_key,
        message: `Enter a valid ${field.label.toLowerCase()}.`,
      });
    }
  }
  return errors;
}
export function taskDraftStaleFields(
  draft: TaskDraft,
  row: WorkbookQueryRow,
): readonly string[] {
  if (Object.keys(draft.values).length === 0) return [];
  return taskGuardFields.filter(
    (field) => !taskFieldEqual(draft.baseline, row, field),
  );
}
export function taskLifecycleChanges(
  draft: TaskDraft,
  row: WorkbookQueryRow,
): readonly RecordPatchChange[] {
  const contract = requireViewContract(taskViewId);
  const status = draft.values["task.status"] ?? taskValue(row, "task.status");
  const values: Record<string, string> = {
    ...draft.values,
    "task.status": status,
  };
  if (status !== "blocked") delete values["task.blocked_reason"];
  if (
    status !== "done" ||
    (taskValue(row, "task.status") !== "done" && !values["task.completed_at"])
  )
    delete values["task.completed_at"];
  return Object.entries(values).flatMap(([fieldKey, value]) => {
    const field = contract.fieldMap[fieldKey];
    if (!field || field.writeKind === "read_only") return [];
    const change = buildGenericPatchChange(field, value, "add", taskViewId);
    // Retain invalid intent for local guard feedback; transport construction
    // still rejects values outside the declared field contract.
    return [change ?? { field_key: fieldKey, value }];
  });
}

/** Memory-only drafts; presentation retargeting never changes record identity. */
export class TaskLifecycleDraftStore {
  private readonly drafts = new Map<string, TaskDraft>();
  private revision = 0;
  private readonly listeners = new Set<() => void>();
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.revision;
  read(row: WorkbookQueryRow): TaskDraft {
    return this.drafts.get(row.record_id) ?? { baseline: row, values: {} };
  }
  capture(recordId: string) {
    return this.drafts.get(recordId) ?? null;
  }
  acknowledge(recordId: string, captured: TaskDraft | null) {
    if (captured && this.drafts.get(recordId) === captured)
      this.clear(recordId);
  }
  update(row: WorkbookQueryRow, field: string, value: string): void {
    const draft = this.read(row);
    this.drafts.set(row.record_id, {
      baseline: draft.baseline,
      values: { ...draft.values, [field]: value },
    });
    this.emit();
  }
  review(row: WorkbookQueryRow, field: string, keepDraft: boolean): void {
    const draft = this.read(row),
      values = { ...draft.values };
    if (!keepDraft) delete values[field];
    this.drafts.set(row.record_id, {
      baseline: {
        ...draft.baseline,
        cells: {
          ...draft.baseline.cells,
          [field]: { value: row.cells[field]?.value ?? null },
        },
      },
      values,
    });
    this.emit();
  }
  clear(recordId?: string): void {
    if (recordId) this.drafts.delete(recordId);
    else this.drafts.clear();
    this.emit();
  }
  private emit() {
    this.revision++;
    for (const listener of this.listeners) listener();
  }
}
