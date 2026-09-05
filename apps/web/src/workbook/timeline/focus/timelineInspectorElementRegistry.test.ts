import { afterEach, describe, expect, it } from "vitest";
import type { WorkbookInspectorSubject } from "../../inspector/workbookInspectorSubject";
import { createTimelineInspectorElementRegistry } from "./timelineInspectorElementRegistry";

const viewSchemaId = "workbook.timeline";
const subject = (recordId: string, rowVersion: number) =>
  ({
    kind: "live",
    label: "Timeline row",
    recordId,
    rowVersion,
    surfaceLabel: "Timeline",
    viewSchemaId,
  }) satisfies WorkbookInspectorSubject;

function scope(activeSubject: WorkbookInspectorSubject | null) {
  return {
    invalidationGeneration: 1,
    lifecycleKey: "incident-1:timeline",
    subject: activeSubject,
  };
}

afterEach(() => {
  document.body.replaceChildren();
});

describe("timeline inspector element registry", () => {
  it("focuses only panels and mentions for the captured canonical subject", () => {
    const activeSubject = subject("record-1", 3);
    const registry = createTimelineInspectorElementRegistry(
      scope(activeSubject),
    );
    const panel = document.createElement("section");
    panel.tabIndex = -1;
    const nested = document.createElement("input");
    panel.append(nested);
    const mention = document.createElement("button");
    document.body.append(panel, mention);
    registry.registerPanel("history", panel);
    registry.registerMention("record-1", "mention-1", mention);
    const identity = {
      recordId: "record-1",
      rowVersion: 3,
      viewSchemaId,
    };

    expect(registry.focusPanel(identity, "history")).toBe(true);
    expect(document.activeElement).toBe(panel);
    expect(registry.focusMention(identity, "record-1", "mention-1")).toBe(true);
    expect(document.activeElement).toBe(mention);
    nested.focus();
    expect(registry.containsActiveElement()).toBe(true);

    expect(registry.focusPanel({ ...identity, rowVersion: 4 }, "history")).toBe(
      false,
    );
    expect(registry.focusMention(identity, "record-2", "mention-1")).toBe(
      false,
    );
  });

  it("targets complete collection members and restores only live semantic triggers", () => {
    const registry = createTimelineInspectorElementRegistry(
      scope(subject("record-1", 3)),
    );
    const target = document.createElement("span");
    target.tabIndex = -1;
    const trigger = document.createElement("button");
    document.body.append(target, trigger);
    const identity = { recordId: "record-1", rowVersion: 3, viewSchemaId };
    registry.registerCollectionItem(
      "record-1",
      "timeline.tags",
      "tag-2",
      target,
    );
    expect(
      registry.focusCollectionItem(identity, "timeline.tags", "tag-2"),
    ).toBe(true);
    expect(document.activeElement).toBe(target);
    expect(
      registry.focusCollectionItem(identity, "timeline.host_refs", "tag-2"),
    ).toBe(false);
    expect(
      registry.focusCollectionItem(
        { ...identity, rowVersion: 4 },
        "timeline.tags",
        "tag-2",
      ),
    ).toBe(false);
    registry.registerCollectionTrigger(
      "record-1",
      "timeline.tags",
      null,
      trigger,
    );
    registry.rememberCollectionReturnFocus("record-1", "timeline.tags", null);
    registry.updateScope(scope(null));
    expect(registry.restoreCollectionReturnFocus()).toBe(true);
    expect(document.activeElement).toBe(trigger);
    registry.rememberCollectionReturnFocus("record-1", "timeline.tags", null);
    registry.updateScope(scope(subject("record-2", 1)));
    expect(registry.restoreCollectionReturnFocus()).toBe(false);
    registry.rememberCollectionReturnFocus("record-1", "timeline.tags", null);
    trigger.remove();
    expect(registry.restoreCollectionReturnFocus()).toBe(false);
  });

  it("clears registrations on lifecycle, generation, subject, and version changes", () => {
    const registry = createTimelineInspectorElementRegistry(
      scope(subject("record-1", 3)),
    );
    const panel = document.createElement("section");
    panel.tabIndex = -1;
    document.body.append(panel);
    registry.registerPanel("evidence", panel);

    registry.updateScope({
      ...scope(subject("record-1", 3)),
      invalidationGeneration: 2,
    });
    expect(
      registry.focusPanel(
        { recordId: "record-1", rowVersion: 3, viewSchemaId },
        "evidence",
      ),
    ).toBe(false);

    registry.registerPanel("evidence", panel);
    registry.updateScope(scope(subject("record-1", 4)));
    expect(
      registry.focusPanel(
        { recordId: "record-1", rowVersion: 4, viewSchemaId },
        "evidence",
      ),
    ).toBe(false);

    registry.registerPanel("evidence", panel);
    registry.updateScope({
      ...scope(subject("record-2", 1)),
      lifecycleKey: "incident-2:timeline",
    });
    expect(
      registry.focusPanel(
        { recordId: "record-2", rowVersion: 1, viewSchemaId },
        "evidence",
      ),
    ).toBe(false);
  });

  it("rejects disconnected, hidden, and disabled elements", () => {
    const activeSubject = subject("record-1", 3);
    const registry = createTimelineInspectorElementRegistry(
      scope(activeSubject),
    );
    const identity = {
      recordId: "record-1",
      rowVersion: 3,
      viewSchemaId,
    };
    const disconnectedPanel = document.createElement("section");
    registry.registerPanel("history", disconnectedPanel);
    expect(registry.focusPanel(identity, "history")).toBe(false);

    const hiddenPanel = document.createElement("section");
    hiddenPanel.hidden = true;
    document.body.append(hiddenPanel);
    registry.registerPanel("history", hiddenPanel);
    expect(registry.focusPanel(identity, "history")).toBe(false);

    const disabledMention = document.createElement("button");
    disabledMention.disabled = true;
    document.body.append(disabledMention);
    registry.registerMention("record-1", "mention-1", disabledMention);
    expect(registry.focusMention(identity, "record-1", "mention-1")).toBe(
      false,
    );
  });
});
