import { randomUUID } from "node:crypto";
import {
  closeSync,
  fsyncSync,
  openSync,
  renameSync,
  unlinkSync,
  writeFileSync,
} from "node:fs";

type AtomicWriteOptions = {
  beforeRename?: (temporaryPath: string) => void;
};

export function atomicWritePrivateFile(
  destinationPath: string,
  contents: string,
  options: AtomicWriteOptions = {},
) {
  const temporaryPath = `${destinationPath}.${process.pid}.${randomUUID()}.tmp`;
  let descriptor: number | null = null;
  try {
    descriptor = openSync(temporaryPath, "wx", 0o600);
    writeFileSync(descriptor, contents, "utf8");
    fsyncSync(descriptor);
    closeSync(descriptor);
    descriptor = null;
    options.beforeRename?.(temporaryPath);
    renameSync(temporaryPath, destinationPath);
  } catch (error) {
    if (descriptor !== null) {
      try {
        closeSync(descriptor);
      } catch {
        // Preserve the publication failure.
      }
    }
    try {
      unlinkSync(temporaryPath);
    } catch {
      // The file may already have been atomically published or removed.
    }
    throw error;
  }
}
