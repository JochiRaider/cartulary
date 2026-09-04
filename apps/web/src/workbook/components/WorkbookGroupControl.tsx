import { gridGroupingSelectTestId } from "@cartulary/ui-contracts";
import { type RefObject, useEffect } from "react";
import {
  parseDeclaredGroupField,
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
} from "../models/workbookGridQueryControls";
import {
  immutableControlLabelStyle,
  selectStyle,
} from "./workbookGridControlStyles";

export function WorkbookGroupControl({
  isOpen,
  onClose,
  onCommand,
  onToggle,
  projection,
  returnFocusRef,
  selectedFieldKey,
  surface,
  triggerRef,
}: {
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly returnFocusRef: RefObject<HTMLElement | null>;
  readonly selectedFieldKey: string | null;
  readonly surface: string;
  readonly triggerRef: RefObject<HTMLSelectElement | null>;
}) {
  useEffect(() => {
    if (isOpen) triggerRef.current?.focus();
  }, [isOpen, triggerRef]);
  const declaredFields = projection.groupOptions.map(
    (option) => option.fieldKey,
  );
  return (
    <label style={groupControlStyle}>
      <span style={immutableControlLabelStyle}>Group:</span>
      <select
        ref={triggerRef}
        aria-label="Group rows"
        data-testid={gridGroupingSelectTestId(surface)}
        style={groupSelectStyle}
        title={projection.activeGroupLabel}
        value={selectedFieldKey ?? ""}
        onChange={(event) => {
          const parsed = parseDeclaredGroupField(
            event.currentTarget.value,
            declaredFields,
          );
          if (parsed === null) return;
          onCommand({
            kind: "group_set",
            fieldKey: parsed.kind === "none" ? null : parsed.fieldKey,
          });
          onClose();
        }}
        onClick={onToggle}
        onKeyDown={(event) => {
          if (event.key === "Escape" && isOpen) {
            event.preventDefault();
            onClose();
            queueMicrotask(() => {
              const preferred = returnFocusRef.current;
              if (preferred?.isConnected) preferred.focus();
              else triggerRef.current?.focus();
            });
          }
        }}
      >
        <option value="">None</option>
        {projection.groupOptions.map((option) => (
          <option key={option.fieldKey} value={option.fieldKey}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

const groupSelectStyle = {
  ...selectStyle,
  minInlineSize: "5.75rem",
  maxInlineSize: "7rem",
};
const groupControlStyle = {
  display: "inline-flex",
  gap: "0.25rem",
  alignItems: "center",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.78rem",
  whiteSpace: "nowrap" as const,
  minWidth: 0,
};
