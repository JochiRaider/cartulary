import { describe, expect, it } from "vitest";

import { incidentStreamMessageDecoder } from "./collaboration.js";
import { errorRegistry } from "./errors.js";
import {
  extensionClientSupportRegistry,
  extensionProfileRegistry,
} from "./extensions.js";
import {
  buildHTTPOperationPath,
  errorEnvelopeDecoder,
  httpOperationBindings,
  validateHTTPOperationResponse,
} from "./http.js";
import { importTargetRegistry } from "./import-targets.js";
import {
  networkFlowContractDescriptor,
  networkFlowDecoders,
  networkFlowErrorRegistry,
  networkFlowMappingRegistry,
  networkFlowPresentationRegistry,
} from "./network-flow.js";
import * as viewSchemasEntrypoint from "./view-schemas.js";
import {
  listViewSchemaRegistryEntries,
  viewSchemaRegistry,
} from "./view-schemas.js";

describe("Protocol-TS authored family entrypoints", () => {
  it("keeps collaboration decoding family-confined and payload-safe", () => {
    const ping = {
      emitted_at: "2026-08-03T23:00:00Z",
      event_id: "event-1",
      ignored: "must-not-leak",
      incident_id: "incident-1",
      payload: { ignored: "must-not-leak" },
      type: "ping",
    };
    const decoded = incidentStreamMessageDecoder.decode(ping);
    expect(decoded).toEqual({
      ok: true,
      value: {
        emitted_at: "2026-08-03T23:00:00Z",
        event_id: "event-1",
        incident_id: "incident-1",
        payload: {},
        type: "ping",
      },
    });
    if (decoded.ok) {
      expect(decoded.value).not.toBe(ping);
    }

    // Only the family differs from the accepted ping above.
    const invalid = incidentStreamMessageDecoder.decode({
      ...ping,
      type: "unknown",
    });
    expect(invalid).toEqual(
      expect.objectContaining({
        ok: false,
        error: expect.objectContaining({
          boundary: "generated_protocol",
          schemaId: "cartulary.ws.incident_stream_message.v1",
        }),
      }),
    );
    expect(JSON.stringify(invalid)).not.toContain("must-not-leak");
  });

  it("exposes stable typed registry singletons by owner", () => {
    for (const registry of [
      errorRegistry,
      extensionClientSupportRegistry,
      extensionProfileRegistry,
      importTargetRegistry,
      networkFlowErrorRegistry,
      viewSchemaRegistry,
    ]) {
      expect(
        Object.isFrozen(registry),
        JSON.stringify(Object.keys(registry)),
      ).toBe(true);
    }
    expect(Object.isFrozen(importTargetRegistry.targets)).toBe(true);
    // Object.isFrozen(undefined) is true, so require targets before checking each.
    expect(importTargetRegistry.targets.length).toBeGreaterThan(0);
    for (const target of importTargetRegistry.targets)
      expect(Object.isFrozen(target)).toBe(true);
    expect(errorRegistry.registry_id).toBe("cartulary.errors.phase3.v1");
    expect(extensionClientSupportRegistry.schema_id).toBe(
      "cartulary.client_extension_support_registry.v1",
    );
    expect(extensionProfileRegistry.schema_id).toBe(
      "cartulary.extension_profile_registry.v1",
    );
    expect(importTargetRegistry.schema_id).toBe(
      "cartulary.import_target_frontend_projection.v1",
    );
  });

  it("owns HTTP metadata, builders, types, and payload-safe decoders", () => {
    expect(httpOperationBindings.getIncident.method).toBe("GET");
    expect(
      buildHTTPOperationPath("getIncident", { incident_id: "incident/1" }),
    ).toBe("/api/v1/incidents/incident%2F1");
    expect(
      validateHTTPOperationResponse("uploadObjectBlobContent", undefined),
    ).toEqual({ ok: true });

    const error = {
      code: "invalid",
      details: { secret: "must-not-leak" },
      message: "Invalid.",
      request_id: "request-error",
      retryable: false,
      status: 400,
    };
    expect(errorEnvelopeDecoder.decode({ error }).ok).toBe(true);
    const { status: _omitted, ...missingStatus } = error;
    const invalid = errorEnvelopeDecoder.decode({ error: missingStatus });
    expect(invalid).toEqual({
      ok: false,
      error: {
        boundary: "generated_protocol",
        instancePath: "/error/status",
        reasonCategory: "required_member",
        schemaId: "cartulary.core_http.ErrorEnvelope.v1",
      },
    });
    expect(JSON.stringify(invalid)).not.toContain("must-not-leak");
  });

  it("keeps Network Flow projections and exact decoders in one family", () => {
    expect(networkFlowContractDescriptor).toEqual({
      profile_id: "network_flow_activity",
      contract_major: 10,
    });
    expect(networkFlowMappingRegistry).toMatchObject({
      profile_id: networkFlowContractDescriptor.profile_id,
      schema_id: "cartulary.network_flow.mapping_registry.v3",
    });
    expect(networkFlowPresentationRegistry).toMatchObject({
      profile_id: networkFlowContractDescriptor.profile_id,
      schema_id: "cartulary.network_flow.presentation.v3",
    });

    const invalid = networkFlowDecoders.tableList.decode({
      schema_id: "cartulary.network_flow_table_list.v3",
      tables: [],
      meta: { count: 0 },
      raw_source_value: "must-not-leak",
    });
    expect(invalid).toEqual(
      expect.objectContaining({
        ok: false,
        error: expect.objectContaining({
          reasonCategory: "unknown_member",
        }),
      }),
    );
    expect(JSON.stringify(invalid)).not.toContain("must-not-leak");
  });

  it("exposes only view-schema-specific registry access", () => {
    expect(Object.keys(viewSchemasEntrypoint).sort()).toEqual([
      "listViewSchemaRegistryEntries",
      "viewReferenceRegistry",
      "viewSchemaRegistry",
    ]);
    expect(listViewSchemaRegistryEntries()).toBe(
      viewSchemaRegistry.view_schemas,
    );
    expect(
      listViewSchemaRegistryEntries().find(
        (entry) => entry.view_schema_id === "cartulary.view.timeline.v2",
      ),
    ).toEqual(
      expect.objectContaining({
        artifact_path: "contracts/view-schemas/cartulary.view.timeline.v2.json",
      }),
    );
  });
});
