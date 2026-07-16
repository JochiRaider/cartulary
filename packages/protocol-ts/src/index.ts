import type { IncidentStreamMessage } from "./generated/collaboration-types.js";
import type {
  EnvelopeMeta,
  ErrorEnvelope,
  ExtensionMappingPreviewEnvelope,
  ExtensionMappingPreviewRequest,
  ExtensionMappingPreviewResource,
  ExtensionDiscoveryEnvelope,
  ExtensionProfileResource as GeneratedExtensionProfileResource,
} from "./generated/core-http-types.js";
import {
  contractArtifactIndex,
  errorArtifacts,
  extensionArtifacts,
  openAPIArtifacts,
  viewSchemaArtifacts,
  wsArtifacts,
} from "./generated/index.js";
import { networkFlowContractDescriptor } from "./generated/network-flow-descriptor.js";
import { networkFlowMappingRegistry } from "./generated/network-flow-mapping-registry.js";
import { networkFlowPresentationRegistry } from "./generated/network-flow-presentation.js";
import type {
  GraphContributorQueryResult,
  GraphQueryResult,
  IndicatorLinkResult,
  RejectedRowsQueryResult,
  TableList,
  TableQueryResult,
} from "./generated/network-flow-types.js";
import * as generatedProtocolValidators from "./generated/protocol-validators.js";

export type * from "./generated/network-flow-types.js";
export type {
  EnvelopeMeta,
  ErrorEnvelope,
  ExtensionMappingPreviewEnvelope,
  ExtensionMappingPreviewRequest,
  ExtensionMappingPreviewResource,
  ExtensionDiscoveryEnvelope,
  GeneratedExtensionProfileResource,
  IncidentStreamMessage,
};
export {
  networkFlowContractDescriptor,
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
};

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

export const accountProtocolSchemaNames = Object.freeze({
  accountPreferencesEnvelope: "AccountPreferencesEnvelope",
  accountPreferencesPutRequest: "AccountPreferencesPutRequest",
  accountPreferencesResource: "AccountPreferencesResource",
  accountProfileEnvelope: "AccountProfileEnvelope",
  accountProfilePatchRequest: "AccountProfilePatchRequest",
  accountProfileResource: "AccountProfileResource",
  densityMode: "DensityMode",
} as const);

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

export type DensityMode = "compact" | "default" | "comfortable";

export type AccountProfileResource = {
  readonly user_id: string;
  readonly email: string;
  readonly display_name: string;
  readonly user_version: number;
  readonly created_at: string;
  readonly updated_at: string;
};

export type AccountPreferencesResource = {
  readonly user_id: string;
  readonly density_mode: DensityMode | null;
  readonly preferences_version: number;
  readonly created_at: string;
  readonly updated_at: string;
};

export type AccountProfilePatchRequest = {
  readonly base_user_version: number;
  readonly client_txn_id: string;
  readonly display_name: string;
};

export type AccountPreferencesPutRequest = {
  readonly base_preferences_version: number;
  readonly client_txn_id: string;
  readonly density_mode: DensityMode | null;
};

export type AccountProfileEnvelope = {
  readonly data: AccountProfileResource;
  readonly meta: EnvelopeMeta;
};

export type AccountPreferencesEnvelope = {
  readonly data: AccountPreferencesResource;
  readonly meta: EnvelopeMeta;
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

type GeneratedValidationError = {
  readonly instancePath?: string;
  readonly keyword?: string;
};

type GeneratedValidator = ((value: unknown) => boolean) & {
  readonly errors?: readonly GeneratedValidationError[] | null;
};

export type DecodeFailure = {
  readonly boundary: "generated_protocol";
  readonly instancePath: string;
  readonly reasonCategory:
    | "constraint_violation"
    | "invalid_format"
    | "invalid_type"
    | "invalid_value"
    | "required_member"
    | "schema_mismatch"
    | "unknown_member";
  readonly schemaId: string;
};

export type DecodeResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: DecodeFailure };

export type Decoder<T> = {
  readonly schemaId: string;
  readonly decode: (value: unknown) => DecodeResult<T>;
};

function generatedValidatorName(schemaId: string): string {
  return `validate${schemaId
    .split(/[^A-Za-z0-9]+/u)
    .filter(Boolean)
    .map((part) => `${part[0]?.toUpperCase() ?? ""}${part.slice(1)}`)
    .join("")}`;
}

function generatedValidator(schemaId: string): GeneratedValidator {
  const candidate = (
    generatedProtocolValidators as Readonly<Record<string, unknown>>
  )[generatedValidatorName(schemaId)];
  if (typeof candidate !== "function") {
    throw new Error(`missing generated protocol validator for ${schemaId}`);
  }
  return candidate as GeneratedValidator;
}

function reasonCategory(
  keyword: string | undefined,
): DecodeFailure["reasonCategory"] {
  switch (keyword) {
    case "additionalProperties":
      return "unknown_member";
    case "required":
      return "required_member";
    case "type":
      return "invalid_type";
    case "const":
    case "enum":
      return "invalid_value";
    case "format":
      return "invalid_format";
    case "oneOf":
    case "anyOf":
      return "schema_mismatch";
    default:
      return "constraint_violation";
  }
}

export function createGeneratedDecoder<T>(schemaId: string): Decoder<T> {
  const validate = generatedValidator(schemaId);
  return Object.freeze({
    schemaId,
    decode(value: unknown): DecodeResult<T> {
      if (validate(value)) {
        return { ok: true, value: value as T };
      }
      const firstError = validate.errors?.[0];
      return {
        ok: false,
        error: {
          boundary: "generated_protocol",
          instancePath: firstError?.instancePath ?? "",
          reasonCategory: reasonCategory(firstError?.keyword),
          schemaId,
        },
      };
    },
  });
}

export const networkFlowDecoders = Object.freeze({
  tableList: createGeneratedDecoder<TableList>(
    "cartulary.network_flow.table_list.v1",
  ),
  tableQueryResult: createGeneratedDecoder<TableQueryResult>(
    "cartulary.network_flow.table_query_result.v1",
  ),
  rejectedRowsQueryResult: createGeneratedDecoder<RejectedRowsQueryResult>(
    "cartulary.network_flow.rejected_rows_query_result.v1",
  ),
  graphQueryResult: createGeneratedDecoder<GraphQueryResult>(
    "cartulary.network_flow.graph_query_result.v1",
  ),
  graphContributorQueryResult:
    createGeneratedDecoder<GraphContributorQueryResult>(
      "cartulary.network_flow.graph_contributor_query_result.v1",
    ),
  indicatorLinkResult: createGeneratedDecoder<IndicatorLinkResult>(
    "cartulary.network_flow.indicator_link_result.v1",
  ),
});

export const extensionDiscoveryDecoder =
  createGeneratedDecoder<ExtensionDiscoveryEnvelope>(
    "cartulary.core_http.ExtensionDiscoveryEnvelope.v1",
  );

export const incidentStreamMessageDecoder =
  createGeneratedDecoder<IncidentStreamMessage>(
    "cartulary.ws.incident_stream_message.v1",
  );

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
