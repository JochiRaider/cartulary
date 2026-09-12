import { coordinationWorkflowTestId } from "@cartulary/ui-contracts";
import { useLayoutEffect, useRef } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { WorkbookInspectorPublicError } from "../../inspector/presentation/WorkbookInspectorFeedback";
import { workbookInspectorErrorPresentation } from "../../inspector/workbookInspectorErrorModel";
import { PartyLinkControls } from "./PartyLinkControls";
import { PartyPatchFeedback } from "./PartyLinkRecovery";
import { partyInputStyle, partyStackStyle } from "./partyLinkStyles";
import type { useGenericPartyLinkWorkflow } from "./useGenericPartyLinkWorkflow";

export function PartyLinkPanel({
  workflow,
}: {
  workflow: ReturnType<typeof useGenericPartyLinkWorkflow>;
}) {
  const { owner, review, snapshot, pair, pairs, scopeKey } = workflow;
  const reader = owner.getReader();
  const root = useRef<HTMLDivElement>(null),
    selector = useRef<HTMLSelectElement>(null);
  const focusClaim = useRef<{ context: string; scope: string } | null>(null);
  useLayoutEffect(() => {
    if (
      focusClaim.current?.context === workflow.focusContext &&
      focusClaim.current.scope !== scopeKey &&
      document.activeElement === document.body
    )
      selector.current?.focus({ preventScroll: true });
    focusClaim.current = null;
    return () => {
      if (root.current?.contains(document.activeElement))
        focusClaim.current = {
          context: workflow.focusContext,
          scope: scopeKey,
        };
    };
  }, [scopeKey, workflow.focusContext]);
  if (!review || !pair || !reader) return null;
  const creations = snapshot.creations.filter(
    (entry) =>
      entry.attempt.review.source.record_id === review.source.record_id &&
      entry.attempt.review.pair.key === pair.key,
  );
  const patches = snapshot.patches.filter(
    (entry) =>
      entry.intent.baseline.record_id === review.source.record_id &&
      entry.intent.partyReview?.pair.key === pair.key,
  );
  return (
    <div ref={root} style={partyStackStyle}>
      <label style={partyStackStyle}>
        Party link field
        <select
          ref={selector}
          style={partyInputStyle}
          aria-label="Party link field"
          data-testid={coordinationWorkflowTestId("party-pair")}
          value={pair.key}
          onChange={(event) => workflow.selectPair(event.target.value)}
        >
          {pairs.map((pair) => (
            <option key={pair.key} value={pair.key}>
              {pair.label}
            </option>
          ))}
        </select>
      </label>
      <PartyLinkControls
        key={scopeKey}
        pair={pair}
        row={review.source}
        reader={reader}
        scopeKey={scopeKey}
        candidateRevision={snapshot.candidateRevision}
        disabled={
          !owner.canSubmit() ||
          owner.blocksRecord(review.source.record_id) ||
          owner.patches.blocksRecord(review.source.record_id)
        }
        onCreate={(draft) => void owner.create(review, draft)}
        onPatch={(action, target) => void owner.patch(review, action, target)}
      />
      {snapshot.preparationFailure?.presentation === review.presentation ? (
        <p role="status">{snapshot.preparationFailure.message}</p>
      ) : null}
      {creations.map((entry) => (
        <section key={entry.id} aria-label="Party creation result">
          <p role="status">
            {entry.receipt
              ? `Party saved: ${String(entry.receipt.data.row.cells["party.display_name"]?.value ?? "Party")}.`
              : entry.phase === "uncertain"
                ? "Party creation outcome is uncertain. The original request is retained."
                : (entry.failure?.message ?? "Saving Party…")}
          </p>
          {entry.failure ? (
            <WorkbookInspectorPublicError
              error={workbookInspectorErrorPresentation(entry.failure)}
            />
          ) : null}
          {entry.phase === "uncertain" ? (
            <WorkbookInspectorActionButton
              tone="secondary"
              disabled={!owner.canSubmit()}
              onClick={() => void owner.replayCreation(entry.id)}
            >
              Replay Party creation
            </WorkbookInspectorActionButton>
          ) : null}
          {entry.receipt &&
          (!entry.linkId ||
            ["rejected", "preparation_failed"].includes(
              patches.find((patch) => patch.id === entry.linkId)?.phase ?? "",
            )) ? (
            <>
              <p
                data-testid={coordinationWorkflowTestId(
                  "party-partial-completion",
                )}
              >
                The Party is saved. Review this source and link it without
                creating again.
              </p>
              <WorkbookInspectorActionButton
                tone="secondary"
                data-testid={coordinationWorkflowTestId(
                  "party-retry-created-link",
                )}
                disabled={
                  !owner.canSubmit() ||
                  owner.patches.blocksRecord(review.source.record_id)
                }
                onClick={() => void owner.linkCreated(entry.id, review)}
              >
                Link saved Party to this source
              </WorkbookInspectorActionButton>
            </>
          ) : null}
        </section>
      ))}
      {patches.map((entry) => (
        <PartyPatchFeedback key={entry.id} owner={owner} entry={entry} />
      ))}
    </div>
  );
}
