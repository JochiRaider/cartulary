import { type CSSProperties, useSyncExternalStore } from "react";
import { terminalCommonJob } from "../services/commonJobContract";
import { useWorkflowAttachment } from "../shared/useWorkflowAttachment";
import type { WorkbookIncidentControlsRendererProps } from "../shared/workbookShellContracts";
import type { AttachWorkflow } from "../shared/workflowAttachment";
import type { ReportingController } from "./ReportingController";

const stack: CSSProperties = {
  display: "grid",
  gap: "var(--ct-spacing-md)",
  minWidth: 0,
};
const actions: CSSProperties = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
};
const button: CSSProperties = {
  justifySelf: "start",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  border: "var(--ct-component-button-secondary-border)",
  padding: "var(--ct-component-button-secondary-padding)",
  font: "inherit",
};
const fileLink: CSSProperties = {
  color: "var(--ct-colors-ink)",
  overflowWrap: "anywhere",
  textUnderlineOffset: "0.2em",
};
export function ReportingFeature({
  controller,
  attachSurface,
  ...surface
}: WorkbookIncidentControlsRendererProps & {
  controller: ReportingController;
  attachSurface: AttachWorkflow<WorkbookIncidentControlsRendererProps>;
}) {
  useWorkflowAttachment(attachSurface, surface, true);
  return <ReportingPanel controller={controller} />;
}
function ReportingPanel({ controller }: { controller: ReportingController }) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const job = state.job;
  const admitted = state.available && state.presented;
  const canStart =
    admitted &&
    state.writable &&
    !state.busy &&
    !state.attempt &&
    !state.cancellation &&
    (!job || terminalCommonJob(job));
  return (
    <section
      aria-label="Snapshots and reports"
      style={stack}
      data-reporting-workflow="true"
    >
      <p>
        Create an immutable incident snapshot, then render an internal draft
        report from that snapshot.
      </p>
      <p>
        Internal drafts may contain sensitive incident data. Creating a draft
        does not approve or publish it.
      </p>
      <div style={actions}>
        <button
          type="button"
          style={button}
          disabled={!canStart}
          onClick={() => void controller.createSnapshot()}
        >
          Create snapshot
        </button>
        <button
          type="button"
          style={button}
          disabled={!canStart || !state.snapshotId}
          onClick={() => void controller.createReport()}
        >
          Render internal report
        </button>
      </div>
      {!state.available && (
        <p role="status">
          Reporting is paused while current access is established.
        </p>
      )}
      {state.available && !state.writable && (
        <p>
          Creating snapshots and reports requires an editor, reviewer, or admin
          role.
        </p>
      )}
      {state.busy && !job && <p role="status">Submitting request…</p>}
      {state.problem && <p role="alert">{state.problem}</p>}
      {state.attempt && !state.busy && (
        <button
          type="button"
          style={button}
          disabled={!admitted || !state.writable}
          onClick={() => void controller.replay()}
        >
          Recover request
        </button>
      )}
      {job && (
        <div
          style={stack}
          data-job-id={job.job_id}
          data-job-purpose={state.kind}
          data-job-cancelable={String(job.cancelable)}
          data-job-status={job.status}
        >
          <p role="status" aria-live="polite" data-job-progress="true">
            {state.kind === "snapshot" ? "Snapshot" : "Report"}:{" "}
            {job.status.replaceAll("_", " ")}. {job.progress.completed}
            {job.progress.total === null
              ? " steps completed"
              : ` of ${job.progress.total} steps completed`}
            .
          </p>
          <div style={actions}>
            <button
              type="button"
              style={button}
              disabled={!admitted || state.busy || !!state.attempt}
              onClick={() => void controller.refresh()}
            >
              Refresh job status
            </button>
            <button
              type="button"
              style={button}
              disabled={
                !admitted ||
                !state.writable ||
                state.busy ||
                (!job.cancelable && !state.cancellation)
              }
              data-job-cancel="true"
              onClick={() => void controller.cancel()}
            >
              {state.cancellation ? "Recover cancellation" : "Cancel job"}
            </button>
          </div>
          {state.paused && (
            <p>
              Observation is paused. Refresh to check the job; server work may
              continue.
            </p>
          )}
        </div>
      )}
      {state.snapshotId && (
        <p>
          Snapshot ready: <code>{state.snapshotId}</code>
        </p>
      )}
      {state.releaseId && (
        <p>
          Internal draft ready: <code>{state.releaseId}</code>
        </p>
      )}
      {admitted && state.writable && (
        <section aria-label="Transferred report files" style={stack}>
          <h3>Transferred report files</h3>
          <p>
            These files preserve report inputs and outputs from the source
            incident. Their approval status is historical.
          </p>
          <button
            type="button"
            style={button}
            disabled={state.retainedBusy}
            onClick={() => void controller.loadRetainedArtifacts()}
          >
            {state.retainedArtifacts === null
              ? "Load transferred report files"
              : "Refresh transferred report files"}
          </button>
          {state.retainedBusy && <p role="status">Loading report files…</p>}
          {state.retainedProblem && <p role="alert">{state.retainedProblem}</p>}
          {state.retainedArtifacts?.length === 0 && (
            <p role="status">
              No transferred report files are retained for this incident.
            </p>
          )}
          {state.retainedArtifacts && state.retainedArtifacts.length > 0 && (
            <ul>
              {state.retainedArtifacts.map((file) => (
                <li key={file.artifact_id}>
                  <a
                    href={file.downloadHref}
                    download={file.filename}
                    style={fileLink}
                  >
                    {file.filename}
                  </a>
                  {" — "}
                  {file.byte_size.toLocaleString()} bytes
                  {file.release_id
                    ? ` · Release ${file.release_id}`
                    : file.snapshot_id
                      ? ` · Snapshot ${file.snapshot_id}`
                      : " · Report catalog"}
                </li>
              ))}
            </ul>
          )}
        </section>
      )}
    </section>
  );
}
