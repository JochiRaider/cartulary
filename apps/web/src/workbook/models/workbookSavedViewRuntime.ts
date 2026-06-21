import type { ViewContract } from "@cartulary/view-contracts";
import {
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  type WorkbookLayoutState,
  type WorkbookQueryState,
  workbookLayoutStateFromSavedViewLayoutJson,
  workbookQueryStateFromSavedViewQueryJson,
} from "./workbookQuery";
import type { SavedViewResource } from "./workbookSavedViews";
import type { WorkbookSheetRef } from "./workbookStartup";
import { knownWorkbookViewSchemaId } from "./workbookSurfaceRegistry";

export type WorkbookSavedViewIdentity = {
  readonly sheetRef: WorkbookSheetRef;
  readonly viewSchemaId: string;
};

export function upsertSavedViewList(
  current: readonly SavedViewResource[],
  savedView: SavedViewResource,
): SavedViewResource[] {
  const next = current.filter(
    (candidate) => candidate.saved_view_id !== savedView.saved_view_id,
  );
  return [...next, savedView].sort((left, right) =>
    left.display_name.localeCompare(right.display_name),
  );
}

export function removeSavedViewList(
  current: readonly SavedViewResource[],
  savedViewId: string,
): SavedViewResource[] {
  return current.filter((candidate) => candidate.saved_view_id !== savedViewId);
}

export function savedViewIdentityForSelection(
  savedView: Pick<SavedViewResource, "saved_view_id" | "view_schema_id">,
): WorkbookSavedViewIdentity {
  return {
    sheetRef: {
      kind: "saved_view",
      id: savedView.saved_view_id,
    },
    viewSchemaId: knownWorkbookViewSchemaId(savedView.view_schema_id),
  };
}

export function baseSurfaceIdentityForViewSchemaId(
  viewSchemaId: string,
): WorkbookSavedViewIdentity {
  const normalized = knownWorkbookViewSchemaId(viewSchemaId);
  return {
    sheetRef: {
      kind: "view_schema",
      id: normalized,
    },
    viewSchemaId: normalized,
  };
}

export function fallbackIdentityAfterSavedViewDelete(
  activeSheetRef: WorkbookSheetRef,
  deletedSavedView: Pick<SavedViewResource, "saved_view_id" | "view_schema_id">,
): WorkbookSavedViewIdentity | null {
  if (
    activeSheetRef.kind !== "saved_view" ||
    activeSheetRef.id !== deletedSavedView.saved_view_id
  ) {
    return null;
  }
  return baseSurfaceIdentityForViewSchemaId(deletedSavedView.view_schema_id);
}

export function savedViewQueryStateForRuntime(
  contract: ViewContract,
  savedView: Pick<SavedViewResource, "query_json">,
) {
  return workbookQueryStateFromSavedViewQueryJson(
    contract,
    savedView.query_json,
  );
}

export function savedViewConfigurationIsModified({
  contract,
  currentLayoutState = {},
  currentQueryState,
  savedView,
}: {
  readonly contract: ViewContract;
  readonly currentLayoutState?: WorkbookLayoutState | undefined;
  readonly currentQueryState: WorkbookQueryState;
  readonly savedView: SavedViewResource | null;
}): boolean {
  if (savedView === null) {
    return false;
  }
  const currentQueryJson = buildSavedViewQueryJson(contract, currentQueryState);
  const currentLayoutJson = buildSavedViewLayoutJson(
    contract,
    currentLayoutState,
  );
  const savedQueryJson = buildSavedViewQueryJson(
    contract,
    workbookQueryStateFromSavedViewQueryJson(contract, savedView.query_json),
  );
  const savedLayoutJson = buildSavedViewLayoutJson(
    contract,
    workbookLayoutStateFromSavedViewLayoutJson(contract, savedView.layout_json),
  );
  return (
    stableJSONStringify(currentQueryJson) !==
      stableJSONStringify(savedQueryJson) ||
    stableJSONStringify(currentLayoutJson) !==
      stableJSONStringify(savedLayoutJson)
  );
}

function stableJSONStringify(value: unknown): string {
  return JSON.stringify(sortObjectKeys(value));
}

function sortObjectKeys(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(sortObjectKeys);
  }
  if (!isObjectRecord(value)) {
    return value;
  }
  return Object.fromEntries(
    Object.entries(value)
      .sort(([left], [right]) => left.localeCompare(right))
      .map(([key, entry]) => [key, sortObjectKeys(entry)]),
  );
}

function isObjectRecord(value: unknown): value is Record<string, unknown> {
  return Boolean(value) && typeof value === "object" && !Array.isArray(value);
}
