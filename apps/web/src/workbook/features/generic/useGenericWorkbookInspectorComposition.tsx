import type { GridDensity, GridInteractionMode } from "@cartulary/grid-adapter";
import { genericWorkbookTestId } from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useContext,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import type { WorkbookProtocolPatchRecordRequest } from "../../adapters/workbookProtocolTypes";
import { useWorkbookHistorySurfaceRefresh } from "../../history/WorkbookHistoryContext";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import { inspectorRecordHistoryActions } from "../../inspector/inspectorCapabilityResolver";
import { useInspectorCreateRelatedWorkflow } from "../../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import { workbookInspectorLocalErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "../../inspector/workbookInspectorSubject";
import {
  buildGenericPatchChange,
  type GenericCollectionMode,
  genericCollectionItems,
  genericCollectionSupportsRemove,
  genericCreateMinimumMessage,
  genericInspectorRowLabel,
  initialGenericCreateDraft,
  partyLinkPairsForContract,
  selectWorkbookEditTarget,
} from "../../models/genericWorkbookModel";
import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import type { GenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { WorkbookMutationCommandPorts } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOwnerBinding } from "../../policies/workbookSurfacePolicy";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import type { WorkbookConflictEntry } from "../../runtime/workbookConflictModel";
import { DecisionSupersessionContext } from "../coordination/DecisionSupersessionContext";
import { DecisionSupersessionEditor } from "../coordination/DecisionSupersessionEditor";
import {
  decisionIneligibility,
  decisionViewId,
  reviewedDecision,
} from "../coordination/decisionSupersessionModel";
import {
  taskFieldEqual,
  taskGuardFields,
  taskValue,
  taskViewId,
} from "../coordination/taskLifecycleModel";
import { useEvidenceWorkbookBindings } from "../evidence/useEvidenceWorkbookBindings";
import { IndicatorLifecycleContext } from "../indicators/IndicatorLifecycleContext";
import type { IndicatorInspectorHandler } from "../indicators/indicatorInspectorHandlers";
import { indicatorLifecycleViewId } from "../indicators/indicatorLifecycleModel";
import { useGenericPartyLinkWorkflow } from "../parties/useGenericPartyLinkWorkflow";
import { GenericWorkbookInspectorPresentation } from "./GenericWorkbookInspectorPresentation";

const noDecisionSnapshot = () => null;
const noDecisionSubscription = () => () => {};

type RecordPatchChange = WorkbookProtocolPatchRecordRequest["changes"][number];

export function useGenericWorkbookInspectorComposition({
  canCreateRows,
  contract,
  createDraft,
  currentIncidentRole,
  currentUserId,
  draftInspectorFields,
  density,
  incidentClosed,
  inspectorResetKey,
  interactionMode,
  mutation,
  mutationCommands,
  onClearSurfaceSelection,
  onRefresh,
  onRestoreFocus,
  onRestoreEvidenceFocus,
  onSelectRecord,
  ownerBindings,
  referenceLoadError,
  referenceOptions,
  refreshReferenceOptions,
  rows,
  selectedRecordId,
  setCreateDraft,
}: {
  readonly canCreateRows: boolean;
  readonly contract: ViewContract;
  readonly createDraft: Record<string, string>;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly density: GridDensity;
  readonly onRestoreEvidenceFocus: (recordId: string) => void;
  readonly draftInspectorFields: readonly ViewFieldContract[];
  readonly incidentClosed: boolean;
  readonly inspectorResetKey: string;
  readonly interactionMode: GridInteractionMode;
  readonly mutation: GenericSurfaceMutationController;
  readonly mutationCommands: WorkbookMutationCommandPorts;
  readonly onClearSurfaceSelection: () => void;
  readonly onRefresh: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void> | void;
  readonly onRestoreFocus: () => void;
  readonly onSelectRecord: (recordId: string) => void;
  readonly ownerBindings: readonly WorkbookOwnerBinding[];
  readonly referenceLoadError: string | null;
  readonly referenceOptions: GenericReferenceOptions;
  readonly refreshReferenceOptions: () => Promise<void> | void;
  readonly rows: readonly WorkbookQueryRow[];
  readonly selectedRecordId: string;
  readonly setCreateDraft: Dispatch<SetStateAction<Record<string, string>>>;
}) {
  const inspectorConfig = contract.inspectorConfig;
  const decisionOwner = useContext(DecisionSupersessionContext);
  const lifecycleOwner = useContext(IndicatorLifecycleContext);
  const lifecycleSnapshot = useSyncExternalStore(
    lifecycleOwner?.subscribe ?? noDecisionSubscription,
    lifecycleOwner?.getSnapshot ?? noDecisionSnapshot,
  );
  const decisionSnapshot = useSyncExternalStore(
    decisionOwner?.subscribe ?? noDecisionSubscription,
    decisionOwner?.getSnapshot ?? noDecisionSnapshot,
  );
  const [decisionOpenKey, setDecisionOpenKey] = useState<string | null>(null);
  const decisionTrigger = useRef<HTMLElement | null>(null);
  const restoreDecisionTrigger = useRef(false);
  useLayoutEffect(() => {
    if (decisionOpenKey === null && restoreDecisionTrigger.current) {
      restoreDecisionTrigger.current = false;
      if (decisionTrigger.current?.isConnected)
        decisionTrigger.current.focus({ preventScroll: true });
    }
  }, [decisionOpenKey]);
  useWorkbookHistorySurfaceRefresh(inspectorConfig.viewSchemaId, () =>
    onRefresh({ requireAcceptance: true }),
  );
  const editableFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind !== "read_only"),
    [contract],
  );
  const [deletedHistorySubject, setDeletedHistorySubject] =
    useState<WorkbookInspectorSubject | null>(null);
  const [relatedFeedback, setRelatedFeedback] =
    useState<WorkbookInspectorFeedback | null>(null);
  const [conflictFocus, setConflictFocus] =
    useState<WorkbookConflictEntry | null>(null);
  const [editFieldKey, setEditFieldKey] = useState("");
  const [otherEditValue, setOtherEditValue] = useState("");
  const inspectorDrafts = mutation.explicitPatches.inspectorDrafts;
  useSyncExternalStore(inspectorDrafts.subscribe, inspectorDrafts.getSnapshot);
  const [linkedNoteSourceRecordId, setLinkedNoteSourceRecordId] = useState("");
  const [indicatorInspectorHandler, setIndicatorInspectorHandler] =
    useState<IndicatorInspectorHandler | null>(null);
  const [editCollectionMode, setEditCollectionMode] =
    useState<GenericCollectionMode>("add");
  const partyLinkExistingPartyIdForReset = useRef<(value: string) => void>(
    () => undefined,
  );
  const subjectRow =
    rows.find((row) => row.record_id === selectedRecordId) ?? null;
  useLayoutEffect(() => {
    // Query rows can arrive before the shell's authority effect. Admit their
    // Records version only once this same-account presentation is authorized.
    if (
      contract.viewSchemaId === indicatorLifecycleViewId &&
      subjectRow &&
      lifecycleSnapshot?.authority?.actorId === currentUserId &&
      lifecycleSnapshot.authority.role === currentIncidentRole
    )
      lifecycleOwner?.acceptRow(subjectRow);
  }, [
    contract.viewSchemaId,
    subjectRow,
    lifecycleOwner,
    lifecycleSnapshot?.authority,
    currentUserId,
    currentIncidentRole,
  ]);
  const subject: WorkbookInspectorSubject | null =
    subjectRow === null
      ? deletedHistorySubject
      : buildWorkbookInspectorSubject({
          config: inspectorConfig,
          kind: "live",
          label: genericInspectorRowLabel(contract, subjectRow),
          recordId: subjectRow.record_id,
          rowVersion: subjectRow.row_version,
          surfaceLabel: contract.title,
        });
  const resetEvidence = useRef<() => void>(() => undefined);
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ cause, scope }) => {
        if (cause !== "retarget") resetEvidence.current();
        setOtherEditValue("");
        setLinkedNoteSourceRecordId("");
        setEditCollectionMode("add");
        partyLinkExistingPartyIdForReset.current("");
        mutation.clearMutationError();
        setRelatedFeedback(null);
        if (cause === "close" || scope === "surface") {
          setDeletedHistorySubject(null);
        }
        if (scope === "surface") onClearSurfaceSelection();
      },
      restoreFocus: onRestoreFocus,
    },
    config: inspectorConfig,
    lifecycleKey: inspectorResetKey,
    subject,
  });
  const isOpen = workbookInspectorStateIsOpen(inspector.snapshot);
  const invalidationKey = `${contract.viewSchemaId}:${inspector.snapshot.invalidationGeneration}`;
  useLayoutEffect(() => {
    if (
      !isOpen ||
      !conflictFocus ||
      subjectRow?.record_id !== conflictFocus.conflict.record_id
    )
      return;
    const id = `${conflictFocus.compoundOperationId ? "task-lifecycle" : "generic-edit"}-${subjectRow.record_id}-${conflictFocus.conflict.field_key}`;
    const control = document.getElementById(id);
    if (control) {
      control.focus({ preventScroll: true });
      setConflictFocus(null);
    }
  }, [conflictFocus, isOpen, subjectRow]);
  const recordHistoryActions = useMemo(
    () => inspectorRecordHistoryActions(inspectorConfig),
    [inspectorConfig],
  );
  const ownerRecordActions = useEvidenceWorkbookBindings({
    mutationCommands: mutationCommands.evidence,
    mutation: {
      beginMutation: mutation.beginMutationReport,
    },
    onRefresh,
    ownerBindings,
    resetKey: inspectorResetKey,
    rows,
    subjectRecordId: subjectRow?.record_id ?? null,
    canRead: currentIncidentRole !== null,
    attachDisabledReason: incidentClosed
      ? "This incident is closed. Evidence attachment is read-only."
      : currentIncidentRole === null || currentIncidentRole === "viewer"
        ? "Your role can read evidence but cannot attach files."
        : interactionMode.kind === "read_only"
          ? interactionMode.label
          : null,
    density,
    onInspect: (recordId) => {
      onSelectRecord(recordId);
      inspector.commands.open();
    },
    onRestoreFocus: onRestoreEvidenceFocus,
  });
  resetEvidence.current = ownerRecordActions.resetLocalState;
  const selectedEdit = selectWorkbookEditTarget({
    fieldKey: editFieldKey,
    fields: editableFields,
    getRecordId: (row: WorkbookQueryRow) => row.record_id,
    recordId: selectedRecordId,
    rows,
  });
  const taskEditRow =
    contract.viewSchemaId === taskViewId ? selectedEdit.row : null;
  const taskEditDraft = taskEditRow ? inspectorDrafts.read(taskEditRow) : null;
  const selectedField = selectedEdit.field?.fieldKey ?? "";
  const editValue = taskEditRow
    ? (taskEditDraft?.values[selectedField] ??
      (selectedEdit.field?.writeKind === "action_payload"
        ? ""
        : taskValue(taskEditRow, selectedField)))
    : otherEditValue;
  const setEditValue = (value: string) => {
    if (taskEditRow) inspectorDrafts.update(taskEditRow, selectedField, value);
    else setOtherEditValue(value);
  };
  const staleEditFields =
    taskEditRow &&
    taskEditDraft &&
    Object.hasOwn(taskEditDraft.values, selectedField)
      ? [
          ...new Set([
            selectedField,
            ...(taskGuardFields.some((field) => field === selectedField)
              ? taskGuardFields
              : []),
          ]),
        ].filter(
          (field) =>
            !taskFieldEqual(taskEditDraft.baseline, taskEditRow, field),
        )
      : [];
  const selectedEditCollectionItems =
    selectedEdit.row !== null && selectedEdit.field !== null
      ? genericCollectionItems(selectedEdit.row, selectedEdit.field.fieldKey)
      : [];
  const createRelatedWorkflow = useInspectorCreateRelatedWorkflow({
    beginMutation: mutation.beginMutationReport,
    currentUserId,
    mutationCommands: mutationCommands.timeline.related,
    onCreated: refreshReferenceOptions,
    onFeedback: setRelatedFeedback,
    selectedSubject:
      subjectRow === null || subject?.kind !== "live"
        ? null
        : { cells: subjectRow.cells, subject },
  });
  const disabledTokens = useMemo(() => {
    const tokens = new Set<InspectorDisabledCondition>();
    if (selectedEdit.row === null) tokens.add("no_row_selected");
    else tokens.add("record_not_deleted");
    tokens.add("rollback_target_unavailable");
    tokens.add("pivot_target_unavailable");
    if (incidentClosed) tokens.add("incident_closed");
    return tokens;
  }, [incidentClosed, selectedEdit.row]);
  const partyLinkPairs = useMemo(
    () => partyLinkPairsForContract(contract),
    [contract],
  );

  useEffect(() => {
    if (selectedEdit.field?.writeKind !== "action_payload") {
      setEditCollectionMode("add");
    } else if (
      !genericCollectionSupportsRemove(selectedEdit.field.fieldKey) &&
      editCollectionMode === "remove"
    ) {
      setEditCollectionMode("add");
    }
  }, [editCollectionMode, selectedEdit.field]);
  useEffect(() => {
    if (selectedEdit.row === null || selectedEdit.field === null) {
      setOtherEditValue("");
      return;
    }
    if (selectedEdit.field.writeKind === "action_payload") {
      setOtherEditValue("");
      return;
    }
    const value = selectedEdit.row.cells[selectedEdit.field.fieldKey]?.value;
    setOtherEditValue(
      value === null || value === undefined ? "" : String(value),
    );
  }, [selectedEdit.field, selectedEdit.row]);

  const submitCreate = async () => {
    if (!canCreateRows) return;
    if (
      !mutationCommands.generic.canCreateRecord({
        contract,
        draft: createDraft,
      })
    ) {
      mutation.setValidationError(genericCreateMinimumMessage(contract));
      return;
    }
    const finish = mutation.beginMutation();
    try {
      const result = await mutationCommands.generic.createRecord({
        contract,
        draft: createDraft,
        linkedNoteSourceRecordId:
          ownerBindings.includes("linked_note_create") &&
          linkedNoteSourceRecordId !== ""
            ? linkedNoteSourceRecordId
            : "",
      });
      if (result.kind === "rejected") {
        mutation.rejectMutationFailure(result.failure);
        return;
      }
      setCreateDraft(initialGenericCreateDraft(contract, currentUserId));
      setLinkedNoteSourceRecordId("");
      await mutation.completeGenericMutation();
    } finally {
      finish();
    }
  };
  const submitEdit = async () => {
    if (selectedEdit.row === null || selectedEdit.field === null) {
      mutation.setValidationError("invalid_mutation_payload");
      return;
    }
    if (staleEditFields.length) {
      mutation.setValidationError(
        "Review changed saved fields before submitting this retained draft.",
      );
      return;
    }
    const change = buildGenericPatchChange(
      selectedEdit.field,
      editValue,
      editCollectionMode,
      contract.viewSchemaId,
    );
    if (change === null) {
      mutation.setValidationError(
        "Provide a value, or leave clearable fields empty to clear them.",
      );
      return;
    }
    const finish = mutation.beginMutation();
    try {
      const payload = await mutation.submitPatchMutation({
        baseline: taskEditDraft?.baseline ?? selectedEdit.row,
        baseRowVersion: selectedEdit.row.row_version,
        changes: [change],
        purpose: "generic-patch",
        recordId: selectedEdit.row.record_id,
        viewSchemaId: contract.viewSchemaId,
      });
      if (payload === null) return;
      if (taskEditRow)
        inspectorDrafts.review(payload.row, selectedField, false);
      else setOtherEditValue("");
      if (contract.viewSchemaId !== taskViewId)
        await mutation.completeGenericMutation();
    } finally {
      finish();
    }
  };
  const submitPartyLinkPatch = async (
    changes: RecordPatchChange[],
    purpose: string,
  ) => {
    if (selectedEdit.row === null) {
      mutation.setValidationError("Select a row before changing a party link.");
      return false;
    }
    const finish = mutation.beginMutation();
    try {
      const payload = await mutation.submitPatchMutation({
        baseline: selectedEdit.row,
        baseRowVersion: selectedEdit.row.row_version,
        changes,
        purpose,
        recordId: selectedEdit.row.record_id,
        viewSchemaId: contract.viewSchemaId,
      });
      if (payload === null) return false;
      if (contract.viewSchemaId !== taskViewId)
        await mutation.completeGenericMutation();
      return true;
    } finally {
      finish();
    }
  };
  const party = useGenericPartyLinkWorkflow({
    mutation: {
      beginMutation: mutation.beginMutationReport,
      rejectMutationFailure: mutation.rejectMutationFailure,
      setValidationError: mutation.setValidationError,
    },
    mutationCommands: mutationCommands.generic,
    originViewSchemaId: contract.viewSchemaId,
    partyLinkPairs,
    resetKey: invalidationKey,
    selectedRow: selectedEdit.row,
    selectedSubject: subject?.kind === "live" ? subject : null,
    submitLinkPatch: submitPartyLinkPatch,
  });
  partyLinkExistingPartyIdForReset.current = party.setPartyLinkExistingPartyId;

  const close = () => inspector.commands.close({ restoreFocus: true });
  const node = isOpen ? (
    <GenericWorkbookInspectorPresentation
      isOpen={isOpen}
      inspector={{
        config: inspectorConfig,
        currentIncidentRole,
        disabledTokens,
        decisionSupersession:
          contract.viewSchemaId === decisionViewId
            ? {
                start: () => {
                  decisionTrigger.current =
                    document.activeElement instanceof HTMLElement
                      ? document.activeElement
                      : null;
                  setDecisionOpenKey(invalidationKey);
                },
                disabledReason:
                  subject?.kind !== "live" || subjectRow === null
                    ? "Select a saved Decision."
                    : decisionOwner === null
                      ? "Decision supersession is unavailable."
                      : (decisionIneligibility(
                          reviewedDecision(
                            decisionOwner.latestRow(subjectRow.record_id) ??
                              subjectRow,
                            decisionSnapshot?.authority?.incidentId ?? "",
                          ),
                          "target",
                        ) ??
                        (decisionOwner.blocksRecord(subjectRow.record_id)
                          ? "This Decision has a pending supersession. Use Decision actions to recover it."
                          : null)),
                content:
                  decisionOpenKey === invalidationKey &&
                  decisionOwner &&
                  subjectRow &&
                  subject?.kind === "live" ? (
                    <DecisionSupersessionEditor
                      key={invalidationKey}
                      owner={decisionOwner}
                      row={subjectRow}
                      lifecycleKey={invalidationKey}
                      originSurface={inspectorResetKey}
                      onCancel={() => {
                        restoreDecisionTrigger.current = true;
                        setDecisionOpenKey(null);
                      }}
                      reconcile={async () => {}}
                    />
                  ) : null,
              }
            : undefined,
        evidenceContent:
          subjectRow === null
            ? null
            : ownerRecordActions.renderInspector(subjectRow),
        history: {
          beginMutation: mutation.beginMutationReport,
          actions: recordHistoryActions,
          canMutate:
            interactionMode.kind === "editable" &&
            currentIncidentRole !== null &&
            currentIncidentRole !== "viewer",
          commands: mutationCommands.records,
          effects: {
            deleteAccepted: (accepted) => {
              setIndicatorInspectorHandler(null);
              createRelatedWorkflow.commands.cancel();
              onSelectRecord("");
              setDeletedHistorySubject(
                buildWorkbookInspectorSubject({
                  config: inspectorConfig,
                  kind: "deleted",
                  label: "Deleted record",
                  recordId: accepted.recordId,
                  rowVersion: accepted.rowVersion,
                  stateLabel: "Deleted",
                  surfaceLabel: contract.title,
                }),
              );
            },
            restoreAccepted: (accepted) => {
              setDeletedHistorySubject(null);
              onSelectRecord(accepted.recordId);
            },
            rollbackAccepted: () => {},
            refresh: () => onRefresh({ requireAcceptance: true }),
          },
        },
        indicator:
          selectedEdit.row === null
            ? null
            : {
                handler: indicatorInspectorHandler,
                onMutationCommitted: onRefresh,
                recordId: selectedEdit.row.record_id,
                rowVersion: selectedEdit.row.row_version,
                select: setIndicatorInspectorHandler,
              },
        mutationError: mutation.mutationError,
        onClose: close,
        referenceLoadError:
          referenceLoadError === null
            ? null
            : workbookInspectorLocalErrorPresentation(referenceLoadError),
        referenceLoadErrorTestId: genericWorkbookTestId("reference-load-error"),
        related: {
          begin: createRelatedWorkflow.commands.begin,
          cancel: createRelatedWorkflow.commands.cancel,
          referenceOptions,
          state: createRelatedWorkflow.snapshot.workflow,
          submit: createRelatedWorkflow.commands.submit,
          updateDraft: createRelatedWorkflow.commands.updateDraft,
        },
        relatedFeedback,
        subject,
        surfaceTitle: contract.title,
      }}
      workflow={{
        subjectRow,
        currentIncidentRole,
        disabledTokens,
        lifecycleDisabled:
          interactionMode.kind !== "editable" ||
          (!!subjectRow &&
            mutation.explicitPatches.blocksRecord(subjectRow.record_id)),
        canCreateRows,
        contract,
        createDraft,
        draftInspectorFields,
        invalidationKey,
        linkedNoteSourceRecordId,
        mutation,
        mutationCommands,
        ownerBindings,
        referenceOptions,
        rows,
        setCreateDraft,
        setLinkedNoteSourceRecordId,
        submitCreate,
        subjectPresent: subject !== null,
      }}
      details={{
        staleEditFields: staleEditFields.map((field) => ({
          field,
          label: contract.fieldMap[field]?.label ?? field,
          saved: taskEditRow ? taskValue(taskEditRow, field) : "",
        })),
        reviewEditField: (field, keepDraft) => {
          if (taskEditRow)
            inspectorDrafts.review(taskEditRow, field, keepDraft);
        },
        collectionItems: selectedEditCollectionItems,
        collectionMode: editCollectionMode,
        contract,
        editableFields,
        editFieldKey,
        editValue,
        mutationPending:
          mutation.mutationPending ||
          (!!taskEditRow &&
            mutation.explicitPatches.blocksRecord(taskEditRow.record_id)),
        onSelectRecord,
        referenceOptions,
        rows,
        selectedEdit,
        selectedRecordId,
        setCollectionMode: setEditCollectionMode,
        setEditFieldKey,
        setEditValue,
        submitEdit,
      }}
      relationships={{
        disabled: mutation.mutationPending,
        party,
        partyLinkPairs,
        referenceOptions,
        rowSelected: selectedEdit.row !== null,
      }}
    />
  ) : undefined;
  return {
    restoreConflictFocus: (conflict: WorkbookConflictEntry) => {
      onSelectRecord(conflict.conflict.record_id);
      if (!conflict.compoundOperationId)
        setEditFieldKey(conflict.conflict.field_key);
      setConflictFocus(conflict);
      inspector.commands.open();
    },
    close,
    invalidationKey,
    isOpen,
    node,
    open: inspector.commands.open,
    ownerRecordActions,
    submitCreate,
  };
}
