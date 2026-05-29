import { describe, expect, it } from "vitest";

import {
  getContractArtifact,
  getErrorRegistryContract,
  getExtensionProfile,
  getExtensionRegistryContract,
  getReasonCodeRegistry,
  getViewSchemaRegistryContract,
  getViewSchemaRegistryEntry,
  listContractArtifactFamilies,
  listExtensionProfiles,
  listViewSchemaRegistryEntries,
  parseContractArtifact,
  requireContractArtifact,
  requireExtensionProfile,
  requireReasonCodeRegistry,
  requireViewSchemaRegistryEntry,
  type ContractArtifact,
  type ExtensionRegistryContract,
  type ViewSchemaRegistryContract,
} from "./index";

const requiredBaseViewSchemaIds = [
  "cartulary.view.timeline.v1",
  "cartulary.view.hosts.v1",
  "cartulary.view.identities.v1",
  "cartulary.view.evidence.v1",
  "cartulary.view.notes.v1",
  "cartulary.view.indicators.v1",
  "cartulary.view.assessments.v1",
  "cartulary.view.task_requests.v1",
  "cartulary.view.decisions.v1",
  "cartulary.view.parties.v1",
  "cartulary.view.comm_log.v1",
  "cartulary.view.handoff.v1",
  "cartulary.view.status_review.v1",
  "cartulary.view.lesson.v1",
] as const;

describe("@cartulary/protocol-ts facade", () => {
  it("exposes generated artifact families through stable facade helpers", () => {
    const families = listContractArtifactFamilies();

    expect(families.openAPIArtifacts.map((artifact) => artifact.path)).toEqual([
      "contracts/openapi/cartulary.openapi.yaml",
    ]);
    expect(families.wsArtifacts.map((artifact) => artifact.path)).toEqual([
      "contracts/ws/index.schema.json",
    ]);
    expect(families.viewSchemaArtifacts.map((artifact) => artifact.path)).toContain(
      "contracts/view-schemas/index.json",
    );
    expect(families.errorArtifacts.map((artifact) => artifact.path)).toEqual([
      "contracts/errors/index.json",
    ]);
    expect(families.extensionArtifacts.map((artifact) => artifact.path)).toEqual([
      "contracts/extensions/index.json",
    ]);

    const artifact: ContractArtifact = requireContractArtifact(
      "contracts/view-schemas/index.json",
    );
    expect(artifact.sha256).toMatch(/^[a-f0-9]{64}$/);
    expect(artifact.json).toContain("cartulary.view_schemas.base.v1");
  });

  it("requires and parses known contract artifacts through the facade", () => {
    expect(getContractArtifact("contracts/view-schemas/index.json")?.path).toBe(
      "contracts/view-schemas/index.json",
    );
    expect(
      parseContractArtifact<ViewSchemaRegistryContract>(
        "contracts/view-schemas/index.json",
      ).registry_id,
    ).toBe("cartulary.view_schemas.base.v1");

    expect(() =>
      requireContractArtifact("contracts/view-schemas/missing.json"),
    ).toThrow(
      "missing contract artifact contracts/view-schemas/missing.json",
    );
  });

  it("exposes stable current-profile view schema identifiers", () => {
    const registry = getViewSchemaRegistryContract();
    const ids = listViewSchemaRegistryEntries().map(
      (entry) => entry.view_schema_id,
    );

    expect(registry.registry_id).toBe("cartulary.view_schemas.base.v1");
    expect(ids).toEqual(
      expect.arrayContaining([...requiredBaseViewSchemaIds]),
    );
    for (const viewSchemaId of requiredBaseViewSchemaIds) {
      expect(requireViewSchemaRegistryEntry(viewSchemaId).artifact_path).toBe(
        `contracts/view-schemas/${viewSchemaId}.json`,
      );
    }
    expect(getViewSchemaRegistryEntry("cartulary.view.missing.v1")).toBeUndefined();
    expect(() =>
      requireViewSchemaRegistryEntry("cartulary.view.missing.v1"),
    ).toThrow(
      "missing view-schema registry entry for cartulary.view.missing.v1",
    );
  });

  it("exposes error and reason-code registries through facade helpers", () => {
    expect(getErrorRegistryContract().registry_id).toBe(
      "cartulary.errors.phase3.v1",
    );
    expect(requireReasonCodeRegistry("invalid_mutation_payload").reason_codes).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: "unknown_view_schema" }),
        expect.objectContaining({ code: "unknown_field_key" }),
      ]),
    );
    expect(getReasonCodeRegistry("missing_error_code")).toBeUndefined();
    expect(() => requireReasonCodeRegistry("missing_error_code")).toThrow(
      "missing reason-code registry for missing_error_code",
    );
  });

  it("exposes extension registry profiles through facade helpers", () => {
    const registry = parseContractArtifact<ExtensionRegistryContract>(
      "contracts/extensions/index.json",
    );

    expect(getExtensionRegistryContract()).toEqual(registry);
    expect(listExtensionProfiles().map((profile) => profile.profile_id)).toEqual([
      "enterprise_authentication",
      "import",
      "incident_portability",
      "reference_pack",
      "snapshot_reporting",
    ]);
    expect(requireExtensionProfile("reference_pack").route_families).toContain(
      "/api/v1/reference-packs",
    );
    expect(getExtensionProfile("missing_profile")).toBeUndefined();
    expect(() => requireExtensionProfile("missing_profile")).toThrow(
      "missing extension profile for missing_profile",
    );
  });
});
