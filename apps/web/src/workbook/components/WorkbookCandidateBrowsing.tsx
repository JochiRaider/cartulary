import { useEffect, useId, useLayoutEffect, useRef, useState } from "react";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookCandidate } from "../ports/WorkbookCandidateReadPort";
import type {
  WorkbookCandidateDiscovery,
  WorkbookCandidateDiscoverySnapshot,
} from "../services/WorkbookCandidateDiscovery";

type ReadAction = "first" | "previous" | "next" | "refresh" | "retry";
type ReadIntent<T extends WorkbookCandidate> = {
  readonly action: ReadAction;
  readonly controller: WorkbookCandidateDiscovery<T>;
  readonly scope: string;
  readonly trigger: HTMLButtonElement;
  owned: boolean;
};

export function workbookCandidateAuthorizationMessage(
  failure: WorkbookOperationFailure | null,
) {
  return failure?.kind === "authentication_required"
    ? "Session authorization needs recovery."
    : "Incident access needs verification.";
}

/** Read-only local recovery. These actions cannot dispatch parent mutations. */
export function WorkbookCandidateBrowsing<T extends WorkbookCandidate>({
  discovery,
}: {
  readonly discovery: WorkbookCandidateDiscoverySnapshot<T> & {
    readonly controller: WorkbookCandidateDiscovery<T>;
    readonly canRead: boolean;
    readonly enabled: boolean;
  };
}) {
  const { page, pending, failure, controller, canRead, enabled } = discovery;
  const statusId = useId();
  const intentRef = useRef<ReadIntent<T> | null>(null);
  const admissionRef = useRef<WorkbookCandidateDiscovery<T> | null>(null);
  const [shownIntent, setShownIntent] = useState<ReadIntent<T> | null>(null);
  const restartRequired =
    failure?.kind === "invalid_contract" ||
    failure?.publicReason?.startsWith("cursor_");
  const retryable = Boolean(
    failure && !discovery.unavailable && !restartRequired,
  );
  const intent =
    shownIntent === intentRef.current &&
    shownIntent?.controller === controller &&
    shownIntent.scope === discovery.scope &&
    canRead &&
    enabled
      ? shownIntent
      : null;
  const focusedIntent =
    intent?.owned &&
    intent.trigger.isConnected &&
    typeof document !== "undefined" &&
    document.activeElement === intent.trigger;

  useEffect(() => {
    if (!shownIntent) return;
    const retire = (event: Event) => {
      const current = intentRef.current;
      if (!current) return;
      if (
        event.target === current.trigger &&
        (event.type === "focusin" ||
          event.type === "pointerdown" ||
          (event instanceof KeyboardEvent &&
            (event.key === "Enter" || event.key === " ")))
      )
        return;
      current.owned = false;
    };
    const events = [
      "focusin",
      "keydown",
      "pointerdown",
      "input",
      "wheel",
      "scroll",
      "touchstart",
    ] as const;
    for (const type of events) document.addEventListener(type, retire, true);
    return () => {
      for (const type of events)
        document.removeEventListener(type, retire, true);
    };
  }, [shownIntent]);
  useEffect(
    () => () => {
      intentRef.current = null;
      admissionRef.current = null;
    },
    [],
  );
  useLayoutEffect(() => {
    const current = intentRef.current;
    if (!current) return;
    const sameScope =
      current.controller === controller &&
      current.scope === discovery.scope &&
      canRead &&
      enabled;
    const capable =
      (current.action === "first" || current.action === "refresh") && enabled
        ? true
        : current.action === "previous"
          ? discovery.previousCount > 0
          : current.action === "next"
            ? Boolean(page?.hasMore)
            : retryable;
    const retainedRetry =
      current.action === "retry" && !capable && current.trigger.isConnected;
    if (
      !sameScope ||
      (!pending &&
        !retainedRetry &&
        (capable ||
          !current.owned ||
          !current.trigger.isConnected ||
          document.activeElement !== current.trigger))
    ) {
      intentRef.current = null;
      setShownIntent(null);
    }
  });

  const capability = (action: ReadAction) =>
    enabled &&
    (action === "previous"
      ? discovery.previousCount > 0
      : action === "next"
        ? Boolean(page?.hasMore)
        : action === "retry"
          ? retryable
          : true);
  const controlState = (action: ReadAction) => {
    const available = capability(action);
    const initiating = intent?.action === action;
    const retained = initiating && (pending || focusedIntent);
    return {
      disabled: !available && !retained,
      "aria-disabled": pending || !available,
      "aria-busy": pending && initiating,
      "aria-describedby": statusId,
      onBlur: (event: React.FocusEvent<HTMLButtonElement>) => {
        const current = intentRef.current;
        if (current?.trigger !== event.currentTarget) return;
        current.owned = false;
        if (!pending) {
          if (action === "retry" && !available && canRead && enabled) {
            // Pointer activation needs stable geometry until the new click lands.
            const retained = { ...current, owned: false };
            intentRef.current = retained;
            setShownIntent(retained);
          } else {
            intentRef.current = null;
            setShownIntent(null);
          }
        }
      },
    };
  };
  const invoke = (action: ReadAction, trigger: HTMLButtonElement) => {
    if (
      !canRead ||
      !capability(action) ||
      pending ||
      controller.getSnapshot().pending ||
      controller.getSnapshot().concealed ||
      admissionRef.current === controller
    )
      return;
    admissionRef.current = controller;
    const read = controller[action]();
    if (controller.getSnapshot().pending) {
      const next = {
        action,
        controller,
        scope: discovery.scope,
        trigger,
        owned: document.activeElement === trigger,
      } satisfies ReadIntent<T>;
      intentRef.current = next;
      setShownIntent(next);
    }
    void read.then(
      () => {
        if (admissionRef.current === controller) admissionRef.current = null;
      },
      () => {
        if (admissionRef.current === controller) admissionRef.current = null;
      },
    );
  };
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-xs)", minWidth: 0 }}>
      <p id={statusId} role={failure ? "alert" : "status"}>
        {!canRead
          ? workbookCandidateAuthorizationMessage(failure)
          : !enabled
            ? "Candidate discovery is paused."
            : discovery.unavailable
              ? "This candidate source is unavailable. Selected identities are retained."
              : failure
                ? `${page ? "The accepted page remains available. " : ""}${failure.message}`
                : pending
                  ? page
                    ? "Loading candidates. The accepted page remains available."
                    : "Loading candidates…"
                  : page
                    ? page.candidates.length === 0
                      ? "No candidates match this query."
                      : `Page ${discovery.pageNumber}: ${page.candidates.length} candidates; ${page.hasMore ? "more available" : "end of this query"}.`
                    : "No candidate page loaded."}
      </p>
      {restartRequired ? (
        <p>Restart this query with First candidates.</p>
      ) : null}
      {canRead ? (
        <div
          style={{
            display: "flex",
            flexWrap: "wrap",
            gap: "var(--ct-spacing-xs)",
          }}
        >
          <Button
            type="button"
            tone="secondary"
            {...controlState("first")}
            onClick={(event) => invoke("first", event.currentTarget)}
          >
            First candidates
          </Button>
          <Button
            type="button"
            tone="secondary"
            {...controlState("previous")}
            onClick={(event) => invoke("previous", event.currentTarget)}
          >
            Previous candidates
          </Button>
          <Button
            type="button"
            tone="secondary"
            {...controlState("next")}
            onClick={(event) => invoke("next", event.currentTarget)}
          >
            Next candidates
          </Button>
          <Button
            type="button"
            tone="secondary"
            {...controlState("refresh")}
            onClick={(event) => invoke("refresh", event.currentTarget)}
          >
            Refresh candidates
          </Button>
          {retryable || intent?.action === "retry" ? (
            <Button
              type="button"
              tone="secondary"
              {...controlState("retry")}
              onClick={(event) => invoke("retry", event.currentTarget)}
            >
              Retry candidates
            </Button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
