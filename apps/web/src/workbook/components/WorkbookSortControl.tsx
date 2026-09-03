import {
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
} from "@cartulary/ui-contracts";
import { useMemo, useRef } from "react";
import { useRegisteredOverlayNavigation } from "../focus/useRegisteredOverlayNavigation";
import {
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
  workbookOrderedSortLimit,
} from "../models/workbookGridQueryControls";
import {
  controlButtonStyle,
  dynamicControlValueStyle,
  immutableControlLabelStyle,
  menuFrameStyle,
  menuItemSelectedStyle,
  menuItemStyle,
  menuStyle,
} from "./workbookGridControlStyles";

export function WorkbookSortControl({
  constrained,
  isOpen,
  onClose,
  onCommand,
  onToggle,
  projection,
  surface,
}: {
  readonly constrained: boolean;
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly surface: string;
}) {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const itemKeys = useMemo(() => sortControlItemKeys(projection), [projection]);
  const initialKey = itemKeys[0] ?? null;
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: initialKey,
    isOpen,
    itemKeys,
    onRequestClose: onClose,
    subjectKey: surface,
    triggerRef,
  });
  const triggerLabel =
    projection.activeSortLabel === null
      ? "Sort"
      : `Sort: ${projection.activeSortLabel}`;
  const selectedByField = new Map(
    projection.sortEntries.map((entry) => [entry.fieldKey, entry]),
  );
  const atLimit = projection.sortEntries.length >= workbookOrderedSortLimit;

  return (
    <div
      style={{
        ...sortMenuFrameStyle,
        ...(constrained ? constrainedSortMenuFrameStyle : null),
      }}
    >
      <button
        ref={triggerRef}
        aria-controls={isOpen ? workbookSortMenuTestId(surface) : undefined}
        aria-expanded={isOpen}
        aria-haspopup="menu"
        aria-label={triggerLabel}
        data-testid={workbookSortMenuTriggerTestId(surface)}
        style={sortControlButtonStyle}
        title={triggerLabel}
        type="button"
        onClick={() => {
          if (!isOpen) navigation.prepareOpen(initialKey);
          onToggle();
        }}
        onKeyDown={(event) => {
          if (event.key !== "ArrowDown") return;
          event.preventDefault();
          event.stopPropagation();
          navigation.prepareOpen(initialKey);
          if (!isOpen) onToggle();
        }}
      >
        <span style={immutableControlLabelStyle}>Sort</span>
        {projection.activeSortLabel === null ? null : (
          <span aria-hidden="true" style={dynamicControlValueStyle}>
            : {projection.activeSortLabel}
          </span>
        )}
      </button>
      {isOpen ? (
        <div
          aria-label="Ordered sort controls"
          data-testid={workbookSortMenuTestId(surface)}
          id={workbookSortMenuTestId(surface)}
          role="menu"
          style={menuStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onKeyDown={(event) => {
            if (event.defaultPrevented || navigation.activeKey === null) return;
            navigation.onItemKeyDown(event, navigation.activeKey);
          }}
        >
          {projection.sortableFields.map((field) => {
            const selected = selectedByField.get(field.fieldKey);
            const primaryKey = `${field.fieldKey}:primary`;
            return (
              <div key={field.fieldKey} role="none" style={sortRowStyle}>
                <button
                  ref={navigation.registerItem(primaryKey)}
                  aria-checked={selected !== undefined}
                  aria-label={
                    selected === undefined
                      ? `Add sort ${field.label}`
                      : `Set ${field.label} ${selected.direction === "asc" ? "descending" : "ascending"}`
                  }
                  data-testid={workbookSortOptionTestId(
                    surface,
                    field.fieldKey,
                  )}
                  disabled={selected === undefined && atLimit}
                  role="menuitemcheckbox"
                  style={{
                    ...menuItemStyle,
                    ...(selected === undefined ? null : menuItemSelectedStyle),
                  }}
                  tabIndex={navigation.tabIndexFor(primaryKey)}
                  type="button"
                  onClick={() => {
                    onCommand(
                      selected === undefined
                        ? { kind: "sort_add", fieldKey: field.fieldKey }
                        : {
                            kind: "sort_set_direction",
                            fieldKey: field.fieldKey,
                            direction:
                              selected.direction === "asc" ? "desc" : "asc",
                          },
                    );
                  }}
                  onKeyDown={(event) => {
                    navigation.onItemKeyDown(event, primaryKey);
                  }}
                >
                  {selected === undefined
                    ? `Add sort: ${field.label}`
                    : `${selected.priority}. ${field.label}: ${selected.direction}`}
                </button>
                {selected === undefined ? null : (
                  <SortEntryActions
                    entry={selected}
                    navigation={navigation}
                    onCommand={onCommand}
                    sortCount={projection.sortEntries.length}
                  />
                )}
              </div>
            );
          })}
          {atLimit ? (
            <p role="status" style={sortLimitStyle}>
              Remove a sort before adding another. The limit is{" "}
              {workbookOrderedSortLimit}.
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}

function SortEntryActions({
  entry,
  navigation,
  onCommand,
  sortCount,
}: {
  readonly entry: WorkbookGridQueryControlProjection["sortEntries"][number];
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly sortCount: number;
}) {
  return (
    <span role="none" style={sortActionsStyle}>
      <SortActionButton
        action="earlier"
        disabled={entry.priority === 1}
        entry={entry}
        navigation={navigation}
        onCommand={onCommand}
      />
      <SortActionButton
        action="later"
        disabled={entry.priority === sortCount}
        entry={entry}
        navigation={navigation}
        onCommand={onCommand}
      />
      <SortActionButton
        action="remove"
        disabled={false}
        entry={entry}
        navigation={navigation}
        onCommand={onCommand}
      />
    </span>
  );
}

function SortActionButton({
  action,
  disabled,
  entry,
  navigation,
  onCommand,
}: {
  readonly action: "earlier" | "later" | "remove";
  readonly disabled: boolean;
  readonly entry: WorkbookGridQueryControlProjection["sortEntries"][number];
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
}) {
  const itemKey = `${entry.fieldKey}:${action}`;
  const label =
    action === "remove"
      ? `Remove ${entry.label} sort`
      : `Move ${entry.label} ${action}`;
  return (
    <button
      ref={navigation.registerItem(itemKey)}
      aria-label={label}
      disabled={disabled}
      role="menuitem"
      tabIndex={navigation.tabIndexFor(itemKey)}
      title={label}
      type="button"
      onClick={() => {
        onCommand(
          action === "remove"
            ? { kind: "sort_remove", fieldKey: entry.fieldKey }
            : {
                kind: "sort_move",
                fieldKey: entry.fieldKey,
                direction: action,
              },
        );
      }}
      onKeyDown={(event) => {
        navigation.onItemKeyDown(event, itemKey);
      }}
    >
      {action === "earlier" ? "↑" : action === "later" ? "↓" : "×"}
    </button>
  );
}

function sortControlItemKeys(
  projection: WorkbookGridQueryControlProjection,
): readonly string[] {
  const selected = new Set(
    projection.sortEntries.map((entry) => entry.fieldKey),
  );
  return projection.sortableFields.flatMap((field) => [
    `${field.fieldKey}:primary`,
    ...(selected.has(field.fieldKey)
      ? [
          `${field.fieldKey}:earlier`,
          `${field.fieldKey}:later`,
          `${field.fieldKey}:remove`,
        ]
      : []),
  ]);
}

const sortMenuFrameStyle = { ...menuFrameStyle, maxInlineSize: "8rem" };
const constrainedSortMenuFrameStyle = { minInlineSize: "3.75rem" };
const sortControlButtonStyle = {
  ...controlButtonStyle,
  inlineSize: "100%",
  maxInlineSize: "100%",
  minInlineSize: 0,
  overflow: "hidden",
};
const sortRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(10rem, 1fr) auto",
  alignItems: "center",
};
const sortActionsStyle = { display: "inline-flex", gap: "0.15rem" };
const sortLimitStyle = {
  margin: "0.3rem 0 0",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
};
