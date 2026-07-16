import {
  requireViewContract,
  type ViewContract,
} from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { apiPath } from "../../services/browserApi";
import {
  abortLatestQuery,
  beginLatestQuery,
  fetchWorkbookJSON,
  handleWorkbookLoadFailure,
  isAbortError,
  type LatestQueryRuntime,
  parseErrorMessage,
  readEnvelope,
  workbookLoadFailureIsAccessLoss,
} from "../../services/workbookApi";
import type { EntityRow } from "../models/entityWorkbookModel";
import { entityRowFromApi } from "../models/entityWorkbookModel";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
import {
  initialWorkbookQueryLoadState,
  type WorkbookQueryLoadState,
} from "../models/workbookGridState";
import {
  buildQueryRequest,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import {
  assessmentsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  notesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import { reconcileWorkbookRecordRows } from "../utils/workbookRowReconciliation";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);
const assessmentsContract = requireViewContract(assessmentsViewSchemaId);

type ViewQueryEnvelope = {
  data: {
    incident_id: string;
    view_schema_id: string;
    rows: EntityApiRow[];
  };
};

export type WorkbookSurfaceLoaderInput = {
  readonly activeContract: ViewContract;
  readonly apiBase: string | undefined;
  readonly assessmentQueryState: WorkbookQueryState;
  readonly genericQueryState: WorkbookQueryState;
  readonly hostQueryState: WorkbookQueryState;
  readonly identityQueryState: WorkbookQueryState;
  readonly incidentId: string;
  readonly onIncidentAccessLost: (() => void) | undefined;
  readonly surface: string;
};

export function useWorkbookSurfaceLoaders({
  activeContract,
  apiBase,
  assessmentQueryState,
  genericQueryState,
  hostQueryState,
  identityQueryState,
  incidentId,
  onIncidentAccessLost,
  surface,
}: WorkbookSurfaceLoaderInput) {
  const [hostRows, setHostRows] = useState<EntityRow[]>([]);
  const [identityRows, setIdentityRows] = useState<EntityRow[]>([]);
  const [entityLoadState, setEntityLoadState] =
    useState<WorkbookQueryLoadState>(initialWorkbookQueryLoadState);
  const [genericRows, setGenericRows] = useState<EntityApiRow[]>([]);
  const [genericLoadState, setGenericLoadState] =
    useState<WorkbookQueryLoadState>(initialWorkbookQueryLoadState);
  const [assessmentRows, setAssessmentRows] = useState<EntityApiRow[]>([]);
  const [assessmentLoadState, setAssessmentLoadState] =
    useState<WorkbookQueryLoadState>(initialWorkbookQueryLoadState);
  const entityAcceptedRowCountRef = useRef(0);
  const assessmentAcceptedRowCountRef = useRef(0);
  const genericAcceptedRowCountRef = useRef(0);
  const genericSurfaceRef = useRef(surface);
  const entityQueryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  const assessmentQueryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });
  const genericQueryRuntimeRef = useRef<LatestQueryRuntime>({
    controller: null,
    sequence: 0,
  });

  const entityIndex = useMemo(() => {
    const index: Record<string, EntityRow> = {};
    for (const row of [...hostRows, ...identityRows]) {
      index[row.recordId] = row;
    }
    return index;
  }, [hostRows, identityRows]);

  const queryEntityView = useCallback(
    async (
      viewSchemaId: string,
      entityType: EntityRow["entityType"],
      queryState: WorkbookQueryState,
      signal: AbortSignal,
    ) => {
      const contract =
        viewSchemaId === hostsViewSchemaId ? hostsContract : identitiesContract;
      const result = await fetchWorkbookJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
        ),
        {
          method: "POST",
          signal,
          body: JSON.stringify(buildQueryRequest(contract, queryState)),
        },
      );
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      if (envelope.data.view_schema_id !== viewSchemaId) {
        throw new Error(
          `Entity surface load returned ${envelope.data.view_schema_id} for ${viewSchemaId}.`,
        );
      }
      return normalizeWorkbookViewRows(
        contract,
        envelope.data.rows,
        `${viewSchemaId} query response`,
      ).map((row) => entityRowFromApi(row, entityType));
    },
    [apiBase, incidentId],
  );

  const loadEntities = useCallback(async () => {
    const request = beginLatestQuery(entityQueryRuntimeRef);
    setEntityLoadState(
      entityAcceptedRowCountRef.current > 0
        ? { kind: "refreshing" }
        : initialWorkbookQueryLoadState,
    );
    try {
      const [nextHosts, nextIdentities] = await Promise.all([
        queryEntityView(
          hostsViewSchemaId,
          "host",
          hostQueryState,
          request.signal,
        ),
        queryEntityView(
          identitiesViewSchemaId,
          "identity",
          identityQueryState,
          request.signal,
        ),
      ]);
      if (!request.isCurrent()) {
        return;
      }
      setHostRows((current) => [
        ...reconcileWorkbookRecordRows(current, nextHosts),
      ]);
      setIdentityRows((current) => [
        ...reconcileWorkbookRecordRows(current, nextIdentities),
      ]);
      entityAcceptedRowCountRef.current =
        nextHosts.length + nextIdentities.length;
      setEntityLoadState({ kind: "ready" });
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Entity load failed.",
        onIncidentAccessLost,
      );
      if (workbookLoadFailureIsAccessLoss(message)) {
        entityAcceptedRowCountRef.current = 0;
        setHostRows([]);
        setIdentityRows([]);
        setEntityLoadState({ kind: "permission_denied", message });
      } else if (entityAcceptedRowCountRef.current > 0) {
        setEntityLoadState({ kind: "stale_error", message });
      } else {
        setEntityLoadState({ kind: "unavailable", message });
      }
    }
  }, [
    hostQueryState,
    identityQueryState,
    onIncidentAccessLost,
    queryEntityView,
  ]);

  const isSpecializedSurface =
    surface === timelineViewSchemaId ||
    surface === hostsViewSchemaId ||
    surface === identitiesViewSchemaId ||
    surface === assessmentsViewSchemaId;

  const loadAssessmentSurface = useCallback(async () => {
    if (surface !== assessmentsViewSchemaId) {
      abortLatestQuery(assessmentQueryRuntimeRef);
      return;
    }
    const request = beginLatestQuery(assessmentQueryRuntimeRef);
    setAssessmentLoadState(
      assessmentAcceptedRowCountRef.current > 0
        ? { kind: "refreshing" }
        : initialWorkbookQueryLoadState,
    );
    try {
      const result = await fetchWorkbookJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${assessmentsViewSchemaId}/query`,
        ),
        {
          method: "POST",
          signal: request.signal,
          body: JSON.stringify(
            buildQueryRequest(assessmentsContract, assessmentQueryState),
          ),
        },
      );
      if (!request.isCurrent()) {
        return;
      }
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      setAssessmentRows(envelope.data.rows);
      assessmentAcceptedRowCountRef.current = envelope.data.rows.length;
      setAssessmentLoadState({ kind: "ready" });
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Assessment load failed.",
        onIncidentAccessLost,
      );
      if (workbookLoadFailureIsAccessLoss(message)) {
        assessmentAcceptedRowCountRef.current = 0;
        setAssessmentRows([]);
        setAssessmentLoadState({ kind: "permission_denied", message });
      } else if (assessmentAcceptedRowCountRef.current > 0) {
        setAssessmentLoadState({ kind: "stale_error", message });
      } else {
        setAssessmentLoadState({ kind: "unavailable", message });
      }
    }
  }, [
    apiBase,
    assessmentQueryState,
    incidentId,
    onIncidentAccessLost,
    surface,
  ]);

  const loadGenericSurface = useCallback(async () => {
    if (isSpecializedSurface) {
      abortLatestQuery(genericQueryRuntimeRef);
      genericSurfaceRef.current = surface;
      genericAcceptedRowCountRef.current = 0;
      setGenericRows([]);
      setGenericLoadState(initialWorkbookQueryLoadState);
      return;
    }
    const requestedSurface = surface;
    if (genericSurfaceRef.current !== requestedSurface) {
      genericSurfaceRef.current = requestedSurface;
      genericAcceptedRowCountRef.current = 0;
      setGenericRows([]);
    }
    const request = beginLatestQuery(genericQueryRuntimeRef);
    setGenericLoadState(
      genericAcceptedRowCountRef.current > 0
        ? { kind: "refreshing" }
        : initialWorkbookQueryLoadState,
    );
    try {
      const result = await fetchWorkbookJSON<ViewQueryEnvelope>(
        apiPath(
          apiBase,
          `/api/v1/incidents/${incidentId}/views/${requestedSurface}/query`,
        ),
        {
          method: "POST",
          signal: request.signal,
          body: JSON.stringify(
            buildQueryRequest(activeContract, genericQueryState),
          ),
        },
      );
      if (!request.isCurrent()) {
        return;
      }
      if (!result.ok) {
        throw new Error(parseErrorMessage(result.payload));
      }
      const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
      if (envelope.data.view_schema_id !== requestedSurface) {
        throw new Error(
          `Surface load returned ${envelope.data.view_schema_id} for ${requestedSurface}.`,
        );
      }
      const nextRows =
        requestedSurface === notesViewSchemaId
          ? normalizeWorkbookViewRows(
              activeContract,
              envelope.data.rows,
              `${requestedSurface} query response`,
            )
          : envelope.data.rows;
      setGenericRows(nextRows);
      genericAcceptedRowCountRef.current = nextRows.length;
      setGenericLoadState({ kind: "ready" });
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Surface load failed.",
        onIncidentAccessLost,
      );
      if (workbookLoadFailureIsAccessLoss(message)) {
        genericAcceptedRowCountRef.current = 0;
        setGenericRows([]);
        setGenericLoadState({ kind: "permission_denied", message });
      } else if (genericAcceptedRowCountRef.current > 0) {
        setGenericLoadState({ kind: "stale_error", message });
      } else {
        setGenericLoadState({ kind: "unavailable", message });
      }
    }
  }, [
    activeContract,
    apiBase,
    genericQueryState,
    incidentId,
    isSpecializedSurface,
    onIncidentAccessLost,
    surface,
  ]);

  useEffect(
    () => () => {
      abortLatestQuery(entityQueryRuntimeRef);
      abortLatestQuery(assessmentQueryRuntimeRef);
      abortLatestQuery(genericQueryRuntimeRef);
    },
    [],
  );

  return {
    assessmentLoadState,
    assessmentRows,
    entityIndex,
    entityLoadState,
    genericLoadState,
    genericRows,
    hostRows,
    identityRows,
    loadAssessmentSurface,
    loadEntities,
    loadGenericSurface,
  };
}
