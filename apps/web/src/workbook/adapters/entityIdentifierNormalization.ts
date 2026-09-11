import { entityIdentifierUnicode as unicode } from "@cartulary/protocol-ts/entities";

type ScalarRange = readonly [number, number];

function contains(ranges: readonly ScalarRange[], scalar: number): boolean {
  let left = 0;
  let right = ranges.length;
  while (left < right) {
    const middle = (left + right) >>> 1;
    if ((ranges[middle]?.[1] ?? -1) < scalar) left = middle + 1;
    else right = middle;
  }
  return left < ranges.length && (ranges[left]?.[0] ?? 0x110000) <= scalar;
}

function caseMap(mapping: {
  readonly source: string;
  readonly target: string;
}) {
  const targets = Array.from(mapping.target);
  return new Map(
    Array.from(mapping.source, (source, index) => [source, targets[index]]),
  );
}

const lowercase = caseMap(unicode.lowercase);
const sidUppercase = caseMap(unicode.sid_uppercase);
const lowercaseClasses = new Set([
  "aad_device_id",
  "fqdn",
  "hostname",
  "aad_object_id",
  "upn",
  "email",
  "sam_account_name",
]);

// Native NFC is stable on the frozen Unicode 16 repertoire. Scalars unassigned
// in that version remain opaque starters even in a newer browser's Unicode.
function normalizedText(raw: string): string | null {
  const scalars = Array.from(raw);
  if (
    scalars.some((value) => {
      const scalar = value.codePointAt(0) ?? 0;
      return scalar >= 0xd800 && scalar <= 0xdfff;
    })
  )
    return null;
  let start = 0;
  let end = scalars.length;
  while (
    start < end &&
    contains(unicode.whitespace_ranges, scalars[start]?.codePointAt(0) ?? 0)
  )
    start++;
  while (
    end > start &&
    contains(unicode.whitespace_ranges, scalars[end - 1]?.codePointAt(0) ?? 0)
  )
    end--;
  let result = "";
  let segment = "";
  for (const value of scalars.slice(start, end)) {
    if (contains(unicode.assigned_ranges, value.codePointAt(0) ?? 0))
      segment += value;
    else {
      result += segment.normalize("NFC") + value;
      segment = "";
    }
  }
  result += segment.normalize("NFC");
  return result === "" ? null : result;
}

export function normalizeEntityIdentifier(
  identifierClass: string,
  raw: string,
): string | null {
  const mapping =
    identifierClass === "sid"
      ? sidUppercase
      : lowercaseClasses.has(identifierClass)
        ? lowercase
        : null;
  if (mapping === null) return null;
  const normalized = normalizedText(raw);
  if (normalized === null) return null;
  const scalars = Array.from(normalized);
  if (
    scalars.some((value) =>
      contains(unicode.identifier_rejected_ranges, value.codePointAt(0) ?? 0),
    )
  )
    return null;
  return scalars.map((value) => mapping.get(value) ?? value).join("");
}

export function normalizeEntityAlias(raw: string): string | null {
  const normalized = normalizedText(raw);
  if (normalized === null) return null;
  const scalars = Array.from(normalized);
  if (
    scalars.length > 256 ||
    scalars.some((value) => {
      const scalar = value.codePointAt(0) ?? 0;
      return scalar <= 31 || (scalar >= 127 && scalar <= 159);
    })
  )
    return null;
  return normalized;
}

export function compareEntityNormalizedValues(
  left: string,
  right: string,
): number {
  const a = Array.from(left, (value) => value.codePointAt(0) ?? 0);
  const b = Array.from(right, (value) => value.codePointAt(0) ?? 0);
  for (let index = 0; index < Math.min(a.length, b.length); index++) {
    if (a[index] !== b[index]) return (a[index] ?? 0) - (b[index] ?? 0);
  }
  return a.length - b.length;
}
