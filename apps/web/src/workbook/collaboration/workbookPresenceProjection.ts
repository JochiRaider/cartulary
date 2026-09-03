import type { SheetRef } from "../../shared/sheetRef";
import {
  isPresenceRecord,
  type PresenceRecord,
  presenceMatchesSheet,
} from "../utils/workbookPresence";

export type WorkbookPresenceProjection = ReadonlyMap<string, PresenceRecord>;

export function initialWorkbookPresenceProjection(): WorkbookPresenceProjection {
  return new Map();
}

function copiedPresence(presence: PresenceRecord): PresenceRecord {
  return { ...presence, sheet_ref: { ...presence.sheet_ref } };
}

function recordValue(value: unknown): Record<string, unknown> | null {
  return value !== null && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : null;
}

export function replaceWorkbookPresenceSnapshot(
  current: WorkbookPresenceProjection,
  payload: unknown,
): WorkbookPresenceProjection {
  const record = recordValue(payload);
  if (record === null || !Array.isArray(record.presences)) return current;
  const next = new Map<string, PresenceRecord>();
  for (const candidate of record.presences) {
    if (!isPresenceRecord(candidate) || next.has(candidate.connection_id)) {
      return current;
    }
    next.set(candidate.connection_id, copiedPresence(candidate));
  }
  return next;
}

export function applyWorkbookPresenceDelta(
  current: WorkbookPresenceProjection,
  payload: unknown,
): WorkbookPresenceProjection {
  const record = recordValue(payload);
  const presence = recordValue(record?.presence);
  const connectionId = presence?.connection_id;
  if (record === null || typeof connectionId !== "string") return current;
  if (record.delta_kind === "remove") {
    if (!current.has(connectionId)) return current;
    const next = new Map(current);
    next.delete(connectionId);
    return next;
  }
  if (record.delta_kind !== "upsert" || !isPresenceRecord(presence)) {
    return current;
  }
  const next = new Map(current);
  next.set(connectionId, copiedPresence(presence));
  return next;
}

export function clearWorkbookPresenceProjection(
  current: WorkbookPresenceProjection,
): WorkbookPresenceProjection {
  return current.size === 0 ? current : new Map();
}

export function activeWorkbookPresenceRecords(input: {
  readonly activeSheetRef: SheetRef;
  readonly connectionId: string | null;
  readonly projection: WorkbookPresenceProjection;
}): readonly PresenceRecord[] {
  return Array.from(input.projection.values())
    .filter((presence) => presenceMatchesSheet(presence, input.activeSheetRef))
    .filter((presence) => presence.connection_id !== input.connectionId)
    .sort((left, right) => {
      if (left.display_name < right.display_name) return -1;
      if (left.display_name > right.display_name) return 1;
      if (left.connection_id < right.connection_id) return -1;
      if (left.connection_id > right.connection_id) return 1;
      return 0;
    });
}
