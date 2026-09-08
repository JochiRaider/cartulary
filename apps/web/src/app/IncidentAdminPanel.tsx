import {
  incidentAdministrationTestId,
  incidentControlsStatusTestId,
  incidentControlsSurfaceTestId,
} from "@cartulary/ui-contracts";
import { type ReactNode, useEffect, useRef, useState } from "react";
import {
  type IncidentResource,
  validIncidentResource,
} from "../shared/incidentResource";
import { observeAccountOperation } from "./accountOperation";
import { readIncidentSummary } from "./api/incidentResourceClient";
import type {
  IncidentControlsLoadState,
  IncidentControlsSection,
} from "./landingAdminTypes";

type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";

type IncidentSummary = IncidentResource;

type IncidentAdminPanelProps = {
  acceptedIncident?: IncidentSummary | null | undefined;
  onIncidentObserved?: ((incident: IncidentSummary) => void) | undefined;
  incidentId: string;
  currentIncidentRole: IncidentRole | null;
  activeSection?: IncidentControlsSection | undefined;
  apiBase?: string | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  onSessionRoleChange?: (() => Promise<void> | void) | undefined;
  lifecycleControls?: ReactNode;
  preferenceControls?: ReactNode;
};

function displayValue(value: string | null | undefined): string {
  return value && value.trim() !== "" ? value : "Unset";
}

export function IncidentAdminPanel({
  acceptedIncident,
  onIncidentObserved,
  incidentId,
  currentIncidentRole,
  activeSection = "summary",
  apiBase,
  onSessionRoleChange,
  lifecycleControls,
  preferenceControls,
}: IncidentAdminPanelProps) {
  const [observedIncident, setObservedIncident] =
    useState<IncidentSummary | null>(null);
  const incident =
    acceptedIncident && validIncidentResource(acceptedIncident, incidentId)
      ? acceptedIncident
      : observedIncident?.incident_id === incidentId
        ? observedIncident
        : null;
  const [surfaceLoadState, setSurfaceLoadState] =
    useState<IncidentControlsLoadState>("loading");
  const [surfaceStatusText, setSurfaceStatusText] = useState(
    "Loading incident controls…",
  );
  const [error, setError] = useState<{ code: string } | null>(null);
  const observedRef = useRef(onIncidentObserved);
  observedRef.current = onIncidentObserved;
  const accessCheckRef = useRef(onSessionRoleChange);
  accessCheckRef.current = onSessionRoleChange;
  const generation = useRef(0);
  useEffect(() => {
    if (activeSection !== "summary") return;
    const request = ++generation.current;
    const current = () => request === generation.current;
    setSurfaceLoadState("loading");
    setSurfaceStatusText("Loading incident controls…");
    setError(null);
    const observation = observeAccountOperation(async (signal) => {
      const result = await readIncidentSummary(incidentId, apiBase, signal);
      if (!current()) return;
      if (!result.ok) {
        setSurfaceLoadState("unavailable");
        setSurfaceStatusText("Incident controls unavailable.");
        setError({ code: result.code });
        if (
          result.code === "incident_not_found" ||
          result.code === "authorization_denied"
        )
          void accessCheckRef.current?.();
        return;
      }
      observedRef.current?.(result.resource);
      setObservedIncident(result.resource);
      setSurfaceLoadState("synced");
      setSurfaceStatusText("Incident controls synced.");
    });
    void observation.result.then((outcome) => {
      if (!current() || outcome.kind === "completed") return;
      ++generation.current;
      setSurfaceLoadState("unavailable");
      setSurfaceStatusText("Incident controls unavailable.");
      setError({ code: "incident_summary_unavailable" });
    });
    return () => {
      ++generation.current;
      observation.cancel();
    };
  }, [activeSection, apiBase, incidentId]);

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
            lifecycleControls,
            preferenceControls,
            incident,
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
      "Inspect incident summary and manage workbook startup preferences.",
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
  lifecycleControls,
  preferenceControls,
  incident,
}: {
  readonly currentIncidentRole: IncidentRole | null;
  readonly lifecycleControls: ReactNode;
  readonly preferenceControls: ReactNode;
  readonly incident: IncidentSummary | null;
}) {
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
              {currentIncidentRole || "Current access unresolved"}
            </dd>
          </div>
        </dl>
      </section>

      {lifecycleControls}

      {preferenceControls}
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
