import type { GridInteractionMode } from "@cartulary/grid-adapter";
import {
  entityInspectorTestId,
  entityMergeControlTestId,
  entityMergePreconditionDetailsTestId,
  entityReusableIdentifierItemTestId,
  entityReusableIdentifiersSectionTestId,
  genericEditFieldSelectTestId,
  genericEditRecordSelectTestId,
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
  useMemo,
  useRef,
  useState,
} from "react";
import { type SheetRef, sheetRefKey } from "../../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../../shared/workbookShellContracts";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import {
  WorkbookRelationshipChip,
  WorkbookRelationshipChipDetails,
} from "../../components/WorkbookRelationshipChip";
import { useWorkbookHistorySurfaceRefresh } from "../../history/WorkbookHistoryContext";
import { useEntityTimelinePreview } from "../../hooks/useEntityTimelinePreview";
import { inspectorRecordHistoryActions } from "../../inspector/inspectorCapabilityResolver";
import {
  WorkbookInspectorConfirmation,
  WorkbookInspectorPublicError,
} from "../../inspector/presentation/WorkbookInspectorFeedback";
import { useInspectorCreateRelatedWorkflow } from "../../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../../inspector/useWorkbookInspectorCoordinator";
import type {
  WorkbookInspectorErrorPresentation,
  WorkbookInspectorFeedback,
} from "../../inspector/workbookInspectorErrorModel";
import {
  workbookInspectorErrorPresentation,
  workbookInspectorLocalErrorPresentation,
} from "../../inspector/workbookInspectorErrorModel";
import {
  buildWorkbookInspectorSubject,
  type WorkbookInspectorSubject,
} from "../../inspector/workbookInspectorSubject";
import { mergeIdentifierOutcomeText } from "../../models/entityMergePlan";
import type { EntityRow } from "../../models/entityWorkbookModel";
import {
  buildGenericPatchChange,
  genericRowLabel,
  selectWorkbookEditTarget,
} from "../../models/genericWorkbookModel";
import { workbookInspectorStateIsOpen } from "../../models/workbookInspectorModel";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type {
  EntityMutationCommandPort,
  RecordRouteCommandPort,
  TimelineRelatedRecordPort,
} from "../../mutations/workbookMutationCommandPorts";
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
  mutationCommands,
  mutationError,
  mutationRuntime,
  mutationPending,
  onClearSurfaceSelection,
  onIncidentAccessLost,
  onRefreshEntities,
  onResetOwnerState,
  onRestoreFocus,
  recordMutationCommands,
  relatedMutationCommands,
  rows,
  selectedEntity,
  setEntityActionFeedback,
  setMutationError,
  setMutationPending,
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
  readonly mutationCommands: EntityMutationCommandPort;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly sheetRef: SheetRef;
  readonly mutationPending: boolean;
  readonly onClearSurfaceSelection: () => void;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly onRefreshEntities: (options?: {
    readonly requireAcceptance?: boolean;
  }) => Promise<void>;
  readonly onResetOwnerState: () => void;
  readonly onRestoreFocus: () => void;
  readonly recordMutationCommands: RecordRouteCommandPort;
  readonly relatedMutationCommands: TimelineRelatedRecordPort;
  readonly rows: readonly EntityRow[];
  readonly selectedEntity: EntityRow | null;
  readonly setEntityActionFeedback: Dispatch<
    SetStateAction<WorkbookInspectorFeedback | null>
  >;
  readonly setMutationError: Dispatch<
    SetStateAction<WorkbookInspectorErrorPresentation | null>
  >;
  readonly setMutationPending: Dispatch<SetStateAction<boolean>>;
  readonly setSelectedRecordId: Dispatch<SetStateAction<string | null>>;
  readonly viewQuery: WorkbookViewQueryPort;
}) {
  const inspectorConfig = contract.inspectorConfig;
  useWorkbookHistorySurfaceRefresh(inspectorConfig.viewSchemaId, () =>
    onRefreshEntities({ requireAcceptance: true }),
  );
  const [deletedHistorySubject, setDeletedHistorySubject] =
    useState<WorkbookInspectorSubject | null>(null);
  const [editRecordId, setEditRecordId] = useState("");
  const [editFieldKey, setEditFieldKey] = useState("");
  const [editValue, setEditValue] = useState("");
  const [aliasDraft, setAliasDraft] = useState("");
  const aliasInputRef = useRef<HTMLInputElement | null>(null);
  const subject: WorkbookInspectorSubject | null =
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
  const { clearTimelinePreview, loadTimelinePreview, timelinePreviewRows } =
    useEntityTimelinePreview({ entityType, viewQuery, onIncidentAccessLost });
  const beginMutation = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const merge = useEntityMergeController({
    canMerge:
      canMerge && !incidentClosed && interactionMode.kind === "editable",
    owner: mutationRuntime.entityMerge,
    originSurface: sheetRefKey(sheetRef),
    hasAffectedDraft: (ids) => {
      if (
        selectedEntity !== null &&
        ids.includes(selectedEntity.recordId) &&
        aliasDraft !== ""
      )
        return true;
      const edited = rows.find((row) => row.recordId === editRecordId);
      const field =
        contract.fields.find((field) => field.fieldKey === editFieldKey) ??
        contract.fields.find((field) => field.writeKind === "direct_value");
      return (
        edited !== undefined &&
        field !== undefined &&
        ids.includes(edited.recordId) &&
        editValue !== String(edited.rawRow.cells[field.fieldKey]?.value ?? "")
      );
    },
    discardAffectedDrafts: (ids) => {
      if (selectedEntity !== null && ids.includes(selectedEntity.recordId))
        setAliasDraft("");
      const edited = rows.find((row) => row.recordId === editRecordId);
      const field =
        contract.fields.find((field) => field.fieldKey === editFieldKey) ??
        contract.fields.find((field) => field.writeKind === "direct_value");
      if (
        edited !== undefined &&
        field !== undefined &&
        ids.includes(edited.recordId)
      )
        setEditValue(String(edited.rawRow.cells[field.fieldKey]?.value ?? ""));
    },
    lifecycleResetKey: inspectorResetKey,
    loadSurvivorPreview: (recordId) =>
      loadTimelinePreview(recordId, { requireAcceptance: true }),
    rows,
    selectedEntity,
  });
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ cause, scope }) => {
        onResetOwnerState();
        setEditRecordId("");
        setEditFieldKey("");
        setEditValue("");
        setAliasDraft("");
        merge.commands.clearPlan();
        clearTimelinePreview();
        setEntityActionFeedback(null);
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
    () => contract.fields.filter((field) => field.writeKind === "direct_value"),
    [contract],
  );
  const referenceOptions = useMemo(() => emptyGenericReferenceOptions(), []);
  const selectedEdit = selectWorkbookEditTarget({
    fieldKey: editFieldKey,
    fields: editableFields,
    getRecordId: (row: EntityRow) => row.recordId,
    recordId: editRecordId,
    rows,
  });
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
  useEffect(() => {
    if (selectedEdit.row === null || selectedEdit.field === null) {
      setEditValue("");
      return;
    }
    const value =
      selectedEdit.row.rawRow.cells[selectedEdit.field.fieldKey]?.value;
    setEditValue(value === null || value === undefined ? "" : String(value));
  }, [selectedEdit.field, selectedEdit.row]);

  async function submitEdit() {
    if (selectedEdit.row === null || selectedEdit.field === null) {
      setMutationError(
        workbookInspectorLocalErrorPresentation("invalid_mutation_payload"),
      );
      return;
    }
    const change = buildGenericPatchChange(selectedEdit.field, editValue);
    if (change === null) {
      setMutationError(
        workbookInspectorLocalErrorPresentation(
          "Provide a value, or leave clearable fields empty to clear them.",
        ),
      );
      return;
    }
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      setMutationPending(true);
      setMutationError(null);
      const result = await mutationCommands.patchRecord({
        baseRowVersion: selectedEdit.row.rowVersion,
        changes: [change],
        purpose: "entity-patch",
        recordId: selectedEdit.row.recordId,
        viewSchemaId: contract.viewSchemaId,
      });
      if (result.kind === "rejected") {
        if (result.failure.kind === "same_field_conflict") {
          mutationRuntime.registerConflict({
            sheetRef,
            conflict: result.failure.conflict,
            focusKey: `${selectedEdit.row.recordId}:${selectedEdit.field.fieldKey}`,
            rowLabel: selectedEdit.row.label,
            surfaceLabel: contract.title,
            viewSchemaId: contract.viewSchemaId,
          });
        }
        setMutationPending(false);
        setMutationError(workbookInspectorErrorPresentation(result.failure));
        return;
      }
      await onRefreshEntities();
      setSelectedRecordId(selectedEdit.row.recordId);
      setMutationPending(false);
    } finally {
      finishMutation();
    }
  }

  async function submitAliasActions(
    actions: Array<
      | { op: "add_alias"; alias_text: string }
      | { op: "remove_alias"; item_ref: string }
    >,
  ) {
    if (selectedEntity === null) {
      setMutationError(
        workbookInspectorLocalErrorPresentation("invalid_mutation_payload"),
      );
      return;
    }
    const [firstAction, ...remainingActions] = actions;
    if (firstAction === undefined) {
      setMutationError(
        workbookInspectorLocalErrorPresentation("invalid_mutation_payload"),
      );
      return;
    }
    const aliasFieldKey =
      entityType === "host" ? "host.aliases" : "identity.aliases";
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      setMutationPending(true);
      setMutationError(null);
      const result = await mutationCommands.patchRecord({
        baseRowVersion: selectedEntity.rowVersion,
        changes: [
          {
            field_key: aliasFieldKey,
            action_payload: {
              actions: [firstAction, ...remainingActions],
              kind: "collection_actions_v1",
            },
          },
        ],
        purpose: `entity-alias-${selectedEntity.recordId}`,
        recordId: selectedEntity.recordId,
        viewSchemaId: contract.viewSchemaId,
      });
      if (result.kind === "rejected") {
        if (result.failure.kind === "same_field_conflict") {
          mutationRuntime.registerConflict({
            sheetRef,
            conflict: result.failure.conflict,
            focusKey: `${selectedEntity.recordId}:${aliasFieldKey}`,
            rowLabel: selectedEntity.label,
            surfaceLabel: contract.title,
            viewSchemaId: contract.viewSchemaId,
          });
        }
        setMutationPending(false);
        setMutationError(workbookInspectorErrorPresentation(result.failure));
        return;
      }
      setAliasDraft("");
      await onRefreshEntities();
      setSelectedRecordId(selectedEntity.recordId);
      setMutationPending(false);
      requestAnimationFrame(() => aliasInputRef.current?.focus());
    } finally {
      finishMutation();
    }
  }

  const close = () => inspector.commands.close({ restoreFocus: true });
  const mergeSnapshot = merge.snapshot;
  const node = isOpen ? (
    <EntityInspectorPresentation
      details={{
        aliasDraft,
        aliasInputRef,
        contract,
        editableFields,
        editFieldKey,
        editRecordId,
        editValue,
        mutationError,
        mutationPending,
        referenceOptions,
        rows,
        selectedEdit,
        selectedEntity,
        setAliasDraft: (value) => {
          merge.commands.invalidateReview();
          setAliasDraft(value);
        },
        setEditFieldKey: (value) => {
          merge.commands.invalidateReview();
          setEditFieldKey(value);
        },
        setEditRecordId: (value) => {
          merge.commands.invalidateReview();
          setEditRecordId(value);
        },
        setEditValue: (value) => {
          if (
            editRecordId === selectedEntity?.recordId ||
            editRecordId === merge.snapshot.loser?.recordId
          )
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
          beginMutation,
          actions: recordHistoryActions,
          canMutate:
            interactionMode.kind === "editable" &&
            currentIncidentRole !== null &&
            currentIncidentRole !== "viewer",
          commands: recordMutationCommands,
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
          referenceOptions,
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
        canMerge: canMerge && mutationRuntime.entityMerge.canSubmit(),
        entityIndex,
        entityType,
        merge,
        rows,
        selectedEntity,
        setEntityActionFeedback,
        timelinePreviewRows,
      }}
    />
  ) : undefined;
  return {
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
    "detailsContent" | "relationshipsContent"
  >;
  readonly isOpen: boolean;
  readonly relationships: EntityRelationshipsProps;
}) {
  if (!isOpen) return undefined;
  return (
    <EntityWorkbookInspector
      {...inspector}
      detailsContent={<EntityDetails {...details} />}
      relationshipsContent={<EntityRelationships {...relationships} />}
    />
  );
}

type EntityDetailsProps = {
  readonly aliasDraft: string;
  readonly aliasInputRef: RefObject<HTMLInputElement | null>;
  readonly contract: ViewContract;
  readonly editableFields: ViewContract["fields"];
  readonly editFieldKey: string;
  readonly editRecordId: string;
  readonly editValue: string;
  readonly mutationError: WorkbookInspectorErrorPresentation | null;
  readonly mutationPending: boolean;
  readonly referenceOptions: ReturnType<typeof emptyGenericReferenceOptions>;
  readonly rows: readonly EntityRow[];
  readonly selectedEdit: EntitySelectedEdit;
  readonly selectedEntity: EntityRow | null;
  readonly setAliasDraft: (value: string) => void;
  readonly setEditFieldKey: (value: string) => void;
  readonly setEditRecordId: (value: string) => void;
  readonly setEditValue: (value: string) => void;
  readonly submitAliasActions: (actions: AliasAction[]) => Promise<void>;
  readonly submitEdit: () => Promise<void>;
};

function EntityDetails(props: EntityDetailsProps) {
  return (
    <>
      <EntityEditCell {...props} />
      <EntityAliases {...props} />
      {props.selectedEntity ? (
        <EntityIdentifiers
          entity={props.selectedEntity}
          entityType={props.selectedEntity.entityType}
        />
      ) : null}
    </>
  );
}

function EntityEditCell(props: EntityDetailsProps) {
  if (props.editableFields.length === 0 || props.rows.length === 0) return null;
  return (
    <section style={inspectorSectionStyle}>
      <h3 style={sectionTitleStyle}>Edit cell</h3>
      <div style={inspectorControlStackStyle}>
        <select
          data-testid={genericEditRecordSelectTestId(
            props.contract.viewSchemaId,
          )}
          style={selectStyle}
          value={props.editRecordId}
          onChange={(event) => props.setEditRecordId(event.target.value)}
        >
          <option value="">Row</option>
          {props.rows.map((row) => (
            <option key={row.recordId} value={row.recordId}>
              {genericRowLabel(props.contract, row.rawRow)}
            </option>
          ))}
        </select>
        <select
          data-testid={genericEditFieldSelectTestId(
            props.contract.viewSchemaId,
          )}
          style={selectStyle}
          value={props.editFieldKey}
          onChange={(event) => props.setEditFieldKey(event.target.value)}
        >
          <option value="">Field</option>
          {props.editableFields.map((field) => (
            <option key={field.fieldKey} value={field.fieldKey}>
              {field.label}
            </option>
          ))}
        </select>
        {props.selectedEdit.field ? (
          <GenericMutationControl
            collectionMode="add"
            field={props.selectedEdit.field}
            referenceOptions={props.referenceOptions}
            testId={genericEditValueTestId(props.contract.viewSchemaId)}
            value={props.editValue}
            onChange={props.setEditValue}
          />
        ) : null}
        <button
          data-testid={genericEditSubmitTestId(props.contract.viewSchemaId)}
          disabled={props.mutationPending}
          style={actionButtonStyle}
          type="button"
          onClick={() => void props.submitEdit()}
        >
          Update
        </button>
      </div>
      {props.mutationError ? (
        <WorkbookInspectorPublicError error={props.mutationError} />
      ) : null}
    </section>
  );
}

function EntityAliases(props: EntityDetailsProps) {
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
              disabled={props.mutationPending}
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
      <div style={aliasAddRowStyle}>
        <input
          ref={props.aliasInputRef}
          aria-label="Alias text"
          maxLength={256}
          style={inputStyle}
          value={props.aliasDraft}
          onChange={(event) => props.setAliasDraft(event.target.value)}
        />
        <button
          disabled={props.mutationPending || props.aliasDraft.trim() === ""}
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
    </section>
  );
}

type EntityRelationshipsProps = {
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

function EntityRelationships(props: EntityRelationshipsProps) {
  if (props.selectedEntity === null) return null;
  return (
    <>
      <EntityMergePresentation
        canMerge={props.canMerge}
        merge={props.merge}
        rows={props.rows}
        selectedEntity={props.selectedEntity}
        setEntityActionFeedback={props.setEntityActionFeedback}
      />
      {props.timelinePreviewRows.length > 0 ? (
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
                    props.entityType === "host" ? "hostRefs" : "identityRefs"
                  ].map((item) => (
                    <details key={item.itemRef}>
                      <summary>
                        <WorkbookRelationshipChip
                          expanded
                          presentation={timelineRelationshipChipPresentation({
                            entityIndex: props.entityIndex,
                            item,
                            sourceRecordId: row.recordId,
                          })}
                        />
                      </summary>
                      <WorkbookRelationshipChipDetails
                        presentation={timelineRelationshipChipPresentation({
                          entityIndex: props.entityIndex,
                          item,
                          sourceRecordId: row.recordId,
                        })}
                      />
                    </details>
                  ))}
                </div>
              </article>
            ))}
          </div>
        </section>
      ) : null}
    </>
  );
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
const inspectorControlStackStyle = { display: "grid", gap: "0.65rem" };
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
