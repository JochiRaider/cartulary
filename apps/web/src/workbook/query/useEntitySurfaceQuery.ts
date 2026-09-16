import { requireViewContract } from "@cartulary/view-contracts";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { boundedRead } from "../../services/asyncObservation";
import type { RecordChangedPayload } from "../collaboration/workbookCollaborationMessages";
import type { WorkbookQueryInvalidationReason } from "../lifecycle/workbookInvalidation";
import {
  type EntityRow,
  entityRowFromApi,
} from "../models/entityWorkbookModel";
import {
  emptyWorkbookQueryState,
  type WorkbookQueryState,
} from "../models/workbookQuery";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import { useGenericSurfaceQuery } from "./useGenericSurfaceQuery";
import type { WorkbookCommittedRecordPort } from "./WorkbookCommittedRecordPort";
import type { WorkbookViewQueryPort } from "./WorkbookViewQueryPort";

const hostsContract = requireViewContract(hostsViewSchemaId);
const identitiesContract = requireViewContract(identitiesViewSchemaId);

export type EntitySurfaceQueryInput = {
  readonly ordinaryCreateOwner?: WorkbookCommittedRecordPort | undefined;
  readonly editOwner?: WorkbookCommittedRecordPort | undefined;
  readonly activeViewSchemaId?: string;
  readonly hostQueryState: WorkbookQueryState;
  readonly identityQueryState: WorkbookQueryState;
  readonly onAuthorityUncertain: (() => void) | undefined;
  readonly viewQuery: WorkbookViewQueryPort;
};

/** Entity conversion and partial indexing stay here; each sheet owns its read lifetime. */
export function useEntitySurfaceQuery(input: EntitySurfaceQueryInput) {
  const references = useEntityReferenceRows(
    input.viewQuery,
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
  // An inactive sheet or a reference observation must not replace this sheet's reader.
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

async function inactiveRead(options?: {
  readonly requireAcceptance?: boolean;
}) {
  if (options?.requireAcceptance)
    throw new Error("Entity refresh requires an active reader.");
}

/** Bounded reference observations are independent of both sheets' authored queries. */
function useEntityReferenceRows(
  reader: WorkbookViewQueryPort,
  active: boolean,
) {
  const [hosts, setHosts] = useState<EntityRow[]>([]);
  const [identities, setIdentities] = useState<EntityRow[]>([]);
  const pending = useRef<AbortController | null>(null);
  const hasAcceptedReferences = useRef(false);
  const requestedReferences = useRef(false);
  const previousReader = useRef(reader);
  const latestReader = useRef(reader);
  latestReader.current = reader;
  const clear = useCallback(() => {
    pending.current?.abort();
    pending.current = null;
    hasAcceptedReferences.current = false;
    requestedReferences.current = false;
    setHosts([]);
    setIdentities([]);
  }, []);
  const refresh = useCallback(
    async (options?: { readonly requireAcceptance?: boolean }) => {
      requestedReferences.current = true;
      pending.current?.abort();
      const controller = new AbortController();
      pending.current = controller;
      const currentReader = latestReader.current;
      if (!currentReader || !active) {
        setHosts([]);
        setIdentities([]);
        if (options?.requireAcceptance)
          throw new Error("Entity references are unavailable.");
        return;
      }
      try {
        const result = await Promise.all(
          [hostsContract, identitiesContract].map((contract) =>
            boundedRead(
              (signal) =>
                currentReader.query({
                  contract,
                  queryState: emptyWorkbookQueryState(),
                  limit: 100,
                  signal,
                }),
              controller.signal,
            ),
          ),
        );
        if (
          controller.signal.aborted ||
          pending.current !== controller ||
          latestReader.current !== currentReader
        ) {
          if (options?.requireAcceptance)
            throw new Error("Entity reference refresh was superseded.");
          return;
        }
        hasAcceptedReferences.current = result.every(
          (item) => item.kind === "accepted",
        );
        const [hostResult, identityResult] = result;
        setHosts(
          hostResult?.kind === "accepted"
            ? hostResult.value.rows.map((row) => entityRowFromApi(row, "host"))
            : [],
        );
        setIdentities(
          identityResult?.kind === "accepted"
            ? identityResult.value.rows.map((row) =>
                entityRowFromApi(row, "identity"),
              )
            : [],
        );
        if (options?.requireAcceptance && !hasAcceptedReferences.current)
          throw new Error("Entity reference refresh was not accepted.");
      } catch (error) {
        if (!controller.signal.aborted && pending.current === controller) {
          setHosts([]);
          setIdentities([]);
        }
        if (options?.requireAcceptance) throw error;
      } finally {
        if (pending.current === controller) pending.current = null;
      }
    },
    [active],
  );
  useEffect(() => {
    const replaced = previousReader.current !== reader;
    previousReader.current = reader;
    // Startup authorization can replace the reader before its first read settles.
    // Complete that obligation once; an accepted reference set needs no eager read.
    if (
      replaced &&
      active &&
      requestedReferences.current &&
      (pending.current !== null || !hasAcceptedReferences.current)
    )
      void refresh();
  }, [active, reader, refresh]);
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
