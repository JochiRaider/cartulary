import { useCallback, useEffect, useMemo, useState } from "react";
import {
  normalizeGenericTextValue,
  type PartyLinkPair,
} from "../../models/genericWorkbookModel";
import type { GenericMutationCommandPort } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookOperationFailure } from "../../mutations/workbookOperationOutcome";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

type PartyLinkMutationOwner = {
  readonly beginMutation: () => void;
  readonly rejectMutationFailure: (failure: WorkbookOperationFailure) => void;
  readonly setValidationError: (message: string) => void;
};

export type GenericPartyLinkWorkflow = {
  readonly clearPartyBoth: () => Promise<void>;
  readonly clearPartyLink: () => Promise<void>;
  readonly clearPartyText: () => Promise<void>;
  readonly createPartyFromText: () => Promise<void>;
  readonly linkExistingParty: () => Promise<void>;
  readonly partialCompletionMessage: string | null;
  readonly partyLinkExistingPartyId: string;
  readonly retryCreatedPartyLink: () => Promise<void>;
  readonly selectedPartyLinkPair: PartyLinkPair | null;
  readonly setPartyLinkExistingPartyId: (value: string) => void;
  readonly setPartyLinkPairKey: (value: string) => void;
};

export function useGenericPartyLinkWorkflow({
  mutation,
  mutationCommands,
  originViewSchemaId,
  partyLinkPairs,
  resetKey,
  selectedRow,
  submitLinkPatch,
}: {
  readonly mutation: PartyLinkMutationOwner;
  readonly mutationCommands: GenericMutationCommandPort;
  readonly originViewSchemaId: string;
  readonly partyLinkPairs: readonly PartyLinkPair[];
  readonly resetKey: string;
  readonly selectedRow: WorkbookQueryRow | null;
  readonly submitLinkPatch: (
    changes: Array<Record<string, unknown>>,
    txnPrefix: string,
  ) => Promise<boolean>;
}): GenericPartyLinkWorkflow {
  const [partyLinkPairKey, setPartyLinkPairKey] = useState("");
  const [partyLinkExistingPartyId, setPartyLinkExistingPartyId] = useState("");
  const [createdPartyId, setCreatedPartyId] = useState<string | null>(null);
  const [partialCompletionMessage, setPartialCompletionMessage] = useState<
    string | null
  >(null);
  const selectedPartyLinkPair = useMemo(
    () =>
      partyLinkPairs.find((pair) => pair.key === partyLinkPairKey) ??
      partyLinkPairs[0] ??
      null,
    [partyLinkPairKey, partyLinkPairs],
  );

  useEffect(() => {
    setPartyLinkPairKey((current) =>
      partyLinkPairs.some((pair) => pair.key === current)
        ? current
        : (partyLinkPairs[0]?.key ?? ""),
    );
  }, [partyLinkPairs]);

  const selectedSubjectKey =
    selectedRow === null
      ? ""
      : `${selectedRow.record_id}:${selectedRow.row_version}:${selectedPartyLinkPair?.key ?? ""}`;
  const lifecycleKey = `${resetKey}:${selectedSubjectKey}`;
  useEffect(() => {
    void lifecycleKey;
    setPartyLinkExistingPartyId("");
    setCreatedPartyId(null);
    setPartialCompletionMessage(null);
  }, [lifecycleKey]);

  const submitCreatedPartyLink = useCallback(
    async (partyId: string) => {
      if (selectedPartyLinkPair === null) {
        mutation.setValidationError("Select a party field first.");
        return false;
      }
      const linked = await submitLinkPatch(
        [{ field_key: selectedPartyLinkPair.refFieldKey, value: partyId }],
        "party-link-created",
      );
      if (linked) {
        setCreatedPartyId(null);
        setPartialCompletionMessage(null);
        return true;
      }
      setCreatedPartyId(partyId);
      setPartialCompletionMessage(
        "The party was created, but the selected row was not linked. Retry the link after resolving the row conflict.",
      );
      return false;
    },
    [mutation, selectedPartyLinkPair, submitLinkPatch],
  );

  const createPartyFromText = useCallback(async () => {
    if (selectedRow === null || selectedPartyLinkPair === null) {
      mutation.setValidationError("Select a row and party field first.");
      return;
    }
    const rawText = normalizeGenericTextValue(
      String(
        selectedRow.cells[selectedPartyLinkPair.textFieldKey]?.value ?? "",
      ),
    );
    if (rawText === "") {
      mutation.setValidationError("Party text is empty.");
      return;
    }
    mutation.beginMutation();
    const created = await mutationCommands.createPartyFromText({
      originViewSchemaId,
      rawText,
    });
    if (created.kind === "rejected") {
      mutation.rejectMutationFailure(created.failure);
      return;
    }
    await submitCreatedPartyLink(created.value.row.record_id);
  }, [
    mutation,
    mutationCommands,
    originViewSchemaId,
    selectedPartyLinkPair,
    selectedRow,
    submitCreatedPartyLink,
  ]);

  const retryCreatedPartyLink = useCallback(async () => {
    if (createdPartyId === null) return;
    await submitCreatedPartyLink(createdPartyId);
  }, [createdPartyId, submitCreatedPartyLink]);

  const linkExistingParty = useCallback(async () => {
    if (selectedPartyLinkPair === null || partyLinkExistingPartyId === "") {
      mutation.setValidationError("Select an existing party.");
      return;
    }
    await submitLinkPatch(
      [
        {
          field_key: selectedPartyLinkPair.refFieldKey,
          value: partyLinkExistingPartyId,
        },
      ],
      "party-link-existing",
    );
  }, [
    mutation,
    partyLinkExistingPartyId,
    selectedPartyLinkPair,
    submitLinkPatch,
  ]);

  const clearPartyLink = useCallback(async () => {
    if (selectedPartyLinkPair === null) {
      mutation.setValidationError("Select a party field first.");
      return;
    }
    await submitLinkPatch(
      [{ field_key: selectedPartyLinkPair.refFieldKey, value: null }],
      "party-clear-link",
    );
  }, [mutation, selectedPartyLinkPair, submitLinkPatch]);

  const clearPartyText = useCallback(async () => {
    if (selectedPartyLinkPair === null) {
      mutation.setValidationError("Select a party field first.");
      return;
    }
    await submitLinkPatch(
      [{ field_key: selectedPartyLinkPair.textFieldKey, value: null }],
      "party-clear-text",
    );
  }, [mutation, selectedPartyLinkPair, submitLinkPatch]);

  const clearPartyBoth = useCallback(async () => {
    if (selectedPartyLinkPair === null) {
      mutation.setValidationError("Select a party field first.");
      return;
    }
    await submitLinkPatch(
      [
        { field_key: selectedPartyLinkPair.textFieldKey, value: null },
        { field_key: selectedPartyLinkPair.refFieldKey, value: null },
      ],
      "party-clear-both",
    );
  }, [mutation, selectedPartyLinkPair, submitLinkPatch]);

  return {
    clearPartyBoth,
    clearPartyLink,
    clearPartyText,
    createPartyFromText,
    linkExistingParty,
    partialCompletionMessage,
    partyLinkExistingPartyId,
    retryCreatedPartyLink,
    selectedPartyLinkPair,
    setPartyLinkExistingPartyId,
    setPartyLinkPairKey,
  };
}
