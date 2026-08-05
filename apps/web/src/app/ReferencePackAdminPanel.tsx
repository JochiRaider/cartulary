import {
  referencePackAdminPanelTestId,
  referencePackCancelButtonTestId,
  referencePackErrorTestId,
  referencePackFileInputTestId,
  referencePackImportButtonTestId,
  referencePackJobStatusTestId,
  referencePackListStatusTestId,
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

import { type APIError, extractError } from "../services/browserApi";
import {
  activateReferencePackVersion,
  cancelReferencePackJob,
  disableReferencePackVersion,
  importReferencePackBundle,
  listReferencePacks,
  loadReferencePackJob,
  type ReferencePackAction,
  refreshReferencePacks,
  reverifyReferencePackVersion,
} from "../services/referencePacks";
import {
  type DeploymentAdminSession,
  defaultReferencePackPaging,
  type ReferencePackJobResource,
  type ReferencePackQuery,
  type ReferencePackVersion,
  terminalReferencePackJobStates,
} from "./referencePackAdminModel";

export type { ReferencePackJobResource } from "./referencePackAdminModel";

type ReferencePackAdminPanelProps = {
  activeJob?: ReferencePackJobResource | null;
  onJobChange?: (job: ReferencePackJobResource | null) => void;
  session: DeploymentAdminSession;
};

type ReferencePackAdminPanelHandle = {
  cancelJob: () => Promise<void>;
  importBundle: () => Promise<void>;
  refreshAll: () => Promise<void>;
  refreshPacks: () => Promise<void>;
  refreshSelected: () => Promise<void>;
};

type ReferencePackListStatus = "idle" | "searching" | "loaded" | "unavailable";

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
  const [packPaging, setPackPaging] = useState(defaultReferencePackPaging);
  const [status, setStatus] = useState("Idle");
  const [listStatus, setListStatus] = useState<ReferencePackListStatus>("idle");
  const pollTimer = useRef<number | null>(null);
  const requestSeqRef = useRef(0);
  const initialLoadRef = useRef(false);
  const acceptedPackQueryRef = useRef<ReferencePackQuery>({
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
      if (!session.is_deployment_admin) {
        return;
      }
      const sequence = requestSeqRef.current + 1;
      requestSeqRef.current = sequence;
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
      setListStatus("searching");
      const result = await listReferencePacks({ cursorToken, query });
      if (requestSeqRef.current !== sequence) {
        return;
      }
      if (!result.ok) {
        setError(extractError(result.payload));
        setListStatus("unavailable");
        return;
      }
      const envelope = result.payload;
      const nextPacks = envelope.data.pack_versions;
      if (options?.append !== true) {
        acceptedPackQueryRef.current = query;
      }
      setError(null);
      setPacks((current) =>
        options?.append === true ? [...current, ...nextPacks] : nextPacks,
      );
      setPackPaging(envelope.meta?.paging ?? defaultReferencePackPaging);
      setListStatus("loaded");
    },
    [
      activeFilter,
      packSearch,
      packVersionStateFilter,
      session.is_deployment_admin,
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
      if (!session.is_deployment_admin) {
        return;
      }
      const result = await loadReferencePackJob(jobID);
      if (!result.ok) {
        setError(extractError(result.payload));
        return;
      }
      const nextJob = result.payload.data;
      updateJob(nextJob);
      setStatus(`Job ${nextJob.status}`);
      if (terminalReferencePackJobStates.has(nextJob.status)) {
        await loadPacks();
      }
    },
    [loadPacks, session.is_deployment_admin, updateJob],
  );

  useEffect(() => {
    if (
      !session.is_deployment_admin ||
      job === null ||
      terminalReferencePackJobStates.has(job.status)
    ) {
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
  }, [job, loadJob, session.is_deployment_admin]);

  useEffect(() => {
    if (!session.is_deployment_admin) {
      requestSeqRef.current += 1;
      initialLoadRef.current = false;
      acceptedPackQueryRef.current = {
        active: "",
        packVersionState: "",
        search: "",
        verificationResult: "",
      };
      setPacks([]);
      setSelectedKeys(new Set());
      setPackPaging(defaultReferencePackPaging);
      setListStatus("idle");
      setError(null);
      updateJob(null);
      return;
    }
    if (initialLoadRef.current) {
      return;
    }
    initialLoadRef.current = true;
    void loadPacks();
  }, [loadPacks, session.is_deployment_admin, updateJob]);

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
    if (!session.is_deployment_admin) {
      return;
    }
    if (file === null) {
      setStatus("Select a bundle first");
      return;
    }
    const result = await importReferencePackBundle(file);
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus("Import failed to start");
      return;
    }
    const nextJob = result.payload.data;
    setError(null);
    updateJob(nextJob);
    setStatus("Import queued");
    setImportDialogOpen(false);
    setFile(null);
    void loadJob(nextJob.job_id);
  }

  async function runPackAction(
    pack: ReferencePackVersion,
    action: ReferencePackAction,
  ) {
    if (!session.is_deployment_admin) {
      return;
    }
    if (action === "reverify") {
      const result = await reverifyReferencePackVersion(pack);
      if (!result.ok) {
        setError(extractError(result.payload));
        setStatus("reverify failed");
        return;
      }
      const nextJob = result.payload.data;
      setError(null);
      updateJob(nextJob);
      setStatus("Reverify queued");
      void loadJob(nextJob.job_id);
      return;
    }
    const result =
      action === "activate"
        ? await activateReferencePackVersion(pack)
        : await disableReferencePackVersion(pack);
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus(`${action} failed`);
      return;
    }
    setError(null);
    setStatus(`${action} complete`);
    await loadPacks();
  }

  async function refreshSelected(all: boolean) {
    if (!session.is_deployment_admin) {
      return;
    }
    if (!all && selectedKeys.size === 0) {
      setStatus("Select packs first");
      return;
    }
    const packKeys = all ? [] : Array.from(selectedKeys).sort();
    const result = await refreshReferencePacks({ all, packKeys });
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus("Refresh failed to start");
      return;
    }
    const nextJob = result.payload.data;
    setError(null);
    updateJob(nextJob);
    setStatus("Refresh queued");
    void loadJob(nextJob.job_id);
  }

  async function cancelJob() {
    if (!session.is_deployment_admin || job === null) {
      return;
    }
    const result = await cancelReferencePackJob(job.job_id);
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatus("Cancel rejected");
      return;
    }
    const nextJob = result.payload.data;
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

      <div
        aria-live="polite"
        data-testid={referencePackListStatusTestId()}
        role="status"
        style={jobStyle}
      >
        {referencePackListStatusMessage(listStatus)}
      </div>

      <div style={jobStyle} data-testid={referencePackJobStatusTestId()}>
        <span>{status}</span>
        {job !== null ? (
          <span>
            {job.status} · {job.progress.completed}/{job.progress.total ?? "?"}
          </span>
        ) : null}
        {job !== null &&
        !terminalReferencePackJobStates.has(job.status) &&
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

function referencePackListStatusMessage(
  status: ReferencePackListStatus,
): string {
  switch (status) {
    case "searching":
      return "Searching reference packs";
    case "loaded":
      return "Reference packs loaded";
    case "unavailable":
      return "Reference packs unavailable";
    default:
      return "Reference packs idle";
  }
}

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
