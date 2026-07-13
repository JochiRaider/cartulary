import type {
  InspectorFeatureGroup,
  ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useState } from "react";
import { apiPath } from "../../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../../services/workbookApi";
import {
  buildGenericCreatePayload,
  genericCreateMinimumMessage,
  initialGenericCreateDraft,
} from "../../models/genericWorkbookModel";
import { evidenceViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { stringifyGridValue } from "../../utils/workbookValueFormat";
import {
  buildAttachedEvidencePatchPayload,
  type WorkbookRow,
} from "../models/workbookTimelineModel";
import type { TimelineMutationEnvelope } from "../services/timelineMutationRequests";

export type TimelineCreateRelatedWorkflowState = {
  readonly featureGroup: InspectorFeatureGroup;
  readonly targetContract: ViewContract;
  readonly sourceRowKey: string;
  readonly draft: Record<string, string>;
  readonly isSubmitting: boolean;
  readonly message: string | null;
};

function recordIdFromMutationEnvelope(
  envelope: TimelineMutationEnvelope,
): string | null {
  const row = envelope.data.row;
  if (row && typeof row === "object" && "record_id" in row) {
    const recordId = (row as { record_id?: unknown }).record_id;
    return typeof recordId === "string" && recordId.trim() !== ""
      ? recordId
      : null;
  }
  return null;
}

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
  apiBase,
  applyRowMutation,
  currentUserId,
  incidentId,
  loadRows,
  selectedRow,
  selectedRowWorkflowKey,
  setInspectorMessage,
  targetContracts,
}: {
  readonly apiBase?: string | undefined;
  readonly applyRowMutation: (
    rowKey: string,
    envelope: TimelineMutationEnvelope,
  ) => unknown;
  readonly currentUserId: string | null;
  readonly incidentId: string;
  readonly loadRows: (options: {
    readonly showLoading: boolean;
  }) => Promise<void>;
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
    const payload = buildGenericCreatePayload(
      activeWorkflow.targetContract,
      activeWorkflow.draft,
      `timeline-create-related-${activeWorkflow.featureGroup.featureGroupKey}-${Date.now()}`,
    );
    if (payload === null) {
      setWorkflow({
        ...activeWorkflow,
        message: genericCreateMinimumMessage(
          activeWorkflow.targetContract.viewSchemaId,
        ),
      });
      return;
    }
    setWorkflow({
      ...activeWorkflow,
      isSubmitting: true,
      message: null,
    });
    const createResult = await fetchWorkbookJSON<TimelineMutationEnvelope>(
      apiPath(
        apiBase,
        `/api/v1/incidents/${incidentId}/views/${activeWorkflow.targetContract.viewSchemaId}/rows`,
      ),
      { method: "POST", body: JSON.stringify(payload) },
    );
    if (!createResult.ok) {
      setWorkflow({
        ...activeWorkflow,
        isSubmitting: false,
        message: parseErrorMessage(createResult.payload),
      });
      return;
    }
    const createEnvelope = readEnvelope<TimelineMutationEnvelope>(
      createResult.payload,
    );
    const createdRecordId = recordIdFromMutationEnvelope(createEnvelope);
    if (createdRecordId === null) {
      setWorkflow({
        ...activeWorkflow,
        isSubmitting: false,
        message: "Created row response did not include a record id.",
      });
      return;
    }

    if (activeWorkflow.targetContract.viewSchemaId === evidenceViewSchemaId) {
      const patchPayload = buildAttachedEvidencePatchPayload(
        sourceRow,
        createdRecordId,
        `timeline-link-created-evidence-${Date.now()}`,
      );
      if (patchPayload === null) {
        setWorkflow({
          ...activeWorkflow,
          isSubmitting: false,
          message: "Created evidence, but the selected row version is stale.",
        });
        return;
      }
      const patchResult = await fetchWorkbookJSON<TimelineMutationEnvelope>(
        apiPath(apiBase, `/api/v1/records/${sourceRow.recordId}`),
        { method: "PATCH", body: JSON.stringify(patchPayload) },
      );
      if (!patchResult.ok) {
        setWorkflow({
          ...activeWorkflow,
          isSubmitting: false,
          message: `Created evidence, but Timeline link failed: ${parseErrorMessage(patchResult.payload)}`,
        });
        return;
      }
      const patchEnvelope = readEnvelope<TimelineMutationEnvelope>(
        patchResult.payload,
      );
      applyRowMutation(sourceRow.key, patchEnvelope);
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
    apiBase,
    applyRowMutation,
    incidentId,
    loadRows,
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
