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
  dependenciesForField?: (fieldKey: string) => readonly string[];
  retainedSelection?: {
    readonly fields: readonly ViewFieldContract[];
    readonly select: (identity: InspectorEditIdentity) => void;
  };
}) {
  const { store, row, field } = input;
  const boundary = JSON.stringify([
    input.viewSchemaId,
    row?.record_id,
    input.presentation,
    input.active,
  ]);
  const authorityGeneration = store.getAuthorityGeneration();
  const currentInput = useRef({ input, boundary });
  currentInput.current = { input, boundary };
  const pendingResume = useRef<{
    key: string;
    boundary: string;
    authorityGeneration: number;
    revision: number;
    trigger: HTMLElement | undefined;
  } | null>(null);
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
    ...new Set([
      identity.fieldKey,
      ...(input.dependenciesForField?.(identity.fieldKey) ?? []),
    ]),
  ];
  const staleFields =
    eligible && row ? store.staleFields(identity, row, dependencies) : [];
  useLayoutEffect(() => {
    const pending = pendingResume.current;
    if (!pending) return;
    pendingResume.current = null;
    if (
      pending.key !== key ||
      pending.boundary !== boundary ||
      pending.authorityGeneration !== store.getAuthorityGeneration() ||
      pending.revision !== draft?.revision ||
      !eligible ||
      !store.canAuthor() ||
      staleFields.length
    )
      return;
    if (pending.trigger)
      resumeFocus.current = { attachment, trigger: pending.trigger };
    store.resume(identity, attachment);
  });
  return {
    identity,
    retainedWork:
      row && input.active
        ? store
            .readRecord(input.viewSchemaId, row.record_id)
            .map((retained) => {
              const target = input.retainedSelection?.fields.find(
                (candidate) =>
                  candidate.fieldKey === retained.identity.fieldKey,
              );
              const changed = store.staleFields(retained.identity, row, [
                retained.identity.fieldKey,
                ...(input.dependenciesForField?.(retained.identity.fieldKey) ??
                  []),
              ]);
              const admitted =
                !!target?.patchWritable &&
                Object.hasOwn(row.cells, retained.identity.fieldKey) &&
                store.canAuthor() &&
                !!input.retainedSelection;
              return {
                identity: retained.identity,
                reviewRequired: changed.length > 0,
                command: admitted
                  ? {
                      kind: changed.length
                        ? ("review" as const)
                        : ("resume" as const),
                      invoke: (trigger?: HTMLElement) => {
                        if (
                          currentInput.current.boundary !== boundary ||
                          store.getAuthorityGeneration() !==
                            authorityGeneration ||
                          !store.canAuthor() ||
                          store.read(retained.identity)?.revision !==
                            retained.revision
                        )
                          return;
                        pendingResume.current = {
                          key: inspectorEditKey(retained.identity),
                          boundary,
                          authorityGeneration,
                          revision: retained.revision,
                          trigger,
                        };
                        currentInput.current.input.retainedSelection?.select(
                          retained.identity,
                        );
                      },
                    }
                  : null,
                discard: () => store.discard(retained.identity),
              };
            })
        : [],
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
        if (!store.staleFields(identity, row, dependencies).length)
          store.resume(identity, attachment);
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
