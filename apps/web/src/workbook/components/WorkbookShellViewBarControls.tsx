import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { useWorkbookShellRuntime } from "../hooks/useWorkbookShellRuntime";
import type { WorkbookChromeMode } from "../layout/workbookResponsiveLayout";
import { ActiveSurfaceSavedViewSelector } from "./ActiveSurfaceSavedViewSelector";
import { WorkbookGridControls } from "./WorkbookGridControls";

type WorkbookShellRuntime = ReturnType<typeof useWorkbookShellRuntime>;

export function WorkbookShellSavedViewControl({
  chromeMode,
  currentIncidentRole,
  currentUserId,
  networkAnalysisActive,
  runtime,
}: {
  readonly chromeMode: WorkbookChromeMode;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly networkAnalysisActive: boolean;
  readonly runtime: WorkbookShellRuntime;
}) {
  if (networkAnalysisActive) return null;
  const { commands, snapshot } = runtime;
  return (
    <ActiveSurfaceSavedViewSelector
      activeViewSchemaId={snapshot.surface}
      chromeMode={chromeMode}
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      isModified={snapshot.activeSavedViewModified}
      onCreateSavedView={commands.createSavedView}
      onDeleteSavedView={commands.deleteSavedView}
      onDuplicateSavedView={commands.duplicateSavedView}
      onResetToSavedView={commands.selectSavedView}
      onSelectBaseSurface={commands.selectWorkbookSurface}
      onSelectSavedView={commands.selectSavedView}
      onSetDefaultSheetRef={commands.setWorkbookDefaultSheetRef}
      onSetHomeSheetRef={commands.setWorkbookHomeSheetRef}
      onUpdateSavedView={commands.updateSavedView}
      savedViewsResource={snapshot.savedViewsResource}
      selectedSheetRef={snapshot.startupSheetRef}
    />
  );
}

export function WorkbookShellQueryControls({
  chromeMode,
  networkAnalysisActive,
  runtime,
}: {
  readonly chromeMode: WorkbookChromeMode;
  readonly networkAnalysisActive: boolean;
  readonly runtime: WorkbookShellRuntime;
}) {
  if (networkAnalysisActive || chromeMode === "below_supported_minimum") {
    return null;
  }
  const { snapshot } = runtime;
  return (
    <WorkbookGridControls
      chromeMode={chromeMode}
      contract={snapshot.activeQueryControls.contract}
      filterDraft={snapshot.activeQueryControls.filterDraft}
      layoutState={snapshot.activeLayoutState}
      onApplyFilter={snapshot.activeQueryControls.onApplyFilter}
      onClearAll={snapshot.activeQueryControls.onClearAll}
      onColumnHiddenChange={snapshot.activeLayoutControls.onColumnHiddenChange}
      onColumnMove={snapshot.activeLayoutControls.onColumnMove}
      onFilterDraftChange={snapshot.activeQueryControls.onFilterDraftChange}
      onGroupByChange={snapshot.activeQueryControls.onGroupByChange}
      onRemoveFilter={snapshot.activeQueryControls.onRemoveFilter}
      onResetColumns={snapshot.activeLayoutControls.onResetColumns}
      onSortChange={snapshot.activeQueryControls.onSortChange}
      queryState={snapshot.activeQueryControls.queryState}
      surface={snapshot.activeQueryControls.surface}
    />
  );
}
