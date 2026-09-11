import {
  useCallback,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import type { WorkbookMutationRuntime } from "../../runtime/WorkbookMutationRuntime";
import type { TimelineEditorDraftRegistry } from "../editing/useTimelineEditorDraftRegistry";
import type { TimelineCommittedRecordIdleResult } from "../models/timelineControllerPorts";
import type { TimelinePendingSavesRefs } from "../models/timelinePendingSaves";
import type { WorkbookRow } from "../models/timelineRowModel";
import type { TimelineCaptureBinding } from "../ports/TimelineRecordActionPort";
import {
  type TimelineCaptureAction,
  type TimelineCaptureReview,
  timelineCaptureIneligibility,
  timelineCaptureSubject,
} from "./timelineCaptureActionModel";
import { timelineCaptureOwnerFor } from "./timelineCaptureOwnerFor";

export function useTimelineCaptureActions(options: {
  readonly runtime: WorkbookMutationRuntime;
  readonly selectedRow: WorkbookRow | null;
  readonly selectedId: string | null;
  readonly isOpen: boolean;
  readonly deleted: boolean;
  readonly concealed: boolean;
  readonly originKey: string;
  readonly rowsRef: { readonly current: WorkbookRow[] };
  readonly drafts: TimelineEditorDraftRegistry;
  readonly pending: TimelinePendingSavesRefs;
  readonly openHistory: (id: string) => void;
  readonly waitForIdle: (
    id: string,
    options: { signal: AbortSignal; refreshIfMissing: boolean },
  ) => Promise<TimelineCommittedRecordIdleResult | null>;
  readonly acceptReceipt: (
    receipt: import("../adapters/timelineCaptureProtocol").TimelineCaptureReceipt,
    baseVersion: number,
  ) => void;
  readonly loadRows: (options: {
    showLoading: boolean;
    requireAcceptance?: boolean;
  }) => Promise<void>;
}) {
  const owner = useMemo(
    () => timelineCaptureOwnerFor(options.runtime),
    [options.runtime],
  );
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const onAccessLost = useCallback(() => owner.loseAccess(), [owner]);
  const applied = useRef(new Set<number>());
  useLayoutEffect(() => {
    if (options.concealed || !snapshot.authority) return;
    for (const entry of snapshot.entries) {
      if (!entry.receipt || applied.current.has(entry.key)) continue;
      applied.current.add(entry.key);
      if (
        entry.receipt.data.row_version >=
        (owner.latestVersion(entry.review.target.recordId) ?? 0)
      )
        options.acceptReceipt(entry.receipt, entry.review.target.rowVersion);
    }
  }, [snapshot, owner, options]);
  const current = useRef(options);
  current.current = options;
  const alive = useRef(true);
  const [editing, setEditing] = useState<string | null>(null);
  const shortcut = useRef<{
    review: TimelineCaptureReview;
    earlier: Promise<void>;
  } | null>(null);
  const target =
    options.selectedRow?.rawRow && !options.deleted && !options.concealed
      ? timelineCaptureSubject(
          options.selectedRow.rawRow,
          options.runtime.scope.incidentId,
        )
      : null;
  const originKey = `${options.originKey}:${options.selectedId ?? ""}:${options.isOpen}`;
  const origin = useRef(originKey);
  const originGeneration = useRef(0);
  if (origin.current !== originKey) originGeneration.current++;
  origin.current = originKey;
  useLayoutEffect(() => {
    alive.current = true;
    return () => {
      alive.current = false;
      shortcut.current = null;
    };
  }, []);
  useLayoutEffect(() => {
    if (
      !options.isOpen ||
      options.concealed ||
      options.deleted ||
      !snapshot.authority
    )
      setEditing(null);
  }, [options.isOpen, options.concealed, options.deleted, snapshot.authority]);
  function reason(row: WorkbookRow | null, action: TimelineCaptureAction) {
    const subject = row?.rawRow
      ? timelineCaptureSubject(row.rawRow, owner.incidentId)
      : null;
    return (
      owner.unavailableReason() ??
      (current.current.concealed
        ? "This Timeline record is unavailable."
        : null) ??
      timelineCaptureIneligibility(subject, action) ??
      (row?.recordId && owner.blocksRecord(row.recordId)
        ? "An earlier Timeline action needs to finish or be recovered."
        : null)
    );
  }
  function binding(
    review: TimelineCaptureReview,
    matches: () => boolean,
    earlier = options.pending.saveQueueRef.current,
  ): TimelineCaptureBinding {
    const expectedOrigin = origin.current;
    const expectedGeneration = originGeneration.current;
    const isCurrent = () =>
      alive.current &&
      originGeneration.current === expectedGeneration &&
      origin.current === expectedOrigin &&
      !current.current.concealed &&
      !current.current.deleted &&
      current.current.isOpen &&
      current.current.selectedId === review.target.recordId;
    return {
      isCurrent,
      matchesReview: () =>
        matches() &&
        current.current.selectedRow?.rawRow?.row_version ===
          review.target.rowVersion,
      prepare: async (signal) => {
        await earlier;
        if (signal.aborted || !isCurrent()) return null;
        const committed = await current.current.waitForIdle(
          review.target.recordId,
          { signal, refreshIfMissing: false },
        );
        if (!committed?.row?.rawRow || signal.aborted || !isCurrent())
          return null;
        const live = current.current.rowsRef.current.find(
          (row) => row.recordId === review.target.recordId,
        );
        if (!live) return null;
        const materialized = current.current.drafts.materializeRow(live);
        if (
          Object.entries(materialized.values).some(
            ([field, value]) =>
              value !==
              committed.row?.committedValues[
                field as keyof typeof materialized.values
              ],
          ) ||
          Object.values(materialized.collectionDrafts).some(
            (value) => value.trim() !== "",
          )
        )
          return null;
        if (
          current.current.runtime.history
            .getSnapshot()
            .some(
              (entry) =>
                entry.attempt.subject.recordId === review.target.recordId &&
                ["preparing", "submitting", "uncertain"].includes(entry.phase),
            )
        )
          return null;
        const subject = timelineCaptureSubject(
          committed.row.rawRow,
          owner.incidentId,
        );
        if (subject.rowVersion !== review.target.rowVersion)
          await current.current.loadRows({
            showLoading: false,
            requireAcceptance: true,
          });
        return subject;
      },
    };
  }
  function makeReview(
    row: WorkbookRow,
    action: TimelineCaptureAction,
  ): TimelineCaptureReview | null {
    const state = owner.getSnapshot();
    if (!row.rawRow || !state.authority || reason(row, action)) return null;
    return Object.freeze({
      action,
      target: timelineCaptureSubject(row.rawRow, owner.incidentId),
      replacement: null,
      reason: null,
      authority: state.authority,
      authorityGeneration: state.generation,
      originKey: `${options.originKey}:${row.recordId}:true`,
    });
  }
  function activate(action: TimelineCaptureAction) {
    if (!options.selectedRow || reason(options.selectedRow, action)) return;
    if (action === "supersede") {
      setEditing(origin.current);
      return;
    }
    const review = makeReview(options.selectedRow, action);
    if (review)
      owner.submit(
        review,
        binding(review, () => true),
      );
  }
  function start(rowKey: string, action: TimelineCaptureAction) {
    if (shortcut.current) return;
    const row = current.current.rowsRef.current.find(
      (candidate) => candidate.key === rowKey,
    );
    if (!row?.recordId || reason(row, action)) return;
    const review = makeReview(row, action);
    if (!review) return;
    shortcut.current = {
      review,
      earlier: current.current.pending.saveQueueRef.current,
    };
    current.current.openHistory(row.recordId);
  }
  useLayoutEffect(() => {
    const pending = shortcut.current;
    if (!pending) return;
    shortcut.current = null;
    if (
      !options.isOpen ||
      options.selectedId !== pending.review.target.recordId ||
      options.concealed ||
      options.deleted
    )
      return;
    if (pending.review.action === "supersede") setEditing(origin.current);
    else
      owner.submit(
        pending.review,
        binding(pending.review, () => true, pending.earlier),
      );
  });
  const additionalDisabledReasons = new Map<string, string>();
  for (const [action, key] of [
    ["mark-reviewed", "timeline.mark_reviewed"],
    ["supersede", "timeline.supersede"],
  ] as const) {
    const disabled = options.deleted
      ? "This Timeline row is deleted."
      : reason(options.selectedRow, action);
    if (disabled) additionalDisabledReasons.set(key, disabled);
  }
  return {
    result:
      [...snapshot.entries]
        .reverse()
        .find(
          (entry) =>
            entry.review.target.recordId === options.selectedId &&
            entry.receipt !== null,
        ) ?? null,
    owner,
    start,
    activate,
    reason,
    additionalDisabledReasons,
    editor:
      editing === originKey &&
      target &&
      target.captureState !== "superseded" &&
      snapshot.authority &&
      !options.concealed
        ? {
            target,
            authority: snapshot.authority,
            generation: snapshot.generation,
            originKey,
            port: owner,
            latestVersion: owner.latestVersion,
            blocked: owner.blocksRecord(target.recordId),
            onPrepare: (review: TimelineCaptureReview, signal: AbortSignal) =>
              owner.prepareReview(
                review,
                binding(review, () => true),
                signal,
              ),
            onConfirm: (
              review: TimelineCaptureReview,
              matches: () => boolean,
            ) => owner.submit(review, binding(review, matches)),
            onCancel: () => setEditing(null),
            onAccessLost,
          }
        : null,
  };
}
