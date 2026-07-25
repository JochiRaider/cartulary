import type { AdministrativeAuditEvent } from "./landingAdminTypes";

export function administrativeAuditEventsFromPayload(
  payload: unknown,
): AdministrativeAuditEvent[] | null {
  if (!payload || typeof payload !== "object") {
    return null;
  }
  const data = (payload as { data?: unknown }).data;
  if (!data || typeof data !== "object" || Array.isArray(data)) {
    return null;
  }
  if (Object.keys(data).length !== 1 || !hasOwn(data, "audit_events")) {
    return null;
  }
  const auditEvents = (data as { audit_events?: unknown }).audit_events;
  if (!Array.isArray(auditEvents)) {
    return null;
  }
  const decoded: AdministrativeAuditEvent[] = [];
  for (const auditEvent of auditEvents) {
    const event = decodeAdministrativeAuditEvent(auditEvent);
    if (event === null) {
      return null;
    }
    decoded.push(event);
  }
  return decoded;
}

const administrativeAuditEventKeys = [
  "action_code",
  "actor_kind",
  "actor_user_id",
  "audit_event_id",
  "changes",
  "occurred_at",
  "reason_code",
  "scope_id",
  "scope_kind",
  "source",
  "target_id",
  "target_kind",
] as const;

function decodeAdministrativeAuditEvent(
  value: unknown,
): AdministrativeAuditEvent | null {
  if (!hasExactKeys(value, administrativeAuditEventKeys)) {
    return null;
  }
  const event = value as Record<string, unknown>;
  if (
    typeof event.audit_event_id !== "string" ||
    typeof event.occurred_at !== "string" ||
    typeof event.action_code !== "string" ||
    typeof event.target_kind !== "string" ||
    !isNullableString(event.actor_user_id) ||
    !isNullableString(event.scope_id) ||
    !isNullableString(event.target_id) ||
    !isNullableString(event.reason_code) ||
    !isOneOf(event.scope_kind, ["deployment", "incident"]) ||
    !isOneOf(event.actor_kind, ["operator", "system", "user"]) ||
    !isOneOf(event.source, ["api", "operator", "startup", "system", "ui"]) ||
    !Array.isArray(event.changes)
  ) {
    return null;
  }
  if (
    (event.scope_kind === "deployment" && event.scope_id !== null) ||
    (event.scope_kind === "incident" && event.scope_id === null) ||
    (event.actor_kind === "user" && event.actor_user_id === null) ||
    (event.actor_kind !== "user" && event.actor_user_id !== null)
  ) {
    return null;
  }
  const changes = [];
  for (const rawChange of event.changes) {
    if (
      !hasExactKeys(rawChange, ["after", "before", "field_path", "value_state"])
    ) {
      return null;
    }
    if (
      typeof rawChange.field_path !== "string" ||
      !isOneOf(rawChange.value_state, ["redacted", "visible"]) ||
      (rawChange.value_state === "redacted" &&
        (rawChange.before !== null || rawChange.after !== null))
    ) {
      return null;
    }
    changes.push({
      field_path: rawChange.field_path,
      value_state: rawChange.value_state,
      before: rawChange.before,
      after: rawChange.after,
    });
  }
  return {
    audit_event_id: event.audit_event_id,
    scope_kind: event.scope_kind,
    scope_id: event.scope_id,
    occurred_at: event.occurred_at,
    actor_kind: event.actor_kind,
    actor_user_id: event.actor_user_id,
    source: event.source,
    action_code: event.action_code,
    target_kind: event.target_kind,
    target_id: event.target_id,
    changes,
    reason_code: event.reason_code,
  };
}

function hasExactKeys<const Key extends string>(
  value: unknown,
  keys: readonly Key[],
): value is Record<Key, unknown> {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return false;
  }
  const actual = Object.keys(value).sort();
  const expected = [...keys].sort();
  return (
    actual.length === expected.length &&
    actual.every((key, index) => key === expected[index])
  );
}

function hasOwn(value: object, key: PropertyKey): boolean {
  return Object.hasOwn(value, key);
}

function isNullableString(value: unknown): value is string | null {
  return value === null || typeof value === "string";
}

function isOneOf<const Value extends string>(
  value: unknown,
  allowed: readonly Value[],
): value is Value {
  return typeof value === "string" && allowed.includes(value as Value);
}
