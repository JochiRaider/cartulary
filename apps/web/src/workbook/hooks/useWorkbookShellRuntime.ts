import { useEffect } from "react";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON, readEnvelope } from "../../services/workbookApi";
import { workbookLayoutStateFromSavedViewLayoutJson } from "../models/workbookQuery";
import { savedViewQueryStateForRuntime } from "../models/workbookSavedViewRuntime";
import { normalizeSavedViewResource } from "../models/workbookSavedViews";
import {
  normalizeWorkbookStartupSelection,
  workbookStartupQueryFromURLParams,
} from "../models/workbookStartup";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import { knownWorkbookViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useWorkbookLayoutController } from "./useWorkbookLayoutController";
import { useWorkbookQueryController } from "./useWorkbookQueryController";
import { useWorkbookSavedViewController } from "./useWorkbookSavedViewController";
import { useWorkbookStartupController } from "./useWorkbookStartupController";

type WorkbookShellMutableRef<T> = {
  current: T;
};

type WorkbookStartupEnvelope = {
  data?: unknown;
};

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
  const workbookLayouts = useWorkbookLayoutController({
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

  useEffect(() => {
    let cancelled = false;
    const startupQuery = workbookStartupQueryFromURLParams(params);
    const selectionVersionAtRequest = surfaceSelectionVersionRef.current;
    const loadStartup = async () => {
      const result = await fetchWorkbookJSON<WorkbookStartupEnvelope>(
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
      const nextSurface =
        startup.selectedSheetRef.kind === "extension_workspace"
          ? null
          : knownWorkbookViewSchemaId(startup.selectedViewSchemaId ?? "");
      const startupSavedView = normalizeSavedViewResource(
        startup.selectedSavedView,
      );
      if (
        startup.selectedSheetRef.kind === "saved_view" &&
        startupSavedView !== null &&
        nextSurface !== null &&
        startupSavedView.saved_view_id === startup.selectedSheetRef.id
      ) {
        const contract = workbookContractForViewSchemaId(nextSurface ?? "");
        upsertSavedView(startupSavedView);
        applyQueryStateForSurface(
          nextSurface,
          savedViewQueryStateForRuntime(contract, startupSavedView),
        );
        applyLayoutStateForSurface(
          nextSurface,
          workbookLayoutStateFromSavedViewLayoutJson(
            contract,
            startupSavedView.layout_json,
          ),
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
    applyLayoutStateForSurface,
    applyStartupIdentity,
    incidentId,
    params,
    surfaceSelectionVersionRef,
    upsertSavedView,
  ]);

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
