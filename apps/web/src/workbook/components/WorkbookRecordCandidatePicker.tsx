import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { type Ref, useId } from "react";
import {
  workbookReferenceIdentityStyle,
  workbookTypography,
} from "./workbookFormStyles";

type WorkbookRecordCandidate = {
  readonly displayText: string;
  readonly recordId: string;
};

/** Native single-choice presentation; the caller owns selection and acceptance. */
export function WorkbookRecordCandidatePicker({
  candidates,
  disabled = false,
  disabledRecordIds = [],
  label,
  onSelect,
  selectedRecordId,
  focusTargetRef,
  testId,
}: {
  readonly candidates: readonly WorkbookRecordCandidate[];
  readonly disabled?: boolean | undefined;
  readonly disabledRecordIds?: readonly string[];
  readonly label: string;
  readonly onSelect: (recordId: string) => void;
  readonly selectedRecordId: string | null;
  readonly focusTargetRef?: Ref<HTMLInputElement>;
  readonly testId: string;
}) {
  const name = useId();
  const firstEnabled = disabled
    ? undefined
    : candidates.find(
        (candidate) => !disabledRecordIds.includes(candidate.recordId),
      )?.recordId;
  return (
    <fieldset
      aria-label={label}
      data-testid={testId}
      data-workbook-single-candidates
      disabled={disabled}
      style={{ margin: 0, padding: 0, border: 0, minInlineSize: 0 }}
    >
      <legend style={{ ...workbookTypography("ui"), padding: 0 }}>
        {label}
      </legend>
      <div
        style={{
          maxBlockSize: `${cartularyDesignPresentation.inspector.candidateChooserMaxBlockSizeRem}rem`,
          overflow: "auto",
          overscrollBehavior: "contain",
          minInlineSize: 0,
          border: "var(--ct-component-text-input-border)",
          borderRadius: "var(--ct-component-text-input-rounded)",
          background: "var(--ct-component-text-input-backgroundColor)",
          color: "var(--ct-component-text-input-textColor)",
        }}
      >
        {candidates.map((candidate) => (
          <label
            key={candidate.recordId}
            style={{
              display: "flex",
              alignItems: "start",
              gap: "var(--ct-spacing-xs)",
              padding: "var(--ct-spacing-xs)",
              minInlineSize: 0,
              ...workbookTypography("ui"),
            }}
          >
            <input
              ref={
                candidate.recordId === firstEnabled ? focusTargetRef : undefined
              }
              type="radio"
              name={name}
              value={candidate.recordId}
              checked={selectedRecordId === candidate.recordId}
              disabled={
                disabled || disabledRecordIds.includes(candidate.recordId)
              }
              aria-label={`${candidate.displayText || "Unnamed reference"} (${candidate.recordId})`}
              onChange={(event) => {
                if (event.currentTarget.checked) onSelect(candidate.recordId);
              }}
              onKeyDown={(event) => {
                if (event.key === "Enter") event.preventDefault();
              }}
              style={{
                flex: "0 0 auto",
                accentColor: "var(--ct-colors-ink)",
                margin: "var(--ct-spacing-xs)",
              }}
            />
            <span
              style={{
                minInlineSize: 0,
                overflowWrap: "anywhere",
                whiteSpace: "pre-wrap",
              }}
            >
              <span>{candidate.displayText || "Unnamed reference"}</span>
              <span
                data-generated-metadata="candidate-reference-id"
                style={workbookReferenceIdentityStyle}
              >
                {candidate.recordId}
              </span>
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}
