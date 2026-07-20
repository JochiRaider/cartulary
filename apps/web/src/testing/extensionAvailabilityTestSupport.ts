import { ExtensionAvailabilityController } from "../extensions/extensionAvailability";

export function readyExtensionAvailability(
  incidentId = "incident-1",
): ExtensionAvailabilityController {
  const controller = new ExtensionAvailabilityController({
    incidentId,
    randomValues: (bytes) => {
      bytes.fill(0xa5);
      return bytes;
    },
  });
  controller.setDiscovery([
    {
      profile_id: "network_flow_activity",
      claimed: true,
      contract_major: 2,
      route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
      workspace_keys: ["network_analysis"],
      capabilities: [],
    },
  ]);
  const tag = controller.reserve();
  if (
    tag === null ||
    !controller.acceptWorkbookStartup(tag, {
      schema_id: "cartulary.extension_workspace_availability.v1",
      incident_id: incidentId,
      workspaces: [
        {
          extension_profile_id: "network_flow_activity",
          workspace_key: "network_analysis",
        },
      ],
    })
  ) {
    throw new Error("extension availability test setup failed");
  }
  return controller;
}
