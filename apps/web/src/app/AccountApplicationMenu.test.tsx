import { cleanup, fireEvent, render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import { AccountApplicationMenu } from "./AccountApplicationMenu";
import type { AccountApplicationMenuProps } from "./landingAdminTypes";

afterEach(cleanup);

const controls = {
  activeSection: "summary",
  items: [
    {
      section: "summary",
      label: "Summary and preferences",
      description: "Incident summary and workbook defaults",
    },
    {
      section: "memberships",
      label: "Memberships",
      description: "Incident access and roles",
    },
  ],
  onSelectSection: vi.fn(),
} satisfies NonNullable<AccountApplicationMenuProps["incidentControls"]>;

function props(
  overrides: Partial<AccountApplicationMenuProps> = {},
): AccountApplicationMenuProps {
  return {
    canOpenDeploymentAdministration: true,
    currentContext: "workbook",
    currentUserLabel: "Analyst",
    currentIncidentRole: "admin",
    incidentControls: controls,
    onOpenAccountSettings: vi.fn(),
    onOpenDeploymentAdministration: vi.fn(),
    onOpenIncidentDirectory: vi.fn(),
    ...overrides,
  };
}

const trigger = () =>
  screen.getByRole("button", { name: "Account and application navigation" });
const item = (name: string) => screen.getByRole("menuitem", { name });

describe("account application menu interaction", () => {
  it("keeps unsupported shortcuts available while consuming menu keys", async () => {
    const user = userEvent.setup();
    const observed = vi.fn();
    render(
      <section aria-label="Keyboard owner" onKeyDown={observed}>
        <AccountApplicationMenu {...props()} />
      </section>,
    );
    await user.click(trigger());
    await user.keyboard("{ArrowDown}{Home}{End}");
    expect(observed).not.toHaveBeenCalled();
    fireEvent.keyDown(document.activeElement ?? document.body, {
      key: "k",
      ctrlKey: true,
    });
    fireEvent.keyDown(document.activeElement ?? document.body, {
      key: "Home",
      ctrlKey: true,
    });
    expect(observed).toHaveBeenCalledTimes(2);
    expect(
      observed.mock.calls.every(([event]) => !event.defaultPrevented),
    ).toBe(true);
  });

  it("hands Controls focus to its destination once for Space and pointer activation", async () => {
    const user = userEvent.setup();
    const destination = document.createElement("button");
    destination.textContent = "Drawer close";
    document.body.append(destination);
    const onSelectSection = vi.fn(() => destination.focus());
    render(
      <AccountApplicationMenu
        {...props({ incidentControls: { ...controls, onSelectSection } })}
      />,
    );
    for (const keyboard of [true, false]) {
      await user.click(trigger());
      await user.click(item("Controls"));
      const section = screen.getByRole("menuitem", {
        name: /Summary and preferences/,
      });
      if (keyboard) await user.keyboard(" ");
      else await user.click(section);
      expect(document.activeElement).toBe(destination);
      expect(trigger().getAttribute("aria-expanded")).toBe("false");
    }
    expect(onSelectSection).toHaveBeenCalledTimes(2);
    expect(onSelectSection).toHaveBeenLastCalledWith("summary", trigger());
    destination.remove();
  });

  it("reconciles nested removal and exits nested Tab without trapping or restoring", async () => {
    const user = userEvent.setup();
    const input = props();
    const { rerender } = render(<AccountApplicationMenu {...input} />);
    await user.click(trigger());
    await user.click(item("Controls"));
    rerender(
      <AccountApplicationMenu
        {...input}
        incidentControls={{ ...controls, items: controls.items.slice(1) }}
      />,
    );
    expect(document.activeElement?.textContent).toContain("Memberships");
    await user.tab({ shift: true });
    expect(document.activeElement).toBe(trigger());
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
    await user.click(trigger());
    await user.click(item("Controls"));
    rerender(
      <AccountApplicationMenu {...input} incidentControls={undefined} />,
    );
    expect(screen.queryByText("Controls")).toBeNull();
    expect(document.activeElement).toBe(item("Account settings"));
  });

  it("invalidates incident identity and never restores on ordinary unmount", async () => {
    const user = userEvent.setup();
    const input = props();
    const { rerender, unmount } = render(
      <AccountApplicationMenu {...input} subjectKey="incident-a" />,
    );
    await user.click(trigger());
    await user.click(item("Controls"));
    rerender(<AccountApplicationMenu {...input} subjectKey="incident-b" />);
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
    await user.click(trigger());
    const focus = vi.spyOn(trigger(), "focus");
    unmount();
    expect(focus).not.toHaveBeenCalled();
  });

  it("opens at the current action and moves by semantic menu items", async () => {
    const user = userEvent.setup();
    render(
      <AccountApplicationMenu
        {...props({ currentContext: "deployment-administration" })}
      />,
    );
    trigger().focus();
    await user.keyboard("{Enter}");
    expect(document.activeElement).toBe(item("Deployment administration"));
    await user.keyboard("{End}");
    expect(document.activeElement).toBe(item("Account settings"));
    await user.keyboard("{ArrowDown}");
    expect(document.activeElement).toBe(item("Incidents"));
    await user.keyboard("{ArrowUp}");
    expect(document.activeElement).toBe(item("Account settings"));
    await user.keyboard("{Home}");
    expect(document.activeElement).toBe(item("Incidents"));
  });

  it("opens Controls and closes the innermost menu before restoring the root trigger", async () => {
    const user = userEvent.setup();
    render(<AccountApplicationMenu {...props()} />);
    await user.click(trigger());
    const invoker = item("Controls");
    invoker.focus();
    await user.keyboard("{ArrowRight}");
    const summary = screen
      .getByText("Summary and preferences")
      .closest("button");
    expect(document.activeElement).toBe(summary);
    await user.keyboard("{End}");
    expect(document.activeElement?.textContent).toContain("Memberships");
    await user.keyboard("{Escape}");
    expect(screen.queryByText("Summary and preferences")).toBeNull();
    expect(document.activeElement).toBe(invoker);
    expect(trigger().getAttribute("aria-expanded")).toBe("true");
    await user.keyboard("{Escape}");
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
    expect(document.activeElement).toBe(trigger());
  });

  it("dismisses on an outside nonfocusable pointer without restoring focus", async () => {
    const user = userEvent.setup();
    render(
      <>
        <AccountApplicationMenu {...props()} />
        <p>Outside</p>
      </>,
    );
    await user.click(trigger());
    fireEvent.pointerDown(screen.getByText("Outside"));
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
  });

  it("closes stale context state without stealing destination focus", async () => {
    const user = userEvent.setup();
    const input = props();
    const { rerender } = render(
      <>
        <AccountApplicationMenu {...input} />
        <button type="button">Destination</button>
      </>,
    );
    await user.click(trigger());
    rerender(
      <>
        <AccountApplicationMenu {...input} currentContext="incidents" />
        <button type="button">Destination</button>
      </>,
    );
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
  });

  it("reconciles a removed focused capability to the next available action", async () => {
    const user = userEvent.setup();
    const input = props();
    const { rerender } = render(<AccountApplicationMenu {...input} />);
    await user.click(trigger());
    item("Deployment administration").focus();
    rerender(
      <AccountApplicationMenu
        {...input}
        canOpenDeploymentAdministration={false}
      />,
    );
    expect(document.activeElement).toBe(item("Controls"));
  });

  it("allows native Tab departure and dispatches an action only once", async () => {
    const user = userEvent.setup();
    const input = props();
    render(
      <>
        <AccountApplicationMenu {...input} />
        <button type="button">Next region</button>
      </>,
    );
    await user.click(trigger());
    item("Account settings").focus();
    await user.tab();
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Next region" }),
    );
    expect(trigger().getAttribute("aria-expanded")).toBe("false");
    await user.click(trigger());
    item("Incidents").focus();
    await user.keyboard("{Enter}");
    expect(input.onOpenIncidentDirectory).toHaveBeenCalledTimes(1);
  });
});
