import type { CSSProperties } from "react";

type WorkbookRecordCandidate = {
  readonly displayText: string;
  readonly recordId: string;
};

export function WorkbookRecordCandidatePicker({
  candidates,
  disabled = false,
  label,
  onSelectedRecordIdsChange,
  selectedRecordIds,
  testId,
}: {
  readonly candidates: readonly WorkbookRecordCandidate[];
  readonly disabled?: boolean | undefined;
  readonly label: string;
  readonly onSelectedRecordIdsChange: (recordIds: string[]) => void;
  readonly selectedRecordIds: readonly string[];
  readonly testId: string;
}) {
  return (
    <label style={labelStyle}>
      {label}
      <select
        data-testid={testId}
        disabled={disabled}
        multiple
        size={Math.min(Math.max(candidates.length, 2), 5)}
        style={selectStyle}
        value={selectedRecordIds}
        onChange={(event) => {
          onSelectedRecordIdsChange(
            Array.from(event.currentTarget.selectedOptions).map(
              (option) => option.value,
            ),
          );
        }}
      >
        {candidates.map((candidate) => (
          <option key={candidate.recordId} value={candidate.recordId}>
            {candidate.displayText}
          </option>
        ))}
      </select>
    </label>
  );
}

const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const selectStyle = {
  boxSizing: "border-box",
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  appearance: "auto",
} satisfies CSSProperties;
