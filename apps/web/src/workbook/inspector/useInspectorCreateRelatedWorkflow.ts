import {
  getViewContract,
  type InspectorFeatureGroup,
  type ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useState } from "react";
import {
  genericCreateMinimumMessage,
  initialGenericCreateDraft,
  workbookCreationAvailable,
} from "../models/genericWorkbookModel";
import type { TimelineRelatedRecordPort } from "../mutations/workbookMutationCommandPorts";
import { stringifyGridValue } from "../utils/workbookValueFormat";

export type InspectorCreateRelatedSubject = {
  readonly cells: Readonly<Record<string, { readonly value: unknown }>>;
  readonly recordId: string;
  readonly rowVersion: number;
};

export type InspectorCreateRelatedWorkflowState = {
  readonly draft: Record<string, string>;
  readonly featureGroup: InspectorFeatureGroup;
  readonly isSubmitting: boolean;
  readonly message: string | null;
  readonly sourceRecordId: string;
  readonly targetContract: ViewContract;
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
      if (
        targetContract === undefined ||
        !workbookCreationAvailable(targetContract)
      ) {
        onMessage("The target view does not allow row creation.");
        return true;
      }
      if (selectedSubject === null) {
        onMessage("Select a saved row before creating a related record.");
        return true;
      }
      const draft = applySeedBindings(
        initialGenericCreateDraft(targetContract, currentUserId),
        featureGroup,
        selectedSubject,
      );
      setWorkflow({
        draft,
        featureGroup,
        isSubmitting: false,
        message: null,
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
            message: null,
          },
    );
  }, []);

  const cancel = useCallback(() => setWorkflow(null), []);

  const submit = useCallback(async () => {
    const active = workflow;
    if (active === null) return;
    setWorkflow({ ...active, isSubmitting: true, message: null });
    const outcome = await mutationCommands.createRelatedRecord({
      contract: active.targetContract,
      draft: active.draft,
      featureGroupKey: active.featureGroup.featureGroupKey,
    });
    if (outcome.kind === "rejected") {
      setWorkflow({
        ...active,
        isSubmitting: false,
        message:
          outcome.failure.kind === "validation"
            ? genericCreateMinimumMessage(active.targetContract)
            : outcome.failure.message,
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

function applySeedBindings(
  draft: Record<string, string>,
  featureGroup: InspectorFeatureGroup,
  subject: InspectorCreateRelatedSubject,
): Record<string, string> {
  const next = { ...draft };
  for (const binding of featureGroup.seedBindings) {
    let value: string | null = null;
    if (binding.source.kind === "selected_record_id") {
      value = subject.recordId;
    } else if (
      binding.source.kind === "selected_field_value" &&
      binding.source.sourceFieldKey !== undefined
    ) {
      const text = stringifyGridValue(
        subject.cells[binding.source.sourceFieldKey]?.value,
      ).trim();
      value = text === "" ? null : text;
    } else if (
      binding.source.kind === "literal" &&
      binding.source.value !== null &&
      binding.source.value !== undefined
    ) {
      value =
        typeof binding.source.value === "string"
          ? binding.source.value
          : JSON.stringify(binding.source.value);
    }
    if (value !== null) next[binding.targetFieldKey] = value;
  }
  return next;
}
