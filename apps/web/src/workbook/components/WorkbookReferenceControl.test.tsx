import { genericEditValueTestId } from "@cartulary/ui-contracts";
import {
  getReferenceFieldContract,
  requireViewContract,
} from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import { afterEach, expect, it, vi } from "vitest";
import { taskAuthority } from "../../testing/taskWorkbookTestSupport";
import { prepareWorkbookInspectorChange } from "../inspector/prepareWorkbookInspectorChange";
import { useWorkbookInspectorEditDraft } from "../inspector/useWorkbookInspectorEditDraft";
import { WorkbookInspectorDraftStore } from "../inspector/WorkbookInspectorDraftStore";
import { WorkbookInspectorEditControl } from "../inspector/WorkbookInspectorEditControl";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import type {
  WorkbookReferencePage,
  WorkbookReferenceReadPort,
  WorkbookReferenceRequest,
} from "../ports/WorkbookReferenceReadPort";
import type { WorkbookCommittedRecordPort } from "../query/WorkbookCommittedRecordPort";
import { WorkbookReferenceContext } from "./WorkbookReferenceControl";

const view = "cartulary.view.task_requests.v1";
const contract = requireViewContract(view);
const id = (number: number) =>
  `00000000-0000-4000-8000-${String(number).padStart(12, "0")}`;
const authoritySnapshot = { authority: taskAuthority };
const evidence = {
  subscribe: () => () => {},
  getSnapshot: () => authoritySnapshot,
  latestRow: () => null,
  latestVersion: () => null,
  acceptRow: () => null,
};
afterEach(cleanup);
function accepted(
  input: WorkbookReferenceRequest,
  offset: number,
): WorkbookPortResult<WorkbookReferencePage> {
  return {
    kind: "accepted",
    value: {
      producingRequest: input,
      canonicalQuery: { filters: [], sort: [] },
      paging: {
        limit: 100,
        hasMore: offset === 0,
        nextCursor: offset === 0 ? "next-page" : null,
      },
      candidates: Array.from({ length: 100 }, (_, i) => ({
        identity: { kind: input.identityKind, id: id(offset + i + 1) },
        displayText: `Candidate ${offset + i + 1}`,
        presentation: "observed",
        viewSchemaId: input.viewSchemaId,
      })),
    },
  };
}
function Form({
  store,
  fieldKey,
  write,
  reader,
  recordEvidence = evidence,
  recordId = id(999),
}: {
  store: WorkbookInspectorDraftStore;
  fieldKey: string;
  write: (change: unknown) => void;
  reader: WorkbookReferenceReadPort;
  recordEvidence?: WorkbookCommittedRecordPort;
  recordId?: string;
}) {
  const field = contract.fieldMap[fieldKey];
  if (!field) throw new Error("Missing test field");
  const edit = useWorkbookInspectorEditDraft({
    store,
    row: {
      record_id: recordId,
      row_version: 1,
      cells: { [fieldKey]: { value: null } },
    },
    field,
    viewSchemaId: view,
    active: true,
    presentation: "test",
  });
  return (
    <WorkbookReferenceContext.Provider
      value={{
        reader,
        evidence: recordEvidence,
        onAuthorityFailure: () => store.setAuthority(null),
      }}
    >
      <WorkbookInspectorEditControl
        edit={edit}
        field={field}
        collectionMode="add"
        testId={genericEditValueTestId(view)}
      />
      <button
        type="button"
        disabled={!edit.canSubmit}
        onClick={() =>
          write(prepareWorkbookInspectorChange(field, edit.value, "add", view))
        }
      >
        Update parent
      </button>
    </WorkbookReferenceContext.Provider>
  );
}
it("stages collection choices across pages and eligible surfaces without editing or submitting the inspector until acceptance", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  const page = vi.fn(async (input: WorkbookReferenceRequest) =>
    accepted(input, input.cursorToken ? 100 : 0),
  );
  const reader = { page };
  render(
    <Form
      store={store}
      fieldKey="task.linked_record_ids"
      write={write}
      reader={reader}
    />,
  );
  fireEvent.change(screen.getByTestId(genericEditValueTestId(view)), {
    target: { value: id(800) },
  });
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const select = await screen.findByRole("listbox", {
    name: "Linked Records candidates",
  });
  const choose = (value: string) => {
    for (const option of (select as HTMLSelectElement).options)
      option.selected = option.value === `record:${value}`;
    fireEvent.change(select);
  };
  choose(id(1));
  fireEvent.click(screen.getByRole("button", { name: "Next" }));
  await screen.findByRole("option", { name: `Candidate 101 (${id(101)})` });
  choose(id(101));
  fireEvent.change(screen.getByLabelText("Reference surface"), {
    target: { value: "cartulary.view.indicators.v1" },
  });
  await waitFor(() =>
    expect(page.mock.lastCall?.[0].viewSchemaId).toBe(
      "cartulary.view.indicators.v1",
    ),
  );
  expect(
    (screen.getByTestId(genericEditValueTestId(view)) as HTMLTextAreaElement)
      .value,
  ).toBe(id(800));
  expect(write).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Cancel references" }));
  expect(
    (screen.getByTestId(genericEditValueTestId(view)) as HTMLTextAreaElement)
      .value,
  ).toBe(id(800));
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  await screen.findByRole("option", { name: `Candidate 1 (${id(1)})` });
  const current = screen.getByRole("listbox", {
    name: "Linked Records candidates",
  }) as HTMLSelectElement;
  const firstOption = current.options[0];
  if (!firstOption) throw new Error("Missing candidate");
  firstOption.selected = true;
  fireEvent.change(current);
  fireEvent.click(screen.getByRole("button", { name: "Use selection" }));
  expect(
    (screen.getByTestId(genericEditValueTestId(view)) as HTMLTextAreaElement)
      .value,
  ).toBe(`${id(800)}\n${id(1)}`);
  expect(write).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Update parent" }));
  expect(write).toHaveBeenCalledWith({
    change: {
      field_key: "task.linked_record_ids",
      action_payload: {
        kind: "collection_actions_v1",
        actions: [
          { op: "add_record_ref", linked_record_id: id(800) },
          { op: "add_record_ref", linked_record_id: id(1) },
        ],
      },
    },
  });
});
it("retries only failed reads while retaining raw inspector work and fences dismissal", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  const page = vi
    .fn()
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Source unavailable" },
    })
    .mockImplementationOnce(async (input: WorkbookReferenceRequest) =>
      accepted(input, 100),
    );
  render(
    <Form
      store={store}
      fieldKey="task.decision_record_id"
      write={write}
      reader={{ page }}
    />,
  );
  fireEvent.change(screen.getByTestId(genericEditValueTestId(view)), {
    target: { value: id(900) },
  });
  fireEvent.click(screen.getByRole("button", { name: "Choose decision" }));
  await screen.findByRole("alert");
  expect(
    (screen.getByTestId(genericEditValueTestId(view)) as HTMLInputElement)
      .value,
  ).toBe(id(900));
  fireEvent.click(screen.getByRole("button", { name: "Retry references" }));
  await screen.findByRole("option", { name: `Candidate 101 (${id(101)})` });
  fireEvent.keyDown(screen.getByRole("dialog"), { key: "Escape" });
  expect(screen.queryByRole("dialog")).toBeNull();
  expect(document.activeElement).toBe(
    screen.getByTestId(genericEditValueTestId(view)),
  );
  expect(write).not.toHaveBeenCalled();
  act(() => store.setAuthority(null));
  expect(
    (screen.getByTestId(genericEditValueTestId(view)) as HTMLInputElement)
      .value,
  ).toBe("");
});

it("keeps a keyboard-activated Next focusable through pending and exhausted reads", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  let resolveNext:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  const heldNext = new Promise<WorkbookPortResult<WorkbookReferencePage>>(
    (resolve) => {
      resolveNext = resolve;
    },
  );
  const page = vi.fn((input: WorkbookReferenceRequest) =>
    input.cursorToken ? heldNext : Promise.resolve(accepted(input, 0)),
  );
  render(
    <>
      <Form
        store={store}
        fieldKey="task.linked_record_ids"
        write={write}
        reader={{ page }}
      />
      <button type="button">Elsewhere</button>
    </>,
  );
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const next = await screen.findByRole("button", { name: "Next" });
  await waitFor(() => expect(next).toHaveProperty("disabled", false));
  next.focus();
  fireEvent.click(next);
  expect(page).toHaveBeenCalledTimes(2);
  expect(document.activeElement).toBe(next);
  expect(next).toHaveProperty("disabled", false);
  expect(next.getAttribute("aria-disabled")).toBe("true");
  fireEvent.click(next);
  expect(page).toHaveBeenCalledTimes(2);
  fireEvent.keyDown(next, { key: "ArrowDown" });
  const nextRequest = page.mock.calls[1]?.[0];
  if (!nextRequest) throw new Error("Missing Next request");
  await act(async () => resolveNext?.(accepted(nextRequest, 100)));
  expect(document.activeElement).toBe(next);
  expect(next).toHaveProperty("disabled", false);
  expect(next.getAttribute("aria-disabled")).toBe("true");
  screen.getByRole("button", { name: "Elsewhere" }).focus();
  await waitFor(() => expect(next).toHaveProperty("disabled", true));
  expect(write).not.toHaveBeenCalled();
});

it("keeps First and Previous as focus anchors during held reads", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  let settleFirst:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  let settlePrevious:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  const heldFirst = new Promise<WorkbookPortResult<WorkbookReferencePage>>(
    (resolve) => {
      settleFirst = resolve;
    },
  );
  const heldPrevious = new Promise<WorkbookPortResult<WorkbookReferencePage>>(
    (resolve) => {
      settlePrevious = resolve;
    },
  );
  let reads = 0;
  const page = vi.fn((input: WorkbookReferenceRequest) => {
    reads += 1;
    if (reads === 3) return heldFirst;
    if (reads === 5) return heldPrevious;
    return Promise.resolve(accepted(input, input.cursorToken ? 100 : 0));
  });
  render(
    <>
      <Form
        store={store}
        fieldKey="task.linked_record_ids"
        write={write}
        reader={{ page }}
      />
      <button type="button">Elsewhere</button>
    </>,
  );
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  await screen.findByRole("option", { name: `Candidate 1 (${id(1)})` });
  fireEvent.click(screen.getByRole("button", { name: "Next" }));
  await screen.findByRole("option", { name: `Candidate 101 (${id(101)})` });
  const first = screen.getByRole("button", { name: "First" });
  first.focus();
  fireEvent.click(first);
  expect(page).toHaveBeenCalledTimes(3);
  expect(first).toHaveProperty("disabled", false);
  expect(first.getAttribute("aria-disabled")).toBe("true");
  fireEvent.click(first);
  expect(page).toHaveBeenCalledTimes(3);
  const firstRequest = page.mock.calls[2]?.[0];
  if (!firstRequest) throw new Error("Missing First request");
  await act(async () => settleFirst?.(accepted(firstRequest, 0)));
  expect(document.activeElement).toBe(first);
  expect(first).toHaveProperty("disabled", false);
  expect(first.getAttribute("aria-disabled")).toBe("true");
  screen.getByRole("button", { name: "Elsewhere" }).focus();
  await waitFor(() => expect(first).toHaveProperty("disabled", true));
  fireEvent.click(screen.getByRole("button", { name: "Next" }));
  await screen.findByRole("option", { name: `Candidate 101 (${id(101)})` });
  const previous = screen.getByRole("button", { name: "Previous" });
  previous.focus();
  fireEvent.click(previous);
  expect(page).toHaveBeenCalledTimes(5);
  expect(previous).toHaveProperty("disabled", false);
  expect(previous.getAttribute("aria-disabled")).toBe("true");
  const previousRequest = page.mock.calls[4]?.[0];
  if (!previousRequest) throw new Error("Missing Previous request");
  await act(async () => settlePrevious?.(accepted(previousRequest, 0)));
  expect(document.activeElement).toBe(previous);
  expect(previous).toHaveProperty("disabled", false);
  expect(previous.getAttribute("aria-disabled")).toBe("true");
  expect(write).not.toHaveBeenCalled();
});

it("does not reclaim focus when a pending reference read completes after departure", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  let settle:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  const held = new Promise<WorkbookPortResult<WorkbookReferencePage>>(
    (resolve) => {
      settle = resolve;
    },
  );
  const page = vi.fn((input: WorkbookReferenceRequest) =>
    input.cursorToken ? held : Promise.resolve(accepted(input, 0)),
  );
  render(
    <>
      <Form
        store={store}
        fieldKey="task.linked_record_ids"
        write={write}
        reader={{ page }}
      />
      <button type="button">Elsewhere</button>
    </>,
  );
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const next = screen.getByRole("button", { name: "Next" });
  await waitFor(() => expect(next).toHaveProperty("disabled", false));
  next.focus();
  fireEvent.click(next);
  const elsewhere = screen.getByRole("button", { name: "Elsewhere" });
  elsewhere.focus();
  const nextRequest = page.mock.calls[1]?.[0];
  if (!nextRequest) throw new Error("Missing Next request");
  await act(async () => settle?.(accepted(nextRequest, 100)));
  expect(document.activeElement).toBe(elsewhere);
  expect(next).toHaveProperty("disabled", true);
  expect(page).toHaveBeenCalledTimes(2);
  expect(write).not.toHaveBeenCalled();
});

it("retires pending read focus on source replacement and popup closure", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  const held: ((result: WorkbookPortResult<WorkbookReferencePage>) => void)[] =
    [];
  const page = vi.fn((input: WorkbookReferenceRequest) =>
    input.cursorToken
      ? new Promise<WorkbookPortResult<WorkbookReferencePage>>((resolve) => {
          held.push(resolve);
        })
      : Promise.resolve(accepted(input, 0)),
  );
  render(
    <>
      <Form
        store={store}
        fieldKey="task.linked_record_ids"
        write={write}
        reader={{ page }}
      />
      <button type="button">Elsewhere</button>
    </>,
  );
  const parent = screen.getByTestId(
    genericEditValueTestId(view),
  ) as HTMLTextAreaElement;
  fireEvent.change(parent, { target: { value: id(800) } });
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const next = screen.getByRole("button", { name: "Next" });
  await waitFor(() => expect(next).toHaveProperty("disabled", false));
  next.focus();
  fireEvent.click(next);
  const surface = screen.getByLabelText("Reference surface");
  surface.focus();
  fireEvent.change(surface, {
    target: { value: "cartulary.view.indicators.v1" },
  });
  const priorRequest = page.mock.calls[1]?.[0];
  if (!priorRequest) throw new Error("Missing prior source request");
  await act(async () => held[0]?.(accepted(priorRequest, 100)));
  expect(document.activeElement).toBe(surface);
  expect(parent.value).toBe(id(800));
  const updatedNext = screen.getByRole("button", { name: "Next" });
  updatedNext.focus();
  fireEvent.click(updatedNext);
  fireEvent.click(screen.getByRole("button", { name: "Cancel references" }));
  const elsewhere = screen.getByRole("button", { name: "Elsewhere" });
  elsewhere.focus();
  const closedRequest = page.mock.calls[3]?.[0];
  if (!closedRequest) throw new Error("Missing closed picker request");
  await act(async () => held[1]?.(accepted(closedRequest, 100)));
  expect(screen.queryByRole("dialog")).toBeNull();
  expect(document.activeElement).toBe(elsewhere);
  expect(parent.value).toBe(id(800));
  expect(write).not.toHaveBeenCalled();
});

it("retires a pending read when the ordinary filter is applied", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  let settle:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  const reader = {
    page: vi.fn((input: WorkbookReferenceRequest) =>
      input.cursorToken
        ? new Promise<WorkbookPortResult<WorkbookReferencePage>>((resolve) => {
            settle = resolve;
          })
        : Promise.resolve(accepted(input, 0)),
    ),
  };
  render(
    <Form
      store={store}
      fieldKey="task.linked_record_ids"
      write={write}
      reader={reader}
    />,
  );
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const surface = screen.getByLabelText("Reference surface");
  fireEvent.change(surface, {
    target: { value: "cartulary.view.indicators.v1" },
  });
  const next = screen.getByRole("button", { name: "Next" });
  await waitFor(() => expect(next).toHaveProperty("disabled", false));
  next.focus();
  fireEvent.click(next);
  const filterField = screen.getByRole("combobox", {
    name: "Reference filter field",
  }) as HTMLSelectElement;
  const selectedFilter = filterField.options[1]?.value;
  if (!selectedFilter) throw new Error("Missing ordinary reference filter");
  fireEvent.change(filterField, { target: { value: selectedFilter } });
  fireEvent.change(
    screen.getByRole("textbox", { name: "Reference filter value" }),
    { target: { value: "needle" } },
  );
  const apply = screen.getByRole("button", { name: "Apply filter" });
  apply.focus();
  fireEvent.click(apply);
  const priorRequest = reader.page.mock.calls[2]?.[0];
  if (!priorRequest) throw new Error("Missing pending page request");
  await act(async () => settle?.(accepted(priorRequest, 100)));
  expect(document.activeElement).toBe(apply);
  expect(reader.page).toHaveBeenCalledTimes(4);
  expect(reader.page.mock.calls[3]?.[0].queryState.filters).toHaveLength(1);
  expect(write).not.toHaveBeenCalled();
});

it("retires a pending read when the parent record is replaced", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  let settle:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  const reader = {
    page: vi.fn((input: WorkbookReferenceRequest) =>
      input.cursorToken
        ? new Promise<WorkbookPortResult<WorkbookReferencePage>>((resolve) => {
            settle = resolve;
          })
        : Promise.resolve(accepted(input, 0)),
    ),
  };
  const viewFor = (recordId: string) => (
    <>
      <Form
        store={store}
        fieldKey="task.linked_record_ids"
        write={write}
        reader={reader}
        recordId={recordId}
      />
      <button type="button">Elsewhere</button>
    </>
  );
  const { rerender } = render(viewFor(id(999)));
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const next = screen.getByRole("button", { name: "Next" });
  await waitFor(() => expect(next).toHaveProperty("disabled", false));
  next.focus();
  fireEvent.click(next);
  const elsewhere = screen.getByRole("button", { name: "Elsewhere" });
  elsewhere.focus();
  rerender(viewFor(id(998)));
  expect(screen.queryByRole("dialog")).toBeNull();
  const request = reader.page.mock.calls[1]?.[0];
  if (!request) throw new Error("Missing pending page request");
  await act(async () => settle?.(accepted(request, 100)));
  expect(document.activeElement).toBe(elsewhere);
  expect(write).not.toHaveBeenCalled();
});

it("conceals a pending picker without restoring obsolete focus on authority loss", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  let rejectRead:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  const held = new Promise<WorkbookPortResult<WorkbookReferencePage>>(
    (resolve) => {
      rejectRead = resolve;
    },
  );
  const page = vi.fn((input: WorkbookReferenceRequest) =>
    input.cursorToken ? held : Promise.resolve(accepted(input, 0)),
  );
  render(
    <>
      <Form
        store={store}
        fieldKey="task.linked_record_ids"
        write={write}
        reader={{ page }}
      />
      <button type="button">Elsewhere</button>
    </>,
  );
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const next = screen.getByRole("button", { name: "Next" });
  await waitFor(() => expect(next).toHaveProperty("disabled", false));
  next.focus();
  fireEvent.click(next);
  const elsewhere = screen.getByRole("button", { name: "Elsewhere" });
  elsewhere.focus();
  await act(async () =>
    rejectRead?.({
      kind: "rejected",
      failure: { kind: "authentication_required", message: "Sign in again" },
    }),
  );
  expect(document.activeElement).toBe(elsewhere);
  expect(
    screen.queryByRole("listbox", { name: "Linked Records candidates" }),
  ).toBeNull();
  expect(write).not.toHaveBeenCalled();
});

it("keeps keyboard Retry present during a held read and permits repeated failure", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  let settleRetry:
    | ((result: WorkbookPortResult<WorkbookReferencePage>) => void)
    | undefined;
  const heldRetry = new Promise<WorkbookPortResult<WorkbookReferencePage>>(
    (resolve) => {
      settleRetry = resolve;
    },
  );
  const page = vi
    .fn()
    .mockResolvedValueOnce({
      kind: "rejected",
      failure: { kind: "retryable", message: "Source unavailable" },
    })
    .mockImplementationOnce(() => heldRetry)
    .mockImplementationOnce(async (input: WorkbookReferenceRequest) =>
      accepted(input, 0),
    );
  render(
    <Form
      store={store}
      fieldKey="task.decision_record_id"
      write={write}
      reader={{ page }}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Choose decision" }));
  const retry = await screen.findByRole("button", {
    name: "Retry references",
  });
  retry.focus();
  fireEvent.click(retry);
  expect(page).toHaveBeenCalledTimes(2);
  expect(retry.isConnected).toBe(true);
  expect(document.activeElement).toBe(retry);
  expect(retry).toHaveProperty("disabled", false);
  expect(retry.getAttribute("aria-disabled")).toBe("true");
  fireEvent.click(retry);
  expect(page).toHaveBeenCalledTimes(2);
  await act(async () =>
    settleRetry?.({
      kind: "rejected",
      failure: { kind: "retryable", message: "Still unavailable" },
    }),
  );
  expect(document.activeElement).toBe(retry);
  expect(retry.getAttribute("aria-disabled")).toBe("false");
  fireEvent.click(retry);
  await screen.findByRole("option", { name: `Candidate 1 (${id(1)})` });
  expect(page).toHaveBeenCalledTimes(3);
  expect(document.activeElement).toBe(retry);
  act(() => screen.getByRole("button", { name: "Cancel references" }).focus());
  expect(screen.queryByRole("button", { name: "Retry references" })).toBeNull();
  expect(write).not.toHaveBeenCalled();
});

it("moves staged removal focus by stable identity and leaves the parent draft untouched", async () => {
  for (const [removed, expected] of [
    [1, 2],
    [2, 3],
    [3, 2],
    [1, 0],
  ] as const) {
    const store = new WorkbookInspectorDraftStore();
    store.setAuthority(taskAuthority);
    const write = vi.fn();
    render(
      <Form
        store={store}
        fieldKey="task.linked_record_ids"
        write={write}
        reader={{ page: async (input) => accepted(input, 0) }}
      />,
    );
    const raw = (expected === 0 ? [id(1)] : [id(1), id(2), id(3)]).join("\n");
    const parent = screen.getByTestId(
      genericEditValueTestId(view),
    ) as HTMLTextAreaElement;
    fireEvent.change(parent, { target: { value: raw } });
    fireEvent.click(
      screen.getByRole("button", { name: "Choose linked records" }),
    );
    const source = await screen.findByRole("button", {
      name: `Remove selected Candidate ${removed}`,
    });
    source.focus();
    fireEvent.click(source);
    expect(document.activeElement).toBe(
      expected === 0
        ? screen.getByRole("listbox", { name: "Linked Records candidates" })
        : screen.getByRole("button", {
            name: `Remove selected Candidate ${expected}`,
          }),
    );
    expect(parent.value).toBe(raw);
    expect(write).not.toHaveBeenCalled();
    cleanup();
  }
});

it("distinguishes duplicate labels and retains off-page identity during removal", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  render(
    <Form
      store={store}
      fieldKey="task.linked_record_ids"
      write={write}
      reader={{
        page: async (input) => {
          const result = accepted(input, 0);
          if (result.kind !== "accepted") return result;
          return {
            ...result,
            value: {
              ...result.value,
              candidates: result.value.candidates.map((item) => ({
                ...item,
                displayText: [id(1), id(2)].includes(item.identity.id)
                  ? "Duplicate"
                  : item.displayText,
              })),
            },
          };
        },
      }}
    />,
  );
  const raw = [id(800), id(1), id(2)].join("\n");
  const parent = screen.getByTestId(
    genericEditValueTestId(view),
  ) as HTMLTextAreaElement;
  fireEvent.change(parent, { target: { value: raw } });
  fireEvent.click(
    screen.getByRole("button", { name: "Choose linked records" }),
  );
  const offPage = await screen.findByRole("button", {
    name: `Remove selected ${id(800)}`,
  });
  offPage.focus();
  fireEvent.click(offPage);
  const first = screen.getByRole("button", {
    name: `Remove selected Duplicate (${id(1)})`,
  });
  expect(document.activeElement).toBe(first);
  fireEvent.click(first);
  expect(document.activeElement).toBe(
    screen.getByRole("button", { name: "Remove selected Duplicate" }),
  );
  expect(parent.value).toBe(raw);
  expect(write).not.toHaveBeenCalled();
});

it("uses a labelled local fallback when a sole selected item has no candidate", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  render(
    <Form
      store={store}
      fieldKey="task.decision_record_id"
      write={write}
      reader={{
        page: async (input) => {
          const result = accepted(input, 0);
          if (result.kind !== "accepted") return result;
          return {
            ...result,
            value: {
              ...result.value,
              candidates: [],
              paging: { limit: 100, hasMore: false, nextCursor: null },
            },
          };
        },
      }}
    />,
  );
  const parent = screen.getByTestId(
    genericEditValueTestId(view),
  ) as HTMLInputElement;
  fireEvent.change(parent, { target: { value: id(800) } });
  fireEvent.click(screen.getByRole("button", { name: "Choose decision" }));
  const candidate = await screen.findByRole("listbox", {
    name: "Decision candidates",
  });
  expect(candidate).toHaveProperty("disabled", true);
  const remove = screen.getByRole("button", {
    name: `Remove selected ${id(800)}`,
  });
  remove.focus();
  fireEvent.click(remove);
  const fallback = screen.getByRole("group", {
    name: "Decision selected references",
  });
  expect(document.activeElement).toBe(fallback);
  expect(fireEvent.keyDown(fallback, { key: "Tab", shiftKey: true })).toBe(
    true,
  );
  expect(parent.value).toBe(id(800));
  expect(write).not.toHaveBeenCalled();
});
it("uses exact collection action identities in declared order with the pending edit limit", () => {
  const surfaces = [
    "task_requests",
    "decisions",
    "findings",
    "comm_log",
    "handoff",
    "status_review",
    "lesson",
  ];
  for (const surface of surfaces) {
    const contract = requireViewContract(`cartulary.view.${surface}.v1`);
    for (const field of contract.fields) {
      const reference = getReferenceFieldContract(
        contract.viewSchemaId,
        field.fieldKey,
      );
      if (reference?.kind !== "collection") continue;
      const removed = prepareWorkbookInspectorChange(
        field,
        "opaque-item-b\nopaque-item-a",
        "remove",
        contract.viewSchemaId,
      );
      expect(removed.change?.action_payload?.actions).toEqual([
        { op: reference.removeOperation, item_ref: "opaque-item-b" },
        { op: reference.removeOperation, item_ref: "opaque-item-a" },
      ]);
      const added = prepareWorkbookInspectorChange(
        field,
        `${id(2)}\n${id(1)}`,
        "add",
        contract.viewSchemaId,
      );
      expect(added.change?.action_payload?.actions).toEqual(
        [2, 1].map((number) => ({
          op: reference.addOperation,
          [reference.identityKind === "party"
            ? "party_id"
            : "linked_record_id"]: id(number),
        })),
      );
      for (const raw of [
        "",
        ` ${id(1)}`,
        `${id(1)} `,
        Array.from({ length: 65 }, (_, i) => id(i)).join("\n"),
      ])
        expect(
          prepareWorkbookInspectorChange(
            field,
            raw,
            "add",
            contract.viewSchemaId,
          ).error,
        ).toBeDefined();
    }
  }
});

it("uses committed target presentation without retargeting staged identity or editing the parent", async () => {
  const store = new WorkbookInspectorDraftStore();
  store.setAuthority(taskAuthority);
  const write = vi.fn();
  const listeners = new Set<() => void>();
  let snapshot = { authority: taskAuthority };
  let row: ReturnType<WorkbookCommittedRecordPort["latestRow"]> = null;
  const recordEvidence: WorkbookCommittedRecordPort = {
    ...evidence,
    subscribe: (listener) => {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    getSnapshot: () => snapshot,
    latestRow: () => row,
  };
  render(
    <Form
      store={store}
      fieldKey="task.decision_record_id"
      write={write}
      reader={{ page: async (input) => accepted(input, 0) }}
      recordEvidence={recordEvidence}
    />,
  );
  fireEvent.click(screen.getByRole("button", { name: "Choose decision" }));
  const select = await screen.findByRole("listbox", {
    name: "Decision candidates",
  });
  fireEvent.change(select, { target: { value: `record:${id(1)}` } });
  await screen.findByRole("button", { name: "Remove selected Candidate 1" });
  act(() => {
    row = {
      record_id: id(1),
      row_version: 2,
      cells: { "decision.summary": { value: "Committed target label" } },
    };
    snapshot = { authority: taskAuthority };
    for (const listener of listeners) listener();
  });
  await screen.findByRole("button", {
    name: "Remove selected Committed target label",
  });
  expect(
    (screen.getByTestId(genericEditValueTestId(view)) as HTMLInputElement)
      .value,
  ).toBe("");
  expect(write).not.toHaveBeenCalled();
  fireEvent.click(screen.getByRole("button", { name: "Use selection" }));
  expect(
    (screen.getByTestId(genericEditValueTestId(view)) as HTMLInputElement)
      .value,
  ).toBe(id(1));
  expect(
    store.read({
      viewSchemaId: view,
      recordId: id(999),
      fieldKey: "task.decision_record_id",
      action: "value",
    })?.references?.[0],
  ).toMatchObject({ recordId: id(1), displayText: "Committed target label" });
  expect(write).not.toHaveBeenCalled();
});
