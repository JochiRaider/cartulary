import { describe, expect, it } from "vitest";
import {
  mapTimelineCollectionEditorIntent,
  mapTimelineScalarEditorIntent,
  mapTimelineWorkAreaInspectorIntent,
} from "./timelineKeyboardIntentModel";

describe("Timeline keyboard intent mapping", () => {
  it("maps scalar navigation without performing focus or save effects", () => {
    expect(
      mapTimelineScalarEditorIntent({
        event: { key: "Escape" },
        focusField: "activitySynopsisText",
        hasCommittedAnchor: true,
        inspectorCanClose: true,
        priorTimelineGridAnchor: true,
        surface: "inspector",
      }),
    ).toEqual({ kind: "restore_prior_grid_focus", preventDefault: true });
    expect(
      mapTimelineScalarEditorIntent({
        event: { key: "ArrowRight", shiftKey: true },
        focusField: "activitySynopsisText",
        hasCommittedAnchor: true,
        inspectorCanClose: false,
        priorTimelineGridAnchor: false,
        surface: "grid",
      }),
    ).toEqual({ kind: "none", preventDefault: true });
    expect(
      mapTimelineScalarEditorIntent({
        event: { key: "Tab" },
        focusField: "activitySynopsisText",
        hasCommittedAnchor: true,
        inspectorCanClose: false,
        priorTimelineGridAnchor: false,
        surface: "grid",
      }),
    ).toMatchObject({
      kind: "save",
      navigateAfterSave: null,
      preserveInputFocus: false,
    });
    expect(
      mapTimelineScalarEditorIntent({
        event: { key: "Enter" },
        focusField: "activitySynopsisText",
        hasCommittedAnchor: false,
        inspectorCanClose: false,
        priorTimelineGridAnchor: false,
        surface: "grid",
      }),
    ).toMatchObject({
      kind: "save",
      preserveInputFocus: true,
      recordBlankRowTiming: true,
    });
  });

  it("maps collection save, navigation, and close intent", () => {
    expect(
      mapTimelineCollectionEditorIntent({
        event: { key: "Enter" },
        hasCommittedAnchor: true,
        inspectorCanClose: false,
      }),
    ).toMatchObject({
      kind: "save",
      navigateAfterSave: { key: "Enter", shiftKey: false },
    });
    expect(
      mapTimelineCollectionEditorIntent({
        event: { key: "ArrowDown" },
        hasCommittedAnchor: true,
        inspectorCanClose: false,
      }),
    ).toEqual({
      kind: "navigate",
      navigation: { key: "ArrowDown", shiftKey: false },
      preventDefault: true,
    });
    expect(
      mapTimelineCollectionEditorIntent({
        event: { key: "Escape" },
        hasCommittedAnchor: true,
        inspectorCanClose: true,
      }),
    ).toEqual({ kind: "close_inspector", preventDefault: true });
  });

  it("maps work-area shortcuts to semantic panel and mention intent", () => {
    const row = {
      collectionValues: {
        hostRefs: [
          {
            itemKind: "unresolved_mention",
            itemRef: "mention-1",
          },
        ],
        identityRefs: [],
      },
      recordId: "record-1",
      rowVersion: 3,
    };
    expect(
      mapTimelineWorkAreaInspectorIntent({
        event: { altKey: true, key: "h" },
        fieldKey: "timeline.activity_synopsis_text",
        row,
      }),
    ).toEqual({ kind: "open_panel", panelId: "history", row });
    expect(
      mapTimelineWorkAreaInspectorIntent({
        event: { ctrlKey: true, key: "k" },
        fieldKey: "timeline.host_refs",
        row,
      }),
    ).toEqual({ itemRef: "mention-1", kind: "quick_link", row });
    expect(
      mapTimelineWorkAreaInspectorIntent({
        event: { ctrlKey: true, key: "k" },
        fieldKey: "timeline.activity_synopsis_text",
        row,
      }),
    ).toEqual({ kind: "none" });
  });
});
