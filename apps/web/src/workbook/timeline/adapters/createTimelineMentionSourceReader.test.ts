import { expect, it, vi } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createTimelineMentionSourceReader } from "./createTimelineMentionSourceReader";

const execute = vi.hoisted(() => vi.fn());
vi.mock("../../adapters/workbookOperationExecutor", () => ({
  createWorkbookOperationExecutor: () => ({ execute }),
}));

it("aborted Review source reading does not request another Timeline page", async () => {
  execute.mockReset();
  let releaseFirst!: (value: unknown) => void;
  execute.mockReturnValueOnce(
    new Promise((resolve) => {
      releaseFirst = resolve;
    }),
  );
  const controller = new AbortController();
  const reader = createTimelineMentionSourceReader({
    apiBase: undefined,
    incidentId: "incident-review",
  });
  const result = reader("source-outside-first-page", controller.signal);
  expect(execute).toHaveBeenCalledOnce();
  controller.abort();
  releaseFirst({
    kind: "accepted",
    value: {
      data: {
        incident_id: "incident-review",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      },
      meta: { paging: { limit: 100, has_more: true, next_cursor: "page-two" } },
    },
  });
  await expect(result).rejects.toMatchObject({ reason: "aborted" });
  await Promise.resolve();
  expect(execute).toHaveBeenCalledOnce();
});
