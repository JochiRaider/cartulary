import test from "node:test";
test("foundation validators preserve bootstrap operation without installed dependencies", async () => {
  await import("./test-foundation-schema-validator.mjs");
});
