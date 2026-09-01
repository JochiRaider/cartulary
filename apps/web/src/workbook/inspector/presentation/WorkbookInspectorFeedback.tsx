import {
  type CSSProperties,
  type KeyboardEvent,
  type ReactNode,
  useEffect,
  useRef,
} from "react";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../workbookInspectorErrorModel";
import { WorkbookInspectorActionButton } from "./WorkbookInspectorActions";
import type { WorkbookInspectorTechnicalField } from "./workbookInspectorPresentationModel";

export function WorkbookInspectorCompactMetadata({
  children,
}: {
  readonly children: ReactNode;
}) {
  return <div style={compactMetadataStyle}>{children}</div>;
}

export function WorkbookInspectorTechnicalDetails({
  fields,
}: {
  readonly fields: readonly WorkbookInspectorTechnicalField[];
}) {
  if (fields.length === 0) return null;
  return (
    <details style={technicalDetailsStyle}>
      <summary>Technical details</summary>
      <dl style={technicalListStyle}>
        {fields.map((field) => (
          <div key={field.label}>
            <dt style={technicalTermStyle}>{field.label}</dt>
            <dd style={technicalValueStyle}>{field.value}</dd>
          </div>
        ))}
      </dl>
    </details>
  );
}

export function WorkbookInspectorPublicError({
  error,
  testId,
}: {
  readonly error: WorkbookInspectorErrorPresentation;
  readonly testId?: string | undefined;
}) {
  return (
    <div
      aria-live="assertive"
      data-testid={testId}
      role="alert"
      style={publicErrorStyle}
    >
      <p style={messageStyle}>{error.primaryMessage}</p>
      <WorkbookInspectorTechnicalDetails fields={error.technicalFields} />
    </div>
  );
}

export function WorkbookInspectorFeedbackView({
  feedback,
  neutralStyle,
  testId,
}: {
  readonly feedback: WorkbookInspectorFeedback | null;
  readonly neutralStyle?: CSSProperties | undefined;
  readonly testId?: string | undefined;
}) {
  if (feedback === null) return null;
  if (feedback.kind === "error") {
    return (
      <WorkbookInspectorPublicError error={feedback.error} testId={testId} />
    );
  }
  return (
    <p
      aria-live={feedback.announcement === "polite" ? "polite" : undefined}
      data-testid={testId}
      role={feedback.announcement === "polite" ? "status" : undefined}
      style={neutralStyle}
    >
      {feedback.message}
    </p>
  );
}

export function WorkbookInspectorConfirmation({
  cancelLabel = "Cancel",
  cancelTestId,
  confirmLabel,
  confirmTestId,
  destructive = false,
  onCancel,
  onConfirm,
  operation,
  subject,
  technicalFields = [],
  testId,
}: {
  readonly cancelLabel?: string | undefined;
  readonly cancelTestId?: string | undefined;
  readonly confirmLabel: string;
  readonly confirmTestId?: string | undefined;
  readonly destructive?: boolean | undefined;
  readonly onCancel: () => void;
  readonly onConfirm: () => void;
  readonly operation: string;
  readonly subject: string;
  readonly technicalFields?: readonly WorkbookInspectorTechnicalField[];
  readonly testId?: string | undefined;
}) {
  const safeControlRef = useRef<HTMLButtonElement>(null);
  const invokingControlRef = useRef<HTMLElement | null>(null);
  useEffect(() => {
    invokingControlRef.current =
      document.activeElement instanceof HTMLElement
        ? document.activeElement
        : null;
    safeControlRef.current?.focus({ preventScroll: true });
    return () => {
      if (invokingControlRef.current?.isConnected) {
        invokingControlRef.current.focus({ preventScroll: true });
      }
    };
  }, []);
  const cancel = () => onCancel();
  const handleKeyDown = (event: KeyboardEvent<HTMLDivElement>) => {
    if (event.key !== "Escape") return;
    event.preventDefault();
    event.stopPropagation();
    cancel();
  };
  return (
    <div
      aria-label={`${operation} confirmation`}
      data-testid={testId}
      role="alertdialog"
      style={confirmationStyle}
      onKeyDown={handleKeyDown}
    >
      <p style={confirmationTextStyle}>
        {operation} <strong>{subject}</strong>?
      </p>
      <WorkbookInspectorTechnicalDetails fields={technicalFields} />
      <div style={confirmationActionsStyle}>
        <WorkbookInspectorActionButton
          data-testid={cancelTestId}
          ref={safeControlRef}
          tone="secondary"
          onClick={cancel}
        >
          {cancelLabel}
        </WorkbookInspectorActionButton>
        <WorkbookInspectorActionButton
          data-testid={confirmTestId}
          tone={destructive ? "destructive" : "primary"}
          onClick={onConfirm}
        >
          {confirmLabel}
        </WorkbookInspectorActionButton>
      </div>
    </div>
  );
}

const compactMetadataStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "var(--ct-spacing-xs)",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-compact-metadata-fontSize)",
} satisfies CSSProperties;
const messageStyle = { margin: 0 } satisfies CSSProperties;
const publicErrorStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  color: "var(--ct-colors-semantic-conflict)",
} satisfies CSSProperties;
const technicalDetailsStyle = {
  color: "var(--ct-colors-ink-muted)",
  fontSize: "var(--ct-typography-compact-metadata-fontSize)",
} satisfies CSSProperties;
const technicalListStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  marginBlockEnd: 0,
} satisfies CSSProperties;
const technicalTermStyle = {
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;
const technicalValueStyle = {
  margin: 0,
  fontFamily: "var(--ct-typography-mono-fontFamily)",
  overflowWrap: "anywhere" as const,
} satisfies CSSProperties;
const confirmationStyle = {
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderColor: "var(--ct-colors-semantic-caution)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-2)",
} satisfies CSSProperties;
const confirmationTextStyle = { margin: 0 } satisfies CSSProperties;
const confirmationActionsStyle = {
  display: "flex",
  justifyContent: "flex-end",
  gap: "var(--ct-spacing-xs)",
} satisfies CSSProperties;
