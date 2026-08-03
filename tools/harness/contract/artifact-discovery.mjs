import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const rootDefault = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../../..");

function resultsRoot(root) {
  const configured = process.env.CARTULARY_TEST_RESULTS_DIR || ".cartulary/test-results";
  return path.isAbsolute(configured) ? configured : path.join(root, configured);
}

function readJSON(file) {
  try {
    return JSON.parse(readFileSync(file, "utf8"));
  } catch {
    return null;
  }
}

function rel(root, file) {
  const value = path.relative(root, file).replaceAll("\\", "/");
  return value.startsWith("../") || value === ".." ? file.replaceAll("\\", "/") : value;
}

function candidates(target, root) {
  const base = resultsRoot(root);
  if (!existsSync(base)) return [];
  return readdirSync(base, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => {
      const file = path.join(base, entry.name, "target-summaries", `${target}.json`);
      const summary = existsSync(file) ? readJSON(file) : null;
      if (summary?.schema_id !== "cartulary.harness_target_summary.v1" || summary.target !== target) {
        return null;
      }
      return {
        kind: "canonical_target_summary",
        path: rel(root, file),
        run_id: entry.name,
        mtime_ms: statSync(file).mtimeMs,
        status: summary.status,
      };
    })
    .filter(Boolean)
    .sort(
      (left, right) =>
        right.mtime_ms - left.mtime_ms || left.path.localeCompare(right.path),
    );
}

export function newestTargetArtifact(target, { root = rootDefault } = {}) {
  const values = candidates(target, root);
  return {
    latest: values[0] ?? null,
    candidates: values,
    expected: [
      rel(
        root,
        path.join(resultsRoot(root), "<run-id>", "target-summaries", `${target}.json`),
      ),
    ],
  };
}

export function helperArtifactReferences(
  helperTargets,
  { root = rootDefault, runId = "" } = {},
) {
  const selectedRun = runId || process.env.CARTULARY_TEST_RUN_ID || "";
  if (!selectedRun) return [];
  return helperTargets
    .map((target) => {
      const file = path.join(resultsRoot(root), selectedRun, "target-summaries", `${target}.json`);
      const summary = existsSync(file) ? readJSON(file) : null;
      if (summary?.schema_id !== "cartulary.harness_target_summary.v1") return null;
      return {
        target,
        latest: rel(root, file),
        step_summaries: [
          {
            label: target,
            status: summary.status,
            artifact: rel(root, file),
            runner_json: "",
            stdout_log: "",
            stderr_log: "",
          },
        ],
      };
    })
    .filter(Boolean);
}
