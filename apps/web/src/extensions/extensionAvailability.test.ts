import { describe, expect, it, vi } from "vitest";
import {
  decodeClientExtensionSupportRegistry,
  decodeExtensionWorkspaceAvailability,
  ExtensionAvailabilityController,
  ExtensionAvailabilityUnavailableError,
  extensionCapabilityActivationFailure,
  packagedClientExtensionSupportRegistry,
} from "./extensionAvailability";

const discovery = [
  {
    profile_id: "network_flow_activity",
    claimable: true,
    claimed: true,
    contract_major: 2,
    route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
    workspace_keys: ["network_analysis"],
    capabilities: [],
  },
] as const;

const availability = {
  schema_id: "cartulary.extension_workspace_availability.v1",
  incident_id: "incident-1",
  workspaces: [
    {
      extension_profile_id: "network_flow_activity",
      workspace_key: "network_analysis",
    },
  ],
};

function deterministicRandom(fill: number) {
  return (bytes: Uint8Array) => {
    bytes.fill(fill);
    return bytes;
  };
}

describe("extension availability lifecycle", () => {
  it("loads the generated standard support registry and rejects capability facts", () => {
    const support = packagedClientExtensionSupportRegistry();
    expect(support).not.toBeNull();
    expect(support?.client_build_class).toBe("standard");
    expect(support?.profiles).toEqual([
      expect.objectContaining({
        profile_id: "import",
        supported_contract_majors: [1],
        workspace_keys: [],
        capability_ids: [],
      }),
      expect.objectContaining({
        profile_id: "network_flow_activity",
        supported_contract_majors: [2],
        workspace_keys: ["network_analysis"],
        capability_ids: [],
      }),
    ]);
    expect(
      decodeClientExtensionSupportRegistry({
        ...support,
        profiles: [{ ...support?.profiles[0], capability_ids: ["execute"] }],
      }),
    ).toBeNull();
    expect(extensionCapabilityActivationFailure()).toEqual({
      code: "extension_capability_not_supported",
    });
  });

  it("admits exact profile routes and discards a response after claim loss", async () => {
    const controller = new ExtensionAvailabilityController({
      incidentId: "incident-1",
      randomValues: deterministicRandom(4),
    });
    const importDiscovery = {
      profile_id: "import",
      claimable: true,
      claimed: true,
      contract_major: 1,
      route_families: ["/api/v1/import-sessions"],
      workspace_keys: [],
      capabilities: [],
    } as const;
    controller.setDiscovery([importDiscovery]);
    expect(
      controller.isRouteAvailable("import", "/api/v1/import-sessions"),
    ).toBe(true);
    expect(
      controller.isRouteAvailable("import", "/api/v1/reference-packs"),
    ).toBe(false);

    let resolveRequest: (value: string) => void = () => undefined;
    const request = controller.runProfileRequest(
      "import",
      "/api/v1/import-sessions",
      () =>
        new Promise<string>((resolve) => {
          resolveRequest = resolve;
        }),
    );
    await vi.waitFor(() =>
      expect(controller.currentTag()?.generation).toBe(2n),
    );
    controller.setDiscovery([{ ...importDiscovery, claimed: false }]);
    resolveRequest("stale");
    await expect(request).rejects.toBeInstanceOf(
      ExtensionAvailabilityUnavailableError,
    );
  });

  it("uses the build-bound browser bootstrap and fails closed when it is malformed", () => {
    const source = packagedClientExtensionSupportRegistry();
    if (source === null) {
      throw new Error("expected generated test support");
    }
    const script = document.createElement("script");
    script.id = "cartulary-client-extension-support-registry";
    script.type = "application/json";
    script.textContent = JSON.stringify({
      ...source,
      client_build_id: "cartulary.web.standard.sha256:bootstrap",
    });
    document.head.append(script);
    expect(packagedClientExtensionSupportRegistry()?.client_build_id).toBe(
      "cartulary.web.standard.sha256:bootstrap",
    );
    script.textContent = "{";
    expect(packagedClientExtensionSupportRegistry()).toBeNull();
    script.remove();
  });

  it("requires the exact discovery, support, authorization, and current-generation intersection", () => {
    const controller = new ExtensionAvailabilityController({
      incidentId: "incident-1",
      randomValues: deterministicRandom(0xab),
    });
    controller.setDiscovery(discovery);
    const tag = controller.reserve();
    if (tag === null) {
      throw new Error("expected availability reservation");
    }
    expect(controller.acceptWorkbookStartup(tag, availability)).toBe(true);
    expect(controller.renderableWorkspaces()).toEqual([
      {
        extensionProfileId: "network_flow_activity",
        workspaceKey: "network_analysis",
      },
    ]);

    controller.setDiscovery([{ ...discovery[0], contract_major: 3 }]);
    expect(controller.renderableWorkspaces()).toEqual([]);
    controller.setDiscovery(discovery);
    const refreshed = controller.currentTag();
    if (refreshed === null) {
      throw new Error("expected refreshed availability tag");
    }
    expect(controller.acceptWorkbookStartup(refreshed, availability)).toBe(
      true,
    );
    expect(controller.renderableWorkspaces()).toHaveLength(1);
    controller.invalidate();
    expect(controller.renderableWorkspaces()).toEqual([]);
  });

  it("ignores stale and malformed availability without retry", async () => {
    const controller = new ExtensionAvailabilityController({
      incidentId: "incident-1",
      randomValues: deterministicRandom(1),
    });
    controller.setDiscovery(discovery);
    const stale = controller.reserve();
    const current = controller.reserve();
    if (stale === null || current === null) {
      throw new Error("expected availability reservations");
    }
    expect(controller.acceptWorkbookStartup(stale, availability)).toBe(false);
    expect(
      controller.acceptWorkbookStartup(current, {
        ...availability,
        incident_id: "incident-2",
      }),
    ).toBe(false);

    const reset = controller.reserve();
    if (
      reset === null ||
      !controller.acceptWorkbookStartup(reset, availability)
    ) {
      throw new Error("expected current availability");
    }
    let resolveFirst: (value: string) => void = () => undefined;
    const first = controller.runRequest(
      () =>
        new Promise<string>((resolve) => {
          resolveFirst = resolve;
        }),
    );
    await vi.waitFor(() =>
      expect(controller.currentTag()?.generation).toBe(4n),
    );
    controller.reserve();
    resolveFirst("stale");
    await expect(first).rejects.toBeInstanceOf(
      ExtensionAvailabilityUnavailableError,
    );
  });

  it("linearizes extension requests while preserving distinct reservation generations", async () => {
    const controller = new ExtensionAvailabilityController({
      incidentId: "incident-1",
      randomValues: deterministicRandom(2),
    });
    controller.setDiscovery(discovery);
    const startup = controller.reserve();
    if (
      startup === null ||
      !controller.acceptWorkbookStartup(startup, availability)
    ) {
      throw new Error("expected current availability");
    }
    const order: string[] = [];
    const first = controller.runRequest(async () => {
      order.push("first-start");
      await Promise.resolve();
      order.push("first-end");
      return controller.currentTag()?.generation;
    });
    const second = controller.runRequest(async () => {
      order.push("second-start");
      return controller.currentTag()?.generation;
    });
    await expect(first).resolves.toBe(3n);
    await expect(second).resolves.toBe(4n);
    expect(order).toEqual(["first-start", "first-end", "second-start"]);
  });

  it("rolls over atomically and disables extensions when secure randomness fails", () => {
    const randomValues = vi
      .fn<(bytes: Uint8Array) => Uint8Array>()
      .mockImplementationOnce(deterministicRandom(1))
      .mockImplementationOnce(deterministicRandom(2));
    const controller = new ExtensionAvailabilityController({
      incidentId: "incident-1",
      initialGeneration: 18_446_744_073_709_551_615n,
      randomValues,
    });
    const initialTag = controller.currentTag();
    const rollover = controller.reserve();
    if (initialTag === null || rollover === null) {
      throw new Error("expected availability rollover");
    }
    const firstEpoch = initialTag.epochId;
    expect(rollover.generation).toBe(1n);
    expect(rollover.epochId).not.toBe(firstEpoch);
    expect(randomValues).toHaveBeenCalledTimes(2);

    const disabled = new ExtensionAvailabilityController({
      incidentId: "incident-1",
      randomValues: () => {
        throw new Error("unavailable");
      },
    });
    expect(disabled.reserve()).toBeNull();
    expect(disabled.renderableWorkspaces()).toEqual([]);
  });

  it("rejects duplicate, unsorted, extra-member, and oversized availability rows", () => {
    expect(
      decodeExtensionWorkspaceAvailability(availability, "incident-1"),
    ).toHaveLength(1);
    expect(
      decodeExtensionWorkspaceAvailability(
        { ...availability, extra: true },
        "incident-1",
      ),
    ).toBeNull();
    expect(
      decodeExtensionWorkspaceAvailability(
        {
          ...availability,
          workspaces: [availability.workspaces[0], availability.workspaces[0]],
        },
        "incident-1",
      ),
    ).toBeNull();
    expect(
      decodeExtensionWorkspaceAvailability(
        {
          ...availability,
          workspaces: Array(65).fill(availability.workspaces[0]),
        },
        "incident-1",
      ),
    ).toBeNull();
  });
});
