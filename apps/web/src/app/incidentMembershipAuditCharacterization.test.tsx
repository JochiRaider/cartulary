import { incidentMembershipAuditRowTestId } from "@cartulary/ui-contracts";
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
  auditEnvelope,
  auditEvent,
} from "../testing/administrativeAuditTestSupport";
import { MembershipAuditTestSurface as AuditSurface } from "../testing/incidentMembershipAuditTestSupport";

const incidentId = "00000000-0000-4000-8000-000000001001";
const event = (id = "00000000-0000-4000-8000-000000002001") =>
  auditEvent({
    audit_event_id: id,
    scope_kind: "incident",
    scope_id: incidentId,
    action_code: "membership_role_changed",
    target_kind: "incident_membership",
    changes: [
      {
        field_path: "role",
        value_state: "visible",
        before: "viewer",
        after: "admin",
      },
    ],
  });
const response = (rows = [event()], cursor: string | null = "next") =>
  new Response(
    JSON.stringify({
      ...auditEnvelope(rows, cursor),
      meta: {
        request_id: "characterization",
        paging: { limit: 100, has_more: cursor !== null, next_cursor: cursor },
      },
    }),
    { status: 200, headers: { "Content-Type": "application/json" } },
  );
function setup() {
  const reads: { url: string; resolve: (value: Response) => void }[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn((input: unknown) => {
      const url = String(input);
      if (url.includes("membership-audit-events"))
        return new Promise<Response>((resolve) => reads.push({ url, resolve }));
      return Promise.resolve(
        new Response(
          JSON.stringify({
            data: {
              incident_id: incidentId,
              name: "Audit incident",
              status: "open",
            },
          }),
          { status: 200 },
        ),
      );
    }),
  );
  const view = render(
    <AuditSurface
      incidentId={incidentId}
      currentIncidentRole="admin"
      activeSection="membership-audit"
    />,
  );
  return { reads, view };
}
async function resolve(
  read: { resolve: (value: Response) => void },
  value = response(),
) {
  await act(async () => read.resolve(value));
}
const apply = () =>
  fireEvent.click(screen.getByRole("button", { name: "Apply filters" }));
const next = () =>
  fireEvent.click(screen.getByRole("button", { name: /Load more|Next page/u }));
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

describe("Incident membership audit characterization", () => {
  it("does not present initial loading as a successful empty result", async () => {
    const { reads } = setup();
    await waitFor(() => expect(reads).toHaveLength(1));
    await waitFor(() => expect(reads).toHaveLength(1));
    expect(screen.queryByText("No membership audit events loaded.")).toBeNull();
  });
  it("continues the accepted query despite unsubmitted filter edits", async () => {
    const { reads } = setup();
    await waitFor(() => expect(reads).toHaveLength(1));
    await resolve(required(reads[0]));
    fireEvent.change(screen.getByRole("combobox", { name: /action/iu }), {
      target: { value: "membership_deleted" },
    });
    next();
    await waitFor(() => expect(reads).toHaveLength(2));
    expect(
      new URL(required(reads[1]).url, "http://localhost").searchParams.has(
        "action_code",
      ),
    ).toBe(false);
  });
  it("coalesces repeated Apply and page admissions", async () => {
    const { reads } = setup();
    await waitFor(() => expect(reads).toHaveLength(1));
    await resolve(required(reads[0]));
    apply();
    apply();
    await waitFor(() => expect(reads).toHaveLength(2));
    await resolve(required(reads[1]));
    next();
    next();
    await waitFor(() => expect(reads).toHaveLength(3));
  });
  it("rejects a delayed success after a newer accepted query", async () => {
    const { reads } = setup();
    await waitFor(() => expect(reads).toHaveLength(1));
    fireEvent.change(screen.getByRole("combobox", { name: /action/iu }), {
      target: { value: "membership_deleted" },
    });
    apply();
    await waitFor(() => expect(reads).toHaveLength(2));
    await resolve(
      required(reads[1]),
      response([event("00000000-0000-4000-8000-000000002002")], null),
    );
    await resolve(required(reads[0]));
    expect(
      screen.queryByTestId(
        incidentMembershipAuditRowTestId(
          "00000000-0000-4000-8000-000000002001",
        ),
      ),
    ).toBeNull();
  });
  it("does not publish delayed audit failure into another section", async () => {
    const { reads, view } = setup();
    await waitFor(() => expect(reads).toHaveLength(1));
    view.rerender(
      <AuditSurface
        incidentId={incidentId}
        currentIncidentRole="admin"
        activeSection="incident-fields"
      />,
    );
    await resolve(
      required(reads[0]),
      new Response(
        JSON.stringify({
          error: {
            code: "service_unavailable",
            message: "old audit failure",
            status: 503,
            retryable: true,
          },
        }),
        { status: 503 },
      ),
    );
    expect(screen.queryByText(/old audit failure/u)).toBeNull();
  });
  it("keeps visible JSON strings distinct from null and structured values", async () => {
    const { reads } = setup();
    await waitFor(() => expect(reads).toHaveLength(1));
    await resolve(
      required(reads[0]),
      response(
        [
          {
            ...event(),
            changes: [
              {
                field_path: "a",
                value_state: "visible",
                before: null,
                after: "null",
              },
              {
                field_path: "b",
                value_state: "visible",
                before: { role: "viewer" },
                after: [false, 0],
              },
            ],
          },
        ],
        null,
      ),
    );
    const inspect = screen.queryByRole("button", { name: /Inspect.*event/u });
    if (inspect) fireEvent.click(inspect);
    expect(screen.queryByText('"null"')).not.toBeNull();
    expect(screen.queryByText(/\[object Object\]/u)).toBeNull();
  });
});

function required<T>(value: T | null | undefined): T {
  if (value == null) throw new Error("Expected audit fixture value");
  return value;
}
