import type { ExtensionAvailabilityTag } from "../../extensions/extensionAvailability";
import type { WorkbookIdentity } from "../hooks/useWorkbookStartupController";
import {
  normalizeSavedViewResource,
  type SavedViewResource,
} from "../models/workbookSavedViews";
import type { WorkbookStartupQuery } from "../models/workbookStartup";
import { isStandardizedWorkbookViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookStartupPort } from "./WorkbookStartupPort";

export type StartupFallbackReason =
  | "availability_reservation_unavailable"
  | "availability_rejected"
  | "selected_extension_not_renderable";

export type WorkbookStartupAdmission = {
  readonly availabilityTag: ExtensionAvailabilityTag;
  readonly incidentId: string;
  readonly queryKey: string;
  readonly requestGeneration: number;
  readonly selectionVersion: number;
};

export type WorkbookStartupAdmissionMachine = {
  readonly active: WorkbookStartupAdmission | null;
  readonly requestGeneration: number;
};

export type WorkbookStartupCommitPlan =
  | { readonly kind: "discard" }
  | {
      readonly kind: "fallback";
      readonly reason: StartupFallbackReason;
    }
  | {
      readonly kind: "apply";
      readonly identity: WorkbookIdentity;
      readonly savedView: SavedViewResource | null;
    };

export function initialWorkbookStartupAdmissionMachine(): WorkbookStartupAdmissionMachine {
  return { active: null, requestGeneration: 0 };
}

function queryKey(query: WorkbookStartupQuery): string {
  return [
    query.sheetRefKind ?? "",
    query.sheetRefId ?? "",
    query.extensionProfileId ?? "",
  ].join("\u0000");
}

export function beginWorkbookStartupAdmission(
  machine: WorkbookStartupAdmissionMachine,
  input: {
    readonly availabilityTag: ExtensionAvailabilityTag;
    readonly incidentId: string;
    readonly query: WorkbookStartupQuery;
    readonly selectionVersion: number;
  },
): {
  readonly admission: WorkbookStartupAdmission;
  readonly machine: WorkbookStartupAdmissionMachine;
} {
  const admission = {
    availabilityTag: { ...input.availabilityTag },
    incidentId: input.incidentId,
    queryKey: queryKey(input.query),
    requestGeneration: machine.requestGeneration + 1,
    selectionVersion: input.selectionVersion,
  };
  return {
    admission,
    machine: {
      active: admission,
      requestGeneration: admission.requestGeneration,
    },
  };
}

export function cancelWorkbookStartupAdmission(
  machine: WorkbookStartupAdmissionMachine,
  admission: WorkbookStartupAdmission,
): WorkbookStartupAdmissionMachine {
  return machine.active?.requestGeneration === admission.requestGeneration
    ? { ...machine, active: null }
    : machine;
}

export function workbookStartupAdmissionIsCurrent(
  machine: WorkbookStartupAdmissionMachine,
  admission: WorkbookStartupAdmission,
  current: {
    readonly incidentId: string;
    readonly query: WorkbookStartupQuery;
    readonly selectionVersion: number;
  },
): boolean {
  const active = machine.active;
  return (
    active !== null &&
    active.requestGeneration === admission.requestGeneration &&
    active.incidentId === admission.incidentId &&
    active.queryKey === admission.queryKey &&
    active.availabilityTag.epochId === admission.availabilityTag.epochId &&
    active.availabilityTag.generation ===
      admission.availabilityTag.generation &&
    current.incidentId === admission.incidentId &&
    queryKey(current.query) === admission.queryKey &&
    current.selectionVersion === admission.selectionVersion
  );
}

type AcceptedStartup = Extract<
  Awaited<ReturnType<WorkbookStartupPort["load"]>>,
  { readonly kind: "accepted" }
>["value"];

export function planAcceptedWorkbookStartup(input: {
  readonly availabilityAccepted: boolean;
  readonly extensionRenderable: boolean;
  readonly startup: AcceptedStartup;
}): WorkbookStartupCommitPlan {
  if (!input.availabilityAccepted) {
    return { kind: "fallback", reason: "availability_rejected" };
  }
  const selection = input.startup.selection;
  if (
    selection.selectedSheetRef.kind === "extension_workspace" &&
    !input.extensionRenderable
  ) {
    return {
      kind: "fallback",
      reason: "selected_extension_not_renderable",
    };
  }
  const viewSchemaId =
    selection.selectedSheetRef.kind === "extension_workspace"
      ? null
      : selection.selectedViewSchemaId;
  if (
    selection.selectedSheetRef.kind !== "extension_workspace" &&
    (viewSchemaId === null ||
      !isStandardizedWorkbookViewSchemaId(viewSchemaId) ||
      (selection.selectedSheetRef.kind === "view_schema" &&
        selection.selectedSheetRef.id !== viewSchemaId))
  ) {
    return { kind: "discard" };
  }
  const savedView = normalizeSavedViewResource(selection.selectedSavedView);
  if (
    selection.selectedSheetRef.kind === "saved_view" &&
    (savedView === null ||
      savedView.saved_view_id !== selection.selectedSheetRef.id ||
      savedView.view_schema_id !== viewSchemaId)
  ) {
    return { kind: "discard" };
  }
  return {
    kind: "apply",
    identity: {
      sheetRef: selection.selectedSheetRef,
      viewSchemaId,
    },
    savedView:
      selection.selectedSheetRef.kind === "saved_view" ? savedView : null,
  };
}
