import { describe, expect, expectTypeOf, it } from "vitest";

import {
  type AccountPreferencesEnvelope,
  type AccountPreferencesPutRequest,
  type AccountPreferencesResource,
  type AccountProfileEnvelope,
  type AccountProfilePatchRequest,
  type AccountProfileResource,
  type ApplyImportSessionRequest,
  buildHTTPOperationPath,
  type CancelJobRequest,
  type ContractArtifact,
  type DensityMode,
  type ErrorEnvelope,
  type EvidenceAttachBlobEnvelope,
  type EvidenceAttachBlobRequest,
  type EvidenceHandleEnvelope,
  type EvidenceHandleIssueRequest,
  type ExtensionRegistryContract,
  encodeHTTPOperationQuery,
  extensionDiscoveryDecoder,
  getContractArtifact,
  getErrorRegistryContract,
  getExtensionProfile,
  getExtensionRegistryContract,
  getReasonCodeRegistry,
  getViewSchemaRegistryContract,
  getViewSchemaRegistryEntry,
  type HTTPOperationRequest,
  type HTTPOperationResponse,
  httpOperationBindings,
  type ListImportUnitsResponse,
  listContractArtifactFamilies,
  listExtensionProfiles,
  listViewSchemaRegistryEntries,
  networkFlowContractDescriptor,
  networkFlowDecoders,
  type ObjectBlobCreateEnvelope,
  type ObjectBlobCreateRequest,
  type ObjectBlobUploadTarget,
  parseContractArtifact,
  type QueryWorkbookViewRequest,
  type QueryWorkbookViewResponse,
  requireContractArtifact,
  requireExtensionProfile,
  requireReasonCodeRegistry,
  requireViewSchemaRegistryEntry,
  type SelectImportUnitResponse,
  type ViewSchemaRegistryContract,
  validateHTTPOperationResponse,
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
  it("exposes deterministic generated HTTP operation bindings without payload leakage", () => {
    expect(
      buildHTTPOperationPath("getDeploymentUser", {
        user_id: "user/id",
      }),
    ).toBe("/api/v1/users/user%2Fid");
    expect(
      encodeHTTPOperationQuery("listDeploymentUsers", {
        search: "Alice Example",
        limit: 25,
      }),
    ).toBe("?limit=25&search=Alice%20Example");
    expect(() =>
      encodeHTTPOperationQuery("getDeploymentUser", {
        unsupported: "must-not-pass",
      }),
    ).toThrow("unexpected query parameter unsupported");

    expect([
      buildHTTPOperationPath("createImportSession"),
      buildHTTPOperationPath("getImportSession", {
        import_session_id: "session/id",
      }),
      buildHTTPOperationPath("listImportUnits", {
        import_session_id: "session/id",
      }),
      buildHTTPOperationPath("getImportUnit", {
        import_session_id: "session/id",
        import_unit_id: "unit/id",
      }),
      buildHTTPOperationPath("getImportUnitPreview", {
        import_session_id: "session/id",
        import_unit_id: "unit/id",
      }),
      buildHTTPOperationPath("createImportUnitRegion", {
        import_session_id: "session/id",
        base_unit_id: "unit/id",
      }),
      buildHTTPOperationPath("previewImportUnitExtensionMapping", {
        import_session_id: "session/id",
        import_unit_id: "unit/id",
      }),
      buildHTTPOperationPath("putImportUnitMapping", {
        import_session_id: "session/id",
        import_unit_id: "unit/id",
      }),
      buildHTTPOperationPath("selectImportUnit", {
        import_session_id: "session/id",
        import_unit_id: "unit/id",
      }),
      buildHTTPOperationPath("skipImportUnit", {
        import_session_id: "session/id",
        import_unit_id: "unit/id",
      }),
      buildHTTPOperationPath("applyImportSession", {
        import_session_id: "session/id",
      }),
      buildHTTPOperationPath("getJob", { job_id: "job/id" }),
      buildHTTPOperationPath("cancelJob", { job_id: "job/id" }),
    ]).toEqual([
      "/api/v1/import-sessions",
      "/api/v1/import-sessions/session%2Fid",
      "/api/v1/import-sessions/session%2Fid/units",
      "/api/v1/import-sessions/session%2Fid/units/unit%2Fid",
      "/api/v1/import-sessions/session%2Fid/units/unit%2Fid/preview",
      "/api/v1/import-sessions/session%2Fid/units/unit%2Fid/regions",
      "/api/v1/import-sessions/session%2Fid/units/unit%2Fid/mapping-preview",
      "/api/v1/import-sessions/session%2Fid/units/unit%2Fid/mapping",
      "/api/v1/import-sessions/session%2Fid/units/unit%2Fid/select",
      "/api/v1/import-sessions/session%2Fid/units/unit%2Fid/skip",
      "/api/v1/import-sessions/session%2Fid/apply",
      "/api/v1/jobs/job%2Fid",
      "/api/v1/jobs/job%2Fid/cancel",
    ]);
    expect(
      encodeHTTPOperationQuery("listImportUnits", {
        cursor_token: "opaque /+ cursor",
        limit: 50,
      }),
    ).toBe("?cursor_token=opaque%20%2F%2B%20cursor&limit=50");
    expect(
      buildHTTPOperationPath("queryWorkbookView", {
        incident_id: "incident/id",
        view_schema_id: "cartulary.view.timeline.v2",
      }),
    ).toBe(
      "/api/v1/incidents/incident%2Fid/views/cartulary.view.timeline.v2/query",
    );
    expect(
      buildHTTPOperationPath("patchRecord", { record_id: "record/id" }),
    ).toBe("/api/v1/records/record%2Fid");
    expect([
      buildHTTPOperationPath("getIncident", { incident_id: "incident/id" }),
      buildHTTPOperationPath("listIncidentMemberships", {
        incident_id: "incident/id",
      }),
      buildHTTPOperationPath("listIncidentSavedViews", {
        incident_id: "incident/id",
      }),
      buildHTTPOperationPath("createIncidentSavedView", {
        incident_id: "incident/id",
      }),
      buildHTTPOperationPath("patchIncidentSavedView", {
        incident_id: "incident/id",
        saved_view_id: "saved/view",
      }),
      buildHTTPOperationPath("deleteIncidentSavedView", {
        incident_id: "incident/id",
        saved_view_id: "saved/view",
      }),
    ]).toEqual([
      "/api/v1/incidents/incident%2Fid",
      "/api/v1/incidents/incident%2Fid/memberships",
      "/api/v1/incidents/incident%2Fid/saved-views",
      "/api/v1/incidents/incident%2Fid/saved-views",
      "/api/v1/incidents/incident%2Fid/saved-views/saved%2Fview",
      "/api/v1/incidents/incident%2Fid/saved-views/saved%2Fview",
    ]);
    expect(
      encodeHTTPOperationQuery("listIncidentSavedViews", {
        cursor_token: "opaque /+ cursor",
        limit: 50,
      }),
    ).toBe("?cursor_token=opaque%20%2F%2B%20cursor&limit=50");
    expect(
      [
        "getIncident",
        "listIncidentMemberships",
        "listIncidentSavedViews",
        "createIncidentSavedView",
        "patchIncidentSavedView",
        "deleteIncidentSavedView",
      ].map(
        (operationID) =>
          httpOperationBindings[
            operationID as keyof typeof httpOperationBindings
          ].method,
      ),
    ).toEqual(["GET", "GET", "GET", "POST", "PATCH", "DELETE"]);

    const applyRequest: ApplyImportSessionRequest = {
      client_txn_id: "txn-apply",
      selected_unit_ids: ["unit-1"],
    };
    const cancelRequest: CancelJobRequest = {
      client_txn_id: "txn-cancel",
    };
    expect(applyRequest.selected_unit_ids).toEqual(["unit-1"]);
    expect(cancelRequest.client_txn_id).toBe("txn-cancel");
    expectTypeOf<ListImportUnitsResponse>().toMatchTypeOf<{
      data: { import_units: unknown[] };
      meta: { request_id: string };
    }>();
    expectTypeOf<SelectImportUnitResponse>().toMatchTypeOf<{
      data: {
        import_session_id: string;
        selected_unit_ids: string[];
        session_status: string;
      };
      meta: { request_id: string };
    }>();
    expectTypeOf<
      HTTPOperationRequest<"queryWorkbookView">
    >().toEqualTypeOf<QueryWorkbookViewRequest>();
    expectTypeOf<
      HTTPOperationResponse<"queryWorkbookView">
    >().toEqualTypeOf<QueryWorkbookViewResponse>();

    const strictWorkbookOperations = [
      "applyWorkbookBulkMutation",
      "createIncidentSavedView",
      "createRecordLinkedNote",
      "createViewRow",
      "deleteIncidentSavedView",
      "deleteRecord",
      "getCurrentUserWorkbookPreferences",
      "getIncident",
      "getIncidentDefaultWorkbookPreferences",
      "getIncidentWorkbookStartup",
      "getRecordHistory",
      "getTimelineTimeConversionProfile",
      "listIncidentMemberships",
      "listIncidentSavedViews",
      "markTimelineRecordReviewed",
      "mergeEntityRecord",
      "pasteWorkbookClipboard",
      "patchIncidentSavedView",
      "patchRecord",
      "putCurrentUserWorkbookPreferences",
      "putIncidentDefaultWorkbookPreferences",
      "putTimelineTimeConversionProfile",
      "queryWorkbookView",
      "resolveEntityMention",
      "resolveRecordSameFieldConflict",
      "restoreRecord",
      "rollbackRecord",
      "supersedeRecord",
    ] as const;
    for (const operationID of strictWorkbookOperations) {
      expect(validateHTTPOperationResponse(operationID, {})).toEqual(
        expect.objectContaining({ ok: false }),
      );
    }

    for (const viewSchemaId of [
      "cartulary.view.findings.v1",
      "cartulary.view.investigative_queries.v1",
      "cartulary.view.forensic_keywords.v1",
    ]) {
      expect(
        validateHTTPOperationResponse("patchRecord", {
          data: {
            change_set_id: "00000000-0000-4000-8000-000000000002",
            row: {
              cells: {
                "extension.value": { value: "strict optional surface" },
              },
              group_values: { "extension.group": "strict" },
              record_id: "00000000-0000-4000-8000-000000000001",
              row_version: 2,
            },
            view_schema_id: viewSchemaId,
          },
          meta: { request_id: "request-optional-surface" },
        }),
      ).toEqual(expect.objectContaining({ ok: true }));
    }
    expect(
      validateHTTPOperationResponse("patchRecord", {
        data: {
          change_set_id: "00000000-0000-4000-8000-000000000002",
          row: {
            cells: {},
            record_id: "00000000-0000-4000-8000-000000000001",
            row_version: 2,
          },
          view_schema_id: "cartulary.view.unsupported.v1",
        },
        meta: { request_id: "request-unsupported-surface" },
      }),
    ).toEqual(expect.objectContaining({ ok: false }));

    const malformedImportUnits = validateHTTPOperationResponse(
      "listImportUnits",
      {
        data: { import_units: [] },
        meta: {
          request_id: "request-test",
          paging: { has_more: true, limit: 50 },
        },
      },
    );
    expect(malformedImportUnits).toEqual(
      expect.objectContaining({
        ok: false,
        schemaId: "cartulary.core_http.ImportUnitsEnvelope.v1",
      }),
    );
    expect(
      validateHTTPOperationResponse("putImportUnitMapping", {
        data: {
          import_session_id: "00000000-0000-4000-8000-000000000001",
          import_unit_id: "00000000-0000-4000-8000-000000000002",
          locator_kind: "csv_file",
          locator: { file: "source" },
          source_rect_a1: "A1:A2",
          header_row_ref: 1,
          data_start_row_ref: 2,
          inferred_row_count: 1,
          inferred_column_count: 1,
          warning_codes: [],
          unit_status: "mapped",
          mapping_fingerprint: "mapping-fingerprint",
          approved_mapping: {
            target_kind: "network_flow_table",
            extension_profile_id: "network_flow_activity",
            owner_mapping_schema_id:
              "cartulary.network_flow_mapping_candidate.v1",
            owner_mapping: { target_kind: "network_flow_table" },
            source_columns: [
              {
                source_column_ordinal: 1,
                source_header_text: "Source IP",
                field_key: null,
                entity_binding_mode: null,
                transform_id: null,
                transform_options: {},
                empty_value_policy: "omit_field",
              },
            ],
          },
        },
        meta: { request_id: "request-test" },
      }),
    ).toEqual(expect.objectContaining({ ok: true }));

    const invalid = validateHTTPOperationResponse(
      "listAdministrativeAuditEvents",
      {
        data: {
          audit_events: [
            {
              unexpected_secret: "must-not-leak",
            },
          ],
        },
        meta: { request_id: "request-test" },
      },
    );
    expect(invalid).toEqual(
      expect.objectContaining({
        ok: false,
        schemaId: "cartulary.core_http.AdministrativeAuditEnvelope.v1",
      }),
    );
    expect(JSON.stringify(invalid)).not.toContain("must-not-leak");
  });

  it("exposes generated artifact families through stable facade helpers", () => {
    const families = listContractArtifactFamilies();

    expect("openAPIArtifacts" in families).toBe(false);
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
    ).toContain("contracts/extensions/generated/profile-registry.json");

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
      "contracts/extensions/generated/profile-registry.json",
    );

    expect(getExtensionRegistryContract()).toEqual(registry);
    expect(
      listExtensionProfiles().map((profile) => profile.profile_id),
    ).toEqual([
      "enterprise_authentication",
      "import",
      "incident_portability",
      "network_flow_activity",
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

  it("decodes exact Network Flow contracts without exposing payload data on failure", () => {
    expect(networkFlowContractDescriptor).toEqual({
      profile_id: "network_flow_activity",
      contract_major: 2,
    });

    const valid = networkFlowDecoders.tableList.decode({
      schema_id: "cartulary.network_flow.table_list.v1",
      tables: [],
      meta: { count: 0 },
    });
    expect(valid).toEqual({
      ok: true,
      value: {
        schema_id: "cartulary.network_flow.table_list.v1",
        tables: [],
        meta: { count: 0 },
      },
    });

    const invalid = networkFlowDecoders.tableList.decode({
      schema_id: "cartulary.network_flow.table_list.v1",
      tables: [],
      meta: { count: 0 },
      raw_source_value: "must-not-leak",
    });
    expect(invalid).toEqual({
      ok: false,
      error: {
        boundary: "generated_protocol",
        instancePath: "",
        reasonCategory: "unknown_member",
        schemaId: "cartulary.network_flow.table_list.v1",
      },
    });
    expect(JSON.stringify(invalid)).not.toContain("must-not-leak");
  });

  it("decodes the Core extension discovery envelope at the transport boundary", () => {
    expect(
      extensionDiscoveryDecoder.decode({
        data: { extensions: [] },
        meta: { request_id: "request-test" },
      }),
    ).toEqual({
      ok: true,
      value: {
        data: { extensions: [] },
        meta: { request_id: "request-test" },
      },
    });

    expect(
      extensionDiscoveryDecoder.decode({
        data: {
          extensions: [
            {
              profile_id: "network_flow_activity",
              claimable: true,
              claimed: true,
              contract_major: 2,
              route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
              workspace_keys: ["network_analysis"],
              capabilities: [],
              future_additive_member: { executable: "must-remain-inert" },
            },
          ],
        },
        meta: { request_id: "request-test" },
      }),
    ).toEqual({
      ok: true,
      value: {
        data: {
          extensions: [
            {
              profile_id: "network_flow_activity",
              claimable: true,
              claimed: true,
              contract_major: 2,
              route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
              workspace_keys: ["network_analysis"],
              capabilities: [],
            },
          ],
        },
        meta: { request_id: "request-test" },
      },
    });
  });

  it("checks evidence protocol values through generated types", () => {
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

  it("checks account preference values through generated types", () => {
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
