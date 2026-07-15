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
} from "../../services/workbookApi";
import type { EntityRow } from "../models/entityWorkbookModel";
import { entityRowFromApi } from "../models/entityWorkbookModel";
import { normalizeWorkbookViewRows } from "../models/workbookContractRows";
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
  const [entityLoadError, setEntityLoadError] = useState<string | null>(null);
  const [genericRows, setGenericRows] = useState<EntityApiRow[]>([]);
  const [genericLoadError, setGenericLoadError] = useState<string | null>(null);
  const [assessmentRows, setAssessmentRows] = useState<EntityApiRow[]>([]);
  const [assessmentLoadError, setAssessmentLoadError] = useState<string | null>(
    null,
  );
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
    setEntityLoadError(null);
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
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Entity load failed.",
        onIncidentAccessLost,
      );
      setEntityLoadError(message);
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
    setAssessmentLoadError(null);
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
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Assessment load failed.",
        onIncidentAccessLost,
      );
      setAssessmentLoadError(message);
      setAssessmentRows([]);
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
      return;
    }
    const requestedSurface = surface;
    const request = beginLatestQuery(genericQueryRuntimeRef);
    setGenericLoadError(null);
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
      setGenericRows(
        requestedSurface === notesViewSchemaId
          ? normalizeWorkbookViewRows(
              activeContract,
              envelope.data.rows,
              `${requestedSurface} query response`,
            )
          : envelope.data.rows,
      );
    } catch (error) {
      if (!request.isCurrent() || isAbortError(error)) {
        return;
      }
      const message = handleWorkbookLoadFailure(
        error,
        "Surface load failed.",
        onIncidentAccessLost,
      );
      setGenericLoadError(message);
      setGenericRows([]);
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
    assessmentLoadError,
    assessmentRows,
    entityIndex,
    entityLoadError,
    genericLoadError,
    genericRows,
    hostRows,
    identityRows,
    loadAssessmentSurface,
    loadEntities,
    loadGenericSurface,
  };
}
