import {
  gridFilterFieldTestId,
  workbookFilterPopoverTriggerTestId,
} from "@cartulary/ui-contracts";
import type { Page } from "@playwright/test";
import { expect } from "../../fixtures";

export async function openFilterPopover(page: Page, viewSchemaId: string) {
  const trigger = page.getByTestId(
    workbookFilterPopoverTriggerTestId(viewSchemaId),
  );
  await expect(trigger).toBeVisible();
  const field = page.getByTestId(gridFilterFieldTestId(viewSchemaId));
  if (!(await field.isVisible().catch(() => false))) {
    await trigger.click();
  }
  await expect(field).toBeVisible();
}
