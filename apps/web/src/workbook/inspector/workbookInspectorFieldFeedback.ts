import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { InspectorEditIdentity } from "./WorkbookInspectorDraftStore";

export type InspectorFieldFeedbackIdentity = InspectorEditIdentity & {
  readonly revision: number;
};

export function inspectorFieldFeedback(
  failure: WorkbookOperationFailure,
  captured: InspectorFieldFeedbackIdentity,
  current: InspectorFieldFeedbackIdentity | null,
): string | null {
  if (
    !current ||
    failure.kind !== "validation" ||
    captured.viewSchemaId !== current.viewSchemaId ||
    captured.recordId !== current.recordId ||
    captured.fieldKey !== current.fieldKey ||
    captured.action !== current.action ||
    captured.revision !== current.revision
  )
    return null;
  return (
    failure.fields?.find((item) => item.field === captured.fieldKey)?.message ??
    null
  );
}
