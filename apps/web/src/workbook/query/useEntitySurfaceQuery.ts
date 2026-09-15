import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import {
  type EntityRow,
  entityRowFromApi,
} from "../models/entityWorkbookModel";
import type { WorkbookQueryState } from "../models/workbookQuery";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { referenceRequirement } from "../policies/workbookSurfacePolicy";
import type { ReferenceQueryBrokerPort } from "../services/referenceQueryBroker";
import { useGenericSurfaceQuery } from "./useGenericSurfaceQuery";
import type { WorkbookCommittedRecordPort } from "./WorkbookCommittedRecordPort";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

export type EntitySurfaceQueryInput = {
  readonly ordinaryCreateOwner?: WorkbookCommittedRecordPort | undefined;
  readonly editOwner?: WorkbookCommittedRecordPort | undefined;
  readonly activeViewSchemaId?: string;
  readonly referenceBroker?: ReferenceQueryBrokerPort;
  readonly hostQueryState: WorkbookQueryState;
  readonly identityQueryState: WorkbookQueryState;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly viewQuery: WorkbookViewQueryPort;
};

/** Entity conversion and partial indexing stay here; each sheet owns its read lifetime. */
export function useEntitySurfaceQuery(input: EntitySurfaceQueryInput) {
  const references = useEntityReferenceRows(
    input.referenceBroker,
    input.activeViewSchemaId === timelineViewSchemaId,
  );
  const shared = {
    ordinaryCreateOwner: input.ordinaryCreateOwner,
    committedRecordOwner: input.editOwner,
    onAuthorityUncertain: input.onAuthorityUncertain,
    viewQuery: input.viewQuery,
  };
  const hosts = useGenericSurfaceQuery({
    ...shared,
    active:
      input.activeViewSchemaId === undefined ||
      input.activeViewSchemaId === hostsViewSchemaId,
    contract: hostsContract,
    viewSchemaId: hostsViewSchemaId,
    queryState: input.hostQueryState,
  });
  const identities = useGenericSurfaceQuery({
    ...shared,
    active:
      input.activeViewSchemaId === undefined ||
      input.activeViewSchemaId === identitiesViewSchemaId,
    contract: identitiesContract,
    viewSchemaId: identitiesViewSchemaId,
    queryState: input.identityQueryState,
  });
  const hostRows = useMemo(
    () => hosts.rows.map((row) => entityRowFromApi(row, "host")),
    [hosts.rows],
  );
  const identityRows = useMemo(
    () => identities.rows.map((row) => entityRowFromApi(row, "identity")),
    [identities.rows],
  );
  const entityIndex = useMemo(
    () =>
      Object.fromEntries(
        [...hostRows, ...identityRows].map((row) => [row.recordId, row]),
      ) as Record<string, EntityRow>,
    [hostRows, identityRows],
  );
  const refreshHosts = hosts.refresh,
    refreshIdentities = identities.refresh;
  const refreshBoth = useCallback(
    async (options?: { readonly requireAcceptance?: boolean }) => {
      await Promise.all([refreshHosts(options), refreshIdentities(options)]);
    },
    [refreshHosts, refreshIdentities],
  );
  // An inactive sheet or a reference broker must not replace this sheet's reader.
  const refresh =
    input.activeViewSchemaId === undefined
      ? refreshBoth
      : input.activeViewSchemaId === hostsViewSchemaId
        ? refreshHosts
        : input.activeViewSchemaId === identitiesViewSchemaId
          ? refreshIdentities
          : input.activeViewSchemaId === timelineViewSchemaId
            ? references.refresh
            : inactiveRead;
  const invalidateHosts = hosts.invalidate,
    invalidateIdentities = identities.invalidate;
  const invalidate = useCallback(
    (reason: WorkbookQueryInvalidationReason) => {
      invalidateHosts(reason);
      invalidateIdentities(reason);
      if (
        reason.kind !== "incident_closed" &&
        reason.kind !== "collaboration_reset_required"
      )
        references.clear();
    },
    [invalidateHosts, invalidateIdentities, references.clear],
  );
  const patchHosts = hosts.applyRecordChanged,
    patchIdentities = identities.applyRecordChanged;
  const applyRecordChanged = useCallback(
    (payload: RecordChangedPayload, viewSchemaId: string) =>
      viewSchemaId === hostsViewSchemaId
        ? patchHosts(payload)
        : patchIdentities(payload),
    [patchHosts, patchIdentities],
  );
  const selected =
    input.activeViewSchemaId === identitiesViewSchemaId ? identities : hosts;
  return {
    hosts,
    identities,
    hostRows,
    identityRows,
    entityIndex,
    references,
    refresh,
    invalidate,
    applyRecordChanged,
    loadState: selected.loadState,
    browser: selected.browser,
    browsing: selected.browsing,
  };
}

async function inactiveRead() {}

const entityReferenceRequirements = [
  referenceRequirement(hostsViewSchemaId),
  referenceRequirement(identitiesViewSchemaId),
];
/** Bounded reference observations are independent of both sheets' authored queries. */
function useEntityReferenceRows(
  broker: ReferenceQueryBrokerPort | undefined,
  active: boolean,
) {
  const [hosts, setHosts] = useState<EntityRow[]>([]);
  const [identities, setIdentities] = useState<EntityRow[]>([]);
  const pending = useRef<AbortController | null>(null);
  const hasAcceptedReferences = useRef(false);
  const requestedReferences = useRef(false);
  const previousBroker = useRef(broker);
  const latestBroker = useRef(broker);
  latestBroker.current = broker;
  const clear = useCallback(() => {
    pending.current?.abort();
    pending.current = null;
    hasAcceptedReferences.current = false;
    requestedReferences.current = false;
    setHosts([]);
    setIdentities([]);
  }, []);
  const refresh = useCallback(async () => {
    requestedReferences.current = true;
    pending.current?.abort();
    const controller = new AbortController();
    pending.current = controller;
    const currentBroker = latestBroker.current;
    if (!currentBroker || !active) {
      setHosts([]);
      setIdentities([]);
      return;
    }
    try {
      const result = await currentBroker.execute(
        entityReferenceRequirements,
        controller.signal,
      );
      if (
        controller.signal.aborted ||
        pending.current !== controller ||
        latestBroker.current !== currentBroker
      )
        return;
      hasAcceptedReferences.current = true;
      setHosts(
        (
          result.find(
            (entry) => entry.requirement.viewSchemaId === hostsViewSchemaId,
          )?.rows ?? []
        ).map((row) => entityRowFromApi(row, "host")),
      );
      setIdentities(
        (
          result.find(
            (entry) =>
              entry.requirement.viewSchemaId === identitiesViewSchemaId,
          )?.rows ?? []
        ).map((row) => entityRowFromApi(row, "identity")),
      );
    } catch {
      if (controller.signal.aborted || pending.current !== controller) return;
      setHosts([]);
      setIdentities([]);
    } finally {
      if (pending.current === controller) pending.current = null;
    }
  }, [active]);
  useEffect(() => {
    const replaced = previousBroker.current !== broker;
    previousBroker.current = broker;
    // Startup authorization can replace the broker before its first read settles.
    // Complete that obligation once; an accepted reference set needs no eager read.
    if (
      replaced &&
      active &&
      requestedReferences.current &&
      (pending.current !== null || !hasAcceptedReferences.current)
    )
      void refresh();
  }, [active, broker, refresh]);
  useEffect(() => {
    if (!active) clear();
    return clear;
  }, [active, clear]);
  const index = useMemo(
    () =>
      Object.fromEntries(
        [...hosts, ...identities].map((row) => [row.recordId, row]),
      ),
    [hosts, identities],
  );
  return { hosts, identities, index, refresh, clear };
}
