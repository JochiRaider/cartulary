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
        "host.reusable_identifiers": {
          value: {
            items: [
              {
                item_ref: "entity_preserved_identifier:host-secondary",
                item_kind: "reusable_identifier",
                identifier_class: "fqdn",
                raw_value: "endpoint-old.example.test",
                normalized_value: "endpoint-old.example.test",
                display_text: "endpoint-old.example.test",
              },
            ],
          },
        },
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
          identifierClass: "aad_device_id",
          value: "device-1",
        },
        {
          key: "host.hostname",
          label: "Hostname",
          identifierClass: "hostname",
          value: "endpoint-01",
        },
      ],
      reusableIdentifiers: [
        {
          itemRef: "entity_preserved_identifier:host-secondary",
          itemKind: "reusable_identifier",
          identifierClass: "fqdn",
          label: "FQDN",
          rawValue: "endpoint-old.example.test",
          normalizedValue: "endpoint-old.example.test",
          displayText: "endpoint-old.example.test",
        },
      ],
    });

    const identity = entityRowFromApi(
      entityRow("identity-1", {
        "identity.display_name": { value: "" },
        "identity.email": { value: "analyst@example.test" },
        "identity.reusable_identifiers": {
          value: {
            items: [
              {
                item_ref: "entity_preserved_identifier:identity-secondary",
                item_kind: "reusable_identifier",
                identifier_class: "sid",
                raw_value: "S-1-5-21-123",
                normalized_value: "s-1-5-21-123",
                display_text: "S-1-5-21-123",
              },
            ],
          },
        },
        "identity.identity_state": { value: "enabled" },
      }),
      "identity",
    );
    expect(identity.label).toBe("analyst@example.test");
    expect(identity.secondaryText).toBe("analyst@example.test");
    expect(identity.reusableIdentifiers).toEqual([
      {
        itemRef: "entity_preserved_identifier:identity-secondary",
        itemKind: "reusable_identifier",
        identifierClass: "sid",
        label: "SID",
        rawValue: "S-1-5-21-123",
        normalizedValue: "s-1-5-21-123",
        displayText: "S-1-5-21-123",
      },
    ]);
  });

  it("builds merge plans without changing case-insensitive duplicate semantics", () => {
    const survivor = entityRowFromApi(
      entityRow("host-survivor", {
        "host.display_name": { value: "Endpoint survivor" },
        "host.fqdn": { value: "survivor.example.test" },
        "host.hostname": { value: "ENDPOINT-01" },
        "host.linked_event_count": { value: 2 },
        "host.reusable_identifiers": {
          value: {
            items: [
              {
                item_ref: "entity_preserved_identifier:survivor-reused",
                item_kind: "reusable_identifier",
                identifier_class: "aad_device_id",
                raw_value: "device-existing",
                normalized_value: "device-existing",
                display_text: "device-existing",
              },
            ],
          },
        },
        "host.aliases": {
          value: { items: [{ raw_text: "primary" }, { raw_text: "shared" }] },
        },
      }),
      "host",
    );
    const loser = entityRowFromApi(
      entityRow("host-loser", {
        "host.display_name": { value: "Endpoint loser" },
        "host.aad_device_id": { value: "device-existing" },
        "host.hostname": { value: "endpoint-01" },
        "host.fqdn": { value: "endpoint-01.example.test" },
        "host.linked_event_count": { value: 1 },
        "host.reusable_identifiers": {
          value: {
            items: [
              {
                item_ref: "entity_preserved_identifier:loser-reused",
                item_kind: "reusable_identifier",
                identifier_class: "fqdn",
                raw_value: "old-endpoint.example.test",
                normalized_value: "old-endpoint.example.test",
                display_text: "old-endpoint.example.test",
              },
            ],
          },
        },
        "host.aliases": {
          value: { items: [{ raw_text: "Shared" }, { raw_text: "secondary" }] },
        },
      }),
      "host",
    );

    const plan = buildMergePlan(survivor, loser);
    expect(plan.identifierLines).toEqual([
      { label: "AAD Device ID", outcome: "Promote device-existing" },
      {
        label: "FQDN",
        outcome: "Carry as reusable endpoint-01.example.test",
      },
      { label: "Hostname", outcome: "Duplicate no-op endpoint-01" },
      {
        label: "FQDN",
        outcome: "Carry as reusable old-endpoint.example.test",
      },
    ]);
    expect(plan.aliasesToCopy).toEqual(["secondary"]);
    expect(plan.duplicateAliases).toEqual(["Shared"]);
    expect(plan.provenanceOnlySummary).toBe(
      "Merge lineage and source provenance are retained server-side; no editable cell value is copied for them.",
    );
    expect(plan.dependencySummary).toBe(
      "Linked events visible on surface: survivor=2, loser=1.",
    );
  });

  it("promotes loser identifiers only when the survivor canonical field is empty", () => {
    const survivor = entityRowFromApi(
      entityRow("identity-survivor", {
        "identity.display_name": { value: "Identity survivor" },
        "identity.upn": { value: "survivor@example.test" },
      }),
      "identity",
    );
    const loser = entityRowFromApi(
      entityRow("identity-loser", {
        "identity.display_name": { value: "Identity loser" },
        "identity.email": { value: "loser@example.test" },
        "identity.upn": { value: "loser@example.test" },
      }),
      "identity",
    );

    expect(buildMergePlan(survivor, loser).identifierLines).toEqual([
      { label: "UPN", outcome: "Carry as reusable loser@example.test" },
      { label: "Email", outcome: "Promote loser@example.test" },
    ]);
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
