import { existsSync, readFileSync } from "node:fs";

import {
  assertObjectKeys,
  requireEnum,
  requireRepoRelativePath,
  requireString,
  requireStringArray,
} from "../../contract/json-shape.mjs";
import { repoPath } from "./common.mjs";
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

function coreHeadingMap(text) {
  const headings = new Map();
  for (const match of text.matchAll(
    /^#{2,6}\s+([0-9]+(?:\.[0-9]+)*[A-Z]?)\.?\s+(.+)$/gm,
  )) {
    headings.set(match[1], `${match[1]} ${match[2].trim()}`);
  }
  return headings;
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
    const absolute = repoPath(root, ownerPath);
    if (!existsSync(absolute)) {
      throw new Error(`${label}.path does not exist: ${ownerPath}`);
    }
    const text = readFileSync(absolute, "utf8");
    const sectionNumbers = sectionNumbersFromOwnerRef(ownerRef.section_ref);
    if (sectionNumbers.length === 0) {
      throw new Error(`${label}.section_ref must name at least one exact section`);
    }
    const headings = coreHeadingMap(text);
    const expectedHeadingText = sectionNumbers
      .map((sectionNumber) => {
        const heading = headings.get(sectionNumber);
        if (!heading) {
          throw new Error(
            `${label}.section_ref references missing section ${sectionNumber}`,
          );
        }
        return heading;
      })
      .join("; ");
    if (ownerRef.heading_text !== expectedHeadingText) {
      throw new Error(
        `${label}.heading_text must match resolved headings for ${ownerRef.section_ref}`,
      );
    }
    for (const reqID of ownerRef.req_ids) {
      if (!text.includes(`**${reqID}**`)) {
        throw new Error(`${label}.req_ids references missing ${reqID}`);
      }
    }
    for (const acID of ownerRef.ac_ids) {
      if (!text.includes(acID)) {
        throw new Error(`${label}.ac_ids references missing ${acID}`);
      }
    }
  }
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
