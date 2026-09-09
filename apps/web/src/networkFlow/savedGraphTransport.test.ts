import { afterEach, describe, expect, it, vi } from "vitest";
import {
  decodeNetworkFlowSavedGraphAccepted,
  normalizeSavedGraphDisplayName,
} from "../services/networkFlowContractAdapter";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { submitNetworkFlowSavedGraphMutation } from "./networkFlowClient";
import {
  captureSavedGraphIntent,
  prepareSavedGraphAttempt,
  type SavedGraphAction,
} from "./savedGraphOperation";
import {
  savedGraphAccepted,
  savedGraphFixture,
  savedGraphTestAuthority,
} from "./savedGraphTestFixtures";

afterEach(() => {
  vi.unstubAllGlobals();
});
function setup(kind: SavedGraphAction = "create") {
  const graph = savedGraphFixture();
  const intent = captureSavedGraphIntent(
    kind,
    savedGraphTestAuthority,
    kind === "create" ? null : graph,
    graph.semantic_query,
  );
  const attempt = prepareSavedGraphAttempt(
    intent,
    "Graph A",
    () => "txn-fixed",
  );
  const fetch = vi.fn();
  vi.stubGlobal("fetch", fetch);
  const send = () =>
    submitNetworkFlowSavedGraphMutation({
      availability: readyExtensionAvailability(graph.incident_id),
      attempt,
      signal: new AbortController().signal,
    });
  return { fetch, send, graph, attempt };
}
function acceptedResponse() {
  const receipt = savedGraphAccepted();
  if (receipt.kind !== "accepted") throw new Error();
  return receipt.value;
}
const response = (data: unknown, status = 202) =>
  new Response(JSON.stringify({ data }), {
    status,
    headers: { "Content-Type": "application/json" },
  });
describe("Saved graph transport integrity", () => {
  it("sends identical captured bytes on replay and accepts only the scoped narrow job receipt", async () => {
    const { fetch, send, attempt } = setup();
    fetch.mockImplementation(() =>
      Promise.resolve(response(acceptedResponse())),
    );
    decodeNetworkFlowSavedGraphAccepted(acceptedResponse());
    await expect(send()).resolves.toMatchObject({ kind: "accepted" });
    await send();
    expect(fetch.mock.calls.map((call) => call[1].body)).toEqual([
      attempt.body,
      attempt.body,
    ]);
    const wrong = acceptedResponse();
    wrong.job.status_route = "https://other.example/jobs/1";
    fetch.mockResolvedValueOnce(response(wrong));
    await expect(send()).rejects.toMatchObject({
      category: "invalid_response",
      certainty: "uncertain",
    });
  });
  it("requires empty 204 retirement on initial response and replay", async () => {
    const { fetch, send } = setup("retire");
    fetch.mockImplementation(() =>
      Promise.resolve(new Response(null, { status: 204 })),
    );
    await expect(send()).resolves.toEqual({ kind: "retired" });
    await expect(send()).resolves.toEqual({ kind: "retired" });
    fetch.mockResolvedValueOnce(response({}, 200));
    await expect(send()).rejects.toMatchObject({
      category: "invalid_response",
      certainty: "uncertain",
    });
  });
  it("preserves typed rejection and uncertain transport outcomes", async () => {
    const { fetch, send } = setup("refresh");
    fetch.mockResolvedValueOnce(
      new Response(
        JSON.stringify({
          error: {
            code: "network_flow_graph_view_version_conflict",
            message: "Conflict",
            status: 409,
            details: {},
          },
        }),
        { status: 409, headers: { "Content-Type": "application/json" } },
      ),
    );
    await expect(send()).rejects.toMatchObject({
      category: "version_conflict",
      certainty: "rejected",
      detail: { status: 409 },
    });
    fetch.mockRejectedValueOnce(new Error("lost response"));
    await expect(send()).rejects.toMatchObject({
      category: "transport",
      certainty: "uncertain",
    });
    fetch.mockResolvedValueOnce(response({}));
    await expect(send()).rejects.toMatchObject({
      category: "invalid_response",
      certainty: "uncertain",
    });
  });
  it("uses NFC and exact whitespace with UTF-8 byte limits and rejects controls before trimming", () => {
    expect(
      normalizeSavedGraphDisplayName(` ${"e\u0301".repeat(32)}\u3000`),
    ).toEqual({ ok: true, name: "é".repeat(32) });
    expect(normalizeSavedGraphDisplayName(`${"é".repeat(32)}a`)).toMatchObject({
      ok: false,
    });
    expect(normalizeSavedGraphDisplayName("\ufeffName\ufeff")).toEqual({
      ok: true,
      name: "\ufeffName\ufeff",
    });
    for (const name of ["\tName", "Name\u0085", "\ud800", "\u3000"])
      expect(normalizeSavedGraphDisplayName(name).ok).toBe(false);
    expect(normalizeSavedGraphDisplayName("😀".repeat(16)).ok).toBe(true);
  });
});
