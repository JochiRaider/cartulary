import { type CSSProperties, useCallback, useEffect, useState } from "react";

const csrfCookieName = "cartulary_csrf";
const csrfHeaderName = "X-CSRF-Token";

type SessionMembership = {
  incident_id: string;
  role: string;
};

type SessionData = {
  user_id: string;
  display_name: string;
  is_deployment_admin: boolean;
  memberships: SessionMembership[];
};

type IncidentData = {
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

type MembershipData = {
  incident_id: string;
  user_id: string;
  display_name: string;
  role: string;
  membership_version: number;
};

type WorkbookPreferences = {
  default_sheet_ref?: unknown;
  home_sheet_ref?: unknown;
};

type ExtensionProfile = {
  profile_id: string;
  claimed: boolean;
  route_families: string[];
};

type APIError = {
  code: string;
  details?: Record<string, unknown>;
  message?: string;
  request_id?: string;
  status?: number;
};

type ProbeResult = {
  status: number;
  body: unknown;
};

function readCookie(name: string): string | null {
  if (typeof document === "undefined") {
    return null;
  }

  const prefix = `${name}=`;
  for (const segment of document.cookie.split(";")) {
    const trimmed = segment.trim();
    if (trimmed.startsWith(prefix)) {
      return decodeURIComponent(trimmed.slice(prefix.length));
    }
  }
  return null;
}

async function fetchJSON<T>(
  input: RequestInfo | URL,
  init?: RequestInit,
): Promise<{
  ok: boolean;
  status: number;
  payload: T | { error?: APIError };
}> {
  const method = (init?.method ?? "GET").toUpperCase();
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (method !== "GET" && method !== "HEAD" && method !== "OPTIONS") {
    headers["Content-Type"] = "application/json";
    const csrfToken = readCookie(csrfCookieName);
    if (csrfToken !== null && csrfToken !== "") {
      headers[csrfHeaderName] = csrfToken;
    }
  }

  const response = await fetch(input, {
    credentials: "include",
    ...init,
    headers,
  });
  const contentType = response.headers.get("Content-Type") ?? "";
  const payload = contentType.includes("application/json")
    ? ((await response.json()) as T | { error?: APIError })
    : ((await response.text()) as unknown as T | { error?: APIError });
  return { ok: response.ok, status: response.status, payload };
}

function apiPath(base: string | undefined, path: string): string {
  const trimmedBase = (base ?? "").trim();
  if (trimmedBase === "") {
    return path;
  }
  return `${trimmedBase.replace(/\/$/, "")}${path}`;
}

function extractError(payload: unknown): APIError | null {
  if (!payload || typeof payload !== "object") {
    return null;
  }
  const error = (payload as { error?: APIError }).error;
  return error ?? null;
}

function prettyJSON(value: unknown): string {
  return JSON.stringify(value, null, 2);
}

export function Phase2Harness({ apiBase }: { apiBase?: string }) {
  const [session, setSession] = useState<SessionData | null>(null);
  const [sessionError, setSessionError] = useState<APIError | null>(null);
  const [incidents, setIncidents] = useState<IncidentData[]>([]);
  const [incidentListError, setIncidentListError] = useState<APIError | null>(
    null,
  );
  const [selectedIncidentId, setSelectedIncidentId] = useState("");
  const [selectedIncident, setSelectedIncident] = useState<IncidentData | null>(
    null,
  );
  const [memberships, setMemberships] = useState<MembershipData[]>([]);
  const [defaultPrefs, setDefaultPrefs] = useState<WorkbookPreferences | null>(
    null,
  );
  const [userPrefs, setUserPrefs] = useState<WorkbookPreferences | null>(null);
  const [incidentSurfaceError, setIncidentSurfaceError] =
    useState<APIError | null>(null);
  const [extensions, setExtensions] = useState<ExtensionProfile[]>([]);
  const [extensionsError, setExtensionsError] = useState<APIError | null>(null);
  const [lastError, setLastError] = useState<APIError | null>(null);
  const [lastProbe, setLastProbe] = useState<ProbeResult | null>(null);
  const [statusText, setStatusText] = useState("Ready");

  const [createIncidentKey, setCreateIncidentKey] = useState("");
  const [createIncidentTitle, setCreateIncidentTitle] = useState("");
  const [patchTLP, setPatchTLP] = useState("");
  const [patchCurrentPhase, setPatchCurrentPhase] = useState("");
  const [patchExternalCase, setPatchExternalCase] = useState("");
  const [membershipEmail, setMembershipEmail] = useState("");
  const [membershipRole, setMembershipRole] = useState("viewer");
  const [membershipRoleDrafts, setMembershipRoleDrafts] = useState<
    Record<string, string>
  >({});

  const loadSession = useCallback(async () => {
    const result = await fetchJSON<{ data: SessionData }>(
      apiPath(apiBase, "/api/v1/auth/session"),
    );
    if (!result.ok) {
      const error = extractError(result.payload);
      setSessionError(error);
      setSession(null);
      return;
    }
    setSessionError(null);
    setSession((result.payload as { data: SessionData }).data);
  }, [apiBase]);

  const loadIncidents = useCallback(async () => {
    const result = await fetchJSON<{ data: { incidents: IncidentData[] } }>(
      apiPath(apiBase, "/api/v1/incidents"),
    );
    if (!result.ok) {
      setIncidentListError(extractError(result.payload));
      setIncidents([]);
      return;
    }
    setIncidentListError(null);
    setIncidents(
      (result.payload as { data: { incidents: IncidentData[] } }).data
        .incidents,
    );
  }, [apiBase]);

  const loadExtensions = useCallback(async () => {
    const result = await fetchJSON<{
      data: { extensions: ExtensionProfile[] };
    }>(apiPath(apiBase, "/api/v1/extensions"));
    if (!result.ok) {
      setExtensionsError(extractError(result.payload));
      setExtensions([]);
      return;
    }
    setExtensionsError(null);
    setExtensions(
      (result.payload as { data: { extensions: ExtensionProfile[] } }).data
        .extensions,
    );
  }, [apiBase]);

  const loadIncidentSurface = useCallback(
    async (incidentId: string) => {
      if (incidentId.trim() === "") {
        setSelectedIncident(null);
        setMemberships([]);
        setDefaultPrefs(null);
        setUserPrefs(null);
        return;
      }

      const [
        incidentResult,
        membershipsResult,
        defaultPrefsResult,
        userPrefsResult,
      ] = await Promise.all([
        fetchJSON<{ data: IncidentData }>(
          apiPath(apiBase, `/api/v1/incidents/${incidentId}`),
        ),
        fetchJSON<{ data: { memberships: MembershipData[] } }>(
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
        setIncidentSurfaceError(extractError(incidentResult.payload));
        setSelectedIncident(null);
        setMemberships([]);
        setDefaultPrefs(null);
        setUserPrefs(null);
        return;
      }

      setIncidentSurfaceError(null);
      const incident = (incidentResult.payload as { data: IncidentData }).data;
      setSelectedIncident(incident);
      setPatchTLP(incident.tlp ?? "");
      setPatchCurrentPhase(incident.current_phase ?? "");
      setPatchExternalCase(incident.primary_external_case_ref ?? "");

      if (membershipsResult.ok) {
        const nextMemberships = (
          membershipsResult.payload as {
            data: { memberships: MembershipData[] };
          }
        ).data.memberships;
        setMemberships(nextMemberships);
        setMembershipRoleDrafts(
          Object.fromEntries(
            nextMemberships.map((membership) => [
              membership.user_id,
              membership.role,
            ]),
          ),
        );
      } else {
        setMemberships([]);
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
    },
    [apiBase],
  );

  useEffect(() => {
    void loadSession();
    void loadIncidents();
    void loadExtensions();
  }, [loadExtensions, loadIncidents, loadSession]);

  useEffect(() => {
    void loadIncidentSurface(selectedIncidentId);
  }, [selectedIncidentId, loadIncidentSurface]);

  async function handleCreateIncident() {
    setStatusText("Creating incident");
    const result = await fetchJSON<{ data: IncidentData }>(
      apiPath(apiBase, "/api/v1/incidents"),
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: `phase2-ui-create-${Date.now()}`,
          incident_key: createIncidentKey,
          title: createIncidentTitle,
        }),
      },
    );
    if (!result.ok) {
      setLastError(extractError(result.payload));
      setStatusText("Create failed");
      return;
    }
    setLastError(null);
    const incident = (result.payload as { data: IncidentData }).data;
    setCreateIncidentKey("");
    setCreateIncidentTitle("");
    await loadSession();
    await loadIncidents();
    setSelectedIncidentId(incident.incident_id);
    setStatusText("Created incident");
  }

  async function handlePatchIncident() {
    if (!selectedIncident) {
      return;
    }
    setStatusText("Patching incident");
    const result = await fetchJSON<{ data: IncidentData }>(
      apiPath(apiBase, `/api/v1/incidents/${selectedIncident.incident_id}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          base_incident_version: selectedIncident.incident_version,
          tlp: patchTLP === "" ? null : patchTLP,
          current_phase: patchCurrentPhase === "" ? null : patchCurrentPhase,
          primary_external_case_ref:
            patchExternalCase === "" ? null : patchExternalCase,
        }),
      },
    );
    if (!result.ok) {
      setLastError(extractError(result.payload));
      setStatusText("Patch failed");
      return;
    }
    setLastError(null);
    await loadIncidents();
    await loadIncidentSurface(selectedIncident.incident_id);
    setStatusText("Patched incident");
  }

  async function handleCreateMembership() {
    if (!selectedIncident) {
      return;
    }
    setStatusText("Creating membership");
    const result = await fetchJSON<{ data: MembershipData }>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${selectedIncident.incident_id}/memberships`,
      ),
      {
        method: "POST",
        body: JSON.stringify({
          client_txn_id: `phase2-ui-membership-${Date.now()}`,
          email: membershipEmail,
          role: membershipRole,
        }),
      },
    );
    if (!result.ok) {
      setLastError(extractError(result.payload));
      setStatusText("Membership create failed");
      return;
    }
    setLastError(null);
    setMembershipEmail("");
    setMembershipRole("viewer");
    await loadIncidentSurface(selectedIncident.incident_id);
    await loadSession();
    setStatusText("Created membership");
  }

  async function handlePatchMembership(membership: MembershipData) {
    if (!selectedIncident) {
      return;
    }
    setStatusText("Patching membership");
    const result = await fetchJSON<{ data: MembershipData }>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${selectedIncident.incident_id}/memberships/${membership.user_id}`,
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
      setLastError(extractError(result.payload));
      setStatusText("Membership patch failed");
      return;
    }
    setLastError(null);
    await loadIncidentSurface(selectedIncident.incident_id);
    await loadSession();
    setStatusText("Patched membership");
  }

  async function handleDeleteMembership(membership: MembershipData) {
    if (!selectedIncident) {
      return;
    }
    setStatusText("Deleting membership");
    const result = await fetchJSON(
      apiPath(
        apiBase,
        `/api/v1/incidents/${selectedIncident.incident_id}/memberships/${membership.user_id}`,
      ),
      {
        method: "DELETE",
        body: JSON.stringify({
          base_membership_version: membership.membership_version,
        }),
      },
    );
    if (!result.ok && result.status !== 204) {
      setLastError(extractError(result.payload));
      setStatusText("Membership delete failed");
      return;
    }
    setLastError(null);
    await loadIncidentSurface(selectedIncident.incident_id);
    await loadSession();
    setStatusText("Deleted membership");
  }

  async function runProbe(
    path: string,
    init?: RequestInit,
    overrideStatusText = "Probe complete",
  ) {
    setStatusText("Running probe");
    const result = await fetchJSON(apiPath(apiBase, path), init);
    const error = extractError(result.payload);
    setLastProbe({
      status: result.status,
      body: result.payload,
    });
    setLastError(error);
    setStatusText(overrideStatusText);
  }

  return (
    <section style={shellStyle}>
      <header style={sectionHeaderStyle}>
        <div>
          <p style={eyebrowStyle}>Phase 2 Harness</p>
          <h1 style={headlineStyle}>Incident control envelope</h1>
          <p style={bodyStyle}>
            Browser-visible create, discovery, patch, membership admin,
            workbook landing, and extension-dispatch verification.
          </p>
        </div>
        <div style={statusCardStyle}>
          <span style={labelStyle}>Status</span>
          <strong data-testid="phase2-status">{statusText}</strong>
        </div>
      </header>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Session</h2>
        {session ? (
          <div style={gridStyle}>
            <div>
              <span style={labelStyle}>User</span>
              <div data-testid="session-user-id">{session.user_id}</div>
            </div>
            <div>
              <span style={labelStyle}>Display Name</span>
              <div>{session.display_name}</div>
            </div>
            <div>
              <span style={labelStyle}>Deployment Admin</span>
              <div data-testid="session-is-deployment-admin">
                {String(session.is_deployment_admin)}
              </div>
            </div>
            <div>
              <span style={labelStyle}>Memberships</span>
              <ul data-testid="session-memberships" style={plainListStyle}>
                {session.memberships.map((membership) => (
                  <li key={membership.incident_id}>
                    {membership.incident_id} - {membership.role}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        ) : (
          <p style={bodyStyle}>
            {sessionError?.code ??
              "Sign in through the backend bootstrap flow first."}
          </p>
        )}
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Incident Create</h2>
        <div style={formRowStyle}>
          <input
            data-testid="create-incident-key"
            placeholder="IR-2026-001"
            style={inputStyle}
            type="text"
            value={createIncidentKey}
            onChange={(event) => {
              setCreateIncidentKey(event.target.value);
            }}
          />
          <input
            data-testid="create-incident-title"
            placeholder="Incident title"
            style={inputStyle}
            type="text"
            value={createIncidentTitle}
            onChange={(event) => {
              setCreateIncidentTitle(event.target.value);
            }}
          />
          <button
            data-testid="create-incident"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void handleCreateIncident();
            }}
          >
            Create incident
          </button>
          <button
            data-testid="probe-invalid-create"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void runProbe(
                "/api/v1/incidents",
                {
                  method: "POST",
                  body: JSON.stringify({
                    client_txn_id: `phase2-ui-invalid-create-${Date.now()}`,
                    incident_key: "IR-PROBE-CREATE",
                    title: "Invalid",
                    initial_memberships: [],
                    unexpected: true,
                  }),
                },
                "Invalid create probe complete",
              );
            }}
          >
            Probe invalid create
          </button>
        </div>
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Incident Discovery</h2>
        {incidentListError ? (
          <p style={bodyStyle}>{incidentListError.code}</p>
        ) : (
          <table style={tableStyle}>
            <thead>
              <tr>
                <th style={thStyle}>Incident</th>
                <th style={thStyle}>Title</th>
                <th style={thStyle}>Version</th>
                <th style={thStyle}>Action</th>
              </tr>
            </thead>
            <tbody data-testid="incident-discovery">
              {incidents.map((incident) => (
                <tr key={incident.incident_id}>
                  <td data-testid={`incident-row-${incident.incident_id}`}>
                    {incident.incident_key}
                  </td>
                  <td>{incident.title}</td>
                  <td>{incident.incident_version}</td>
                  <td>
                    <button
                      data-testid={`select-incident-${incident.incident_id}`}
                      style={secondaryButtonStyle}
                      type="button"
                      onClick={() => {
                        setSelectedIncidentId(incident.incident_id);
                      }}
                    >
                      Open
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Selected Incident</h2>
        {selectedIncident ? (
          <>
            <div style={gridStyle}>
              <div>
                <span style={labelStyle}>Incident ID</span>
                <div data-testid="current-incident-id">
                  {selectedIncident.incident_id}
                </div>
              </div>
              <div>
                <span style={labelStyle}>Incident Key</span>
                <div data-testid="current-incident-key">
                  {selectedIncident.incident_key}
                </div>
              </div>
              <div>
                <span style={labelStyle}>Title</span>
                <div data-testid="current-incident-title">
                  {selectedIncident.title}
                </div>
              </div>
              <div>
                <span style={labelStyle}>Version</span>
                <div data-testid="current-incident-version">
                  {selectedIncident.incident_version}
                </div>
              </div>
            </div>

            <div style={formRowStyle}>
              <input
                data-testid="patch-tlp"
                placeholder="tlp"
                style={inputStyle}
                type="text"
                value={patchTLP}
                onChange={(event) => {
                  setPatchTLP(event.target.value);
                }}
              />
              <input
                data-testid="patch-current-phase"
                placeholder="current_phase"
                style={inputStyle}
                type="text"
                value={patchCurrentPhase}
                onChange={(event) => {
                  setPatchCurrentPhase(event.target.value);
                }}
              />
              <input
                data-testid="patch-primary-external-case-ref"
                placeholder="primary_external_case_ref"
                style={inputStyle}
                type="text"
                value={patchExternalCase}
                onChange={(event) => {
                  setPatchExternalCase(event.target.value);
                }}
              />
              <button
                data-testid="patch-incident"
                style={buttonStyle}
                type="button"
                onClick={() => {
                  void handlePatchIncident();
                }}
              >
                Patch incident
              </button>
              <button
                data-testid="probe-invalid-patch"
                style={secondaryButtonStyle}
                type="button"
                onClick={() => {
                  void runProbe(
                    `/api/v1/incidents/${selectedIncident.incident_id}`,
                    {
                      method: "PATCH",
                      body: JSON.stringify({
                        base_incident_version: selectedIncident.incident_version,
                        title: "forbidden",
                        unknown: true,
                      }),
                    },
                    "Invalid patch probe complete",
                  );
                }}
              >
                Probe invalid patch
              </button>
            </div>

            <section style={nestedCardStyle}>
              <h3 style={subheadStyle}>Workbook Landing Surface</h3>
              <div style={gridStyle}>
                <div>
                  <span style={labelStyle}>Default Sheet Ref</span>
                  <pre
                    data-testid="default-workbook-pref"
                    style={jsonBlockStyle}
                  >
                    {prettyJSON(defaultPrefs?.default_sheet_ref ?? null)}
                  </pre>
                </div>
                <div>
                  <span style={labelStyle}>Home Sheet Ref</span>
                  <pre data-testid="user-workbook-pref" style={jsonBlockStyle}>
                    {prettyJSON(userPrefs?.home_sheet_ref ?? null)}
                  </pre>
                </div>
              </div>
            </section>

            <section style={nestedCardStyle}>
              <h3 style={subheadStyle}>Memberships</h3>
              <div style={formRowStyle}>
                <input
                  data-testid="membership-email"
                  placeholder="member@example.test"
                  style={inputStyle}
                  type="text"
                  value={membershipEmail}
                  onChange={(event) => {
                    setMembershipEmail(event.target.value);
                  }}
                />
                <select
                  data-testid="membership-role"
                  style={inputStyle}
                  value={membershipRole}
                  onChange={(event) => {
                    setMembershipRole(event.target.value);
                  }}
                >
                  <option value="viewer">viewer</option>
                  <option value="editor">editor</option>
                  <option value="reviewer">reviewer</option>
                  <option value="admin">admin</option>
                </select>
                <button
                  data-testid="create-membership"
                  style={buttonStyle}
                  type="button"
                  onClick={() => {
                    void handleCreateMembership();
                  }}
                >
                  Add membership
                </button>
              </div>
              <table style={tableStyle}>
                <thead>
                  <tr>
                    <th style={thStyle}>User</th>
                    <th style={thStyle}>Role</th>
                    <th style={thStyle}>Version</th>
                    <th style={thStyle}>Actions</th>
                  </tr>
                </thead>
                <tbody data-testid="membership-list">
                  {memberships.map((membership) => (
                    <tr key={membership.user_id}>
                      <td data-testid={`membership-row-${membership.user_id}`}>
                        {membership.display_name}
                      </td>
                      <td>
                        <select
                          data-testid={`membership-role-input-${membership.user_id}`}
                          style={inputStyle}
                          value={
                            membershipRoleDrafts[membership.user_id] ??
                            membership.role
                          }
                          onChange={(event) => {
                            setMembershipRoleDrafts((current) => ({
                              ...current,
                              [membership.user_id]: event.target.value,
                            }));
                          }}
                        >
                          <option value="viewer">viewer</option>
                          <option value="editor">editor</option>
                          <option value="reviewer">reviewer</option>
                          <option value="admin">admin</option>
                        </select>
                      </td>
                      <td
                        data-testid={`membership-version-${membership.user_id}`}
                      >
                        {membership.membership_version}
                      </td>
                      <td style={actionRowStyle}>
                        <button
                          data-testid={`patch-membership-${membership.user_id}`}
                          style={secondaryButtonStyle}
                          type="button"
                          onClick={() => {
                            void handlePatchMembership(membership);
                          }}
                        >
                          Patch
                        </button>
                        <button
                          data-testid={`delete-membership-${membership.user_id}`}
                          style={dangerButtonStyle}
                          type="button"
                          onClick={() => {
                            void handleDeleteMembership(membership);
                          }}
                        >
                          Delete
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </section>
          </>
        ) : (
          <p style={bodyStyle}>
            {incidentSurfaceError?.code ??
              "Select an incident from discovery or create a new one."}
          </p>
        )}
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Extension Discovery</h2>
        <div style={formRowStyle}>
          <button
            data-testid="reload-extensions"
            style={buttonStyle}
            type="button"
            onClick={() => {
              void loadExtensions();
            }}
          >
            Reload extensions
          </button>
          <button
            data-testid="probe-extensions-pagination"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void runProbe(
                "/api/v1/extensions?cursor_token=opaque",
                undefined,
                "Extension pagination probe complete",
              );
            }}
          >
            Probe pagination rejection
          </button>
          <button
            data-testid="probe-base-route"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void runProbe("/readyz", undefined, "Base route probe complete");
            }}
          >
            Probe base route
          </button>
          <button
            data-testid="probe-reserved-root"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void runProbe(
                "/api/v1/import-sessions",
                undefined,
                "Reserved root probe complete",
              );
            }}
          >
            Probe reserved root
          </button>
          <button
            data-testid="probe-reserved-descendant"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              const userId =
                session?.user_id ?? "00000000-0000-0000-0000-000000000000";
              void runProbe(
                `/api/v1/users/${userId}/auth-bindings/provider`,
                undefined,
                "Reserved descendant probe complete",
              );
            }}
          >
            Probe reserved descendant
          </button>
          <button
            data-testid="probe-outside-reserved"
            style={secondaryButtonStyle}
            type="button"
            onClick={() => {
              void runProbe(
                "/api/v1/outside-reserved-families",
                undefined,
                "Outside reserved probe complete",
              );
            }}
          >
            Probe outside reserved
          </button>
        </div>
        {extensionsError ? (
          <p style={bodyStyle}>{extensionsError.code}</p>
        ) : (
          <table style={tableStyle}>
            <thead>
              <tr>
                <th style={thStyle}>Profile</th>
                <th style={thStyle}>Claimed</th>
                <th style={thStyle}>Route Families</th>
              </tr>
            </thead>
            <tbody data-testid="extensions-list">
              {extensions.map((extension) => (
                <tr key={extension.profile_id}>
                  <td data-testid={`extension-${extension.profile_id}`}>
                    {extension.profile_id}
                  </td>
                  <td>{String(extension.claimed)}</td>
                  <td>{extension.route_families.join(", ")}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section style={cardStyle}>
        <h2 style={subheadStyle}>Last Result</h2>
        <div style={gridStyle}>
          <div>
            <span style={labelStyle}>Last Error Code</span>
            <div data-testid="last-error-code">{lastError?.code ?? ""}</div>
          </div>
          <div>
            <span style={labelStyle}>Last Error Details</span>
            <pre data-testid="last-error-details" style={jsonBlockStyle}>
              {prettyJSON(lastError?.details ?? null)}
            </pre>
          </div>
          <div>
            <span style={labelStyle}>Last Probe Status</span>
            <div data-testid="last-probe-status">
              {lastProbe ? String(lastProbe.status) : ""}
            </div>
          </div>
          <div>
            <span style={labelStyle}>Last Probe Payload</span>
            <pre data-testid="last-probe-payload" style={jsonBlockStyle}>
              {prettyJSON(lastProbe?.body ?? null)}
            </pre>
          </div>
        </div>
      </section>
    </section>
  );
}

const shellStyle: CSSProperties = {
  display: "grid",
  gap: "1rem",
};

const sectionHeaderStyle: CSSProperties = {
  display: "flex",
  justifyContent: "space-between",
  alignItems: "flex-start",
  gap: "1rem",
};

const cardStyle: CSSProperties = {
  border: "1px solid rgb(203 213 225)",
  borderRadius: "1rem",
  padding: "1rem",
  backgroundColor: "rgb(248 250 252)",
};

const nestedCardStyle: CSSProperties = {
  ...cardStyle,
  marginTop: "1rem",
  backgroundColor: "white",
};

const statusCardStyle: CSSProperties = {
  minWidth: "12rem",
  padding: "0.75rem 1rem",
  borderRadius: "0.75rem",
  backgroundColor: "rgb(255 247 237)",
  border: "1px solid rgb(253 186 116)",
  display: "grid",
  gap: "0.25rem",
};

const gridStyle: CSSProperties = {
  display: "grid",
  gap: "1rem",
  gridTemplateColumns: "repeat(auto-fit, minmax(14rem, 1fr))",
};

const formRowStyle: CSSProperties = {
  display: "flex",
  gap: "0.75rem",
  flexWrap: "wrap",
  alignItems: "center",
};

const actionRowStyle: CSSProperties = {
  display: "flex",
  gap: "0.5rem",
  flexWrap: "wrap",
};

const inputStyle: CSSProperties = {
  minWidth: "12rem",
  padding: "0.65rem 0.8rem",
  borderRadius: "0.75rem",
  border: "1px solid rgb(148 163 184)",
  backgroundColor: "white",
};

const buttonStyle: CSSProperties = {
  padding: "0.7rem 1rem",
  borderRadius: "999px",
  border: "none",
  backgroundColor: "rgb(15 23 42)",
  color: "white",
  cursor: "pointer",
};

const secondaryButtonStyle: CSSProperties = {
  ...buttonStyle,
  backgroundColor: "rgb(226 232 240)",
  color: "rgb(15 23 42)",
};

const dangerButtonStyle: CSSProperties = {
  ...buttonStyle,
  backgroundColor: "rgb(153 27 27)",
};

const tableStyle: CSSProperties = {
  width: "100%",
  borderCollapse: "collapse",
};

const thStyle: CSSProperties = {
  textAlign: "left",
  padding: "0.6rem",
  borderBottom: "1px solid rgb(203 213 225)",
};

const eyebrowStyle: CSSProperties = {
  margin: 0,
  textTransform: "uppercase",
  letterSpacing: "0.08em",
  color: "rgb(124 45 18)",
  fontSize: "0.8rem",
};

const headlineStyle: CSSProperties = {
  margin: "0.2rem 0 0.5rem",
  fontSize: "1.8rem",
};

const subheadStyle: CSSProperties = {
  margin: "0 0 0.75rem",
  fontSize: "1.1rem",
};

const bodyStyle: CSSProperties = {
  margin: 0,
  color: "rgb(71 85 105)",
};

const labelStyle: CSSProperties = {
  display: "block",
  marginBottom: "0.2rem",
  color: "rgb(71 85 105)",
  fontSize: "0.8rem",
  textTransform: "uppercase",
  letterSpacing: "0.06em",
};

const plainListStyle: CSSProperties = {
  listStyle: "none",
  padding: 0,
  margin: 0,
  display: "grid",
  gap: "0.2rem",
};

const jsonBlockStyle: CSSProperties = {
  margin: 0,
  padding: "0.75rem",
  backgroundColor: "rgb(15 23 42)",
  color: "rgb(226 232 240)",
  borderRadius: "0.75rem",
  overflowX: "auto",
  whiteSpace: "pre-wrap",
};
