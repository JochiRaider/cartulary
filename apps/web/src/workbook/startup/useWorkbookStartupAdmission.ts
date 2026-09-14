import { useEffect, useRef, useState } from "react";
import type {
  ExtensionAvailabilityTag,
  ExtensionWorkspaceIdentity,
} from "../../extensions/extensionAvailability";
import type { WorkbookIdentity } from "../hooks/useWorkbookStartupController";
import {
  type WorkbookLayoutState,
  type WorkbookQueryState,
  workbookLayoutStateFromSavedViewLayoutJson,
} from "../models/workbookQuery";
import { savedViewQueryStateForRuntime } from "../models/workbookSavedViewRuntime";
import type { SavedViewResource } from "../models/workbookSavedViews";
import { workbookStartupQueryFromURLParams } from "../models/workbookStartup";
import { workbookContractForViewSchemaId } from "../models/workbookSurfaceQueryRuntime";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import type {
  WorkbookStartupAvailability,
  WorkbookStartupPort,
} from "./WorkbookStartupPort";
import {
  beginWorkbookStartupAdmission,
  cancelWorkbookStartupAdmission,
  initialWorkbookStartupAdmissionMachine,
  planAcceptedWorkbookStartup,
  type StartupFallbackReason,
  type WorkbookStartupCommitPlan,
  workbookStartupAdmissionIsCurrent,
} from "./workbookStartupAdmissionMachine";

export interface WorkbookStartupAvailabilityPort {
  waitForDiscovery(signal: AbortSignal): Promise<boolean>;
  reserve(): ExtensionAvailabilityTag | null;
  acceptWorkbookStartup(
    tag: ExtensionAvailabilityTag,
    availability: WorkbookStartupAvailability,
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

function applyWorkbookStartupCommitPlan(input: {
  readonly onAvailabilityChange: () => void;
  readonly plan: WorkbookStartupCommitPlan;
  readonly savedViewStatePort: WorkbookStartupSavedViewStatePort;
  readonly selectionPort: WorkbookStartupSelectionPort;
}): void {
  input.onAvailabilityChange();
  const plan = input.plan;
  if (plan.kind === "discard") return;
  if (plan.kind === "fallback") {
    input.selectionPort.selectTimeline(plan.reason);
    return;
  }
  if (plan.savedView !== null && plan.identity.viewSchemaId !== null) {
    const contract = workbookContractForViewSchemaId(
      plan.identity.viewSchemaId,
    );
    input.savedViewStatePort.upsertSavedView(plan.savedView);
    input.savedViewStatePort.applyQueryStateForSurface(
      plan.identity.viewSchemaId,
      savedViewQueryStateForRuntime(contract, plan.savedView),
    );
    input.savedViewStatePort.applyLayoutStateForSurface(
      plan.identity.viewSchemaId,
      workbookLayoutStateFromSavedViewLayoutJson(
        contract,
        plan.savedView.layout_json,
      ),
    );
  }
  input.selectionPort.applyStartupIdentity(plan.identity);
}

export function useWorkbookStartupAdmission({
  incidentId,
  urlParams,
  availabilityPort,
  selectionPort,
  savedViewStatePort,
  startupPort,
  onAuthorityUncertain,
  onAvailabilityChange,
}: {
  readonly incidentId: string;
  readonly urlParams: URLSearchParams;
  readonly availabilityPort: WorkbookStartupAvailabilityPort;
  readonly selectionPort: WorkbookStartupSelectionPort;
  readonly savedViewStatePort: WorkbookStartupSavedViewStatePort;
  readonly startupPort: WorkbookStartupPort;
  readonly onAuthorityUncertain?: (() => void) | undefined;
  readonly onAvailabilityChange: () => void;
}): { readonly pending: boolean } {
  const [pending, setPending] = useState(true);
  const launchQuery = urlParams.toString();
  const admissionMachineRef = useRef(initialWorkbookStartupAdmissionMachine());

  useEffect(() => {
    const controller = new AbortController();
    setPending(true);
    const startupQuery = workbookStartupQueryFromURLParams(
      new URLSearchParams(launchQuery),
    );
    const loadStartup = async () => {
      const availabilityTag = availabilityPort.reserve();
      const started = beginWorkbookStartupAdmission(
        admissionMachineRef.current,
        {
          availabilityTag,
          incidentId,
          query: startupQuery,
          selectionVersion: selectionPort.readSelectionVersion(),
        },
      );
      admissionMachineRef.current = started.machine;
      const result = await startupPort.load({
        query: startupQuery,
        signal: controller.signal,
      });
      const isCurrent = () =>
        workbookStartupAdmissionIsCurrent(
          admissionMachineRef.current,
          started.admission,
          {
            incidentId,
            query: startupQuery,
            selectionVersion: selectionPort.readSelectionVersion(),
          },
        );
      if (
        controller.signal.aborted ||
        result.kind === "aborted" ||
        !isCurrent()
      ) {
        return;
      }
      if (result.kind === "rejected") {
        admissionMachineRef.current = cancelWorkbookStartupAdmission(
          admissionMachineRef.current,
          started.admission,
        );
        if (
          workbookFailureLifecycle(result.failure).kind ===
          "authority_unavailable"
        ) {
          onAuthorityUncertain?.();
        }
        return;
      }
      const selectedSheetRef = result.value.selection.selectedSheetRef;
      if (
        selectedSheetRef.kind === "extension_workspace" &&
        !(await availabilityPort.waitForDiscovery(controller.signal))
      )
        return;
      if (controller.signal.aborted || !isCurrent()) return;
      const availabilityAccepted =
        started.admission.availabilityTag !== null &&
        availabilityPort.acceptWorkbookStartup(
          started.admission.availabilityTag,
          result.value.availability,
        );
      if (!isCurrent()) return;
      const plan = planAcceptedWorkbookStartup({
        availabilityAccepted,
        extensionRenderable:
          selectedSheetRef.kind !== "extension_workspace" ||
          availabilityPort.isRenderable({
            extensionProfileId: selectedSheetRef.extension_profile_id,
            workspaceKey: selectedSheetRef.workspace_key,
          }),
        startup: result.value,
      });
      admissionMachineRef.current = cancelWorkbookStartupAdmission(
        admissionMachineRef.current,
        started.admission,
      );
      applyWorkbookStartupCommitPlan({
        onAvailabilityChange,
        plan,
        savedViewStatePort,
        selectionPort,
      });
    };

    void loadStartup().finally(() => {
      if (!controller.signal.aborted) setPending(false);
    });
    return () => {
      controller.abort();
      const active = admissionMachineRef.current.active;
      if (active !== null) {
        admissionMachineRef.current = cancelWorkbookStartupAdmission(
          admissionMachineRef.current,
          active,
        );
      }
    };
  }, [
    availabilityPort,
    incidentId,
    onAvailabilityChange,
    onAuthorityUncertain,
    savedViewStatePort,
    selectionPort,
    startupPort,
    launchQuery,
  ]);
  return { pending };
}
