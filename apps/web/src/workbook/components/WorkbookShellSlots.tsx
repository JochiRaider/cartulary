import {
  type WorkbookShellSlot,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties, KeyboardEventHandler, ReactNode } from "react";

export const workbookShellId = workbookShellReadyTestId();

export function WorkbookShellSlotRegion({
  children,
  inert,
  onKeyDown,
  slot,
  style,
  viewSchemaId,
}: {
  readonly children: ReactNode;
  readonly inert?: boolean | undefined;
  readonly onKeyDown?: KeyboardEventHandler<HTMLElement> | undefined;
  readonly slot: WorkbookShellSlot;
  readonly style?: CSSProperties | undefined;
  readonly viewSchemaId?: string | undefined;
}) {
  return (
    <section
      aria-label={workbookShellSlotLabel(slot)}
      data-testid={workbookShellSlotTestId(slot)}
      data-workbook-slot={slot}
      data-view-schema-id={viewSchemaId}
      data-workbook-shell-id={workbookShellId}
      onKeyDown={onKeyDown}
      inert={inert || undefined}
      style={style}
    >
      {children}
    </section>
  );
}
