import {
  requireViewContract,
  type ViewContract,
  type ViewFieldContract,
} from "@cartulary/view-contracts";
import { describe, expect, it } from "vitest";
import {
  emptyGenericReferenceOptions,
  type GenericReferenceOptions,
  genericFieldUsesReferenceOptions,
  referenceOptionsForField,
} from "./workbookReferenceOptions";
import {
  commLogViewSchemaId,
  decisionsViewSchemaId,
  taskRequestsViewSchemaId,
} from "./workbookSurfaceRegistry";

function requireField(
  contract: ViewContract,
  fieldKey: string,
): ViewFieldContract {
  const field = contract.fieldMap[fieldKey];
  if (!field) {
    throw new Error(`Missing field ${fieldKey} on ${contract.viewSchemaId}`);
  }
  return field;
}

function referenceOptions(): GenericReferenceOptions {
  return {
    ...emptyGenericReferenceOptions(),
    allRecords: [
      { label: "Duplicate", recordId: "record-1", viewSchemaId: "notes" },
      { label: "Duplicate", recordId: "record-2", viewSchemaId: "notes" },
    ],
    decisions: [
      {
        label: "Decision",
        recordId: "decision-1",
        viewSchemaId: decisionsViewSchemaId,
      },
    ],
    parties: [{ label: "Party", recordId: "party-1", viewSchemaId: "parties" }],
    taskRequests: [
      {
        label: "Task",
        recordId: "task-1",
        viewSchemaId: taskRequestsViewSchemaId,
      },
    ],
  };
}

describe("workbookReferenceOptions", () => {
  it("returns same-incident reference buckets without relabeling or deduping identities", () => {
    const commLog = requireViewContract(commLogViewSchemaId);
    const decisions = requireViewContract(decisionsViewSchemaId);
    const options = referenceOptions();

    expect(
      referenceOptionsForField(
        requireField(commLog, "comm_log.audience_party_ids"),
        options,
      ),
    ).toEqual(options.parties);
    expect(
      referenceOptionsForField(
        requireField(commLog, "comm_log.decision_ids"),
        options,
      ),
    ).toEqual(options.decisions);
    expect(
      referenceOptionsForField(
        requireField(decisions, "decision.support_refs"),
        options,
      ),
    ).toEqual(options.allRecords);
  });

  it("detects fields that use reference choices and leaves scalar fields free text", () => {
    const commLog = requireViewContract(commLogViewSchemaId);
    expect(
      genericFieldUsesReferenceOptions(
        requireField(commLog, "comm_log.audience_party_ids"),
      ),
    ).toBe(true);
    expect(
      genericFieldUsesReferenceOptions(
        requireField(commLog, "comm_log.summary"),
      ),
    ).toBe(false);
  });
});
