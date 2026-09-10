import { networkAnalysisTestId } from "@cartulary/ui-contracts";
import type { CSSProperties } from "react";
import { NetworkFlowButton } from "./NetworkFlowControls";
import type { NetworkFlowPageNavigation } from "./useNetworkFlowPagedQuery";

export type NetworkFlowPageFeedback = {
  readonly message: string;
  readonly retained: boolean;
};

export function pageFailureFeedback(
  page: NetworkFlowPageNavigation,
): NetworkFlowPageFeedback {
  const attempt = page.failed;
  const failure =
    attempt?.command === "restart"
      ? "Could not restart at page one."
      : attempt?.command === "refresh"
        ? `Could not refresh page ${attempt.destination}.`
        : attempt
          ? `Could not load page ${attempt.destination}.`
          : "Results are unavailable.";
  const retained = page.pageNumber !== null;
  const guidance =
    page.recovery === "correct_request"
      ? "Review the applied query and correct the request."
      : page.recovery === "reduce_scope"
        ? "Reduce the query scope or limits."
        : "";
  return {
    retained,
    message: [
      failure,
      retained ? `Showing page ${page.pageNumber}.` : "",
      page.error?.message,
      guidance,
    ]
      .filter(Boolean)
      .join(" "),
  };
}

export function NetworkFlowQueryPagination({
  page,
  onRefreshResource,
}: {
  readonly page: NetworkFlowPageNavigation;
  readonly onRefreshResource: () => void;
}) {
  const pending = page.pending;
  const displayed =
    page.pageNumber === null ? "No page displayed" : `Page ${page.pageNumber}`;
  const progress =
    pending?.command === "restart"
      ? (page.notice ?? "Restarting at page one.")
      : pending?.command === "refresh"
        ? `Refreshing page ${pending.destination}`
        : pending
          ? `Loading page ${pending.destination}`
          : page.error === null
            ? page.notice
            : null;
  const recovery =
    page.recovery === "retry" && page.failed
      ? {
          label:
            page.failed.command === "refresh"
              ? "Retry refresh"
              : `Retry page ${page.failed.destination}`,
          invoke: page.retry,
          enabled: true,
        }
      : page.recovery === "refresh_resource"
        ? { label: "Refresh tables", invoke: onRefreshResource, enabled: true }
        : {
            label: "Restart query",
            invoke: page.restart,
            enabled:
              page.canRestart &&
              (page.pageNumber !== null || page.failed !== null),
          };
  const recoveryDisabled =
    !recovery.enabled ||
    (pending !== null &&
      (recovery.invoke !== page.restart || pending.command === "restart"));
  return (
    <nav
      aria-label="Network Flow result pages"
      className="network-flow-pagination"
      style={paginationStyle}
    >
      <NetworkFlowButton
        data-testid={networkAnalysisTestId("page-previous")}
        aria-disabled={!page.canPrevious || pending !== null}
        variant="secondary"
        onClick={page.previousPage}
      >
        Previous
      </NetworkFlowButton>
      <span
        aria-live="polite"
        aria-atomic="true"
        data-testid={networkAnalysisTestId("page-status")}
      >
        {displayed}
        {progress ? ` · ${progress}` : ""}
      </span>
      <NetworkFlowButton
        data-testid={networkAnalysisTestId("page-next")}
        aria-disabled={!page.canNext || pending !== null}
        variant="secondary"
        onClick={page.nextPage}
      >
        Next
      </NetworkFlowButton>
      <NetworkFlowButton
        aria-disabled={!page.canRefresh || pending !== null}
        variant="secondary"
        onClick={page.refresh}
      >
        Refresh page
      </NetworkFlowButton>
      <NetworkFlowButton
        aria-disabled={recoveryDisabled}
        variant="secondary"
        onClick={() => {
          if (!recoveryDisabled) recovery.invoke();
        }}
      >
        {recovery.label}
      </NetworkFlowButton>
    </nav>
  );
}

const paginationStyle = {
  display: "flex",
  alignItems: "center",
  justifyContent: "flex-end",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-sm)",
  padding: "var(--ct-spacing-xs) var(--ct-spacing-md)",
  borderBlockStart: "var(--ct-border-hairline)",
  background: "var(--ct-colors-surface-1)",
  fontSize: "0.8125rem",
} satisfies CSSProperties;
