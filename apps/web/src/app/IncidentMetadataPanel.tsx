import {
  incidentAdministrationTestId,
  incidentControlsActionMessageTestId,
  incidentControlsStatusTestId,
  incidentControlsSurfaceTestId,
} from "@cartulary/ui-contracts";
import {
  type CSSProperties,
  type RefObject,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type {
  WorkbookDensityMode,
  WorkbookIncidentControlsRendererProps,
} from "../shared/workbookShellContracts";
import { AccountDialog } from "./AccountDialog";
import type { IncidentMetadataController } from "./incidentMetadataController";
import {
  changedMetadataFields,
  type IncidentMetadataField,
  metadataDirty,
  metadataEditRole,
  metadataFields,
  metadataLabels,
  metadataTLPs,
} from "./incidentMetadataModel";
import { primaryButtonStyle, secondaryButtonStyle } from "./landingAdminStyles";

export function IncidentMetadataFeature({
  controller,
  bindSurface,
  ...surface
}: WorkbookIncidentControlsRendererProps & {
  controller: IncidentMetadataController;
  bindSurface: (surface: WorkbookIncidentControlsRendererProps | null) => void;
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
    <IncidentMetadataPanel controller={controller} density={surface.density} />
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
  alignItems: "center",
  gap: "var(--ct-spacing-sm)",
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
  background: "var(--ct-colors-surface-2)",
};
const input: CSSProperties = {
  padding: "var(--ct-component-text-input-padding)",
  border: "var(--ct-component-text-input-border)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
};
const inputIDs: Record<
  IncidentMetadataField,
  Parameters<typeof incidentAdministrationTestId>[0]
> = {
  description: "patch-description",
  severity: "patch-severity",
  tlp: "patch-tlp",
  current_phase: "patch-current-phase",
  primary_external_case_ref: "patch-external-case",
};
const css = `
[data-incident-metadata] :is(input,select,textarea,button):focus-visible { outline: var(--ct-border-focus); outline-offset: var(--ct-spacing-xs); }
[data-incident-metadata] :is(input,select,textarea,button) { font: inherit; max-inline-size: 100%; }
[data-incident-metadata] :is(input,select,textarea) { box-sizing: border-box; min-inline-size: 0; inline-size: 100%; }
[data-incident-metadata] button[aria-disabled="true"] { cursor: not-allowed; color: var(--ct-colors-ink-muted) !important; background: var(--ct-colors-surface-2) !important; border: var(--ct-border-hairline) !important; }
[data-incident-metadata] [data-metadata-field] { padding: var(--metadata-cell-padding); }
`;
export function IncidentMetadataPanel({
  controller,
  density = "default",
}: {
  controller: IncidentMetadataController;
  density?: WorkbookDensityMode | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const saveFocus = useRef<HTMLButtonElement>(null);
  const refreshFocus = useRef<HTMLButtonElement>(null);
  const lastFocusedControl = useRef<HTMLElement | null>(null);
  useLayoutEffect(() => {
    if (
      document.activeElement === document.body &&
      lastFocusedControl.current &&
      !lastFocusedControl.current.isConnected
    )
      refreshFocus.current?.focus();
  });
  const { draft, resource, operation } = state;
  const busy = state.read === "loading" || state.read === "refreshing";
  const [delayed, setDelayed] = useState(false);
  useEffect(() => {
    setDelayed(false);
    if (state.read !== "loading") return;
    const generation = state.readGeneration;
    const timer = setTimeout(() => {
      if (controller.getSnapshot().readGeneration === generation)
        setDelayed(true);
    }, 2_000);
    return () => clearTimeout(timer);
  }, [controller, state.read, state.readGeneration]);
  const visible = state.authority !== null && state.access === "ready";
  const editable =
    visible &&
    metadataEditRole(state.authority?.role) &&
    resource?.status === "active";
  const needsReview =
    draft?.reviewRequired ||
    operation.kind === "conflicted" ||
    operation.kind === "uncertain";
  const status =
    state.access === "unavailable"
      ? "Current incident access could not be checked. Your local work is retained."
      : state.read === "failed"
        ? resource && visible
          ? "Metadata refresh failed. Displayed values may be stale."
          : "Incident metadata unavailable."
        : state.read === "loading"
          ? delayed
            ? "Still loading this surface"
            : "Loading promoted fields…"
          : state.read === "refreshing"
            ? "Refreshing promoted fields…"
            : !visible
              ? "Checking current incident access…"
              : "Incident controls synced.";
  const operationText =
    operation.kind === "pending"
      ? operation.stage === "authorization"
        ? "Checking access before saving…"
        : "Saving promoted incident fields…"
      : operation.kind === "confirmed"
        ? "Saved promoted incident fields."
        : operation.kind === "conflicted"
          ? "The incident changed. Your exact draft is retained; review current values before saving again."
          : operation.kind === "uncertain"
            ? "The save result is uncertain. Observing current values cannot establish whether this attempt succeeded."
            : operation.kind === "rejected"
              ? operation.problem.code === "incident_closed"
                ? "The save was rejected because the incident is closed."
                : operation.problem.code === "authorization_denied"
                  ? "The save was rejected. Current incident access must be checked."
                  : "The submitted changes were rejected. Review the affected input before saving again."
              : state.reviewNotice ||
                (metadataDirty(draft) ? "Unsaved changes." : "");
  const assertion =
    operation.kind === "conflicted" ||
    operation.kind === "rejected" ||
    state.read === "failed";
  const readOnlyReason =
    resource?.status === "closed"
      ? "This incident is closed. Promoted fields are read-only; current members can still read them."
      : "Promoted fields are read-only. A current reviewer or admin role is required to edit this incident.";
  return (
    <section
      aria-label="Promoted incident fields"
      onFocusCapture={(event) => {
        if (event.target instanceof HTMLElement)
          lastFocusedControl.current = event.target;
      }}
      data-incident-metadata=""
      data-metadata-operation={operation.kind}
      data-testid={incidentControlsSurfaceTestId()}
      data-incident-controls-section="incident-fields"
      data-incident-controls-load-state={
        state.read === "failed" ? "unavailable" : busy ? "loading" : "synced"
      }
      style={{
        ...stack,
        fontSize: `var(--ct-density-${density}-fontSize)`,
        lineHeight: `var(--ct-density-${density}-lineHeight)`,
        ...{
          "--metadata-cell-padding": `var(--ct-density-${density}-cellPadding)`,
        },
      }}
    >
      <style>{css}</style>
      <header style={actions}>
        <h3 style={text}>Promoted fields</h3>
        <button
          type="button"
          style={secondaryButtonStyle}
          ref={refreshFocus}
          aria-disabled={
            busy || operation.kind === "pending" || !state.authority
          }
          onClick={() => {
            if (!busy && operation.kind !== "pending") controller.refresh();
          }}
        >
          Refresh
        </button>
      </header>
      <div
        role={assertion ? "alert" : "status"}
        aria-live={assertion ? "assertive" : "polite"}
        aria-atomic="true"
      >
        <p data-testid={incidentControlsStatusTestId()} style={muted}>
          {status}
        </p>
        <p data-testid={incidentControlsActionMessageTestId()} style={text}>
          {operationText}
        </p>
        {operation.kind !== "idle" && metadataDirty(draft) ? (
          <p style={muted}>Unsaved local changes remain.</p>
        ) : null}
      </div>
      {state.read === "failed" ? (
        <button
          type="button"
          style={secondaryButtonStyle}
          onClick={controller.refresh}
        >
          Check access and refresh
        </button>
      ) : null}
      {state.transportPending &&
      operation.kind !== "pending" &&
      operation.kind !== "uncertain" ? (
        <p style={muted}>
          A previous request is still settling. Saving becomes available when it
          finishes.
        </p>
      ) : null}
      {visible && resource && draft ? (
        <>
          <p style={muted}>
            Version {resource.incident_version}. Only changed promoted fields
            are submitted.
          </p>
          {!editable ? (
            <p
              data-testid={incidentAdministrationTestId("patch-readonly-note")}
              style={muted}
            >
              {readOnlyReason}
            </p>
          ) : null}
          {editable ? (
            <form
              style={stack}
              onSubmit={(event) => {
                event.preventDefault();
                controller.save();
              }}
            >
              {metadataFields.map((field) => {
                const error = state.fieldErrors[field];
                const errorId = `metadata-${field}-error`;
                const hintId = `metadata-${field}-hint`;
                const attributes = {
                  id: `metadata-${field}`,
                  "data-testid": incidentAdministrationTestId(inputIDs[field]),
                  "aria-invalid": error ? true : undefined,
                  "aria-describedby": `${hintId}${error ? ` ${errorId}` : ""}`,
                  style: input,
                  value: draft.values[field],
                };
                return (
                  <div
                    key={field}
                    data-metadata-field=""
                    style={{ ...stack, gap: "var(--ct-spacing-xs)" }}
                  >
                    <label htmlFor={attributes.id}>
                      {metadataLabels[field]}
                    </label>
                    {field === "description" ? (
                      <textarea
                        {...attributes}
                        rows={4}
                        onChange={(event) =>
                          controller.change(field, event.target.value)
                        }
                      />
                    ) : field === "tlp" ? (
                      <select
                        {...attributes}
                        onChange={(event) =>
                          controller.change(field, event.target.value)
                        }
                      >
                        <option value="">Unset</option>
                        {metadataTLPs.map((token) => (
                          <option key={token} value={token}>
                            {token}
                          </option>
                        ))}
                      </select>
                    ) : (
                      <input
                        {...attributes}
                        onChange={(event) =>
                          controller.change(field, event.target.value)
                        }
                      />
                    )}
                    <p id={hintId} style={muted}>
                      {field === "description"
                        ? "Multiline text, up to 16,384 characters after normalization. Empty clears the description."
                        : field === "tlp"
                          ? "Select a canonical TLP value. Unset clears TLP."
                          : "Text, up to 128 characters after normalization. Empty clears this field."}
                    </p>
                    {error ? (
                      <p
                        id={errorId}
                        style={{
                          ...text,
                          color: "var(--ct-colors-semantic-conflict)",
                        }}
                      >
                        {error.message}
                      </p>
                    ) : null}
                  </div>
                );
              })}
              <div style={actions}>
                <button
                  type="submit"
                  ref={saveFocus}
                  data-testid={incidentAdministrationTestId("patch-button")}
                  style={primaryButtonStyle}
                  aria-disabled={!controller.canSave()}
                >
                  Save promoted fields
                </button>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  aria-disabled={!metadataDirty(draft)}
                  onClick={() => {
                    if (metadataDirty(draft)) controller.discard();
                  }}
                >
                  Discard changes
                </button>
              </div>
            </form>
          ) : (
            <dl style={stack}>
              {metadataFields.map((field) => (
                <div key={field} data-metadata-field="">
                  <dt>{metadataLabels[field]}</dt>
                  <dd style={text}>{resource[field] ?? "Unset"}</dd>
                </div>
              ))}
            </dl>
          )}
          {!editable && metadataDirty(draft) ? (
            <div style={stack}>
              <p style={muted}>
                Your local changes are retained. Editing eligibility and current
                values must be reviewed before saving.
              </p>
              <button
                type="button"
                style={secondaryButtonStyle}
                onClick={controller.discard}
              >
                Discard changes
              </button>
            </div>
          ) : null}
          {needsReview ? (
            <section aria-label="Review promoted field changes" style={card}>
              <h4 style={text}>Review current values and intended changes</h4>
              <p style={muted}>
                {operation.kind === "uncertain"
                  ? "This comparison is an observation, not a save receipt. A new attempt requires explicit review and a separate Save."
                  : "Use this version only after reviewing the current incident and your intended changes."}
              </p>
              {operation.kind === "uncertain" ? (
                <details>
                  <summary>Unconfirmed submitted attempt</summary>
                  <dl style={stack}>
                    {operation.attempt.fields.map((field) => (
                      <div key={field}>
                        <dt>{metadataLabels[field]}</dt>
                        <dd style={text}>
                          {operation.attempt.values[field] ||
                            "Clear this field"}
                        </dd>
                      </div>
                    ))}
                  </dl>
                </details>
              ) : null}
              {changedMetadataFields(draft).map((field) => (
                <div key={field} style={stack}>
                  <strong>{metadataLabels[field]}</strong>
                  <div style={stack}>
                    <div>
                      <span>Current value</span>
                      <p style={text}>{resource[field] ?? "Unset"}</p>
                    </div>
                    <div>
                      <span>Intended change</span>
                      <p style={text}>
                        {draft.values[field] === ""
                          ? "Clear this field"
                          : draft.values[field]}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
              {operation.kind === "uncertain" && !metadataDirty(draft) ? (
                <p style={muted}>
                  No local changes remain. The previous attempt is still
                  unconfirmed.
                </p>
              ) : null}
              <div style={actions}>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  onClick={controller.refresh}
                  aria-disabled={busy || operation.kind === "pending"}
                >
                  Observe current values
                </button>
                <button
                  type="button"
                  style={secondaryButtonStyle}
                  onClick={() => {
                    if (controller.canReview()) {
                      controller.review();
                      saveFocus.current?.focus();
                    }
                  }}
                  aria-disabled={!controller.canReview()}
                >
                  Use this version
                </button>
              </div>
            </section>
          ) : null}
          {operation.kind === "confirmed" && state.read === "failed" ? (
            <p style={muted}>
              The save is confirmed. Retry only the follow-up read; the saved
              request will not be resent.
            </p>
          ) : null}
        </>
      ) : null}
    </section>
  );
}
export function IncidentMetadataDepartureDialog({
  controller,
  fallbackFocusRef,
}: {
  controller: IncidentMetadataController;
  fallbackFocusRef?: RefObject<HTMLElement | null> | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  if (!state.departure) return null;
  return (
    <AccountDialog
      labelledBy="metadata-departure-title"
      describedBy="metadata-departure-description"
      fallbackFocusRef={fallbackFocusRef}
      onClose={() => controller.resolveDeparture("stay")}
      style={card}
    >
      {(dismiss) => (
        <>
          <h2 id="metadata-departure-title" style={text}>
            Leave metadata work?
          </h2>
          <p id="metadata-departure-description" style={text}>
            Leaving discards local changes and forgets recovery for this
            incident. It cannot cancel a pending server operation or establish
            whether an uncertain save succeeded.
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
