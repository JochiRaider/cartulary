import {
  coreIndicatorTypes,
  type QueryWorkbookViewRequest,
  type QueryWorkbookViewResponse,
} from "@cartulary/protocol-ts/http";
import {
  indicatorsViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import { fetchHTTPOperation } from "./browserApi";

export function indicatorTargetChange(message: {
  readonly type: string;
  readonly payload?: unknown;
}): {
  readonly id: string;
  readonly removed: boolean;
  readonly clientTxnId: string;
} | null {
  if (
    message.type !== "record_changed" ||
    message.payload === null ||
    typeof message.payload !== "object"
  )
    return null;
  const payload = message.payload as {
    readonly record_id?: unknown;
    readonly client_txn_id?: unknown;
    readonly affected_views?: readonly {
      readonly view_schema_id: string;
      readonly change_kind: string;
    }[];
  };
  const change = Array.isArray(payload.affected_views)
    ? payload.affected_views.find(
        (view) => view.view_schema_id === indicatorsViewSchemaId,
      )
    : undefined;
  return change !== undefined &&
    typeof payload.record_id === "string" &&
    typeof payload.client_txn_id === "string"
    ? {
        id: payload.record_id,
        removed: change.change_kind === "remove",
        clientTxnId: payload.client_txn_id,
      }
    : null;
}

export type IndicatorLinkTargetOption = {
  readonly id: string;
  readonly value: string;
};
export type IndicatorLinkTargetPage = {
  readonly items: readonly IndicatorLinkTargetOption[];
  readonly nextCursor: string | null;
};
export type IndicatorLinkTargetQuery = {
  readonly incidentId: string;
  readonly candidateValue: string;
  readonly cursor: string | null;
};
export type IndicatorLinkTargetPort = (
  query: IndicatorLinkTargetQuery,
  signal: AbortSignal,
) => Promise<IndicatorLinkTargetPage>;
export type CoreAtomicIPType = Extract<
  (typeof coreIndicatorTypes)[number],
  "ipv4_addr" | "ipv6_addr"
>;

/** Core registry adapter: callers supply canonical endpoint bytes, never labels. */
export function coreAtomicIPType(value: string): CoreAtomicIPType | null {
  const family = canonicalIPFamily(value);
  const token = family === 4 ? "ipv4_addr" : family === 6 ? "ipv6_addr" : null;
  return (
    coreIndicatorTypes.find(
      (kind): kind is CoreAtomicIPType => kind === token,
    ) ?? null
  );
}

export function compatibleAtomicIPTarget(
  target: {
    readonly indicator_type: string;
    readonly value_kind: string;
    readonly normalized_value: string | null;
  },
  candidate: string,
): boolean {
  const type = coreAtomicIPType(candidate);
  return (
    type !== null &&
    target.indicator_type === type &&
    target.value_kind === "atomic" &&
    target.normalized_value === candidate
  );
}

export async function queryCompatibleIndicatorTargets(
  apiBase: string | undefined,
  query: IndicatorLinkTargetQuery,
  signal: AbortSignal,
): Promise<IndicatorLinkTargetPage> {
  const type = coreAtomicIPType(query.candidateValue);
  const contract = requireViewContract(indicatorsViewSchemaId);
  if (
    type === null ||
    !["indicator.indicator_type", "indicator.value_kind"].every((key) =>
      contract.fieldMap[key]?.filterOps.includes("eq"),
    )
  )
    throw new Error("indicator_discovery_contract_unavailable");
  const request: QueryWorkbookViewRequest =
    query.cursor === null
      ? {
          limit: 100,
          filters: [
            {
              field_key: "indicator.indicator_type",
              op: "eq",
              arg: { value: type },
            },
            {
              field_key: "indicator.value_kind",
              op: "eq",
              arg: { value: "atomic" },
            },
          ],
        }
      : { limit: 100, cursor_token: query.cursor };
  signal.throwIfAborted();
  const result = await fetchHTTPOperation<QueryWorkbookViewResponse>({
    apiBase,
    operationID: "queryWorkbookView",
    pathParameters: {
      incident_id: query.incidentId,
      view_schema_id: indicatorsViewSchemaId,
    },
    init: { method: "POST", body: JSON.stringify(request), signal },
  });
  if (!result.ok) throw new Error("indicator_discovery_failed");
  const { data, meta } = result.payload;
  const paging = meta.paging;
  if (
    data.incident_id !== query.incidentId ||
    data.view_schema_id !== indicatorsViewSchemaId ||
    data.rows.length > 100 ||
    paging === undefined ||
    paging.limit !== 100 ||
    paging.has_more !== (paging.next_cursor !== null) ||
    (paging.next_cursor !== null && paging.next_cursor.length === 0) ||
    new Set(data.rows.map((row) => row.record_id)).size !== data.rows.length
  )
    throw new Error("invalid_indicator_discovery_page");
  const items = [];
  for (const row of data.rows) {
    const type = row.cells["indicator.indicator_type"]?.value;
    const kind = row.cells["indicator.value_kind"]?.value;
    const normalized = row.cells["indicator.normalized_value"]?.value;
    if (
      typeof type !== "string" ||
      typeof kind !== "string" ||
      (normalized !== null && typeof normalized !== "string")
    )
      throw new Error("invalid_indicator_discovery_row");
    if (
      compatibleAtomicIPTarget(
        {
          indicator_type: type,
          value_kind: kind,
          normalized_value: normalized,
        },
        query.candidateValue,
      )
    )
      items.push({ id: row.record_id, value: query.candidateValue });
  }
  return { items, nextCursor: paging.next_cursor };
}

// Exact Core IP spelling: IPv4 dotted decimal and RFC 5952 IPv6, excluding
// zones, mapped/coerced IPv4, dotted suffixes, brackets, CIDRs and defanging.
function canonicalIPFamily(value: string): 4 | 6 | null {
  if (/^(?:0|[1-9]\d{0,2})(?:\.(?:0|[1-9]\d{0,2})){3}$/u.test(value)) {
    return value.split(".").every((part) => Number(part) <= 255) ? 4 : null;
  }
  if (!/^[0-9a-f:]+$/u.test(value) || !value.includes(":")) return null;
  const halves = value.split("::");
  if (halves.length > 2) return null;
  const left = halves[0] === "" ? [] : (halves[0]?.split(":") ?? []);
  const right =
    halves.length === 1 || halves[1] === ""
      ? []
      : (halves[1]?.split(":") ?? []);
  if (
    ![...left, ...right].every((part) =>
      /^(?:0|[1-9a-f][0-9a-f]{0,3})$/u.test(part),
    )
  )
    return null;
  const missing = 8 - left.length - right.length;
  if (halves.length === 1 ? missing !== 0 : missing < 1) return null;
  const words = [
    ...left,
    ...Array<string>(halves.length === 2 ? missing : 0).fill("0"),
    ...right,
  ];
  if (words.slice(0, 5).every((word) => word === "0") && words[5] === "ffff")
    return null;
  let bestStart = -1;
  let bestLength = 1;
  for (let index = 0; index < words.length; ) {
    if (words[index] !== "0") {
      index++;
      continue;
    }
    const start = index;
    while (words[index] === "0") index++;
    if (index - start > bestLength) {
      bestStart = start;
      bestLength = index - start;
    }
  }
  const canonical =
    bestStart < 0
      ? words.join(":")
      : `${words.slice(0, bestStart).join(":")}::${words.slice(bestStart + bestLength).join(":")}`;
  return canonical === value ? 6 : null;
}
