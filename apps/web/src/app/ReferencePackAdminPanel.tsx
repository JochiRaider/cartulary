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
import { X } from "lucide-react";
import {
  type CSSProperties,
  forwardRef,
  useCallback,
  useEffect,
  useImperativeHandle,
  useMemo,
  useRef,
  useState,
} from "react";

import {
  type APIError,
  clientTxnID,
  csrfCookieName,
  csrfHeaderName,
  extractError,
  fetchJSON,
  readCookie,
} from "../services/browserApi";
import type { SessionData } from "./phase1Client";

type ReferencePackVersion = {
  pack_key: string;
  pack_version: string;
  pack_kind: string;
  pack_version_state: string;
  active: boolean;
  source_identifier: string | null;
  manifest_sha256: string;
  payload_sha256: string;
  pack_contract_version: string;
  verification_method: string;
  verification_result: string;
  signer_key_id: string | null;
  previous_active_version: string | null;
  imported_by_user_id: string | null;
  imported_at: string;
  activated_by_user_id: string | null;
  activated_at: string | null;
};

export type ReferencePackJobResource = {
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
  meta?: {
    paging?: ReferencePackPaging;
  };
};

type ReferencePackPaging = {
  limit?: number;
  has_more: boolean;
  next_cursor: string | null;
};

type JobEnvelope = {
  data: ReferencePackJobResource;
};

type ReferencePackAdminPanelProps = {
  activeJob?: ReferencePackJobResource | null;
  onJobChange?: (job: ReferencePackJobResource | null) => void;
  session: SessionData;
};

export type ReferencePackAdminPanelHandle = {
  cancelJob: () => Promise<void>;
  importBundle: () => Promise<void>;
  refreshAll: () => Promise<void>;
  refreshPacks: () => Promise<void>;
  refreshSelected: () => Promise<void>;
};

const terminalJobStates = new Set(["succeeded", "failed", "canceled"]);
const defaultPackPaging: ReferencePackPaging = {
  limit: 100,
  has_more: false,
  next_cursor: null,
};

export const ReferencePackAdminPanel = forwardRef<
  ReferencePackAdminPanelHandle,
  ReferencePackAdminPanelProps
>(function ReferencePackAdminPanel({ activeJob, onJobChange, session }, ref) {
  const [packs, setPacks] = useState<ReferencePackVersion[]>([]);
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(new Set());
  const [file, setFile] = useState<File | null>(null);
  const [importDialogOpen, setImportDialogOpen] = useState(false);
  const [internalJob, setInternalJob] =
    useState<ReferencePackJobResource | null>(null);
  const [error, setError] = useState<APIError | null>(null);
  const [packSearch, setPackSearch] = useState("");
  const [packVersionStateFilter, setPackVersionStateFilter] = useState("");
  const [verificationResultFilter, setVerificationResultFilter] = useState("");
  const [activeFilter, setActiveFilter] = useState("");
  const [packPaging, setPackPaging] = useState(defaultPackPaging);
  const [status, setStatus] = useState("Idle");
  const pollTimer = useRef<number | null>(null);
  const requestSeqRef = useRef(0);
  const initialLoadRef = useRef(false);
  const acceptedPackQueryRef = useRef({
    active: "",
    packVersionState: "",
    search: "",
    verificationResult: "",
  });
  const job = activeJob === undefined ? internalJob : activeJob;

  const activePackKeys = useMemo(() => {
    return Array.from(new Set(packs.map((pack) => pack.pack_key))).sort();
  }, [packs]);

  const loadPacks = useCallback(
    async (options?: { append?: boolean; cursorToken?: string | null }) => {
      const sequence = requestSeqRef.current + 1;
      requestSeqRef.current = sequence;
      const params = new URLSearchParams({ limit: "100" });
      const cursorToken = options?.cursorToken?.trim() ?? "";
      const query =
        options?.append === true
          ? acceptedPackQueryRef.current
          : {
              active: activeFilter,
              packVersionState: packVersionStateFilter,
              search: packSearch.trim(),
              verificationResult: verificationResultFilter,
            };
      if (cursorToken !== "") {
        params.set("cursor_token", cursorToken);
      }
      if (query.search !== "") {
        params.set("search", query.search);
      }
      if (query.packVersionState !== "") {
        params.set("pack_version_state", query.packVersionState);
      }
      if (query.verificationResult !== "") {
        params.set("verification_result", query.verificationResult);
      }
      if (query.active !== "") {
        params.set("active", query.active);
      }

      setStatus(
        options?.append === true
          ? "Loading more reference packs"
          : "Searching reference packs",
      );
      const result = await fetchJSON<ReferencePackListEnvelope>(
        `/api/v1/reference-packs?${params.toString()}`,
      );
      if (requestSeqRef.current !== sequence) {
        return;
      }
      if (!result.ok) {
        setError(extractError(result.payload));
        setStatus("Reference packs unavailable");
        return;
      }
      const envelope = result.payload as ReferencePackListEnvelope;
      const nextPacks = envelope.data.pack_versions;
      if (options?.append !== true) {
        acceptedPackQueryRef.current = query;
      }
      setError(null);
      setPacks((current) =>
        options?.append === true ? [...current, ...nextPacks] : nextPacks,
      );
      setPackPaging(envelope.meta?.paging ?? defaultPackPaging);
      setStatus("Reference packs loaded");
    },
    [
      activeFilter,
      packSearch,
      packVersionStateFilter,
      verificationResultFilter,
    ],
  );

  const updateJob = useCallback(
    (nextJob: ReferencePackJobResource | null) => {
      setInternalJob(nextJob);
      onJobChange?.(nextJob);
    },
    [onJobChange],
  );

  const loadJob = useCallback(
    async (jobID: string) => {
      const result = await fetchJSON<JobEnvelope>(`/api/v1/jobs/${jobID}`);
      if (!result.ok) {
        setError(extractError(result.payload));
        return;
      }
      const nextJob = (result.payload as JobEnvelope).data;
      updateJob(nextJob);
      setStatus(`Job ${nextJob.status}`);
      if (terminalJobStates.has(nextJob.status)) {
        await loadPacks();
      }
    },
    [loadPacks, updateJob],
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

  useEffect(() => {
    if (initialLoadRef.current) {
      return;
    }
    initialLoadRef.current = true;
    void loadPacks();
  }, [loadPacks]);

  useImperativeHandle(ref, () => ({
    cancelJob,
    importBundle: submitUpload,
    refreshAll: () => refreshSelected(true),
    refreshPacks: loadPacks,
    refreshSelected: () => refreshSelected(false),
  }));

  if (!session.is_deployment_admin) {
    return (
      <section data-testid={referencePackAdminPanelTestId()} style={panelStyle}>
        <div style={panelHeaderStyle}>
          <div>
            <p style={eyebrowStyle}>Reference packs</p>
            <h2 style={titleStyle}>Pack operations</h2>
          </div>
        </div>
        <p style={emptyStyle}>
          Deployment admin access is required for reference-pack import,
          refresh, activation, and verification actions.
        </p>
      </section>
    );
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
    updateJob(nextJob);
    setStatus("Import queued");
    setImportDialogOpen(false);
    setFile(null);
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
      updateJob(nextJob);
      setStatus("Reverify queued");
      void loadJob(nextJob.job_id);
      return;
    }
    setStatus(`${action} complete`);
    await loadPacks();
  }

  async function refreshSelected(all: boolean) {
    if (!all && selectedKeys.size === 0) {
      setStatus("Select packs first");
      return;
    }
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
    updateJob(nextJob);
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
    updateJob(nextJob);
    setStatus("Cancel requested");
  }

  return (
    <section data-testid={referencePackAdminPanelTestId()} style={panelStyle}>
      <div style={panelHeaderStyle}>
        <div>
          <p style={eyebrowStyle}>Reference packs</p>
          <h2 style={titleStyle}>Pack operations</h2>
        </div>
        <div style={headerActionRowStyle}>
          <button
            data-testid={
              importDialogOpen ? undefined : referencePackImportButtonTestId()
            }
            type="button"
            style={primaryButtonStyle}
            onClick={() => setImportDialogOpen(true)}
          >
            Import pack
          </button>
          <button
            data-testid={referencePackReloadButtonTestId()}
            type="button"
            style={buttonStyle}
            onClick={() => void loadPacks()}
          >
            Refresh
          </button>
        </div>
      </div>

      <div style={searchRowStyle}>
        <label style={filterFieldStyle} htmlFor="reference-pack-search">
          Search reference packs
          <input
            aria-label="Search reference packs"
            id="reference-pack-search"
            style={filterInputStyle}
            value={packSearch}
            onChange={(event) => {
              setPackSearch(event.target.value);
            }}
            onKeyDown={(event) => {
              if (event.key === "Enter") {
                void loadPacks();
              }
            }}
            placeholder="Pack key, kind, version, source"
          />
        </label>
        <label style={filterFieldStyle} htmlFor="reference-pack-state-filter">
          State
          <select
            aria-label="Reference pack state"
            id="reference-pack-state-filter"
            style={filterInputStyle}
            value={packVersionStateFilter}
            onChange={(event) => {
              setPackVersionStateFilter(event.target.value);
            }}
          >
            <option value="">Any state</option>
            <option value="staged">staged</option>
            <option value="verified_available">verified_available</option>
            <option value="disabled">disabled</option>
            <option value="failed">failed</option>
            <option value="missing">missing</option>
          </select>
        </label>
        <label
          style={filterFieldStyle}
          htmlFor="reference-pack-verification-filter"
        >
          Verification
          <select
            aria-label="Reference pack verification result"
            id="reference-pack-verification-filter"
            style={filterInputStyle}
            value={verificationResultFilter}
            onChange={(event) => {
              setVerificationResultFilter(event.target.value);
            }}
          >
            <option value="">Any verification</option>
            <option value="pending">pending</option>
            <option value="passed">passed</option>
            <option value="failed">failed</option>
          </select>
        </label>
        <label style={filterFieldStyle} htmlFor="reference-pack-active-filter">
          Active
          <select
            aria-label="Reference pack active state"
            id="reference-pack-active-filter"
            style={filterInputStyle}
            value={activeFilter}
            onChange={(event) => {
              setActiveFilter(event.target.value);
            }}
          >
            <option value="">Any active state</option>
            <option value="true">true</option>
            <option value="false">false</option>
          </select>
        </label>
        <button
          type="button"
          style={buttonStyle}
          onClick={() => void loadPacks()}
        >
          Search
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
        <table style={packTableStyle}>
          <thead>
            <tr>
              <th style={tableHeaderCellStyle}>Pack</th>
              <th style={tableHeaderCellStyle}>State</th>
              <th style={tableHeaderCellStyle}>Verification</th>
              <th style={tableHeaderCellStyle}>Active</th>
              <th style={tableHeaderActionCellStyle}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {packs.map((pack) => {
              const canActivate =
                pack.pack_version_state === "verified_available" &&
                !pack.active;
              const canDisable =
                pack.pack_version_state === "verified_available";
              const canReverify = pack.pack_version_state !== "staged";
              return (
                <tr
                  key={`${pack.pack_key}/${pack.pack_version}`}
                  style={packRowStyle}
                  data-testid={referencePackRowTestId(
                    pack.pack_key,
                    pack.pack_version,
                  )}
                >
                  <td style={primaryCellStyle}>
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
                        <strong>{pack.pack_key}</strong>
                        <span style={mutedTextStyle}>
                          @{pack.pack_version} · {pack.pack_kind}
                        </span>
                      </span>
                    </label>
                  </td>
                  <td style={tableCellStyle}>{pack.pack_version_state}</td>
                  <td style={tableCellStyle}>
                    {pack.verification_result}
                    <span style={mutedTextStyle}>
                      {pack.verification_method}
                    </span>
                  </td>
                  <td style={tableCellStyle}>{pack.active ? "Yes" : "No"}</td>
                  <td style={tableActionCellStyle}>
                    <span style={actionButtonsStyle}>
                      <button
                        type="button"
                        style={buttonStyle}
                        disabled={!canActivate}
                        onClick={() => void runPackAction(pack, "activate")}
                      >
                        Activate
                      </button>
                      <button
                        type="button"
                        style={buttonStyle}
                        disabled={!canDisable}
                        onClick={() => void runPackAction(pack, "disable")}
                      >
                        Disable
                      </button>
                      <button
                        type="button"
                        style={buttonStyle}
                        disabled={!canReverify}
                        onClick={() => void runPackAction(pack, "reverify")}
                      >
                        Reverify
                      </button>
                    </span>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        {packs.length === 0 ? <p style={emptyStyle}>No packs loaded</p> : null}
        {packPaging.has_more && packPaging.next_cursor !== null ? (
          <button
            type="button"
            style={buttonStyle}
            onClick={() =>
              void loadPacks({
                append: true,
                cursorToken: packPaging.next_cursor,
              })
            }
          >
            Load more
          </button>
        ) : null}
      </div>

      <p data-testid={referencePackErrorTestId()} style={errorStyle}>
        {error?.code ?? ""}
      </p>
      <p style={emptyStyle}>{activePackKeys.length} pack keys loaded</p>
      {importDialogOpen ? (
        <div style={dialogBackdropStyle}>
          <section
            aria-label="Import reference pack"
            aria-modal="true"
            role="dialog"
            style={dialogStyle}
          >
            <header style={dialogHeaderStyle}>
              <div>
                <p style={eyebrowStyle}>Reference packs</p>
                <h3 style={titleStyle}>Import pack</h3>
              </div>
              <button
                aria-label="Close reference pack import"
                style={iconButtonStyle}
                type="button"
                onClick={() => setImportDialogOpen(false)}
              >
                <X aria-hidden="true" size={16} />
              </button>
            </header>
            <label style={filterFieldStyle}>
              Bundle file
              <input
                aria-label="Reference pack bundle file"
                data-testid={referencePackFileInputTestId()}
                style={fileInputStyle}
                type="file"
                onChange={(event) => {
                  setFile(event.currentTarget.files?.[0] ?? null);
                }}
              />
            </label>
            <div style={dialogButtonRowStyle}>
              <button
                type="button"
                style={buttonStyle}
                onClick={() => setImportDialogOpen(false)}
              >
                Cancel
              </button>
              <button
                data-testid={referencePackImportButtonTestId()}
                type="button"
                style={primaryButtonStyle}
                onClick={() => void submitUpload()}
              >
                Import
              </button>
            </div>
          </section>
        </div>
      ) : null}
    </section>
  );
});

const panelStyle = {
  boxSizing: "border-box" as const,
  minWidth: 0,
  padding: "1.25rem",
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
};

const panelHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "center",
};

const eyebrowStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.75rem",
  textTransform: "uppercase" as const,
};

const titleStyle = {
  margin: "0.2rem 0 0",
  fontSize: "1.05rem",
};

const searchRowStyle = {
  display: "grid",
  gridTemplateColumns:
    "minmax(18rem, 1.4fr) repeat(3, minmax(10rem, 1fr)) auto",
  gap: "0.75rem",
  marginTop: "1rem",
  minWidth: 0,
  alignItems: "end",
} satisfies CSSProperties;

const filterFieldStyle = {
  display: "grid",
  gap: "0.35rem",
  minWidth: 0,
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
  fontWeight: 700,
} satisfies CSSProperties;

const filterInputStyle = {
  boxSizing: "border-box" as const,
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  color: "var(--ct-component-text-input-textColor)",
  padding: "var(--ct-component-text-input-padding)",
} satisfies CSSProperties;

const fileInputStyle = {
  boxSizing: "border-box" as const,
  minWidth: 0,
  maxWidth: "100%",
  color: "var(--ct-colors-ink-muted)",
} satisfies CSSProperties;

const headerActionRowStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.5rem",
  justifyContent: "flex-end",
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

const dialogBackdropStyle = {
  position: "fixed" as const,
  inset: 0,
  zIndex: 40,
  display: "grid",
  placeItems: "center",
  padding: "1.5rem",
  background: "rgba(10, 13, 18, 0.68)",
} satisfies CSSProperties;

const dialogStyle = {
  width: "min(42rem, 100%)",
  display: "grid",
  gap: "1rem",
  padding: "1.25rem",
  borderRadius: "var(--ct-rounded-md)",
  border: "var(--ct-border-strong)",
  background: "var(--ct-colors-surface-1)",
  boxShadow: "var(--ct-elevation-panel)",
} satisfies CSSProperties;

const dialogHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "flex-start",
} satisfies CSSProperties;

const iconButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "2rem",
  height: "2rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink)",
  cursor: "pointer",
} satisfies CSSProperties;

const dialogButtonRowStyle = {
  display: "flex",
  justifyContent: "flex-end",
  flexWrap: "wrap" as const,
  gap: "0.5rem",
} satisfies CSSProperties;

const packListStyle = {
  marginTop: "0.75rem",
  minWidth: 0,
  overflowX: "auto" as const,
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-rounded-sm)",
  background: "var(--ct-colors-surface-1)",
};

const packTableStyle = {
  width: "100%",
  minWidth: "58rem",
  borderCollapse: "collapse" as const,
} satisfies CSSProperties;

const packRowStyle = {
  borderBottom: "var(--ct-border-hairline)",
  fontSize: "0.82rem",
};

const packLabelStyle = {
  display: "flex",
  alignItems: "center",
  gap: "0.4rem",
  minWidth: 0,
};

const tableHeaderCellStyle = {
  padding: "0.7rem 0.85rem",
  borderBottom: "var(--ct-border-hairline)",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.68rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  textAlign: "left" as const,
  whiteSpace: "nowrap" as const,
} satisfies CSSProperties;

const tableHeaderActionCellStyle = {
  ...tableHeaderCellStyle,
  textAlign: "right" as const,
} satisfies CSSProperties;

const tableCellStyle = {
  padding: "0.75rem 0.85rem",
  color: "var(--ct-colors-ink-muted)",
  verticalAlign: "top" as const,
} satisfies CSSProperties;

const primaryCellStyle = {
  ...tableCellStyle,
  color: "var(--ct-colors-ink)",
} satisfies CSSProperties;

const tableActionCellStyle = {
  ...tableCellStyle,
  textAlign: "right" as const,
} satisfies CSSProperties;

const actionButtonsStyle = {
  display: "inline-flex",
  flexWrap: "wrap" as const,
  justifyContent: "flex-end",
  gap: "0.45rem",
} satisfies CSSProperties;

const mutedTextStyle = {
  display: "block",
  color: "var(--ct-colors-ink-subtle)",
  fontSize: "0.76rem",
  fontWeight: 400,
} satisfies CSSProperties;

const buttonStyle = {
  border: "var(--ct-component-button-secondary-border)",
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "var(--ct-component-button-secondary-padding)",
  cursor: "pointer",
};

const primaryButtonStyle = {
  ...buttonStyle,
  border: "none",
  color: "var(--ct-component-button-primary-textColor)",
  background: "var(--ct-component-button-primary-backgroundColor)",
  padding: "var(--ct-component-button-primary-padding)",
};

const errorStyle = {
  minHeight: "1.25rem",
  margin: "0.75rem 0 0",
  color: "var(--ct-colors-semantic-conflict)",
  fontSize: "0.82rem",
};

const emptyStyle = {
  margin: "0.5rem 0 0",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.82rem",
};
