import { ChevronRight, RefreshCw } from "lucide-react";
import { Fragment, useCallback, useEffect, useRef, useState } from "react";

import {
  type APIError,
  extractError,
  fetchJSON,
  publicErrorView,
} from "../services/browserApi";
import { administrativeAuditEventsFromPayload } from "./deploymentAuditContract";
import { formatNullableDateTime } from "./LandingAdminDisplay";
import {
  auditChangeTableStyle,
  auditDetailCellStyle,
  auditDetailMetaGridStyle,
  auditDetailPanelStyle,
  auditFilterGridStyle,
  dataTableStyle,
  definitionLabelStyle,
  emptyStateStyle,
  errorTextStyle,
  inputStyle,
  labelBlockStyle,
  metadataTextStyle,
  monoMutedStyle,
  nullValueStyle,
  primaryCellStyle,
  redactedBadgeStyle,
  secondaryButtonStyle,
  sectionEyebrowStyle,
  sectionTitleStyle,
  statusTextStyle,
  surfaceHeaderStyle,
  surfacePanelStyle,
  tableActionCellStyle,
  tableCellStyle,
  tableHeaderActionCellStyle,
  tableHeaderCellStyle,
  tableLinkButtonStyle,
  tableRowStyle,
  tableShellStyle,
} from "./landingAdminStyles";
import type { AdministrativeAuditEvent } from "./landingAdminTypes";

export function AdministrativeAuditPanel() {
  const [events, setEvents] = useState<AdministrativeAuditEvent[]>([]);
  const [expandedAuditEventID, setExpandedAuditEventID] = useState<
    string | null
  >(null);
  const [actorUserID, setActorUserID] = useState("");
  const [actionCode, setActionCode] = useState("");
  const [targetKind, setTargetKind] = useState("");
  const [targetID, setTargetID] = useState("");
  const [occurredAtGTE, setOccurredAtGTE] = useState("");
  const [occurredAtLT, setOccurredAtLT] = useState("");
  const [status, setStatus] = useState("Administrative audit idle.");
  const [error, setError] = useState<APIError | null>(null);
  const initialAuditLoadedRef = useRef(false);

  const loadAudit = useCallback(async () => {
    setStatus("Loading administrative audit.");
    const params = new URLSearchParams({ limit: "100" });
    for (const [key, value] of [
      ["actor_user_id", actorUserID],
      ["action_code", actionCode],
      ["target_kind", targetKind],
      ["target_id", targetID],
      ["occurred_at_gte", occurredAtGTE],
      ["occurred_at_lt", occurredAtLT],
    ] as const) {
      const trimmed = value.trim();
      if (trimmed !== "") {
        params.set(key, trimmed);
      }
    }
    const result = await fetchJSON<{
      data: { audit_events: AdministrativeAuditEvent[] };
    }>(`/api/v1/administrative-audit-events?${params.toString()}`);
    const nextError = extractError(result.payload);
    setError(nextError);
    if (!result.ok) {
      setStatus("Administrative audit unavailable.");
      return;
    }
    const nextEvents = administrativeAuditEventsFromPayload(result.payload);
    if (nextEvents === null) {
      setEvents([]);
      setError({
        code: "invalid_administrative_audit_response",
        details: {},
        message: "Administrative audit returned an invalid response.",
        retryable: true,
        status: 502,
      });
      setStatus("Administrative audit unavailable.");
      return;
    }
    setEvents(nextEvents);
    setStatus("Administrative audit loaded.");
  }, [
    actionCode,
    actorUserID,
    occurredAtGTE,
    occurredAtLT,
    targetID,
    targetKind,
  ]);

  useEffect(() => {
    if (initialAuditLoadedRef.current) {
      return;
    }
    initialAuditLoadedRef.current = true;
    void loadAudit();
  }, [loadAudit]);

  return (
    <section style={surfacePanelStyle}>
      <div style={surfaceHeaderStyle}>
        <div>
          <p style={sectionEyebrowStyle}>Administrative audit</p>
          <h2 style={sectionTitleStyle}>Deployment events</h2>
        </div>
        <button
          style={secondaryButtonStyle}
          type="button"
          onClick={() => {
            void loadAudit();
          }}
        >
          <RefreshCw size={15} />
          Refresh
        </button>
      </div>
      <div style={auditFilterGridStyle}>
        <label style={labelBlockStyle}>
          Actor user ID
          <input
            aria-label="Actor user id"
            style={inputStyle}
            value={actorUserID}
            onChange={(event) => setActorUserID(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Action code
          <input
            aria-label="Action code"
            style={inputStyle}
            value={actionCode}
            onChange={(event) => setActionCode(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Target kind
          <select
            aria-label="Target kind"
            style={inputStyle}
            value={targetKind}
            onChange={(event) => setTargetKind(event.target.value)}
          >
            <option value="">Any target kind</option>
            <option value="user">User</option>
            <option value="account_preferences">Account preferences</option>
            <option value="auth_binding">Auth binding</option>
            <option value="backup_set">Backup set</option>
            <option value="restore_operation">Restore operation</option>
            <option value="legacy_administrative_event">Legacy event</option>
          </select>
        </label>
        <label style={labelBlockStyle}>
          Target ID
          <input
            aria-label="Target id"
            style={inputStyle}
            value={targetID}
            onChange={(event) => setTargetID(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Occurred at or after
          <input
            aria-label="Occurred at or after"
            style={inputStyle}
            value={occurredAtGTE}
            onChange={(event) => setOccurredAtGTE(event.target.value)}
          />
        </label>
        <label style={labelBlockStyle}>
          Occurred before
          <input
            aria-label="Occurred before"
            style={inputStyle}
            value={occurredAtLT}
            onChange={(event) => setOccurredAtLT(event.target.value)}
          />
        </label>
      </div>
      <div style={tableShellStyle}>
        <table style={dataTableStyle}>
          <thead>
            <tr>
              <th style={tableHeaderCellStyle}>Occurred</th>
              <th style={tableHeaderCellStyle}>Action</th>
              <th style={tableHeaderCellStyle}>Actor</th>
              <th style={tableHeaderCellStyle}>Target</th>
              <th style={tableHeaderCellStyle}>Reason</th>
              <th style={tableHeaderCellStyle}>Changes</th>
              <th style={tableHeaderActionCellStyle}>Details</th>
            </tr>
          </thead>
          <tbody>
            {events.map((event) => {
              const expanded = expandedAuditEventID === event.audit_event_id;
              return (
                <Fragment key={event.audit_event_id}>
                  <tr key={event.audit_event_id} style={tableRowStyle}>
                    <td style={primaryCellStyle}>
                      {formatNullableDateTime(event.occurred_at)}
                      <p style={metadataTextStyle}>{event.source ?? ""}</p>
                    </td>
                    <td style={tableCellStyle}>
                      {formatAuditActionLabel(event.action_code)}
                    </td>
                    <td style={tableCellStyle}>
                      {formatAuditActor(event.actor_kind, event.actor_user_id)}
                    </td>
                    <td style={tableCellStyle}>{formatAuditTarget(event)}</td>
                    <td style={tableCellStyle}>
                      {event.reason_code ?? "No reason"}
                    </td>
                    <td style={tableCellStyle}>{event.changes.length}</td>
                    <td style={tableActionCellStyle}>
                      <button
                        aria-expanded={expanded}
                        style={tableLinkButtonStyle}
                        type="button"
                        onClick={() => {
                          setExpandedAuditEventID(
                            expanded ? null : event.audit_event_id,
                          );
                        }}
                      >
                        {expanded ? "Hide" : "Inspect"}
                        <ChevronRight size={15} />
                      </button>
                    </td>
                  </tr>
                  {expanded ? (
                    <tr key={`${event.audit_event_id}-details`}>
                      <td colSpan={7} style={auditDetailCellStyle}>
                        <div style={auditDetailPanelStyle}>
                          <div style={auditDetailMetaGridStyle}>
                            <div>
                              <span style={definitionLabelStyle}>
                                Audit event ID
                              </span>
                              <div style={monoMutedStyle}>
                                {event.audit_event_id}
                              </div>
                            </div>
                            <div>
                              <span style={definitionLabelStyle}>
                                Raw action code
                              </span>
                              <div style={monoMutedStyle}>
                                {event.action_code}
                              </div>
                            </div>
                            <div>
                              <span style={definitionLabelStyle}>
                                Target kind
                              </span>
                              <div style={monoMutedStyle}>
                                {event.target_kind}
                              </div>
                            </div>
                            <div>
                              <span style={definitionLabelStyle}>
                                Target ID
                              </span>
                              <div style={monoMutedStyle}>
                                {event.target_id ?? "No target ID"}
                              </div>
                            </div>
                          </div>
                          {event.changes.length > 0 ? (
                            <table style={auditChangeTableStyle}>
                              <thead>
                                <tr>
                                  <th style={tableHeaderCellStyle}>Field</th>
                                  <th style={tableHeaderCellStyle}>Before</th>
                                  <th style={tableHeaderCellStyle}>After</th>
                                </tr>
                              </thead>
                              <tbody>
                                {event.changes.map((change) => (
                                  <tr
                                    key={`${event.audit_event_id}-${change.field_path}`}
                                  >
                                    <td style={primaryCellStyle}>
                                      <span style={monoMutedStyle}>
                                        {change.field_path}
                                      </span>
                                    </td>
                                    <td style={tableCellStyle}>
                                      {renderAuditValue(
                                        change.value_state,
                                        change.before,
                                      )}
                                    </td>
                                    <td style={tableCellStyle}>
                                      {renderAuditValue(
                                        change.value_state,
                                        change.after,
                                      )}
                                    </td>
                                  </tr>
                                ))}
                              </tbody>
                            </table>
                          ) : (
                            <p style={emptyStateStyle}>
                              No field-level changes were published for this
                              event.
                            </p>
                          )}
                        </div>
                      </td>
                    </tr>
                  ) : null}
                </Fragment>
              );
            })}
          </tbody>
        </table>
      </div>
      {events.length === 0 ? (
        <p style={emptyStateStyle}>No administrative audit events loaded.</p>
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

function renderAuditValue(valueState: "redacted" | "visible", value: unknown) {
  if (valueState === "redacted") {
    return <span style={redactedBadgeStyle}>Redacted</span>;
  }
  if (value === null) {
    return <span style={nullValueStyle}>null</span>;
  }
  if (typeof value === "string") {
    return <span>{value}</span>;
  }
  return <span>{JSON.stringify(value)}</span>;
}

function formatAuditActionLabel(actionCode: string) {
  const label = auditActionLabels[actionCode];
  return label ?? actionCode;
}

function formatAuditActor(
  actorKind: AdministrativeAuditEvent["actor_kind"],
  actorUserID: string | null,
) {
  if (actorUserID !== null && actorUserID !== "") {
    return actorUserID;
  }
  return actorKind ?? "system";
}

function formatAuditTarget(event: AdministrativeAuditEvent) {
  const targetID = event.target_id ?? "";
  switch (event.target_kind) {
    case "user":
      return targetID === "" ? "User" : `User ${targetID}`;
    case "account_preferences":
      return targetID === ""
        ? "Account preferences"
        : `Account preferences for ${targetID}`;
    case "auth_binding":
      return targetID === "" ? "Enterprise authentication binding" : targetID;
    case "backup_set":
      return targetID === "" ? "Backup set" : `Backup set ${targetID}`;
    case "restore_operation":
      return targetID === ""
        ? "Restore operation"
        : `Restore operation ${targetID}`;
    case "legacy_administrative_event":
      return "Legacy event";
    default:
      return targetID === ""
        ? event.target_kind
        : `${event.target_kind}: ${targetID}`;
  }
}

const auditActionLabels: Record<string, string> = {
  account_preferences_updated: "Account preferences updated",
  auth_binding_created: "Authentication binding created",
  auth_binding_retired: "Authentication binding retired",
  auth_binding_rotated: "Authentication binding rotated",
  backup_created: "Backup created",
  bootstrap_admin_created: "Bootstrap admin created",
  deployment_admin_granted: "Deployment admin granted",
  deployment_admin_revoked: "Deployment admin revoked",
  legacy_administrative_event: "Legacy administrative event",
  password_changed: "Password changed",
  password_reset: "Password reset",
  restore_completed: "Restore completed",
  restore_failed: "Restore failed",
  restore_started: "Restore started",
  restore_verification_completed: "Restore verification completed",
  sessions_revoked: "Sessions revoked",
  totp_enrollment_begun: "TOTP enrollment begun",
  totp_enrollment_completed: "TOTP enrollment completed",
  totp_reset: "TOTP reset",
  user_created: "User created",
  user_profile_updated: "User profile updated",
  user_status_changed: "User status changed",
};
