import { useMemo } from "react";
import type {
  ExtensionAvailabilityController,
  ExtensionAvailabilityTag,
  ExtensionWorkspaceIdentity,
} from "../../extensions/extensionAvailability";
import { useWorkbookColumnLayoutController } from "../layout/useWorkbookColumnLayoutController";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { SavedViewBinding } from "../savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";
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
  onAuthorityUncertain,
  surfaceSelectionVersionRef,
  extensionAvailability,
  onExtensionAvailabilityChange,
  savedViewOwner,
  bindWorkbookSavedViews,
  authorizationRecovered,
  apiBase,
  startupPort,
}: {
  readonly incidentId: string;
  readonly onAuthorityUncertain?: (() => void) | undefined;
  readonly surfaceSelectionVersionRef: WorkbookShellMutableRef<number>;
  readonly extensionAvailability: ExtensionAvailabilityController;
  readonly onExtensionAvailabilityChange: () => void;
  readonly savedViewOwner: WorkbookSavedViewController;
  readonly bindWorkbookSavedViews: (binding: SavedViewBinding | null) => void;
  readonly authorizationRecovered: SavedViewBinding["authorizationRecovered"];
  readonly apiBase?: string | undefined;
  readonly startupPort: WorkbookStartupPort;
}) {
  const startupController = useWorkbookStartupController({
    incidentId,
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
  } = startupController.commands;
  const workbookQueries = useWorkbookQueryController({
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
    controller: savedViewOwner,
    bindWorkbook: bindWorkbookSavedViews,
    incidentId,
    selectionGeneration: surfaceSelectionVersionRef.current,
    authorizationRecovered,
    apiBase,
    startupSheetRef,
  });
  const { activeSavedViewModified, savedViewsResource } =
    savedViewController.snapshot;
  const { selectSavedView, upsertSavedView } = savedViewController.commands;

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
      waitForDiscovery: (signal: AbortSignal) =>
        extensionAvailability.waitForDiscovery(signal),
    }),
    [extensionAvailability],
  );
  const startupAdmission = useWorkbookStartupAdmission({
    incidentId,
    urlParams: params,
    availabilityPort: startupAvailabilityPort,
    selectionPort: startupSelectionPort,
    savedViewStatePort: startupSavedViewStatePort,
    startupPort,
    onAuthorityUncertain,
    onAvailabilityChange: onExtensionAvailabilityChange,
  });

  return {
    savedViewOwner,
    commands: {
      acknowledgeGridEntryFocus,
      cancelGridEntryFocus,
      selectWorkbookSurface,
      selectExtensionWorkspace,
      setAssessmentQueryState,
      setGenericQueryState,
      setHostQueryState,
      setIdentityQueryState,
      setTimelineQueryState,
      selectSavedView,
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
      savedViewsResource,
      sheetReloadToken,
      startupSheetRef,
      startupPending: startupAdmission.pending,
      surface,
      timelineQueryState,
    },
  };
}
