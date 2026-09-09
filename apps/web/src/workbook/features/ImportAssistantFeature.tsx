import { workbookImportAssistantTestId } from "@cartulary/ui-contracts";
import { useEffect, useSyncExternalStore } from "react";
import type { WorkbookImportController } from "../../imports/WorkbookImportController";
import {
  importApplyBlocker,
  importOutcomeViews,
  terminalImportSession,
} from "../../imports/workbookImportState";
import { terminalCommonJob } from "../../services/commonJobContract";
import { importFailureMessage } from "../../services/importClient";

import type { WorkbookDensityMode } from "../../shared/workbookShellContracts";
import {
  ImportOperationNotice,
  importActionsStyle,
  importSectionStyle,
} from "./ImportOperationNotice";
import { ImportUnitCard } from "./ImportUnitCard";

export function ImportAssistantFeature({
  controller,
  onNavigateToView,
  density = "default",
}: {
  readonly controller: WorkbookImportController;
  readonly onNavigateToView: (viewSchemaId: string) => void;
  readonly density?: WorkbookDensityMode | undefined;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  useEffect(() => {
    controller.setPresented(true);
    return () => controller.setPresented(false);
  }, [controller]);
  const pending =
    state.operation?.phase === "pending" ||
    state.operation?.phase === "uncertain" ||
    state.actionPending;
  const terminal =
    state.session !== null && terminalImportSession(state.session);
  const applyBlocker = importApplyBlocker(state);
  const selected = state.session?.selected_unit_ids ?? [];
  const mutable =
    state.canWrite &&
    !pending &&
    !terminal &&
    state.session?.session_status !== "applying" &&
    !(state.job && !terminalCommonJob(state.job));
  return (
    <div
      data-testid={workbookImportAssistantTestId()}
      className="workbook-import-assistant"
      data-density={density}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-md)",
        minWidth: 0,
        overflowWrap: "anywhere",
        fontSize: `var(--ct-density-${density}-fontSize)`,
        lineHeight: `var(--ct-density-${density}-lineHeight)`,
      }}
    >
      <style>{importControlCss}</style>
      <p
        role="status"
        aria-live="polite"
        aria-atomic="true"
        style={{ margin: 0 }}
      >
        {state.message}
      </p>
      {state.access === "active" ? (
        <>
          <section aria-label="Import source" style={importSectionStyle}>
            <h3 style={{ margin: 0 }}>Source</h3>
            <p style={{ margin: 0 }}>
              Review a CSV or XLSX workbook. Closing this drawer keeps the
              current import in this tab; observation may continue.
            </p>
            {!state.session && !state.job ? (
              <div style={importActionsStyle}>
                <label style={{ minWidth: 0 }}>
                  Source workbook
                  <input
                    style={{ maxWidth: "100%" }}
                    type="file"
                    accept=".csv,.xlsx,text/csv,application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                    disabled={!state.canWrite || pending}
                    onChange={(event) =>
                      controller.chooseFile(
                        event.currentTarget.files?.[0] ?? null,
                      )
                    }
                  />
                </label>
                <button
                  type="button"
                  disabled={!state.canWrite || !state.file || pending}
                  onClick={() => void controller.upload()}
                >
                  Upload and discover
                </button>
              </div>
            ) : null}
            {state.file || state.session ? (
              <p style={{ margin: 0 }}>
                Workbook: {state.session?.original_filename ?? state.file?.name}
              </p>
            ) : null}
            {state.operation?.attempt.kind === "upload" ? (
              <ImportOperationNotice
                operation={state.operation}
                controller={controller}
                canRetry={state.canWrite}
              />
            ) : null}
            {state.job ? (
              <section aria-label="Import job" style={importSectionStyle}>
                <h4 style={{ margin: 0 }}>
                  {state.jobPurpose === "apply" ? "Apply job" : "Discovery job"}
                  : {state.job.status.replaceAll("_", " ")}
                </h4>
                <p style={{ margin: 0 }}>
                  Progress: {state.job.progress.completed}/
                  {state.job.progress.total ?? "unknown"}.{" "}
                  {state.observing
                    ? "Observing server work."
                    : terminalCommonJob(state.job)
                      ? "Job is terminal."
                      : "Observation is paused; server work may continue."}
                </p>
                {state.job.status === "failed" ? (
                  <p>
                    Server work failed. Review the session and unit outcomes for
                    any committed work. {state.job.error_summary?.code}
                  </p>
                ) : null}
                {state.job.status === "cancel_requested" ||
                state.job.status === "canceled" ? (
                  <p>
                    Cancellation does not roll back committed units. Review
                    their outcomes below.
                  </p>
                ) : null}
                {state.observationFailure ? (
                  <p role="alert">
                    Observation:{" "}
                    {importFailureMessage(state.observationFailure)}
                  </p>
                ) : null}
                <div style={importActionsStyle}>
                  <button
                    type="button"
                    disabled={state.cancellation?.phase === "pending"}
                    onClick={() =>
                      state.observing
                        ? controller.stopObservation()
                        : void controller.resumeObservation()
                    }
                  >
                    {state.observing
                      ? "Pause observation"
                      : "Refresh / Resume job"}
                  </button>
                  {!terminalCommonJob(state.job) ? (
                    <button
                      type="button"
                      disabled={!controller.canCancel()}
                      onClick={() => void controller.cancel()}
                    >
                      {state.cancellation?.phase === "pending"
                        ? "Requesting cancellation…"
                        : state.cancellation?.phase === "uncertain"
                          ? "Retry exact cancellation"
                          : "Cancel import"}
                    </button>
                  ) : null}
                </div>
                {state.cancellation?.failure ? (
                  <p role="alert">
                    Cancellation:{" "}
                    {importFailureMessage(state.cancellation.failure)}
                  </p>
                ) : null}
              </section>
            ) : null}
          </section>
          {state.loading || state.loadFailure || state.session ? (
            <section
              aria-label="Import session review"
              style={importSectionStyle}
            >
              <h3 style={{ margin: 0 }}>
                {terminal ? "Outcomes" : "Units and mapping"}
              </h3>
              {state.loading ? (
                <p>Loading session and unit resources…</p>
              ) : null}
              {state.loadFailure ? (
                <p role="alert">
                  Session review: {importFailureMessage(state.loadFailure)}
                </p>
              ) : null}
              <button
                type="button"
                disabled={state.loading || pending}
                onClick={() => void controller.refresh()}
              >
                Refresh session and outcomes
              </button>
              {state.session ? (
                <>
                  <p>
                    Session: {state.session.session_status.replaceAll("_", " ")}
                    . Persisted selection: {selected.length} unit
                    {selected.length === 1 ? "" : "s"}.
                  </p>
                  {state.session.session_status === "partially_applied" ? (
                    <p>
                      Some units were applied and others failed. Applied work
                      remains committed.
                    </p>
                  ) : null}
                  {[...new Set(state.session.nonblocking_warning_codes)].map(
                    (code) => (
                      <p key={code}>Warning: {code}</p>
                    ),
                  )}
                  {state.session.blocking_diagnostics.map((diagnostic) => (
                    <p key={JSON.stringify(diagnostic)}>
                      Diagnostic
                      {typeof diagnostic.import_unit_id === "string"
                        ? ` for unit ${diagnostic.import_unit_id}`
                        : ""}
                      :{" "}
                      {typeof diagnostic.code === "string"
                        ? diagnostic.code
                        : typeof diagnostic.reason_code === "string"
                          ? diagnostic.reason_code
                          : "Review the affected unit."}
                    </p>
                  ))}
                  {!state.loading && state.units.length === 0 ? (
                    <p>
                      No import units were discovered. Choose another workbook
                      in a new import.
                    </p>
                  ) : null}
                </>
              ) : null}
              {state.units.map((item, index) => (
                <ImportUnitCard
                  key={item.unit.import_unit_id}
                  item={item}
                  index={index}
                  selected={selected.includes(item.unit.import_unit_id)}
                  mutable={mutable}
                  state={state}
                  controller={controller}
                />
              ))}
            </section>
          ) : null}
          {state.session && !terminal ? (
            <section aria-label="Apply review" style={importSectionStyle}>
              <h3 style={{ margin: 0 }}>Apply</h3>
              <p id="import-apply-readiness">
                {applyBlocker ??
                  "Selected units are ready to apply. The server will recheck mapping, selection, overlap, and duplicate imports."}
              </p>
              <button
                type="button"
                aria-describedby="import-apply-readiness"
                disabled={applyBlocker !== null}
                onClick={() => void controller.apply()}
              >
                Apply {selected.length} selected unit
                {selected.length === 1 ? "" : "s"}
              </button>
              {state.operation?.attempt.kind === "apply" ? (
                <ImportOperationNotice
                  operation={state.operation}
                  controller={controller}
                  canRetry={state.canWrite}
                />
              ) : null}
            </section>
          ) : null}
          {importOutcomeViews(state).length ? (
            <nav aria-label="Imported result views" style={importActionsStyle}>
              {importOutcomeViews(state).map((view) => (
                <button
                  key={view.id}
                  type="button"
                  disabled={state.actionPending}
                  onClick={() =>
                    void controller.openResult(view.id, onNavigateToView)
                  }
                >
                  Open {view.title}
                </button>
              ))}
            </nav>
          ) : null}
          {(state.session || state.job) &&
          !pending &&
          (!state.job || terminalCommonJob(state.job)) ? (
            <button
              type="button"
              disabled={!state.canWrite}
              onClick={() => controller.startNew()}
            >
              Start a new import
            </button>
          ) : null}
        </>
      ) : null}
    </div>
  );
}

const importControlCss = `
.workbook-import-assistant :is(button, input, select, textarea) {
  box-sizing: border-box;
  min-inline-size: 0;
  max-inline-size: 100%;
  font: inherit;
  color: var(--ct-colors-ink);
  background: var(--ct-colors-surface-2);
  border: var(--ct-border-hairline);
  border-radius: var(--ct-rounded-sm);
  padding: var(--ct-spacing-xs) var(--ct-spacing-sm);
}
.workbook-import-assistant button { justify-self: start; inline-size: fit-content; cursor: pointer; }
.workbook-import-assistant :is(input, select, textarea) { background: var(--ct-colors-surface-1); }
.workbook-import-assistant :is(button, input, select, textarea):disabled { opacity: 0.65; cursor: default; }
.workbook-import-assistant :is(button, input, select, textarea, summary, section[tabindex]):focus-visible {
  outline: var(--ct-component-focus-ring-border);
  outline-offset: var(--ct-component-focus-ring-offset);
}
.workbook-import-assistant summary { cursor: pointer; }
`;
