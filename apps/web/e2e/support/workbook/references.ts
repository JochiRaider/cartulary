import { genericCreateFieldTestId } from "@cartulary/ui-contracts";
import { listViewContracts } from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";

function createFieldLabel(testId: string) {
  return listViewContracts()
    .flatMap((contract) => contract.fields)
    .find((field) => genericCreateFieldTestId(field.fieldKey) === testId)
    ?.label;
}

export async function openReferenceCandidates(
  page: Page,
  testId: string,
  referenceView?: string,
) {
  const label =
    createFieldLabel(testId) ??
    (await page.getByTestId(testId).getAttribute("aria-label"))?.replace(
      / value$/u,
      "",
    );
  if (!label) throw new Error(`Missing reference label for ${testId}`);
  const name = `Choose ${label.toLowerCase()}`;
  const picker = page
    .getByRole("region", { name, exact: true })
    .or(page.getByRole("dialog", { name, exact: true }));
  if (!(await picker.isVisible())) {
    // Details and create-related workflows can expose the same field label.
    // Find the named discovery action belonging to this exact raw control.
    let owner = page.getByTestId(testId).locator("..");
    for (
      let depth = 0;
      depth < 5 &&
      !(await owner.getByRole("button", { name, exact: true }).count());
      depth++
    )
      owner = owner.locator("..");
    await owner.getByRole("button", { name, exact: true }).click();
  }
  await expect(picker).toBeVisible();
  const surface = picker.getByRole("combobox", {
    name: "Reference surface",
    exact: true,
  });
  if (referenceView && (await surface.count()))
    await surface.selectOption(referenceView);
  const names = [label, `${label} candidates`];
  const candidates = names
    .map((name) =>
      picker
        .getByRole("combobox", { name, exact: true })
        .or(picker.getByRole("listbox", { name, exact: true })),
    )
    .reduce((left, right) => left.or(right));
  return { picker, candidates };
}

export async function selectReferenceCandidates(
  page: Page,
  testId: string,
  recordIds: string | readonly string[],
  referenceView?: string,
) {
  const { picker, candidates } = await openReferenceCandidates(
    page,
    testId,
    referenceView,
  );
  const values: string[] = [];
  for (const recordId of typeof recordIds === "string"
    ? [recordIds]
    : recordIds) {
    const option = candidates.locator(
      `option[value="${recordId}"], option[value$=":${recordId}"]`,
    );
    await expect(option).toHaveCount(1);
    const value = await option.getAttribute("value");
    if (value === null)
      throw new Error(`Missing candidate identity for ${recordId}`);
    values.push(value);
  }
  await candidates.selectOption(values);
  await picker
    .getByRole("button", { name: /^(Apply references|Use selection)$/u })
    .click();
  await expect(picker).toBeHidden();
}
