#!/usr/bin/env node
import { validateSchemaSync } from "../contract/index.mjs";
import { runObservationSchema } from "../contract/live-observation.mjs";
import { emptyObservation, observationFailure, observeRun, renderObservation } from "./run-observation.mjs";

const controller = new AbortController();
const interrupt = (signal) => controller.abort(Object.assign(new Error("reader interrupted"), { failure_class: "interrupted", failure_reason: "cancelled_or_interrupted", signal }));
const onINT = () => interrupt("SIGINT");
const onTERM = () => interrupt("SIGTERM");
process.once("SIGINT", onINT);
process.once("SIGTERM", onTERM);
let json = process.argv.includes("--json");
let admitted = false;
try {
  validateSchemaSync(runObservationSchema, emptyObservation());
  admitted = true;
  const options = {};
  const flags = new Map([["--results-dir", "resultsDir"], ["--run-id", "runId"], ["--target", "target"], ["--after-revision", "afterRevision"], ["--wait-seconds", "waitSeconds"]]);
  const seen = new Set();
  for (let index = 2; index < process.argv.length; index++) {
    const flag = process.argv[index];
    if (seen.has(flag)) throw Object.assign(new Error("duplicate input"), { failure_class: "config", failure_reason: "usage_error" });
    seen.add(flag);
    if (flag === "--json") { json = true; continue; }
    if (!flags.has(flag) || index + 1 >= process.argv.length) throw Object.assign(new Error("invalid input"), { failure_class: "config", failure_reason: "usage_error" });
    options[flags.get(flag)] = process.argv[++index];
  }
  const result = await observeRun(options, { signal: controller.signal });
  process.stdout.write(json ? `${JSON.stringify(result)}\n` : renderObservation(result));
  process.exitCode = result.operation_exit_code;
} catch (error) {
  if (!admitted) {
    process.stderr.write("[CONFIG] test-run-status requires installed schema-validation tooling\n");
    process.exitCode = 2;
  } else {
    const result = observationFailure(error);
    validateSchemaSync(runObservationSchema, result);
    process.stdout.write(json ? `${JSON.stringify(result)}\n` : renderObservation(result));
    process.exitCode = result.operation_exit_code;
  }
} finally {
  process.removeListener("SIGINT", onINT);
  process.removeListener("SIGTERM", onTERM);
}
