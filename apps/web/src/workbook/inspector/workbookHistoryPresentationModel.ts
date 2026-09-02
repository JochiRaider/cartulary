import type {
  WorkbookHistoryEventPresentation,
  WorkbookInspectorTechnicalField,
} from "./presentation/workbookInspectorPresentationModel";
import type {
  RecordHistoryItem,
  RecordHistoryRollbackAction,
} from "./workbookRecordHistoryModel";

export function workbookHistoryEventPresentation(
  item: RecordHistoryItem,
  actorLabel?: string,
): WorkbookHistoryEventPresentation {
  return {
    ...(actorLabel === undefined ? {} : { actorLabel }),
    committedAt: item.committed_at,
    key: item.history_item_ref,
    operation: item.operation,
    summary: item.diff_summary.summary,
    technicalFields: workbookHistoryTechnicalFields(item),
  };
}

export function workbookHistoryRollbackLabel(
  action: RecordHistoryRollbackAction,
): string {
  if (action === "history_entry") return "Rollback entry";
  if (action === "change_set") return "Rollback change set";
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
