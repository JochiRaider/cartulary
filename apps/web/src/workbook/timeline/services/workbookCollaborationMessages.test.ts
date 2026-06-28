import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  buildWorkbookPresenceInput,
  buildWorkbookPresenceUpdateMessage,
  buildWorkbookSocketSessionMessage,
} from "./workbookCollaborationMessages";

describe("workbook collaboration presence messages", () => {
  it("FE-U-P7-01 builds default timeline presence payloads for WebSocket session establishment", () => {
    expect(buildWorkbookPresenceInput()).toEqual({
      sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
      mode: "viewing",
    });
  });

  it("FE-U-P7-01 includes field_key only for editing presence payloads", () => {
    const sheetRef = { kind: "saved_view" as const, id: "saved-view-1" };

    expect(
      buildWorkbookPresenceInput(
        {
          fieldKey: "timeline.activity_synopsis_text",
          mode: "viewing",
          recordId: "record-1",
        },
        sheetRef,
      ),
    ).toEqual({
      sheet_ref: sheetRef,
      mode: "viewing",
      record_id: "record-1",
    });

    expect(
      buildWorkbookPresenceInput(
        {
          fieldKey: "timeline.activity_synopsis_text",
          mode: "editing",
          recordId: "record-1",
        },
        sheetRef,
      ),
    ).toEqual({
      sheet_ref: sheetRef,
      mode: "editing",
      record_id: "record-1",
      field_key: "timeline.activity_synopsis_text",
    });
  });

  it("FE-U-P7-01 builds presence_update messages without changing event names", () => {
    expect(
      buildWorkbookPresenceUpdateMessage(
        { fieldKey: null, mode: "viewing", recordId: "record-1" },
        { kind: "view_schema", id: timelineViewSchemaId },
      ),
    ).toEqual({
      type: "presence_update",
      payload: {
        presence: {
          sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
          mode: "viewing",
          record_id: "record-1",
        },
      },
    });
  });

  it("FE-U-P7-01 builds hello and resume session messages from socket resume state", () => {
    const sheetRef = { kind: "view_schema" as const, id: timelineViewSchemaId };
    const presence = {
      fieldKey: "timeline.activity_synopsis_text",
      mode: "editing" as const,
      recordId: "record-1",
    };

    expect(
      buildWorkbookSocketSessionMessage({
        clientInstanceId: "client-1",
        lastSeenStreamSeq: 7,
        presence,
        resumeToken: null,
        sheetRef,
      }),
    ).toEqual({
      type: "hello",
      payload: {
        client_instance_id: "client-1",
        presence: {
          sheet_ref: sheetRef,
          mode: "editing",
          record_id: "record-1",
          field_key: "timeline.activity_synopsis_text",
        },
      },
    });

    expect(
      buildWorkbookSocketSessionMessage({
        clientInstanceId: "client-1",
        lastSeenStreamSeq: 7,
        presence,
        resumeToken: "resume-1",
        sheetRef,
      }),
    ).toEqual({
      type: "resume",
      payload: {
        client_instance_id: "client-1",
        resume_token: "resume-1",
        last_seen_stream_seq: 7,
        presence: {
          sheet_ref: sheetRef,
          mode: "editing",
          record_id: "record-1",
          field_key: "timeline.activity_synopsis_text",
        },
      },
    });
  });
});
