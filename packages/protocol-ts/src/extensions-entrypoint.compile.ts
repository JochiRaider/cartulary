import {
  type ExtensionClientSupportRegistry,
  type ExtensionProfile,
  extensionClientSupportRegistry,
  extensionProfileRegistry,
} from "@cartulary/protocol-ts/extensions";

export const typedClientSupportRegistry: ExtensionClientSupportRegistry =
  extensionClientSupportRegistry;
export const firstExtensionProfile: ExtensionProfile =
  extensionProfileRegistry.profiles[0];
