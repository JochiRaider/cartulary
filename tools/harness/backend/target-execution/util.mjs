import path from "node:path";

export function compareStrings(left, right) {
  return String(left).localeCompare(String(right));
}

export function sleep(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

export function nowUTC() {
  return new Date().toISOString();
}

export function monotonicMs() {
  return Number(process.hrtime.bigint() / 1_000_000n);
}

export function clampDurationMs(value) {
  const numeric = Number.parseInt(String(value ?? "0"), 10);
  if (!Number.isInteger(numeric) || numeric < 0) {
    return 0;
  }
  return numeric;
}

export function captureStart() {
  return {
    startTime: nowUTC(),
    startMs: monotonicMs(),
  };
}

export function captureFinish(started) {
  return {
    startTime: started.startTime,
    endTime: nowUTC(),
    durationMs: clampDurationMs(monotonicMs() - started.startMs),
  };
}

export function resolvePath(repoRoot, value) {
  return path.isAbsolute(value) ? value : path.join(repoRoot, value);
}

export function relToRepo(ctx, value) {
  if (!value) {
    return "";
  }
  const normalized = String(value).replaceAll("\\", "/");
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = path.relative(ctx.repoRoot, value).replaceAll("\\", "/");
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

export function slugifyLabel(label) {
  return String(label)
    .toLowerCase()
    .replace(/[^a-z0-9]+/gu, "-")
    .replace(/^-+|-+$/gu, "")
    .replace(/--+/gu, "-");
}
