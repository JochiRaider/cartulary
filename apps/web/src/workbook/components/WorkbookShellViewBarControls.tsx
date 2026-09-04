import { sheetRefKey } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { useWorkbookShellRuntime } from "../hooks/useWorkbookShellRuntime";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import type { WorkbookViewBarWorkingSetBinding } from "./WorkbookViewBar";

type WorkbookShellRuntime = ReturnType<typeof useWorkbookShellRuntime>;

export function workbookShellViewBarWorkingSet({
  chromeMode,
  currentIncidentRole,
  currentUserId,
  incidentId,
  networkAnalysisActive,
  runtime,
}: {
  readonly chromeMode: WorkbookChromeMode;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly incidentId: string;
  readonly networkAnalysisActive: boolean;
  readonly runtime: WorkbookShellRuntime;
}): WorkbookViewBarWorkingSetBinding {
  if (networkAnalysisActive) return { query: null, savedView: null };

  const { commands, snapshot } = runtime;
  const selectedSheetRef = snapshot.startupSheetRef;
  const selectedSavedView =
    selectedSheetRef.kind === "saved_view" &&
    (snapshot.savedViewsResource.kind === "ready" ||
      snapshot.savedViewsResource.kind === "invalid_selection")
      ? (snapshot.savedViewsResource.savedViews.find(
          (savedView) => savedView.saved_view_id === selectedSheetRef.id,
        ) ?? null)
      : null;
  const subjectKey = `${incidentId}:${snapshot.surface}:${sheetRefKey(selectedSheetRef)}:${selectedSavedView?.saved_view_version ?? 0}`;
  return {
    query:
      chromeMode === "below_supported_minimum"
        ? null
        : {
            contract: snapshot.activeQueryControls.contract,
            filterDraft: snapshot.activeQueryControls.filterDraft,
            layoutState: snapshot.activeLayoutState,
            onApplyFilter: snapshot.activeQueryControls.onApplyFilter,
            onClearFilters: snapshot.activeQueryControls.onClearFilters,
            onColumnHiddenChange:
              snapshot.activeLayoutControls.onColumnHiddenChange,
            onColumnMove: snapshot.activeLayoutControls.onColumnMove,
            onFilterDraftChange:
              snapshot.activeQueryControls.onFilterDraftChange,
            onGroupByChange: snapshot.activeQueryControls.onGroupByChange,
            onRemoveFilter: snapshot.activeQueryControls.onRemoveFilter,
            onResetColumns: snapshot.activeLayoutControls.onResetColumns,
            onSortChange: snapshot.activeQueryControls.onSortChange,
            queryState: snapshot.activeQueryControls.queryState,
            surface: snapshot.activeQueryControls.surface,
            subjectKey,
          },
    savedView: {
      activeViewSchemaId: snapshot.surface,
      currentIncidentRole,
      currentUserId,
      isModified: snapshot.activeSavedViewModified,
      onCreateSavedView: commands.createSavedView,
      onDeleteSavedView: commands.deleteSavedView,
      onDuplicateSavedView: commands.duplicateSavedView,
      onResetToSavedView: commands.selectSavedView,
      onSelectBaseSurface: commands.selectWorkbookSurface,
      onSelectSavedView: commands.selectSavedView,
      onSetDefaultSheetRef: commands.setWorkbookDefaultSheetRef,
      onSetHomeSheetRef: commands.setWorkbookHomeSheetRef,
      onUpdateSavedView: commands.updateSavedView,
      savedViewsResource: snapshot.savedViewsResource,
      selectedSheetRef,
    },
  };
}
