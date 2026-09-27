import type { WorkbookRow } from "./timelineRowModel";

export type TimelineEvidenceLink = {
  readonly itemRef: string;
  readonly evidenceRecordId: string;
  readonly title: string;
};

export type TimelineEvidenceLinks =
  | { readonly kind: "unavailable" }
  | { readonly kind: "empty"; readonly items: readonly [] }
  | {
      readonly kind: "available";
      readonly items: readonly TimelineEvidenceLink[];
    };

const observationTokens = new WeakMap<object, number>();
let nextObservationToken = 0;
function observationToken(row: WorkbookRow): number | null {
  if (!row.rawRow) return null;
  let token = observationTokens.get(row.rawRow);
  if (token === undefined) {
    token = ++nextObservationToken;
    observationTokens.set(row.rawRow, token);
  }
  return token;
}

function object(value: unknown): Record<string, unknown> | null {
  return value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

/** A link's returned target identity is independent of its item_ref and label. */
export function readTimelineEvidenceLinks(
  row: WorkbookRow,
): TimelineEvidenceLinks {
  const value = object(
    row.rawRow?.cells["timeline.attached_evidence_ids"]?.value,
  );
  if (!value || !Array.isArray(value.items)) return { kind: "unavailable" };
  if (value.items.length === 0) return { kind: "empty", items: [] };
  const items: TimelineEvidenceLink[] = [];
  const refs = new Set<string>();
  const targets = new Set<string>();
  for (const entry of value.items) {
    const item = object(entry);
    const itemRef = item?.item_ref;
    const target = item?.linked_record_id;
    const label = item?.display_text;
    if (
      item?.item_kind !== "record_ref" ||
      typeof itemRef !== "string" ||
      itemRef.trim() === "" ||
      typeof target !== "string" ||
      target.trim() === "" ||
      typeof label !== "string" ||
      label.trim() === "" ||
      refs.has(itemRef) ||
      targets.has(target)
    )
      return { kind: "unavailable" };
    refs.add(itemRef);
    targets.add(target);
    items.push({ itemRef, evidenceRecordId: target, title: label });
  }
  return { kind: "available", items };
}

export function timelineEvidenceLinkKey(row: WorkbookRow): string {
  const links = readTimelineEvidenceLinks(row);
  return JSON.stringify([
    row.recordId,
    row.rowVersion,
    row.rawRow?.observation?.scope,
    observationToken(row),
    links,
  ]);
}
