import { useCallback, useEffect, useRef, useState } from "react";
import { isAbortError } from "../../services/workbookApi";
import { genericReferenceOptionsFromRows } from "../models/genericWorkbookModel";
import {
  emptyGenericReferenceOptions,
  type GenericReferenceOptions,
} from "../models/workbookReferenceOptions";
import { requireWorkbookSurfaceRegistration } from "../models/workbookSurfaceRegistration";
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
} from "../models/workbookSurfaceRegistry";
import { referenceQueryBroker } from "../services/referenceQueryBroker";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";

const allRecordViewSchemaIds = [
  timelineViewSchemaId,
  hostsViewSchemaId,
  identitiesViewSchemaId,
  evidenceViewSchemaId,
  notesViewSchemaId,
  findingsViewSchemaId,
  investigativeQueriesViewSchemaId,
  forensicKeywordsViewSchemaId,
  taskRequestsViewSchemaId,
  decisionsViewSchemaId,
  partiesViewSchemaId,
] as const;

export function useOwnerReferenceOptions({
  apiBase,
  authorizationEpoch,
  incidentId,
  viewSchemaId,
}: {
  readonly apiBase: string | undefined;
  readonly authorizationEpoch: string;
  readonly incidentId: string;
  readonly viewSchemaId: string;
}) {
  const requirements =
    requireWorkbookSurfaceRegistration(viewSchemaId).policy
      .referenceRequirements;
  const [refreshVersion, setRefreshVersion] = useState(0);
  const contextVersionRef = useRef(0);
  const [referenceOptions, setReferenceOptions] =
    useState<GenericReferenceOptions>(() => emptyGenericReferenceOptions());
  const [referenceLoadError, setReferenceLoadError] = useState<string | null>(
    null,
  );

  const refreshReferenceOptions = useCallback(() => {
    setRefreshVersion((current) => current + 1);
  }, []);

  useEffect(() => {
    void refreshVersion;
    const controller = new AbortController();
    const contextVersion = contextVersionRef.current + 1;
    contextVersionRef.current = contextVersion;
    setReferenceLoadError(null);
    if (requirements.length === 0) {
      setReferenceOptions(emptyGenericReferenceOptions());
      return () => controller.abort();
    }
    void referenceQueryBroker
      .execute(
        requirements,
        { apiBase, authorizationEpoch, incidentId },
        controller.signal,
      )
      .then((results) => {
        if (
          controller.signal.aborted ||
          contextVersionRef.current !== contextVersion
        ) {
          return;
        }
        const rowsByView = new Map<string, readonly EntityApiRow[]>(
          results.map((result) => [
            result.requirement.viewSchemaId,
            result.rows,
          ]),
        );
        const optionsFor = (targetViewSchemaId: string) =>
          genericReferenceOptionsFromRows(targetViewSchemaId, [
            ...(rowsByView.get(targetViewSchemaId) ?? []),
          ]);
        const next: GenericReferenceOptions = {
          parties: optionsFor(partiesViewSchemaId),
          taskRequests: optionsFor(taskRequestsViewSchemaId),
          decisions: optionsFor(decisionsViewSchemaId),
          evidence: optionsFor(evidenceViewSchemaId),
          hosts: optionsFor(hostsViewSchemaId),
          identities: optionsFor(identitiesViewSchemaId),
          notes: optionsFor(notesViewSchemaId),
          timeline: optionsFor(timelineViewSchemaId),
          noteSourceRecords: [],
          allRecords: [],
        };
        next.noteSourceRecords = [
          ...next.timeline,
          ...next.hosts,
          ...next.identities,
          ...next.evidence,
        ];
        next.allRecords = allRecordViewSchemaIds.flatMap(optionsFor);
        setReferenceOptions(next);
      })
      .catch((error: unknown) => {
        if (
          controller.signal.aborted ||
          contextVersionRef.current !== contextVersion ||
          isAbortError(error)
        ) {
          return;
        }
        setReferenceOptions(emptyGenericReferenceOptions());
        setReferenceLoadError(
          error instanceof Error ? error.message : "Reference options failed.",
        );
      });
    return () => controller.abort();
  }, [apiBase, authorizationEpoch, incidentId, refreshVersion, requirements]);

  return {
    referenceLoadError,
    referenceOptions,
    refreshReferenceOptions,
  };
}
