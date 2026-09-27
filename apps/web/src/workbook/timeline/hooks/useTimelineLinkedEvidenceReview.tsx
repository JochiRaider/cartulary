import { useCallback, useEffect, useRef, useState } from "react";
import { resolveSolePreviewableEvidence } from "../../features/evidence/resolveSolePreviewableEvidence";
import {
  type EvidenceAccessTarget,
  useEvidenceHandleAccess,
} from "../../features/evidence/useEvidenceHandleAccess";
import type { EvidenceCapabilityPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookReadScope } from "../../query/WorkbookQueryRow";
import {
  sameWorkbookReadScope,
  workbookRowIsAdmissible,
} from "../../query/workbookRowObservation";
import type { TimelineInspectorElementRegistry } from "../focus/timelineInspectorElementRegistry";
import {
  readTimelineEvidenceLinks,
  type TimelineEvidenceLink,
  timelineEvidenceLinkKey,
} from "../models/timelineEvidenceLinks";
import type { WorkbookRow } from "../models/timelineRowModel";

function targetFor(
  row: WorkbookRow,
  link: TimelineEvidenceLink,
): EvidenceAccessTarget {
  return {
    recordId: link.evidenceRecordId,
    ...(row.recordId ? { sourceRecordId: row.recordId } : {}),
    title: link.title,
    identity: JSON.stringify([
      timelineEvidenceLinkKey(row),
      link.itemRef,
      link.evidenceRecordId,
    ]),
  };
}

type SpaceIntent = {
  readonly row: WorkbookRow;
  readonly sequence: number;
  readonly scopeKey: string;
};

export function useTimelineLinkedEvidenceReview(input: {
  readonly port: EvidenceCapabilityPort;
  readonly rowsRef: { readonly current: readonly WorkbookRow[] };
  readonly selectedRowId: string | null;
  readonly readScope: WorkbookReadScope | null;
  readonly canRead: boolean;
  readonly inspectorOpen: boolean;
  readonly scopeKey: string;
  readonly elementRegistry: TimelineInspectorElementRegistry;
  readonly onAccessFailure: () => Promise<void> | void;
  readonly onRestoreGridFocus: (recordId: string) => void;
}) {
  const inputRef = useRef(input);
  inputRef.current = input;
  const [spaceIntent, setSpaceIntent] = useState<SpaceIntent | null>(null);
  const [spaceState, setSpaceState] = useState<{
    readonly kind: "checking" | "indeterminate";
    readonly sourceKey: string;
  } | null>(null);
  const intentSequence = useRef(0);
  const probeController = useRef<AbortController | null>(null);
  const priorInspectorOpen = useRef(input.inspectorOpen);
  const cancelSpace = useCallback(() => {
    intentSequence.current += 1;
    probeController.current?.abort();
    probeController.current = null;
    setSpaceIntent(null);
    setSpaceState(null);
  }, []);
  const currentRow = useCallback(
    (recordId: string) =>
      inputRef.current.rowsRef.current.find(
        (row) => row.recordId === recordId,
      ) ?? null,
    [],
  );
  useEffect(() => {
    if (priorInspectorOpen.current && !input.inspectorOpen) cancelSpace();
    priorInspectorOpen.current = input.inspectorOpen;
  }, [input.inspectorOpen, cancelSpace]);
  const access = useEvidenceHandleAccess({
    port: input.port,
    canRead: input.canRead,
    scopeKey: input.scopeKey,
    isCurrent: (target) => {
      if (
        !inputRef.current.canRead ||
        !inputRef.current.inspectorOpen ||
        inputRef.current.selectedRowId === null
      )
        return false;
      const row = currentRow(inputRef.current.selectedRowId);
      if (
        !row?.rawRow ||
        !inputRef.current.readScope ||
        !workbookRowIsAdmissible(
          row.rawRow,
          row.recordId ?? "",
          inputRef.current.readScope,
        )
      )
        return false;
      const links = readTimelineEvidenceLinks(row);
      return (
        links.kind === "available" &&
        links.items.some(
          (link) =>
            link.evidenceRecordId === target.recordId &&
            targetFor(row, link).identity === target.identity,
        )
      );
    },
    onAccessFailure: input.onAccessFailure,
    onRestoreFocus: (target) => {
      if (inputRef.current.selectedRowId !== target.sourceRecordId) return;
      const row = target.sourceRecordId
        ? currentRow(target.sourceRecordId)
        : null;
      if (
        row?.recordId &&
        row.rowVersion !== null &&
        inputRef.current.elementRegistry.focusEvidenceList({
          recordId: row.recordId,
          rowVersion: row.rowVersion,
          viewSchemaId: "cartulary.view.timeline.v2",
        })
      )
        return;
      if (row?.recordId) inputRef.current.onRestoreGridFocus(row.recordId);
      else if (target.sourceRecordId)
        inputRef.current.onRestoreGridFocus(target.sourceRecordId);
    },
  });
  const startSpace = useCallback(
    (row: WorkbookRow) => {
      cancelSpace();
      setSpaceIntent({
        row,
        sequence: intentSequence.current,
        scopeKey: inputRef.current.scopeKey,
      });
    },
    [cancelSpace],
  );
  const sourceKey = spaceIntent
    ? timelineEvidenceLinkKey(
        currentRow(spaceIntent.row.recordId ?? "") ?? spaceIntent.row,
      )
    : "";
  useEffect(() => {
    const readScope = inputRef.current.readScope;
    if (
      !spaceIntent ||
      !input.canRead ||
      !input.inspectorOpen ||
      !readScope ||
      spaceIntent.scopeKey !== input.scopeKey ||
      input.selectedRowId !== spaceIntent.row.recordId
    )
      return;
    const row = currentRow(spaceIntent.row.recordId ?? "");
    if (
      !row?.rawRow ||
      !spaceIntent.row.rawRow ||
      !workbookRowIsAdmissible(row.rawRow, row.recordId ?? "", readScope) ||
      !sameWorkbookReadScope(
        spaceIntent.row.rawRow.observation?.scope,
        readScope,
      ) ||
      sourceKey !== timelineEvidenceLinkKey(spaceIntent.row)
    ) {
      cancelSpace();
      return;
    }
    const links = readTimelineEvidenceLinks(row);
    if (links.kind !== "available") {
      setSpaceIntent(null);
      return;
    }
    const controller = new AbortController();
    probeController.current = controller;
    const sequence = spaceIntent.sequence;
    let settled = false;
    const stillCurrent = () =>
      !controller.signal.aborted &&
      intentSequence.current === sequence &&
      inputRef.current.selectedRowId === row.recordId &&
      inputRef.current.canRead &&
      inputRef.current.inspectorOpen &&
      sameWorkbookReadScope(
        inputRef.current.readScope,
        row.rawRow?.observation?.scope,
      ) &&
      timelineEvidenceLinkKey(currentRow(row.recordId ?? "") ?? row) ===
        timelineEvidenceLinkKey(row);
    const open = async () => {
      if (links.items.length === 1) {
        const link = links.items[0];
        if (link && stillCurrent())
          void access.issue(targetFor(row, link), "preview", null);
        probeController.current = null;
        settled = true;
        setSpaceIntent(null);
        return;
      }
      setSpaceState({ kind: "checking", sourceKey });
      const result = await resolveSolePreviewableEvidence(
        links.items.map((link) => link.evidenceRecordId),
        input.port,
        controller.signal,
        stillCurrent,
      );
      if (!stillCurrent()) return;
      probeController.current = null;
      settled = true;
      setSpaceState(
        result.kind === "indeterminate"
          ? { kind: "indeterminate", sourceKey }
          : null,
      );
      setSpaceIntent(null);
      if (result.kind === "sole") {
        const link = links.items.find(
          (item) => item.evidenceRecordId === result.recordId,
        );
        if (link) void access.issue(targetFor(row, link), "preview", null);
      }
    };
    void open();
    return () => {
      controller.abort();
      if (!settled && intentSequence.current === sequence) {
        intentSequence.current += 1;
        setSpaceIntent((current) =>
          current?.sequence === sequence ? null : current,
        );
        setSpaceState(null);
      }
    };
  }, [
    spaceIntent,
    input.canRead,
    input.inspectorOpen,
    input.scopeKey,
    input.selectedRowId,
    input.port,
    sourceKey,
    access.issue,
    cancelSpace,
    currentRow,
  ]);
  useEffect(() => {
    const cancel = () => {
      if (probeController.current) cancelSpace();
    };
    document.addEventListener("keydown", cancel, true);
    document.addEventListener("pointerdown", cancel, true);
    document.addEventListener("wheel", cancel, true);
    return () => {
      document.removeEventListener("keydown", cancel, true);
      document.removeEventListener("pointerdown", cancel, true);
      document.removeEventListener("wheel", cancel, true);
      probeController.current?.abort();
    };
  }, [cancelSpace]);
  const displayedRow = input.selectedRowId
    ? currentRow(input.selectedRowId)
    : null;
  const spaceStatus =
    spaceState !== null &&
    displayedRow !== null &&
    spaceState.sourceKey === timelineEvidenceLinkKey(displayedRow)
      ? spaceState.kind
      : "idle";
  return {
    access,
    startSpace,
    cancelSpace,
    spaceStatus,
    issue: (
      row: WorkbookRow,
      link: TimelineEvidenceLink,
      kind: "preview" | "download",
      invoker: HTMLElement,
    ) => {
      cancelSpace();
      return access.issue(targetFor(row, link), kind, invoker);
    },
  };
}
