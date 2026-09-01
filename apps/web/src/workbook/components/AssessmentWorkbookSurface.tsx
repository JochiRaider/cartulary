import {
  type GridColumn,
  type GridDataRow,
  type GridGroupingDescriptor,
  type GridHandle,
  GridViewport,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  assessmentCreateControlTestId,
  assessmentCreatePanelTestId,
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
} from "@cartulary/ui-contracts";
import type { InspectorDisabledCondition } from "@cartulary/view-contracts";
import {
  type InspectorFeatureGroup,
  requireViewContract,
  resolveHeaderSortFieldKey,
} from "@cartulary/view-contracts";
import {
  type ReactNode,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { useWorkbookCollaborationCoordinator } from "../collaboration/useWorkbookCollaborationCoordinator";
import type { WorkbookCollaborationCoordinator } from "../collaboration/WorkbookCollaborationCoordinator";
import {
  useWorkbookGridContinuity,
  WorkbookContinuityCell,
} from "../continuity/useWorkbookGridContinuity";
import type {
  WorkbookContinuityPort,
  WorkbookContinuityToken,
} from "../continuity/workbookContinuityPort";
import { useAssessmentCreationController } from "../features/assessments/useAssessmentCreationController";
import { useAssessmentSupportCandidates } from "../hooks/useAssessmentSupportCandidates";
import { InspectorCreateRelatedWorkflow } from "../inspector/InspectorCreateRelatedWorkflow";
import { WorkbookInspectorPublicError } from "../inspector/presentation/WorkbookInspectorFeedback";
import {
  WorkbookInspectorPanelSection,
  WorkbookInspectorShell,
} from "../inspector/presentation/WorkbookInspectorShell";
import type { WorkbookInspectorSubjectPresentation } from "../inspector/presentation/workbookInspectorPresentationModel";
import { useInspectorCreateRelatedWorkflow } from "../inspector/useInspectorCreateRelatedWorkflow";
import { useWorkbookInspectorCoordinator } from "../inspector/useWorkbookInspectorCoordinator";
import { WorkbookInspectorContextualActions } from "../inspector/WorkbookInspectorContextualActions";
import { WorkbookInspectorRecordHistory } from "../inspector/WorkbookInspectorRecordHistory";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import {
  WorkbookSurfaceLayout,
  workbookSurfaceGridShellStyle,
} from "../layout/WorkbookSurfaceLayout";
import { applyWorkbookLayoutToColumns } from "../layout/workbookColumnLayout";
import {
  assessmentColumnWidth,
  isAssessmentConfidenceBand,
} from "../models/assessmentWorkbookModel";
import type { EntityRow } from "../models/entityWorkbookModel";
import {
  enumValuesFor,
  genericCellLabel,
  genericInspectorRowLabel,
} from "../models/genericWorkbookModel";
import { workbookGridRows } from "../models/workbookContractRows";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../models/workbookGridState";
import { inspectorPanelIsDeclared } from "../models/workbookInspectorModel";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { emptyGenericReferenceOptions } from "../models/workbookReferenceOptions";
import { assessmentsViewSchemaId } from "../models/workbookSurfaceRegistry";
import type {
  AssessmentMutationCommandPort,
  RecordRouteCommandPort,
  TimelineRelatedRecordPort,
} from "../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import { useWorkbookMutationRuntime } from "../runtime/useWorkbookMutationRuntime";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { WorkbookCellPresenceMarker } from "./WorkbookPresenceMarkers";
import { WorkbookRecordCandidatePicker } from "./WorkbookRecordCandidatePicker";
import {
  type WorkbookConflictActivation,
  WorkbookSurfaceStatusStrip,
} from "./WorkbookStatusStrip";
import { WorkbookViewBar } from "./WorkbookViewBar";

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

export type AssessmentWorkbookSurfaceProps = {
  assessmentRows: WorkbookQueryRow[];
  continuityResetKey: string;
  currentIncidentRole: WorkbookIncidentRole | null;
  currentUserId: string | null;
  inspectorResetKey: string;
  queryControls?: ReactNode | undefined;
  savedViewSelector?: ReactNode | undefined;
  hostRows: EntityRow[];
  identityRows: EntityRow[];
  layout: WorkbookSurfaceLayoutOwner;
  loadState: WorkbookQueryLoadState;
  mutationRuntime: WorkbookMutationRuntime;
  mutationCommands: AssessmentMutationCommandPort;
  onActivateConflict?: WorkbookConflictActivation | undefined;
  recordMutationCommands: RecordRouteCommandPort;
  relatedMutationCommands: TimelineRelatedRecordPort;
  collaborationProjection: WorkbookCollaborationCoordinator;
  onClearFilters: () => void;
  onRefreshAssessmentRows: () => Promise<void>;
  onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  queryState: WorkbookQueryState;
  viewQuery: WorkbookViewQueryPort;
};

export function AssessmentWorkbookSurface({
  assessmentRows,
  continuityResetKey,
  currentIncidentRole,
  currentUserId,
  inspectorResetKey,
  queryControls,
  savedViewSelector,
  hostRows,
  identityRows,
  layout,
  loadState,
  mutationRuntime,
  mutationCommands,
  onActivateConflict,
  recordMutationCommands,
  relatedMutationCommands,
  collaborationProjection,
  onClearFilters,
  onRefreshAssessmentRows,
  onSortChange,
  queryState,
  viewQuery,
}: AssessmentWorkbookSurfaceProps) {
  const {
    commands: { onColumnReorder, onColumnWidthChange },
    snapshot: {
      chromeMode,
      density,
      incidentClosed,
      interactionMode,
      showStatusPresence,
      state: layoutState,
    },
  } = layout;
  const [selectedAssessmentRecordId, setSelectedAssessmentRecordId] = useState<
    string | null
  >(null);
  const [selectedAssessmentSnapshot, setSelectedAssessmentSnapshot] =
    useState<WorkbookQueryRow | null>(null);
  const inspectorContinuityTokenRef = useRef<WorkbookContinuityToken | null>(
    null,
  );
  const continuityPortRef = useRef<WorkbookContinuityPort | null>(null);
  const mutation = useWorkbookMutationRuntime(
    mutationRuntime,
    assessmentsViewSchemaId,
  );
  const collaboration = useWorkbookCollaborationCoordinator(
    collaborationProjection,
  );
  const inspectorConfig = assessmentsContract.inspectorConfig;
  const showWorkflowPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "workflow",
  );
  const supportCandidates = useAssessmentSupportCandidates({
    viewQuery,
  });
  const roleCanCreate =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
  const canCreate = interactionMode.kind === "editable" && roleCanCreate;
  const selectedAssessment =
    assessmentRows.find(
      (row) => row.record_id === selectedAssessmentRecordId,
    ) ??
    (selectedAssessmentSnapshot?.record_id === selectedAssessmentRecordId
      ? selectedAssessmentSnapshot
      : null);
  const createRelatedReferenceOptions = useMemo(
    () => emptyGenericReferenceOptions(),
    [],
  );
  const beginAssessmentMutation = useCallback(
    () => mutationRuntime.beginExplicitMutation(),
    [mutationRuntime],
  );
  const subjectRecordIds = useMemo(
    () => ({
      host: hostRows.map((row) => row.recordId),
      identity: identityRows.map((row) => row.recordId),
    }),
    [hostRows, identityRows],
  );
  const assessmentCreation = useAssessmentCreationController({
    beginMutation: beginAssessmentMutation,
    lifecycleResetKey: inspectorResetKey,
    mutationCommands,
    onRefreshAssessmentRows,
    subjectRecordIds,
  });
  const createRelatedWorkflow = useInspectorCreateRelatedWorkflow({
    currentUserId,
    mutationCommands: relatedMutationCommands,
    onCreated: onRefreshAssessmentRows,
    onMessage: (message) => {
      if (message !== null) assessmentCreation.commands.rejectStart(message);
    },
    selectedSubject:
      selectedAssessment === null
        ? null
        : {
            cells: selectedAssessment.cells,
            recordId: selectedAssessment.record_id,
            rowVersion: selectedAssessment.row_version,
          },
  });
  const { draft, draftMode, isSubmitting, message } =
    assessmentCreation.snapshot;
  const subjectRows = draft.subjectType === "host" ? hostRows : identityRows;
  const inspector = useWorkbookInspectorCoordinator({
    actionPorts: {
      resetOwnerState: ({ scope }) => {
        assessmentCreation.commands.reset();
        if (scope === "surface") {
          continuityPortRef.current?.clear();
          setSelectedAssessmentRecordId(null);
          setSelectedAssessmentSnapshot(null);
        }
      },
      restoreFocus: () => {
        const token = inspectorContinuityTokenRef.current;
        inspectorContinuityTokenRef.current = null;
        if (token !== null) {
          continuityPortRef.current?.restore(token);
        }
      },
    },
    config: inspectorConfig,
    lifecycleKey: inspectorResetKey,
    subject:
      selectedAssessment === null
        ? null
        : {
            recordId: selectedAssessment.record_id,
            rowVersion: selectedAssessment.row_version,
            viewSchemaId: assessmentsViewSchemaId,
          },
  });
  const isInspectorOpen = inspector.snapshot.isOpen;
  const assessmentInspectorDisabledTokens = useMemo(() => {
    const tokens = new Set<InspectorDisabledCondition>();
    if (selectedAssessment === null) {
      tokens.add("no_row_selected");
    } else {
      tokens.add("record_not_deleted");
    }
    tokens.add("rollback_target_unavailable");
    if (!roleCanCreate) {
      tokens.add("authorization_lost");
    } else if (incidentClosed) {
      tokens.add("incident_closed");
    }
    return tokens;
  }, [incidentClosed, roleCanCreate, selectedAssessment]);
  const stateOptions = enumValuesFor(
    assessmentsContract,
    "assessment.assessment_state",
    ["unknown", "suspected", "confirmed", "disproven", "cleared"],
  );
  const confidenceBandOptions = enumValuesFor(
    assessmentsContract,
    "assessment.confidence_band",
    ["unset", "low", "medium", "high"],
  ).filter(isAssessmentConfidenceBand);
  const anchorColumns = useMemo<readonly GridColumn<WorkbookQueryRow>[]>(
    () =>
      applyWorkbookLayoutToColumns(
        assessmentsContract,
        assessmentsContract.fields.map((field) => ({
          fieldKey: field.fieldKey,
          headerTestId: gridSortHeaderTestId(
            assessmentsViewSchemaId,
            field.fieldKey,
          ),
          label: field.label,
          width: assessmentColumnWidth(field.fieldKey),
          renderCell: () => null,
          sortableFieldKey: resolveHeaderSortFieldKey(
            assessmentsContract,
            field.fieldKey,
          ),
        })),
        layoutState,
      ),
    [layoutState],
  );
  const gridRows = useMemo<readonly GridDataRow<WorkbookQueryRow>[]>(
    () =>
      workbookGridRows({
        getRecordId: (row) => row.record_id,
        getRowVersion: (row) => row.row_version,
        rows: assessmentRows,
        surface: assessmentsViewSchemaId,
      }),
    [assessmentRows],
  );
  const grouping =
    useMemo<GridGroupingDescriptor<WorkbookQueryRow> | null>(() => {
      const fieldKey = queryState.groupBy;
      if (fieldKey === null) {
        return null;
      }
      return {
        fieldKey,
        formatLabel: (value) => genericCellLabel(value),
        getTestId: (groupFieldKey, _value, label) =>
          label === null
            ? undefined
            : gridGroupRowTestId(assessmentsViewSchemaId, groupFieldKey, label),
        getValue: (row) => {
          const value = row.cells[fieldKey]?.value;
          return value === null ||
            typeof value === "boolean" ||
            typeof value === "number" ||
            typeof value === "string"
            ? value
            : null;
        },
        label: assessmentsContract.fieldMap[fieldKey]?.label ?? fieldKey,
      };
    }, [queryState.groupBy]);
  const gridHandleRef = useRef<GridHandle | null>(null);
  const assessmentFocus = useWorkbookGridContinuity({
    columns: anchorColumns,
    continuityResetKey,
    gridHandleRef,
    viewSchemaId: assessmentsViewSchemaId,
  });
  continuityPortRef.current = assessmentFocus.port;
  const dataState = workbookGridDataState({
    emptyAction: canCreate
      ? {
          label: "Add assessment",
          onInvoke: openStandaloneDraft,
        }
      : undefined,
    emptyMessage: "No assessments have been recorded.",
    loadState,
    onClearFilters,
    onRetry: () => void onRefreshAssessmentRows(),
    queryState,
    rowCount: gridRows.length,
    surfaceLabel: assessmentsContract.title,
  });
  const columns: readonly GridColumn<WorkbookQueryRow>[] = anchorColumns.map(
    (field) => ({
      ...field,
      getClipboardValue: (row) =>
        genericCellLabel(row.cells[field.fieldKey]?.value),
      renderCell: ({ row }) => (
        <WorkbookContinuityCell
          continuity={assessmentFocus.port}
          fieldKey={field.fieldKey}
          recordId={row.record_id}
          viewSchemaId={assessmentsViewSchemaId}
        >
          {genericCellLabel(row.cells[field.fieldKey]?.value)}
          <WorkbookCellPresenceMarker
            fieldKey={field.fieldKey}
            fieldLabel={field.label}
            presences={collaborationProjection.editingPresenceForCell(
              row.record_id,
              field.fieldKey,
            )}
            recordId={row.record_id}
          />
        </WorkbookContinuityCell>
      ),
    }),
  );

  function openStandaloneDraft() {
    inspectorContinuityTokenRef.current = assessmentFocus.port.capture();
    assessmentCreation.commands.openStandalone(hostRows[0]?.recordId ?? "");
    inspector.commands.open();
  }

  function cancelAssessmentDraft() {
    assessmentCreation.commands.cancel();
    inspector.commands.close({ restoreFocus: true });
  }

  function selectAssessment(recordId: string) {
    setSelectedAssessmentRecordId(recordId);
    const row = assessmentRows.find(
      (candidate) => candidate.record_id === recordId,
    );
    if (row !== undefined) {
      setSelectedAssessmentSnapshot(row);
    }
  }

  async function executeAssessmentRecordLifecycle(
    featureGroup: InspectorFeatureGroup,
  ): Promise<boolean> {
    if (
      featureGroup.routeBinding.kind !== "record_action" ||
      featureGroup.routeBinding.owner !== "record_delete_route"
    ) {
      return false;
    }
    if (selectedAssessment === null) {
      assessmentCreation.commands.rejectStart(
        "Select a saved row before running this action.",
      );
      return true;
    }
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      const outcome = await recordMutationCommands.execute({
        action: "delete",
        baseRowVersion: selectedAssessment.row_version,
        reason: "Deleted from the Assessments inspector",
        recordId: selectedAssessment.record_id,
      });
      if (outcome.kind === "rejected") {
        assessmentCreation.commands.rejectFailure(outcome.failure);
        return true;
      }
      setSelectedAssessmentRecordId(null);
      setSelectedAssessmentSnapshot(null);
      inspector.commands.completeAction();
      await onRefreshAssessmentRows();
      return true;
    } finally {
      finishMutation();
    }
  }

  function beginAssessmentFeatureAction(featureGroup: InspectorFeatureGroup) {
    if (
      featureGroup.featureGroupKey !== "create_related.assessment" &&
      createRelatedWorkflow.commands.begin(featureGroup)
    ) {
      return;
    }
    if (featureGroup.routeBinding.kind === "record_action") {
      void executeAssessmentRecordLifecycle(featureGroup);
      return;
    }
    if (featureGroup.featureGroupKey !== "create_related.assessment") {
      return;
    }
    if (
      featureGroup.routeBinding.kind !== "view_row_create" ||
      featureGroup.routeBinding.owner !== "view_row_create_route" ||
      featureGroup.routeBinding.targetViewSchemaId !== assessmentsViewSchemaId
    ) {
      assessmentCreation.commands.rejectStart(
        "Assessment follow-on creation is unavailable.",
      );
      return;
    }
    if (!canCreate) {
      assessmentCreation.commands.rejectStart(
        "Assessment creation requires an active editor role.",
      );
      return;
    }
    if (!assessmentCreation.commands.openFollowOn(selectedAssessment)) return;
    if (!inspector.snapshot.isOpen) {
      inspectorContinuityTokenRef.current = assessmentFocus.port.capture();
    }
    inspector.commands.open();
  }

  useEffect(() => {
    if (selectedAssessmentRecordId === null) {
      return;
    }
    const refreshed = assessmentRows.find(
      (row) => row.record_id === selectedAssessmentRecordId,
    );
    if (refreshed !== undefined) {
      setSelectedAssessmentSnapshot(refreshed);
    }
  }, [assessmentRows, selectedAssessmentRecordId]);

  useEffect(
    () =>
      mutationRuntime.registerSurface(
        assessmentsViewSchemaId,
        onRefreshAssessmentRows,
      ),
    [mutationRuntime, onRefreshAssessmentRows],
  );
  const inspectorSubjectPresentation: WorkbookInspectorSubjectPresentation | null =
    selectedAssessment === null
      ? null
      : {
          label: genericInspectorRowLabel(
            assessmentsContract,
            selectedAssessment,
          ),
          recordId: selectedAssessment.record_id,
          rowVersion: selectedAssessment.row_version,
          stateLabel: `Follow-on subject: ${draft.subjectRecordId || "not selected"}`,
          surfaceLabel: assessmentsContract.title,
        };

  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={
        isInspectorOpen && showWorkflowPanel ? (
          <WorkbookInspectorShell
            accessibleLabel="Compromise Assessments inspector"
            eyebrow="Create"
            heading={
              draftMode === "follow_on"
                ? "Append follow-on assessment"
                : "Append assessment"
            }
            mode="creation"
            noRowHeading="Append assessment"
            subject={inspectorSubjectPresentation}
            testId={assessmentCreatePanelTestId()}
            viewSchemaId={assessmentsViewSchemaId}
            onClose={cancelAssessmentDraft}
          >
            {inspectorConfig.panels.map((panel) => (
              <WorkbookInspectorPanelSection
                config={inspectorConfig}
                key={panel.panelId}
                panelId={panel.panelId}
              >
                {inspectorSubjectPresentation === null ? null : (
                  <WorkbookInspectorContextualActions
                    config={inspectorConfig}
                    currentIncidentRole={currentIncidentRole}
                    disabledTokens={assessmentInspectorDisabledTokens}
                    featureGroups={inspectorConfig.featureGroups.filter(
                      (featureGroup) =>
                        featureGroup.panelId === panel.panelId &&
                        (featureGroup.routeBinding.kind === "view_row_create" ||
                          (featureGroup.routeBinding.kind === "record_action" &&
                            featureGroup.routeBinding.owner ===
                              "record_delete_route")),
                    )}
                    subject={inspectorSubjectPresentation}
                    onAction={beginAssessmentFeatureAction}
                  />
                )}
                {panel.panelId === "relationships" &&
                selectedAssessment !== null ? (
                  <p style={bodyStyle}>
                    Supporting records:{" "}
                    {genericCellLabel(
                      selectedAssessment.cells["assessment.support_refs"]
                        ?.value,
                    )}
                  </p>
                ) : null}
                {createRelatedWorkflow.snapshot.workflow?.featureGroup
                  .panelId === panel.panelId ? (
                  <InspectorCreateRelatedWorkflow
                    referenceOptions={createRelatedReferenceOptions}
                    state={createRelatedWorkflow.snapshot.workflow}
                    onCancel={createRelatedWorkflow.commands.cancel}
                    onSubmit={() => {
                      void createRelatedWorkflow.commands.submit();
                    }}
                    onUpdateDraft={createRelatedWorkflow.commands.updateDraft}
                  />
                ) : null}
                {panel.panelId === "history" ? (
                  <WorkbookInspectorRecordHistory
                    canMutate={canCreate}
                    commands={recordMutationCommands}
                    subject={
                      selectedAssessment === null
                        ? null
                        : {
                            recordId: selectedAssessment.record_id,
                            rowVersion: selectedAssessment.row_version,
                          }
                    }
                    onMessage={assessmentCreation.commands.rejectStart}
                    onRefresh={onRefreshAssessmentRows}
                  />
                ) : null}
              </WorkbookInspectorPanelSection>
            ))}
            <div style={inspectorSectionStyle}>
              <label style={labelStyle}>
                Subject type
                <select
                  data-testid={assessmentCreateControlTestId("subject-type")}
                  disabled={isSubmitting}
                  style={selectStyle}
                  value={draft.subjectType}
                  onChange={(event) => {
                    const subjectType =
                      event.target.value === "identity" ? "identity" : "host";
                    const nextRows =
                      subjectType === "host" ? hostRows : identityRows;
                    assessmentCreation.commands.updateDraft((current) => ({
                      ...current,
                      subjectType,
                      subjectRecordId: nextRows[0]?.recordId ?? "",
                    }));
                  }}
                >
                  {enumValuesFor(
                    assessmentsContract,
                    "assessment.subject_type",
                    ["host", "identity"],
                  ).map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>

              <label style={labelStyle}>
                Subject
                <select
                  data-testid={assessmentCreateControlTestId("subject")}
                  disabled={isSubmitting}
                  style={selectStyle}
                  value={draft.subjectRecordId}
                  onChange={(event) => {
                    assessmentCreation.commands.updateDraft((current) => ({
                      ...current,
                      subjectRecordId: event.target.value,
                    }));
                  }}
                >
                  <option value="">Select subject</option>
                  {subjectRows.map((row) => (
                    <option key={row.recordId} value={row.recordId}>
                      {row.label}
                    </option>
                  ))}
                </select>
              </label>

              <label style={labelStyle}>
                State
                <select
                  data-testid={assessmentCreateControlTestId("state")}
                  disabled={isSubmitting}
                  style={selectStyle}
                  value={draft.assessmentState}
                  onChange={(event) => {
                    assessmentCreation.commands.updateDraft((current) => ({
                      ...current,
                      assessmentState: event.target.value,
                    }));
                  }}
                >
                  {stateOptions.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>

              <label style={labelStyle}>
                Confidence
                <select
                  data-testid={assessmentCreateControlTestId("confidence-band")}
                  disabled={isSubmitting}
                  style={selectStyle}
                  value={draft.confidenceBand}
                  onChange={(event) => {
                    const confidenceBand = isAssessmentConfidenceBand(
                      event.target.value,
                    )
                      ? event.target.value
                      : "unset";
                    assessmentCreation.commands.updateDraft((current) => ({
                      ...current,
                      confidenceBand,
                    }));
                  }}
                >
                  {confidenceBandOptions.map((value) => (
                    <option key={value} value={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </label>

              <label style={labelStyle}>
                Rationale
                <textarea
                  data-testid={assessmentCreateControlTestId("rationale")}
                  disabled={isSubmitting}
                  rows={4}
                  style={textareaStyle}
                  value={draft.rationale}
                  onChange={(event) => {
                    assessmentCreation.commands.updateDraft((current) => ({
                      ...current,
                      rationale: event.target.value,
                    }));
                  }}
                />
              </label>

              <label style={labelStyle}>
                Assessed
                <input
                  data-testid={assessmentCreateControlTestId("assessed-at")}
                  disabled={isSubmitting}
                  placeholder="RFC3339 timestamp"
                  style={inputStyle}
                  type="text"
                  value={draft.assessedAt}
                  onChange={(event) => {
                    assessmentCreation.commands.updateDraft((current) => ({
                      ...current,
                      assessedAt: event.target.value,
                    }));
                  }}
                />
              </label>

              <WorkbookRecordCandidatePicker
                candidates={supportCandidates}
                disabled={isSubmitting}
                label="Support refs"
                selectedRecordIds={draft.supportRecordIds}
                testId={assessmentCreateControlTestId("support-refs")}
                onSelectedRecordIdsChange={(supportRecordIds) => {
                  assessmentCreation.commands.updateDraft((current) => ({
                    ...current,
                    supportRecordIds,
                  }));
                }}
              />

              <button
                data-testid={assessmentCreateControlTestId("submit")}
                disabled={!canCreate || isSubmitting}
                style={secondaryActionButtonStyle}
                type="button"
                onClick={() => {
                  void assessmentCreation.commands.submit(canCreate);
                }}
              >
                Create assessment
              </button>
              {message ? (
                <WorkbookInspectorPublicError
                  error={message}
                  testId={assessmentCreateControlTestId("message")}
                />
              ) : null}
            </div>
          </WorkbookInspectorShell>
        ) : undefined
      }
      onRequestInspectorClose={cancelAssessmentDraft}
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(assessmentsViewSchemaId)}
        >
          <SemanticDataGrid
            ref={gridHandleRef}
            activeRowIdentity={
              selectedAssessmentRecordId === null
                ? null
                : {
                    kind: "core_record",
                    recordId: selectedAssessmentRecordId,
                  }
            }
            columns={columns}
            columnWidths={layoutState.columnWidths}
            dataState={dataState}
            density={density}
            grouping={grouping}
            interactionMode={interactionMode}
            onActiveCellChange={(anchor) => {
              const recordId =
                anchor?.rowIdentity.kind === "core_record"
                  ? anchor.rowIdentity.recordId
                  : null;
              if (recordId !== null) {
                selectAssessment(recordId);
              }
              if (recordId === null || anchor === null) {
                assessmentFocus.port.clear();
              } else {
                assessmentFocus.port.select({
                  fieldKey: anchor.fieldKey,
                  recordId,
                  viewSchemaId: assessmentsViewSchemaId,
                });
              }
              collaborationProjection.publishPresence({
                fieldKey: null,
                mode: recordId === null ? "idle" : "viewing",
                recordId,
              });
            }}
            onColumnReorder={onColumnReorder}
            onColumnWidthChange={onColumnWidthChange}
            onSelectRow={(rowIdentity) => {
              if (rowIdentity.kind === "core_record") {
                selectAssessment(rowIdentity.recordId);
              }
            }}
            onSortChange={onSortChange}
            dataRows={gridRows}
            sort={queryState.sort}
            surface={{
              kind: "view_schema",
              viewSchemaId: assessmentsViewSchemaId,
            }}
          />
        </GridViewport>
      }
      statusStrip={
        <WorkbookSurfaceStatusStrip
          activeSheetPresenceRecords={collaboration.activeSheetPresenceRecords}
          mutationError={mutation.secondaryMessage}
          mutationState={mutation.primaryLabel}
          onActivateConflict={onActivateConflict}
          showPresence={showStatusPresence}
          workbookFocusAnchor={assessmentFocus.snapshot.anchor}
        />
      }
      viewBar={
        <WorkbookViewBar
          addRowDisabled={!canCreate}
          chromeMode={chromeMode}
          queryControls={queryControls}
          savedViewControls={savedViewSelector}
          onAddRow={() => {
            openStandaloneDraft();
          }}
          onInspectorToggle={() => {
            inspectorContinuityTokenRef.current =
              assessmentFocus.port.capture();
            inspector.commands.open();
          }}
          surface={assessmentsViewSchemaId}
        />
      }
      viewSchemaId={assessmentsViewSchemaId}
    />
  );
}

const bodyStyle = {
  margin: 0,
  lineHeight: 1.5,
  color: "var(--ct-colors-ink-muted)",
};

const gridShellStyle = {
  ...workbookSurfaceGridShellStyle,
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

const textareaStyle = {
  ...inputStyle,
  resize: "vertical" as const,
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

const selectStyle = {
  ...inputStyle,
  appearance: "auto" as const,
};
