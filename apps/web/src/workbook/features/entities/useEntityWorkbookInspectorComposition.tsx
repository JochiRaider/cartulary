import type { GridInteractionMode } from "@cartulary/grid-adapter";
import {
  entityInspectorTestId,
  entityMergeControlTestId,
  entityMergePreconditionDetailsTestId,
  entityReusableIdentifierItemTestId,
  entityReusableIdentifiersSectionTestId,
  genericEditSubmitTestId,
  genericEditValueTestId,
  timelinePreviewRowTestId,
} from "@cartulary/ui-contracts";
import type {
  InspectorDisabledCondition,
  ViewContract,
} from "@cartulary/view-contracts";
import { X } from "lucide-react";
import {
  type Dispatch,
  type RefObject,
  type SetStateAction,
  useCallback,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { type SheetRef, sheetRefKey } from "../../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import {
  WorkbookRelationshipChip,
  WorkbookRelationshipChipDetails,
} from "../../components/WorkbookRelationshipChip";
import { useWorkbookHistorySurfaceRefresh } from "../../history/WorkbookHistoryContext";
import { useEntityTimelinePreview } from "../../hooks/useEntityTimelinePreview";
import { inspectorRecordHistoryActions } from "../../inspector/inspectorCapabilityResolver";
import { prepareWorkbookInspectorChange } from "../../inspector/prepareWorkbookInspectorChange";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorPublicError,
} from "../../inspector/presentation/WorkbookInspectorFeedback";
import {
  ownedInspectorRegion,
  type PresentInspectorRegion,
  type WorkbookInspectorPanelContentModel,
  type WorkbookInspectorPanelData,
} from "../../inspector/presentation/WorkbookInspectorPanelContent";
import { useInspectorCreateRelatedWorkflow } from "../../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import {
  useWorkbookInspectorEditDraft,
  type WorkbookInspectorEditDraft,
} from "../../inspector/useWorkbookInspectorEditDraft";
import { useWorkbookInspectorFieldFeedback } from "../../inspector/useWorkbookInspectorFieldFeedback";
import { WorkbookExplicitPatchRecovery } from "../../inspector/WorkbookExplicitPatchRecovery";
import { WorkbookInspectorDetails } from "../../inspector/WorkbookInspectorDetails";
import { WorkbookInspectorDraftFeedback } from "../../inspector/WorkbookInspectorDraftFeedback";
import { WorkbookInspectorEditControl } from "../../inspector/WorkbookInspectorEditControl";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import { buildWorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import { mergeIdentifierOutcomeText } from "../../models/entityMergePlan";
import type { EntityRow } from "../../models/entityWorkbookModel";
import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import type { TimelineRelatedRecordPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";
import type { WorkbookViewQueryPort } from "../../query/WorkbookViewQueryPort";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import { timelineRelationshipChipPresentation } from "../../timeline/models/workbookMentionChips";
import { EntityWorkbookInspector } from "./EntityWorkbookInspector";
import { useEntityMergeController } from "./useEntityMergeController";

export function useEntityWorkbookInspectorComposition({
  sheetRef,
  canMerge,
  contract,
  currentIncidentRole,
  currentUserId,
  entityActionFeedback,
  entityIndex,
  entityType,
  incidentClosed,
  inspectorResetKey,
  interactionMode,
  mutationError,
  mutationRuntime,
  onClearSurfaceSelection,
  onAuthorityUncertain,
  onRefreshEntities,
  onRestoreFocus,
  relatedMutationCommands,
  rows,
  selectedEntity,
  setEntityActionFeedback,
  setMutationError,
  setSelectedRecordId,
  viewQuery,
}: {
  readonly canMerge: boolean;
  readonly contract: ViewContract;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly entityActionFeedback: WorkbookInspectorFeedback | null;
  readonly entityIndex: Record<string, EntityRow>;
  readonly entityType: EntityRow["entityType"];
  readonly incidentClosed: boolean;
  readonly inspectorResetKey: string;
  readonly interactionMode: GridInteractionMode;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly sheetRef: SheetRef;
  readonly onClearSurfaceSelection: () => void;
  readonly onAuthorityUncertain?: (() => void) | undefined;
  readonly onRefreshEntities: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void>;
  readonly onRestoreFocus: () => void;
  readonly relatedMutationCommands: TimelineRelatedRecordPort;
  readonly rows: readonly EntityRow[];
  readonly selectedEntity: EntityRow | null;
  readonly setEntityActionFeedback: Dispatch<
    SetStateAction<WorkbookInspectorFeedback | null>
  >;
  readonly setMutationError: Dispatch<
    SetStateAction<WorkbookInspectorErrorPresentation | null>
  >;
  readonly setSelectedRecordId: Dispatch<SetStateAction<string | null>>;
  readonly viewQuery: WorkbookViewQueryPort;
}) {
  const inspectorConfig = contract.inspectorConfig;
  useWorkbookHistorySurfaceRefresh(inspectorConfig.viewSchemaId, () =>
    onRefreshEntities({ requireAcceptance: true }),
  );
  const [deletedHistorySubject, setDeletedHistorySubject] =
    useState<WorkbookRecordSubject | null>(null);
  const [editFieldKey, setEditFieldKey] = useState("");
  const aliasInputRef = useRef<HTMLInputElement | null>(null);
  const subject: WorkbookRecordSubject | null =
    selectedEntity === null
      ? deletedHistorySubject
      : buildWorkbookInspectorSubject({
          config: inspectorConfig,
          kind: "live",
          label: selectedEntity.label,
          recordId: selectedEntity.recordId,
          rowVersion: selectedEntity.rowVersion,
          stateLabel: selectedEntity.state,
          surfaceLabel: contract.title,
        });
  const {
    clearTimelinePreview,
    loadTimelinePreview,
    timelinePreviewRows,
    timelinePreviewState,
    timelinePreviewNotice,
  } = useEntityTimelinePreview({
    entityType,
    viewQuery,
    onAuthorityUncertain,
    authorityIdentity: inspectorResetKey,
  });
  const beginMutation = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const merge = useEntityMergeController({
    canMerge:
      canMerge && !incidentClosed && interactionMode.kind === "editable",
    owner: mutationRuntime.entityMerge,
    originSurface: sheetRefKey(sheetRef),
    hasAffectedDraft: (ids) => mutationRuntime.inspectorDrafts.hasRecords(ids),
    discardAffectedDrafts: (ids) =>
      mutationRuntime.inspectorDrafts.discardRecords(ids),
    lifecycleResetKey: inspectorResetKey,
    loadSurvivorPreview: (recordId) =>
      loadTimelinePreview(recordId, { requireAcceptance: true }),
    rows,
    selectedEntity,
  });
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ cause, scope }) => {
        if (cause !== "record_updated") setEditFieldKey("");
        merge.commands.clearPlan();
        if (cause !== "record_updated") clearTimelinePreview();
        setEntityActionFeedback(null);
        setMutationError(null);
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
  const editableFields = useMemo(
    () =>
      contract.fields.filter(
        (field) => field.patchWritable && field.writeKind === "direct_value",
      ),
    [contract],
  );

  const selectedEdit = {
    row: selectedEntity,
    field:
      editableFields.find((field) => field.fieldKey === editFieldKey) ?? null,
  };
  const edit = useWorkbookInspectorEditDraft({
    store: mutationRuntime.inspectorDrafts,
    row: selectedEntity?.rawRow ?? null,
    field: selectedEdit.field,
    viewSchemaId: contract.viewSchemaId,
    presentation: inspectorResetKey,
    active: isOpen,
  });
  const aliasEdit = useWorkbookInspectorEditDraft({
    store: mutationRuntime.inspectorDrafts,
    row: selectedEntity?.rawRow ?? null,
    field:
      contract.fieldMap[
        entityType === "host" ? "host.aliases" : "identity.aliases"
      ] ?? null,
    viewSchemaId: contract.viewSchemaId,
    action: "add_alias",
    presentation: inspectorResetKey,
    active: isOpen && interactionMode.kind === "editable",
  });
  const aliasRemove = useWorkbookInspectorEditDraft({
    store: mutationRuntime.inspectorDrafts,
    row: selectedEntity?.rawRow ?? null,
    field: contract.fieldMap[`${entityType}.aliases`] ?? null,
    viewSchemaId: contract.viewSchemaId,
    action: "remove_alias",
    presentation: inspectorResetKey,
    active: isOpen && interactionMode.kind === "editable",
  });
  useSyncExternalStore(
    mutationRuntime.explicitPatches.subscribe,
    mutationRuntime.explicitPatches.getSnapshot,
  );
  const mutationPending =
    !!selectedEntity &&
    mutationRuntime.explicitPatches.blocksRecord(selectedEntity.recordId);
  useLayoutEffect(() => {
    if (mutationRuntime.explicitPatches.getSnapshot().authority)
      for (const row of rows)
        if (mutationRuntime.explicitPatches.latestRow(row.recordId))
          mutationRuntime.explicitPatches.observeQuery(row.rawRow);
  }, [rows, mutationRuntime]);
  const editFeedback = useWorkbookInspectorFieldFeedback(edit);
  const aliasFeedback = useWorkbookInspectorFieldFeedback(aliasEdit);
  const aliasRemoveFeedback = useWorkbookInspectorFieldFeedback(aliasRemove);
  const aliasDraft = aliasEdit.value ?? "";
  const setEditValue = edit.update;
  const setAliasDraft = aliasEdit.update;
  const recordHistoryActions = useMemo(
    () => inspectorRecordHistoryActions(inspectorConfig),
    [inspectorConfig],
  );
  const related = useInspectorCreateRelatedWorkflow({
    beginMutation,
    currentUserId,
    mutationCommands: relatedMutationCommands,
    onCreated: onRefreshEntities,
    onFeedback: setEntityActionFeedback,
    selectedSubject:
      selectedEntity === null || subject?.kind !== "live"
        ? null
        : { cells: selectedEntity.rawRow.cells, subject },
  });
  const disabledTokens = useMemo(() => {
    const tokens = new Set<InspectorDisabledCondition>();
    if (selectedEntity === null) tokens.add("no_row_selected");
    else tokens.add("record_not_deleted");
    tokens.add("rollback_target_unavailable");
    tokens.add("pivot_target_unavailable");
    if (rows.length < 2) tokens.add("merge_target_unavailable");
    if (incidentClosed) tokens.add("incident_closed");
    return tokens;
  }, [incidentClosed, rows.length, selectedEntity]);

  useEffect(() => {
    if (selectedEntity === null) clearTimelinePreview();
  }, [clearTimelinePreview, selectedEntity]);
  useEffect(() => {
    if (!isOpen || selectedEntity === null) {
      clearTimelinePreview();
      return;
    }
    void loadTimelinePreview(selectedEntity.recordId);
  }, [clearTimelinePreview, isOpen, loadTimelinePreview, selectedEntity]);

  async function submitEdit() {
    if (
      !edit.canSubmit ||
      selectedEdit.row === null ||
      selectedEdit.field === null
    ) {
      editFeedback.rejectLocal(
        "Select an editable field before submitting.",
        false,
      );
      return;
    }
    const prepared = prepareWorkbookInspectorChange(
      selectedEdit.field,
      edit.value,
      "add",
      contract.viewSchemaId,
    );
    if (prepared.error) {
      editFeedback.rejectLocal(prepared.error);
      return;
    }
    if (!prepared.change) return;
    const change = prepared.change;
    const captured = edit.capture();
    const onFailure = editFeedback.capture();
    editFeedback.clear();
    setMutationError(null);
    const result = await mutationRuntime.explicitPatches.submit(
      {
        baseline: edit.baseline ?? selectedEdit.row.rawRow,
        changes: [change],
        purpose: "entity-patch",
        viewSchemaId: contract.viewSchemaId,
        sheetRef,
        surfaceLabel: contract.title,
        ...(captured.draft
          ? { authoringRevision: captured.draft.revision }
          : {}),
        presentationIdentity: captured.attachment,
      },
      [
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
    );
    if (result?.failure && edit.isCurrent(captured)) onFailure(result.failure);
  }

  async function submitAliasActions(actions: AliasAction[]) {
    const first = actions[0];
    if (!first || !selectedEntity) return;
    const editing = first.op === "add_alias" ? aliasEdit : aliasRemove;
    const feedback =
      first.op === "add_alias" ? aliasFeedback : aliasRemoveFeedback;
    if (!editing.canSubmit) return;
    const field = contract.fieldMap[`${entityType}.aliases`];
    if (!field?.patchWritable) return;
    const prepared = prepareWorkbookInspectorChange(
      field,
      first.op === "add_alias" ? first.alias_text : first.item_ref,
      first.op === "add_alias" ? "add" : "remove",
      contract.viewSchemaId,
    );
    if (prepared.error) {
      feedback.rejectLocal(prepared.error);
      return;
    }
    if (!prepared.change) return;
    const captured = editing.capture();
    const onFailure = feedback.capture();
    feedback.clear();
    const focusedControl = document.activeElement;
    const complete = () => {
      editing.complete(captured);
      if (
        editing.isCurrent(captured) &&
        focusedControl instanceof HTMLElement &&
        focusedControl !== document.body &&
        (document.activeElement === focusedControl ||
          (!focusedControl.isConnected &&
            document.activeElement === document.body))
      )
        aliasInputRef.current?.focus({ preventScroll: true });
    };
    setMutationError(null);
    const result = await mutationRuntime.explicitPatches.submit(
      {
        baseline: editing.baseline ?? selectedEntity.rawRow,
        changes: [prepared.change],
        purpose: `entity-alias-${first.op}`,
        viewSchemaId: contract.viewSchemaId,
        sheetRef,
        surfaceLabel: contract.title,
        ...(captured.draft
          ? { authoringRevision: captured.draft.revision }
          : {}),
        presentationIdentity: captured.attachment,
      },
      [
        {
          prepare: async () => {
            if (!editing.isCurrent(captured))
              throw new Error(
                "The inspector changed before dispatch. Review the original aliases again.",
              );
          },
          acknowledged: complete,
          conflictResolved: () => editing.complete(captured),
        },
      ],
    );
    if (result?.failure && editing.isCurrent(captured))
      onFailure(result.failure);
  }

  const close = () => inspector.commands.close({ restoreFocus: true });
  const mergeSnapshot = merge.snapshot;
  const node = isOpen ? (
    <EntityInspectorPresentation
      details={{
        patches: mutationRuntime.explicitPatches,
        disabledReason:
          interactionMode.kind !== "editable"
            ? interactionMode.label
            : !mutationRuntime.inspectorDrafts.canAuthor()
              ? "Current access permits reading only."
              : null,
        aliasDraft,
        aliasInputRef,
        contract,
        editableFields,
        editFieldKey,
        edit,
        aliasEdit,
        editFeedback,
        aliasFeedback,
        aliasRemoveFeedback,
        aliasRemove,
        mutationError,
        mutationPending,
        rows,
        selectedEdit,
        selectedEntity,
        setAliasDraft: (value) => {
          merge.commands.invalidateReview();
          setAliasDraft(value);
        },
        setEditFieldKey: (value) => {
          merge.commands.invalidateReview();
          setMutationError(null);
          setEditFieldKey(value);
        },
        setEditValue: (value) => {
          merge.commands.invalidateReview();
          setEditValue(value);
        },
        submitAliasActions,
        submitEdit,
      }}
      inspector={{
        actionFeedback: entityActionFeedback,
        config: inspectorConfig,
        currentIncidentRole,
        disabledTokens,
        feedbackTestId: entityMergeControlTestId("message"),
        history: {
          actions: recordHistoryActions,
          canMutate:
            interactionMode.kind === "editable" &&
            currentIncidentRole !== null &&
            currentIncidentRole !== "viewer",
          effects: {
            deleteAccepted: (accepted) => {
              related.commands.cancel();
              clearTimelinePreview();
              setSelectedRecordId(null);
              setDeletedHistorySubject(
                buildWorkbookInspectorSubject({
                  config: inspectorConfig,
                  kind: "deleted",
                  label: "Deleted entity",
                  recordId: accepted.recordId,
                  rowVersion: accepted.rowVersion,
                  stateLabel: "Deleted",
                  surfaceLabel: contract.title,
                }),
              );
            },
            restoreAccepted: (accepted) => {
              setDeletedHistorySubject(null);
              setSelectedRecordId(accepted.recordId);
            },
            rollbackAccepted: () => {},
            refresh: () => onRefreshEntities({ requireAcceptance: true }),
          },
        },
        mergeFeedback: mergeSnapshot.feedback,
        mergePreconditionDetails:
          mergeSnapshot.preconditionDetails.length === 0 ||
          selectedEntity === null ? null : (
            <ul
              data-testid={entityMergePreconditionDetailsTestId(
                entityType,
                selectedEntity.recordId,
              )}
              style={flatListStyle}
            >
              {mergeSnapshot.preconditionDetails.map((line) => (
                <li key={line.key}>
                  {line.label}: {line.value}
                </li>
              ))}
            </ul>
          ),
        onClose: close,
        related: {
          begin: related.commands.begin,
          cancel: related.commands.cancel,
          state: related.snapshot.workflow,
          submit: related.commands.submit,
          updateDraft: related.commands.updateDraft,
        },
        subject,
        surfaceTitle: contract.title,
        testId: entityInspectorTestId(entityType),
      }}
      isOpen={isOpen}
      relationships={{
        timelinePreviewNotice,
        refreshTimelinePreview: (recordId) => {
          void loadTimelinePreview(recordId);
        },
        canMerge: canMerge && mutationRuntime.entityMerge.canSubmit(),
        entityIndex,
        entityType,
        merge,
        rows,
        selectedEntity,
        setEntityActionFeedback,
        timelinePreviewRows:
          timelinePreviewState.recordId === selectedEntity?.recordId
            ? timelinePreviewRows
            : [],
        timelinePreviewState,
      }}
    />
  ) : undefined;
  return {
    captureFindFocus: (target: EventTarget | null) => {
      const owner = [edit, aliasEdit, aliasRemove].find(
        (draft) => draft.controlRef.current === target,
      );
      if (!owner || !(target instanceof HTMLElement)) return null;
      const captured = owner.capture();
      return {
        restore: () => {
          if (
            !owner.isCurrent(captured) ||
            !target.isConnected ||
            owner.controlRef.current !== target
          )
            return false;
          target.focus();
          return document.activeElement === target;
        },
      };
    },
    close,
    isOpen,
    node,
    open: inspector.commands.open,
    openForRecord: (recordId: string) => {
      setSelectedRecordId(recordId);
      setEntityActionFeedback(null);
      merge.commands.reset();
      inspector.commands.open();
    },
  };
}

type EntityInspectorProps = Parameters<typeof EntityWorkbookInspector>[0];
type EntitySelectedEdit = {
  readonly field: ViewContract["fields"][number] | null;
  readonly row: EntityRow | null;
};
type AliasAction =
  | { op: "add_alias"; alias_text: string }
  | { op: "remove_alias"; item_ref: string };

function EntityInspectorPresentation({
  details,
  inspector,
  isOpen,
  relationships,
}: {
  readonly details: EntityDetailsProps;
  readonly inspector: Omit<
    EntityInspectorProps,
    "detailsContent" | "relationshipsContent" | "evidenceContent"
  >;
  readonly isOpen: boolean;
  readonly relationships: EntityRelationshipsProps;
}) {
  if (!isOpen) return undefined;
  return (
    <EntityWorkbookInspector
      {...inspector}
      detailsContent={<EntityDetails {...details} />}
      evidenceContent={
        <p>
          Evidence linked to this{" "}
          {details.selectedEntity?.entityType ?? "record"}:{" "}
          {String(
            details.selectedEntity?.rawRow.cells[
              `${details.selectedEntity.entityType}.evidence_count`
            ]?.value ?? "Unavailable",
          )}
        </p>
      }
      relationshipsContent={[
        ownedInspectorRegion("timeline-preview", (present) => (
          <EntityRelationships {...relationships} present={present} />
        )),
      ]}
    />
  );
}

type EntityDetailsProps = {
  readonly patches: WorkbookMutationRuntime["explicitPatches"];
  readonly disabledReason: string | null;
  readonly editFeedback: ReturnType<typeof useWorkbookInspectorFieldFeedback>;
  readonly aliasFeedback: ReturnType<typeof useWorkbookInspectorFieldFeedback>;
  readonly aliasRemoveFeedback: ReturnType<
    typeof useWorkbookInspectorFieldFeedback
  >;
  readonly aliasDraft: string;
  readonly aliasInputRef: RefObject<HTMLInputElement | null>;
  readonly contract: ViewContract;
  readonly editableFields: ViewContract["fields"];
  readonly editFieldKey: string;
  readonly edit: WorkbookInspectorEditDraft;
  readonly aliasEdit: WorkbookInspectorEditDraft;
  readonly aliasRemove: WorkbookInspectorEditDraft;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly mutationPending: boolean;
  readonly rows: readonly EntityRow[];
  readonly selectedEdit: EntitySelectedEdit;
  readonly selectedEntity: EntityRow | null;
  readonly setAliasDraft: (value: string) => void;
  readonly setEditFieldKey: (value: string) => void;
  readonly setEditValue: (value: string | null) => void;
  readonly submitAliasActions: (actions: AliasAction[]) => Promise<void>;
  readonly submitEdit: () => Promise<void>;
};

function EntityDetails(props: EntityDetailsProps) {
  const editor = useEntityEditSlots(props);
  const aliases = useRef<HTMLDetailsElement>(null);
  return (
    <>
      {props.selectedEntity ? (
        <WorkbookInspectorDetails
          contract={props.contract}
          row={props.selectedEntity.rawRow}
          editableFields={props.editableFields}
          activeField={props.editFieldKey}
          onEdit={props.setEditFieldKey}
          onDetach={() => props.setEditFieldKey("")}
          disabledReason={props.disabledReason}
          canSubmit={!props.mutationPending && props.edit.canSubmit}
          onSubmit={() => void props.submitEdit()}
          editor={editor}
          retainedWork={props.edit.retainedWork}
          onReviewDraft={(identity) => props.setEditFieldKey(identity.fieldKey)}
          collectionDestinations={{
            [`${props.selectedEntity.entityType}.aliases`]: () => {
              if (!aliases.current) return;
              aliases.current.open = true;
              aliases.current
                .querySelector("summary")
                ?.focus({ preventScroll: true });
              aliases.current.scrollIntoView?.({ block: "nearest" });
            },
          }}
        />
      ) : null}
      <details ref={aliases}>
        <summary>Manage aliases</summary>
        <EntityAliases {...props} />
      </details>
      {props.selectedEntity ? (
        <EntityIdentifiers
          entity={props.selectedEntity}
          entityType={props.selectedEntity.entityType}
        />
      ) : null}
    </>
  );
}

function useEntityEditSlots(props: EntityDetailsProps) {
  const feedbackId = useId();
  if (!props.selectedEdit.field || !props.selectedEntity)
    return {
      content: null,
      actions: null,
      feedback: null,
      retainedDraft: null,
    };
  return {
    content: (
      <WorkbookInspectorEditControl
        invalid={props.editFeedback.message !== null}
        describedBy={props.editFeedback.message ? feedbackId : undefined}
        edit={{ ...props.edit, update: props.setEditValue }}
        ariaLabel={props.selectedEdit.field.label}
        collectionMode="add"
        field={props.selectedEdit.field}
        testId={genericEditValueTestId(props.contract.viewSchemaId)}
      />
    ),
    actions: (
      <WorkbookInspectorActionButton
        data-testid={genericEditSubmitTestId(props.contract.viewSchemaId)}
        disabled={props.mutationPending || !props.edit.canSubmit}
        tone="primary"
        type="button"
        onClick={() => void props.submitEdit()}
      >
        Update
      </WorkbookInspectorActionButton>
    ),
    feedback: (
      <>
        {props.editFeedback.message ? (
          <p id={feedbackId} role="alert">
            {props.editFeedback.message}
          </p>
        ) : null}
        {props.editFeedback.actionError ? (
          <WorkbookInspectorPublicError
            error={props.editFeedback.actionError}
          />
        ) : null}
        {props.mutationError ? (
          <WorkbookInspectorPublicError error={props.mutationError} />
        ) : null}
        <WorkbookExplicitPatchRecovery
          owner={props.patches}
          viewSchemaId={props.contract.viewSchemaId}
          recordId={props.selectedEntity.recordId}
          fieldKey={props.selectedEdit.field.fieldKey}
        />
      </>
    ),
    retainedDraft: (
      <WorkbookInspectorDraftFeedback
        edit={props.edit}
        contract={props.contract}
        row={props.selectedEntity?.rawRow ?? null}
      />
    ),
  };
}

function EntityAliases(props: EntityDetailsProps) {
  const feedbackId = useId();
  if (props.selectedEntity === null) return null;
  return (
    <section style={inspectorSectionStyle}>
      <h3 style={sectionTitleStyle}>Aliases</h3>
      <div style={entityAliasListStyle}>
        {props.selectedEntity.aliases.map((alias) => (
          <span key={alias.itemRef} style={tagChipStyle}>
            {alias.displayText}
            <button
              aria-label={`Remove alias ${alias.displayText}`}
              disabled={props.mutationPending || !props.aliasRemove.canSubmit}
              style={aliasRemoveButtonStyle}
              type="button"
              onClick={() =>
                void props.submitAliasActions([
                  { op: "remove_alias", item_ref: alias.itemRef },
                ])
              }
            >
              <X aria-hidden="true" size={12} />
            </button>
          </span>
        ))}
      </div>
      <WorkbookInspectorDraftFeedback
        edit={props.aliasEdit}
        contract={props.contract}
        row={props.selectedEntity.rawRow}
      />
      <div style={aliasAddRowStyle}>
        <input
          ref={(element) => {
            props.aliasInputRef.current = element;
            props.aliasEdit.controlRef.current = element;
          }}
          disabled={!props.aliasEdit.canEdit}
          aria-invalid={props.aliasFeedback.message ? true : undefined}
          aria-describedby={
            props.aliasFeedback.message ? feedbackId : undefined
          }
          aria-label="Alias text"
          maxLength={256}
          style={inputStyle}
          value={props.aliasDraft}
          onChange={(event) => props.setAliasDraft(event.target.value)}
        />
        <button
          disabled={
            props.mutationPending ||
            !props.aliasEdit.canSubmit ||
            props.aliasDraft.trim() === ""
          }
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() =>
            void props.submitAliasActions([
              { op: "add_alias", alias_text: props.aliasDraft },
            ])
          }
        >
          Add alias
        </button>
      </div>
      {props.aliasFeedback.message ? (
        <p id={feedbackId} role="alert">
          {props.aliasFeedback.message}
        </p>
      ) : null}
      {props.aliasFeedback.actionError ? (
        <WorkbookInspectorPublicError error={props.aliasFeedback.actionError} />
      ) : null}
      {props.aliasRemoveFeedback.message ? (
        <p role="alert">{props.aliasRemoveFeedback.message}</p>
      ) : null}
      {props.aliasRemoveFeedback.actionError ? (
        <WorkbookInspectorPublicError
          error={props.aliasRemoveFeedback.actionError}
        />
      ) : null}
    </section>
  );
}

type EntityRelationshipsProps = {
  readonly timelinePreviewNotice: ReturnType<
    typeof useEntityTimelinePreview
  >["timelinePreviewNotice"];
  readonly refreshTimelinePreview: (recordId: string) => void;
  readonly timelinePreviewState: ReturnType<
    typeof useEntityTimelinePreview
  >["timelinePreviewState"];
  readonly canMerge: boolean;
  readonly entityIndex: Record<string, EntityRow>;
  readonly entityType: EntityRow["entityType"];
  readonly merge: ReturnType<typeof useEntityMergeController>;
  readonly rows: readonly EntityRow[];
  readonly selectedEntity: EntityRow | null;
  readonly setEntityActionFeedback: Dispatch<
    SetStateAction<WorkbookInspectorFeedback | null>
  >;
  readonly timelinePreviewRows: ReturnType<
    typeof useEntityTimelinePreview
  >["timelinePreviewRows"];
};

function EntityRelationships(
  props: EntityRelationshipsProps & {
    readonly present: PresentInspectorRegion;
  },
) {
  const selected = props.selectedEntity;
  if (selected === null) return null;
  const current = props.timelinePreviewState.recordId === selected.recordId;
  const content: WorkbookInspectorPanelContentModel =
    props.timelinePreviewRows.length && current
      ? {
          kind: "populated",
          content:
            props.timelinePreviewRows.length > 0 ? (
              <section style={inspectorSectionStyle}>
                <h3 style={sectionTitleStyle}>Dependent Timeline</h3>
                <div style={timelinePreviewStackStyle}>
                  {props.timelinePreviewRows.map((row) => (
                    <article
                      key={row.recordId ?? row.key}
                      data-testid={
                        row.recordId === null
                          ? undefined
                          : timelinePreviewRowTestId(row.recordId)
                      }
                      style={timelinePreviewCardStyle}
                    >
                      <p style={noticeTitleStyle}>
                        {row.values.activitySynopsisText || "Untitled row"}
                      </p>
                      <div style={relationshipItemsWrapStyle}>
                        {row.collectionValues[
                          props.entityType === "host"
                            ? "hostRefs"
                            : "identityRefs"
                        ].map((item) => (
                          <details key={item.itemRef}>
                            <summary>
                              <WorkbookRelationshipChip
                                expanded
                                presentation={timelineRelationshipChipPresentation(
                                  {
                                    entityIndex: props.entityIndex,
                                    item,
                                    sourceRecordId: row.recordId,
                                  },
                                )}
                              />
                            </summary>
                            <WorkbookRelationshipChipDetails
                              presentation={timelineRelationshipChipPresentation(
                                {
                                  entityIndex: props.entityIndex,
                                  item,
                                  sourceRecordId: row.recordId,
                                },
                              )}
                            />
                          </details>
                        ))}
                      </div>
                    </article>
                  ))}
                </div>
              </section>
            ) : null,
        }
      : {
          kind: "empty",
          message:
            "No matching records in the loaded Timeline window. This preview does not establish that there are no relationships elsewhere.",
        };
  const state = props.timelinePreviewState;
  const data: WorkbookInspectorPanelData = !current
    ? {
        state: "unavailable",
        cause: "not_requested",
        message: "Timeline preview has not been loaded.",
      }
    : state.state === "initial_loading"
      ? { state: "initial_loading" }
      : state.state === "unavailable"
        ? {
            state: "unavailable",
            cause: "load_failed",
            message: state.message ?? "Could not load the Timeline preview.",
          }
        : state.state === "stale_failure"
          ? {
              state: "stale_failure",
              content,
              message:
                state.message ?? "Could not refresh the Timeline preview.",
            }
          : { state: state.state, content };
  return props.present({
    access: "readable",
    data,
    ...(current && props.timelinePreviewNotice
      ? { notice: props.timelinePreviewNotice }
      : {}),
    commands: (
      <>
        <WorkbookInspectorActionButton
          tone="secondary"
          disabled={
            state.state === "initial_loading" || state.state === "refreshing"
          }
          onClick={() => props.refreshTimelinePreview(selected.recordId)}
        >
          Refresh Timeline preview
        </WorkbookInspectorActionButton>
        <EntityMergePresentation
          canMerge={props.canMerge}
          merge={props.merge}
          rows={props.rows}
          selectedEntity={selected}
          setEntityActionFeedback={props.setEntityActionFeedback}
        />
      </>
    ),
  });
}

function EntityIdentifiers({
  entity,
  entityType,
}: {
  readonly entity: EntityRow;
  readonly entityType: EntityRow["entityType"];
}) {
  return (
    <>
      <section style={inspectorSectionStyle}>
        <h3 style={sectionTitleStyle}>Identifiers</h3>
        <ul style={flatListStyle}>
          {entity.identifiers.length > 0 ? (
            entity.identifiers.map((identifier) => (
              <li key={identifier.key}>
                {identifier.label}: {identifier.value}
              </li>
            ))
          ) : (
            <li>No exact-match identifiers visible.</li>
          )}
        </ul>
      </section>
      <section
        data-testid={entityReusableIdentifiersSectionTestId(
          entityType,
          entity.recordId,
        )}
        style={reusableIdentifierSectionStyle}
      >
        <div style={sectionHeadingRowStyle}>
          <h3 style={sectionTitleStyle}>Reusable identifiers</h3>
          <span style={readOnlyBadgeStyle}>Read-only</span>
        </div>
        <ul style={flatListStyle}>
          {entity.reusableIdentifiers.length > 0 ? (
            entity.reusableIdentifiers.map((identifier) => (
              <li
                data-testid={entityReusableIdentifierItemTestId(
                  entityType,
                  entity.recordId,
                  identifier.itemRef,
                )}
                key={identifier.itemRef}
              >
                {identifier.label}: {identifier.displayText}
              </li>
            ))
          ) : (
            <li>No reusable identifiers carried forward.</li>
          )}
        </ul>
      </section>
    </>
  );
}

function EntityMergePresentation({
  canMerge,
  merge,
  rows,
  selectedEntity,
  setEntityActionFeedback,
}: {
  readonly canMerge: boolean;
  readonly merge: ReturnType<typeof useEntityMergeController>;
  readonly rows: readonly EntityRow[];
  readonly selectedEntity: EntityRow;
  readonly setEntityActionFeedback: Dispatch<
    SetStateAction<WorkbookInspectorFeedback | null>
  >;
}) {
  const regionRef = useRef<HTMLElement | null>(null);
  if (!canMerge) {
    return (
      <section
        ref={regionRef}
        tabIndex={-1}
        aria-label="Entity merge review"
        style={inspectorSectionStyle}
      >
        <h3 style={sectionTitleStyle}>Merge</h3>
        <p style={bodyStyle}>
          Merging is unavailable with the current incident access.
        </p>
      </section>
    );
  }
  const { candidateId, loser, reason, reviewed, hasAffectedDraft } =
    merge.snapshot;
  const plan = reviewed?.plan ?? merge.snapshot.plan;
  return (
    <section
      ref={regionRef}
      tabIndex={-1}
      aria-label="Entity merge review"
      style={inspectorSectionStyle}
    >
      <h3 style={sectionTitleStyle}>Merge</h3>
      <label style={labelStyle}>
        Merge loser
        <select
          data-testid={entityMergeControlTestId("loser-record")}
          style={selectStyle}
          value={candidateId}
          onChange={(event) => {
            setEntityActionFeedback(null);
            merge.commands.selectCandidate(event.target.value);
          }}
        >
          <option value="">Select duplicate</option>
          {rows
            .filter(
              (row) =>
                row.recordId !== selectedEntity.recordId &&
                row.entityType === selectedEntity.entityType &&
                ["stub", "canonical"].includes(row.state),
            )
            .map((row) => (
              <option key={row.recordId} value={row.recordId}>
                {row.label} ({row.recordId})
              </option>
            ))}
        </select>
      </label>
      <label style={labelStyle}>
        Merge reason
        <input
          data-testid={entityMergeControlTestId("reason")}
          style={inputStyle}
          type="text"
          value={reason}
          onChange={(event) => merge.commands.setReason(event.target.value)}
        />
      </label>
      {loser && plan ? (
        <div
          data-testid={entityMergeControlTestId("plan")}
          style={mergePlanStyle}
        >
          <p style={noticeTitleStyle}>
            Survivor {selectedEntity.label} absorbs loser {loser.label}
          </p>
          <p style={bodyStyle}>
            Survivor record {selectedEntity.recordId}
            <br />
            Loser record {loser.recordId}
          </p>
          <ul style={flatListStyle}>
            {plan.identifierOutcomes.map((line) => (
              <li key={`${line.identifierClass}:${line.normalizedValue}`}>
                {line.label}: {mergeIdentifierOutcomeText(line)}
              </li>
            ))}
            <li>
              Aliases to copy:{" "}
              {plan.aliasesToCopy.length > 0
                ? plan.aliasesToCopy.join(", ")
                : "none"}
            </li>
            <li>
              Alias duplicate no-op:{" "}
              {plan.duplicateAliases.length > 0
                ? plan.duplicateAliases.join(", ")
                : "none"}
            </li>
            <li>Provenance-only values: {plan.provenanceOnlySummary}</li>
            <li>{plan.dependencySummary}</li>
          </ul>
          {plan.issues.length > 0 ? (
            <ul>
              {plan.issues.map((issue) => (
                <li key={issue}>{issue}</li>
              ))}
            </ul>
          ) : null}
          {hasAffectedDraft ? (
            <div>
              <p style={bodyStyle}>
                Finish or explicitly discard changes to both merge participants
                before reviewing.
              </p>
              <button
                style={secondaryActionButtonStyle}
                type="button"
                onClick={merge.commands.discardDrafts}
              >
                Discard participant drafts
              </button>
            </div>
          ) : null}
          {reviewed === null ? (
            <button
              data-testid={entityMergeControlTestId("review")}
              disabled={!plan.valid || hasAffectedDraft}
              style={secondaryActionButtonStyle}
              type="button"
              onClick={merge.commands.review}
            >
              Review merge
            </button>
          ) : (
            <>
              <p style={bodyStyle}>Reviewed reason: {reviewed.reason}</p>
              <WorkbookInspectorConfirmation
                operation="Merge"
                subject={`${reviewed.loser.label} (${reviewed.loser.recordId}, version ${reviewed.loser.baseRowVersion}) into ${reviewed.survivor.label} (${reviewed.survivor.recordId}, version ${reviewed.survivor.baseRowVersion})`}
                destructive
                confirmLabel="Confirm merge"
                confirmTestId={entityMergeControlTestId("confirm")}
                cancelTestId={entityMergeControlTestId("cancel")}
                onCancel={() => {
                  regionRef.current?.focus({ preventScroll: true });
                  merge.commands.invalidateReview();
                }}
                onConfirm={() => {
                  regionRef.current?.focus({ preventScroll: true });
                  void merge.commands.confirm();
                }}
              />
            </>
          )}
        </div>
      ) : (
        <button
          data-testid={entityMergeControlTestId("start")}
          style={secondaryActionButtonStyle}
          type="button"
          onClick={() => {
            setEntityActionFeedback(null);
            merge.commands.start();
          }}
        >
          Start merge
        </button>
      )}
    </section>
  );
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};
const inputStyle = {
  boxSizing: "border-box" as const,
  display: "block",
  minWidth: 0,
  width: "100%",
  borderRadius: "var(--ct-component-text-input-rounded)",
  border: "var(--ct-component-text-input-border)",
  background: "var(--ct-component-text-input-backgroundColor)",
  padding: "0.65rem 0.75rem",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
};
const actionButtonStyle = {
  borderRadius: "var(--ct-component-button-secondary-rounded)",
  border: "var(--ct-component-button-secondary-border)",
  background: "var(--ct-component-button-secondary-backgroundColor)",
  color: "var(--ct-component-button-secondary-textColor)",
  padding: "0.55rem 0.9rem",
  font: "inherit",
  cursor: "pointer",
};
const secondaryActionButtonStyle = {
  ...actionButtonStyle,
  background: "var(--ct-colors-surface-3)",
};
const labelStyle = {
  display: "grid",
  gap: "0.4rem",
  fontSize: "0.95rem",
  color: "var(--ct-colors-ink-muted)",
};
const inspectorSectionStyle = {
  display: "grid",
  gap: "0.75rem",
  marginBottom: "1rem",
};
const sectionTitleStyle = { margin: 0, fontSize: "1rem" };
const sectionHeadingRowStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  gap: "0.5rem",
};
const reusableIdentifierSectionStyle = {
  ...inspectorSectionStyle,
  borderInlineStart: "var(--ct-border-strong)",
  paddingInlineStart: "0.75rem",
};
const readOnlyBadgeStyle = {
  border: "var(--ct-border-hairline)",
  borderRadius: "999px",
  color: "var(--ct-colors-ink-muted)",
  fontSize: "0.75rem",
  lineHeight: 1,
  padding: "0.2rem 0.45rem",
};
const relationshipItemsWrapStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.4rem",
  marginBottom: "0.55rem",
  maxWidth: "100%",
  minWidth: 0,
};
const relationshipChipStyle = {
  display: "inline-flex",
  alignItems: "center",
  gap: "0.35rem",
  borderRadius: "var(--ct-component-chip-rounded)",
  padding: "var(--ct-component-chip-padding)",
  font: "inherit",
  lineHeight: 1.2,
  maxWidth: "100%",
  minWidth: 0,
  overflowWrap: "anywhere" as const,
};
const entityAliasListStyle = {
  display: "flex",
  flexWrap: "wrap" as const,
  gap: "0.35rem",
};
const tagChipStyle = {
  ...relationshipChipStyle,
  border: "var(--ct-component-chip-border)",
  background: "var(--ct-component-chip-backgroundColor)",
  color: "var(--ct-component-chip-textColor)",
};
const aliasRemoveButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  border: 0,
  background: "transparent",
  color: "inherit",
  cursor: "pointer",
  padding: 0,
};
const aliasAddRowStyle = {
  display: "grid",
  gridTemplateColumns: "minmax(0, 1fr) auto",
  gap: "0.5rem",
};
const noticeTitleStyle = { margin: 0, fontSize: "0.95rem", fontWeight: 600 };
const selectStyle = { ...inputStyle, appearance: "auto" as const };
const mergePlanStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.9rem",
  display: "grid",
  gap: "0.65rem",
};
const flatListStyle = {
  margin: 0,
  paddingLeft: "1.2rem",
  display: "grid",
  gap: "0.35rem",
};
const timelinePreviewStackStyle = { display: "grid", gap: "0.75rem" };
const timelinePreviewCardStyle = {
  borderRadius: "var(--ct-rounded-lg)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  padding: "0.85rem",
  display: "grid",
  gap: "0.55rem",
};
