import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useMemo,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookActiveSurfacePort } from "../collaboration/workbookSurfacePort";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import type { WorkbookQueryState } from "../models/workbookQuery";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { useAssessmentSurfaceQuery } from "../query/useAssessmentSurfaceQuery";
import { useEntitySurfaceQuery } from "../query/useEntitySurfaceQuery";
import { useGenericSurfaceQuery } from "../query/useGenericSurfaceQuery";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import type { ReferenceQueryBrokerPort } from "../services/referenceQueryBroker";
import type { WorkbookSurfacesFacadeProps } from "../surfaces/WorkbookSurfacesFacade";

type QueryStateOwner = {
  readonly state: WorkbookQueryState;
  readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
};

type WorkbookSurfaceQueriesOptions = {
  readonly activeContract: ViewContract;
  readonly assessment: QueryStateOwner;
  readonly generic: QueryStateOwner;
  readonly hosts: QueryStateOwner;
  readonly identities: QueryStateOwner;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly referenceBroker: ReferenceQueryBrokerPort;
  readonly sheetRef: SheetRef;
  readonly surface: string;
  readonly timeline: QueryStateOwner;
  readonly viewQuery: WorkbookViewQueryPort;
};

/** Owns loading, invalidation, and collaboration projection for query surfaces. */
export function useWorkbookSurfaceQueries({
  activeContract,
  assessment,
  generic,
  hosts,
  identities,
  onIncidentAccessLost,
  referenceBroker,
  sheetRef,
  surface,
  timeline,
  viewQuery,
}: WorkbookSurfaceQueriesOptions): {
  readonly activeSurfacePort: WorkbookActiveSurfacePort | null;
  readonly facadeQueries: WorkbookSurfacesFacadeProps["queries"];
  readonly invalidateAll: (reason: WorkbookQueryInvalidationReason) => void;
  readonly refreshProjection: {
    readonly assessment: () => Promise<void>;
    readonly entities: () => Promise<void>;
    readonly generic: () => Promise<void>;
  };
} {
  const genericSurfaceActive =
    surface !== timelineViewSchemaId &&
    surface !== hostsViewSchemaId &&
    surface !== identitiesViewSchemaId &&
    surface !== assessmentsViewSchemaId;
  const genericQuery = useGenericSurfaceQuery({
    active: genericSurfaceActive,
    contract: activeContract,
    onIncidentAccessLost,
    queryState: generic.state,
    viewQuery,
    viewSchemaId: surface,
  });
  const assessmentQuery = useAssessmentSurfaceQuery({
    active: surface === assessmentsViewSchemaId,
    onIncidentAccessLost,
    queryState: assessment.state,
    viewQuery,
  });
  const entityQuery = useEntitySurfaceQuery({
    hostQueryState: hosts.state,
    identityQueryState: identities.state,
    onIncidentAccessLost,
    viewQuery,
  });
  const {
    applyRecordChanged: applyGenericRecordChanged,
    invalidate: invalidateGeneric,
    loadState: genericLoadState,
    refresh: refreshGeneric,
    rows: genericRows,
  } = genericQuery;
  const {
    applyRecordChanged: applyAssessmentRecordChanged,
    invalidate: invalidateAssessment,
    loadState: assessmentLoadState,
    refresh: refreshAssessment,
    rows: assessmentRows,
  } = assessmentQuery;
  const {
    applyRecordChanged: applyEntityRecordChanged,
    entityIndex,
    hostRows,
    identityRows,
    invalidate: invalidateEntities,
    loadState: entityLoadState,
    refresh: refreshEntities,
  } = entityQuery;
  const invalidateAll = useCallback(
    (reason: WorkbookQueryInvalidationReason) => {
      invalidateGeneric(reason);
      invalidateAssessment(reason);
      invalidateEntities(reason);
    },
    [invalidateAssessment, invalidateEntities, invalidateGeneric],
  );
  const activeSurfacePort = useMemo<WorkbookActiveSurfacePort | null>(() => {
    if (
      sheetRef.kind === "extension_workspace" ||
      surface === timelineViewSchemaId
    ) {
      return null;
    }
    const entitySurface =
      surface === hostsViewSchemaId || surface === identitiesViewSchemaId;
    const assessmentSurface = surface === assessmentsViewSchemaId;
    const refresh = assessmentSurface
      ? refreshAssessment
      : entitySurface
        ? refreshEntities
        : refreshGeneric;
    return {
      identity: { sheetRef, viewSchemaId: surface },
      applyRecordChanged: (payload) =>
        entitySurface
          ? applyEntityRecordChanged(payload, surface)
          : assessmentSurface
            ? applyAssessmentRecordChanged(payload)
            : applyGenericRecordChanged(payload),
      invalidate: assessmentSurface
        ? invalidateAssessment
        : entitySurface
          ? invalidateEntities
          : invalidateGeneric,
      refresh: async () => {
        await refresh();
      },
    };
  }, [
    applyAssessmentRecordChanged,
    applyEntityRecordChanged,
    applyGenericRecordChanged,
    invalidateAssessment,
    invalidateEntities,
    invalidateGeneric,
    refreshAssessment,
    refreshEntities,
    refreshGeneric,
    sheetRef,
    surface,
  ]);
  const facadeQueries = useMemo<WorkbookSurfacesFacadeProps["queries"]>(
    () => ({
      assessment: {
        loadState: assessmentLoadState,
        refresh: refreshAssessment,
        rows: assessmentRows,
        setState: assessment.setState,
        state: assessment.state,
      },
      entities: {
        hosts: {
          rows: hostRows,
          setState: hosts.setState,
          state: hosts.state,
        },
        identities: {
          rows: identityRows,
          setState: identities.setState,
          state: identities.state,
        },
        index: entityIndex,
        loadState: entityLoadState,
        refresh: refreshEntities,
      },
      generic: {
        loadState: genericLoadState,
        refresh: refreshGeneric,
        rows: genericRows,
        setState: generic.setState,
        state: generic.state,
      },
      referenceBroker,
      timeline: {
        setState: timeline.setState,
        state: timeline.state,
      },
      viewQuery,
    }),
    [
      assessment.setState,
      assessment.state,
      assessmentLoadState,
      assessmentRows,
      entityIndex,
      entityLoadState,
      generic.setState,
      generic.state,
      genericLoadState,
      genericRows,
      hostRows,
      hosts.setState,
      hosts.state,
      identities.setState,
      identities.state,
      identityRows,
      referenceBroker,
      refreshAssessment,
      refreshEntities,
      refreshGeneric,
      timeline.setState,
      timeline.state,
      viewQuery,
    ],
  );

  return {
    activeSurfacePort,
    facadeQueries,
    invalidateAll,
    refreshProjection: {
      assessment: refreshAssessment,
      entities: refreshEntities,
      generic: refreshGeneric,
    },
  };
}
