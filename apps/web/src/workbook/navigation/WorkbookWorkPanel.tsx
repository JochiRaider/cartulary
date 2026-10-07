import { useLayoutEffect, useRef, useState, useSyncExternalStore } from "react";
import { useWorkbookSecondaryPanel } from "../../shared/WorkbookRecoveryBoundary";
import { useWorkbookCommand } from "../commands/WorkbookCommands";
import {
  workbookFormHeadingStyle,
  workbookQuietCommandStyle,
} from "../components/workbookFormStyles";
import { WorkbookAuxiliaryDock } from "../layout/WorkbookAuxiliaryDock";
import {
  decisionsViewSchemaId,
  handoffViewSchemaId,
  statusReviewViewSchemaId,
  taskRequestsViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { workbookPinIdentity } from "./WorkbookSessionNavigation";
import { useWorkbookWorkbench } from "./WorkbookWorkbenchContext";

const subscribeEmpty = () => () => {};
const emptySnapshot = () => null;
export function WorkbookWorkControls() {
  const workbench = useWorkbookWorkbench();
  const state = useSyncExternalStore(
    workbench?.session.subscribe ?? subscribeEmpty,
    workbench?.session.getSnapshot ?? emptySnapshot,
  );
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const close = () => setOpen(false);
  const coordinated = useWorkbookSecondaryPanel(open, close);
  const before = useRef(false);
  const focusedOpening = useRef(false);
  if (!open) focusedOpening.current = false;
  useLayoutEffect(() => {
    if (!state?.readable) setOpen(false);
  }, [state?.readable]);
  useLayoutEffect(() => {
    if (
      before.current &&
      !open &&
      !coordinated.current &&
      document.activeElement === document.body
    )
      trigger.current?.focus();
    before.current = open;
  }, [open, coordinated]);
  useWorkbookCommand({
    id: "work.open",
    family: "Follow up",
    label: "Open Work",
    terms: ["session", "working set", "pins", "tasks"],
    targetKind: "shell",
    availability: () =>
      state?.readable ? null : "Incident access is being checked.",
    invoke: () => {
      setOpen(true);
      return true;
    },
  });
  useWorkbookCommand({
    id: "view.return",
    family: "View",
    label: "Return to previous context",
    terms: ["back", "navigation"],
    targetKind: "shell",
    availability: () =>
      state?.trail.length
        ? null
        : "There is no previous context in this session.",
    invoke: () => {
      workbench?.returnToOrigin();
      return true;
    },
  });
  if (!workbench || !state) return null;
  return (
    <>
      <button
        style={workbookQuietCommandStyle}
        ref={trigger}
        type="button"
        aria-expanded={open}
        data-grid-editor-external-action="true"
        onClick={() => setOpen((value) => !value)}
      >
        Work
      </button>
      {open && state.readable ? (
        <WorkbookAuxiliaryDock
          label="Work"
          onClose={close}
          onNavigationClose={() => {
            coordinated.current = true;
            setOpen(false);
          }}
        >
          <WorkContents onClose={close} focusedOpening={focusedOpening} />
        </WorkbookAuxiliaryDock>
      ) : null}
    </>
  );
}
function WorkContents({
  onClose,
  focusedOpening,
}: {
  readonly onClose: () => void;
  readonly focusedOpening: { current: boolean };
}) {
  const workbench = useWorkbookWorkbench();
  const state = useSyncExternalStore(
    workbench?.session.subscribe ?? subscribeEmpty,
    workbench?.session.getSnapshot ?? emptySnapshot,
  );
  const heading = useRef<HTMLHeadingElement>(null);
  useLayoutEffect(() => {
    if (!focusedOpening.current) {
      heading.current?.focus();
      focusedOpening.current = true;
    }
  }, [focusedOpening]);
  if (!workbench || !state?.readable) return null;
  return (
    <div>
      <header
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
        }}
      >
        <h2 style={workbookFormHeadingStyle} ref={heading} tabIndex={-1}>
          Work
        </h2>
        <button
          style={workbookQuietCommandStyle}
          type="button"
          onClick={onClose}
        >
          Close Work
        </button>
      </header>
      <h3 style={workbookFormHeadingStyle}>Session working set</h3>
      <p>Session only · {state.pins.length}/20 pins</p>
      <p>
        Use Notes, Task Requests, Decisions, or saved views for durable
        investigative intent.
      </p>
      <button
        style={workbookQuietCommandStyle}
        type="button"
        onClick={workbench.pinCurrentView}
      >
        {workbench.pinViewLabel}
      </button>
      <p role="status">{state.message}</p>
      {state.pins.length ? (
        <ol>
          {state.pins.map((pin) => (
            <li key={workbookPinIdentity(pin)}>
              <button
                style={workbookQuietCommandStyle}
                type="button"
                disabled={state.pending}
                data-workbook-navigation="true"
                onClick={() => workbench.openPin(pin)}
              >
                {pin.label}
              </button>{" "}
              <button
                style={workbookQuietCommandStyle}
                type="button"
                aria-label={`Remove ${pin.label} from working set`}
                onClick={() => workbench.session.remove(pin)}
              >
                Remove
              </button>
            </li>
          ))}
        </ol>
      ) : (
        <p>No pins yet. Pin a view here or a record from its inspector.</p>
      )}
      <h3 style={workbookFormHeadingStyle}>Coordination</h3>
      <nav aria-label="Coordination views">
        {[
          ["Task Requests", taskRequestsViewSchemaId],
          ["Decisions", decisionsViewSchemaId],
          ["Handoffs", handoffViewSchemaId],
          ["Status Reviews", statusReviewViewSchemaId],
        ].map(([label, id]) => (
          <p key={id}>
            <button
              style={workbookQuietCommandStyle}
              type="button"
              data-workbook-navigation="true"
              onClick={() => {
                workbench.open({
                  sheetRef: { kind: "view_schema", id: id ?? "" },
                });
              }}
            >
              {label}
            </button>
          </p>
        ))}
      </nav>
    </div>
  );
}
export function WorkbookNavigationNotice() {
  const workbench = useWorkbookWorkbench();
  const state = useSyncExternalStore(
    workbench?.session.subscribe ?? subscribeEmpty,
    workbench?.session.getSnapshot ?? emptySnapshot,
  );
  if (!workbench || (!workbench.message && !state?.pending)) return null;
  return (
    <div
      role="status"
      style={{
        padding: "var(--ct-spacing-xs)",
        borderBottom: "var(--ct-border-hairline)",
      }}
    >
      {state?.pending ? "Opening destination…" : workbench.message}{" "}
      <button
        style={workbookQuietCommandStyle}
        type="button"
        onClick={workbench.cancelNavigation}
      >
        Cancel navigation
      </button>{" "}
      {workbench.retry ? (
        <button
          style={workbookQuietCommandStyle}
          type="button"
          data-workbook-navigation="true"
          onClick={workbench.retry}
        >
          Retry navigation
        </button>
      ) : null}{" "}
      {workbench.openBase ? (
        <button
          style={workbookQuietCommandStyle}
          type="button"
          data-workbook-navigation="true"
          onClick={workbench.openBase}
        >
          Open base view
        </button>
      ) : null}
    </div>
  );
}

export function WorkbookReturnControl() {
  const workbench = useWorkbookWorkbench();
  const state = useSyncExternalStore(
    workbench?.session.subscribe ?? subscribeEmpty,
    workbench?.session.getSnapshot ?? emptySnapshot,
  );
  if (!workbench || !state) return null;
  return (
    <div style={{ display: "flex", gap: "var(--ct-spacing-xs)" }}>
      <button
        style={workbookQuietCommandStyle}
        type="button"
        disabled={!state.readable || !state.trail.length || state.pending}
        data-workbook-navigation="true"
        onClick={() => workbench.returnToOrigin()}
      >
        Return
      </button>
      <details style={{ position: "relative" }}>
        <summary
          style={{ ...workbookQuietCommandStyle, listStyle: "none" }}
          aria-label="Return options"
        >
          ▾
        </summary>
        <button
          style={{
            ...workbookQuietCommandStyle,
            position: "absolute",
            insetBlockStart: "100%",
            insetInlineStart: 0,
            zIndex: 30,
            whiteSpace: "nowrap",
            background: "var(--ct-colors-surface-1)",
          }}
          type="button"
          disabled={
            !state.readable || !state.trail.at(-1)?.recordId || state.pending
          }
          data-workbook-navigation="true"
          onClick={(event) => {
            event.currentTarget.closest("details")?.removeAttribute("open");
            workbench.returnToOrigin(true);
          }}
        >
          Return and inspect
        </button>
      </details>
    </div>
  );
}
