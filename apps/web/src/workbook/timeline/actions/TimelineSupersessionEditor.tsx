import { timelineCaptureActionTestId } from "@cartulary/ui-contracts";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { TimelineCandidatePort } from "./TimelineCandidatePort";
import {
  normalizeTimelineReason,
  type TimelineCaptureAuthority,
  type TimelineCaptureReview,
  type TimelineCaptureSubject,
  timelineCaptureIneligibility,
  timelineReplacementIneligibility,
  timelineSupersessionConsequence,
} from "./timelineCaptureActionModel";
import { useTimelineCandidates } from "./useTimelineCandidates";

export function TimelineSupersessionEditor({
  target,
  authority,
  generation,
  originKey,
  port,
  latestVersion,
  blocked,
  onPrepare,
  onConfirm,
  onCancel,
  onAccessLost,
}: {
  readonly target: TimelineCaptureSubject;
  readonly authority: TimelineCaptureAuthority;
  readonly generation: number;
  readonly originKey: string;
  readonly port: TimelineCandidatePort;
  readonly latestVersion: (id: string) => number | null;
  readonly blocked: boolean;
  readonly onPrepare: (
    review: TimelineCaptureReview,
    signal: AbortSignal,
  ) => Promise<boolean>;
  readonly onConfirm: (
    review: TimelineCaptureReview,
    isCurrent: () => boolean,
  ) => boolean;
  readonly onCancel: () => void;
  readonly onAccessLost: () => void;
}) {
  const candidates = useTimelineCandidates(
    port,
    `${originKey}:${generation}`,
    onAccessLost,
  );
  const [reason, setReason] = useState("");
  const [replacement, setReplacement] = useState<TimelineCaptureSubject | null>(
    null,
  );
  const [preparing, setPreparing] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [, renderReview] = useState(0);
  const reviewRef = useRef<{
    value: TimelineCaptureReview;
    key: string;
  } | null>(null);
  const preparation = useRef<AbortController | null>(null);
  const alive = useRef(true);
  const reasonInput = useRef<HTMLTextAreaElement>(null);
  const confirmButton = useRef<HTMLButtonElement>(null);
  const reviewButton = useRef<HTMLButtonElement>(null);
  const focusReview = useRef(false);
  const focusBack = useRef(false);
  useEffect(() => {
    alive.current = true;
    return () => {
      alive.current = false;
      preparation.current?.abort();
    };
  }, []);
  const selected =
    replacement === null
      ? null
      : (candidates.rows.find((row) => row.recordId === replacement.recordId) ??
        replacement);
  const normalized = normalizeTimelineReason(reason);
  const problem =
    timelineCaptureIneligibility(target, "supersede") ??
    (selected && timelineReplacementIneligibility(selected, target));
  const key = JSON.stringify([
    target,
    selected,
    reason,
    authority,
    generation,
    originKey,
    latestVersion(target.recordId),
    selected && latestVersion(selected.recordId),
  ]);
  const keyRef = useRef(key);
  keyRef.current = key;
  if (reviewRef.current && reviewRef.current.key !== key) {
    focusBack.current = document.activeElement === confirmButton.current;
    reviewRef.current = null;
  }
  const review = reviewRef.current?.value ?? null;
  const eligible =
    !blocked &&
    !preparing &&
    !authority.closed &&
    ["reviewer", "admin"].includes(authority.role) &&
    !!authority.actorId &&
    !!authority.sessionIdentity &&
    !problem &&
    normalized !== null &&
    (latestVersion(target.recordId) ?? target.rowVersion) <=
      target.rowVersion &&
    (selected === null ||
      (latestVersion(selected.recordId) ?? selected.rowVersion) <=
        selected.rowVersion);
  useLayoutEffect(() => {
    if (focusReview.current) {
      focusReview.current = false;
      confirmButton.current?.focus({ preventScroll: true });
    }
    if (focusBack.current) {
      focusBack.current = false;
      reasonInput.current?.focus({ preventScroll: true });
    }
  });
  async function prepare() {
    if (!eligible || preparation.current || normalized === null) return;
    const controller = new AbortController();
    preparation.current = controller;
    const capturedKey = keyRef.current;
    const value: TimelineCaptureReview = Object.freeze({
      action: "supersede",
      target,
      replacement: selected,
      reason: normalized,
      authority: Object.freeze({ ...authority }),
      authorityGeneration: generation,
      originKey,
    });
    setPreparing(true);
    setMessage(null);
    try {
      const ready = await onPrepare(value, controller.signal);
      if (!alive.current || controller.signal.aborted) return;
      if (!ready || keyRef.current !== capturedKey) {
        setMessage(
          "The row or reviewed inputs changed, or earlier edits need attention. Review the current row again. No supersession was sent.",
        );
        return;
      }
      reviewRef.current = { value, key: capturedKey };
      focusReview.current = true;
      renderReview((value) => value + 1);
    } catch {
      if (alive.current && !controller.signal.aborted)
        setMessage(
          "Earlier edits could not finish. No supersession was sent; your reason is retained.",
        );
    } finally {
      if (preparation.current === controller) preparation.current = null;
      if (alive.current) setPreparing(false);
    }
  }
  function confirm() {
    const captured = reviewRef.current;
    if (!captured || !eligible || captured.key !== keyRef.current) return;
    const isCurrent = () =>
      alive.current &&
      reviewRef.current === captured &&
      captured.key === keyRef.current;
    if (!onConfirm(captured.value, isCurrent))
      setMessage(
        "This action is unavailable. Check current access and earlier edits, then review again.",
      );
  }
  const selectedOnPage =
    selected &&
    candidates.rows.some((row) => row.recordId === selected.recordId);
  return (
    <section
      aria-label="Timeline supersession"
      data-testid={timelineCaptureActionTestId("editor", target.recordId)}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-sm)",
        paddingBlock: "var(--ct-spacing-sm)",
        overflowWrap: "anywhere",
      }}
    >
      <strong>Supersede Timeline row</strong>
      <p>
        Target: {target.label} ({target.recordId}), version {target.rowVersion}
      </p>
      <p>{timelineSupersessionConsequence}</p>
      {review ? (
        <>
          <section aria-label="Review Timeline supersession">
            <p>
              Supersede {review.target.label} ({review.target.recordId}),
              version {review.target.rowVersion}.
            </p>
            <p style={{ whiteSpace: "pre-wrap" }}>Reason: {review.reason}</p>
            <p>
              {review.replacement
                ? `Replacement: ${review.replacement.label} (${review.replacement.recordId})`
                : "No replacement will be linked."}
            </p>
            <p>
              The change and reason will be attributed to your account in
              History.
            </p>
          </section>
          <WorkbookInspectorActionButton
            ref={confirmButton}
            tone="destructive"
            disabled={!eligible}
            data-testid={timelineCaptureActionTestId(
              "confirm",
              target.recordId,
            )}
            onClick={confirm}
          >
            Confirm supersede
          </WorkbookInspectorActionButton>
          <WorkbookInspectorActionButton
            onClick={() => {
              reviewRef.current = null;
              focusBack.current = true;
              renderReview((value) => value + 1);
            }}
          >
            Back
          </WorkbookInspectorActionButton>
        </>
      ) : (
        <>
          <label>
            Reason for supersession
            <textarea
              ref={reasonInput}
              value={reason}
              aria-invalid={reason !== "" && normalized === null}
              data-testid={timelineCaptureActionTestId(
                "reason",
                target.recordId,
              )}
              style={{
                display: "block",
                boxSizing: "border-box",
                width: "100%",
                font: "inherit",
                color: "inherit",
                background: "var(--ct-colors-surface-1)",
              }}
              onChange={(event) => {
                setReason(event.target.value);
                setMessage(null);
              }}
            />
          </label>
          {reason !== "" && normalized === null ? (
            <p role="alert">
              Enter a non-empty reason of at most 4096 characters, without
              unsupported control characters.
            </p>
          ) : null}
          <label>
            Replacement Timeline row (optional)
            <select
              value={selected?.recordId ?? ""}
              data-testid={timelineCaptureActionTestId(
                "replacement",
                target.recordId,
              )}
              style={{
                display: "block",
                width: "100%",
                font: "inherit",
                color: "inherit",
                background: "var(--ct-colors-surface-1)",
              }}
              onChange={(event) => {
                setReplacement(
                  candidates.rows.find(
                    (row) => row.recordId === event.target.value,
                  ) ?? null,
                );
                setMessage(null);
              }}
            >
              <option value="">No replacement</option>
              {selected && !selectedOnPage ? (
                <option value={selected.recordId}>
                  {selected.label} · {selected.recordId}
                </option>
              ) : null}
              {candidates.rows
                .filter((row) => row.recordId !== target.recordId)
                .map((row) => (
                  <option
                    key={row.recordId}
                    value={row.recordId}
                    disabled={
                      timelineReplacementIneligibility(row, target) !== null
                    }
                  >
                    {row.label} · {row.context} · {row.recordId}
                    {row.captureState === "superseded" ? " (superseded)" : ""}
                  </option>
                ))}
            </select>
          </label>
          <p role="status">
            {candidates.state === "loading"
              ? "Loading Timeline replacements…"
              : candidates.state === "failed"
                ? candidates.error
                : candidates.rows.some(
                      (row) =>
                        timelineReplacementIneligibility(row, target) === null,
                    )
                  ? "Choose a replacement or leave No replacement selected."
                  : "No eligible replacements on this page."}
          </p>
          {candidates.state === "failed" ? (
            <WorkbookInspectorActionButton onClick={candidates.retry}>
              Retry replacements
            </WorkbookInspectorActionButton>
          ) : null}
          <div>
            <WorkbookInspectorActionButton
              disabled={
                candidates.state === "loading" || !candidates.canPrevious
              }
              onClick={candidates.previous}
            >
              Previous replacements
            </WorkbookInspectorActionButton>
            <WorkbookInspectorActionButton
              disabled={candidates.state !== "ready" || !candidates.nextCursor}
              onClick={candidates.next}
            >
              Next replacements
            </WorkbookInspectorActionButton>
          </div>
          <WorkbookInspectorActionButton
            ref={reviewButton}
            disabled={!eligible}
            onClick={() => void prepare()}
            data-testid={timelineCaptureActionTestId("review", target.recordId)}
          >
            Review supersession
          </WorkbookInspectorActionButton>
        </>
      )}
      {problem ? <p>{problem}</p> : null}
      {preparing ? (
        <p role="status">Waiting for earlier edits before review…</p>
      ) : null}
      {message ? <p role="status">{message}</p> : null}
      <WorkbookInspectorActionButton onClick={onCancel}>
        Cancel supersession
      </WorkbookInspectorActionButton>
    </section>
  );
}
