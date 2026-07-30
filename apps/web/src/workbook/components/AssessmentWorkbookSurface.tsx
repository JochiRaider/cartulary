import {
  type GridColumn,
  type GridDataRow,
  type GridDensity,
  type GridGroupingDescriptor,
  type GridHandle,
  type GridInteractionMode,
  GridViewport,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  assessmentCreateControlTestId,
  assessmentCreatePanelTestId,
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  type InspectorFeatureGroup,
  requireViewContract,
  resolveHeaderSortFieldKey,
} from "@cartulary/view-contracts";
import { X } from "lucide-react";
import { type ReactNode, useEffect, useMemo, useRef, useState } from "react";
import { apiPath, clientTxnID } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
} from "../../services/workbookApi";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { useAssessmentSupportCandidates } from "../hooks/useAssessmentSupportCandidates";
import { useInspectorLifecycleReset } from "../hooks/useInspectorLifecycleReset";
import {
  type AssessmentApiRow,
  type AssessmentCreateDraft,
  assessmentColumnWidth,
  buildAssessmentCreatePayload,
  followOnAssessmentDraft,
  initialAssessmentDraft,
  isAssessmentConfidenceBand,
} from "../models/assessmentWorkbookModel";
import type { EntityRow } from "../models/entityWorkbookModel";
import {
  enumValuesFor,
  genericCellLabel,
} from "../models/genericWorkbookModel";
import { workbookGridRows } from "../models/workbookContractRows";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../models/workbookGridState";
import {
  inspectorPanelIsDeclared,
  selectInspectorConfig,
} from "../models/workbookInspectorModel";
import {
  applyWorkbookLayoutToColumns,
  type WorkbookResolvedLayoutState,
} from "../models/workbookLayout";
import type { ViewMutationEnvelope } from "../models/workbookMutations";
import type { WorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookChromeMode } from "../models/workbookResponsiveLayout";
import { assessmentsViewSchemaId } from "../models/workbookSurfaceRegistry";
import { useWorkbookCollaborationProjection } from "../runtime/useWorkbookCollaborationProjection";
import { useWorkbookMutationRuntime } from "../runtime/useWorkbookMutationRuntime";
import type { WorkbookCollaborationProjection } from "../runtime/WorkbookCollaborationProjection";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import {
  FocusableWorkbookCell,
  useWorkbookGridFocus,
} from "../utils/workbookGridFocus";
import {
  type InspectorDisabledToken,
  WorkbookInspectorPanelSection,
} from "./WorkbookInspectorFeatureGroups";
import { WorkbookCellPresenceMarker } from "./WorkbookPresenceMarkers";
import { WorkbookRecordCandidatePicker } from "./WorkbookRecordCandidatePicker";
import { WorkbookSurfaceStatusStrip } from "./WorkbookStatusStrip";
import {
  WorkbookSurfaceFrame,
  workbookSurfaceGridShellStyle,
  workbookSurfaceInspectorPanelStyle,
} from "./WorkbookSurfaceFrame";
import { WorkbookViewBar } from "./WorkbookViewBar";

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

export type AssessmentWorkbookSurfaceProps = {
  chromeMode: WorkbookChromeMode;
  apiBase?: string | undefined;
  assessmentRows: AssessmentApiRow[];
  currentIncidentRole: WorkbookIncidentRole | null;
  density: GridDensity;
  inspectorResetKey: string;
  queryControls?: ReactNode | undefined;
  savedViewSelector?: ReactNode | undefined;
  showStatusPresence: boolean;
  hostRows: EntityRow[];
  identityRows: EntityRow[];
  incidentId: string;
  layoutState: WorkbookResolvedLayoutState;
  loadState: WorkbookQueryLoadState;
  mutationRuntime: WorkbookMutationRuntime;
  collaborationProjection: WorkbookCollaborationProjection;
  interactionMode: GridInteractionMode;
  onClearFilters: () => void;
  onRefreshAssessmentRows: () => Promise<void>;
  onColumnReorder: (sourceFieldKey: string, targetFieldKey: string) => void;
  onColumnWidthChange: (fieldKey: string, width: number) => void;
  onSortChange: (sort: WorkbookQueryState["sort"]) => void;
  queryState: WorkbookQueryState;
};

export function AssessmentWorkbookSurface({
  chromeMode,
  apiBase,
  assessmentRows,
  currentIncidentRole,
  density,
  inspectorResetKey,
  queryControls,
  savedViewSelector,
  showStatusPresence,
  hostRows,
  identityRows,
  incidentId,
  layoutState,
  loadState,
  mutationRuntime,
  collaborationProjection,
  interactionMode,
  onClearFilters,
  onRefreshAssessmentRows,
  onColumnReorder,
  onColumnWidthChange,
  onSortChange,
  queryState,
}: AssessmentWorkbookSurfaceProps) {
  const [selectedAssessmentRecordId, setSelectedAssessmentRecordId] = useState<
    string | null
  >(null);
  const [selectedAssessmentSnapshot, setSelectedAssessmentSnapshot] =
    useState<AssessmentApiRow | null>(null);
  const mutation = useWorkbookMutationRuntime(mutationRuntime);
  const collaboration = useWorkbookCollaborationProjection(
    collaborationProjection,
  );
  const [draft, setDraft] = useState<AssessmentCreateDraft>(() =>
    initialAssessmentDraft(assessmentsContract),
  );
  const [draftMode, setDraftMode] = useState<"follow_on" | "standalone">(
    "standalone",
  );
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const inspectorConfig = selectInspectorConfig(assessmentsContract);
  const showWorkflowPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "workflow",
  );
  const supportCandidates = useAssessmentSupportCandidates({
    apiBase,
    incidentId,
  });
  const [message, setMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const subjectRows = draft.subjectType === "host" ? hostRows : identityRows;
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
  const assessmentInspectorDisabledTokens = useMemo(() => {
    const tokens = new Set<InspectorDisabledToken>();
    if (selectedAssessment === null) {
      tokens.add("no_row_selected");
    }
    if (!roleCanCreate) {
      tokens.add("authorization_lost");
    } else if (interactionMode.kind === "read_only") {
      tokens.add("incident_closed");
    }
    return tokens;
  }, [interactionMode.kind, roleCanCreate, selectedAssessment]);
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
  const anchorColumns = useMemo<readonly GridColumn<AssessmentApiRow>[]>(
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
  const gridRows = useMemo<readonly GridDataRow<AssessmentApiRow>[]>(
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
    useMemo<GridGroupingDescriptor<AssessmentApiRow> | null>(() => {
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
  const assessmentFocus = useWorkbookGridFocus({
    columns: anchorColumns,
    gridHandleRef,
    surface: assessmentsViewSchemaId,
  });
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
  const columns: readonly GridColumn<AssessmentApiRow>[] = anchorColumns.map(
    (field) => ({
      ...field,
      getClipboardValue: (row) =>
        genericCellLabel(row.cells[field.fieldKey]?.value),
      renderCell: ({ row }) => (
        <FocusableWorkbookCell
          fieldKey={field.fieldKey}
          focus={assessmentFocus}
          recordId={row.record_id}
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
        </FocusableWorkbookCell>
      ),
    }),
  );

  function standaloneAssessmentDraft(): AssessmentCreateDraft {
    return initialAssessmentDraft(assessmentsContract, {
      subjectRecordId: hostRows[0]?.recordId ?? "",
      subjectType: "host",
    });
  }

  function openStandaloneDraft() {
    setDraft(standaloneAssessmentDraft());
    setDraftMode("standalone");
    setMessage(null);
    setIsInspectorOpen(true);
  }

  function cancelAssessmentDraft() {
    setDraft(initialAssessmentDraft(assessmentsContract));
    setDraftMode("standalone");
    setMessage(null);
    setIsInspectorOpen(false);
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

  function beginAssessmentFeatureAction(featureGroup: InspectorFeatureGroup) {
    if (featureGroup.featureGroupKey !== "create_related.assessment") {
      return;
    }
    if (
      featureGroup.routeBinding.kind !== "view_row_create" ||
      featureGroup.routeBinding.owner !== "view_row_create_route" ||
      featureGroup.routeBinding.targetViewSchemaId !== assessmentsViewSchemaId
    ) {
      setMessage("Assessment follow-on creation is unavailable.");
      return;
    }
    if (!canCreate) {
      setMessage("Assessment creation requires an active editor role.");
      return;
    }
    if (selectedAssessment === null) {
      setMessage("Select an assessment before creating a follow-on.");
      return;
    }
    const followOnDraft = followOnAssessmentDraft(
      assessmentsContract,
      selectedAssessment,
    );
    if (followOnDraft === null) {
      setMessage("The selected assessment has no valid subject.");
      return;
    }
    setDraft(followOnDraft);
    setDraftMode("follow_on");
    setMessage(null);
    setIsInspectorOpen(true);
  }

  useInspectorLifecycleReset(inspectorResetKey, () => {
    setIsInspectorOpen(false);
    setDraft(initialAssessmentDraft(assessmentsContract));
    setDraftMode("standalone");
    setMessage(null);
  });

  useEffect(() => {
    if (draftMode === "follow_on") {
      return;
    }
    setDraft((current) => {
      if (
        current.subjectRecordId !== "" &&
        subjectRows.some((row) => row.recordId === current.subjectRecordId)
      ) {
        return current;
      }
      return {
        ...current,
        subjectRecordId: subjectRows[0]?.recordId ?? "",
      };
    });
  }, [draftMode, subjectRows]);

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

  async function submitAssessment() {
    if (!canCreate) return;
    const submittedDraft = draft;
    const payload = buildAssessmentCreatePayload(
      submittedDraft,
      clientTxnID("assessment"),
    );
    if (payload === null) {
      setMessage("Subject, state, and rationale are required.");
      return;
    }

    setIsSubmitting(true);
    setMessage(null);
    const finishMutation = mutationRuntime.beginExplicitMutation();
    try {
      const result = await fetchWorkbookJSON<ViewMutationEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/rows`,
        ),
        {
          method: "POST",
          body: JSON.stringify(payload),
        },
      );
      if (!result.ok) {
        setMessage(parseErrorMessage(result.payload));
        return;
      }
      await onRefreshAssessmentRows();
      setDraft(
        initialAssessmentDraft(assessmentsContract, {
          subjectType: submittedDraft.subjectType,
          subjectRecordId: submittedDraft.subjectRecordId,
        }),
      );
      setMessage("Assessment created.");
    } finally {
      finishMutation();
      setIsSubmitting(false);
    }
  }

  return (
    <WorkbookSurfaceFrame
      chromeMode={chromeMode}
      inspector={
        isInspectorOpen && showWorkflowPanel ? (
          <aside
            data-testid={assessmentCreatePanelTestId()}
            style={inspectorShellStyle}
          >
            <div style={inspectorHeaderStyle}>
              <div style={inspectorTitleRowStyle}>
                <div>
                  <p style={eyebrowStyle}>Create</p>
                  <h2 style={inspectorTitleStyle}>
                    {draftMode === "follow_on"
                      ? "Append follow-on assessment"
                      : "Append assessment"}
                  </h2>
                </div>
                <button
                  aria-label="Close inspector"
                  data-testid={workbookInspectorCloseButtonTestId(
                    assessmentsViewSchemaId,
                  )}
                  style={inspectorCloseButtonStyle}
                  type="button"
                  onClick={cancelAssessmentDraft}
                >
                  <X aria-hidden="true" size={16} />
                </button>
              </div>
            </div>
            {inspectorConfig.panels.map((panel) => (
              <WorkbookInspectorPanelSection
                config={inspectorConfig}
                disabledTokens={assessmentInspectorDisabledTokens}
                key={panel.panelId}
                panelId={panel.panelId}
                onFeatureAction={beginAssessmentFeatureAction}
              >
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
                    setDraft((current) => ({
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
                    setDraft((current) => ({
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
                    setDraft((current) => ({
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
                    setDraft((current) => ({
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
                    setDraft((current) => ({
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
                    setDraft((current) => ({
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
                  setDraft((current) => ({
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
                  void submitAssessment();
                }}
              >
                Create assessment
              </button>
              {message ? (
                <p
                  data-testid={assessmentCreateControlTestId("message")}
                  style={bodyStyle}
                >
                  {message}
                </p>
              ) : null}
            </div>
          </aside>
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
              assessmentFocus.update(recordId, anchor?.fieldKey ?? "");
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
          showPresence={showStatusPresence}
          workbookFocusAnchor={assessmentFocus.anchor}
        />
      }
      viewBar={
        <WorkbookViewBar
          addRowDisabled={!canCreate}
          queryControls={queryControls}
          savedViewControls={savedViewSelector}
          onAddRow={() => {
            openStandaloneDraft();
          }}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
          }}
          surface={assessmentsViewSchemaId}
        />
      }
      viewSchemaId={assessmentsViewSchemaId}
    />
  );
}

const eyebrowStyle = {
  margin: 0,
  fontSize: "0.78rem",
  letterSpacing: "0.12em",
  textTransform: "uppercase" as const,
  color: "var(--ct-colors-accent)",
};

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

const inspectorShellStyle = {
  ...workbookSurfaceInspectorPanelStyle,
};

const inspectorHeaderStyle = {
  display: "grid",
  gap: "0.35rem",
  marginBottom: "1rem",
};

const inspectorTitleRowStyle = {
  display: "flex",
  alignItems: "start",
  justifyContent: "space-between",
  gap: "1rem",
};

const inspectorCloseButtonStyle = {
  display: "inline-flex",
  alignItems: "center",
  justifyContent: "center",
  width: "1.9rem",
  height: "1.9rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "transparent",
  color: "var(--ct-colors-ink-muted)",
  cursor: "pointer",
};

const inspectorTitleStyle = {
  margin: 0,
  fontSize: "1.25rem",
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
