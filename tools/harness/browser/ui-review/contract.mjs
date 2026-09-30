import { reviewPins } from "./policy.mjs";
import { parseStrictJSON, validateSchemaSync } from "../../contract/index.mjs";
import { validateBundleSemantics } from "./bundle-semantics.mjs";

export const schemaID = (name) => `cartulary.ui_review_${name}.v${["command_result", "receipt", "action"].includes(name) ? 2 : 1}`;
export const commandID = (name) => `cartulary.harness.command.${name.replaceAll("-", "_")}.v1`;
export const commands = Object.freeze(["ui-review", "ui-review-status", "ui-browser", "ui-capture", "ui-analyze", "ui-review-report", "ui-review-stop"]);
export const limits = Object.freeze({ request: 65536, png: 32 * 1024 ** 2, pixels: 16777216, dimension: 8192, files: 64, bundle: 128 * 1024 ** 2, storage: 512 * 1024 ** 2, bundles: 100, component: 8 * 1024 ** 2, report: 32 * 1024 ** 2, snapshot: 1048576, lock: 5000, action: 10000, operation: 30000, lifetime: 8 * 3600000 });
import { ReviewFailure, failureMappings, failureFromRecord } from "./failure.mjs";
export { ReviewFailure, failureRecord, failureMappings } from "./failure.mjs";
export function validate(name, value, code = "invalid_artifact") {
  try {
    validateSchemaSync(schemaID(name), value);
    if (name === "command_result" || name === "receipt") {
      if ((value.status === "ok") !== (value.failures.length === 0)) throw new Error("outcome mismatch");
      if ((value.status === "ok") !== (value.exit_code === 0)) throw new Error("exit mismatch");
      for (const failure of value.failures) {
        failureFromRecord(failure);
        const [kind, reason] = failureMappings[failure.diagnostic_code];
        if (failure.failure_class !== kind || failure.failure_reason !== reason) throw new Error("failure mismatch");
      }
    }
    if (name === "command_result" && (value.status === "error" || ["closed", "failed"].includes(value.state)) && value.private_refs.length) throw new Error("private terminal output");
    if (name === "receipt") {
      const terminal = ["closed", "failed"].includes(value.state);
      if (terminal !== (value.cleanup !== "not_terminal") || terminal !== (value.operation_id === null)) throw new Error("receipt lifecycle");
      if (terminal && value.command_id !== commandID("ui-review")) throw new Error("terminal owner");
    }
    if (name === "session" && value.terminal_receipt && !["closed", "failed"].includes(value.state)) throw new Error("live terminal reference");
    if (name === "bundle") {
      validateBundleSemantics(value);
      for (const values of [value.limitations, value.binding?.fixture_ids ?? []]) if (values.some((item, index) => index > 0 && item <= values[index - 1])) throw new Error("unordered set");
      if (value.source.kind === "reference_image" && value.source.import_ref.metadata !== null) throw new Error("reference metadata");
      if (value.source.kind === "canonical_visual" && value.source.import_ref.metadata === null) throw new Error("missing canonical metadata");
      const refs = [...Object.values(value.components).filter(Boolean), ...value.derived.map((entry) => entry.ref)];
      if (new Set(refs.map((ref) => ref.path)).size !== refs.length) throw new Error("duplicate component");
      for (const ref of refs) if (ref.path.split("/").some((part) => !part || part === "." || part === "..")) throw new Error("unnormalized reference");
      for (const entry of value.derived) if ((entry.kind === "crop") !== (entry.rectangle !== null)) throw new Error("crop rectangle");
    }
    if (name === "observations") {
      if ((value.axe.status === "completed") !== (value.axe.engine_version === reviewPins()["axe-core"]) || (value.axe.status !== "completed" && value.axe.engine_version !== null)) throw new Error("axe engine identity");
      if (value.axe.violations.concat(value.axe.incomplete).reduce((sum, entry) => sum + entry.node_refs.length, 0) > 1000) throw new Error("axe occurrence limit");
      if (value.accessibility_snapshot !== null && Buffer.byteLength(value.accessibility_snapshot) > limits.snapshot) throw new Error("snapshot limit");
      if (value.axe.status !== "completed" && (value.axe.violations.length || value.axe.incomplete.length)) throw new Error("unavailable findings");
      for (const channel of [value.console, value.network]) for (let index = 1; index < channel.records.length; index++) if (channel.records[index].sequence <= channel.records[index - 1].sequence) throw new Error("channel ordering");
    }
  }
  catch (cause) { throw new ReviewFailure(code, { cause }); }
  return value;
}
export function parseRequest(bytes, name) {
  try {
    if (bytes.length < 1 || bytes.length > limits.request) throw new Error("request size");
    const value = parseStrictJSON(new TextDecoder("utf-8", { fatal: true }).decode(bytes), "review request");
    validate(name, value, "invalid_request");
    function utf8(item, field = "") {
      const maximum = field === "text" ? 16384 : field === "path" ? (name === "action" ? 2048 : 4096) : field === "run_root" ? 4096 : 1024;
      if (typeof item === "string" && Buffer.byteLength(item) > maximum) throw new Error("string limit");
      if (Array.isArray(item)) for (const child of item) utf8(child, field);
      else if (item && typeof item === "object") for (const [key, child] of Object.entries(item)) utf8(child, key);
    }
    utf8(value);
    if (name === "action" && value.action === "resize" && value.parameters.width * value.parameters.height > 8294400) throw new Error("viewport area");
    if (name === "action" && value.action === "navigate" && (!/^\/(?!\/)/u.test(value.parameters.path) || /[\\#\u0000-\u001f]/u.test(value.parameters.path))) throw new Error("navigation path");
    if (name === "analysis_request") {
      value.operations ??= ["contact_sheet"];
      value.comparison ??= null;
      value.crops ??= [];
      if (value.operations.includes("exact_diff") !== (value.comparison !== null) || value.operations.includes("crop") !== (value.crops.length > 0)) throw new Error("analysis variant");
      value.operations.sort();
    }
    if (name === "capture_request" && value.source === "page") {
      value.binding ??= null; value.scope ??= { kind: "viewport" }; value.targets ??= []; value.include_axe ??= true;
      if (value.scope.kind === "element" && value.targets.length === 64 && !value.targets.some((target) => JSON.stringify(target) === JSON.stringify(value.scope.target))) throw new Error("target union limit");
    }
    return value;
  } catch (cause) { throw new ReviewFailure("invalid_request", { cause }); }
}
export function emptyCounts() { return { images: 0, observed_elements: 0, axe_violations: 0, axe_incomplete: 0, console_errors: 0, failed_requests: 0 }; }
export function result(command, fields = {}) {
  return validate("command_result", { schema_id: schemaID("command_result"), command_id: commandID(command), session_id: null, operation_id: null, state: null, epoch: null, status: "ok", exit_code: 0, failures: [], receipt: null, bundle_id: null, private_refs: [], ...fields });
}
