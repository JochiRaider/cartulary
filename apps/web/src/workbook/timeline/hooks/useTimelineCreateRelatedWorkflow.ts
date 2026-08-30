import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useState } from "react";
import {
  genericCreateMinimumMessage,
  initialGenericCreateDraft,
  workbookCreationAvailable,
} from "../../models/genericWorkbookModel";
import { evidenceViewSchemaId } from "../../models/workbookSurfaceRegistry";
import type {
  TimelineRelatedEvidenceLinked,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import type { WorkbookRow } from "../models/workbookTimelineModel";

export type TimelineCreateRelatedWorkflowState = {
  readonly featureGroup: InspectorFeatureGroup;
  readonly targetContract: ViewContract;
  readonly sourceRowKey: string;
  readonly draft: Record<string, string>;
  readonly isSubmitting: boolean;
  readonly message: string | null;
};

function applyTimelineCreateRelatedSeedBindings(
  draft: Record<string, string>,
  featureGroup: InspectorFeatureGroup,
  selectedRow: WorkbookRow,
): Record<string, string> {
  const next = { ...draft };
  for (const binding of featureGroup.seedBindings) {
    const value = timelineCreateRelatedSeedValue(binding.source, selectedRow);
    if (value !== null) {
      next[binding.targetFieldKey] = value;
    }
  }
  return next;
}

function timelineCreateRelatedSeedValue(
  source: InspectorFeatureGroup["seedBindings"][number]["source"],
  selectedRow: WorkbookRow,
): string | null {
  switch (source.kind) {
    case "selected_record_id":
      return selectedRow.recordId;
    case "selected_field_value": {
      if (source.sourceFieldKey === undefined) {
        return null;
      }
      if (selectedRow.rawRow === null) {
        return null;
      }
      const value = selectedRow.rawRow.cells[source.sourceFieldKey]?.value;
      const text = stringifyGridValue(value).trim();
      return text === "" ? null : text;
    }
    case "literal":
      if (source.value === null || source.value === undefined) {
        return null;
      }
      return typeof source.value === "string"
        ? source.value
        : JSON.stringify(source.value);
  }
}

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
  readonly setInspectorMessage: (message: string | null) => void;
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
      if (!workbookCreationAvailable(targetContract)) {
        setInspectorMessage("The target view does not allow row creation.");
        return;
      }
      if (
        selectedRow?.recordId === null ||
        selectedRow?.recordId === undefined
      ) {
        setInspectorMessage("Select a row before creating a related record.");
        return;
      }
      const seededDraft = applyTimelineCreateRelatedSeedBindings(
        initialGenericCreateDraft(targetContract, currentUserId),
        featureGroup,
        selectedRow,
      );
      setWorkflow({
        featureGroup,
        targetContract,
        sourceRowKey: selectedRowWorkflowKey,
        draft: seededDraft,
        isSubmitting: false,
        message: null,
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
              message: null,
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
      message: null,
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
        message:
          createResult.failure.kind === "validation"
            ? genericCreateMinimumMessage(activeWorkflow.targetContract)
            : createResult.failure.message,
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
          message: `Created evidence, but Timeline link failed: ${patchResult.failure.message}`,
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
