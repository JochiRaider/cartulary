import { useRef, useState } from "react";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookInspectorEditDraft } from "./useWorkbookInspectorEditDraft";
import {
  type WorkbookInspectorErrorPresentation,
  workbookInspectorErrorPresentation,
} from "./workbookInspectorErrorModel";
import {
  type InspectorFieldFeedbackIdentity,
  inspectorFieldFeedback,
} from "./workbookInspectorFieldFeedback";

// Feedback belongs to the captured authoring revision. Rendering a different
// attachment or newer draft never reassigns an asynchronous failure.
export function useWorkbookInspectorFieldFeedback(
  edit: WorkbookInspectorEditDraft,
) {
  const [feedback, setFeedback] = useState<{
    identity: InspectorFieldFeedbackIdentity;
    attachment: string;
    failure: WorkbookOperationFailure;
  } | null>(null);
  const current = { ...edit.identity, revision: edit.draft?.revision ?? 0 };
  const token = JSON.stringify([edit.attachment, current.revision]);
  const latest = useRef(token);
  latest.current = token;
  const visible =
    feedback?.attachment === edit.attachment &&
    feedback.identity.revision === current.revision &&
    edit.canResume
      ? feedback
      : null;
  const message = visible
    ? inspectorFieldFeedback(visible.failure, visible.identity, current)
    : null;
  const actionError: WorkbookInspectorErrorPresentation | null =
    visible && !message
      ? workbookInspectorErrorPresentation(visible.failure)
      : null;
  return {
    message,
    actionError,
    clear: () => setFeedback(null),
    capture: () => {
      const identity = current;
      const attachment = edit.attachment;
      return (failure: WorkbookOperationFailure) => {
        if (latest.current === token)
          setFeedback({ identity, attachment, failure });
      };
    },
    rejectLocal: (message: string, field = true) =>
      setFeedback({
        identity: current,
        attachment: edit.attachment,
        failure: {
          kind: "validation",
          message,
          ...(field ? { fields: [{ field: current.fieldKey, message }] } : {}),
        },
      }),
  };
}
