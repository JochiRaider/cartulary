import type { ViewContract } from "@cartulary/view-contracts";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  workbookLayoutStateFromSavedViewLayoutJson,
  workbookQueryStateFromSavedViewQueryJson,
} from "./workbookQuery";
import { isStandardizedWorkbookViewSchemaId } from "./workbookSurfaceRegistry";

export type SavedViewResource = {
  saved_view_id: string;
  view_schema_id: string;
  display_name: string;
  scope: "private" | "shared" | "system";
  query_json: unknown;
  layout_json: unknown;
  owner_user_id: string | null;
  saved_view_version: number;
};

export type SavedViewListEnvelope = {
  data: {
    saved_views: SavedViewResource[];
  };
  meta?: {
    paging?: {
      has_more?: boolean;
      next_cursor?: string | null;
    };
  };
};

export type SavedViewEnvelope = {
  data: SavedViewResource;
};

export function normalizeSavedViewResource(
  value: unknown,
): SavedViewResource | null {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    return null;
  }
  const record = value as Record<string, unknown>;
  if (
    typeof record.saved_view_id !== "string" ||
    typeof record.view_schema_id !== "string" ||
    typeof record.display_name !== "string" ||
    !isStandardizedWorkbookViewSchemaId(record.view_schema_id)
  ) {
    return null;
  }
  const scope = normalizeSavedViewScope(record.scope);
  if (scope === null) {
    return null;
  }
  const version =
    typeof record.saved_view_version === "number" &&
    Number.isSafeInteger(record.saved_view_version)
      ? record.saved_view_version
      : 0;
  return {
    saved_view_id: record.saved_view_id,
    view_schema_id: record.view_schema_id,
    display_name: record.display_name,
    scope,
    query_json: record.query_json ?? {},
    layout_json: record.layout_json ?? {},
    owner_user_id:
      typeof record.owner_user_id === "string" ? record.owner_user_id : null,
    saved_view_version: version,
  };
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
