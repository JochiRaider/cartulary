import type {
  GridCellAnchor,
  GridFocusTarget,
  GridHandle,
} from "@cartulary/grid-adapter";
import {
  type KeyboardEvent as ReactKeyboardEvent,
  type MouseEvent as ReactMouseEvent,
  type PointerEvent as ReactPointerEvent,
  type RefObject,
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { focusWorkbookShellFallback } from "../../components/workbookFocusFallback";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineRowContextMenuPosition } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/timelineRowModel";

type Invocation = {
  readonly anchor: GridCellAnchor & {
    readonly rowIdentity: {
      readonly kind: "core_record";
      readonly recordId: string;
    };
  };
  readonly position: TimelineRowContextMenuPosition;
  readonly scroll: { readonly top: number; readonly left: number };
  readonly scopeKey: string;
};

/** Timeline's row-action admission, independent of Inspector selection. */
export function useTimelineRowActionMenu({
  gridHandleRef,
  readable,
  rows,
  scopeKey,
}: {
  readonly gridHandleRef: RefObject<GridHandle | null>;
  readonly readable: boolean;
  readonly rows: readonly WorkbookRow[];
  readonly scopeKey: string;
}) {
  const [invocation, setInvocation] = useState<Invocation | null>(null);
  const menuRef = useRef<HTMLDivElement | null>(null);
  const menuOwnedFocus = useRef(false);
  const pendingFocus = useRef<{
    controller: AbortController;
    scopeKey: string;
  } | null>(null);
  const current = useRef({ readable, rows, scopeKey });
  current.current = { readable, rows, scopeKey };

  const ownsFocus = useCallback(
    () =>
      menuRef.current?.contains(document.activeElement) === true ||
      (menuOwnedFocus.current && document.activeElement === document.body),
    [],
  );
  const cancelFocus = useCallback(() => {
    pendingFocus.current?.controller.abort();
    pendingFocus.current = null;
  }, []);

  const restoreFocus = useCallback(
    async (source: Invocation | null) => {
      cancelFocus();
      const state = current.current;
      if (source && source.scopeKey !== state.scopeKey) return;
      const controller = new AbortController();
      pendingFocus.current = { controller, scopeKey: state.scopeKey };
      const grid = gridHandleRef.current;
      const options = { signal: controller.signal, preserveSelection: true };
      let destination: GridFocusTarget | null = null;
      // Own the entire chain, including the gaps between adapter requests.
      const cancel = () => controller.abort();
      const externalFocus = (event: FocusEvent) => {
        if (
          destination?.kind === "root" &&
          event.target === grid?.getScrollElement()
        )
          return;
        if (destination?.kind === "cell" && event.target instanceof Element) {
          const cell = semanticCell(event.target);
          if (
            cell.fieldKey === destination.anchor.fieldKey &&
            destination.anchor.rowIdentity.kind === "core_record" &&
            cell.recordId === destination.anchor.rowIdentity.recordId
          )
            return;
        }
        cancel();
      };
      document.addEventListener("focusin", externalFocus, true);
      document.addEventListener("pointerdown", cancel, true);
      document.addEventListener("keydown", cancel, true);
      try {
        if (grid && state.readable) {
          const presented = grid.presentation?.getSnapshot();
          if (
            source &&
            state.rows.some(
              (row) => row.recordId === source.anchor.rowIdentity.recordId,
            ) &&
            presented?.rowIdentities.some(
              (row) =>
                row.kind === "core_record" &&
                row.recordId === source.anchor.rowIdentity.recordId,
            )
          ) {
            const fieldKey = presented.fieldKeys.includes(
              source.anchor.fieldKey,
            )
              ? source.anchor.fieldKey
              : presented.fieldKeys[0];
            if (fieldKey) {
              destination = {
                kind: "cell",
                anchor: { ...source.anchor, fieldKey },
              };
              const result = await grid.requestFocus(destination, options);
              if (result !== "unavailable") return;
            }
          }
          if (controller.signal.aborted) return;
          destination = { kind: "root" };
          const result = await grid.requestFocus(destination, options);
          if (result !== "unavailable") return;
        }
        if (!controller.signal.aborted) focusWorkbookShellFallback();
      } finally {
        document.removeEventListener("focusin", externalFocus, true);
        document.removeEventListener("pointerdown", cancel, true);
        document.removeEventListener("keydown", cancel, true);
        if (pendingFocus.current?.controller === controller)
          pendingFocus.current = null;
      }
    },
    [cancelFocus, gridHandleRef],
  );

  const close = useCallback(() => {
    menuOwnedFocus.current = false;
    setInvocation(null);
  }, []);
  const dismissToGrid = useCallback(() => {
    const restore = ownsFocus();
    close();
    if (restore) void restoreFocus(null);
  }, [close, ownsFocus, restoreFocus]);

  // Both event paths use this one admission decision. The work-area can also
  // contain Inspector controls and React portals; ancestry alone is insufficient.
  const admit = useCallback(
    (
      event:
        | ReactMouseEvent<HTMLDivElement>
        | ReactKeyboardEvent<HTMLDivElement>,
    ) => {
      if (
        event.defaultPrevented ||
        !current.current.readable ||
        !(event.target instanceof Element)
      )
        return null;
      if ("isComposing" in event.nativeEvent && event.nativeEvent.isComposing)
        return null;
      const target = event.target;
      if (
        target.closest(
          "input, textarea, select, button, a, summary, [contenteditable], [popover], [role='alertdialog'], [role='menuitem'], [role='menuitemcheckbox'], [role='menuitemradio'], [role='spinbutton'], [role='slider'], [role='switch'], [role='tab'], [role='button'], [role='checkbox'], [role='link'], [role='textbox'], [role='combobox'], [role='menu'], [role='dialog'], [role='listbox'], [role='option'], [data-grid-editor-interaction], [data-grid-editor-toolbar], [data-grid-prevent-cell-edit='true']",
        )
      )
        return null;
      const grid = gridHandleRef.current;
      if (!grid?.getScrollElement()?.contains(target)) return null;
      const { recordId, fieldKey } = semanticCell(target);
      const presented = grid.presentation?.getSnapshot();
      if (
        !recordId ||
        !fieldKey ||
        !presented?.fieldKeys.includes(fieldKey) ||
        !presented.rowIdentities.some(
          (row) => row.kind === "core_record" && row.recordId === recordId,
        ) ||
        !current.current.rows.some((row) => row.recordId === recordId)
      )
        return null;
      return {
        fieldKey,
        rowIdentity: { kind: "core_record" as const, recordId },
        surface: {
          kind: "view_schema" as const,
          viewSchemaId: timelineViewSchemaId,
        },
      };
    },
    [gridHandleRef],
  );
  const open = useCallback(
    (
      event:
        | ReactMouseEvent<HTMLDivElement>
        | ReactKeyboardEvent<HTMLDivElement>,
      position?: TimelineRowContextMenuPosition,
    ) => {
      const anchor = admit(event);
      if (!anchor) return;
      const rect = gridHandleRef.current?.getAnchorRect(anchor);
      if (!position && !rect) return;
      event.preventDefault();
      event.stopPropagation();
      cancelFocus();
      const scroll = gridHandleRef.current?.getScrollElement();
      setInvocation({
        anchor,
        scroll: { top: scroll?.scrollTop ?? 0, left: scroll?.scrollLeft ?? 0 },
        scopeKey: current.current.scopeKey,
        position: position ?? {
          x: (rect?.left ?? 0) + 12,
          y: (rect?.top ?? 0) + 12,
        },
      });
    },
    [admit, cancelFocus, gridHandleRef],
  );
  const handleTimelineGridPointerDown = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
      // Borrow focus directly from authoring when contextmenu arrives. The
      // browser's intermediate cell focus would otherwise accept a departure.
      if (event.button === 2 && admit(event)) event.preventDefault();
    },
    [admit],
  );
  const handleTimelineGridContextMenu = useCallback(
    (event: ReactMouseEvent<HTMLDivElement>) => {
      open(event, { x: event.clientX, y: event.clientY });
    },
    [open],
  );
  const handleTimelineGridContextKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLDivElement>) => {
      if (event.altKey || event.ctrlKey || event.metaKey) return;
      if (
        (event.key === "ContextMenu" && !event.shiftKey) ||
        (event.key === "F10" && event.shiftKey)
      )
        open(event);
    },
    [open],
  );

  const onRestoreFocus = useCallback(() => {
    if (invocation) void restoreFocus(invocation);
  }, [invocation, restoreFocus]);
  const onFocusChange = useCallback((owned: boolean) => {
    menuOwnedFocus.current = owned;
  }, []);
  const onExternalScroll = useCallback(
    (event: Event) => {
      const grid = gridHandleRef.current?.getScrollElement();
      // Native inputs can reset their own text scroll offset when focus is
      // borrowed. A descendant scroll does not move the invoking grid cell.
      if (
        event.target instanceof Node &&
        event.target !== grid &&
        grid?.contains(event.target)
      )
        return;
      // A cell can be revealed before contextmenu but deliver scroll afterward.
      if (
        invocation &&
        event.target === grid &&
        grid?.scrollTop === invocation.scroll.top &&
        grid.scrollLeft === invocation.scroll.left
      )
        return;
      dismissToGrid();
    },
    [invocation, gridHandleRef, dismissToGrid],
  );

  useLayoutEffect(() => {
    if (
      pendingFocus.current &&
      (!readable || pendingFocus.current.scopeKey !== scopeKey)
    )
      cancelFocus();
  }, [scopeKey, readable, cancelFocus]);
  useLayoutEffect(() => {
    if (!invocation) return;
    const validate = () => {
      const presented = gridHandleRef.current?.presentation?.getSnapshot();
      if (
        !readable ||
        scopeKey !== invocation.scopeKey ||
        !rows.some(
          (row) => row.recordId === invocation.anchor.rowIdentity.recordId,
        ) ||
        !presented?.fieldKeys.includes(invocation.anchor.fieldKey) ||
        !presented.rowIdentities.some(
          (row) =>
            row.kind === "core_record" &&
            row.recordId === invocation.anchor.rowIdentity.recordId,
        )
      ) {
        cancelFocus();
        const restore = ownsFocus();
        close();
        if (restore)
          void restoreFocus(
            scopeKey === invocation.scopeKey ? invocation : null,
          );
      }
    };
    validate();
    return gridHandleRef.current?.presentation?.subscribe(validate);
  }, [
    invocation,
    rows,
    readable,
    scopeKey,
    gridHandleRef,
    cancelFocus,
    close,
    ownsFocus,
    restoreFocus,
  ]);
  useLayoutEffect(
    () => () => {
      cancelFocus();
      if (ownsFocus()) focusWorkbookShellFallback();
    },
    [cancelFocus, ownsFocus],
  );

  const row =
    invocation && readable && invocation.scopeKey === scopeKey
      ? (rows.find(
          (row) => row.recordId === invocation.anchor.rowIdentity.recordId,
        ) ?? null)
      : null;
  return {
    commands: {
      handleTimelineGridPointerDown,
      handleTimelineGridContextMenu,
      handleTimelineGridContextKeyDown,
    },
    menu:
      invocation === null
        ? null
        : {
            position: invocation.position,
            row,
            menuRef,
            onClose: close,
            onDismissToGrid: dismissToGrid,
            onExternalScroll,
            onRestoreFocus,
            onFocusChange,
          },
  };
}

function semanticCell(target: Element) {
  const cell =
    target.closest<HTMLElement>("[data-grid-field-key]") ??
    target
      .closest("[role='gridcell']")
      ?.querySelector<HTMLElement>("[data-grid-field-key]");
  return {
    fieldKey: cell?.dataset.gridFieldKey,
    recordId: target.closest<HTMLElement>("[data-grid-record-id]")?.dataset
      .gridRecordId,
  };
}
