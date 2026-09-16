import {
  normalizeSavedViewResource,
  type SavedViewResource,
} from "./workbookSavedViews";

export function normalizeWorkbookSavedViewPage(input: {
  readonly incidentId: string;
  readonly viewSchemaId?: string | undefined;
  readonly limit: number;
  readonly paging:
    | {
        readonly has_more: boolean;
        readonly limit: number;
        readonly next_cursor: string | null;
      }
    | undefined;
  readonly savedViews: readonly unknown[];
}): {
  readonly nextCursor: string | null;
  readonly savedViews: readonly SavedViewResource[];
} | null {
  const paging = input.paging;
  if (
    paging === undefined ||
    !Number.isSafeInteger(input.limit) ||
    input.limit < 1 ||
    input.limit > 500 ||
    input.savedViews.length > input.limit ||
    paging.limit !== input.limit ||
    (paging.has_more &&
      (input.savedViews.length === 0 ||
        paging.next_cursor === null ||
        paging.next_cursor.trim() === "")) ||
    (!paging.has_more && paging.next_cursor !== null)
  ) {
    return null;
  }
  const savedViews: SavedViewResource[] = [];
  const ids = new Set<string>();
  for (const candidate of input.savedViews) {
    if (
      candidate === null ||
      typeof candidate !== "object" ||
      Array.isArray(candidate) ||
      !("incident_id" in candidate) ||
      candidate.incident_id !== input.incidentId
    ) {
      return null;
    }
    const savedView = normalizeSavedViewResource(candidate);
    if (
      savedView === null ||
      savedView.saved_view_version < 1 ||
      ids.has(savedView.saved_view_id) ||
      (input.viewSchemaId !== undefined &&
        savedView.view_schema_id !== input.viewSchemaId)
    ) {
      return null;
    }
    ids.add(savedView.saved_view_id);
    savedViews.push(savedView);
  }
  return {
    nextCursor: paging.has_more ? paging.next_cursor : null,
    savedViews,
  };
}
