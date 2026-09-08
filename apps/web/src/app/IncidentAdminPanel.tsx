import {
  incidentAdministrationTestId,
  incidentControlsActionMessageTestId,
  incidentControlsStatusTestId,
  incidentControlsSurfaceTestId,
} from "@cartulary/ui-contracts";
import { getViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useRef, useState } from "react";
import {
  type APIError,
  clientTxnID,
  extractError,
  fetchHTTPOperation,
  fetchJSON,
} from "../services/browserApi";
import type { SheetRef } from "../shared/sheetRef";
import { isSheetRef } from "../shared/sheetRef";
import { useTransientMessageController } from "../shared/useTransientMessageController";
import type {
  CloseIncidentRequest,
  CloseIncidentResponse,
  ReopenIncidentRequest,
  ReopenIncidentResponse,
} from "./api/publicHttpTypes";
import type {
  IncidentControlsLoadState,
  IncidentControlsSection,
} from "./landingAdminTypes";

type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";

type IncidentSummary = CloseIncidentResponse["data"];

type WorkbookPreferences = {
  default_sheet_ref?: SheetRef | null;
  home_sheet_ref?: SheetRef | null;
};

type PreferenceSlot = {
  readonly sheetRef: SheetRef | null;
  readonly status: "loading" | "loaded" | "unavailable";
};

type WorkbookPreferenceField = "default_sheet_ref" | "home_sheet_ref";

type IncidentAdminPanelProps = {
  acceptedIncident?: IncidentSummary | null | undefined;
  onIncidentObserved?: ((incident: IncidentSummary) => void) | undefined;
  incidentId: string;
  currentIncidentRole: IncidentRole | null;
  activeSection?: IncidentControlsSection | undefined;
  apiBase?: string | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  onSessionRoleChange?: (() => Promise<void> | void) | undefined;
};

type IncidentSurfaceLoadTarget = {
  readonly activeSection: IncidentControlsSection;
  readonly apiBase: string | undefined;
  readonly incidentId: string;
};

type IncidentActionMessage = {
  readonly incidentId: string;
  readonly text: string;
  readonly transient: boolean;
};

function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

function displayValue(value: string | null | undefined): string {
  return value && value.trim() !== "" ? value : "Unset";
}

function unavailablePreferenceSlot(): PreferenceSlot {
  return { sheetRef: null, status: "unavailable" };
}

function loadingPreferenceSlot(): PreferenceSlot {
  return { sheetRef: null, status: "loading" };
}

function loadedPreferenceSlot(sheetRef: SheetRef | null): PreferenceSlot {
  return { sheetRef, status: "loaded" };
}

function preferenceSlotFromPayload(
  payload: unknown,
  field: WorkbookPreferenceField,
): PreferenceSlot {
  if (!payload || typeof payload !== "object" || Array.isArray(payload)) {
    return unavailablePreferenceSlot();
  }
  const data = (payload as { readonly data?: unknown }).data;
  if (!data || typeof data !== "object" || Array.isArray(data)) {
    return unavailablePreferenceSlot();
  }
  const record = data as Record<string, unknown>;
  if (!Object.hasOwn(record, field)) {
    return unavailablePreferenceSlot();
  }
  const value = record[field];
  if (value === null) {
    return loadedPreferenceSlot(null);
  }
  if (isSheetRef(value)) {
    return loadedPreferenceSlot({ ...value });
  }
  return unavailablePreferenceSlot();
}

function formatSheetRef(slot: PreferenceSlot): string {
  if (slot.status === "loading") {
    return "Loading…";
  }
  if (slot.status === "unavailable") {
    return "Unavailable";
  }
  const sheetRef = slot.sheetRef;
  if (sheetRef === null) {
    return "Unset";
  }
  if (sheetRef.kind === "view_schema") {
    const contract = getViewContract(sheetRef.id);
    const label = contract?.title ?? sheetRef.id;
    return `View schema: ${label} (${sheetRef.id})`;
  }
  if (sheetRef.kind === "saved_view") {
    return `Saved view: ${sheetRef.id}`;
  }
  return `Extension workspace: ${sheetRef.extension_profile_id}/${sheetRef.workspace_key}`;
}

export function IncidentAdminPanel({
  acceptedIncident,
  onIncidentObserved,
  incidentId,
  currentIncidentRole,
  activeSection = "summary",
  apiBase,
  onIncidentAccessLost,
  onSessionRoleChange,
}: IncidentAdminPanelProps) {
  const [incident, setIncident] = useState<IncidentSummary | null>(null);
  const [defaultPreference, setDefaultPreference] = useState<PreferenceSlot>(
    loadingPreferenceSlot,
  );
  const [userPreference, setUserPreference] = useState<PreferenceSlot>(
    loadingPreferenceSlot,
  );
  const [lifecycleReason, setLifecycleReason] = useState("");
  const [surfaceLoadState, setSurfaceLoadState] =
    useState<IncidentControlsLoadState>("loading");
  const [surfaceStatusText, setSurfaceStatusText] = useState(
    "Loading incident controls…",
  );
  const [actionMessageState, setActionMessageState] =
    useState<IncidentActionMessage>({ incidentId, text: "", transient: false });
  const [error, setError] = useState<APIError | null>(null);
  const acceptedIncidentRef = useRef(acceptedIncident);
  acceptedIncidentRef.current = acceptedIncident;
  useEffect(() => {
    if (acceptedIncident?.incident_id === incidentId)
      setIncident((current) =>
        current?.incident_id === incidentId &&
        current.incident_version > acceptedIncident.incident_version
          ? current
          : acceptedIncident,
      );
  }, [acceptedIncident, incidentId]);
  const loadRequestIdRef = useRef(0);
  const activeSectionRef = useRef(activeSection);
  const apiBaseRef = useRef(apiBase);
  const incidentIdRef = useRef(incidentId);

  activeSectionRef.current = activeSection;
  apiBaseRef.current = apiBase;
  incidentIdRef.current = incidentId;
  const actionMessage =
    actionMessageState.incidentId === incidentId ? actionMessageState.text : "";

  function setActionMessageForIncident(
    messageIncidentId: string,
    text: string,
    transient = false,
  ) {
    if (messageIncidentId !== incidentIdRef.current) {
      return;
    }
    setActionMessageState({ incidentId: messageIncidentId, text, transient });
  }

  const transientActionMessage = useTransientMessageController({
    actionAvailable: false,
    enabled:
      actionMessageState.incidentId === incidentId &&
      actionMessageState.text !== "" &&
      actionMessageState.transient,
    messageKey: `${actionMessageState.incidentId}:${actionMessageState.text}`,
    onDismiss: () =>
      setActionMessageState((current) =>
        current.incidentId === incidentIdRef.current && current.transient
          ? { ...current, text: "", transient: false }
          : current,
      ),
  });

  const refreshSessionRole = useCallback(async () => {
    await onSessionRoleChange?.();
  }, [onSessionRoleChange]);

  const loadIncidentSurface = useCallback(
    async (target?: IncidentSurfaceLoadTarget) => {
      const requestId = loadRequestIdRef.current + 1;
      loadRequestIdRef.current = requestId;
      const requestedSection =
        target?.activeSection ?? activeSectionRef.current;
      const requestedApiBase = target?.apiBase ?? apiBaseRef.current;
      const requestedIncidentId = target?.incidentId ?? incidentIdRef.current;
      const isLatestRequest = () => loadRequestIdRef.current === requestId;

      setSurfaceLoadState("loading");
      setSurfaceStatusText("Loading incident controls…");
      if (requestedSection === "summary") {
        setDefaultPreference(loadingPreferenceSlot());
        setUserPreference(loadingPreferenceSlot());
      }

      const incidentRequest = fetchJSON<{ data: IncidentSummary }>(
        apiPath(requestedApiBase, `/api/v1/incidents/${requestedIncidentId}`),
      );
      const defaultPrefsRequest =
        requestedSection === "summary"
          ? fetchJSON<{ data: WorkbookPreferences }>(
              apiPath(
                requestedApiBase,
                `/api/v1/incidents/${requestedIncidentId}/workbook-preferences/default`,
              ),
            )
          : Promise.resolve(null);
      const userPrefsRequest =
        requestedSection === "summary"
          ? fetchJSON<{ data: WorkbookPreferences }>(
              apiPath(
                requestedApiBase,
                `/api/v1/incidents/${requestedIncidentId}/workbook-preferences/me`,
              ),
            )
          : Promise.resolve(null);

      const [incidentResult, defaultPrefsResult, userPrefsResult] =
        await Promise.all([
          incidentRequest,
          defaultPrefsRequest,
          userPrefsRequest,
        ]);

      if (!isLatestRequest()) {
        return;
      }

      if (!incidentResult.ok) {
        const incidentError = extractError(incidentResult.payload);
        setError(incidentError);
        setIncident(null);
        setDefaultPreference(unavailablePreferenceSlot());
        setUserPreference(unavailablePreferenceSlot());
        setSurfaceLoadState("unavailable");
        setSurfaceStatusText("Incident controls unavailable.");
        if (
          incidentError?.code === "incident_not_found" ||
          incidentError?.code === "authorization_denied"
        ) {
          onIncidentAccessLost?.();
        }
        return;
      }

      const nextIncident = (incidentResult.payload as { data: IncidentSummary })
        .data;
      const published = acceptedIncidentRef.current;
      setIncident(
        published?.incident_id === requestedIncidentId &&
          published.incident_version > nextIncident.incident_version
          ? published
          : nextIncident,
      );

      let partialFailure = false;

      if (requestedSection === "summary") {
        const nextDefaultPreference = defaultPrefsResult?.ok
          ? preferenceSlotFromPayload(
              defaultPrefsResult.payload,
              "default_sheet_ref",
            )
          : unavailablePreferenceSlot();
        const nextUserPreference = userPrefsResult?.ok
          ? preferenceSlotFromPayload(userPrefsResult.payload, "home_sheet_ref")
          : unavailablePreferenceSlot();
        setDefaultPreference(nextDefaultPreference);
        setUserPreference(nextUserPreference);
        partialFailure =
          nextDefaultPreference.status === "unavailable" ||
          nextUserPreference.status === "unavailable";
      }

      setError(null);
      setSurfaceLoadState(partialFailure ? "partial" : "synced");
      setSurfaceStatusText(
        partialFailure
          ? "Incident summary synced; workbook preferences unavailable."
          : "Incident controls synced.",
      );
    },
    [onIncidentAccessLost],
  );

  useEffect(() => {
    void loadIncidentSurface({ activeSection, apiBase, incidentId });
  }, [activeSection, apiBase, incidentId, loadIncidentSurface]);

  useEffect(() => {
    setActionMessageState((current) =>
      current.incidentId === incidentId
        ? current
        : { incidentId, text: "", transient: false },
    );
  }, [incidentId]);

  async function handleLifecycle(action: "close" | "reopen") {
    if (!incident || currentIncidentRole !== "admin") {
      return;
    }
    const actionIncidentId = incident.incident_id;
    if (lifecycleReason.trim() === "") {
      setActionMessageForIncident(
        actionIncidentId,
        "Lifecycle reason is required.",
      );
      return;
    }
    setActionMessageForIncident(
      actionIncidentId,
      action === "close" ? "Closing incident…" : "Reopening incident…",
    );
    const operationID = action === "close" ? "closeIncident" : "reopenIncident";
    const request: CloseIncidentRequest | ReopenIncidentRequest = {
      base_incident_version: incident.incident_version,
      client_txn_id: clientTxnID(`incident-${action}`),
      reason: lifecycleReason.trim(),
    };
    const result = await fetchHTTPOperation<
      CloseIncidentResponse | ReopenIncidentResponse
    >({
      apiBase,
      operationID,
      pathParameters: {
        incident_id: incident.incident_id,
      },
      init: {
        method: "POST",
        body: JSON.stringify(request),
      },
    });
    if (!result.ok) {
      setError(extractError(result.payload));
      if (result.status === 409) {
        await loadIncidentSurface();
      }
      setActionMessageForIncident(
        actionIncidentId,
        result.status === 409
          ? "Incident changed; refreshed current state. Review and retry."
          : action === "close"
            ? "Incident close failed."
            : "Incident reopen failed.",
      );
      return;
    }
    const nextIncident = (
      result.payload as CloseIncidentResponse | ReopenIncidentResponse
    ).data;
    setError(null);
    setIncident(nextIncident);
    onIncidentObserved?.(nextIncident);
    setLifecycleReason("");
    await refreshSessionRole();
    setActionMessageForIncident(
      actionIncidentId,
      action === "close" ? "Incident closed." : "Incident reopened.",
      true,
    );
  }

  const activeSectionMeta = incidentControlsSectionMeta[activeSection];

  return (
    <section
      aria-busy={surfaceLoadState === "loading"}
      data-incident-controls-load-state={surfaceLoadState}
      data-incident-controls-section={activeSection}
      data-testid={incidentControlsSurfaceTestId()}
      style={panelStyle}
    >
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Incident shell</p>
          <h2 style={titleStyle}>{activeSectionMeta.title}</h2>
          <p style={bodyStyle}>{activeSectionMeta.description}</p>
        </div>
        <div style={statusCardStyle}>
          <span style={labelStyle}>Status</span>
          <strong
            aria-live="polite"
            data-testid={incidentControlsStatusTestId()}
            role="status"
          >
            {surfaceStatusText}
          </strong>
        </div>
      </div>

      <p
        {...transientActionMessage}
        aria-live="polite"
        data-testid={incidentControlsActionMessageTestId()}
        role="status"
        style={actionMessageStyle}
      >
        {actionMessage}
      </p>

      {error ? (
        <p
          aria-live="assertive"
          data-testid={incidentAdministrationTestId("admin-error-code")}
          role="alert"
          style={errorStyle}
        >
          {error.code}
        </p>
      ) : (
        <p
          data-testid={incidentAdministrationTestId("admin-error-code")}
          style={errorStyle}
        >
          {""}
        </p>
      )}

      {activeSection === "summary" ? (
        <div style={gridStyle}>
          {renderIncidentSummary({
            currentIncidentRole,
            defaultPreference,
            lifecycleReason,
            incident,
            onLifecycleReasonChange: setLifecycleReason,
            onClose: () => handleLifecycle("close"),
            onReopen: () => handleLifecycle("reopen"),
            userPreference,
          })}
        </div>
      ) : null}
    </section>
  );
}

const incidentControlsSectionMeta = {
  "import-assistant": {
    title: "Import workbook",
    description:
      "Discover, map, select, and apply structured workbook source data.",
  },
  summary: {
    title: "Summary and preferences",
    description:
      "Read incident summary fields and workbook bootstrap defaults.",
  },
  "incident-fields": {
    title: "Promoted fields",
    description: "Update promoted incident fields when your role allows edits.",
  },
  memberships: {
    title: "Memberships",
    description: "Review incident membership roles and manage access.",
  },
  "membership-audit": {
    title: "Membership audit",
    description:
      "Review incident-scoped membership changes with exact actor and target filters.",
  },
} satisfies Record<
  IncidentControlsSection,
  { readonly description: string; readonly title: string }
>;

function renderIncidentSummary({
  currentIncidentRole,
  defaultPreference,
  lifecycleReason,
  incident,
  onClose,
  onLifecycleReasonChange,
  onReopen,
  userPreference,
}: {
  readonly currentIncidentRole: IncidentRole | null;
  readonly defaultPreference: PreferenceSlot;
  readonly lifecycleReason: string;
  readonly incident: IncidentSummary | null;
  readonly onClose: () => Promise<void> | void;
  readonly onLifecycleReasonChange: (value: string) => void;
  readonly onReopen: () => Promise<void> | void;
  readonly userPreference: PreferenceSlot;
}) {
  const canLifecycle = currentIncidentRole === "admin" && incident !== null;
  const lifecycleReasonReady = lifecycleReason.trim() !== "";
  return (
    <>
      <section style={cardStyle}>
        <div style={cardHeaderStyle}>
          <div>
            <p style={cardEyebrowStyle}>Direct retrieval</p>
            <h3 style={cardTitleStyle}>Incident summary</h3>
          </div>
          <span
            data-testid={incidentAdministrationTestId("summary-version")}
            style={versionBadgeStyle}
          >
            Version {incident?.incident_version ?? "?"}
          </span>
        </div>

        <dl style={definitionGridStyle}>
          <div>
            <dt style={labelStyle}>Incident key</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-key")}
              style={valueStyle}
            >
              {incident?.incident_key ?? "Loading…"}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Title</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-title")}
              style={valueStyle}
            >
              {incident?.title ?? "Loading…"}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Status</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-status")}
              style={valueStyle}
            >
              {incident?.status === "closed"
                ? "Closed, read-only"
                : (incident?.status ?? "Loading…")}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Description</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-description")}
              style={valueStyle}
            >
              {displayValue(incident?.description)}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Severity</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-severity")}
              style={valueStyle}
            >
              {displayValue(incident?.severity)}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Closed at</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-closed-at")}
              style={valueStyle}
            >
              {incident?.closed_at ?? "Unset"}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>TLP</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-tlp")}
              style={valueStyle}
            >
              {displayValue(incident?.tlp)}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Current phase</dt>
            <dd
              data-testid={incidentAdministrationTestId(
                "summary-current-phase",
              )}
              style={valueStyle}
            >
              {displayValue(incident?.current_phase)}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Primary external case</dt>
            <dd
              data-testid={incidentAdministrationTestId(
                "summary-primary-external-case-ref",
              )}
              style={valueStyle}
            >
              {displayValue(incident?.primary_external_case_ref)}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>Current role</dt>
            <dd
              data-testid={incidentAdministrationTestId("summary-role")}
              style={valueStyle}
            >
              {currentIncidentRole || "viewer"}
            </dd>
          </div>
        </dl>
      </section>

      <section style={cardStyle}>
        <div style={cardHeaderStyle}>
          <div>
            <p style={cardEyebrowStyle}>Lifecycle</p>
            <h3 style={cardTitleStyle}>Close or reopen</h3>
          </div>
        </div>
        <div style={inlineFormStyle}>
          <label style={fieldLabelStyle}>
            Reason
            <input
              data-testid={incidentAdministrationTestId("lifecycle-reason")}
              style={inputStyle}
              value={lifecycleReason}
              onChange={(event) => {
                onLifecycleReasonChange(event.target.value);
              }}
            />
          </label>
          <button
            data-testid={incidentAdministrationTestId("close-button")}
            disabled={
              !canLifecycle ||
              !lifecycleReasonReady ||
              incident?.status !== "active"
            }
            style={primaryButtonStyle}
            type="button"
            onClick={() => {
              void onClose();
            }}
          >
            Close incident
          </button>
          <button
            data-testid={incidentAdministrationTestId("reopen-button")}
            disabled={
              !canLifecycle ||
              !lifecycleReasonReady ||
              incident?.status !== "closed"
            }
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void onReopen();
            }}
          >
            Reopen incident
          </button>
        </div>
      </section>

      <section style={cardStyle}>
        <div style={cardHeaderStyle}>
          <div>
            <p style={cardEyebrowStyle}>Workbook preferences</p>
            <h3 style={cardTitleStyle}>Bootstrap defaults</h3>
          </div>
        </div>

        <dl style={definitionGridStyle}>
          <div>
            <dt style={labelStyle}>Incident default sheet</dt>
            <dd
              data-testid={incidentAdministrationTestId(
                "pref-default-sheet-ref",
              )}
              style={valueStyle}
            >
              {formatSheetRef(defaultPreference)}
            </dd>
          </div>
          <div>
            <dt style={labelStyle}>My home sheet</dt>
            <dd
              data-testid={incidentAdministrationTestId("pref-home-sheet-ref")}
              style={valueStyle}
            >
              {formatSheetRef(userPreference)}
            </dd>
          </div>
        </dl>
      </section>
    </>
  );
}

const panelStyle = {
  marginBottom: "1.5rem",
  display: "grid",
  gap: "1rem",
};

const headerStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "1rem",
  alignItems: "start",
  flexWrap: "wrap" as const,
};

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.76rem",
  letterSpacing: "0.18em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

const titleStyle = {
  margin: "0.35rem 0 0.4rem",
  fontSize: "1.5rem",
  lineHeight: 1.15,
};

const bodyStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-muted)",
  maxWidth: "42rem",
};

const actionMessageStyle = {
  margin: 0,
  minHeight: "1.25rem",
  color: "var(--ct-colors-ink-muted)",
};

const errorStyle = {
  margin: 0,
  color: "var(--ct-colors-semantic-conflict)",
  minHeight: "1.25rem",
};

const statusCardStyle = {
  minWidth: "14rem",
  padding: "0.9rem 1rem",
  borderRadius: "var(--ct-rounded-lg)",
  background: "var(--ct-colors-surface-2)",
  border: "var(--ct-border-hairline)",
  display: "grid",
  gap: "0.35rem",
};

const labelStyle = {
  margin: 0,
  fontSize: "0.8rem",
  fontWeight: 600,
  color: "var(--ct-colors-ink-muted)",
};

const gridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(20rem, 1fr))",
  gap: "1rem",
};

const cardStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  padding: "1rem",
  display: "grid",
  gap: "0.85rem",
};

const cardHeaderStyle = {
  display: "flex",
  justifyContent: "space-between",
  gap: "0.75rem",
  alignItems: "start",
};

const cardEyebrowStyle = {
  margin: 0,
  fontSize: "0.74rem",
  letterSpacing: "0.16em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-ink-subtle)",
};

const cardTitleStyle = {
  margin: "0.3rem 0 0",
  fontSize: "1.05rem",
};

const versionBadgeStyle = {
  alignSelf: "start",
  borderRadius: "var(--ct-rounded-pill)",
  padding: "0.35rem 0.7rem",
  background: "var(--ct-colors-surface-3)",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 600,
};

const definitionGridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.85rem",
  margin: 0,
};

const valueStyle = {
  margin: "0.25rem 0 0",
  color: "var(--ct-colors-ink)",
  wordBreak: "break-word" as const,
};

const inlineFormStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(12rem, 1fr))",
  gap: "0.85rem",
  alignItems: "end",
};

const fieldLabelStyle = {
  display: "grid",
  gap: "0.35rem",
  color: "var(--ct-colors-ink-muted)",
  fontWeight: 600,
  fontSize: "0.88rem",
};

const inputStyle = {
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  padding: "var(--ct-component-text-input-padding)",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  background: "var(--ct-component-text-input-backgroundColor)",
};

const primaryButtonStyle = {
  borderRadius: "var(--ct-component-button-primary-rounded)",
  border: "none",
  padding: "var(--ct-component-button-primary-padding)",
  background: "var(--ct-component-button-primary-backgroundColor)",
  color: "var(--ct-component-button-primary-textColor)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
};

const secondaryButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  padding: "var(--ct-component-button-secondary-padding)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
};
