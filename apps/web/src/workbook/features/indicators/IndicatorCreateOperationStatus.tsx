import { indicatorCreateTestId } from "@cartulary/ui-contracts";
import { useLayoutEffect, useRef } from "react";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import type {
  IndicatorCreateOperation,
  IndicatorCreateOwnerPort,
} from "./indicatorCreateOperation";
import { observationStack, observationText } from "./observationStyles";

export function IndicatorCreateOperationStatus({
  owner,
  entry,
}: {
  owner: IndicatorCreateOwnerPort;
  entry: IndicatorCreateOperation;
}) {
  const container = useRef<HTMLDivElement>(null),
    status = useRef<HTMLParagraphElement>(null);
  const retainFocus = container.current?.contains(document.activeElement);
  useLayoutEffect(() => {
    if (retainFocus && document.activeElement === document.body)
      status.current?.focus();
  });
  const pending =
    entry.transportPending ||
    entry.phase === "preparing" ||
    entry.phase === "submitting";
  return (
    <div
      ref={container}
      style={observationStack}
      data-testid={indicatorCreateTestId("result")}
    >
      <p ref={status} role="status" tabIndex={-1} style={observationText}>
        {entry.receipt
          ? `Indicator available: ${String(entry.receipt.row.cells["indicator.display_value"]?.value ?? "")}`
          : entry.phase === "uncertain"
            ? "Canonical create outcome unknown. The server may have committed it; the original request is retained."
            : entry.phase === "rejected"
              ? "Canonical create was not accepted. Your draft is retained."
              : entry.phase === "preparing"
                ? "Checking the persisted observation and access…"
                : "Submitting canonical create…"}
      </p>
      {entry.failure ? (
        <p style={observationText}>{entry.failure.message}</p>
      ) : null}
      {entry.receipt ? (
        <>
          <p style={observationText}>
            {entry.refresh === "complete"
              ? "Indicator and history refreshed."
              : entry.refresh === "refreshing"
                ? "Refreshing Indicator and history…"
                : "Indicator result retained. Refresh is still required."}
          </p>
          <details>
            <summary>Canonical create receipt</summary>
            <dl>
              <dt>Indicator record</dt>
              <dd>{entry.receipt.row.record_id}</dd>
              <dt>Returned version</dt>
              <dd>{entry.receipt.row.row_version}</dd>
              <dt>Change set</dt>
              <dd>{entry.receipt.response.data.change_set_id}</dd>
            </dl>
          </details>
        </>
      ) : null}
      {entry.phase === "uncertain" ? (
        <WorkbookInspectorActionButton
          disabled={pending || !owner.canReplay()}
          onClick={() => void owner.replay(entry.attempt.id)}
        >
          Replay original canonical create
        </WorkbookInspectorActionButton>
      ) : null}
      {entry.receipt && entry.refresh !== "complete" ? (
        <WorkbookInspectorActionButton
          disabled={pending || entry.refresh === "refreshing"}
          onClick={() => void owner.refresh(entry.attempt.id)}
        >
          Retry Indicator refresh
        </WorkbookInspectorActionButton>
      ) : null}
    </div>
  );
}
