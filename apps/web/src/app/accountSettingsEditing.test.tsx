import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import {
  installLandingShellFetch,
  sessionResource,
} from "../testing/appShellTestSupport";
import {
  deferred,
  findFetchCallsByPath,
  jsonResponse,
} from "../testing/fetchMockTestSupport";
import { AppRoot } from "./AppRoot";

const userId = "00000000-0000-4000-8000-000000000001";
const timestamp = "2026-09-05T12:00:00Z";
const profile = (display_name = "Operator", user_version = 1) => ({
  user_id: userId,
  email: "operator@example.test",
  display_name,
  user_version,
  created_at: timestamp,
  updated_at: timestamp,
});
const preferences = (
  density_mode: "compact" | "default" | "comfortable" | null = null,
  preferences_version = 1,
) => ({
  user_id: userId,
  density_mode,
  preferences_version,
  created_at: timestamp,
  updated_at: timestamp,
});
const response = (data: object) =>
  jsonResponse({ data, meta: { request_id: "settings-test" } });
const nameInput = () => screen.getByRole("textbox", { name: "Display name" });
const save = () => screen.getByRole("button", { name: "Save profile" });
let fetchMock: ReturnType<typeof vi.fn>;
let profileWrite: ReturnType<typeof deferred<Response>>;
let preferenceWrite: ReturnType<typeof deferred<Response>>;
let preferenceRead: ReturnType<
  typeof vi.fn<() => Response | Promise<Response>>
>;
let profileRead: ReturnType<typeof vi.fn<() => Response | Promise<Response>>>;

beforeEach(() => {
  window.history.replaceState({}, "", "/");
  fetchMock = vi.fn();
  profileWrite = deferred<Response>();
  preferenceWrite = deferred<Response>();
  preferenceRead = vi.fn(() => response(preferences()));
  profileRead = vi.fn(() => response(profile()));
  vi.stubGlobal("fetch", fetchMock);
  installLandingShellFetch(fetchMock, {
    session: sessionResource({ user_id: userId }),
    accountPreferences: () => preferenceRead(),
    extraRoutes: [
      {
        method: "GET",
        url: "/api/v1/account/profile",
        handler: () => profileRead(),
      },
      {
        method: "PATCH",
        url: "/api/v1/account/profile",
        handler: () => profileWrite.promise,
      },
      {
        method: "PUT",
        url: "/api/v1/account/preferences",
        handler: () => preferenceWrite.promise,
      },
    ],
  });
});
afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});

async function openSettings() {
  render(<AppRoot />);
  fireEvent.click(
    await screen.findByRole("button", {
      name: "Account and application navigation",
    }),
  );
  fireEvent.click(screen.getByRole("menuitem", { name: "Account settings" }));
}
async function openProfile() {
  await openSettings();
  await waitFor(() =>
    expect((nameInput() as HTMLInputElement).value).toBe("Operator"),
  );
}
function edit(value: string) {
  fireEvent.change(nameInput(), { target: { value } });
}
function writes() {
  return findFetchCallsByPath(fetchMock, "/api/v1/account/profile", "PATCH");
}

describe("account settings editing", () => {
  it("retries an unavailable profile without changing opening focus", async () => {
    profileRead.mockResolvedValueOnce(
      jsonResponse({ error: { code: "unavailable" } }, 503),
    );
    await openSettings();
    await screen.findByRole("button", { name: "Retry load" });
    expect(document.activeElement).toBe(
      screen.getByRole("button", { name: "Close" }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Retry load" }));
    await waitFor(() =>
      expect((nameInput() as HTMLInputElement).value).toBe("Operator"),
    );
  });
  it("associates native-submit validation with the field and preserves input focus", async () => {
    const user = userEvent.setup();
    await openProfile();
    await user.click(nameInput());
    await user.clear(nameInput());
    await user.type(nameInput(), "   {Enter}");
    expect(writes()).toHaveLength(0);
    expect(nameInput().getAttribute("aria-invalid")).toBe("true");
    const feedback = document.getElementById(
      nameInput().getAttribute("aria-describedby") ?? "",
    );
    expect(feedback?.textContent).toBe("Enter a display name.");
    expect(document.activeElement).toBe(nameInput());
    await user.clear(nameInput());
    await user.type(nameInput(), "Keyboard name{Enter}");
    expect(writes()).toHaveLength(1);
    expect(nameInput().getAttribute("aria-invalid")).toBe("false");
  });
  it("keeps a dirty draft through a background refresh and exposes explicit review", async () => {
    await openProfile();
    edit("Local draft");
    const refreshed = deferred<Response>();
    profileRead.mockReturnValueOnce(refreshed.promise);
    nameInput().focus();
    fireEvent.click(screen.getByRole("button", { name: "Refresh profile" }));
    await act(async () =>
      refreshed.resolve(response(profile("External name", 2))),
    );
    expect((nameInput() as HTMLInputElement).value).toBe("Local draft");
    expect(document.activeElement).toBe(nameInput());
    expect(screen.getByText("External name")).toBeTruthy();
    expect((save() as HTMLButtonElement).disabled).toBe(true);
    fireEvent.click(screen.getByRole("button", { name: "Review my edit" }));
    fireEvent.click(save());
    expect(JSON.parse(String(writes()[0]?.[1]?.body))).toMatchObject({
      base_user_version: 2,
      display_name: "Local draft",
    });
  });
  it("offers exact recovery for a malformed success without announcing rejection", async () => {
    await openProfile();
    edit("Retained name");
    fireEvent.click(save());
    await act(async () =>
      profileWrite.resolve(jsonResponse({ data: { display_name: "invalid" } })),
    );
    expect(screen.getByRole("button", { name: "Retry save" })).toBeTruthy();
    expect((nameInput() as HTMLInputElement).value).toBe("Retained name");
    expect(
      screen.getByText(/could not confirm whether your profile was saved/),
    ).toBeTruthy();
  });
  it("suppresses duplicate activation at the submission boundary", async () => {
    await openProfile();
    edit("First edit");
    act(() => {
      save().click();
      save().click();
    });
    expect(writes()).toHaveLength(1);
  });
  it("retains a newer draft when a captured save completes", async () => {
    await openProfile();
    edit("First edit");
    fireEvent.click(save());
    edit("Newer edit");
    await act(async () => {
      profileWrite.resolve(response(profile("First edit", 2)));
    });
    expect((nameInput() as HTMLInputElement).value).toBe("Newer edit");
  });
  it("retains drafts across section switches and dismissal", async () => {
    await openProfile();
    edit("Retained edit");
    fireEvent.click(screen.getByRole("tab", { name: "Appearance" }));
    fireEvent.click(screen.getByRole("tab", { name: "Profile" }));
    expect((nameInput() as HTMLInputElement).value).toBe("Retained edit");
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    fireEvent.click(
      screen.getByRole("button", {
        name: "Account and application navigation",
      }),
    );
    fireEvent.click(screen.getByRole("menuitem", { name: "Account settings" }));
    expect((nameInput() as HTMLInputElement).value).toBe("Retained edit");
  });
  it("does not initiate session publication after application disposal", async () => {
    await openProfile();
    edit("Retired edit");
    fireEvent.click(save());
    cleanup();
    const before = findFetchCallsByPath(
      fetchMock,
      "/api/v1/auth/session",
      "GET",
    ).length;
    await act(async () => {
      profileWrite.resolve(response(profile("Retired edit", 2)));
    });
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/auth/session", "GET"),
    ).toHaveLength(before);
  });
  it("retains the appearance draft after a definitive version conflict", async () => {
    await openProfile();
    fireEvent.click(screen.getByRole("tab", { name: "Appearance" }));
    await waitFor(() =>
      expect(
        (
          screen.getByRole("button", {
            name: "Refresh appearance",
          }) as HTMLButtonElement
        ).disabled,
      ).toBe(false),
    );
    fireEvent.click(screen.getByRole("radio", { name: "Comfortable" }));
    const refreshed = deferred<Response>();
    const reads = preferenceRead.mock.calls.length;
    preferenceRead.mockImplementation(() => refreshed.promise);
    fireEvent.click(screen.getByRole("button", { name: "Save appearance" }));
    await act(async () => {
      preferenceWrite.resolve(
        jsonResponse(
          { error: { code: "preferences_version_conflict", retryable: false } },
          409,
        ),
      );
    });
    await waitFor(() =>
      expect(preferenceRead.mock.calls.length).toBeGreaterThan(reads),
    );
    await act(async () => {
      refreshed.resolve(response(preferences("compact", 2)));
    });
    expect(
      (screen.getByRole("radio", { name: "Comfortable" }) as HTMLInputElement)
        .checked,
    ).toBe(true);
    expect(
      findFetchCallsByPath(fetchMock, "/api/v1/account/preferences", "PUT"),
    ).toHaveLength(1);
  });
});
