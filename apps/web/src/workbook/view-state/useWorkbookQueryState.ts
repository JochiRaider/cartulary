import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useReducer,
} from "react";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";

export type WorkbookQueryStateEntry = {
  readonly filterDraft: FilterDraft;
  readonly queryState: WorkbookQueryState;
};

type WorkbookQueryStateStore = {
  readonly entries: Readonly<Record<string, WorkbookQueryStateEntry>>;
};

type WorkbookQueryStateAction =
  | {
      readonly type: "ensure";
      readonly viewSchemaIds: readonly string[];
    }
  | {
      readonly type: "update";
      readonly update: (
        current: WorkbookQueryStateEntry,
      ) => WorkbookQueryStateEntry;
      readonly viewSchemaId: string;
    }
  | {
      readonly type: "reset";
      readonly viewSchemaId: string;
    };

function newQueryEntry(
  viewSchemaId: string,
  contract: ViewContract = workbookContractForViewSchemaId(viewSchemaId),
): WorkbookQueryStateEntry {
  return {
    filterDraft: defaultFilterDraft(contract),
    queryState: emptyWorkbookQueryState(),
  };
}

function initializeWorkbookQueryState(
  viewSchemaIds: readonly string[],
): WorkbookQueryStateStore {
  return {
    entries: Object.fromEntries(
      [...new Set(viewSchemaIds)].map((viewSchemaId) => [
        viewSchemaId,
        newQueryEntry(viewSchemaId),
      ]),
    ),
  };
}

function workbookQueryStateReducer(
  current: WorkbookQueryStateStore,
  action: WorkbookQueryStateAction,
): WorkbookQueryStateStore {
  if (action.type === "ensure") {
    const missing = action.viewSchemaIds.filter(
      (viewSchemaId) => current.entries[viewSchemaId] === undefined,
    );
    if (missing.length === 0) {
      return current;
    }
    return {
      entries: {
        ...current.entries,
        ...Object.fromEntries(
          missing.map((viewSchemaId) => [
            viewSchemaId,
            newQueryEntry(viewSchemaId),
          ]),
        ),
      },
    };
  }
  if (action.type === "reset") {
    return {
      entries: {
        ...current.entries,
        [action.viewSchemaId]: newQueryEntry(action.viewSchemaId),
      },
    };
  }
  const previous =
    current.entries[action.viewSchemaId] ?? newQueryEntry(action.viewSchemaId);
  return {
    entries: {
      ...current.entries,
      [action.viewSchemaId]: action.update(previous),
    },
  };
}

function resolveStateAction<T>(current: T, action: SetStateAction<T>): T {
  return typeof action === "function"
    ? (action as (current: T) => T)(current)
    : action;
}

export function useWorkbookQueryState(viewSchemaIds: readonly string[]) {
  const [store, dispatch] = useReducer(
    workbookQueryStateReducer,
    viewSchemaIds,
    initializeWorkbookQueryState,
  );

  useEffect(() => {
    dispatch({ type: "ensure", viewSchemaIds });
  }, [viewSchemaIds]);

  const entryFor = useCallback(
    (viewSchemaId: string) =>
      store.entries[viewSchemaId] ?? newQueryEntry(viewSchemaId),
    [store.entries],
  );

  const updateEntry = useCallback(
    (
      viewSchemaId: string,
      update: (current: WorkbookQueryStateEntry) => WorkbookQueryStateEntry,
    ) => {
      dispatch({ type: "update", update, viewSchemaId });
    },
    [],
  );

  const resetEntry = useCallback((viewSchemaId: string) => {
    dispatch({ type: "reset", viewSchemaId });
  }, []);

  const setQueryStateForSurface = useCallback(
    (viewSchemaId: string, action: SetStateAction<WorkbookQueryState>) => {
      updateEntry(viewSchemaId, (current) => ({
        ...current,
        queryState: resolveStateAction(current.queryState, action),
      }));
    },
    [updateEntry],
  );

  const setFilterDraftForSurface = useCallback(
    (viewSchemaId: string, action: SetStateAction<FilterDraft>) => {
      updateEntry(viewSchemaId, (current) => ({
        ...current,
        filterDraft: resolveStateAction(current.filterDraft, action),
      }));
    },
    [updateEntry],
  );

  return {
    entryFor,
    resetEntry,
    setFilterDraftForSurface,
    setQueryStateForSurface,
    updateEntry,
  };
}

export type WorkbookQueryStateSetter = Dispatch<
  SetStateAction<WorkbookQueryState>
>;
