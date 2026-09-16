import { indicatorLifecycleTestId } from "@cartulary/ui-contracts";
import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import { IndicatorLifecycleOperationStatus } from "./IndicatorLifecycleOperationStatus";
import type { IndicatorLifecycleOwnerPort } from "./indicatorLifecycleOperation";

export function WorkbookIndicatorLifecycleRecovery({
  owner,
}: {
  owner: IndicatorLifecycleOwnerPort;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: readonly WorkbookRecoveryItem[] = snapshot.authority
    ? snapshot.entries.map((entry, order) => ({
        id: entry.attempt.id,
        label: "Indicator interval",
        origin: entry.attempt.draft.label,
        sheetRef: { kind: "view_schema", id: "cartulary.view.indicators.v1" },
        refreshViews:
          entry.receipt && entry.reconciliation !== "complete"
            ? ["cartulary.view.indicators.v1"]
            : [],
        order,
        summary: entry.receipt
          ? entry.reconciliation === "complete"
            ? "Completed"
            : "Saved; refresh required"
          : entry.phase === "uncertain"
            ? "Outcome unconfirmed"
            : entry.phase === "rejected"
              ? "Review required"
              : "In progress",
        attention:
          entry.receipt && entry.reconciliation === "complete"
            ? "completed"
            : entry.receipt ||
                entry.phase === "uncertain" ||
                entry.phase === "rejected"
              ? "attention"
              : "progress",
      }))
    : [];
  const selected = useWorkbookRecoverySource("indicator-lifecycle", items);
  return (
    <WorkbookRecoveryDetail source="indicator-lifecycle" item={selected}>
      <section
        data-testid={indicatorLifecycleTestId("recovery")}
        aria-label="Indicator interval recovery"
      >
        {snapshot.entries
          .filter((entry) => entry.attempt.id === selected)
          .map((entry) => (
            <article
              key={entry.attempt.id}
              aria-label={`Interval for ${entry.attempt.draft.label}`}
            >
              <strong>{entry.attempt.draft.label}</strong>
              <p>
                {entry.attempt.values.lifecycle_state.replaceAll("_", " ")} from{" "}
                {entry.attempt.values.valid_from} (UTC)
              </p>
              <IndicatorLifecycleOperationStatus owner={owner} entry={entry} />
            </article>
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
