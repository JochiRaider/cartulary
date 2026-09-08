import { useLayoutEffect, useState } from "react";
import { vi } from "vitest";
import { observeAccountOperation } from "../app/accountOperation";
import type { AuthorizationRecoveryResult } from "../shared/authorizationRecovery";
import type { SheetRef } from "../shared/sheetRef";
import { createWorkbookPreferenceAdapter } from "../workbook/adapters/createWorkbookPreferenceAdapter";
import type { WorkbookPreferencePort } from "../workbook/ports/WorkbookPreferencePort";
import { WorkbookPreferenceController } from "../workbook/preferences/WorkbookPreferenceController";
import type {
  DefaultPreference,
  HomePreference,
  PreferenceAuthority,
  PreferenceResult,
} from "../workbook/preferences/workbookPreferenceModel";
export const preferenceIncidentId = "00000000-0000-4000-8000-000000001001";
export const preferenceActorId = "00000000-0000-4000-8000-000000000001";
export const preferenceTimeline = {
  kind: "view_schema",
  id: "cartulary.view.timeline.v2",
} as const;
export const preferenceHosts = {
  kind: "view_schema",
  id: "cartulary.view.hosts.v1",
} as const;
export const preferenceAuthority: PreferenceAuthority = {
  incidentId: preferenceIncidentId,
  actorId: preferenceActorId,
  lifetime: "session-1",
  role: "admin",
};
export const homePreference = (
  home_sheet_ref: SheetRef | null = null,
): HomePreference => ({
  incident_id: preferenceIncidentId,
  user_id: preferenceActorId,
  created_at: "2026-08-01T00:00:00Z",
  updated_at: "2026-08-01T00:00:00Z",
  home_sheet_ref,
});
export const defaultPreference = (
  default_sheet_ref: SheetRef | null = null,
): DefaultPreference => ({
  incident_id: preferenceIncidentId,
  updated_by_user_id: preferenceActorId,
  created_at: "2026-08-01T00:00:00Z",
  updated_at: "2026-08-01T00:00:00Z",
  default_sheet_ref,
});
export const preferenceAccepted = <T>(value: T): PreferenceResult<T> => ({
  kind: "accepted",
  value,
});
export const preferenceFailed = {
  kind: "rejected",
  status: 500,
  problem: { code: "internal_error" },
} as const;
export const preferenceUncertain = {
  kind: "uncertain",
  status: 0,
  problem: { code: "transport" },
} as const;
export function preferenceDeferred<T>() {
  let resolve!: (value: T) => void;
  const promise = new Promise<T>((done) => {
    resolve = done;
  });
  return { promise, resolve };
}
export async function flushPreferences() {
  for (let i = 0; i < 20; i++) await Promise.resolve();
}
export function preferenceFixture() {
  let authority: PreferenceAuthority | null = preferenceAuthority;
  let home = homePreference();
  let defaults = defaultPreference();
  const port = {
    readHome: vi.fn(async () => preferenceAccepted(home)),
    readDefault: vi.fn(async () => preferenceAccepted(defaults)),
    setHomeSheet: vi.fn<WorkbookPreferencePort["setHomeSheet"]>(
      async ({ sheetRef }) => {
        home = homePreference(sheetRef);
        return preferenceAccepted(home);
      },
    ),
    setDefaultSheet: vi.fn<WorkbookPreferencePort["setDefaultSheet"]>(
      async ({ sheetRef }) => {
        defaults = defaultPreference(sheetRef);
        return preferenceAccepted(defaults);
      },
    ),
  };
  const lost = vi.fn();
  const recover = vi.fn(
    async (): Promise<AuthorizationRecoveryResult> =>
      authority
        ? {
            kind: "authorized",
            userId: authority.actorId,
            role: authority.role,
          }
        : { kind: "session_lost" },
  );
  const controller = new WorkbookPreferenceController({
    port: () => port,
    observe: observeAccountOperation,
    recover,
    lost,
    isCurrent: (a) =>
      a.actorId === authority?.actorId &&
      a.incidentId === authority?.incidentId &&
      a.lifetime === authority?.lifetime,
  });
  const setAuthority = (next: PreferenceAuthority | null) => {
    authority = next;
    controller.setAuthority(next);
  };
  setAuthority(authority);
  controller.setSurface({
    sheetRef: preferenceTimeline,
    available: true,
    label: "Timeline",
  });
  return { controller, port, recover, lost, setAuthority };
}

export function usePreferenceTestController(
  incidentId = preferenceIncidentId,
  actorId = preferenceActorId,
) {
  const [controller] = useState(
    () =>
      new WorkbookPreferenceController({
        port: () =>
          createWorkbookPreferenceAdapter({
            apiBase: undefined,
            incidentId,
            actorId,
          }),
        isCurrent: () => true,
        lost: () => {},
        observe: observeAccountOperation,
        recover: async () => ({
          kind: "authorized",
          userId: actorId,
          role: "admin",
        }),
      }),
  );
  useLayoutEffect(() => {
    controller.setAuthority({
      incidentId,
      actorId,
      lifetime: "session",
      role: "admin",
    });
    controller.setSurface({
      sheetRef: preferenceTimeline,
      label: "Timeline",
      available: true,
    });
    controller.setInspectionActive(true);
    return () => controller.dispose();
  }, [controller, incidentId, actorId]);
  return controller;
}
