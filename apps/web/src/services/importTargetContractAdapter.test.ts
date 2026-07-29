import { readFileSync } from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import {
  importTargetRegistryDigest,
  importTargetSemantics,
  requireClaimGatedAnalyticalImportTarget,
  selectableViewImportTargets,
} from "./importTargetContractAdapter";
import { networkFlowMappingMetadata } from "./networkFlowContractAdapter";

const servicesDirectory = path.dirname(fileURLToPath(import.meta.url));
const sourceDirectory = path.resolve(servicesDirectory, "..");

describe("generated import target frontend semantics", () => {
  it("partitions all 18 generated rows without a client fallback registry", () => {
    expect(importTargetRegistryDigest).toMatch(/^[a-f0-9]{64}$/u);
    expect(importTargetSemantics).toHaveLength(18);
    expect(
      new Set(importTargetSemantics.map((row) => row.target_id)).size,
    ).toBe(18);
    expect(importTargetSemantics.map((row) => row.registry_order)).toEqual(
      Array.from({ length: 18 }, (_, index) => index),
    );

    expect(selectableViewImportTargets).toHaveLength(14);
    expect(
      importTargetSemantics.filter(
        (row) => row.public_projection_disposition === "hidden_reserved",
      ),
    ).toHaveLength(3);
    expect(
      requireClaimGatedAnalyticalImportTarget(
        networkFlowMappingMetadata.target_kind,
        networkFlowMappingMetadata.profile_id,
      ),
    ).toMatchObject({
      target_kind: "network_flow_table",
      availability_kind: "claim_gated",
      activation_policy: "extension_claim_required",
      public_projection_disposition: "extension_claim_gated",
    });
    expect(() =>
      requireClaimGatedAnalyticalImportTarget(
        networkFlowMappingMetadata.target_kind,
        "unclaimed_future_profile",
      ),
    ).toThrow("invalid analytical import target binding");
  });

  it("keeps generated identity consumption in the shared adapter", () => {
    const assistant = readFileSync(
      path.join(
        sourceDirectory,
        "workbook/features/ImportAssistantFeature.tsx",
      ),
      "utf8",
    );
    const networkFlowController = readFileSync(
      path.join(
        sourceDirectory,
        "networkFlow/useNetworkFlowImportController.ts",
      ),
      "utf8",
    );

    expect(assistant).toContain("selectableViewImportTargets");
    expect(assistant).not.toContain("importableViewSchemaIds");
    expect(assistant).not.toMatch(/cartulary\.view\.[a-z_]+\.v[0-9]+/u);
    expect(networkFlowController).toContain("networkFlowImportTarget");
    expect(networkFlowController).not.toContain(
      'targetKind: "network_flow_table"',
    );
    expect(networkFlowController).not.toContain(
      'extensionProfileId: "network_flow_activity"',
    );
  });
});
