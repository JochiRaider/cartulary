import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { sheetRefKey } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { useWorkbookQueryController } from "../hooks/useWorkbookQueryController";
import type { useWorkbookShellRuntime } from "../hooks/useWorkbookShellRuntime";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import type { WorkbookResolvedLayoutState } from "../layout/workbookColumnLayout";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import { workbookQueryStateFromSavedViewQueryJson } from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookNavigationActions } from "../navigation/WorkbookWorkbenchContext";
import type { WorkbookPreferenceController } from "../preferences/WorkbookPreferenceController";
import type { WorkbookViewBarWorkingSetBinding } from "./WorkbookViewBar";

type WorkbookShellRuntime = ReturnType<typeof useWorkbookShellRuntime>;

export function workbookQueryViewBarBinding({
  queryControls,
  layoutState,
  layoutControls,
  subjectKey,
  onApplyPreset,
}: {
  readonly queryControls: ReturnType<
    typeof useWorkbookQueryController
  >["snapshot"]["activeQueryControls"];
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly layoutControls: Pick<
    WorkbookSurfaceLayoutOwner["commands"],
    | "sizing"
    | "freezing"
    | "onColumnHiddenChange"
    | "onColumnMove"
    | "onResetColumns"
  >;
  readonly subjectKey: string;
  readonly onApplyPreset?: ((id: string) => void) | undefined;
}): NonNullable<WorkbookViewBarWorkingSetBinding["query"]> {
  return {
    onApplyPreset,
    contract: queryControls.contract,
    filterDraft: queryControls.filterDraft,
    layoutState,
    sizing: layoutControls.sizing,
    freezing: layoutControls.freezing,
    onApplyFilter: queryControls.onApplyFilter,
    onClearFilters: queryControls.onClearFilters,
    onColumnHiddenChange: layoutControls.onColumnHiddenChange,
    onColumnMove: layoutControls.onColumnMove,
    onFilterDraftChange: queryControls.onFilterDraftChange,
    onGroupByChange: queryControls.onGroupByChange,
    onRemoveFilter: queryControls.onRemoveFilter,
    onResetColumns: layoutControls.onResetColumns,
    onSortChange: queryControls.onSortChange,
    queryState: queryControls.queryState,
    requestedFilters: queryControls.requestedFilters,
    requestedGroupBy: queryControls.requestedGroupBy,
    requestedSort: queryControls.requestedSort,
    surface: queryControls.surface,
    subjectKey,
  };
}

export function workbookShellViewBarWorkingSet({
  chromeMode,
  currentIncidentRole,
  currentUserId,
  incidentId,
  networkAnalysisActive,
  navigation,
  runtime,
  preferenceController,
  onInspectPreferences,
}: {
  readonly chromeMode: WorkbookChromeMode;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly incidentId: string;
  readonly networkAnalysisActive: boolean;
  readonly navigation: WorkbookNavigationActions;
  readonly runtime: WorkbookShellRuntime;
  readonly preferenceController?: WorkbookPreferenceController | undefined;
  readonly onInspectPreferences?:
    | ((target?: HTMLElement | null) => void)
    | undefined;
}): WorkbookViewBarWorkingSetBinding {
  if (networkAnalysisActive) return { query: null, savedView: null };

  const { commands, snapshot } = runtime;
  const selectedSheetRef = snapshot.startupSheetRef;
  const selectedSavedView = snapshot.savedViewsResource.selectedSavedView;
  const subjectKey = `${incidentId}:${snapshot.surface}:${sheetRefKey(selectedSheetRef)}:${selectedSavedView?.saved_view_version ?? 0}`;
  return {
    query:
      chromeMode === "below_supported_minimum"
        ? null
        : workbookQueryViewBarBinding({
            queryControls: snapshot.activeQueryControls,
            layoutState: snapshot.activeLayoutState,
            layoutControls: snapshot.activeLayoutControls,
            subjectKey,
            onApplyPreset:
              snapshot.surface === timelineViewSchemaId
                ? (id) => {
                    const preset =
                      cartularyDesignPresentation.workbookWorkbench.presets.find(
                        (item) => item.id === id,
                      );
                    if (preset)
                      commands.applyQueryStateForSurface(
                        snapshot.surface,
                        workbookQueryStateFromSavedViewQueryJson(
                          snapshot.activeContract,
                          preset,
                        ),
                      );
                  }
                : undefined,
          }),
    savedView: {
      activeViewSchemaId: snapshot.surface,
      currentIncidentRole,
      currentUserId,
      isModified: snapshot.activeSavedViewModified,
      controller: runtime.savedViewOwner,
      navigation,
      preferenceController,
      onInspectPreferences,
      savedViewsResource: snapshot.savedViewsResource,
      selectedSheetRef,
    },
  };
}
