import {
  type WorkbookShellSlot,
  workbookShellReadyTestId,
  workbookShellSlotLabel,
  workbookShellSlotTestId,
} from "@cartulary/ui-contracts";
import type { CSSProperties, ReactNode } from "react";

export const workbookShellId = workbookShellReadyTestId();

export function WorkbookShellSlotRegion({
  children,
  inert,
  slot,
  style,
  viewSchemaId,
}: {
  readonly children: ReactNode;
  readonly inert?: boolean | undefined;
  readonly slot: WorkbookShellSlot;
  readonly style?: CSSProperties | undefined;
  readonly viewSchemaId?: string | undefined;
}) {
  return (
    <section
      aria-label={workbookShellSlotLabel(slot)}
      data-testid={workbookShellSlotTestId(slot)}
      data-view-schema-id={viewSchemaId}
      data-workbook-shell-id={workbookShellId}
      inert={inert || undefined}
      style={style}
    >
      {children}
    </section>
  );
}
