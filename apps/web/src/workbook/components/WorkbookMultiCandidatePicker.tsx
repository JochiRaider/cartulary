import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import type { Ref } from "react";
import { workbookTypography } from "./workbookFormStyles";

/** Presentation only: each native activation names one identity and its new state. */
export function WorkbookMultiCandidatePicker({
  candidates,
  selectedKeys,
  label,
  disabled,
  testId,
  focusTargetRef,
  onToggle,
}: {
  readonly candidates: readonly {
    readonly key: string;
    readonly displayText: string;
    readonly identityText: string;
    readonly disabled?: boolean;
  }[];
  readonly selectedKeys: readonly string[];
  readonly label: string;
  readonly disabled: boolean;
  readonly testId?: string;
  readonly focusTargetRef?: Ref<HTMLInputElement>;
  readonly onToggle: (key: string, checked: boolean) => void;
}) {
  const firstEnabled = disabled
    ? undefined
    : candidates.find((candidate) => !candidate.disabled)?.key;
  return (
    <fieldset
      aria-label={label}
      data-testid={testId}
      data-workbook-multi-candidates
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
            key={candidate.key}
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
              ref={candidate.key === firstEnabled ? focusTargetRef : undefined}
              type="checkbox"
              value={candidate.key}
              checked={selectedKeys.includes(candidate.key)}
              disabled={disabled || candidate.disabled}
              aria-label={`${candidate.displayText || candidate.identityText} (${candidate.identityText})`}
              onChange={(event) =>
                onToggle(candidate.key, event.currentTarget.checked)
              }
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
                whiteSpace: "normal",
              }}
            >
              <span>{candidate.displayText || candidate.identityText}</span>
              <span
                data-generated-metadata="candidate-reference-id"
                style={{
                  display: "block",
                  ...workbookTypography("metadata"),
                  color: "var(--ct-colors-ink-muted)",
                }}
              >
                {candidate.identityText}
              </span>
            </span>
          </label>
        ))}
      </div>
    </fieldset>
  );
}
