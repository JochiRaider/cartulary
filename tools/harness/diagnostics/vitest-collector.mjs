import { writeFileSync } from "node:fs";
import path from "node:path";

function originalError(error) {
  return { name: String(error.name ?? "Error"), message: String(error.message ?? ""), stack: String(error.stack ?? "") };
}

function catalogTitle(test) {
  const names = [test.name];
  for (let parent = test.parent; parent && parent !== test.module; parent = parent.parent) names.unshift(parent.name);
  return names.join(" ");
}

// Uses only the supported Reporter/TestCase/TestSuite APIs; it does not modify tasks.
export default class VitestCollector {
  onTestRunEnd(modules, unhandledErrors, reason) {
    const output = process.env.CARTULARY_VITEST_CAPTURE_FILE;
    const root = process.env.CARTULARY_VITEST_REPO_ROOT;
    if (!output || !root) throw new Error("Vitest collector requires invocation-private output");
    const observations = [];
    for (const module of modules) {
      const ownerPath = path.relative(root, module.moduleId).replaceAll("\\", "/");
      const project = module.project.name;
      const addSuite = (suite, title) => {
        const errors = suite.errors();
        if (!errors.length) return;
        observations.push({ project, owner_path: ownerPath, title, test_name: "", scope: "suite", status: "failed", duration_ms: 0, timeout_ms: null, errors: errors.map(originalError) });
      };
      addSuite(module, "(suite load)");
      for (const suite of module.children.allSuites()) addSuite(suite, `${suite.fullName} (suite)`);
      for (const test of module.children.allTests()) {
        const result = test.result();
        observations.push({
          project, owner_path: ownerPath, title: catalogTitle(test), test_name: test.name, scope: "test",
          status: result.state, duration_ms: test.diagnostic()?.duration ?? 0,
          timeout_ms: test.options.timeout ?? test.project.config.testTimeout ?? null,
          errors: (result.errors ?? []).map(originalError),
        });
      }
    }
    if (unhandledErrors.length) observations.push({
      project: "", owner_path: "", title: "(unhandled runner error)", test_name: "", scope: "run",
      status: "failed", duration_ms: 0, timeout_ms: null, errors: unhandledErrors.map(originalError),
    });
    writeFileSync(output, JSON.stringify({
      schema_id: "cartulary.vitest_failure_details.v2",
      invocation_id: process.env.CARTULARY_VITEST_INVOCATION_ID,
      run_id: process.env.CARTULARY_TEST_RUN_ID,
      run_status: reason, generated_at: new Date().toISOString(),
      runner_json: "", stdout_log: "", stderr_log: "", observations, failures: [],
    }), { flag: "wx", mode: 0o600 });
  }
}
