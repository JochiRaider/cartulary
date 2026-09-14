import { genericEditValueTestId } from "@cartulary/ui-contracts";
import {
  listViewContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import {
  act,
  cleanup,
  fireEvent,
  render,
  renderHook,
  screen,
} from "@testing-library/react";
import { afterEach, expect, it } from "vitest";
import { taskAuthority } from "../../testing/taskWorkbookTestSupport";
import { emptyGenericReferenceOptions } from "../models/workbookReferenceOptions";
import { useWorkbookInspectorEditDraft } from "./useWorkbookInspectorEditDraft";
import { WorkbookInspectorDraftStore } from "./WorkbookInspectorDraftStore";
import { WorkbookInspectorEditControl } from "./WorkbookInspectorEditControl";

afterEach(cleanup);
it("retains selected reference identities outside the current option page and clears with explicit null", () => {
  const contract = requireViewContract("cartulary.view.task_requests.v1"),
    field = contract.fieldMap["task.decision_record_id"],
    store = new WorkbookInspectorDraftStore(),
    referenceId = "00000000-0000-4000-8000-000000000905";
  if (!field) throw new Error("Missing Decision reference field");
  store.setAuthority(taskAuthority);
  const row = {
    record_id: "task",
    row_version: 1,
    cells: { [field.fieldKey]: { value: null } },
  };
  function Form({ present }: { present: boolean }) {
    const edit = useWorkbookInspectorEditDraft({
      store,
      row,
      field: field ?? null,
      viewSchemaId: contract.viewSchemaId,
      active: true,
      presentation: "form",
    });
    if (!field) return null;
    return (
      <WorkbookInspectorEditControl
        edit={edit}
        field={field}
        collectionMode="add"
        referenceOptions={{
          ...emptyGenericReferenceOptions(),
          decisions: present
            ? [
                {
                  recordId: referenceId,
                  label: "Reviewed Decision",
                  viewSchemaId: "cartulary.view.decisions.v1",
                },
              ]
            : [],
        }}
        testId={genericEditValueTestId(contract.viewSchemaId)}
      />
    );
  }
  const view = render(<Form present />);
  fireEvent.change(
    screen.getByTestId(genericEditValueTestId(contract.viewSchemaId)),
    {
      target: { value: referenceId },
    },
  );
  view.rerender(<Form present={false} />);
  expect(
    (
      screen.getByTestId(
        genericEditValueTestId(contract.viewSchemaId),
      ) as HTMLSelectElement
    ).value,
  ).toBe(referenceId);
  expect(
    screen.getByRole("option", { name: "Reviewed Decision" }),
  ).toBeTruthy();
  fireEvent.click(screen.getByRole("button", { name: `Clear ${field.label}` }));
  expect(
    store.read({
      viewSchemaId: contract.viewSchemaId,
      recordId: row.record_id,
      fieldKey: field.fieldKey,
      action: "value",
    })?.value,
  ).toBeNull();
  expect(screen.getByRole("status").textContent).toContain("will be cleared");
});
it("binds every retained field and action to its original subject through refresh detachment and explicit return", () => {
  for (const contract of listViewContracts().filter(
    (c) => !c.viewSchemaId.includes("timeline"),
  ))
    for (const field of contract.fields.filter((f) => f.patchWritable))
      for (const action of field.writeKind === "action_payload"
        ? ["add", "remove"]
        : ["value"]) {
        const store = new WorkbookInspectorDraftStore();
        store.setAuthority(taskAuthority);
        const row = {
          record_id: "original",
          row_version: 1,
          cells: Object.fromEntries(
            contract.fields.map((f) => [f.fieldKey, { value: null }]),
          ),
        };
        const initial: Parameters<typeof useWorkbookInspectorEditDraft>[0] = {
          store,
          row,
          field,
          action,
          viewSchemaId: contract.viewSchemaId,
          presentation: "sheet",
          active: true,
        };
        const hook = renderHook(useWorkbookInspectorEditDraft, {
          initialProps: initial,
        });
        act(() => hook.result.current.update(" unfinished "));
        hook.rerender({
          ...initial,
          row: { ...structuredClone(row), row_version: 2 },
        });
        expect(hook.result.current.value).toBe(" unfinished ");
        expect(hook.result.current.canSubmit).toBe(true);
        hook.rerender({ ...initial, row: { ...row, record_id: "other" } });
        expect(hook.result.current.value).toBe("");
        hook.rerender({ ...initial, row: null });
        expect(hook.result.current.draft).toBeNull();
        expect(hook.result.current.canSubmit).toBe(false);
        hook.rerender({ ...initial });
        expect(hook.result.current.needsResume).toBe(true);
        expect(hook.result.current.canSubmit).toBe(false);
        act(() => hook.result.current.resume());
        expect(hook.result.current.canSubmit).toBe(true);
        const captured = hook.result.current.capture();
        hook.rerender({ ...initial, active: false });
        expect(hook.result.current.isCurrent(captured)).toBe(false);
        expect(hook.result.current.value).toBe("");
        hook.rerender({ ...initial });
        expect(hook.result.current.needsResume).toBe(true);
        hook.rerender({ ...initial, field: null });
        expect(hook.result.current.canSubmit).toBe(false);
        expect(hook.result.current.value).toBe("");
        hook.rerender(initial);
        for (const authority of [
          { ...taskAuthority, role: "viewer" as const },
          { ...taskAuthority, role: "" as const },
          { ...taskAuthority, closed: true },
          null,
        ]) {
          act(() => store.setAuthority(authority));
          expect(hook.result.current.canSubmit).toBe(false);
          act(() => hook.result.current.update("unauthorized"));
          if (!authority) expect(hook.result.current.value).toBe("");
        }
        act(() =>
          store.setAuthority({ ...taskAuthority, sessionIdentity: "renewed" }),
        );
        expect(hook.result.current.needsResume).toBe(true);
        act(() => hook.result.current.resume());
        expect(hook.result.current.value).toBe(" unfinished ");
        expect(hook.result.current.canSubmit).toBe(true);
        act(() =>
          store.setAuthority({ ...taskAuthority, actorId: "replacement" }),
        );
        expect(hook.result.current.value).toBe("");
        expect(hook.result.current.canSubmit).toBe(false);
        hook.unmount();
      }
});
