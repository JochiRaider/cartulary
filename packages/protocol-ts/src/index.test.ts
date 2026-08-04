import { describe, expect, expectTypeOf, it } from "vitest";

import * as protocolFacade from "./index";
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
  incidentStreamMessageDecoder,
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
  it("characterizes the exact root runtime surface and stable artifact identity", () => {
    expect(Object.keys(protocolFacade).sort()).toEqual([
      "buildHTTPOperationPath",
      "createGeneratedDecoder",
      "decodeExtensionDiscoveryItem",
      "encodeHTTPOperationQuery",
      "extensionDiscoveryDecoder",
      "getContractArtifact",
      "getErrorRegistryContract",
      "getExtensionProfile",
      "getExtensionRegistryContract",
      "getNetworkFlowErrorRegistry",
      "getReasonCodeRegistry",
      "getViewSchemaRegistryContract",
      "getViewSchemaRegistryEntry",
      "httpOperationBindings",
      "importTargetRegistry",
      "incidentStreamMessageDecoder",
      "listContractArtifactFamilies",
      "listErrorArtifacts",
      "listExtensionArtifacts",
      "listExtensionProfiles",
      "listReasonCodeRegistries",
      "listViewSchemaArtifacts",
      "listViewSchemaRegistryEntries",
      "listWSArtifacts",
      "networkFlowContractDescriptor",
      "networkFlowDecoders",
      "networkFlowMappingRegistry",
      "networkFlowPresentationRegistry",
      "parseContractArtifact",
      "requireContractArtifact",
      "requireContractArtifactJSON",
      "requireExtensionProfile",
      "requireReasonCodeRegistry",
      "requireViewSchemaRegistryEntry",
      "validateHTTPOperationResponse",
    ]);

    const firstFamilies = listContractArtifactFamilies();
    expect(listContractArtifactFamilies()).toBe(firstFamilies);
    expect(protocolFacade.listWSArtifacts()).toBe(firstFamilies.wsArtifacts);
    expect(protocolFacade.listViewSchemaArtifacts()).toBe(
      firstFamilies.viewSchemaArtifacts,
    );
    expect(protocolFacade.listErrorArtifacts()).toBe(
      firstFamilies.errorArtifacts,
    );
    expect(protocolFacade.listExtensionArtifacts()).toBe(
      firstFamilies.extensionArtifacts,
    );
    expect(Object.isFrozen(firstFamilies)).toBe(true);
    expect(Object.isFrozen(incidentStreamMessageDecoder)).toBe(true);
    expect(Object.isFrozen(networkFlowDecoders)).toBe(true);
    expect(protocolFacade.incidentStreamMessageDecoder).toBe(
      incidentStreamMessageDecoder,
    );
    expect(protocolFacade.networkFlowDecoders.tableList).toBe(
      networkFlowDecoders.tableList,
    );

    expect(
      Object.values(firstFamilies)
        .flat()
        .map((artifact) => `${artifact.path}:${artifact.sha256}`),
    ).toEqual([
      "contracts/ws/index.schema.json:1f0155f872991f00afe7c2ea83269b656f8ab873cfcadd411e0a8a0a1425eeaa",
      "contracts/view-schemas/cartulary.view.assessments.v1.json:d45c00b3df0ce14104aa6cb338700cfda81c5efa7662c40dda335140edc53640",
      "contracts/view-schemas/cartulary.view.comm_log.v1.json:75ce6aa473202be2e4eb702ea75d6761ce3211badf04325b5ab16f3ef5c10348",
      "contracts/view-schemas/cartulary.view.decisions.v1.json:1bbf524b237a07e45dafa9e13872882a3c1ffcc0a360cfeac9ab8f0109c57677",
      "contracts/view-schemas/cartulary.view.evidence.v1.json:9784f4262d447bcde2d0dd3decb0d512d2c46a293c5a6fbf20d81202dfdbaf79",
      "contracts/view-schemas/cartulary.view.findings.v1.json:2a0593f2196030a3747dc3ee666040068794f2ac595fe997791bd7b1ec627210",
      "contracts/view-schemas/cartulary.view.forensic_keywords.v1.json:9fae534b5548cee5952ce811540f82f9d9f6bd6cb5dc715109d75ac002dfc577",
      "contracts/view-schemas/cartulary.view.handoff.v1.json:3920ca6215c0ed2aa995a2e15ab4f9f727700d5b640c93c646f399107dddda8d",
      "contracts/view-schemas/cartulary.view.hosts.v1.json:ec1bdcd1375b826fd0e483d74df49c5c70f0cf7de0cba298a42b2a328ec50489",
      "contracts/view-schemas/cartulary.view.identities.v1.json:76f9dee4f302ff471a149c1954ba77a6efefb231e3db4c92439c6b4a1099f0cc",
      "contracts/view-schemas/cartulary.view.indicators.v1.json:4f8f8263073cd59e550ac5fa137631cbcf4c435bb1bde792236d4c9d9ea5d815",
      "contracts/view-schemas/cartulary.view.investigative_queries.v1.json:667230d66a48965c758be280396d733c3b221947a8d6e9416cc5257fd2309984",
      "contracts/view-schemas/cartulary.view.lesson.v1.json:b5c546a6b65c11de5fe3cc929634d17d62d93e6d6764fcc01f8f27027f235b2e",
      "contracts/view-schemas/cartulary.view.notes.v1.json:8ed810f3cf03e3ed9e4de79d7b41c1c305ccac71ac8ef4278c63717fd39c858b",
      "contracts/view-schemas/cartulary.view.parties.v1.json:9d9d9b9d3770b81518c732ec424c6eae772744c77d55017841cdc69263eb09a9",
      "contracts/view-schemas/cartulary.view.status_review.v1.json:7e6d6bd4bf970a99ffe044e6a07777c0dfb9e1cf5c1aa4f1ae6cce4e2bc1328f",
      "contracts/view-schemas/cartulary.view.task_requests.v1.json:4009aed0c9e8ac17c9f1075eb0a4180f37bbf12c2f6fd7d6f4eef89cf20d6e45",
      "contracts/view-schemas/cartulary.view.timeline.v2.json:d21045fadef3d1ee6bcdddd73c5abce5fc21d0b2cec49fbbe57741629712bc1d",
      "contracts/view-schemas/index.json:a3cc8a9f2001e2d40150bfc8e2fbed62401118fb5b1aa854aa1be3573f1fe0f4",
      "contracts/errors/index.json:e5a857dda85be46d51134bc324ef8028aff0dc3bae4c5e049c7c9a173c0d2a22",
      "contracts/extensions/generated/client-support-registry.json:e7f906285d9138feed9c7e5105fa79681496d6337052390a0fa5ec59272de470",
      "contracts/extensions/generated/descriptors/enterprise_authentication.json:703a46d1c5274109574b912555168c33c28ab678290f43b3e43598edd2116957",
      "contracts/extensions/generated/descriptors/import.json:e47b16467ac8030a82169e17d9421d1a77fe87de0b8f77a8290242dcc0497e53",
      "contracts/extensions/generated/descriptors/incident_portability.json:a4ada5b51aa648f8aea6c16cae979663519b2e59cad36e429bb3cf73b058db55",
      "contracts/extensions/generated/descriptors/network_flow_activity.json:d4472be715f9ac9e7273de6cd0389bbbf63e158a429647547a84d0192766f327",
      "contracts/extensions/generated/descriptors/reference_pack.json:d0b7a5d664a97b113d4376b0e13088a5ce8315d859fe536f5ce9f77c4f16efc5",
      "contracts/extensions/generated/descriptors/snapshot_reporting.json:6a1f47e4f2036eddab4b932c88815a96733d9a4e5c4b69d269285c94f79b9536",
      "contracts/extensions/generated/profile-registry.json:780023873673b58663db9156497cf32867726b48a1d3560b7465f3c5e0fc2ed9",
      "contracts/network-flow/errors.v1.json:00b89a0de2b99037b25c661012707dd1fa81d0335de05bc31146cd2e3e2eff56",
      "contracts/network-flow/frontend-entrypoints.v3.json:d6e4c65d82c01a5ade347e5cd992ec7e7234cffcf6137859a4dccc4b4a90401e",
      "contracts/network-flow/index.json:daf4be3faf5e9848ec5ac70ce06ec202ff25c0636a5fa4955ba88aca9ee6626d",
      "contracts/network-flow/key-rings.v1.schema.json:e71f02f71dad4f06b997a80928497acb831a8a2794be31a56cf09fb11b44cc1a",
      "contracts/network-flow/mapping-registry.v2.json:2f406bfeb37e2f5cd12a9901f34c22ed1ceb88a98bcf0a326d6d12f7845fd381",
      "contracts/network-flow/presentation.v2.json:74567c1d348c7c31861a85584a2ccbc63cc666a970f4f2d2d538c158d5358f6c",
      "contracts/network-flow/routes.v1.json:8c3cef00088d29b29ad6b630ac08a68a212a711fd7ce86d8fbd650169f0d92fd",
      "contracts/network-flow/schemas.v1.json:94463c8964c51158d89a3d102c821a3d9a7082f918179d1f7d3c74c65a1e1ca9",
      "contracts/network-flow/timezone/tzdb-2026c.provenance.json:bf77ec3f4efc3a31e37426adae4b0c8ed574f69da6d9dbc76587da7034e91aa5",
      "contracts/network-flow/unicode17-nfc.provenance.json:10bcf7b12536e39da236678c60e01a6e32f7fb9e509ec628563cfec97b7d5ea4",
    ]);
  });

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
                "extension.value": {
                  value: "strict optional surface",
                  future_cell_member: { must_remain_inert: true },
                },
              },
              group_values: { "extension.group": "strict" },
              record_id: "00000000-0000-4000-8000-000000000001",
              row_version: 2,
              future_row_member: ["must", "remain", "inert"],
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

    const objectBlobResponse = {
      data: {
        accepted_contract: {
          byte_size: 42,
          content_type_hint: "text/plain",
          filename_hint: "evidence.txt",
          incident_id: "00000000-0000-4000-8000-000000000001",
          sha256_hex: null,
        },
        incident_id: "00000000-0000-4000-8000-000000000001",
        object_blob_id: "00000000-0000-4000-8000-000000000002",
        pending_expires_at: "2026-08-04T01:00:00Z",
        target_expires_at: "2026-08-04T00:30:00Z",
        upload_state: "pending",
        upload_target: {
          expires_at: "2026-08-04T00:30:00Z",
          headers: { "X-Upload-Contract": "generated-protocol" },
          href: "/api/v1/object-uploads/upload-token",
          method: "PUT",
        },
      },
      meta: { request_id: "request-create-blob" },
    };
    expect(
      validateHTTPOperationResponse("createObjectBlobSlot", objectBlobResponse),
    ).toEqual(expect.objectContaining({ ok: true }));
    expect(
      validateHTTPOperationResponse("createObjectBlobSlot", {
        ...objectBlobResponse,
        data: {
          ...objectBlobResponse.data,
          upload_target: {
            ...objectBlobResponse.data.upload_target,
            headers: { "X-Upload-Contract": 42 },
          },
        },
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

  it("characterizes all WebSocket message families and payload-free failures", () => {
    const messages = [
      {
        type: "hello",
        payload: { client_instance_id: "client-1", presence: {} },
      },
      {
        type: "resume",
        payload: {
          client_instance_id: "client-1",
          last_seen_stream_seq: 1,
          presence: {},
          resume_token: "resume-1",
        },
      },
      { type: "pong", payload: {} },
      { type: "presence_update", payload: { presence: {} } },
      {
        type: "hello_ack",
        payload: {
          connection_id: "connection-1",
          heartbeat_interval_ms: 10_000,
          presence_ttl_ms: 30_000,
          resume_token: "resume-1",
          resume_window_ms: 60_000,
          server_time: "2026-08-03T23:00:00Z",
        },
      },
      {
        type: "resume_ack",
        payload: {
          resume_token: "resume-2",
          server_high_water_stream_seq: 4,
          status: "resumed",
        },
      },
      { type: "presence_snapshot", payload: { presences: [] } },
      {
        type: "presence_delta",
        payload: { delta_kind: "upsert", presence: {} },
      },
      {
        type: "record_changed",
        payload: {
          actor_user_id: "user-1",
          affected_views: [],
          change_set_id: "change-1",
          changed_field_keys: [],
          client_txn_id: "txn-1",
          record_id: "record-1",
          row_version: 2,
        },
        stream_seq: 1,
      },
      {
        type: "extension_resource_changed",
        payload: {
          change_kind: "invalidate",
          extension_profile_id: "network_flow_activity",
          reason_code: "changed",
          resource_id: "resource-1",
          resource_kind: "table",
        },
        stream_seq: 2,
      },
      {
        type: "job_progress",
        payload: {
          job_id: "job-1",
          progress: { completed: 0, total: null },
          scope: { kind: "deployment" },
          status: "queued",
          updated_at: "2026-08-03T23:00:00Z",
        },
        stream_seq: 3,
      },
      { type: "ping", payload: {} },
      {
        type: "error",
        payload: {
          code: "invalid_message",
          message: "Invalid.",
          retryable: false,
        },
      },
      {
        type: "session_revoked",
        payload: { reason_code: "membership_removed" },
      },
    ] as const;

    expect(messages.map((message) => message.type)).toEqual([
      "hello",
      "resume",
      "pong",
      "presence_update",
      "hello_ack",
      "resume_ack",
      "presence_snapshot",
      "presence_delta",
      "record_changed",
      "extension_resource_changed",
      "job_progress",
      "ping",
      "error",
      "session_revoked",
    ]);
    for (const message of messages) {
      const result = incidentStreamMessageDecoder.decode(message);
      expect(result).toEqual({ ok: true, value: message });
      if (result.ok) {
        expect(result.value).toBe(message);
      }
    }

    // Non-replayable messages tolerate additive members at the message level.
    expect(
      incidentStreamMessageDecoder.decode({
        type: "ping",
        payload: {},
        stream_seq: 99,
      }),
    ).toEqual(expect.objectContaining({ ok: true }));

    for (const invalid of [
      { type: "unknown", payload: { secret: "must-not-leak" } },
      { type: "hello", payload: null, secret: "must-not-leak" },
      {
        type: "record_changed",
        payload: messages[8].payload,
        secret: "must-not-leak",
      },
      {
        type: "extension_resource_changed",
        payload: messages[9].payload,
        stream_seq: 0,
        secret: "must-not-leak",
      },
      {
        type: "job_progress",
        payload: { ...messages[10].payload, unexpected: "must-not-leak" },
        stream_seq: 3,
      },
    ]) {
      const result = incidentStreamMessageDecoder.decode(invalid);
      expect(result).toEqual(
        expect.objectContaining({
          ok: false,
          error: expect.objectContaining({
            boundary: "generated_protocol",
            schemaId: "cartulary.ws.incident_stream_message.v1",
          }),
        }),
      );
      expect(JSON.stringify(result)).not.toContain("must-not-leak");
      expect(result).not.toHaveProperty("value");
    }
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

    const knownItem = {
      capabilities: [],
      claimable: true,
      claimed: true,
      contract_major: 2,
      profile_id: "network_flow_activity",
      route_families: [
        "/api/v1/incidents/{incident_id}/network-flow",
        "/api/v1/network-flow",
      ],
      workspace_keys: ["network_analysis", "network_tables"],
      future_additive_member: { executable: "must-remain-inert" },
    };
    const decodedKnownItem =
      protocolFacade.decodeExtensionDiscoveryItem(knownItem);
    expect(Object.keys(decodedKnownItem).sort()).toEqual([
      "capabilities",
      "claimable",
      "claimed",
      "contract_major",
      "profile_id",
      "route_families",
      "workspace_keys",
    ]);
    expect(decodedKnownItem).not.toHaveProperty("future_additive_member");

    for (const invalidItem of [
      { ...knownItem, profile_id: "Invalid-Profile" },
      {
        ...knownItem,
        route_families: ["/api/v1/z", "/api/v1/a"],
      },
      {
        ...knownItem,
        workspace_keys: ["network_analysis", "network_analysis"],
      },
      { ...knownItem, capabilities: ["future_execution"] },
      { ...knownItem, contract_major: null },
    ]) {
      expect(() =>
        protocolFacade.decodeExtensionDiscoveryItem(invalidItem),
      ).toThrow("invalid extension discovery");
    }

    for (const malformedEnvelope of [
      null,
      {},
      { data: { extensions: "not-an-array" }, meta: { request_id: "req" } },
      {
        data: {
          extensions: [knownItem, { ...knownItem, profile_id: "invalid-id!" }],
        },
        meta: { request_id: "req" },
      },
      {
        data: { extensions: [knownItem] },
        meta: { request_id: 42 },
      },
    ]) {
      const result = extensionDiscoveryDecoder.decode(malformedEnvelope);
      expect(result).toEqual(
        expect.objectContaining({
          ok: false,
          error: expect.objectContaining({
            boundary: "generated_protocol",
            schemaId: "cartulary.core_http.ExtensionDiscoveryEnvelope.v1",
          }),
        }),
      );
      expect(result).not.toHaveProperty("value");
      expect(JSON.stringify(result)).not.toContain("must-remain-inert");
    }
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
