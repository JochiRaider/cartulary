import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";
import {
  buildMergePlan,
  entityContractColumnWidth,
  entityGroupLabel,
  entityRowFromApi,
} from "./entityWorkbookModel";
import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
} from "./workbookSurfaceRegistry";

function entityRow(
  recordId: string,
  cells: EntityApiRow["cells"],
): EntityApiRow {
  return {
    record_id: recordId,
    row_version: 7,
    cells,
  };
}

describe("entityWorkbookModel", () => {
  it("normalizes host and identity rows with label fallbacks, aliases, identifiers, and counts", () => {
    const host = entityRowFromApi(
      entityRow("host-1", {
        "host.display_name": { value: "" },
        "host.hostname": { value: "endpoint-01" },
        "host.host_state": { value: "active" },
        "host.linked_event_count": { value: 3 },
        "host.aliases": {
          value: {
            items: [
              { raw_text: "endpoint-short" },
              { alias_text: "endpoint-alias" },
              { display_text: "endpoint-display" },
            ],
          },
        },
        "host.aad_device_id": { value: "device-1" },
      }),
      "host",
    );
    expect(host).toMatchObject({
      entityType: "host",
      recordId: "host-1",
      label: "endpoint-01",
      secondaryText: "endpoint-01",
      state: "active",
      aliasTexts: ["endpoint-short", "endpoint-alias", "endpoint-display"],
      linkedEventCount: 3,
      identifiers: [
        {
          key: "host.aad_device_id",
          label: "AAD Device ID",
          value: "device-1",
        },
        { key: "host.hostname", label: "Hostname", value: "endpoint-01" },
      ],
    });

    const identity = entityRowFromApi(
      entityRow("identity-1", {
        "identity.display_name": { value: "" },
        "identity.email": { value: "analyst@example.test" },
        "identity.identity_state": { value: "enabled" },
      }),
      "identity",
    );
    expect(identity.label).toBe("analyst@example.test");
    expect(identity.secondaryText).toBe("analyst@example.test");
  });

  it("builds merge plans without changing case-insensitive duplicate semantics", () => {
    const survivor = entityRowFromApi(
      entityRow("host-survivor", {
        "host.display_name": { value: "Endpoint survivor" },
        "host.hostname": { value: "ENDPOINT-01" },
        "host.linked_event_count": { value: 2 },
        "host.aliases": {
          value: { items: [{ raw_text: "primary" }, { raw_text: "shared" }] },
        },
      }),
      "host",
    );
    const loser = entityRowFromApi(
      entityRow("host-loser", {
        "host.display_name": { value: "Endpoint loser" },
        "host.hostname": { value: "endpoint-01" },
        "host.fqdn": { value: "endpoint-01.example.test" },
        "host.linked_event_count": { value: 1 },
        "host.aliases": {
          value: { items: [{ raw_text: "Shared" }, { raw_text: "secondary" }] },
        },
      }),
      "host",
    );

    expect(buildMergePlan(survivor, loser)).toEqual({
      identifierLines: [
        { label: "FQDN", outcome: "Promote endpoint-01.example.test" },
        { label: "Hostname", outcome: "Duplicate no-op endpoint-01" },
      ],
      aliasesToCopy: ["secondary"],
      duplicateAliases: ["Shared"],
      provenanceOnlySummary: "Not exposed on this surface.",
      dependencySummary:
        "Linked events visible on surface: survivor=2, loser=1.",
    });
  });

  it("keeps entity grouping and column widths contract-key based", () => {
    const hosts = requireViewContract(hostsViewSchemaId);
    const identities = requireViewContract(identitiesViewSchemaId);
    const host = entityRowFromApi(
      entityRow("host-1", {
        "host.display_name": { value: "Endpoint" },
        "host.host_state": { value: "" },
      }),
      "host",
    );

    expect(entityGroupLabel(host, "host.host_state")).toBe("Unassigned");
    const displayNameField = hosts.fieldMap["host.display_name"];
    const aliasesField = identities.fieldMap["identity.aliases"];
    expect(displayNameField).toBeDefined();
    expect(aliasesField).toBeDefined();
    if (displayNameField === undefined || aliasesField === undefined) {
      throw new Error("Expected entity contract fields to be present.");
    }
    expect(entityContractColumnWidth(displayNameField)).toBe(240);
    expect(entityContractColumnWidth(aliasesField)).toBe(320);
  });
});
