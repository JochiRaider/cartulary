import {
  type GridColumn,
  type GridDensity,
  type GridRow,
  GridTable,
  GridViewport,
} from "@cartulary/grid-adapter";
import {
  assessmentCreatePanelTestId,
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
  workbookInspectorCloseButtonTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  resolveHeaderSortFieldKey,
  visibleFields,
} from "@cartulary/view-contracts";
import { X } from "lucide-react";
import { type ReactNode, useEffect, useMemo, useState } from "react";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
} from "../../services/workbookApi";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import { useAssessmentSupportRows } from "../hooks/useAssessmentSupportRows";
import {
  assessmentColumnWidth,
  initialAssessmentDraft,
  isAssessmentConfidenceBand,
  supportRowLabel,
} from "../models/assessmentWorkbookModel";
import type { EntityRow } from "../models/entityWorkbookModel";
import {
  enumValuesFor,
  genericCellLabel,
} from "../models/genericWorkbookModel";
import { workbookGridRows } from "../models/workbookContractRows";
import {
  inspectorPanelIsDeclared,
  selectInspectorConfig,
} from "../models/workbookInspectorModel";
import type { ViewMutationEnvelope } from "../models/workbookMutations";
import type { WorkbookQueryState } from "../models/workbookQuery";
import { assessmentsViewSchemaId } from "../models/workbookSurfaceRegistry";
import {
  type AssessmentCreateDraft,
  buildAssessmentCreatePayload,
  type EntityApiRow,
} from "../timeline/models/workbookTimelineModel";
import {
  FocusableWorkbookCell,
  useWorkbookGridFocus,
} from "../utils/workbookGridFocus";
import {
  type InspectorDisabledToken,
  WorkbookInspectorPanelSection,
} from "./WorkbookInspectorFeatureGroups";
import { WorkbookSheetToolbar } from "./WorkbookSheetToolbar";
import { WorkbookSurfaceStatusStrip } from "./WorkbookStatusStrip";
import {
  WorkbookSurfaceFrame,
  workbookSurfaceGridShellStyle,
  workbookSurfaceInspectorPanelStyle,
  workbookSurfaceOverlayPanelStyle,
} from "./WorkbookSurfaceFrame";

const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

export type AssessmentWorkbookSurfaceProps = {
  apiBase?: string | undefined;
  assessmentRows: EntityApiRow[];
  currentIncidentRole: WorkbookIncidentRole | null;
  density: GridDensity;
  inspectorResetKey: string;
  savedViewSelector?: ReactNode | undefined;
  hostRows: EntityRow[];
  identityRows: EntityRow[];
  incidentId: string;
  loadError: string | null;
  onRefreshAssessmentRows: () => Promise<void>;
  onToggleSort: (fieldKey: string) => void;
  queryState: WorkbookQueryState;
};

export function AssessmentWorkbookSurface({
  apiBase,
  assessmentRows,
  currentIncidentRole,
  density,
  inspectorResetKey,
  savedViewSelector,
  hostRows,
  identityRows,
  incidentId,
  loadError,
  onRefreshAssessmentRows,
  onToggleSort,
  queryState,
}: AssessmentWorkbookSurfaceProps) {
  const [draft, setDraft] = useState<AssessmentCreateDraft>(() =>
    initialAssessmentDraft(assessmentsContract),
  );
  const [isInspectorOpen, setIsInspectorOpen] = useState(false);
  const inspectorConfig = selectInspectorConfig(assessmentsContract);
  const showWorkflowPanel = inspectorPanelIsDeclared(
    inspectorConfig,
    "workflow",
  );
  const supportRows = useAssessmentSupportRows({ apiBase, incidentId });
  const [message, setMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const assessmentInspectorDisabledTokens = useMemo(
    () => new Set<InspectorDisabledToken>(),
    [],
  );
  const subjectRows = draft.subjectType === "host" ? hostRows : identityRows;
  const canCreate =
    currentIncidentRole === "editor" ||
    currentIncidentRole === "reviewer" ||
    currentIncidentRole === "admin";
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
  const anchorColumns: readonly GridColumn<EntityApiRow>[] = visibleFields(
    assessmentsContract,
  ).map((field) => ({
    fieldKey: field.fieldKey,
    headerTestId: gridSortHeaderTestId(assessmentsViewSchemaId, field.fieldKey),
    label: field.label,
    width: assessmentColumnWidth(field.fieldKey),
    renderCell: () => null,
    sortableFieldKey: resolveHeaderSortFieldKey(
      assessmentsContract,
      field.fieldKey,
    ),
  }));
  const gridRows: readonly GridRow<EntityApiRow>[] = workbookGridRows({
    getRecordId: (row) => row.record_id,
    rows: assessmentRows,
    surface: assessmentsViewSchemaId,
  });
  const assessmentFocus = useWorkbookGridFocus({
    columns: anchorColumns,
    getGroupLabel: (row, fieldKey) =>
      genericCellLabel(row.cells[fieldKey]?.value),
    groupBy: queryState.groupBy,
    rows: gridRows,
    surface: assessmentsViewSchemaId,
  });
  const columns: readonly GridColumn<EntityApiRow>[] = anchorColumns.map(
    (field) => ({
      ...field,
      renderCell: (row) => (
        <FocusableWorkbookCell
          fieldKey={field.fieldKey}
          focus={assessmentFocus}
          recordId={row.record_id}
        >
          {genericCellLabel(row.cells[field.fieldKey]?.value)}
        </FocusableWorkbookCell>
      ),
    }),
  );

  useEffect(() => {
    if (inspectorResetKey === "") {
      return;
    }
    setIsInspectorOpen(false);
    setDraft(initialAssessmentDraft(assessmentsContract));
    setMessage(null);
  }, [inspectorResetKey]);

  useEffect(() => {
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
  }, [subjectRows]);

  async function submitAssessment() {
    const payload = buildAssessmentCreatePayload(
      draft,
      `assessment-${Date.now()}`,
    );
    if (payload === null) {
      setMessage("Subject, state, and rationale are required.");
      return;
    }

    setIsSubmitting(true);
    setMessage(null);
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
      setDraft((current) => ({
        ...initialAssessmentDraft(assessmentsContract),
        subjectType: current.subjectType,
        subjectRecordId: current.subjectRecordId,
      }));
      setMessage("Assessment created.");
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <WorkbookSurfaceFrame
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
                  <h2 style={inspectorTitleStyle}>Append assessment</h2>
                </div>
                <button
                  aria-label="Close inspector"
                  data-testid={workbookInspectorCloseButtonTestId(
                    assessmentsViewSchemaId,
                  )}
                  style={inspectorCloseButtonStyle}
                  type="button"
                  onClick={() => {
                    setIsInspectorOpen(false);
                  }}
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
              />
            ))}
            <div style={inspectorSectionStyle}>
              <label style={labelStyle}>
                Subject type
                <select
                  data-testid="assessment-create-subject-type"
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
                  data-testid="assessment-create-subject"
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
                  data-testid="assessment-create-state"
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
                  data-testid="assessment-create-confidence-band"
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
                  data-testid="assessment-create-rationale"
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
                  data-testid="assessment-create-assessed-at"
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

              <label style={labelStyle}>
                Support refs
                <select
                  data-testid="assessment-create-support-refs"
                  multiple
                  size={Math.min(Math.max(supportRows.length, 2), 5)}
                  style={selectStyle}
                  value={draft.supportRecordIds}
                  onChange={(event) => {
                    const supportRecordIds = Array.from(
                      event.currentTarget.selectedOptions,
                    ).map((option) => option.value);
                    setDraft((current) => ({
                      ...current,
                      supportRecordIds,
                    }));
                  }}
                >
                  {supportRows.map((row) => (
                    <option key={row.record_id} value={row.record_id}>
                      {supportRowLabel(row)}
                    </option>
                  ))}
                </select>
              </label>

              <button
                data-testid="assessment-create-submit"
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
                <p data-testid="assessment-create-message" style={bodyStyle}>
                  {message}
                </p>
              ) : null}
            </div>
          </aside>
        ) : undefined
      }
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(assessmentsViewSchemaId)}
        >
          <GridTable
            columns={columns}
            density={density}
            getGroupLabel={(row, fieldKey) =>
              genericCellLabel(row.cells[fieldKey]?.value)
            }
            getGroupRowTestId={(fieldKey, value) =>
              gridGroupRowTestId(assessmentsViewSchemaId, fieldKey, value)
            }
            groupBy={queryState.groupBy}
            onToggleSort={onToggleSort}
            rows={gridRows}
            sort={queryState.sort}
          />
        </GridViewport>
      }
      statusStrip={
        <WorkbookSurfaceStatusStrip
          mutationState="Saved"
          workbookFocusAnchor={assessmentFocus.anchor}
        />
      }
      viewBar={
        <WorkbookSheetToolbar
          addRowDisabled={!canCreate}
          leading={savedViewSelector}
          onAddRow={() => {
            setIsInspectorOpen(true);
          }}
          onInspectorToggle={() => {
            setIsInspectorOpen(true);
          }}
          surface={assessmentsViewSchemaId}
        />
      }
      viewSchemaId={assessmentsViewSchemaId}
      workAreaOverlays={
        loadError ? (
          <p
            data-testid="assessment-surface-load-error"
            style={surfaceNoticeOverlayStyle}
          >
            {loadError}
          </p>
        ) : undefined
      }
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

const surfaceNoticeOverlayStyle = {
  ...workbookSurfaceOverlayPanelStyle,
  margin: 0,
  padding: "0.85rem 1rem",
  borderRadius: "var(--ct-rounded-sm)",
  border: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-2)",
  color: "var(--ct-colors-ink-muted)",
  boxShadow: "var(--ct-elevation-popover)",
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
