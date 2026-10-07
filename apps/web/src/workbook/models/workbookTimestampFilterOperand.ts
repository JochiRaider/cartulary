/** Query operand syntax, independent of writable-field and workflow contracts. */
export const timestampFilterGuidance =
  "Enter a real timestamp with Z or a numeric timezone offset, for example 2026-04-18T00:00:00Z or 2026-04-18T02:00:00+02:00.";

/** Parse explicit-zone query text without calendar rollover or millisecond rounding. */
export function timestampFilterInstant(value: unknown): bigint | null {
  if (typeof value !== "string") return null;
  const match =
    /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.(\d+))?(Z|([+-])(\d{2}):(\d{2}))$/u.exec(
      value,
    );
  if (!match) return null;
  const year = Number(match[1]),
    month = Number(match[2]),
    day = Number(match[3]);
  const hour = Number(match[4]),
    minute = Number(match[5]),
    second = Number(match[6]);
  const offsetHour = Number(match[10] ?? 0),
    offsetMinute = Number(match[11] ?? 0);
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
  // setUTCFullYear preserves four-digit years 0000..0099 without Date.UTC's 1900 adjustment.
  date.setUTCFullYear(year, month - 1, day);
  date.setUTCHours(hour, minute, second, 0);
  if (
    date.getUTCFullYear() !== year ||
    date.getUTCMonth() !== month - 1 ||
    date.getUTCDate() !== day
  )
    return null;
  const offset = (offsetHour * 60 + offsetMinute) * (match[9] === "-" ? -1 : 1);
  // Match Go's existing nanosecond truncation, retaining the original draft/wire text.
  const fraction = (match[7] ?? "").slice(0, 9).padEnd(9, "0");
  return (
    BigInt(date.getTime() - offset * 60_000) * 1_000_000n + BigInt(fraction)
  );
}
