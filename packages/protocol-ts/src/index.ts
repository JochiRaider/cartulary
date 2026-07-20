import type { IncidentStreamMessage } from "./generated/collaboration-types.js";
import type {
  EnvelopeMeta,
  ErrorEnvelope,
  ExtensionDiscoveryEnvelope,
  ExtensionMappingPreviewEnvelope,
  ExtensionMappingPreviewRequest,
  ExtensionMappingPreviewResource,
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
  ImportPreviewResult,
  IndicatorLinkResult,
  RejectedRowsQueryResult,
  SourceProfileList,
  TableList,
  TableMutationResult,
  TableQueryResult,
} from "./generated/network-flow-types.js";
import * as generatedProtocolValidators from "./generated/protocol-validators.js";

export type * from "./generated/network-flow-types.js";
export type {
  EnvelopeMeta,
  ErrorEnvelope,
  ExtensionDiscoveryEnvelope,
  ExtensionMappingPreviewEnvelope,
  ExtensionMappingPreviewRequest,
  ExtensionMappingPreviewResource,
  GeneratedExtensionProfileResource,
  IncidentStreamMessage,
};
export {
  networkFlowContractDescriptor,
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
};

export type NetworkFlowErrorRetryAction =
  | "correct_request"
  | "refresh_resource"
  | "restart_query"
  | "reduce_scope_or_limits"
  | "retry_with_backoff"
  | "do_not_retry";

export type NetworkFlowErrorContract = {
  readonly code: string;
  readonly http_status: number | null;
  readonly retry_action: NetworkFlowErrorRetryAction;
  readonly scope: string;
};

export type NetworkFlowErrorRegistry = {
  readonly schema_id: "cartulary.network_flow_error_contracts.v1";
  readonly errors: readonly NetworkFlowErrorContract[];
};

export function getNetworkFlowErrorRegistry(): NetworkFlowErrorRegistry {
  return parseContractArtifact<NetworkFlowErrorRegistry>(
    "contracts/network-flow/errors.v1.json",
  );
}

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
  readonly surface_kind: "built_in_sheet" | "system_view";
  readonly surface_status:
    | "required_built_in_sheet"
    | "required_system_view"
    | "standardized_optional_workbook_surface";
  readonly source_record_types: readonly string[];
  readonly required_reference_pack_keys: readonly string[];
  readonly artifact_path: string;
};

export type ViewSchemaRegistryContract = {
  readonly registry_id: string;
  readonly note?: string;
  readonly view_schemas: readonly ViewSchemaRegistryEntry[];
};

export type ExtensionProfileEntry = {
  readonly profile_id: string;
  readonly claimable: boolean;
  readonly contract_major: number | null;
  readonly route_families: readonly string[];
  readonly workspace_keys: readonly string[];
  readonly capability_ids: readonly string[];
};

export type ExtensionRegistryContract = {
  readonly schema_id: "cartulary.extension_profile_registry.v1";
  readonly profiles: readonly ExtensionProfileEntry[];
};

const extensionTokenPattern = /^[a-z][a-z0-9_]{0,63}$/u;
const extensionRoutePattern = /^\/api\/v1\/[^?#%\\]+$/u;

function requireSortedUniqueStrings(
  value: unknown,
  label: string,
  predicate: (item: string) => boolean,
): string[] {
  if (
    !Array.isArray(value) ||
    !value.every((item) => typeof item === "string")
  ) {
    throw new Error(`invalid extension discovery ${label}`);
  }
  const result = [...value];
  for (let index = 0; index < result.length; index += 1) {
    if (
      !predicate(result[index] ?? "") ||
      (index > 0 && (result[index - 1] ?? "") >= (result[index] ?? ""))
    ) {
      throw new Error(`invalid extension discovery ${label}`);
    }
  }
  return result;
}

// decodeExtensionDiscoveryItem intentionally reads only the seven known v1
// members. Unknown additive members remain inert and are never returned to callers.
export function decodeExtensionDiscoveryItem(
  value: unknown,
): GeneratedExtensionProfileResource {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error("invalid extension discovery item");
  }
  const item = value as Record<string, unknown>;
  if (
    typeof item.profile_id !== "string" ||
    !extensionTokenPattern.test(item.profile_id) ||
    typeof item.claimable !== "boolean" ||
    typeof item.claimed !== "boolean" ||
    (item.contract_major !== null &&
      (typeof item.contract_major !== "number" ||
        !Number.isSafeInteger(item.contract_major) ||
        item.contract_major < 1))
  ) {
    throw new Error("invalid extension discovery scalar");
  }
  const routeFamilies = requireSortedUniqueStrings(
    item.route_families,
    "route_families",
    (route) => extensionRoutePattern.test(route),
  );
  const workspaceKeys = requireSortedUniqueStrings(
    item.workspace_keys,
    "workspace_keys",
    (key) => extensionTokenPattern.test(key),
  );
  const capabilities = requireSortedUniqueStrings(
    item.capabilities,
    "capabilities",
    (key) => extensionTokenPattern.test(key),
  );
  if (
    capabilities.length !== 0 ||
    (item.claimable && item.contract_major === null)
  ) {
    throw new Error("invalid extension discovery capability or contract state");
  }
  return {
    profile_id: item.profile_id,
    claimable: item.claimable,
    claimed: item.claimed,
    contract_major: item.contract_major,
    route_families: routeFamilies,
    workspace_keys: workspaceKeys,
    capabilities,
  };
}

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
  tableMutationResult: createGeneratedDecoder<TableMutationResult>(
    "cartulary.network_flow.table_mutation_result.v1",
  ),
  tableQueryResult: createGeneratedDecoder<TableQueryResult>(
    "cartulary.network_flow.table_query_result.v1",
  ),
  rejectedRowsQueryResult: createGeneratedDecoder<RejectedRowsQueryResult>(
    "cartulary.network_flow.rejected_rows_query_result.v1",
  ),
  sourceProfileList: createGeneratedDecoder<SourceProfileList>(
    "cartulary.network_flow.source_profile_list.v1",
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
  importPreviewResult: createGeneratedDecoder<ImportPreviewResult>(
    "cartulary.network_flow.import_preview_result.v1",
  ),
});

const strictExtensionDiscoveryEnvelopeDecoder =
  createGeneratedDecoder<ExtensionDiscoveryEnvelope>(
    "cartulary.core_http.ExtensionDiscoveryEnvelope.v1",
  );

export const extensionDiscoveryDecoder: Decoder<ExtensionDiscoveryEnvelope> =
  Object.freeze({
    schemaId: strictExtensionDiscoveryEnvelopeDecoder.schemaId,
    decode(value: unknown): DecodeResult<ExtensionDiscoveryEnvelope> {
      try {
        if (!value || typeof value !== "object" || Array.isArray(value)) {
          throw new Error("invalid envelope");
        }
        const envelope = value as Record<string, unknown>;
        if (
          !envelope.data ||
          typeof envelope.data !== "object" ||
          Array.isArray(envelope.data)
        ) {
          throw new Error("invalid data");
        }
        const data = envelope.data as Record<string, unknown>;
        if (!Array.isArray(data.extensions)) {
          throw new Error("invalid extensions");
        }
        const sanitized: ExtensionDiscoveryEnvelope = {
          data: {
            extensions: data.extensions.map((item) =>
              decodeExtensionDiscoveryItem(item),
            ),
          },
          meta: envelope.meta as EnvelopeMeta,
        };
        return strictExtensionDiscoveryEnvelopeDecoder.decode(sanitized);
      } catch {
        return {
          ok: false,
          error: {
            boundary: "generated_protocol",
            instancePath: "",
            reasonCategory: "constraint_violation",
            schemaId: strictExtensionDiscoveryEnvelopeDecoder.schemaId,
          },
        };
      }
    },
  });

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
    "contracts/extensions/generated/profile-registry.json",
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
