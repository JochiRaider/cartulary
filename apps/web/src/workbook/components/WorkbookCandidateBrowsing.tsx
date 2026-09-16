import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import type { WorkbookCandidate } from "../ports/WorkbookCandidateReadPort";
import type {
  WorkbookCandidateDiscovery,
  WorkbookCandidateDiscoverySnapshot,
} from "../services/WorkbookCandidateDiscovery";

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
  const restartRequired =
    failure?.kind === "invalid_contract" ||
    failure?.publicReason?.startsWith("cursor_");
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-xs)", minWidth: 0 }}>
      <p role={failure ? "alert" : "status"}>
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
            disabled={pending || !enabled}
            onClick={() => void controller.first()}
          >
            First candidates
          </Button>
          <Button
            type="button"
            tone="secondary"
            disabled={pending || !enabled || !discovery.previousCount}
            onClick={() => void controller.previous()}
          >
            Previous candidates
          </Button>
          <Button
            type="button"
            tone="secondary"
            disabled={pending || !enabled || !page?.hasMore}
            onClick={() => void controller.next()}
          >
            Next candidates
          </Button>
          <Button
            type="button"
            tone="secondary"
            disabled={pending || !enabled}
            onClick={() => void controller.refresh()}
          >
            Refresh candidates
          </Button>
          {failure && !discovery.unavailable && !restartRequired ? (
            <Button
              type="button"
              tone="secondary"
              disabled={pending || !enabled}
              onClick={() => void controller.retry()}
            >
              Retry candidates
            </Button>
          ) : null}
        </div>
      ) : null}
    </div>
  );
}
