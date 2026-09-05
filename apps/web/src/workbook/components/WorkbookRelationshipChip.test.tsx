import { relationshipChipTestId } from "@cartulary/ui-contracts";
import { fireEvent, render, screen, within } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import {
  type CollectionItem,
  timelineRelationshipChipPresentation,
} from "../timeline/models/workbookMentionChips";
import { WorkbookRelationshipChip } from "./WorkbookRelationshipChip";

it("WorkbookRelationshipChip preserves state details selectors and optional selection", () => {
  const inspect = vi.fn();
  const parentKeyDown = vi.fn();
  const item: CollectionItem = {
    itemRef: "entity_mention:source-mention",
    entityType: "host",
    itemKind: "resolved_ref",
    rawText: "Raw alias Ω",
    displayText: "Canonical host",
    resolvedRecordId: "host-target",
    mentionRowVersion: 1,
    resolutionMethod: "auto_match",
    autoResolved: true,
    provenance: "auto_match",
    confidence: 100,
    matchedAliasText: "Raw alias Ω",
  };
  const automatic = timelineRelationshipChipPresentation({
    entityIndex: { "host-target": { label: "Canonical host" } },
    item,
    sourceRecordId: "source-row",
  });
  const manual = timelineRelationshipChipPresentation({
    entityIndex: {},
    item: {
      ...item,
      itemRef: "mention-manual",
      resolutionMethod: "explicit_resolve_route",
      autoResolved: false,
      provenance: "manual",
      confidence: null,
      matchedAliasText: null,
    },
    selected: true,
  });
  const unresolved = timelineRelationshipChipPresentation({
    entityIndex: {},
    item: {
      ...item,
      itemRef: "mention-unresolved",
      itemKind: "unresolved_mention",
      rawText: "  manual resolution é Ω  ",
      displayText: "unrelated display text",
      resolvedRecordId: null,
      resolutionMethod: null,
      autoResolved: false,
      provenance: null,
      confidence: null,
      matchedAliasText: null,
    },
  });
  expect(manual.state).toBe("resolved");
  expect(manual.resolution.method).toBe("manual");
  expect(automatic.source).toMatchObject({
    recordId: "source-row",
    fieldKey: "timeline.host_refs",
    kind: "entity_mention",
  });
  expect(unresolved.rawText).toBe("  manual resolution é Ω  ");
  render(
    <table>
      <tbody>
        <tr>
          <td onKeyDown={parentKeyDown}>
            <WorkbookRelationshipChip
              presentation={{ ...automatic, onSelect: inspect }}
            />
            <WorkbookRelationshipChip
              presentation={{ ...manual, onSelect: inspect }}
            />
            <WorkbookRelationshipChip presentation={unresolved} />
            <WorkbookRelationshipChip
              presentation={{
                ...unresolved,
                state: "dismissed",
                selectorIdentity: "dismissed",
              }}
            />
            <WorkbookRelationshipChip
              presentation={{
                ...manual,
                entityType: "identity",
                selectorIdentity: "identity",
              }}
            />
          </td>
        </tr>
      </tbody>
    </table>,
  );
  const autoChip = screen.getByRole("button", {
    name: "Auto-resolved host: Canonical host; matched Raw alias Ω",
  });
  expect(within(autoChip).getByText("auto").style.clipPath).toBe("");
  expect(
    screen
      .getByTestId(relationshipChipTestId("mention-unresolved"))
      .getAttribute("aria-label"),
  ).toBe("Unresolved host mention:   manual resolution é Ω  ");
  expect(
    screen
      .getByTestId(relationshipChipTestId("dismissed"))
      .getAttribute("aria-label"),
  ).toBe("Dismissed mention:   manual resolution é Ω  ");
  expect(
    screen.getByLabelText("Resolved identity: Canonical host"),
  ).toBeTruthy();
  const manualChip = screen.getByRole("button", {
    name: "Resolved host: Canonical host",
  });
  expect(manualChip.getAttribute("aria-pressed")).toBe("true");
  expect(manualChip.style.boxShadow).toContain("var(--ct-colors-accent)");
  expect(within(manualChip).queryByText("M")).toBeNull();
  expect(
    within(screen.getByTestId(relationshipChipTestId("dismissed"))).getByText(
      "dismissed",
    ),
  ).toBeTruthy();
  fireEvent.click(autoChip);
  expect(inspect).toHaveBeenCalledTimes(1);
  for (const key of ["Enter", " ", "F2"]) fireEvent.keyDown(autoChip, { key });
  expect(parentKeyDown).not.toHaveBeenCalled();
  expect(inspect).toHaveBeenCalledTimes(1);
});
