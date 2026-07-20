import {
  type GeneratedExtensionProfileResource,
  parseContractArtifact as parseGeneratedContractArtifact,
} from "@cartulary/protocol-ts";

export type { GeneratedExtensionProfileResource };

export function parseExtensionContractArtifact<T>(path: string): T {
  return parseGeneratedContractArtifact<T>(path);
}
