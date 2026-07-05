import path from "node:path";

import { requireString } from "../../contract/json-shape.mjs";
import {
  phaseIDPattern,
  phaseLedgerFilenamePattern,
  phaseMapFilenamePattern,
} from "./constants.mjs";

export function repoPath(root, relativePath) {
  return path.join(root, relativePath);
}

export function phaseNumber(phaseID) {
  const match = /^FE-P(0|[1-9]\d*)$/.exec(phaseID);
  if (!match) {
    throw new Error(`frontend phase id ${phaseID} must match FE-P<N>`);
  }
  return match[1];
}

export function phaseFromMapPath(manifestPath, label) {
  const match = phaseMapFilenamePattern.exec(path.posix.basename(manifestPath));
  if (!match) {
    throw new Error(`${label} must end with fe_p<N>_test_map.json`);
  }
  return `FE-P${match[1]}`;
}

export function phaseFromLedgerPath(ledgerPath, label) {
  const match = phaseLedgerFilenamePattern.exec(
    path.posix.basename(ledgerPath),
  );
  if (!match) {
    throw new Error(`${label} must end with fe_p<N>_coverage_ledger.md`);
  }
  return `FE-P${match[1]}`;
}

export function requirePhaseID(value, label) {
  return requireString(value, label, { pattern: phaseIDPattern });
}

export function entryTitles(entry) {
  if (Array.isArray(entry?.titles)) {
    return entry.titles.filter((title) => typeof title === "string");
  }
  return typeof entry?.title === "string" ? [entry.title] : [];
}
