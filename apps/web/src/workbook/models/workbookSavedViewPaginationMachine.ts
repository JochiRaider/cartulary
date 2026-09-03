import {
  normalizeSavedViewResource,
  type SavedViewResource,
} from "./workbookSavedViews";

export type WorkbookSavedViewPaginationMachine = {
  readonly nextCursor: string | null;
  readonly savedViews: readonly SavedViewResource[];
  readonly seenCursorTokens: readonly string[];
  readonly seenSavedViewIds: readonly string[];
  readonly subjectGeneration: number;
};

export type WorkbookSavedViewPagePlan =
  | {
      readonly kind: "continue";
      readonly machine: WorkbookSavedViewPaginationMachine;
    }
  | {
      readonly kind: "complete";
      readonly savedViews: readonly SavedViewResource[];
    }
  | { readonly kind: "invalid"; readonly message: string };

export function startWorkbookSavedViewPagination(
  subjectGeneration: number,
): WorkbookSavedViewPaginationMachine {
  return {
    nextCursor: null,
    savedViews: [],
    seenCursorTokens: [],
    seenSavedViewIds: [],
    subjectGeneration,
  };
}

export function workbookSavedViewPaginationIsCurrent(
  machine: WorkbookSavedViewPaginationMachine,
  currentSubjectGeneration: number,
): boolean {
  return machine.subjectGeneration === currentSubjectGeneration;
}

export function normalizeWorkbookSavedViewPage(input: {
  readonly incidentId: string;
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
    paging.limit !== input.limit ||
    (paging.has_more &&
      (paging.next_cursor === null || paging.next_cursor.trim() === "")) ||
    (!paging.has_more && paging.next_cursor !== null)
  ) {
    return null;
  }
  const savedViews: SavedViewResource[] = [];
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
    if (savedView === null || savedView.saved_view_version < 1) {
      return null;
    }
    savedViews.push(savedView);
  }
  return {
    nextCursor: paging.has_more ? paging.next_cursor : null,
    savedViews,
  };
}

export function acceptWorkbookSavedViewPage(
  machine: WorkbookSavedViewPaginationMachine,
  page: {
    readonly nextCursor: string | null;
    readonly savedViews: readonly SavedViewResource[];
  },
): WorkbookSavedViewPagePlan {
  const savedViews = [...machine.savedViews];
  const seenSavedViewIds = new Set(machine.seenSavedViewIds);
  for (const candidate of page.savedViews) {
    const savedView = normalizeSavedViewResource(candidate);
    if (savedView === null) {
      return {
        kind: "invalid",
        message: "Saved-view listing returned an invalid resource.",
      };
    }
    if (seenSavedViewIds.has(savedView.saved_view_id)) {
      return {
        kind: "invalid",
        message: "Saved-view listing returned a duplicate resource.",
      };
    }
    seenSavedViewIds.add(savedView.saved_view_id);
    savedViews.push(savedView);
  }
  if (page.nextCursor === null) {
    return { kind: "complete", savedViews };
  }
  if (
    typeof page.nextCursor !== "string" ||
    page.nextCursor.trim() === "" ||
    machine.seenCursorTokens.includes(page.nextCursor)
  ) {
    return {
      kind: "invalid",
      message: "Saved-view listing returned a cyclic cursor.",
    };
  }
  return {
    kind: "continue",
    machine: {
      ...machine,
      nextCursor: page.nextCursor,
      savedViews,
      seenCursorTokens: [...machine.seenCursorTokens, page.nextCursor],
      seenSavedViewIds: [...seenSavedViewIds],
    },
  };
}
