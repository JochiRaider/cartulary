import {
  rowHistoryOpenButtonTestId,
  rowInspectButtonTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowSupersedeButtonTestId,
  workbookRowContextMenuTestId,
} from "@cartulary/ui-contracts";
import {
  type KeyboardEvent as ReactKeyboardEvent,
  type RefObject,
  useLayoutEffect,
} from "react";
import { useRegisteredOverlayNavigation } from "../../../shared/useRegisteredOverlayNavigation";
import { workbookViewportOverlayScrollableStyle } from "../../layout/workbookShellStyles";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineRowContextMenuPosition } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/timelineRowModel";
import { actionButtonStyle } from "./TimelineWorkbookStyles";

type TimelineRowContextMenuProps = {
  readonly position: TimelineRowContextMenuPosition;
  readonly reviewDisabledReason: string | null;
  readonly supersedeDisabledReason: string | null;
  readonly menuRef: RefObject<HTMLDivElement | null>;
  readonly onRestoreFocus: () => void;
  readonly onDismissToGrid: () => void;
  readonly onExternalScroll: (event: Event) => void;
  readonly onFocusChange: (owned: boolean) => void;
  readonly row: WorkbookRow | null;
  readonly onClose: () => void;
  readonly onInspectRow: (recordId: string) => void;
  readonly onMarkReviewed: (rowKey: string) => void;
  readonly onOpenHistory: (recordId: string) => void;
  readonly onSupersede: (rowKey: string) => void;
};

export function TimelineRowContextMenu({
  position,
  reviewDisabledReason,
  supersedeDisabledReason,
  menuRef,
  onRestoreFocus,
  onDismissToGrid,
  onExternalScroll,
  onFocusChange,
  row,
  onClose,
  onInspectRow,
  onMarkReviewed,
  onOpenHistory,
  onSupersede,
}: TimelineRowContextMenuProps) {
  useLayoutEffect(() => {
    const menu = menuRef.current;
    if (!menu) return;
    // Invocation rectangles are viewport pixels; fixed-position CSS lengths
    // inherit document zoom. Measure the existing overlay in that same space.
    const rect = menu.getBoundingClientRect();
    const scale = menu.offsetWidth > 0 ? rect.width / menu.offsetWidth : 1;
    const viewport = window.visualViewport;
    const width = (viewport?.width ?? window.innerWidth) / (scale || 1);
    const height = (viewport?.height ?? window.innerHeight) / (scale || 1);
    const margin = contextMenuMarginPx;
    menu.style.maxInlineSize = `${Math.max(0, width - 2 * margin)}px`;
    menu.style.maxBlockSize = `${Math.max(0, height - 2 * margin)}px`;
    const measured = menu.getBoundingClientRect();
    menu.style.left = `${Math.max(margin, Math.min(position.x / (scale || 1), width - measured.width / (scale || 1) - margin))}px`;
    menu.style.top = `${Math.max(margin, Math.min(position.y / (scale || 1), height - measured.height / (scale || 1) - margin))}px`;
  });
  const recordId = row?.recordId ?? null;
  const itemKeys = ["inspect", "history", "mark-reviewed", "supersede"];
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: "inspect",
    isOpen: recordId !== null,
    itemKeys,
    onRequestClose: onClose,
    subjectKey: recordId ?? "unavailable",
    onRestoreFocus,
    reconcileItems: true,
    restoreFocusOnSubjectChange: false,
  });

  useLayoutEffect(() => {
    const closeForPointer = (event: PointerEvent) => {
      const menu = menuRef.current;
      if (
        menu !== null &&
        (!(event.target instanceof Node) || !menu.contains(event.target))
      ) {
        onClose();
      }
    };
    const closeForResize = () => {
      onDismissToGrid();
    };
    const closeForScroll = (event: Event) => {
      const menu = menuRef.current;
      if (menu === null) {
        return;
      }
      const scrollTarget = event.target;
      if (scrollTarget instanceof Node && menu.contains(scrollTarget)) {
        return;
      }
      onExternalScroll(event);
    };
    window.addEventListener("pointerdown", closeForPointer, true);
    window.addEventListener("resize", closeForResize);
    window.visualViewport?.addEventListener("resize", closeForResize);
    window.addEventListener("scroll", closeForScroll, true);
    return () => {
      window.removeEventListener("pointerdown", closeForPointer, true);
      window.removeEventListener("resize", closeForResize);
      window.visualViewport?.removeEventListener("resize", closeForResize);
      window.removeEventListener("scroll", closeForScroll, true);
    };
  }, [menuRef, onClose, onDismissToGrid, onExternalScroll]);

  if (row?.recordId === null || row?.recordId === undefined) {
    return null;
  }

  const availableRecordId = row.recordId;
  const closeAfterAction = (action: () => void) => {
    action();
    onClose();
  };

  return (
    <div
      aria-label="Timeline row actions"
      data-grid-editor-external-action="true"
      data-testid={workbookRowContextMenuTestId(
        timelineViewSchemaId,
        availableRecordId,
      )}
      ref={menuRef}
      role="dialog"
      style={actionPopoverStyle}
      onFocus={(event) => {
        onFocusChange(true);
        const menu = event.currentTarget;
        const item = event.target;
        const bounds = menu.getBoundingClientRect();
        const scale =
          menu.offsetHeight > 0 ? bounds.height / menu.offsetHeight : 1;
        const target = item.getBoundingClientRect();
        // Reveal only inside this overlay; scrolling the grid dismisses it.
        if (target.bottom > bounds.bottom)
          menu.scrollTop += (target.bottom - bounds.bottom) / (scale || 1);
        else if (target.top < bounds.top)
          menu.scrollTop -= (bounds.top - target.top) / (scale || 1);
      }}
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget))
          onFocusChange(false);
        navigation.onOverlayBlur(event);
      }}
      onContextMenu={(event) => {
        event.preventDefault();
      }}
      onKeyDown={(event: ReactKeyboardEvent<HTMLDivElement>) => {
        if (
          !event.defaultPrevented &&
          event.key === "Escape" &&
          navigation.activeKey !== null
        ) {
          navigation.onItemKeyDown(event, navigation.activeKey);
        }
      }}
    >
      <button
        ref={navigation.registerItem("inspect")}
        aria-description="Opens this record in the Inspector."
        data-testid={rowInspectButtonTestId(availableRecordId)}
        style={timelineActionButtonStyle}
        tabIndex={navigation.tabIndexFor("inspect")}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onInspectRow(availableRecordId);
          });
        }}
        onKeyDown={(event) => navigation.onItemKeyDown(event, "inspect")}
        onFocus={() => navigation.onItemFocus("inspect")}
      >
        Inspect
      </button>
      <button
        ref={navigation.registerItem("history")}
        aria-description="Opens this record’s History in the Inspector."
        data-testid={rowHistoryOpenButtonTestId(availableRecordId)}
        style={timelineActionButtonStyle}
        tabIndex={navigation.tabIndexFor("history")}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onOpenHistory(availableRecordId);
          });
        }}
        onKeyDown={(event) => navigation.onItemKeyDown(event, "history")}
        onFocus={() => navigation.onItemFocus("history")}
      >
        History
      </button>
      <button
        ref={navigation.registerItem("mark-reviewed")}
        aria-description="Marks this record reviewed and opens History."
        data-testid={timelineRowMarkReviewedButtonTestId(availableRecordId)}
        disabled={reviewDisabledReason !== null}
        title={reviewDisabledReason ?? undefined}
        style={timelineActionButtonStyle}
        tabIndex={navigation.tabIndexFor("mark-reviewed")}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onMarkReviewed(row.key);
          });
        }}
        onKeyDown={(event) => navigation.onItemKeyDown(event, "mark-reviewed")}
        onFocus={() => navigation.onItemFocus("mark-reviewed")}
      >
        Mark reviewed
      </button>
      <button
        ref={navigation.registerItem("supersede")}
        aria-description="Opens review before superseding this record."
        data-testid={timelineRowSupersedeButtonTestId(availableRecordId)}
        disabled={supersedeDisabledReason !== null}
        title={supersedeDisabledReason ?? undefined}
        style={timelineActionButtonStyle}
        tabIndex={navigation.tabIndexFor("supersede")}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onSupersede(row.key);
          });
        }}
        onKeyDown={(event) => navigation.onItemKeyDown(event, "supersede")}
        onFocus={() => navigation.onItemFocus("supersede")}
      >
        Supersede
      </button>
    </div>
  );
}

const contextMenuWidthPx = 240;
const contextMenuMarginPx = 8;

const actionPopoverStyle = {
  ...workbookViewportOverlayScrollableStyle,
  position: "fixed" as const,
  boxSizing: "border-box" as const,
  zIndex: 30,
  display: "grid",
  gap: "0.35rem",
  inlineSize: `${contextMenuWidthPx}px`,
  padding: "0.45rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  boxShadow: "var(--ct-elevation-popover)",
};

const timelineActionButtonStyle = {
  ...actionButtonStyle,
  boxSizing: "border-box" as const,
  fontSize: "0.85rem",
  lineHeight: 1.1,
  padding: "0.45rem 0.3rem",
  width: "100%",
};
