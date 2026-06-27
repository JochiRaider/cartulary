import { describe, expect, it } from "vitest";

import {
  type AccountPreferencesEnvelope,
  type AccountPreferencesPutRequest,
  type AccountPreferencesResource,
  type AccountProfileEnvelope,
  type AccountProfilePatchRequest,
  type AccountProfileResource,
  accountProtocolSchemaNames,
  type ContractArtifact,
  type DensityMode,
  type ErrorEnvelope,
  type EvidenceAttachBlobEnvelope,
  type EvidenceAttachBlobRequest,
  type EvidenceHandleEnvelope,
  type EvidenceHandleIssueRequest,
  type ExtensionRegistryContract,
  evidenceProtocolSchemaNames,
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
  type ObjectBlobCreateEnvelope,
  type ObjectBlobCreateRequest,
  type ObjectBlobUploadTarget,
  parseContractArtifact,
  requireContractArtifact,
  requireExtensionProfile,
  requireReasonCodeRegistry,
  requireViewSchemaRegistryEntry,
  type ViewSchemaRegistryContract,
} from "./index";

const requiredBaseViewSchemaIds = [
  "cartulary.view.timeline.v2",
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
    expect(
      families.viewSchemaArtifacts.map((artifact) => artifact.path),
    ).toContain("contracts/view-schemas/index.json");
    expect(families.errorArtifacts.map((artifact) => artifact.path)).toEqual([
      "contracts/errors/index.json",
    ]);
    expect(
      families.extensionArtifacts.map((artifact) => artifact.path),
    ).toEqual(["contracts/extensions/index.json"]);

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
    ).toThrow("missing contract artifact contracts/view-schemas/missing.json");
  });

  it("exposes stable current-profile view schema identifiers", () => {
    const registry = getViewSchemaRegistryContract();
    const ids = listViewSchemaRegistryEntries().map(
      (entry) => entry.view_schema_id,
    );

    expect(registry.registry_id).toBe("cartulary.view_schemas.base.v1");
    expect(ids).toEqual(expect.arrayContaining([...requiredBaseViewSchemaIds]));
    for (const viewSchemaId of requiredBaseViewSchemaIds) {
      expect(requireViewSchemaRegistryEntry(viewSchemaId).artifact_path).toBe(
        `contracts/view-schemas/${viewSchemaId}.json`,
      );
    }
    expect(
      getViewSchemaRegistryEntry("cartulary.view.missing.v1"),
    ).toBeUndefined();
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
    expect(
      requireReasonCodeRegistry("invalid_mutation_payload").reason_codes,
    ).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: "unknown_view_schema" }),
        expect.objectContaining({ code: "unknown_field_key" }),
      ]),
    );
    expect(
      requireReasonCodeRegistry("invalid_startup_request").reason_codes,
    ).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: "ambiguous_explicit_sheet_ref" }),
        expect.objectContaining({ code: "unsupported_sheet_ref_kind" }),
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
    expect(
      listExtensionProfiles().map((profile) => profile.profile_id),
    ).toEqual([
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

  it("anchors evidence protocol facade types to generated OpenAPI schema names", () => {
    const openAPI = parseContractArtifact<{
      components: { schemas: Record<string, unknown> };
    }>("contracts/openapi/cartulary.openapi.yaml");

    expect(Object.values(evidenceProtocolSchemaNames)).toEqual([
      "EnvelopeMeta",
      "ErrorEnvelope",
      "EvidenceAttachBlobEnvelope",
      "EvidenceAttachBlobRequest",
      "EvidenceHandleEnvelope",
      "EvidenceHandleIssueRequest",
      "ObjectBlobCreateEnvelope",
      "ObjectBlobCreateRequest",
      "ObjectBlobUploadTarget",
    ]);
    for (const schemaName of Object.values(evidenceProtocolSchemaNames)) {
      expect(openAPI.components.schemas[schemaName]).toBeDefined();
    }

    const createRequest = {
      incident_id: "incident-1",
      client_txn_id: "txn-create-blob",
      byte_size: 42,
      filename_hint: "evidence.txt",
      content_type_hint: "text/plain",
      sha256_hex: null,
    } satisfies ObjectBlobCreateRequest;
    const uploadTarget = {
      href: "/api/v1/object-uploads/upload-token",
      method: "PUT",
      expires_at: "2026-06-07T12:00:00Z",
      headers: { "X-Upload-Contract": "generated-protocol" },
    } satisfies ObjectBlobUploadTarget;
    const createEnvelope = {
      data: {
        accepted_contract: {
          byte_size: 42,
          content_type_hint: "text/plain",
          filename_hint: "evidence.txt",
          incident_id: "incident-1",
          sha256_hex: null,
        },
        incident_id: "incident-1",
        object_blob_id: "object-blob-1",
        pending_expires_at: "2026-06-07T12:00:00Z",
        target_expires_at: "2026-06-07T12:00:00Z",
        upload_state: "pending",
        upload_target: uploadTarget,
      },
      meta: { request_id: "req-create-blob" },
    } satisfies ObjectBlobCreateEnvelope;
    const attachRequest = {
      object_blob_id: createEnvelope.data.object_blob_id,
      base_row_version: 1,
      client_txn_id: "txn-attach-blob",
    } satisfies EvidenceAttachBlobRequest;
    const attachEnvelope = {
      data: {
        change_set_id: "change-1",
        object_blob_id: attachRequest.object_blob_id,
        row: {
          record_id: "evidence-1",
          row_version: 2,
          cells: {},
        },
        view_schema_id: "cartulary.view.evidence.v1",
      },
      meta: { request_id: "req-attach-blob" },
    } satisfies EvidenceAttachBlobEnvelope;
    const handleRequest = {} satisfies EvidenceHandleIssueRequest;
    const handleEnvelope = {
      data: {
        content_type: "text/plain",
        disposition: "inline",
        evidence_lifecycle_state: "available",
        expires_at: "2026-06-07T12:00:00Z",
        filename: "evidence.txt",
        handle_kind: "preview",
        href: "/api/v1/evidence-handles/preview-token",
        incident_id: "incident-1",
        media_class: "text",
        method: "GET",
        object_blob_id: "object-blob-1",
        preview_kind: "text_inline",
        record_id: "evidence-1",
        sha256: null,
        single_use: true,
        size_bytes: 42,
        upload_state: "available",
      },
      meta: { request_id: "req-handle" },
    } satisfies EvidenceHandleEnvelope;
    const errorEnvelope = {
      error: {
        status: 409,
        code: "evidence_attach_rejected",
        message: "Evidence attach was rejected.",
        request_id: "req-error",
        retryable: false,
        details: { reason_code: "blob_failed" },
      },
    } satisfies ErrorEnvelope;

    expect(createRequest.byte_size).toBe(42);
    expect(attachEnvelope.data.row.row_version).toBe(2);
    expect(handleRequest).toEqual({});
    expect(handleEnvelope.data.href).toBe(
      "/api/v1/evidence-handles/preview-token",
    );
    expect(errorEnvelope.error.details.reason_code).toBe("blob_failed");
  });

  it("anchors account preference protocol facade types to generated OpenAPI schema names and routes", () => {
    const openAPI = parseContractArtifact<{
      components: { schemas: Record<string, unknown> };
      paths: Record<
        string,
        Record<string, { operationId?: string; requestBody?: unknown }>
      >;
    }>("contracts/openapi/cartulary.openapi.yaml");

    expect(Object.values(accountProtocolSchemaNames)).toEqual([
      "AccountPreferencesEnvelope",
      "AccountPreferencesPutRequest",
      "AccountPreferencesResource",
      "AccountProfileEnvelope",
      "AccountProfilePatchRequest",
      "AccountProfileResource",
      "DensityMode",
    ]);
    for (const schemaName of Object.values(accountProtocolSchemaNames)) {
      expect(openAPI.components.schemas[schemaName]).toBeDefined();
    }
    expect(openAPI.paths["/api/v1/account/profile"]?.get?.operationId).toBe(
      "getCurrentAccountProfile",
    );
    expect(openAPI.paths["/api/v1/account/profile"]?.patch?.operationId).toBe(
      "patchCurrentAccountProfile",
    );
    expect(openAPI.paths["/api/v1/account/preferences"]?.get?.operationId).toBe(
      "getCurrentAccountPreferences",
    );
    expect(openAPI.paths["/api/v1/account/preferences"]?.put?.operationId).toBe(
      "putCurrentAccountPreferences",
    );

    const densityMode = "compact" satisfies DensityMode;
    const profileResource = {
      user_id: "af0f2c88-1fd2-42ad-9631-4fbbef243f30",
      email: "operator@example.test",
      display_name: "Operator",
      user_version: 1,
      created_at: "2026-06-07T12:00:00Z",
      updated_at: "2026-06-07T12:00:00Z",
    } satisfies AccountProfileResource;
    const preferencesResource = {
      user_id: profileResource.user_id,
      density_mode: densityMode,
      preferences_version: 2,
      created_at: "2026-06-07T12:00:00Z",
      updated_at: "2026-06-07T12:05:00Z",
    } satisfies AccountPreferencesResource;
    const profilePatchRequest = {
      base_user_version: profileResource.user_version,
      client_txn_id: "txn-profile",
      display_name: "Operator Prime",
    } satisfies AccountProfilePatchRequest;
    const preferencesPutRequest = {
      base_preferences_version: preferencesResource.preferences_version,
      client_txn_id: "txn-preferences",
      density_mode: null,
    } satisfies AccountPreferencesPutRequest;
    const profileEnvelope = {
      data: profileResource,
      meta: { request_id: "req-profile" },
    } satisfies AccountProfileEnvelope;
    const preferencesEnvelope = {
      data: preferencesResource,
      meta: { request_id: "req-preferences" },
    } satisfies AccountPreferencesEnvelope;

    expect(profilePatchRequest.display_name).toBe("Operator Prime");
    expect(preferencesPutRequest.density_mode).toBeNull();
    expect(profileEnvelope.data.user_id).toBe(profileResource.user_id);
    expect(preferencesEnvelope.data.density_mode).toBe("compact");
  });
});
