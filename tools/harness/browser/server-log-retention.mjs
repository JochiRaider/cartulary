// One retention path for every browser server generation (TH-HARNESS-REQ-861).
// The private log directory carries a marker naming the generation and the
// cohort that run against the server whose output currently occupies
// server.log. Reset retains a replaced generation; final settlement retains the
// last one. Both copy a bounded tail through harness redaction
// (TH-HARNESS-REQ-600/601) into owner-only session evidence.
import { closeSync, constants, fstatSync, openSync, readSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { parseStrictJSON, redactString } from "../contract/index.mjs";
import { atomicLocalFile, privateDirectory, readLocalFile } from "../runtime/secure-local-files.mjs";

export const serverGenerationMarker = "server-generation.json";
export const retainedLogTailBytes = 1048576;
const logNames = ["server.log", "web.log"];
const cohortLabel = /^[A-Za-z0-9][A-Za-z0-9_.-]{0,199}$/u;
const sessionArtifactSuffix = ["_shared", "test-services", null, "browser-sessions", null];

function invalid(message) { return Object.assign(new Error(message), { failure_class: "artifact", failure_reason: "artifact_error" }); }

function inside(parent, child) {
  const relative = path.relative(path.resolve(parent), path.resolve(child));
  return relative !== "" && !relative.startsWith("..") && !path.isAbsolute(relative);
}

// A session artifact directory is <run root>/_shared/test-services/<suite>/browser-sessions/<session>.
function validArtifactDirectory(directory, runRoot) {
  if (typeof directory !== "string" || !path.isAbsolute(directory)) return false;
  const parts = path.resolve(directory).split(path.sep).slice(-sessionArtifactSuffix.length);
  if (parts.length !== sessionArtifactSuffix.length || sessionArtifactSuffix.some((part, index) => part !== null && parts[index] !== part)) return false;
  if (parts.some((part) => !/^[A-Za-z0-9_.-]{1,255}$/u.test(part) || part === "." || part === "..")) return false;
  return runRoot === undefined || inside(runRoot, directory);
}

export function generationFolder(generation, cohort) {
  return `generation-${String(generation).padStart(3, "0")}-${cohort}`;
}

// Mark the log directory before a backend generation is spawned. The log is
// truncated here so a failed spawn can never leave predecessor output under
// the new generation's label.
export function markServerGeneration(logDirectory, { generation, cohort, artifactDirectory }) {
  if (!Number.isSafeInteger(generation) || generation < 1 || generation > 999999 || !cohortLabel.test(cohort) || !validArtifactDirectory(artifactDirectory)) {
    throw invalid("invalid server generation marker");
  }
  privateDirectory(logDirectory);
  closeSync(openSync(path.join(logDirectory, "server.log"), constants.O_WRONLY | constants.O_CREAT | constants.O_TRUNC | constants.O_NOFOLLOW | constants.O_NONBLOCK, 0o600));
  atomicLocalFile(path.join(logDirectory, serverGenerationMarker), `${JSON.stringify({ generation, cohort, artifact_directory: path.resolve(artifactDirectory) })}\n`, { replace: true, maximumReplacementBytes: 4096 });
}

function readMarker(logDirectory) {
  let bytes;
  try { bytes = readLocalFile(path.join(logDirectory, serverGenerationMarker), { maximum: 4096 }); }
  catch (error) { if (error.code === "ENOENT") return null; throw error; }
  const marker = parseStrictJSON(bytes.toString("utf8"));
  if (!marker || Array.isArray(marker) || Object.keys(marker).sort().join(",") !== "artifact_directory,cohort,generation" ||
      !Number.isSafeInteger(marker.generation) || marker.generation < 1 || !cohortLabel.test(marker.cohort)) throw invalid("invalid server generation marker");
  return marker;
}

function readTail(file) {
  let descriptor;
  try { descriptor = openSync(file, constants.O_RDONLY | constants.O_NOFOLLOW | constants.O_NONBLOCK); }
  catch (error) { if (error.code === "ENOENT") return null; throw error; }
  try {
    const info = fstatSync(descriptor);
    if (!info.isFile() || info.uid !== process.getuid() || (info.mode & 0o077) !== 0) throw invalid("unsafe browser process log");
    const length = Math.min(info.size, retainedLogTailBytes);
    const bytes = Buffer.alloc(length);
    let offset = 0;
    while (offset < length) {
      const count = readSync(descriptor, bytes, offset, length - offset, info.size - length + offset);
      if (!count) break;
      offset += count;
    }
    return bytes.subarray(0, offset);
  } finally { closeSync(descriptor); }
}

// Retain the marked generation once. An existing destination is that same
// generation's earlier retention and is never replaced. Returns the folder, or
// null when no backend generation was ever started.
export function retainServerGeneration(logDirectory, { runRoot } = {}) {
  const marker = readMarker(logDirectory);
  if (!marker) return null;
  if (!validArtifactDirectory(marker.artifact_directory, runRoot)) throw invalid("server generation marker names a foreign artifact directory");
  const folder = path.join(path.resolve(marker.artifact_directory), "process-diagnostics", generationFolder(marker.generation, marker.cohort));
  privateDirectory(folder);
  for (const name of logNames) {
    const tail = readTail(path.join(logDirectory, name));
    if (tail === null) continue;
    try { atomicLocalFile(path.join(folder, name), redactString(tail.toString("utf8"))); }
    catch (error) { if (error.code !== "EEXIST") throw error; }
  }
  return folder;
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const [operation, logDirectory, ...args] = process.argv.slice(2);
  try {
    if (operation === "mark" && args.length === 3) {
      markServerGeneration(logDirectory, { generation: Number(args[0]), cohort: args[1], artifactDirectory: args[2] });
    } else if (operation === "retain" && args.length === 0) {
      retainServerGeneration(logDirectory);
    } else throw invalid("unknown server log retention operation");
  } catch {
    process.stderr.write(`browser server log ${operation ?? "operation"} failed\n`);
    process.exitCode = 11;
  }
}
