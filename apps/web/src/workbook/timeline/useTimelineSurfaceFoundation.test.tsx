import { requireViewContract } from "@cartulary/view-contracts";
import { act, renderHook } from "@testing-library/react";
import { useState } from "react";
import { expect, it } from "vitest";
import {
  defaultFilterDraft,
  emptyWorkbookQueryState,
} from "../models/workbookQuery";
import { timelineViewSchemaId } from "../models/workbookSurfaceRegistry";
import type { WorkbookPendingMutationPort } from "../ports/WorkbookPendingMutationPort";
import { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import { useTimelineSurfaceFoundation } from "./composition/useTimelineSurfaceFoundation";
import type { TimelineWorkbookSurfaceRuntime } from "./models/timelineWorkbookSurfaceRuntime";

const timelineContract = requireViewContract(timelineViewSchemaId);

function createMutationRuntime() {
  const pendingMutationPort: WorkbookPendingMutationPort = {
    execute: async () => {
      throw new Error("foundation test does not execute mutations");
    },
  };
  return new WorkbookMutationRuntime(
    {
      clientInstanceId: "foundation-test",
      incidentId: "incident-1",
    },
    { create: () => "foundation-test-transaction" },
    pendingMutationPort,
  );
}

it("useTimelineSurfaceFoundation owns stable adapter row query and pending foundations", () => {
  const mutationRuntime = createMutationRuntime();
  const mutationCommands = {
    identity: {
      createLogicalActionId: () => "foundation-logical-action",
    },
  } as unknown as TimelineWorkbookSurfaceRuntime["mutationCommands"];
  const { result, rerender } = renderHook(
    ({ apiBase }) => {
      const [filterDraft, setFilterDraft] = useState(() =>
        defaultFilterDraft(timelineContract),
      );
      const [state, setState] = useState(() => emptyWorkbookQueryState());
      return useTimelineSurfaceFoundation({
        apiBase,
        clipboardPaste: {
          paste: async () => ({
            clientTxnId: null,
            outcome: {
              kind: "rejected",
              failure: { kind: "terminal", message: "not used" },
            },
          }),
        },
        incidentId: "incident-1",
        mutationCommands,
        mutationRuntime,
        query: { filterDraft, setFilterDraft, setState, state },
      });
    },
    { initialProps: { apiBase: undefined as string | undefined } },
  );
  const initialHistoryPort = result.current.ports.history;
  const initialPendingRefs = result.current.refs.pendingSaves;
  const initialRowsRef = result.current.refs.rows;

  expect(result.current.snapshot.rows).toHaveLength(1);
  expect(result.current.snapshot.rows[0]?.recordId).toBeNull();
  expect(initialRowsRef.current).toBe(result.current.snapshot.rows);

  act(() => {
    result.current.commands.editor.activateCollectionInput(
      "record-1:tags:grid",
    );
  });
  expect(result.current.snapshot.editor.activeCollectionInputKey).toBe(
    "record-1:tags:grid",
  );
  act(() => {
    result.current.commands.editor.deactivateCollectionInput(
      "record-2:tags:grid",
    );
  });
  expect(result.current.snapshot.editor.activeCollectionInputKey).toBe(
    "record-1:tags:grid",
  );
  act(() => {
    result.current.commands.editor.deactivateCollectionInput(
      "record-1:tags:grid",
    );
  });
  expect(result.current.snapshot.editor.activeCollectionInputKey).toBeNull();

  act(() => {
    result.current.commands.rows.replaceRows([]);
  });

  expect(result.current.snapshot.rows).toEqual([]);
  expect(initialRowsRef.current).toBe(result.current.snapshot.rows);

  rerender({ apiBase: undefined });

  expect(result.current.ports.history).toBe(initialHistoryPort);
  expect(result.current.refs.pendingSaves).toBe(initialPendingRefs);
  expect(result.current.refs.rows).toBe(initialRowsRef);
  expect(result.current.snapshot.query.queryState).toEqual(
    emptyWorkbookQueryState(),
  );
});
