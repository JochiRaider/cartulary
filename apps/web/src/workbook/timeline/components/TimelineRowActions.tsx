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
  useEffect,
  useMemo,
  useRef,
} from "react";
import { workbookViewportOverlayScrollableStyle } from "../../layout/workbookShellStyles";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type { TimelineRowContextMenuPosition } from "../models/timelineControllerPorts";
import type { WorkbookRow } from "../models/workbookTimelineModel";
import { actionButtonStyle, inputStyle } from "./TimelineWorkbookStyles";

type TimelineRowContextMenuProps = {
  readonly position: TimelineRowContextMenuPosition;
  readonly replacementDraft: string;
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

  useEffect(() => {
    const menu = menuRef.current;
    const firstAction = menu?.querySelector<HTMLButtonElement>(
      "button:not(:disabled)",
    );
    firstAction?.focus({ preventScroll: true });
  }, []);

  useEffect(() => {
    const closeForPointer = (event: PointerEvent) => {
      const menu = menuRef.current;
      if (menu !== null && !menu.contains(event.target as Node | null)) {
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

  const recordId = row.recordId;
  const closeAfterAction = (action: () => void) => {
    action();
    onClose();
  };

  return (
    <div
      aria-label="Timeline row actions"
      data-testid={workbookRowContextMenuTestId(timelineViewSchemaId, recordId)}
      ref={menuRef}
      role="dialog"
      style={menuStyle}
      onContextMenu={(event) => {
        event.preventDefault();
      }}
      onKeyDown={(event: ReactKeyboardEvent<HTMLDivElement>) => {
        if (event.key === "Escape") {
          event.preventDefault();
          event.stopPropagation();
          onClose();
        }
      }}
    >
      <button
        data-testid={rowInspectButtonTestId(recordId)}
        style={timelineActionButtonStyle}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onInspectRow(recordId);
          });
        }}
      >
        Inspect
      </button>
      <button
        data-testid={rowHistoryOpenButtonTestId(recordId)}
        style={timelineActionButtonStyle}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onOpenHistory(recordId);
          });
        }}
      >
        History
      </button>
      <button
        data-testid={timelineRowMarkReviewedButtonTestId(recordId)}
        disabled={
          row.captureState === "reviewed" || row.captureState === "superseded"
        }
        style={timelineActionButtonStyle}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onMarkReviewed(row.key);
          });
        }}
      >
        Mark reviewed
      </button>
      <input
        aria-label="Replacement record id"
        data-testid={timelineRowReplacementInputTestId(recordId)}
        placeholder="Replacement record id"
        style={timelineReplacementInputStyle}
        type="text"
        value={replacementDraft}
        onChange={(event) => {
          onReplacementDraftChange(row.key, event.target.value);
        }}
      />
      <button
        data-testid={timelineRowSupersedeButtonTestId(recordId)}
        disabled={
          row.captureState === "superseded" || replacementDraft.trim() === ""
        }
        style={timelineActionButtonStyle}
        type="button"
        onClick={() => {
          closeAfterAction(() => {
            onSupersede(row.key);
          });
        }}
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
