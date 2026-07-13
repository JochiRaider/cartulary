#!/usr/bin/env node

import {
  existsSync,
  readFileSync,
} from "node:fs";
import path from "node:path";
import { artifactFailureRecord } from "../../contract/failure-taxonomy.mjs";
import {
  repoRoot,
  validateSchemaSync,
} from "../../contract/index.mjs";

const govulncheckFindingsSchemaID = "cartulary.govulncheck_findings.v1";

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function relToRepo(value) {
  if (!value) {
    return "";
  }
  const normalized = normalizePath(value);
  if (!path.isAbsolute(value)) {
    return normalized;
  }
  const relative = normalizePath(path.relative(repoRoot, value));
  if (!relative.startsWith("../") && relative !== "..") {
    return relative;
  }
  return normalized;
}

function oneLineError(error) {
  return String(error instanceof Error ? error.message : error)
    .split(/\r?\n/u)[0]
    .trim();
}

export function loadGovulncheckFindingsFile(file) {
  if (!file || !existsSync(file)) {
    return { findings: null, error: null, artifact: "" };
  }
  const artifact = relToRepo(file);
  let findings;
  try {
    findings = JSON.parse(readFileSync(file, "utf8"));
  } catch (error) {
    return {
      findings: null,
      error: `invalid JSON: ${oneLineError(error)}`,
      artifact,
    };
  }
  try {
    validateSchemaSync(govulncheckFindingsSchemaID, findings);
  } catch (error) {
    return {
      findings: null,
      error: oneLineError(error),
      artifact,
    };
  }
  return { findings, error: null, artifact };
}

export function govulncheckArtifactFailure(file, error, defaults = {}) {
  const artifact = relToRepo(file);
  const detail = error ? `: ${error}` : "";
  return artifactFailureRecord(
    `invalid Govulncheck findings artifact ${artifact}${detail}`,
    {
      ...defaults,
      source: "govulncheck",
      label: "govulncheck-findings",
      artifact,
    },
  );
}
