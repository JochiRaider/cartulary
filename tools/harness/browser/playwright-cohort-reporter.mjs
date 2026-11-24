import path from "node:path";
import { fileURLToPath } from "node:url";
import { redactValue } from "../contract/index.mjs";
import { atomicLocalFile } from "../runtime/secure-local-files.mjs";
import { publishRowResult } from "../execution/runners/row-results.mjs";
import { adaptPlaywrightReport, playwrightReportMaximumBytes } from "../execution/runners/playwright.mjs";
import { loadTestCatalog, playwrightSelection } from "../test-catalog/index.mjs";

// Public Reporter callbacks are delivered after test fixtures and applicable
// afterAll hooks settle. Worker/global teardown errors are invocation failures
// while independently settled test outcomes remain intact.
export class PlaywrightCohortCollector {
  constructor({ rows, selection, runRoot, checkpointFile, publishRows = true }) {
    Object.assign(this, { rows, selection, runRoot, checkpointFile, publishRows });
    this.report = { suites: [], errors: [] };
    this.completed = new Set();
    this.started = new Date();
  }
  onTestEnd(test, result) {
    const titles = [test.title];
    for (let suite = test.parent; suite?.type === "describe"; suite = suite.parent) titles.unshift(suite.title);
    const project = test.parent.project().name;
    const file = test.location.file.replaceAll("\\", "/").replace(/^.*?apps\/web\//u, "apps/web/");
    const matches = this.selection.filter((entry) => entry.file === file && entry.project_id === project &&
      JSON.stringify(entry.title_path) === JSON.stringify(titles));
    const spec = { file, title: test.title, tests: [{ projectName: project, results: [{
      status: result.status, duration: result.duration, retry: result.retry,
      errors: result.errors,
      attachments: result.attachments.map((attachment) => ({ ...attachment,
        ...(attachment.body ? { body: attachment.body.toString("base64") } : {}),
      })),
    }] }] };
    let body = { specs: [spec] };
    for (const title of titles.slice(0, -1).reverse()) body = { suites: [{ title, file, ...body }] };
    this.report.suites.push({ title: "", file, ...body });
    for (const entry of matches) this.completed.add(JSON.stringify([entry.row_id, entry.file, entry.title_path]));
    this.flush();
  }
  onError(error) {
    this.report.errors.push(error);
    this.flush();
  }
  flush() {
    const report = redactValue(this.report);
    const bytes = Buffer.from(`${JSON.stringify(report)}\n`);
    if (bytes.length > playwrightReportMaximumBytes) throw new Error("Playwright checkpoint exceeds report byte limit");
    atomicLocalFile(this.checkpointFile, bytes, { replace: true, maximumReplacementBytes: playwrightReportMaximumBytes });
    if (!this.publishRows) return;
    const finishedAt = new Date().toISOString();
    for (const result of adaptPlaywrightReport(this.rows, report, 0, null, this.selection)) {
      const selected = this.selection.filter((entry) => entry.row_id === result.row_id);
      if (!selected.every((entry) => this.completed.has(JSON.stringify([entry.row_id, entry.file, entry.title_path])))) continue;
      publishRowResult(this.runRoot, { schema_id: "cartulary.harness_row_result.v6", ...result,
        runner: "playwright", started_at: this.started.toISOString(), finished_at: finishedAt,
        wall_duration_ms: Date.now() - this.started.getTime() });
    }
  }
}

export default class PlaywrightCohortReporter extends PlaywrightCohortCollector {
  constructor() {
    const root = fileURLToPath(new URL("../../../", import.meta.url));
    const runRoot = path.resolve(root, process.env.CARTULARY_TEST_RESULTS_DIR, process.env.CARTULARY_TEST_RUN_ID);
    // The browser group CLI hands its validated cohort to the reporter.
    const group = process.env.CARTULARY_BROWSER_COHORT ?? "";
    const rowIDs = (process.env.CARTULARY_BROWSER_COHORT_ROWS ?? "").split(",").filter(Boolean);
    const target = process.env.CARTULARY_TEST_TARGET ?? "";
    if (!/^cohort-[a-f0-9]{64}$/u.test(group) || !rowIDs.length || !/^[A-Za-z0-9_.-]+$/u.test(target)) {
      throw new Error("Playwright reporter requires its browser cohort");
    }
    const catalog = loadTestCatalog(root);
    const rows = rowIDs.map((id) => catalog.rowByID.get(id));
    if (rows.some((row) => row?.runner !== "playwright")) throw new Error("Playwright reporter requires Playwright rows");
    const checkpointFile = path.join(runRoot, target, "browser-groups", group, "playwright-checkpoint.json");
    super({ rows, selection: playwrightSelection(rows), runRoot, checkpointFile });
  }
}
