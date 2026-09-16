import { indicatorCreateTestId } from "@cartulary/ui-contracts";
import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { IndicatorCreateOperationStatus } from "./IndicatorCreateOperationStatus";
import type { IndicatorCreateOwnerPort } from "./indicatorCreateOperation";
import { observationStack, observationText } from "./observationStyles";

export function WorkbookIndicatorCreateRecovery({
  owner,
}: {
  owner: IndicatorCreateOwnerPort;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: readonly WorkbookRecoveryItem[] = snapshot.authority
    ? snapshot.entries.map((entry, order) => ({
        id: entry.attempt.id,
        label: "Canonical Indicator creation",
        origin: entry.attempt.observation.observed_text,
        sheetRef: { kind: "view_schema", id: "cartulary.view.indicators.v1" },
        refreshViews:
          entry.receipt && entry.refresh !== "complete"
            ? ["cartulary.view.indicators.v1"]
            : [],
        order,
        summary: entry.receipt
          ? entry.refresh === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : entry.phase === "uncertain"
            ? "Outcome unconfirmed"
            : entry.phase === "rejected"
              ? "Review required"
              : "In progress",
        attention:
          entry.receipt && entry.refresh === "complete"
            ? "completed"
            : entry.receipt ||
                entry.phase === "uncertain" ||
                entry.phase === "rejected"
              ? "attention"
              : "progress",
      }))
    : [];
  const selected = useWorkbookRecoverySource("indicator-create", items);
  return (
    <WorkbookRecoveryDetail source="indicator-create" item={selected}>
      <section
        data-testid={indicatorCreateTestId("recovery")}
        aria-label="Canonical Indicator recovery"
      >
        {snapshot.entries
          .filter((entry) => entry.attempt.id === selected)
          .map((entry) => (
            <article key={entry.attempt.id} style={observationStack}>
              <p style={observationText}>
                Observed text: {entry.attempt.observation.observed_text}
              </p>
              <p style={observationText}>
                Source: {entry.attempt.observation.source_record_id}
              </p>
              <IndicatorCreateOperationStatus owner={owner} entry={entry} />
            </article>
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
