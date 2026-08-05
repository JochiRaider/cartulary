import { mkdirSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";

function normalizeEnvPath(value: string | undefined) {
  const trimmed = value?.trim() ?? "";
  return trimmed === "" ? null : trimmed;
}

export function isExternalServerHarnessMode() {
  return process.env.CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER === "1";
}

export function sharedPlaywrightStateDir() {
  return normalizeEnvPath(process.env.CARTULARY_PLAYWRIGHT_STATE_DIR);
}

export function usesSharedPlaywrightState() {
  return sharedPlaywrightStateDir() !== null;
}

function ensureSharedPlaywrightStateDir() {
  const directory = sharedPlaywrightStateDir();
  if (directory === null) {
    return null;
  }
  mkdirSync(directory, { recursive: true });
  return directory;
}

export function resolvePlaywrightStateFile(fileName: string) {
  const directory = ensureSharedPlaywrightStateDir();
  return directory === null
    ? join(tmpdir(), fileName)
    : join(directory, fileName);
}

export function resolvePlaywrightStateDirectory(directoryName: string) {
  const directory = ensureSharedPlaywrightStateDir();
  return directory === null
    ? join(tmpdir(), directoryName)
    : join(directory, directoryName);
}
