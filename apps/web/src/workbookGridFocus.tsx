import {
  buildGridPresentationRows,
  type GridCellAnchor,
  type GridColumn,
  type GridNavigationIntent,
  type GridRow,
  navigateGridCellAnchor,
} from "@cartulary/grid-adapter";
import {
  dataTestIdSelector,
  rowCellTestId,
  type WorkbookSurface,
} from "@cartulary/ui-contracts";
import {
  type ClipboardEvent as ReactClipboardEvent,
  type KeyboardEvent as ReactKeyboardEvent,
  type ReactNode,
  useCallback,
  useState,
} from "react";
import { mapWorkbookKeyboardCommand } from "./workbookKeyboard";
import { focusableCellStyle, visuallyHiddenStyle } from "./workbookStyles";

export type WorkbookFocusAnchor = GridCellAnchor & {
  readonly surface: WorkbookSurface;
};

export type WorkbookGridFocusRuntime = {
  readonly anchor: WorkbookFocusAnchor | null;
  readonly navigate: (
    current: GridCellAnchor,
    intent: GridNavigationIntent,
  ) => void;
  readonly update: (recordId: string | null, fieldKey: string) => void;
};

export function formatWorkbookFocusAnchor(anchor: WorkbookFocusAnchor | null) {
  return anchor === null
    ? "cleared"
    : `${anchor.surface}:${anchor.recordId}:${anchor.fieldKey}`;
}

export function useWorkbookGridFocus<Row>({
  columns,
  getGroupLabel,
  groupBy,
  rows,
  surface,
}: {
  readonly columns: readonly GridColumn<Row>[];
  readonly getGroupLabel?: (
    row: Row,
    fieldKey: string,
  ) => string | null | undefined;
  readonly groupBy?: string | null | undefined;
  readonly rows: readonly GridRow<Row>[];
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
      setAnchor({ fieldKey, recordId, surface });
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
          getGroupLabel,
          groupBy,
          rows,
        }),
      });
      if (nextAnchor === null) {
        setAnchor(null);
        return;
      }
      setAnchor({ ...nextAnchor, surface });
      window.setTimeout(() => {
        const element = document.querySelector<HTMLElement>(
          dataTestIdSelector(
            rowCellTestId(nextAnchor.recordId, nextAnchor.fieldKey),
          ),
        );
        element?.focus({ preventScroll: true });
      }, 0);
    },
    [columns, getGroupLabel, groupBy, rows, surface],
  );

  return { anchor, navigate, update };
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
      onKeyDown={(event: ReactKeyboardEvent<HTMLElement>) => {
        const command = mapWorkbookKeyboardCommand(event);
        if (command.preventDefault) {
          event.preventDefault();
        }
        if (command.kind === "navigate") {
          focus.navigate({ fieldKey, recordId }, command.intent);
        }
      }}
      style={focusableCellStyle}
      // biome-ignore lint/a11y/noNoninteractiveTabindex: The grid adapter owns cell semantics; this wrapper keeps keyboard focus anchored inside the rendered cell.
      tabIndex={0}
    >
      {children}
    </span>
  );
}
