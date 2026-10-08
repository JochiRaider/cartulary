import type { ViewContract } from "@cartulary/view-contracts";
import { type SetStateAction, useCallback, useMemo, useRef } from "react";
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
import type { WorkbookSurfacesFacadeProps } from "../surfaces/WorkbookSurfacesFacade";

type WorkbookSurfaceQueriesOptions = {
  readonly ordinaryCreateOwner?: WorkbookCommittedRecordPort;
  readonly explicitPatchOwner?: WorkbookExplicitPatchOwner;
  readonly decisionOwner?: DecisionSupersessionOwnerPort;
  readonly indicatorOwner?: WorkbookCommittedRecordPort;
  readonly assessmentOwner?: AssessmentCommittedRecordPort;
  readonly activeContract: ViewContract;
  readonly queryStateForSurface: (viewSchemaId: string) => WorkbookQueryState;
  readonly setQueryStateForSurface: (
    viewSchemaId: string,
    action: SetStateAction<WorkbookQueryState>,
  ) => void;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly sheetRef: SheetRef;
  readonly surface: string;
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
  queryStateForSurface,
  setQueryStateForSurface,
  onAuthorityUncertain,
  sheetRef,
  surface,
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
  // Each callback retains the exact schema it was composed for, including generic
  // surfaces after navigation. Unrelated state updates do not replace setters.
  const setters = useMemo(() => {
    const bind =
      (viewSchemaId: string) => (action: SetStateAction<WorkbookQueryState>) =>
        setQueryStateForSurface(viewSchemaId, action);
    return {
      assessment: bind(assessmentsViewSchemaId),
      generic: bind(surface),
      hosts: bind(hostsViewSchemaId),
      identities: bind(identitiesViewSchemaId),
      timeline: bind(timelineViewSchemaId),
    };
  }, [setQueryStateForSurface, surface]);
  const timelineQueryState = queryStateForSurface(timelineViewSchemaId);
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
    queryState: queryStateForSurface(surface),
    viewQuery,
    viewSchemaId: surface,
  });
  const assessmentQuery = useAssessmentSurfaceQuery({
    committedRecords: assessmentOwner,
    active: surface === assessmentsViewSchemaId,
    onAuthorityUncertain,
    queryState: queryStateForSurface(assessmentsViewSchemaId),
    viewQuery,
  });
  const entityQuery = useEntitySurfaceQuery({
    activeViewSchemaId: surface,
    ordinaryCreateOwner,
    editOwner: explicitPatchOwner,
    hostQueryState: queryStateForSurface(hostsViewSchemaId),
    identityQueryState: queryStateForSurface(identitiesViewSchemaId),
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
        browsingRegistry.invalidateAll();
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
        const browser = browsingRegistry.find(surface);
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
        setState: setters.assessment,
        state: assessmentQuery.acceptedQueryState,
      },
      entities: {
        hosts: {
          rows:
            surface === timelineViewSchemaId
              ? entityQuery.references.hosts
              : hostRows,
          setState: setters.hosts,
          state: entityQuery.hosts.acceptedQueryState,
        },
        identities: {
          rows:
            surface === timelineViewSchemaId
              ? entityQuery.references.identities
              : identityRows,
          setState: setters.identities,
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
        setState: setters.generic,
        state: genericQuery.acceptedQueryState,
      },
      timeline: {
        setState: setters.timeline,
        state: timelineQueryState,
      },
      viewQuery,
    }),
    [
      setters.assessment,
      assessmentLoadState,
      assessmentRows,
      assessmentQuery.acceptedQueryState,
      entityIndex,
      entityLoadState,
      setters.generic,
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
      setters.hosts,
      setters.identities,
      identityRows,
      refreshAssessment,
      refreshEntities,
      refreshGeneric,
      setters.timeline,
      timelineQueryState,
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
