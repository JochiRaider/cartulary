import type { ChangeEvent } from "react";
import { useCallback, useRef, useState } from "react";
import type { ExtensionAvailabilityController } from "../extensions/extensionAvailability";
import {
  approveSelectAndApplyExtensionImport,
  type ExtensionImportDiscovery,
  ImportMappingPreviewStaleError,
  previewExtensionImportMapping,
  uploadAndDiscoverExtensionImport,
} from "../imports/importCoordinator";
import {
  decodeNetworkFlowImportPreviewResult,
  type NetworkFlowImportPreviewResult,
} from "../services/networkFlowContractAdapter";
import {
  buildNetworkFlowMappingCandidate,
  createNetworkFlowMappingDraft,
  type NetworkFlowMappingDraft,
  networkFlowMappingDraftReadyForPreview,
} from "./networkFlowImportModel";

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
  incidentId,
  onError,
  onImported,
  onMessage,
}: {
  readonly availability: ExtensionAvailabilityController;
  readonly apiBase: string | undefined;
  readonly canImport: boolean;
  readonly incidentId: string;
  readonly onError: (message: string | null) => void;
  readonly onImported: (tableId: string) => Promise<void>;
  readonly onMessage: (message: string) => void;
}) {
  const operationGeneration = useRef(0);
  const [stage, setStage] = useState<NetworkFlowImportStage>("idle");
  const [discovery, setDiscovery] = useState<ExtensionImportDiscovery | null>(
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
    setDiscovery(null);
    setDraft(null);
    setPreview(null);
    setPreviewCandidateKey(null);
    setStage("idle");
  }, []);

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
      if (file === null || stage !== "idle" || !canImport) {
        return;
      }
      const generation = operationGeneration.current + 1;
      operationGeneration.current = generation;
      setStage("discovering");
      onError(null);
      try {
        const nextDiscovery = await uploadAndDiscoverExtensionImport({
          availability,
          apiBase,
          incidentId,
          file,
          transactionPrefix: "nf-import",
          onProgress: onMessage,
        });
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
      }
    },
    [availability, apiBase, canImport, incidentId, onError, onMessage, stage],
  );

  const requestPreview = useCallback(async () => {
    if (
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
    setStage("previewing");
    setPreview(null);
    setPreviewCandidateKey(null);
    onError(null);
    try {
      const resource = await previewExtensionImportMapping<unknown>({
        availability,
        apiBase,
        discovery,
        candidate: extensionCandidate(candidate),
      });
      if (generation !== operationGeneration.current) {
        return;
      }
      if (
        resource.schema_id !==
          "cartulary.imports.extension_mapping_preview_result.v1" ||
        resource.import_session_id !== discovery.sessionId ||
        resource.import_unit_id !== discovery.unit.import_unit_id ||
        resource.target_kind !== "network_flow_table" ||
        resource.extension_profile_id !== "network_flow_activity" ||
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
    }
  }, [availability, apiBase, discovery, draft, onError, onMessage, stage]);

  const apply = useCallback(async () => {
    if (discovery === null || draft === null || preview === null) {
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
    setStage("applying");
    onError(null);
    try {
      const refs = await approveSelectAndApplyExtensionImport({
        availability,
        apiBase,
        discovery,
        candidate: extensionCandidate(candidate),
        expectedMappingFingerprint: preview.mapping_fingerprint,
        transactionPrefix: "nf-import",
        onProgress: onMessage,
      });
      if (generation !== operationGeneration.current) {
        return;
      }
      const importedTable = refs.find(
        (resource) =>
          resource.kind === "network_flow_table" && resource.id.trim() !== "",
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
    }
  }, [
    availability,
    apiBase,
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
    targetKind: "network_flow_table",
    extensionProfileId: "network_flow_activity",
    ownerMappingSchemaId: "cartulary.network_flow.mapping_candidate.v1",
    ownerMapping: { ...candidate },
  } as const;
}

function candidateKey(
  candidate: ReturnType<typeof buildNetworkFlowMappingCandidate>,
): string {
  return JSON.stringify(candidate);
}

function importErrorMessage(caught: unknown): string {
  return caught instanceof Error ? caught.message : "import_failed";
}
