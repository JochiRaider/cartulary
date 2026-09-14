import type { ViewFieldContract } from "@cartulary/view-contracts";
import { useId, useLayoutEffect, useRef, useSyncExternalStore } from "react";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import {
  type InspectorEditIdentity,
  inspectorEditKey,
  type WorkbookInspectorDraftStore,
} from "./WorkbookInspectorDraftStore";

export function useWorkbookInspectorEditDraft(input: {
  store: WorkbookInspectorDraftStore;
  row: WorkbookQueryRow | null;
  field: ViewFieldContract | null;
  viewSchemaId: string;
  action?: string;
  presentation: string;
  active: boolean;
  dependencies?: readonly string[];
}) {
  const { store, row, field } = input;
  const controlRef = useRef<HTMLElement | null>(null);
  const resumeFocus = useRef<{
    attachment: string;
    trigger: HTMLElement;
  } | null>(null);
  useSyncExternalStore(store.subscribe, store.getSnapshot);
  const id = useId();
  const identity: InspectorEditIdentity = {
    viewSchemaId: input.viewSchemaId,
    recordId: row?.record_id ?? "",
    fieldKey: field?.fieldKey ?? "",
    action: input.action ?? "value",
  };
  const key = inspectorEditKey(identity);
  const eligible =
    input.active &&
    !!row &&
    !!field?.patchWritable &&
    Object.hasOwn(row.cells, field.fieldKey) &&
    store.canRead();
  const presentationKey = JSON.stringify([
    id,
    input.presentation,
    key,
    eligible,
  ]);
  const observed = useRef({ key: presentationKey, generation: 0 });
  if (observed.current.key !== presentationKey)
    observed.current = {
      key: presentationKey,
      generation: observed.current.generation + 1,
    };
  const attachment = JSON.stringify([
    presentationKey,
    observed.current.generation,
  ]);
  const latest = useRef({ attachment, eligible });
  latest.current = { attachment, eligible };
  useLayoutEffect(() => () => store.detach(attachment), [store, attachment]);
  const draft = eligible ? store.read(identity) : null;
  const needsResume = !!draft && draft.attachment !== attachment;
  useLayoutEffect(() => {
    const pending = resumeFocus.current;
    if (!pending) return;
    resumeFocus.current = null;
    if (
      eligible &&
      !needsResume &&
      store.canAuthor() &&
      pending.attachment === attachment &&
      (document.activeElement === pending.trigger ||
        (!pending.trigger.isConnected &&
          document.activeElement === document.body))
    )
      controlRef.current?.focus({ preventScroll: true });
  });
  const saved =
    eligible && row && field ? row.cells[field.fieldKey]?.value : null;
  const value = draft
    ? draft.value
    : field?.writeKind === "action_payload"
      ? ""
      : saved === null || saved === undefined
        ? ""
        : String(saved);
  const dependencies = [
    ...new Set([identity.fieldKey, ...(input.dependencies ?? [])]),
  ];
  const staleFields =
    eligible && row ? store.staleFields(identity, row, dependencies) : [];
  return {
    identity,
    attachment,
    draft,
    value,
    needsResume,
    staleFields,
    controlRef,
    canResume: eligible && store.canAuthor(),
    canEdit: eligible && store.canAuthor() && !needsResume,
    canSubmit:
      eligible && store.canAuthor() && !needsResume && staleFields.length === 0,
    baseline: draft?.baseline ?? row,
    update: (value: string | null) => {
      if (eligible && row) store.update(identity, row, value, attachment);
    },
    selectReferences: (
      references: Parameters<WorkbookInspectorDraftStore["update"]>[4],
    ) => {
      if (eligible && row && references)
        store.update(
          identity,
          row,
          references.map((item) => item.recordId).join("\n"),
          attachment,
          references,
        );
    },
    resume: (trigger?: HTMLElement) => {
      if (eligible && store.canAuthor()) {
        if (trigger) resumeFocus.current = { attachment, trigger };
        store.resume(identity, attachment);
      }
    },
    discard: (trigger?: HTMLElement) => {
      if (trigger) resumeFocus.current = { attachment, trigger };
      store.discard(identity);
    },
    review: (changed: string, keepDraft: boolean, trigger?: HTMLElement) => {
      if (eligible && row) {
        if (trigger) resumeFocus.current = { attachment, trigger };
        store.review(identity, row, changed, keepDraft);
      }
    },
    capture: () => ({ draft: store.capture(identity), attachment }),
    isCurrent: (captured: { attachment: string }) =>
      latest.current.eligible &&
      latest.current.attachment === captured.attachment,
    complete: (captured: {
      draft: ReturnType<WorkbookInspectorDraftStore["capture"]>;
    }) => store.acknowledge(captured.draft),
  };
}
export type WorkbookInspectorEditDraft = ReturnType<
  typeof useWorkbookInspectorEditDraft
>;
