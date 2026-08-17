import {
  closeSync,
  createWriteStream,
  mkdirSync,
  openSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { rename, rm } from "node:fs/promises";
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

export function createAtomicNDJSONWriter(file, serialize = JSON.stringify) {
  mkdirSync(path.dirname(file), { recursive: true, mode: 0o700 });
  const temporary = `${file}.tmp-${process.pid}-${temporaryCounter++}`;
  const stream = createWriteStream(temporary, {
    encoding: "utf8",
    flags: "wx",
    highWaterMark: 64 * 1024,
    mode: 0o600,
  });
  let closed = false;
  let failed = null;
  stream.on("error", (error) => {
    failed ??= error;
  });
  return {
    stagingFile: temporary,
    async write(record) {
      if (closed) throw new Error("canonical event writer is closed");
      if (failed) throw failed;
      await new Promise((resolve, reject) => {
        stream.write(`${serialize(record)}\n`, (error) => {
          if (error) reject(error);
          else resolve();
        });
      });
      if (failed) throw failed;
    },
    async close() {
      if (closed) throw new Error("canonical event writer is already closed");
      closed = true;
      try {
        await new Promise((resolve, reject) => {
          stream.end((error) => {
            if (error) reject(error);
            else resolve();
          });
        });
        if (failed) throw failed;
        await rename(temporary, file);
      } catch (error) {
        stream.destroy();
        await rm(temporary, { force: true });
        throw error;
      }
    },
    async abort() {
      if (closed) return;
      closed = true;
      stream.destroy();
      await rm(temporary, { force: true });
    },
  };
}
