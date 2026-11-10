import { describe, expect, it } from "vitest";
import { workbookMappingRequest } from "../imports/workbookImportMapping";
import {
  importTestPreview,
  importTestUnit,
} from "../testing/workbookImportTestSupport";
import {
  importTargetRegistryDigest,
  importTargetSemantics,
  requireClaimGatedAnalyticalImportTarget,
  selectableViewImportTargets,
  workbookImportTargets,
} from "./importTargetContractAdapter";
import { networkFlowMappingMetadata } from "./networkFlowContractAdapter";

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

  it("maps selectable targets and policies through the workbook consumer", () => {
    const unit = importTestUnit();
    for (const target of workbookImportTargets) {
      const columns = target.fields.map((field, index) => ({
        source_column_ordinal: index + 1,
        source_header_text: field.label,
      }));
      const preview = { ...importTestPreview(unit), columns };
      const fields = Object.fromEntries(
        target.fields.map((field, index) => [index + 1, field.fieldKey]),
      );
      const request = workbookMappingRequest(
        unit,
        preview,
        {
          targetViewSchemaId: target.contract.viewSchemaId,
          fields,
          dirty: true,
        },
        "consumer-mapping",
      );
      expect(request).toMatchObject({
        target_view_schema_id: target.semantics.target_view_schema_id,
        unknown_column_policy: target.semantics.default_unknown_column_policy,
      });
      expect(request?.source_columns.map((column) => column.field_key)).toEqual(
        target.fields.map((field) => field.fieldKey),
      );
    }
    for (const target of importTargetSemantics.filter(
      (row) => row.public_projection_disposition !== "selectable",
    )) {
      expect(
        workbookMappingRequest(
          unit,
          importTestPreview(unit),
          {
            targetViewSchemaId:
              target.target_view_schema_id ?? target.target_kind,
            fields: {},
            dirty: true,
          },
          "blocked",
        ),
      ).toBeNull();
    }
    expect(
      workbookMappingRequest(
        unit,
        importTestPreview(unit),
        {
          targetViewSchemaId: "unknown-future-target",
          fields: {},
          dirty: true,
        },
        "unknown",
      ),
    ).toBeNull();
  });
});
