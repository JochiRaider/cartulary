import {
  rowHistoryOpenButtonTestId,
  rowInspectButtonTestId,
  timelineRowMarkReviewedButtonTestId,
  timelineRowReplacementInputTestId,
  timelineRowSupersedeButtonTestId,
  workbookRowContextMenuTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type KeyboardEvent as ReactKeyboardEvent,
  type RefObject,
  useEffect,
  useMemo,
  useRef,
} from "react";
import { useRegisteredOverlayNavigation } from "../../focus/useRegisteredOverlayNavigation";
import { workbookViewportOverlayScrollableStyle } from "../../layout/workbookShellStyles";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineRowContextMenuPosition } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/timelineRowModel";
import { actionButtonStyle, inputStyle } from "./TimelineWorkbookStyles";

type TimelineRowContextMenuProps = {
  readonly position: TimelineRowContextMenuPosition;
  readonly replacementDraft: string;
  readonly fallbackFocusTargetRef: RefObject<HTMLElement | null>;
  readonly returnFocusTargetRef: RefObject<HTMLElement | null>;
  readonly row: WorkbookRow | null;
  readonly onClose: () => void;
  readonly onInspectRow: (recordId: string) => void;
  readonly onMarkReviewed: (rowKey: string) => void;
  readonly onOpenHistory: (recordId: string) => void;
  readonly onReplacementDraftChange: (rowKey: string, value: string) => void;
  readonly onSupersede: (rowKey: string) => void;
};

export function TimelineRowContextMenu({
  position,
  replacementDraft,
  fallbackFocusTargetRef,
  returnFocusTargetRef,
  row,
  onClose,
  onInspectRow,
  onMarkReviewed,
  onOpenHistory,
  onReplacementDraftChange,
  onSupersede,
}: TimelineRowContextMenuProps) {
  const menuRef = useRef<HTMLDivElement | null>(null);
  const menuStyle = useMemo(
    () => ({
      ...actionPopoverStyle,
      ...clampedContextMenuPosition(position),
    }),
    [position],
  );
  const recordId = row?.recordId ?? null;
  const itemKeys = ["inspect", "history", "mark-reviewed", "supersede"];
  const navigation = useRegisteredOverlayNavigation({
    fallbackFocusRef: fallbackFocusTargetRef,
    initialItemKey: "inspect",
    isOpen: recordId !== null,
    itemKeys,
    onRequestClose: onClose,
    subjectKey: recordId ?? "unavailable",
    triggerRef: returnFocusTargetRef,
  });

  useEffect(() => {
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
      onClose();
    };
    const closeForScroll = (event: Event) => {
      const menu = menuRef.current;
      if (menu === null) {
        onClose();
        return;
      }
      const activeElement = document.activeElement;
      if (activeElement instanceof Node && menu.contains(activeElement)) {
        return;
      }
      const scrollTarget = event.target;
      if (scrollTarget instanceof Node && menu.contains(scrollTarget)) {
        return;
      }
      onClose();
    };
    window.addEventListener("pointerdown", closeForPointer, true);
    window.addEventListener("resize", closeForResize);
    window.addEventListener("scroll", closeForScroll, true);
    return () => {
      window.removeEventListener("pointerdown", closeForPointer, true);
      window.removeEventListener("resize", closeForResize);
      window.removeEventListener("scroll", closeForScroll, true);
    };
  }, [onClose]);

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
      data-testid={workbookRowContextMenuTestId(
        timelineViewSchemaId,
        availableRecordId,
      )}
      ref={menuRef}
      role="dialog"
      style={menuStyle}
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
      >
        Inspect
      </button>
      <button
        ref={navigation.registerItem("history")}
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
      >
        History
      </button>
      <button
        ref={navigation.registerItem("mark-reviewed")}
        data-testid={timelineRowMarkReviewedButtonTestId(availableRecordId)}
        disabled={
          row.captureState === "reviewed" || row.captureState === "superseded"
        }
        style={timelineActionButtonStyle}
        tabIndex={navigation.tabIndexFor("mark-reviewed")}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onMarkReviewed(row.key);
          });
        }}
        onKeyDown={(event) => navigation.onItemKeyDown(event, "mark-reviewed")}
      >
        Mark reviewed
      </button>
      <input
        aria-label="Replacement record id"
        data-testid={timelineRowReplacementInputTestId(availableRecordId)}
        placeholder="Replacement record id"
        style={timelineReplacementInputStyle}
        type="text"
        value={replacementDraft}
        onChange={(event) => {
          onReplacementDraftChange(row.key, event.target.value);
        }}
      />
      <button
        ref={navigation.registerItem("supersede")}
        data-testid={timelineRowSupersedeButtonTestId(availableRecordId)}
        disabled={
          row.captureState === "superseded" || replacementDraft.trim() === ""
        }
        style={timelineActionButtonStyle}
        tabIndex={navigation.tabIndexFor("supersede")}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onSupersede(row.key);
          });
        }}
        onKeyDown={(event) => navigation.onItemKeyDown(event, "supersede")}
      >
        Supersede
      </button>
    </div>
  );
}

function clampedContextMenuPosition(
  position: TimelineRowContextMenuPosition,
): Pick<CSSProperties, "left" | "top"> {
  const viewportWidth =
    typeof window === "undefined"
      ? position.x + contextMenuWidthPx
      : window.innerWidth;
  const viewportHeight =
    typeof window === "undefined"
      ? position.y + contextMenuHeightPx
      : window.innerHeight;
  const left = Math.max(
    contextMenuMarginPx,
    Math.min(
      position.x,
      viewportWidth - contextMenuWidthPx - contextMenuMarginPx,
    ),
  );
  const top = Math.max(
    contextMenuMarginPx,
    Math.min(
      position.y,
      viewportHeight - contextMenuHeightPx - contextMenuMarginPx,
    ),
  );
  return { left, top };
}

const contextMenuWidthPx = 240;
const contextMenuHeightPx = 248;
const contextMenuMarginPx = 8;

const timelineReplacementInputStyle = {
  ...inputStyle,
  boxSizing: "border-box" as const,
  fontSize: "0.82rem",
  width: "100%",
};

const actionPopoverStyle = {
  ...workbookViewportOverlayScrollableStyle,
  position: "fixed" as const,
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
