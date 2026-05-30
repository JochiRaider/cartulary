import {
  incidentMembershipAdminNoteTestId,
  incidentMembershipCreateButtonTestId,
  incidentMembershipDeleteButtonTestId,
  incidentMembershipEmailInputTestId,
  incidentMembershipListTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleDisplayTestId,
  incidentMembershipRoleInputTestId,
  incidentMembershipRoleSelectTestId,
  incidentMembershipRowTestId,
  incidentMembershipVersionTestId,
} from "@cartulary/ui-contracts";
import { useCallback, useEffect, useState } from "react";

import {
  type APIError,
  clientTxnID,
  extractError,
  fetchJSON,
} from "./browserApi";

type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";
type MembershipRole = Exclude<IncidentRole, "">;

type IncidentSummary = {
  incident_id: string;
  incident_key: string;
  title: string;
  description: string | null;
  severity: string | null;
  tlp: string | null;
  current_phase: string | null;
  primary_external_case_ref: string | null;
  incident_version: number;
};

type MembershipRecord = {
  incident_id: string;
  user_id: string;
  display_name: string;
  role: MembershipRole;
  membership_version: number;
};

type WorkbookPreferences = {
  default_sheet_ref?: string | null;
  home_sheet_ref?: string | null;
};

type IncidentAdminPanelProps = {
  incidentId: string;
  currentIncidentRole: IncidentRole | null;
  apiBase?: string | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  onIncidentSnapshot?: ((incident: IncidentSummary) => void) | undefined;
  onSessionRoleChange?: (() => Promise<void> | void) | undefined;
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

function upsertMembershipRoleDrafts(records: MembershipRecord[]) {
  return Object.fromEntries(
    records.map((record) => [record.user_id, record.role]),
  ) as Record<string, MembershipRole>;
}

export function IncidentAdminPanel({
  incidentId,
  currentIncidentRole,
  apiBase,
  onIncidentAccessLost,
  onIncidentSnapshot,
  onSessionRoleChange,
}: IncidentAdminPanelProps) {
  const [incident, setIncident] = useState<IncidentSummary | null>(null);
  const [memberships, setMemberships] = useState<MembershipRecord[]>([]);
  const [defaultPrefs, setDefaultPrefs] = useState<WorkbookPreferences | null>(
    null,
  );
  const [userPrefs, setUserPrefs] = useState<WorkbookPreferences | null>(null);
  const [patchTLP, setPatchTLP] = useState("");
  const [patchCurrentPhase, setPatchCurrentPhase] = useState("");
  const [patchExternalCase, setPatchExternalCase] = useState("");
  const [membershipEmail, setMembershipEmail] = useState("");
  const [membershipRole, setMembershipRole] =
    useState<MembershipRole>("viewer");
  const [membershipRoleDrafts, setMembershipRoleDrafts] = useState<
    Record<string, MembershipRole>
  >({});
  const [statusText, setStatusText] = useState("Loading incident controls…");
  const [error, setError] = useState<APIError | null>(null);

  const canEditIncident =
    currentIncidentRole === "reviewer" || currentIncidentRole === "admin";
  const canManageMemberships = currentIncidentRole === "admin";

  const refreshSessionRole = useCallback(async () => {
    await onSessionRoleChange?.();
  }, [onSessionRoleChange]);

  const loadIncidentSurface = useCallback(async () => {
    const [
      incidentResult,
      membershipsResult,
      defaultPrefsResult,
      userPrefsResult,
    ] = await Promise.all([
      fetchJSON<{ data: IncidentSummary }>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}`),
      ),
      fetchJSON<{ data: { memberships: MembershipRecord[] } }>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}/memberships`),
      ),
      fetchJSON<{ data: WorkbookPreferences }>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/workbook-preferences/default`,
        ),
      ),
      fetchJSON<{ data: WorkbookPreferences }>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/workbook-preferences/me`,
        ),
      ),
    ]);

    if (!incidentResult.ok) {
      const incidentError = extractError(incidentResult.payload);
      setError(incidentError);
      setIncident(null);
      setMemberships([]);
      setDefaultPrefs(null);
      setUserPrefs(null);
      setStatusText("Incident controls unavailable.");
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
    setIncident(nextIncident);
    onIncidentSnapshot?.(nextIncident);
    setPatchTLP(nextIncident.tlp ?? "");
    setPatchCurrentPhase(nextIncident.current_phase ?? "");
    setPatchExternalCase(nextIncident.primary_external_case_ref ?? "");

    if (membershipsResult.ok) {
      const nextMemberships = (
        membershipsResult.payload as {
          data: { memberships: MembershipRecord[] };
        }
      ).data.memberships;
      setMemberships(nextMemberships);
      setMembershipRoleDrafts(upsertMembershipRoleDrafts(nextMemberships));
    } else {
      setMemberships([]);
      setMembershipRoleDrafts({});
    }

    setDefaultPrefs(
      defaultPrefsResult.ok
        ? (defaultPrefsResult.payload as { data: WorkbookPreferences }).data
        : null,
    );
    setUserPrefs(
      userPrefsResult.ok
        ? (userPrefsResult.payload as { data: WorkbookPreferences }).data
        : null,
    );
    setError(null);
    setStatusText("Incident controls synced.");
  }, [apiBase, incidentId, onIncidentAccessLost, onIncidentSnapshot]);

  useEffect(() => {
    void loadIncidentSurface();
  }, [loadIncidentSurface]);

  async function handlePatchIncident() {
    if (!incident) {
      return;
    }

    setStatusText("Saving promoted incident fields…");
    const result = await fetchJSON<{ data: IncidentSummary }>(
      apiPath(apiBase, `/api/v1/incidents/${incident.incident_id}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          base_incident_version: incident.incident_version,
          tlp: patchTLP === "" ? null : patchTLP,
          current_phase: patchCurrentPhase === "" ? null : patchCurrentPhase,
          primary_external_case_ref:
            patchExternalCase === "" ? null : patchExternalCase,
        }),
      },
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatusText("Incident update failed.");
      return;
    }

    setError(null);
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setStatusText("Saved promoted incident fields.");
  }

  async function handleCreateMembership() {
    if (!incident) {
      return;
    }

    setStatusText("Adding membership…");
    const result = await fetchJSON<{ data: MembershipRecord }>(
      apiPath(apiBase, `/api/v1/incidents/${incident.incident_id}/memberships`),
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: clientTxnID("incident-membership"),
          email: membershipEmail.trim(),
          role: membershipRole,
        }),
      },
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatusText("Membership create failed.");
      return;
    }

    setError(null);
    setMembershipEmail("");
    setMembershipRole("viewer");
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setStatusText("Added membership.");
  }

  async function handlePatchMembership(membership: MembershipRecord) {
    if (!incident) {
      return;
    }

    setStatusText("Updating membership…");
    const result = await fetchJSON<{ data: MembershipRecord }>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incident.incident_id}/memberships/${membership.user_id}`,
      ),
      {
        method: "PATCH",
        body: JSON.stringify({
          base_membership_version: membership.membership_version,
          role: membershipRoleDrafts[membership.user_id] ?? membership.role,
        }),
      },
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setStatusText("Membership update failed.");
      return;
    }

    setError(null);
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setStatusText("Updated membership.");
  }

  async function handleDeleteMembership(membership: MembershipRecord) {
    if (!incident) {
      return;
    }

    setStatusText("Removing membership…");
    const result = await fetchJSON(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incident.incident_id}/memberships/${membership.user_id}`,
      ),
      {
        method: "DELETE",
        body: JSON.stringify({
          base_membership_version: membership.membership_version,
        }),
      },
    );
    if (!result.ok && result.status !== 204) {
      setError(extractError(result.payload));
      setStatusText("Membership delete failed.");
      return;
    }

    setError(null);
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setStatusText("Removed membership.");
  }

  return (
    <section aria-busy={incident === null} style={panelStyle}>
      <div style={headerStyle}>
        <div>
          <p style={eyebrowStyle}>Incident shell</p>
          <h2 style={titleStyle}>Summary and admin controls</h2>
          <p style={bodyStyle}>
            The ordinary workbook route now carries visible Phase 2 summary,
            preference, patch, and membership behavior.
          </p>
        </div>
        <div style={statusCardStyle}>
          <span style={labelStyle}>Status</span>
          <strong
            aria-live="polite"
            data-testid="incident-admin-status"
            role="status"
          >
            {statusText}
          </strong>
        </div>
      </div>

      {error ? (
        <p
          aria-live="assertive"
          data-testid="incident-admin-error-code"
          role="alert"
          style={errorStyle}
        >
          {error.code}
        </p>
      ) : (
        <p data-testid="incident-admin-error-code" style={errorStyle}>
          {""}
        </p>
      )}

      <div style={gridStyle}>
        <section style={cardStyle}>
          <div style={cardHeaderStyle}>
            <div>
              <p style={cardEyebrowStyle}>Direct retrieval</p>
              <h3 style={cardTitleStyle}>Incident summary</h3>
            </div>
            <span
              data-testid="incident-summary-version"
              style={versionBadgeStyle}
            >
              Version {incident?.incident_version ?? "?"}
            </span>
          </div>

          <dl style={definitionGridStyle}>
            <div>
              <dt style={labelStyle}>Incident key</dt>
              <dd data-testid="incident-summary-key" style={valueStyle}>
                {incident?.incident_key ?? "Loading…"}
              </dd>
            </div>
            <div>
              <dt style={labelStyle}>Title</dt>
              <dd data-testid="incident-summary-title" style={valueStyle}>
                {incident?.title ?? "Loading…"}
              </dd>
            </div>
            <div>
              <dt style={labelStyle}>TLP</dt>
              <dd data-testid="incident-summary-tlp" style={valueStyle}>
                {displayValue(incident?.tlp)}
              </dd>
            </div>
            <div>
              <dt style={labelStyle}>Current phase</dt>
              <dd
                data-testid="incident-summary-current-phase"
                style={valueStyle}
              >
                {displayValue(incident?.current_phase)}
              </dd>
            </div>
            <div>
              <dt style={labelStyle}>Primary external case</dt>
              <dd
                data-testid="incident-summary-primary-external-case-ref"
                style={valueStyle}
              >
                {displayValue(incident?.primary_external_case_ref)}
              </dd>
            </div>
            <div>
              <dt style={labelStyle}>Current role</dt>
              <dd data-testid="incident-summary-role" style={valueStyle}>
                {currentIncidentRole || "viewer"}
              </dd>
            </div>
          </dl>
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
                data-testid="incident-pref-default-sheet-ref"
                style={valueStyle}
              >
                {displayValue(defaultPrefs?.default_sheet_ref)}
              </dd>
            </div>
            <div>
              <dt style={labelStyle}>My home sheet</dt>
              <dd data-testid="incident-pref-home-sheet-ref" style={valueStyle}>
                {displayValue(userPrefs?.home_sheet_ref)}
              </dd>
            </div>
          </dl>
        </section>
      </div>

      <div style={gridStyle}>
        <section style={cardStyle}>
          <div style={cardHeaderStyle}>
            <div>
              <p style={cardEyebrowStyle}>Promoted fields only</p>
              <h3 style={cardTitleStyle}>Incident update</h3>
            </div>
          </div>

          {canEditIncident ? (
            <div style={formGridStyle}>
              <label style={fieldLabelStyle}>
                TLP
                <input
                  data-testid="incident-patch-tlp"
                  style={inputStyle}
                  value={patchTLP}
                  onChange={(event) => {
                    setPatchTLP(event.target.value);
                  }}
                  placeholder="amber"
                />
              </label>
              <label style={fieldLabelStyle}>
                Current phase
                <input
                  data-testid="incident-patch-current-phase"
                  style={inputStyle}
                  value={patchCurrentPhase}
                  onChange={(event) => {
                    setPatchCurrentPhase(event.target.value);
                  }}
                  placeholder="containment"
                />
              </label>
              <label style={fieldLabelStyle}>
                Primary external case
                <input
                  data-testid="incident-patch-external-case"
                  style={inputStyle}
                  value={patchExternalCase}
                  onChange={(event) => {
                    setPatchExternalCase(event.target.value);
                  }}
                  placeholder="CASE-1234"
                />
              </label>
              <button
                data-testid="incident-patch-button"
                style={primaryButtonStyle}
                type="button"
                onClick={() => {
                  void handlePatchIncident();
                }}
              >
                Save promoted fields
              </button>
            </div>
          ) : (
            <p
              data-testid="incident-patch-readonly-note"
              style={mutedBodyStyle}
            >
              Promoted incident fields are read-only for this incident role.
            </p>
          )}
        </section>

        <section style={cardStyle}>
          <div style={cardHeaderStyle}>
            <div>
              <p style={cardEyebrowStyle}>Membership surface</p>
              <h3 style={cardTitleStyle}>Incident memberships</h3>
            </div>
          </div>

          {canManageMemberships ? (
            <div style={inlineFormStyle}>
              <label style={fieldLabelStyle}>
                User email
                <input
                  data-testid={incidentMembershipEmailInputTestId()}
                  style={inputStyle}
                  value={membershipEmail}
                  onChange={(event) => {
                    setMembershipEmail(event.target.value);
                  }}
                  placeholder="analyst@example.test"
                />
              </label>
              <label style={fieldLabelStyle}>
                Role
                <select
                  data-testid={incidentMembershipRoleSelectTestId()}
                  style={inputStyle}
                  value={membershipRole}
                  onChange={(event) => {
                    setMembershipRole(event.target.value as MembershipRole);
                  }}
                >
                  <option value="viewer">viewer</option>
                  <option value="editor">editor</option>
                  <option value="reviewer">reviewer</option>
                  <option value="admin">admin</option>
                </select>
              </label>
              <button
                data-testid={incidentMembershipCreateButtonTestId()}
                style={primaryButtonStyle}
                type="button"
                onClick={() => {
                  void handleCreateMembership();
                }}
              >
                Add membership
              </button>
            </div>
          ) : (
            <p
              data-testid={incidentMembershipAdminNoteTestId()}
              style={mutedBodyStyle}
            >
              Only incident admins can add, change, or remove memberships.
            </p>
          )}

          <div
            data-testid={incidentMembershipListTestId()}
            style={membershipListStyle}
          >
            {memberships.map((membership) => (
              <article
                key={membership.user_id}
                data-testid={incidentMembershipRowTestId(membership.user_id)}
                style={membershipCardStyle}
              >
                <div style={membershipMetaStyle}>
                  <strong>{membership.display_name}</strong>
                  <span style={valueStyle}>{membership.user_id}</span>
                  <span
                    data-testid={incidentMembershipVersionTestId(
                      membership.user_id,
                    )}
                    style={subtleValueStyle}
                  >
                    Version {membership.membership_version}
                  </span>
                </div>

                {canManageMemberships ? (
                  <div style={membershipControlStyle}>
                    <select
                      data-testid={incidentMembershipRoleInputTestId(
                        membership.user_id,
                      )}
                      style={inputStyle}
                      value={
                        membershipRoleDrafts[membership.user_id] ??
                        membership.role
                      }
                      onChange={(event) => {
                        setMembershipRoleDrafts((current) => ({
                          ...current,
                          [membership.user_id]: event.target
                            .value as MembershipRole,
                        }));
                      }}
                    >
                      <option value="viewer">viewer</option>
                      <option value="editor">editor</option>
                      <option value="reviewer">reviewer</option>
                      <option value="admin">admin</option>
                    </select>
                    <button
                      data-testid={incidentMembershipPatchButtonTestId(
                        membership.user_id,
                      )}
                      style={secondaryButtonStyle}
                      type="button"
                      onClick={() => {
                        void handlePatchMembership(membership);
                      }}
                    >
                      Save role
                    </button>
                    <button
                      data-testid={incidentMembershipDeleteButtonTestId(
                        membership.user_id,
                      )}
                      style={dangerButtonStyle}
                      type="button"
                      onClick={() => {
                        void handleDeleteMembership(membership);
                      }}
                    >
                      Remove
                    </button>
                  </div>
                ) : (
                  <p
                    data-testid={incidentMembershipRoleDisplayTestId(
                      membership.user_id,
                    )}
                    style={valueStyle}
                  >
                    {membership.role}
                  </p>
                )}
              </article>
            ))}
          </div>
        </section>
      </div>
    </section>
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
  color: "rgb(55 92 86)",
};

const titleStyle = {
  margin: "0.35rem 0 0.4rem",
  fontSize: "1.5rem",
  lineHeight: 1.15,
};

const bodyStyle = {
  margin: 0,
  color: "rgb(63 83 78)",
  maxWidth: "42rem",
};

const mutedBodyStyle = {
  margin: 0,
  color: "rgb(94 107 103)",
};

const errorStyle = {
  margin: 0,
  color: "rgb(151 53 38)",
  minHeight: "1.25rem",
};

const statusCardStyle = {
  minWidth: "14rem",
  padding: "0.9rem 1rem",
  borderRadius: "1rem",
  background: "rgb(239 245 240)",
  border: "1px solid rgb(194 210 201)",
  display: "grid",
  gap: "0.35rem",
};

const labelStyle = {
  margin: 0,
  fontSize: "0.8rem",
  fontWeight: 600,
  color: "rgb(63 83 78)",
};

const gridStyle = {
  display: "grid",
  gridTemplateColumns: "repeat(auto-fit, minmax(20rem, 1fr))",
  gap: "1rem",
};

const cardStyle = {
  borderRadius: "1.2rem",
  border: "1px solid rgb(204 216 210)",
  background: "rgb(250 248 242)",
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
  color: "rgb(92 108 103)",
};

const cardTitleStyle = {
  margin: "0.3rem 0 0",
  fontSize: "1.05rem",
};

const versionBadgeStyle = {
  alignSelf: "start",
  borderRadius: "999px",
  padding: "0.35rem 0.7rem",
  background: "rgb(233 239 234)",
  color: "rgb(43 69 63)",
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
  color: "rgb(29 47 43)",
  wordBreak: "break-word" as const,
};

const subtleValueStyle = {
  color: "rgb(94 107 103)",
  fontSize: "0.85rem",
};

const formGridStyle = {
  display: "grid",
  gap: "0.85rem",
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
  color: "rgb(43 69 63)",
  fontWeight: 600,
  fontSize: "0.88rem",
};

const inputStyle = {
  borderRadius: "0.8rem",
  border: "1px solid rgb(187 202 195)",
  padding: "0.75rem 0.85rem",
  font: "inherit",
  color: "rgb(24 38 35)",
  background: "rgb(255 255 255)",
};

const primaryButtonStyle = {
  borderRadius: "999px",
  border: "none",
  padding: "0.8rem 1rem",
  background: "rgb(34 84 73)",
  color: "white",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
};

const secondaryButtonStyle = {
  borderRadius: "999px",
  border: "1px solid rgb(134 164 154)",
  padding: "0.75rem 0.95rem",
  background: "rgb(244 248 245)",
  color: "rgb(34 84 73)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
};

const dangerButtonStyle = {
  borderRadius: "999px",
  border: "1px solid rgb(210 176 168)",
  padding: "0.75rem 0.95rem",
  background: "rgb(250 242 239)",
  color: "rgb(138 52 38)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
};

const membershipListStyle = {
  display: "grid",
  gap: "0.75rem",
};

const membershipCardStyle = {
  borderRadius: "1rem",
  border: "1px solid rgb(214 222 217)",
  padding: "0.85rem",
  background: "rgb(255 255 255 / 0.65)",
  display: "grid",
  gap: "0.75rem",
};

const membershipMetaStyle = {
  display: "grid",
  gap: "0.2rem",
};

const membershipControlStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(9rem, 1fr) auto auto",
  gap: "0.65rem",
  alignItems: "center",
};
