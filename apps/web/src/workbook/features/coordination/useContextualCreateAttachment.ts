import type { InspectorFeatureGroup } from "@cartulary/view-contracts";
import {
  useCallback,
  useContext,
  useEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import { ContextualCreateContext } from "./ContextualCreateContext";
import { isContextualCreateFeature } from "./contextualCreateModel";

const noSubscribe = () => () => {};
const noSnapshot = () => null;
export function useContextualCreateAttachment(
  subject: WorkbookInspectorLiveRowBinding | null,
) {
  const context = useContext(ContextualCreateContext);
  const token = useRef(Symbol("contextual-create-inspector")).current;
  const current = useRef({ context, subject });
  current.current = { context, subject };
  const snapshot = useSyncExternalStore(
    context?.owner.subscribe ?? noSubscribe,
    context?.owner.getSnapshot ?? noSnapshot,
  );
  const sourceId = subject?.subject.recordId;
  const sourceView = subject?.subject.viewSchemaId;
  const sourceVersion = subject?.subject.rowVersion;
  const sheet = JSON.stringify(context?.sheetRef);
  useEffect(() => {
    if (!context) return;
    const draft = context.owner.getSnapshot().draft;
    if (
      draft &&
      (draft.source.recordId !== sourceId ||
        draft.source.viewSchemaId !== sourceView ||
        JSON.stringify(draft.presentation.sheetRef) !== sheet)
    )
      context.owner.detach(token);
    if (sourceId && sourceVersion)
      context.owner.observe(sourceId, sourceVersion);
  }, [context, sourceId, sourceVersion, sourceView, sheet, token]);
  useEffect(() => () => context?.owner.detach(token), [context?.owner, token]);
  const draft = snapshot?.draft;
  const workflow: InspectorRelatedRecordWorkflowState | null =
    draft && snapshot?.attachment === token
      ? {
          workflowId: token,
          subject: {
            kind: "live",
            recordId: draft.source.recordId,
            rowVersion: draft.source.rowVersion,
            viewSchemaId: draft.source.viewSchemaId,
            label: draft.presentation.label,
            surfaceLabel: draft.presentation.surfaceLabel,
          },
          draft: { ...draft.values },
          targetContract: draft.target,
          featureGroup: draft.feature,
          error: null,
          phase: "editing",
        }
      : null;
  const begin = useCallback(
    (feature: InspectorFeatureGroup) => {
      if (!isContextualCreateFeature(feature.featureGroupKey)) return false;
      const { context, subject } = current.current;
      if (context && subject)
        context.owner.begin(subject, feature, context.sheetRef, token);
      return true;
    },
    [token],
  );
  const detach = useCallback(
    () => current.current.context?.owner.detach(token),
    [token],
  );
  const update = useCallback(
    (field: string, value: string) =>
      current.current.context?.owner.update(field, value),
    [],
  );
  return { workflow, begin, detach, update };
}
