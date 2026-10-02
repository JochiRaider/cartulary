import assert from "node:assert/strict";
import test from "node:test";
import { publicWorkflow } from "./ui-review-public-workflow.mjs";
import { realStackCleanupCases } from "./ui-review-fixture-cleanup.mjs";

const profile = process.env.CARTULARY_UI_REVIEW_TEST_PROFILE;
assert.ok(["default", "network_flow_claimed"].includes(profile), "Make must select an exact review profile");
test(`public seeded editor/viewer and artifact workflows qualify ${profile} with complete private cleanup`, async () => {
  const result = await publicWorkflow({ seeded: true, profile });
  assert.equal(result.private_cleanup, "complete");
  assert.equal(result.retained_projection, "structural_only");
  if (profile === "default") await realStackCleanupCases();
});
