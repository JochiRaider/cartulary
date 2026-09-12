import { assessmentCreateControlTestId } from "@cartulary/ui-contracts";
import {
  assessmentsViewSchemaId,
  requireViewContract,
} from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  renderHook,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createAssessmentCandidateReader } from "../../adapters/createAssessmentCandidateReader";
import {
  type AssessmentCreateDraft,
  initialAssessmentDraft,
} from "../../models/assessmentWorkbookModel";
import { emptyWorkbookQueryState } from "../../models/workbookQuery";
import {
  AssessmentSubjectPicker,
  AssessmentSupportPicker,
} from "./AssessmentDiscovery";
import type { AssessmentCandidateReadPort } from "./assessmentCandidatePort";
import { useAssessmentCandidates } from "./useAssessmentCandidates";

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
const query = emptyWorkbookQueryState();
const candidate = (recordId: string) => ({
  recordId,
  displayText: `Record ${recordId}`,
});
const page = (ids: string[], nextCursor: string | null = null) => ({
  kind: "accepted" as const,
  value: {
    candidates: ids.map(candidate),
    hasMore: nextCursor !== null,
    nextCursor,
  },
});

describe("Assessment discovery", () => {
  it("reads only the authorized selected view and reduces minimal Timeline rows with opaque paging", async () => {
    const fetch = vi.fn(
      async (_url: unknown, _init: RequestInit) =>
        new Response(
          JSON.stringify({
            data: {
              incident_id: "00000000-0000-4000-8000-000000000001",
              view_schema_id: "cartulary.view.timeline.v2",
              rows: [
                {
                  record_id: "00000000-0000-4000-8000-000000000102",
                  row_version: 1,
                  cells: {
                    "timeline.activity_synopsis_text": {
                      value: "Readable support",
                    },
                  },
                },
              ],
            },
            meta: {
              request_id: "query",
              query: { filters: [], sort: [] },
              paging: {
                limit: 100,
                has_more: true,
                next_cursor: "opaque-next",
              },
            },
          }),
          { status: 200, headers: { "Content-Type": "application/json" } },
        ),
    );
    vi.stubGlobal("fetch", fetch);
    const reader = createAssessmentCandidateReader({
      apiBase: undefined,
      incidentId: "00000000-0000-4000-8000-000000000001",
    });
    const result = await reader.support({
      queryState: query,
      cursor: "opaque-first",
      signal: new AbortController().signal,
    });
    expect(result).toEqual({
      kind: "accepted",
      value: {
        candidates: [
          {
            recordId: "00000000-0000-4000-8000-000000000102",
            displayText: "Readable support",
          },
        ],
        hasMore: true,
        nextCursor: "opaque-next",
      },
    });
    expect(fetch).toHaveBeenCalledTimes(1);
    expect(String(fetch.mock.calls[0]?.[0])).toContain(
      "cartulary.view.timeline.v2/query",
    );
    expect(JSON.parse(String(fetch.mock.calls[0]?.[1].body))).toMatchObject({
      limit: 100,
      cursor_token: "opaque-first",
    });
  });

  it("keeps failed pages distinct from empty results and retries the same cursor without losing prior candidates", async () => {
    const read = vi
      .fn()
      .mockResolvedValueOnce(page(["a"], "next"))
      .mockResolvedValueOnce({
        kind: "rejected",
        failure: { kind: "retryable", message: "Read failed" },
      })
      .mockResolvedValueOnce(page(["b"]));
    const { result } = renderHook(() =>
      useAssessmentCandidates(read, query, 0),
    );
    await waitFor(() => expect(result.current.phase).toBe("ready"));
    await act(() => result.current.loadMore());
    expect(result.current.phase).toBe("failed");
    expect(result.current.stale).toBe(true);
    expect(result.current.candidates).toEqual([candidate("a")]);
    await act(() => result.current.retry());
    expect(read.mock.calls[1]?.[0].cursor).toBe("next");
    expect(read.mock.calls[2]?.[0].cursor).toBe("next");
    expect(result.current.candidates).toEqual([candidate("a"), candidate("b")]);
  });

  it("rejects cyclic paging and ignores an obsolete filter response", async () => {
    let finish!: (value: ReturnType<typeof page>) => void;
    const read = vi
      .fn()
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            finish = resolve;
          }),
      )
      .mockResolvedValueOnce(page(["new"], "cycle"))
      .mockResolvedValueOnce(page(["ignored"], "cycle"));
    const { result, rerender } = renderHook(
      ({ revision }) => useAssessmentCandidates(read, query, revision),
      { initialProps: { revision: 0 } },
    );
    rerender({ revision: 1 });
    await waitFor(() => expect(result.current.phase).toBe("ready"));
    await act(async () => finish(page(["old"])));
    expect(result.current.candidates).toEqual([candidate("new")]);
    await act(() => result.current.loadMore());
    expect(result.current.phase).toBe("failed");
    expect(result.current.error).toContain("paging changed");
  });

  it("preserves an explicitly selected subject absent from refreshed candidate pages", async () => {
    const reader: AssessmentCandidateReadPort = {
      subjects: vi.fn(async () => page(["other"])),
      support: vi.fn(async () => page([])),
    };
    const update = vi.fn();
    const draft = {
      ...initialAssessmentDraft(requireViewContract(assessmentsViewSchemaId)),
      subjectRecordId: "selected",
      subjectDisplayText: "Selected subject",
    };
    render(
      <AssessmentSubjectPicker
        reader={reader}
        draft={draft}
        disabled={false}
        revision={0}
        update={update}
      />,
    );
    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toContain("All matching"),
    );
    expect(
      (
        screen.getByTestId(
          assessmentCreateControlTestId("subject"),
        ) as HTMLSelectElement
      ).value,
    ).toBe("selected");
    fireEvent.click(screen.getByRole("button", { name: "Reload candidates" }));
    await waitFor(() => expect(reader.subjects).toHaveBeenCalledTimes(2));
    expect(update).not.toHaveBeenCalled();
  });

  it("keeps subject intent across ordering filters paging and empty refreshes", async () => {
    const subjects = vi
      .fn()
      .mockResolvedValueOnce(page(["selected"], "next"))
      .mockResolvedValueOnce(page(["other"]))
      .mockResolvedValue(page([]));
    const reader: AssessmentCandidateReadPort = {
      subjects,
      support: vi.fn(async () => page([])),
    };
    const update = vi.fn();
    const draft = {
      ...initialAssessmentDraft(requireViewContract(assessmentsViewSchemaId)),
      subjectRecordId: "selected",
      subjectDisplayText: "Selected subject",
    };
    render(
      <AssessmentSubjectPicker
        reader={reader}
        draft={draft}
        disabled={false}
        revision={0}
        update={update}
      />,
    );
    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toContain(
        "More candidates",
      ),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Load more candidates" }),
    );
    await waitFor(() => expect(subjects).toHaveBeenCalledTimes(2));
    expect(subjects.mock.calls[1]?.[1].cursor).toBe("next");
    fireEvent.change(screen.getByLabelText("Subject order"), {
      target: { value: "host.display_name:desc" },
    });
    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toContain(
        "No candidates found",
      ),
    );
    expect(subjects.mock.calls[2]?.[1].queryState.sort).toEqual([
      { fieldKey: "host.display_name", direction: "desc" },
    ]);
    const filter = screen.getByLabelText("Subject filter State");
    fireEvent.change(filter, { target: { value: "selected-state" } });
    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toContain(
        "No candidates match",
      ),
    );
    fireEvent.click(screen.getByRole("button", { name: "Reload candidates" }));
    await waitFor(() => expect(subjects).toHaveBeenCalledTimes(5));
    expect(
      (
        screen.getByTestId(
          assessmentCreateControlTestId("subject"),
        ) as HTMLSelectElement
      ).value,
    ).toBe("selected");
    expect(update).not.toHaveBeenCalled();
  });

  it("stages support selection and cancels with Escape without changing the draft", async () => {
    const reader: AssessmentCandidateReadPort = {
      subjects: vi.fn(async () => page([])),
      support: vi.fn(async () => page(["a", "b"])),
    };
    const update = vi.fn();
    const draft: AssessmentCreateDraft = {
      ...initialAssessmentDraft(requireViewContract(assessmentsViewSchemaId)),
      supportRecordIds: ["a"],
    };
    render(
      <AssessmentSupportPicker
        reader={reader}
        draft={draft}
        disabled={false}
        revision={0}
        update={update}
      />,
    );
    const trigger = screen.getByRole("button", { name: "Choose support" });
    fireEvent.click(trigger);
    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toContain("All matching"),
    );
    const select = screen.getByTestId(
      assessmentCreateControlTestId("support-refs"),
    ) as HTMLSelectElement;
    const [first, second] = Array.from(select.options);
    if (!first || !second) throw new Error("Expected two support candidates.");
    first.selected = false;
    second.selected = true;
    fireEvent.change(select);
    fireEvent.keyDown(select, { key: "Escape" });
    expect(update).not.toHaveBeenCalled();
    expect(document.activeElement).toBe(trigger);
    fireEvent.click(trigger);
    await waitFor(() =>
      expect(screen.getByRole("status").textContent).toContain("All matching"),
    );
    fireEvent.click(
      screen.getByRole("button", { name: "Apply support selection" }),
    );
    expect(update.mock.calls[0]?.[0](draft).supportRecordIds).toEqual(["a"]);
  });
});
