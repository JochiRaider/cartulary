// Only the exact harness adapter contract admits this fixture. A global union
// of titles would run the failing case that belongs to the other selected file.
import { describe, expect, test } from "vitest";

describe("scope", () => {
  test("one", () => {
    throw new Error("title selected only in sibling file");
  });
  test("independent", () => {
    expect(
      (globalThis as { cohortSentinel?: string }).cohortSentinel,
    ).toBeUndefined();
  });
});
