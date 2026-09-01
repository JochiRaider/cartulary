import {
  getViewContract,
  type InspectorFeatureGroup,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useState } from "react";
import { genericCreateMinimumMessage } from "../models/genericWorkbookModel";
import type { TimelineRelatedRecordPort } from "../mutations/workbookMutationCommandPorts";
import {
  buildInspectorRelatedRecordDraft,
  type InspectorRelatedRecordFormModel,
} from "./inspectorRelatedRecordModel";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "./workbookInspectorErrorModel";

export type InspectorCreateRelatedSubject = {
  readonly cells: Readonly<Record<string, { readonly value: unknown }>>;
  readonly recordId: string;
  readonly rowVersion: number;
};

export type InspectorCreateRelatedWorkflowState =
  InspectorRelatedRecordFormModel & {
    readonly sourceRecordId: string;
  };

export function useInspectorCreateRelatedWorkflow({
  currentUserId,
  mutationCommands,
  onCreated,
  onMessage,
  selectedSubject,
}: {
  readonly currentUserId: string | null;
  readonly mutationCommands: TimelineRelatedRecordPort;
  readonly onCreated: () => Promise<void> | void;
  readonly onMessage: (message: string | null) => void;
  readonly selectedSubject: InspectorCreateRelatedSubject | null;
}) {
  const [workflow, setWorkflow] =
    useState<InspectorCreateRelatedWorkflowState | null>(null);

  useEffect(() => {
    if (
      workflow !== null &&
      workflow.sourceRecordId !== selectedSubject?.recordId
    ) {
      setWorkflow(null);
    }
  }, [selectedSubject?.recordId, workflow]);

  const begin = useCallback(
    (featureGroup: InspectorFeatureGroup): boolean => {
      if (
        featureGroup.routeBinding.kind !== "view_row_create" ||
        featureGroup.routeBinding.owner !== "view_row_create_route" ||
        featureGroup.routeBinding.targetViewSchemaId === undefined
      ) {
        return false;
      }
      const targetContract = getViewContract(
        featureGroup.routeBinding.targetViewSchemaId,
      );
      if (targetContract === undefined) {
        onMessage("The target view does not allow row creation.");
        return true;
      }
      if (selectedSubject === null) {
        onMessage("Select a saved row before creating a related record.");
        return true;
      }
      const result = buildInspectorRelatedRecordDraft({
        currentUserId,
        featureGroup,
        subject: selectedSubject,
        targetContract,
      });
      if (result.kind === "invalid_target") {
        onMessage("The target view does not allow row creation.");
        return true;
      }
      setWorkflow({
        draft: result.draft,
        error: null,
        featureGroup,
        isSubmitting: false,
        sourceRecordId: selectedSubject.recordId,
        targetContract,
      });
      onMessage(null);
      return true;
    },
    [currentUserId, onMessage, selectedSubject],
  );

  const updateDraft = useCallback((fieldKey: string, value: string) => {
    setWorkflow((current) =>
      current === null
        ? null
        : {
            ...current,
            draft: { ...current.draft, [fieldKey]: value },
            error: null,
          },
    );
  }, []);

  const cancel = useCallback(() => setWorkflow(null), []);

  const submit = useCallback(async () => {
    const active = workflow;
    if (active === null) return;
    setWorkflow({ ...active, isSubmitting: true, error: null });
    const outcome = await mutationCommands.createRelatedRecord({
      contract: active.targetContract,
      draft: active.draft,
      featureGroupKey: active.featureGroup.featureGroupKey,
    });
    if (outcome.kind === "rejected") {
      setWorkflow({
        ...active,
        isSubmitting: false,
        error:
          outcome.failure.kind === "validation"
            ? workbookInspectorLocalErrorPresentation(
                genericCreateMinimumMessage(active.targetContract),
              )
            : workbookInspectorErrorPresentation(outcome.failure),
      });
      return;
    }
    setWorkflow(null);
    onMessage(
      `Created ${active.targetContract.title} record ${outcome.value.recordId}.`,
    );
    await onCreated();
  }, [mutationCommands, onCreated, onMessage, workflow]);

  return {
    commands: { begin, cancel, submit, updateDraft },
    snapshot: { workflow },
  };
}
