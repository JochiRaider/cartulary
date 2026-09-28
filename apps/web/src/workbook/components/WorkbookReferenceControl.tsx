import {
  getViewContract,
  type ReferenceFieldContract,
} from "@cartulary/view-contracts";
import {
  createContext,
  type RefCallback,
  useContext,
  useEffect,
  useId,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import { DecisionSupersessionContext } from "../features/coordination/DecisionSupersessionContext";
import { NoteCreateContext } from "../features/notes/NoteCreateContext";
import { WorkbookInspectorActionButton as Button } from "../inspector/presentation/WorkbookInspectorActions";
import { genericInspectorRowLabel } from "../models/genericWorkbookModel";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type { WorkbookOperationFailure } from "../mutations/workbookOperationOutcome";
import {
  type WorkbookReference,
  type WorkbookReferenceReadPort,
  workbookReferenceKey,
} from "../ports/WorkbookReferenceReadPort";
import type { WorkbookCommittedRecordPort } from "../query/WorkbookCommittedRecordPort";
import { WorkbookReferenceSelection } from "../services/WorkbookReferenceSelection";
import { useSelectedReferenceRemovalFocus } from "./useSelectedReferenceRemovalFocus";
import { menuStyle } from "./workbookGridControlStyles";

export const WorkbookReferenceContext = createContext<{
  readonly reader: WorkbookReferenceReadPort;
  readonly evidence: WorkbookCommittedRecordPort;
  readonly actorPresentation?:
    | Readonly<{
        userId: string;
        displayName: string;
      }>
    | undefined;
  readonly onAuthorityFailure: (failure: WorkbookOperationFailure) => void;
} | null>(null);
const noSubscribe = () => () => {};
const noSnapshot = () => null;
type Props = Readonly<{
  field: ReferenceFieldContract;
  label: string;
  value: string;
  sourceRecordId?: string | undefined;
  retained?:
    | readonly {
        readonly recordId: string;
        readonly displayText: string;
        readonly viewSchemaId: string;
      }[]
    | undefined;
  disabled?: boolean | undefined;
  compact?: boolean | undefined;
  invalid?: boolean | undefined;
  describedBy?: string | undefined;
  id?: string | undefined;
  testId: string;
  focusTargetRef?:
    | RefCallback<
        | HTMLInputElement
        | HTMLTextAreaElement
        | HTMLSelectElement
        | HTMLButtonElement
      >
    | undefined;
  onChange: (value: string) => void;
  onAccept: (references: readonly WorkbookReference[]) => void;
}>;

/** Raw parent input and an optional staged picker. Reading never edits that input. */
export function WorkbookReferenceControl(props: Props) {
  const boundary = useContext(WorkbookReferenceContext);
  const decisions = useContext(DecisionSupersessionContext);
  useSyncExternalStore(
    boundary?.evidence.subscribe ?? noSubscribe,
    boundary?.evidence.getSnapshot ?? noSnapshot,
  );
  useSyncExternalStore(
    decisions?.subscribe ?? noSubscribe,
    decisions?.getSnapshot ?? noSnapshot,
  );
  const [open, setOpen] = useState(false);
  const trigger = useRef<HTMLButtonElement>(null);
  const input = useRef<HTMLInputElement | HTMLTextAreaElement | null>(null);
  const panel = useRef<HTMLDivElement>(null);
  const panelId = useId();
  const present = (item: WorkbookReference): WorkbookReference => {
    if (
      item.identity.kind === "incident_member" &&
      item.identity.id === boundary?.actorPresentation?.userId
    )
      return {
        ...item,
        displayText: boundary.actorPresentation.displayName,
        presentation: "observed",
      };
    const row =
      decisions?.latestRow(item.identity.id) ??
      boundary?.evidence.latestRow(item.identity.id);
    const contract = getViewContract(item.viewSchemaId);
    return row && contract
      ? {
          ...item,
          displayText: genericInspectorRowLabel(contract, row),
          presentation: "observed",
        }
      : item;
  };
  const selected: WorkbookReference[] = (
    props.value
      ? props.field.kind === "direct"
        ? [props.value]
        : props.value.split("\n").filter(Boolean)
      : []
  ).map((id) => {
    const retained = props.retained?.find((item) => item.recordId === id);
    const viewSchemaId =
      retained?.viewSchemaId ||
      (props.field.identityKind === "incident_member"
        ? "incident_members"
        : props.field.targetViewSchemaIds.length === 1
          ? props.field.targetViewSchemaIds[0]
          : "") ||
      "";
    return present({
      identity: { kind: props.field.identityKind, id },
      viewSchemaId,
      displayText: retained?.displayText || id,
      presentation: retained?.displayText ? "retained" : "unresolved",
    });
  });
  const close = () => {
    setOpen(false);
    if (input.current?.isConnected)
      input.current.focus({ preventScroll: true });
    else if (trigger.current?.isConnected)
      trigger.current.focus({ preventScroll: true });
  };
  const attachment = useRef({
    reader: boundary?.reader,
    field: props.field,
    recordId: props.sourceRecordId,
    disabled: props.disabled,
  });
  useLayoutEffect(() => {
    const current = {
      reader: boundary?.reader,
      field: props.field,
      recordId: props.sourceRecordId,
      disabled: props.disabled,
    };
    if (
      attachment.current.reader !== current.reader ||
      attachment.current.field !== current.field ||
      attachment.current.recordId !== current.recordId ||
      attachment.current.disabled !== current.disabled
    ) {
      attachment.current = current;
      setOpen(false);
    }
  }, [boundary?.reader, props.field, props.disabled, props.sourceRecordId]);
  useLayoutEffect(() => {
    const popup = panel.current;
    const anchor = trigger.current;
    if (!open || !popup || !anchor || props.disabled) return;
    popup.showPopover?.();
    const position = () => {
      const viewport = window.visualViewport;
      const left = viewport?.offsetLeft ?? 0,
        top = viewport?.offsetTop ?? 0;
      const width = viewport?.width ?? window.innerWidth,
        height = viewport?.height ?? window.innerHeight;
      const bounds = anchor.getBoundingClientRect();
      // Fixed CSS coordinates must account for inherited CSS zoom. Rectangles
      // and visualViewport are in screen CSS pixels; declared sizes are local.
      const cssWidth = Number.parseFloat(getComputedStyle(popup).width);
      const scale =
        cssWidth > 0 ? popup.getBoundingClientRect().width / cssWidth : 1;
      popup.style.maxWidth = `${width / scale}px`;
      popup.style.maxHeight = `${height / scale}px`;
      const size = popup.getBoundingClientRect();
      popup.style.left = `${Math.max(left, Math.min(bounds.left, left + width - size.width)) / scale}px`;
      popup.style.top = `${Math.max(top, Math.min(bounds.bottom, top + height - size.height)) / scale}px`;
    };
    position();
    const observer = new ResizeObserver(position);
    observer.observe(popup);
    window.addEventListener("resize", position);
    window.addEventListener("scroll", position, true);
    window.visualViewport?.addEventListener("resize", position);
    popup
      .querySelector<HTMLElement>(
        "select:not(:disabled),input:not(:disabled),button:not(:disabled)",
      )
      ?.focus({ preventScroll: true });
    return () => {
      observer.disconnect();
      window.removeEventListener("resize", position);
      window.removeEventListener("scroll", position, true);
      window.visualViewport?.removeEventListener("resize", position);
    };
  }, [open, props.disabled]);
  const inputProps = {
    id: props.id,
    "aria-label": `${props.label} value`,
    "aria-describedby": props.describedBy,
    "aria-invalid": props.invalid,
    "data-testid": props.testId,
    disabled: props.disabled,
    value: props.value,
    ref: (element: HTMLInputElement | HTMLTextAreaElement | null) => {
      input.current = element;
      props.focusTargetRef?.(element);
    },
    onChange: (
      event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
    ) => props.onChange(event.currentTarget.value),
    style: props.compact
      ? {
          ...fieldStyle,
          position: "absolute" as const,
          inset: 0,
          height: "100%",
          padding: "var(--cartulary-grid-cell-padding)",
          paddingInlineEnd: "2.5em",
          border: "none",
          background: "transparent",
        }
      : fieldStyle,
  };
  return (
    <div style={{ minWidth: 0 }}>
      {props.field.kind === "direct" ? (
        <input
          {...inputProps}
          type="text"
          autoComplete="off"
          spellCheck={false}
        />
      ) : (
        <textarea {...inputProps} rows={2} />
      )}
      <Button
        ref={trigger}
        data-grid-editor-interaction="true"
        type="button"
        tone="secondary"
        disabled={props.disabled || !boundary}
        aria-label={`Choose ${props.label.toLowerCase()}`}
        aria-expanded={open}
        aria-controls={open ? panelId : undefined}
        style={
          props.compact
            ? {
                position: "absolute",
                insetInlineEnd: 0,
                top: 0,
                bottom: 0,
                padding: "0 var(--ct-spacing-xs)",
                minWidth: 0,
              }
            : undefined
        }
        onClick={() => setOpen(true)}
      >
        {props.compact ? "…" : `Choose ${props.label.toLowerCase()}`}
      </Button>
      {!props.compact && selected.length ? (
        <span role="status">
          {selected.map((item) => item.displayText).join(", ")}
          {selected.some((item) => item.presentation === "unresolved")
            ? " — selected identifier; label not loaded"
            : ""}
        </span>
      ) : null}
      {open && !props.disabled && boundary ? (
        <div
          ref={panel}
          id={panelId}
          popover="auto"
          role="dialog"
          aria-label={`Choose ${props.label.toLowerCase()}`}
          data-grid-editor-interaction="true"
          style={{
            ...menuStyle,
            position: "fixed",
            inset: "auto",
            margin: 0,
            overflow: "auto",
            width: "min(30rem, 100vw)",
            boxSizing: "border-box",
            color: "var(--ct-colors-ink)",
          }}
          onToggle={(event) => {
            if (event.newState === "closed") setOpen(false);
          }}
          onKeyDown={(event) => {
            if (event.key === "Escape") {
              event.preventDefault();
              event.stopPropagation();
              close();
            }
            if (event.key === "Tab") {
              const controls = Array.from(
                event.currentTarget.querySelectorAll<HTMLElement>(
                  "button:not(:disabled),input:not(:disabled),select:not(:disabled),textarea:not(:disabled)",
                ),
              );
              const index = controls.indexOf(
                document.activeElement as HTMLElement,
              );
              if (
                (event.shiftKey && index === 0) ||
                (!event.shiftKey && index === controls.length - 1)
              ) {
                event.preventDefault();
                controls[event.shiftKey ? controls.length - 1 : 0]?.focus();
              }
            }
            event.stopPropagation();
          }}
        >
          <ReferencePicker
            {...props}
            boundary={boundary}
            selected={selected}
            present={present}
            onCancel={close}
            onAccept={(items) => {
              close();
              props.onAccept(items);
            }}
          />
        </div>
      ) : null}
    </div>
  );
}

type ReadAction = "first" | "previous" | "next" | "retry";
type ReadIntent = {
  readonly action: ReadAction;
  readonly controller: WorkbookReferenceSelection;
  readonly scopeKey: string;
  readonly trigger: HTMLButtonElement;
  owned: boolean;
};

/** Only the focused initiating control is retained; reads stay with the controller. */
function useReferenceReadControls(
  controller: WorkbookReferenceSelection,
  snapshot: ReturnType<WorkbookReferenceSelection["getSnapshot"]>,
  scopeKey: string,
) {
  const intentRef = useRef<ReadIntent | null>(null);
  const admissionRef = useRef<WorkbookReferenceSelection | null>(null);
  const [shownIntent, setShownIntent] = useState<ReadIntent | null>(null);
  const intent =
    shownIntent === intentRef.current &&
    shownIntent?.controller === controller &&
    shownIntent.scopeKey === scopeKey &&
    !snapshot.concealed
      ? shownIntent
      : null;
  const focusedIntent =
    intent?.owned &&
    intent.trigger.isConnected &&
    document.activeElement === intent.trigger;
  const retire = () => {
    intentRef.current = null;
    admissionRef.current = null;
    setShownIntent(null);
  };
  const capable = (action: ReadAction) =>
    action === "first"
      ? snapshot.pageNumber > 1
      : action === "previous"
        ? snapshot.previousCount > 0
        : action === "next"
          ? Boolean(snapshot.page?.paging.hasMore)
          : Boolean(snapshot.failure);

  useEffect(() => {
    if (!shownIntent) return;
    const leave = (event: Event) => {
      const current = intentRef.current;
      if (!current) return;
      if (
        document.activeElement === current.trigger &&
        (event.target === current.trigger ||
          event.type === "wheel" ||
          event.type === "scroll")
      )
        return;
      current.owned = false;
    };
    const events = [
      "focusin",
      "keydown",
      "pointerdown",
      "input",
      "wheel",
      "scroll",
      "touchstart",
    ] as const;
    for (const type of events) document.addEventListener(type, leave, true);
    return () => {
      for (const type of events)
        document.removeEventListener(type, leave, true);
    };
  }, [shownIntent]);
  useEffect(
    () => () => {
      intentRef.current = null;
      admissionRef.current = null;
    },
    [],
  );
  useLayoutEffect(() => {
    const current = intentRef.current;
    if (!current) return;
    const sameScope =
      current.controller === controller &&
      current.scopeKey === scopeKey &&
      !snapshot.concealed;
    const available = capable(current.action);
    if (
      !sameScope ||
      (!snapshot.loading &&
        (available ||
          !current.owned ||
          !current.trigger.isConnected ||
          document.activeElement !== current.trigger))
    ) {
      intentRef.current = null;
      setShownIntent(null);
    }
  });

  const control = (action: ReadAction) => {
    const available = capable(action);
    const initiating = intent?.action === action;
    return {
      disabled:
        !available && !(initiating && (snapshot.loading || focusedIntent)),
      "aria-disabled": snapshot.loading || !available,
      "aria-busy": snapshot.loading && initiating,
      onBlur: (event: React.FocusEvent<HTMLButtonElement>) => {
        const current = intentRef.current;
        if (current?.trigger !== event.currentTarget) return;
        current.owned = false;
        if (!snapshot.loading) {
          intentRef.current = null;
          setShownIntent(null);
        }
      },
    };
  };
  const invoke = (action: ReadAction, trigger: HTMLButtonElement) => {
    if (
      !capable(action) ||
      snapshot.loading ||
      snapshot.concealed ||
      controller.getSnapshot().loading ||
      controller.getSnapshot().concealed ||
      admissionRef.current === controller
    )
      return;
    admissionRef.current = controller;
    const read = controller[action]();
    if (controller.getSnapshot().loading) {
      const next = {
        action,
        controller,
        scopeKey,
        trigger,
        owned: document.activeElement === trigger,
      } satisfies ReadIntent;
      intentRef.current = next;
      setShownIntent(next);
    }
    void read.then(
      () => {
        if (admissionRef.current === controller) admissionRef.current = null;
      },
      () => {
        if (admissionRef.current === controller) admissionRef.current = null;
      },
    );
  };
  return { control, invoke, intent, retire };
}

function ReferencePicker(
  props: Props & {
    boundary: NonNullable<React.ContextType<typeof WorkbookReferenceContext>>;
    selected: readonly WorkbookReference[];
    present: (item: WorkbookReference) => WorkbookReference;
    onCancel: () => void;
  },
) {
  const latest = useRef(props);
  latest.current = props;
  const controller = useMemo(
    () =>
      new WorkbookReferenceSelection({
        field: props.field,
        reader: props.boundary.reader,
        selected: latest.current.selected,
        sourceRecordId: props.sourceRecordId,
        onAuthorityFailure: (failure) =>
          latest.current.boundary.onAuthorityFailure(failure),
      }),
    [props.field, props.boundary.reader, props.sourceRecordId],
  );
  const snapshot = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const note = useContext(NoteCreateContext);
  const noteSnapshot = useSyncExternalStore(
    note?.owner.subscribe ?? noSubscribe,
    note?.owner.getSnapshot ?? noSnapshot,
  );
  const acceptedNotes =
    noteSnapshot?.entries
      .filter((entry) => entry.receipt)
      .map((entry) => entry.attempt.clientTxnId)
      .join(",") ?? "";
  useEffect(() => {
    void controller.first();
    return () => controller.dispose();
  }, [controller]);
  const previousNoteRevision = useRef(acceptedNotes);
  useEffect(() => {
    if (previousNoteRevision.current !== acceptedNotes) {
      previousNoteRevision.current = acceptedNotes;
      if (snapshot.source === "cartulary.view.notes.v1")
        void controller.first();
    }
  }, [controller, acceptedNotes, snapshot.source]);
  const source = getViewContract(snapshot.source);
  const filters =
    source?.fields.filter(
      (field) =>
        source.filterableFieldMap[field.fieldKey] &&
        field.readKind === "text" &&
        field.filterOps.some(
          (op) => op === "prefix" || op === "full_text" || op === "eq",
        ),
    ) ?? [];
  const [filterField, setFilterField] = useState("");
  const [filterValue, setFilterValue] = useState("");
  const filter = filters.find((field) => field.fieldKey === filterField);
  const filterOp = filter?.filterOps.find(
    (op) => op === "prefix" || op === "full_text" || op === "eq",
  );
  const selectedKeys = snapshot.selected.map(workbookReferenceKey);
  const presentedSelected = snapshot.selected.map(props.present);
  const pickerId = useId();
  const selector = useRef<HTMLSelectElement>(null);
  const selectionGroup = useRef<HTMLFieldSetElement>(null);
  const scopeKey = `${pickerId}:${props.field.viewSchemaId}:${props.field.fieldKey}:${props.sourceRecordId ?? ""}:${snapshot.source}:${JSON.stringify(snapshot.queryState)}`;
  const removalFocus = useSelectedReferenceRemovalFocus({
    ids: selectedKeys,
    scopeKey,
    disabled: Boolean(props.disabled || snapshot.concealed),
    fallback: () => selector.current,
    groupRef: selectionGroup,
  });
  const reads = useReferenceReadControls(controller, snapshot, scopeKey);
  if (snapshot.concealed)
    return (
      <div role="status">
        Reference access needs recovery. Your parent draft follows session
        recovery.
        <Button type="button" tone="secondary" onClick={props.onCancel}>
          Close references
        </Button>
      </div>
    );
  return (
    <div style={{ display: "grid", gap: "var(--ct-spacing-sm)", minWidth: 0 }}>
      {controller.sources().length > 1 ? (
        <label>
          Reference surface
          <select
            aria-label="Reference surface"
            style={fieldStyle}
            value={snapshot.source}
            onChange={(event) => {
              reads.retire();
              setFilterField("");
              setFilterValue("");
              void controller.replace(event.currentTarget.value);
            }}
          >
            {controller.sources().map((view) => (
              <option key={view} value={view}>
                {getViewContract(view)?.title ?? "Incident members"}
              </option>
            ))}
          </select>
        </label>
      ) : null}
      {filters.length ? (
        <div>
          <label>
            Filter field
            <select
              style={fieldStyle}
              aria-label="Reference filter field"
              value={filterField}
              onChange={(event) => setFilterField(event.currentTarget.value)}
            >
              <option value="">All candidates</option>
              {filters.map((field) => (
                <option value={field.fieldKey} key={field.fieldKey}>
                  {field.label}
                </option>
              ))}
            </select>
          </label>
          {filter ? (
            <label>
              {filterOp === "prefix"
                ? "Starts with"
                : filterOp === "full_text"
                  ? "Search text"
                  : "Equals"}
              <input
                aria-label="Reference filter value"
                style={fieldStyle}
                value={filterValue}
                onChange={(event) => setFilterValue(event.currentTarget.value)}
              />
            </label>
          ) : null}
          <Button
            type="button"
            tone="secondary"
            disabled={!!filter && !filterValue}
            onClick={() => {
              reads.retire();
              void controller.replace(snapshot.source, {
                ...emptyWorkbookQueryState(),
                filters:
                  filter && filterOp && filterValue
                    ? [
                        {
                          fieldKey: filter.fieldKey,
                          op: filterOp,
                          arg:
                            filterOp === "full_text"
                              ? { query: filterValue }
                              : { value: filterValue },
                        },
                      ]
                    : [],
              });
            }}
          >
            Apply filter
          </Button>
        </div>
      ) : null}
      {snapshot.loading ? <p role="status">Loading references…</p> : null}
      {snapshot.failure ? (
        <p role="alert">
          {snapshot.failure.phase === "continuation"
            ? "Could not load another page. The accepted page is retained. "
            : "This source could not be loaded. "}
          {snapshot.failure.detail.message}
        </p>
      ) : null}
      {snapshot.failure || reads.intent?.action === "retry" ? (
        <Button
          type="button"
          tone="secondary"
          {...reads.control("retry")}
          onClick={(event) => reads.invoke("retry", event.currentTarget)}
        >
          Retry references
        </Button>
      ) : null}
      {snapshot.page ? (
        <>
          <label>
            {props.label}
            <select
              ref={selector}
              aria-label={`${props.label} candidates`}
              style={fieldStyle}
              disabled={!snapshot.page.candidates.length}
              multiple={props.field.kind === "collection"}
              size={Math.min(6, Math.max(2, snapshot.page.candidates.length))}
              value={
                props.field.kind === "collection"
                  ? selectedKeys
                  : snapshot.page.candidates.some(
                        (item) =>
                          workbookReferenceKey(item) === selectedKeys[0],
                      )
                    ? selectedKeys[0]
                    : ""
              }
              onChange={(event) =>
                controller.selectPage(
                  Array.from(event.currentTarget.selectedOptions).map(
                    (option) => option.value,
                  ),
                )
              }
            >
              {props.field.kind === "direct" ? (
                <option value="" disabled>
                  Select a candidate on this page
                </option>
              ) : null}
              {snapshot.page.candidates.map(props.present).map((item) => (
                <option
                  key={workbookReferenceKey(item)}
                  value={workbookReferenceKey(item)}
                >
                  {item.displayText} ({item.identity.id})
                </option>
              ))}
            </select>
          </label>
          <p role="status">
            Page {snapshot.pageNumber}: {snapshot.page.candidates.length}{" "}
            candidates
            {snapshot.page.paging.hasMore
              ? "; more available"
              : "; end of this source"}
            .
          </p>
        </>
      ) : null}
      <div>
        <Button
          type="button"
          tone="secondary"
          {...reads.control("first")}
          onClick={(event) => reads.invoke("first", event.currentTarget)}
        >
          First
        </Button>
        <Button
          type="button"
          tone="secondary"
          {...reads.control("previous")}
          onClick={(event) => reads.invoke("previous", event.currentTarget)}
        >
          Previous
        </Button>
        <Button
          type="button"
          tone="secondary"
          {...reads.control("next")}
          onClick={(event) => reads.invoke("next", event.currentTarget)}
        >
          Next
        </Button>
      </div>
      <fieldset
        ref={selectionGroup}
        aria-label={`${props.label} selected references`}
        tabIndex={-1}
        data-reference-focus-fallback
        style={{ margin: 0, padding: 0, border: 0, minWidth: 0 }}
      >
        <p>
          Selected for this edit: {snapshot.selected.length}. Selection is
          retained across pages and sources.
        </p>
        <ul
          style={{
            margin: 0,
            paddingInlineStart: "var(--ct-spacing-lg)",
            overflowWrap: "anywhere",
          }}
        >
          {presentedSelected.map((item) => (
            <li key={workbookReferenceKey(item)}>
              {item.displayText}
              {item.presentation === "unresolved" ? " (label not loaded)" : ""}
              <Button
                ref={removalFocus.buttonRef(workbookReferenceKey(item))}
                type="button"
                tone="secondary"
                aria-label={`Remove selected ${item.displayText}${presentedSelected.filter((selected) => selected.displayText === item.displayText).length > 1 ? ` (${item.identity.id})` : ""}`}
                onClick={(event) =>
                  removalFocus.remove(
                    workbookReferenceKey(item),
                    event.currentTarget,
                    () => controller.remove(workbookReferenceKey(item)),
                  )
                }
              >
                Remove
              </Button>
            </li>
          ))}
        </ul>
      </fieldset>
      {snapshot.selectionError ? (
        <p role="alert">{snapshot.selectionError}</p>
      ) : null}
      <Button
        type="button"
        tone="secondary"
        disabled={!!snapshot.selectionError || !snapshot.selected.length}
        onClick={() => props.onAccept(snapshot.selected.map(props.present))}
      >
        Use selection
      </Button>
      <Button type="button" tone="secondary" onClick={props.onCancel}>
        Cancel references
      </Button>
    </div>
  );
}
const fieldStyle = {
  width: "100%",
  minWidth: 0,
  boxSizing: "border-box",
  font: "inherit",
  color: "var(--ct-component-text-input-textColor)",
  background: "var(--ct-component-text-input-backgroundColor)",
  border: "var(--ct-component-text-input-border)",
} as const;
