import type { InspectorFeatureGroup } from "@cartulary/view-contracts";
import { useCallback } from "react";
import { useContextualCreateAttachment } from "../features/coordination/useContextualCreateAttachment";
import { useCoordinationCreateAttachment } from "../features/coordination/useCoordinationCreateAttachment";
import { useNoteCreateAttachment } from "../features/notes/useNoteCreateAttachment";
import type { WorkbookInspectorLiveRowBinding } from "./workbookInspectorSubject";

/** Presentation attaches to explicit retained owners; unbound additive actions are omitted. */
export function useInspectorCreateRelatedWorkflow({
  selectedSubject,
}: {
  readonly selectedSubject: WorkbookInspectorLiveRowBinding | null;
}) {
  const note = useNoteCreateAttachment(selectedSubject);
  const coordination = useCoordinationCreateAttachment(selectedSubject);
  const contextual = useContextualCreateAttachment(selectedSubject);
  const begin = useCallback(
    (feature: InspectorFeatureGroup): boolean =>
      coordination.begin(feature) ||
      note.begin(feature) ||
      contextual.begin(feature),
    [coordination.begin, note.begin, contextual.begin],
  );
  const cancel = useCallback(() => {
    note.detach();
    coordination.detach();
    contextual.detach();
  }, [note.detach, coordination.detach, contextual.detach]);
  return {
    commands: { begin, cancel },
    snapshot: {
      workflow: coordination.workflow ?? note.workflow ?? contextual.workflow,
    },
  };
}
