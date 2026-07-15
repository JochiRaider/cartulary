import {
  buildGridPresentationRows,
  type GridCellAnchor,
  type GridColumn,
  type GridGroupingDescriptor,
  type GridHandle,
  type GridNavigationIntent,
  type GridRecordRow,
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
  type RefObject,
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
  readonly rows: readonly GridRecordRow<Row>[];
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
      setAnchor({ ...nextAnchor, surface });
      window.setTimeout(() => {
        gridHandleRef?.current?.focusAnchor(nextAnchor);
        const element = document.querySelector<HTMLElement>(
          dataTestIdSelector(
            rowCellTestId(nextAnchor.recordId, nextAnchor.fieldKey),
          ),
        );
        element?.focus({ preventScroll: true });
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
      onKeyDown={(event: ReactKeyboardEvent<HTMLElement>) => {
        const command = mapWorkbookKeyboardCommand(event);
        if (command.preventDefault) {
          event.preventDefault();
          // The semantic cell focus runtime owns these navigation commands.
          // Keep RDG's root key handler from also moving its presentation
          // selection from the independently focused header cell.
          event.stopPropagation();
        }
        if (command.kind === "navigate") {
          focus.navigate(
            { fieldKey, recordId, viewSchemaId: focus.viewSchemaId },
            command.intent,
          );
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
