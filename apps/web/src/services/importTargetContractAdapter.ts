import {
  type ImportTargetFrontendRow,
  importTargetRegistry,
} from "@cartulary/protocol-ts/import-targets";

export type ImportUnknownColumnPolicy =
  | "preserve_raw_capture"
  | "preserve_custom_attrs"
  | "reject_if_unmapped";

export type SelectableViewImportTarget = ImportTargetFrontendRow & {
  readonly target_kind: "view_schema";
  readonly target_view_schema_id: string;
  readonly extension_profile_id: null;
  readonly availability_kind: "enabled";
  readonly activation_policy: "always";
  readonly default_unknown_column_policy: ImportUnknownColumnPolicy;
  readonly public_projection_disposition: "selectable";
};

export type ClaimGatedAnalyticalImportTarget = ImportTargetFrontendRow & {
  readonly target_view_schema_id: null;
  readonly extension_profile_id: string;
  readonly availability_kind: "claim_gated";
  readonly activation_policy: "extension_claim_required";
  readonly public_projection_disposition: "extension_claim_gated";
};

export const importTargetRegistryDigest = importTargetRegistry.registry_sha256;
export const importTargetSemantics = importTargetRegistry.targets;

export const selectableViewImportTargets: readonly SelectableViewImportTarget[] =
  importTargetRegistry.targets.flatMap((row) => {
    if (row.public_projection_disposition !== "selectable") {
      return [];
    }
    if (
      row.target_kind !== "view_schema" ||
      row.target_view_schema_id === null ||
      row.extension_profile_id !== null ||
      row.availability_kind !== "enabled" ||
      row.activation_policy !== "always" ||
      !isUnknownColumnPolicy(row.default_unknown_column_policy)
    ) {
      throw new Error(`invalid selectable import target ${row.target_id}`);
    }
    return [row as SelectableViewImportTarget];
  });

export function requireClaimGatedAnalyticalImportTarget(
  targetKind: string,
  extensionProfileId: string,
): ClaimGatedAnalyticalImportTarget {
  const matches = importTargetRegistry.targets.filter(
    (row) =>
      row.target_kind === targetKind &&
      row.extension_profile_id === extensionProfileId,
  );
  const row = matches[0];
  if (
    matches.length !== 1 ||
    row === undefined ||
    row.target_view_schema_id !== null ||
    row.extension_profile_id === null ||
    row.availability_kind !== "claim_gated" ||
    row.activation_policy !== "extension_claim_required" ||
    row.public_projection_disposition !== "extension_claim_gated"
  ) {
    throw new Error(
      `invalid analytical import target binding ${targetKind}:${extensionProfileId}`,
    );
  }
  return row as ClaimGatedAnalyticalImportTarget;
}

function isUnknownColumnPolicy(
  value: string,
): value is ImportUnknownColumnPolicy {
  return (
    value === "preserve_raw_capture" ||
    value === "preserve_custom_attrs" ||
    value === "reject_if_unmapped"
  );
}
