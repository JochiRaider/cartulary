import { indicatorsViewSchemaId } from "@cartulary/view-contracts";
import { afterEach, describe, expect, it, vi } from "vitest";
import type { NetworkFlowIndicatorLinkResult } from "../services/networkFlowContractAdapter";
import {
  coreAtomicIPType,
  type IndicatorLinkTargetPage,
  queryCompatibleIndicatorTargets,
} from "../services/networkFlowIndicatorAdapter";
import { readyExtensionAvailability } from "../testing/extensionAvailabilityTestSupport";
import { IndicatorLinkTargetDiscovery } from "./IndicatorLinkTargetDiscovery";
import {
  type IndicatorLinkTransport,
  NetworkFlowIndicatorLinkController,
} from "./NetworkFlowIndicatorLinkController";
import { linkNetworkFlowIndicator } from "./networkFlowClient";
import {
  captureIndicatorLinkAttempt,
  type IndicatorLinkAuthority,
  IndicatorLinkWriteError,
  type NetworkFlowIndicatorLinkCandidate,
  validateIndicatorLinkReceipt,
} from "./networkFlowIndicatorLinkOperation";
import { networkFlowWorkspaceStatus } from "./networkFlowWorkspaceStatus";

const incidentId = "11111111-1111-4111-8111-111111111111";
const actorId = "22222222-2222-4222-8222-222222222222";
const indicatorId = "33333333-3333-4333-8333-333333333333";
const source = {
  network_flow_table_id: `nft_${"a".repeat(32)}`,
  network_flow_row_id: `nfr_${"b".repeat(64)}`,
  source_row_number: 1,
  mapping_fingerprint: "c".repeat(64),
};
const candidate: NetworkFlowIndicatorLinkCandidate = {
  candidateValue: "192.0.2.10",
  label: "Source endpoint",
  key: "source",
  selector: {
    kind: "row_field_value",
    network_flow_table_id: source.network_flow_table_id,
    network_flow_row_id: source.network_flow_row_id,
    field_key: "network_flow.src_ip",
  },
  sourceRefs: [source],
  sourceTableIds: [source.network_flow_table_id],
  sourceTableRefs: [],
};
const receipt = (): NetworkFlowIndicatorLinkResult => ({
  schema_id: "cartulary.network_flow_indicator_link_result.v1",
  duplicate: false,
  binding: {
    network_flow_indicator_binding_id: `nfb_${"d".repeat(32)}`,
    incident_id: incidentId,
    candidate_value: candidate.candidateValue,
    target_indicator_ref: {
      indicator_id: indicatorId,
      indicator_type: "ipv4_addr",
      value_kind: "atomic",
      normalized_value: candidate.candidateValue,
    },
    selector_kind: "row_field_value",
    source_row_refs: [source],
    source_row_refs_truncated: false,
    source_row_refs_total_count: 1,
    created_observation_refs: [],
    created_by_user_id: actorId,
    created_at: "2026-09-09T20:00:00Z",
  },
});
const authority = (): IndicatorLinkAuthority => ({
  incidentId,
  actorId,
  session: "session-1",
  sessionResolved: true,
  role: "editor",
  open: true,
  available: true,
  profileAvailable: true,
  availabilityTag: { epochId: "epoch", generation: 1n },
});
function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: unknown) => void;
  const promise = new Promise<T>((yes, no) => {
    resolve = yes;
    reject = no;
  });
  return { promise, resolve, reject };
}
async function flush() {
  for (let index = 0; index < 8; index++) await Promise.resolve();
}
async function setup(submit?: IndicatorLinkTransport["submit"]) {
  let current = authority();
  const pending = deferred<NetworkFlowIndicatorLinkResult>();
  const send = vi.fn(
    submit ??
      ((_attempt, _signal, dispatch) => {
        dispatch();
        return pending.promise;
      }),
  );
  const controller = new NetworkFlowIndicatorLinkController();
  controller.bind(
    {
      sourceLimit: async () => 16,
      submit: send,
      discoverTargets: async () => ({ items: [], nextCursor: null }),
    },
    () => current,
  );
  controller.activate();
  await flush();
  controller.openDraft(candidate);
  controller.editDraft({ confirmation: candidate.candidateValue });
  return {
    controller,
    pending,
    send,
    change: (patch: Partial<IndicatorLinkAuthority>, notify = true) => {
      current = { ...current, ...patch };
      if (notify) controller.revalidateAuthority();
    },
  };
}
afterEach(() => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  vi.unstubAllGlobals();
});

describe("Indicator link captured operation", () => {
  it("discovers one compatible visible page with stable IDs and opaque continuation", async () => {
    const envelope = (value: string) => ({
      data: {
        incident_id: incidentId,
        view_schema_id: indicatorsViewSchemaId,
        rows: [
          {
            record_id: indicatorId,
            row_version: 1,
            cells: {
              "indicator.indicator_type": { value: "ipv4_addr" },
              "indicator.value_kind": { value: "atomic" },
              "indicator.normalized_value": { value },
            },
          },
        ],
      },
      meta: {
        request_id: "req-indicator-discovery",
        query: { filters: [], sort: [] },
        paging: { limit: 100, has_more: true, next_cursor: "opaque-page-2" },
      },
    });
    const fetch = vi.fn().mockImplementation(
      async () =>
        new Response(JSON.stringify(envelope(candidate.candidateValue)), {
          status: 200,
          headers: { "content-type": "application/json" },
        }),
    );
    vi.stubGlobal("fetch", fetch);
    const signal = new AbortController().signal;
    const query = {
      incidentId,
      candidateValue: candidate.candidateValue,
      cursor: null,
    };
    expect(
      await queryCompatibleIndicatorTargets(undefined, query, signal),
    ).toEqual({
      items: [{ id: indicatorId, value: candidate.candidateValue }],
      nextCursor: "opaque-page-2",
    });
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(JSON.parse(String(fetch.mock.calls[0]?.[1]?.body))).toEqual({
      limit: 100,
      filters: [
        {
          field_key: "indicator.indicator_type",
          op: "eq",
          arg: { value: "ipv4_addr" },
        },
        {
          field_key: "indicator.value_kind",
          op: "eq",
          arg: { value: "atomic" },
        },
      ],
    });
    await queryCompatibleIndicatorTargets(
      undefined,
      { ...query, cursor: "opaque-page-2" },
      signal,
    );
    expect(JSON.parse(String(fetch.mock.calls[1]?.[1]?.body))).toEqual({
      limit: 100,
      cursor_token: "opaque-page-2",
    });
    fetch.mockImplementationOnce(
      async () =>
        new Response(JSON.stringify(envelope("192.0.2.11")), {
          status: 200,
          headers: { "content-type": "application/json" },
        }),
    );
    expect(
      (await queryCompatibleIndicatorTargets(undefined, query, signal)).items,
    ).toEqual([]);
  });
  it("fences discovery generations and bounds failed reads independently of writes", async () => {
    vi.useFakeTimers();
    const discovery = new IndicatorLinkTargetDiscovery();
    let context = {
      key: "one",
      incidentId,
      candidateValue: candidate.candidateValue,
    };
    const first = deferred<IndicatorLinkTargetPage>();
    const second = deferred<IndicatorLinkTargetPage>();
    const port = vi
      .fn()
      .mockReturnValueOnce(first.promise)
      .mockReturnValueOnce(second.promise)
      .mockResolvedValue({ items: [], nextCursor: null });
    discovery.bind(port, () => context);
    discovery.load();
    discovery.load();
    expect(port).toHaveBeenCalledTimes(1);
    context = { ...context, key: "new-session" };
    discovery.revalidate();
    discovery.load();
    first.resolve({
      items: [{ id: indicatorId, value: candidate.candidateValue }],
      nextCursor: "old",
    });
    await flush();
    expect(discovery.getSnapshot().items).toEqual([]);
    await vi.advanceTimersByTimeAsync(30_000);
    expect(discovery.getSnapshot().phase).toBe("failed");
    discovery.retry();
    await flush();
    expect(discovery.getSnapshot().phase).toBe("ready");
    second.resolve({
      items: [{ id: indicatorId, value: candidate.candidateValue }],
      nextCursor: "late",
    });
    await flush();
    expect(discovery.getSnapshot().items).toEqual([]);
    discovery.clear();
  });
  it("projects one status by precedence and expires success without discarding the receipt", async () => {
    vi.useFakeTimers();
    const { controller, pending } = await setup();
    controller.submit();
    await flush();
    const base = {
      importStatus: null,
      linkStatus: controller.getSnapshot().status,
      graphState: "loading",
      hasGraph: false,
      rejectedRows: 2,
    } as const;
    expect(networkFlowWorkspaceStatus(base)).toBe("link_pending");
    for (const importStatus of [
      "validation_failed",
      "mapping_required",
      "validating",
    ] as const)
      expect(networkFlowWorkspaceStatus({ ...base, importStatus })).toBe(
        importStatus,
      );
    pending.resolve(receipt());
    await flush();
    expect(
      networkFlowWorkspaceStatus({
        ...base,
        linkStatus: controller.getSnapshot().status,
      }),
    ).toBe("graph_pending");
    expect(
      networkFlowWorkspaceStatus({
        ...base,
        graphState: "ready",
        hasGraph: true,
        linkStatus: controller.getSnapshot().status,
      }),
    ).toBe("link_committed");
    controller.dismiss();
    controller.reopen();
    expect(controller.getSnapshot().status).toBe("link_committed");
    await vi.advanceTimersByTimeAsync(5000);
    expect(controller.getSnapshot().status).toBeNull();
    expect(controller.getSnapshot().settlement?.kind).toBe("confirmed");
    controller.done();
    expect(controller.getSnapshot().draft).toBeNull();
    controller.dispose();
  });
  it("admits once synchronously and freezes the exact request", async () => {
    const { controller, pending, send } = await setup();
    expect(controller.submit()).toBe(true);
    expect(controller.submit()).toBe(false);
    await flush();
    expect(send).toHaveBeenCalledTimes(1);
    const attempt = controller.getSnapshot().attempt;
    expect(attempt?.body).toBe(JSON.stringify(attempt?.request));
    expect(Object.isFrozen(attempt?.request.selector)).toBe(true);
    expect(attempt?.request.client_txn_id).toMatch(/^nf-indicator-link-/u);
    pending.resolve(receipt());
    await flush();
    expect(controller.getSnapshot().settlement?.kind).toBe("confirmed");
    expect(controller.getSnapshot().presentation).toBe("attempt");
    controller.dispose();
  });
  it("invalidates exact confirmation without silently normalizing or retargeting", async () => {
    const { controller, send } = await setup();
    controller.editDraft({ confirmation: `${candidate.candidateValue} ` });
    expect(controller.submit()).toBe(false);
    expect(controller.getSnapshot().draft?.feedback?.field).toBe(
      "confirmation",
    );
    controller.editDraft({ confirmation: candidate.candidateValue });
    controller.editDraft({
      targetMode: "existing_indicator",
      existingId: indicatorId,
    });
    expect(controller.getSnapshot().draft?.confirmation).toBe("");
    controller.editDraft({ confirmation: candidate.candidateValue });
    controller.onTargetChange({
      id: indicatorId,
      removed: false,
      clientTxnId: "other-request",
    });
    expect(controller.getSnapshot().draft?.confirmation).toBe("");
    expect(controller.getSnapshot().draft?.applicable).toBe(false);
    controller.openDraft(candidate);
    controller.editDraft({
      targetMode: "existing_indicator",
      existingId: indicatorId,
    });
    controller.onTargetChange({
      id: indicatorId,
      removed: true,
      clientTxnId: "remove-request",
    });
    expect(controller.getSnapshot().draft).toBeNull();
    controller.openDraft(candidate);
    controller.editDraft({ confirmation: candidate.candidateValue });
    controller.setSelectionContext("new query");
    expect(controller.submit()).toBe(false);
    expect(send).not.toHaveBeenCalled();
    controller.dispose();
  });
  it("fences actual queued dispatch after authority changes", async () => {
    const availability = readyExtensionAvailability(incidentId);
    const barrier = deferred<void>();
    const queue = availability.runProfileRequest(
      "network_flow_activity",
      "/api/v1/incidents/{incident_id}/network-flow",
      () => barrier.promise,
    );
    const fetch = vi.fn();
    vi.stubGlobal("fetch", fetch);
    const { controller, change } = await setup(
      (attempt, signal, authorizeDispatch) =>
        linkNetworkFlowIndicator({
          availability,
          attempt,
          signal,
          authorizeDispatch,
        }),
    );
    expect(controller.submit()).toBe(true);
    await flush();
    change({ role: "reviewer" }, false);
    barrier.resolve();
    await queue;
    await flush();
    expect(fetch).not.toHaveBeenCalled();
    expect(controller.getSnapshot().settlement?.kind).toBe("not_dispatched");
    controller.dispose();

    vi.useFakeTimers();
    let queuedDispatch = () => {};
    const queued = await setup((_attempt, _signal, dispatch) => {
      queuedDispatch = dispatch;
      return new Promise(() => {});
    });
    queued.controller.submit();
    await flush();
    await vi.advanceTimersByTimeAsync(30_000);
    expect(queued.controller.getSnapshot().settlement?.kind).toBe(
      "not_dispatched",
    );
    expect(queuedDispatch).toThrow();
    queued.controller.dispose();
  });
  it("bounds observation and explicitly replays unchanged bytes while failed replay preserves uncertainty", async () => {
    vi.useFakeTimers();
    const { controller, send } = await setup();
    controller.submit();
    await flush();
    const attempt = controller.getSnapshot().attempt;
    await vi.advanceTimersByTimeAsync(30_000);
    expect(controller.getSnapshot().settlement?.kind).toBe("uncertain");
    controller.dismiss();
    expect(controller.getSnapshot().attempt).toBe(attempt);
    controller.reopen();
    send.mockImplementationOnce((_attempt, _signal, dispatch) => {
      dispatch();
      return Promise.reject(
        new IndicatorLinkWriteError("rejected", {
          kind: "denied",
          field: "target",
          message: "Target not visible.",
        }),
      );
    });
    expect(controller.replay()).toBe(true);
    expect(controller.replay()).toBe(false);
    await flush();
    expect(send.mock.calls[1]?.[0]).toBe(attempt);
    expect(controller.getSnapshot().settlement?.kind).toBe("uncertain");
    expect(controller.getSnapshot().attempt?.body).toBe(attempt?.body);
    controller.dispose();
  });
  it("accepts a late current receipt without closing or replacing a newer draft", async () => {
    vi.useFakeTimers();
    const { controller, pending } = await setup();
    controller.submit();
    await flush();
    await vi.advanceTimersByTimeAsync(30_000);
    controller.openDraft({
      ...candidate,
      key: "other",
      candidateValue: "192.0.2.20",
    });
    controller.editDraft({ existingId: indicatorId });
    expect(controller.submit()).toBe(false);
    const draft = controller.getSnapshot().draft;
    pending.resolve(receipt());
    await flush();
    expect(controller.getSnapshot().settlement?.kind).toBe("confirmed");
    expect(controller.getSnapshot().presentation).toBe("draft");
    expect(controller.getSnapshot().draft).toBe(draft);
    expect(controller.getSnapshot().status).toBeNull();
    controller.dispose();

    const abandoned = await setup();
    abandoned.controller.submit();
    await flush();
    abandoned.controller.abandonRecovery();
    abandoned.controller.openDraft(candidate);
    abandoned.controller.editDraft({ confirmation: candidate.candidateValue });
    const next = deferred<NetworkFlowIndicatorLinkResult>();
    abandoned.send.mockImplementationOnce((_attempt, _signal, dispatch) => {
      dispatch();
      return next.promise;
    });
    expect(abandoned.controller.submit()).toBe(true);
    await flush();
    const newAttempt = abandoned.controller.getSnapshot().attempt;
    abandoned.pending.resolve(receipt());
    await flush();
    expect(abandoned.controller.getSnapshot().attempt).toBe(newAttempt);
    expect(abandoned.controller.getSnapshot().settlement?.kind).toBe("pending");
    next.resolve(receipt());
    await flush();
    expect(abandoned.controller.getSnapshot().settlement?.kind).toBe(
      "confirmed",
    );
    abandoned.controller.dispose();
  });
  it("pauses session recovery and purges different actors incidents claims and removed sources", async () => {
    const { controller, pending, change } = await setup();
    controller.submit();
    await flush();
    change({ sessionResolved: false, actorId: null });
    expect(controller.getSnapshot().hidden).toBe(true);
    expect(controller.getSnapshot().settlement?.kind).toBe("uncertain");
    change({ sessionResolved: true, actorId, session: "session-2" });
    pending.resolve(receipt());
    await flush();
    expect(controller.getSnapshot().settlement?.kind).toBe("uncertain");
    change({ actorId: indicatorId });
    expect(controller.getSnapshot().attempt).toBeNull();
    expect(controller.getSnapshot().draft).toBeNull();
    controller.openDraft(candidate);
    change({ profileAvailable: false });
    expect(controller.getSnapshot().draft).toBeNull();
    controller.dispose();
    const second = await setup();
    second.controller.submit();
    await flush();
    second.controller.onResourceChange({
      resourceKind: "network_flow_table",
      resourceId: source.network_flow_table_id,
      changeKind: "remove",
      reasonCode: "soft_deleted",
    });
    expect(second.controller.getSnapshot().attempt).toBeNull();
    expect(second.controller.getSnapshot().draft).toBeNull();
    second.controller.dispose();
  });
  it("withdraws writes synchronously while retaining copyable drafts and never automatically resubmits", async () => {
    const { controller, change, send } = await setup();
    controller.onResourceChange({
      resourceKind: "*",
      resourceId: "*",
      changeKind: "remove",
      reasonCode: "incident_closed",
    });
    expect(controller.submit()).toBe(false);
    expect(controller.getSnapshot().draft?.candidate).toEqual(candidate);
    change({ open: false });
    change({ open: true });
    await flush();
    expect(send).not.toHaveBeenCalled();
    expect(controller.getSnapshot().draft?.confirmation).toBe("");
    controller.dispose();
  });
  it("distinguishes reuse metadata from binding identity and validates source refs", () => {
    const attempt = captureIndicatorLinkAttempt({
      authority: authority(),
      candidate,
      sourceLimit: 16,
      selectionRevision: 1,
      draftRevision: 1,
      target: { mode: "create_indicator", indicator_type: "ipv4_addr" },
      confirmation: candidate.candidateValue,
    });
    const reused = receipt();
    reused.duplicate = true;
    reused.binding.selector_kind = "graph_vertex";
    reused.binding.created_by_user_id = indicatorId;
    reused.binding.source_row_refs_truncated = true;
    reused.binding.source_row_refs_total_count = 20;
    expect(validateIndicatorLinkReceipt(reused, 200, attempt)).toEqual(reused);
    expect(() => validateIndicatorLinkReceipt(receipt(), 200, attempt)).toThrow(
      IndicatorLinkWriteError,
    );
    const invalid = receipt();
    invalid.binding.source_row_refs = [source, source];
    expect(() => validateIndicatorLinkReceipt(invalid, 201, attempt)).toThrow(
      IndicatorLinkWriteError,
    );
    const wrong = receipt();
    wrong.binding.target_indicator_ref.normalized_value = "192.0.2.11";
    expect(() => validateIndicatorLinkReceipt(wrong, 201, attempt)).toThrow(
      IndicatorLinkWriteError,
    );
    const targetId = "a3333333-3333-4333-8333-333333333333";
    const existing = {
      ...attempt,
      request: {
        ...attempt.request,
        target: {
          mode: "existing_indicator" as const,
          indicator_id: targetId.toUpperCase(),
        },
      },
    };
    const canonicalReceipt = receipt();
    canonicalReceipt.binding.target_indicator_ref.indicator_id = targetId;
    expect(
      validateIndicatorLinkReceipt(canonicalReceipt, 201, existing),
    ).toEqual(canonicalReceipt);
  });
  it("derives exact IPv4 and IPv6 types from Core registry facts", () => {
    expect(coreAtomicIPType("192.0.2.1")).toBe("ipv4_addr");
    expect(coreAtomicIPType("2001:db8::1")).toBe("ipv6_addr");
    expect(coreAtomicIPType("::")).toBe("ipv6_addr");
    for (const value of [
      "192.000.2.1",
      "192.0.2.1 ",
      "2001:DB8::1",
      "2001:0db8::1",
      "::ffff:c000:201",
      "::ffff:192.0.2.1",
      "[2001:db8::1]",
      "fe80::1%eth0",
      "2001:db8::1/64",
    ])
      expect(coreAtomicIPType(value)).toBeNull();
  });
  it("uses captured transport bytes CSRF and a validated receipt", async () => {
    const attempt = captureIndicatorLinkAttempt({
      authority: authority(),
      candidate,
      sourceLimit: 16,
      selectionRevision: 1,
      draftRevision: 1,
      target: { mode: "create_indicator", indicator_type: "ipv4_addr" },
      confirmation: candidate.candidateValue,
    });
    vi.spyOn(document, "cookie", "get").mockReturnValue("cartulary_csrf=csrf");
    const fetch = vi.fn().mockResolvedValue(
      new Response(JSON.stringify({ data: receipt(), meta: {} }), {
        status: 201,
        headers: { "content-type": "application/json" },
      }),
    );
    vi.stubGlobal("fetch", fetch);
    const dispatch = vi.fn();
    const signal = new AbortController().signal;
    await expect(
      linkNetworkFlowIndicator({
        availability: readyExtensionAvailability(incidentId),
        attempt,
        signal,
        authorizeDispatch: dispatch,
      }),
    ).resolves.toEqual(receipt());
    expect(dispatch).toHaveBeenCalledOnce();
    const init = fetch.mock.calls[0]?.[1] as RequestInit;
    expect(init.body).toBe(attempt.body);
    expect(init.signal).toBe(signal);
    expect(new Headers(init.headers).get("X-CSRF-Token")).toBe("csrf");

    const rejection = {
      error: {
        code: "network_flow_invalid_indicator_target",
        status: 400,
        request_id: "req-link-rejection",
        message: "Invalid target",
        retryable: false,
        details: {
          reason_code: "target_type_mismatch",
          retry_action: "correct_request",
          selector_kind: "row_field_value",
          field_key: "network_flow.src_ip",
          target_mode: "create_indicator",
          resolved_candidate_value: candidate.candidateValue,
        },
      },
    };
    const timeout = {
      error: {
        ...rejection.error,
        status: 503,
        code: "service_unavailable",
        retryable: true,
        details: {
          reason_code: "extension_transaction_timeout",
          operation_id: `network-flow-indicator-link:${incidentId}:${attempt.request.client_txn_id}`,
          timeout_seconds: 30,
        },
      },
    };
    for (const [status, payload, certainty] of [
      [400, rejection, "rejected"],
      [
        400,
        {
          error: {
            ...rejection.error,
            details: { reason_code: "target_type_mismatch" },
          },
        },
        "uncertain",
      ],
      [
        400,
        {
          error: {
            ...rejection.error,
            details: {
              ...rejection.error.details,
              reason_code: "unregistered",
            },
          },
        },
        "uncertain",
      ],
      [503, timeout, "rejected"],
      [
        503,
        {
          error: {
            ...timeout.error,
            details: {
              ...timeout.error.details,
              operation_id: "other-attempt",
            },
          },
        },
        "uncertain",
      ],
      [200, { data: receipt(), meta: {} }, "uncertain"],
    ] as const) {
      fetch.mockResolvedValueOnce(
        new Response(JSON.stringify(payload), {
          status,
          headers: { "content-type": "application/json" },
        }),
      );
      await expect(
        linkNetworkFlowIndicator({
          availability: readyExtensionAvailability(incidentId),
          attempt,
          signal,
          authorizeDispatch: dispatch,
        }),
      ).rejects.toMatchObject({ certainty });
    }
    fetch.mockResolvedValueOnce(
      new Response(JSON.stringify(rejection), {
        status: 400,
        headers: { "content-type": "application/json" },
      }),
    );
    await expect(
      linkNetworkFlowIndicator({
        availability: readyExtensionAvailability(incidentId),
        attempt,
        signal,
        authorizeDispatch: dispatch,
      }),
    ).rejects.toMatchObject({
      certainty: "rejected",
      feedback: { kind: "validation", field: "target" },
    });
  });
});
