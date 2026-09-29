import { readFileSync } from "node:fs";
import path from "node:path";

// Authored machine projections only. Control can load these facts without
// resolving browser/native packages or interpreting a document.
export const repoRoot = path.resolve(import.meta.dirname, "../../../..");
const read = (file) => JSON.parse(readFileSync(path.join(repoRoot, file), "utf8"));
let pins;
export function reviewPins() {
  return pins ??= Object.freeze(read("tools/toolchain_pins.json").ui_review);
}
