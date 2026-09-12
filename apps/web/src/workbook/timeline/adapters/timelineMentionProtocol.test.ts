import { afterEach, expect, it, vi } from "vitest";
import {
  mentionReceipt,
  mentionReview,
} from "../../../testing/timelineMentionTestSupport";
import {
  errorEnvelope,
  successEnvelope,
} from "../../../testing/timelineWorkbookTestSupport";
import { createTimelineMentionResolutionAdapter } from "./createTimelineMentionResolutionAdapter";

afterEach(() => vi.unstubAllGlobals());
it("Mention transport preserves complete receipts raw text and opaque selector separation", async () => {
  const adapter = createTimelineMentionResolutionAdapter({ apiBase: "/base" });
  for (const action of [
    "resolve_item",
    "dismiss_item",
    "revert_to_unresolved",
  ] as const) {
    const base = mentionReview();
    const review = mentionReview({
      subject: {
        ...base.subject,
        state: action === "revert_to_unresolved" ? "dismissed" : "unresolved",
      },
      intent: action === "resolve_item" ? base.intent : { action },
    });
    const receipt = mentionReceipt(review);
    const fetchMock = vi.fn(
      async (_url: RequestInfo | URL, _init?: RequestInit) =>
        successEnvelope(receipt),
    );
    vi.stubGlobal("fetch", fetchMock);
    const attempt = adapter.capture(review, "original-secure-key");
    expect(attempt.path).toBe(
      `/base/api/v1/entity-mentions/${review.subject.mentionId}/resolve`,
    );
    expect(JSON.parse(attempt.body)).toEqual({
      action,
      client_txn_id: "original-secure-key",
      base_mention_row_version: 2,
      ...(review.intent.action === "resolve_item"
        ? { resolved_record_id: review.intent.resolvedRecordId }
        : {}),
    });
    await expect(
      adapter.send(attempt, new AbortController().signal),
    ).resolves.toEqual({ kind: "accepted", receipt });
    expect(fetchMock.mock.calls[0]?.[1]?.body).toBe(attempt.body);
  }
});
it("Mention transport treats missing required nulls inconsistent metadata and malformed success as uncertain", async () => {
  const adapter = createTimelineMentionResolutionAdapter({
      apiBase: undefined,
    }),
    review = mentionReview({ intent: { action: "dismiss_item" } });
  const receipt = mentionReceipt(review),
    attempt = adapter.capture(review, "original-key");
  const invalid: unknown[] = [];
  for (const key of Object.keys(receipt)) {
    const data = { ...receipt };
    delete (data as Record<string, unknown>)[key];
    invalid.push(data);
  }
  for (const key of Object.keys(receipt.entity_mention)) {
    const mention = { ...receipt.entity_mention };
    delete mention[key];
    invalid.push({ ...receipt, entity_mention: mention });
  }
  invalid.push(
    { ...receipt, active_link: null },
    { ...receipt, active_link: {} },
    { ...receipt, incident_id: "10000000-0000-4000-8000-000000000002" },
    { ...receipt, source_record: { ...receipt.source_record, row_version: 4 } },
    {
      ...receipt,
      entity_mention: {
        ...receipt.entity_mention,
        raw_text: review.subject.rawText.trim(),
      },
    },
    {
      ...receipt,
      entity_mention: {
        ...receipt.entity_mention,
        resolution_method: "auto_match",
      },
    },
    {
      ...receipt,
      entity_mention: { ...receipt.entity_mention, row_version: 9 },
    },
    {
      ...receipt,
      entity_mention: {
        ...receipt.entity_mention,
        source_record_id: review.subject.mentionId,
      },
    },
    {
      ...receipt,
      entity_mention: {
        ...receipt.entity_mention,
        resolution_status: "unresolved",
      },
    },
  );
  for (const data of invalid) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => successEnvelope(data)),
    );
    await expect(
      adapter.send(attempt, new AbortController().signal),
    ).resolves.toEqual({ kind: "uncertain" });
  }
  const resolved = mentionReceipt(),
    resolvedAttempt = adapter.capture(mentionReview(), "resolve-key");
  for (const data of [
    { ...resolved, active_link: undefined },
    {
      ...resolved,
      active_link: {
        ...resolved.active_link,
        link_type: "observed_as_identity",
      },
    },
    {
      ...resolved,
      active_link: {
        ...resolved.active_link,
        dst_record_id: review.subject.sourceRecordId,
      },
    },
    {
      ...resolved,
      entity_mention: {
        ...resolved.entity_mention,
        resolved_by_user_id: review.subject.mentionId,
      },
    },
    {
      ...resolved,
      entity_mention: { ...resolved.entity_mention, resolved_at: null },
    },
  ]) {
    vi.stubGlobal(
      "fetch",
      vi.fn(async () => successEnvelope(data)),
    );
    await expect(
      adapter.send(resolvedAttempt, new AbortController().signal),
    ).resolves.toEqual({ kind: "uncertain" });
  }
});
it("Mention transport separates confirmed rejection from lost responses and server ambiguity", async () => {
  const adapter = createTimelineMentionResolutionAdapter({
      apiBase: undefined,
    }),
    attempt = adapter.capture(mentionReview(), "exact-key");
  for (const response of [
    async () => {
      throw new TypeError("Lost response");
    },
    async () => new Response("invalid success", { status: 200 }),
    async () => errorEnvelope("internal_error", 500),
  ]) {
    vi.stubGlobal("fetch", vi.fn(response));
    await expect(
      adapter.send(attempt, new AbortController().signal),
    ).resolves.toEqual({ kind: "uncertain" });
  }
  vi.stubGlobal(
    "fetch",
    vi.fn(async () => errorEnvelope("row_version_conflict", 409)),
  );
  await expect(
    adapter.send(attempt, new AbortController().signal),
  ).resolves.toMatchObject({
    kind: "rejected",
    failure: { publicCode: "row_version_conflict" },
  });
});
