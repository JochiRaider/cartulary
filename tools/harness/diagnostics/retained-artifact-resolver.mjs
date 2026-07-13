import {
  lstatSync,
  readdirSync,
} from "node:fs";
import path from "node:path";

const retainedLogReadLimits = Object.freeze({
  maxFiles: 4096,
  maxFileBytes: 16 * 1024 * 1024,
  maxTotalBytes: 256 * 1024 * 1024,
});

const logFileFormats = new Set(["log", "text"]);
const safeSegmentPattern = /^[A-Za-z0-9._@+,:=-]+$/u;

function artifactError(message) {
  const error = new Error(message);
  error.failure_class = "artifact";
  error.failure_reason = "artifact_error";
  error.exit_code = 11;
  return error;
}

function validateRelativePath(value, label) {
  if (typeof value !== "string" || value.length === 0 || path.isAbsolute(value)) {
    throw artifactError(`${label}.path must be a non-empty run-root-relative path`);
  }
  const normalized = value.replaceAll("\\", "/");
  const segments = normalized.split("/");
  if (
    normalized !== value ||
    segments.some(
      (segment) =>
        segment === "" ||
        segment === "." ||
        segment === ".." ||
        !safeSegmentPattern.test(segment),
    )
  ) {
    throw artifactError(`${label}.path is not a normalized safe relative path`);
  }
  return segments;
}

function resolveWithoutSymlinks(runDir, ref, label) {
  const segments = validateRelativePath(ref.path, label);
  let current = path.resolve(runDir);
  const root = current;
  for (const segment of segments) {
    current = path.join(current, segment);
    let stat;
    try {
      stat = lstatSync(current);
    } catch {
      throw artifactError(`${label}.path does not exist: ${ref.path}`);
    }
    if (stat.isSymbolicLink()) {
      throw artifactError(`${label}.path contains a symlink: ${ref.path}`);
    }
  }
  const relative = path.relative(root, current);
  if (relative === "" || relative === ".." || relative.startsWith(`..${path.sep}`)) {
    throw artifactError(`${label}.path escapes the selected run root`);
  }
  return current;
}

function validateReferenceShape(ref, label) {
  if (!ref || typeof ref !== "object" || Array.isArray(ref)) {
    throw artifactError(`${label} must be a structured retained harness artifact reference`);
  }
  if (typeof ref.role !== "string" || ref.role.length === 0) {
    throw artifactError(`${label}.role must be non-empty`);
  }
  if (ref.path_kind !== "file" && ref.path_kind !== "directory") {
    throw artifactError(`${label}.path_kind must be file or directory`);
  }
  if (ref.path_kind === "file" && !logFileFormats.has(ref.format)) {
    throw artifactError(`${label}.format must be log or text for log replay`);
  }
  if (ref.path_kind === "directory" && Object.hasOwn(ref, "format")) {
    throw artifactError(`${label}.format is forbidden for directory references`);
  }
}

function validateFile(file, label, limits) {
  const stat = lstatSync(file);
  if (!stat.isFile()) {
    throw artifactError(`${label} must resolve to a regular file`);
  }
  if (stat.size > limits.maxFileBytes) {
    throw artifactError(`${label} exceeds the ${limits.maxFileBytes}-byte file limit`);
  }
  return stat.size;
}

function directoryLogFiles(directory, label, limits) {
  const directoryStat = lstatSync(directory);
  if (!directoryStat.isDirectory()) {
    throw artifactError(`${label} must resolve to a directory`);
  }
  const files = [];
  let totalBytes = 0;
  for (const entry of readdirSync(directory, { withFileTypes: true }).sort((left, right) =>
    left.name.localeCompare(right.name),
  )) {
    if (entry.isSymbolicLink()) {
      throw artifactError(`${label} contains a symlink entry: ${entry.name}`);
    }
    if (!entry.isFile() || !entry.name.endsWith(".log")) {
      continue;
    }
    const file = path.join(directory, entry.name);
    totalBytes += validateFile(file, `${label}/${entry.name}`, limits);
    files.push(file);
    if (files.length > limits.maxFiles) {
      throw artifactError(`${label} exceeds the ${limits.maxFiles}-file directory limit`);
    }
    if (totalBytes > limits.maxTotalBytes) {
      throw artifactError(`${label} exceeds the ${limits.maxTotalBytes}-byte aggregate limit`);
    }
  }
  return { files, totalBytes };
}

export function resolveRetainedLogArtifacts(
  runDir,
  refs,
  limits = retainedLogReadLimits,
) {
  const resolved = [];
  let totalBytes = 0;
  for (const [index, ref] of (refs ?? []).entries()) {
    const label = `log_artifacts[${index + 1}]`;
    validateReferenceShape(ref, label);
    const target = resolveWithoutSymlinks(runDir, ref, label);
    if (ref.path_kind === "file") {
      totalBytes += validateFile(target, label, limits);
      resolved.push({ role: ref.role, file: target });
    } else {
      const directory = directoryLogFiles(target, label, limits);
      totalBytes += directory.totalBytes;
      for (const file of directory.files) {
        resolved.push({ role: ref.role, file });
      }
    }
    if (resolved.length > limits.maxFiles) {
      throw artifactError(`selected log artifacts exceed the ${limits.maxFiles}-file limit`);
    }
    if (totalBytes > limits.maxTotalBytes) {
      throw artifactError(`selected log artifacts exceed the ${limits.maxTotalBytes}-byte aggregate limit`);
    }
  }
  return resolved;
}
