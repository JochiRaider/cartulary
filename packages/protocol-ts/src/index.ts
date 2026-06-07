import {
  contractArtifactIndex,
  errorArtifacts,
  extensionArtifacts,
  openAPIArtifacts,
  viewSchemaArtifacts,
  wsArtifacts,
} from "./generated/index.js";

export type ContractArtifact = {
  readonly path: string;
  readonly json: string;
  readonly sha256: string;
};

export type ErrorCodeEntry = {
  readonly code: string;
  readonly http_status: number;
  readonly summary: string;
};

export type ReasonCodeEntry = {
  readonly code: string;
  readonly summary: string;
};

export type ReasonCodeRegistry = {
  readonly error_code: string;
  readonly reason_codes: readonly ReasonCodeEntry[];
};

export type ErrorRegistryContract = {
  readonly registry_id: string;
  readonly note?: string;
  readonly errors: readonly ErrorCodeEntry[];
  readonly reason_registries?: readonly ReasonCodeRegistry[];
};

export type ViewSchemaRegistryEntry = {
  readonly view_schema_id: string;
  readonly title: string;
  readonly surface_kind: string;
  readonly source_record_types: readonly string[];
  readonly artifact_path: string;
};

export type ViewSchemaRegistryContract = {
  readonly registry_id: string;
  readonly note?: string;
  readonly view_schemas: readonly ViewSchemaRegistryEntry[];
};

export type ExtensionProfileEntry = {
  readonly profile_id: string;
  readonly route_families: readonly string[];
};

export type ExtensionRegistryContract = {
  readonly registry_id: string;
  readonly note?: string;
  readonly profiles: readonly ExtensionProfileEntry[];
};

export const evidenceProtocolSchemaNames = Object.freeze({
  envelopeMeta: "EnvelopeMeta",
  errorEnvelope: "ErrorEnvelope",
  evidenceAttachBlobEnvelope: "EvidenceAttachBlobEnvelope",
  evidenceAttachBlobRequest: "EvidenceAttachBlobRequest",
  evidenceHandleEnvelope: "EvidenceHandleEnvelope",
  evidenceHandleIssueRequest: "EvidenceHandleIssueRequest",
  objectBlobCreateEnvelope: "ObjectBlobCreateEnvelope",
  objectBlobCreateRequest: "ObjectBlobCreateRequest",
  objectBlobUploadTarget: "ObjectBlobUploadTarget",
} as const);

export type EnvelopeMeta = {
  readonly request_id: string;
  readonly paging?: unknown;
};

export type ErrorEnvelope = {
  readonly error: {
    readonly code: string;
    readonly conflict?: Record<string, unknown>;
    readonly details: Record<string, unknown>;
    readonly message: string;
    readonly request_id: string;
    readonly retryable: boolean;
    readonly status: number;
  };
};

export type ViewCell = {
  readonly value?: unknown;
  readonly [key: string]: unknown;
};

export type ViewRow = {
  readonly cells: Record<string, ViewCell>;
  readonly record_id: string;
  readonly row_version: number;
  readonly [key: string]: unknown;
};

export type ViewMutationData = {
  readonly change_set_id: string;
  readonly row: ViewRow;
  readonly view_schema_id: string;
};

export type ObjectBlobCreateRequest = {
  readonly byte_size: number;
  readonly client_txn_id: string;
  readonly content_type_hint?: string | null;
  readonly filename_hint?: string | null;
  readonly incident_id: string;
  readonly sha256_hex?: string | null;
};

export type ObjectBlobUploadTarget = {
  readonly expires_at: string;
  readonly headers: Record<string, string>;
  readonly href: string;
  readonly method: "PUT";
};

export type ObjectBlobCreateEnvelope = {
  readonly data: {
    readonly accepted_contract: {
      readonly byte_size: number;
      readonly content_type_hint: string | null;
      readonly filename_hint: string | null;
      readonly incident_id: string;
      readonly sha256_hex: string | null;
    };
    readonly incident_id: string;
    readonly object_blob_id: string;
    readonly pending_expires_at: string;
    readonly target_expires_at: string;
    readonly upload_state: "pending";
    readonly upload_target: ObjectBlobUploadTarget;
  };
  readonly meta: EnvelopeMeta;
};

export type EvidenceAttachBlobRequest = {
  readonly base_row_version: number;
  readonly client_txn_id: string;
  readonly object_blob_id: string;
};

export type EvidenceAttachBlobEnvelope = {
  readonly data: ViewMutationData & {
    readonly object_blob_id: string;
  };
  readonly meta: EnvelopeMeta;
};

export type EvidenceHandleIssueRequest = Record<string, never>;

export type EvidenceHandleEnvelope = {
  readonly data: {
    readonly content_type: string;
    readonly disposition: "inline" | "attachment";
    readonly evidence_lifecycle_state: string;
    readonly expires_at: string;
    readonly filename: string;
    readonly handle_kind: "preview" | "download";
    readonly href: string;
    readonly incident_id: string;
    readonly media_class: string;
    readonly method: "GET";
    readonly object_blob_id: string;
    readonly preview_kind?: string;
    readonly record_id: string;
    readonly sha256: string | null;
    readonly single_use: boolean;
    readonly size_bytes: number;
    readonly upload_state: string;
  };
  readonly meta: EnvelopeMeta;
};

const openAPIArtifactList = Object.freeze([...openAPIArtifacts]);
const wsArtifactList = Object.freeze([...wsArtifacts]);
const viewSchemaArtifactList = Object.freeze([...viewSchemaArtifacts]);
const errorArtifactList = Object.freeze([...errorArtifacts]);
const extensionArtifactList = Object.freeze([...extensionArtifacts]);
const contractArtifactLists = Object.freeze({
  openAPIArtifacts: openAPIArtifactList,
  wsArtifacts: wsArtifactList,
  viewSchemaArtifacts: viewSchemaArtifactList,
  errorArtifacts: errorArtifactList,
  extensionArtifacts: extensionArtifactList,
});

export function listOpenAPIArtifacts(): readonly ContractArtifact[] {
  return openAPIArtifactList;
}

export function listWSArtifacts(): readonly ContractArtifact[] {
  return wsArtifactList;
}

export function listViewSchemaArtifacts(): readonly ContractArtifact[] {
  return viewSchemaArtifactList;
}

export function listErrorArtifacts(): readonly ContractArtifact[] {
  return errorArtifactList;
}

export function listExtensionArtifacts(): readonly ContractArtifact[] {
  return extensionArtifactList;
}

export function listContractArtifactFamilies(): Readonly<{
  openAPIArtifacts: readonly ContractArtifact[];
  wsArtifacts: readonly ContractArtifact[];
  viewSchemaArtifacts: readonly ContractArtifact[];
  errorArtifacts: readonly ContractArtifact[];
  extensionArtifacts: readonly ContractArtifact[];
}> {
  return contractArtifactLists;
}

export function getContractArtifact(
  artifactPath: string,
): ContractArtifact | undefined {
  return contractArtifactIndex[artifactPath];
}

export function requireContractArtifact(
  artifactPath: string,
): ContractArtifact {
  const artifact = getContractArtifact(artifactPath);
  if (!artifact) {
    throw new Error(`missing contract artifact ${artifactPath}`);
  }
  return artifact;
}

export function requireContractArtifactJSON(artifactPath: string): string {
  return requireContractArtifact(artifactPath).json;
}

export function parseContractArtifact<T>(artifactPath: string): T {
  return JSON.parse(requireContractArtifactJSON(artifactPath)) as T;
}

export function getErrorRegistryContract(): ErrorRegistryContract {
  return parseContractArtifact<ErrorRegistryContract>(
    "contracts/errors/index.json",
  );
}

export function getViewSchemaRegistryContract(): ViewSchemaRegistryContract {
  return parseContractArtifact<ViewSchemaRegistryContract>(
    "contracts/view-schemas/index.json",
  );
}

export function listViewSchemaRegistryEntries(): readonly ViewSchemaRegistryEntry[] {
  return Object.freeze([...getViewSchemaRegistryContract().view_schemas]);
}

export function getViewSchemaRegistryEntry(
  viewSchemaId: string,
): ViewSchemaRegistryEntry | undefined {
  return listViewSchemaRegistryEntries().find(
    (entry) => entry.view_schema_id === viewSchemaId,
  );
}

export function requireViewSchemaRegistryEntry(
  viewSchemaId: string,
): ViewSchemaRegistryEntry {
  const entry = getViewSchemaRegistryEntry(viewSchemaId);
  if (!entry) {
    throw new Error(`missing view-schema registry entry for ${viewSchemaId}`);
  }
  return entry;
}

export function getExtensionRegistryContract(): ExtensionRegistryContract {
  return parseContractArtifact<ExtensionRegistryContract>(
    "contracts/extensions/index.json",
  );
}

export function listExtensionProfiles(): readonly ExtensionProfileEntry[] {
  return Object.freeze([...getExtensionRegistryContract().profiles]);
}

export function getExtensionProfile(
  profileId: string,
): ExtensionProfileEntry | undefined {
  return listExtensionProfiles().find(
    (profile) => profile.profile_id === profileId,
  );
}

export function requireExtensionProfile(
  profileId: string,
): ExtensionProfileEntry {
  const profile = getExtensionProfile(profileId);
  if (!profile) {
    throw new Error(`missing extension profile for ${profileId}`);
  }
  return profile;
}

export function listReasonCodeRegistries(): readonly ReasonCodeRegistry[] {
  return Object.freeze([
    ...(getErrorRegistryContract().reason_registries ?? []),
  ]);
}

export function getReasonCodeRegistry(
  errorCode: string,
): ReasonCodeRegistry | undefined {
  return listReasonCodeRegistries().find(
    (registry) => registry.error_code === errorCode,
  );
}

export function requireReasonCodeRegistry(
  errorCode: string,
): ReasonCodeRegistry {
  const registry = getReasonCodeRegistry(errorCode);
  if (!registry) {
    throw new Error(`missing reason-code registry for ${errorCode}`);
  }
  return registry;
}
