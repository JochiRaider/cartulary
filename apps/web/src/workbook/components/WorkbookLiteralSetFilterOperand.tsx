import {
  type AriaAttributes,
  type RefCallback,
  useLayoutEffect,
  useRef,
} from "react";
import {
  appendFilterDraftMember,
  type FilterDraftControl,
  type FilterDraftMember,
} from "../models/workbookQuery";
import {
  inputStyle,
  secondaryButtonStyle,
  stackedLabelStyle,
} from "./workbookGridControlStyles";

/** Only presentation and pending focus live here; raw members belong to FilterDraft. */
export function WorkbookLiteralSetFilterOperand({
  members,
  onChange,
  valueLabel = "Value",
  addLabel = "Add value",
  removeLabel,
  placeholder,
  valueTestId,
  feedbackFor,
  registerItem,
}: {
  readonly members: readonly FilterDraftMember[];
  readonly onChange: (members: readonly FilterDraftMember[]) => void;
  readonly valueLabel?: string;
  readonly addLabel?: string;
  readonly removeLabel?: string;
  readonly placeholder?: string | undefined;
  readonly valueTestId?: string | undefined;
  readonly feedbackFor: (
    control: FilterDraftControl,
  ) => Pick<AriaAttributes, "aria-invalid" | "aria-describedby">;
  readonly registerItem?:
    | ((key: string) => RefCallback<HTMLElement>)
    | undefined;
}) {
  const inputs = useRef(new Map<number, HTMLTextAreaElement>());
  const add = useRef<HTMLButtonElement>(null);
  const pendingFocus = useRef<number | "add" | null>(null);
  useLayoutEffect(() => {
    const target = pendingFocus.current;
    if (target === null) return;
    pendingFocus.current = null;
    (target === "add" ? add.current : inputs.current.get(target))?.focus();
  });
  return (
    <div
      style={{ display: "grid", gap: "var(--ct-spacing-xs)", minInlineSize: 0 }}
    >
      <span style={stackedLabelStyle}>
        Each value is one member. Commas remain part of a value.
      </span>
      {members.map((member, index) => (
        <div
          key={member.id}
          style={{
            display: "flex",
            alignItems: "end",
            gap: "var(--ct-spacing-xs)",
            minInlineSize: 0,
          }}
        >
          <label style={{ ...stackedLabelStyle, flex: 1, minInlineSize: 0 }}>
            Value {index + 1}
            <textarea
              rows={1}
              ref={(element) => {
                if (element) inputs.current.set(member.id, element);
                else inputs.current.delete(member.id);
                registerItem?.(`member:${member.id}`)(element);
              }}
              aria-label={`${valueLabel} ${index + 1}`}
              {...feedbackFor(`member:${member.id}`)}
              data-testid={index === 0 ? valueTestId : undefined}
              placeholder={placeholder}
              style={{
                ...inputStyle,
                minInlineSize: 0,
                inlineSize: "100%",
                resize: "vertical",
              }}
              value={member.value}
              onChange={(event) => {
                const value = event.currentTarget.value;
                onChange(
                  members.map((item) =>
                    item.id === member.id ? { ...item, value } : item,
                  ),
                );
              }}
            />
          </label>
          <button
            type="button"
            ref={registerItem?.(`member_remove:${member.id}`)}
            style={secondaryButtonStyle}
            aria-label={`${removeLabel ?? `Remove ${valueLabel.toLowerCase()}`} ${index + 1}`}
            onClick={() => {
              pendingFocus.current =
                members[index + 1]?.id ?? members[index - 1]?.id ?? "add";
              onChange(members.filter((item) => item.id !== member.id));
            }}
          >
            Remove
          </button>
        </div>
      ))}
      <button
        type="button"
        ref={(element) => {
          add.current = element;
          registerItem?.("member_add")(element);
        }}
        {...(members.length === 0 ? feedbackFor("value") : {})}
        style={secondaryButtonStyle}
        onClick={() => {
          const next = appendFilterDraftMember(members);
          pendingFocus.current = next.at(-1)?.id ?? "add";
          onChange(next);
        }}
      >
        {addLabel}
      </button>
    </div>
  );
}
