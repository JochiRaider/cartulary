export function validateProfileClaims(manifest, entries, manifestPath) {
  const entryByID = new Map(entries.map((entry) => [entry.id, entry]));
  for (const claim of manifest.profile_claims ?? []) {
    const label = `manifest ${manifestPath} profile_claims.${claim.profile_id}`;
    if (!claim.required_ac_ids.includes(claim.claim_ac_id)) {
      throw new Error(`${label} claim_ac_id must be listed in required_ac_ids`);
    }
    for (const [field, values] of [
      ["required_ac_ids", claim.required_ac_ids],
      ["direct_evidence_ids", claim.direct_evidence_ids],
      ["aggregate_ac_ids", claim.aggregate_ac_ids],
    ]) {
      const uniqueValues = new Set(values);
      if (uniqueValues.size !== values.length) {
        throw new Error(`${label} ${field} must not contain duplicates`);
      }
    }
    for (const aggregateAC of claim.aggregate_ac_ids) {
      if (!claim.required_ac_ids.includes(aggregateAC)) {
        throw new Error(`${label} aggregate_ac_ids must be a subset of required_ac_ids`);
      }
    }
    if (!claim.claimed) {
      continue;
    }
    if (claim.direct_evidence_ids.length === 0) {
      throw new Error(`${label} claimed profiles must declare direct_evidence_ids`);
    }
    for (const evidenceID of claim.direct_evidence_ids) {
      const entry = entryByID.get(evidenceID);
      if (!entry || entry.coverage !== "authoritative") {
        throw new Error(`${label} direct_evidence_id ${evidenceID} must name an authoritative row`);
      }
      const status = entry.claim_status ?? "implemented";
      if (status !== "implemented") {
        throw new Error(
          `${label} direct_evidence_id ${evidenceID} must have claim_status=implemented`,
        );
      }
    }
  }
}
