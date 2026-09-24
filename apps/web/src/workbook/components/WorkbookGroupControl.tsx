import { gridGroupingSelectTestId } from "@cartulary/ui-contracts";
import { type RefObject, useEffect, useRef } from "react";
import {
  parseDeclaredGroupField,
  type WorkbookGridQueryCommand,
  type WorkbookGridQueryControlProjection,
} from "../models/workbookGridQueryControls";
import { visuallyHiddenStyle } from "../utils/workbookStyles";
import {
  immutableControlLabelStyle,
  selectStyle,
} from "./workbookGridControlStyles";

export function WorkbookGroupControl({
  groupUnapplied,
  isOpen,
  onClose,
  onCommand,
  onToggle,
  projection,
  returnFocusRef,
  selectedFieldKey,
  subjectKey,
  surface,
  triggerRef,
}: {
  readonly groupUnapplied: boolean;
  readonly isOpen: boolean;
  readonly onClose: () => void;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly onToggle: () => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly returnFocusRef: RefObject<HTMLElement | null>;
  readonly selectedFieldKey: string | null;
  readonly subjectKey: string;
  readonly surface: string;
  readonly triggerRef: RefObject<HTMLSelectElement | null>;
}) {
  const currentSubjectKey = useRef(subjectKey);
  currentSubjectKey.current = subjectKey;
  useEffect(() => {
    if (isOpen) triggerRef.current?.focus();
  }, [isOpen, triggerRef]);
  const declaredFields = projection.groupOptions.map(
    (option) => option.fieldKey,
  );
  const requestedLabel =
    projection.groupOptions.find(
      (option) => option.fieldKey === selectedFieldKey,
    )?.label ??
    selectedFieldKey ??
    "None";
  const statusId = `${gridGroupingSelectTestId(surface)}-unapplied`;
  return (
    <label style={groupControlStyle}>
      <span style={immutableControlLabelStyle}>Group:</span>
      <select
        ref={triggerRef}
        aria-describedby={groupUnapplied ? statusId : undefined}
        aria-label="Group rows"
        data-testid={gridGroupingSelectTestId(surface)}
        style={groupSelectStyle}
        title={requestedLabel}
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
              if (currentSubjectKey.current !== subjectKey) return;
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
      {groupUnapplied ? (
        <span id={statusId} role="status" style={groupUnappliedStyle}>
          Unapplied
          <span style={visuallyHiddenStyle}>
            {`. Requested ${requestedLabel}; retained results grouped by ${projection.activeGroupLabel}.`}
          </span>
        </span>
      ) : null}
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
const groupUnappliedStyle = {
  flex: "0 0 auto",
  color: "var(--ct-colors-ink-muted)",
};
