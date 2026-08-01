import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
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
import type { WorkbookIncidentPort } from "../ports/WorkbookIncidentPort";
import { workbookOperationFailureIsAccessLoss } from "../ports/WorkbookPortResult";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { isAbortError } from "../query/workbookLatestRequest";
import type { ReferenceQueryBrokerPort } from "../services/referenceQueryBroker";

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

export function useIncidentMemberReferenceOptions({
  enabled,
  incidentPort,
  onIncidentAccessLost,
  refreshVersion = 0,
}: {
  readonly enabled: boolean;
  readonly incidentPort: WorkbookIncidentPort;
  readonly onIncidentAccessLost?: (() => void) | undefined;
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
    void incidentPort
      .listMembers({ signal: controller.signal })
      .then((result) => {
        if (controller.signal.aborted || result.kind === "aborted") return;
        if (result.kind === "rejected") {
          if (workbookOperationFailureIsAccessLoss(result.failure)) {
            onIncidentAccessLost?.();
          }
          setOptions([]);
          setError(result.failure.message);
          return;
        }
        setOptions(
          result.value.members.map((membership) => ({
            label: `${membership.displayName} (${membership.userId})`,
            recordId: membership.userId,
            viewSchemaId: "incident_member",
          })),
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
  }, [enabled, incidentPort, onIncidentAccessLost, refreshVersion]);

  return { error, options };
}

export function useOwnerReferenceOptions({
  incidentPort,
  onIncidentAccessLost,
  referenceQueryBroker,
  viewSchemaId,
}: {
  readonly incidentPort: WorkbookIncidentPort;
  readonly onIncidentAccessLost?: (() => void) | undefined;
  readonly referenceQueryBroker: ReferenceQueryBrokerPort;
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
      enabled: needsIncidentMembers,
      incidentPort,
      onIncidentAccessLost,
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
        : referenceQueryBroker.execute(requirements, controller.signal);
    void referenceRequest
      .then((results) => {
        if (
          controller.signal.aborted ||
          contextVersionRef.current !== contextVersion
        ) {
          return;
        }
        const rowsByView = new Map<string, readonly WorkbookQueryRow[]>(
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
  }, [referenceQueryBroker, refreshVersion, requirements]);

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
