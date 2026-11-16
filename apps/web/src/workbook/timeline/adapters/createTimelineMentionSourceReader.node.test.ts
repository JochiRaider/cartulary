import { expect, it, vi } from "vitest";
import { timelineViewSchemaId } from "../../models/workbookSurfaceRegistry";
import { createTimelineMentionSourceReader } from "./createTimelineMentionSourceReader";

const execute = vi.hoisted(() => vi.fn());
vi.mock("../../adapters/workbookOperationExecutor", () => ({
  createWorkbookOperationExecutor: () => ({ execute }),
}));

const firstPage = {
  kind: "accepted",
  value: {
    data: {
      incident_id: "incident-review",
      view_schema_id: timelineViewSchemaId,
      rows: [],
    },
    meta: { paging: { limit: 100, has_more: true, next_cursor: "page-two" } },
  },
};

it("aborted Review source reading does not request another Timeline page", async () => {
  execute.mockReset();
  execute.mockResolvedValueOnce(firstPage).mockResolvedValueOnce({
    kind: "accepted",
    value: {
      data: {
        incident_id: "incident-review",
        view_schema_id: timelineViewSchemaId,
        rows: [],
      },
      meta: { paging: { limit: 100, has_more: false, next_cursor: null } },
    },
  });
  await expect(
    createTimelineMentionSourceReader({
      apiBase: undefined,
      incidentId: "incident-review",
    })("source-outside-first-page", new AbortController().signal),
  ).rejects.toThrow(
    "The mention source is no longer in the active Timeline. Open History for its changes.",
  );
  expect(execute).toHaveBeenCalledTimes(2);
  expect(execute.mock.calls[1]?.[0].request).toEqual({
    limit: 100,
    cursor_token: "page-two",
  });

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
  releaseFirst(firstPage);
  await expect(result).rejects.toMatchObject({ reason: "aborted" });
  await new Promise((resolve) => setTimeout(resolve, 0));
  expect(execute).toHaveBeenCalledOnce();
});
