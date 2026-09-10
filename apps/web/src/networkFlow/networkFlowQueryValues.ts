import { networkFlowQueryMetadata } from "../services/networkFlowContractAdapter";

export type QueryValueKind =
  (typeof networkFlowQueryMetadata.fields)[number]["kind"];
export type ParsedQueryScalar = {
  readonly value: string | number;
  readonly canonical: string;
  readonly order: bigint | string;
};
export type ScalarResult =
  | { readonly ok: true; readonly scalar: ParsedQueryScalar }
  | { readonly ok: false; readonly message: string };

const invalid = (message: string): ScalarResult => ({ ok: false, message });
const decimal = /^(?:0|[1-9][0-9]*)$/u;
const uint64Maximum = (1n << 64n) - 1n;

export function parseQueryScalar(
  kind: QueryValueKind,
  text: string,
  cidr = false,
): ScalarResult {
  switch (kind) {
    case "ip": {
      const parts = text.split("/");
      if (cidr ? parts.length !== 2 : parts.length !== 1)
        return invalid(
          cidr
            ? "Enter an IP address and prefix length."
            : "Enter an IPv4 or IPv6 address.",
        );
      const address = parseQueryIP(parts[0] ?? "");
      if (!address)
        return invalid(
          "Enter a valid IPv4 or IPv6 literal without a zone or leading-zero octets.",
        );
      let bits = address.bits;
      let canonical = address.canonical;
      if (cidr) {
        const prefix = parts[1] ?? "";
        if (
          !decimal.test(prefix) ||
          BigInt(prefix) > BigInt(address.family === 4 ? 32 : 128)
        )
          return invalid("The prefix must be 0–32 for IPv4 or 0–128 for IPv6.");
        const shift = BigInt(
          (address.family === 4 ? 32 : 128) - Number(prefix),
        );
        bits = (bits >> shift) << shift;
        canonical = formatIP(address.family, bits) + "/" + prefix;
      }
      return {
        ok: true,
        scalar: {
          value: text,
          canonical,
          order: (BigInt(address.family) << 128n) + bits,
        },
      };
    }
    case "counter":
    case "port":
    case "protocol":
    case "positive_integer": {
      if (!decimal.test(text))
        return invalid(
          "Enter a decimal integer without spaces, signs, leading zeroes, fractions, or exponents.",
        );
      const value = BigInt(text);
      const minimum = kind === "positive_integer" ? 1n : 0n;
      const maximum =
        kind === "counter"
          ? uint64Maximum
          : kind === "port"
            ? BigInt(networkFlowQueryMetadata.portMaximum)
            : kind === "protocol"
              ? BigInt(networkFlowQueryMetadata.protocolMaximum)
              : BigInt(Number.MAX_SAFE_INTEGER);
      if (value < minimum || value > maximum)
        return invalid(
          `Enter an integer from ${minimum} through ${maximum}${kind === "positive_integer" ? " (exact browser integer range)" : ""}.`,
        );
      return {
        ok: true,
        scalar: {
          value: kind === "counter" ? text : Number(value),
          canonical: text,
          order: value,
        },
      };
    }
    case "timestamp": {
      const parsed = parseQueryTimestamp(text);
      return parsed === null
        ? invalid(
            "Enter a real UTC timestamp ending in Z with at most six fractional digits.",
          )
        : {
            ok: true,
            scalar: {
              value: text,
              canonical: parsed.canonical,
              order: parsed.microseconds,
            },
          };
    }
    case "text":
      if (
        Array.from(text).length > networkFlowQueryMetadata.textMaximum ||
        Array.from(text).some((character) => {
          const code = character.codePointAt(0) ?? 0;
          return (
            code < 32 ||
            (code >= 127 && code <= 159) ||
            (code >= 0xd800 && code <= 0xdfff)
          );
        })
      )
        return invalid(
          "Use at most 256 Unicode characters without control characters.",
        );
      return {
        ok: true,
        scalar: { value: text, canonical: text, order: text },
      };
  }
}

export function compareQueryScalars(
  a: ParsedQueryScalar,
  b: ParsedQueryScalar,
): number {
  if (typeof a.order === "bigint" && typeof b.order === "bigint")
    return a.order < b.order ? -1 : a.order > b.order ? 1 : 0;
  const left = Array.from(a.canonical, (v) => v.codePointAt(0) ?? 0);
  const right = Array.from(b.canonical, (v) => v.codePointAt(0) ?? 0);
  for (let i = 0; i < Math.min(left.length, right.length); i++) {
    const delta = (left[i] ?? 0) - (right[i] ?? 0);
    if (delta !== 0) return delta;
  }
  return left.length - right.length;
}

export function parseQueryTimestamp(
  text: string,
): { canonical: string; microseconds: bigint } | null {
  const match =
    /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,6}))?Z$/u.exec(text);
  if (!match || text.startsWith("0000")) return null;
  const seconds = match[1] ?? "";
  const ms = Date.parse(seconds + "Z");
  if (
    !Number.isFinite(ms) ||
    new Date(ms).toISOString().slice(0, 19) !== seconds
  )
    return null;
  const fraction = (match[2] ?? "").replace(/0+$/u, "");
  return {
    canonical: seconds + (fraction ? "." + fraction : "") + "Z",
    microseconds: BigInt(ms) * 1000n + BigInt((match[2] ?? "").padEnd(6, "0")),
  };
}

type ParsedIP = { family: 4 | 6; bits: bigint; canonical: string };
function ipv4(text: string): bigint | null {
  const parts = text.split(".");
  if (
    parts.length !== 4 ||
    parts.some((p) => !decimal.test(p) || p.length > 3 || Number(p) > 255)
  )
    return null;
  return parts.reduce((n, p) => (n << 8n) + BigInt(p), 0n);
}
export function parseQueryIP(text: string): ParsedIP | null {
  if (!text.includes(":")) {
    const bits = ipv4(text);
    return bits === null
      ? null
      : { family: 4, bits, canonical: formatIP(4, bits) };
  }
  if (/[^0-9a-fA-F:.]/u.test(text)) return null;
  let source = text;
  if (source.includes(".")) {
    const lastColon = source.lastIndexOf(":");
    const tail = ipv4(source.slice(lastColon + 1));
    if (tail === null) return null;
    source =
      source.slice(0, lastColon + 1) +
      (tail >> 16n).toString(16) +
      ":" +
      (tail & 65535n).toString(16);
  }
  const halves = source.split("::");
  if (halves.length > 2) return null;
  const first = halves[0] ? halves[0].split(":") : [];
  const second = halves[1] ? halves[1].split(":") : [];
  const count = first.length + second.length;
  if (halves.length === 1 ? count !== 8 : count >= 8) return null;
  if ([...first, ...second].some((p) => !/^[0-9a-fA-F]{1,4}$/u.test(p)))
    return null;
  const groups = [...first, ...Array<string>(8 - count).fill("0"), ...second];
  const bits = groups.reduce((n, p) => (n << 16n) + BigInt("0x" + p), 0n);
  return { family: 6, bits, canonical: formatIP(6, bits) };
}
function formatIP(family: 4 | 6, bits: bigint): string {
  if (family === 4)
    return [24n, 16n, 8n, 0n]
      .map((shift) => String((bits >> shift) & 255n))
      .join(".");
  const groups = Array.from({ length: 8 }, (_, i) =>
    Number((bits >> BigInt((7 - i) * 16)) & 65535n),
  );
  let bestStart = -1;
  let bestLength = 1;
  for (let i = 0; i < 8; ) {
    if (groups[i] !== 0) {
      i++;
      continue;
    }
    let end = i;
    while (end < 8 && groups[end] === 0) end++;
    if (end - i > bestLength) {
      bestStart = i;
      bestLength = end - i;
    }
    i = end;
  }
  if (bestStart < 0) return groups.map((v) => v.toString(16)).join(":");
  return (
    groups
      .slice(0, bestStart)
      .map((v) => v.toString(16))
      .join(":") +
    "::" +
    groups
      .slice(bestStart + bestLength)
      .map((v) => v.toString(16))
      .join(":")
  );
}
