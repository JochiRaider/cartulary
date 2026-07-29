import { semanticJSONSHA256 } from "./semantic-json.mjs";

const familyIDPattern =
  /^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,62}$/u;
const rowIDPattern =
  /^(?:module|platform|app|web|package|harness)\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,62}\.[a-z][a-z0-9_]{0,127}_[0-9a-f]{10}$/u;

function normalizedSemanticText(value, label) {
  const normalized = String(value ?? "").trim();
  if (normalized.length === 0 || normalized.length > 500) {
    throw new Error(`${label} must contain 1..500 characters after trimming`);
  }
  if (/[\u0000-\u001f\u007f]/u.test(normalized)) {
    throw new Error(`${label} must not contain control characters`);
  }
  return normalized;
}

function semanticSlug(value) {
  return value
    .normalize("NFKD")
    .replace(/[\u0300-\u036f]/gu, "")
    .replace(/([a-z0-9])([A-Z])/gu, "$1 $2")
    .replace(/[^A-Za-z0-9]+/gu, "_")
    .replace(/^_+|_+$/gu, "")
    .toLowerCase()
    .replace(/_+/gu, "_");
}

export function deriveTestRowID({ familyID, claim, selectorKey }) {
  const family = String(familyID ?? "").trim();
  if (!familyIDPattern.test(family)) {
    throw new Error("FAMILY_ID must be an owner-qualified test family ID");
  }
  const normalizedClaim = normalizedSemanticText(claim, "CLAIM");
  const normalizedSelector = normalizedSemanticText(selectorKey, "SELECTOR_KEY");
  const digest = semanticJSONSHA256({
    claim: normalizedClaim,
    family_id: family,
    selector_key: normalizedSelector,
  }).slice(0, 10);
  const availableStemLength = 191 - family.length - 1 - 1 - digest.length;
  if (availableStemLength < 1) {
    throw new Error("FAMILY_ID leaves no room for a semantic row segment");
  }
  const stem = (semanticSlug(normalizedClaim) || "behavior")
    .slice(0, Math.min(48, availableStemLength))
    .replace(/_+$/u, "") || "behavior";
  const rowID = `${family}.${stem}_${digest}`;
  if (!rowIDPattern.test(rowID) || rowID.length > 191) {
    throw new Error("derived row ID does not satisfy the test-catalog contract");
  }
  return rowID;
}
