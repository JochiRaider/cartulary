import {
  cleanup,
  fireEvent,
  render,
  screen,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it } from "vitest";
import { WorkbookIncidentIdentityDisclosure } from "./WorkbookIncidentIdentityDisclosure";

describe("Workbook incident heading", () => {
  afterEach(cleanup);

  it("keeps one full identity heading and a separate keyboard disclosure through identity changes", async () => {
    const user = userEvent.setup();
    const content = (incidentKey: string, title: string) => (
      <>
        <WorkbookIncidentIdentityDisclosure
          incidentKey={incidentKey}
          title={title}
        />
        <button type="button">Next control</button>
      </>
    );
    const { rerender } = render(
      content("INC-001", "A complete and deliberately long incident title"),
    );
    const heading = screen.getByRole("heading", {
      level: 1,
      name: "INC-001 A complete and deliberately long incident title",
    });
    expect(heading.tagName).toBe("H1");
    const trigger = within(heading).getByRole("button");
    trigger.focus();
    await user.keyboard("{Enter}");
    expect(trigger.getAttribute("aria-expanded")).toBe("true");
    const region = screen.getByRole("region", { name: "Incident identity" });
    expect(heading.contains(region)).toBe(false);
    expect(trigger.getAttribute("aria-controls")).toBe(region.id);
    await user.keyboard("{Escape}");
    expect(
      screen.queryByRole("region", { name: "Incident identity" }),
    ).toBeNull();
    expect(document.activeElement).toBe(trigger);
    await user.keyboard(" ");
    expect(trigger.getAttribute("aria-expanded")).toBe("true");
    await user.tab();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Next control" }),
    );
    expect(trigger.getAttribute("aria-expanded")).toBe("false");
    await user.click(trigger);
    fireEvent.pointerDown(screen.getByRole("button", { name: "Next control" }));
    expect(trigger.getAttribute("aria-expanded")).toBe("false");
    rerender(content("INC-001", "Updated title"));
    expect(
      screen.getByRole("heading", { level: 1, name: "INC-001 Updated title" }),
    ).toBe(heading);
    rerender(content("INC-RENAMED", "Updated title"));
    expect(
      screen.getByRole("heading", {
        level: 1,
        name: "INC-RENAMED Updated title",
      }),
    ).toBe(heading);
    rerender(content("INC-002", "Replacement identity"));
    expect(screen.getAllByRole("heading", { level: 1 })).toHaveLength(1);
    expect(
      screen.getByRole("heading", {
        level: 1,
        name: "INC-002 Replacement identity",
      }),
    ).toBe(heading);
    expect(screen.queryByText("INC-001")).toBeNull();
    await user.click(within(heading).getByRole("button"));
    expect(
      screen.getByRole("region", { name: "Incident identity" }).textContent,
    ).toBe("INC-002Replacement identity");
  });
});
