import {
  assertObjectKeys,
  requireEnum,
  requireRepoRelativePath,
  requireString,
  requireStringArray,
} from "../../contract/json-shape.mjs";
import {
  blockerKeys,
  ownerRefKeys,
  validOwnerResolutionStatuses,
  validOwnerRoles,
  validOwnerSourceKeys,
} from "./constants.mjs";

export function sectionNumbersFromOwnerRef(sectionRef) {
  const normalized = sectionRef
    .replace(/^Core\s+\d+\s+Sections?\s+/i, "")
    .replace(/^Sections?\s+/i, "");
  if (/\bthrough\b/.test(normalized)) {
    throw new Error(`section_ref must list exact sections, got ${sectionRef}`);
  }
  return [...normalized.matchAll(/\d+(?:\.\d+)*[A-Z]?/g)].map(
    (match) => match[0],
  );
}

export function validateOwnerRef(root, ownerRef, label, row) {
  assertObjectKeys(ownerRef, ownerRefKeys, label);
  const sourceKey = requireEnum(
    ownerRef.source_key,
    `${label}.source_key`,
    validOwnerSourceKeys,
  );
  const ownerPath = requireRepoRelativePath(ownerRef.path, `${label}.path`);
  requireString(ownerRef.section_ref, `${label}.section_ref`);
  requireString(ownerRef.heading_text, `${label}.heading_text`);
  requireStringArray(ownerRef.req_ids, `${label}.req_ids`);
  requireStringArray(ownerRef.ac_ids, `${label}.ac_ids`);
  requireEnum(ownerRef.role, `${label}.role`, validOwnerRoles);
  const resolutionStatus = requireEnum(
    ownerRef.resolution_status,
    `${label}.resolution_status`,
    validOwnerResolutionStatuses,
  );
  if (resolutionStatus !== "resolved" && row.claim_status !== "blocked") {
    throw new Error(
      `${label} unresolved owner refs are valid only on blocked rows`,
    );
  }
  if (sourceKey.startsWith("core") && resolutionStatus === "resolved") {
    const sectionNumbers = sectionNumbersFromOwnerRef(ownerRef.section_ref);
    if (sectionNumbers.length === 0) {
      throw new Error(`${label}.section_ref must name at least one exact section`);
    }
  }
  void root;
  void ownerPath;
  return ownerRef;
}

export function validateBlocker(blocker, label) {
  assertObjectKeys(blocker, blockerKeys, label);
  requireString(blocker.blocker_id, `${label}.blocker_id`);
  requireString(blocker.reason_code, `${label}.reason_code`);
  requireString(blocker.description, `${label}.description`);
  requireString(blocker.resolution_owner, `${label}.resolution_owner`);
  return blocker;
}
