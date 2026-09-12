import { afterEach, expect, it, vi } from "vitest";
import {
  errorResponse,
  jsonResponse,
} from "../../testing/fetchMockTestSupport";
import {
  canonicalCreateContract,
  canonicalCreateResponse,
  canonicalCreateValues,
} from "../../testing/indicatorCreateTestSupport";
import {
  observationAuthority,
  testObservation,
} from "../../testing/observationTestSupport";
import {
  indicatorCreateAvailable,
  indicatorCreateConstraints,
  indicatorCreateErrors,
  indicatorCreateRequest,
  indicatorCreateSeed,
  indicatorCreateTypes,
} from "../features/indicators/indicatorCreateModel";
import { createIndicatorCreateTransport } from "./createIndicatorCreateTransport";

afterEach(() => vi.unstubAllGlobals());
const port = () =>
  createIndicatorCreateTransport({
    apiBase: "https://original.test",
    incidentId: observationAuthority.incidentId,
  });
const capture = () =>
  port().capture(
    observationAuthority,
    1,
    testObservation,
    canonicalCreateContract,
    canonicalCreateValues,
    "original-secure-id",
  );
it("Canonical transport retains closed request bytes and complete accepted results with neutral status semantics", async () => {
  const attempt = capture(),
    response = canonicalCreateResponse(),
    fetch = vi.fn<typeof globalThis.fetch>();
  vi.stubGlobal("fetch", fetch);
  expect(JSON.parse(attempt.body)).toEqual({
    client_txn_id: "original-secure-id",
    ...canonicalCreateValues,
  });
  for (const status of [201, 200]) {
    fetch.mockResolvedValueOnce(jsonResponse(response, status));
    const result = await port().send(attempt, new AbortController().signal);
    expect(result.kind).toBe("accepted");
    if (result.kind === "accepted") {
      expect(result.receipt.response).toEqual(response);
      expect(result.receipt.status).toBe(status);
    }
  }
  expect(fetch.mock.calls.map(([url]) => url)).toEqual([
    attempt.path,
    attempt.path,
  ]);
  expect(fetch.mock.calls.map(([, init]) => init?.body)).toEqual([
    attempt.body,
    attempt.body,
  ]);
  expect(
    fetch.mock.calls.every(([, init]) => init?.credentials === "include"),
  ).toBe(true);
  for (const field of [
    "observation_id",
    "origin_locator",
    "resolved_indicator_record_id",
    "indicator.observation_count",
  ])
    expect(() =>
      port().capture(
        observationAuthority,
        1,
        testObservation,
        canonicalCreateContract,
        { ...canonicalCreateValues, [field]: "not admitted" },
        "txn",
      ),
    ).toThrow();
});
it("Canonical transport rejects invalid receipt bindings and classifies malformed or lost responses as uncertainty", async () => {
  const attempt = capture(),
    valid = canonicalCreateResponse();
  const invalid = [
    { ...valid, data: { ...valid.data, view_schema_id: "wrong" } },
    { ...valid, data: { ...valid.data, change_set_id: "" } },
    ...[
      { record_id: "bad" },
      { row_version: 0 },
      { cells: {} },
      {
        cells: {
          ...valid.data.row.cells,
          "indicator.indicator_type": { value: "text" },
        },
      },
    ].map((patch) => ({
      ...valid,
      data: { ...valid.data, row: { ...valid.data.row, ...patch } },
    })),
  ];
  for (const response of [
    ...invalid.map((value) => jsonResponse(value, 201)),
    jsonResponse(valid, 202),
    new Response("broken", { status: 201 }),
    new Response("broken", { status: 409 }),
    errorResponse("internal_error", 500),
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => response),
    );
    expect(
      (await port().send(attempt, new AbortController().signal)).kind,
    ).toBe("uncertain");
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => {
      throw new Error("response lost");
    }),
  );
  expect((await port().send(attempt, new AbortController().signal)).kind).toBe(
    "uncertain",
  );
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => errorResponse("invalid_mutation_payload", 400)),
  );
  expect((await port().send(attempt, new AbortController().signal)).kind).toBe(
    "rejected",
  );
});

it("Indicator canonical request constraints preserve exact vocabularies identity and omission semantics", () => {
  expect(indicatorCreateTypes).toEqual([
    "ipv4_addr",
    "ipv6_addr",
    "domain_name",
    "url",
    "sha256",
    "email_addr",
    "registry_key",
    "process_name",
    "text",
  ]);
  expect(indicatorCreateConstraints.valueKinds).toEqual([
    "atomic",
    "pattern",
    "reference",
  ]);
  const seed = indicatorCreateSeed(testObservation);
  expect(seed).toEqual({
    "indicator.indicator_type": "domain_name",
    "indicator.value_kind": "",
    "indicator.display_value": testObservation.normalized_candidate,
  });
  expect(
    indicatorCreateRequest(canonicalCreateContract, seed, "txn"),
  ).toBeNull();
  const valid = {
    ...seed,
    "indicator.value_kind": "atomic",
    "indicator.display_value": "EXAMPLE[.]COM",
    "indicator.stix_pattern": "",
  };
  expect(indicatorCreateRequest(canonicalCreateContract, valid, "txn")).toEqual(
    {
      client_txn_id: "txn",
      "indicator.indicator_type": "domain_name",
      "indicator.value_kind": "atomic",
      "indicator.display_value": "EXAMPLE[.]COM",
    },
  );
  for (const key of [
    "observation_id",
    "origin_kind",
    "resolved_indicator_record_id",
    "indicator.observation_count",
  ])
    expect(
      indicatorCreateRequest(
        canonicalCreateContract,
        { ...valid, [key]: "bad" },
        "txn",
      ),
    ).toBeNull();
  expect(
    indicatorCreateRequest(
      canonicalCreateContract,
      { ...valid, "indicator.indicator_type": "DOMAIN_NAME" },
      "txn",
    ),
  ).toBeNull();
  expect(
    indicatorCreateRequest(
      canonicalCreateContract,
      { ...valid, "indicator.value_kind": "hash" },
      "txn",
    ),
  ).toBeNull();
  expect(
    indicatorCreateRequest(
      canonicalCreateContract,
      { ...valid, "indicator.hash_algorithm": "sha256" },
      "txn",
    ),
  ).toBeNull();
  expect(
    indicatorCreateRequest(
      canonicalCreateContract,
      {
        ...valid,
        "indicator.hash_algorithm": "sha256",
        "indicator.hash_value": "aF01",
      },
      "txn",
    ),
  ).not.toBeNull();
  expect(
    indicatorCreateRequest(
      canonicalCreateContract,
      {
        ...valid,
        "indicator.indicator_type": "ipv4_addr",
        "indicator.hash_algorithm": "sha256",
        "indicator.hash_value": "aF01",
      },
      "txn",
    ),
  ).toBeNull();
  expect(
    indicatorCreateErrors(canonicalCreateContract, {
      ...valid,
      "indicator.indicator_type": "sha256",
    })["indicator.display_value"],
  ).toBeTruthy();
  expect(
    indicatorCreateAvailable({
      ...canonicalCreateContract,
      createCapable: false,
    }),
  ).toBe(false);
});
