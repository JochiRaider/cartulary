import { incidentLandingTestId } from "@cartulary/ui-contracts";
import { type RefObject, useId, useLayoutEffect, useRef } from "react";
import type {
  IncidentCreationBinding,
  IncidentCreationField,
} from "./incidentCreationModel";
import {
  detailsStyle,
  detailsSummaryStyle,
  errorTextStyle,
  formGridStyle,
  inputStyle,
  labelBlockStyle,
  primaryButtonStyle,
  secondaryButtonStyle,
  subsectionTitleStyle,
  textAreaStyle,
} from "./landingAdminStyles";

const fields = [
  "incident_key",
  "title",
  "description",
  "severity",
  "tlp",
  "current_phase",
  "primary_external_case_ref",
] as const;
const labels: Record<IncidentCreationField, string> = {
  incident_key: "Incident key",
  title: "Title",
  description: "Description",
  severity: "Severity",
  tlp: "TLP",
  current_phase: "Current phase",
  primary_external_case_ref: "External case",
};
const selectors = {
  incident_key: "incident-key",
  title: "incident-title",
  description: "create-description",
  severity: "create-severity",
  tlp: "create-tlp",
  current_phase: "create-current-phase",
  primary_external_case_ref: "create-external-case",
} as const;

export function IncidentCreationForm({
  creation: { controller, state },
  triggerRef,
  headingRef,
}: {
  creation: IncidentCreationBinding;
  triggerRef: RefObject<HTMLButtonElement | null>;
  headingRef: RefObject<HTMLHeadingElement | null>;
}) {
  const id = useId();
  const controls = useRef(
    new Map<
      IncidentCreationField,
      HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement
    >(),
  );
  const actionRef = useRef<HTMLButtonElement>(null);
  const wasOpen = useRef(false);
  const restoreFocus = useRef(false);
  const focusNew = useRef(false);
  const operation = state.operation;
  const busy =
    operation.kind === "submitting" ||
    (operation.kind === "created" && operation.handoff === "opening");
  const editable =
    operation.kind === "editing" ||
    (operation.kind === "rejected" && !operation.transactionConflict);

  useLayoutEffect(() => {
    if (state.open && (!wasOpen.current || focusNew.current)) {
      focusNew.current = false;
      const invalid = fields.find((field) => state.errors[field] !== undefined);
      const target = invalid
        ? controls.current.get(invalid)
        : editable || operation.kind === "submitting"
          ? controls.current.get("incident_key")
          : actionRef.current;
      target?.focus();
    } else if (!state.open && wasOpen.current && restoreFocus.current) {
      restoreFocus.current = false;
      (triggerRef.current?.isConnected
        ? triggerRef.current
        : headingRef.current
      )?.focus({ preventScroll: true });
    }
    wasOpen.current = state.open;
  });

  function close() {
    restoreFocus.current = true;
    controller.close();
  }
  const feedback = (
    <p
      id={`${id}-status`}
      role={
        state.message === ""
          ? undefined
          : state.announcement === "assertive"
            ? "alert"
            : "status"
      }
      data-testid={incidentLandingTestId("create-status")}
      style={{ margin: 0, overflowWrap: "anywhere" }}
    >
      {state.message}
    </p>
  );
  if (!state.open)
    return state.message === "" ? null : (
      <div style={formGridStyle}>{feedback}</div>
    );

  function field(field: IncidentCreationField) {
    const error = state.errors[field];
    const common = {
      id: `${id}-${field}`,
      name: field,
      "data-testid": incidentLandingTestId(selectors[field]),
      "aria-invalid": error ? (true as const) : undefined,
      "aria-describedby": error ? `${id}-${field}-error` : undefined,
      value: state.draft[field],
    };
    return (
      <div
        key={field}
        style={{
          ...labelBlockStyle,
          gridColumn: field === "description" ? "1 / -1" : undefined,
        }}
      >
        <label htmlFor={common.id}>{labels[field]}</label>
        {field === "description" ? (
          <textarea
            {...common}
            ref={(element) => {
              if (element) controls.current.set(field, element);
              else controls.current.delete(field);
            }}
            readOnly={!editable}
            style={textAreaStyle}
            onChange={(event) => controller.change(field, event.target.value)}
          />
        ) : field === "tlp" ? (
          <select
            {...common}
            ref={(element) => {
              if (element) controls.current.set(field, element);
              else controls.current.delete(field);
            }}
            disabled={!editable}
            style={inputStyle}
            onChange={(event) => {
              const value = event.target.value;
              if (
                value === "" ||
                value === "TLP:CLEAR" ||
                value === "TLP:GREEN" ||
                value === "TLP:AMBER" ||
                value === "TLP:AMBER+STRICT" ||
                value === "TLP:RED"
              )
                controller.change(field, value);
            }}
          >
            <option value="">Unset</option>
            <option value="TLP:CLEAR">Clear</option>
            <option value="TLP:GREEN">Green</option>
            <option value="TLP:AMBER">Amber</option>
            <option value="TLP:AMBER+STRICT">Amber strict</option>
            <option value="TLP:RED">Red</option>
          </select>
        ) : (
          <input
            {...common}
            ref={(element) => {
              if (element) controls.current.set(field, element);
              else controls.current.delete(field);
            }}
            required={field === "incident_key" || field === "title"}
            readOnly={!editable}
            style={inputStyle}
            onChange={(event) => controller.change(field, event.target.value)}
          />
        )}
        {error ? (
          <span
            id={`${id}-${field}-error`}
            style={{ ...errorTextStyle, margin: 0, overflowWrap: "anywhere" }}
          >
            {error}
          </span>
        ) : null}
      </div>
    );
  }

  return (
    <form
      aria-label="New incident"
      aria-busy={busy}
      noValidate
      data-testid={incidentLandingTestId("create-form")}
      style={{
        ...formGridStyle,
        minWidth: 0,
        padding: "var(--ct-spacing-md)",
        border: "var(--ct-border-hairline)",
        background: "var(--ct-colors-surface-1)",
      }}
      onSubmit={(event) => {
        event.preventDefault();
        controller.submit();
      }}
      onKeyDown={(event) => {
        if (
          event.key === "Escape" &&
          !event.defaultPrevented &&
          !event.nativeEvent.isComposing
        ) {
          event.preventDefault();
          event.stopPropagation();
          close();
        }
        if (
          event.key === "Enter" &&
          !(event.target instanceof HTMLTextAreaElement) &&
          (event.repeat ||
            event.nativeEvent.isComposing ||
            event.keyCode === 229)
        )
          event.preventDefault();
      }}
    >
      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          alignItems: "center",
          justifyContent: "space-between",
          gap: "var(--ct-spacing-sm)",
        }}
      >
        <h3 style={subsectionTitleStyle}>New incident</h3>
        <button
          type="button"
          aria-label="Close new incident"
          style={secondaryButtonStyle}
          onClick={close}
        >
          Close
        </button>
      </div>
      {operation.kind !== "created" ? (
        <>
          <div
            style={{
              ...formGridStyle,
              gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
            }}
          >
            {field("incident_key")}
            {field("title")}
          </div>
          <details
            open={state.detailsOpen}
            style={detailsStyle}
            onToggle={(event) => {
              if (event.currentTarget.open !== state.detailsOpen)
                controller.setDetailsOpen(event.currentTarget.open);
            }}
          >
            <summary style={detailsSummaryStyle}>More details</summary>
            <div
              style={{
                ...formGridStyle,
                marginTop: "var(--ct-spacing-sm)",
                gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
              }}
            >
              {fields.slice(2).map(field)}
            </div>
          </details>
        </>
      ) : null}
      {feedback}
      {operation.kind === "uncertain" || operation.kind === "submitting" ? (
        <p
          style={{
            margin: 0,
            overflowWrap: "anywhere",
            color: "var(--ct-colors-ink-muted)",
          }}
        >
          Closing this form does not cancel creation. Keep this tab open to
          retain recovery.
        </p>
      ) : null}
      <div
        style={{
          display: "flex",
          flexWrap: "wrap",
          gap: "var(--ct-spacing-sm)",
        }}
      >
        {operation.kind === "created" ? (
          <>
            <button
              ref={actionRef}
              type="button"
              style={primaryButtonStyle}
              disabled={busy}
              onClick={() => {
                void controller.openCreated();
              }}
            >
              Open created incident
            </button>
            <button
              type="button"
              style={secondaryButtonStyle}
              disabled={busy}
              onClick={() => {
                focusNew.current = true;
                controller.startNewAttempt();
              }}
            >
              New incident
            </button>
          </>
        ) : operation.kind === "uncertain" ? (
          <button
            ref={actionRef}
            type="button"
            style={primaryButtonStyle}
            onClick={controller.retry}
          >
            Retry creation
          </button>
        ) : operation.kind === "rejected" && operation.transactionConflict ? (
          <button
            ref={actionRef}
            type="button"
            style={primaryButtonStyle}
            onClick={() => {
              focusNew.current = true;
              controller.startNewAttempt();
            }}
          >
            Start new attempt
          </button>
        ) : (
          <button
            ref={actionRef}
            data-testid={incidentLandingTestId("create-submit-button")}
            style={primaryButtonStyle}
            disabled={busy}
            aria-busy={busy}
            type="submit"
          >
            Create and open
            {busy ? <span> — pending</span> : null}
          </button>
        )}
      </div>
    </form>
  );
}
