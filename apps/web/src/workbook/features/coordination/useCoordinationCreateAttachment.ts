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
import { CoordinationCreateContext } from "./CoordinationCreateContext";
import { coordinationVariant } from "./coordinationCreateModel";

const noSubscribe = () => () => {};
const noSnapshot = () => null;
export function useCoordinationCreateAttachment(
  subject: WorkbookInspectorLiveRowBinding | null,
  open = true,
) {
  const context = useContext(CoordinationCreateContext);
  const token = useRef(Symbol("coordination-inspector")).current;
  const current = useRef({ context, subject });
  current.current = { context, subject };
  const state = useSyncExternalStore(
    context?.owner.subscribe ?? noSubscribe,
    context?.owner.getSnapshot ?? noSnapshot,
  );
  const sourceId = subject?.subject.recordId;
  const sourceVersion = subject?.subject.rowVersion;
  const sourceView = subject?.subject.viewSchemaId;
  const sheet = JSON.stringify(context?.sheetRef);
  useEffect(() => {
    if (!context) return;
    const origin = context.owner.getSnapshot().draft?.origin;
    if (
      !open ||
      (origin &&
        (origin.subject?.recordId !== sourceId ||
          origin.subject?.viewSchemaId !== sourceView ||
          JSON.stringify(origin.sheetRef) !== sheet))
    )
      context.owner.detach(token);
    if (sourceId && sourceVersion)
      context.owner.observe(sourceId, sourceVersion);
  }, [context, open, sourceId, sourceVersion, sourceView, sheet, token]);
  useEffect(() => () => context?.owner.detach(token), [context?.owner, token]);
  const draft = state?.draft;
  const workflow: InspectorRelatedRecordWorkflowState | null =
    draft?.origin.subject && draft.origin.feature && state?.attachment === token
      ? {
          workflowId: token,
          subject: draft.origin.subject,
          draft: Object.fromEntries(
            Object.entries(draft.values).map(([key, value]) => [
              key,
              value ?? "",
            ]),
          ),
          targetContract: draft.target,
          featureGroup: draft.origin.feature,
          error: null,
          phase: "editing",
        }
      : null;
  const begin = useCallback(
    (feature: InspectorFeatureGroup) => {
      if (!coordinationVariant(feature.featureGroupKey)) return false;
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
  const actions = context?.owner.captureDraftActions(token);
  const update = (key: string, value: string) => actions?.update(key, value);
  return { workflow, begin, detach, update };
}
