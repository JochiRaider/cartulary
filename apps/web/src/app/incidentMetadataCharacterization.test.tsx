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
  metadataDeferred,
  metadataEnvelope,
  metadataIncident,
  metadataIncidentId,
  metadataJSON,
} from "../testing/incidentMetadataTestSupport";
import { MetadataTestSurface } from "./incidentMetadataTestSurface";

const subject = (
  activeSection: "incident-fields" | "summary" | null = "incident-fields",
  incidentId = metadataIncidentId,
  role: "reviewer" | "viewer" = "reviewer",
) => (
  <MetadataTestSurface
    activeSection={activeSection}
    incidentId={incidentId}
    currentIncidentRole={role}
  />
);
function setup(closed = false) {
  const writes: {
    payload: unknown;
    response: ReturnType<typeof metadataDeferred<Response>>;
  }[] = [];
  let failure = false;
  const fetch = vi.fn((input: unknown, init?: RequestInit) => {
    if (init?.method === "PATCH") {
      const response = metadataDeferred<Response>();
      writes.push({ payload: JSON.parse(String(init.body)), response });
      return response.promise;
    }
    if (String(input).includes("workbook-preferences"))
      return Promise.resolve(
        metadataJSON({
          data: { default_sheet_ref: null, home_sheet_ref: null },
        }),
      );
    if (failure)
      return Promise.resolve(
        metadataJSON({ error: { code: "internal_error" } }, 500),
      );
    return Promise.resolve(
      metadataJSON(
        metadataEnvelope(
          metadataIncident({
            incident_id: String(input).split("/").at(-1) ?? metadataIncidentId,
            ...(closed
              ? { status: "closed", closed_at: "2026-08-02T00:00:00Z" }
              : {}),
          }),
        ),
      ),
    );
  });
  vi.stubGlobal("fetch", fetch);
  return {
    writes,
    view: render(subject()),
    failReads: () => {
      failure = true;
    },
  };
}
async function edit(value = "critical") {
  const input = await screen.findByLabelText("Severity");
  fireEvent.change(input, { target: { value } });
}
function save() {
  fireEvent.click(screen.getByRole("button", { name: /Save/ }));
}
async function settle(
  writes: ReturnType<typeof setup>["writes"],
  response: Response,
) {
  await waitFor(() => expect(writes).toHaveLength(1));
  await act(async () => writes[0]?.response.resolve(response));
}
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
describe("Incident metadata characterization", () => {
  it("sends only the changed field and draft base version", async () => {
    const { writes } = setup();
    await edit();
    save();
    await waitFor(() => expect(writes).toHaveLength(1));
    expect(writes[0]?.payload).toEqual({
      base_incident_version: 1,
      severity: "critical",
    });
  });
  it("admits repeated Save activation only once", async () => {
    const { writes } = setup();
    await edit();
    save();
    save();
    await act(async () => {});
    expect(writes).toHaveLength(1);
  });
  it("preserves newer typing when the submitted revision completes", async () => {
    const { writes } = setup();
    await edit();
    save();
    await edit("newer exact input  ");
    await settle(
      writes,
      metadataJSON(
        metadataEnvelope(
          metadataIncident({ severity: "critical", incident_version: 2 }),
        ),
      ),
    );
    expect((screen.getByLabelText("Severity") as HTMLInputElement).value).toBe(
      "newer exact input  ",
    );
  });
  it("retains exact input across section changes and their reads", async () => {
    const { view } = setup();
    await edit("  retained  ");
    view.rerender(subject("summary"));
    await act(async () => {});
    view.rerender(subject());
    await screen.findByLabelText("Severity");
    await act(async () => {});
    expect((screen.getByLabelText("Severity") as HTMLInputElement).value).toBe(
      "  retained  ",
    );
  });
  it("retains exact input across drawer closure", async () => {
    const { view } = setup();
    await edit("retained after close");
    view.rerender(subject(null));
    view.rerender(subject());
    expect(
      ((await screen.findByLabelText("Severity")) as HTMLInputElement).value,
    ).toBe("retained after close");
  });
  it("consumes acknowledgement despite failed subsequent reads", async () => {
    const { writes, failReads } = setup();
    await edit();
    save();
    failReads();
    await settle(
      writes,
      metadataJSON(
        metadataEnvelope(
          metadataIncident({ severity: "critical", incident_version: 2 }),
        ),
      ),
    );
    await screen.findByText("Saved promoted incident fields.");
    expect((screen.getByLabelText("Severity") as HTMLInputElement).value).toBe(
      "critical",
    );
  });
  it("preserves conflicted input and exposes explicit version review", async () => {
    const { writes } = setup();
    await edit("  conflicted  ");
    save();
    await settle(
      writes,
      metadataJSON({ error: { code: "incident_version_conflict" } }, 409),
    );
    expect((screen.getByLabelText("Severity") as HTMLInputElement).value).toBe(
      "  conflicted  ",
    );
    expect(
      screen.queryByRole("button", { name: "Use this version" }),
    ).not.toBeNull();
  });
  it("excludes an obsolete rejection from another incident", async () => {
    const { writes, view } = setup();
    await edit();
    save();
    await waitFor(() => expect(writes).toHaveLength(1));
    view.rerender(
      subject("incident-fields", "00000000-0000-4000-8000-000000001002"),
    );
    await act(async () => {});
    await settle(
      writes,
      metadataJSON({ error: { code: "invalid_incident_patch" } }, 400),
    );
    expect(screen.queryByText("invalid_incident_patch")).toBeNull();
  });
  it("explains closure while keeping authorized metadata visible", async () => {
    setup(true);
    await act(async () => {});
    expect(screen.queryByText(/incident is closed/i)).not.toBeNull();
    expect(screen.queryByText("Original description")).not.toBeNull();
  });
  it("explains insufficient role while keeping authorized values visible", async () => {
    const { view } = setup();
    await screen.findByLabelText("Severity");
    view.rerender(subject("incident-fields", metadataIncidentId, "viewer"));
    await act(async () => {});
    expect(screen.queryByText(/reviewer or admin/i)).not.toBeNull();
    expect(screen.queryByText("Original description")).not.toBeNull();
  });
});
