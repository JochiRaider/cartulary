import {
  coordinationWorkflowTestId,
  genericCreateFieldTestId,
} from "@cartulary/ui-contracts";
import { requireViewContract } from "@cartulary/view-contracts";
import { useId, useLayoutEffect, useRef, useState } from "react";
import { GenericMutationControl } from "../../components/GenericMutationControl";
import { WorkbookRecordCandidatePicker } from "../../components/WorkbookRecordCandidatePicker";
import { WorkbookInspectorActionButton } from "../../inspector/presentation/WorkbookInspectorActions";
import { extractEmailFromPartyText } from "../../models/genericWorkbookModel";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";
import { buildGenericCreateRequest } from "../generic/genericCreateRequestBuilder";
import {
  type PartyAction,
  type PartyLinkReadPort,
  type PartyPair,
  partyCreateDraft,
  partyState,
  partyValues,
  partyViewId,
} from "./partyLinkModel";
import {
  partyInputStyle,
  partyStackStyle as stackStyle,
} from "./partyLinkStyles";
import { usePartyCandidates } from "./usePartyCandidates";

export function PartyLinkControls({
  pair,
  row,
  reader,
  scopeKey,
  candidateRevision = 0,
  disabled,
  onCreate,
  onPatch,
}: {
  pair: PartyPair;
  row: WorkbookQueryRow;
  reader: PartyLinkReadPort;
  scopeKey: string;
  candidateRevision?: number;
  disabled: boolean;
  onCreate: (draft: Readonly<Record<string, string>>) => void;
  onPatch: (action: PartyAction, target?: string) => void;
}) {
  const formId = useId();
  const candidates = usePartyCandidates(reader, scopeKey, candidateRevision);
  const createTrigger = useRef<HTMLButtonElement>(null),
    form = useRef<HTMLFormElement>(null),
    restoreCreateFocus = useRef(false);
  const [selection, setSelection] = useState({
    recordId: "",
    revision: candidateRevision,
  });
  const target =
    selection.revision === candidateRevision ? selection.recordId : "";
  const setTarget = (recordId: string) =>
    setSelection({ recordId, revision: candidateRevision });
  const [filter, setFilter] = useState("");
  const [creating, setCreating] = useState(false);
  useLayoutEffect(() => {
    if (creating)
      form.current
        ?.querySelector<HTMLElement>("input, select")
        ?.focus({ preventScroll: true });
    else if (restoreCreateFocus.current) {
      restoreCreateFocus.current = false;
      createTrigger.current?.focus({ preventScroll: true });
    }
  }, [creating]);
  const cancel = () => {
    restoreCreateFocus.current = true;
    setCreating(false);
  };
  const { text, reference } = partyValues(row, pair);
  const [draft, setDraft] = useState(() => partyCreateDraft(text));
  const [includeEmail, setIncludeEmail] = useState(false);
  const proposal = extractEmailFromPartyText(text);
  const contract = requireViewContract(partyViewId);
  const reviewed = {
    ...draft,
    "party.primary_email": includeEmail
      ? (draft["party.primary_email"] ?? "")
      : "",
  };
  const button = (
    id: Parameters<typeof coordinationWorkflowTestId>[0],
    label: string,
    unavailable: boolean,
    action: () => void,
  ) => (
    <WorkbookInspectorActionButton
      tone="secondary"
      ref={id === "party-create-from-text" ? createTrigger : undefined}
      data-testid={coordinationWorkflowTestId(id)}
      disabled={disabled || unavailable}
      onClick={action}
    >
      {label}
    </WorkbookInspectorActionButton>
  );
  const field = (key: string) => {
    const entry = contract.fieldMap[key];
    if (!entry?.createWritable) return null;
    return (
      <label key={key} htmlFor={`${formId}-${key}`} style={stackStyle}>
        {entry.label}
        <GenericMutationControl
          id={`${formId}-${key}`}
          ariaLabel={`${entry.label} value`}
          collectionMode="add"
          field={entry}
          testId={genericCreateFieldTestId(key)}
          value={draft[key] ?? ""}
          onChange={(value) =>
            setDraft((previous) => ({ ...previous, [key]: value }))
          }
        />
      </label>
    );
  };
  const linked = candidates.rows.find(
    (candidate) => candidate.record_id === reference,
  );
  return (
    <section aria-label={`${pair.label} Party`} style={stackStyle}>
      <p style={{ margin: 0 }}>{partyState(row, pair)}</p>
      <p
        style={{ margin: 0, whiteSpace: "pre-wrap", overflowWrap: "anywhere" }}
      >
        Source wording: {text || "None"}
      </p>
      <p style={{ margin: 0 }}>
        Party link:{" "}
        {reference
          ? String(
              linked?.cells["party.display_name"]?.value ??
                "Linked Party; label not currently loaded",
            )
          : "None"}
      </p>
      {button(
        "party-create-from-text",
        "Create party from text",
        !text || creating,
        () => {
          setDraft(partyCreateDraft(text));
          setIncludeEmail(false);
          setCreating(true);
        },
      )}
      {creating ? (
        <form
          ref={form}
          aria-label="Review Party creation"
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              cancel();
            }
          }}
          style={stackStyle}
          onSubmit={(event) => {
            event.preventDefault();
            if (
              !disabled &&
              buildGenericCreateRequest(contract, reviewed, "review")
            )
              onCreate(reviewed);
          }}
        >
          <p style={{ margin: 0 }}>
            Review the Party to save and link. An exact email or external
            reference match may reuse an existing Party without changing it.
            Source wording is preserved.
          </p>
          <fieldset
            disabled={disabled}
            style={{ ...stackStyle, border: 0, padding: 0 }}
          >
            {field("party.display_name")}
            {field("party.party_kind")}
            <label>
              <input
                type="checkbox"
                checked={includeEmail}
                onChange={(event) => {
                  setIncludeEmail(event.target.checked);
                  if (event.target.checked && !draft["party.primary_email"])
                    setDraft((previous) => ({
                      ...previous,
                      "party.primary_email": proposal ?? "",
                    }));
                }}
              />
              Include email{proposal ? ` proposal: ${proposal}` : " (optional)"}
            </label>
            {includeEmail ? field("party.primary_email") : null}
            <details>
              <summary>Optional Party details</summary>
              <div style={stackStyle}>
                {contract.fields
                  .filter(
                    (entry) =>
                      entry.createWritable &&
                      ![
                        "party.display_name",
                        "party.party_kind",
                        "party.primary_email",
                      ].includes(entry.fieldKey),
                  )
                  .map((entry) => field(entry.fieldKey))}
              </div>
            </details>
          </fieldset>
          <WorkbookInspectorActionButton
            tone="secondary"
            disabled={
              disabled ||
              !buildGenericCreateRequest(contract, reviewed, "review")
            }
            onClick={() => onCreate(reviewed)}
          >
            Save Party and link
          </WorkbookInspectorActionButton>
          <WorkbookInspectorActionButton tone="secondary" onClick={cancel}>
            Cancel Party review
          </WorkbookInspectorActionButton>
        </form>
      ) : null}
      <label style={stackStyle}>
        Filter loaded Parties
        <input
          style={partyInputStyle}
          value={filter}
          onChange={(event) => setFilter(event.target.value)}
        />
      </label>
      <WorkbookRecordCandidatePicker
        label="Existing party"
        testId={coordinationWorkflowTestId("party-existing")}
        selection="single"
        disabled={disabled || candidates.phase !== "ready"}
        candidates={candidates.rows
          .map((candidate) => ({
            recordId: candidate.record_id,
            displayText: String(
              candidate.cells["party.display_name"]?.value ??
                candidate.record_id,
            ),
          }))
          .filter(
            (candidate) =>
              candidate.recordId === target ||
              candidate.displayText
                .toLowerCase()
                .includes(filter.toLowerCase()),
          )}
        selectedRecordIds={target ? [target] : []}
        onSelectedRecordIdsChange={(ids) => setTarget(ids[0] ?? "")}
      />
      <p role="status" style={{ margin: 0 }}>
        {candidates.phase === "loading"
          ? "Loading Parties…"
          : (candidates.error ??
            (candidates.rows.length
              ? `${candidates.rows.length} Parties loaded.${candidates.hasMore ? " More Parties are available." : ""}`
              : "No eligible Parties found."))}
      </p>
      {candidates.phase === "failed" ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          onClick={() => {
            setTarget("");
            void candidates.retry();
          }}
        >
          Retry Party read
        </WorkbookInspectorActionButton>
      ) : null}
      {candidates.hasMore ? (
        <WorkbookInspectorActionButton
          tone="secondary"
          disabled={candidates.phase !== "ready"}
          onClick={() => void candidates.loadMore()}
        >
          Load more Parties
        </WorkbookInspectorActionButton>
      ) : null}
      <WorkbookInspectorActionButton
        tone="secondary"
        disabled={candidates.phase === "loading"}
        onClick={() => {
          setTarget("");
          void candidates.reload();
        }}
      >
        Reload Parties
      </WorkbookInspectorActionButton>
      {button(
        "party-link-existing",
        "Link existing party",
        !target ||
          candidates.phase !== "ready" ||
          !candidates.rows.some((row) => row.record_id === target) ||
          target === reference,
        () => onPatch("link", target),
      )}
      {button("party-clear-link", "Clear party link", !reference, () =>
        onPatch("clear_link"),
      )}
      {button("party-clear-text", "Clear party text", !text, () =>
        onPatch("clear_text"),
      )}
      {button("party-clear-both", "Clear both", !text && !reference, () =>
        onPatch("clear_both"),
      )}
    </section>
  );
}
