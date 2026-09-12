// Synthetic runner observations for wrapper mechanics; real Vitest is exercised separately.
import { readFileSync, writeFileSync } from "node:fs";
import path from "node:path";

const report = JSON.parse(readFileSync(process.argv[2], "utf8"));
const observations = [];
function errors(messages) {
  if (!messages.length) return [];
  const message = messages.find((value) => !value.startsWith("Error: STACK_TRACE_ERROR")) ?? messages[0];
  const line = message.split("\n")[0];
  const split = line.indexOf(": ");
  return [{ name: split < 0 ? "Error" : line.slice(0, split), message: split < 0 ? line : line.slice(split + 2), stack: messages[0] }];
}
for (const suite of report.testResults) {
  const base = { project: "fixture", owner_path: path.relative(process.cwd(), suite.name), duration_ms: 1, timeout_ms: 15000 };
  for (const assertion of suite.assertionResults) observations.push({
    ...base, title: assertion.fullName, test_name: assertion.title, scope: "test",
    status: assertion.status, errors: errors(assertion.failureMessages),
  });
  if (!suite.assertionResults.length && suite.status === "failed") observations.push({
    ...base, title: "(suite load)", test_name: "", scope: "suite", status: "failed", errors: errors([suite.message]),
  });
}
writeFileSync(process.env.CARTULARY_VITEST_CAPTURE_FILE, JSON.stringify({
  schema_id: "cartulary.vitest_failure_details.v2", invocation_id: process.env.CARTULARY_VITEST_INVOCATION_ID,
  run_id: process.env.CARTULARY_TEST_RUN_ID, run_status: report.success ? "passed" : "failed",
  generated_at: new Date().toISOString(), runner_json: "", stdout_log: "", stderr_log: "",
  observations, failures: [],
}), { mode: 0o600 });
