import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  membershipAuthority as authority,
  membershipJSON as json,
  membershipEnvelope,
  membershipFixture,
} from "../testing/incidentMembershipManagementTestSupport";
import { MembershipTestSurface } from "./incidentMembershipManagementTestSurface";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
describe("Incident membership management presentation", () => {
  it("issues versioned create patch and bodyless delete with deliberate review and retained receipts", async () => {
    let rows: ReturnType<typeof membershipFixture>[] = [];
    const writes: {
      method: string;
      body: Record<string, unknown>;
      headers: Headers;
    }[] = [];
    vi.stubGlobal(
      "fetch",
      vi.fn(async (_input: unknown, init?: RequestInit) => {
        const method = init?.method ?? "GET";
        if (method === "GET") return json(membershipEnvelope(rows));
        const body = JSON.parse(String(init?.body)) as Record<string, unknown>;
        writes.push({ method, body, headers: new Headers(init?.headers) });
        if (method === "DELETE") {
          rows = [];
          return new Response(null, { status: 204 });
        }
        rows = [
          membershipFixture({
            role: method === "POST" ? "viewer" : "reviewer",
            membership_version: method === "POST" ? 1 : 2,
          }),
        ];
        return json(
          { data: rows[0], meta: { request_id: "write" } },
          method === "POST" ? 201 : 200,
        );
      }),
    );
    render(
      <MembershipTestSurface
        incidentId={authority.incidentId}
        currentIncidentRole="admin"
        activeSection="memberships"
      />,
    );
    await screen.findByText("No memberships on this page.");
    fireEvent.click(
      screen.getByRole("button", { name: "Add existing account" }),
    );
    fireEvent.change(screen.getByRole("textbox", { name: "User email" }), {
      target: { value: " analyst@example.test " },
    });
    fireEvent.click(screen.getByRole("button", { name: "Add membership" }));
    await screen.findByText("Membership added.");
    await screen.findByText("Analyst");
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    fireEvent.change(
      screen.getByRole("combobox", { name: /Role for Analyst/u }),
      { target: { value: "reviewer" } },
    );
    fireEvent.click(screen.getByRole("button", { name: "Save role" }));
    await screen.findByText("Membership role saved.");
    await screen.findByText("Version 2");
    fireEvent.click(
      screen.getByRole("button", {
        name: /Remove incident access for Analyst/u,
      }),
    );
    expect(writes).toHaveLength(2);
    expect(
      screen.getByText(
        /account and its access to other incidents are not deleted/u,
      ),
    ).toBeTruthy();
    const confirm = screen.getByRole("button", { name: "Confirm removal" });
    confirm.focus();
    fireEvent.click(confirm);
    await screen.findByText("Incident membership removed.");
    await waitFor(() => expect(screen.queryByText("Analyst")).toBeNull());
    expect(document.activeElement).toBe(screen.getByRole("heading"));
    expect(writes.map((write) => write.method)).toEqual([
      "POST",
      "PATCH",
      "DELETE",
    ]);
    expect(writes[0]?.body).toEqual({
      email: "analyst@example.test",
      role: "viewer",
      client_txn_id: expect.any(String),
    });
    expect(writes[1]?.body).toEqual({
      base_membership_version: 1,
      role: "reviewer",
    });
    expect(writes[2]?.body).toEqual({ base_membership_version: 2 });
    for (const write of writes)
      expect(write.headers.get("Content-Type")).toBe("application/json");
  });
  it("keeps non-admin incident browsing and excludes privileged forms", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => json(membershipEnvelope())),
    );
    const view = render(
      <MembershipTestSurface
        incidentId={authority.incidentId}
        currentIncidentRole="viewer"
        activeSection="memberships"
      />,
    );
    await screen.findByText("Analyst");
    expect(screen.getByText(/Only incident admins/u)).toBeTruthy();
    expect(
      screen.queryByRole("button", { name: "Add existing account" }),
    ).toBeNull();
    view.rerender(
      <MembershipTestSurface
        incidentId={authority.incidentId}
        currentIncidentRole="admin"
        activeSection="memberships"
      />,
    );
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    expect(
      screen.getByRole("combobox", { name: /Role for Analyst/u }),
    ).toBeTruthy();
    view.rerender(
      <MembershipTestSurface
        incidentId={authority.incidentId}
        currentIncidentRole="reviewer"
        activeSection="memberships"
      />,
    );
    expect(screen.queryByRole("combobox")).toBeNull();
    expect(screen.getByText("Analyst")).toBeTruthy();
  });
  it("requires Stay or Discard before replacing dirty input and preserves explicit cancellation", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => json(membershipEnvelope())),
    );
    render(
      <MembershipTestSurface
        incidentId={authority.incidentId}
        currentIncidentRole="admin"
        activeSection="memberships"
      />,
    );
    await screen.findByText("Analyst");
    fireEvent.click(
      screen.getByRole("button", { name: "Add existing account" }),
    );
    fireEvent.change(screen.getByRole("textbox"), {
      target: { value: "draft@example.test" },
    });
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    expect(
      screen.getByRole("dialog", { name: "Discard this membership draft?" }),
    ).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Stay" }));
    expect((screen.getByRole("textbox") as HTMLInputElement).value).toBe(
      "draft@example.test",
    );
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Discard draft" }));
    expect(screen.queryByRole("textbox")).toBeNull();
    expect(
      screen.getByRole("combobox", { name: /Role for Analyst/u }),
    ).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    await act(async () => {});
    expect(screen.queryByRole("form")).toBeNull();
  });
});
