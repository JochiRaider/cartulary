import {
  isPresenceRecord,
  type PresenceRecord,
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
