import { useEffect, useRef } from "react";
import type {
  ExtensionAvailabilityTag,
  ExtensionWorkspaceIdentity,
} from "../../extensions/extensionAvailability";
import { apiPath } from "../../services/browserApi";
import { fetchWorkbookJSON, readEnvelope } from "../../services/workbookApi";
import type { WorkbookIdentity } from "../hooks/useWorkbookStartupController";
import {
  type WorkbookLayoutState,
  type WorkbookQueryState,
  workbookLayoutStateFromSavedViewLayoutJson,
} from "../models/workbookQuery";
import { savedViewQueryStateForRuntime } from "../models/workbookSavedViewRuntime";
import {
  normalizeSavedViewResource,
  type SavedViewResource,
} from "../models/workbookSavedViews";
import {
  normalizeWorkbookStartupSelection,
  workbookStartupQueryFromURLParams,
} from "../models/workbookStartup";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import { knownWorkbookViewSchemaId } from "../models/workbookSurfaceRegistry";

export type StartupFallbackReason =
  | "availability_reservation_unavailable"
  | "availability_rejected"
  | "selected_extension_not_renderable";

export interface WorkbookStartupAvailabilityPort {
  reserve(): ExtensionAvailabilityTag | null;
  acceptWorkbookStartup(
    tag: ExtensionAvailabilityTag,
    availability: unknown,
  ): boolean;
  isRenderable(identity: ExtensionWorkspaceIdentity): boolean;
}

export interface WorkbookStartupSelectionPort {
  readSelectionVersion(): number;
  applyStartupIdentity(identity: WorkbookIdentity): void;
  selectTimeline(reason: StartupFallbackReason): void;
}

export interface WorkbookStartupSavedViewStatePort {
  upsertSavedView(savedView: SavedViewResource): void;
  applyQueryStateForSurface(
    viewSchemaId: string,
    query: WorkbookQueryState,
  ): void;
  applyLayoutStateForSurface(
    viewSchemaId: string,
    layout: WorkbookLayoutState,
  ): void;
}

type WorkbookStartupEnvelope = {
  readonly data?: unknown;
};

type StartupAdmissionGuard = {
  readonly incidentId: string;
  readonly requestOrdinal: number;
  readonly selectionVersionAtDispatch: number;
  readonly availabilityTag: ExtensionAvailabilityTag;
};

export function useWorkbookStartupAdmission({
  apiBase,
  incidentId,
  urlParams,
  availabilityPort,
  selectionPort,
  savedViewStatePort,
  onAvailabilityChange,
}: {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly urlParams: URLSearchParams;
  readonly availabilityPort: WorkbookStartupAvailabilityPort;
  readonly selectionPort: WorkbookStartupSelectionPort;
  readonly savedViewStatePort: WorkbookStartupSavedViewStatePort;
  readonly onAvailabilityChange: () => void;
}): void {
  const requestOrdinalRef = useRef(0);

  useEffect(() => {
    let cancelled = false;
    const startupQuery = workbookStartupQueryFromURLParams(urlParams);
    const selectionVersionAtDispatch = selectionPort.readSelectionVersion();

    const loadStartup = async () => {
      const availabilityTag = availabilityPort.reserve();
      if (availabilityTag === null) {
        selectionPort.selectTimeline("availability_reservation_unavailable");
        return;
      }
      const requestOrdinal = requestOrdinalRef.current + 1;
      requestOrdinalRef.current = requestOrdinal;
      const guard: StartupAdmissionGuard = {
        incidentId,
        requestOrdinal,
        selectionVersionAtDispatch,
        availabilityTag,
      };
      const result = await fetchWorkbookJSON<WorkbookStartupEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/workbook-startup${startupQuery}`,
        ),
      );
      if (
        cancelled ||
        !result.ok ||
        requestOrdinalRef.current !== guard.requestOrdinal ||
        guard.incidentId !== incidentId
      ) {
        return;
      }
      const envelope = readEnvelope<WorkbookStartupEnvelope>(result.payload);
      const startupRecord = envelope.data as Record<string, unknown>;
      if (
        !availabilityPort.acceptWorkbookStartup(
          guard.availabilityTag,
          startupRecord.extension_workspace_availability,
        )
      ) {
        onAvailabilityChange();
        selectionPort.selectTimeline("availability_rejected");
        return;
      }
      onAvailabilityChange();
      const startup = normalizeWorkbookStartupSelection(envelope.data);
      if (!startup) {
        return;
      }
      if (
        guard.selectionVersionAtDispatch !==
        selectionPort.readSelectionVersion()
      ) {
        return;
      }
      if (
        startup.selectedSheetRef.kind === "extension_workspace" &&
        !availabilityPort.isRenderable({
          extensionProfileId: startup.selectedSheetRef.extension_profile_id,
          workspaceKey: startup.selectedSheetRef.workspace_key,
        })
      ) {
        selectionPort.selectTimeline("selected_extension_not_renderable");
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
        const contract = workbookContractForViewSchemaId(nextSurface);
        savedViewStatePort.upsertSavedView(startupSavedView);
        savedViewStatePort.applyQueryStateForSurface(
          nextSurface,
          savedViewQueryStateForRuntime(contract, startupSavedView),
        );
        savedViewStatePort.applyLayoutStateForSurface(
          nextSurface,
          workbookLayoutStateFromSavedViewLayoutJson(
            contract,
            startupSavedView.layout_json,
          ),
        );
      }
      selectionPort.applyStartupIdentity({
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
    availabilityPort,
    incidentId,
    onAvailabilityChange,
    savedViewStatePort,
    selectionPort,
    urlParams,
  ]);
}
