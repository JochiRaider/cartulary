import { afterEach, describe, expect, it, vi } from "vitest";
import { createReferenceQueryBroker } from "./referenceQueryBroker";

const requirement = {
  requirementId: "parties",
  resourceId: "view:cartulary.view.parties.v1:rows",
  viewSchemaId: "cartulary.view.parties.v1",
} as const;

type Deferred<T> = {
  readonly promise: Promise<T>;
  readonly resolve: (value: T) => void;
};

function deferred<T>(): Deferred<T> {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((resolvePromise) => {
    resolve = resolvePromise;
  });
  return { promise, resolve };
}

function queryResponse(): Response {
  return new Response(
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
  );
}

function brokerContext(
  overrides: {
    readonly authorizationGeneration?: string;
    readonly incidentId?: string;
  } = {},
) {
  return {
    authorizationGeneration:
      overrides.authorizationGeneration ?? "authorization-1",
    incidentId: overrides.incidentId ?? "incident-1",
  };
}

describe("ReferenceQueryBroker", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("deduplicates in-flight work while keeping shared-consumer cancellation local", async () => {
    const pending = deferred<Response>();
    const fetchMock = vi.fn(
      (_input: RequestInfo | URL, _init?: RequestInit) => pending.promise,
    );
    vi.stubGlobal("fetch", fetchMock);
    const broker = createReferenceQueryBroker(brokerContext());
    const first = new AbortController();
    const second = new AbortController();

    const obsolete = broker.execute([requirement], first.signal);
    const current = broker.execute([requirement], second.signal);
    expect(fetchMock).toHaveBeenCalledTimes(1);
    const underlyingSignal = (fetchMock.mock.calls[0]?.[1] as RequestInit)
      .signal;

    first.abort();
    expect(underlyingSignal?.aborted).toBe(false);
    pending.resolve(queryResponse());

    await expect(obsolete).rejects.toMatchObject({ name: "AbortError" });
    await expect(current).resolves.toHaveLength(1);
    broker.dispose();
  });

  it("keeps identical queries isolated between Workbook broker instances", async () => {
    const fetchMock = vi.fn((_input: RequestInfo | URL, _init?: RequestInit) =>
      Promise.resolve(queryResponse()),
    );
    vi.stubGlobal("fetch", fetchMock);
    const firstShell = createReferenceQueryBroker(brokerContext());
    const secondShell = createReferenceQueryBroker(brokerContext());

    await Promise.all([
      firstShell.execute([requirement], new AbortController().signal),
      secondShell.execute([requirement], new AbortController().signal),
    ]);

    expect(fetchMock).toHaveBeenCalledTimes(2);
    firstShell.dispose();
    secondShell.dispose();
  });

  it("binds incident and authorization generation at construction", async () => {
    const fetchMock = vi.fn((_input: RequestInfo | URL, _init?: RequestInit) =>
      Promise.resolve(queryResponse()),
    );
    vi.stubGlobal("fetch", fetchMock);
    const first = createReferenceQueryBroker(brokerContext());
    const incidentSwitch = createReferenceQueryBroker(
      brokerContext({ incidentId: "incident-2" }),
    );
    const authorizationChange = createReferenceQueryBroker(
      brokerContext({ authorizationGeneration: "authorization-2" }),
    );

    await first.execute([requirement], new AbortController().signal);
    await incidentSwitch.execute([requirement], new AbortController().signal);
    await authorizationChange.execute(
      [requirement],
      new AbortController().signal,
    );

    expect(fetchMock).toHaveBeenCalledTimes(3);
    expect(String(fetchMock.mock.calls[0]?.[0])).toContain(
      "/incidents/incident-1/",
    );
    expect(String(fetchMock.mock.calls[1]?.[0])).toContain(
      "/incidents/incident-2/",
    );
    first.dispose();
    incidentSwitch.dispose();
    authorizationChange.dispose();
  });

  it("aborts and rejects late work after typed invalidation or disposal", async () => {
    const pending = deferred<Response>();
    const fetchMock = vi.fn(
      (_input: RequestInfo | URL, _init?: RequestInit) => pending.promise,
    );
    vi.stubGlobal("fetch", fetchMock);
    const broker = createReferenceQueryBroker(brokerContext());
    const result = broker.execute([requirement], new AbortController().signal);
    const rejected = expect(result).rejects.toMatchObject({
      name: "AbortError",
    });
    const underlyingSignal = (fetchMock.mock.calls[0]?.[1] as RequestInit)
      .signal;

    broker.invalidate({ kind: "session_unavailable" });
    expect(underlyingSignal?.aborted).toBe(true);
    pending.resolve(queryResponse());
    await rejected;
    await expect(
      broker.execute([requirement], new AbortController().signal),
    ).rejects.toMatchObject({ name: "AbortError" });

    expect(() => {
      broker.dispose();
      broker.dispose();
    }).not.toThrow();
  });

  it("aborts underlying work after its final consumer is cancelled", async () => {
    const pending = deferred<Response>();
    const fetchMock = vi.fn(
      (_input: RequestInfo | URL, _init?: RequestInit) => pending.promise,
    );
    vi.stubGlobal("fetch", fetchMock);
    const broker = createReferenceQueryBroker(brokerContext());
    const consumer = new AbortController();
    const result = broker.execute([requirement], consumer.signal);
    const rejected = expect(result).rejects.toMatchObject({
      name: "AbortError",
    });
    const underlyingSignal = (fetchMock.mock.calls[0]?.[1] as RequestInit)
      .signal;

    consumer.abort();
    expect(underlyingSignal?.aborted).toBe(true);
    await rejected;
    broker.dispose();
  });
});
