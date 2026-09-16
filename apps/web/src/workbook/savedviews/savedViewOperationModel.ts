import type { AuthorizationRecoveryResult } from "../../shared/authorizationRecovery";
import { type SheetRef, sheetRefKey } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type {
  WorkbookSavedViewLayoutJson,
  WorkbookSavedViewQueryJson,
} from "../models/workbookQuery";
import type { SavedViewResource } from "../models/workbookSavedViews";
import type {
  SavedViewObserver,
  SavedViewProblem,
  WorkbookSavedViewChanges,
  WorkbookSavedViewDefinition,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";
import type { SavedViewDiscoverySnapshot } from "./SavedViewDiscovery";
import type { SavedViewObservation } from "./SavedViewResourceObserver";

export type SavedViewAuthority = {
  readonly incidentId: string;
  readonly actorId: string;
  readonly lifetime: string;
  readonly role: WorkbookIncidentRole;
  readonly apiBase?: string | undefined;
};
export type SavedViewSubject = {
  readonly viewSchemaId: string;
  readonly savedViewId: string | null;
  readonly savedViewVersion: number | null;
};
export type SavedViewDraft = {
  readonly displayName: string;
  readonly scope: "private" | "shared";
  readonly generation: number;
  readonly edited: boolean;
};
export type SavedViewIntent = {
  readonly kind: "create" | "update" | "duplicate" | "delete" | "reset";
};
export type SavedViewAttempt = {
  readonly id: number;
  readonly kind: Exclude<SavedViewIntent["kind"], "reset">;
  readonly authority: SavedViewAuthority;
  readonly subject: SavedViewSubject;
  readonly base: SavedViewResource | null;
  readonly definition: WorkbookSavedViewDefinition;
  readonly changes: WorkbookSavedViewChanges;
  readonly selectionGeneration: number;
  readonly workingGeneration: number;
  readonly formGeneration: number;
};
export type SavedViewOperation =
  | { readonly kind: "idle" }
  | {
      readonly kind: "pending";
      readonly attempt: SavedViewAttempt;
      readonly stage: "authorization" | "write";
    }
  | {
      readonly kind: "confirmed";
      readonly attempt: SavedViewAttempt;
      readonly resource: SavedViewResource | null;
    }
  | {
      readonly kind: "conflict" | "rejected" | "uncertain";
      readonly attempt: SavedViewAttempt;
      readonly problem: SavedViewProblem;
    }
  | { readonly kind: "reviewed"; readonly attempt: SavedViewAttempt };
export type SavedViewSnapshot = {
  readonly authority: SavedViewAuthority | null;
  readonly access: "ready" | "checking" | "unavailable";
  readonly observations: ReadonlyMap<string, SavedViewObservation>;
  readonly discovery: SavedViewDiscoverySnapshot;
  readonly activationId: string | null;
  readonly refreshing: boolean;
  readonly resourceProblem: SavedViewProblem | null;
  readonly observation: number | null;
  readonly transportPending: boolean;
  readonly operation: SavedViewOperation;
  readonly drafts: ReadonlyMap<string, SavedViewDraft>;
  readonly notice: string | null;
  readonly reviewName: string | null;
};
export type SavedViewBinding = {
  readonly apiBase?: string | undefined;
  readonly incidentId: string;
  readonly subject: SavedViewSubject;
  readonly sheetRef: SheetRef;
  readonly selectionGeneration: number;
  readonly workingGeneration: number;
  readonly queryJson: WorkbookSavedViewQueryJson;
  readonly layoutJson: WorkbookSavedViewLayoutJson;
  readonly applyConfiguration: (
    viewSchemaId: string,
    query: WorkbookSavedViewQueryJson,
    layout: WorkbookSavedViewLayoutJson,
  ) => void;
  readonly select: (resource: SavedViewResource) => void;
  readonly deleted: (resource: SavedViewResource) => void;
  readonly unavailable: () => void;
  readonly authorizationRecovered: (
    access: Extract<AuthorizationRecoveryResult, { kind: "authorized" }>,
  ) => void;
};
export type SavedViewControllerPorts = {
  readonly port: (authority: SavedViewAuthority) => WorkbookSavedViewPort;
  readonly isCurrent: (authority: SavedViewAuthority) => boolean;
  readonly observe: SavedViewObserver;
  readonly recover: (
    authority: SavedViewAuthority,
    signal: AbortSignal,
    current: () => boolean,
  ) => Promise<AuthorizationRecoveryResult>;
  readonly lost: (
    reason: "session" | "incident",
    authority: SavedViewAuthority,
  ) => void;
};
export function savedViewSubjectKey(
  subject: Pick<SavedViewSubject, "viewSchemaId" | "savedViewId">,
) {
  return `${subject.viewSchemaId}:${sheetRefKey(subject.savedViewId === null ? { kind: "view_schema", id: subject.viewSchemaId } : { kind: "saved_view", id: subject.savedViewId })}`;
}
export function savedViewOutcome(operation: SavedViewOperation): string {
  switch (operation.kind) {
    case "idle":
      return "";
    case "pending":
      return operation.stage === "authorization"
        ? "Checking access for the saved-view action…"
        : "Saving view configuration…";
    case "conflict":
      return "This saved view changed. Your submitted configuration is retained for review.";
    case "rejected":
      return operation.problem.message;
    case "uncertain":
      return "The saved-view outcome is uncertain. The server may have committed the request. Your submitted configuration is retained.";
    case "reviewed":
      return "Recovery ended by your choice. No receipt was inferred from the resource observation.";
    case "confirmed": {
      if (operation.attempt.kind === "delete")
        return "Saved view deleted. Source rows are unchanged.";
      if (operation.attempt.kind === "update")
        return operation.resource?.saved_view_version ===
          operation.attempt.base?.saved_view_version
          ? "Saved configuration confirmed; no persisted change was needed."
          : "Saved view updated.";
      return operation.attempt.kind === "duplicate"
        ? "Saved view duplicated."
        : "Saved view created.";
    }
  }
}
