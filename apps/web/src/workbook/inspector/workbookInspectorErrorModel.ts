import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookInspectorTechnicalField } from "./presentation/workbookInspectorPresentationModel";

export type WorkbookInspectorErrorPresentation = {
  readonly primaryMessage: string;
  readonly technicalFields: readonly WorkbookInspectorTechnicalField[];
};

export type WorkbookInspectorFeedback =
  | string
  | WorkbookInspectorErrorPresentation;

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

export function workbookInspectorLocalErrorPresentation(
  message: string,
): WorkbookInspectorErrorPresentation {
  return { primaryMessage: message, technicalFields: [] };
}

export function workbookInspectorFeedbackPresentation(
  feedback: WorkbookInspectorFeedback,
): WorkbookInspectorErrorPresentation {
  return typeof feedback === "string"
    ? workbookInspectorLocalErrorPresentation(feedback)
    : feedback;
}
