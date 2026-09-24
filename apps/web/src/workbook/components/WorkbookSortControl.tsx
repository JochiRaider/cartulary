import {
  workbookSortAppliedEntryTestId,
  workbookSortMenuTestId,
  workbookSortMenuTriggerTestId,
  workbookSortOptionTestId,
} from "@cartulary/ui-contracts";
import { type RefObject, useMemo } from "react";
import { useRegisteredOverlayNavigation } from "../../shared/useRegisteredOverlayNavigation";
import {
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
  workbookOrderedSortLimit,
} from "../models/workbookGridQueryControls";
import {
  controlButtonStyle,
  menuFrameStyle,
  menuItemSelectedStyle,
  menuItemStyle,
  menuStyle,
} from "./workbookGridControlStyles";

export function WorkbookSortControl({
  constrained,
  editorProjection,
  isOpen,
  onClose,
  onCommand,
  onToggle,
  projection,
  returnFocusRef,
  requestedFieldKey,
  surface,
  sortUnapplied,
  triggerRef,
}: {
  readonly constrained: boolean;
  readonly editorProjection: WorkbookGridQueryControlProjection;
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly returnFocusRef: RefObject<HTMLElement | null>;
  readonly requestedFieldKey: string | null;
  readonly surface: string;
  readonly sortUnapplied: boolean;
  readonly triggerRef: RefObject<HTMLButtonElement | null>;
}) {
  const itemKeys = useMemo(
    () => sortControlItemKeys(editorProjection),
    [editorProjection],
  );
  const requestedKey =
    requestedFieldKey === null ? null : `${requestedFieldKey}:direction`;
  const initialKey =
    requestedKey !== null && itemKeys.includes(requestedKey)
      ? requestedKey
      : (itemKeys[0] ?? null);
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: initialKey,
    isOpen,
    itemKeys,
    onRequestClose: onClose,
    preferredReturnFocusRef: returnFocusRef,
    reconcileItemKey: sortReconciledItemKey,
    reconcileItems: true,
    restoreFocusOnSubjectChange: false,
    restoreFocusOnUnmount: false,
    subjectKey: `${surface}:${requestedFieldKey ?? "trigger"}`,
    triggerRef,
  });
  const atLimit =
    editorProjection.sortEntries.length >= workbookOrderedSortLimit;

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
        aria-label={
          projection.sortEntries.length === 0
            ? "Sort, no user sorts"
            : `Sort, ${projection.sortEntries.length} applied`
        }
        data-testid={workbookSortMenuTriggerTestId(surface)}
        style={sortControlButtonStyle}
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
        Sort
      </button>
      {isOpen ? (
        <div
          aria-label="Ordered sort controls"
          data-testid={workbookSortMenuTestId(surface)}
          id={workbookSortMenuTestId(surface)}
          role="menu"
          style={sortMenuStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onFocusCapture={navigation.onOverlayFocus}
          onKeyDown={navigation.onOverlayKeyDown}
        >
          {sortUnapplied ? (
            <p role="status" style={sortLimitStyle}>
              Sort changes are unapplied until the query is accepted.
            </p>
          ) : null}
          <fieldset style={sectionStyle}>
            <legend>
              {sortUnapplied ? "Requested sorts" : "Applied sorts"}
            </legend>
            {editorProjection.sortEntries.length === 0 ? (
              <span style={emptyStateStyle}>No user sort override.</span>
            ) : (
              editorProjection.sortEntries.map((entry) => (
                <div
                  key={entry.fieldKey}
                  data-testid={workbookSortAppliedEntryTestId(
                    surface,
                    entry.fieldKey,
                  )}
                  role="none"
                  style={sortRowStyle}
                >
                  <button
                    ref={navigation.registerItem(`${entry.fieldKey}:direction`)}
                    aria-checked="true"
                    aria-label={`Set ${entry.label} ${
                      entry.direction === "asc" ? "descending" : "ascending"
                    }`}
                    data-testid={workbookSortOptionTestId(
                      surface,
                      entry.fieldKey,
                    )}
                    role="menuitemcheckbox"
                    style={{ ...menuItemStyle, ...menuItemSelectedStyle }}
                    tabIndex={navigation.tabIndexFor(
                      `${entry.fieldKey}:direction`,
                    )}
                    type="button"
                    onClick={() => {
                      onCommand({
                        kind: "sort_set_direction",
                        fieldKey: entry.fieldKey,
                        direction: entry.direction === "asc" ? "desc" : "asc",
                      });
                    }}
                    onKeyDown={(event) => {
                      navigation.onItemKeyDown(
                        event,
                        `${entry.fieldKey}:direction`,
                      );
                    }}
                  >
                    {entry.priority}. {entry.label}: {entry.direction}
                  </button>
                  <SortEntryActions
                    entry={entry}
                    navigation={navigation}
                    onCommand={onCommand}
                    sortCount={editorProjection.sortEntries.length}
                  />
                </div>
              ))
            )}
          </fieldset>
          <fieldset style={sectionStyle}>
            <legend>Add sort</legend>
            {editorProjection.unusedSortableFields.map((field) => {
              const itemKey = `${field.fieldKey}:add`;
              return (
                <button
                  key={field.fieldKey}
                  ref={navigation.registerItem(itemKey)}
                  aria-label={`Add sort ${field.label}`}
                  data-testid={workbookSortOptionTestId(
                    surface,
                    field.fieldKey,
                  )}
                  disabled={atLimit}
                  role="menuitem"
                  style={menuItemStyle}
                  tabIndex={navigation.tabIndexFor(itemKey)}
                  type="button"
                  onClick={() => {
                    onCommand({ kind: "sort_add", fieldKey: field.fieldKey });
                  }}
                  onKeyDown={(event) => {
                    navigation.onItemKeyDown(event, itemKey);
                  }}
                >
                  {field.label}
                </button>
              );
            })}
            {atLimit ? (
              <p role="status" style={sortLimitStyle}>
                Remove a sort before adding another. The limit is{" "}
                {workbookOrderedSortLimit}.
              </p>
            ) : null}
          </fieldset>
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
      style={controlButtonStyle}
      tabIndex={navigation.tabIndexFor(itemKey)}
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
      onKeyDown={(event) => navigation.onItemKeyDown(event, itemKey)}
    >
      {action === "earlier" ? "↑" : action === "later" ? "↓" : "×"}
    </button>
  );
}

function sortControlItemKeys(
  projection: WorkbookGridQueryControlProjection,
): readonly string[] {
  return [
    ...projection.sortEntries.flatMap((entry) => [
      `${entry.fieldKey}:direction`,
      `${entry.fieldKey}:earlier`,
      `${entry.fieldKey}:later`,
      `${entry.fieldKey}:remove`,
    ]),
    ...projection.unusedSortableFields.map((field) => `${field.fieldKey}:add`),
  ];
}

function sortReconciledItemKey(
  previousKey: string,
  previousKeys: readonly string[],
  eligibleKeys: readonly string[],
): string | null {
  const separator = previousKey.lastIndexOf(":");
  if (separator < 0) return null;
  const fieldKey = previousKey.slice(0, separator);
  const action = previousKey.slice(separator + 1);
  const directionKey = `${fieldKey}:direction`;
  if (eligibleKeys.includes(directionKey)) return directionKey;
  if (action !== "add") {
    const priorDirections = previousKeys.filter((key) =>
      key.endsWith(":direction"),
    );
    const oldIndex = priorDirections.indexOf(directionKey);
    if (oldIndex >= 0) {
      for (const key of priorDirections.slice(oldIndex + 1)) {
        if (eligibleKeys.includes(key)) return key;
      }
      for (const key of priorDirections.slice(0, oldIndex).reverse()) {
        if (eligibleKeys.includes(key)) return key;
      }
    }
  }
  const sameFieldAdd = `${fieldKey}:add`;
  if (eligibleKeys.includes(sameFieldAdd)) return sameFieldAdd;
  return eligibleKeys.find((key) => key.endsWith(":add")) ?? null;
}

const sortMenuFrameStyle = { ...menuFrameStyle, flex: "0 0 auto" };
const constrainedSortMenuFrameStyle = { minInlineSize: "max-content" };
const sortControlButtonStyle = { ...controlButtonStyle, flex: "0 0 auto" };
const sortMenuStyle = {
  ...menuStyle,
  inlineSize: "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 92vw)",
};
const sectionStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  border: 0,
  padding: 0,
  margin: 0,
  minWidth: 0,
};
const emptyStateStyle = {
  color: "var(--ct-colors-ink-muted)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
};
const sortRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(10rem, 1fr) auto",
  alignItems: "center",
};
const sortActionsStyle = {
  display: "inline-flex",
  gap: "var(--ct-spacing-xxs)",
};
const sortLimitStyle = {
  margin: "var(--ct-spacing-xs) 0 0",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
};
