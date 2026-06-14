import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useState } from "react";
import { apiPath } from "./browserApi";
import { genericReferenceOptionsFromRows } from "./genericWorkbookModel";
import { fetchJSON, parseErrorMessage, readEnvelope } from "./workbookApi";
import { buildQueryRequest, emptyWorkbookQueryState } from "./workbookQuery";
import {
  emptyGenericReferenceOptions,
  type GenericReferenceOptions,
} from "./workbookReferenceOptions";
import {
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  findingsViewSchemaId,
  forensicKeywordsViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  investigativeQueriesViewSchemaId,
  notesViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  timelineViewSchemaId,
} from "./workbookSurfaceRegistry";
import type { EntityApiRow } from "./workbookTimelineModel";

type ViewQueryEnvelope = {
  data: {
    view_schema_id: string;
    rows: EntityApiRow[];
  };
};

const genericReferenceTargetViewSchemaIds = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  partiesViewSchemaId,
  taskRequestsViewSchemaId,
  decisionsViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
  findingsViewSchemaId,
  investigativeQueriesViewSchemaId,
  forensicKeywordsViewSchemaId,
] as const;

export function useGenericReferenceOptions({
  apiBase,
  incidentId,
}: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
}) {
  const [referenceOptions, setReferenceOptions] =
    useState<GenericReferenceOptions>(() => emptyGenericReferenceOptions());
  const [referenceLoadError, setReferenceLoadError] = useState<string | null>(
    null,
  );

  const refreshReferenceOptions = useCallback(async () => {
    setReferenceLoadError(null);
    const loaded = await Promise.all(
      genericReferenceTargetViewSchemaIds.map(async (viewSchemaId) => {
        const targetContract = requireViewContract(viewSchemaId);
        const result = await fetchJSON<ViewQueryEnvelope>(
          apiPath(
            apiBase,
            `/api/v1/incidents/${incidentId}/views/${viewSchemaId}/query`,
          ),
          {
            method: "POST",
            body: JSON.stringify(
              buildQueryRequest(targetContract, emptyWorkbookQueryState()),
            ),
          },
        );
        if (!result.ok) {
          throw new Error(parseErrorMessage(result.payload));
        }
        const envelope = readEnvelope<ViewQueryEnvelope>(result.payload);
        return [viewSchemaId, envelope.data.rows] as const;
      }),
    );
    const rowsByView = Object.fromEntries(loaded) as Record<
      string,
      EntityApiRow[]
    >;
    const next: GenericReferenceOptions = {
      parties: genericReferenceOptionsFromRows(
        partiesViewSchemaId,
        rowsByView[partiesViewSchemaId] ?? [],
      ),
      taskRequests: genericReferenceOptionsFromRows(
        taskRequestsViewSchemaId,
        rowsByView[taskRequestsViewSchemaId] ?? [],
      ),
      decisions: genericReferenceOptionsFromRows(
        decisionsViewSchemaId,
        rowsByView[decisionsViewSchemaId] ?? [],
      ),
      evidence: genericReferenceOptionsFromRows(
        evidenceViewSchemaId,
        rowsByView[evidenceViewSchemaId] ?? [],
      ),
      hosts: genericReferenceOptionsFromRows(
        hostsViewSchemaId,
        rowsByView[hostsViewSchemaId] ?? [],
      ),
      identities: genericReferenceOptionsFromRows(
        identitiesViewSchemaId,
        rowsByView[identitiesViewSchemaId] ?? [],
      ),
      notes: genericReferenceOptionsFromRows(
        notesViewSchemaId,
        rowsByView[notesViewSchemaId] ?? [],
      ),
      timeline: genericReferenceOptionsFromRows(
        timelineViewSchemaId,
        rowsByView[timelineViewSchemaId] ?? [],
      ),
      noteSourceRecords: [],
      allRecords: [],
    };
    next.noteSourceRecords = [
      ...next.timeline,
      ...next.hosts,
      ...next.identities,
      ...next.evidence,
    ];
    next.allRecords = [
      ...next.timeline,
      ...next.hosts,
      ...next.identities,
      ...next.evidence,
      ...next.notes,
      ...genericReferenceOptionsFromRows(
        findingsViewSchemaId,
        rowsByView[findingsViewSchemaId] ?? [],
      ),
      ...genericReferenceOptionsFromRows(
        investigativeQueriesViewSchemaId,
        rowsByView[investigativeQueriesViewSchemaId] ?? [],
      ),
      ...genericReferenceOptionsFromRows(
        forensicKeywordsViewSchemaId,
        rowsByView[forensicKeywordsViewSchemaId] ?? [],
      ),
      ...next.taskRequests,
      ...next.decisions,
      ...next.parties,
    ];
    setReferenceOptions(next);
  }, [apiBase, incidentId]);

  useEffect(() => {
    let isCurrent = true;
    refreshReferenceOptions().catch((error: unknown) => {
      if (!isCurrent) {
        return;
      }
      setReferenceOptions(emptyGenericReferenceOptions());
      setReferenceLoadError(
        error instanceof Error ? error.message : "Reference options failed.",
      );
    });
    return () => {
      isCurrent = false;
    };
  }, [refreshReferenceOptions]);

  return {
    referenceLoadError,
    referenceOptions,
    refreshReferenceOptions,
  };
}
