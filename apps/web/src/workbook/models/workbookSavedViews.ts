import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import { validateDisplayName } from "../../shared/displayName";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  workbookLayoutStateFromSavedViewLayoutJson,
  workbookQueryStateFromSavedViewQueryJson,
} from "./workbookQuery";
import { isStandardizedWorkbookViewSchemaId } from "./workbookSurfaceRegistry";

export type SavedViewResource = {
  incident_id: string;
  created_at: string;
  updated_at: string;
  saved_view_id: string;
  view_schema_id: string;
  display_name: string;
  scope: "private" | "shared" | "system";
  query_json: unknown;
  layout_json: unknown;
  owner_user_id: string | null;
  saved_view_version: number;
};

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

export function normalizeSavedViewResource(
  value: unknown,
): SavedViewResource | null {
  if (!isRecord(value)) {
    return null;
  }
  const record = value;
  if (
    typeof record.saved_view_id !== "string" ||
    typeof record.view_schema_id !== "string" ||
    typeof record.display_name !== "string" ||
    !isStandardizedWorkbookViewSchemaId(record.view_schema_id)
  ) {
    return null;
  }
  const scope = normalizeSavedViewScope(record.scope);
  const name = validateDisplayName(record.display_name);
  if (
    scope === null ||
    name.error !== null ||
    name.value !== record.display_name ||
    typeof record.incident_id !== "string" ||
    record.incident_id === "" ||
    typeof record.created_at !== "string" ||
    !Number.isFinite(Date.parse(record.created_at)) ||
    typeof record.updated_at !== "string" ||
    !Number.isFinite(Date.parse(record.updated_at)) ||
    (record.owner_user_id !== null &&
      typeof record.owner_user_id !== "string") ||
    (scope !== "system" &&
      (typeof record.owner_user_id !== "string" || record.owner_user_id === ""))
  ) {
    return null;
  }
  const version =
    typeof record.saved_view_version === "number" &&
    Number.isSafeInteger(record.saved_view_version)
      ? record.saved_view_version
      : 0;
  const contract = requireViewContract(record.view_schema_id);
  if (
    version < 1 ||
    !isRecord(record.query_json) ||
    !isRecord(record.layout_json) ||
    !savedViewJSONEqual(
      record.query_json,
      savedViewQueryJsonForPersistence(contract, record.query_json),
    ) ||
    !savedViewJSONEqual(
      record.layout_json,
      savedViewLayoutJsonForPersistence(contract, record.layout_json),
    )
  )
    return null;
  return {
    incident_id: record.incident_id,
    created_at: record.created_at,
    updated_at: record.updated_at,
    saved_view_id: record.saved_view_id,
    view_schema_id: record.view_schema_id,
    display_name: record.display_name,
    scope,
    query_json: structuredClone(record.query_json),
    layout_json: structuredClone(record.layout_json),
    owner_user_id:
      typeof record.owner_user_id === "string" ? record.owner_user_id : null,
    saved_view_version: version,
  };
}

/** Structural JSON equality preserves array order and ignores object member order. */
export function savedViewJSONEqual(left: unknown, right: unknown): boolean {
  if (left === right) return true;
  if (Array.isArray(left) && Array.isArray(right)) {
    return (
      left.length === right.length &&
      left.every((entry, index) => savedViewJSONEqual(entry, right[index]))
    );
  }
  if (!isRecord(left) || !isRecord(right)) return false;
  const keys = Object.keys(left);
  return (
    keys.length === Object.keys(right).length &&
    keys.every(
      (key) =>
        Object.hasOwn(right, key) && savedViewJSONEqual(left[key], right[key]),
    )
  );
}

function normalizeSavedViewScope(
  value: unknown,
): SavedViewResource["scope"] | null {
  return value === "private" || value === "shared" || value === "system"
    ? value
    : null;
}

export function canMutateSavedView(
  savedView: SavedViewResource | null,
  currentUserId: string | null,
  currentIncidentRole: string | null,
): boolean {
  if (savedView === null || savedView.scope === "system") {
    return false;
  }
  if (currentIncidentRole === "admin") {
    return true;
  }
  return (
    currentUserId !== null &&
    savedView.owner_user_id !== null &&
    savedView.owner_user_id === currentUserId
  );
}

export function savedViewQueryJsonForPersistence(
  contract: ViewContract,
  value: unknown,
) {
  return buildSavedViewQueryJson(
    contract,
    workbookQueryStateFromSavedViewQueryJson(contract, value),
  );
}

export function savedViewLayoutJsonForPersistence(
  contract: ViewContract,
  value: unknown,
) {
  return buildSavedViewLayoutJson(
    contract,
    workbookLayoutStateFromSavedViewLayoutJson(contract, value),
  );
}
