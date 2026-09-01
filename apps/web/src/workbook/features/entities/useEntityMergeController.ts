import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  type WorkbookInspectorErrorPresentation,
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "../../inspector/workbookInspectorErrorModel";
import {
  buildMergePlan,
  type EntityMergePlan,
  type EntityRow,
} from "../../models/entityWorkbookModel";
import type { EntityMutationCommandPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFieldFailure } from "../../mutations/workbookOperationOutcome";

export type EntityMergePreconditionDetail = {
  readonly key: string;
  readonly label: string;
  readonly value: string;
};

export type EntityMergeController = {
  readonly commands: {
    readonly clearPlan: () => void;
    readonly confirm: () => Promise<void>;
    readonly reset: () => void;
    readonly selectCandidate: (recordId: string) => void;
    readonly setReason: (reason: string) => void;
    readonly start: () => void;
  };
  readonly snapshot: {
    readonly candidateId: string;
    readonly loser: EntityRow | null;
    readonly message: WorkbookInspectorErrorPresentation | null;
    readonly plan: EntityMergePlan | null;
    readonly preconditionDetails: readonly EntityMergePreconditionDetail[];
    readonly reason: string;
  };
};

const defaultMergeReason = "Merge duplicate entity";

const mergeDetailLabels: Readonly<Record<string, string>> = {
  reason_code: "Reason",
  record_type: "Record type",
  identifier_class: "Identifier class",
  normalized_value: "Normalized value",
  blocking_record_id: "Blocking record",
  survivor_record_id: "Survivor record",
  loser_record_id: "Loser record",
  survivor_base_row_version: "Survivor supplied version",
  loser_base_row_version: "Loser supplied version",
  survivor_current_row_version: "Survivor current version",
  loser_current_row_version: "Loser current version",
};

function preconditionDetails(
  fields: readonly WorkbookOperationFieldFailure[] | undefined,
): readonly EntityMergePreconditionDetail[] {
  if (fields === undefined) return [];
  return fields.flatMap((field) => {
    const label = mergeDetailLabels[field.field];
    return label === undefined
      ? []
      : [{ key: field.field, label, value: field.message }];
  });
}

export function useEntityMergeController({
  canMerge,
  clearDrafts,
  lifecycleResetKey,
  loadSurvivorPreview,
  mutationCommands,
  onRefreshEntities,
  retargetSurvivor,
  rows,
  selectedEntity,
}: {
  readonly canMerge: boolean;
  readonly clearDrafts: () => void;
  readonly lifecycleResetKey: string;
  readonly loadSurvivorPreview: (recordId: string) => Promise<void>;
  readonly mutationCommands: EntityMutationCommandPort;
  readonly onRefreshEntities: () => Promise<void>;
  readonly retargetSurvivor: (recordId: string) => void;
  readonly rows: readonly EntityRow[];
  readonly selectedEntity: EntityRow | null;
}): EntityMergeController {
  const [candidateId, setCandidateId] = useState("");
  const [message, setMessage] =
    useState<WorkbookInspectorErrorPresentation | null>(null);
  const [details, setDetails] = useState<
    readonly EntityMergePreconditionDetail[]
  >([]);
  const [reason, setReason] = useState(defaultMergeReason);
  const generationRef = useRef(0);
  const admittedLifecycleKeyRef = useRef<string | null>(null);
  const lifecycleKey = `${lifecycleResetKey}:${canMerge}:${selectedEntity?.recordId ?? "none"}:${selectedEntity?.rowVersion ?? 0}`;
  const previousLifecycleKeyRef = useRef(lifecycleKey);

  const resetState = useCallback(() => {
    setCandidateId("");
    setMessage(null);
    setDetails([]);
    setReason(defaultMergeReason);
  }, []);

  const reset = useCallback(() => {
    generationRef.current += 1;
    admittedLifecycleKeyRef.current = null;
    resetState();
  }, [resetState]);

  const clearPlan = useCallback(() => {
    setCandidateId("");
    setDetails([]);
  }, []);

  useEffect(() => {
    if (previousLifecycleKeyRef.current === lifecycleKey) return;
    previousLifecycleKeyRef.current = lifecycleKey;
    if (admittedLifecycleKeyRef.current === lifecycleKey) {
      admittedLifecycleKeyRef.current = null;
      return;
    }
    generationRef.current += 1;
    admittedLifecycleKeyRef.current = null;
    resetState();
  }, [lifecycleKey, resetState]);

  useEffect(
    () => () => {
      generationRef.current += 1;
    },
    [],
  );

  const loser =
    rows.find((candidate) => candidate.recordId === candidateId) ?? null;
  const plan = useMemo(
    () =>
      selectedEntity === null || loser === null
        ? null
        : buildMergePlan(selectedEntity, loser),
    [loser, selectedEntity],
  );

  const selectCandidate = useCallback((recordId: string) => {
    setCandidateId(recordId);
    setMessage(null);
    setDetails([]);
  }, []);

  const start = useCallback(() => {
    setMessage(
      workbookInspectorLocalErrorPresentation(
        "Select a loser to review the merge plan.",
      ),
    );
    setDetails([]);
  }, []);

  const confirm = useCallback(async () => {
    if (!canMerge || selectedEntity === null || loser === null) return;
    const generation = generationRef.current;
    const survivor = selectedEntity;
    setMessage(null);
    setDetails([]);
    const outcome = await mutationCommands.merge({
      loserBaseRowVersion: loser.rowVersion,
      loserRecordId: loser.recordId,
      reason,
      survivorBaseRowVersion: survivor.rowVersion,
      survivorRecordId: survivor.recordId,
    });
    if (generationRef.current !== generation) return;
    if (outcome.kind === "rejected") {
      setMessage(workbookInspectorErrorPresentation(outcome.failure));
      setDetails(
        outcome.failure.kind === "validation"
          ? preconditionDetails(outcome.failure.fields)
          : [],
      );
      return;
    }

    admittedLifecycleKeyRef.current = `${lifecycleResetKey}:${canMerge}:${survivor.recordId}:${outcome.value.survivorRowVersion}`;
    clearDrafts();
    setDetails([]);
    setMessage(
      workbookInspectorLocalErrorPresentation(
        `Merged ${loser.label} into ${survivor.label} (${outcome.value.recordType}).`,
      ),
    );
    await onRefreshEntities();
    if (generationRef.current !== generation) return;
    await loadSurvivorPreview(survivor.recordId);
    if (generationRef.current !== generation) return;
    retargetSurvivor(survivor.recordId);
    setCandidateId("");
  }, [
    canMerge,
    clearDrafts,
    lifecycleResetKey,
    loadSurvivorPreview,
    loser,
    mutationCommands,
    onRefreshEntities,
    reason,
    retargetSurvivor,
    selectedEntity,
  ]);

  return {
    commands: { clearPlan, confirm, reset, selectCandidate, setReason, start },
    snapshot: {
      candidateId,
      loser,
      message,
      plan,
      preconditionDetails: details,
      reason,
    },
  };
}
