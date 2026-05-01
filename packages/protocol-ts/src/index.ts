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
