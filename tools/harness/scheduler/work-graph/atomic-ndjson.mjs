import {
  closeSync,
  mkdirSync,
  openSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

let temporaryCounter = 0;

// writeAtomicNDJSON serializes and writes one record at a time so evidence size
// is bounded by the largest individual event rather than by the complete run.
export function writeAtomicNDJSON(file, records, serialize = JSON.stringify) {
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  const temporary = `${file}.tmp-${process.pid}-${temporaryCounter++}`;
  let descriptor;
  try {
    descriptor = openSync(temporary, "wx", 0o600);
    for (const record of records) {
      writeFileSync(descriptor, `${serialize(record)}\n`);
    }
    closeSync(descriptor);
    descriptor = undefined;
    renameSync(temporary, file);
  } catch (error) {
    if (descriptor !== undefined) {
      try {
        closeSync(descriptor);
      } catch {
        // Preserve the original write failure.
      }
    }
    rmSync(temporary, { force: true });
    throw error;
  }
}
