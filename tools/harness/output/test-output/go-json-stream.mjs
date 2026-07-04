import { readFileSync } from "node:fs";

export function handleGoJSONStream() {
  flushGoJSONStream(readFileSync(0, "utf8"), true);
  return 0;
}

export function flushGoJSONStream(buffer, flushAll) {
  const lines = buffer.split(/\r?\n/);
  const pending = flushAll ? "" : (lines.pop() ?? "");
  for (const line of lines) {
    if (!line.trim()) {
      continue;
    }
    try {
      const entry = JSON.parse(line);
      if (typeof entry.Output === "string") {
        process.stdout.write(entry.Output);
      }
    } catch {
      // Go's -json stream is expected on stdout; ignore malformed lines.
    }
  }
  if (flushAll && pending.trim()) {
    try {
      const entry = JSON.parse(pending);
      if (typeof entry.Output === "string") {
        process.stdout.write(entry.Output);
      }
    } catch {
      // Ignore trailing malformed output.
    }
  }
  return pending;
}
