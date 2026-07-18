import {
  assertObjectKeys,
  requireEnum,
  requireString,
} from "../../contract/json-shape.mjs";
import {
  claimKeys,
  core03Section14OwnerRefPattern,
  core03Section48OwnerRefPattern,
  core03SortingFilteringGroupingReqIDs,
  validClaimPublicationIntents,
  validClosureScopes,
} from "./constants.mjs";
import { sectionNumbersFromOwnerRef } from "./owner-refs.mjs";
import { baseAuthoritativePlaywrightTitleIndex } from "./source-index.mjs";

export function validateClaim(claim, label, row) {
  assertObjectKeys(claim, claimKeys, label);
  requireString(claim.statement, `${label}.statement`);
  requireEnum(
    claim.claim_publication_intent,
    `${label}.claim_publication_intent`,
    validClaimPublicationIntents,
  );
  requireEnum(claim.closure_scope, `${label}.closure_scope`, validClosureScopes);
  return claim;
}

export function validateRowMetadata(row, label) {
  const evidenceClass = row.evidence_class;
  if (evidenceClass === "product_conformance") {
    if (row.core_req_ids.length === 0 || row.core_ac_ids.length === 0) {
      throw new Error(
        `${label} product_conformance rows must declare core_req_ids[] and core_ac_ids[]`,
      );
    }
    return;
  }
  if (
    evidenceClass === "design_direction" ||
    evidenceClass === "implementation_support"
  ) {
    if (row.core_req_ids.length !== 0 || row.core_ac_ids.length !== 0) {
      throw new Error(
        `${label} ${evidenceClass} rows must not declare Core requirement or acceptance IDs`,
      );
    }
    if (row.support_or_design_ac_ids.length === 0) {
      throw new Error(
        `${label} ${evidenceClass} rows must declare support_or_design_ac_ids[]`,
      );
    }
    return;
  }
  if (evidenceClass === "claim_publication_boundary") {
    for (const id of row.core_req_ids) {
      if (!id.startsWith("REQ-05-")) {
        throw new Error(
          `${label} claim_publication_boundary Core IDs must be Core 05 IDs`,
        );
      }
    }
    return;
  }
  if (evidenceClass === "TODO_owner_lookup") {
    throw new Error(
      `${label} TODO_owner_lookup rows are not valid in current frontend maps`,
    );
  }
}

export function validateVisualAccessibilityEvidenceBoundary(row, label) {
  const visualOrAccessibility =
    row.layer === "visual" ||
    row.layer === "accessibility" ||
    row.id.startsWith("FE-V-") ||
    row.id.startsWith("FE-A11Y-");
  if (!visualOrAccessibility) {
    return;
  }
  if (row.evidence_class === "product_conformance") {
    throw new Error(
      `${label} FE visual and accessibility rows must not use product_conformance; use design_direction, implementation_support, or an explicit Core 05 claim_publication_boundary route`,
    );
  }
  if (row.evidence_class !== "claim_publication_boundary") {
    return;
  }
  const hasCore05Owner = row.owner_refs.some(
    (ownerRef) =>
      ownerRef.source_key === "core05" &&
      ownerRef.role === "claim_publication_owner" &&
      ownerRef.resolution_status === "resolved",
  );
  if (!hasCore05Owner) {
    throw new Error(
      `${label} FE visual and accessibility claim_publication_boundary rows must cite a resolved Core 05 claim_publication_owner`,
    );
  }
  if (row.claim.claim_publication_intent !== "claim_bearing_publication") {
    throw new Error(
      `${label} FE visual and accessibility claim_publication_boundary rows must declare claim_publication_intent=claim_bearing_publication`,
    );
  }
}

export function validateCore03SortingFilteringGroupingOwnerRefs(row, label) {
  const coversSortingFilteringGrouping = row.core_req_ids.some((id) =>
    core03SortingFilteringGroupingReqIDs.has(id),
  );
  if (!coversSortingFilteringGrouping) {
    return;
  }

  const citesCore03Section14 = row.owner_refs.some((ownerRef) =>
    ownerRef.source_key === "core03" &&
    (sectionNumbersFromOwnerRef(ownerRef.section_ref).includes("14") ||
      core03Section14OwnerRefPattern.test(ownerRef.heading_text)),
  );
  const citesCore03Section48 = row.owner_refs.some((ownerRef) =>
    ownerRef.source_key === "core03" &&
    (sectionNumbersFromOwnerRef(ownerRef.section_ref).includes("4.8") ||
      core03Section48OwnerRefPattern.test(ownerRef.heading_text)),
  );
  if (!citesCore03Section14 || citesCore03Section48) {
    throw new Error(
      `${label} rows covering REQ-03-223..REQ-03-235 must cite Core 03 Section 14 and must not cite Core 03 Section 4.8`,
    );
  }
}

function requiresBrowserScenarioTitleOwnership(row) {
  return (
    row.claim_status === "implemented" &&
    row.targets.some(
      (target) =>
        target.target_name.startsWith("browser-e2e") &&
        target.scenario_title_required,
    )
  );
}

export function validateFrontendBrowserScenarioTitleOwnership(
  root,
  row,
  label,
  titleOwners,
) {
  if (!requiresBrowserScenarioTitleOwnership(row)) {
    return;
  }
  const baseTitles = baseAuthoritativePlaywrightTitleIndex(root);
  for (const title of row.scenario_titles) {
    const baseOwner = baseTitles.get(title);
    if (baseOwner) {
      throw new Error(
        `${label}.scenario_titles reuses base authoritative Playwright title owned by ${baseOwner}`,
      );
    }
    const existingOwner = titleOwners.get(title);
    if (existingOwner && existingOwner !== row.id) {
      throw new Error(
        `${label}.scenario_titles duplicates frontend browser title owned by ${existingOwner}`,
      );
    }
    if (
      /^(?:FE-[A-Z]+-P\d+(?:-[A-Z0-9]+)*|[UIEV]-\d+(?:-[A-Z0-9]+)+)\s+/u.test(title)
      || /^(?:support\s+)?(?:Phase|Sprint)\s+\d+\s+/iu.test(title)
    ) {
      throw new Error(
        `${label}.scenario_titles must use semantic titles without delivery or legacy row prefixes`,
      );
    }
    titleOwners.set(title, row.id);
  }
}
