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

    const invalid = incidentStreamMessageDecoder.decode({
      type: "unknown",
      payload: { secret: "must-not-leak" },
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
      expect(Object.isFrozen(registry)).toBe(true);
    }
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

    const invalid = errorEnvelopeDecoder.decode({
      error: { code: "invalid", secret: "must-not-leak" },
    });
    expect(invalid).toEqual(
      expect.objectContaining({
        ok: false,
        error: expect.objectContaining({
          boundary: "generated_protocol",
          schemaId: "cartulary.core_http.ErrorEnvelope.v1",
        }),
      }),
    );
    expect(JSON.stringify(invalid)).not.toContain("must-not-leak");
  });

  it("keeps Network Flow projections and exact decoders in one family", () => {
    expect(networkFlowContractDescriptor).toEqual({
      profile_id: "network_flow_activity",
      contract_major: 5,
    });
    expect(networkFlowMappingRegistry).toBeDefined();
    expect(networkFlowPresentationRegistry).toBeDefined();

    const invalid = networkFlowDecoders.tableList.decode({
      schema_id: "cartulary.network_flow.table_list.v1",
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
