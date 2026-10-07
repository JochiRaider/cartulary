import type { GridHandle } from "@cartulary/grid-adapter";
import { cleanup, render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  useWorkbookCommand,
  WorkbookCommandsControl,
  WorkbookCommandsProvider,
} from "./WorkbookCommands";
import type { WorkbookCommandDescriptor } from "./workbookCommandIndex";

afterEach(cleanup);

function Contribution({ command }: { command: WorkbookCommandDescriptor }) {
  useWorkbookCommand(command);
  return null;
}

function setup(count = 1, targetKind: "shell" | "record" = "record") {
  let recordId = "record-a";
  let blocked: string | null = null;
  const invoke = vi.fn(() => true);
  const grid = {
    getActiveCell: () => ({
      rowIdentity: { kind: "core_record", recordId },
      fieldKey: "note.title",
    }),
    getSelectedRecordIds: () => [],
  } as unknown as GridHandle;
  const commands: WorkbookCommandDescriptor[] = Array.from(
    { length: count },
    (_, index) => ({
      id: `inspect.${index}`,
      family: "Inspect",
      label: `Read source ${String(index).padStart(2, "0")}`,
      terms: ["source"],
      targetKind,
      availability: () => blocked,
      invoke,
    }),
  );
  const content = (readable: boolean) => (
    <WorkbookCommandsProvider
      grid={() => grid}
      readable={readable}
      surface="cartulary.view.notes.v1"
    >
      <input aria-label="Retained draft" defaultValue="Exact local draft" />
      {commands.map((command) => (
        <Contribution key={command.id} command={command} />
      ))}
      <WorkbookCommandsControl />
    </WorkbookCommandsProvider>
  );
  const view = render(content(true));
  return {
    user: userEvent.setup(),
    invoke,
    changeRecord: () => {
      recordId = "record-b";
    },
    block: () => {
      blocked = "Review the current saved version first.";
    },
    revoke: () => view.rerender(content(false)),
  };
}

describe("Commands interaction", () => {
  it("keeps the captured record and revalidates selection and owner eligibility before invocation", async () => {
    const h = setup();
    await h.user.click(screen.getByRole("button", { name: "Commands" }));
    h.changeRecord();
    await h.user.click(
      screen.getByRole("button", { name: "Inspect · Read source 00" }),
    );
    expect(h.invoke).not.toHaveBeenCalled();
    expect(screen.getByRole("status").textContent).toContain(
      "selection changed",
    );
    await h.user.click(screen.getByRole("button", { name: "Cancel" }));
    await h.user.click(screen.getByRole("button", { name: "Commands" }));
    h.block();
    await h.user.click(
      screen.getByRole("button", { name: "Inspect · Read source 00" }),
    );
    expect(h.invoke).not.toHaveBeenCalled();
    expect(screen.getByRole("status").textContent).toBe(
      "Review the current saved version first.",
    );
    expect(screen.getByRole("region", { name: "Commands" })).toBeTruthy();
  });

  it("pages all commands without execution and restores the borrowed draft on cancel", async () => {
    const h = setup(21, "shell");
    const draft = screen.getByRole("textbox", { name: "Retained draft" });
    draft.focus();
    await h.user.click(screen.getByRole("button", { name: "Commands" }));
    const panel = screen.getByRole("region", { name: "Commands" });
    expect(
      within(panel).getAllByRole("button", { name: /^Inspect ·/ }),
    ).toHaveLength(20);
    await h.user.click(screen.getByRole("button", { name: "Next commands" }));
    expect(
      within(panel).getAllByRole("button", { name: /^Inspect ·/ }),
    ).toHaveLength(1);
    await h.user.click(
      screen.getByRole("textbox", { name: "Search commands" }),
    );
    await h.user.keyboard("source 00{Enter}");
    expect(h.invoke).not.toHaveBeenCalled();
    expect(screen.getByRole("status").textContent).toBe(
      "1 commands. Page 1 of 1.",
    );
    await h.user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(document.activeElement).toBe(draft);
    expect((draft as HTMLInputElement).value).toBe("Exact local draft");
  });

  it("conceals open command context immediately when incident authority becomes uncertain", async () => {
    const h = setup();
    await h.user.click(screen.getByRole("button", { name: "Commands" }));
    h.revoke();
    expect(screen.queryByRole("region", { name: "Commands" })).toBeNull();
    expect(
      (
        screen.getByRole("button", {
          name: "Commands",
        }) as HTMLButtonElement
      ).disabled,
    ).toBe(true);
    expect(h.invoke).not.toHaveBeenCalled();
  });
});
