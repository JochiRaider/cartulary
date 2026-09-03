import { gridGroupingSelectTestId } from "@cartulary/ui-contracts";
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
  compact,
  constrained,
  onCommand,
  projection,
  selectedFieldKey,
  surface,
}: {
  readonly compact: boolean;
  readonly constrained: boolean;
  readonly onCommand: (command: WorkbookGridQueryCommand) => void;
  readonly projection: WorkbookGridQueryControlProjection;
  readonly selectedFieldKey: string | null;
  readonly surface: string;
}) {
  const declaredFields = projection.groupOptions.map(
    (option) => option.fieldKey,
  );
  return (
    <label
      style={{
        ...groupControlStyle,
        ...(constrained ? constrainedGroupControlStyle : null),
        ...(compact ? condensedGroupControlStyle : null),
      }}
    >
      <span style={immutableControlLabelStyle}>Group:</span>
      <select
        aria-label="Group rows"
        data-testid={gridGroupingSelectTestId(surface)}
        style={{
          ...groupSelectStyle,
          ...(constrained ? constrainedGroupSelectStyle : null),
        }}
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
const constrainedGroupSelectStyle = {
  inlineSize: "100%",
  minInlineSize: 0,
  maxInlineSize: "100%",
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
const constrainedGroupControlStyle = {
  display: "grid",
  gridTemplateColumns: "max-content minmax(0, 1fr)",
  minInlineSize: 0,
};
const condensedGroupControlStyle = { maxInlineSize: "5.75rem" };
