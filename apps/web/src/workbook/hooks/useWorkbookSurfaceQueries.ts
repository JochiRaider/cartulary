import type { ViewContract } from "@cartulary/view-contracts";
import {
  type Dispatch,
  type SetStateAction,
  useCallback,
  useMemo,
  useRef,
} from "react";
import type { SheetRef } from "../../shared/sheetRef";
import type { WorkbookActiveSurfacePort } from "../collaboration/workbookSurfacePort";
import type { AssessmentCommittedRecordPort } from "../features/assessments/assessmentOperation";
import type { DecisionSupersessionOwnerPort } from "../features/coordination/decisionSupersessionOperation";
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
import type { WorkbookCommittedRecordPort } from "../query/WorkbookCommittedRecordPort";
import { useWorkbookBrowsingRegistry } from "../query/WorkbookQueryBrowsingContext";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";
import type { WorkbookExplicitPatchOwner } from "../runtime/WorkbookExplicitPatchOwner";
import type { ReferenceQueryBrokerPort } from "../services/referenceQueryBroker";
import type { WorkbookSurfacesFacadeProps } from "../surfaces/WorkbookSurfacesFacade";

type QueryStateOwner = {
  readonly state: WorkbookQueryState;
  readonly setState: Dispatch<SetStateAction<WorkbookQueryState>>;
};

type WorkbookSurfaceQueriesOptions = {
  readonly ordinaryCreateOwner?: WorkbookCommittedRecordPort;
  readonly explicitPatchOwner?: WorkbookExplicitPatchOwner;
  readonly decisionOwner?: DecisionSupersessionOwnerPort;
  readonly indicatorOwner?: WorkbookCommittedRecordPort;
  readonly assessmentOwner?: AssessmentCommittedRecordPort;
  readonly activeContract: ViewContract;
  readonly assessment: QueryStateOwner;
  readonly generic: QueryStateOwner;
  readonly hosts: QueryStateOwner;
  readonly identities: QueryStateOwner;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly referenceBroker: ReferenceQueryBrokerPort;
  readonly sheetRef: SheetRef;
  readonly surface: string;
  readonly timeline: QueryStateOwner;
  readonly viewQuery: WorkbookViewQueryPort;
};

/** Owns loading, invalidation, and collaboration projection for query surfaces. */
export function useWorkbookSurfaceQueries({
  ordinaryCreateOwner,
  explicitPatchOwner,
  decisionOwner,
  indicatorOwner,
  assessmentOwner,
  activeContract,
  assessment,
  generic,
  hosts,
  identities,
  onAuthorityUncertain,
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
    readonly assessment: (options?: {
      readonly requireAcceptance?: boolean;
    }) => Promise<void>;
    readonly entities: (options?: {
      readonly requireAcceptance?: boolean;
    }) => Promise<void>;
    readonly generic: (options?: {
      readonly requireAcceptance?: boolean;
    }) => Promise<void>;
  };
} {
  const browsingRegistry = useWorkbookBrowsingRegistry();
  const genericSurfaceActive =
    surface !== timelineViewSchemaId &&
    surface !== hostsViewSchemaId &&
    surface !== identitiesViewSchemaId &&
    surface !== assessmentsViewSchemaId;
  const genericQuery = useGenericSurfaceQuery({
    ordinaryCreateOwner,
    explicitPatchOwner,
    decisionOwner,
    indicatorOwner,
    active: genericSurfaceActive,
    contract: activeContract,
    onAuthorityUncertain,
    queryState: generic.state,
    viewQuery,
    viewSchemaId: surface,
  });
  const assessmentQuery = useAssessmentSurfaceQuery({
    committedRecords: assessmentOwner,
    active: surface === assessmentsViewSchemaId,
    onAuthorityUncertain,
    queryState: assessment.state,
    viewQuery,
  });
  const entityQuery = useEntitySurfaceQuery({
    referenceBroker,
    activeViewSchemaId: surface,
    ordinaryCreateOwner,
    editOwner: explicitPatchOwner,
    hostQueryState: hosts.state,
    identityQueryState: identities.state,
    onAuthorityUncertain,
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
  const currentInvalidators = useRef([
    invalidateGeneric,
    invalidateAssessment,
    invalidateEntities,
  ]);
  currentInvalidators.current = [
    invalidateGeneric,
    invalidateAssessment,
    invalidateEntities,
  ];
  const invalidateAll = useCallback(
    (reason: WorkbookQueryInvalidationReason) => {
      if (
        reason.kind !== "incident_closed" &&
        reason.kind !== "collaboration_reset_required"
      )
        browsingRegistry?.invalidateAll();
      for (const invalidate of currentInvalidators.current) invalidate(reason);
    },
    [browsingRegistry],
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
      refresh: async (options) => {
        const read = () =>
          refresh({
            requireAcceptance: options?.reason === "authorization_recovered",
          });
        const browser = browsingRegistry?.find(surface);
        if (options?.reason === "record_changed" && browser)
          await browser.reconcile(read);
        else await read();
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
    browsingRegistry,
  ]);
  const facadeQueries = useMemo<WorkbookSurfacesFacadeProps["queries"]>(
    () => ({
      assessment: {
        loadState: assessmentLoadState,
        refresh: refreshAssessment,
        rows: assessmentRows,
        setState: assessment.setState,
        state: assessmentQuery.acceptedQueryState,
      },
      entities: {
        hosts: {
          rows:
            surface === timelineViewSchemaId
              ? entityQuery.references.hosts
              : hostRows,
          setState: hosts.setState,
          state: entityQuery.hosts.acceptedQueryState,
        },
        identities: {
          rows:
            surface === timelineViewSchemaId
              ? entityQuery.references.identities
              : identityRows,
          setState: identities.setState,
          state: entityQuery.identities.acceptedQueryState,
        },
        index:
          surface === timelineViewSchemaId
            ? entityQuery.references.index
            : entityIndex,
        loadState: entityLoadState,
        refresh: refreshEntities,
      },
      generic: {
        loadState: genericLoadState,
        refresh: refreshGeneric,
        rows: genericRows,
        setState: generic.setState,
        state: genericQuery.acceptedQueryState,
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
      assessmentLoadState,
      assessmentRows,
      assessmentQuery.acceptedQueryState,
      entityIndex,
      entityLoadState,
      generic.setState,
      genericLoadState,
      genericRows,
      genericQuery.acceptedQueryState,
      hostRows,
      entityQuery.hosts.acceptedQueryState,
      entityQuery.identities.acceptedQueryState,
      entityQuery.references.hosts,
      entityQuery.references.identities,
      entityQuery.references.index,
      surface,
      hosts.setState,
      identities.setState,
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
