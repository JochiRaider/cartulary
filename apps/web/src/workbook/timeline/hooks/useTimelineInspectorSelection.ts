import { useMemo, useState } from "react";
import {
  buildInspectorMentions,
  type DismissedMention,
} from "../models/workbookMentionChips";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export function useTimelineInspectorSelection({
  currentIncidentRole,
  dismissedMentionsByRow,
  rows,
  selectedMentionRef,
}: {
  readonly currentIncidentRole: string | null | undefined;
  readonly dismissedMentionsByRow: Record<string, DismissedMention[]>;
  readonly rows: readonly WorkbookRow[];
  readonly selectedMentionRef: string | null;
}) {
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const [selectedRowId, setSelectedRowId] = useState<string | null>(null);
  const selectedRow = useMemo(
    () =>
      rows.find(
        (row) => row.recordId !== null && row.recordId === selectedRowId,
      ) ?? null,
    [rows, selectedRowId],
  );
  const draftRow = useMemo(
    () => rows.find((row) => row.recordId === null) ?? null,
    [rows],
  );
  const dismissedForSelectedRow = selectedRow?.recordId
    ? (dismissedMentionsByRow[selectedRow.recordId] ?? [])
    : [];
  const inspectorMentions = useMemo(
    () =>
      buildInspectorMentions(selectedRow ?? undefined, dismissedForSelectedRow),
    [dismissedForSelectedRow, selectedRow],
  );
  const selectedMention =
    inspectorMentions.find((item) => item.itemRef === selectedMentionRef) ??
    inspectorMentions[0] ??
    null;
  const canManageMentions =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const selectedRowWorkflowKey =
    selectedRow?.recordId && selectedRow.rowVersion !== null
      ? `${selectedRow.recordId}:${selectedRow.rowVersion}`
      : (selectedRow?.recordId ?? "");

  return {
    commands: {
      setIsInspectorOpen,
      setSelectedRowId,
    },
    snapshot: {
      canManageMentions,
      draftRow,
      dismissedForSelectedRow,
      isInspectorOpen,
      inspectorMentions,
      selectedMention,
      selectedRow,
      selectedRowId,
      selectedRowWorkflowKey,
    },
  };
}
