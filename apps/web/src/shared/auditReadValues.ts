export const auditFilterFields = [
  "actor_user_id",
  "action_code",
  "target_kind",
  "target_id",
  "occurred_at_gte",
  "occurred_at_lt",
] as const;
export type AuditField = (typeof auditFilterFields)[number];
export type AuditInputs = Readonly<Record<AuditField, string>>;
export type AuditFieldErrors = Partial<Record<AuditField, string>>;
const uuidPattern =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/iu;

/** Match the route's UTC RFC3339Nano normalization without millisecond rounding. */
export function normalizeAuditInstant(
  input: string,
): { text: string; nanoseconds: bigint } | null {
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?(Z|([+-])(\d{2}):(\d{2}))$/u.exec(
      input,
    );
  if (!match) return null;
  const year = Number(match[1]);
  const month = Number(match[2]);
  const day = Number(match[3]);
  const hour = Number(match[4]);
  const minute = Number(match[5]);
  const second = Number(match[6]);
  const offsetHour = Number(match[10] ?? 0);
  const offsetMinute = Number(match[11] ?? 0);
  if (
    month < 1 ||
    month > 12 ||
    day < 1 ||
    hour > 23 ||
    minute > 59 ||
    second > 59 ||
    offsetHour > 23 ||
    offsetMinute > 59
  )
    return null;
  const date = new Date(0);
  date.setUTCFullYear(year, month - 1, day);
  date.setUTCHours(hour, minute, second, 0);
  if (
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  )
    return null;
  const offset = (offsetHour * 60 + offsetMinute) * (match[9] === "-" ? -1 : 1);
  date.setTime(date.getTime() - offset * 60_000);
  if (date.getUTCFullYear() < 0 || date.getUTCFullYear() > 9999) return null;
  // The existing Go route admits fractional digits at nanosecond precision.
  const fraction = (match[7] ?? "").slice(0, 9).padEnd(9, "0");
  const significant = fraction.replace(/0+$/u, "");
  return {
    text: `${date.toISOString().slice(0, 19)}${significant ? `.${significant}` : ""}Z`,
    nanoseconds: BigInt(date.getTime()) * 1_000_000n + BigInt(fraction),
  };
}

export function normalizeAuditFilters(inputs: AuditInputs) {
  const values = { ...inputs };
  const errors: AuditFieldErrors = {};
  for (const field of auditFilterFields) {
    values[field] = inputs[field].trim();
    const value = values[field];
    if (
      value &&
      (value === "null" ||
        value.includes(",") ||
        value.startsWith("[") ||
        value.endsWith("]"))
    )
      errors[field] = "Enter one exact value, or leave this filter blank.";
  }
  if (values.actor_user_id && !uuidPattern.test(values.actor_user_id))
    errors.actor_user_id = "Enter a user ID in UUID format.";
  values.actor_user_id = values.actor_user_id.toLowerCase();
  if (values.target_id && !values.target_kind)
    errors.target_kind = "Choose a target kind when supplying a target ID.";
  const lower = values.occurred_at_gte
    ? normalizeAuditInstant(values.occurred_at_gte)
    : null;
  const upper = values.occurred_at_lt
    ? normalizeAuditInstant(values.occurred_at_lt)
    : null;
  for (const [field, parsed] of [
    ["occurred_at_gte", lower],
    ["occurred_at_lt", upper],
  ] as const) {
    if (values[field] && !parsed)
      errors[field] =
        "Enter a valid RFC 3339 timestamp with Z or a numeric timezone offset.";
    if (parsed) values[field] = parsed.text;
  }
  if (lower && upper && lower.nanoseconds >= upper.nanoseconds)
    errors.occurred_at_lt =
      "The exclusive upper bound must be later than the inclusive lower bound.";
  return { values, errors };
}

export function formatAuditJSON(value: unknown): string {
  return JSON.stringify(value, null, 2);
}
