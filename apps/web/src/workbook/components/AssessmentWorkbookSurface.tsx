import {
  type GridColumn,
  type GridDataRow,
  type GridGroupingDescriptor,
  type GridHandle,
  GridViewport,
  SemanticDataGrid,
} from "@cartulary/grid-adapter";
import {
  gridGroupRowTestId,
  gridShellTestId,
  gridSortHeaderTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  resolveHeaderSortFieldKey,
} from "@cartulary/view-contracts";
import { type ReactNode, useEffect, useMemo, useRef, useState } from "react";
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
import { useAssessmentWorkbookInspectorComposition } from "../features/assessments/useAssessmentWorkbookInspectorComposition";
import { useWorkbookSemanticGridFocus } from "../hooks/useWorkbookSemanticGridFocus";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import {
  WorkbookSurfaceLayout,
  workbookSurfaceGridShellStyle,
} from "../layout/WorkbookSurfaceLayout";
import { applyWorkbookLayoutToColumns } from "../layout/workbookColumnLayout";
import { assessmentColumnWidth } from "../models/assessmentWorkbookModel";
import type { EntityRow } from "../models/entityWorkbookModel";
import { genericCellLabel } from "../models/genericWorkbookModel";
import { workbookGridRows } from "../models/workbookContractRows";
import type { WorkbookGridEntryFocusOwner } from "../models/workbookGridEntryFocus";
import {
  type WorkbookQueryLoadState,
  workbookGridDataState,
} from "../models/workbookGridState";
import type { WorkbookQueryState } from "../models/workbookQuery";
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
  gridEntryFocus: WorkbookGridEntryFocusOwner;
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
  gridEntryFocus,
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
  const assessmentInspector = useAssessmentWorkbookInspectorComposition({
    canCreate,
    contract: assessmentsContract,
    currentIncidentRole,
    currentUserId,
    hostRows,
    identityRows,
    incidentClosed,
    inspectorResetKey,
    interactionMode,
    mutationCommands,
    mutationRuntime,
    onCaptureFocus: () => {
      inspectorContinuityTokenRef.current = assessmentFocus.port.capture();
    },
    onClearSelectedAssessment: () => {
      setSelectedAssessmentRecordId(null);
      setSelectedAssessmentSnapshot(null);
    },
    onClearSurfaceSelection: () => {
      continuityPortRef.current?.clear();
      setSelectedAssessmentRecordId(null);
      setSelectedAssessmentSnapshot(null);
    },
    onRefreshAssessmentRows,
    onRestoreFocus: () => {
      const token = inspectorContinuityTokenRef.current;
      inspectorContinuityTokenRef.current = null;
      if (token !== null) continuityPortRef.current?.restore(token);
    },
    onSelectAssessment: selectAssessment,
    recordMutationCommands,
    relatedMutationCommands,
    roleCanCreate,
    selectedAssessment,
    viewQuery,
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
  const registerGridHandle = useWorkbookSemanticGridFocus({
    dataRows: gridRows,
    dataState,
    focusOwner: gridEntryFocus,
    gridHandleRef,
    visibleColumns: columns,
    viewSchemaId: assessmentsViewSchemaId,
  });

  function openStandaloneDraft() {
    inspectorContinuityTokenRef.current = assessmentFocus.port.capture();
    assessmentInspector.openStandalone(hostRows[0]?.recordId ?? "");
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
  return (
    <WorkbookSurfaceLayout
      chromeMode={chromeMode}
      inspector={assessmentInspector.node}
      onRequestInspectorClose={assessmentInspector.close}
      primaryGrid={
        <GridViewport
          blockSizing="fill"
          style={gridShellStyle}
          testId={gridShellTestId(assessmentsViewSchemaId)}
        >
          <SemanticDataGrid
            ref={registerGridHandle}
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
            assessmentInspector.open();
          }}
          surface={assessmentsViewSchemaId}
        />
      }
      viewSchemaId={assessmentsViewSchemaId}
    />
  );
}

const gridShellStyle = {
  ...workbookSurfaceGridShellStyle,
};
