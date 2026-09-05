import { describe, expect, expectTypeOf, it } from "vitest";
import { incidentStreamMessageDecoder } from "./entrypoints/collaboration.js";
import { errorRegistry } from "./entrypoints/errors.js";
import {
  extensionClientSupportRegistry,
  extensionProfileRegistry,
} from "./entrypoints/extensions.js";
import {
  type ApplyImportSessionRequest,
  type AttachBlobToEvidenceRecordRequest,
  type AttachBlobToEvidenceRecordResponse,
  buildHTTPOperationPath,
  type CancelJobRequest,
  type CreateIncidentMembershipRequest,
  type CreateIncidentMembershipResponse,
  type CreateIncidentRequest,
  type CreateIncidentResponse,
  type CreateObjectBlobSlotRequest,
  type CreateObjectBlobSlotResponse,
  type CreateViewRowResponse,
  type ErrorEnvelope,
  encodeHTTPOperationQuery,
  type GetCurrentAccountPreferencesResponse,
  type GetCurrentAccountProfileResponse,
  type GetDeploymentUserResponse,
  type GetRecordHistoryResponse,
  type HTTPOperationRequest,
  type HTTPOperationResponse,
  httpOperationBindings,
  type IssueEvidencePreviewHandleRequest,
  type IssueEvidencePreviewHandleResponse,
  type ListImportUnitsResponse,
  type LoginLocalUserRequest,
  type LoginLocalUserResponse,
  type LogoutCurrentSessionRequest,
  type LogoutCurrentSessionResponse,
  type PatchCurrentAccountProfileRequest,
  type PatchCurrentAccountProfileResponse,
  type PutCurrentAccountPreferencesRequest,
  type PutCurrentAccountPreferencesResponse,
  type QueryWorkbookViewRequest,
  type QueryWorkbookViewResponse,
  type RecordHistoryData,
  type RecordHistoryItem,
  type SafeUserResource,
  type SelectImportUnitResponse,
  type SessionResource,
  type ViewCell,
  type ViewRow,
  validateHTTPOperationResponse,
} from "./entrypoints/http.js";
import {
  networkFlowContractDescriptor,
  networkFlowDecoders,
  networkFlowErrorRegistry,
} from "./entrypoints/network-flow.js";
import {
  listViewSchemaRegistryEntries,
  viewSchemaRegistry,
} from "./entrypoints/view-schemas.js";

type ObjectBlobUploadTarget =
  CreateObjectBlobSlotResponse["data"]["upload_target"];
type AccountProfileResource = GetCurrentAccountProfileResponse["data"];
type AccountPreferencesResource = GetCurrentAccountPreferencesResponse["data"];
type DensityMode = Exclude<AccountPreferencesResource["density_mode"], null>;

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

function expectDeepFrozen(
  value: unknown,
  visited = new WeakSet<object>(),
): void {
  if (value === null || typeof value !== "object" || visited.has(value)) return;
  visited.add(value);
  expect(Object.isFrozen(value)).toBe(true);
  for (const child of Object.values(value)) expectDeepFrozen(child, visited);
}

describe("@cartulary/protocol-ts family conformance", () => {
  it("emits owner-selected registries as deeply readonly stable values", () => {
    for (const registry of [
      errorRegistry,
      extensionClientSupportRegistry,
      extensionProfileRegistry,
      networkFlowErrorRegistry,
      viewSchemaRegistry,
    ]) {
      expectDeepFrozen(registry);
    }
    expect(errorRegistry.registry_id).toBe("cartulary.errors.phase3.v1");
    expect(extensionProfileRegistry.schema_id).toBe(
      "cartulary.extension_profile_registry.v1",
    );
    expect(networkFlowErrorRegistry.schema_id).toBe(
      "cartulary.network_flow_error_contracts.v1",
    );
    expect(viewSchemaRegistry.registry_id).toBe(
      "cartulary.view_schemas.base.v1",
    );
  });

  it("requires complete incident paging and consistent continuation metadata", () => {
    const validate = (paging: unknown) =>
      validateHTTPOperationResponse("listVisibleIncidents", {
        data: { incidents: [] },
        meta: {
          request_id: "request-test",
          ...(paging === undefined ? {} : { paging }),
        },
      });
    for (const paging of [
      { limit: 100, has_more: false, next_cursor: null },
      { limit: 1, has_more: true, next_cursor: "opaque next +/%" },
      { limit: 500, has_more: false, next_cursor: null },
    ])
      expect(validate(paging).ok).toBe(true);
    for (const paging of [
      undefined,
      null,
      {},
      { limit: 100, has_more: false },
      { limit: 0, has_more: false, next_cursor: null },
      { limit: 501, has_more: false, next_cursor: null },
      { limit: 1.5, has_more: false, next_cursor: null },
      { limit: 100, has_more: true, next_cursor: null },
      { limit: 100, has_more: true, next_cursor: "" },
      { limit: 100, has_more: false, next_cursor: "unexpected" },
    ])
      expect(validate(paging).ok, JSON.stringify(paging)).toBe(false);
  });

  it("exposes deterministic generated HTTP operation bindings without payload leakage", () => {
    expect([
      buildHTTPOperationPath("loginLocalUser"),
      buildHTTPOperationPath("logoutCurrentSession"),
      buildHTTPOperationPath("createIncident"),
      buildHTTPOperationPath("createIncidentMembership", {
        incident_id: "incident/id",
      }),
    ]).toEqual([
      "/api/v1/auth/login",
      "/api/v1/auth/logout",
      "/api/v1/incidents",
      "/api/v1/incidents/incident%2Fid/memberships",
    ]);
    expect(
      [
        "loginLocalUser",
        "logoutCurrentSession",
        "createIncident",
        "createIncidentMembership",
      ].map((operationID) => {
        const binding =
          httpOperationBindings[
            operationID as keyof typeof httpOperationBindings
          ];
        return [binding.method, binding.success_statuses];
      }),
    ).toEqual([
      ["POST", [200]],
      ["POST", [200]],
      ["POST", [200, 201]],
      ["POST", [200, 201]],
    ]);

    const loginRequest: LoginLocalUserRequest = {
      password: "password",
      username: "operator@example.test",
    };
    const incidentRequest: CreateIncidentRequest = {
      client_txn_id: "txn-incident",
      incident_key: "INC-1",
      title: "Incident",
    };
    const membershipRequest: CreateIncidentMembershipRequest = {
      client_txn_id: "txn-membership",
      email: "member@example.test",
      role: "reviewer",
    };
    const logoutRequest: LogoutCurrentSessionRequest = undefined;
    expect(loginRequest.username).toBe("operator@example.test");
    expect(incidentRequest.incident_key).toBe("INC-1");
    expect(membershipRequest).toEqual(
      expect.objectContaining({ role: "reviewer" }),
    );
    expect(logoutRequest).toBeUndefined();
    expectTypeOf<
      LoginLocalUserResponse["data"]
    >().toEqualTypeOf<SessionResource>();
    expectTypeOf<LogoutCurrentSessionResponse["data"]>().toMatchTypeOf<{
      logged_out: true;
    }>();
    expectTypeOf<CreateIncidentResponse["data"]>().toMatchTypeOf<{
      incident_id: string;
    }>();
    expectTypeOf<CreateIncidentMembershipResponse["data"]>().toMatchTypeOf<{
      role: "viewer" | "editor" | "reviewer" | "admin";
    }>();
    expectTypeOf<
      GetDeploymentUserResponse["data"]
    >().toEqualTypeOf<SafeUserResource>();
    expectTypeOf<
      CreateViewRowResponse["data"]["row"]
    >().toEqualTypeOf<ViewRow>();
    expectTypeOf<ViewRow["cells"][string]>().toEqualTypeOf<ViewCell>();
    expectTypeOf<
      GetRecordHistoryResponse["data"]
    >().toEqualTypeOf<RecordHistoryData>();
    expectTypeOf<
      RecordHistoryData["items"][number]
    >().toEqualTypeOf<RecordHistoryItem>();

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

    const futureVocabularyAuditEnvelope = {
      data: {
        audit_events: [
          {
            action_code: "future_owner_action",
            actor_kind: "system",
            actor_user_id: null,
            audit_event_id: "00000000-0000-4000-8000-000000000001",
            changes: [
              {
                after: "safe",
                before: null,
                field_path: "future_field",
                value_state: "visible",
              },
            ],
            occurred_at: "2026-08-04T02:00:00Z",
            reason_code: null,
            scope_id: null,
            scope_kind: "deployment",
            source: "system",
            target_id: "future-target",
            target_kind: "future_owner_target",
          },
        ],
      },
      meta: { request_id: "request-future-audit" },
    };
    for (const operationID of [
      "listAdministrativeAuditEvents",
      "listIncidentMembershipAuditEvents",
    ] as const) {
      expect(
        validateHTTPOperationResponse(
          operationID,
          futureVocabularyAuditEnvelope,
        ),
      ).toEqual({ ok: true });
    }

    for (const unsafeAuditEnvelope of [
      {
        ...futureVocabularyAuditEnvelope,
        data: {
          audit_events: [
            {
              ...futureVocabularyAuditEnvelope.data.audit_events[0],
              changes: [
                {
                  after: "must-not-leak",
                  before: null,
                  field_path: "password",
                  value_state: "redacted",
                },
              ],
            },
          ],
        },
      },
      {
        ...futureVocabularyAuditEnvelope,
        data: {
          audit_events: [
            {
              ...futureVocabularyAuditEnvelope.data.audit_events[0],
              changes: [
                {
                  after: { nested_secret: "must-not-leak" },
                  before: null,
                  field_path: "safe_field",
                  value_state: "visible",
                },
              ],
            },
          ],
        },
      },
      {
        ...futureVocabularyAuditEnvelope,
        data: {
          audit_events: [
            {
              ...futureVocabularyAuditEnvelope.data.audit_events[0],
              future_structure: true,
            },
          ],
        },
      },
    ]) {
      const validation = validateHTTPOperationResponse(
        "listAdministrativeAuditEvents",
        unsafeAuditEnvelope,
      );
      expect(validation).toEqual(
        expect.objectContaining({
          ok: false,
          schemaId: "cartulary.core_http.AdministrativeAuditEnvelope.v1",
        }),
      );
      expect(JSON.stringify(validation)).not.toContain("must-not-leak");
    }
  });

  it("exposes stable current-profile view schema identifiers", () => {
    const registry = viewSchemaRegistry;
    const ids = listViewSchemaRegistryEntries().map(
      (entry) => entry.view_schema_id,
    );

    expect(registry.registry_id).toBe("cartulary.view_schemas.base.v1");
    expect(ids).toEqual(expect.arrayContaining([...requiredBaseViewSchemaIds]));
    for (const viewSchemaId of requiredBaseViewSchemaIds) {
      expect(
        registry.view_schemas.find(
          (entry) => entry.view_schema_id === viewSchemaId,
        )?.artifact_path,
      ).toBe(`contracts/view-schemas/${viewSchemaId}.json`);
    }
    expect(
      registry.view_schemas.find(
        (entry) => entry.view_schema_id === "cartulary.view.missing.v1",
      ),
    ).toBeUndefined();
  });

  it("exposes the typed error and reason-code registry", () => {
    const requireReasonCodeRegistry = (errorCode: string) => {
      const registry = errorRegistry.reason_registries.find(
        (candidate) => candidate.error_code === errorCode,
      );
      if (!registry) throw new Error(`missing reason registry ${errorCode}`);
      return registry;
    };
    expect(errorRegistry.registry_id).toBe("cartulary.errors.phase3.v1");
    expect(
      requireReasonCodeRegistry("invalid_mutation_payload").reason_codes,
    ).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: "unknown_view_schema" }),
        expect.objectContaining({ code: "unknown_field_key" }),
      ]),
    );
    const startupReasonCodes = requireReasonCodeRegistry(
      "invalid_startup_request",
    ).reason_codes;
    expect(startupReasonCodes).toEqual(
      expect.arrayContaining([
        expect.objectContaining({ code: "unsupported_sheet_ref_kind" }),
      ]),
    );
    expect(
      startupReasonCodes.map((reason) => String(reason.code)),
    ).not.toContain("ambiguous_explicit_sheet_ref");
    expect(
      errorRegistry.reason_registries.find(
        (registry) => String(registry.error_code) === "missing_error_code",
      ),
    ).toBeUndefined();
    expect(() => requireReasonCodeRegistry("missing_error_code")).toThrow(
      "missing reason registry missing_error_code",
    );
  });

  it("exposes typed extension registry profiles", () => {
    expect(
      extensionProfileRegistry.profiles.map((profile) => profile.profile_id),
    ).toEqual([
      "enterprise_authentication",
      "import",
      "incident_portability",
      "network_flow_activity",
      "reference_pack",
      "snapshot_reporting",
    ]);
    expect(
      extensionProfileRegistry.profiles.find(
        (profile) => profile.profile_id === "reference_pack",
      )?.route_families,
    ).toContain("/api/v1/reference-packs");
    expect(
      extensionProfileRegistry.profiles.find(
        (profile) => String(profile.profile_id) === "missing_profile",
      ),
    ).toBeUndefined();
  });

  it("implements extension discovery tolerance through the HTTP owner policy", () => {
    const additiveEnvelope = {
      data: {
        extensions: [
          {
            capabilities: [],
            claimable: true,
            claimed: true,
            contract_major: 2,
            future_additive_member: { executable: "must-remain-inert" },
            profile_id: "network_flow_activity",
            route_families: ["/api/v1/incidents/{incident_id}/network-flow"],
            workspace_keys: ["network_analysis"],
          },
        ],
      },
      meta: { request_id: "request-test" },
    };
    expect(
      validateHTTPOperationResponse(
        "listDeploymentExtensions",
        additiveEnvelope,
      ),
    ).toEqual({ ok: true });
    for (const invalidEnvelope of [
      {
        ...additiveEnvelope,
        data: {
          extensions: [
            {
              ...additiveEnvelope.data.extensions[0],
              route_families: ["/api/v1/z", "/api/v1/a"],
            },
          ],
        },
      },
      { ...additiveEnvelope, future_envelope_member: true },
    ]) {
      const validation = validateHTTPOperationResponse(
        "listDeploymentExtensions",
        invalidEnvelope,
      );
      expect(validation).toEqual(
        expect.objectContaining({
          ok: false,
          schemaId: "cartulary.core_http.ExtensionDiscoveryEnvelope.v1",
        }),
      );
      expect(JSON.stringify(validation)).not.toContain("must-remain-inert");
    }
  });

  it("decodes exact Network Flow contracts without exposing payload data on failure", () => {
    expect(networkFlowContractDescriptor).toEqual({
      profile_id: "network_flow_activity",
      contract_major: 5,
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
        instancePath: "/raw_source_value",
        reasonCategory: "unknown_member",
        schemaId: "cartulary.network_flow.table_list.v1",
      },
    });
    expect(JSON.stringify(invalid)).not.toContain("must-not-leak");
  });

  it("decodes every server WebSocket family with its complete envelope", () => {
    const envelope = {
      emitted_at: "2026-08-03T23:00:00Z",
      event_id: "event-1",
      incident_id: "incident-1",
    } as const;
    const messages = [
      {
        ...envelope,
        type: "hello_ack",
        payload: {
          connection_id: "connection-1",
          heartbeat_interval_ms: 15_000,
          presence_ttl_ms: 45_000,
          resume_token: "resume-1",
          resume_window_ms: 300_000,
          server_time: "2026-08-03T23:00:00Z",
        },
      },
      {
        ...envelope,
        type: "resume_ack",
        payload: {
          resume_token: "resume-2",
          server_high_water_stream_seq: 4,
          status: "replayed",
        },
      },
      { ...envelope, type: "presence_snapshot", payload: { presences: [] } },
      {
        ...envelope,
        type: "presence_delta",
        payload: {
          delta_kind: "remove",
          presence: { connection_id: "connection-1" },
        },
      },
      {
        ...envelope,
        type: "record_changed",
        payload: {
          actor_user_id: "user-1",
          affected_views: [
            {
              change_kind: "invalidate",
              view_schema_id: "cartulary.view.timeline.v2",
            },
          ],
          change_set_id: "change-1",
          changed_field_keys: [],
          client_txn_id: "txn-1",
          record_id: "record-1",
          row_version: 2,
        },
        stream_seq: 1,
      },
      {
        ...envelope,
        type: "extension_resource_changed",
        payload: {
          change_kind: "invalidate",
          extension_profile_id: "network_flow_activity",
          reason_code: "renamed",
          resource_id: "resource-1",
          resource_kind: "network_flow_table",
        },
        stream_seq: 2,
      },
      {
        ...envelope,
        type: "job_progress",
        payload: {
          job_id: "job-1",
          progress: { completed: 0, total: null },
          scope: { incident_id: "incident-1", kind: "incident" },
          status: "queued",
          updated_at: "2026-08-03T23:00:00Z",
        },
        stream_seq: 3,
      },
      { ...envelope, type: "ping", payload: {} },
      {
        ...envelope,
        type: "error",
        payload: {
          code: "invalid_message",
          message: "Invalid.",
          retryable: false,
        },
      },
      {
        ...envelope,
        type: "session_revoked",
        payload: { reason_code: "membership_removed" },
      },
    ] as const;

    expect(messages.map((message) => message.type)).toEqual([
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
        expect(result.value).not.toBe(message);
        expect(result.value.payload).not.toBe(message.payload);
      }
    }
  });

  it("projects replayable messages to fresh known members at every additive boundary", () => {
    const envelope = {
      emitted_at: "2026-08-03T23:00:00Z",
      event_id: "event-1",
      incident_id: "incident-1",
      ignored_envelope: { route: "/must-not-dispatch" },
    } as const;
    const fixtures = [
      {
        input: {
          ...envelope,
          type: "record_changed",
          payload: {
            actor_user_id: "user-1",
            affected_views: [
              {
                change_kind: "patch",
                ignored_view: "must-not-render",
                patch_cells: {
                  cells: {
                    "timeline.title": {
                      ignored_cell: "must-not-render",
                      value: { retained_owner_value: true },
                    },
                  },
                  group_values: {
                    "timeline.group": { retained_owner_value: true },
                  },
                  ignored_patch: "must-not-render",
                  record_id: "record-1",
                  row_version: 2,
                },
                view_schema_id: "cartulary.view.timeline.v2",
              },
            ],
            change_set_id: "change-1",
            changed_field_keys: ["timeline.title"],
            client_txn_id: "txn-1",
            ignored_payload: { action: "must-not-execute" },
            record_id: "record-1",
            row_version: 2,
          },
          stream_seq: 1,
        },
        forbidden: [
          "ignored_envelope",
          "ignored_payload",
          "ignored_view",
          "ignored_patch",
          "ignored_cell",
        ],
      },
      {
        input: {
          ...envelope,
          type: "extension_resource_changed",
          payload: {
            change_kind: "invalidate",
            extension_profile_id: "network_flow_activity",
            ignored_payload: { discriminator: "must-not-dispatch" },
            reason_code: "retained_owner_reason",
            resource_id: "resource-1",
            resource_kind: "network_flow_table",
            workspace_refs: [
              {
                extension_profile_id: "network_flow_activity",
                ignored_workspace: "must-not-render",
                kind: "extension_workspace",
                workspace_key: "network_analysis",
              },
            ],
          },
          stream_seq: 2,
        },
        forbidden: ["ignored_envelope", "ignored_payload", "ignored_workspace"],
      },
      {
        input: {
          ...envelope,
          type: "job_progress",
          payload: {
            error_summary: {
              code: "job_failed",
              details: { retained_owner_detail: { nested: true } },
              ignored_summary: "must-not-render",
              message: "Failed.",
              retryable: false,
            },
            ignored_payload: { route: "/must-not-follow" },
            job_id: "job-1",
            progress: { completed: 1, ignored_progress: 100, total: 2 },
            scope: {
              ignored_scope: "deployment",
              incident_id: "incident-1",
              kind: "incident",
            },
            status: "failed",
            updated_at: "2026-08-03T23:00:00Z",
          },
          stream_seq: 3,
        },
        forbidden: [
          "ignored_envelope",
          "ignored_payload",
          "ignored_progress",
          "ignored_scope",
          "ignored_summary",
        ],
      },
    ] as const;

    for (const fixture of fixtures) {
      const result = incidentStreamMessageDecoder.decode(fixture.input);
      expect(result.ok).toBe(true);
      if (!result.ok) continue;
      expect(result.value).not.toBe(fixture.input);
      const encoded = JSON.stringify(result.value);
      for (const member of fixture.forbidden) {
        expect(encoded).not.toContain(member);
      }
      expect(encoded).toContain("retained_owner");
    }
  });

  it("rejects incomplete envelopes, invalid payloads, and non-server families", () => {
    const validRecordPayload = {
      actor_user_id: "user-1",
      affected_views: [
        {
          change_kind: "invalidate",
          view_schema_id: "cartulary.view.timeline.v2",
        },
      ],
      change_set_id: "change-1",
      changed_field_keys: [],
      client_txn_id: "txn-1",
      record_id: "record-1",
      row_version: 2,
    } as const;
    for (const invalid of [
      { type: "unknown", payload: { secret: "must-not-leak" } },
      { type: "hello", payload: { secret: "must-not-leak" } },
      {
        emitted_at: "2026-08-03T23:00:00Z",
        incident_id: "incident-1",
        payload: validRecordPayload,
        stream_seq: 1,
        type: "record_changed",
      },
      {
        emitted_at: "2026-08-03T23:00:00Z",
        event_id: "event-1",
        incident_id: "incident-1",
        payload: validRecordPayload,
        type: "record_changed",
      },
      {
        emitted_at: "2026-08-03T23:00:00Z",
        event_id: "event-1",
        incident_id: "incident-1",
        payload: {},
        stream_seq: 99,
        type: "ping",
      },
      {
        emitted_at: "2026-08-03T23:00:00Z",
        event_id: "event-1",
        incident_id: "incident-1",
        payload: { ...validRecordPayload, affected_views: [] },
        stream_seq: 1,
        type: "record_changed",
      },
      {
        emitted_at: "2026-08-03T23:00:00Z",
        event_id: "event-1",
        incident_id: "incident-1",
        payload: {
          job_id: "job-1",
          progress: { completed: 0, total: null },
          scope: { kind: "deployment" },
          status: "queued",
          updated_at: "2026-08-03T23:00:00Z",
        },
        stream_seq: 3,
        type: "job_progress",
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

  it("checks evidence protocol values through generated types", () => {
    const createRequest = {
      incident_id: "incident-1",
      client_txn_id: "txn-create-blob",
      byte_size: 42,
      filename_hint: "evidence.txt",
      content_type_hint: "text/plain",
      sha256_hex: null,
    } satisfies CreateObjectBlobSlotRequest;
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
    } satisfies CreateObjectBlobSlotResponse;
    const attachRequest = {
      object_blob_id: createEnvelope.data.object_blob_id,
      base_row_version: 1,
      client_txn_id: "txn-attach-blob",
    } satisfies AttachBlobToEvidenceRecordRequest;
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
    } satisfies AttachBlobToEvidenceRecordResponse;
    const handleRequest = {} satisfies IssueEvidencePreviewHandleRequest;
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
    } satisfies IssueEvidencePreviewHandleResponse;
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
    } satisfies PatchCurrentAccountProfileRequest;
    const preferencesPutRequest = {
      base_preferences_version: preferencesResource.preferences_version,
      client_txn_id: "txn-preferences",
      density_mode: null,
    } satisfies PutCurrentAccountPreferencesRequest;
    const profileEnvelope = {
      data: profileResource,
      meta: { request_id: "req-profile" },
    } satisfies PatchCurrentAccountProfileResponse;
    const preferencesEnvelope = {
      data: preferencesResource,
      meta: { request_id: "req-preferences" },
    } satisfies PutCurrentAccountPreferencesResponse;

    expect(profilePatchRequest.display_name).toBe("Operator Prime");
    expect(preferencesPutRequest.density_mode).toBeNull();
    expect(profileEnvelope.data.user_id).toBe(profileResource.user_id);
    expect(preferencesEnvelope.data.density_mode).toBe("compact");
  });
});
