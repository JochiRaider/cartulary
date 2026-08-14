import { cartularyAc043PerformanceContract } from "@cartulary/ui-contracts";
import { describe, expect, it } from "vitest";

import { parseAc043RuntimeBundle } from "./ac043Snapshot";

const snapshotKey = "a".repeat(64);

function runtimeBundle() {
  return {
    schema_id: "cartulary.performance_fixture_runtime.v1",
    fixture_profile_id: "ac043_large_grid_snapshot_v1",
    snapshot_key: snapshotKey,
    background_accounts: Array.from({ length: 24 }, (_, index) => ({
      email: `private-${index}@example.test`,
      password: `private-runtime-password-${String(index).padStart(2, "0")}`,
    })),
  };
}

describe("AC-043 snapshot runtime bundle", () => {
  it("preserves the typed supported-envelope fixture and traffic shape", () => {
    expect(cartularyAc043PerformanceContract.fixture).toEqual({
      fixtureId: "cartulary.perf.large_grid.v1",
      seed: 20260405,
      timelineRows: 20_000,
      hostRows: 1_000,
      identityRows: 1_000,
      tagAssignments: 1_000,
      mentionAssignments: 1_000,
      linkAssignments: 1_000,
      analystSessions: 25,
      backgroundSessions: 24,
      backgroundUpdateIntervalMs: 5_000,
      backgroundUpdatesPerSecond: 4.8,
      trafficTraceId: "cartulary.perf.live_updates_25sessions.v1",
    });
  });

  it("accepts exactly 24 private accounts for the admitted snapshot", () => {
    const parsed = parseAc043RuntimeBundle(JSON.stringify(runtimeBundle()), {
      fixtureProfileId: "ac043_large_grid_snapshot_v1",
      snapshotKey,
    });

    expect(parsed.backgroundAccounts).toHaveLength(24);
    expect(parsed.fixtureProfileId).toBe("ac043_large_grid_snapshot_v1");
    expect(parsed.snapshotKey).toBe(snapshotKey);
  });

  it("rejects stale identity, duplicate accounts, and extra fields", () => {
    const stale = runtimeBundle();
    stale.snapshot_key = "b".repeat(64);
    expect(() =>
      parseAc043RuntimeBundle(JSON.stringify(stale), {
        fixtureProfileId: "ac043_large_grid_snapshot_v1",
        snapshotKey,
      }),
    ).toThrow(/identity is inconsistent/u);

    const duplicate = runtimeBundle();
    duplicate.background_accounts[1] = duplicate.background_accounts[0] as {
      email: string;
      password: string;
    };
    expect(() =>
      parseAc043RuntimeBundle(JSON.stringify(duplicate), {
        fixtureProfileId: "ac043_large_grid_snapshot_v1",
        snapshotKey,
      }),
    ).toThrow(/duplicate accounts/u);

    expect(() =>
      parseAc043RuntimeBundle(
        JSON.stringify({ ...runtimeBundle(), runtime_path: "/private" }),
        {
          fixtureProfileId: "ac043_large_grid_snapshot_v1",
          snapshotKey,
        },
      ),
    ).toThrow(/invalid field set/u);
  });
});
