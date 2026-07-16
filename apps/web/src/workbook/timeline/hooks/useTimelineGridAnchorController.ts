import {
  buildGridPresentationRows,
  type GridCellAnchor,
  type GridColumn,
  type GridDataRow,
  type GridHandle,
  type GridNavigationIntent,
  type GridPasteTargetResolution,
  navigateGridCellAnchor,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import { useCallback } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  clipboardGridDimensions,
  clipboardTextLooksTabular,
} from "../../utils/workbookClipboard";
import type { WorkbookFocusAnchor } from "../../utils/workbookGridFocus";
import type { TimelinePasteTargetResolution } from "../models/timelineControllerPorts";
import {
  timelineGroupLabel,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

type TimelineReadonlyRef<T> = {
  readonly current: T;
};

function timelineClipboardShouldDispatchTabular(
  fieldKey: string,
  clipboardText: string,
) {
  if (clipboardTextLooksTabular(clipboardText)) {
    return true;
  }
  return (
    fieldKey === "timeline.activity_utc_text" &&
    clipboardGridDimensions(clipboardText).columnCount > 1
  );
}

function timelinePasteColumnsFromStart(
  columns: readonly GridColumn<WorkbookRow>[],
  startFieldKey: string,
  pastedColumnCount: number,
) {
  const startColumnIndex = columns.findIndex(
    (column) => column.fieldKey === startFieldKey,
  );
  if (startColumnIndex < 0) {
    return null;
  }
  const targetColumns = columns
    .slice(startColumnIndex, startColumnIndex + pastedColumnCount)
    .map((column) => column.fieldKey);
  return targetColumns.length < 1 ? null : targetColumns;
}

function resolveDraftTimelinePasteTargets({
  columns,
  pastedColumnCount,
  pastedRowCount,
  startFieldKey,
  viewSchemaId,
}: {
  readonly columns: readonly GridColumn<WorkbookRow>[];
  readonly pastedColumnCount: number;
  readonly pastedRowCount: number;
  readonly startFieldKey: string;
  readonly viewSchemaId: string;
}): GridPasteTargetResolution | null {
  const targetColumns = timelinePasteColumnsFromStart(
    columns,
    startFieldKey,
    pastedColumnCount,
  );
  if (targetColumns === null || pastedRowCount < 1) {
    return null;
  }
  return {
    columns: targetColumns,
    rowTargets: Array.from({ length: pastedRowCount }, (_, createIndex) => ({
      createIndex,
      kind: "create" as const,
      surface: { kind: "view_schema" as const, viewSchemaId },
    })),
  };
}

export function useTimelineGridAnchorController({
  gridHandleRef,
  groupBy,
  rowsRef,
  timelineAnchorColumnsRef,
  timelineAnchorRowsRef,
  updateTimelineSurfaceFocusAnchor,
  updateWorkbookFocusAnchor,
}: {
  readonly gridHandleRef: TimelineReadonlyRef<GridHandle | null>;
  readonly groupBy: string | null;
  readonly rowsRef: TimelineReadonlyRef<readonly WorkbookRow[]>;
  readonly timelineAnchorColumnsRef: TimelineReadonlyRef<
    readonly GridColumn<WorkbookRow>[]
  >;
  readonly timelineAnchorRowsRef: TimelineReadonlyRef<
    readonly GridDataRow<WorkbookRow>[]
  >;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
  readonly updateWorkbookFocusAnchor: (anchor: null) => void;
}) {
  const restoreTimelineFocusAnchor = useCallback(
    (
      anchor:
        | GridCellAnchor
        | Pick<WorkbookFocusAnchor, "fieldKey" | "recordId" | "viewSchemaId">,
    ) => {
      const semanticAnchor: GridCellAnchor =
        "rowIdentity" in anchor
          ? anchor
          : {
              fieldKey: anchor.fieldKey,
              rowIdentity: {
                kind: "core_record",
                recordId: anchor.recordId,
              },
              surface: {
                kind: "view_schema",
                viewSchemaId: anchor.viewSchemaId,
              },
            };
      const restored =
        gridHandleRef.current?.focusAnchor(semanticAnchor) ?? false;
      if (!restored) {
        window.setTimeout(() => {
          gridHandleRef.current?.focusAnchor(semanticAnchor);
        }, 0);
      }
      return restored;
    },
    [gridHandleRef],
  );

  const currentTimelineAnchorFor = useCallback(
    (rowKey: string, fieldKey: string): GridCellAnchor | null => {
      const row = rowsRef.current.find((candidate) => candidate.key === rowKey);
      if (row?.recordId === null || row?.recordId === undefined) {
        updateWorkbookFocusAnchor(null);
        return null;
      }
      const anchor = {
        fieldKey,
        rowIdentity: { kind: "core_record" as const, recordId: row.recordId },
        surface: {
          kind: "view_schema" as const,
          viewSchemaId: timelineViewSchemaId,
        },
      };
      updateTimelineSurfaceFocusAnchor(
        anchor.rowIdentity.recordId,
        anchor.fieldKey,
      );
      return anchor;
    },
    [rowsRef, updateTimelineSurfaceFocusAnchor, updateWorkbookFocusAnchor],
  );

  const resolveTimelinePasteTargetResolution = useCallback(
    (
      rowKey: string,
      fieldKey: string,
      clipboardText: string,
    ): TimelinePasteTargetResolution | null => {
      if (!timelineClipboardShouldDispatchTabular(fieldKey, clipboardText)) {
        return null;
      }

      const dimensions = clipboardGridDimensions(clipboardText);
      const row = rowsRef.current.find((candidate) => candidate.key === rowKey);
      const isDraftTarget =
        row?.recordId === null ||
        (row === undefined && rowKey.startsWith("draft-"));

      if (isDraftTarget) {
        updateWorkbookFocusAnchor(null);
        const targetResolution = resolveDraftTimelinePasteTargets({
          columns: timelineAnchorColumnsRef.current,
          pastedColumnCount: dimensions.columnCount,
          pastedRowCount: dimensions.rowCount,
          startFieldKey: fieldKey,
          viewSchemaId: timelineViewSchemaId,
        });
        return targetResolution === null
          ? null
          : { anchor: null, targetResolution };
      }

      const recordId = row?.recordId;
      if (recordId === undefined || recordId === null) {
        return null;
      }

      const anchor = {
        fieldKey,
        rowIdentity: { kind: "core_record" as const, recordId },
        surface: {
          kind: "view_schema" as const,
          viewSchemaId: timelineViewSchemaId,
        },
      };
      const presentationRows = buildGridPresentationRows({
        grouping:
          groupBy === null
            ? null
            : {
                fieldKey: groupBy,
                formatLabel: (value) => (value === null ? null : String(value)),
                getValue: (candidate) => timelineGroupLabel(candidate, groupBy),
              },
        rows: timelineAnchorRowsRef.current,
      });
      const targetResolution = resolveGridPasteTargets({
        columns: timelineAnchorColumnsRef.current,
        current: anchor,
        pastedColumnCount: dimensions.columnCount,
        pastedRowCount: dimensions.rowCount,
        presentationRows,
      });
      if (targetResolution === null) {
        return null;
      }

      updateTimelineSurfaceFocusAnchor(
        anchor.rowIdentity.recordId,
        anchor.fieldKey,
      );
      return { anchor, targetResolution };
    },
    [
      groupBy,
      rowsRef,
      timelineAnchorColumnsRef,
      timelineAnchorRowsRef,
      updateTimelineSurfaceFocusAnchor,
      updateWorkbookFocusAnchor,
    ],
  );

  const navigateTimelineFocusAnchor = useCallback(
    (current: GridCellAnchor, intent: GridNavigationIntent) => {
      const nextAnchor = navigateGridCellAnchor({
        columns: timelineAnchorColumnsRef.current,
        current,
        intent,
        presentationRows: buildGridPresentationRows({
          grouping:
            groupBy === null
              ? null
              : {
                  fieldKey: groupBy,
                  formatLabel: (value) =>
                    value === null ? null : String(value),
                  getValue: (row) => timelineGroupLabel(row, groupBy),
                },
          rows: timelineAnchorRowsRef.current,
        }),
      });
      if (nextAnchor === null) {
        updateWorkbookFocusAnchor(null);
        return;
      }
      updateTimelineSurfaceFocusAnchor(
        nextAnchor.rowIdentity.kind === "core_record"
          ? nextAnchor.rowIdentity.recordId
          : null,
        nextAnchor.fieldKey,
      );
      const restoredNow = restoreTimelineFocusAnchor(nextAnchor);
      window.setTimeout(() => {
        if (restoredNow) {
          return;
        }
        restoreTimelineFocusAnchor(nextAnchor);
      }, 0);
    },
    [
      groupBy,
      restoreTimelineFocusAnchor,
      timelineAnchorColumnsRef,
      timelineAnchorRowsRef,
      updateTimelineSurfaceFocusAnchor,
      updateWorkbookFocusAnchor,
    ],
  );

  return {
    currentTimelineAnchorFor,
    navigateTimelineFocusAnchor,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
  };
}
