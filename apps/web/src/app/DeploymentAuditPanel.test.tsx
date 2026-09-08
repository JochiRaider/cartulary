import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StrictMode } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  auditActorId,
  auditEnvelope,
  auditEvent,
} from "../testing/administrativeAuditTestSupport";
import {
  deferred,
  errorResponse,
  jsonResponse,
} from "../testing/fetchMockTestSupport";
import { AdministrativeAuditPanel } from "./DeploymentAuditPanel";
import { useAdministrativeAudit } from "./useAdministrativeAudit";

function AuditHarness({
  active = true,
  lifetime = "test-session",
  visible = true,
}: {
  active?: boolean;
  lifetime?: string | null;
  visible?: boolean;
}) {
  const controller = useAdministrativeAudit({
    authority: lifetime ? { lifetime, actorId: auditActorId } : null,
    active,
    isCurrent: (authority) => authority.lifetime === lifetime,
    confirmAccess: async () => ({ kind: "authorized" }),
    authorizationFailed: () => {},
  });
  return visible ? (
    <section hidden={!active}>
      <AdministrativeAuditPanel controller={controller} active={active} />
    </section>
  ) : null;
}

function envelope(action = "user_created", more = false) {
  return {
    data: {
      audit_events: [
        {
          audit_event_id: "00000000-0000-4000-8000-000000002001",
          scope_kind: "deployment",
          scope_id: null,
          occurred_at: "2026-05-24T00:00:00Z",
          actor_kind: "user",
          actor_user_id: "00000000-0000-4000-8000-000000000001",
          source: "ui",
          action_code: action,
          target_kind: "user",
          target_id: "00000000-0000-4000-8000-000000000002",
          reason_code: null,
          changes: [
            {
              field_path: "display_name",
              value_state: "visible",
              before: null,
              after: "Target User",
            },
          ],
        },
      ],
    },
    meta: {
      request_id: "audit-test",
      paging: {
        limit: 100,
        has_more: more,
        next_cursor: more ? "next-page" : null,
      },
    },
  };
}

describe("Administrative audit browsing", () => {
  const fetchMock = vi.fn<typeof fetch>();
  beforeEach(() => {
    fetchMock.mockReset();
    vi.stubGlobal("fetch", fetchMock);
  });
  afterEach(() => {
    cleanup();
    vi.unstubAllGlobals();
  });

  it("refreshes only the applied exact filters", async () => {
    fetchMock.mockImplementation(async () => jsonResponse(envelope()));
    render(<AuditHarness />);
    await screen.findByText("User created");
    fireEvent.change(screen.getByLabelText(/Actor user id/i), {
      target: { value: "00000000-0000-4000-8000-000000000003" },
    });
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect(String(fetchMock.mock.calls[1]?.[0])).toBe(
      "/api/v1/administrative-audit-events?limit=100",
    );
  });

  it("provides an explicit native Apply filters form", async () => {
    fetchMock.mockImplementation(async () => jsonResponse(envelope()));
    render(<AuditHarness />);
    await screen.findByText("User created");
    expect(
      screen.getByRole("button", { name: "Apply filters" }).closest("form"),
    ).not.toBeNull();
  });

  it("exposes continuation when more audit events exist", async () => {
    fetchMock.mockImplementation(async () =>
      jsonResponse(envelope("user_created", true)),
    );
    render(<AuditHarness />);
    await screen.findByText("User created");
    expect(screen.getByRole("button", { name: "Next page" })).toBeTruthy();
  });

  it("does not label an unresolved initial read as empty", async () => {
    const read = deferred<Response>();
    fetchMock.mockReturnValue(read.promise);
    render(<AuditHarness />);
    expect(
      screen.queryByText(/No administrative audit events loaded/),
    ).toBeNull();
    await act(async () => read.resolve(jsonResponse(envelope())));
  });

  it("does not read a hidden inactive panel", async () => {
    fetchMock.mockImplementation(async () => jsonResponse(envelope()));
    render(<AuditHarness active={false} />);
    await act(async () => {});
    expect(fetchMock).not.toHaveBeenCalled();
  });

  it("keeps the admitted refresh after a superseded initial response settles", async () => {
    const old = deferred<Response>();
    fetchMock
      .mockReturnValueOnce(old.promise)
      .mockImplementation(async () => jsonResponse(envelope("password_reset")));
    render(<AuditHarness />);
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await screen.findByText("Password reset");
    await act(async () => old.resolve(jsonResponse(envelope())));
    expect(screen.queryByText("User created")).toBeNull();
    expect(screen.getByText("Password reset")).toBeTruthy();
  });

  it("applies with Enter and associates target and timestamp errors locally", async () => {
    fetchMock.mockImplementation(async () => jsonResponse(envelope()));
    const user = userEvent.setup();
    render(<AuditHarness />);
    await screen.findByText("User created");
    const target = screen.getByLabelText("Target ID");
    await user.type(target, "target");
    await user.keyboard("{Enter}");
    expect(
      screen.getByLabelText("Target kind").getAttribute("aria-invalid"),
    ).toBe("true");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await user.selectOptions(screen.getByLabelText("Target kind"), "user");
    await user.type(
      screen.getByLabelText("Occurred at or after"),
      "2026-05-24T02:00:00+02:00",
    );
    await user.type(
      screen.getByLabelText("Occurred before"),
      "2026-05-24T00:00:00Z",
    );
    await user.keyboard("{Enter}");
    expect(
      screen.getByLabelText("Occurred before").getAttribute("aria-describedby"),
    ).toContain("occurred_at_lt-error");
    expect(fetchMock).toHaveBeenCalledTimes(1);
    await user.clear(screen.getByLabelText("Occurred before"));
    await user.type(
      screen.getByLabelText("Occurred before"),
      "2026-05-25T00:00:00Z",
    );
    await user.keyboard("{Enter}");
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    const url = new URL(
      String(fetchMock.mock.calls[1]?.[0]),
      "http://cartulary.test",
    );
    expect(Object.fromEntries(url.searchParams)).toMatchObject({
      target_kind: "user",
      target_id: "target",
      occurred_at_gte: "2026-05-24T00:00:00Z",
      occurred_at_lt: "2026-05-25T00:00:00Z",
    });
  });

  it("preserves expanded identity and moves focus only when replacement removes it", async () => {
    fetchMock.mockImplementation(async () => jsonResponse(envelope()));
    const user = userEvent.setup();
    render(<AuditHarness />);
    await screen.findByText("User created");
    await user.click(
      screen.getByRole("button", { name: /^Inspect User created/ }),
    );
    const hide = screen.getByRole("button", { name: /^Hide User created/ });
    const detailsId = hide.getAttribute("aria-controls");
    expect(document.getElementById(detailsId ?? "")).not.toBeNull();
    const same = deferred<Response>();
    fetchMock.mockReturnValueOnce(same.promise);
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    hide.focus();
    await act(async () => same.resolve(jsonResponse(envelope())));
    expect(document.activeElement).toBe(hide);
    const replacement = deferred<Response>();
    fetchMock.mockReturnValueOnce(replacement.promise);
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    hide.focus();
    await act(async () => replacement.resolve(jsonResponse(auditEnvelope([]))));
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Refresh" }),
    );
    expect(document.getElementById(detailsId ?? "")).toBeNull();
  });

  it("renders additive vocabulary and distinct visible JSON values versus redaction", async () => {
    const changes = [
      {
        field_path: "a_empty",
        value_state: "visible" as const,
        before: "",
        after: "<img onerror=alert(1)>",
      },
      {
        field_path: "b_false",
        value_state: "visible" as const,
        before: false,
        after: true,
      },
      {
        field_path: "c_null",
        value_state: "visible" as const,
        before: null,
        after: "null",
      },
      {
        field_path: "d_object",
        value_state: "visible" as const,
        before: { nested: [0, false] },
        after: ["safe"],
      },
      {
        field_path: "e_zero",
        value_state: "visible" as const,
        before: 0,
        after: 1,
      },
      {
        field_path: "password",
        value_state: "redacted" as const,
        before: null,
        after: null,
      },
    ];
    fetchMock.mockImplementation(async () =>
      jsonResponse(
        auditEnvelope([
          auditEvent({
            action_code: "constructor",
            target_kind: "__proto__",
            changes,
          }),
        ]),
      ),
    );
    const user = userEvent.setup();
    const view = render(<AuditHarness />);
    await screen.findByText("constructor");
    await user.click(
      screen.getByRole("button", { name: /^Inspect constructor/ }),
    );
    const details = within(
      screen.getByRole("region", { name: /^Details for constructor/ }),
    );
    for (const text of ['""', "false", "null", '"null"', "0"])
      expect(details.getByText(text)).toBeTruthy();
    expect(details.getAllByText("Redacted")).toHaveLength(2);
    expect(view.container.querySelector("img")).toBeNull();
    expect(details.getByText(/"nested"/)).toBeTruthy();
  });

  it("retains page and inputs across route mounts but clears a retired lifetime", async () => {
    fetchMock.mockImplementation(async () => jsonResponse(envelope()));
    const view = render(<AuditHarness />);
    await screen.findByText("User created");
    fireEvent.change(screen.getByLabelText("Target ID"), {
      target: { value: "retained" },
    });
    view.rerender(<AuditHarness active={false} visible={false} />);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    view.rerender(<AuditHarness />);
    await waitFor(() => expect(fetchMock).toHaveBeenCalledTimes(2));
    expect((screen.getByLabelText("Target ID") as HTMLInputElement).value).toBe(
      "retained",
    );
    const old = deferred<Response>();
    fetchMock.mockReturnValueOnce(old.promise);
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    view.rerender(<AuditHarness lifetime={null} />);
    await act(async () => old.resolve(jsonResponse(envelope())));
    expect(screen.queryByText("User created")).toBeNull();
    expect(screen.queryByDisplayValue("retained")).toBeNull();
  });

  it("distinguishes unavailable empty and stale states with explicit retries", async () => {
    fetchMock
      .mockImplementationOnce(async () => errorResponse("internal_error", 500))
      .mockImplementation(async () => jsonResponse(auditEnvelope([])));
    const user = userEvent.setup();
    render(<AuditHarness />);
    await user.click(await screen.findByRole("button", { name: "Retry read" }));
    await screen.findByText("No deployment audit events are available.");
    fetchMock.mockImplementationOnce(async () => jsonResponse(envelope()));
    await user.click(screen.getByRole("button", { name: "Refresh" }));
    await screen.findByText("User created");
    fetchMock.mockImplementationOnce(async () =>
      errorResponse("internal_error", 500),
    );
    await user.click(screen.getByRole("button", { name: "Refresh" }));
    await screen.findByRole("button", { name: "Retry refresh" });
    expect(screen.getByText("User created")).toBeTruthy();
    expect(screen.getAllByRole("status")).toHaveLength(1);
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("survives Strict Mode cleanup and document visibility fencing", async () => {
    fetchMock.mockImplementation(async () => jsonResponse(envelope()));
    render(
      <StrictMode>
        <AuditHarness />
      </StrictMode>,
    );
    await screen.findByText("User created");
    const initialCount = fetchMock.mock.calls.length;
    const visibility = vi
      .spyOn(document, "visibilityState", "get")
      .mockReturnValue("hidden");
    act(() => document.dispatchEvent(new Event("visibilitychange")));
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    expect(fetchMock).toHaveBeenCalledTimes(initialCount);
    visibility.mockReturnValue("visible");
    act(() => document.dispatchEvent(new Event("visibilitychange")));
    await waitFor(() =>
      expect(fetchMock).toHaveBeenCalledTimes(initialCount + 1),
    );
    visibility.mockRestore();
  });
});
