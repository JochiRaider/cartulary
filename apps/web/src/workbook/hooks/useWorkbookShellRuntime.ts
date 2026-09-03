import { useMemo } from "react";
import type {
  ExtensionAvailabilityController,
  ExtensionAvailabilityTag,
  ExtensionWorkspaceIdentity,
} from "../../extensions/extensionAvailability";
import { useWorkbookColumnLayoutController } from "../layout/useWorkbookColumnLayoutController";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookPreferencePort } from "../ports/WorkbookPreferencePort";
import type { WorkbookSavedViewPort } from "../ports/WorkbookSavedViewPort";
import { useWorkbookStartupAdmission } from "../startup/useWorkbookStartupAdmission";
import type {
  WorkbookStartupAvailability,
  WorkbookStartupPort,
} from "../startup/WorkbookStartupPort";
import { useWorkbookQueryController } from "./useWorkbookQueryController";
import { useWorkbookSavedViewController } from "./useWorkbookSavedViewController";
import { useWorkbookStartupController } from "./useWorkbookStartupController";

type WorkbookShellMutableRef<T> = {
  current: T;
};

export function useWorkbookShellRuntime({
  incidentId,
  onIncidentAccessLost,
  surfaceSelectionVersionRef,
  extensionAvailability,
  onExtensionAvailabilityChange,
  preferencePort,
  savedViewPort,
  startupPort,
}: {
  readonly incidentId: string;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly surfaceSelectionVersionRef: WorkbookShellMutableRef<number>;
  readonly extensionAvailability: ExtensionAvailabilityController;
  readonly onExtensionAvailabilityChange: () => void;
  readonly preferencePort: WorkbookPreferencePort;
  readonly savedViewPort: WorkbookSavedViewPort;
  readonly startupPort: WorkbookStartupPort;
}) {
  const startupController = useWorkbookStartupController({
    incidentId,
    preferencePort,
    surfaceSelectionVersionRef,
  });
  const { gridEntryFocusRequest, sheetReloadToken, startupSheetRef, surface } =
    startupController.snapshot;
  const { params } = startupController.refs;
  const {
    acknowledgeGridEntryFocus,
    applyStartupIdentity,
    applyWorkbookIdentity,
    cancelGridEntryFocus,
    selectExtensionWorkspace,
    selectWorkbookSurface,
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
    applyLayoutStateForSurface,
    applyQueryStateForSurface,
    applyWorkbookIdentity,
    currentLayoutStateForSurface,
    currentQueryStateForSurface,
    onIncidentAccessLost,
    savedViewPort,
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
  const startupAvailabilityPort = useMemo(
    () => ({
      acceptWorkbookStartup: (
        tag: ExtensionAvailabilityTag,
        availability: WorkbookStartupAvailability,
      ) =>
        extensionAvailability.acceptWorkbookStartupWorkspaces(
          tag,
          availability.workspaces,
        ),
      isRenderable: (identity: ExtensionWorkspaceIdentity) =>
        extensionAvailability.isRenderable(identity),
      reserve: () => extensionAvailability.reserve(),
    }),
    [extensionAvailability],
  );
  useWorkbookStartupAdmission({
    incidentId,
    urlParams: params,
    availabilityPort: startupAvailabilityPort,
    selectionPort: startupSelectionPort,
    savedViewStatePort: startupSavedViewStatePort,
    startupPort,
    onIncidentAccessLost,
    onAvailabilityChange: onExtensionAvailabilityChange,
  });

  return {
    commands: {
      acknowledgeGridEntryFocus,
      cancelGridEntryFocus,
      createSavedView,
      deleteSavedView,
      duplicateSavedView,
      selectWorkbookSurface,
      selectExtensionWorkspace,
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
      gridEntryFocusRequest,
      savedViews,
      sheetReloadToken,
      startupSheetRef,
      surface,
      timelineQueryState,
    },
  };
}
