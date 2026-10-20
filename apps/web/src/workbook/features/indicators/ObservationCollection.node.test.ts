import { expect, it, vi } from "vitest";
import type { WorkbookPortResult } from "../../ports/WorkbookPortResult";
import { ObservationCollection } from "./ObservationCollection";
import type { ObservationPage } from "./observationOperation";

type Item = { id: string; version: number; text: string };
const page = (
  items: Item[],
  nextCursor: string | null = null,
): WorkbookPortResult<ObservationPage<Item>> => ({
  kind: "accepted",
  value: { items, nextCursor, hasMore: nextCursor !== null },
});
it("Observation pages retain data retry the failed cursor and deduplicate identity not text", async () => {
  const one = { id: "one", version: 2, text: "same" },
    two = { id: "two", version: 1, text: "same" };
  const read = vi
    .fn()
    .mockResolvedValueOnce(page([one], "second"))
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "retry" },
    })
    .mockResolvedValueOnce(page([{ ...one, version: 1 }, two]));
  const pages = new ObservationCollection<Item>(
    read,
    (item) => item.id,
    (item) => item.version,
  );
  await pages.load();
  await pages.more();
  expect(pages.getSnapshot().items).toEqual([one]);
  await pages.retry();
  expect(read.mock.calls.map((call) => call[0])).toEqual([
    null,
    "second",
    "second",
  ]);
  expect(pages.getSnapshot().items).toEqual([one, two]);
});
it("Observation pages reject nonprogress and fence obsolete reads", async () => {
  const one = { id: "one", version: 1, text: "same" };
  const read = vi
    .fn()
    .mockResolvedValueOnce(page([one], "second"))
    .mockResolvedValueOnce(page([one], "third"));
  const pages = new ObservationCollection<Item>(
    read,
    (item) => item.id,
    (item) => item.version,
  );
  await pages.load();
  await pages.more();
  expect(pages.getSnapshot().restartRequired).toBe(true);
  let resolve!: (value: WorkbookPortResult<ObservationPage<Item>>) => void;
  read.mockImplementationOnce(
    () =>
      new Promise((done) => {
        resolve = done;
      }),
  );
  const loading = pages.restart();
  pages.dispose();
  resolve(page([]));
  await loading;
  expect(pages.getSnapshot().items).toEqual([one]);
});
