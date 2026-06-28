import {
  buildGridPresentationRows,
  type GridCellAnchor,
  type GridColumn,
  type GridNavigationIntent,
  type GridPasteTargetResolution,
  type GridRow,
  navigateGridCellAnchor,
  resolveGridPasteTargets,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  gridRowGutterTestId,
  rowCellTestId,
  timelineCollectionInputTestId,
  timelineRowVersionTestId,
} from "@cartulary/ui-contracts";
import { useCallback } from "react";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  clipboardGridDimensions,
  clipboardTextLooksTabular,
} from "../../utils/workbookClipboard";
import {
  inputFocusKey,
  timelineFieldBinding,
  timelineFocusFieldForFieldKey,
  timelineGroupLabel,
  type WorkbookRow,
} from "../models/workbookTimelineModel";

type TimelineReadonlyRef<T> = {
  readonly current: T;
};

export type TimelinePasteTargetResolution = {
  readonly anchor: GridCellAnchor | null;
  readonly targetResolution: GridPasteTargetResolution;
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
}: {
  readonly columns: readonly GridColumn<WorkbookRow>[];
  readonly pastedColumnCount: number;
  readonly pastedRowCount: number;
  readonly startFieldKey: string;
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
    })),
  };
}

export function useTimelineGridAnchorController({
  groupBy,
  resolveInputElement,
  rowsRef,
  timelineAnchorColumnsRef,
  timelineAnchorRowsRef,
  updateTimelineSurfaceFocusAnchor,
  updateWorkbookFocusAnchor,
}: {
  readonly groupBy: string | null;
  readonly resolveInputElement: (focusKey: string) => HTMLElement | null;
  readonly rowsRef: TimelineReadonlyRef<readonly WorkbookRow[]>;
  readonly timelineAnchorColumnsRef: TimelineReadonlyRef<
    readonly GridColumn<WorkbookRow>[]
  >;
  readonly timelineAnchorRowsRef: TimelineReadonlyRef<
    readonly GridRow<WorkbookRow>[]
  >;
  readonly updateTimelineSurfaceFocusAnchor: (
    recordId: string | null,
    fieldKey: string,
  ) => void;
  readonly updateWorkbookFocusAnchor: (anchor: null) => void;
}) {
  const resolveTimelineAnchorElement = useCallback(
    (anchor: GridCellAnchor) => {
      const focusField = timelineFocusFieldForFieldKey(anchor.fieldKey);
      if (focusField !== null) {
        const inputElement = resolveInputElement(
          inputFocusKey(anchor.recordId, focusField, "grid"),
        );
        if (inputElement !== null) {
          return inputElement;
        }
      }
      const binding = timelineFieldBinding(anchor.fieldKey);
      if (binding.kind === "collection") {
        const collectionInput = document.querySelector<HTMLInputElement>(
          dataTestIdSelector(
            timelineCollectionInputTestId(anchor.recordId, anchor.fieldKey),
          ),
        );
        if (collectionInput !== null) {
          return collectionInput;
        }
      }
      const testId =
        anchor.fieldKey === "timeline.capture_state"
          ? gridRowGutterTestId(timelineViewSchemaId, anchor.recordId)
          : anchor.fieldKey === "row_version"
            ? timelineRowVersionTestId(anchor.recordId)
            : rowCellTestId(anchor.recordId, anchor.fieldKey);
      return document.querySelector<HTMLElement>(dataTestIdSelector(testId));
    },
    [resolveInputElement],
  );

  const restoreTimelineFocusAnchor = useCallback(
    (anchor: GridCellAnchor) => {
      const element = resolveTimelineAnchorElement(anchor);
      if (element === null) {
        return false;
      }
      if (!element.hasAttribute("tabindex")) {
        element.tabIndex = -1;
      }
      element.focus({ preventScroll: true });
      return document.activeElement === element;
    },
    [resolveTimelineAnchorElement],
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
        recordId: row.recordId,
      };
      updateTimelineSurfaceFocusAnchor(anchor.recordId, anchor.fieldKey);
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
        recordId,
      };
      const presentationRows = buildGridPresentationRows({
        getGroupLabel: (candidate, groupFieldKey) =>
          timelineGroupLabel(candidate, groupFieldKey),
        groupBy,
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

      updateTimelineSurfaceFocusAnchor(anchor.recordId, anchor.fieldKey);
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
          getGroupLabel: (row, fieldKey) => timelineGroupLabel(row, fieldKey),
          groupBy,
          rows: timelineAnchorRowsRef.current,
        }),
      });
      if (nextAnchor === null) {
        updateWorkbookFocusAnchor(null);
        return;
      }
      updateTimelineSurfaceFocusAnchor(
        nextAnchor.recordId,
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
    resolveTimelineAnchorElement,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
  };
}
