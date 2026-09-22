import {
  incidentMembershipCreateButtonTestId,
  incidentMembershipEmailInputTestId,
  incidentMembershipPatchButtonTestId,
  incidentMembershipRoleInputTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { MembershipTestSurface as MembershipSurface } from "../testing/incidentMembershipManagementSurfaceTestSupport";

const incidentId = "00000000-0000-4000-8000-000000001001";
const userId = "00000000-0000-4000-8000-000000000002";
const member = {
  incident_id: incidentId,
  user_id: userId,
  display_name: "Analyst",
  role: "viewer",
  membership_version: 1,
  joined_at: "2026-08-01T00:00:00Z",
  added_by_user_id: userId,
  updated_at: "2026-08-01T00:00:00Z",
  updated_by_user_id: userId,
};
const json = (data: unknown, status = 200) =>
  new Response(JSON.stringify(data), {
    status,
    headers: { "Content-Type": "application/json" },
  });
const envelope = (rows = [member], next: string | null = null) => ({
  data: { memberships: rows },
  meta: {
    request_id: "membership-characterization",
    paging: { limit: 100, has_more: next !== null, next_cursor: next },
  },
});
const error = (code: string, status: number) =>
  json(
    {
      error: { code, status, retryable: false, message: "Unsafe server prose" },
    },
    status,
  );
const subject = (
  id = incidentId,
  section: "memberships" | "incident-fields" = "memberships",
) => (
  <MembershipSurface
    incidentId={id}
    currentIncidentRole="admin"
    activeSection={section}
  />
);
function setup(next: string | null = null) {
  const writes: { resolve: (response: Response) => void; method: string }[] =
    [];
  let holdReads = false;
  const reads: ((response: Response) => void)[] = [];
  const fetch = vi.fn((input: unknown, init?: RequestInit) => {
    const url = String(input);
    if (init?.method && init.method !== "GET")
      return new Promise<Response>((resolve) =>
        writes.push({ resolve, method: init.method ?? "GET" }),
      );
    if (url.includes("/memberships")) {
      if (holdReads)
        return new Promise<Response>((resolve) => reads.push(resolve));
      return Promise.resolve(
        json(
          envelope(
            [{ ...member, incident_id: url.split("/")[4] ?? incidentId }],
            next,
          ),
        ),
      );
    }
    return Promise.resolve(
      json({
        data: {
          incident_id: url.split("/").at(-1),
          incident_key: "MM",
          title: "Memberships",
          status: "active",
          incident_version: 1,
        },
      }),
    );
  });
  vi.stubGlobal("fetch", fetch);
  const view = render(subject());
  return {
    view,
    writes,
    reads,
    fetch,
    holdReads: () => {
      holdReads = true;
    },
  };
}
async function ready() {
  await screen.findByText("Analyst");
  if (
    !screen.queryByTestId(incidentMembershipEmailInputTestId()) &&
    !screen.queryByTestId(incidentMembershipRoleInputTestId(userId))
  )
    fireEvent.click(
      screen.getByRole("button", { name: "Add existing account" }),
    );
}
function add(value = "analyst@example.test") {
  fireEvent.change(screen.getByTestId(incidentMembershipEmailInputTestId()), {
    target: { value },
  });
  fireEvent.click(screen.getByTestId(incidentMembershipCreateButtonTestId()));
}
async function settle(
  write: { resolve: (response: Response) => void } | undefined,
  result: Response,
) {
  if (!write) throw new Error("Missing admitted write");
  await act(async () => write.resolve(result));
}
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
describe("Incident membership management characterization", () => {
  it("makes server continuation reachable", async () => {
    setup("next-members");
    await ready();
    expect(screen.queryByRole("button", { name: "Next page" })).not.toBeNull();
  });
  it("admits a repeated create activation only once", async () => {
    const { writes } = setup();
    await ready();
    add();
    add();
    await act(async () => {});
    expect(writes).toHaveLength(1);
  });
  it("preserves a newer add draft when the captured create completes", async () => {
    const { writes } = setup();
    await ready();
    add();
    fireEvent.change(screen.getByTestId(incidentMembershipEmailInputTestId()), {
      target: { value: "newer@example.test" },
    });
    await waitFor(() => expect(writes).toHaveLength(1));
    await settle(
      writes[0],
      json({ data: member, meta: { request_id: "created" } }, 201),
    );
    expect(
      (
        screen.getByTestId(
          incidentMembershipEmailInputTestId(),
        ) as HTMLInputElement
      ).value,
    ).toBe("newer@example.test");
  });
  it("preserves a role draft through a section reload", async () => {
    const { view } = setup();
    await ready();
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    fireEvent.change(
      screen.getByTestId(incidentMembershipRoleInputTestId(userId)),
      { target: { value: "reviewer" } },
    );
    view.rerender(subject(incidentId, "incident-fields"));
    view.rerender(subject());
    await ready();
    expect(
      (
        screen.getByTestId(
          incidentMembershipRoleInputTestId(userId),
        ) as HTMLSelectElement
      ).value,
    ).toBe("reviewer");
  });
  it("does not publish an older write rejection into another incident", async () => {
    const { view, writes } = setup();
    await ready();
    add();
    await waitFor(() => expect(writes).toHaveLength(1));
    view.rerender(subject("00000000-0000-4000-8000-000000001002"));
    await ready();
    await waitFor(() => expect(writes).toHaveLength(1));
    await settle(writes[0], error("user_inactive", 409));
    expect(screen.queryByText("user_inactive")).toBeNull();
    expect(screen.queryByText(/inactive account/iu)).toBeNull();
  });
  it("acknowledges a write before follow-up reads finish", async () => {
    const { writes, holdReads, reads } = setup();
    await ready();
    holdReads();
    add();
    await waitFor(() => expect(writes).toHaveLength(1));
    await settle(
      writes[0],
      json({ data: member, meta: { request_id: "created" } }, 201),
    );
    await waitFor(() => expect(reads.length).toBeGreaterThan(0));
    expect(
      screen.queryByText(/Added membership|Membership added/u),
    ).not.toBeNull();
  });
  it("separates uncertain writes from rejected version conflicts", async () => {
    const { writes } = setup();
    await ready();
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    fireEvent.click(
      screen.getByTestId(incidentMembershipPatchButtonTestId(userId)),
    );
    await waitFor(() => expect(writes).toHaveLength(1));
    await settle(writes[0], error("membership_version_conflict", 409));
    expect(
      screen.queryByRole("button", { name: /Observe current membership/u }),
    ).not.toBeNull();
  });
  it("names role controls for their member", async () => {
    setup();
    await ready();
    fireEvent.click(
      screen.getByRole("button", { name: /Change role for Analyst/u }),
    );
    expect(
      screen.queryByRole("combobox", { name: /Role for Analyst/u }),
    ).not.toBeNull();
  });
});
