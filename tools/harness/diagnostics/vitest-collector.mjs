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
  constructor(root, invocationID, runID) { this.root = root; this.invocationID = invocationID; this.runID = runID; }
  onUserConsoleLog(log) { (log.type === "stderr" ? process.stderr : process.stdout).write(log.content); }
  onTestRunEnd(modules, unhandledErrors, reason) {
    const root = this.root;
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
    this.latest = {
      schema_id: "cartulary.vitest_failure_details.v3",
      invocation_id: this.invocationID,
      run_id: this.runID,
      run_status: reason, generated_at: new Date().toISOString(),
      runner_json: "", stdout_log: "", stderr_log: "", observations, failures: [],
      final_complete: false, completed_files: modules.map((module) => ({ project: module.project.name, owner_path: path.relative(root, module.moduleId).replaceAll("\\", "/") })),
    };
  }
}
