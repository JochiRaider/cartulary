import { X } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";

import {
  type APIError,
  clientTxnID,
  extractError,
  fetchJSON,
  publicErrorView,
} from "../services/browserApi";
import { requestMultipartJSON } from "../services/httpTransport";
import {
  buttonRowEndStyle,
  createDialogStyle,
  dialogBackdropStyle,
  dialogHeaderStyle,
  errorTextStyle,
  formGridStyle,
  iconButtonStyle,
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
  surfaceHeaderStyle,
  surfacePanelStyle,
} from "./landingAdminStyles";

type JobResource = {
  job_id?: string;
  status?: string;
  cancelable?: boolean;
  progress?: {
    completed?: number;
    total?: number | null;
  };
  result_summary?: { code?: string; resource_refs?: unknown[] } | null;
  error_summary?: { code?: string } | null;
};

type JobResourceRef = {
  kind?: unknown;
  id?: unknown;
};

const terminalJobStates = new Set(["succeeded", "failed", "canceled"]);

function importedIncidentIDFromJob(job: JobResource | null): string | null {
  const refs = job?.result_summary?.resource_refs;
  if (!Array.isArray(refs)) {
    return null;
  }
  const incidentRefs = refs.filter((ref): ref is JobResourceRef => {
    if (typeof ref !== "object" || ref === null || Array.isArray(ref)) {
      return false;
    }
    const candidate = ref as JobResourceRef;
    return candidate.kind === "incident" && typeof candidate.id === "string";
  });
  if (incidentRefs.length !== 1) {
    return null;
  }
  const incidentID = incidentRefs[0]?.id;
  return typeof incidentID === "string" && incidentID.trim() !== ""
    ? incidentID
    : null;
}

export function IncidentImportPanel({
  onOpenImportedIncident,
}: {
  onOpenImportedIncident: (incidentId: string) => void;
}) {
  const [file, setFile] = useState<File | null>(null);
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [job, setJob] = useState<JobResource | null>(null);
  const [status, setStatus] = useState("Incident import idle.");
  const [error, setError] = useState<APIError | null>(null);
  const pollTimer = useRef<number | null>(null);
  const importedIncidentID = importedIncidentIDFromJob(job);

  const loadJob = useCallback(async (jobID: string) => {
    const result = await fetchJSON<{ data: JobResource }>(
      `/api/v1/jobs/${jobID}`,
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      return;
    }
    const nextJob = (result.payload as { data: JobResource }).data;
    setJob(nextJob);
    setStatus(`Incident import job ${nextJob.status ?? "updated"}.`);
  }, []);

  useEffect(() => {
    if (
      job?.job_id === undefined ||
      job.status === undefined ||
      terminalJobStates.has(job.status)
    ) {
      return;
    }
    pollTimer.current = window.setTimeout(() => {
      void loadJob(job.job_id ?? "");
    }, 1000);
    return () => {
      if (pollTimer.current !== null) {
        window.clearTimeout(pollTimer.current);
        pollTimer.current = null;
      }
    };
  }, [job, loadJob]);

  async function submitImport() {
    if (file === null) {
      setStatus("Select an incident bundle first.");
      return;
    }
    const form = new FormData();
    form.append(
      "metadata",
      new Blob(
        [
          JSON.stringify({
            client_txn_id: clientTxnID("incident-import"),
          }),
        ],
        { type: "application/json" },
      ),
    );
    form.append("file", file);
    setStatus("Submitting incident import.");
    const response = await requestMultipartJSON<{
      data?: JobResource;
      error?: APIError;
    }>("/api/v1/incident-bundles/import", form);
    const payload = response.payload;
    if (!response.ok) {
      setError(extractError(payload));
      setStatus("Incident import failed to start.");
      return;
    }
    const nextJob = "data" in payload ? (payload.data ?? null) : null;
    setError(null);
    setJob(nextJob);
    setStatus(
      `Incident import queued${nextJob?.job_id ? `: ${nextJob.job_id}` : "."}`,
    );
    setImportDialogOpen(false);
    setFile(null);
    if (nextJob?.job_id) {
      void loadJob(nextJob.job_id);
    }
  }

  return (
    <section style={surfacePanelStyle}>
      <div style={surfaceHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Incident portability</p>
          <h2 style={sectionTitleStyle}>Incident import</h2>
        </div>
        <button
          style={primaryButtonStyle}
          type="button"
          onClick={() => setImportDialogOpen(true)}
        >
          Import incident
        </button>
      </div>
      <div style={jobPanelStyle}>
        <p style={strongTextStyle}>Import progress</p>
        <p style={metadataTextStyle}>
          {job === null
            ? "No import job is active."
            : `${job.status ?? "queued"} · ${job.progress?.completed ?? 0}/${job.progress?.total ?? "?"}`}
        </p>
        {importedIncidentID !== null ? (
          <button
            style={primaryButtonStyle}
            type="button"
            onClick={() => {
              onOpenImportedIncident(importedIncidentID);
            }}
          >
            Open imported incident
          </button>
        ) : null}
      </div>
      {importDialogOpen ? (
        <div style={dialogBackdropStyle}>
          <section
            aria-label="Import incident bundle"
            aria-modal="true"
            role="dialog"
            style={createDialogStyle}
          >
            <header style={dialogHeaderStyle}>
              <div>
                <p style={sectionEyebrowStyle}>Staged import</p>
                <h3 style={subsectionTitleStyle}>Select incident bundle</h3>
              </div>
              <button
                aria-label="Close incident import"
                style={iconButtonStyle}
                type="button"
                onClick={() => setImportDialogOpen(false)}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <div style={formGridStyle}>
              <label style={labelBlockStyle}>
                Bundle file
                <input
                  aria-label="Incident bundle file"
                  style={inputStyle}
                  type="file"
                  onChange={(event) => {
                    setFile(event.currentTarget.files?.[0] ?? null);
                  }}
                />
              </label>
            </div>
            <div style={buttonRowEndStyle}>
              <button
                style={secondaryButtonStyle}
                type="button"
                onClick={() => setImportDialogOpen(false)}
              >
                Cancel
              </button>
              <button
                style={primaryButtonStyle}
                type="button"
                onClick={() => {
                  void submitImport();
                }}
              >
                Start import
              </button>
            </div>
          </section>
        </div>
      ) : null}
      <p aria-live="polite" role="status" style={statusTextStyle}>
        {status}
      </p>
      <p
        aria-live="assertive"
        role={error === null ? undefined : "alert"}
        style={errorTextStyle}
      >
        {publicErrorView(error)?.code ?? ""}
      </p>
    </section>
  );
}
