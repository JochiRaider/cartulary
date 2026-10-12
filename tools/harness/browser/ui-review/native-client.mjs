// Input compilation and output adaptation only. Native review owns sessions,
// admission, recovery, cleanup, registry authority and retained publication.
import path from "node:path";
import { spawn } from "../../workspace/child-process.mjs";
import { workspaceLayout } from "../../../workspace_layout.generated.mjs";
import { generateRunId, validateResultRoot, validateRunId } from "../../contract/harness-contract.mjs";
import { parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { readInput } from "./session-files.mjs";
import { limits, parseRequest, ReviewFailure } from "./contract.mjs";
import { emitResult } from "./output.mjs";
import { repoRoot } from "./policy.mjs";

export function reviewEnvironment(environment = process.env) {
  const copy = { ...environment };
  for (const name of Object.keys(copy)) {
    if (/^(?:UI_|REVIEW_PROFILE$|MAKEFLAGS$|MAKEOVERRIDES$|MFLAGS$|NODE_OPTIONS$|NODE_PATH$|NODE_V8_COVERAGE$|DEBUG$|PWDEBUG$|PWTEST_|PLAYWRIGHT_(?!BROWSERS_PATH$)|OTEL_|CARTULARY_(?:MAKE_|HARNESS_|BROWSER_|WEB_E2E_|PGTEST_|S3_)|CARTULARY__)/u.test(name)) delete copy[name];
  }
  return copy;
}
export function nativeReview(message, { onProgress = () => {}, environment = process.env } = {}) {
  const request = { schema_id: "cartulary.harness_review_command.v1", ...message };
  validateSchemaSync(request.schema_id, request);
  const input = Buffer.from(JSON.stringify(request));
  if (input.length > limits.frame) throw new ReviewFailure("invalid_request");
  const binary = path.join(repoRoot, workspaceLayout.toolbin, "cartulary-harness");
  const worker = spawn(binary, ["review"], { cwd: repoRoot, env: reviewEnvironment(environment), detached: true, stdio: ["pipe", "pipe", "ignore"] });
  return new Promise((resolve, reject) => {
    let bytes = Buffer.alloc(0), terminal, failed;
    const abort = () => worker.kill("SIGINT"), terminate = () => worker.kill("SIGTERM");
    const fail = (error) => { failed ??= error; worker.kill("SIGTERM"); };
    process.on("SIGINT", abort); process.on("SIGTERM", terminate);
    worker.stdin.on("error", () => fail(new ReviewFailure("session_lost")));
    worker.once("error", () => { failed = new ReviewFailure("tool_configuration"); });
    worker.stdout.on("data", (part) => {
      bytes = Buffer.concat([bytes, part]);
      for (;;) {
        const end = bytes.indexOf(10);
        if (end < 0) { if (bytes.length >= limits.frame) fail(new ReviewFailure("invalid_artifact")); break; }
        if (end + 1 > limits.frame) { fail(new ReviewFailure("invalid_artifact")); break; }
        const line = bytes.subarray(0, end); bytes = bytes.subarray(end + 1);
        try {
          const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(line));
          if (Object.keys(value).sort().join(",") !== "kind,value" || terminal) throw new ReviewFailure("invalid_artifact");
          if (value.kind === "terminal") { validateSchemaSync("cartulary.ui_review_command_result.v2", value.value); terminal = value.value; }
          else if (["preparing", "ready"].includes(value.kind) && ["ui-review", "browser-design-review", "browser-design-review-smoke"].includes(message.command) && typeof value.value === "string" && value.value.length < 4096 && !/[\u0000-\u001f]/u.test(value.value)) onProgress(value);
          else if (value.kind === "design_ready" && message.command.startsWith("browser-design-review")) { validateSchemaSync("cartulary.harness_review_design_ready.v1", value.value); onProgress(value); }
          else throw new ReviewFailure("invalid_artifact");
        } catch (error) { fail(error instanceof ReviewFailure ? error : new ReviewFailure("invalid_artifact")); }
      }
    });
    worker.once("close", (code) => {
      process.off("SIGINT", abort); process.off("SIGTERM", terminate);
      if (failed || !terminal || bytes.length || code !== terminal.exit_code) reject(failed ?? new ReviewFailure("session_lost"));
      else resolve(terminal);
    });
    worker.stdin.end(input);
  });
}
export async function execute(command, input, profile) {
  let value;
  if (command === "ui-review") {
    let locator;
    const results = validateResultRoot(process.env.CARTULARY_TEST_RESULTS_DIR || undefined, { root: repoRoot, create: false });
    const runID = process.env.CARTULARY_TEST_RUN_ID ? validateRunId(process.env.CARTULARY_TEST_RUN_ID) : generateRunId();
    try {
      value = await nativeReview({ command, mode: input.UI_MODE, origin: input.UI_ORIGIN ?? null, profile: input.REVIEW_PROFILE ?? null,
        tool_profile: profile, results_root: results, run_id: runID, preparation_policy: "installed_only" }, { onProgress: ({ kind, value }) => {
        locator = value;
        process.stdout.write(kind === "preparing" ? `UI review preparing. UI_SESSION=${value}\n` : `UI review ready (${input.UI_MODE}). UI_SESSION=${value}\nPrivate review links expire on stop or session failure.\n`);
      } });
    } catch (error) {
      if (!locator) throw error;
      value = await nativeReview({ command: "ui-review-stop", locator, request: null, bundle_id: null });
    }
  } else {
    let request = null;
    if (input.UI_REQUEST) request = parseRequest(readInput(input.UI_REQUEST, { extension: ".json", maximum: limits.request }), command === "ui-browser" ? "action" : command === "ui-capture" ? "capture_request" : "analysis_request");
    value = await nativeReview({ command, locator: input.UI_SESSION, request, bundle_id: input.UI_BUNDLE ?? null });
  }
  emitResult(value, input.output); process.exitCode = value.exit_code;
}
