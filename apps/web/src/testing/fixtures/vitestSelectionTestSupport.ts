// Executed only by the harness adapter contract with an explicit include and
// exact title selection. The ordinary frontend test profile does not admit
// *TestSupport.ts. This deliberately failing sentinel must remain unselected.
import { describe, expect, test } from "vitest";

let count = 0;
describe("scope", () => {
  test("one", () => expect(count++).toBe(0));
  test("two", () => expect(count++).toBe(1));
});
test("prefix scope one", () => {
  throw new Error("unselected suffix collision executed");
});
