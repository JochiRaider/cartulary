import { describe, expect, it } from "vitest";
import { buildInspectorMentions } from "./timeline/models/workbookMentionChips";

describe("FE-P5 mention chip state model", () => {
  it("preserves closed mention chip states by stable identifiers and field keys", () => {
    const row = {
      recordId: "record-1",
      collectionValues: {
        hostRefs: [
          {
            itemRef: "entity_mention:11111111-1111-4111-8111-111111111111",
            entityType: "host" as const,
            itemKind: "unresolved_mention",
            displayText: "WS-023?",
            rawText: " WS-023? ",
            resolvedRecordId: null,
            mentionRowVersion: 1,
            resolutionMethod: null,
            autoResolved: false,
            provenance: "source_text",
            confidence: null,
            matchedAliasText: null,
          },
          {
            itemRef: "entity_mention:22222222-2222-4222-8222-222222222222",
            entityType: "host" as const,
            itemKind: "resolved_ref",
            displayText: "Server 02",
            rawText: "server-02",
            resolvedRecordId: "host-2",
            mentionRowVersion: 2,
            resolutionMethod: "legacy_import",
            autoResolved: false,
            provenance: "curated",
            confidence: 87,
            matchedAliasText: null,
          },
          {
            itemRef: "entity_mention:33333333-3333-4333-8333-333333333333",
            entityType: "host" as const,
            itemKind: "resolved_ref",
            displayText: "VPN Gateway",
            rawText: " vpn   gateway ",
            resolvedRecordId: "host-3",
            mentionRowVersion: 3,
            resolutionMethod: "auto_match",
            autoResolved: true,
            provenance: "auto_match",
            confidence: 100,
            matchedAliasText: "VPN Gateway",
          },
        ],
        identityRefs: [
          {
            itemRef: "entity_mention:44444444-4444-4444-8444-444444444444",
            entityType: "identity" as const,
            itemKind: "resolved_ref",
            displayText: "Alex Analyst",
            rawText: " alex.analyst@example.test ",
            resolvedRecordId: "identity-4",
            mentionRowVersion: 4,
            resolutionMethod: "explicit_resolve_route",
            autoResolved: false,
            provenance: "manual",
            confidence: null,
            matchedAliasText: null,
          },
        ],
      },
    };

    const mentions = buildInspectorMentions(row, [
      {
        rowRecordId: "record-1",
        fieldKey: "timeline.host_refs",
        entityType: "host",
        itemRef: "entity_mention:55555555-5555-4555-8555-555555555555",
        rawText: "old-host",
        resolvedRecordId: "host-5",
        mentionRowVersion: 5,
        resolutionMethod: "explicit_resolve_route",
        autoResolved: false,
        displayText: "Old Host",
        provenance: "manual",
        confidence: 42,
        matchedAliasText: "Old Host Alias",
      },
    ]);

    expect(
      mentions.map((mention) => [
        mention.itemRef,
        mention.status,
        mention.chipState,
        mention.anchor,
        mention.sourceKind,
        mention.isActiveRelationshipValue,
        mention.priorTargetEntityRecordId,
      ]),
    ).toEqual([
      [
        "entity_mention:11111111-1111-4111-8111-111111111111",
        "unresolved",
        "unresolved",
        {
          recordId: "record-1",
          fieldKey: "timeline.host_refs",
          itemRef: "entity_mention:11111111-1111-4111-8111-111111111111",
          entityMentionId: "11111111-1111-4111-8111-111111111111",
          targetEntityRecordId: null,
        },
        "entity_mention",
        true,
        null,
      ],
      [
        "entity_mention:22222222-2222-4222-8222-222222222222",
        "resolved",
        "resolved",
        {
          recordId: "record-1",
          fieldKey: "timeline.host_refs",
          itemRef: "entity_mention:22222222-2222-4222-8222-222222222222",
          entityMentionId: "22222222-2222-4222-8222-222222222222",
          targetEntityRecordId: "host-2",
        },
        "entity_mention",
        true,
        null,
      ],
      [
        "entity_mention:33333333-3333-4333-8333-333333333333",
        "resolved",
        "auto_resolved",
        {
          recordId: "record-1",
          fieldKey: "timeline.host_refs",
          itemRef: "entity_mention:33333333-3333-4333-8333-333333333333",
          entityMentionId: "33333333-3333-4333-8333-333333333333",
          targetEntityRecordId: "host-3",
        },
        "entity_mention",
        true,
        null,
      ],
      [
        "entity_mention:44444444-4444-4444-8444-444444444444",
        "resolved",
        "resolved",
        {
          recordId: "record-1",
          fieldKey: "timeline.identity_refs",
          itemRef: "entity_mention:44444444-4444-4444-8444-444444444444",
          entityMentionId: "44444444-4444-4444-8444-444444444444",
          targetEntityRecordId: "identity-4",
        },
        "entity_mention",
        true,
        null,
      ],
      [
        "entity_mention:55555555-5555-4555-8555-555555555555",
        "dismissed",
        "dismissed",
        {
          recordId: "record-1",
          fieldKey: "timeline.host_refs",
          itemRef: "entity_mention:55555555-5555-4555-8555-555555555555",
          entityMentionId: "55555555-5555-4555-8555-555555555555",
          targetEntityRecordId: null,
        },
        "entity_mention",
        false,
        "host-5",
      ],
    ]);

    expect(mentions[0]).toEqual(
      expect.objectContaining({
        rawText: " WS-023? ",
        resolvedRecordId: null,
        provenance: "source_text",
      }),
    );
    expect(mentions[2]).toEqual(
      expect.objectContaining({
        rawText: " vpn   gateway ",
        provenance: "auto_match",
        confidence: 100,
        matchedAliasText: "VPN Gateway",
      }),
    );
    expect(mentions[3]).toEqual(
      expect.objectContaining({
        rawText: " alex.analyst@example.test ",
        provenance: "manual",
        resolvedRecordId: "identity-4",
        resolutionMethod: "explicit_resolve_route",
      }),
    );
    expect(mentions[4]).toEqual(
      expect.objectContaining({
        displayText: "Old Host",
        provenance: "manual",
        confidence: 42,
        matchedAliasText: "Old Host Alias",
        resolvedRecordId: null,
      }),
    );
  });
});
