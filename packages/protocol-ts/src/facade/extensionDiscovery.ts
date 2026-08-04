import type {
  EnvelopeMeta,
  ExtensionDiscoveryEnvelope,
  ExtensionProfileResource as GeneratedExtensionProfileResource,
} from "../generated/core-http-types.js";

import {
  createGeneratedDecoder,
  type DecodeResult,
  type Decoder,
} from "./runtimeValidation.js";

export type { ExtensionDiscoveryEnvelope, GeneratedExtensionProfileResource };

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

// Unknown additive members remain inert and are never returned to callers.
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
