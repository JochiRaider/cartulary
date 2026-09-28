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
  return (
    <div
      data-testid={evidenceAccessStateTestId(recordId, context)}
      data-evidence-state-key={access.stateKey}
      style={compact ? rowStyle : inspectorStyle}
    >
      {compact ? null : <p style={accessHeadingStyle}>File access</p>}
      <EvidenceHandleButtons
        canRead={canRead}
        canPreview={access.canPreview}
        canDownload={access.canDownload}
        retryKind={access.retryKind}
        busyKind={access.busyKind}
        context={context}
        descriptionId={messageId}
        recordId={recordId}
        onIssue={onIssue}
      />
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

/** Evidence-owned access affordances shared with linked-record presentations. */
export function EvidenceHandleButtons({
  canRead,
  canPreview,
  canDownload,
  retryKind,
  busyKind,
  context,
  descriptionId,
  labelSuffix,
  recordId,
  onIssue,
}: {
  readonly canRead: boolean;
  readonly canPreview: boolean;
  readonly canDownload: boolean;
  readonly retryKind?: "preview" | "download" | null;
  readonly busyKind?: "preview" | "download" | null;
  readonly context: EvidenceAccessContext;
  readonly descriptionId?: string;
  readonly labelSuffix?: string;
  readonly recordId: string;
  readonly onIssue: (
    kind: Exclude<EvidenceOperationKind, "attach">,
    invoker: HTMLButtonElement,
  ) => void;
}) {
  const controlStyle =
    context === "row" ? workbookGridActionButtonStyle : evidenceButtonStyle;
  return (
    <div style={context === "row" ? rowStyle : buttonRowStyle}>
      {(["preview", "download"] as const).map((kind) => {
        const allowed = kind === "preview" ? canPreview : canDownload;
        const label = `${retryKind === kind ? "Retry " : ""}${kind === "preview" ? "Preview" : "Download"}`;
        return (
          <button
            key={kind}
            type="button"
            data-testid={
              kind === "preview"
                ? evidencePreviewButtonTestId(recordId, context)
                : evidenceDownloadButtonTestId(recordId, context)
            }
            disabled={!canRead}
            aria-disabled={!allowed}
            aria-busy={busyKind === kind ? true : undefined}
            aria-describedby={descriptionId}
            aria-label={labelSuffix ? `${label} ${labelSuffix}` : undefined}
            style={{
              ...controlStyle,
              ...(!allowed ? unavailableControlStyle : {}),
            }}
            onClick={(event) => {
              if (canRead && allowed) onIssue(kind, event.currentTarget);
            }}
          >
            {label}
          </button>
        );
      })}
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
