import type {
  GridCellAnchor,
  GridClipboardInput,
  GridColumn,
  GridHandle,
  GridNavigationIntent,
  GridPasteTargetResolution,
} from "@cartulary/grid-adapter";
import { useCallback } from "react";
import type {
  WorkbookContinuityAnchor,
  WorkbookContinuityPort,
} from "../../continuity/workbookContinuityPort";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelinePasteTargetResolution } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/workbookTimelineModel";

type TimelineReadonlyRef<T> = {
  readonly current: T;
};

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
  continuityPort,
  gridHandleRef,
  rowsRef,
  timelineAnchorColumnsRef,
  updateTimelineSurfaceFocusAnchor,
  updateWorkbookFocusAnchor,
}: {
  readonly continuityPort: WorkbookContinuityPort;
  readonly gridHandleRef: TimelineReadonlyRef<GridHandle | null>;
  readonly rowsRef: TimelineReadonlyRef<readonly WorkbookRow[]>;
  readonly timelineAnchorColumnsRef: TimelineReadonlyRef<
    readonly GridColumn<WorkbookRow>[]
  >;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
  readonly updateWorkbookFocusAnchor: (anchor: null) => void;
}) {
  const restoreTimelineFocusAnchor = useCallback(
    (anchor: GridCellAnchor | WorkbookContinuityAnchor): boolean => {
      const semanticAnchor: WorkbookContinuityAnchor | null =
        "rowIdentity" in anchor
          ? anchor.rowIdentity.kind === "core_record"
            ? {
                fieldKey: anchor.fieldKey,
                recordId: anchor.rowIdentity.recordId,
                viewSchemaId:
                  anchor.surface.kind === "view_schema"
                    ? anchor.surface.viewSchemaId
                    : timelineViewSchemaId,
              }
            : null
          : anchor;
      return semanticAnchor === null
        ? gridHandleRef.current?.focusAnchor(anchor as GridCellAnchor) === true
        : continuityPort.focus(semanticAnchor);
    },
    [continuityPort, gridHandleRef],
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
      input: GridClipboardInput,
    ): TimelinePasteTargetResolution | null => {
      if (input.kind !== "table") {
        return null;
      }

      const dimensions = {
        columnCount: input.values[0]?.length ?? 0,
        rowCount: input.values.length,
      };
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
      const targetResolution =
        gridHandleRef.current?.planPasteTargets(anchor, dimensions) ?? null;
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
      gridHandleRef,
      rowsRef,
      timelineAnchorColumnsRef,
      updateTimelineSurfaceFocusAnchor,
      updateWorkbookFocusAnchor,
    ],
  );

  const navigateTimelineFocusAnchor = useCallback(
    (current: GridCellAnchor, intent: GridNavigationIntent) => {
      const nextAnchor =
        gridHandleRef.current?.moveFocus(current, intent) ?? null;
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
    },
    [
      gridHandleRef,
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
