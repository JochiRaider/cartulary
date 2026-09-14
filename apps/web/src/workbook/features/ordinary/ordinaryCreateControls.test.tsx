import { requireViewContract } from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { emptyGenericReferenceOptions } from "../../models/workbookReferenceOptions";
import type { WorkbookAuthoringReadPort } from "../../ports/WorkbookAuthoringReadPort";
import { OrdinaryCreateControl } from "./OrdinaryCreateControl";
import { OrdinaryCreateNotice } from "./OrdinaryCreateNotice";
import { ordinaryCreateContributions } from "./ordinaryCreateContributions";
import { WorkbookOrdinaryCreateOwner } from "./WorkbookOrdinaryCreateOwner";

const id = "10000000-0000-4000-8000-000000000001";
const second = "10000000-0000-4000-8000-000000000002";
const contract = requireViewContract("cartulary.view.handoff.v1");
const field = contract.fieldMap["handoff.incoming_owner_user_id"];
function setup() {
  if (!field) throw new Error("Missing incoming member contract.");
  const owner = new WorkbookOrdinaryCreateOwner(
    id,
    ordinaryCreateContributions,
  );
  owner.setAuthority({
    actorId: id,
    incidentId: id,
    sessionIdentity: "account",
    role: "editor",
    closed: false,
  });
  const reader: WorkbookAuthoringReadPort = {
    availableViews: vi.fn(async () => []),
    verify: vi.fn(async () => {}),
    page: vi.fn(async (input) => ({
      kind: "accepted" as const,
      value: {
        candidates: [
          {
            recordId: input.cursor ? second : id,
            displayText: input.cursor ? "Second member" : "First member",
            viewSchemaId: "incident_members",
          },
        ],
        hasMore: !input.cursor,
        nextCursor: input.cursor ? null : "page2",
      },
    })),
  };
  owner.configureReader(reader);
  const renderControl = () =>
    render(
      <OrdinaryCreateControl
        owner={owner}
        contract={contract}
        field={field}
        surface="grid"
        collectionMode="add"
        referenceOptions={emptyGenericReferenceOptions()}
        testId="incoming"
        value={
          owner.getSnapshot().schemas[contract.viewSchemaId]?.values[
            field.fieldKey
          ] ?? ""
        }
        onChange={(value) =>
          owner.update(contract.viewSchemaId, field.fieldKey, value)
        }
      />,
    );
  return { owner, reader, renderControl };
}
afterEach(cleanup);
describe("ordinary workbook reference controls", () => {
  it("keeps raw authoring copyable outside the read-only grid and conceals it with authority", () => {
    const { owner } = setup();
    owner.update(
      contract.viewSchemaId,
      "handoff.current_state_summary",
      "  exact\nretained text  ",
    );
    render(<OrdinaryCreateNotice owner={owner} view={contract.viewSchemaId} />);
    expect(screen.queryByRole("textbox")).toBeNull();
    act(() =>
      owner.setAuthority({
        actorId: id,
        incidentId: id,
        sessionIdentity: "account",
        role: "viewer",
        closed: false,
      }),
    );
    const input = screen.getByRole("textbox", {
      name: "Current State retained authoring",
    }) as HTMLTextAreaElement;
    expect(input.readOnly).toBe(true);
    expect(input.value).toBe("  exact\nretained text  ");
    act(() => owner.suspend());
    expect(screen.queryByRole("textbox")).toBeNull();
    expect(screen.queryByText("Unfinished row (read only)")).toBeNull();
  });
  it("stages paged selection, retains off-page identity and returns focus on Escape", async () => {
    const { owner, reader, renderControl } = setup();
    const control = renderControl();
    fireEvent.click(
      screen.getByRole("button", { name: "Choose incoming owner" }),
    );
    const select = await screen.findByTestId("incoming-options");
    await screen.findByRole("option", { name: "First member" });
    fireEvent.change(select, { target: { value: id } });
    expect(
      owner.getSnapshot().schemas[contract.viewSchemaId]?.draft.values,
    ).toEqual({});
    fireEvent.click(
      screen.getByRole("button", { name: "Load more references" }),
    );
    await screen.findByRole("option", { name: "Second member" });
    fireEvent.click(screen.getByRole("button", { name: "Apply references" }));
    expect(
      owner.getSnapshot().schemas[contract.viewSchemaId]?.draft.values[
        field?.fieldKey ?? ""
      ],
    ).toBe(id);
    control.unmount();
    vi.mocked(reader.page).mockResolvedValue({
      kind: "accepted",
      value: { candidates: [], hasMore: false, nextCursor: null },
    });
    renderControl();
    const trigger = screen.getByRole("button", {
      name: "Choose incoming owner",
    });
    fireEvent.click(trigger);
    await screen.findByText("No available references.");
    expect(screen.getByRole("option", { name: "First member" })).toBeTruthy();
    fireEvent.keyDown(
      screen.getByRole("region", { name: "Choose incoming owner" }),
      { key: "Escape" },
    );
    expect(document.activeElement).toBe(trigger);
    expect(
      owner.getSnapshot().schemas[contract.viewSchemaId]?.draft.references[
        field?.fieldKey ?? ""
      ]?.[0]?.recordId,
    ).toBe(id);
  });
  it("exposes failed reference recovery and conceals retained labels when authorization suspends", async () => {
    const { owner, reader, renderControl } = setup();
    vi.mocked(reader.availableViews).mockRejectedValueOnce(
      new Error("lost response"),
    );
    renderControl();
    fireEvent.click(
      screen.getByRole("button", { name: "Choose incoming owner" }),
    );
    await screen.findByRole("button", { name: "Retry surfaces" });
    fireEvent.click(screen.getByRole("button", { name: "Retry surfaces" }));
    await screen.findByRole("option", { name: "First member" });
    act(() => owner.suspend());
    expect(screen.queryByTestId("incoming")).toBeNull();
    expect(screen.queryByText("First member")).toBeNull();
  });
});
