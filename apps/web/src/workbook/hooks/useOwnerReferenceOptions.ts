import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  isAbortError,
  readEnvelope,
} from "../../services/workbookApi";
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

type MembershipEnvelope = {
  readonly data: {
    readonly memberships: readonly {
      readonly display_name: string;
      readonly user_id: string;
    }[];
  };
};

export function useIncidentMemberReferenceOptions({
  apiBase,
  enabled,
  incidentId,
  refreshVersion = 0,
}: {
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly refreshVersion?: number;
}) {
  const [options, setOptions] = useState<
    GenericReferenceOptions["incidentMembers"]
  >([]);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    void refreshVersion;
    const controller = new AbortController();
    setError(null);
    if (!enabled) {
      setOptions([]);
      return () => controller.abort();
    }
    void fetchWorkbookJSON<MembershipEnvelope>(
      apiPath(apiBase, `/api/v1/incidents/${incidentId}/memberships`),
      { signal: controller.signal },
    )
      .then((result) => {
        if (controller.signal.aborted) return;
        if (!result.ok) {
          throw new Error("Incident member references are unavailable.");
        }
        setOptions(
          readEnvelope<MembershipEnvelope>(result.payload).data.memberships.map(
            (membership) => ({
              label: `${membership.display_name} (${membership.user_id})`,
              recordId: membership.user_id,
              viewSchemaId: "incident_member",
            }),
          ),
        );
      })
      .catch((loadError: unknown) => {
        if (controller.signal.aborted || isAbortError(loadError)) return;
        setOptions([]);
        setError(
          loadError instanceof Error
            ? loadError.message
            : "Incident member references failed.",
        );
      });
    return () => controller.abort();
  }, [apiBase, enabled, incidentId, refreshVersion]);

  return { error, options };
}

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
  const needsIncidentMembers = requireViewContract(viewSchemaId).fields.some(
    (field) =>
      field.directReferenceContractId === "incident_member_user_ref_v1",
  );
  const [refreshVersion, setRefreshVersion] = useState(0);
  const { error: incidentMemberLoadError, options: incidentMemberOptions } =
    useIncidentMemberReferenceOptions({
      apiBase,
      enabled: needsIncidentMembers,
      incidentId,
      refreshVersion,
    });
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
    const referenceRequest =
      requirements.length === 0
        ? Promise.resolve([])
        : referenceQueryBroker.execute(
            requirements,
            { apiBase, authorizationEpoch, incidentId },
            controller.signal,
          );
    void referenceRequest
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
          incidentMembers: [],
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

  const resolvedReferenceOptions = useMemo(
    () => ({
      ...referenceOptions,
      incidentMembers: incidentMemberOptions,
    }),
    [incidentMemberOptions, referenceOptions],
  );

  return {
    referenceLoadError: referenceLoadError ?? incidentMemberLoadError,
    referenceOptions: resolvedReferenceOptions,
    refreshReferenceOptions,
  };
}
