import type { InspectorPanelId } from "@cartulary/view-contracts";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookInspectorTechnicalField } from "./presentation/workbookInspectorPresentationModel";

type WorkbookInspectorNoticeDestination =
  | {
      readonly kind: "field";
      readonly panel: "details";
      readonly fieldKey: string;
      readonly action: string;
      readonly revision: number;
    }
  | {
      readonly kind: "relationship_item";
      readonly panel: "relationships";
      readonly fieldKey: string;
      readonly itemRef: string;
    }
  | {
      readonly kind: "region";
      readonly panel: InspectorPanelId;
      readonly regionId: string;
    }
  | {
      readonly kind: "feature";
      readonly panel: InspectorPanelId;
      readonly featureGroupKey: string;
    }
  | { readonly kind: "panel"; readonly panel: InspectorPanelId }
  | { readonly kind: "inspector" };

export type WorkbookInspectorNotice = {
  readonly context: {
    readonly authority: string;
    readonly subject:
      | {
          readonly kind: "record";
          readonly viewSchemaId: string;
          readonly recordId: string;
        }
      | {
          readonly kind: "creation";
          readonly viewSchemaId: string;
          readonly attachment: string;
        };
  };
  readonly destination: WorkbookInspectorNoticeDestination;
  readonly attemptId: string;
  readonly transitionId: string;
  readonly feedback: WorkbookInspectorFeedback;
  readonly announcement: "none" | "polite" | "assertive";
};

/** Retained by the producing owner, never by a mounted inspector. */
export class WorkbookInspectorNoticeLedger {
  private readonly emitted = new Set<string>();
  consume = (notice: WorkbookInspectorNotice): boolean => {
    const key = workbookInspectorNoticeIdentity(notice);
    if (notice.announcement === "none" || this.emitted.has(key)) return false;
    this.emitted.add(key);
    return true;
  };
  clear() {
    this.emitted.clear();
  }
}

export function workbookInspectorNoticeIdentity(
  notice: WorkbookInspectorNotice,
) {
  return JSON.stringify([
    notice.context,
    notice.destination,
    notice.attemptId,
    notice.transitionId,
  ]);
}

export type WorkbookInspectorErrorPresentation = {
  readonly primaryMessage: string;
  readonly technicalFields: readonly WorkbookInspectorTechnicalField[];
};

export type WorkbookInspectorFeedback = (
  | {
      readonly kind: "message";
      readonly message: string;
      readonly announcement: "none" | "polite";
    }
  | {
      readonly kind: "error";
      readonly error: WorkbookInspectorErrorPresentation;
    }
) & {
  readonly destination?: WorkbookInspectorNoticeDestination;
  readonly sourceRecordId?: string;
};

export function targetWorkbookInspectorFeedback(
  feedback: WorkbookInspectorFeedback,
  sourceRecordId: string,
  destination: WorkbookInspectorNoticeDestination,
): WorkbookInspectorFeedback {
  return { ...feedback, sourceRecordId, destination };
}

const rowVersionConflictMessage =
  "This row changed; refresh it before retrying.";

export function workbookInspectorErrorPresentation(
  failure: WorkbookOperationFailure,
): WorkbookInspectorErrorPresentation {
  if (failure.publicCode === "row_version_conflict") {
    return {
      primaryMessage: rowVersionConflictMessage,
      technicalFields: [
        { label: "Public error code", value: failure.publicCode },
        { label: "Server message", value: failure.message },
      ],
    };
  }
  return {
    primaryMessage: failure.message,
    technicalFields:
      failure.publicCode === undefined
        ? []
        : [{ label: "Public error code", value: failure.publicCode }],
  };
}

export function workbookInspectorMessageFeedback(
  message: string,
  announcement: "none" | "polite",
): WorkbookInspectorFeedback {
  return { announcement, kind: "message", message };
}

export function workbookInspectorOperationFailureFeedback(
  failure: WorkbookOperationFailure,
): WorkbookInspectorFeedback {
  return { error: workbookInspectorErrorPresentation(failure), kind: "error" };
}

export function workbookInspectorLocalErrorFeedback(
  message: string,
): WorkbookInspectorFeedback {
  return {
    error: workbookInspectorLocalErrorPresentation(message),
    kind: "error",
  };
}

export function workbookInspectorLocalErrorPresentation(
  message: string,
): WorkbookInspectorErrorPresentation {
  return { primaryMessage: message, technicalFields: [] };
}
