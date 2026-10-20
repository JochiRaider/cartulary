import { existsSync, lstatSync, readFileSync } from "node:fs";
import path from "node:path";
import { parseStrictJSON, validateSchemaSync } from "../contract/index.mjs";
import { readCanonicalUnitEvents, reduceCanonicalUnitIntervals, notLaunchedUnitIDs } from "../evidence-accounting/index.mjs";

function safeID(value) { return value.replaceAll(/[^A-Za-z0-9_.+-]+/gu, "-"); }

export function readBrowserUnitResult(base, unitID) {
  const file = path.join(base, "unit-results", `${safeID(unitID)}.json`);
  if (!existsSync(file)) {
    throw new Error(`browser target unit result is missing: ${unitID}`);
  }
  const stat = lstatSync(file);
  if (!stat.isFile() || stat.isSymbolicLink() || stat.size > 1024 * 1024) {
    throw new Error(`browser target unit result exceeds its bounded JSON contract: ${file}`);
  }
  const result = parseStrictJSON(readFileSync(file, "utf8"));
  validateSchemaSync("cartulary.harness_unit_result.v1", result);
  if (result.unit_id !== unitID) {
    throw new Error(`browser target unit result identity mismatch: ${unitID}`);
  }
  return result;
}


// This projection reads the existing canonical lifecycle; it never infers launch
// from missing files or publishes replacement runner evidence.
export async function readUnlaunchedBrowserFailures(base, options, groups, source) {
  const failures = [];
  const prefix = options.resetPrefix.replace(/^browser_reset:/u, "browser_group:");
  const ids = options.groups.map(group => `${prefix}${safeID(group)}`);
  // Resolve only the selected groups and their declared dependency closure.
  // Other branches may still run; their open waits are not a summary defect.
  const registrations = new Map();
  for await (const event of readCanonicalUnitEvents(source.file, { prefix: source.prefix })) {
    if (event.event !== "queued") continue;
    if (registrations.has(event.unit_id)) throw new Error("duplicate canonical registration");
    registrations.set(event.unit_id, event);
  }
  const selected = new Set(), pending = [...ids];
  while (pending.length) {
    const id = pending.pop();
    if (selected.has(id)) continue;
    const registration = registrations.get(id);
    if (!registration) throw new Error("browser dependency lacks canonical registration");
    selected.add(id); pending.push(...registration.needs);
  }
  const state = await reduceCanonicalUnitIntervals(source.file, { prefix: source.prefix, unitIDs: [...selected] });
  if ([...selected].some(id => !state.terminals.has(id))) throw new Error("browser dependencies have incomplete canonical outcomes");
  const notLaunched = notLaunchedUnitIDs(state);
  for (const group of options.groups) {
    const id = `${prefix}${safeID(group)}`;
    if (!state.registrations.has(id) || !state.terminals.has(id)) throw new Error("browser target lacks complete canonical group outcomes");
    const retained = groups.find(entry => entry.result.group_id === group);
    if (retained) {
      if (notLaunched.has(id)) throw new Error("unlaunched browser group has contradictory runner evidence");
      continue;
    }
    if (!notLaunched.has(id)) throw new Error("launched browser group is missing required results");
    const result = readBrowserUnitResult(base, id), terminal = state.terminals.get(id);
    const reference = `${options.groupTargets.get(group)}/browser-groups/${group}/browser-group-result.json`;
    const skipped = terminal.status === "skipped" && terminal.failure_reason === "dependency_failure";
    if (result.status !== terminal.status || result.failure_class !== (terminal.failure_class ?? null) ||
        result.failure_reason !== (skipped ? null : terminal.failure_reason) || !result.evidence_outputs.includes(reference)) {
      throw new Error("unlaunched browser group evidence contradicts canonical outcome");
    }
  }
  // Skipped descendants have no new primary failure. Retain the failed ancestor
  // from the same declared closure, including setup outside the browser stage.
  failures.push(...[...state.terminals.values()].filter(event => event.failure_class && event.failure_reason).map(event => ({
    failure_class: event.failure_class, failure_reason: event.failure_reason, scheduler_event_sequence: event.seq,
  })));
  if (!failures.length) throw new Error("unlaunched browser target has no causal failure");
  return failures;
}
