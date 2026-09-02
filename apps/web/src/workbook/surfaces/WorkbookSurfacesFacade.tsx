import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type ReactNode,
  type SetStateAction,
  useState,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookIncidentRole } from "../../shared/workbookShellContracts";
import type { WorkbookCollaborationCoordinator } from "../collaboration/WorkbookCollaborationCoordinator";
import { AssessmentWorkbookSurface } from "../components/AssessmentWorkbookSurface";
import { EntityWorkbookSurface } from "../components/EntityWorkbookSurface";
import { ContractWorkbookSurface } from "../components/GenericWorkbookSurface";
import type { WorkbookSurfaceLayoutOwner } from "../layout/useWorkbookLayoutFacade";
import type { EntityRow } from "../models/entityWorkbookModel";
import type { WorkbookQueryLoadState } from "../models/workbookGridState";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
  replaceWorkbookSort,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
import type { WorkbookMutationCommandPorts } from "../mutations/workbookMutationCommandPorts";
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { ReferenceQueryBrokerPort } from "../services/referenceQueryBroker";
import { TimelineWorkbook } from "../timeline/components/TimelineWorkbook";

export type WorkbookSurfacesFacadeProps = {
  readonly collaboration: {
    readonly projection: WorkbookCollaborationCoordinator;
  };
  readonly continuity: {
    readonly resetKey: string;
  };
  readonly incident: {
    readonly apiBase?: string | undefined;
    readonly currentIncidentRole: WorkbookIncidentRole | null;
    readonly currentUserId: string | null;
    readonly incidentPort: WorkbookIncidentPort;
    readonly incidentId: string;
    readonly onIncidentAccessLost: (() => void) | undefined;
  };
  readonly inspector: {
    readonly resetKey: string;
  };
  readonly layout: WorkbookSurfaceLayoutOwner;
  readonly mutations: {
    readonly commands: WorkbookMutationCommandPorts;
    readonly onActivateConflict:
      | ((invoker: HTMLButtonElement) => void)
      | undefined;
    readonly pending: WorkbookPendingMutationPort;
    readonly runtime: WorkbookMutationRuntime;
  };
  readonly queries: {
    readonly viewQuery: WorkbookViewQueryPort;
    readonly assessment: {
      readonly loadState: WorkbookQueryLoadState;
      readonly refresh: () => Promise<void>;
      readonly rows: WorkbookQueryRow[];
      readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
      readonly state: WorkbookQueryState;
    };
    readonly entities: {
      readonly hosts: {
        readonly rows: EntityRow[];
        readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
        readonly state: WorkbookQueryState;
      };
      readonly identities: {
        readonly rows: EntityRow[];
        readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
        readonly state: WorkbookQueryState;
      };
      readonly index: Record<string, EntityRow>;
      readonly loadState: WorkbookQueryLoadState;
      readonly refresh: () => Promise<void>;
    };
    readonly generic: {
      readonly loadState: WorkbookQueryLoadState;
      readonly refresh: () => Promise<void>;
      readonly rows: WorkbookQueryRow[];
      readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
      readonly state: WorkbookQueryState;
    };
    readonly referenceBroker: ReferenceQueryBrokerPort;
    readonly timeline: {
      readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
      readonly state: WorkbookQueryState;
    };
  };
  readonly viewState: {
    readonly activeContract: ViewContract;
    readonly queryControls?: ReactNode | undefined;
    readonly savedViewSelector?: ReactNode | undefined;
    readonly sheetRef: SheetRef;
    readonly sheetReloadToken: number;
    readonly surface: string;
  };
};

export function WorkbookSurfacesFacade({
  collaboration,
  continuity,
  incident,
  inspector,
  layout,
  mutations,
  queries,
  viewState,
}: WorkbookSurfacesFacadeProps) {
  const {
    apiBase,
    currentIncidentRole,
    currentUserId,
    incidentPort,
    incidentId,
    onIncidentAccessLost,
  } = incident;
  const {
    activeContract,
    queryControls,
    savedViewSelector,
    sheetRef,
    sheetReloadToken,
    surface,
  } = viewState;
  const {
    assessment,
    entities,
    generic,
    referenceBroker: referenceQueryBroker,
    timeline,
    viewQuery,
  } = queries;
  const {
    commands: mutationCommands,
    onActivateConflict,
    pending: pendingMutationPort,
    runtime: mutationRuntime,
  } = mutations;
  const { projection: collaborationProjection } = collaboration;
  const inspectorResetKey = inspector.resetKey;
  const continuityResetKey = continuity.resetKey;
  const assessmentLoadState = assessment.loadState;
  const assessmentQueryState = assessment.state;
  const assessmentRows = assessment.rows;
  const entityIndex = entities.index;
  const entityLoadState = entities.loadState;
  const genericLoadState = generic.loadState;
  const genericQueryState = generic.state;
  const genericRows = generic.rows;
  const hostQueryState = entities.hosts.state;
  const hostRows = entities.hosts.rows;
  const identityQueryState = entities.identities.state;
  const identityRows = entities.identities.rows;
  const timelineQueryState = timeline.state;
  const loadAssessmentSurface = assessment.refresh;
  const loadEntities = entities.refresh;
  const loadGenericSurface = generic.refresh;
  const setAssessmentQueryState = assessment.setState;
  const setGenericQueryState = generic.setState;
  const setHostQueryState = entities.hosts.setState;
  const setIdentityQueryState = entities.identities.setState;
  const setTimelineQueryState = timeline.setState;
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
          pendingMutationPort,
          mutationCommands: mutationCommands.timeline,
          indicatorWorkflow: mutationCommands.indicators,
          incident: {
            id: incidentId,
            apiBase,
            continuityResetKey,
            currentUserId,
            currentRole: currentIncidentRole,
            incidentPort,
            sheetRef,
            inspectorResetKey,
            reloadToken: sheetReloadToken,
          },
          query: {
            viewQuery,
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
          layout,
          onActivateConflict,
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
        key={activeContract.viewSchemaId}
        continuityResetKey={continuityResetKey}
        currentIncidentRole={currentIncidentRole}
        currentUserId={currentUserId}
        entityIndex={entityIndex}
        entityType={isHosts ? "host" : "identity"}
        inspectorResetKey={inspectorResetKey}
        layout={layout}
        mutationRuntime={mutationRuntime}
        mutationCommands={mutationCommands.entity}
        onActivateConflict={onActivateConflict}
        recordMutationCommands={mutationCommands.records}
        relatedMutationCommands={mutationCommands.timeline.related}
        collaborationProjection={collaborationProjection}
        loadState={entityLoadState}
        onRefreshEntities={loadEntities}
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
        viewQuery={viewQuery}
      />
    );
  }

  if (registration.renderer === "assessment") {
    return (
      <AssessmentWorkbookSurface
        assessmentRows={assessmentRows}
        currentIncidentRole={currentIncidentRole}
        currentUserId={currentUserId}
        continuityResetKey={continuityResetKey}
        hostRows={hostRows}
        identityRows={identityRows}
        inspectorResetKey={inspectorResetKey}
        layout={layout}
        mutationRuntime={mutationRuntime}
        mutationCommands={mutationCommands.assessment}
        onActivateConflict={onActivateConflict}
        recordMutationCommands={mutationCommands.records}
        relatedMutationCommands={mutationCommands.timeline.related}
        collaborationProjection={collaborationProjection}
        loadState={assessmentLoadState}
        onRefreshAssessmentRows={loadAssessmentSurface}
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
        viewQuery={viewQuery}
      />
    );
  }

  return (
    <ContractWorkbookSurface
      key={activeContract.viewSchemaId}
      contract={activeContract}
      continuityResetKey={continuityResetKey}
      currentIncidentRole={currentIncidentRole}
      currentUserId={currentUserId}
      incidentPort={incidentPort}
      inspectorResetKey={inspectorResetKey}
      loadState={genericLoadState}
      layout={layout}
      mutationRuntime={mutationRuntime}
      mutationCommands={mutationCommands}
      onActivateConflict={onActivateConflict}
      referenceQueryBroker={referenceQueryBroker}
      collaborationProjection={collaborationProjection}
      sheetRef={sheetRef}
      onIncidentAccessLost={onIncidentAccessLost}
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
    />
  );
}
