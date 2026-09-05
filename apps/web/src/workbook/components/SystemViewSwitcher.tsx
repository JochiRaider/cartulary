import {
  systemViewSwitcherGroupTestId,
  systemViewSwitcherMenuTestId,
  systemViewSwitcherOptionTestId,
  systemViewSwitcherTriggerTestId,
} from "@cartulary/ui-contracts";
import { useRef, useState } from "react";
import { useRegisteredOverlayNavigation } from "../../shared/useRegisteredOverlayNavigation";
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
  const triggerRef = useRef<HTMLButtonElement | null>(null);
  const activeSystemEntryIndex = systemViewSwitcherEntries.findIndex(
    (entry) => entry.viewSchemaId === activeViewSchemaId,
  );
  const activeSystemEntry =
    activeSystemEntryIndex === -1
      ? null
      : (systemViewSwitcherEntries[activeSystemEntryIndex] ?? null);
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: activeSystemEntry?.viewSchemaId ?? null,
    isOpen,
    itemKeys: systemViewSwitcherEntries.map((entry) => entry.viewSchemaId),
    onRequestClose: () => setIsOpen(false),
    subjectKey: "system-view-switcher",
    triggerRef,
  });

  const openMenu = () => {
    navigation.prepareOpen(activeSystemEntry?.viewSchemaId ?? null);
    setIsOpen(true);
  };

  const selectOption = (viewSchemaId: string) => {
    navigation.close({ restoreTriggerFocus: false });
    onSelect(viewSchemaId);
  };

  return (
    <fieldset
      aria-label="System view switcher"
      style={systemViewSwitcherStyle}
      onBlur={navigation.onOverlayBlur}
    >
      <button
        ref={triggerRef}
        aria-controls={isOpen ? systemViewSwitcherMenuTestId() : undefined}
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-label="System views"
        data-active={activeSystemEntry === null ? "false" : "true"}
        data-testid={systemViewSwitcherTriggerTestId()}
        data-view-schema-id={activeSystemEntry?.viewSchemaId ?? ""}
        style={systemViewSwitcherTriggerStyle}
        type="button"
        onClick={() => {
          if (isOpen) {
            navigation.close({ restoreTriggerFocus: false });
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
            event.stopPropagation();
            navigation.close({ restoreTriggerFocus: true });
          }
        }}
      >
        <span>System views</span>
      </button>
      {isOpen ? (
        <div
          data-testid={systemViewSwitcherMenuTestId()}
          id={systemViewSwitcherMenuTestId()}
          role="menu"
          style={systemViewSwitcherMenuStyle}
          tabIndex={-1}
          onKeyDown={(event) => {
            if (event.defaultPrevented || navigation.activeKey === null) return;
            navigation.onItemKeyDown(event, navigation.activeKey);
          }}
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
                const isSelected = entry.viewSchemaId === activeViewSchemaId;
                return (
                  <button
                    key={entry.viewSchemaId}
                    ref={navigation.registerItem(entry.viewSchemaId)}
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
                    tabIndex={navigation.tabIndexFor(entry.viewSchemaId)}
                    type="button"
                    onMouseDown={(event) => {
                      event.preventDefault();
                    }}
                    onClick={() => {
                      selectOption(entry.viewSchemaId);
                    }}
                    onKeyDown={(event) => {
                      navigation.onItemKeyDown(event, entry.viewSchemaId);
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
