import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useState } from "react";
import {
  buildInspectorRelatedRecordDraft,
  type InspectorRelatedRecordFormModel,
} from "../../inspector/inspectorRelatedRecordModel";
import {
  type WorkbookInspectorFeedback,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "../../inspector/workbookInspectorErrorModel";
import { genericCreateMinimumMessage } from "../../models/genericWorkbookModel";
import { evidenceViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  TimelineRelatedEvidenceLinked,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export type TimelineCreateRelatedWorkflowState =
  InspectorRelatedRecordFormModel & {
    readonly sourceRowKey: string;
  };

export function useTimelineCreateRelatedWorkflow({
  applyAcceptedRowMutation,
  currentUserId,
  loadRows,
  mutationCommands,
  selectedRow,
  selectedRowWorkflowKey,
  setInspectorMessage,
  targetContracts,
}: {
  readonly applyAcceptedRowMutation: (
    rowKey: string,
    accepted: Pick<TimelineRelatedEvidenceLinked, "row" | "viewSchemaId">,
  ) => unknown;
  readonly currentUserId: string | null;
  readonly loadRows: (options: {
    readonly showLoading: boolean;
  }) => Promise<void>;
  readonly mutationCommands: TimelineRelatedRecordPort;
  readonly selectedRow: WorkbookRow | null;
  readonly selectedRowWorkflowKey: string;
  readonly setInspectorMessage: (
    message: WorkbookInspectorFeedback | null,
  ) => void;
  readonly targetContracts: ReadonlyMap<string, ViewContract>;
}) {
  const [workflow, setWorkflow] =
    useState<TimelineCreateRelatedWorkflowState | null>(null);

  useEffect(() => {
    setWorkflow((current) =>
      current === null || current.sourceRowKey === selectedRowWorkflowKey
        ? current
        : null,
    );
  }, [selectedRowWorkflowKey]);

  const cancelWorkflow = useCallback(() => {
    setWorkflow(null);
  }, []);

  const beginWorkflow = useCallback(
    (featureGroup: InspectorFeatureGroup) => {
      if (
        featureGroup.routeBinding.kind !== "view_row_create" ||
        featureGroup.routeBinding.owner !== "view_row_create_route"
      ) {
        setInspectorMessage("Inspector action is unavailable.");
        return;
      }
      const targetViewSchemaId = featureGroup.routeBinding.targetViewSchemaId;
      const targetContract =
        targetViewSchemaId === undefined
          ? undefined
          : targetContracts.get(targetViewSchemaId);
      if (targetContract === undefined) {
        setInspectorMessage("Inspector action is unavailable.");
        return;
      }
      if (
        selectedRow?.recordId === null ||
        selectedRow?.recordId === undefined
      ) {
        setInspectorMessage("Select a row before creating a related record.");
        return;
      }
      const result = buildInspectorRelatedRecordDraft({
        currentUserId,
        featureGroup,
        subject: {
          cells: selectedRow.rawRow?.cells ?? {},
          recordId: selectedRow.recordId,
        },
        targetContract,
      });
      if (result.kind === "invalid_target") {
        setInspectorMessage("The target view does not allow row creation.");
        return;
      }
      setWorkflow({
        featureGroup,
        targetContract,
        sourceRowKey: selectedRowWorkflowKey,
        draft: result.draft,
        error: null,
        isSubmitting: false,
      });
      setInspectorMessage(null);
    },
    [
      currentUserId,
      selectedRow,
      selectedRowWorkflowKey,
      setInspectorMessage,
      targetContracts,
    ],
  );

  const updateWorkflowDraft = useCallback(
    (featureGroupKey: string, fieldKey: string, value: string) => {
      setWorkflow((current) =>
        current?.featureGroup.featureGroupKey === featureGroupKey
          ? {
              ...current,
              draft: {
                ...current.draft,
                [fieldKey]: value,
              },
              error: null,
            }
          : current,
      );
    },
    [],
  );

  const submitWorkflow = useCallback(async () => {
    const activeWorkflow = workflow;
    const sourceRow = selectedRow;
    if (
      activeWorkflow === null ||
      sourceRow === null ||
      sourceRow.recordId === null
    ) {
      return;
    }
    setWorkflow({
      ...activeWorkflow,
      isSubmitting: true,
      error: null,
    });
    const createResult = await mutationCommands.createRelatedRecord({
      contract: activeWorkflow.targetContract,
      draft: activeWorkflow.draft,
      featureGroupKey: activeWorkflow.featureGroup.featureGroupKey,
    });
    if (createResult.kind === "rejected") {
      setWorkflow({
        ...activeWorkflow,
        isSubmitting: false,
        error:
          createResult.failure.kind === "validation"
            ? workbookInspectorLocalErrorPresentation(
                genericCreateMinimumMessage(activeWorkflow.targetContract),
              )
            : workbookInspectorErrorPresentation(createResult.failure),
      });
      return;
    }
    const createdRecordId = createResult.value.recordId;

    if (activeWorkflow.targetContract.viewSchemaId === evidenceViewSchemaId) {
      const patchResult = await mutationCommands.linkCreatedEvidence({
        sourceRow,
        createdRecordId,
      });
      if (patchResult.kind === "rejected") {
        setWorkflow({
          ...activeWorkflow,
          isSubmitting: false,
          error: workbookInspectorErrorPresentation(patchResult.failure),
        });
        return;
      }
      applyAcceptedRowMutation(sourceRow.key, patchResult.value);
      await loadRows({ showLoading: false });
      setWorkflow(null);
      setInspectorMessage(`Created and linked evidence ${createdRecordId}.`);
      return;
    }

    setWorkflow(null);
    setInspectorMessage(
      `Created related ${activeWorkflow.targetContract.viewSchemaId} row ${createdRecordId}.`,
    );
  }, [
    applyAcceptedRowMutation,
    loadRows,
    mutationCommands,
    selectedRow,
    setInspectorMessage,
    workflow,
  ]);

  return {
    beginWorkflow,
    cancelWorkflow,
    submitWorkflow,
    updateWorkflowDraft,
    workflow,
  };
}
