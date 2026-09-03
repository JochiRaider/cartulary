import { describe, expect, it } from "vitest";
import { normalizeWorkbookStartupSelection } from "../models/workbookStartup";
import {
  hostsViewSchemaId,
  timelineViewSchemaId,
} from "../models/workbookSurfaceRegistry";
import {
  beginWorkbookStartupAdmission,
  cancelWorkbookStartupAdmission,
  initialWorkbookStartupAdmissionMachine,
  planAcceptedWorkbookStartup,
  workbookStartupAdmissionIsCurrent,
} from "./workbookStartupAdmissionMachine";

const availabilityTag = { epochId: "a".repeat(64), generation: 4n };
const canonicalQuery = {
  sheetRefId: hostsViewSchemaId,
  sheetRefKind: "view_schema",
};

function startupSelection(overrides: Readonly<Record<string, unknown>> = {}) {
  const selection = normalizeWorkbookStartupSelection({
    cleared_pointers: [],
    default_sheet_ref: null,
    home_sheet_ref: null,
    selected_saved_view: null,
    selected_sheet_ref: { kind: "view_schema", id: hostsViewSchemaId },
    selected_view_schema_id: hostsViewSchemaId,
    source: "explicit",
    ...overrides,
  });
  if (selection === null) throw new Error("invalid startup fixture");
  return { availability: { workspaces: [] }, selection };
}

describe("Workbook startup admission machine", () => {
  it("keys admission by incident, canonical query, generation, selection, and availability", () => {
    const first = beginWorkbookStartupAdmission(
      initialWorkbookStartupAdmissionMachine(),
      {
        availabilityTag,
        incidentId: "incident-1",
        query: canonicalQuery,
        selectionVersion: 7,
      },
    );
    expect(
      workbookStartupAdmissionIsCurrent(first.machine, first.admission, {
        incidentId: "incident-1",
        query: canonicalQuery,
        selectionVersion: 7,
      }),
    ).toBe(true);
    for (const current of [
      {
        incidentId: "incident-2",
        query: canonicalQuery,
        selectionVersion: 7,
      },
      {
        incidentId: "incident-1",
        query: { sheetRefId: "saved-1", sheetRefKind: "saved_view" },
        selectionVersion: 7,
      },
      {
        incidentId: "incident-1",
        query: canonicalQuery,
        selectionVersion: 8,
      },
    ]) {
      expect(
        workbookStartupAdmissionIsCurrent(
          first.machine,
          first.admission,
          current,
        ),
      ).toBe(false);
    }
    const second = beginWorkbookStartupAdmission(first.machine, {
      availabilityTag: { ...availabilityTag, generation: 5n },
      incidentId: "incident-1",
      query: canonicalQuery,
      selectionVersion: 7,
    });
    expect(
      workbookStartupAdmissionIsCurrent(second.machine, first.admission, {
        incidentId: "incident-1",
        query: canonicalQuery,
        selectionVersion: 7,
      }),
    ).toBe(false);
    expect(
      cancelWorkbookStartupAdmission(second.machine, second.admission).active,
    ).toBeNull();
  });

  it("plans base, saved-view, extension, and availability outcomes exactly", () => {
    expect(
      planAcceptedWorkbookStartup({
        availabilityAccepted: true,
        extensionRenderable: true,
        startup: startupSelection(),
      }),
    ).toMatchObject({
      kind: "apply",
      identity: {
        sheetRef: { kind: "view_schema", id: hostsViewSchemaId },
        viewSchemaId: hostsViewSchemaId,
      },
      savedView: null,
    });
    const savedView = {
      display_name: "Hosts",
      layout_json: {},
      owner_user_id: "user-1",
      query_json: {},
      saved_view_id: "saved-1",
      saved_view_version: 1,
      scope: "private",
      view_schema_id: hostsViewSchemaId,
    };
    expect(
      planAcceptedWorkbookStartup({
        availabilityAccepted: true,
        extensionRenderable: true,
        startup: startupSelection({
          selected_saved_view: savedView,
          selected_sheet_ref: { kind: "saved_view", id: "saved-1" },
        }),
      }),
    ).toMatchObject({ kind: "apply", savedView });
    expect(
      planAcceptedWorkbookStartup({
        availabilityAccepted: false,
        extensionRenderable: true,
        startup: startupSelection(),
      }),
    ).toEqual({ kind: "fallback", reason: "availability_rejected" });
    expect(
      planAcceptedWorkbookStartup({
        availabilityAccepted: true,
        extensionRenderable: false,
        startup: startupSelection({
          selected_sheet_ref: {
            kind: "extension_workspace",
            extension_profile_id: "network_flow_activity",
            workspace_key: "network_analysis",
          },
          selected_view_schema_id: null,
        }),
      }),
    ).toEqual({
      kind: "fallback",
      reason: "selected_extension_not_renderable",
    });
  });

  it("discards inconsistent selected identities without changing state", () => {
    const validViewSchemaStartup = startupSelection();
    expect(
      planAcceptedWorkbookStartup({
        availabilityAccepted: true,
        extensionRenderable: true,
        startup: {
          ...validViewSchemaStartup,
          selection: {
            ...validViewSchemaStartup.selection,
            selectedSavedView: null,
            selectedSheetRef: { kind: "saved_view", id: "missing" },
          },
        },
      }),
    ).toEqual({ kind: "discard" });
    expect(
      planAcceptedWorkbookStartup({
        availabilityAccepted: true,
        extensionRenderable: true,
        startup: {
          ...validViewSchemaStartup,
          selection: {
            ...validViewSchemaStartup.selection,
            selectedSheetRef: {
              kind: "view_schema",
              id: timelineViewSchemaId,
            },
            selectedViewSchemaId: "future.view.v1",
          },
        },
      }),
    ).toEqual({ kind: "discard" });
  });
});
