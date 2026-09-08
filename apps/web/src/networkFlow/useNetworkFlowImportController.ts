import type { ChangeEvent } from "react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  boundedImportRead,
  loadImportResources,
  observeImportJob,
  submitImportAttempt,
} from "../imports/importCoordinator";
import {
  captureImportWrite,
  captureWorkbookUpload,
} from "../imports/importRequests";
import { clientTxnID } from "../services/browserApi";
import {
  ImportClient,
  type ImportReadResult,
  type ImportWriteReceipt,
  importFailureMessage,
} from "../services/importClient";
import type { ImportJobResource } from "../services/importContractAdapter";
import { importSessionIdFromReceipt } from "../services/importJobContract";
import { requireClaimGatedAnalyticalImportTarget } from "../services/importTargetContractAdapter";
import {
  decodeNetworkFlowImportPreviewResult,
  type NetworkFlowImportPreviewResult,
  networkFlowMappingCandidateSchemaId,
  networkFlowMappingMetadata,
} from "../services/networkFlowContractAdapter";
import {
  buildNetworkFlowMappingCandidate,
  createNetworkFlowMappingDraft,
  type NetworkFlowImportDiscovery,
  type NetworkFlowMappingDraft,
  networkFlowApprovalRequest,
  networkFlowApprovedPreviewMatches,
  networkFlowMappingDraftReadyForPreview,
} from "./networkFlowImportModel";

const networkFlowImportTarget = requireClaimGatedAnalyticalImportTarget(
  networkFlowMappingMetadata.target_kind,
  networkFlowMappingMetadata.profile_id,
);

export type NetworkFlowImportStage =
  | "idle"
  | "discovering"
  | "ready"
  | "previewing"
  | "applying";

export function useNetworkFlowImportController({
  availability,
  apiBase,
  canImport,
  actorId,
  incidentId,
  onError,
  onImported,
  onMessage,
}: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase: string | undefined;
  readonly canImport: boolean;
  readonly actorId: string | null;
  readonly incidentId: string;
  readonly onError: (message: string | null) => void;
  readonly onImported: (tableId: string) => Promise<void>;
  readonly onMessage: (message: string) => void;
}) {
  const operationGeneration = useRef(0);
  const active = useRef<AbortController | null>(null);
  const knownJob = useRef<ImportJobResource | null>(null);
  const admission = useRef(false);
  const current = useRef({ canImport, actorId });
  current.current = { canImport, actorId };
  const client = useMemo(
    () => new ImportClient({ availability, apiBase, incidentId }),
    [availability, apiBase, incidentId],
  );
  const scope = useMemo(
    () => ({
      incidentId,
      actorId: actorId ?? "",
      lifetime: availability.clientInstanceId,
    }),
    [incidentId, actorId, availability],
  );
  const requireCurrent = useCallback(
    (generation: number) => {
      if (
        operationGeneration.current !== generation ||
        !current.current.canImport ||
        current.current.actorId !== scope.actorId
      )
        throw new Error("Import access changed. Review current access.");
    },
    [scope],
  );
  const submit = useCallback(
    async (
      attempt: Parameters<ImportClient["send"]>[0],
      generation: number,
      signal: AbortSignal,
    ): Promise<ImportWriteReceipt> => {
      requireCurrent(generation);
      const result = await submitImportAttempt({
        signal,
        upload: attempt.kind === "upload",
        send: (requestSignal) => {
          requireCurrent(generation);
          return client.send(attempt, requestSignal);
        },
      });
      requireCurrent(generation);
      if (result.kind !== "accepted")
        throw new ImportStageFailure(importFailureMessage(result.failure));
      if (result.receipt.kind === "job") knownJob.current = result.receipt.job;
      return result.receipt;
    },
    [client, requireCurrent],
  );
  const observe = useCallback(
    async (
      job: ImportJobResource,
      signal: AbortSignal,
      generation: number,
      sessionId?: string,
    ) => {
      const outcome = await observeImportJob({
        initial: job,
        signal,
        read: (id, requestSignal) =>
          client.readJob(
            id,
            requestSignal,
            sessionId,
            sessionId ? "apply" : "discovery",
          ),
        onJob: (next) => {
          if (operationGeneration.current === generation)
            knownJob.current = next;
        },
      });
      requireCurrent(generation);
      if (outcome.kind === "paused")
        throw new ImportStageFailure(importFailureMessage(outcome.failure));
      if (outcome.job.status !== "succeeded")
        throw new ImportStageFailure(
          "Import job ended without success. Committed units remain committed.",
        );
      return outcome.job;
    },
    [client, requireCurrent],
  );
  const [stage, setStage] = useState<NetworkFlowImportStage>("idle");
  const [discovery, setDiscovery] = useState<NetworkFlowImportDiscovery | null>(
    null,
  );
  const [draft, setDraft] = useState<NetworkFlowMappingDraft | null>(null);
  const [preview, setPreview] = useState<NetworkFlowImportPreviewResult | null>(
    null,
  );
  const [previewCandidateKey, setPreviewCandidateKey] = useState<string | null>(
    null,
  );

  const reset = useCallback(() => {
    operationGeneration.current += 1;
    active.current?.abort();
    active.current = null;
    admission.current = false;
    knownJob.current = null;
    setDiscovery(null);
    setDraft(null);
    setPreview(null);
    setPreviewCandidateKey(null);
    setStage("idle");
  }, []);

  useEffect(() => {
    if (!scope.actorId) reset();
    return () => {
      operationGeneration.current++;
      active.current?.abort();
    };
  }, [scope, reset]);

  const updateDraft = useCallback(
    (
      update:
        | NetworkFlowMappingDraft
        | ((current: NetworkFlowMappingDraft) => NetworkFlowMappingDraft),
    ) => {
      setDraft((current) => {
        if (current === null) {
          return current;
        }
        return typeof update === "function" ? update(current) : update;
      });
      setPreview(null);
      setPreviewCandidateKey(null);
      setStage("ready");
    },
    [],
  );

  const handleImportChange = useCallback(
    async (event: ChangeEvent<HTMLInputElement>) => {
      const file = event.target.files?.[0] ?? null;
      event.target.value = "";
      if (
        file === null ||
        stage !== "idle" ||
        !canImport ||
        actorId === null ||
        admission.current
      ) {
        return;
      }
      const generation = operationGeneration.current + 1;
      operationGeneration.current = generation;
      admission.current = true;
      const stop = new AbortController();
      active.current = stop;
      setStage("discovering");
      onError(null);
      try {
        const receipt = await submit(
          captureWorkbookUpload(scope, file, clientTxnID("nf-import-upload")),
          generation,
          stop.signal,
        );
        if (receipt.kind !== "job") throw new Error("Invalid upload receipt");
        const job = await observe(receipt.job, stop.signal, generation);
        const sessionId = importSessionIdFromReceipt(job);
        if (!sessionId) throw new Error("Missing import session");
        const resources = requireRead(
          await loadImportResources(client, sessionId, stop.signal),
        );
        const unit = resources.units[0];
        if (!unit)
          throw new ImportStageFailure("No import units were discovered.");
        const preview = requireRead(
          await boundedImportRead(
            (signal) => client.preview(unit, signal),
            stop.signal,
          ),
        );
        const nextDiscovery = { sessionId, unit, preview };
        if (generation !== operationGeneration.current) {
          return;
        }
        setDiscovery(nextDiscovery);
        setDraft(createNetworkFlowMappingDraft(nextDiscovery.preview.columns));
        setStage("ready");
        onMessage("Review the discovered mapping before approval.");
      } catch (caught) {
        if (generation === operationGeneration.current) {
          setStage("idle");
          onError(importErrorMessage(caught));
        }
      } finally {
        if (generation === operationGeneration.current)
          admission.current = false;
      }
    },
    [
      actorId,
      client,
      scope,
      canImport,
      onError,
      onMessage,
      stage,
      submit,
      observe,
    ],
  );

  const requestPreview = useCallback(async () => {
    if (
      admission.current ||
      !canImport ||
      discovery === null ||
      draft === null ||
      stage !== "ready" ||
      !networkFlowMappingDraftReadyForPreview(draft)
    ) {
      return;
    }
    const generation = operationGeneration.current + 1;
    operationGeneration.current = generation;
    const candidate = buildNetworkFlowMappingCandidate(
      draft,
      discovery.preview.columns,
    );
    admission.current = true;
    const stop = new AbortController();
    active.current = stop;
    setStage("previewing");
    setPreview(null);
    setPreviewCandidateKey(null);
    onError(null);
    try {
      const resource = requireRead(
        await boundedImportRead(
          (signal) =>
            client.previewMapping(
              discovery.unit,
              extensionCandidate(candidate),
              signal,
            ),
          stop.signal,
        ),
      );
      if (generation !== operationGeneration.current) {
        return;
      }
      if (
        resource.schema_id !==
          "cartulary.imports.extension_mapping_preview_result.v1" ||
        resource.import_session_id !== discovery.sessionId ||
        resource.import_unit_id !== discovery.unit.import_unit_id ||
        resource.target_kind !== networkFlowImportTarget.target_kind ||
        resource.extension_profile_id !==
          networkFlowImportTarget.extension_profile_id ||
        resource.owner_result_schema_id !==
          "cartulary.network_flow.import_preview_result.v1"
      ) {
        throw new Error("invalid_import_mapping_preview_wrapper");
      }
      const ownerResult = decodeNetworkFlowImportPreviewResult(
        resource.owner_result,
      );
      setPreview(ownerResult);
      setPreviewCandidateKey(candidateKey(candidate));
      setStage("ready");
      onMessage("Mapping preview is ready for explicit approval.");
    } catch (caught) {
      if (generation === operationGeneration.current) {
        setStage("ready");
        onError(importErrorMessage(caught));
      }
    } finally {
      if (generation === operationGeneration.current) admission.current = false;
    }
  }, [client, canImport, discovery, draft, onError, onMessage, stage]);

  const apply = useCallback(async () => {
    if (
      admission.current ||
      !canImport ||
      discovery === null ||
      draft === null ||
      preview === null
    ) {
      return;
    }
    const candidate = buildNetworkFlowMappingCandidate(
      draft,
      discovery.preview.columns,
    );
    if (
      stage !== "ready" ||
      previewCandidateKey === null ||
      previewCandidateKey !== candidateKey(candidate)
    ) {
      setPreview(null);
      setPreviewCandidateKey(null);
      onError("Mapping changed. Generate a new preview before applying.");
      return;
    }
    const generation = operationGeneration.current + 1;
    operationGeneration.current = generation;
    admission.current = true;
    const stop = new AbortController();
    active.current = stop;
    setStage("applying");
    onError(null);
    try {
      const approved = await submit(
        captureImportWrite({
          kind: "mapping",
          scope,
          sessionId: discovery.sessionId,
          unitId: discovery.unit.import_unit_id,
          body: networkFlowApprovalRequest(
            discovery,
            extensionCandidate(candidate),
            clientTxnID("nf-import-mapping"),
          ),
        }),
        generation,
        stop.signal,
      );
      if (approved.kind !== "unit") throw new Error("Invalid mapping receipt");
      if (
        !networkFlowApprovedPreviewMatches(
          approved.unit,
          preview.mapping_fingerprint,
        )
      )
        throw new ImportMappingPreviewStaleError();
      setDiscovery({ ...discovery, unit: approved.unit });
      await submit(
        captureImportWrite({
          kind: "select",
          scope,
          sessionId: discovery.sessionId,
          unitId: discovery.unit.import_unit_id,
          body: { client_txn_id: clientTxnID("nf-import-selection") },
        }),
        generation,
        stop.signal,
      );
      const receipt = await submit(
        captureImportWrite({
          kind: "apply",
          scope,
          sessionId: discovery.sessionId,
          body: {
            client_txn_id: clientTxnID("nf-import-apply"),
            selected_unit_ids: [discovery.unit.import_unit_id],
          },
        }),
        generation,
        stop.signal,
      );
      if (receipt.kind !== "job") throw new Error("Invalid apply receipt");
      const applied = await observe(
        receipt.job,
        stop.signal,
        generation,
        discovery.sessionId,
      );
      const refs = applied.result_summary?.resource_refs ?? [];
      if (generation !== operationGeneration.current) {
        return;
      }
      const importedTable = refs.find(
        (resource) =>
          resource.kind === networkFlowImportTarget.target_kind &&
          resource.id.trim() !== "",
      );
      if (importedTable === undefined) {
        throw new Error("network_flow_table_not_returned");
      }
      await onImported(importedTable.id);
      onMessage("Import applied.");
      reset();
    } catch (caught) {
      if (generation !== operationGeneration.current) {
        return;
      }
      setStage("ready");
      if (caught instanceof ImportMappingPreviewStaleError) {
        setPreview(null);
        setPreviewCandidateKey(null);
        onError(
          "The approved mapping no longer matches this preview. Review and preview it again.",
        );
        return;
      }
      onError(importErrorMessage(caught));
    } finally {
      if (generation === operationGeneration.current) admission.current = false;
    }
  }, [
    canImport,
    scope,
    submit,
    observe,
    discovery,
    draft,
    onError,
    onImported,
    onMessage,
    preview,
    previewCandidateKey,
    reset,
    stage,
  ]);

  const currentCandidateKey =
    discovery === null || draft === null
      ? null
      : candidateKey(
          buildNetworkFlowMappingCandidate(draft, discovery.preview.columns),
        );
  return {
    apply,
    canApply:
      stage === "ready" &&
      preview !== null &&
      previewCandidateKey !== null &&
      previewCandidateKey === currentCandidateKey,
    discovery,
    draft,
    handleImportChange,
    importing: stage === "discovering" || stage === "applying",
    mappingOpen: discovery !== null && draft !== null,
    preview,
    requestPreview,
    reset,
    stage,
    updateDraft,
  };
}

function extensionCandidate(
  candidate: ReturnType<typeof buildNetworkFlowMappingCandidate>,
) {
  return {
    target_kind: networkFlowImportTarget.target_kind,
    extension_profile_id: networkFlowImportTarget.extension_profile_id,
    owner_mapping_schema_id: networkFlowMappingCandidateSchemaId,
    owner_mapping: { ...candidate },
  } as const;
}

function candidateKey(
  candidate: ReturnType<typeof buildNetworkFlowMappingCandidate>,
): string {
  return JSON.stringify(candidate);
}

function importErrorMessage(caught: unknown): string {
  return caught instanceof ImportStageFailure
    ? caught.message
    : "The import response could not be validated. Review the current import before continuing.";
}

class ImportStageFailure extends Error {}
class ImportMappingPreviewStaleError extends Error {}
function requireRead<T>(result: ImportReadResult<T>): T {
  if (result.kind === "received") return result.value;
  throw new ImportStageFailure(importFailureMessage(result.failure));
}
