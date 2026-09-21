import { decisionSupersessionTestId } from "@cartulary/ui-contracts";
import {
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  workbookFormInputStyle as inputStyle,
  workbookFormFieldsStyle as sectionStyle,
  workbookFormHeadingStyle,
} from "../../components/workbookFormStyles";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import {
  type DecisionSupersessionReview,
  decisionConsequence,
  decisionIneligibility,
  normalizeDecisionReason,
  reviewedDecision,
} from "./decisionSupersessionModel";
import type { DecisionSupersessionOwnerPort } from "./decisionSupersessionOperation";
import { useDecisionCandidates } from "./useDecisionCandidates";

export function DecisionSupersessionEditor({
  owner,
  row,
  lifecycleKey,
  originSurface,
  onCancel,
  reconcile,
}: {
  readonly owner: DecisionSupersessionOwnerPort;
  readonly row: WorkbookQueryRow;
  readonly lifecycleKey: string;
  readonly originSurface: string;
  readonly onCancel: () => void;
  readonly reconcile: () => Promise<void>;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const candidates = useDecisionCandidates(
    owner,
    `${lifecycleKey}:${snapshot.generation}`,
  );
  const [replacementId, selectReplacement] = useState("");
  const [reason, setReason] = useState("");
  const [notice, setNotice] = useState<string | null>(null);
  const [, renderReview] = useState(0);
  const reviewRef = useRef<{
    value: DecisionSupersessionReview;
    key: string;
  } | null>(null);
  const alive = useRef(true);
  useEffect(() => {
    alive.current = true;
    return () => {
      alive.current = false;
    };
  }, []);
  const confirmButton = useRef<HTMLButtonElement>(null);
  const reviewButton = useRef<HTMLButtonElement>(null);
  const focusReview = useRef(false);
  const noticeElement = useRef<HTMLParagraphElement>(null);
  const focusNotice = useRef(false);
  const reasonElement = useRef<HTMLTextAreaElement>(null);
  const focusInvalidation = useRef(false);
  const incidentId = snapshot.authority?.incidentId ?? "";
  const cachedTarget = owner.latestRow(row.record_id);
  const target = reviewedDecision(
    cachedTarget && cachedTarget.row_version >= row.row_version
      ? cachedTarget
      : row,
    incidentId,
  );
  const replacementRow = candidates.rows.find(
    (candidate) => candidate.record_id === replacementId,
  );
  const replacement = replacementRow
    ? reviewedDecision(
        owner.latestRow(replacementId) ?? replacementRow,
        incidentId,
      )
    : null;
  const targetProblem = decisionIneligibility(target, "target");
  const replacementProblem = replacement
    ? decisionIneligibility(replacement, "replacement", target)
    : null;
  const normalized = normalizeDecisionReason(reason);
  const eligible =
    owner.canSubmit() &&
    !targetProblem &&
    replacement !== null &&
    !replacementProblem &&
    normalized !== null &&
    candidates.state === "ready" &&
    !owner.blocksRecord(target.recordId) &&
    !owner.blocksRecord(replacement.recordId) &&
    (owner.latestVersion(target.recordId) ?? target.baseRowVersion) ===
      target.baseRowVersion &&
    (owner.latestVersion(replacement.recordId) ??
      replacement.baseRowVersion) === replacement.baseRowVersion;
  const key = JSON.stringify([
    lifecycleKey,
    originSurface,
    snapshot.generation,
    target,
    replacement,
    reason,
    candidates.state,
    owner.latestVersion(target.recordId),
    replacement && owner.latestVersion(replacement.recordId),
  ]);
  const keyRef = useRef(key);
  keyRef.current = key;
  if (reviewRef.current && reviewRef.current.key !== key) {
    focusInvalidation.current =
      document.activeElement === confirmButton.current;
    reviewRef.current = null;
  }
  const reviewed = reviewRef.current?.value ?? null;
  const presentation = `${lifecycleKey}:${originSurface}:${snapshot.generation}`;
  const presentationRef = useRef(presentation);
  presentationRef.current = presentation;
  useLayoutEffect(() => {
    if (focusReview.current) {
      focusReview.current = false;
      (reviewed ? confirmButton : reviewButton).current?.focus({
        preventScroll: true,
      });
    }
    if (focusInvalidation.current) {
      focusInvalidation.current = false;
      reasonElement.current?.focus({ preventScroll: true });
    }
    if (focusNotice.current && notice !== null) {
      focusNotice.current = false;
      noticeElement.current?.focus({ preventScroll: true });
    }
  }, [reviewed, notice]);
  function invalidateInputs() {
    reviewRef.current = null;
    setNotice(null);
  }
  function review() {
    if (!eligible || !snapshot.authority || !replacement || normalized === null)
      return;
    reviewRef.current = {
      key,
      value: Object.freeze({
        target,
        replacement,
        reason: normalized,
        authority: Object.freeze({ ...snapshot.authority }),
        authorityGeneration: snapshot.generation,
        originSurface,
        lifecycleKey,
      }),
    };
    focusReview.current = true;
    renderReview((value) => value + 1);
  }
  function confirm() {
    const captured = reviewRef.current;
    if (!captured || captured.key !== keyRef.current || !eligible) return;
    const attempt = owner.admit(captured.value, {
      isCurrent: () =>
        alive.current && presentationRef.current === presentation,
      matchesReview: () =>
        reviewRef.current === captured && keyRef.current === captured.key,
      reconcile,
    });
    if (!attempt) {
      setNotice(
        "This review could not be admitted. Check current access and earlier writes, then review again.",
      );
      return;
    }
    focusNotice.current = true;
    setNotice(
      "Supersession admitted. Its progress and recovery remain in Recovery.",
    );
    void owner.execute(attempt);
  }
  if (!snapshot.authority) return null;
  return (
    <section
      aria-label="Decision supersession"
      data-testid={decisionSupersessionTestId("editor")}
      style={sectionStyle}
    >
      <h4 style={workbookFormHeadingStyle}>Supersede Decision</h4>
      <p>
        Target: {target.label} ({target.recordId}) — {target.status}
      </p>
      {targetProblem ? <p role="status">{targetProblem}</p> : null}
      <label>
        Superseding Decision
        <select
          aria-label="Superseding Decision"
          data-testid={decisionSupersessionTestId("replacement")}
          value={replacementId}
          style={inputStyle}
          onChange={(event) => {
            invalidateInputs();
            selectReplacement(event.target.value);
          }}
        >
          <option value="">Select a replacement</option>
          {candidates.rows.map((candidate) => {
            const record = reviewedDecision(
              owner.latestRow(candidate.record_id) ?? candidate,
              incidentId,
            );
            const problem = decisionIneligibility(
              record,
              "replacement",
              target,
            );
            return (
              <option
                key={record.recordId}
                value={record.recordId}
                disabled={problem !== null}
              >
                {record.label} ({record.recordId}) — {record.status}
                {record.isSuperseded ? ", superseded" : ""}
                {problem ? ` — ${problem}` : ""}
              </option>
            );
          })}
        </select>
      </label>
      <p role="status">
        {candidates.state === "loading"
          ? "Loading Decision candidates…"
          : (candidates.error ??
            (candidates.hasMore
              ? "More Decisions are available. These results are incomplete."
              : candidates.rows.some(
                    (candidate) =>
                      decisionIneligibility(
                        reviewedDecision(candidate, incidentId),
                        "replacement",
                        target,
                      ) === null,
                  )
                ? "All visible Decision candidates loaded."
                : "No eligible replacement Decisions are available."))}
      </p>
      {candidates.state === "failed" ? (
        <WorkbookInspectorActionButton onClick={() => void candidates.retry()}>
          Retry candidates
        </WorkbookInspectorActionButton>
      ) : null}
      {candidates.hasMore && candidates.nextCursor ? (
        <WorkbookInspectorActionButton
          disabled={candidates.state === "loading"}
          onClick={() => void candidates.more()}
        >
          Load more Decisions
        </WorkbookInspectorActionButton>
      ) : null}
      <WorkbookInspectorActionButton
        disabled={candidates.state === "loading"}
        onClick={() => void candidates.refresh()}
      >
        Refresh candidates
      </WorkbookInspectorActionButton>
      {replacementId && !replacement ? (
        <p role="status">
          {candidates.hasMore
            ? "The selected replacement is not in the loaded results yet."
            : "The selected replacement is no longer available."}
        </p>
      ) : replacementProblem ? (
        <p role="status">{replacementProblem}</p>
      ) : null}
      <label>
        Reason (required)
        <textarea
          ref={reasonElement}
          aria-label="Decision supersession reason"
          data-testid={decisionSupersessionTestId("reason")}
          aria-invalid={reason !== "" && normalized === null}
          value={reason}
          style={inputStyle}
          onChange={(event) => {
            invalidateInputs();
            setReason(event.target.value);
          }}
        />
      </label>
      {reason !== "" && normalized === null ? (
        <p role="alert">
          Enter a nonempty reason of at most 4096 Unicode characters, without
          unsupported controls.
        </p>
      ) : null}
      <p>
        {decisionConsequence(target.status)} The superseding Decision will link
        to this target. The reason explains this change; it is not an approval.
      </p>
      {reviewed ? (
        <section
          aria-label="Review Decision supersession"
          data-testid={decisionSupersessionTestId("review")}
        >
          <strong>Review supersession</strong>
          <p>
            Target: {reviewed.target.label} ({reviewed.target.recordId}),{" "}
            {reviewed.target.status}, version {reviewed.target.baseRowVersion}.
          </p>
          <p>
            Replacement: {reviewed.replacement.label} (
            {reviewed.replacement.recordId}), {reviewed.replacement.status},
            loaded version {reviewed.replacement.baseRowVersion}.
          </p>
          <p>
            {decisionConsequence(reviewed.target.status)} The replacement is
            checked again by the server.
          </p>
          <p style={{ whiteSpace: "pre-wrap" }}>Reason: {reviewed.reason}</p>
          <WorkbookInspectorActionButton
            ref={confirmButton}
            tone="destructive"
            disabled={!eligible}
            data-testid={decisionSupersessionTestId("confirm")}
            onClick={confirm}
          >
            Confirm supersession
          </WorkbookInspectorActionButton>
          <WorkbookInspectorActionButton
            onClick={() => {
              reviewRef.current = null;
              focusReview.current = true;
              renderReview((value) => value + 1);
            }}
          >
            Cancel review
          </WorkbookInspectorActionButton>
        </section>
      ) : (
        <WorkbookInspectorActionButton
          ref={reviewButton}
          disabled={!eligible}
          data-testid={decisionSupersessionTestId("review-action")}
          onClick={review}
        >
          Review supersession
        </WorkbookInspectorActionButton>
      )}
      {notice ? (
        <p ref={noticeElement} tabIndex={-1} role="status">
          {notice}
        </p>
      ) : null}
      <WorkbookInspectorActionButton onClick={onCancel}>
        Close supersession
      </WorkbookInspectorActionButton>
    </section>
  );
}
