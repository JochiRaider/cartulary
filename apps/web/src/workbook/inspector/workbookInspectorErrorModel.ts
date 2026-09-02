import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookInspectorTechnicalField } from "./presentation/workbookInspectorPresentationModel";

export type WorkbookInspectorErrorPresentation = {
  readonly primaryMessage: string;
  readonly technicalFields: readonly WorkbookInspectorTechnicalField[];
};

export type WorkbookInspectorFeedback =
  | {
      readonly kind: "message";
      readonly message: string;
      readonly announcement: "none" | "polite";
    }
  | {
      readonly kind: "error";
      readonly error: WorkbookInspectorErrorPresentation;
    };

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
