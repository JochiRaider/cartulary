import { incidentAdministrationTestId } from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type ReactNode,
  type RefObject,
  useId,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { IncidentResource } from "../shared/incidentResource";
import type {
  WorkbookDensityMode,
  WorkbookIncidentControlsRendererProps,
} from "../shared/workbookShellContracts";
import { AccountDialog } from "./AccountDialog";
import { IncidentAdminPanel } from "./IncidentAdminPanel";
import type { IncidentLifecycleController } from "./incidentLifecycleController";
import { lifecycleProblemMessage } from "./incidentLifecycleModel";
import { primaryButtonStyle, secondaryButtonStyle } from "./landingAdminStyles";

export function IncidentLifecycleFeature({
  controller,
  bindSurface,
  acceptedIncident,
  onIncidentObserved,
  preferenceControls,
  ...surface
}: WorkbookIncidentControlsRendererProps & {
  controller: IncidentLifecycleController;
  preferenceControls?: ReactNode;
  bindSurface: (surface: WorkbookIncidentControlsRendererProps | null) => void;
  acceptedIncident: IncidentResource | null;
  onIncidentObserved: (resource: IncidentResource) => void;
}) {
  const latest = useRef({ bindSurface, surface });
  latest.current = { bindSurface, surface };
  useLayoutEffect(() => bindSurface(surface));
  useLayoutEffect(() => {
    const visibility = () => latest.current.bindSurface(latest.current.surface);
    document.addEventListener("visibilitychange", visibility);
    return () => {
      document.removeEventListener("visibilitychange", visibility);
      latest.current.bindSurface(null);
    };
  }, []);
  return (
    <IncidentAdminPanel
      {...surface}
      acceptedIncident={acceptedIncident}
      preferenceControls={preferenceControls}
      onIncidentObserved={onIncidentObserved}
      lifecycleControls={
        <IncidentLifecyclePanel
          controller={controller}
          density={surface.density}
        />
      }
    />
  );
}
const stack: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
  minWidth: 0,
};
const actions: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
  alignItems: "center",
  minWidth: 0,
};
const text: CSSProperties = {
  margin: 0,
  overflowWrap: "anywhere",
  whiteSpace: "pre-wrap",
};
const muted: CSSProperties = { ...text, color: "var(--ct-colors-ink-muted)" };
const card: CSSProperties = {
  ...stack,
  padding: "var(--ct-spacing-md)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
};
const css = `
[data-incident-lifecycle] :is(button,textarea):focus-visible { outline: var(--ct-border-focus); outline-offset: var(--ct-spacing-xs); }
[data-incident-lifecycle] :is(button,textarea) { font: inherit; max-inline-size: 100%; }
[data-incident-lifecycle] button[aria-disabled="true"] { cursor: not-allowed; color: var(--ct-colors-ink-muted) !important; background: var(--ct-colors-surface-2) !important; border: var(--ct-border-hairline) !important; }
`;
export function IncidentLifecyclePanel({
  controller,
  density = "default",
}: {
  controller: IncidentLifecycleController;
  density?: WorkbookDensityMode | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const prefix = useId();
  const heading = useRef<HTMLHeadingElement>(null);
  const refresh = useRef<HTMLButtonElement>(null);
  const lastFocus = useRef<HTMLElement | null>(null);
  useLayoutEffect(() => {
    if (
      state.active &&
      document.visibilityState !== "hidden" &&
      document.activeElement === document.body &&
      lastFocus.current &&
      !lastFocus.current.isConnected
    )
      refresh.current?.focus();
  });
  const { draft, resource, operation, review } = state;
  const pending = operation.kind === "pending";
  const admin = state.authority?.role === "admin";
  const hasAttempt = "attempt" in operation;
  const actionLabel = hasAttempt
    ? operation.attempt.action === "close"
      ? "Close"
      : "Reopen"
    : "Action";
  const problem = "problem" in operation ? operation.problem : undefined;
  const reviewResource =
    review?.resource ?? (pending ? operation.attempt.resource : null);
  const reviewedAction =
    review?.action ?? (pending ? operation.attempt.action : null);
  const status = pending
    ? operation.stage === "authorization"
      ? "Checking current administrator access…"
      : `${actionLabel} request sent. Waiting for acknowledgement…`
    : operation.kind === "confirmed"
      ? `${actionLabel} confirmed.`
      : operation.kind === "uncertain"
        ? `${actionLabel} result is uncertain. Current incident state cannot establish whether this action succeeded.`
        : operation.kind === "rejected"
          ? lifecycleProblemMessage(operation.problem)
          : "";
  const blocked = !admin
    ? "Only current incident admins can close, reopen or replay. Other current members retain authorized reads."
    : pending
      ? "One lifecycle action is already being checked or sent."
      : state.transportPending
        ? "The original dispatch has not settled. A timeout cannot establish server cancellation; another lifecycle dispatch remains unavailable."
        : operation.kind === "uncertain" ||
            problem?.code === "client_txn_conflict"
          ? "Resolve or explicitly forget this recovery before reviewing a new action."
          : state.access !== "ready"
            ? "Current administrator access must be checked before acting."
            : state.read !== "ready"
              ? "Refresh current incident state before reviewing a new action."
              : draft.reason === ""
                ? "A reason is required before confirmation."
                : "";
  return (
    <section
      data-incident-lifecycle=""
      data-density={density}
      data-testid={incidentAdministrationTestId("lifecycle-panel")}
      aria-labelledby={`${prefix}-heading`}
      onFocusCapture={(event) => {
        lastFocus.current = event.target as HTMLElement;
      }}
      style={{
        ...card,
        padding: `var(--ct-density-${density}-cellPadding)`,
        fontSize: `var(--ct-density-${density}-fontSize)`,
      }}
    >
      <style>{css}</style>
      <h3 ref={heading} id={`${prefix}-heading`} tabIndex={-1} style={text}>
        Incident lifecycle
      </h3>
      <p style={muted}>
        Close makes incident source data read-only. Current members keep
        authorized reads, membership administration, saved views and workbook
        preferences under their ordinary permissions.
      </p>
      <p style={muted}>
        Closing does not save or discard local work. Reopen permits new
        authorized source-data writes; retained rejected drafts require a fresh
        user action.
      </p>
      <p
        style={text}
        data-testid={incidentAdministrationTestId("lifecycle-current")}
      >
        {resource
          ? `${resource.incident_key} — ${resource.title}\nCurrent accepted state: ${resource.status === "closed" ? "Closed, read-only" : "Active"} · Version ${resource.incident_version}`
          : "Loading incident identity…"}
      </p>
      <p
        role="status"
        aria-live={state.active ? "polite" : "off"}
        style={muted}
      >
        {state.access === "unavailable"
          ? "Current access could not be checked. Local work is retained."
          : state.read === "failed"
            ? "Current incident state could not be refreshed. Displayed values may be stale."
            : state.read === "loading"
              ? "Loading lifecycle controls…"
              : state.read === "refreshing"
                ? "Refreshing current incident…"
                : ""}
      </p>
      <label htmlFor={`${prefix}-reason`} style={stack}>
        Reason
        <textarea
          id={`${prefix}-reason`}
          data-testid={incidentAdministrationTestId("lifecycle-reason")}
          rows={4}
          value={draft.reason}
          readOnly={!admin}
          aria-describedby={`${prefix}-reason-help${state.fieldError ? ` ${prefix}-reason-error` : ""}`}
          aria-invalid={state.fieldError !== null}
          style={{
            boxSizing: "border-box",
            width: "100%",
            minWidth: 0,
            resize: "vertical",
            padding: "var(--ct-component-text-input-padding)",
            border: "var(--ct-component-text-input-border)",
            borderRadius: "var(--ct-component-text-input-rounded)",
            background: "var(--ct-component-text-input-backgroundColor)",
            color: "var(--ct-component-text-input-textColor)",
          }}
          onChange={(event) => controller.changeReason(event.target.value)}
        />
      </label>
      <p id={`${prefix}-reason-help`} style={muted}>
        A reason is required. Up to 4096 characters after text normalization;
        Enter adds a new line.
      </p>
      {state.fieldError ? (
        <p
          id={`${prefix}-reason-error`}
          role="alert"
          aria-live={state.active ? "assertive" : "off"}
          style={text}
        >
          {state.fieldError.message}
        </p>
      ) : null}
      <div style={actions}>
        <button
          type="button"
          data-testid={incidentAdministrationTestId("close-button")}
          style={secondaryButtonStyle}
          aria-disabled={!controller.canPropose("close")}
          aria-describedby={`${prefix}-disabled`}
          onClick={() => controller.propose("close")}
        >
          Close incident
        </button>
        <button
          type="button"
          data-testid={incidentAdministrationTestId("reopen-button")}
          style={secondaryButtonStyle}
          aria-disabled={!controller.canPropose("reopen")}
          aria-describedby={`${prefix}-disabled`}
          onClick={() => controller.propose("reopen")}
        >
          Reopen incident
        </button>
        <button
          type="button"
          style={secondaryButtonStyle}
          aria-disabled={draft.reason === ""}
          onClick={() => {
            if (draft.reason !== "") controller.discardReason();
          }}
        >
          Discard reason
        </button>
      </div>
      <p id={`${prefix}-disabled`} style={muted}>
        {blocked ||
          (resource?.status === "active"
            ? "Close is available for this active incident; Reopen requires a closed incident."
            : "Reopen is available for this closed incident; Close requires an active incident.")}
      </p>
      {reviewResource ? (
        <section
          data-testid={incidentAdministrationTestId("lifecycle-review")}
          style={card}
          aria-label="Lifecycle action review"
          aria-busy={pending}
        >
          <h4 style={text}>
            {reviewedAction === "close"
              ? "Review Close incident"
              : "Review Reopen incident"}
          </h4>
          <p style={text}>
            {reviewResource.incident_key} — {reviewResource.title}
            {"\n"}
            {reviewResource.incident_id}
            {"\n"}Reviewed {reviewResource.status}, version{" "}
            {reviewResource.incident_version}.
          </p>
          <p style={muted}>
            {reviewedAction === "close"
              ? "Source-state writes will become unavailable. The incident remains visible to its current members."
              : "New authorized source-state writes will become available. Rejected drafts will not resume automatically."}
          </p>
          {pending ? (
            <p style={text}>
              Captured reason: {operation.attempt.payload.reason}
            </p>
          ) : null}
          <div style={actions}>
            <button
              type="button"
              data-testid={incidentAdministrationTestId("lifecycle-confirm")}
              style={primaryButtonStyle}
              aria-disabled={!controller.canConfirm()}
              aria-describedby={`${prefix}-disabled`}
              onClick={controller.confirm}
            >
              {reviewedAction === "close"
                ? "Confirm Close incident"
                : "Confirm Reopen incident"}
            </button>
            {!pending ? (
              <button
                type="button"
                style={secondaryButtonStyle}
                onClick={controller.cancelReview}
              >
                Cancel review
              </button>
            ) : null}
          </div>
        </section>
      ) : draft.action && !pending && operation.kind !== "uncertain" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          aria-disabled={!controller.canPropose(draft.action)}
          onClick={controller.review}
        >
          Review current incident
        </button>
      ) : null}
      <p
        data-testid={incidentAdministrationTestId("lifecycle-outcome")}
        role={
          operation.kind === "rejected" && !state.fieldError
            ? "alert"
            : "status"
        }
        aria-live={
          state.active
            ? operation.kind === "rejected" && !state.fieldError
              ? "assertive"
              : "polite"
            : "off"
        }
        style={text}
      >
        {status}
      </p>
      {operation.kind === "confirmed" &&
      (state.read === "failed" || state.access === "unavailable") ? (
        <p style={text}>
          The action is confirmed, but current state could not be refreshed.
          Refresh current incident; do not resubmit the action.
        </p>
      ) : null}
      {operation.kind === "uncertain" ? (
        <>
          {problem ? (
            <p style={muted}>
              {lifecycleProblemMessage(problem)} The earlier action remains
              unconfirmed.
            </p>
          ) : null}
          <p style={muted}>
            Replay sends the original action and captured reason, even if the
            current incident has changed. It does not send your newer reason
            input.
          </p>
          <button
            type="button"
            data-testid={incidentAdministrationTestId("lifecycle-replay")}
            style={primaryButtonStyle}
            aria-disabled={!controller.canReplay()}
            aria-describedby={`${prefix}-disabled`}
            onClick={controller.replay}
          >
            Replay original action
          </button>
        </>
      ) : null}
      {operation.kind === "uncertain" ||
      problem?.code === "client_txn_conflict" ? (
        <div style={stack}>
          <p style={muted}>
            Forgetting recovery removes only this local record. It cannot cancel
            or undo a server operation.
          </p>
          <button
            type="button"
            style={secondaryButtonStyle}
            aria-disabled={!controller.canForget()}
            onClick={controller.forgetRecovery}
          >
            Forget recovery
          </button>
        </div>
      ) : null}
      <p
        style={muted}
        data-testid={incidentAdministrationTestId("lifecycle-notice")}
      >
        {state.notice}
      </p>
      <button
        ref={refresh}
        type="button"
        data-testid={incidentAdministrationTestId("lifecycle-refresh")}
        style={secondaryButtonStyle}
        aria-disabled={pending}
        onClick={controller.refresh}
      >
        Refresh current incident
      </button>
    </section>
  );
}

export function IncidentLifecycleDepartureDialog({
  controller,
  fallbackFocusRef,
}: {
  controller: IncidentLifecycleController;
  fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (!state.departure) return null;
  return (
    <AccountDialog
      labelledBy="lifecycle-departure-title"
      describedBy="lifecycle-departure-description"
      fallbackFocusRef={fallbackFocusRef}
      onClose={() => controller.resolveDeparture("stay")}
      style={card}
    >
      {(dismiss) => (
        <>
          <h2 id="lifecycle-departure-title" style={text}>
            Leave lifecycle work?
          </h2>
          <p id="lifecycle-departure-description" style={text}>
            Leaving discards the local reason and forgets recovery for this
            incident. It cannot cancel a server operation or establish whether
            an uncertain action succeeded.
          </p>
          <div style={actions}>
            <button
              type="button"
              style={secondaryButtonStyle}
              onClick={dismiss}
            >
              Stay
            </button>
            <button
              type="button"
              style={secondaryButtonStyle}
              onClick={() => controller.resolveDeparture("discard")}
            >
              Leave and forget recovery
            </button>
          </div>
        </>
      )}
    </AccountDialog>
  );
}
