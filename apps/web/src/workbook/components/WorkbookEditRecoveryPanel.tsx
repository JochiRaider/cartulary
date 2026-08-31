import {
  workbookEditRecoveryDiscardButtonTestId,
  workbookEditRecoveryRetryButtonTestId,
  workbookEditRecoveryTestId,
} from "@cartulary/ui-contracts";
import { type CSSProperties, forwardRef, useRef, useState } from "react";
import type {
  WorkbookEditRecoveryActionResult,
  WorkbookMutationSnapshot,
} from "../runtime/WorkbookMutationRuntime";
import { RecoverySurface } from "./RecoverySurface";

type WorkbookBlockedEdit = NonNullable<WorkbookMutationSnapshot["blockedEdit"]>;

export const WorkbookEditRecoveryPanel = forwardRef<
  HTMLElement,
  {
    readonly blockedEdit: WorkbookBlockedEdit;
    readonly onDiscard: () => Promise<WorkbookEditRecoveryActionResult>;
    readonly onFocusWithinChange: (focused: boolean) => void;
    readonly onRetry: () => Promise<WorkbookEditRecoveryActionResult>;
  }
>(function WorkbookEditRecoveryPanel(
  { blockedEdit, onDiscard, onFocusWithinChange, onRetry },
  ref,
) {
  const [actionMessage, setActionMessage] = useState<string | null>(null);
  const [transitioning, setTransitioning] = useState(false);
  const retryButtonRef = useRef<HTMLButtonElement | null>(null);
  const discardButtonRef = useRef<HTMLButtonElement | null>(null);
  const transitionGuardRef = useRef(false);
  const currentUnitIdRef = useRef(blockedEdit.unitId);
  currentUnitIdRef.current = blockedEdit.unitId;

  const runAction = async (
    action: "discard" | "retry",
    execute: () => Promise<WorkbookEditRecoveryActionResult>,
  ) => {
    if (transitionGuardRef.current) return;
    transitionGuardRef.current = true;
    if (retryButtonRef.current !== null) retryButtonRef.current.disabled = true;
    if (discardButtonRef.current !== null) {
      discardButtonRef.current.disabled = true;
    }
    setTransitioning(true);
    setActionMessage(null);
    const unitId = blockedEdit.unitId;
    try {
      const result = await execute();
      if (!result.ok && currentUnitIdRef.current === unitId) {
        setActionMessage(
          action === "retry"
            ? "The queued edit could not be retried. No queued work was changed."
            : "The blocked edit could not be discarded. No queued work was changed.",
        );
      }
    } finally {
      if (currentUnitIdRef.current === unitId) {
        transitionGuardRef.current = false;
        if (retryButtonRef.current !== null) {
          retryButtonRef.current.disabled = false;
        }
        if (discardButtonRef.current !== null) {
          discardButtonRef.current.disabled = false;
        }
        setTransitioning(false);
      }
    }
  };

  return (
    <RecoverySurface
      aria-label="Workbook edit recovery"
      data-testid={workbookEditRecoveryTestId()}
      ref={ref}
      tabIndex={-1}
      onBlurCapture={(event) => {
        if (transitionGuardRef.current) return;
        if (!event.currentTarget.contains(event.relatedTarget as Node | null)) {
          onFocusWithinChange(false);
        }
      }}
      onFocusCapture={() => onFocusWithinChange(true)}
    >
      <div>
        <p style={eyebrowStyle}>Local edit needs attention</p>
        <h2 style={titleStyle}>Queued edits</h2>
      </div>
      <p
        aria-atomic="true"
        aria-label="Queued edit recovery message"
        aria-live="polite"
        key={blockedEdit.unitId}
        role="status"
        style={bodyStyle}
      >
        {actionMessage ?? blockedEdit.message}
      </p>
      <div style={buttonRowStyle}>
        {blockedEdit.kind === "client_txn_conflict" ? (
          <button
            data-testid={workbookEditRecoveryRetryButtonTestId()}
            disabled={transitioning}
            onClick={() => void runAction("retry", onRetry)}
            ref={retryButtonRef}
            style={secondaryButtonStyle}
            type="button"
          >
            Retry with a new request ID
          </button>
        ) : null}
        <button
          data-testid={workbookEditRecoveryDiscardButtonTestId()}
          disabled={transitioning}
          onClick={() => void runAction("discard", onDiscard)}
          ref={discardButtonRef}
          style={destructiveButtonStyle}
          type="button"
        >
          Discard blocked edit
        </button>
      </div>
    </RecoverySurface>
  );
});

const eyebrowStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  fontSize: "0.78rem",
  fontWeight: 800,
  letterSpacing: "0.08em",
  textTransform: "uppercase",
} satisfies CSSProperties;

const titleStyle = {
  margin: "0.2rem 0 0",
  overflowWrap: "anywhere",
} satisfies CSSProperties;

const bodyStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;

const buttonRowStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
  minInlineSize: 0,
} satisfies CSSProperties;

const baseButtonStyle = {
  maxInlineSize: "100%",
  minInlineSize: 0,
  padding: "var(--ct-component-button-secondary-padding)",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
  overflowWrap: "anywhere",
  whiteSpace: "normal",
} satisfies CSSProperties;

const secondaryButtonStyle = {
  ...baseButtonStyle,
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
} satisfies CSSProperties;

const destructiveButtonStyle = {
  ...baseButtonStyle,
  border: "1px solid var(--ct-colors-semantic-destructive)",
  background: "var(--ct-component-button-danger-backgroundColor)",
  color: "var(--ct-component-button-danger-textColor)",
} satisfies CSSProperties;
