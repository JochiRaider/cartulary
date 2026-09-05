import {
  type EvidenceAccessContext,
  evidenceAccessMessageTestId,
  evidenceAccessStateTestId,
  evidenceAttachFileInputTestId,
  evidenceDownloadButtonTestId,
  evidencePreviewButtonTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, useRef } from "react";
import type {
  EvidenceAccessPresentation,
  EvidenceOperationKind,
} from "../../evidence/evidenceAccessPresentation";

export function EvidenceAccessActions({
  access,
  attachDisabledReason,
  attaching,
  canRead,
  context,
  onAttach,
  onInspect,
  onIssue,
  recordId,
  title,
}: {
  readonly access: EvidenceAccessPresentation;
  readonly attachDisabledReason: string | null;
  readonly attaching: boolean;
  readonly canRead: boolean;
  readonly context: EvidenceAccessContext;
  readonly onAttach: (file: File) => void;
  readonly onInspect: () => void;
  readonly onIssue: (
    kind: Exclude<EvidenceOperationKind, "attach">,
    invoker: HTMLButtonElement,
  ) => void;
  readonly recordId: string;
  readonly title: string;
}) {
  const fileInput = useRef<HTMLInputElement>(null);
  const compact = context === "row";
  const messageId = evidenceAccessMessageTestId(recordId, context);
  const controlStyle = compact ? compactButtonStyle : evidenceButtonStyle;
  return (
    <div
      data-testid={evidenceAccessStateTestId(recordId, context)}
      data-evidence-state-key={access.stateKey}
      style={compact ? rowStyle : inspectorStyle}
    >
      {compact ? null : <p style={evidenceMessageStyle}>{title}</p>}
      {compact ? null : (
        <dl style={metadataStyle}>
          <dt>Lifecycle</dt>
          <dd style={valueStyle}>{access.lifecycleLabel}</dd>
          <dt>File</dt>
          <dd style={valueStyle}>{access.uploadLabel}</dd>
        </dl>
      )}
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
        <button
          type="button"
          disabled={!canRead || attachDisabledReason !== null || attaching}
          aria-label={`Attach file to ${title}`}
          aria-busy={attaching || undefined}
          title={attachDisabledReason ?? undefined}
          style={controlStyle}
          onClick={() => fileInput.current?.click()}
        >
          Attach
        </button>
        <input
          ref={fileInput}
          type="file"
          hidden
          aria-label={`Attach file to ${title}`}
          data-testid={evidenceAttachFileInputTestId(recordId, context)}
          disabled={!canRead || attachDisabledReason !== null || attaching}
          accept="image/*,.txt,.pdf,text/plain,application/pdf"
          onChange={(event) => {
            const file = event.currentTarget.files?.[0];
            event.currentTarget.value = "";
            if (file) onAttach(file);
          }}
        />
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
      ) : (
        <>
          <p
            id={messageId}
            data-testid={messageId}
            style={evidenceMessageStyle}
          >
            {access.message}
          </p>
          {attachDisabledReason === null ? null : (
            <p style={evidenceMessageStyle}>{attachDisabledReason}</p>
          )}
        </>
      )}
    </div>
  );
}

export const evidenceButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  font: "inherit",
  cursor: "pointer",
} satisfies CSSProperties;
const compactButtonStyle = {
  ...evidenceButtonStyle,
  padding: "0 var(--ct-spacing-xs)",
  minBlockSize: 0,
  maxBlockSize: "100%",
  lineHeight: "inherit",
  whiteSpace: "nowrap",
} satisfies CSSProperties;
const stateButtonStyle = {
  ...compactButtonStyle,
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
  gap: "var(--ct-spacing-sm)",
  minInlineSize: 0,
} satisfies CSSProperties;
const buttonRowStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
const metadataStyle = {
  display: "grid",
  gridTemplateColumns: "auto minmax(0, 1fr)",
  gap: "var(--ct-spacing-xs) var(--ct-spacing-sm)",
  margin: 0,
} satisfies CSSProperties;
const valueStyle = {
  margin: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
export const evidenceMessageStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;

const unavailableControlStyle = {
  color: "var(--ct-colors-ink-muted)",
  cursor: "default",
} satisfies CSSProperties;
