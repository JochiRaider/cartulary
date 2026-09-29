import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import {
  type CSSProperties,
  type FocusEvent,
  type ReactNode,
  type RefObject,
  useId,
  useLayoutEffect,
  useRef,
} from "react";
import { workbookTypography } from "../components/workbookFormStyles";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { WorkbookInspectorActionButton as Button } from "./presentation/WorkbookInspectorActions";
import { revealWorkbookInspectorField } from "./presentation/workbookInspectorFieldReveal";
import {
  type WorkbookInspectorDisabledReason,
  workbookInspectorDisabledReasonKey,
  workbookInspectorDisabledReasonText,
} from "./presentation/workbookInspectorPresentationModel";
import { WorkbookInspectorSavedDetails } from "./WorkbookInspectorSavedDetails";

type WorkbookInspectorEditorSlots = {
  readonly content: ReactNode;
  readonly actions: ReactNode;
  readonly feedback: ReactNode;
  readonly retainedDraft: ReactNode;
};

export type WorkbookInspectorFieldFocusRequest = {
  readonly revision: number;
  readonly viewSchemaId: string;
  readonly recordId: string;
  readonly fieldKey: string;
  readonly trigger: HTMLElement;
};

/** Explicit edit controls, commands and feedback, independent of any draft owner. */
type WorkbookInspectorEditPresentation = {
  readonly contract: ViewContract;
  readonly row: WorkbookQueryRow;
  readonly editableFields: readonly ViewFieldContract[];
  readonly activeField: string;
  readonly activeAction: string;
  readonly attachmentId: string;
  readonly controlRef: RefObject<HTMLElement | null>;
  readonly focusRequest?: WorkbookInspectorFieldFocusRequest | null | undefined;
  readonly onEdit: (fieldKey: string) => void;
  readonly onDetach: () => void;
  readonly onSubmit: () => void;
  readonly canSubmit: boolean;
  readonly editor: WorkbookInspectorEditorSlots;
  readonly retainedWork: readonly {
    readonly identity: { readonly fieldKey: string; readonly action: string };
    readonly command: {
      readonly kind: "resume" | "review";
      readonly invoke: (trigger?: HTMLElement) => void;
    } | null;
    readonly discard: () => void;
  }[];
  readonly collectionDestinations?:
    | Readonly<Record<string, (() => void) | undefined>>
    | undefined;
  readonly disabledReason?: WorkbookInspectorDisabledReason | null | undefined;
};

/** Saved values always come from the accepted row, never from editor text. */
export function WorkbookInspectorDetails({
  contract,
  row,
  editableFields,
  activeField,
  activeAction,
  attachmentId,
  controlRef,
  focusRequest,
  onEdit,
  onDetach,
  onSubmit,
  canSubmit,
  editor,
  disabledReason,
  retainedWork,
  collectionDestinations,
}: WorkbookInspectorEditPresentation) {
  const reasonId = useId();
  const reasonText = disabledReason
    ? workbookInspectorDisabledReasonText(disabledReason)
    : null;
  const editButtons = useRef(new Map<string, HTMLButtonElement>());
  const fieldElements = useRef(new Map<string, HTMLElement>());
  const attachment = useRef<HTMLFieldSetElement>(null);
  const previousField = useRef("");
  const consumedFocusRequest = useRef<number | null>(null);
  const focusIdentity = JSON.stringify([
    contract.viewSchemaId,
    row.record_id,
    activeField,
    activeAction,
    attachmentId,
  ]);
  const latestFocusIdentity = useRef(focusIdentity);
  latestFocusIdentity.current = focusIdentity;
  const revealOwner = useRef<{
    identity: string;
    body: HTMLElement;
    target: HTMLElement;
    top: number;
    left: number;
  } | null>(null);
  const closingField = useRef<{
    viewSchemaId: string;
    recordId: string;
    fieldKey: string;
    action: string;
    trigger: Element | null;
  } | null>(null);
  const buttonKey = (fieldKey: string, action: string) =>
    JSON.stringify([fieldKey, action]);
  const registerEditButton = (
    fieldKey: string,
    action: string,
    element: HTMLButtonElement | null,
  ) => {
    const key = buttonKey(fieldKey, action);
    if (element) editButtons.current.set(key, element);
    else editButtons.current.delete(key);
  };
  const reveal = (
    field: HTMLElement,
    target: HTMLElement,
    retainOwnership: boolean,
  ) => {
    const body = field.closest<HTMLElement>("[data-inspector-scroll-body]");
    if (!body?.contains(target)) return;
    revealWorkbookInspectorField(body, field, target);
    revealOwner.current = retainOwnership
      ? {
          identity: focusIdentity,
          body,
          target,
          top: body.scrollTop,
          left: body.scrollLeft,
        }
      : null;
  };
  const focusEditor = () => {
    const primary = controlRef.current;
    const target =
      primary?.isConnected && primary.matches(":not(:disabled)")
        ? primary
        : attachment.current?.querySelector<HTMLElement>(
            "[data-inspector-review-control]:not(:disabled), [data-inspector-resume-control]:not(:disabled)",
          );
    if (target && attachment.current?.contains(target))
      target.focus({ preventScroll: true });
  };
  useLayoutEffect(() => {
    const closing = closingField.current;
    closingField.current = null;
    if (
      closing &&
      !activeField &&
      !disabledReason &&
      closing.viewSchemaId === contract.viewSchemaId &&
      closing.recordId === row.record_id &&
      (document.activeElement === document.body ||
        document.activeElement === closing.trigger)
    ) {
      const button =
        editButtons.current.get(buttonKey(closing.fieldKey, closing.action)) ??
        [...editButtons.current.entries()].find(([key]) =>
          key.startsWith(`[${JSON.stringify(closing.fieldKey)},`),
        )?.[1];
      if (button?.isConnected && button.matches(":not(:disabled)"))
        button.focus({ preventScroll: true });
    }
    const fieldChanged = !!activeField && activeField !== previousField.current;
    const requested =
      focusRequest && focusRequest.revision !== consumedFocusRequest.current;
    const validRequest =
      requested &&
      focusRequest.viewSchemaId === contract.viewSchemaId &&
      focusRequest.recordId === row.record_id &&
      focusRequest.fieldKey === activeField &&
      document.activeElement === focusRequest.trigger;
    if (requested) consumedFocusRequest.current = focusRequest.revision;
    if (!disabledReason && (fieldChanged || validRequest)) focusEditor();
    previousField.current = activeField;
  });
  useLayoutEffect(() => {
    if (!activeField || disabledReason) return;
    const field = fieldElements.current.get(activeField);
    const body = field?.closest<HTMLElement>("[data-inspector-scroll-body]");
    if (!field || !body) return;
    let frame: number | null = null;
    const refresh = () => {
      if (frame !== null) return;
      frame = requestAnimationFrame(() => {
        frame = null;
        const owner = revealOwner.current;
        if (
          !owner ||
          owner.identity !== focusIdentity ||
          latestFocusIdentity.current !== focusIdentity ||
          owner.body !== body ||
          document.activeElement !== owner.target ||
          !attachment.current?.contains(owner.target)
        )
          return;
        revealWorkbookInspectorField(body, field, owner.target);
        owner.top = body.scrollTop;
        owner.left = body.scrollLeft;
      });
    };
    const onScroll = () => {
      const owner = revealOwner.current;
      if (
        owner?.identity === focusIdentity &&
        (Math.abs(body.scrollTop - owner.top) > 1 ||
          Math.abs(body.scrollLeft - owner.left) > 1)
      )
        revealOwner.current = null;
    };
    const resize =
      typeof ResizeObserver === "undefined"
        ? null
        : new ResizeObserver(refresh);
    resize?.observe(body);
    resize?.observe(field);
    if (attachment.current) resize?.observe(attachment.current);
    if (controlRef.current) resize?.observe(controlRef.current);
    for (
      let ancestor = body.parentElement;
      ancestor;
      ancestor = ancestor.parentElement
    ) {
      const style = getComputedStyle(ancestor);
      if (
        [style.overflowX, style.overflowY].some((value) =>
          ["auto", "scroll", "hidden", "clip"].includes(value),
        )
      )
        resize?.observe(ancestor);
    }
    const view = body.ownerDocument.defaultView;
    body.addEventListener("scroll", onScroll);
    view?.addEventListener("resize", refresh);
    view?.visualViewport?.addEventListener("resize", refresh);
    body.ownerDocument.fonts?.addEventListener("loadingdone", refresh);
    return () => {
      if (frame !== null) cancelAnimationFrame(frame);
      resize?.disconnect();
      body.removeEventListener("scroll", onScroll);
      view?.removeEventListener("resize", refresh);
      view?.visualViewport?.removeEventListener("resize", refresh);
      body.ownerDocument.fonts?.removeEventListener("loadingdone", refresh);
      if (revealOwner.current?.identity === focusIdentity)
        revealOwner.current = null;
    };
  }, [activeField, disabledReason, focusIdentity, controlRef]);
  const detach = () => {
    closingField.current = {
      viewSchemaId: contract.viewSchemaId,
      recordId: row.record_id,
      fieldKey: activeField,
      action: activeAction,
      trigger: document.activeElement,
    };
    onDetach();
  };
  const fields = new Map(
    contract.fields.map((field) => {
      const cell = row.cells[field.fieldKey];
      const editable = editableFields.some(
        (candidate) => candidate.fieldKey === field.fieldKey,
      );
      const retained = retainedWork.filter(
        (work) => work.identity.fieldKey === field.fieldKey,
      );
      return [
        field.fieldKey,
        {
          controls: (
            <>
              {editable && field.fieldKey !== activeField && retained.length ? (
                retained.map((work) => (
                  <Button
                    key={work.identity.action}
                    aria-label={`${work.command?.kind === "review" ? "Review" : "Resume"} ${retained.length > 1 ? `${work.identity.action} ` : ""}draft for ${field.label}`}
                    tone="quiet"
                    data-inspector-edit-field={field.fieldKey}
                    ref={(element) =>
                      registerEditButton(
                        field.fieldKey,
                        work.identity.action,
                        element,
                      )
                    }
                    aria-describedby={disabledReason ? reasonId : undefined}
                    disabled={!cell || !!disabledReason || !work.command}
                    onClick={(event) =>
                      work.command?.invoke(event.currentTarget)
                    }
                  >
                    {work.command?.kind === "review" ? "Review" : "Resume"}{" "}
                    {retained.length > 1 ? `${work.identity.action} ` : ""}draft
                  </Button>
                ))
              ) : editable && field.fieldKey !== activeField ? (
                <Button
                  tone="quiet"
                  aria-label={`${field.readKind === "collection" ? "Manage" : "Edit"} ${field.label}`}
                  data-inspector-edit-field={field.fieldKey}
                  ref={(element) =>
                    registerEditButton(field.fieldKey, "value", element)
                  }
                  disabled={!cell || !!disabledReason}
                  aria-describedby={disabledReason ? reasonId : undefined}
                  title={reasonText ?? undefined}
                  onClick={() => onEdit(field.fieldKey)}
                >
                  {field.readKind === "collection" ? "Manage" : "Edit"}
                </Button>
              ) : collectionDestinations?.[field.fieldKey] ? (
                <Button
                  tone="quiet"
                  aria-label={`Manage ${field.label}`}
                  onClick={collectionDestinations[field.fieldKey]}
                >
                  Manage
                </Button>
              ) : null}
            </>
          ),
          attachment: (
            <>
              {field.fieldKey !== activeField && retained.length ? (
                <dd style={fullRowStyle}>
                  {retained.map((work) => (
                    <span key={work.identity.action}>
                      <Button
                        tone="quiet"
                        aria-label={`Discard ${retained.length > 1 ? `${work.identity.action} ` : ""}draft for ${field.label}`}
                        onClick={work.discard}
                      >
                        Discard{" "}
                        {retained.length > 1 ? `${work.identity.action} ` : ""}
                        draft
                      </Button>
                    </span>
                  ))}
                </dd>
              ) : null}
              {field.fieldKey === activeField ? (
                <dd style={fullRowStyle}>
                  <fieldset
                    aria-label={`Unsaved change: ${field.label}`}
                    ref={attachment}
                    style={editorStyle}
                    data-inspector-editor-field={field.fieldKey}
                    aria-describedby={disabledReason ? reasonId : undefined}
                    onKeyDown={(event) => {
                      if (
                        event.defaultPrevented ||
                        event.nativeEvent.isComposing
                      )
                        return;
                      if (event.key === "Escape") {
                        event.preventDefault();
                        event.stopPropagation();
                        detach();
                      } else if (
                        event.key === "Enter" &&
                        (event.ctrlKey || event.metaKey)
                      ) {
                        event.preventDefault();
                        event.stopPropagation();
                        if (!event.repeat && canSubmit) onSubmit();
                      }
                    }}
                  >
                    <legend style={labelStyle}>
                      Unsaved change: {field.label}
                    </legend>
                    {editor.content}
                    <div style={actionsStyle}>
                      {editor.actions}
                      <Button tone="quiet" onClick={detach}>
                        Close editor
                      </Button>
                    </div>
                    <p style={retentionStyle}>
                      Update saves this field. Closing keeps unfinished work in
                      this session.
                    </p>
                    {editor.feedback}
                    {editor.retainedDraft}
                  </fieldset>
                </dd>
              ) : null}
            </>
          ),
        },
      ] as const;
    }),
  );
  return (
    <>
      {disabledReason ? (
        <p
          id={reasonId}
          data-inspector-read-only-reason={workbookInspectorDisabledReasonKey(
            disabledReason,
          )}
          style={retentionStyle}
        >
          {reasonText}
        </p>
      ) : null}
      <WorkbookInspectorSavedDetails
        contract={contract}
        row={row}
        fields={fields}
        describedBy={disabledReason ? reasonId : undefined}
        onFieldElement={(fieldKey, element) => {
          if (element) fieldElements.current.set(fieldKey, element);
          else fieldElements.current.delete(fieldKey);
        }}
        onFocusCapture={(event: FocusEvent<HTMLElement>) => {
          const target = event.target;
          if (!(target instanceof HTMLElement)) return;
          const field = [...fieldElements.current.values()].find((element) =>
            element.contains(target),
          );
          if (!field) return;
          const editorFocused =
            !!activeField &&
            field === fieldElements.current.get(activeField) &&
            !!attachment.current?.contains(target) &&
            !disabledReason;
          reveal(field, target, editorFocused);
        }}
      />
    </>
  );
}

const labelStyle = {
  ...workbookTypography("metadata"),
  color: "var(--ct-colors-ink-muted)",
  minInlineSize: 0,
  overflowWrap: "anywhere",
} satisfies CSSProperties;
const fullRowStyle = {
  margin: 0,
  gridColumn: "1 / -1",
  minInlineSize: 0,
} satisfies CSSProperties;
const editorStyle = {
  display: "grid",
  gap: "var(--ct-spacing-xs)",
  margin: 0,
  padding: "var(--ct-spacing-sm)",
  border: "var(--ct-border-hairline)",
  borderRadius: "var(--ct-component-text-input-rounded)",
  minInlineSize: 0,
} satisfies CSSProperties;
const actionsStyle = {
  display: "flex",
  flexWrap: "wrap",
  gap: "var(--ct-spacing-xs)",
  alignItems: "center",
} satisfies CSSProperties;
const retentionStyle = {
  ...workbookTypography("metadata"),
  color: "var(--ct-colors-ink-muted)",
  margin: 0,
} satisfies CSSProperties;
