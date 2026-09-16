import {
  genericCreateFieldTestId,
  genericCreateSubmitTestId,
  gridShellTestId,
  workbookInspectorCloseButtonTestId,
  workbookInspectorFeatureActionTestId,
} from "@cartulary/ui-contracts";
import {
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import { expect, type Page } from "@playwright/test";
import { createIncident } from "../incidents/fixtures";
import { createIncidentMemberUser } from "../incidents/memberships";
import {
  uniqueEmail,
  uniqueIncidentKey,
  uniqueTxn,
} from "../runtime/fixtureIdentity";
import { createViewRow } from "./query";
import { openRecoveryItem, recoveryEntry } from "./recovery";
import {
  openGenericInspectorForRecord,
  openTimelineInspector,
} from "./rowMutations";

export const coordinationMatrix = {
  [timelineViewSchemaId]: ["comm_log", "handoff", "status_review", "lesson"],
  "cartulary.view.task_requests.v1": ["comm_log", "status_review", "lesson"],
  "cartulary.view.decisions.v1": ["comm_log", "status_review"],
  "cartulary.view.comm_log.v1": ["status_review"],
  "cartulary.view.handoff.v1": ["status_review"],
  "cartulary.view.status_review.v1": ["comm_log"],
} as const;
export async function openCoordinationFixture(
  page: Page,
  variant = "lesson",
  view: string = timelineViewSchemaId,
  navigate: (url: string) => Promise<unknown> = (url) => page.goto(url),
) {
  const incident = await createIncident(
    page,
    uniqueIncidentKey("CCA"),
    "Contextual coordination authoring",
  );
  const member = await createIncidentMemberUser(page, incident, {
    display_name: "Coordination owner",
    email: uniqueEmail("coordination-owner"),
    initial_password: "CoordinationOwner1!",
    role: "editor",
    is_deployment_admin: false,
    mfa_required: false,
  });
  const values: Record<string, string> = {};
  const sourceContract = requireViewContract(view);
  for (const key of sourceContract.minimumCreateFieldSets[0] ?? []) {
    const field = sourceContract.fieldMap[key];
    values[key] = key.endsWith("_user_id")
      ? member.user_id
      : (field?.enumValues?.[0] ?? "Reviewed coordination source");
  }
  if (view === timelineViewSchemaId)
    values["timeline.activity_synopsis_text"] = "Reviewed coordination source";
  const source = await createViewRow(page, incident, view, {
    client_txn_id: uniqueTxn("coordination-source"),
    ...values,
  });
  await navigate(
    `/?incident_id=${incident}&view_schema_id=${encodeURIComponent(view)}`,
  );
  await expect(page.getByTestId(gridShellTestId(view))).toBeVisible();
  if (view === timelineViewSchemaId)
    await openTimelineInspector(page, source.record_id);
  else await openGenericInspectorForRecord(page, view, source.record_id);
  const target = requireViewContract(`cartulary.view.${variant}.v1`);
  const action = page.getByTestId(
    workbookInspectorFeatureActionTestId(view, `create_related.${variant}`),
  );
  await action.focus();
  await action.press("Enter");
  const form = page.getByRole("region", {
    name: `Create ${target.title}`,
    exact: true,
  });
  await expect(form).toBeVisible();
  return { incident, source, view, target, variant, action, form, member };
}
export async function fillCoordinationMinimum(
  f: Awaited<ReturnType<typeof openCoordinationFixture>>,
) {
  for (const key of f.target.minimumCreateFieldSets[0] ?? []) {
    const field = f.target.fieldMap[key];
    if (!field) throw new Error("Missing field");
    if (field.directReferenceContractId) {
      await f.form
        .getByRole("button", {
          name: `Choose ${field.label.toLowerCase()}`,
          exact: true,
        })
        .click();
      await expect(
        f.form.getByRole("option", { name: "Coordination owner", exact: true }),
      ).toBeAttached();
      await f.form
        .getByTestId(genericCreateFieldTestId(key))
        .selectOption(f.member.user_id);
      await f.form
        .getByRole("button", { name: "Apply references", exact: true })
        .click();
    } else if (field.enumValues)
      await f.form
        .getByTestId(genericCreateFieldTestId(key))
        .selectOption(field.enumValues[0] as string);
    else
      await f.form
        .getByTestId(genericCreateFieldTestId(key))
        .fill("  Authored coordination summary  ");
  }
}
export async function retainCoordinationUncertainResult(
  page: Page,
  f: Awaited<ReturnType<typeof openCoordinationFixture>>,
) {
  const path = `**/incidents/${f.incident}/views/${f.target.viewSchemaId}/rows`;
  let reachedTransport: () => void = () => {};
  const captured = new Promise<void>((resolve) => {
    reachedTransport = resolve;
  });
  await page.route(path, async (route) => {
    const response = await route.fetch();
    expect(response.ok()).toBe(true);
    await route.abort("failed");
    reachedTransport();
  });
  await f.form
    .getByTestId(genericCreateSubmitTestId(f.target.viewSchemaId))
    .click();
  await captured;
  const summary = recoveryEntry(page);
  await expect(summary).toBeVisible();
  await page.getByTestId(workbookInspectorCloseButtonTestId(f.view)).click();
  await openRecoveryItem(page, /^Coordination creation ·/);
  const recovery = page.getByRole("region", {
    name: "Retained Coordination authoring",
    exact: true,
  });
  await expect(
    recovery.getByRole("button", { name: "Recover submission", exact: true }),
  ).toBeEnabled();
  return { path, summary, recovery };
}
