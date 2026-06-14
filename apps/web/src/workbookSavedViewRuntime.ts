import type { ViewContract } from "@cartulary/view-contracts";
import { workbookQueryStateFromSavedViewQueryJson } from "./workbookQuery";
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
