import type { GridInteractionMode } from "@cartulary/grid-adapter";
import { genericWorkbookTestId } from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
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
import { useEvidenceWorkbookBindings } from "../evidence/useEvidenceWorkbookBindings";
import type { IndicatorInspectorHandler } from "../indicators/indicatorInspectorHandlers";
import { useGenericPartyLinkWorkflow } from "../parties/useGenericPartyLinkWorkflow";
import { GenericWorkbookInspectorPresentation } from "./GenericWorkbookInspectorPresentation";

export function useGenericWorkbookInspectorComposition({
  canCreateRows,
  contract,
  createDraft,
  currentIncidentRole,
  currentUserId,
  draftInspectorFields,
  incidentClosed,
  inspectorResetKey,
  interactionMode,
  mutation,
  mutationCommands,
  onClearSurfaceSelection,
  onRefresh,
  onRestoreFocus,
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
  readonly draftInspectorFields: readonly ViewFieldContract[];
  readonly incidentClosed: boolean;
  readonly inspectorResetKey: string;
  readonly interactionMode: GridInteractionMode;
  readonly mutation: GenericSurfaceMutationController;
  readonly mutationCommands: WorkbookMutationCommandPorts;
  readonly onClearSurfaceSelection: () => void;
  readonly onRefresh: () => Promise<void> | void;
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
  const editableFields = useMemo(
    () => contract.fields.filter((field) => field.writeKind !== "read_only"),
    [contract],
  );
  const [deletedHistorySubject, setDeletedHistorySubject] =
    useState<WorkbookInspectorSubject | null>(null);
  const [relatedFeedback, setRelatedFeedback] =
    useState<WorkbookInspectorFeedback | null>(null);
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
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
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ cause, scope }) => {
        setEditValue("");
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
  const recordHistoryActions = useMemo(
    () => inspectorRecordHistoryActions(inspectorConfig),
    [inspectorConfig],
  );
  const ownerRecordActions = useEvidenceWorkbookBindings({
    mutationCommands: mutationCommands.evidence,
    mutation: {
      beginMutation: mutation.beginMutation,
      markMutationConflict: mutation.markMutationConflict,
      markMutationSaved: mutation.markMutationSaved,
    },
    onRefresh,
    ownerBindings,
    resetKey: invalidationKey,
  });
  const selectedEdit = selectWorkbookEditTarget({
    fieldKey: editFieldKey,
    fields: editableFields,
    getRecordId: (row: WorkbookQueryRow) => row.record_id,
    recordId: selectedRecordId,
    rows,
  });
  const selectedEditCollectionItems =
    selectedEdit.row !== null && selectedEdit.field !== null
      ? genericCollectionItems(selectedEdit.row, selectedEdit.field.fieldKey)
      : [];
  const createRelatedWorkflow = useInspectorCreateRelatedWorkflow({
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
      setEditValue("");
      return;
    }
    if (selectedEdit.field.writeKind === "action_payload") {
      setEditValue("");
      return;
    }
    const value = selectedEdit.row.cells[selectedEdit.field.fieldKey]?.value;
    setEditValue(value === null || value === undefined ? "" : String(value));
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
    mutation.beginMutation();
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
  };
  const submitEdit = async () => {
    if (selectedEdit.row === null || selectedEdit.field === null) {
      mutation.setValidationError("invalid_mutation_payload");
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
    const payload = await mutation.submitPatchMutation({
      baseRowVersion: selectedEdit.row.row_version,
      changes: [change],
      purpose: "generic-patch",
      recordId: selectedEdit.row.record_id,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) return;
    setEditValue("");
    await mutation.completeGenericMutation();
  };
  const submitPartyLinkPatch = async (
    changes: Array<Record<string, unknown>>,
    purpose: string,
  ) => {
    if (selectedEdit.row === null) {
      mutation.setValidationError("Select a row before changing a party link.");
      return false;
    }
    const payload = await mutation.submitPatchMutation({
      baseRowVersion: selectedEdit.row.row_version,
      changes,
      purpose,
      recordId: selectedEdit.row.record_id,
      viewSchemaId: contract.viewSchemaId,
    });
    if (payload === null) return false;
    await mutation.completeGenericMutation();
    return true;
  };
  const party = useGenericPartyLinkWorkflow({
    mutation: {
      beginMutation: mutation.beginMutation,
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
        evidenceContent:
          subjectRow === null
            ? null
            : ownerRecordActions.renderRecordActions(subjectRow),
        history: {
          actions: recordHistoryActions,
          canMutate:
            interactionMode.kind === "editable" &&
            currentIncidentRole !== null &&
            currentIncidentRole !== "viewer",
          commands: mutationCommands.records,
          effects: {
            deleteAccepted: async (accepted) => {
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
              await onRefresh();
            },
            restoreAccepted: async (accepted) => {
              await onRefresh();
              setDeletedHistorySubject(null);
              onSelectRecord(accepted.recordId);
            },
            rollbackAccepted: async () => {
              await onRefresh();
            },
          },
        },
        indicator:
          selectedEdit.row === null
            ? null
            : {
                handler: indicatorInspectorHandler,
                onMutationCommitted: onRefresh,
                port: mutationCommands.indicators,
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
        collectionItems: selectedEditCollectionItems,
        collectionMode: editCollectionMode,
        contract,
        editableFields,
        editFieldKey,
        editValue,
        mutationState: mutation.mutationState,
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
        disabled: mutation.mutationState === "Syncing",
        party,
        partyLinkPairs,
        referenceOptions,
        rowSelected: selectedEdit.row !== null,
      }}
    />
  ) : undefined;
  return {
    close,
    invalidationKey,
    isOpen,
    node,
    open: inspector.commands.open,
    ownerRecordActions,
    submitCreate,
  };
}
