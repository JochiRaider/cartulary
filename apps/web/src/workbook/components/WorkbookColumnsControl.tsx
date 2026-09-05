import {
  workbookColumnsMenuTestId,
  workbookColumnsMenuTriggerTestId,
} from "@cartulary/ui-contracts";
import { useMemo, useRef } from "react";
import { useRegisteredOverlayNavigation } from "../../shared/useRegisteredOverlayNavigation";
import type {
  WorkbookGridQueryCommand,
  WorkbookGridQueryControlProjection,
} from "../models/workbookGridQueryControls";
import {
  controlButtonStyle,
  fixedMenuFrameStyle,
  menuItemStyle,
  menuStyle,
} from "./workbookGridControlStyles";

export function WorkbookColumnsControl({
  isOpen,
  onClose,
  onCommand,
  onToggle,
  projection,
  surface,
}: {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly surface: string;
}) {
  const triggerRef = useRef<HTMLButtonElement>(null);
  const itemKeys = useMemo(
    () => [
      ...projection.columns.flatMap((column) => [
        `${column.fieldKey}:visibility`,
        `${column.fieldKey}:earlier`,
        `${column.fieldKey}:later`,
      ]),
      "reset",
    ],
    [projection.columns],
  );
  const initialKey = itemKeys[0] ?? "reset";
  const navigation = useRegisteredOverlayNavigation({
    initialItemKey: initialKey,
    isOpen,
    itemKeys,
    onRequestClose: onClose,
    subjectKey: surface,
    triggerRef,
  });

  return (
    <div style={fixedMenuFrameStyle}>
      <button
        ref={triggerRef}
        aria-controls={isOpen ? workbookColumnsMenuTestId(surface) : undefined}
        aria-expanded={isOpen}
        aria-haspopup="menu"
        data-testid={workbookColumnsMenuTriggerTestId(surface)}
        style={controlButtonStyle}
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
        Columns
      </button>
      {isOpen ? (
        <div
          aria-label="Column controls"
          data-testid={workbookColumnsMenuTestId(surface)}
          id={workbookColumnsMenuTestId(surface)}
          role="menu"
          style={columnsMenuStyle}
          tabIndex={-1}
          onBlur={navigation.onOverlayBlur}
          onKeyDown={(event) => {
            if (event.defaultPrevented || navigation.activeKey === null) return;
            navigation.onItemKeyDown(event, navigation.activeKey);
          }}
        >
          {projection.columns.map((column) => (
            <div key={column.fieldKey} role="none" style={columnMenuRowStyle}>
              <button
                ref={navigation.registerItem(`${column.fieldKey}:visibility`)}
                aria-checked={!column.hidden}
                role="menuitemcheckbox"
                style={menuItemStyle}
                tabIndex={navigation.tabIndexFor(
                  `${column.fieldKey}:visibility`,
                )}
                type="button"
                onClick={() => {
                  onCommand({
                    kind: "column_set_hidden",
                    fieldKey: column.fieldKey,
                    hidden: !column.hidden,
                  });
                }}
                onKeyDown={(event) => {
                  navigation.onItemKeyDown(
                    event,
                    `${column.fieldKey}:visibility`,
                  );
                }}
              >
                {column.label}
              </button>
              <ColumnMoveButton
                column={column}
                direction="earlier"
                disabled={column.position === 0}
                navigation={navigation}
                onCommand={onCommand}
              />
              <ColumnMoveButton
                column={column}
                direction="later"
                disabled={column.position === projection.columns.length - 1}
                navigation={navigation}
                onCommand={onCommand}
              />
            </div>
          ))}
          <button
            ref={navigation.registerItem("reset")}
            role="menuitem"
            style={menuItemStyle}
            tabIndex={navigation.tabIndexFor("reset")}
            type="button"
            onClick={() => {
              onCommand({ kind: "columns_reset" });
            }}
            onKeyDown={(event) => {
              navigation.onItemKeyDown(event, "reset");
            }}
          >
            Reset columns
          </button>
        </div>
      ) : null}
    </div>
  );
}

function ColumnMoveButton({
  column,
  direction,
  disabled,
  navigation,
  onCommand,
}: {
  readonly column: WorkbookGridQueryControlProjection["columns"][number];
  readonly direction: "earlier" | "later";
  readonly disabled: boolean;
  readonly navigation: ReturnType<typeof useRegisteredOverlayNavigation>;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
}) {
  const itemKey = `${column.fieldKey}:${direction}`;
  return (
    <button
      ref={navigation.registerItem(itemKey)}
      aria-label={`Move ${column.label} ${direction}`}
      disabled={disabled}
      role="menuitem"
      tabIndex={navigation.tabIndexFor(itemKey)}
      type="button"
      onClick={() => {
        onCommand({
          kind: "column_move",
          fieldKey: column.fieldKey,
          direction,
        });
      }}
      onKeyDown={(event) => {
        navigation.onItemKeyDown(event, itemKey);
      }}
    >
      {direction === "earlier" ? "↑" : "↓"}
    </button>
  );
}

const columnMenuRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(10rem, 1fr) 2rem 2rem",
  alignItems: "center",
};
const columnsMenuStyle = {
  ...menuStyle,
  insetInlineStart: "auto",
  insetInlineEnd: 0,
  inlineSize: "min(var(--ct-layout-viewBarOverlayMaxInlineSize), 92vw)",
};
