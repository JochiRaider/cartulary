import { ChevronDown, ChevronRight } from "lucide-react";
import {
  type FocusEvent,
  type RefObject,
  useEffect,
  useId,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  workbookQuietCommandStyle,
  workbookTypography,
} from "../../components/workbookFormStyles";
import { EvidenceFileRecovery } from "../../features/evidence/EvidenceFileRecovery";
import type {
  TimelineFileSnapshot,
  WorkbookTimelineFileOwner,
} from "../../features/evidence/WorkbookTimelineFileOwner";
import { workbookSurfaceFeedbackStyle } from "../../layout/WorkbookSurfaceLayout";
import { visuallyHiddenStyle } from "../../utils/workbookStyles";

function concealed(element: HTMLElement) {
  const closed = element.closest("details:not([open])");
  return !!closed && !closed.querySelector("summary")?.contains(element);
}

/** Restore only a control this region lost; never supersede newer external focus. */
function useAttachmentFocus(fallback: RefObject<HTMLElement | null>) {
  const region = useRef<HTMLDivElement>(null);
  const focused = useRef<{
    element: HTMLElement;
    workId: string | undefined;
  } | null>(null);
  useLayoutEffect(() => {
    const lost = focused.current;
    if (!lost || (lost.element.isConnected && !concealed(lost.element))) return;
    if (
      document.activeElement !== document.body &&
      document.activeElement !== lost.element
    )
      return;
    const work = lost.workId
      ? region.current?.querySelector<HTMLElement>(
          `[data-attachment-work-id="${CSS.escape(lost.workId)}"]`,
        )
      : null;
    const target = work && !concealed(work) ? work : fallback.current;
    focused.current = null;
    target?.focus();
  });
  return {
    ref: region,
    onFocusCapture: (event: FocusEvent<HTMLElement>) => {
      if (event.target instanceof HTMLElement)
        focused.current = {
          element: event.target,
          workId: event.target.closest<HTMLElement>("[data-attachment-work-id]")
            ?.dataset.attachmentWorkId,
        };
    },
    onBlurCapture: (event: FocusEvent<HTMLElement>) => {
      if (
        event.relatedTarget instanceof HTMLElement &&
        event.relatedTarget !== document.body &&
        !event.currentTarget.contains(event.relatedTarget)
      )
        focused.current = null;
    },
  };
}

/** Timeline-specific binding; the shared Evidence renderer retains its other consumers. */
export function TimelineAttachmentDetails({
  owner,
  files,
  presentation,
  fallbackRef,
}: {
  readonly owner: WorkbookTimelineFileOwner;
  readonly files: readonly TimelineFileSnapshot[];
  readonly presentation: "grid" | "inspector";
  readonly fallbackRef: RefObject<HTMLElement | null>;
}) {
  const focus = useAttachmentFocus(fallbackRef);
  const completed = files.filter((entry) => entry.feedbackKind === "completed");
  const render = (entry: TimelineFileSnapshot) => (
    <section
      key={entry.workId}
      data-attachment-work-id={entry.workId}
      data-evidence-work-id={entry.workId}
      tabIndex={-1}
      aria-label={`Attachment detail: ${entry.filename}`}
    >
      <EvidenceFileRecovery
        {...entry}
        presentation={presentation}
        canDiscard={entry.feedbackKind !== "completed" && entry.canDiscard}
        source={entry.sourceLabel}
        onConfirmReview={() => owner.confirmReview(entry.key)}
        onReview={() => void owner.review(entry.key)}
        onResume={() => void owner.resume(entry.key)}
        onFreshSlot={() => owner.freshSlot(entry.key)}
        onNewId={() => owner.newRequestId(entry.key)}
        onDiscard={() => owner.discard(entry.key)}
        onRefresh={() => void owner.refresh(entry.key)}
      />
    </section>
  );
  return (
    <div {...focus}>
      {files.filter((entry) => entry.feedbackKind !== "completed").map(render)}
      {completed.length ? (
        <details>
          <summary
            style={{
              ...workbookQuietCommandStyle,
              display: "list-item",
              whiteSpace: "normal",
            }}
          >
            Completed attachments ({completed.length})
          </summary>
          {completed.map(render)}
        </details>
      ) : null}
    </div>
  );
}

/** Only this leaf observes Timeline file progress; records still render through the grid owner. */
export function TimelineAttachmentFeedback({
  owner,
  fallbackRef,
}: {
  readonly owner: WorkbookTimelineFileOwner;
  readonly fallbackRef: RefObject<HTMLElement | null>;
}) {
  const files = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const admission = useSyncExternalStore(
    owner.subscribe,
    owner.getAdmissionNotice,
  );
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const DisclosureIcon = open ? ChevronDown : ChevronRight;
  const focus = useAttachmentFocus(fallbackRef);
  const detailsId = useId();
  const noticeId = useId();
  // Initial retained outcomes are readable, but mounting/reattaching does not announce them again.
  const seen = useRef({
    owner,
    outcomes: new Map(
      files.map((entry) => [entry.workId, entry.outcomeIdentity]),
    ),
    admission: admission?.identity,
  });
  const [emission, setEmission] = useState<{
    identity: string;
    message: string;
    priority: "assertive" | "polite";
  } | null>(null);
  useEffect(() => {
    if (seen.current.owner !== owner) {
      seen.current = {
        owner,
        outcomes: new Map(
          files.map((entry) => [entry.workId, entry.outcomeIdentity]),
        ),
        admission: admission?.identity,
      };
      setEmission(null);
      setOpen(false);
      return;
    }
    const changed = files.filter(
      (entry) =>
        seen.current.outcomes.get(entry.workId) !== entry.outcomeIdentity,
    );
    const admissionChanged =
      admission && admission.identity !== seen.current.admission;
    seen.current.outcomes = new Map(
      files.map((entry) => [entry.workId, entry.outcomeIdentity]),
    );
    seen.current.admission = admission?.identity;
    const attention = changed.filter(
      (entry) => entry.feedbackKind === "needs_attention",
    );
    const progress = changed.filter(
      (entry) => entry.feedbackKind === "in_progress",
    );
    const announced = attention.length ? attention : progress;
    if (admissionChanged)
      setEmission({
        identity: [
          `admission:${admission.identity}`,
          ...attention.map((entry) => entry.outcomeIdentity),
        ].join("|"),
        message: [
          admission.message,
          ...attention.map((entry) => `${entry.filename}: ${entry.message}`),
        ].join(" "),
        priority: "assertive",
      });
    else if (announced.length)
      setEmission({
        identity: announced.map((entry) => entry.outcomeIdentity).join("|"),
        message: announced
          .map((entry) => `${entry.filename}: ${entry.message}`)
          .join(" "),
        priority: attention.length ? "assertive" : "polite",
      });
    else if (!files.length || changed.length) setEmission(null);
  }, [owner, files, admission]);
  const attention = files.filter(
    (entry) => entry.feedbackKind === "needs_attention",
  ).length;
  const progress = files.filter(
    (entry) => entry.feedbackKind === "in_progress",
  ).length;
  const completed = files.length - attention - progress;
  const visible = files.length > 0 || admission !== null;
  // Conceal prior live copy in the same render as its owner snapshot, before effects run.
  const liveEmission =
    visible && seen.current.owner === owner ? emission : null;
  return (
    <>
      {visible ? (
        <section
          {...focus}
          aria-label="Timeline attachments"
          style={{
            ...workbookSurfaceFeedbackStyle,
            flexShrink: 0,
            scrollPaddingBlock: "var(--ct-spacing-xs)",
            ...workbookTypography("ui"),
          }}
          onKeyDown={(event) => {
            if (event.key === "Escape" && open) {
              event.preventDefault();
              event.stopPropagation();
              setOpen(false);
              trigger.current?.focus();
            }
          }}
        >
          {admission ? (
            <div
              id={noticeId}
              style={{
                padding: "var(--ct-spacing-xs)",
                overflowWrap: "anywhere",
                flexShrink: 0,
              }}
            >
              {admission.message}
            </div>
          ) : null}
          {files.length ? (
            <>
              <button
                ref={trigger}
                type="button"
                aria-expanded={open}
                aria-controls={detailsId}
                aria-describedby={admission ? noticeId : undefined}
                style={{
                  ...workbookQuietCommandStyle,
                  justifyContent: "flex-start",
                  whiteSpace: "normal",
                  flexShrink: 0,
                }}
                onClick={() => setOpen((value) => !value)}
              >
                <DisclosureIcon
                  aria-hidden="true"
                  style={{
                    inlineSize: "var(--ct-component-icon-inline-size)",
                    blockSize: "var(--ct-component-icon-inline-size)",
                    flexShrink: 0,
                  }}
                />
                <span>
                  Attachments: {attention} need attention, {progress} in
                  progress, {completed} completed
                </span>
              </button>
              <div id={detailsId} hidden={!open} style={{ flexShrink: 0 }}>
                {open ? (
                  <TimelineAttachmentDetails
                    owner={owner}
                    files={files}
                    presentation="grid"
                    fallbackRef={trigger}
                  />
                ) : null}
              </div>
            </>
          ) : null}
        </section>
      ) : null}
      <div style={visuallyHiddenStyle}>
        <div role="alert" aria-live="assertive" aria-atomic="true">
          <span
            key={
              liveEmission?.priority === "assertive"
                ? liveEmission.identity
                : "idle"
            }
          >
            {liveEmission?.priority === "assertive" ? liveEmission.message : ""}
          </span>
        </div>
        <div role="status" aria-live="polite" aria-atomic="true">
          <span
            key={
              liveEmission?.priority === "polite"
                ? liveEmission.identity
                : "idle"
            }
          >
            {liveEmission?.priority === "polite" ? liveEmission.message : ""}
          </span>
        </div>
      </div>
    </>
  );
}
