import { relationshipChipTestId } from "@cartulary/ui-contracts";
import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { WorkbookRelationshipChip } from "./WorkbookRelationshipChip";

it("WorkbookRelationshipChip preserves state details selectors and optional selection", () => {
  const onSelect = vi.fn();
  render(
    <div>
      <WorkbookRelationshipChip
        presentation={{
          label: "Unresolved host",
          selected: false,
          selectorIdentity: "mention-unresolved",
          state: "unresolved",
        }}
      />
      <WorkbookRelationshipChip
        presentation={{
          accessibleDetail: "manual resolution",
          label: "Resolved host",
          onSelect,
          selected: true,
          selectorIdentity: "mention-resolved",
          state: "resolved",
        }}
      />
      <WorkbookRelationshipChip
        presentation={{
          accessibleDetail: "matched host-alias",
          label: "Automatic host",
          selected: false,
          selectorIdentity: "mention-auto",
          state: "auto_resolved",
        }}
      />
      <WorkbookRelationshipChip
        presentation={{
          label: "Dismissed host",
          selected: false,
          selectorIdentity: "mention-dismissed",
          state: "dismissed",
        }}
      />
    </div>,
  );

  expect(
    screen
      .getByTestId(relationshipChipTestId("mention-unresolved"))
      .getAttribute("aria-label"),
  ).toBe("Unresolved Unresolved host");
  expect(
    screen
      .getByTestId(relationshipChipTestId("mention-auto"))
      .getAttribute("aria-label"),
  ).toBe("Auto-resolved Automatic host; matched host-alias");
  expect(
    screen
      .getByTestId(relationshipChipTestId("mention-dismissed"))
      .getAttribute("aria-label"),
  ).toBe("Dismissed Dismissed host");

  const selected = screen.getByRole("button", {
    name: "Resolved Resolved host; manual resolution",
  });
  fireEvent.click(selected);
  expect(onSelect).toHaveBeenCalledTimes(1);
  expect(selected.style.boxShadow).toContain("var(--ct-colors-accent)");
});
