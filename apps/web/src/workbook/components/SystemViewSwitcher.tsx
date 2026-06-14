import {
  systemViewSwitcherGroupTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
} from "@cartulary/ui-contracts";
import {
  type KeyboardEvent as ReactKeyboardEvent,
  useCallback,
  useEffect,
  useRef,
  useState,
} from "react";
import { listSystemWorkbookSurfaceGroups } from "../models/workbookSurfaceRegistry";

const systemWorkbookSurfaceGroups = listSystemWorkbookSurfaceGroups();
const systemViewSwitcherEntries = systemWorkbookSurfaceGroups.flatMap((group) =>
  group.entries.map((entry) => ({
    contract: entry.contract,
    groupToken: group.token,
    viewSchemaId: entry.viewSchemaId,
  })),
);

export function SystemViewSwitcher({
  activeViewSchemaId,
  onSelect,
}: {
  readonly activeViewSchemaId: string;
  readonly onSelect: (viewSchemaId: string) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [activeIndex, setActiveIndex] = useState(0);
  const containerRef = useRef<HTMLFieldSetElement | null>(null);
  const blurCloseTimerRef = useRef<number | null>(null);
  const deferredFocusTimerRef = useRef<number | null>(null);
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const optionRefs = useRef(new Map<string, HTMLButtonElement>());
  const activeSystemEntryIndex = systemViewSwitcherEntries.findIndex(
    (entry) => entry.viewSchemaId === activeViewSchemaId,
  );
  const activeSystemEntry =
    activeSystemEntryIndex === -1
      ? null
      : (systemViewSwitcherEntries[activeSystemEntryIndex] ?? null);

  const clearBlurCloseTimer = useCallback(() => {
    if (blurCloseTimerRef.current !== null) {
      window.clearTimeout(blurCloseTimerRef.current);
      blurCloseTimerRef.current = null;
    }
  }, []);

  const clearDeferredFocusTimer = useCallback(() => {
    if (deferredFocusTimerRef.current !== null) {
      window.clearTimeout(deferredFocusTimerRef.current);
      deferredFocusTimerRef.current = null;
    }
  }, []);

  useEffect(
    () => () => {
      clearBlurCloseTimer();
      clearDeferredFocusTimer();
    },
    [clearBlurCloseTimer, clearDeferredFocusTimer],
  );

  const deferFocus = useCallback(
    (focusTarget: () => HTMLElement | null) => {
      clearDeferredFocusTimer();
      deferredFocusTimerRef.current = window.setTimeout(() => {
        deferredFocusTimerRef.current = null;
        focusTarget()?.focus({ preventScroll: true });
      }, 0);
    },
    [clearDeferredFocusTimer],
  );

  const focusOption = useCallback(
    (index: number) => {
      const entry = systemViewSwitcherEntries[index];
      if (entry === undefined) {
        return;
      }
      deferFocus(() => optionRefs.current.get(entry.viewSchemaId) ?? null);
    },
    [deferFocus],
  );

  const scheduleBlurClose = useCallback(() => {
    clearBlurCloseTimer();
    blurCloseTimerRef.current = window.setTimeout(() => {
      blurCloseTimerRef.current = null;
      const activeElement = document.activeElement;
      if (
        activeElement instanceof Node &&
        containerRef.current?.contains(activeElement)
      ) {
        return;
      }
      setIsOpen(false);
    }, 0);
  }, [clearBlurCloseTimer]);

  const openMenu = useCallback(() => {
    clearBlurCloseTimer();
    const nextIndex =
      activeSystemEntryIndex === -1 ? 0 : activeSystemEntryIndex;
    setActiveIndex(nextIndex);
    setIsOpen(true);
    focusOption(nextIndex);
  }, [activeSystemEntryIndex, clearBlurCloseTimer, focusOption]);

  const closeMenu = useCallback(
    (options: { readonly restoreTriggerFocus: boolean }) => {
      clearBlurCloseTimer();
      clearDeferredFocusTimer();
      setIsOpen(false);
      if (options.restoreTriggerFocus) {
        deferFocus(() => triggerRef.current);
      }
    },
    [clearBlurCloseTimer, clearDeferredFocusTimer, deferFocus],
  );

  const handleInternalPointerDown = useCallback(() => {
    clearBlurCloseTimer();
  }, [clearBlurCloseTimer]);

  const moveOptionFocus = useCallback(
    (nextIndex: number) => {
      const optionCount = systemViewSwitcherEntries.length;
      if (optionCount === 0) {
        return;
      }
      const wrappedIndex = (nextIndex + optionCount) % optionCount;
      setActiveIndex(wrappedIndex);
      focusOption(wrappedIndex);
    },
    [focusOption],
  );

  const selectOption = useCallback(
    (viewSchemaId: string) => {
      if (viewSchemaId === "") {
        return;
      }
      clearBlurCloseTimer();
      clearDeferredFocusTimer();
      setIsOpen(false);
      onSelect(viewSchemaId);
    },
    [clearBlurCloseTimer, clearDeferredFocusTimer, onSelect],
  );

  const handleOptionKeyDown = useCallback(
    (event: ReactKeyboardEvent<HTMLButtonElement>, index: number) => {
      switch (event.key) {
        case "ArrowDown":
          event.preventDefault();
          moveOptionFocus(index + 1);
          break;
        case "ArrowUp":
          event.preventDefault();
          moveOptionFocus(index - 1);
          break;
        case "Home":
          event.preventDefault();
          moveOptionFocus(0);
          break;
        case "End":
          event.preventDefault();
          moveOptionFocus(systemViewSwitcherEntries.length - 1);
          break;
        case "Escape":
          event.preventDefault();
          closeMenu({ restoreTriggerFocus: true });
          break;
        case "Enter":
          event.preventDefault();
          selectOption(systemViewSwitcherEntries[index]?.viewSchemaId ?? "");
          break;
        default:
          break;
      }
    },
    [closeMenu, moveOptionFocus, selectOption],
  );

  return (
    <fieldset
      aria-label="System view switcher"
      ref={containerRef}
      style={systemViewSwitcherStyle}
      onPointerDownCapture={handleInternalPointerDown}
      onBlur={(event) => {
        const nextFocus = event.relatedTarget;
        if (
          nextFocus instanceof Node &&
          containerRef.current?.contains(nextFocus)
        ) {
          return;
        }
        scheduleBlurClose();
      }}
    >
      <button
        ref={triggerRef}
        aria-controls={isOpen ? systemViewSwitcherMenuTestId() : undefined}
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-label="System views"
        data-testid={systemViewSwitcherTriggerTestId()}
        data-view-schema-id={activeSystemEntry?.viewSchemaId ?? ""}
        style={systemViewSwitcherTriggerStyle}
        type="button"
        onClick={() => {
          if (isOpen) {
            closeMenu({ restoreTriggerFocus: false });
            return;
          }
          openMenu();
        }}
        onKeyDown={(event) => {
          if (event.key === "Enter" || event.key === " ") {
            event.preventDefault();
            openMenu();
          }
          if (event.key === "Escape" && isOpen) {
            event.preventDefault();
            closeMenu({ restoreTriggerFocus: true });
          }
        }}
      >
        <span>System views</span>
        <span aria-hidden="true" style={systemViewSwitcherValueStyle}>
          {activeSystemEntry?.contract.title ?? "Select view"}
        </span>
      </button>
      {isOpen ? (
        <div
          data-testid={systemViewSwitcherMenuTestId()}
          id={systemViewSwitcherMenuTestId()}
          role="menu"
          style={systemViewSwitcherMenuStyle}
        >
          {systemWorkbookSurfaceGroups.map((group) => (
            <fieldset
              key={group.token}
              aria-label={group.label}
              data-testid={systemViewSwitcherGroupTestId(group.token)}
              style={systemViewSwitcherGroupStyle}
            >
              <legend style={systemViewSwitcherGroupLabelStyle}>
                {group.label}
              </legend>
              {group.entries.map((entry) => {
                const optionIndex = systemViewSwitcherEntries.findIndex(
                  (option) => option.viewSchemaId === entry.viewSchemaId,
                );
                const isSelected = entry.viewSchemaId === activeViewSchemaId;
                return (
                  <button
                    key={entry.viewSchemaId}
                    ref={(node) => {
                      if (node === null) {
                        optionRefs.current.delete(entry.viewSchemaId);
                        return;
                      }
                      optionRefs.current.set(entry.viewSchemaId, node);
                    }}
                    aria-checked={isSelected}
                    data-testid={systemViewSwitcherOptionTestId(
                      group.token,
                      entry.viewSchemaId,
                    )}
                    data-view-schema-id={entry.viewSchemaId}
                    role="menuitemradio"
                    style={{
                      ...systemViewSwitcherOptionStyle,
                      ...(isSelected
                        ? systemViewSwitcherOptionSelectedStyle
                        : null),
                    }}
                    tabIndex={optionIndex === activeIndex ? 0 : -1}
                    type="button"
                    onMouseDown={(event) => {
                      event.preventDefault();
                    }}
                    onClick={() => {
                      selectOption(entry.viewSchemaId);
                    }}
                    onKeyDown={(event) => {
                      handleOptionKeyDown(event, optionIndex);
                    }}
                  >
                    {entry.contract.title}
                  </button>
                );
              })}
            </fieldset>
          ))}
        </div>
      ) : null}
    </fieldset>
  );
}

const surfaceTabStyle = {
  borderRadius: 0,
  border: 0,
  borderBottom: "2px solid transparent",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  padding: "0.85rem 0.55rem 0.7rem",
  font: "inherit",
  cursor: "pointer",
  whiteSpace: "nowrap" as const,
};

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

const systemViewSwitcherStyle = {
  position: "relative" as const,
  minWidth: "10rem",
  border: 0,
  margin: 0,
  padding: 0,
};

const systemViewSwitcherTriggerStyle = {
  ...surfaceTabStyle,
  width: "100%",
  display: "flex",
  justifyContent: "space-between",
  alignItems: "center",
  gap: "0.75rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  padding: "0.45rem 0.65rem",
};

const systemViewSwitcherValueStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.85rem",
  fontWeight: 500,
  overflow: "hidden",
  textOverflow: "ellipsis",
  whiteSpace: "nowrap" as const,
};

const systemViewSwitcherMenuStyle = {
  position: "absolute" as const,
  zIndex: 10,
  insetBlockStart: "calc(100% + 0.35rem)",
  insetInlineStart: 0,
  width: "min(26rem, 80vw)",
  maxHeight: "28rem",
  overflowY: "auto" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-md)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
  padding: "0.5rem",
};

const systemViewSwitcherGroupStyle = {
  display: "grid",
  gap: "0.25rem",
  padding: "0.35rem 0",
  border: 0,
  margin: 0,
  minInlineSize: 0,
};

const systemViewSwitcherGroupLabelStyle = {
  ...eyebrowStyle,
  margin: "0.2rem 0.45rem",
  padding: 0,
};

const systemViewSwitcherOptionStyle = {
  border: "0",
  borderRadius: "var(--ct-rounded-sm)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
  font: "inherit",
  padding: "0.5rem 0.55rem",
  textAlign: "left" as const,
};

const systemViewSwitcherOptionSelectedStyle = {
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink)",
  fontWeight: 700,
};
