import { expect, it, vi } from "vitest";
import { indicatorLifecycleConstraints } from "../../adapters/indicatorLifecycleProtocol";
import { IndicatorLifecycleDraftStore } from "./IndicatorLifecycleDraftStore";
import {
  emptyLifecycleValues,
  lifecycleReadTimestamp,
  lifecycleUTCTimestamp,
  validateLifecycleDraft,
} from "./indicatorLifecycleModel";
import { IndicatorLifecyclePaging } from "./indicatorLifecyclePaging";

it("Indicator lifecycle validates UTC dates, equal bounds, nullable fields and exact constraints", () => {
  expect(lifecycleReadTimestamp("2026-09-11T08:00:00.123456789-04:00")).toBe(
    "2026-09-11T12:00:00.123456789Z",
  );
  expect(lifecycleReadTimestamp("2026-01-01T00:00:00+02:30")).toBe(
    "2025-12-31T21:30:00Z",
  );
  expect(lifecycleReadTimestamp("2026-02-30T12:00:00-04:00")).toBeNull();
  expect(lifecycleReadTimestamp("2026-09-11T12:00:00+25:00")).toBeNull();
  expect(lifecycleUTCTimestamp("2026-09-11T12:30")).toBe(
    "2026-09-11T12:30:00Z",
  );
  expect(lifecycleUTCTimestamp("2024-02-29T12:30:01.120000")).toBe(
    "2024-02-29T12:30:01.12Z",
  );
  for (const raw of [
    "",
    "2025-02-29T12:00",
    "2026-04-31T12:00",
    "2026-13-01T12:00",
    "2026-09-11T24:00",
    "2026-09-11T12:60",
    "2026-09-11T12:00:60",
    "2026-09-11T12:00:00+02:00",
    "garbage",
  ])
    expect(lifecycleUTCTimestamp(raw), raw).toBeNull();
  const values = { ...emptyLifecycleValues, validFrom: "2026-09-11T12:30" };
  expect(validateLifecycleDraft(values).values).toMatchObject({
    valid_from: "2026-09-11T12:30:00Z",
    valid_to: null,
    confidence: null,
    rationale: null,
    assessor: null,
    support_refs: [],
  });
  expect(
    validateLifecycleDraft({ ...values, validTo: values.validFrom }).errors,
  ).toEqual([]);
  expect(
    validateLifecycleDraft({ ...values, validTo: "2026-09-11T12:29" }).errors[0]
      ?.field,
  ).toBe("validTo");
  for (const confidence of ["0", "100"])
    expect(
      validateLifecycleDraft({ ...values, confidence }).values?.confidence,
    ).toBe(Number(confidence));
  for (const confidence of ["-1", "101", "0.5", "NaN", " ", "1e1"])
    expect(
      validateLifecycleDraft({ ...values, confidence }).errors[0]?.field,
    ).toBe("confidence");
  for (const state of indicatorLifecycleConstraints.states)
    expect(
      validateLifecycleDraft({ ...values, state }).values?.lifecycle_state,
    ).toBe(state);
  for (const state of ["ACTIVE", " active", "false-positive", "unknown"])
    expect(validateLifecycleDraft({ ...values, state }).values).toBeNull();
  expect(
    validateLifecycleDraft({
      ...values,
      assessor: " analyst ",
      rationale: "  context\n",
    }).values,
  ).toMatchObject({ assessor: " analyst ", rationale: "  context\n" });
  expect(
    validateLifecycleDraft({ ...values, assessor: "\0" }).values,
  ).toBeNull();
});

it("Indicator lifecycle rejects malformed and duplicate support without reordering authored references", () => {
  const first = {
      recordId: "00000000-0000-4000-8000-000000000002",
      label: "Second",
      viewSchemaId: "view",
    },
    second = { ...first, recordId: "00000000-0000-4000-8000-000000000001" };
  const values = {
    ...emptyLifecycleValues,
    validFrom: "2026-09-11T12:30",
    support: [first, second],
  };
  expect(validateLifecycleDraft(values).values?.support_refs).toEqual([
    first.recordId,
    second.recordId,
  ]);
  expect(
    validateLifecycleDraft({ ...values, support: [first, first] }).values,
  ).toBeNull();
  expect(
    validateLifecycleDraft({
      ...values,
      support: [{ ...first, recordId: "foreign-format" }],
    }).values,
  ).toBeNull();
});

it("Indicator lifecycle drafts retain raw values independently for each subject", () => {
  const store = new IndicatorLifecycleDraftStore(vi.fn());
  store.open("a", "First", 1);
  store.update("a", {
    ...emptyLifecycleValues,
    validFrom: "invalid retained",
    confidence: "0",
  });
  store.open("b", "Second", 9);
  expect(store.open("a", "First", 2).baseRowVersion).toBe(1);
  expect(store.get("a")?.values.validFrom).toBe("invalid retained");
  expect(store.get("b")?.values.validFrom).toBe("");
  store.review("a", 2);
  expect(store.get("a")?.baseRowVersion).toBe(2);
  store.discard("b");
  expect(store.get("a")?.values.confidence).toBe("0");
  store.clear();
  expect(store.get("a")).toBeNull();
});

it("Indicator lifecycle pages retain results, retry the failed cursor and rebuild loaded pages", async () => {
  const first = { id: "newer", detail: "same" },
    second = { id: "older" };
  const read = vi
    .fn()
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { items: [first], hasMore: true, nextCursor: "opaque" },
    })
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Read failed" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [{ detail: "same", id: "newer" }, second],
        hasMore: false,
        nextCursor: null,
      },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { items: [first], hasMore: true, nextCursor: "new-opaque" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { items: [second], hasMore: false, nextCursor: null },
    });
  const pages = new IndicatorLifecyclePaging<{ id: string }>(
    read,
    (item) => item.id,
  );
  await pages.load();
  await pages.more();
  expect(pages.getSnapshot()).toMatchObject({
    phase: "failed",
    items: [first],
    hasMore: true,
  });
  await pages.retry();
  expect(read.mock.calls[2]?.[0]).toBe("opaque");
  expect(pages.getSnapshot().items).toEqual([first, second]);
  await pages.refresh();
  expect(read.mock.calls.slice(3).map((call) => call[0])).toEqual([
    null,
    "new-opaque",
  ]);
  expect(pages.getSnapshot().items).toEqual([first, second]);
});

it("Indicator lifecycle rejects invalid continuation and ignores late pages after restart", async () => {
  let release!: (value: unknown) => void;
  const read = vi
    .fn()
    .mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          release = resolve;
        }),
    )
    .mockResolvedValueOnce({
      kind: "accepted",
      value: { items: [{ id: "current" }], hasMore: true, nextCursor: "next" },
    })
    .mockResolvedValueOnce({
      kind: "accepted",
      value: {
        items: [{ id: "current", changed: true }],
        hasMore: false,
        nextCursor: null,
      },
    });
  const pages = new IndicatorLifecyclePaging<{ id: string }>(
    read,
    (item) => item.id,
  );
  const old = pages.load();
  await pages.restart();
  release({
    kind: "accepted",
    value: { items: [{ id: "late" }], hasMore: false, nextCursor: null },
  });
  await old;
  expect(pages.getSnapshot().items).toEqual([{ id: "current" }]);
  await pages.more();
  expect(pages.getSnapshot()).toMatchObject({
    phase: "failed",
    restartRequired: true,
    items: [{ id: "current" }],
  });
});
