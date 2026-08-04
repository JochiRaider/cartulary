import { extensionClientSupportRegistry as generatedExtensionClientSupportRegistry } from "../generated/extension-client-support-registry.js";
import { extensionProfileRegistry } from "../generated/extension-profile-registry.js";

type ExtensionClientSupportProfile = {
  readonly capability_ids: readonly string[];
  readonly profile_id: string;
  readonly public_schema_ids: readonly string[];
  readonly supported_contract_majors: readonly number[];
  readonly workspace_keys: readonly string[];
};
export type ExtensionClientSupportRegistry = {
  readonly asset_set_sha256: string;
  readonly client_build_class: "standard";
  readonly client_build_id: string;
  readonly profiles: readonly ExtensionClientSupportProfile[];
  readonly schema_id: "cartulary.client_extension_support_registry.v1";
};
export type ExtensionProfile =
  (typeof extensionProfileRegistry)["profiles"][number];

export const extensionClientSupportRegistry: ExtensionClientSupportRegistry =
  generatedExtensionClientSupportRegistry;

export { extensionProfileRegistry };
