import type { GridDensity } from "@cartulary/grid-adapter";
import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import type { Dispatch, ReactNode, SetStateAction } from "react";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { EntityRow } from "../models/entityWorkbookModel";
import {
  toggleSortField,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import type { WorkbookSheetRef } from "../models/workbookStartup";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { TimelineWorkbook } from "../timeline/components/TimelineWorkbook";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import { AssessmentWorkbookSurface } from "./AssessmentWorkbookSurface";
import { EntityWorkbookSurface } from "./EntityWorkbookSurface";
import { GenericWorkbookSurface } from "./GenericWorkbookSurface";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

export type WorkbookActiveSurfaceProps = {
  readonly activeContract: ViewContract;
  readonly apiBase?: string | undefined;
  readonly assessmentLoadError: string | null;
  readonly assessmentQueryState: WorkbookQueryState;
  readonly assessmentRows: EntityApiRow[];
  readonly currentIncidentRole: WorkbookIncidentRole | null;
  readonly currentUserId: string | null;
  readonly density: GridDensity;
  readonly entityIndex: Record<string, EntityRow>;
  readonly genericLoadError: string | null;
  readonly genericQueryState: WorkbookQueryState;
  readonly genericRows: EntityApiRow[];
  readonly hostQueryState: WorkbookQueryState;
  readonly hostRows: EntityRow[];
  readonly identityQueryState: WorkbookQueryState;
  readonly identityRows: EntityRow[];
  readonly incidentId: string;
  readonly inspectorResetKey: string;
  readonly loadAssessmentSurface: () => Promise<void>;
  readonly loadEntities: () => Promise<void>;
  readonly loadGenericSurface: () => Promise<void>;
  readonly savedViewSelector?: ReactNode | undefined;
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
  assessmentLoadError,
  assessmentQueryState,
  assessmentRows,
  currentIncidentRole,
  currentUserId,
  density,
  entityIndex,
  genericLoadError,
  genericQueryState,
  genericRows,
  hostQueryState,
  hostRows,
  identityQueryState,
  identityRows,
  incidentId,
  inspectorResetKey,
  loadAssessmentSurface,
  loadEntities,
  loadGenericSurface,
  savedViewSelector,
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
  if (surface === timelineViewSchemaId) {
    return (
      <TimelineWorkbook
        apiBase={apiBase}
        currentIncidentRole={currentIncidentRole}
        currentUserId={currentUserId}
        density={density}
        entityIndex={entityIndex}
        hostEntities={hostRows}
        identityEntities={identityRows}
        incidentId={incidentId}
        inspectorResetKey={inspectorResetKey}
        onQueryStateChange={setTimelineQueryState}
        onRefreshEntities={loadEntities}
        queryState={timelineQueryState}
        reloadToken={sheetReloadToken}
        renderInlineQueryControls={false}
        savedViewSelector={savedViewSelector}
        sheetRef={sheetRef}
      />
    );
  }

  if (surface === hostsViewSchemaId || surface === identitiesViewSchemaId) {
    return (
      <EntityWorkbookSurface
        apiBase={apiBase}
        currentIncidentRole={currentIncidentRole}
        density={density}
        entityIndex={entityIndex}
        entityType={surface === hostsViewSchemaId ? "host" : "identity"}
        incidentId={incidentId}
        inspectorResetKey={inspectorResetKey}
        onRefreshEntities={loadEntities}
        onToggleSort={(fieldKey) => {
          if (surface === hostsViewSchemaId) {
            setHostQueryState((current) =>
              toggleSortField(hostsContract, current, fieldKey),
            );
            return;
          }
          setIdentityQueryState((current) =>
            toggleSortField(identitiesContract, current, fieldKey),
          );
        }}
        queryState={
          surface === hostsViewSchemaId ? hostQueryState : identityQueryState
        }
        rows={surface === hostsViewSchemaId ? hostRows : identityRows}
        savedViewSelector={savedViewSelector}
      />
    );
  }

  if (surface === assessmentsViewSchemaId) {
    return (
      <AssessmentWorkbookSurface
        apiBase={apiBase}
        assessmentRows={assessmentRows}
        currentIncidentRole={currentIncidentRole}
        density={density}
        hostRows={hostRows}
        identityRows={identityRows}
        incidentId={incidentId}
        inspectorResetKey={inspectorResetKey}
        loadError={assessmentLoadError}
        onRefreshAssessmentRows={loadAssessmentSurface}
        onToggleSort={(fieldKey) => {
          setAssessmentQueryState((current) =>
            toggleSortField(assessmentsContract, current, fieldKey),
          );
        }}
        queryState={assessmentQueryState}
        savedViewSelector={savedViewSelector}
      />
    );
  }

  return (
    <GenericWorkbookSurface
      key={activeContract.viewSchemaId}
      apiBase={apiBase}
      contract={activeContract}
      currentUserId={currentUserId}
      density={density}
      incidentId={incidentId}
      inspectorResetKey={inspectorResetKey}
      loadError={genericLoadError}
      onRefresh={loadGenericSurface}
      onToggleSort={(fieldKey) => {
        setGenericQueryState((current) =>
          toggleSortField(activeContract, current, fieldKey),
        );
      }}
      queryState={genericQueryState}
      rows={genericRows}
      savedViewSelector={savedViewSelector}
    />
  );
}
