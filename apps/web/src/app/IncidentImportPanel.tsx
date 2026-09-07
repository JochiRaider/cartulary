import {
  incidentImportJobTestId,
  incidentImportTestId,
} from "@cartulary/ui-contracts";
import { useId } from "react";
import {
  importedIncidentTarget,
  terminalImportJob,
} from "./api/incidentImportClient";
import {
  admissionUnresolved,
  cancelableImport,
  importStatusLabel,
  openableImport,
} from "./incidentImportModel";
import {
  errorTextStyle,
  formGridStyle,
  inputStyle,
  jobPanelStyle,
  labelBlockStyle,
  metadataTextStyle,
  primaryButtonStyle,
  secondaryButtonStyle,
  sectionEyebrowStyle,
  sectionTitleStyle,
  statusTextStyle,
  strongTextStyle,
  subsectionTitleStyle,
  surfacePanelStyle,
  visuallyHiddenStyle,
} from "./landingAdminStyles";
import type { IncidentImportPresentation } from "./useIncidentImport";

const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
  alignItems: "center",
} as const;
const localStyles = `
[data-incident-import] :is(button,input):focus-visible { outline: var(--ct-border-focus); outline-offset: var(--ct-component-focus-ring-offset); }
[data-incident-import] button:is(:disabled,[aria-disabled="true"]) { cursor: not-allowed !important; color: var(--ct-colors-ink-subtle) !important; background: var(--ct-colors-surface-3) !important; }
[data-incident-import] button { white-space: normal; overflow-wrap: anywhere; min-width: 0; max-width: 100%; }
[data-incident-import] input { box-sizing: border-box; min-width: 0; max-width: 100%; }
[data-incident-import], [data-incident-import] :is(section,form) { min-width: 0; grid-template-columns: minmax(0, 1fr); overflow-wrap: anywhere; }
[data-incident-import] progress { appearance: none; box-sizing: border-box; height: var(--ct-spacing-sm); border: var(--ct-border-hairline); border-radius: var(--ct-rounded-pill); background: var(--ct-colors-surface-3); }
[data-incident-import] progress:indeterminate { background: repeating-linear-gradient(135deg, var(--ct-colors-accent) 0, var(--ct-colors-accent) var(--ct-spacing-xs), var(--ct-colors-surface-3) var(--ct-spacing-xs), var(--ct-colors-surface-3) var(--ct-spacing-sm)); }
[data-incident-import] progress::-webkit-progress-bar { background: inherit; border-radius: inherit; }
[data-incident-import] progress::-webkit-progress-value { background: var(--ct-colors-accent); border-radius: inherit; }
[data-incident-import] progress::-moz-progress-bar { background: var(--ct-colors-accent); border-radius: inherit; }
`;

export function IncidentImportPanel({
  binding,
}: {
  binding: IncidentImportPresentation;
}) {
  const {
    controller,
    state,
    fileInputRef,
    jobHeadingRef,
    retryAdmission,
    retryAdmissionRef,
    cancel,
  } = binding;
  const id = useId();
  const entry = state.selectedJobId
    ? state.jobs[state.selectedJobId]
    : undefined;
  const job = entry?.job;
  const unresolved = admissionUnresolved(state);
  const uploadLocked = unresolved || state.access !== "ready";
  const admission = state.admission;
  const opening = state.navigation.kind === "opening";
  const target = entry ? openableImport(entry) : null;
  const handoffFailed =
    state.navigation.kind === "unavailable" ||
    state.navigation.kind === "access_lost";
  return (
    <section data-incident-import="" style={surfacePanelStyle}>
      <style>{localStyles}</style>
      <header>
        <p style={sectionEyebrowStyle}>Incident portability</p>
        <h2 style={sectionTitleStyle}>Incident import</h2>
      </header>
      <form
        aria-label="Import incident bundle"
        data-testid={incidentImportTestId("form")}
        style={formGridStyle}
        onSubmit={(event) => {
          event.preventDefault();
          controller.submit();
        }}
      >
        <label htmlFor={`${id}-file`} style={labelBlockStyle}>
          Incident bundle file
        </label>
        <input
          id={`${id}-file`}
          ref={fileInputRef}
          data-testid={incidentImportTestId("file")}
          type="file"
          accept=".zip,.tar,.gz,.tgz,application/zip,application/x-tar,application/gzip,application/x-gzip,application/octet-stream"
          style={inputStyle}
          disabled={uploadLocked}
          aria-invalid={state.fieldError !== null}
          aria-describedby={`${id}-file-help${state.fieldError ? ` ${id}-file-error` : ""}`}
          onChange={(event) =>
            controller.selectFile(event.currentTarget.files?.[0] ?? null)
          }
        />
        <p id={`${id}-file-help`} style={metadataTextStyle}>
          {state.selectedFile ? `Selected: ${state.selectedFile.name}. ` : ""}
          Upload a whole-incident bundle. The server validates its contents.
        </p>
        {state.fieldError ? (
          <p id={`${id}-file-error`} style={errorTextStyle}>
            Select an incident bundle file.
          </p>
        ) : null}
        <div style={actionsStyle}>
          <button
            style={primaryButtonStyle}
            type="submit"
            disabled={uploadLocked}
            aria-busy={admission.kind === "pending"}
          >
            Start import
          </button>
          {state.selectedFile && !unresolved ? (
            <button
              style={secondaryButtonStyle}
              type="button"
              onClick={() => controller.selectFile(null)}
            >
              Clear selection
            </button>
          ) : null}
        </div>
      </form>
      {state.access === "checking" ? (
        <p style={statusTextStyle}>
          Import access is being checked. Upload is unavailable until the
          session is confirmed.
        </p>
      ) : null}
      <div
        data-testid={incidentImportTestId("admission")}
        style={formGridStyle}
      >
        {admission.kind === "pending" ? (
          <>
            <p style={statusTextStyle}>
              Uploading bundle and awaiting admission…
            </p>
            <progress aria-label="Upload admission" style={{ width: "100%" }} />
          </>
        ) : null}
        {admission.kind === "uncertain" ? (
          <>
            <p style={errorTextStyle}>
              Admission is unconfirmed. The server may have accepted this
              import.
            </p>
            <p style={metadataTextStyle}>
              Retry sends the same file and request. Keep this tab open to
              retain recovery.
            </p>
            <div style={actionsStyle}>
              <button
                type="button"
                style={primaryButtonStyle}
                onClick={retryAdmission}
                ref={retryAdmissionRef}
              >
                Retry admission
              </button>
            </div>
          </>
        ) : null}
        {admission.kind === "rejected" ? (
          <p style={errorTextStyle}>
            {admission.problem === "conflict"
              ? "This request conflicts with a previous admission. No new import was confirmed."
              : "Import admission was rejected. Review the selected bundle before submitting again."}
          </p>
        ) : null}
        {admission.kind === "accepted" ? (
          <p style={statusTextStyle}>Import accepted. Observe its job below.</p>
        ) : null}
      </div>
      <section aria-labelledby={`${id}-jobs`} style={formGridStyle}>
        <h3 id={`${id}-jobs`} style={subsectionTitleStyle}>
          Imports in this session
        </h3>
        <p style={metadataTextStyle}>
          Only imports started here are listed. This list and upload recovery
          stay in this tab until reload, sign-out or access loss. Server jobs
          may continue when you leave.
        </p>
        {state.order.length === 0 ? (
          <p style={statusTextStyle}>No imports are known in this session.</p>
        ) : (
          <>
            <div style={actionsStyle}>
              <button
                type="button"
                style={secondaryButtonStyle}
                onClick={state.paused ? controller.resume : controller.pause}
              >
                {state.paused ? "Resume updates" : "Pause updates"}
              </button>
              {state.paused ? <span>Automatic updates paused.</span> : null}
            </div>
            <ul
              aria-label="Imports in this session"
              data-testid={incidentImportTestId("jobs")}
              style={{
                ...formGridStyle,
                listStyle: "none",
                padding: 0,
                margin: 0,
              }}
            >
              {[...state.order].reverse().map((jobId) => {
                const item = state.jobs[jobId];
                if (!item) return null;
                return (
                  <li key={jobId}>
                    <button
                      type="button"
                      data-testid={incidentImportJobTestId(jobId)}
                      aria-current={
                        jobId === state.selectedJobId ? "true" : undefined
                      }
                      onClick={() => controller.selectJob(jobId)}
                      style={{
                        ...secondaryButtonStyle,
                        width: "100%",
                        justifyContent: "flex-start",
                        textAlign: "left",
                      }}
                    >
                      {item.filename} — {importStatusLabel[item.job.status]}
                      {item.observation.kind === "failed" ||
                      item.observation.kind === "stale"
                        ? " (last observed)"
                        : item.observation.kind === "unavailable"
                          ? " (job unavailable)"
                          : ""}
                    </button>
                  </li>
                );
              })}
            </ul>
          </>
        )}
      </section>
      {entry && job ? (
        <section
          aria-labelledby={`${id}-detail`}
          data-testid={incidentImportTestId("detail")}
          style={{ ...jobPanelStyle, ...formGridStyle }}
        >
          <h3
            id={`${id}-detail`}
            ref={jobHeadingRef}
            tabIndex={-1}
            style={subsectionTitleStyle}
          >
            {importStatusLabel[job.status]}
          </h3>
          <p style={strongTextStyle}>{entry.filename}</p>
          {!terminalImportJob(job) ? (
            <>
              <progress
                data-testid={incidentImportTestId("progress")}
                aria-label="Import processing"
                aria-valuetext={
                  job.progress.total === null
                    ? `${job.progress.completed} completed; total unknown`
                    : `${job.progress.completed} of ${job.progress.total} completed`
                }
                value={
                  job.progress.total === null
                    ? undefined
                    : job.progress.completed
                }
                max={job.progress.total ?? undefined}
                style={{
                  width: "100%",
                  accentColor: "var(--ct-colors-accent)",
                }}
              />
              <p style={metadataTextStyle}>
                {job.progress.total === null
                  ? `${job.progress.completed} completed; total unknown.`
                  : `${job.progress.completed} of ${job.progress.total} completed.`}
              </p>
            </>
          ) : null}
          {entry.observation.kind === "reading" ? (
            <p style={metadataTextStyle}>Refreshing job status…</p>
          ) : null}
          {entry.observation.kind === "stale" ? (
            <p style={metadataTextStyle}>
              Showing the last validated status. Refresh to check current
              actions.
            </p>
          ) : null}
          {entry.observation.kind === "failed" ? (
            <p style={errorTextStyle}>
              Observation unavailable. The last validated status may be stale.
            </p>
          ) : null}
          {entry.observation.kind === "unavailable" ? (
            <p style={errorTextStyle}>
              This job is no longer available. It may have expired or become
              inaccessible. This does not undo a committed import.
            </p>
          ) : null}
          {entry.cancellation.kind === "pending" ? (
            <p style={statusTextStyle}>Requesting cancellation…</p>
          ) : null}
          {entry.cancellation.kind === "uncertain" ? (
            <p style={errorTextStyle}>
              Cancellation is unconfirmed. Check the job status; retry
              cancellation only if it remains available.
            </p>
          ) : null}
          {entry.cancellation.kind === "rejected" ? (
            <p style={errorTextStyle}>
              Cancellation was rejected. The job’s authoritative status
              determines its outcome.
            </p>
          ) : null}
          {job.status === "cancel_requested" ? (
            <p style={statusTextStyle}>
              Cancellation has been requested. Waiting for the server’s final
              outcome.
            </p>
          ) : null}
          {job.status === "failed" ? (
            <p style={errorTextStyle}>
              The server could not import this bundle. Review the bundle before
              starting a new import.
            </p>
          ) : null}
          {job.status === "canceled" ? (
            <p style={statusTextStyle}>The server confirmed cancellation.</p>
          ) : null}
          {job.status === "succeeded" &&
          importedIncidentTarget(job) === null ? (
            <p style={errorTextStyle}>
              The job succeeded, but its imported incident reference is
              incomplete or unsupported. Refresh the result to check again.
            </p>
          ) : null}
          {handoffFailed ? (
            <p style={errorTextStyle}>
              The import succeeded, but the workbook could not be opened. Retry
              opening the imported incident.
            </p>
          ) : null}
          {opening ? (
            <p style={statusTextStyle}>Opening imported incident…</p>
          ) : null}
          <div style={actionsStyle}>
            {target ? (
              <button
                type="button"
                style={primaryButtonStyle}
                aria-disabled={opening}
                aria-busy={opening}
                onClick={controller.open}
              >
                Open imported incident
              </button>
            ) : null}
            <button
              type="button"
              style={secondaryButtonStyle}
              disabled={entry.observation.kind === "reading" || opening}
              onClick={() => controller.refresh()}
            >
              {entry.observation.kind === "failed"
                ? "Retry observation"
                : "Refresh job status"}
            </button>
            {(job.status === "queued" || job.status === "running") &&
            job.cancelable ? (
              <button
                type="button"
                style={secondaryButtonStyle}
                aria-disabled={!cancelableImport(entry)}
                aria-busy={entry.cancellation.kind === "pending"}
                onClick={cancel}
              >
                {entry.cancellation.kind === "uncertain"
                  ? "Retry cancellation"
                  : "Cancel import"}
              </button>
            ) : null}
          </div>
        </section>
      ) : null}
      <div
        data-testid={incidentImportTestId("feedback")}
        role={state.announcement.priority === "assertive" ? "alert" : "status"}
        aria-live={state.announcement.priority}
        aria-atomic="true"
        style={visuallyHiddenStyle}
      >
        <span key={state.announcement.sequence}>{state.announcement.text}</span>
      </div>
    </section>
  );
}
