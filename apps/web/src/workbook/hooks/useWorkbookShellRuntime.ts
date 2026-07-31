import { useMemo } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import { useWorkbookColumnLayoutController } from "../layout/useWorkbookColumnLayoutController";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useWorkbookStartupAdmission } from "../startup/useWorkbookStartupAdmission";
import { useWorkbookQueryController } from "./useWorkbookQueryController";
import { useWorkbookSavedViewController } from "./useWorkbookSavedViewController";
import { useWorkbookStartupController } from "./useWorkbookStartupController";

type WorkbookShellMutableRef<T> = {
  current: T;
};

export function useWorkbookShellRuntime({
  apiBase,
  incidentId,
  onIncidentAccessLost,
  surfaceSelectionVersionRef,
  extensionAvailability,
  onExtensionAvailabilityChange,
}: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly surfaceSelectionVersionRef: WorkbookShellMutableRef<number>;
  readonly extensionAvailability: ExtensionAvailabilityController;
  readonly onExtensionAvailabilityChange: () => void;
}) {
  const startupController = useWorkbookStartupController({
    apiBase,
    incidentId,
    surfaceSelectionVersionRef,
  });
  const {
    pendingGridFocusSurface,
    sheetReloadToken,
    startupSheetRef,
    surface,
  } = startupController.snapshot;
  const { params } = startupController.refs;
  const {
    applyStartupIdentity,
    applyWorkbookIdentity,
    selectExtensionWorkspace,
    selectWorkbookSurface,
    setPendingGridFocusSurface,
    setWorkbookDefaultSheetRef,
    setWorkbookHomeSheetRef,
  } = startupController.commands;
  const workbookQueries = useWorkbookQueryController({
    startupSheetRef,
    surface,
  });
  const {
    activeContract,
    activeQueryControls,
    assessmentQueryState,
    genericQueryState,
    hostQueryState,
    identityQueryState,
    timelineQueryState,
  } = workbookQueries.snapshot;
  const {
    applyQueryStateForSurface,
    currentQueryStateForSurface,
    setAssessmentQueryState,
    setGenericQueryState,
    setHostQueryState,
    setIdentityQueryState,
    setTimelineQueryState,
  } = workbookQueries.commands;
  const workbookLayouts = useWorkbookColumnLayoutController({
    activeContract,
    startupSheetRef,
  });
  const { activeLayoutControls, activeLayoutState } = workbookLayouts.snapshot;
  const { applyLayoutStateForSurface, currentLayoutStateForSurface } =
    workbookLayouts.commands;

  const savedViewController = useWorkbookSavedViewController({
    activeContract,
    apiBase,
    applyLayoutStateForSurface,
    applyQueryStateForSurface,
    applyWorkbookIdentity,
    currentLayoutStateForSurface,
    currentQueryStateForSurface,
    incidentId,
    onIncidentAccessLost,
    startupSheetRef,
  });
  const { activeSavedViewModified, savedViews } = savedViewController.snapshot;
  const {
    createSavedView,
    deleteSavedView,
    duplicateSavedView,
    selectSavedView,
    updateSavedView,
    upsertSavedView,
  } = savedViewController.commands;

  const startupSelectionPort = useMemo(
    () => ({
      applyStartupIdentity,
      readSelectionVersion: () => surfaceSelectionVersionRef.current,
      selectTimeline: () => selectWorkbookSurface(timelineViewSchemaId),
    }),
    [applyStartupIdentity, selectWorkbookSurface, surfaceSelectionVersionRef],
  );
  const startupSavedViewStatePort = useMemo(
    () => ({
      applyLayoutStateForSurface,
      applyQueryStateForSurface,
      upsertSavedView,
    }),
    [applyLayoutStateForSurface, applyQueryStateForSurface, upsertSavedView],
  );
  useWorkbookStartupAdmission({
    apiBase,
    incidentId,
    urlParams: params,
    availabilityPort: extensionAvailability,
    selectionPort: startupSelectionPort,
    savedViewStatePort: startupSavedViewStatePort,
    onAvailabilityChange: onExtensionAvailabilityChange,
  });

  return {
    commands: {
      createSavedView,
      deleteSavedView,
      duplicateSavedView,
      selectWorkbookSurface,
      selectExtensionWorkspace,
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
      activeLayoutControls,
      activeLayoutState,
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
