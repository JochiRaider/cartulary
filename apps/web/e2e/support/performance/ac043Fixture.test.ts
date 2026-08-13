import { describe, expect, it } from "vitest";

import { ac043FixtureCacheKey, timelineFixtureMatrix } from "./ac043Fixture";

describe("AC-043 performance fixture", () => {
  it("defines the deterministic supported-envelope shape and cache invalidation key", () => {
    const matrix = timelineFixtureMatrix(0, 20_000);
    expect(matrix).toHaveLength(20_000);
    expect(matrix.filter((row) => row[1] !== "")).toHaveLength(1_000);
    expect(matrix.filter((row) => row[2] !== "")).toHaveLength(1_000);
    expect(matrix.filter((row) => row[3] !== "")).toHaveLength(1_000);
    expect(matrix.filter((row) => row[4] !== "")).toHaveLength(1_000);
    expect(matrix[0]).toEqual([
      "Performance Timeline 00000 perf-host-0000 perf-identity-0000@example.test https://fixture-0000.example.test/trace",
      "perf-host-0000",
      "perf-identity-0000@example.test",
      "perf-tag-0000",
      "https://fixture-0000.example.test/trace",
    ]);
    expect(matrix[19]?.slice(1)).toEqual(["", "", "", ""]);
    expect(matrix[20]?.slice(1)).toEqual([
      "perf-host-0001",
      "perf-identity-0001@example.test",
      "perf-tag-0001",
      "https://fixture-0001.example.test/trace",
    ]);
    expect(
      ac043FixtureCacheKey({
        migrationDigest: "m1",
        sourceContractDigest: "s1",
      }),
    ).not.toBe(
      ac043FixtureCacheKey({
        migrationDigest: "m2",
        sourceContractDigest: "s1",
      }),
    );
    expect(
      ac043FixtureCacheKey({
        migrationDigest: "m1",
        sourceContractDigest: "s1",
      }),
    ).not.toBe(
      ac043FixtureCacheKey({
        migrationDigest: "m1",
        sourceContractDigest: "s2",
      }),
    );
  });
});
