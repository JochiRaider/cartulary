import type { SameFieldConflictPayload } from "./workbookTimelineModel";

export type SameFieldConflictFields = Record<string, unknown> & {
  base_row_version: number;
  conflict_resolution_class: string;
  conflict_token: string;
  current_row_version: number;
  field_key: string;
  record_id: string;
};

export function parseSameFieldConflictFields(
  conflict: unknown,
): SameFieldConflictFields | null {
  if (!conflict || typeof conflict !== "object") {
    return null;
  }
  const object = conflict as Record<string, unknown>;
  if (
    typeof object.conflict_token !== "string" ||
    object.conflict_token.trim() === "" ||
    typeof object.record_id !== "string" ||
    object.record_id.trim() === "" ||
    typeof object.field_key !== "string" ||
    object.field_key.trim() === "" ||
    typeof object.conflict_resolution_class !== "string" ||
    object.conflict_resolution_class.trim() === "" ||
    typeof object.base_row_version !== "number" ||
    typeof object.current_row_version !== "number"
  ) {
    return null;
  }
  return object as SameFieldConflictFields;
}

export function parseSameFieldConflict(
  payload: unknown,
): SameFieldConflictPayload | null {
  if (!payload || typeof payload !== "object" || !("error" in payload)) {
    return null;
  }
  const error = payload.error;
  if (
    !error ||
    typeof error !== "object" ||
    !("code" in error) ||
    error.code !== "same_field_conflict" ||
    !("conflict" in error)
  ) {
    return null;
  }
  const object = parseSameFieldConflictFields(error.conflict);
  if (object === null) {
    return null;
  }
  const parsed: SameFieldConflictPayload = {
    conflict_token: object.conflict_token,
    record_id: object.record_id,
    field_key: object.field_key,
    conflict_resolution_class: object.conflict_resolution_class,
    base_row_version: object.base_row_version,
    current_row_version: object.current_row_version,
    client_value: object.client_value,
    server_value: object.server_value,
    base_value: object.base_value,
    suggested_merged_value: object.suggested_merged_value,
  };
  if (typeof object.server_updated_by === "string") {
    parsed.server_updated_by = object.server_updated_by;
  }
  if (typeof object.server_updated_at === "string") {
    parsed.server_updated_at = object.server_updated_at;
  }
  return parsed;
}
