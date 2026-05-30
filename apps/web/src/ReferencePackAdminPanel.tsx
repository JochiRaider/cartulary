import {
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackErrorTestId,
  referencePackFileInputTestId,
  referencePackImportButtonTestId,
  referencePackJobStatusTestId,
  referencePackRefreshAllButtonTestId,
  referencePackRefreshSelectedButtonTestId,
  referencePackReloadButtonTestId,
  referencePackRowTestId,
} from "@cartulary/ui-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";

import {
  type APIError,
  clientTxnID,
  csrfCookieName,
  csrfHeaderName,
  extractError,
  fetchJSON,
  readCookie,
} from "./browserApi";
import type { SessionData } from "./phase1Client";

type ReferencePackVersion = {
  pack_key: string;
  pack_version: string;
  pack_kind: string;
  pack_version_state: string;
  active: boolean;
  verification_result: string;
};

type JobResource = {
  job_id: string;
  status: string;
  cancelable: boolean;
  progress: {
    completed: number;
    total: number | null;
  };
  result_summary: { code: string } | null;
  error_summary: { code: string; details?: Record<string, unknown> } | null;
};

type ReferencePackListEnvelope = {
  data: {
    pack_versions: ReferencePackVersion[];
  };
};

type JobEnvelope = {
  data: JobResource;
};

type ReferencePackAdminPanelProps = {
  session: SessionData;
};

const terminalJobStates = new Set(["succeeded", "failed", "canceled"]);

export function ReferencePackAdminPanel({
  session,
}: ReferencePackAdminPanelProps) {
  const [packs, setPacks] = useState<ReferencePackVersion[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set());
  const [file, setFile] = useState<File | null>(null);
  const [job, setJob] = useState<JobResource | null>(null);
  const [error, setError] = useState<APIError | null>(null);
  const [status, setStatus] = useState("Idle");
  const pollTimer = useRef<number | null>(null);

  const activePackKeys = useMemo(() => {
    return Array.from(new Set(packs.map((pack) => pack.pack_key))).sort();
  }, [packs]);

  const loadPacks = useCallback(async () => {
    const result = await fetchJSON<ReferencePackListEnvelope>(
      "/api/v1/reference-packs?limit=100",
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus("Reference packs unavailable");
      return;
    }
    setError(null);
    setPacks((result.payload as ReferencePackListEnvelope).data.pack_versions);
    setStatus("Reference packs loaded");
  }, []);

  const loadJob = useCallback(
    async (jobID: string) => {
      const result = await fetchJSON<JobEnvelope>(`/api/v1/jobs/${jobID}`);
      if (!result.ok) {
        setError(extractError(result.payload));
        return;
      }
      const nextJob = (result.payload as JobEnvelope).data;
      setJob(nextJob);
      setStatus(`Job ${nextJob.status}`);
      if (terminalJobStates.has(nextJob.status)) {
        await loadPacks();
      }
    },
    [loadPacks],
  );

  useEffect(() => {
    if (job === null || terminalJobStates.has(job.status)) {
      return;
    }
    pollTimer.current = window.setTimeout(() => {
      void loadJob(job.job_id);
    }, 1000);
    return () => {
      if (pollTimer.current !== null) {
        window.clearTimeout(pollTimer.current);
        pollTimer.current = null;
      }
    };
  }, [job, loadJob]);

  if (!session.is_deployment_admin) {
    return null;
  }

  async function submitUpload() {
    if (file === null) {
      setStatus("Select a bundle first");
      return;
    }
    const form = new FormData();
    form.append(
      "metadata",
      new Blob(
        [
          JSON.stringify({
            client_txn_id: clientTxnID("reference-pack-import"),
          }),
        ],
        { type: "application/json" },
      ),
    );
    form.append("file", file);
    const headers = new Headers();
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken !== null && csrfToken !== "") {
      headers.set(csrfHeaderName, csrfToken);
    }
    const response = await fetch("/api/v1/reference-packs/import", {
      method: "POST",
      credentials: "include",
      headers,
      body: form,
    });
    const payload = (await response.json()) as
      | JobEnvelope
      | { error: APIError };
    if (!response.ok) {
      setError(extractError(payload));
      setStatus("Import failed to start");
      return;
    }
    const nextJob = (payload as JobEnvelope).data;
    setError(null);
    setJob(nextJob);
    setStatus("Import queued");
    void loadJob(nextJob.job_id);
  }

  async function runPackAction(
    pack: ReferencePackVersion,
    action: "activate" | "disable" | "reverify",
  ) {
    const result = await fetchJSON<JobEnvelope | { data: unknown }>(
      `/api/v1/reference-packs/${encodeURIComponent(pack.pack_key)}/${encodeURIComponent(pack.pack_version)}/${action}`,
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: clientTxnID(`reference-pack-${action}`),
        }),
      },
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus(`${action} failed`);
      return;
    }
    setError(null);
    if (action === "reverify") {
      const nextJob = (result.payload as JobEnvelope).data;
      setJob(nextJob);
      setStatus("Reverify queued");
      void loadJob(nextJob.job_id);
      return;
    }
    setStatus(`${action} complete`);
    await loadPacks();
  }

  async function refreshSelected(all: boolean) {
    const packKeys = all ? [] : Array.from(selectedKeys).sort();
    const body: Record<string, unknown> = {
      client_txn_id: clientTxnID("reference-pack-refresh"),
    };
    if (!all) {
      body.pack_keys = packKeys;
    }
    const result = await fetchJSON<JobEnvelope>(
      "/api/v1/reference-packs/refresh",
      {
        method: "POST",
        body: JSON.stringify(body),
      },
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus("Refresh failed to start");
      return;
    }
    const nextJob = (result.payload as JobEnvelope).data;
    setError(null);
    setJob(nextJob);
    setStatus("Refresh queued");
    void loadJob(nextJob.job_id);
  }

  async function cancelJob() {
    if (job === null) {
      return;
    }
    const result = await fetchJSON<JobEnvelope>(
      `/api/v1/jobs/${job.job_id}/cancel`,
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: clientTxnID("reference-pack-cancel"),
        }),
      },
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus("Cancel rejected");
      return;
    }
    const nextJob = (result.payload as JobEnvelope).data;
    setError(null);
    setJob(nextJob);
    setStatus("Cancel requested");
  }

  return (
    <section data-testid={referencePackAdminPanelTestId()} style={panelStyle}>
      <div style={panelHeaderStyle}>
        <div>
          <p style={eyebrowStyle}>Reference packs</p>
          <h2 style={titleStyle}>Pack operations</h2>
        </div>
        <button
          data-testid={referencePackReloadButtonTestId()}
          type="button"
          style={buttonStyle}
          onClick={() => void loadPacks()}
        >
          Refresh
        </button>
      </div>

      <div style={uploadRowStyle}>
        <input
          aria-label="Reference pack bundle file"
          data-testid={referencePackFileInputTestId()}
          type="file"
          onChange={(event) => {
            setFile(event.currentTarget.files?.[0] ?? null);
          }}
        />
        <button
          data-testid={referencePackImportButtonTestId()}
          type="button"
          style={primaryButtonStyle}
          onClick={() => void submitUpload()}
        >
          Import
        </button>
      </div>

      <div style={jobStyle} data-testid={referencePackJobStatusTestId()}>
        <span>{status}</span>
        {job !== null ? (
          <span>
            {job.status} · {job.progress.completed}/{job.progress.total ?? "?"}
          </span>
        ) : null}
        {job !== null &&
        !terminalJobStates.has(job.status) &&
        job.cancelable ? (
          <button
            data-testid={referencePackCancelButtonTestId()}
            type="button"
            style={buttonStyle}
            onClick={() => void cancelJob()}
          >
            Cancel
          </button>
        ) : null}
      </div>

      <div style={actionBarStyle}>
        <button
          data-testid={referencePackRefreshAllButtonTestId()}
          type="button"
          style={buttonStyle}
          onClick={() => void refreshSelected(true)}
        >
          Refresh all
        </button>
        <button
          data-testid={referencePackRefreshSelectedButtonTestId()}
          type="button"
          style={buttonStyle}
          disabled={selectedKeys.size === 0}
          onClick={() => void refreshSelected(false)}
        >
          Refresh selected
        </button>
      </div>

      <div style={packListStyle}>
        {packs.map((pack) => (
          <div
            key={`${pack.pack_key}/${pack.pack_version}`}
            style={packRowStyle}
            data-testid={referencePackRowTestId(
              pack.pack_key,
              pack.pack_version,
            )}
          >
            <label style={packLabelStyle}>
              <input
                type="checkbox"
                checked={selectedKeys.has(pack.pack_key)}
                onChange={(event) => {
                  setSelectedKeys((current) => {
                    const next = new Set(current);
                    if (event.currentTarget.checked) {
                      next.add(pack.pack_key);
                    } else {
                      next.delete(pack.pack_key);
                    }
                    return next;
                  });
                }}
              />
              <span>
                {pack.pack_key}@{pack.pack_version}
              </span>
            </label>
            <span>{pack.pack_version_state}</span>
            <span>{pack.active ? "active" : ""}</span>
            <button
              type="button"
              style={buttonStyle}
              disabled={pack.active}
              onClick={() => void runPackAction(pack, "activate")}
            >
              Activate
            </button>
            <button
              type="button"
              style={buttonStyle}
              onClick={() => void runPackAction(pack, "disable")}
            >
              Disable
            </button>
            <button
              type="button"
              style={buttonStyle}
              onClick={() => void runPackAction(pack, "reverify")}
            >
              Reverify
            </button>
          </div>
        ))}
        {packs.length === 0 ? <p style={emptyStyle}>No packs loaded</p> : null}
      </div>

      <p data-testid={referencePackErrorTestId()} style={errorStyle}>
        {error?.code ?? ""}
      </p>
      <p style={emptyStyle}>{activePackKeys.length} pack keys visible</p>
    </section>
  );
}

const panelStyle = {
  minWidth: 0,
  padding: "1.25rem",
  borderRadius: "0.75rem",
  border: "1px solid rgb(185 204 196 / 0.8)",
  background: "rgb(255 255 255 / 0.72)",
};

const panelHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "center",
};

const eyebrowStyle = {
  margin: 0,
  color: "rgb(70 96 89)",
  fontSize: "0.75rem",
  textTransform: "uppercase" as const,
};

const titleStyle = {
  margin: "0.2rem 0 0",
  fontSize: "1.05rem",
};

const uploadRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  gap: "0.75rem",
  marginTop: "1rem",
};

const actionBarStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.5rem",
  marginTop: "0.75rem",
};

const jobStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  alignItems: "center",
  gap: "0.75rem",
  minHeight: "2.25rem",
  marginTop: "0.75rem",
  fontSize: "0.85rem",
};

const packListStyle = {
  display: "grid",
  gap: "0.45rem",
  marginTop: "0.75rem",
};

const packRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(10rem, 1fr) auto auto repeat(3, auto)",
  gap: "0.5rem",
  alignItems: "center",
  minHeight: "2.25rem",
  fontSize: "0.82rem",
};

const packLabelStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  minWidth: 0,
};

const buttonStyle = {
  border: "1px solid rgb(139 165 157)",
  borderRadius: "0.45rem",
  background: "rgb(255 255 255)",
  padding: "0.45rem 0.65rem",
  cursor: "pointer",
};

const primaryButtonStyle = {
  ...buttonStyle,
  color: "white",
  borderColor: "rgb(22 95 75)",
  background: "rgb(22 95 75)",
};

const errorStyle = {
  minHeight: "1.25rem",
  margin: "0.75rem 0 0",
  color: "rgb(142 45 36)",
  fontSize: "0.82rem",
};

const emptyStyle = {
  margin: "0.5rem 0 0",
  color: "rgb(70 96 89)",
  fontSize: "0.82rem",
};
