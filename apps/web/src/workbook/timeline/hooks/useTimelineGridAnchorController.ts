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
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelinePasteTargetResolution } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/timelineRowModel";

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
  mappedFields,
  pastedRowCount,
  startFieldKey,
  viewSchemaId,
}: {
  readonly columns: readonly GridColumn<WorkbookRow>[];
  readonly mappedFields?: readonly string[] | undefined;
  readonly pastedColumnCount: number;
  readonly pastedRowCount: number;
  readonly startFieldKey: string;
  readonly viewSchemaId: string;
}): GridPasteTargetResolution | null {
  const targetColumns =
    mappedFields ??
    timelinePasteColumnsFromStart(columns, startFieldKey, pastedColumnCount);
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

type TimelinePasteDimensions = {
  readonly fieldKeys?: readonly string[] | undefined;
  readonly columnCount: number;
  readonly rowCount: number;
};

function timelinePasteDimensions(
  input: GridClipboardInput,
): TimelinePasteDimensions | null {
  return input.kind === "table"
    ? {
        fieldKeys: input.fieldKeys,
        columnCount: input.values[0]?.length ?? 0,
        rowCount: input.values.length,
      }
    : null;
}

function savedTimelinePasteResolution({
  dimensions,
  fieldKey,
  gridHandle,
  recordId,
}: {
  readonly dimensions: TimelinePasteDimensions;
  readonly fieldKey: string;
  readonly gridHandle: GridHandle | null;
  readonly recordId: string;
}): TimelinePasteTargetResolution | null {
  const anchor: GridCellAnchor = {
    fieldKey,
    rowIdentity: { kind: "core_record", recordId },
    surface: { kind: "view_schema", viewSchemaId: timelineViewSchemaId },
  };
  const targetResolution =
    gridHandle?.planPasteTargets(anchor, dimensions) ?? null;
  return targetResolution === null ? null : { anchor, targetResolution };
}

export function useTimelineGridAnchorController({
  continuityPort,
  editorDraftRegistry,
  gridHandleRef,
  rowsRef,
  timelineAnchorColumnsRef,
  updateTimelineSurfaceFocusAnchor,
  updateWorkbookFocusAnchor,
}: {
  readonly continuityPort: WorkbookContinuityPort;
  readonly editorDraftRegistry: TimelineEditorDraftRegistry;
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
    async (
      anchor: GridCellAnchor | WorkbookContinuityAnchor,
      options?: { readonly presentation: "cell" },
    ): Promise<boolean> => {
      if (!("rowIdentity" in anchor)) {
        return continuityPort.focus(anchor);
      }
      if (
        options?.presentation === "cell" ||
        anchor.rowIdentity.kind !== "core_record"
      ) {
        return (
          (await gridHandleRef.current?.requestFocus({
            kind: "cell",
            anchor,
          })) === "focused"
        );
      }
      return continuityPort.focus({
        fieldKey: anchor.fieldKey,
        recordId: anchor.rowIdentity.recordId,
        viewSchemaId:
          anchor.surface.kind === "view_schema"
            ? anchor.surface.viewSchemaId
            : timelineViewSchemaId,
      });
    },
    [continuityPort, gridHandleRef],
  );

  const currentTimelineAnchorFor = useCallback(
    (rowKey: string, fieldKey: string): GridCellAnchor | null => {
      const row = rowsRef.current.find(
        (candidate) =>
          candidate.key === editorDraftRegistry.resolveRowKey(rowKey),
      );
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
    [
      rowsRef,
      editorDraftRegistry,
      updateTimelineSurfaceFocusAnchor,
      updateWorkbookFocusAnchor,
    ],
  );

  const resolveTimelinePasteTargetResolution = useCallback(
    (
      rowKey: string,
      fieldKey: string,
      input: GridClipboardInput,
    ): TimelinePasteTargetResolution | null => {
      const dimensions = timelinePasteDimensions(input);
      if (dimensions === null) return null;
      const row = rowsRef.current.find((candidate) => candidate.key === rowKey);
      const isDraftTarget =
        row?.recordId === null ||
        (row === undefined && rowKey.startsWith("draft-"));

      if (isDraftTarget) {
        updateWorkbookFocusAnchor(null);
        const targetResolution = resolveDraftTimelinePasteTargets({
          columns: timelineAnchorColumnsRef.current,
          mappedFields: dimensions.fieldKeys,
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
      if (recordId === undefined || recordId === null) return null;
      const resolution = savedTimelinePasteResolution({
        dimensions,
        fieldKey,
        gridHandle: gridHandleRef.current,
        recordId,
      });
      if (resolution === null) return null;

      updateTimelineSurfaceFocusAnchor(recordId, fieldKey);
      return resolution;
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

  const prepareTimelineCollectionNavigation = useCallback(
    (rowKey: string, fieldKey: string, intent: GridNavigationIntent) => {
      const handle = gridHandleRef.current;
      const anchor = currentTimelineAnchorFor(rowKey, fieldKey);
      if (anchor) return handle?.prepareNavigation?.(anchor, intent) ?? null;
      const fields = timelineAnchorColumnsRef.current
        .filter(
          (column) =>
            column.renderDraftCell !== undefined &&
            column.draftWritable === true,
        )
        .map((column) => column.fieldKey);
      const previous = rowsRef.current
        .filter((row) => row.recordId !== null)
        .at(-1);
      const next =
        intent.key === "Tab"
          ? fields[fields.indexOf(fieldKey) + (intent.shiftKey ? -1 : 1)]
          : intent.key === "Enter" && !intent.shiftKey
            ? fieldKey
            : undefined;
      const previousField =
        intent.key === "Tab"
          ? timelineAnchorColumnsRef.current.at(-1)?.fieldKey
          : fieldKey;
      return () => {
        if (!handle || !gridHandleRef.current) return;
        if (next !== undefined) {
          if (
            timelineAnchorColumnsRef.current.some(
              (column) => column.fieldKey === next && column.draftWritable,
            )
          )
            void handle.requestFocus({ kind: "draft", fieldKey: next });
        } else if (intent.shiftKey && previous?.recordId && previousField) {
          if (rowsRef.current.some((row) => row.recordId === previous.recordId))
            void handle.requestFocus({
              kind: "cell",
              anchor: {
                surface: {
                  kind: "view_schema",
                  viewSchemaId: timelineViewSchemaId,
                },
                rowIdentity: {
                  kind: "core_record",
                  recordId: previous.recordId,
                },
                fieldKey: previousField,
              },
            });
        } else if (intent.key === "Tab")
          handle.focusAdjacentRegion?.(intent.shiftKey === true);
      };
    },
    [
      currentTimelineAnchorFor,
      gridHandleRef,
      rowsRef,
      timelineAnchorColumnsRef,
    ],
  );
  const navigateTimelineDraftFocus = useCallback(
    (rowKey: string, fieldKey: string, intent: GridNavigationIntent) => {
      const committed = currentTimelineAnchorFor(rowKey, fieldKey);
      if (committed !== null) {
        gridHandleRef.current?.moveFocus(committed, intent);
        return;
      }
      prepareTimelineCollectionNavigation(rowKey, fieldKey, intent)?.();
    },
    [
      currentTimelineAnchorFor,
      gridHandleRef,
      prepareTimelineCollectionNavigation,
    ],
  );
  return {
    prepareTimelineCollectionNavigation,
    navigateTimelineDraftFocus,
    currentTimelineAnchorFor,
    navigateTimelineFocusAnchor,
    resolveTimelinePasteTargetResolution,
    restoreTimelineFocusAnchor,
  };
}
