import {
  type Artifact,
  contractArtifactIndex,
  errorArtifacts,
  extensionArtifacts,
  openAPIArtifacts,
  viewSchemaArtifacts,
  wsArtifacts,
} from "./generated/index.js";

export type ContractArtifact = Artifact;

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
  readonly errors: readonly ErrorCodeEntry[];
  readonly reason_registries?: readonly ReasonCodeRegistry[];
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
