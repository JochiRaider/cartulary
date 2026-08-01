import {
  incidentAdministrationTestId,
  incidentControlsActionMessageTestId,
  incidentControlsStatusTestId,
  incidentControlsSurfaceTestId,
  incidentMembershipAdminNoteTestId,
  incidentMembershipAuditRowTestId,
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
import type {
  CloseIncidentRequest,
  CloseIncidentResponse,
  ListIncidentMembershipAuditEventsResponse,
  ReopenIncidentRequest,
  ReopenIncidentResponse,
} from "./api/publicHttpTypes";
import type {
  IncidentControlsLoadState,
  IncidentControlsSection,
} from "./landingAdminTypes";

type IncidentRole = "viewer" | "editor" | "reviewer" | "admin" | "";
type MembershipRole = Exclude<IncidentRole, "">;

type IncidentSummary = CloseIncidentResponse["data"];

type MembershipRecord = {
  incident_id: string;
  user_id: string;
  display_name: string;
  role: MembershipRole;
  membership_version: number;
};

type MembershipAuditEvent =
  ListIncidentMembershipAuditEventsResponse["data"]["audit_events"][number];

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
  incidentId: string;
  currentIncidentRole: IncidentRole | null;
  activeSection?: IncidentControlsSection | undefined;
  apiBase?: string | undefined;
  onIncidentAccessLost?: (() => void) | undefined;
  onIncidentSnapshot?: ((incident: IncidentSummary) => void) | undefined;
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

function upsertMembershipRoleDrafts(records: MembershipRecord[]) {
  return Object.fromEntries(
    records.map((record) => [record.user_id, record.role]),
  ) as Record<string, MembershipRole>;
}

export function IncidentAdminPanel({
  incidentId,
  currentIncidentRole,
  activeSection = "summary",
  apiBase,
  onIncidentAccessLost,
  onIncidentSnapshot,
  onSessionRoleChange,
}: IncidentAdminPanelProps) {
  const [incident, setIncident] = useState<IncidentSummary | null>(null);
  const [memberships, setMemberships] = useState<MembershipRecord[]>([]);
  const [defaultPreference, setDefaultPreference] = useState<PreferenceSlot>(
    loadingPreferenceSlot,
  );
  const [userPreference, setUserPreference] = useState<PreferenceSlot>(
    loadingPreferenceSlot,
  );
  const [patchDescription, setPatchDescription] = useState("");
  const [patchSeverity, setPatchSeverity] = useState("");
  const [patchTLP, setPatchTLP] = useState("");
  const [patchCurrentPhase, setPatchCurrentPhase] = useState("");
  const [patchExternalCase, setPatchExternalCase] = useState("");
  const [lifecycleReason, setLifecycleReason] = useState("");
  const [membershipEmail, setMembershipEmail] = useState("");
  const [membershipRole, setMembershipRole] =
    useState<MembershipRole>("viewer");
  const [membershipRoleDrafts, setMembershipRoleDrafts] = useState<
    Record<string, MembershipRole>
  >({});
  const [membershipAuditEvents, setMembershipAuditEvents] = useState<
    MembershipAuditEvent[]
  >([]);
  const [membershipAuditAction, setMembershipAuditAction] = useState("");
  const [membershipAuditActor, setMembershipAuditActor] = useState("");
  const [membershipAuditTargetID, setMembershipAuditTargetID] = useState("");
  const [membershipAuditNextCursor, setMembershipAuditNextCursor] = useState<
    string | null
  >(null);
  const [membershipAuditStatus, setMembershipAuditStatus] = useState(
    "Membership audit idle.",
  );
  const membershipAuditFiltersRef = useRef({
    action: membershipAuditAction,
    actor: membershipAuditActor,
    targetID: membershipAuditTargetID,
  });
  const [surfaceLoadState, setSurfaceLoadState] =
    useState<IncidentControlsLoadState>("loading");
  const [surfaceStatusText, setSurfaceStatusText] = useState(
    "Loading incident controls…",
  );
  const [actionMessageState, setActionMessageState] =
    useState<IncidentActionMessage>({ incidentId, text: "" });
  const [error, setError] = useState<APIError | null>(null);
  const loadRequestIdRef = useRef(0);
  const activeSectionRef = useRef(activeSection);
  const apiBaseRef = useRef(apiBase);
  const incidentIdRef = useRef(incidentId);

  activeSectionRef.current = activeSection;
  apiBaseRef.current = apiBase;
  incidentIdRef.current = incidentId;
  membershipAuditFiltersRef.current = {
    action: membershipAuditAction,
    actor: membershipAuditActor,
    targetID: membershipAuditTargetID,
  };

  const canEditIncident =
    incident?.status !== "closed" &&
    (currentIncidentRole === "reviewer" || currentIncidentRole === "admin");
  const canManageMemberships = currentIncidentRole === "admin";
  const actionMessage =
    actionMessageState.incidentId === incidentId ? actionMessageState.text : "";

  function setActionMessageForIncident(
    messageIncidentId: string,
    text: string,
  ) {
    if (messageIncidentId !== incidentIdRef.current) {
      return;
    }
    setActionMessageState({ incidentId: messageIncidentId, text });
  }

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
      const membershipsRequest =
        requestedSection === "memberships"
          ? fetchJSON<{ data: { memberships: MembershipRecord[] } }>(
              apiPath(
                requestedApiBase,
                `/api/v1/incidents/${requestedIncidentId}/memberships`,
              ),
            )
          : Promise.resolve(null);
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

      const [
        incidentResult,
        membershipsResult,
        defaultPrefsResult,
        userPrefsResult,
      ] = await Promise.all([
        incidentRequest,
        membershipsRequest,
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
        setMemberships([]);
        setMembershipRoleDrafts({});
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
      setIncident(nextIncident);
      onIncidentSnapshot?.(nextIncident);
      setPatchDescription(nextIncident.description ?? "");
      setPatchSeverity(nextIncident.severity ?? "");
      setPatchTLP(nextIncident.tlp ?? "");
      setPatchCurrentPhase(nextIncident.current_phase ?? "");
      setPatchExternalCase(nextIncident.primary_external_case_ref ?? "");

      let partialFailure = false;

      if (requestedSection === "memberships") {
        if (membershipsResult?.ok) {
          const nextMemberships = (
            membershipsResult.payload as {
              data: { memberships: MembershipRecord[] };
            }
          ).data.memberships;
          setMemberships(nextMemberships);
          setMembershipRoleDrafts(upsertMembershipRoleDrafts(nextMemberships));
        } else {
          partialFailure = true;
          setMemberships([]);
          setMembershipRoleDrafts({});
        }
      }

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
          ? requestedSection === "summary"
            ? "Incident summary synced; workbook preferences unavailable."
            : "Incident controls synced; memberships unavailable."
          : "Incident controls synced.",
      );
    },
    [onIncidentAccessLost, onIncidentSnapshot],
  );

  useEffect(() => {
    void loadIncidentSurface({ activeSection, apiBase, incidentId });
  }, [activeSection, apiBase, incidentId, loadIncidentSurface]);

  useEffect(() => {
    setActionMessageState((current) =>
      current.incidentId === incidentId ? current : { incidentId, text: "" },
    );
  }, [incidentId]);

  const loadMembershipAudit = useCallback(
    async (cursorToken?: string) => {
      if (currentIncidentRole !== "admin") {
        setMembershipAuditEvents([]);
        setMembershipAuditNextCursor(null);
        setMembershipAuditStatus(
          "Only incident admins can review membership audit.",
        );
        return;
      }
      setMembershipAuditStatus("Loading membership audit…");
      const query: Record<string, string | number> = { limit: 50 };
      const filters = membershipAuditFiltersRef.current;
      if (filters.action !== "") {
        query.action_code = filters.action;
      }
      if (filters.actor.trim() !== "") {
        query.actor_user_id = filters.actor.trim();
      }
      if (filters.targetID.trim() !== "") {
        query.target_kind = "incident_membership";
        query.target_id = filters.targetID.trim();
      }
      if (cursorToken !== undefined) {
        query.cursor_token = cursorToken;
      }
      const result =
        await fetchHTTPOperation<ListIncidentMembershipAuditEventsResponse>({
          apiBase,
          operationID: "listIncidentMembershipAuditEvents",
          pathParameters: { incident_id: incidentId },
          query,
        });
      if (!result.ok) {
        setError(extractError(result.payload));
        setMembershipAuditStatus("Membership audit unavailable.");
        return;
      }
      const payload =
        result.payload as ListIncidentMembershipAuditEventsResponse;
      if (typeof payload.meta.paging === "undefined") {
        setError({
          code: "invalid_membership_audit_response",
          details: {
            instance_path: "/meta/paging",
            schema_id: "cartulary.core_http.AdministrativeAuditEnvelope.v1",
          },
          message: "Membership audit response omitted paging metadata.",
          retryable: true,
          status: 502,
        });
        setMembershipAuditStatus("Membership audit unavailable.");
        return;
      }
      setMembershipAuditEvents((current) =>
        cursorToken === undefined
          ? payload.data.audit_events
          : [...current, ...payload.data.audit_events],
      );
      setMembershipAuditNextCursor(payload.meta.paging.next_cursor);
      setError(null);
      setMembershipAuditStatus(
        payload.data.audit_events.length === 0 && cursorToken === undefined
          ? "No membership audit events matched."
          : "Membership audit loaded.",
      );
    },
    [apiBase, currentIncidentRole, incidentId],
  );

  useEffect(() => {
    if (activeSection === "membership-audit") {
      void loadMembershipAudit();
    }
  }, [activeSection, loadMembershipAudit]);

  async function handlePatchIncident() {
    if (!incident) {
      return;
    }

    const actionIncidentId = incident.incident_id;
    setActionMessageForIncident(
      actionIncidentId,
      "Saving promoted incident fields…",
    );
    const result = await fetchJSON<{ data: IncidentSummary }>(
      apiPath(apiBase, `/api/v1/incidents/${incident.incident_id}`),
      {
        method: "PATCH",
        body: JSON.stringify({
          base_incident_version: incident.incident_version,
          description: patchDescription.trim() === "" ? null : patchDescription,
          severity: patchSeverity.trim() === "" ? null : patchSeverity,
          tlp: patchTLP === "" ? null : patchTLP,
          current_phase:
            patchCurrentPhase.trim() === "" ? null : patchCurrentPhase,
          primary_external_case_ref:
            patchExternalCase.trim() === "" ? null : patchExternalCase,
        }),
      },
    );
    if (!result.ok) {
      setError(extractError(result.payload));
      setActionMessageForIncident(actionIncidentId, "Incident update failed.");
      return;
    }

    setError(null);
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setActionMessageForIncident(
      actionIncidentId,
      "Saved promoted incident fields.",
    );
  }

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
    onIncidentSnapshot?.(nextIncident);
    setLifecycleReason("");
    await refreshSessionRole();
    setActionMessageForIncident(
      actionIncidentId,
      action === "close" ? "Incident closed." : "Incident reopened.",
    );
  }

  async function handleCreateMembership() {
    if (!incident) {
      return;
    }

    const actionIncidentId = incident.incident_id;
    setActionMessageForIncident(actionIncidentId, "Adding membership…");
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
      setActionMessageForIncident(
        actionIncidentId,
        "Membership create failed.",
      );
      return;
    }

    setError(null);
    setMembershipEmail("");
    setMembershipRole("viewer");
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setActionMessageForIncident(actionIncidentId, "Added membership.");
  }

  async function handlePatchMembership(membership: MembershipRecord) {
    if (!incident) {
      return;
    }

    const actionIncidentId = incident.incident_id;
    setActionMessageForIncident(actionIncidentId, "Updating membership…");
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
      setActionMessageForIncident(
        actionIncidentId,
        "Membership update failed.",
      );
      return;
    }

    setError(null);
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setActionMessageForIncident(actionIncidentId, "Updated membership.");
  }

  async function handleDeleteMembership(membership: MembershipRecord) {
    if (!incident) {
      return;
    }

    const actionIncidentId = incident.incident_id;
    setActionMessageForIncident(actionIncidentId, "Removing membership…");
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
      setActionMessageForIncident(
        actionIncidentId,
        "Membership delete failed.",
      );
      return;
    }

    setError(null);
    await Promise.all([loadIncidentSurface(), refreshSessionRole()]);
    setActionMessageForIncident(actionIncidentId, "Removed membership.");
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

      {activeSection === "incident-fields" ? (
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
                Description
                <textarea
                  data-testid={incidentAdministrationTestId(
                    "patch-description",
                  )}
                  style={textAreaStyle}
                  value={patchDescription}
                  onChange={(event) => {
                    setPatchDescription(event.target.value);
                  }}
                />
              </label>
              <label style={fieldLabelStyle}>
                Severity
                <input
                  data-testid={incidentAdministrationTestId("patch-severity")}
                  style={inputStyle}
                  value={patchSeverity}
                  onChange={(event) => {
                    setPatchSeverity(event.target.value);
                  }}
                  placeholder="high"
                />
              </label>
              <label style={fieldLabelStyle}>
                TLP
                <select
                  data-testid={incidentAdministrationTestId("patch-tlp")}
                  style={inputStyle}
                  value={patchTLP}
                  onChange={(event) => {
                    setPatchTLP(event.target.value);
                  }}
                >
                  <option value="">Unset</option>
                  <option value="TLP:CLEAR">TLP:CLEAR</option>
                  <option value="TLP:GREEN">TLP:GREEN</option>
                  <option value="TLP:AMBER">TLP:AMBER</option>
                  <option value="TLP:AMBER+STRICT">TLP:AMBER+STRICT</option>
                  <option value="TLP:RED">TLP:RED</option>
                </select>
              </label>
              <label style={fieldLabelStyle}>
                Current phase
                <input
                  data-testid={incidentAdministrationTestId(
                    "patch-current-phase",
                  )}
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
                  data-testid={incidentAdministrationTestId(
                    "patch-external-case",
                  )}
                  style={inputStyle}
                  value={patchExternalCase}
                  onChange={(event) => {
                    setPatchExternalCase(event.target.value);
                  }}
                  placeholder="CASE-1234"
                />
              </label>
              <button
                data-testid={incidentAdministrationTestId("patch-button")}
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
              data-testid={incidentAdministrationTestId("patch-readonly-note")}
              style={mutedBodyStyle}
            >
              Promoted incident fields are read-only for this incident role.
            </p>
          )}
        </section>
      ) : null}

      {activeSection === "memberships" ? (
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
      ) : null}

      {activeSection === "membership-audit" ? (
        <section style={cardStyle}>
          <div style={cardHeaderStyle}>
            <div>
              <p style={cardEyebrowStyle}>Membership audit</p>
              <h3 style={cardTitleStyle}>Incident membership audit</h3>
            </div>
          </div>
          {canManageMemberships ? (
            <>
              <div style={formGridStyle}>
                <label style={fieldLabelStyle}>
                  Action
                  <select
                    aria-label="Membership audit action"
                    style={inputStyle}
                    value={membershipAuditAction}
                    onChange={(event) => {
                      setMembershipAuditAction(event.target.value);
                    }}
                  >
                    <option value="">Any action</option>
                    <option value="membership_created">Created</option>
                    <option value="membership_role_changed">
                      Role changed
                    </option>
                    <option value="membership_deleted">Deleted</option>
                  </select>
                </label>
                <label style={fieldLabelStyle}>
                  Actor user ID
                  <input
                    aria-label="Membership audit actor user id"
                    style={inputStyle}
                    value={membershipAuditActor}
                    onChange={(event) => {
                      setMembershipAuditActor(event.target.value);
                    }}
                  />
                </label>
                <label style={fieldLabelStyle}>
                  Target user ID
                  <input
                    aria-label="Membership audit target user id"
                    style={inputStyle}
                    value={membershipAuditTargetID}
                    onChange={(event) => {
                      setMembershipAuditTargetID(event.target.value);
                    }}
                  />
                </label>
                <button
                  data-testid="membership-audit-apply-filters"
                  style={secondaryButtonStyle}
                  type="button"
                  onClick={() => {
                    void loadMembershipAudit();
                  }}
                >
                  Apply filters
                </button>
              </div>
              <p
                aria-live="polite"
                data-testid="membership-audit-status"
                role="status"
                style={mutedBodyStyle}
              >
                {membershipAuditStatus}
              </p>
              {membershipAuditEvents.length === 0 ? (
                <p data-testid="membership-audit-empty" style={bodyStyle}>
                  No membership audit events loaded.
                </p>
              ) : (
                <ul
                  aria-label="Incident membership audit events"
                  data-testid="membership-audit-list"
                  style={membershipListStyle}
                >
                  {membershipAuditEvents.map((event) => (
                    <li
                      data-testid={incidentMembershipAuditRowTestId(
                        event.audit_event_id,
                      )}
                      key={event.audit_event_id}
                      style={membershipCardStyle}
                    >
                      <div>
                        <strong>{formatMembershipAuditAction(event)}</strong>
                        <p style={mutedBodyStyle}>
                          {new Date(event.occurred_at).toLocaleString(
                            undefined,
                            {
                              dateStyle: "medium",
                              timeStyle: "medium",
                              timeZone: "UTC",
                            },
                          )}{" "}
                          UTC
                        </p>
                      </div>
                      <dl style={definitionGridStyle}>
                        <div>
                          <dt style={labelStyle}>Actor</dt>
                          <dd style={valueStyle}>
                            {event.actor_user_id ?? event.actor_kind}
                          </dd>
                        </div>
                        <div>
                          <dt style={labelStyle}>Target</dt>
                          <dd style={valueStyle}>
                            {event.target_id ?? "Unset"}
                          </dd>
                        </div>
                      </dl>
                      <ul>
                        {event.changes.map((change) => (
                          <li key={change.field_path}>
                            {change.field_path}:{" "}
                            {change.value_state === "redacted"
                              ? "redacted"
                              : `${String(change.before)} → ${String(change.after)}`}
                          </li>
                        ))}
                      </ul>
                    </li>
                  ))}
                </ul>
              )}
              {membershipAuditNextCursor !== null ? (
                <button
                  data-testid="membership-audit-load-more"
                  style={secondaryButtonStyle}
                  type="button"
                  onClick={() => {
                    void loadMembershipAudit(membershipAuditNextCursor);
                  }}
                >
                  Load more
                </button>
              ) : null}
            </>
          ) : (
            <p data-testid="membership-audit-note" style={mutedBodyStyle}>
              Only incident admins can review incident membership audit.
            </p>
          )}
        </section>
      ) : null}
    </section>
  );
}

function formatMembershipAuditAction(event: MembershipAuditEvent): string {
  switch (event.action_code) {
    case "membership_created":
      return "Membership created";
    case "membership_role_changed":
      return "Membership role changed";
    case "membership_deleted":
      return "Membership deleted";
    default:
      return event.action_code;
  }
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

const mutedBodyStyle = {
  margin: 0,
  color: "var(--ct-colors-ink-subtle)",
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

const subtleValueStyle = {
  color: "var(--ct-colors-ink-subtle)",
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

const textAreaStyle = {
  ...inputStyle,
  minHeight: "5.5rem",
  resize: "vertical" as const,
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

const dangerButtonStyle = {
  borderRadius: "var(--ct-component-button-danger-rounded)",
  border: "1px solid var(--ct-colors-semantic-destructive)",
  padding: "var(--ct-component-button-danger-padding)",
  background: "var(--ct-component-button-danger-backgroundColor)",
  color: "var(--ct-component-button-danger-textColor)",
  font: "inherit",
  fontWeight: 700,
  cursor: "pointer",
};

const membershipListStyle = {
  display: "grid",
  gap: "0.75rem",
};

const membershipCardStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  padding: "0.85rem",
  background: "var(--ct-colors-surface-2)",
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
