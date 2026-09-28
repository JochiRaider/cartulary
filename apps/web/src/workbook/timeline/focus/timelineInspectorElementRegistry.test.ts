import { afterEach, describe, expect, it } from "vitest";
import type { WorkbookRecordSubject } from "../../ports/WorkbookRecordSubject";
import {
  createTimelineInspectorElementRegistry,
  type TimelineInspectorElementRegistry,
} from "./timelineInspectorElementRegistry";

const viewSchemaId = "workbook.timeline";
const subject = (recordId: string, rowVersion: number) =>
  ({
    kind: "live",
    label: "Timeline row",
    recordId,
    rowVersion,
    surfaceLabel: "Timeline",
    viewSchemaId,
  }) satisfies WorkbookRecordSubject;

function scope(activeSubject: WorkbookRecordSubject | null) {
  return {
    reviewGeneration: 1,
    lifecycleKey: "incident-1:timeline",
    subject: activeSubject,
  };
}

function attachNavigator(
  registry: TimelineInspectorElementRegistry,
  defaults: Partial<Record<string, HTMLElement>> = {},
) {
  registry.registerDestinationNavigator((panelId, target) => {
    if (target === null) return "pending";
    const destination = target ?? defaults[panelId];
    if (!destination?.isConnected) return "pending";
    destination.focus({ preventScroll: true });
    return document.activeElement === destination ? "applied" : "pending";
  });
}

afterEach(() => {
  document.body.replaceChildren();
});

describe("timeline inspector element registry", () => {
  it("fulfills first-open Evidence list focus only when the matching element registers", () => {
    const registry = createTimelineInspectorElementRegistry(scope(null));
    attachNavigator(registry);
    const identity = { recordId: "record-1", rowVersion: 3, viewSchemaId };
    const list = document.createElement("ul");
    list.tabIndex = -1;
    document.body.append(list);
    expect(registry.focusEvidenceList(identity)).toBe(false);
    registry.updateScope(scope(subject("record-1", 3)));
    expect(document.activeElement).not.toBe(list);
    registry.registerEvidenceList("record-1", list);
    expect(document.activeElement).toBe(list);
    list.blur();
    expect(registry.focusEvidenceList(identity)).toBe(true);
    registry.cancelPendingFocus();
    list.blur();
    registry.focusEvidenceList({ ...identity, rowVersion: 4 });
    registry.registerEvidenceList("record-1", list);
    expect(document.activeElement).not.toBe(list);
  });

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
    attachNavigator(registry, { history: panel });
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
    attachNavigator(registry);
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
    attachNavigator(registry);
    registry.registerPanel("evidence", panel);

    registry.updateScope({
      ...scope(subject("record-1", 3)),
      reviewGeneration: 2,
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
    attachNavigator(registry);
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

  it("does not fulfill superseded, invalidated, closed, or unavailable focus after late registration", () => {
    const identity = { recordId: "record-1", rowVersion: 3, viewSchemaId };
    const registry = createTimelineInspectorElementRegistry(scope(null));
    const evidence = document.createElement("ul");
    evidence.tabIndex = -1;
    const relationships = document.createElement("section");
    relationships.tabIndex = -1;
    document.body.append(evidence, relationships);
    attachNavigator(registry);
    expect(registry.focusEvidenceList(identity)).toBe(false);
    expect(registry.focusPanel(identity, "relationships")).toBe(false);
    registry.updateScope(scope(subject("record-1", 3)));
    registry.registerEvidenceList("record-1", evidence);
    expect(document.activeElement).not.toBe(evidence);
    registry.registerPanel("relationships", relationships);
    expect(document.activeElement).toBe(relationships);

    relationships.blur();
    registry.focusEvidenceList(identity);
    registry.cancelPendingFocus();
    evidence.blur();
    registry.registerEvidenceList("record-1", evidence);
    expect(document.activeElement).not.toBe(evidence);

    registry.focusPanel(identity, "history");
    registry.updateScope(scope(subject("record-1", 4)));
    const history = document.createElement("section");
    history.tabIndex = -1;
    document.body.append(history);
    registry.registerPanel("history", history);
    expect(document.activeElement).not.toBe(history);

    const nextIdentity = { ...identity, rowVersion: 4 };
    registry.focusEvidenceList(nextIdentity);
    registry.updateScope(scope(null));
    registry.registerEvidenceList("record-1", evidence);
    expect(document.activeElement).not.toBe(evidence);

    registry.updateScope(scope(subject("record-1", 4)));
    registry.registerDestinationNavigator((panelId) =>
      panelId === "evidence" ? "unavailable" : "pending",
    );
    registry.focusEvidenceList(nextIdentity);
    registry.registerDestinationNavigator((_panelId, target) => {
      if (target instanceof HTMLElement) target.focus();
      return "applied";
    });
    registry.registerEvidenceList("record-1", evidence);
    expect(document.activeElement).not.toBe(evidence);

    registry.updateScope({
      ...scope(subject("record-1", 4)),
      authorityKey: "editor:session-1",
    });
    attachNavigator(registry);
    registry.registerPanel("relationships", null);
    relationships.blur();
    registry.focusPanel(nextIdentity, "relationships");
    registry.updateScope({
      ...scope(subject("record-1", 4)),
      authorityKey: "viewer:session-1",
    });
    registry.registerPanel("relationships", relationships);
    expect(document.activeElement).not.toBe(relationships);
  });
});
