import type { InspectorFeatureGroup } from "@cartulary/view-contracts";
import {
  useCallback,
  useContext,
  useEffect,
  useLayoutEffect,
  useRef,
  useSyncExternalStore,
} from "react";
import type { InspectorRelatedRecordWorkflowState } from "../../inspector/inspectorRelatedRecordModel";
import type { WorkbookInspectorLiveRowBinding } from "../../inspector/workbookInspectorSubject";
import { TimelineRelatedEvidenceContext } from "./TimelineRelatedEvidenceContext";
import { relatedEvidenceFeature } from "./timelineRelatedEvidenceModel";

const noSubscribe = () => () => {},
  noSnapshot = () => null;
export function useTimelineRelatedEvidenceAttachment(
  subject: WorkbookInspectorLiveRowBinding | null,
  isOpen = true,
) {
  const context = useContext(TimelineRelatedEvidenceContext),
    token = useRef(Symbol("related-evidence-inspector")).current;
  const current = useRef({ context, subject });
  current.current = { context, subject };
  const state = useSyncExternalStore(
    context?.owner.subscribe ?? noSubscribe,
    context?.owner.getSnapshot ?? noSnapshot,
  );
  const sourceId = subject?.subject.recordId,
    version = subject?.subject.rowVersion,
    sheet = JSON.stringify(context?.sheetRef);
  const attachedOrigin = useRef<{ id: string; sheet: string } | null>(null);
  useLayoutEffect(() => {
    const origin = attachedOrigin.current;
    if (origin && (!isOpen || sourceId !== origin.id || sheet !== origin.sheet))
      context?.owner.detach(token);
    if (sourceId && version) context?.owner.observe(sourceId, version);
  }, [context?.owner, sourceId, version, sheet, token, isOpen]);
  useEffect(() => () => context?.owner.detach(token), [context?.owner, token]);
  const draft = state?.draft;
  const workflow: InspectorRelatedRecordWorkflowState | null =
    draft && state.attachment === token
      ? {
          workflowId: token,
          subject: {
            kind: "live",
            ...draft.source,
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
      if (feature.featureGroupKey !== relatedEvidenceFeature) return false;
      const { context, subject } = current.current;
      if (
        context &&
        subject &&
        context.owner.begin(subject, feature, context.sheetRef, token)
      )
        attachedOrigin.current = {
          id: subject.subject.recordId,
          sheet: JSON.stringify(context.sheetRef),
        };
      return true;
    },
    [token],
  );
  const detach = useCallback(
    () => current.current.context?.owner.detach(token),
    [token],
  );
  const notice = useCallback(
    () => current.current.context?.owner.getSnapshot().message ?? null,
    [],
  );
  return { workflow, begin, detach, notice };
}
