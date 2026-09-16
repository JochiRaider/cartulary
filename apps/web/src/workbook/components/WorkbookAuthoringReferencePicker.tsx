import { getViewContract } from "@cartulary/view-contracts";
import { useCallback, useContext, useEffect, useRef, useState } from "react";
import { useWorkbookAuthoringInventory } from "../hooks/useWorkbookAuthoringInventory";
import {
  useWorkbookCandidateDiscovery,
  WorkbookCandidateAuthorityContext,
} from "../hooks/useWorkbookCandidateDiscovery";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type {
  WorkbookAuthoringReadPort,
  WorkbookAuthoringSelection,
} from "../ports/WorkbookAuthoringReadPort";
import type { WorkbookCandidateReader } from "../ports/WorkbookCandidateReadPort";
import {
  WorkbookCandidateBrowsing,
  workbookCandidateAuthorizationMessage,
} from "./WorkbookCandidateBrowsing";
import { WorkbookCandidateQueryControl } from "./WorkbookCandidateQueryControl";
import { WorkbookCandidateSelection } from "./WorkbookCandidateSelection";
import { inputStyle } from "./workbookGridControlStyles";

export type WorkbookAuthoringReferencePickerProps = Readonly<{
  label: string;
  targetKey: string;
  testId: string;
  views: readonly string[];
  initialView?: string;
  multiple: boolean;
  captureRowVersion?: boolean;
  maximum: number;
  selected: readonly WorkbookAuthoringSelection[];
  reader: Pick<WorkbookAuthoringReadPort, "page" | "availableViews">;
  revision: number;
  disabled: boolean;
  surfaceLabel?: string;
  regionLabel?: string;
  applyLabel?: string;
  cancelLabel?: string;
  onApply: (selected: readonly WorkbookAuthoringSelection[]) => void;
  onCancel: () => void;
}>;

/** Staging is local to the attached control; the parent alone applies or submits. */
export function WorkbookAuthoringReferencePicker(
  props: WorkbookAuthoringReferencePickerProps,
) {
  const authority = useContext(WorkbookCandidateAuthorityContext);
  const freshness = `${authority.identity}:${props.revision}`;
  const [staged, setStaged] = useState({ freshness, items: props.selected });
  let selected = staged.items;
  if (staged.freshness !== freshness) {
    selected = selected.map((item) => ({ ...item, displayText: "" }));
    setStaged({ freshness, items: selected });
  }
  const [view, setView] = useState(
    props.initialView ??
      props.selected[0]?.viewSchemaId ??
      props.views[0] ??
      "",
  );
  const [query, setQuery] = useState(emptyWorkbookQueryState);
  const inventory = useWorkbookAuthoringInventory(
    props.reader,
    props.views,
    props.targetKey,
    props.revision,
  );
  const read = useCallback<WorkbookCandidateReader<WorkbookAuthoringSelection>>(
    async (input) => {
      const result = await props.reader.page({ ...input, viewSchemaId: view });
      if (result.kind !== "accepted") return result;
      return {
        kind: "accepted",
        value: {
          ...result.value,
          candidates: result.value.candidates.map((item) => ({
            recordId: item.recordId,
            displayText: item.displayText,
            viewSchemaId: item.viewSchemaId,
            ...(props.captureRowVersion && item.row
              ? { rowVersion: item.row.row_version }
              : {}),
          })),
        },
      };
    },
    [props.reader, props.captureRowVersion, view],
  );
  const available = inventory.views.includes(view);
  const page = useWorkbookCandidateDiscovery(
    read,
    query,
    `${props.targetKey}:${view}`,
    props.revision,
    available && !props.disabled,
  );
  const concealed =
    !authority.canRead || page.concealed || inventory.discovery.concealed;
  const cancel = useRef<HTMLButtonElement>(null);
  useEffect(() => {
    cancel.current?.focus({ preventScroll: true });
  }, []);
  return (
    <section
      aria-label={props.regionLabel ?? `Choose ${props.label.toLowerCase()}`}
      style={{
        display: "grid",
        gap: "var(--ct-spacing-sm)",
        minWidth: 0,
        border: "var(--ct-border-hairline)",
        padding: "var(--ct-spacing-sm)",
      }}
      onKeyDown={(event) => {
        event.stopPropagation();
        if (event.key === "Escape") {
          event.preventDefault();
          props.onCancel();
        }
      }}
    >
      {props.views.length > 1 ? (
        <label>
          {props.surfaceLabel ?? "Reference surface"}
          <select
            aria-label={props.surfaceLabel ?? "Reference surface"}
            style={inputStyle}
            value={view}
            disabled={concealed}
            onChange={(event) => {
              setView(event.currentTarget.value);
              setQuery(emptyWorkbookQueryState());
            }}
          >
            {[...new Set([view, ...inventory.views])].map((id) => (
              <option key={id} value={id}>
                {getViewContract(id)?.title ?? "Members"}
              </option>
            ))}
          </select>
        </label>
      ) : null}
      {!inventory.membershipOnly &&
      (!inventory.discovery.page || inventory.discovery.failure) ? (
        <div>
          <p role={inventory.discovery.failure ? "alert" : "status"}>
            {inventory.discovery.concealed
              ? workbookCandidateAuthorizationMessage(
                  inventory.discovery.failure,
                )
              : (inventory.discovery.failure?.message ??
                "Loading reference surfaces…")}
          </p>
          {inventory.discovery.failure && !inventory.discovery.concealed ? (
            <Button
              type="button"
              tone="secondary"
              onClick={() => void inventory.discovery.controller.retry()}
            >
              Retry surfaces
            </Button>
          ) : null}
        </div>
      ) : null}
      {!available && inventory.discovery.page ? (
        <p role="status">
          This reference surface is unavailable. Your selection is retained.
        </p>
      ) : null}
      {view !== "incident_members" && getViewContract(view) && !concealed ? (
        <WorkbookCandidateQueryControl
          key={view}
          view={view}
          label={props.label}
          query={query}
          onApply={setQuery}
        />
      ) : null}
      <WorkbookCandidateSelection
        candidates={page.page?.candidates ?? []}
        selected={selected}
        multiple={props.multiple}
        maximum={props.maximum}
        label={props.label}
        testId={props.testId}
        disabled={props.disabled || concealed}
        concealed={concealed}
        onChange={(items) => setStaged({ freshness, items })}
      />
      {available ? <WorkbookCandidateBrowsing discovery={page} /> : null}
      <Button
        tone="secondary"
        type="button"
        disabled={
          props.disabled || concealed || selected.length > props.maximum
        }
        onClick={() => props.onApply(selected)}
      >
        {props.applyLabel ?? "Apply references"}
      </Button>
      <Button
        ref={cancel}
        tone="secondary"
        type="button"
        onClick={props.onCancel}
      >
        {props.cancelLabel ?? "Cancel references"}
      </Button>
    </section>
  );
}
