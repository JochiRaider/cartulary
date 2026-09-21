import type { RecordHistoryItem } from "../adapters/workbookHistoryResponse";
import type { RecordHistoryRollbackAction } from "../history/workbookHistoryItem";

import type {
  WorkbookHistoryEventPresentation,
  WorkbookInspectorTechnicalField,
} from "./presentation/workbookInspectorPresentationModel";

export function workbookHistoryEventPresentation(
  item: RecordHistoryItem,
  actorLabel?: string,
): WorkbookHistoryEventPresentation {
  const units = item.diff_summary.units.map((unit) => ({
    key: unit.unit_ref,
    title: historyUnitTitle(unit),
    recordIds: unit.record_ids,
    changes: unit.changes.map((change) => ({
      fieldKey: change.field_key,
      label: historyFieldLabel(change.field_key),
      before: historyValueText(change.before),
      after: historyValueText(change.after),
    })),
  }));
  return {
    actorLabel:
      actorLabel?.trim() ||
      `Identifier ${item.source_actor_id ?? item.actor_user_id}`,
    committedAt: item.committed_at,
    key: item.history_item_ref,
    operation:
      item.operation === "row_restore"
        ? "Row fields restored"
        : item.operation.startsWith("rollback")
          ? "Reversed"
          : item.diff_summary.units.some((unit) => unit.operation === "merge")
            ? "Merged"
            : historyOperationText(item.diff_summary.units[0].operation),
    summary:
      units.length === 1
        ? (units[0]?.title ?? "Record changed")
        : `${units.length} changes: ${[...new Set(item.diff_summary.units.map((unit) => historyKindLabel[unit.kind]))].join(", ")}`,
    units,
    technicalFields: workbookHistoryTechnicalFields(item),
  };
}

type HistoryUnit = RecordHistoryItem["diff_summary"]["units"][number];
const historyKindLabel = {
  field: "Field",
  link: "Link",
  mention: "Mention",
  tag: "Tag",
  evidence_association: "Evidence association",
  capture_state: "Capture state",
  record: "Record",
  entity_identifier: "Entity identifier",
  indicator_observation: "Indicator observation",
  indicator_interval: "Indicator interval",
} satisfies Record<HistoryUnit["kind"], string>;

function historyOperationText(operation: HistoryUnit["operation"]): string {
  return {
    create: "Created",
    update: "Updated",
    delete: "Deleted",
    restore: "Restored",
    add: "Added",
    remove: "Removed",
    merge: "Merged",
  }[operation];
}

function historyUnitTitle(unit: HistoryUnit): string {
  const field = unit.changes[0];
  const label =
    unit.kind === "field" && field
      ? historyFieldLabel(field.field_key)
      : historyKindLabel[unit.kind];
  return `${label} ${historyOperationText(unit.operation).toLowerCase()}`;
}

function historyFieldLabel(fieldKey: string): string {
  const member = fieldKey.split(".").at(-1) ?? fieldKey;
  const text = member.replaceAll("_", " ");
  return text.charAt(0).toUpperCase() + text.slice(1);
}

function historyValueText(
  value: HistoryUnit["changes"][number]["before"],
): string {
  if (value.state === "absent") return "Not present";
  if (value.state === "null") return "No value (null)";
  if (value.value === "") return "Empty text";
  if (typeof value.value === "boolean") return value.value ? "True" : "False";
  if (Array.isArray(value.value))
    return value.value.length ? value.value.join("\n") : "No items";
  return typeof value.value === "string"
    ? `“${value.value}”`
    : String(value.value);
}

export function workbookHistoryRollbackLabel(
  action: RecordHistoryRollbackAction,
): string {
  if (action === "history_entry") return "Reverse history entry";
  if (action === "change_set") return "Reverse change set";
  return "Restore row fields";
}

export function workbookHistoryPendingTechnicalFields({
  recordId,
  rowVersion,
}: {
  readonly recordId: string;
  readonly rowVersion: number | null;
}): readonly WorkbookInspectorTechnicalField[] {
  return [
    { label: "Record ID", value: recordId },
    {
      label: "Row version",
      value: rowVersion === null ? "unknown" : String(rowVersion),
    },
  ];
}

function workbookHistoryTechnicalFields(
  item: RecordHistoryItem,
): readonly WorkbookInspectorTechnicalField[] {
  return [
    { label: "Actor ID", value: item.actor_user_id },
    ...(item.source_actor_id
      ? [{ label: "Source actor ID", value: item.source_actor_id }]
      : []),
    { label: "Source operation", value: item.operation },
    { label: "History reference", value: item.history_item_ref },
    { label: "Change set ID", value: item.change_set_id },
    ...(item.history_entry_ref === undefined
      ? []
      : [{ label: "History entry", value: item.history_entry_ref }]),
    ...(item.revision_no === undefined
      ? []
      : [{ label: "Revision", value: String(item.revision_no) }]),
  ];
}
