import { afterEach, describe, expect, it, vi } from "vitest";
import { ReferenceQueryBroker } from "./referenceQueryBroker";

const requirement = {
  requirementId: "parties",
  resourceId: "view:cartulary.view.parties.v1:rows",
  viewSchemaId: "cartulary.view.parties.v1",
} as const;

describe("ReferenceQueryBroker", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("deduplicates only in-flight incident/authorization queries and rejects obsolete consumers", async () => {
    let resolveResponse!: (response: Response) => void;
    let requestCount = 0;
    const fetchMock = vi.fn(() => {
      requestCount += 1;
      if (requestCount > 1) {
        return Promise.resolve(
          new Response(
            JSON.stringify({
              data: {
                view_schema_id: requirement.viewSchemaId,
                rows: [],
              },
            }),
            {
              status: 200,
              headers: { "Content-Type": "application/json" },
            },
          ),
        );
      }
      return new Promise<Response>((resolve) => {
        resolveResponse = resolve;
      });
    });
    vi.stubGlobal("fetch", fetchMock);
    const broker = new ReferenceQueryBroker();
    const first = new AbortController();
    const second = new AbortController();
    const context = {
      incidentId: "incident-1",
      authorizationEpoch: "user-1",
    };
    const obsolete = broker.execute([requirement], context, first.signal);
    const current = broker.execute([requirement], context, second.signal);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    first.abort();
    resolveResponse(
      new Response(
        JSON.stringify({
          data: { view_schema_id: requirement.viewSchemaId, rows: [] },
        }),
        { status: 200, headers: { "Content-Type": "application/json" } },
      ),
    );
    await expect(obsolete).rejects.toMatchObject({ name: "AbortError" });
    await expect(current).resolves.toHaveLength(1);

    await broker.execute(
      [requirement],
      { ...context, incidentId: "incident-2" },
      new AbortController().signal,
    );
    expect(fetchMock).toHaveBeenCalledTimes(2);

    await broker.execute([requirement], context, new AbortController().signal);
    expect(fetchMock).toHaveBeenCalledTimes(3);

    await broker.execute(
      [requirement],
      { ...context, authorizationEpoch: "user-2" },
      new AbortController().signal,
    );
    expect(fetchMock).toHaveBeenCalledTimes(4);
  });
});
