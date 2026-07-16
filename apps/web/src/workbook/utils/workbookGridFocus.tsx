import {
  buildGridPresentationRows,
  type GridCellAnchor,
  type GridColumn,
  type GridDataRow,
  type GridGroupingDescriptor,
  type GridHandle,
  type GridNavigationIntent,
  navigateGridCellAnchor,
} from "@cartulary/grid-adapter";
import { rowCellTestId, type WorkbookSurface } from "@cartulary/ui-contracts";
import {
  type ClipboardEvent as ReactClipboardEvent,
  type ReactNode,
  type RefObject,
  useCallback,
  useState,
} from "react";
import { focusableCellStyle, visuallyHiddenStyle } from "./workbookStyles";

export type WorkbookFocusAnchor = {
  readonly fieldKey: string;
  readonly recordId: string;
  readonly surface: WorkbookSurface;
  readonly viewSchemaId: string;
};

export type WorkbookGridFocusRuntime = {
  readonly anchor: WorkbookFocusAnchor | null;
  readonly navigate: (
    current: GridCellAnchor,
    intent: GridNavigationIntent,
  ) => void;
  readonly update: (recordId: string | null, fieldKey: string) => void;
  readonly viewSchemaId: string;
};

function formatWorkbookFocusAnchor(anchor: WorkbookFocusAnchor | null) {
  return anchor === null
    ? "cleared"
    : `${anchor.surface}:${anchor.recordId}:${anchor.fieldKey}`;
}

export function useWorkbookGridFocus<Row>({
  columns,
  grouping,
  rows,
  surface,
  gridHandleRef,
}: {
  readonly columns: readonly GridColumn<Row>[];
  readonly gridHandleRef?: RefObject<GridHandle | null> | undefined;
  readonly grouping?: GridGroupingDescriptor<Row> | null | undefined;
  readonly rows: readonly GridDataRow<Row>[];
  readonly surface: WorkbookSurface;
}): WorkbookGridFocusRuntime {
  const [anchor, setAnchor] = useState<WorkbookFocusAnchor | null>(null);

  const update = useCallback(
    (recordId: string | null, fieldKey: string) => {
      if (
        recordId === null ||
        recordId.trim() === "" ||
        !columns.some((column) => column.fieldKey === fieldKey)
      ) {
        setAnchor(null);
        return;
      }
      setAnchor({ fieldKey, recordId, surface, viewSchemaId: surface });
    },
    [columns, surface],
  );

  const navigate = useCallback(
    (current: GridCellAnchor, intent: GridNavigationIntent) => {
      const nextAnchor = navigateGridCellAnchor({
        columns,
        current,
        intent,
        presentationRows: buildGridPresentationRows({
          grouping,
          rows,
        }),
      });
      if (nextAnchor === null) {
        setAnchor(null);
        return;
      }
      if (nextAnchor.rowIdentity.kind !== "core_record") {
        setAnchor(null);
        return;
      }
      setAnchor({
        fieldKey: nextAnchor.fieldKey,
        recordId: nextAnchor.rowIdentity.recordId,
        surface,
        viewSchemaId: surface,
      });
      window.setTimeout(() => {
        gridHandleRef?.current?.focusAnchor(nextAnchor);
      }, 0);
    },
    [columns, gridHandleRef, grouping, rows, surface],
  );

  return { anchor, navigate, update, viewSchemaId: surface };
}

export function WorkbookFocusAnchorStatus({
  anchor,
}: {
  readonly anchor: WorkbookFocusAnchor | null;
}) {
  return (
    <span data-testid="workbook-focus-anchor" style={visuallyHiddenStyle}>
      {formatWorkbookFocusAnchor(anchor)}
    </span>
  );
}

export function FocusableWorkbookCell({
  children,
  fieldKey,
  focus,
  onPaste,
  recordId,
}: {
  readonly children: ReactNode;
  readonly fieldKey: string;
  readonly focus: WorkbookGridFocusRuntime;
  readonly onPaste?: (
    event: ReactClipboardEvent<HTMLElement>,
    anchor: { readonly fieldKey: string; readonly recordId: string },
  ) => void;
  readonly recordId: string;
}) {
  return (
    // biome-ignore lint/a11y: The grid adapter owns the gridcell role; this wrapper is only the focus/paste anchor inside that cell.
    <span
      data-testid={rowCellTestId(recordId, fieldKey)}
      onFocus={() => {
        focus.update(recordId, fieldKey);
      }}
      onCopy={(event) => {
        event.clipboardData.setData(
          "text/plain",
          event.currentTarget.textContent ?? "",
        );
        event.preventDefault();
        focus.update(recordId, fieldKey);
      }}
      onPaste={
        onPaste
          ? (event) => {
              onPaste(event, { fieldKey, recordId });
            }
          : undefined
      }
      style={focusableCellStyle}
    >
      {children}
    </span>
  );
}
