import type { WorkbookSurface } from "@cartulary/ui-contracts";
import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useEffect,
  useMemo,
  useState,
} from "react";
import { apiPath } from "../../services/browserApi";
import {
  fetchJSON,
  handleWorkbookLoadFailure,
  parseErrorMessage,
  readEnvelope,
} from "../../services/workbookApi";
import {
  applyFilterDraft,
  buildSavedViewLayoutJson,
  buildSavedViewQueryJson,
  defaultFilterDraft,
  emptyWorkbookQueryState,
  type FilterDraft,
  removeFilterField,
  toggleSortField,
  updateGroupBy,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import {
  baseSurfaceIdentityForViewSchemaId,
  fallbackIdentityAfterSavedViewDelete,
  removeSavedViewList,
  savedViewConfigurationIsModified,
  savedViewIdentityForSelection,
  savedViewQueryStateForRuntime,
  upsertSavedViewList,
} from "../models/workbookSavedViewRuntime";
import {
  normalizeSavedViewResource,
  type SavedViewEnvelope,
  type SavedViewListEnvelope,
  type SavedViewResource,
  savedViewLayoutJsonForPersistence,
  savedViewQueryJsonForPersistence,
} from "../models/workbookSavedViews";
import {
  normalizeWorkbookStartupSelection,
  type WorkbookSheetRef,
  workbookStartupQueryFromURLParams,
} from "../models/workbookStartup";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  knownWorkbookViewSchemaId,
  listWorkbookSurfaceRegistryEntries,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";

type WorkbookShellMutableRef<T> = {
  current: T;
};

type FilterDraftSetter = Dispatch<SetStateAction<FilterDraft>>;
type WorkbookQueryStateSetter = Dispatch<SetStateAction<WorkbookQueryState>>;

export type WorkbookShellActiveQueryControls = {
  readonly contract: ViewContract;
  readonly filterDraft: FilterDraft;
  readonly onApplyFilter: (draft: FilterDraft) => void;
  readonly onClearAll: () => void;
  readonly onFilterDraftChange: FilterDraftSetter;
  readonly onGroupByChange: (groupBy: string | null) => void;
  readonly onRemoveFilter: (fieldKey: string) => void;
  readonly onToggleSort: (fieldKey: string) => void;
  readonly queryState: WorkbookQueryState;
  readonly surface: WorkbookSurface;
};

type WorkbookStartupEnvelope = {
  data?: unknown;
};

const timelineContract = requireViewContract(timelineViewSchemaId);
const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);
const allWorkbookContracts = listWorkbookSurfaceRegistryEntries().map(
  (entry) => entry.contract,
);

function workbookContractForViewSchemaId(viewSchemaId: string): ViewContract {
  return (
    allWorkbookContracts.find(
      (contract) => contract.viewSchemaId === viewSchemaId,
    ) ?? timelineContract
  );
}

export function useWorkbookShellRuntime({
  apiBase,
  incidentId,
  onIncidentAccessLost,
  surfaceSelectionVersionRef,
}: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly surfaceSelectionVersionRef: WorkbookShellMutableRef<number>;
}) {
  const params = useMemo(() => new URLSearchParams(window.location.search), []);
  const initialViewSchemaId = useMemo(() => {
    const explicit = params.get("view_schema_id");
    return explicit
      ? knownWorkbookViewSchemaId(explicit)
      : timelineViewSchemaId;
  }, [params]);
  const [surface, setSurface] = useState(initialViewSchemaId);
  const [startupSheetRef, setStartupSheetRef] = useState<WorkbookSheetRef>(
    () => ({ kind: "view_schema", id: initialViewSchemaId }),
  );
  const [sheetReloadToken, setSheetReloadToken] = useState(0);
  const [pendingGridFocusSurface, setPendingGridFocusSurface] = useState<
    string | null
  >(null);
  const [savedViews, setSavedViews] = useState<SavedViewResource[]>([]);
  const [timelineQueryState, setTimelineQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [timelineFilterDraft, setTimelineFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(timelineContract),
  );
  const [hostQueryState, setHostQueryState] = useState<WorkbookQueryState>(() =>
    emptyWorkbookQueryState(),
  );
  const [identityQueryState, setIdentityQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [hostFilterDraft, setHostFilterDraft] = useState<FilterDraft>(() =>
    defaultFilterDraft(hostsContract),
  );
  const [identityFilterDraft, setIdentityFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(identitiesContract),
  );
  const [assessmentQueryState, setAssessmentQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const [assessmentFilterDraft, setAssessmentFilterDraft] =
    useState<FilterDraft>(() => defaultFilterDraft(assessmentsContract));
  const [genericQueryState, setGenericQueryState] =
    useState<WorkbookQueryState>(() => emptyWorkbookQueryState());
  const activeContract = useMemo(
    () => workbookContractForViewSchemaId(surface),
    [surface],
  );
  const [genericFilterDraft, setGenericFilterDraft] = useState<FilterDraft>(
    () => defaultFilterDraft(activeContract),
  );

  const applyWorkbookIdentity = useCallback(
    (
      identity: {
        readonly sheetRef: WorkbookSheetRef;
        readonly viewSchemaId: string;
      },
      options: {
        readonly bumpSelectionVersion?: boolean;
        readonly focusFirstGridTarget?: boolean;
        readonly reloadSheet?: boolean;
      } = {},
    ) => {
      if (options.bumpSelectionVersion !== false) {
        surfaceSelectionVersionRef.current += 1;
      }
      setSurface(identity.viewSchemaId);
      setStartupSheetRef({ ...identity.sheetRef });
      if (options.focusFirstGridTarget) {
        setPendingGridFocusSurface(identity.viewSchemaId);
      }
      if (options.reloadSheet) {
        setSheetReloadToken((current) => current + 1);
      }
    },
    [surfaceSelectionVersionRef],
  );

  const selectWorkbookSurface = useCallback(
    (
      viewSchemaId: string,
      options: { readonly focusFirstGridTarget?: boolean } = {},
    ) => {
      applyWorkbookIdentity(baseSurfaceIdentityForViewSchemaId(viewSchemaId), {
        focusFirstGridTarget: options.focusFirstGridTarget === true,
      });
    },
    [applyWorkbookIdentity],
  );

  const selectSavedViewIdentity = useCallback(
    (
      savedView: Pick<SavedViewResource, "saved_view_id" | "view_schema_id">,
    ) => {
      applyWorkbookIdentity(savedViewIdentityForSelection(savedView), {
        reloadSheet: true,
      });
    },
    [applyWorkbookIdentity],
  );

  const currentQueryStateForSurface = useCallback(
    (viewSchemaId: string): WorkbookQueryState => {
      if (viewSchemaId === timelineViewSchemaId) {
        return timelineQueryState;
      }
      if (viewSchemaId === hostsViewSchemaId) {
        return hostQueryState;
      }
      if (viewSchemaId === identitiesViewSchemaId) {
        return identityQueryState;
      }
      if (viewSchemaId === assessmentsViewSchemaId) {
        return assessmentQueryState;
      }
      return genericQueryState;
    },
    [
      assessmentQueryState,
      genericQueryState,
      hostQueryState,
      identityQueryState,
      timelineQueryState,
    ],
  );

  const applyQueryStateForSurface = useCallback(
    (viewSchemaId: string, queryState: WorkbookQueryState) => {
      const contract = workbookContractForViewSchemaId(viewSchemaId);
      if (viewSchemaId === timelineViewSchemaId) {
        setTimelineQueryState(queryState);
        setTimelineFilterDraft(defaultFilterDraft(timelineContract));
        return;
      }
      if (viewSchemaId === hostsViewSchemaId) {
        setHostQueryState(queryState);
        setHostFilterDraft(defaultFilterDraft(hostsContract));
        return;
      }
      if (viewSchemaId === identitiesViewSchemaId) {
        setIdentityQueryState(queryState);
        setIdentityFilterDraft(defaultFilterDraft(identitiesContract));
        return;
      }
      if (viewSchemaId === assessmentsViewSchemaId) {
        setAssessmentQueryState(queryState);
        setAssessmentFilterDraft(defaultFilterDraft(assessmentsContract));
        return;
      }
      setGenericQueryState(queryState);
      setGenericFilterDraft(defaultFilterDraft(contract));
    },
    [],
  );

  const selectSavedView = useCallback(
    (savedView: SavedViewResource) => {
      const nextSurface = knownWorkbookViewSchemaId(savedView.view_schema_id);
      const contract = workbookContractForViewSchemaId(nextSurface);
      applyQueryStateForSurface(
        nextSurface,
        savedViewQueryStateForRuntime(contract, savedView),
      );
      selectSavedViewIdentity(savedView);
    },
    [applyQueryStateForSurface, selectSavedViewIdentity],
  );

  const applyStartupIdentity = useCallback(
    (identity: {
      readonly sheetRef: WorkbookSheetRef;
      readonly viewSchemaId: string;
    }) => {
      applyWorkbookIdentity(identity, { bumpSelectionVersion: false });
    },
    [applyWorkbookIdentity],
  );

  const upsertSavedView = useCallback((savedView: SavedViewResource) => {
    setSavedViews((current) => upsertSavedViewList(current, savedView));
  }, []);

  const replaceSavedViews = useCallback(
    (nextSavedViews: SavedViewResource[]) => {
      setSavedViews(nextSavedViews);
    },
    [],
  );

  const deleteSavedViewIdentity = useCallback(
    (savedView: SavedViewResource, activeSheetRef: WorkbookSheetRef) => {
      setSavedViews((current) =>
        removeSavedViewList(current, savedView.saved_view_id),
      );
      const fallback = fallbackIdentityAfterSavedViewDelete(
        activeSheetRef,
        savedView,
      );
      if (fallback !== null) {
        applyWorkbookIdentity(fallback, { reloadSheet: true });
      }
    },
    [applyWorkbookIdentity],
  );

  const createSavedView = useCallback(
    async (input: {
      readonly displayName: string;
      readonly scope: "private" | "shared";
    }) => {
      const contract = activeContract;
      const queryState = currentQueryStateForSurface(contract.viewSchemaId);
      const result = await fetchJSON<SavedViewEnvelope>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}/saved-views`),
        {
          method: "POST",
          body: JSON.stringify({
            view_schema_id: contract.viewSchemaId,
            display_name: input.displayName,
            scope: input.scope,
            query_json: buildSavedViewQueryJson(contract, queryState),
            layout_json: buildSavedViewLayoutJson(contract),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const savedView = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (savedView === null) {
        throw new Error("Saved-view create returned an invalid resource.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [
      activeContract,
      apiBase,
      currentQueryStateForSurface,
      incidentId,
      selectSavedView,
      upsertSavedView,
    ],
  );

  const duplicateSavedView = useCallback(
    async (source: SavedViewResource) => {
      const contract = workbookContractForViewSchemaId(source.view_schema_id);
      const result = await fetchJSON<SavedViewEnvelope>(
        apiPath(apiBase, `/api/v1/incidents/${incidentId}/saved-views`),
        {
          method: "POST",
          body: JSON.stringify({
            view_schema_id: source.view_schema_id,
            display_name: `${source.display_name} Copy`,
            scope: "private",
            query_json: savedViewQueryJsonForPersistence(
              contract,
              source.query_json,
            ),
            layout_json: savedViewLayoutJsonForPersistence(
              contract,
              source.layout_json,
            ),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const savedView = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (savedView === null) {
        throw new Error("Saved-view duplicate returned an invalid resource.");
      }
      upsertSavedView(savedView);
      selectSavedView(savedView);
      return savedView;
    },
    [apiBase, incidentId, selectSavedView, upsertSavedView],
  );

  const updateSavedView = useCallback(
    async (
      savedView: SavedViewResource,
      input: {
        readonly displayName: string;
        readonly scope: "private" | "shared";
      },
    ) => {
      const contract = workbookContractForViewSchemaId(
        savedView.view_schema_id,
      );
      const queryState = currentQueryStateForSurface(savedView.view_schema_id);
      const result = await fetchJSON<SavedViewEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/saved-views/${savedView.saved_view_id}`,
        ),
        {
          method: "PATCH",
          body: JSON.stringify({
            base_saved_view_version: savedView.saved_view_version,
            display_name: input.displayName,
            scope: input.scope,
            query_json: buildSavedViewQueryJson(contract, queryState),
            layout_json: buildSavedViewLayoutJson(contract),
          }),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const updated = normalizeSavedViewResource(
        readEnvelope<SavedViewEnvelope>(result.payload).data,
      );
      if (updated === null) {
        throw new Error("Saved-view update returned an invalid resource.");
      }
      upsertSavedView(updated);
      return updated;
    },
    [apiBase, currentQueryStateForSurface, incidentId, upsertSavedView],
  );

  const deleteSavedView = useCallback(
    async (savedView: SavedViewResource) => {
      const result = await fetchJSON<Record<string, unknown>>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/saved-views/${savedView.saved_view_id}`,
        ),
        { method: "DELETE" },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      deleteSavedViewIdentity(savedView, startupSheetRef);
    },
    [apiBase, deleteSavedViewIdentity, incidentId, startupSheetRef],
  );

  const setWorkbookHomeSheetRef = useCallback(async () => {
    const result = await fetchJSON<Record<string, unknown>>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/workbook-preferences/me`,
      ),
      {
        method: "PUT",
        body: JSON.stringify({ home_sheet_ref: startupSheetRef }),
      },
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
  }, [apiBase, incidentId, startupSheetRef]);

  const setWorkbookDefaultSheetRef = useCallback(async () => {
    const result = await fetchJSON<Record<string, unknown>>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/workbook-preferences/default`,
      ),
      {
        method: "PUT",
        body: JSON.stringify({ default_sheet_ref: startupSheetRef }),
      },
    );
    if (!result.ok) {
      throw new Error(parseErrorMessage(result.payload));
    }
  }, [apiBase, incidentId, startupSheetRef]);

  const applyTimelineFilter = useCallback((draft: FilterDraft) => {
    applyFilterDraftToQuery(
      setTimelineQueryState,
      setTimelineFilterDraft,
      draft,
    );
  }, []);

  const applyHostFilter = useCallback((draft: FilterDraft) => {
    applyFilterDraftToQuery(setHostQueryState, setHostFilterDraft, draft);
  }, []);

  const applyIdentityFilter = useCallback((draft: FilterDraft) => {
    applyFilterDraftToQuery(
      setIdentityQueryState,
      setIdentityFilterDraft,
      draft,
    );
  }, []);

  const applyAssessmentFilter = useCallback((draft: FilterDraft) => {
    applyFilterDraftToQuery(
      setAssessmentQueryState,
      setAssessmentFilterDraft,
      draft,
    );
  }, []);

  const applyGenericFilter = useCallback((draft: FilterDraft) => {
    applyFilterDraftToQuery(setGenericQueryState, setGenericFilterDraft, draft);
  }, []);

  const clearActiveQueryControls = useCallback(() => {
    if (surface === timelineViewSchemaId) {
      setTimelineQueryState(emptyWorkbookQueryState());
      setTimelineFilterDraft(defaultFilterDraft(timelineContract));
      return;
    }
    if (surface === hostsViewSchemaId) {
      setHostQueryState(emptyWorkbookQueryState());
      setHostFilterDraft(defaultFilterDraft(hostsContract));
      return;
    }
    if (surface === identitiesViewSchemaId) {
      setIdentityQueryState(emptyWorkbookQueryState());
      setIdentityFilterDraft(defaultFilterDraft(identitiesContract));
      return;
    }
    if (surface === assessmentsViewSchemaId) {
      setAssessmentQueryState(emptyWorkbookQueryState());
      setAssessmentFilterDraft(defaultFilterDraft(assessmentsContract));
      return;
    }
    setGenericQueryState(emptyWorkbookQueryState());
    setGenericFilterDraft(defaultFilterDraft(activeContract));
  }, [activeContract, surface]);

  useEffect(() => {
    let cancelled = false;
    const startupQuery = workbookStartupQueryFromURLParams(params);
    const selectionVersionAtRequest = surfaceSelectionVersionRef.current;
    const loadStartup = async () => {
      const result = await fetchJSON<WorkbookStartupEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/workbook-startup${startupQuery}`,
        ),
      );
      if (cancelled || !result.ok) {
        return;
      }
      const envelope = readEnvelope<WorkbookStartupEnvelope>(result.payload);
      const startup = normalizeWorkbookStartupSelection(envelope.data);
      if (!startup) {
        return;
      }
      if (selectionVersionAtRequest !== surfaceSelectionVersionRef.current) {
        return;
      }
      const nextSurface = knownWorkbookViewSchemaId(
        startup.selectedViewSchemaId,
      );
      const startupSavedView = normalizeSavedViewResource(
        startup.selectedSavedView,
      );
      if (
        startup.selectedSheetRef.kind === "saved_view" &&
        startupSavedView !== null &&
        startupSavedView.saved_view_id === startup.selectedSheetRef.id
      ) {
        const contract = workbookContractForViewSchemaId(nextSurface);
        upsertSavedView(startupSavedView);
        applyQueryStateForSurface(
          nextSurface,
          savedViewQueryStateForRuntime(contract, startupSavedView),
        );
      }
      applyStartupIdentity({
        sheetRef: startup.selectedSheetRef,
        viewSchemaId: nextSurface,
      });
    };
    void loadStartup();
    return () => {
      cancelled = true;
    };
  }, [
    apiBase,
    applyQueryStateForSurface,
    applyStartupIdentity,
    incidentId,
    params,
    surfaceSelectionVersionRef,
    upsertSavedView,
  ]);

  useEffect(() => {
    let cancelled = false;
    const nextSavedViews: SavedViewResource[] = [];
    const loadSavedViews = async () => {
      let cursorToken: string | null = null;
      do {
        const query = new URLSearchParams({ limit: "100" });
        if (cursorToken !== null) {
          query.set("cursor_token", cursorToken);
        }
        const result = await fetchJSON<SavedViewListEnvelope>(
          apiPath(
            apiBase,
            `/api/v1/incidents/${incidentId}/saved-views?${query.toString()}`,
          ),
        );
        if (cancelled) {
          return;
        }
        if (!result.ok) {
          handleWorkbookLoadFailure(
            parseErrorMessage(result.payload),
            "Saved views load failed.",
            onIncidentAccessLost,
          );
          replaceSavedViews([]);
          return;
        }

        const envelope = readEnvelope<SavedViewListEnvelope>(result.payload);
        for (const savedView of envelope.data.saved_views) {
          const normalized = normalizeSavedViewResource(savedView);
          if (normalized !== null) {
            nextSavedViews.push(normalized);
          }
        }
        const paging = envelope.meta?.paging;
        cursorToken =
          paging?.has_more === true && paging.next_cursor
            ? paging.next_cursor
            : null;
      } while (cursorToken !== null);

      if (!cancelled) {
        replaceSavedViews(nextSavedViews);
      }
    };

    void loadSavedViews();
    return () => {
      cancelled = true;
    };
  }, [apiBase, incidentId, onIncidentAccessLost, replaceSavedViews]);

  useEffect(() => {
    if (startupSheetRef.kind === "saved_view") {
      return;
    }
    setGenericQueryState(emptyWorkbookQueryState());
    setGenericFilterDraft(defaultFilterDraft(activeContract));
  }, [activeContract, startupSheetRef.kind]);

  useEffect(() => {
    const next = new URLSearchParams(window.location.search);
    next.set("incident_id", incidentId);
    if (startupSheetRef.kind === "saved_view") {
      next.delete("view_schema_id");
      next.set("sheet_ref_kind", startupSheetRef.kind);
      next.set("sheet_ref_id", startupSheetRef.id);
    } else {
      next.set("view_schema_id", surface);
      next.delete("sheet_ref_kind");
      next.delete("sheet_ref_id");
    }
    next.delete("surface");
    window.history.replaceState({}, "", `/?${next.toString()}`);
  }, [incidentId, startupSheetRef, surface]);

  const activeQueryControls: WorkbookShellActiveQueryControls =
    surface === timelineViewSchemaId
      ? {
          contract: timelineContract,
          filterDraft: timelineFilterDraft,
          onApplyFilter: applyTimelineFilter,
          onClearAll: clearActiveQueryControls,
          onFilterDraftChange: setTimelineFilterDraft,
          onGroupByChange: (groupBy: string | null) => {
            setTimelineQueryState((current) =>
              updateGroupBy(timelineContract, current, groupBy),
            );
          },
          onRemoveFilter: (fieldKey: string) => {
            setTimelineQueryState((current) =>
              removeFilterField(current, fieldKey),
            );
          },
          onToggleSort: (fieldKey: string) => {
            setTimelineQueryState((current) =>
              toggleSortField(timelineContract, current, fieldKey),
            );
          },
          queryState: timelineQueryState,
          surface: timelineViewSchemaId as WorkbookSurface,
        }
      : surface === hostsViewSchemaId
        ? {
            contract: hostsContract,
            filterDraft: hostFilterDraft,
            onApplyFilter: applyHostFilter,
            onClearAll: clearActiveQueryControls,
            onFilterDraftChange: setHostFilterDraft,
            onGroupByChange: (groupBy: string | null) => {
              setHostQueryState((current) =>
                updateGroupBy(hostsContract, current, groupBy),
              );
            },
            onRemoveFilter: (fieldKey: string) => {
              setHostQueryState((current) =>
                removeFilterField(current, fieldKey),
              );
            },
            onToggleSort: (fieldKey: string) => {
              setHostQueryState((current) =>
                toggleSortField(hostsContract, current, fieldKey),
              );
            },
            queryState: hostQueryState,
            surface: hostsViewSchemaId as WorkbookSurface,
          }
        : surface === identitiesViewSchemaId
          ? {
              contract: identitiesContract,
              filterDraft: identityFilterDraft,
              onApplyFilter: applyIdentityFilter,
              onClearAll: clearActiveQueryControls,
              onFilterDraftChange: setIdentityFilterDraft,
              onGroupByChange: (groupBy: string | null) => {
                setIdentityQueryState((current) =>
                  updateGroupBy(identitiesContract, current, groupBy),
                );
              },
              onRemoveFilter: (fieldKey: string) => {
                setIdentityQueryState((current) =>
                  removeFilterField(current, fieldKey),
                );
              },
              onToggleSort: (fieldKey: string) => {
                setIdentityQueryState((current) =>
                  toggleSortField(identitiesContract, current, fieldKey),
                );
              },
              queryState: identityQueryState,
              surface: identitiesViewSchemaId as WorkbookSurface,
            }
          : surface === assessmentsViewSchemaId
            ? {
                contract: assessmentsContract,
                filterDraft: assessmentFilterDraft,
                onApplyFilter: applyAssessmentFilter,
                onClearAll: clearActiveQueryControls,
                onFilterDraftChange: setAssessmentFilterDraft,
                onGroupByChange: (groupBy: string | null) => {
                  setAssessmentQueryState((current) =>
                    updateGroupBy(assessmentsContract, current, groupBy),
                  );
                },
                onRemoveFilter: (fieldKey: string) => {
                  setAssessmentQueryState((current) =>
                    removeFilterField(current, fieldKey),
                  );
                },
                onToggleSort: (fieldKey: string) => {
                  setAssessmentQueryState((current) =>
                    toggleSortField(assessmentsContract, current, fieldKey),
                  );
                },
                queryState: assessmentQueryState,
                surface: assessmentsViewSchemaId as WorkbookSurface,
              }
            : {
                contract: activeContract,
                filterDraft: genericFilterDraft,
                onApplyFilter: applyGenericFilter,
                onClearAll: clearActiveQueryControls,
                onFilterDraftChange: setGenericFilterDraft,
                onGroupByChange: (groupBy: string | null) => {
                  setGenericQueryState((current) =>
                    updateGroupBy(activeContract, current, groupBy),
                  );
                },
                onRemoveFilter: (fieldKey: string) => {
                  setGenericQueryState((current) =>
                    removeFilterField(current, fieldKey),
                  );
                },
                onToggleSort: (fieldKey: string) => {
                  setGenericQueryState((current) =>
                    toggleSortField(activeContract, current, fieldKey),
                  );
                },
                queryState: genericQueryState,
                surface: surface as WorkbookSurface,
              };

  const activeSavedView =
    startupSheetRef.kind === "saved_view"
      ? (savedViews.find(
          (savedView) => savedView.saved_view_id === startupSheetRef.id,
        ) ?? null)
      : null;
  const activeSavedViewModified = savedViewConfigurationIsModified({
    contract: activeContract,
    currentQueryState: currentQueryStateForSurface(activeContract.viewSchemaId),
    savedView:
      activeSavedView?.view_schema_id === activeContract.viewSchemaId
        ? activeSavedView
        : null,
  });

  return {
    commands: {
      createSavedView,
      deleteSavedView,
      duplicateSavedView,
      selectWorkbookSurface,
      setPendingGridFocusSurface,
      setWorkbookDefaultSheetRef,
      setWorkbookHomeSheetRef,
      setAssessmentQueryState,
      setGenericQueryState,
      setHostQueryState,
      setIdentityQueryState,
      setTimelineQueryState,
      selectSavedView,
      updateSavedView,
    },
    snapshot: {
      activeContract,
      activeQueryControls,
      activeSavedViewModified,
      assessmentQueryState,
      genericQueryState,
      hostQueryState,
      identityQueryState,
      pendingGridFocusSurface,
      savedViews,
      sheetReloadToken,
      startupSheetRef,
      surface,
      timelineQueryState,
    },
  };
}

function clearAppliedFilterDraft(current: FilterDraft): FilterDraft {
  return {
    ...current,
    booleanValue: "",
    value: "",
  };
}

function applyFilterDraftToQuery(
  setQueryState: WorkbookQueryStateSetter,
  setFilterDraft: FilterDraftSetter,
  draft: FilterDraft,
): void {
  setQueryState((current) => applyFilterDraft(current, draft));
  setFilterDraft(clearAppliedFilterDraft);
}
