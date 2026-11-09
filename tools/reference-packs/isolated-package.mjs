import { readFileSync, mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";
import { isolated, retainIsolation } from "../reference-pack-modules/isolation.mjs";

const [binary, inputPath, action, work] = process.argv.slice(2);
if (!work) throw new Error("explicit producer isolation inputs required");
const input = JSON.parse(readFileSync(inputPath, "utf8"));
const readonly = [binary, inputPath], writable = [];
const add = (target, fields, list) => {
  for (const field of fields) if (target?.[field]) {
    const value = target[field];
    if (!path.isAbsolute(value) || value === "/") throw new Error("absolute producer capability required");
    if (list === writable) mkdirSync(value, { recursive: true, mode: 0o700 });
    list.push(value);
  }
};
add(input, ["source_root", "source_lock", "source_profile", "field_policy", "source_inventory", "editorial", "license_decision", "output_policy", "approval"], readonly);
add(input, ["output_root", "ledger_root"], writable);
add(input.envelope, ["logical_root", "signing_context", "previous_envelope"], readonly);
add(input.envelope, ["output_root", "ledger_root"], writable);
const cwd = path.join(work, "empty");mkdirSync(cwd, { recursive: true });readonly.push(cwd);
const minimal = readonly.filter((item, index, all) => all.indexOf(item) === index && !all.some(other => other !== item && item.startsWith(other + path.sep)));
const manifest = { schema_id: "cartulary.reference_pack_isolation.v1", readonly: minimal, writable: [...new Set(writable)], cwd, environment: {}, command: [binary, "--action", action, "--input", inputPath] };
writeFileSync(path.join(work, "isolation.json"), JSON.stringify(manifest), { flag: "wx", mode: 0o600 });
const report = isolated(manifest);
retainIsolation("corpus", manifest, { action, report: JSON.parse(report) });
process.stdout.write(report);
