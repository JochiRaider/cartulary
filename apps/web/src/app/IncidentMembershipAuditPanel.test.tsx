import {
  incidentAdministrationTestId,
  incidentMembershipAuditRowTestId,
} from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { MembershipAuditTestSurface } from "../testing/incidentMembershipAuditTestSupport";

const jsonResponse = (payload: unknown) =>
  new Response(JSON.stringify(payload), {
    status: 200,
    headers: { "Content-Type": "application/json" },
  });
describe("Incident membership audit presentation", () => {
  let fetchMock: ReturnType<typeof vi.fn>;
  beforeEach(() => {
    fetchMock = vi.fn();
    vi.stubGlobal("fetch", fetchMock);
  });
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });
  it("keeps membership audit placement inside incident admin controls", async () => {
    let auditReads = 0;
    fetchMock.mockImplementation((input) => {
      const url = String(input);
      if (url === "/api/v1/incidents/00000000-0000-4000-8000-000000001001") {
        return Promise.resolve(jsonResponse({ data: {} }));
      }
      if (
        url.startsWith(
          "/api/v1/incidents/00000000-0000-4000-8000-000000001001/membership-audit-events?",
        )
      ) {
        auditReads += 1;
        const continuation = url.includes("cursor_token=membership-next");
        return Promise.resolve(
          jsonResponse({
            data: {
              audit_events: [
                membershipAuditEvent(
                  continuation
                    ? "00000000-0000-4000-8000-000000002002"
                    : "00000000-0000-4000-8000-000000002001",
                  continuation
                    ? "membership_deleted"
                    : "membership_role_changed",
                ),
              ],
            },
            meta: {
              paging: {
                has_more: !continuation,
                limit: 100,
                next_cursor: continuation ? null : "membership-next",
              },
              request_id: `request-${auditReads}`,
            },
          }),
        );
      }
      throw new Error(`unexpected fetch: ${url}`);
    });

    const view = render(
      <MembershipAuditTestSurface
        activeSection="membership-audit"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );

    expect(
      await screen.findByTestId(
        incidentMembershipAuditRowTestId(
          "00000000-0000-4000-8000-000000002001",
        ),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(
        incidentAdministrationTestId("membership-audit-status"),
      ).textContent,
    ).toContain("Page 1:");
    fireEvent.click(screen.getByRole("button", { name: /Inspect.*event/u }));
    expect(screen.getAllByText("Redacted")).toHaveLength(2);
    fireEvent.click(screen.getByRole("button", { name: "Next page" }));
    expect(
      await screen.findByTestId(
        incidentMembershipAuditRowTestId(
          "00000000-0000-4000-8000-000000002002",
        ),
      ),
    ).toBeTruthy();
    expect(auditReads).toBe(2);
    expect(
      fetchMock.mock.calls.some(([input]) =>
        String(input).includes("/api/v1/administrative-audit-events"),
      ),
    ).toBe(false);

    view.rerender(
      <MembershipAuditTestSurface
        activeSection="membership-audit"
        currentIncidentRole="reviewer"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );
    expect(
      await screen.findByTestId(
        incidentAdministrationTestId("membership-audit-note"),
      ),
    ).toBeTruthy();
    expect(
      screen.getByTestId(incidentAdministrationTestId("membership-audit-note"))
        .textContent,
    ).toContain("Only incident admins");
    expect(auditReads).toBe(2);
  });
  it("uses a native form with local exact target and timestamp errors", async () => {
    fetchMock.mockImplementation(() =>
      Promise.resolve(
        jsonResponse({
          data: { audit_events: [] },
          meta: {
            request_id: "empty",
            paging: { limit: 100, has_more: false, next_cursor: null },
          },
        }),
      ),
    );
    render(
      <MembershipAuditTestSurface
        activeSection="membership-audit"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );
    await screen.findByText("No membership audit events yet.");
    const form = screen.getByRole("form", { name: "Membership audit filters" });
    fireEvent.change(screen.getByRole("textbox", { name: "Target ID" }), {
      target: { value: "member" },
    });
    fireEvent.submit(form);
    expect(
      screen
        .getByRole("combobox", { name: "Target kind" })
        .getAttribute("aria-invalid"),
    ).toBe("true");
    expect(document.activeElement).toBe(
      screen.getByRole("combobox", { name: "Target kind" }),
    );
    fireEvent.change(screen.getByRole("combobox", { name: "Target kind" }), {
      target: { value: "incident_membership" },
    });
    fireEvent.change(
      screen.getByRole("textbox", { name: "Occurred at or after" }),
      { target: { value: "2026-05-24T12:00:00" } },
    );
    fireEvent.submit(form);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    fireEvent.change(
      screen.getByRole("textbox", { name: "Occurred at or after" }),
      { target: { value: "2026-05-24T12:00:00.000000001Z" } },
    );
    fireEvent.submit(form);
    await screen.findByText(
      "No membership audit events match the applied filters.",
    );
    expect(fetchMock).toHaveBeenCalledTimes(2);
  });
  it("distinguishes unavailable initial reads from stale refresh and explicit retry", async () => {
    let mode: "failed" | "ready" = "failed";
    fetchMock.mockImplementation(() =>
      mode === "failed"
        ? Promise.resolve(
            new Response(
              JSON.stringify({
                error: {
                  code: "service_unavailable",
                  status: 503,
                  retryable: true,
                },
              }),
              { status: 503, headers: { "Content-Type": "application/json" } },
            ),
          )
        : Promise.resolve(
            jsonResponse({
              data: {
                audit_events: [
                  membershipAuditEvent(
                    "00000000-0000-4000-8000-000000002001",
                    "membership_created",
                  ),
                ],
              },
              meta: {
                request_id: "ready",
                paging: { limit: 100, has_more: false, next_cursor: null },
              },
            }),
          ),
    );
    render(
      <MembershipAuditTestSurface
        activeSection="membership-audit"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );
    await screen.findByText(
      "Membership audit is unavailable. Try the read again.",
    );
    expect(screen.queryByText("No membership audit events yet.")).toBeNull();
    mode = "ready";
    fireEvent.click(screen.getByRole("button", { name: "Try the read again" }));
    await screen.findByText("Membership created");
    mode = "failed";
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await screen.findByText(/Membership audit refresh failed/u);
    expect(screen.getByText("Membership created")).toBeTruthy();
    expect(
      screen.getByText(/Previous results; displayed events may be stale/u),
    ).toBeTruthy();
  });
  it("preserves expanded identity and focused controls until replacement removes the event", async () => {
    let id = "00000000-0000-4000-8000-000000002001";
    fetchMock.mockImplementation(() =>
      Promise.resolve(
        jsonResponse({
          data: {
            audit_events: [membershipAuditEvent(id, "membership_created")],
          },
          meta: {
            request_id: "ready",
            paging: { limit: 100, has_more: false, next_cursor: null },
          },
        }),
      ),
    );
    render(
      <MembershipAuditTestSurface
        activeSection="membership-audit"
        currentIncidentRole="admin"
        incidentId="00000000-0000-4000-8000-000000001001"
      />,
    );
    const inspect = await screen.findByRole("button", {
      name: /Inspect.*event/u,
    });
    inspect.focus();
    fireEvent.click(inspect);
    const refresh = screen.getByRole("button", { name: "Refresh" });
    fireEvent.click(refresh);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    await act(async () => {});
    expect(inspect.getAttribute("aria-expanded")).toBe("true");
    expect(document.activeElement).toBe(inspect);
    id = "00000000-0000-4000-8000-000000002002";
    fireEvent.click(refresh);
    await waitFor(() => expect(inspect.isConnected).toBe(false));
    expect(document.activeElement).toBe(refresh);
  });
});
function membershipAuditEvent(auditEventID: string, actionCode: string) {
  return {
    action_code: actionCode,
    actor_kind: "user",
    actor_user_id: "00000000-0000-4000-8000-000000000010",
    audit_event_id: auditEventID,
    changes: [
      {
        after: null,
        before: null,
        field_path: "role",
        value_state: "redacted",
      },
    ],
    occurred_at: "2026-07-26T12:00:00Z",
    reason_code: null,
    scope_id: "00000000-0000-4000-8000-000000001001",
    scope_kind: "incident",
    source: "ui",
    target_id: "00000000-0000-4000-8000-000000000020",
    target_kind: "incident_membership",
  };
}
