import {
  type EvidenceAccessContext,
  evidenceAccessMessageTestId,
  evidenceAccessStateTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import {
  workbookFormButtonStyle,
  workbookGridActionButtonStyle,
  workbookTypography,
} from "../../components/workbookFormStyles";
import type {
  EvidenceAccessPresentation,
  EvidenceOperationKind,
} from "../../evidence/evidenceAccessPresentation";

/** Access eligibility is supplied independently for each command by the owner. */
export function EvidenceAccessActions({
  access,
  canRead,
  context,
  onInspect,
  onIssue,
  recordId,
  title,
}: {
  readonly access: EvidenceAccessPresentation;
  readonly canRead: boolean;
  readonly context: EvidenceAccessContext;
  readonly onInspect: () => void;
  readonly onIssue: (
    kind: Exclude<EvidenceOperationKind, "attach">,
    invoker: HTMLButtonElement,
  ) => void;
  readonly recordId: string;
  readonly title: string;
}) {
  const compact = context === "row";
  const messageId = evidenceAccessMessageTestId(recordId, context);
  const controlStyle = compact
    ? workbookGridActionButtonStyle
    : evidenceButtonStyle;
  return (
    <div
      data-testid={evidenceAccessStateTestId(recordId, context)}
      data-evidence-state-key={access.stateKey}
      style={compact ? rowStyle : inspectorStyle}
    >
      {compact ? null : <p style={accessHeadingStyle}>File access</p>}
      <div style={compact ? rowStyle : buttonRowStyle}>
        <button
          type="button"
          data-testid={evidencePreviewButtonTestId(recordId, context)}
          disabled={!canRead}
          aria-disabled={!access.canPreview}
          aria-describedby={messageId}
          style={{
            ...controlStyle,
            ...(!access.canPreview ? unavailableControlStyle : {}),
          }}
          onClick={(event) =>
            access.canPreview &&
            canRead &&
            onIssue("preview", event.currentTarget)
          }
        >
          Preview
        </button>
        <button
          type="button"
          data-testid={evidenceDownloadButtonTestId(recordId, context)}
          disabled={!canRead}
          aria-disabled={!access.canDownload}
          aria-describedby={messageId}
          style={{
            ...controlStyle,
            ...(!access.canDownload ? unavailableControlStyle : {}),
          }}
          onClick={(event) =>
            access.canDownload &&
            canRead &&
            onIssue("download", event.currentTarget)
          }
        >
          Download
        </button>
      </div>
      {compact ? (
        <button
          id={messageId}
          data-testid={messageId}
          type="button"
          style={stateButtonStyle}
          aria-label={`${title}: ${access.message} Inspect evidence details.`}
          onClick={onInspect}
        >
          {access.label}
        </button>
      ) : null}
    </div>
  );
}

export const evidenceButtonStyle = {
  ...workbookFormButtonStyle,
  cursor: "pointer",
} satisfies CSSProperties;
const stateButtonStyle = {
  ...workbookGridActionButtonStyle,
  background: "transparent",
  borderColor: "transparent",
  color: "var(--ct-colors-ink-muted)",
  textDecoration: "underline",
} satisfies CSSProperties;
const rowStyle = {
  display: "flex",
  alignItems: "center",
  gap: "var(--ct-spacing-xs)",
  blockSize: "100%",
  minInlineSize: "max-content",
} satisfies CSSProperties;
const inspectorStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  minInlineSize: 0,
} satisfies CSSProperties;
const buttonRowStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const accessHeadingStyle = {
  ...workbookTypography("section-heading"),
  margin: 0,
} satisfies CSSProperties;
export const evidenceMessageStyle = {
  ...workbookTypography("metadata"),
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const unavailableControlStyle = {
  color: "var(--ct-colors-ink-muted)",
  cursor: "default",
} satisfies CSSProperties;
