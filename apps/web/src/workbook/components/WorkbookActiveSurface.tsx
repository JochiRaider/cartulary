import type { GridDensity, GridInteractionMode } from "@cartulary/grid-adapter";
import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type ReactNode,
  type SetStateAction,
  useState,
} from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { AssessmentApiRow } from "../models/assessmentWorkbookModel";
import type { EntityRow } from "../models/entityWorkbookModel";
import type { WorkbookQueryLoadState } from "../models/workbookGridState";
import type { WorkbookResolvedLayoutState } from "../models/workbookLayout";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  replaceWorkbookSort,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import type { WorkbookChromeMode } from "../models/workbookResponsiveLayout";
import type { WorkbookSheetRef } from "../models/workbookStartup";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import type { WorkbookCollaborationProjection } from "../runtime/WorkbookCollaborationProjection";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { TimelineWorkbook } from "../timeline/components/TimelineWorkbook";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import { AssessmentWorkbookSurface } from "./AssessmentWorkbookSurface";
import { EntityWorkbookSurface } from "./EntityWorkbookSurface";
import { ContractWorkbookSurface } from "./GenericWorkbookSurface";

export type WorkbookActiveSurfaceProps = {
  readonly activeContract: ViewContract;
  readonly apiBase?: string | undefined;
  readonly assessmentLoadState: WorkbookQueryLoadState;
  readonly assessmentQueryState: WorkbookQueryState;
  readonly assessmentRows: AssessmentApiRow[];
  readonly authorizationEpoch: string;
  readonly chromeMode: WorkbookChromeMode;
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly density: GridDensity;
  readonly entityIndex: Record<string, EntityRow>;
  readonly genericLoadState: WorkbookQueryLoadState;
  readonly genericQueryState: WorkbookQueryState;
  readonly genericRows: EntityApiRow[];
  readonly hostQueryState: WorkbookQueryState;
  readonly hostRows: EntityRow[];
  readonly identityQueryState: WorkbookQueryState;
  readonly identityRows: EntityRow[];
  readonly incidentId: string;
  readonly interactionMode: GridInteractionMode;
  readonly inspectorResetKey: string;
  readonly layoutState: WorkbookResolvedLayoutState;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly queryControls?: ReactNode | undefined;
  readonly collaborationProjection: WorkbookCollaborationProjection;
  readonly loadAssessmentSurface: () => Promise<void>;
  readonly loadEntities: () => Promise<void>;
  readonly loadGenericSurface: () => Promise<void>;
  readonly entityLoadState: WorkbookQueryLoadState;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly onColumnReorder: (
    sourceFieldKey: string,
    targetFieldKey: string,
  ) => void;
  readonly onColumnHiddenChange: (fieldKey: string, hidden: boolean) => void;
  readonly onColumnMove: (
    fieldKey: string,
    direction: "earlier" | "later",
  ) => void;
  readonly onResetColumns: () => void;
  readonly onColumnWidthChange: (fieldKey: string, width: number) => void;
  readonly savedViewSelector?: ReactNode | undefined;
  readonly showStatusPresence: boolean;
  readonly setAssessmentQueryState: Dispatch<
    SetStateAction<WorkbookQueryState>
  >;
  readonly setGenericQueryState: Dispatch<SetStateAction<WorkbookQueryState>>;
  readonly setHostQueryState: Dispatch<SetStateAction<WorkbookQueryState>>;
  readonly setIdentityQueryState: Dispatch<SetStateAction<WorkbookQueryState>>;
  readonly setTimelineQueryState: Dispatch<SetStateAction<WorkbookQueryState>>;
  readonly sheetRef: WorkbookSheetRef;
  readonly sheetReloadToken: number;
  readonly surface: string;
  readonly timelineQueryState: WorkbookQueryState;
};

export function WorkbookActiveSurface({
  activeContract,
  apiBase,
  assessmentLoadState,
  assessmentQueryState,
  assessmentRows,
  authorizationEpoch,
  chromeMode,
  currentIncidentRole,
  currentUserId,
  density,
  entityIndex,
  genericLoadState,
  genericQueryState,
  genericRows,
  hostQueryState,
  hostRows,
  identityQueryState,
  identityRows,
  incidentId,
  interactionMode,
  inspectorResetKey,
  layoutState,
  mutationRuntime,
  queryControls,
  collaborationProjection,
  loadAssessmentSurface,
  loadEntities,
  loadGenericSurface,
  entityLoadState,
  onIncidentAccessLost,
  onColumnHiddenChange,
  onColumnMove,
  onColumnReorder,
  onColumnWidthChange,
  onResetColumns,
  savedViewSelector,
  showStatusPresence,
  setAssessmentQueryState,
  setGenericQueryState,
  setHostQueryState,
  setIdentityQueryState,
  setTimelineQueryState,
  sheetRef,
  sheetReloadToken,
  surface,
  timelineQueryState,
}: WorkbookActiveSurfaceProps) {
  const registration = requireWorkbookSurfaceRegistration(surface);
  const [timelineFilterDraft, setTimelineFilterDraft] = useState(() =>
    defaultFilterDraft(
      requireWorkbookSurfaceRegistration("cartulary.view.timeline.v2").contract,
    ),
  );
  if (registration.renderer === "timeline") {
    return (
      <TimelineWorkbook
        runtime={{
          attachCollaborationSession: false,
          collaborationProjection,
          mutationRuntime,
          incident: {
            id: incidentId,
            apiBase,
            currentUserId,
            currentRole: currentIncidentRole,
            sheetRef,
            inspectorResetKey,
            reloadToken: sheetReloadToken,
          },
          query: {
            state: timelineQueryState,
            setState: setTimelineQueryState,
            filterDraft: timelineFilterDraft,
            setFilterDraft: setTimelineFilterDraft,
            renderInlineControls: false,
            savedViewSelector,
            viewBarQueryControls: queryControls,
          },
          entities: {
            hosts: hostRows,
            identities: identityRows,
            index: entityIndex,
            refresh: loadEntities,
          },
          layout: {
            chromeMode,
            density,
            interactionMode,
            state: layoutState,
            setColumnHidden: onColumnHiddenChange,
            moveColumn: onColumnMove,
            reorderColumn: onColumnReorder,
            setColumnWidth: onColumnWidthChange,
            resetColumns: onResetColumns,
            showStatusPresence,
          },
          onIncidentAccessLost,
        }}
      />
    );
  }

  if (
    registration.renderer === "entity_hosts" ||
    registration.renderer === "entity_identities"
  ) {
    const isHosts = registration.renderer === "entity_hosts";
    return (
      <EntityWorkbookSurface
        apiBase={apiBase}
        chromeMode={chromeMode}
        currentIncidentRole={currentIncidentRole}
        density={density}
        entityIndex={entityIndex}
        entityType={isHosts ? "host" : "identity"}
        incidentId={incidentId}
        interactionMode={interactionMode}
        inspectorResetKey={inspectorResetKey}
        layoutState={layoutState}
        mutationRuntime={mutationRuntime}
        collaborationProjection={collaborationProjection}
        loadState={entityLoadState}
        onRefreshEntities={loadEntities}
        onColumnReorder={onColumnReorder}
        onColumnWidthChange={onColumnWidthChange}
        onClearFilters={() => {
          if (isHosts) setHostQueryState(emptyWorkbookQueryState());
          else setIdentityQueryState(emptyWorkbookQueryState());
        }}
        onSortChange={(sort) => {
          if (isHosts) {
            setHostQueryState((current) =>
              replaceWorkbookSort(registration.contract, current, sort),
            );
            return;
          }
          setIdentityQueryState((current) =>
            replaceWorkbookSort(registration.contract, current, sort),
          );
        }}
        queryState={isHosts ? hostQueryState : identityQueryState}
        queryControls={queryControls}
        rows={isHosts ? hostRows : identityRows}
        savedViewSelector={savedViewSelector}
        showStatusPresence={showStatusPresence}
      />
    );
  }

  if (registration.renderer === "assessment") {
    return (
      <AssessmentWorkbookSurface
        apiBase={apiBase}
        chromeMode={chromeMode}
        assessmentRows={assessmentRows}
        currentIncidentRole={currentIncidentRole}
        density={density}
        hostRows={hostRows}
        identityRows={identityRows}
        incidentId={incidentId}
        inspectorResetKey={inspectorResetKey}
        layoutState={layoutState}
        mutationRuntime={mutationRuntime}
        collaborationProjection={collaborationProjection}
        loadState={assessmentLoadState}
        interactionMode={interactionMode}
        onRefreshAssessmentRows={loadAssessmentSurface}
        onColumnReorder={onColumnReorder}
        onColumnWidthChange={onColumnWidthChange}
        onClearFilters={() => {
          setAssessmentQueryState(emptyWorkbookQueryState());
        }}
        onSortChange={(sort) => {
          setAssessmentQueryState((current) =>
            replaceWorkbookSort(registration.contract, current, sort),
          );
        }}
        queryState={assessmentQueryState}
        queryControls={queryControls}
        savedViewSelector={savedViewSelector}
        showStatusPresence={showStatusPresence}
      />
    );
  }

  return (
    <ContractWorkbookSurface
      key={activeContract.viewSchemaId}
      apiBase={apiBase}
      authorizationEpoch={authorizationEpoch}
      chromeMode={chromeMode}
      contract={activeContract}
      currentUserId={currentUserId}
      density={density}
      incidentId={incidentId}
      inspectorResetKey={inspectorResetKey}
      loadState={genericLoadState}
      interactionMode={interactionMode}
      layoutState={layoutState}
      mutationRuntime={mutationRuntime}
      collaborationProjection={collaborationProjection}
      sheetRef={sheetRef}
      onColumnReorder={onColumnReorder}
      onColumnWidthChange={onColumnWidthChange}
      onClearFilters={() => {
        setGenericQueryState(emptyWorkbookQueryState());
      }}
      onRefresh={loadGenericSurface}
      onSortChange={(sort) => {
        setGenericQueryState((current) =>
          replaceWorkbookSort(activeContract, current, sort),
        );
      }}
      queryState={genericQueryState}
      queryControls={queryControls}
      rows={genericRows}
      savedViewSelector={savedViewSelector}
      showStatusPresence={showStatusPresence}
    />
  );
}
