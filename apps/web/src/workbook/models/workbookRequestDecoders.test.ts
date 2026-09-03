import {
  assessmentsViewSchemaId,
  notesViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  buildPatchRecordRequest,
  decodeCollectionActions,
  decodeCreateRecordLinkedNoteRequest,
  decodeCreateViewRowRequest,
  decodeRecordPatchChanges,
} from "./workbookRequestDecoders";

describe("Workbook request decoders", () => {
  it("constructs exact create, linked-note, collection, and patch requests", () => {
    const timeline = requireViewContract(timelineViewSchemaId);
    const create = {
      client_txn_id: "txn-create",
      "timeline.activity_synopsis_text": "Observed activity",
      "timeline.tags": {
        actions: [{ op: "add_tag", tag_name: "priority" }],
        kind: "collection_actions_v1",
      },
    };
    expect(decodeCreateViewRowRequest(timeline, create)).toEqual(create);

    const linkedNote = {
      client_txn_id: "txn-note",
      "note.body": "Linked context",
      "note.tags": {
        actions: [{ op: "add_tag", tag_name: "triage" }],
        kind: "collection_actions_v1",
      },
    };
    expect(decodeCreateRecordLinkedNoteRequest(linkedNote)).toEqual(linkedNote);
    expect(decodeCollectionActions(linkedNote["note.tags"])).toEqual(
      linkedNote["note.tags"],
    );

    expect(
      buildPatchRecordRequest({
        baseRowVersion: 4,
        changes: [
          { field_key: "timeline.activity_synopsis_text", value: "Updated" },
          {
            action_payload: {
              actions: [{ item_ref: "tag:old", op: "remove_tag" }],
              kind: "collection_actions_v1",
            },
            field_key: "timeline.tags",
          },
        ],
        clientTxnId: "txn-patch",
        viewSchemaId: timelineViewSchemaId,
      }),
    ).toEqual({
      base_row_version: 4,
      changes: [
        { field_key: "timeline.activity_synopsis_text", value: "Updated" },
        {
          action_payload: {
            actions: [{ item_ref: "tag:old", op: "remove_tag" }],
            kind: "collection_actions_v1",
          },
          field_key: "timeline.tags",
        },
      ],
      client_txn_id: "txn-patch",
      view_schema_id: timelineViewSchemaId,
    });
  });

  it("fails closed on malformed and future create or collection members", () => {
    const timeline = requireViewContract(timelineViewSchemaId);
    const notes = requireViewContract(notesViewSchemaId);
    const assessments = requireViewContract(assessmentsViewSchemaId);
    expect(
      decodeCreateViewRowRequest(timeline, {
        client_txn_id: "txn-create",
        "timeline.future_field": "not admitted",
      }),
    ).toBeNull();
    expect(
      decodeCreateViewRowRequest(timeline, {
        client_txn_id: "txn-create",
        "timeline.tags": {
          actions: [{ op: "future_collection_action", tag_name: "unsafe" }],
          kind: "collection_actions_v1",
        },
      }),
    ).toBeNull();
    expect(
      decodeCreateViewRowRequest(notes, {
        client_txn_id: "",
        "note.body": "not sent",
      }),
    ).toBeNull();
    expect(
      decodeCreateViewRowRequest(assessments, {
        client_txn_id: "txn-assessment",
        "assessment.assessed_at": "not-a-timestamp",
      }),
    ).toBeNull();
    expect(
      decodeCreateViewRowRequest(assessments, {
        client_txn_id: "txn-assessment",
        "assessment.assessment_state": "future_state",
      }),
    ).toBeNull();
    expect(
      decodeCreateViewRowRequest(assessments, {
        client_txn_id: "txn-assessment",
        "assessment.confidence_score": 1.5,
      }),
    ).toBeNull();
    expect(
      decodeCollectionActions({
        actions: [],
        kind: "collection_actions_v1",
      }),
    ).toBeNull();
    expect(
      decodeCollectionActions({
        actions: [{ op: "add_tag", tag_name: "" }],
        kind: "collection_actions_v1",
      }),
    ).toBeNull();
    expect(
      decodeCreateRecordLinkedNoteRequest({
        client_txn_id: "txn-note",
        "timeline.activity_synopsis_text": "wrong owner",
      }),
    ).toBeNull();
  });

  it("rejects empty, ambiguous, and malformed patch changes", () => {
    expect(decodeRecordPatchChanges([])).toBeNull();
    expect(
      decodeRecordPatchChanges([
        { field_key: "timeline.tags", value: "x", action_payload: {} },
      ]),
    ).toBeNull();
    expect(
      decodeRecordPatchChanges([
        { field_key: "", value: "x" },
        { field_key: "timeline.tags", unexpected: true, value: "x" },
      ]),
    ).toBeNull();
    expect(
      buildPatchRecordRequest({
        baseRowVersion: 0,
        changes: [{ field_key: "timeline.activity_synopsis_text", value: "x" }],
        clientTxnId: "txn-patch",
        viewSchemaId: timelineViewSchemaId,
      }),
    ).toBeNull();
  });
});
