import { describe, expect, it } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import {
  buildWorkbookPresenceInput,
  buildWorkbookPresenceUpdateMessage,
} from "./workbookCollaborationMessages";

describe("workbook collaboration presence messages", () => {
  it("builds default timeline presence payloads for WebSocket session establishment", () => {
    expect(buildWorkbookPresenceInput()).toEqual({
      sheet_ref: { kind: "view_schema", id: timelineViewSchemaId },
      mode: "viewing",
    });
  });

  it("includes field_key only for editing presence payloads", () => {
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

  it("builds presence_update messages without changing event names", () => {
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

  it("transmits extension workspace presence without record or field anchors", () => {
    const sheetRef = {
      kind: "extension_workspace" as const,
      extension_profile_id: "network_flow_activity",
      workspace_key: "network_analysis",
    };
    expect(
      buildWorkbookPresenceInput(
        {
          fieldKey: "network_flow.src_ip",
          mode: "editing",
          recordId: "network-flow-row-1",
        },
        sheetRef,
      ),
    ).toEqual({
      sheet_ref: sheetRef,
      mode: "editing",
    });
  });
});
