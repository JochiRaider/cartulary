import path from "node:path";
import { fileURLToPath } from "node:url";

import { loadTestCatalog } from "../test-catalog/index.mjs";

const scriptDir = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(scriptDir, "..", "..", "..");

function normalizeManifestFile(file) {
  const normalized = String(file ?? "").replaceAll("\\", "/");
  if (normalized.startsWith("apps/web/e2e/")) {
    return normalized;
  }
  if (normalized.startsWith("e2e/")) {
    return `apps/web/${normalized}`;
  }
  return normalized;
}

function compareEntries(left, right) {
  if (left.stage !== right.stage) {
    return left.stage.localeCompare(right.stage);
  }
  if (left.file !== right.file) {
    return left.file.localeCompare(right.file);
  }
  if (left.title !== right.title) {
    return left.title.localeCompare(right.title);
  }
  return left.id.localeCompare(right.id);
}

function browserEntry(row) {
  return {
    id: row.row_id,
    stage: row.selector.stage,
    file: normalizeManifestFile(row.selector.file),
    title: row.selector.titles[0],
    titles: [...row.selector.titles],
    execution_dependency: "browser_functional",
    default_check_required: row.default_check,
    runtime_profile_id: row.runtime_profile_id,
  };
}

function browserRows(root = repoRoot) {
  return loadTestCatalog(root).rows.filter(
    (row) =>
      row.runner === "playwright" &&
      row.selector.stage === "webserver_backed" &&
      row.status === "active",
  );
}

export function browserDurationBaselineEntries(root = repoRoot) {
  return browserRows(root).map(browserEntry).sort(compareEntries);
}

export function selectedEntriesForPlan(
  root = repoRoot,
  { stage = "", frontendRowIDs = new Set(), defaultCheckOnly = false } = {},
) {
  if (stage || frontendRowIDs.size > 0) {
    throw new Error(
      "stage and frontend-row browser selection are retired; select semantic catalog rows",
    );
  }
  return browserRows(root)
    .filter((row) => !defaultCheckOnly || row.default_check)
    .map(browserEntry)
    .sort(compareEntries);
}

export function browserDefaultCheckRowIndex(root = repoRoot) {
  return new Map(
    loadTestCatalog(root).rows
      .filter((row) => row.runner === "playwright" && row.default_check)
      .map((row) => [
        row.row_id,
        {
          id: row.row_id,
          stage: row.selector.stage,
          runtime_profile_id: row.runtime_profile_id,
          resource_profile_id: row.resource_profile_id,
          fixture_profile_id: row.fixture_profile_id,
        },
      ]),
  );
}
