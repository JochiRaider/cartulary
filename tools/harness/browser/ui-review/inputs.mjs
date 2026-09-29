import { readFileSync } from "node:fs";
import path from "node:path";
import { commands, ReviewFailure } from "./contract.mjs";
import { repoRoot } from "./policy.mjs";
import { resolveOutputMode } from "../../contract/index.mjs";

export function resolveInputs(command, environment = process.env) {
  const fail = () => { throw new ReviewFailure("invalid_request"); };
  if (!commands.includes(command)) fail();
  const manifest = JSON.parse(readFileSync(path.join(repoRoot, "tools/task_surface_owner.json")));
  const entry = manifest.targets.find((target) => target.name === command);
  const inputs = entry.input_contract.inputs;
  const origins = Object.fromEntries((environment.CARTULARY_MAKE_INPUT_SOURCES ?? "").trim().split(/\s+/u).filter(Boolean).map((item) => item.split("=")));
  const allowed = new Set([...inputs.map((input) => input.name), ...manifest.global_inputs.map((input) => input.name), "CARTULARY_OUTPUT_MODE", "CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID", "VERBOSE", "CI_VERBOSE"]);
  for (const [name, origin] of Object.entries(origins)) if (origin === "cli" && !allowed.has(name)) fail();
  const resolved = {};
  for (const input of inputs) {
    const origin = origins[input.name];
    const explicit = origin === "cli" || (input.name === "REVIEW_PROFILE" && origin === "env");
    if (!explicit) { if (input.required) fail(); continue; }
    const value = environment[input.name];
    if (typeof value !== "string" || value.trim() === "" || /[\u0000-\u001f\u007f]/u.test(value) || Buffer.byteLength(value) > 4096) fail();
    resolved[input.name] = input.type === "path" ? value : value.trim();
    if (input.values && !input.values.includes(resolved[input.name])) fail();
  }
  if (command === "ui-review") {
    resolved.UI_MODE ??= "seeded";
    if (resolved.UI_MODE !== "seeded" && resolved.REVIEW_PROFILE !== undefined) fail();
    if (resolved.UI_MODE === "seeded") resolved.REVIEW_PROFILE ??= "network_flow_claimed";
    if (resolved.UI_MODE !== "dev" && resolved.UI_ORIGIN !== undefined) fail();
    if (resolved.UI_MODE === "dev") {
      if (!/^http:\/\/(?:127\.0\.0\.1|\[::1\]):[0-9]+\/?$/u.test(resolved.UI_ORIGIN ?? "")) fail();
      const url = new URL(resolved.UI_ORIGIN);
      const port = Number(resolved.UI_ORIGIN.match(/:(\d+)\/?$/u)[1]);
      if (port < 1 || port > 65535) fail();
      resolved.UI_ORIGIN = url.origin;
    }
  }
  let output;
  try { output = resolveOutputMode(environment, command); } catch { fail(); }
  if (!["quiet", "summary", "ci", "verbose", "debug", "machine"].includes(output) || (command === "ui-review" && output === "machine")) fail();
  if (resolved.UI_BUNDLE && !/^bundle-[1-9][0-9]*$/u.test(resolved.UI_BUNDLE)) fail();
  return { ...resolved, output };
}
