import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { type SheetRef, sheetRefsEqual } from "../../shared/sheetRef";
import {
  isPresenceRecord,
  type PresenceRecord,
} from "../utils/workbookPresence";

export type PresenceDisplayMode = "editing" | "focused" | "viewing" | "idle";
export type PresenceUser = {
  readonly connection_id: string;
  readonly user_id: string;
  readonly display_name: string;
  readonly mode: PresenceDisplayMode;
};
export type PresenceScope = {
  readonly users: readonly PresenceUser[];
  readonly shown: readonly PresenceUser[];
  readonly overflow: number;
};
export type WorkbookPresencePresentation = {
  readonly header: PresenceScope;
  readonly rows: ReadonlyMap<string, PresenceScope>;
  readonly cells: ReadonlyMap<string, ReadonlyMap<string, PresenceScope>>;
};

export const emptyPresenceScope: PresenceScope = {
  users: [],
  shown: [],
  overflow: 0,
};
export const emptyWorkbookPresence: WorkbookPresencePresentation = {
  header: emptyPresenceScope,
  rows: new Map(),
  cells: new Map(),
};

type Instant = { readonly secondMs: number; readonly fraction: string };
type Candidate = {
  readonly record: PresenceRecord;
  readonly observed: Instant;
  readonly name: string;
  readonly mode: PresenceDisplayMode;
};

/** UTC seconds plus decimal fraction: do not truncate nanosecond observations. */
function parseInstant(value: string): Instant | null {
  const match =
    /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d+))?(Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/iu.exec(
      value,
    );
  if (match === null) return null;
  const calendar = match[1]?.toUpperCase();
  const calendarMs = Date.parse(`${calendar}Z`);
  const secondMs = Date.parse(`${calendar}${match[3]?.toUpperCase()}`);
  if (
    !Number.isFinite(calendarMs) ||
    !Number.isFinite(secondMs) ||
    new Date(calendarMs).toISOString() !== `${calendar}.000Z`
  )
    return null;
  return { secondMs, fraction: (match[2] ?? "").replace(/0+$/u, "") };
}

function compareInstant(left: Instant, right: Instant): number {
  if (left.secondMs !== right.secondMs) return left.secondMs - right.secondMs;
  const width = Math.max(left.fraction.length, right.fraction.length);
  return compareCodePoints(
    left.fraction.padEnd(width, "0"),
    right.fraction.padEnd(width, "0"),
  );
}

function compareCodePoints(left: string, right: string): number {
  const a = Array.from(left, (value) => value.codePointAt(0) ?? 0);
  const b = Array.from(right, (value) => value.codePointAt(0) ?? 0);
  for (let index = 0; index < Math.min(a.length, b.length); index += 1) {
    const difference = (a[index] ?? 0) - (b[index] ?? 0);
    if (difference !== 0) return difference;
  }
  return a.length - b.length;
}

function compareObservation(left: Candidate, right: Candidate): number {
  return (
    compareInstant(right.observed, left.observed) ||
    compareCodePoints(left.name, right.name) ||
    compareCodePoints(left.record.connection_id, right.record.connection_id)
  );
}

function compactScope(
  candidates: readonly Candidate[],
  capacity: number,
): PresenceScope {
  if (candidates.length === 0) return emptyPresenceScope;
  const states = new Map<string, Candidate>();
  for (const candidate of candidates) {
    const record = candidate.record;
    const key = JSON.stringify([
      record.user_id,
      record.mode,
      record.record_id ?? null,
      record.field_key ?? null,
    ]);
    const current = states.get(key);
    if (current === undefined || compareObservation(candidate, current) < 0)
      states.set(key, candidate);
  }
  const priority = cartularyDesignPresentation.presence.mode_priority;
  const ordered = Array.from(states.values()).sort(
    (left, right) =>
      priority[left.mode] - priority[right.mode] ||
      compareObservation(left, right),
  );
  const users = new Map<string, PresenceUser>();
  for (const { record, mode } of ordered) {
    if (!users.has(record.user_id))
      users.set(record.user_id, {
        connection_id: record.connection_id,
        user_id: record.user_id,
        display_name: record.display_name,
        mode,
      });
  }
  const representatives = Array.from(users.values());
  return {
    users: representatives,
    shown: representatives.slice(0, capacity),
    overflow: Math.max(0, representatives.length - capacity),
  };
}

/** Separate local display projection; never use this order for wire/cache/resume identity. */
export function projectWorkbookPresence(input: {
  readonly records: Iterable<PresenceRecord>;
  readonly activeSheetRef: SheetRef;
  readonly connectionId: string | null;
  readonly nowMs: number;
}): {
  readonly presentation: WorkbookPresencePresentation;
  readonly nextExpiryAtMs: number | null;
} {
  const candidates: Candidate[] = [];
  let nextExpiryAtMs: number | null = null;
  const now = parseInstant(new Date(input.nowMs).toISOString());
  for (const record of input.records) {
    if (
      !isPresenceRecord(record) ||
      record.connection_id === input.connectionId ||
      !sheetRefsEqual(record.sheet_ref, input.activeSheetRef)
    )
      continue;
    const observed = parseInstant(record.observed_at);
    const expires = parseInstant(record.expires_at);
    if (
      observed === null ||
      expires === null ||
      now === null ||
      compareInstant(expires, now) <= 0
    )
      continue;
    const expiryMs =
      expires.secondMs +
      Number(expires.fraction.slice(0, 3).padEnd(3, "0")) +
      (expires.fraction.length > 3 ? 1 : 0);
    nextExpiryAtMs =
      nextExpiryAtMs === null ? expiryMs : Math.min(nextExpiryAtMs, expiryMs);
    candidates.push({
      record,
      observed,
      name: record.display_name.normalize("NFC"),
      mode:
        record.mode === "viewing" && record.record_id !== undefined
          ? "focused"
          : record.mode,
    });
  }
  const rowCandidates = new Map<string, Candidate[]>();
  const cellCandidates = new Map<string, Map<string, Candidate[]>>();
  if (input.activeSheetRef.kind !== "extension_workspace") {
    for (const candidate of candidates) {
      const { record } = candidate;
      if (record.record_id === undefined) continue;
      const row = rowCandidates.get(record.record_id) ?? [];
      row.push(candidate);
      rowCandidates.set(record.record_id, row);
      if (record.mode !== "editing" || record.field_key === undefined) continue;
      const fields =
        cellCandidates.get(record.record_id) ?? new Map<string, Candidate[]>();
      const cell = fields.get(record.field_key) ?? [];
      cell.push(candidate);
      fields.set(record.field_key, cell);
      cellCandidates.set(record.record_id, fields);
    }
  }
  const capacities = cartularyDesignPresentation.presence.capacities;
  return {
    presentation: {
      header: compactScope(candidates, capacities.header),
      rows: new Map(
        Array.from(rowCandidates, ([row, entries]) => [
          row,
          compactScope(entries, capacities.row),
        ]),
      ),
      cells: new Map(
        Array.from(cellCandidates, ([row, fields]) => [
          row,
          new Map(
            Array.from(fields, ([field, entries]) => [
              field,
              compactScope(entries, capacities.cell),
            ]),
          ),
        ]),
      ),
    },
    nextExpiryAtMs,
  };
}

export function presenceForRow(
  presentation: WorkbookPresencePresentation,
  recordId: string | null,
): PresenceScope {
  return recordId === null
    ? emptyPresenceScope
    : (presentation.rows.get(recordId) ?? emptyPresenceScope);
}
export function presenceForCell(
  presentation: WorkbookPresencePresentation,
  recordId: string | null,
  fieldKey: string,
): PresenceScope {
  return recordId === null
    ? emptyPresenceScope
    : (presentation.cells.get(recordId)?.get(fieldKey) ?? emptyPresenceScope);
}

export function samePresencePresentation(
  left: WorkbookPresencePresentation,
  right: WorkbookPresencePresentation,
): boolean {
  const sameScope = (a: PresenceScope, b: PresenceScope) =>
    a.users.length === b.users.length &&
    a.users.every((user, index) => {
      const other = b.users[index];
      return (
        other !== undefined &&
        user.connection_id === other.connection_id &&
        user.user_id === other.user_id &&
        user.display_name === other.display_name &&
        user.mode === other.mode
      );
    });
  return (
    sameScope(left.header, right.header) &&
    left.rows.size === right.rows.size &&
    left.cells.size === right.cells.size &&
    Array.from(left.rows).every(([row, scope]) =>
      sameScope(scope, right.rows.get(row) ?? emptyPresenceScope),
    ) &&
    Array.from(left.cells).every(([row, fields]) => {
      const otherFields = right.cells.get(row);
      return (
        otherFields !== undefined &&
        fields.size === otherFields.size &&
        Array.from(fields).every(([field, scope]) =>
          sameScope(scope, otherFields.get(field) ?? emptyPresenceScope),
        )
      );
    })
  );
}
