import { indicatorObservationTestId } from "@cartulary/ui-contracts";
import { useSyncExternalStore } from "react";
import {
  useWorkbookRecoverySource,
  WorkbookRecoveryDetail,
} from "../../../shared/WorkbookRecoveryBoundary";
import type { WorkbookRecoveryItem } from "../../../shared/workbookRecoveryNavigation";
import {
  workbookFormFieldsStyle as observationStack,
  workbookFormPreservedTextStyle as observationText,
} from "../../components/workbookFormStyles";
import { ObservationOperationStatus } from "./ObservationOperationStatus";
import {
  type ObservationOwnerPort,
  observationIntentSource,
} from "./observationOperation";
export function WorkbookObservationRecovery({
  owner,
}: {
  owner: ObservationOwnerPort;
}) {
  const snapshot = useSyncExternalStore(owner.subscribe, owner.getSnapshot);
  const items: readonly WorkbookRecoveryItem[] = snapshot.authority
    ? snapshot.entries.map((entry, order) => ({
        id: entry.attempt.id,
        label: "Indicator observation",
        origin:
          entry.attempt.intent.action === "create"
            ? entry.attempt.intent.selection.text
            : entry.attempt.intent.observation.observed_text,
        sheetRef: null,
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
  const selected = useWorkbookRecoverySource("observations", items);
  return (
    <WorkbookRecoveryDetail source="observations" item={selected}>
      <section
        data-testid={indicatorObservationTestId("recovery")}
        aria-label="Indicator observation recovery"
      >
        {snapshot.entries
          .filter((entry) => entry.attempt.id === selected)
          .map((entry) => (
            <article key={entry.attempt.id} style={observationStack}>
              <strong>{entry.attempt.intent.action} observation</strong>
              <p style={observationText}>
                {entry.attempt.intent.action === "create"
                  ? entry.attempt.intent.selection.text
                  : entry.attempt.intent.observation.observed_text}
              </p>
              <p style={observationText}>
                Source: {observationIntentSource(entry.attempt.intent)}
              </p>
              <ObservationOperationStatus owner={owner} entry={entry} />
            </article>
          ))}
      </section>
    </WorkbookRecoveryDetail>
  );
}
