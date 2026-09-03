import { render, waitFor } from "@testing-library/react";
import { useEffect } from "react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  ExtensionAvailabilityController,
  type ExtensionDiscoveryProfile,
} from "../../extensions/extensionAvailability";
import { useWorkbookExtensionAvailability } from "./useWorkbookExtensionAvailability";

const claimedImportProfile: ExtensionDiscoveryProfile = {
  profile_id: "import",
  claimable: true,
  claimed: true,
  contract_major: 1,
  route_families: ["/api/v1/import-sessions"],
  workspace_keys: [],
  capabilities: [],
};

function ExtensionAvailabilityHarness({
  onCommit,
  onRender,
  profiles,
}: {
  readonly onCommit: (
    controller: ExtensionAvailabilityController,
    revision: number,
  ) => void;
  readonly onRender: (setDiscoveryCalls: number) => void;
  readonly profiles: readonly ExtensionDiscoveryProfile[];
}) {
  onRender(
    vi.mocked(ExtensionAvailabilityController.prototype.setDiscovery).mock.calls
      .length,
  );
  const availability = useWorkbookExtensionAvailability({
    clientInstanceId: "client-1",
    incidentId: "incident-1",
    profiles,
  });
  useEffect(() => {
    onCommit(availability.controller, availability.revision);
  }, [availability.controller, availability.revision, onCommit]);
  return null;
}

describe("useWorkbookExtensionAvailability", () => {
  afterEach(() => vi.restoreAllMocks());

  it("mutates discovery after render and preserves one controller per subject", async () => {
    const originalSetDiscovery =
      ExtensionAvailabilityController.prototype.setDiscovery;
    const setDiscovery = vi
      .spyOn(ExtensionAvailabilityController.prototype, "setDiscovery")
      .mockImplementation(function (
        this: ExtensionAvailabilityController,
        profiles,
      ) {
        return originalSetDiscovery.call(this, profiles);
      });
    const commits: Array<{
      readonly controller: ExtensionAvailabilityController;
      readonly revision: number;
    }> = [];
    const renderCallCounts: number[] = [];
    const onCommit = (
      controller: ExtensionAvailabilityController,
      revision: number,
    ) => commits.push({ controller, revision });
    const view = render(
      <ExtensionAvailabilityHarness
        onCommit={onCommit}
        onRender={(count) => renderCallCounts.push(count)}
        profiles={[claimedImportProfile]}
      />,
    );

    expect(renderCallCounts[0]).toBe(0);
    await waitFor(() => expect(commits.at(-1)?.revision).toBe(1));
    const initialController = commits.at(-1)?.controller;
    expect(setDiscovery).toHaveBeenCalledTimes(1);

    view.rerender(
      <ExtensionAvailabilityHarness
        onCommit={onCommit}
        onRender={(count) => renderCallCounts.push(count)}
        profiles={[{ ...claimedImportProfile }]}
      />,
    );
    await waitFor(() => expect(setDiscovery).toHaveBeenCalledTimes(2));
    expect(commits.at(-1)?.controller).toBe(initialController);
    expect(commits.at(-1)?.revision).toBe(1);

    view.rerender(
      <ExtensionAvailabilityHarness
        onCommit={onCommit}
        onRender={(count) => renderCallCounts.push(count)}
        profiles={[{ ...claimedImportProfile, claimed: false }]}
      />,
    );
    await waitFor(() => expect(commits.at(-1)?.revision).toBe(2));
    expect(commits.at(-1)?.controller).toBe(initialController);
    expect(
      renderCallCounts.every(
        (count) => count <= setDiscovery.mock.calls.length,
      ),
    ).toBe(true);
  });
});
