import { readFileSync } from "node:fs";
import path from "node:path";

const implementationTestingGuidePath = path.join(
  "docs",
  "guides",
  "cartulary_implementation_testing_guide.md",
);

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, String.raw`\$&`);
}

function phaseIDPatternSource(layerPrefix, phaseNumber, separator) {
  const normalizedLayerPrefix = layerPrefix.endsWith(separator)
    ? layerPrefix
    : `${layerPrefix}${separator}`;
  return `${escapeRegex(normalizedLayerPrefix)}${phaseNumber}${escapeRegex(
    separator,
  )}(?:[A-Z0-9]+${escapeRegex(separator)})*\\d{2}`;
}

export function phaseIDRegex(layerPrefix, phaseNumber) {
  return new RegExp(`^${phaseIDPatternSource(layerPrefix, phaseNumber, "-")}$`);
}

function claimedPhaseIDRegex(phaseNumber, separator) {
  return new RegExp(
    String.raw`\b[UIEV]${escapeRegex(separator)}${phaseNumber}${escapeRegex(
      separator,
    )}(?:[A-Z0-9]+${escapeRegex(separator)})*\d{2}\b`,
    "g",
  );
}

export function validateExpectedIDs(expectedIDs, phaseNumber, manifestPath) {
  const seen = new Set();
  for (const id of expectedIDs) {
    if (typeof id !== "string" || id.trim() === "") {
      throw new Error(`manifest ${manifestPath} has an invalid expected_id: ${JSON.stringify(id)}`);
    }
    const layerPrefix = `${id[0] ?? ""}-`;
    if (!phaseIDRegex(layerPrefix, phaseNumber).test(id)) {
      throw new Error(`manifest ${manifestPath} has expected_id ${id} that does not belong to phase${phaseNumber}`);
    }
    if (seen.has(id)) {
      throw new Error(`manifest ${manifestPath} has duplicate expected_id ${id}`);
    }
    seen.add(id);
  }
}

export function loadGuideExpectedIDs(root, phaseNumber) {
  const source = readFileSync(path.join(root, implementationTestingGuidePath), "utf8");
  const pattern = new RegExp(
    String.raw`^\|\s*([UIEV]-${phaseNumber}(?:-[A-Z0-9]+)*-\d{2})\s*\|`,
  );
  const ids = new Set();
  for (const line of source.split(/\r?\n/)) {
    const match = pattern.exec(line);
    if (match?.[1]) {
      ids.add(match[1]);
    }
  }
  return Array.from(ids).sort();
}

export function extractClaimedPhaseIDs(source, phaseNumber) {
  const hyphenMatches = source.match(claimedPhaseIDRegex(phaseNumber, "-")) ?? [];
  const underscoreMatches = source.match(claimedPhaseIDRegex(phaseNumber, "_")) ?? [];
  return new Set([
    ...hyphenMatches,
    ...underscoreMatches.map((value) => value.replaceAll("_", "-")),
  ]);
}
