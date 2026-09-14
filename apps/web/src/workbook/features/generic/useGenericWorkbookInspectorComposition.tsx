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
import type { SheetRef } from "../../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { useWorkbookHistorySurfaceRefresh } from "../../history/WorkbookHistoryContext";
import type { GenericSurfaceMutationController } from "../../hooks/useGenericSurfaceMutationController";
import { inspectorRecordHistoryActions } from "../../inspector/inspectorCapabilityResolver";
import { prepareWorkbookInspectorChange } from "../../inspector/prepareWorkbookInspectorChange";
import { useInspectorCreateRelatedWorkflow } from "../../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import { useWorkbookInspectorEditDraft } from "../../inspector/useWorkbookInspectorEditDraft";
import type { WorkbookInspectorFeedback } from "../../inspector/workbookInspectorErrorModel";
import { workbookInspectorLocalErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "../../inspector/workbookInspectorSubject";
import {
  type GenericCollectionMode,
  genericCollectionItems,
  genericCollectionSupportsRemove,
  genericInspectorRowLabel,
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
  taskGuardFields,
  taskViewId,
} from "../coordination/taskLifecycleModel";
import { useEvidenceWorkbookBindings } from "../evidence/useEvidenceWorkbookBindings";
import { IndicatorLifecycleContext } from "../indicators/IndicatorLifecycleContext";
import type { IndicatorInspectorHandler } from "../indicators/indicatorInspectorHandlers";
import { indicatorLifecycleViewId } from "../indicators/indicatorLifecycleModel";
import {
  NoteCreateContext,
  noteSheetAttachment,
} from "../notes/NoteCreateContext";
import { useGenericPartyLinkWorkflow } from "../parties/useGenericPartyLinkWorkflow";
import { GenericWorkbookInspectorPresentation } from "./GenericWorkbookInspectorPresentation";

const noDecisionSnapshot = () => null;
const noDecisionSubscription = () => () => {};

export function useGenericWorkbookInspectorComposition({
  canCreateRows,
  sheetRef,
  contract,
  createDraft,
  draftDisabled,
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
  readonly sheetRef: SheetRef;
  readonly canCreateRows: boolean;
  readonly contract: ViewContract;
  readonly createDraft: Record<string, string>;
  readonly draftDisabled: boolean;
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
    () =>
      contract.fields.filter(
        (field) => field.patchWritable && field.writeKind !== "read_only",
      ),
    [contract],
  );
  const [deletedHistorySubject, setDeletedHistorySubject] =
    useState<WorkbookInspectorSubject | null>(null);
  const [relatedFeedback, setRelatedFeedback] =
    useState<WorkbookInspectorFeedback | null>(null);
  const [conflictFocus, setConflictFocus] =
    useState<WorkbookConflictEntry | null>(null);
  const [editFieldKey, setEditFieldKey] = useState(
    () => editableFields[0]?.fieldKey ?? "",
  );
  const note = useContext(NoteCreateContext);
  const [indicatorInspectorHandler, setIndicatorInspectorHandler] =
    useState<IndicatorInspectorHandler | null>(null);
  const [editCollectionMode, setEditCollectionMode] =
    useState<GenericCollectionMode>("add");
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
  const selectedEdit = {
    row: subjectRow,
    field:
      editableFields.find((field) => field.fieldKey === editFieldKey) ?? null,
  };
  const edit = useWorkbookInspectorEditDraft({
    store: mutation.inspectorDrafts,
    row: subjectRow,
    field: selectedEdit.field,
    viewSchemaId: contract.viewSchemaId,
    action:
      selectedEdit.field?.writeKind === "action_payload"
        ? editCollectionMode
        : "value",
    presentation: inspectorResetKey,
    active: isOpen && interactionMode.kind === "editable",
    dependencies:
      contract.viewSchemaId === taskViewId &&
      taskGuardFields.some((field) => field === editFieldKey)
        ? taskGuardFields
        : [],
  });
  const staleEditFields = edit.staleFields;
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

  const submitCreate = async () => {
    if (!canCreateRows) return;
    if (ownerBindings.includes("linked_note_create")) {
      if (!note) return;
      if (!note.owner.getSnapshot().draft)
        note.owner.beginSheet(note.sheetRef, noteSheetAttachment);
      note.owner.resume(noteSheetAttachment);
      await note.owner.submit(noteSheetAttachment);
      return;
    }
    await mutation.ordinaryCreate.submit(contract.viewSchemaId);
  };

  const submitEdit = async () => {
    if (
      !edit.canSubmit ||
      selectedEdit.row === null ||
      selectedEdit.field === null
    ) {
      mutation.setValidationError("invalid_mutation_payload");
      return;
    }
    if (staleEditFields.length) {
      mutation.setValidationError(
        "Review changed saved fields before submitting this retained draft.",
      );
      return;
    }
    const prepared = prepareWorkbookInspectorChange(
      selectedEdit.field,
      edit.value,
      editCollectionMode,
      contract.viewSchemaId,
    );
    if (prepared.error) {
      mutation.setValidationError(prepared.error);
      return;
    }
    if (!prepared.change) return;
    const change = prepared.change;
    const captured = edit.capture();
    const finish = mutation.beginMutation();
    try {
      const payload = await mutation.submitPatchMutation({
        baseline: edit.baseline ?? selectedEdit.row,
        ...(captured.draft
          ? { authoringRevision: captured.draft.revision }
          : {}),
        presentationIdentity: captured.attachment,
        isCurrent: () => edit.isCurrent(captured),
        contributions: [
          {
            prepare: async () => {
              if (!edit.isCurrent(captured))
                throw new Error(
                  "The inspector changed before dispatch. Resume your original draft to submit it.",
                );
            },
            acknowledged: () => edit.complete(captured),
            conflictResolved: () => edit.complete(captured),
          },
        ],
        baseRowVersion: selectedEdit.row.row_version,
        changes: [change],
        purpose: "generic-patch",
        recordId: selectedEdit.row.record_id,
        viewSchemaId: contract.viewSchemaId,
      });
      if (payload === null) return;
    } finally {
      finish();
    }
  };
  const party = useGenericPartyLinkWorkflow({
    owner: mutation.partyLinks,
    contract,
    sheetRef,
    row: selectedEdit.row,
    sourceLabel: subject?.label ?? contract.title,
    resetKey: invalidationKey,
    fieldKey: selectedEdit.field?.fieldKey ?? "",
    visible: isOpen,
  });

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
        draftDisabled,
        draftInspectorFields,
        invalidationKey,
        mutation,
        mutationCommands,
        ownerBindings,
        referenceOptions,
        rows,
        setCreateDraft,
        submitCreate,
        subjectPresent: subject !== null,
      }}
      details={{
        edit,
        collectionItems: selectedEditCollectionItems,
        collectionMode: editCollectionMode,
        contract,
        editableFields,
        editFieldKey,
        mutationPending:
          mutation.mutationPending ||
          (!!subjectRow &&
            mutation.explicitPatches.blocksRecord(subjectRow.record_id)),
        referenceOptions,
        rows,
        selectedEdit,
        selectedRecordId,
        setCollectionMode: (mode) => {
          mutation.clearMutationError();
          setEditCollectionMode(mode);
        },
        setEditFieldKey: (key) => {
          mutation.clearMutationError();
          setEditFieldKey(key);
        },
        submitEdit,
      }}
      relationships={{ party }}
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
