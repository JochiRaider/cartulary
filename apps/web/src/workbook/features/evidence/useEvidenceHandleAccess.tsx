import {
  evidencePreviewFrameTestId,
  evidencePreviewPanelTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  useCallback,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import {
  boundedRead,
  ObservationStopped,
} from "../../../services/asyncObservation";
import {
  type EvidenceOperationState,
  evidenceOperationFeedback,
} from "../../evidence/evidenceAccessPresentation";
import { workbookSurfaceOverlayPanelStyle } from "../../layout/WorkbookSurfaceLayout";
import type {
  EvidenceCapabilityPort,
  EvidenceHandleOutcome,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import {
  evidenceButtonStyle,
  evidenceMessageStyle,
} from "./EvidenceAccessActions";

export type EvidenceAccessTarget = Readonly<{
  recordId: string;
  identity: string;
  title: string;
  sourceRecordId?: string;
}>;

type Ticket = Readonly<{
  target: EvidenceAccessTarget;
  lifetime: number;
  sequence: number;
  kind: "preview" | "download";
  controller: AbortController;
}>;
type Preview = Readonly<{
  ticket: Ticket;
  href: string | null;
  invoker: HTMLElement | null;
}>;
type Operation = Readonly<{
  identity: string;
  sequence: number;
  state: EvidenceOperationState;
}>;

const unknownFailure: WorkbookOperationFailure = {
  kind: "terminal",
  message: "Evidence request failed.",
};

function accessWasLost(failure: WorkbookOperationFailure) {
  return (
    failure.kind === "authentication_required" ||
    failure.kind === "authorization_lost" ||
    failure.presentation?.family === "permission_or_incident_access_loss"
  );
}

function usableInvoker(element: HTMLElement | null): element is HTMLElement {
  return (
    !!element &&
    element.isConnected &&
    element.getClientRects().length > 0 &&
    !element.closest("[inert], [hidden], [aria-hidden='true']")
  );
}

/** Shared Evidence owner for access requests and preview/download effects. */
export function useEvidenceHandleAccess(input: {
  readonly port: EvidenceCapabilityPort;
  readonly canRead: boolean;
  readonly scopeKey: string;
  readonly isCurrent: (target: EvidenceAccessTarget) => boolean;
  readonly onAccessFailure: () => Promise<void> | void;
  readonly onRestoreFocus: (target: EvidenceAccessTarget) => void;
}) {
  const inputRef = useRef(input);
  inputRef.current = input;
  const lifetime = useRef(0);
  const sequence = useRef(0);
  const denied = useRef(false);
  const [accessLost, setAccessLost] = useState(false);
  const pending = useRef(new Map<string, Ticket>());
  const latestFeedback = useRef(new Map<string, number>());
  const previewRef = useRef<Preview | null>(null);
  const previewPanelRef = useRef<HTMLElement | null>(null);
  const closeButtonRef = useRef<HTMLButtonElement | null>(null);
  const [preview, setPreview] = useState<Preview | null>(null);
  const [operations, setOperations] = useState<Record<string, Operation>>({});
  const [announcement, setAnnouncement] = useState<{
    text: string;
    sequence: number;
    priority: "polite" | "assertive";
  } | null>(null);

  const targetCurrent = useCallback(
    (ticket: Ticket) =>
      ticket.lifetime === lifetime.current &&
      !ticket.controller.signal.aborted &&
      inputRef.current.canRead &&
      !denied.current &&
      inputRef.current.isCurrent(ticket.target),
    [],
  );

  const clearPreview = useCallback((restore = false) => {
    const old = previewRef.current;
    const ownsFocus =
      previewPanelRef.current?.contains(document.activeElement) ?? false;
    previewRef.current = null;
    setPreview(null);
    if (old === null) return;
    old.ticket.controller.abort();
    if (
      pending.current.get(old.ticket.target.recordId)?.sequence ===
      old.ticket.sequence
    )
      pending.current.delete(old.ticket.target.recordId);
    setOperations((current) => {
      if (current[old.ticket.target.recordId]?.sequence !== old.ticket.sequence)
        return current;
      const next = { ...current };
      delete next[old.ticket.target.recordId];
      return next;
    });
    if (
      latestFeedback.current.get(old.ticket.target.recordId) ===
      old.ticket.sequence
    )
      latestFeedback.current.delete(old.ticket.target.recordId);
    if (restore && ownsFocus) {
      if (usableInvoker(old.invoker)) old.invoker.focus();
      else inputRef.current.onRestoreFocus(old.ticket.target);
    }
  }, []);
  const reset = useCallback(() => {
    lifetime.current += 1;
    for (const ticket of pending.current.values()) ticket.controller.abort();
    pending.current.clear();
    latestFeedback.current.clear();
    previewRef.current = null;
    setPreview(null);
    setOperations({});
    setAnnouncement(null);
    denied.current = false;
    setAccessLost(false);
  }, []);
  const observedScope = useRef({
    scopeKey: input.scopeKey,
    canRead: input.canRead,
  });
  useLayoutEffect(() => {
    const old = observedScope.current;
    observedScope.current = {
      scopeKey: input.scopeKey,
      canRead: input.canRead,
    };
    if (old.scopeKey !== input.scopeKey || old.canRead !== input.canRead)
      reset();
  }, [input.scopeKey, input.canRead, reset]);
  useLayoutEffect(() => {
    const current = previewRef.current;
    if (current && !input.isCurrent(current.ticket.target)) clearPreview();
    for (const [key, ticket] of pending.current) {
      if (!input.isCurrent(ticket.target)) {
        ticket.controller.abort();
        pending.current.delete(key);
      }
    }
    setOperations((current) => {
      const kept = Object.entries(current).filter(([recordId, operation]) =>
        input.isCurrent({ recordId, identity: operation.identity, title: "" }),
      );
      return kept.length === Object.keys(current).length
        ? current
        : Object.fromEntries(kept);
    });
  });
  useLayoutEffect(
    () => () => {
      lifetime.current += 1;
      for (const ticket of pending.current.values()) ticket.controller.abort();
      pending.current.clear();
      previewRef.current = null;
    },
    [],
  );

  const publish = useCallback(
    (ticket: Ticket, state: EvidenceOperationState) => {
      if (
        !targetCurrent(ticket) ||
        latestFeedback.current.get(ticket.target.recordId) !== ticket.sequence
      )
        return;
      setOperations((current) => ({
        ...current,
        [ticket.target.recordId]: {
          identity: ticket.target.identity,
          sequence: ticket.sequence,
          state,
        },
      }));
      const feedback = evidenceOperationFeedback(state);
      if (feedback.announcement !== "none")
        setAnnouncement({
          text: `${ticket.target.title}: ${feedback.message}`,
          sequence: ticket.sequence,
          priority: feedback.announcement,
        });
    },
    [targetCurrent],
  );

  const issue = useCallback(
    async (
      target: EvidenceAccessTarget,
      kind: "preview" | "download",
      invoker: HTMLElement | null,
    ): Promise<EvidenceHandleOutcome | null> => {
      if (
        !inputRef.current.canRead ||
        denied.current ||
        !inputRef.current.isCurrent(target)
      )
        return null;
      const key = target.recordId;
      const prior = pending.current.get(key);
      if (
        prior &&
        prior.kind === kind &&
        prior.target.identity === target.identity
      )
        return null;
      prior?.controller.abort();
      pending.current.delete(key);
      if (
        prior?.kind === "preview" &&
        previewRef.current?.ticket.sequence === prior.sequence
      )
        clearPreview();
      if (kind === "preview") clearPreview();
      const ticket: Ticket = {
        target,
        kind,
        lifetime: lifetime.current,
        sequence: ++sequence.current,
        controller: new AbortController(),
      };
      pending.current.set(key, ticket);
      latestFeedback.current.set(target.recordId, ticket.sequence);
      if (kind === "preview") {
        const value = { ticket, href: null, invoker };
        previewRef.current = value;
        setPreview(value);
      }
      publish(ticket, { kind: "pending", operation: kind });
      let outcome: EvidenceHandleOutcome | null = null;
      let deadline = false;
      try {
        outcome = await boundedRead(
          (signal) =>
            inputRef.current.port.issueHandle({
              evidenceRecordId: target.recordId,
              kind,
              signal,
            }),
          ticket.controller.signal,
        );
      } catch (error) {
        if (error instanceof ObservationStopped) {
          if (error.reason === "aborted") return null;
          deadline = true;
        } else outcome = { kind: "rejected", failure: unknownFailure };
      }
      if (
        !targetCurrent(ticket) ||
        pending.current.get(key)?.sequence !== ticket.sequence
      )
        return null;
      pending.current.delete(key);
      if (deadline) {
        if (kind === "preview") {
          const ownsFocus =
            previewPanelRef.current?.contains(document.activeElement) ?? false;
          previewRef.current = null;
          setPreview(null);
          if (ownsFocus) {
            if (usableInvoker(invoker)) invoker.focus();
            else inputRef.current.onRestoreFocus(target);
          }
        }
        publish(ticket, { kind: "deadline", operation: kind });
        return null;
      }
      if (outcome === null) return null;
      if (outcome.kind === "rejected") {
        if (kind === "preview") {
          const restoreFocus =
            previewPanelRef.current?.contains(document.activeElement) ?? false;
          previewRef.current = null;
          setPreview(null);
          if (restoreFocus) {
            if (usableInvoker(invoker)) invoker.focus();
            else inputRef.current.onRestoreFocus(target);
          }
        }
        publish(ticket, {
          kind: "rejected",
          operation: kind,
          failure: outcome.failure,
        });
        if (accessWasLost(outcome.failure)) {
          denied.current = true;
          reset();
          denied.current = true;
          setAccessLost(true);
          void inputRef.current.onAccessFailure();
        }
        return outcome;
      }
      if (kind === "preview") {
        const next = { ticket, href: outcome.value.href, invoker };
        previewRef.current = next;
        setPreview(next);
      } else {
        const anchor = document.createElement("a");
        anchor.href = outcome.value.href;
        anchor.download = outcome.value.filename || "evidence";
        anchor.rel = "noopener";
        document.body.append(anchor);
        anchor.click();
        anchor.remove();
      }
      publish(ticket, { kind: "accepted", operation: kind });
      return outcome;
    },
    [clearPreview, publish, reset, targetCurrent],
  );

  const visiblePreview =
    preview !== null &&
    input.canRead &&
    !denied.current &&
    preview.ticket.lifetime === lifetime.current &&
    input.isCurrent(preview.ticket.target);
  const visiblePreviewSequence = visiblePreview
    ? preview.ticket.sequence
    : null;
  useLayoutEffect(() => {
    if (visiblePreviewSequence !== null)
      closeButtonRef.current?.focus({ preventScroll: true });
  }, [visiblePreviewSequence]);
  const overlay = visiblePreview ? (
    <section
      ref={previewPanelRef}
      data-testid={evidencePreviewPanelTestId()}
      aria-label={`Evidence preview: ${preview.ticket.target.title}`}
      style={previewPanelStyle}
    >
      <div style={previewHeaderStyle}>
        <h2 style={previewTitleStyle}>{preview.ticket.target.title}</h2>
        <button
          ref={closeButtonRef}
          type="button"
          style={evidenceButtonStyle}
          onClick={() => clearPreview(true)}
        >
          Close
        </button>
      </div>
      {preview.href === null ? (
        <p style={evidenceMessageStyle}>Opening preview…</p>
      ) : (
        <iframe
          key={preview.ticket.sequence}
          data-testid={evidencePreviewFrameTestId(
            preview.ticket.target.recordId,
          )}
          src={preview.href}
          title={`Evidence preview ${preview.ticket.target.title}`}
          style={previewFrameStyle}
        />
      )}
    </section>
  ) : null;
  const announce = (text: string, priority: "polite" | "assertive") =>
    setAnnouncement({ text, sequence: ++sequence.current, priority });
  return {
    operations,
    announcement,
    issue,
    overlay,
    closePreview: visiblePreview ? () => clearPreview(true) : undefined,
    reset,
    announce,
    accessLost,
  };
}

const previewPanelStyle = {
  ...workbookSurfaceOverlayPanelStyle,
  zIndex: 9,
  display: "grid",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-panel-padding)",
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-popover)",
} satisfies CSSProperties;
const previewHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "start",
  gap: "var(--ct-spacing-sm)",
  minInlineSize: 0,
} satisfies CSSProperties;
const previewTitleStyle = {
  margin: 0,
  fontSize: "var(--ct-typography-section-heading-fontSize)",
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const previewFrameStyle = {
  inlineSize: "100%",
  blockSize: "min(28rem, 34vh)",
  minBlockSize: 0,
  border: "var(--ct-border-hairline)",
  background: "Canvas",
  colorScheme: "light",
} satisfies CSSProperties;
