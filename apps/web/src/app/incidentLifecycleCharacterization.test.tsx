import { incidentAdministrationTestId } from "@cartulary/ui-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { LifecycleTestSurface } from "../testing/incidentLifecycleSurfaceTestSupport";
import {
  metadataDeferred as deferred,
  metadataEnvelope as envelope,
  metadataIncident as incident,
  metadataIncidentId as incidentId,
  metadataJSON as json,
} from "../testing/incidentMetadataTestSupport";

const closed = () =>
  incident({
    status: "closed",
    closed_at: "2026-08-02T00:00:00Z",
    incident_version: 2,
  });
const id = (part: Parameters<typeof incidentAdministrationTestId>[0]) =>
  screen.getByTestId(incidentAdministrationTestId(part));
function setup() {
  const writes: {
    payload: Record<string, unknown>;
    response: ReturnType<typeof deferred<Response>>;
  }[] = [];
  let resource = incident();
  const sessionRead = deferred<void>();
  vi.stubGlobal(
    "fetch",
    vi.fn((url: unknown, init?: RequestInit) => {
      if (init?.method === "POST") {
        const response = deferred<Response>();
        writes.push({ payload: JSON.parse(String(init.body)), response });
        return response.promise;
      }
      if (String(url).includes("workbook-preferences"))
        return Promise.resolve(
          json({ data: { default_sheet_ref: null, home_sheet_ref: null } }),
        );
      return Promise.resolve(
        json(
          envelope({
            ...resource,
            incident_id: String(url).split("/").at(-1) ?? incidentId,
          }),
        ),
      );
    }),
  );
  const surface = (
    section: "summary" | "memberships" | null = "summary",
    subjectId = incidentId,
    acceptedIncident?: ReturnType<typeof incident>,
    waitForSession = false,
  ) => (
    <LifecycleTestSurface
      activeSection={section}
      currentIncidentRole="admin"
      incidentId={subjectId}
      acceptedIncident={acceptedIncident}
      onSessionRoleChange={
        waitForSession ? () => sessionRead.promise : undefined
      }
    />
  );
  return {
    writes,
    surface,
    view: render(surface()),
    observe: (next: ReturnType<typeof incident>) => {
      resource = next;
    },
    sessionRead,
  };
}
async function reason(value = "Initial reason") {
  await waitFor(() =>
    expect(id("summary-version").textContent).not.toContain("?"),
  );
  await waitFor(() =>
    expect(id("lifecycle-reason").hasAttribute("readonly")).toBe(false),
  );
  if (id("lifecycle-outcome").textContent === "")
    await waitFor(() =>
      expect(id("close-button").getAttribute("aria-disabled")).toBe("false"),
    );
  fireEvent.change(id("lifecycle-reason"), { target: { value } });
}
function close() {
  fireEvent.click(id("close-button"));
  const confirm = screen.queryByTestId(
    incidentAdministrationTestId("lifecycle-confirm"),
  );
  if (confirm) fireEvent.click(confirm);
}
async function settle(
  writes: ReturnType<typeof setup>["writes"],
  response: Response,
) {
  await waitFor(() => expect(writes.length).toBeGreaterThan(0));
  await act(async () => writes[0]?.response.resolve(response));
}
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
describe("Incident lifecycle characterization", () => {
  it("admits same-tick Close activation once", async () => {
    const { writes } = setup();
    await reason();
    fireEvent.click(id("close-button"));
    act(() => {
      fireEvent.click(id("lifecycle-confirm"));
      fireEvent.click(id("lifecycle-confirm"));
    });
    await act(async () => {});
    expect(writes).toHaveLength(1);
  });
  it("keeps the transaction and captured reason after an uncertain response", async () => {
    const { writes } = setup();
    await reason();
    close();
    await settle(
      writes,
      json({ error: { code: "invalid_public_contract_response" } }, 502),
    );
    await reason("Newer input");
    fireEvent.click(id("lifecycle-replay"));
    await waitFor(() => expect(writes).toHaveLength(2));
    expect(writes[1]?.payload).toEqual(writes[0]?.payload);
  });
  it("distinguishes transaction conflict from changed incident recovery", async () => {
    const { writes } = setup();
    await reason();
    close();
    await settle(writes, json({ error: { code: "client_txn_conflict" } }, 409));
    expect(screen.queryByText(/transaction.*conflict/i)).not.toBeNull();
    expect(
      screen.queryByText(
        "Incident changed; refreshed current state. Review and retry.",
      ),
    ).toBeNull();
  });
  it("explains an illegal transition using its current-state mismatch", async () => {
    const { writes } = setup();
    await reason();
    close();
    await settle(
      writes,
      json(
        {
          error: {
            code: "illegal_transition",
            details: { reason_code: "incident_already_closed" },
          },
        },
        409,
      ),
    );
    expect(screen.queryByText(/already closed/i)).not.toBeNull();
  });
  it("requires renewed review after an incident version conflict", async () => {
    const { writes, observe } = setup();
    await reason();
    close();
    observe(incident({ incident_version: 2 }));
    await settle(
      writes,
      json({ error: { code: "incident_version_conflict" } }, 409),
    );
    expect(
      screen.queryByRole("button", { name: "Review current incident" }),
    ).not.toBeNull();
  });
  it("preserves newer reason input after delayed confirmation", async () => {
    const { writes } = setup();
    await reason();
    close();
    await reason("  newer exact input\nsecond line  ");
    await settle(writes, json(envelope(closed())));
    expect((id("lifecycle-reason") as HTMLInputElement).value).toBe(
      "  newer exact input\nsecond line  ",
    );
  });
  it("excludes a delayed rejection from another incident", async () => {
    const { writes, view, surface } = setup();
    await reason();
    close();
    await waitFor(() => expect(writes).toHaveLength(1));
    view.rerender(surface("summary", "00000000-0000-4000-8000-000000001002"));
    await act(async () => {});
    await settle(
      writes,
      json({ error: { code: "invalid_incident_lifecycle_request" } }, 400),
    );
    expect(screen.queryByText("invalid_incident_lifecycle_request")).toBeNull();
  });
  it("acknowledges success before subsequent authorization observation settles", async () => {
    const { writes, view, surface } = setup();
    view.rerender(surface("summary", incidentId, undefined, true));
    await reason();
    close();
    await settle(writes, json(envelope(closed())));
    expect(screen.queryByText(/Close confirmed/i)).not.toBeNull();
  });
  it("does not replace a newer accepted incident with an old action receipt", async () => {
    const { writes, view, surface } = setup();
    await reason();
    close();
    await waitFor(() => expect(writes).toHaveLength(1));
    const newer = incident({ incident_version: 3 });
    view.rerender(surface("summary", incidentId, newer));
    await act(async () => {});
    await settle(writes, json(envelope(closed())));
    expect(id("summary-version").textContent).toContain("3");
    expect(id("summary-status").textContent).toBe("active");
  });
  it("retains raw reason across ordinary drawer closure", async () => {
    const { view, surface } = setup();
    await reason("  retained\nreason  ");
    view.rerender(surface(null));
    view.rerender(surface());
    await reasonReadiness();
    expect((id("lifecycle-reason") as HTMLInputElement).value).toBe(
      "  retained\nreason  ",
    );
  });
  it("retains reason across section changes", async () => {
    const { view, surface } = setup();
    await reason("retained reason");
    view.rerender(surface("memberships"));
    view.rerender(surface());
    await reasonReadiness();
    expect((id("lifecycle-reason") as HTMLInputElement).value).toBe(
      "retained reason",
    );
  });
});
async function reasonReadiness() {
  await waitFor(() =>
    expect(id("summary-version").textContent).not.toContain("?"),
  );
}
