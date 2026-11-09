// Delivery validation is distinct from source-specific factual verification.
import { createHash } from "node:crypto";
import { readFileSync, lstatSync, realpathSync } from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { pathToFileURL } from "node:url";

const require = createRequire(import.meta.url);
const contracts = JSON.parse(readFileSync(new URL("./audit-contracts.json", import.meta.url)));
const canonical = value => Array.isArray(value) ? `[${value.map(canonical).join(",")}]` : value !== null && typeof value === "object" ? `{${Object.keys(value).sort().map(key => `${JSON.stringify(key)}:${canonical(value[key])}`).join(",")}}` : JSON.stringify(value);
const sha = bytes => createHash("sha256").update(bytes).digest("hex");
const regular = file => {
  if (!lstatSync(file).isFile()) throw new Error("source-audit input must be a regular file");
  return readFileSync(file);
};
function contractFor(producer) {
  if (!Object.hasOwn(contracts, producer)) throw new Error("unsupported source-audit producer");
  return contracts[producer];
}
// Capture identities before launching the producer, so a stale report cannot
// satisfy delivery merely by containing plausible hashes. Inputs are explicit;
// embedded policy identities come from the current producer-owned files.
export function auditContext(producer, lock, builder, builtAt = "", editorial = "", sourceRoot = "") {
  const contract = contractFor(producer), executable = regular(builder);
  const selected = regular(lock);
  let lockBytes = selected;
  if (contract.retained_lock) {
    if (!sourceRoot) throw new Error("explicit source-audit directory required");
    lockBytes = regular(path.join(sourceRoot, contract.retained_lock.path));
    const retained = JSON.parse(new TextDecoder("utf-8", {fatal: true}).decode(lockBytes));
    if (retained.schema_id !== contract.retained_lock.schema_id || retained.inventory_sha256 !== sha(selected)) throw new Error("source-audit inventory binding mismatch");
    if (contract.retained_lock.encoding === "canonical") lockBytes = Buffer.from(canonical(retained));
  }
  const fields = {source_lock_sha256: sha(lockBytes)};
  for (const [key, file] of Object.entries(contract.bindings)) fields[key] = sha(regular(key === "editorial_sha256" && editorial ? editorial : file));
  if (producer === "attack") {
    fields.built_at = builtAt;
    fields.source_profile_id = JSON.parse(regular(contract.bindings.source_profile_sha256)).source_profile_id;
  }
  return {producer, fields, builder: {builder_id: `cartulary.reference-pack.${producer}`, builder_version: "dataset-v1", builder_source_sha256: sha(executable), size_bytes: executable.length}};
}
export function readAuditReport(producer, file, context) {
  const contract = contractFor(producer), raw = regular(file);
  const report = JSON.parse(new TextDecoder("utf-8", {fatal: true}).decode(raw));
  const name = contract.schema_id.split(/[^A-Za-z0-9]+/u).filter(Boolean).map(part => part[0].toUpperCase() + part.slice(1)).join("");
  const validate = require(`./generated/validate${name}.cjs`);
  if (!validate(report) || !raw.equals(Buffer.from(canonical(report)))) throw new Error("invalid source-audit report contract");
  if ((producer === "attack" && report.status !== "source_audit_passed") ||
      (producer === "windows" && report.status !== "accepted") ||
      (Object.hasOwn(report, "issues") && report.issues.length)) throw new Error("source audit did not succeed");
  const fieldKeys = ["source_lock_sha256", ...Object.keys(contract.bindings), ...(producer === "attack" ? ["built_at", "source_profile_id"] : [])].sort();
  const builderKeys = ["builder_id", "builder_version", "builder_source_sha256", "size_bytes"].sort();
  if (!context || context.producer !== producer ||
      canonical(Object.keys(context.fields ?? {}).sort()) !== canonical(fieldKeys) ||
      canonical(Object.keys(context.builder ?? {}).sort()) !== canonical(builderKeys)) throw new Error("source-audit invocation context required");
  for (const [key, expected] of Object.entries(context.fields)) {
    if (typeof expected !== "string" || report[key] !== expected) throw new Error(`source-audit invocation mismatch: ${key}`);
  }
  for (const [key, expected] of Object.entries(context.builder)) {
    if (report.builder[key] !== expected) throw new Error(`source-audit builder mismatch: ${key}`);
  }
  return report;
}
if (process.argv[1] && import.meta.url === pathToFileURL(realpathSync(process.argv[1])).href) {
  const [action, producer, ...args] = process.argv.slice(2);
  if (action === "context") process.stdout.write(JSON.stringify(auditContext(producer, ...args)));
  else if (action === "validate") readAuditReport(producer, args[0], JSON.parse(regular(args[1])));
  else throw new Error("unsupported audit report action");
}
