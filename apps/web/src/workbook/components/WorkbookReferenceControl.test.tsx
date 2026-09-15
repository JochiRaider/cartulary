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
}: {
  store: WorkbookInspectorDraftStore;
  fieldKey: string;
  write: (change: unknown) => void;
  reader: WorkbookReferenceReadPort;
  recordEvidence?: WorkbookCommittedRecordPort;
}) {
  const field = contract.fieldMap[fieldKey];
  if (!field) throw new Error("Missing test field");
  const edit = useWorkbookInspectorEditDraft({
    store,
    row: {
      record_id: id(999),
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
