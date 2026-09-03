import { requireViewContract } from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  fullWorkbookViewRow,
  workbookCollectionValue,
} from "../../../testing/timelineWorkbookTestSupport";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  planTimelineMentionEntityCreation,
  planTimelineMentionResolution,
  timelineMentionSubject,
} from "./timelineMentionActionPlan";
import { normalizeTimelineFullRow, rowFromApi } from "./timelineRowModel";
import type { InspectorMention } from "./workbookMentionChips";

const timeline = requireViewContract(timelineViewSchemaId);
const recordId = "record-1";
const targetId = "host-1";
const mentionId = "mention-1";
const itemRef = `entity_mention:${mentionId}`;
const context = {
  authorized: true,
  capabilityAvailable: true,
  surfaceKey: "view_schema:timeline",
};
const mention: InspectorMention = {
  anchor: {
    entityMentionId: mentionId,
    fieldKey: "timeline.host_refs",
    itemRef,
    recordId,
    targetEntityRecordId: null,
  },
  autoResolved: false,
  chipState: "unresolved",
  confidence: null,
  displayText: "server.example",
  entityType: "host",
  fieldKey: "timeline.host_refs",
  isActiveRelationshipValue: true,
  itemRef,
  matchedAliasText: null,
  mentionRowVersion: 2,
  priorTargetEntityRecordId: null,
  provenance: null,
  rawText: "server.example",
  resolutionMethod: null,
  resolvedRecordId: null,
  rowRecordId: recordId,
  sourceKind: "entity_mention",
  status: "unresolved",
};
const subject = timelineMentionSubject(mention, context.surfaceKey);

describe("Timeline Mention action plan", () => {
  it("re-reads the exact mention version and validates an existing target", () => {
    const plan = planTimelineMentionResolution({
      action: "resolve_item",
      context,
      knownEntityTypes: new Map<string, "host" | "identity">([
        [targetId, "host"],
      ]),
      mention,
      resolvedRecordId: targetId,
      rows: [mentionRow(7)],
      subject,
    });
    expect(plan).toMatchObject({
      kind: "dispatch",
      request: {
        baseMentionRowVersion: 7,
        expectedSourceRecordId: recordId,
        mentionId,
        resolvedRecordId: targetId,
      },
    });
  });

  it("rejects invalid targets, access loss, surface changes, and removed mentions", () => {
    const base = {
      action: "resolve_item" as const,
      context,
      knownEntityTypes: new Map<string, "host" | "identity">([
        [targetId, "host"],
      ]),
      mention,
      resolvedRecordId: targetId,
      rows: [mentionRow(7)],
      subject,
    };
    expect(
      planTimelineMentionResolution({
        ...base,
        resolvedRecordId: "missing",
      }),
    ).toEqual({ kind: "reject", reason: "target_missing" });
    expect(
      planTimelineMentionResolution({
        ...base,
        knownEntityTypes: new Map([[targetId, "identity"]]),
      }),
    ).toEqual({ kind: "reject", reason: "target_missing" });
    expect(
      planTimelineMentionResolution({
        ...base,
        context: { ...context, authorized: false },
      }),
    ).toEqual({ kind: "reject", reason: "authorization_lost" });
    expect(
      planTimelineMentionResolution({
        ...base,
        context: { ...context, surfaceKey: "saved_view:other" },
      }),
    ).toEqual({ kind: "reject", reason: "surface_changed" });
    expect(
      planTimelineMentionResolution({ ...base, rows: [mentionRow(null)] }),
    ).toEqual({ kind: "reject", reason: "mention_missing" });
  });

  it("admits entity creation only for a current non-empty mention", () => {
    expect(
      planTimelineMentionEntityCreation({
        context,
        mention,
        rows: [mentionRow(4)],
        subject,
      }),
    ).toMatchObject({ kind: "dispatch" });
    expect(
      planTimelineMentionEntityCreation({
        context,
        mention: { ...mention, rawText: " " },
        rows: [mentionRow(4)],
        subject,
      }),
    ).toEqual({ kind: "reject", reason: "mention_missing" });
  });
});

function mentionRow(mentionRowVersion: number | null) {
  return rowFromApi(
    normalizeTimelineFullRow(
      fullWorkbookViewRow(timeline, recordId, 3, {
        "timeline.host_refs": workbookCollectionValue(
          false,
          mentionRowVersion === null
            ? []
            : [
                {
                  auto_resolved: false,
                  confidence: null,
                  display_text: "server.example",
                  entity_type: "host",
                  item_kind: "unresolved_mention",
                  item_ref: itemRef,
                  matched_alias_text: null,
                  mention_row_version: mentionRowVersion,
                  provenance: null,
                  raw_text: "server.example",
                  resolution_method: null,
                  resolved_record_id: null,
                },
              ],
        ),
      }),
      "mention plan fixture",
    ),
  );
}
