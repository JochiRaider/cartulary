import { writeFileSync } from "node:fs";
import path from "node:path";
import { parseStrictJSON, readRetainedArtifactBytes } from "../harness/contract/index.mjs";

// The command adapter resolves caller-supplied paths before moving to an empty
// working directory. The producer still admits its own closed typed input;
// this adapter never adds defaults or discovers files on its behalf.
const [inputPath, outputPath] = process.argv.slice(2);
if (!inputPath || !outputPath) throw new Error("explicit input and output required");
const absolute = path.resolve(inputPath);
const raw = new TextDecoder("utf-8", { fatal: true }).decode(readRetainedArtifactBytes(path.dirname(absolute), path.basename(absolute), 4194304));
const input = parseStrictJSON(raw);
// Do not turn a fractional or exponent-form number into an integer accepted by
// the producer's closed input. String tokens are skipped as complete tokens.
for (const token of raw.matchAll(/"(?:\\.|[^"\\])*"|-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?/gu)) {
  if (!token[0].startsWith('"') && (!/^(?:0|-?[1-9]\d*)$/u.test(token[0]) || !Number.isSafeInteger(Number(token[0])))) throw new Error("producer input requires exact safe integers");
}
const resolve = (object, fields) => {
  if (!object || typeof object !== "object" || Array.isArray(object)) return;
  for (const field of fields) {
    if (typeof object[field] === "string" && object[field] !== "") object[field] = path.resolve(object[field]);
  }
};
resolve(input, ["source_root", "source_lock", "source_profile", "field_policy", "source_inventory", "editorial", "license_decision", "output_policy", "output_root", "ledger_root", "approval"]);
resolve(input.envelope, ["logical_root", "output_root", "signing_context", "ledger_root", "previous_envelope"]);
writeFileSync(outputPath, JSON.stringify(input), { flag: "wx", mode: 0o600 });
