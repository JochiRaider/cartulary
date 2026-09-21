import {
  type CSSProperties,
  type KeyboardEvent,
  type ReactNode,
  useEffect,
  useRef,
} from "react";
import { useWorkbookInspectorNotice } from "../useWorkbookInspectorNotice";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
  WorkbookInspectorNotice,
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
  announce = true,
  testId,
}: {
  readonly announce?: boolean;
  readonly error: WorkbookInspectorErrorPresentation;
  readonly testId?: string | undefined;
}) {
  return (
    <div
      aria-live={announce ? "assertive" : undefined}
      data-testid={testId}
      role={announce ? "alert" : undefined}
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
  announce = true,
}: {
  readonly feedback: WorkbookInspectorFeedback | null;
  readonly neutralStyle?: CSSProperties | undefined;
  readonly testId?: string | undefined;
  readonly announce?: boolean;
}) {
  if (feedback === null) return null;
  if (feedback.kind === "error") {
    return (
      <WorkbookInspectorPublicError
        error={feedback.error}
        testId={testId}
        announce={announce}
      />
    );
  }
  return (
    <p
      aria-live={
        announce && feedback.announcement === "polite" ? "polite" : undefined
      }
      data-testid={testId}
      role={
        announce && feedback.announcement === "polite" ? "status" : undefined
      }
      style={neutralStyle}
    >
      {feedback.message}
    </p>
  );
}

export function WorkbookInspectorNoticeView({
  notice,
  consume,
  visible = true,
}: {
  readonly notice: WorkbookInspectorNotice;
  readonly consume: (notice: WorkbookInspectorNotice) => boolean;
  readonly visible?: boolean;
}) {
  const emission = useWorkbookInspectorNotice(notice, consume);
  const message =
    emission?.feedback.kind === "error"
      ? emission.feedback.error.primaryMessage
      : emission?.feedback.message;
  return (
    <>
      {visible ? (
        <WorkbookInspectorFeedbackView
          feedback={notice.feedback}
          announce={false}
        />
      ) : null}
      <span
        style={{
          position: "absolute",
          inlineSize: 1,
          blockSize: 1,
          overflow: "hidden",
          clipPath: "inset(50%)",
        }}
        role={emission?.announcement === "assertive" ? "alert" : "status"}
        aria-live={
          emission?.announcement === "assertive" ? "assertive" : "polite"
        }
        aria-atomic="true"
      >
        <span
          key={
            emission ? `${emission.attemptId}:${emission.transitionId}` : "idle"
          }
        >
          {message}
        </span>
      </span>
    </>
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
  useEffect(() => {
    safeControlRef.current?.focus({ preventScroll: true });
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
