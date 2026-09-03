import type {
  GridCellAnchor,
  GridColumn,
  GridHandle,
} from "@cartulary/grid-adapter";
import {
  rowCellTestId,
  workbookFocusAnchorTestId,
} from "@cartulary/ui-contracts";
import {
  type ClipboardEvent as ReactClipboardEvent,
  type ReactNode,
  type RefObject,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  focusableCellStyle,
  visuallyHiddenStyle,
} from "../utils/workbookStyles";
import {
  createWorkbookContinuityPort,
  type WorkbookContinuityAnchor,
  type WorkbookContinuityPort,
} from "./workbookContinuityPort";

type GridViewportSnapshot = {
  readonly focusTarget: HTMLElement | null;
  readonly scrollLeft: number;
  readonly scrollTop: number;
};

export type WorkbookGridContinuityRuntime = {
  readonly port: WorkbookContinuityPort;
  readonly snapshot: {
    readonly anchor: WorkbookContinuityAnchor | null;
  };
};

export function useWorkbookGridContinuity<Row>({
  columns,
  continuityResetKey,
  gridHandleRef,
  selectionRef,
  viewSchemaId,
}: {
  readonly columns: readonly GridColumn<Row>[];
  readonly continuityResetKey: string;
  readonly gridHandleRef?: RefObject<GridHandle | null> | undefined;
  readonly selectionRef?:
    | { current: WorkbookContinuityAnchor | null }
    | undefined;
  readonly viewSchemaId: string;
}): WorkbookGridContinuityRuntime {
  const [anchor, setAnchor] = useState<WorkbookContinuityAnchor | null>(null);
  const columnsRef = useRef(columns);
  columnsRef.current = columns;
  const viewSchemaIdRef = useRef(viewSchemaId);
  viewSchemaIdRef.current = viewSchemaId;
  const continuityResetKeyRef = useRef(continuityResetKey);
  const port = useMemo(
    () =>
      createWorkbookContinuityPort({
        capture: () => {
          const scrollElement = gridHandleRef?.current?.getScrollElement();
          return {
            focusTarget:
              document.activeElement instanceof HTMLElement
                ? document.activeElement
                : null,
            scrollLeft: scrollElement?.scrollLeft ?? 0,
            scrollTop: scrollElement?.scrollTop ?? 0,
          } satisfies GridViewportSnapshot;
        },
        focus: (target) => {
          if (
            target.viewSchemaId !== viewSchemaIdRef.current ||
            !columnsRef.current.some(
              (column) => column.fieldKey === target.fieldKey,
            )
          ) {
            return false;
          }
          const gridAnchor = continuityGridAnchor(target);
          const focused =
            gridHandleRef?.current?.focusAnchor(gridAnchor) ?? false;
          if (!focused) {
            window.setTimeout(() => {
              gridHandleRef?.current?.focusAnchor(gridAnchor);
            }, 0);
          }
          return focused;
        },
        restore: (target, privateSnapshot) => {
          if (target === null) {
            return false;
          }
          if (
            target.viewSchemaId !== viewSchemaIdRef.current ||
            !columnsRef.current.some(
              (column) => column.fieldKey === target.fieldKey,
            )
          ) {
            return false;
          }
          const scrollElement = gridHandleRef?.current?.getScrollElement();
          if (
            scrollElement !== null &&
            scrollElement !== undefined &&
            isGridViewportSnapshot(privateSnapshot)
          ) {
            scrollElement.scrollLeft = privateSnapshot.scrollLeft;
            scrollElement.scrollTop = privateSnapshot.scrollTop;
          }
          if (
            isGridViewportSnapshot(privateSnapshot) &&
            privateSnapshot.focusTarget?.isConnected
          ) {
            privateSnapshot.focusTarget.focus({ preventScroll: true });
            return document.activeElement === privateSnapshot.focusTarget;
          }
          const gridAnchor = continuityGridAnchor(target);
          gridHandleRef?.current?.scrollToAnchor(gridAnchor);
          return gridHandleRef?.current?.focusAnchor(gridAnchor) ?? false;
        },
        select: (target) => {
          if (selectionRef !== undefined) {
            selectionRef.current = target;
          }
          setAnchor(target);
        },
      }),
    [gridHandleRef, selectionRef],
  );

  useEffect(() => () => port.dispose(), [port]);

  useEffect(() => {
    if (continuityResetKeyRef.current === continuityResetKey) {
      return;
    }
    continuityResetKeyRef.current = continuityResetKey;
    port.clear();
  }, [continuityResetKey, port]);

  useEffect(() => {
    if (anchor !== null && anchor.viewSchemaId !== viewSchemaIdRef.current) {
      port.clear();
    }
  }, [anchor, port]);

  return { port, snapshot: { anchor } };
}

export function WorkbookContinuityCell({
  children,
  continuity,
  fieldKey,
  onPaste,
  recordId,
  viewSchemaId,
}: {
  readonly children: ReactNode;
  readonly continuity: WorkbookContinuityPort;
  readonly fieldKey: string;
  readonly onPaste?: (
    event: ReactClipboardEvent<HTMLElement>,
    anchor: { readonly fieldKey: string; readonly recordId: string },
  ) => void;
  readonly recordId: string;
  readonly viewSchemaId: string;
}) {
  return (
    // biome-ignore lint/a11y: The grid adapter owns the gridcell role; this wrapper is only the focus/paste anchor inside that cell.
    <span
      data-testid={rowCellTestId(recordId, fieldKey)}
      onFocus={() => {
        continuity.select({ fieldKey, recordId, viewSchemaId });
      }}
      onCopy={(event) => {
        event.clipboardData.setData(
          "text/plain",
          event.currentTarget.textContent ?? "",
        );
        event.preventDefault();
        continuity.select({ fieldKey, recordId, viewSchemaId });
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

export function WorkbookContinuityAnchorStatus({
  anchor,
}: {
  readonly anchor: WorkbookContinuityAnchor | null;
}) {
  return (
    <span data-testid={workbookFocusAnchorTestId()} style={visuallyHiddenStyle}>
      {anchor === null
        ? "cleared"
        : `${anchor.viewSchemaId}:${anchor.recordId}:${anchor.fieldKey}`}
    </span>
  );
}

function continuityGridAnchor(
  anchor: WorkbookContinuityAnchor,
): GridCellAnchor {
  return {
    fieldKey: anchor.fieldKey,
    rowIdentity: { kind: "core_record", recordId: anchor.recordId },
    surface: {
      kind: "view_schema",
      viewSchemaId: anchor.viewSchemaId,
    },
  };
}

function isGridViewportSnapshot(value: unknown): value is GridViewportSnapshot {
  if (value === null || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  if (
    !("focusTarget" in value) ||
    !("scrollLeft" in value) ||
    !("scrollTop" in value)
  ) {
    return false;
  }
  return (
    (value.focusTarget === null || value.focusTarget instanceof HTMLElement) &&
    typeof value.scrollLeft === "number" &&
    typeof value.scrollTop === "number"
  );
}
