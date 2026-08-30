import { describe, expect, it } from "vitest";

import {
  evidenceViewSchemaId,
  getViewContract,
  indicatorsViewSchemaId,
  listViewContracts,
  requireViewContract,
  resolveHeaderSortFieldKey,
  timelineViewSchemaId,
} from "./index";
import { listProjectedViewContracts } from "./projection";

describe("view contracts", () => {
  it("initializes all generated view artifacts by stable identity", () => {
    expect(listViewContracts()).toHaveLength(17);
    for (const contract of listViewContracts()) {
      expect(getViewContract(contract.viewSchemaId)).toBe(contract);
      expect(requireViewContract(contract.viewSchemaId)).toBe(contract);
      expect(contract.inspectorConfig.viewSchemaId).toBe(contract.viewSchemaId);
    }
    expect(() => requireViewContract("cartulary.view.missing.v1")).toThrow(
      "Unknown view schema contract: cartulary.view.missing.v1",
    );
  });

  it("matches generated contract projections by shared identity", () => {
    expect(listViewContracts()).toBe(listProjectedViewContracts());
  });

  it("freezes contracts and Inspector metadata", () => {
    expect(Object.isFrozen(listViewContracts())).toBe(true);
    expect(Object.isFrozen(listViewContracts()[0]?.fieldMap)).toBe(true);
    expect(Object.isFrozen(listViewContracts()[0]?.inspectorConfig)).toBe(true);
  });

  it("derives the four exact Indicator feature signatures from the owner registry", () => {
    const specialized = [
      requireViewContract(timelineViewSchemaId),
      requireViewContract(indicatorsViewSchemaId),
    ]
      .flatMap((contract) => contract.inspectorConfig.featureGroups)
      .filter((group) => group.featureGroupKey.startsWith("indicator."));
    expect(
      specialized.map((group) => [
        group.featureGroupKey,
        group.panelId,
        group.routeBinding.kind,
        group.routeBinding.owner,
        group.routeBinding.actionKey,
      ]),
    ).toEqual([
      [
        "indicator.observations.manage",
        "relationships",
        "indicator_observations",
        "indicator_observations_route",
        "indicator.observations.manage",
      ],
      [
        "indicator.observations.pivot",
        "relationships",
        "indicator_observations",
        "indicator_observations_route",
        "indicator.observations.pivot",
      ],
      [
        "indicator.lifecycle.read",
        "history",
        "indicator_lifecycle",
        "indicator_lifecycle_route",
        "indicator.lifecycle.read",
      ],
      [
        "indicator.lifecycle.manage",
        "history",
        "indicator_lifecycle",
        "indicator_lifecycle_route",
        "indicator.lifecycle.manage",
      ],
    ]);
  });

  it("requires explicit boolean grid_editable inputs", () => {
    for (const contract of listViewContracts()) {
      expect(typeof contract.createCapable).toBe("boolean");
      expect(Array.isArray(contract.createInputs)).toBe(true);
      for (const input of contract.createInputs) {
        expect(input.inputKey).not.toBe("");
        expect(input.valueContractId).not.toBe("");
        expect(typeof input.required).toBe("boolean");
        expect(typeof input.nullable).toBe("boolean");
      }
      for (const field of contract.fields) {
        expect(typeof field.gridEditable).toBe("boolean");
        expect(typeof field.createWritable).toBe("boolean");
        if (field.gridEditable) {
          expect(field.createWritable).toBe(true);
        }
      }
    }
  });

  it("requires an exact inline_create policy", () => {
    const semanticCreateExceptions = new Set([
      evidenceViewSchemaId,
      indicatorsViewSchemaId,
      timelineViewSchemaId,
    ]);
    for (const contract of listViewContracts()) {
      expect(typeof contract.permitsZeroFieldCreate).toBe("boolean");
      expect(Array.isArray(contract.minimumCreateFieldSets)).toBe(true);
      if (!semanticCreateExceptions.has(contract.viewSchemaId)) {
        expect(contract.minimumCreateFieldSets.length).toBeGreaterThan(0);
      }
      for (const fieldSet of contract.minimumCreateFieldSets) {
        expect(fieldSet.length).toBeGreaterThan(0);
        for (const fieldKey of fieldSet) {
          expect(contract.fieldMap[fieldKey]?.createWritable).toBe(true);
        }
      }
    }
  });

  it("exposes sortable, filterable, groupable, and inline-create metadata", () => {
    expect(
      listViewContracts().some((contract) =>
        Object.values(contract.sortableFieldMap).includes(true),
      ),
    ).toBe(true);
    expect(
      listViewContracts().some((contract) =>
        Object.values(contract.filterableFieldMap).includes(true),
      ),
    ).toBe(true);
    expect(
      listViewContracts().some((contract) =>
        Object.values(contract.groupableFieldMap).includes(true),
      ),
    ).toBe(true);
  });

  it("exposes synthetic filter predicates as filter-only fields", () => {
    const contract = listViewContracts().find((candidate) =>
      candidate.filterFields.some(
        (fieldKey) =>
          !candidate.fields.some((field) => field.fieldKey === fieldKey),
      ),
    );
    expect(contract).toBeDefined();
    const syntheticKey = contract?.filterFields.find(
      (fieldKey) =>
        !contract.fields.some((field) => field.fieldKey === fieldKey),
    );
    expect(syntheticKey).toBeDefined();
    expect(contract?.fieldMap[syntheticKey ?? ""]?.readKind).toBe(
      "synthetic_filter",
    );
  });

  it("resolves header sort keys from contract metadata", () => {
    const candidate = listViewContracts()
      .flatMap((contract) =>
        contract.fields.map((field) => ({ contract, field })),
      )
      .find(({ field }) => field.headerSortFieldKey !== null);
    expect(candidate).toBeDefined();
    if (candidate) {
      expect(
        resolveHeaderSortFieldKey(candidate.contract, candidate.field.fieldKey),
      ).toBe(candidate.field.headerSortFieldKey);
    }
  });

  it("exposes mutation metadata needed by workbook controls", () => {
    const timeline = requireViewContract(timelineViewSchemaId);
    expect(
      timeline.inspectorConfig.featureGroups.some((group) => group.mutates),
    ).toBe(true);
  });

  it("exposes inspector config by stable view_schema_id and semantic keys", () => {
    for (const contract of listViewContracts()) {
      expect(contract.inspectorConfig.viewSchemaId).toBe(contract.viewSchemaId);
      expect(contract.inspectorConfig.defaultOpen).toBe(false);
      expect(contract.inspectorConfig.unsupportedFeatureBehavior).toBe(
        "omit_feature",
      );
    }
  });

  it("uses generated Inspector panels and feature groups validated before browser initialization", () => {
    for (const contract of listViewContracts()) {
      const panelIds = contract.inspectorConfig.panels.map(
        (panel) => panel.panelId,
      );
      const featureKeys = contract.inspectorConfig.featureGroups.map(
        (group) => group.featureGroupKey,
      );
      expect(new Set(panelIds).size).toBe(panelIds.length);
      expect(new Set(featureKeys).size).toBe(featureKeys.length);
      for (const group of contract.inspectorConfig.featureGroups) {
        expect(panelIds).toContain(group.panelId);
      }
    }
  });
});
