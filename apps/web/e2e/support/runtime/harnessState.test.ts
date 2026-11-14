// @vitest-environment node

import {
  chmodSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  statSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

const admission = vi.hoisted(() => ({ borrow: vi.fn(), acquisition: vi.fn() }));
vi.mock("../../../../../tools/harness/runtime/session-client.mjs", () => ({
  borrowRuntime: admission.borrow,
}));
vi.mock("../../../../../tools/harness/browser/browser-acquisition.mjs", () => ({
  readBrowserAcquisition: admission.acquisition,
}));

import {
  readBrowserState,
  withBrowserSetupLock,
  writeBrowserState,
} from "../../../../../tools/harness/browser/private-state.mjs";

let root: string;
let session: string;
let state: string;
let runtime: { root: string; close: ReturnType<typeof vi.fn> };

beforeEach(() => {
  root = mkdtempSync(join(tmpdir(), "cartulary-browser-state-"));
  session = join(root, "session");
  state = join(session, "runtime-root", "playwright-state");
  mkdirSync(state, { recursive: true, mode: 0o700 });
  runtime = { root, close: vi.fn() };
  admission.borrow.mockReset().mockReturnValue(runtime);
  admission.acquisition
    .mockReset()
    .mockReturnValue({ session_root: session, suite_id: "suite" });
  for (const [name, value] of Object.entries({
    CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED: "1",
    CARTULARY_TEST_RESULTS_DIR: join(root, "results"),
    CARTULARY_TEST_RUN_ID: "state-unit",
    CARTULARY_BROWSER_ACQUISITION_FILE: join(
      session,
      "acquisition",
      "owner.json",
    ),
    CARTULARY_WEB_E2E_RUNTIME_ROOT: join(session, "runtime-root"),
    CARTULARY_PLAYWRIGHT_STATE_DIR: state,
    CARTULARY_TEST_SUITE_ID: "suite",
  }))
    vi.stubEnv(name, value);
});

afterEach(() => {
  vi.unstubAllEnvs();
  rmSync(root, { recursive: true, force: true });
});

describe("harnessState", () => {
  it("rejects missing or unsafe owned-state bindings without creating or repairing paths", async () => {
    // Importing pure parsers/authentication helpers must not acquire state.
    await import("../auth/workerAdmin");
    await import("../auth/suiteAdmin");
    expect(admission.borrow).not.toHaveBeenCalled();

    for (const variable of [
      "CARTULARY_WEB_E2E_ATTACHMENT_VALIDATED",
      "CARTULARY_TEST_RESULTS_DIR",
      "CARTULARY_TEST_RUN_ID",
      "CARTULARY_BROWSER_ACQUISITION_FILE",
    ]) {
      const value = process.env[variable];
      vi.stubEnv(variable, "");
      expect(() => writeBrowserState("secret", "PrivateValue!")).toThrow(
        "owned runtime binding",
      );
      expect(existsSync(join(state, "secret"))).toBe(false);
      vi.stubEnv(variable, value);
    }
    for (const name of ["../escape", "/absolute", ".", "..", "nested/file"]) {
      expect(() => readBrowserState(name)).toThrow("owned runtime binding");
    }
    admission.borrow.mockImplementationOnce(() => {
      throw new Error("private owner marker mismatch: secret-value");
    });
    expect(() => readBrowserState("secret")).toThrow(
      /^browser private state requires its owned runtime binding \(runtime\)$/,
    );
    admission.acquisition.mockImplementationOnce(() => {
      throw new Error("foreign acquisition: secret-value");
    });
    expect(() => readBrowserState("secret")).toThrow(
      /^browser private state requires its owned runtime binding \(acquisition\)$/,
    );

    for (const variable of [
      "CARTULARY_PLAYWRIGHT_STATE_DIR",
      "CARTULARY_WEB_E2E_RUNTIME_ROOT",
      "CARTULARY_TEST_SUITE_ID",
    ]) {
      const value = process.env[variable];
      vi.stubEnv(variable, join(root, "foreign"));
      expect(() => readBrowserState("secret")).toThrow("owned runtime binding");
      expect(existsSync(join(root, "foreign"))).toBe(false);
      vi.stubEnv(variable, value);
    }
    for (const directory of [
      root,
      session,
      join(session, "runtime-root"),
      state,
    ]) {
      chmodSync(directory, 0o755);
      expect(() => readBrowserState("secret")).toThrow("owned runtime binding");
      expect(statSync(directory).mode & 0o777).toBe(0o755);
      chmodSync(directory, 0o700);
    }
    rmSync(state, { recursive: true });
    const foreign = join(root, "foreign");
    mkdirSync(foreign, { mode: 0o700 });
    symlinkSync(foreign, state);
    expect(() => writeBrowserState("secret", "PrivateValue!")).toThrow(
      "owned runtime binding",
    );
    expect(existsSync(join(foreign, "secret"))).toBe(false);
    rmSync(state);
    expect(() => readBrowserState("secret")).toThrow("owned runtime binding");
    expect(existsSync(state)).toBe(false);
  });

  it("preserves private publication isolation and borrowed runtime lifetime", async () => {
    expect(readBrowserState("secret")).toBeNull();
    writeBrowserState("secret", "old-private-value");
    expect(readBrowserState("secret")).toBe("old-private-value");
    writeBrowserState("secret", "new-private-value");
    expect(readBrowserState("secret")).toBe("new-private-value");
    expect(statSync(join(state, "secret")).mode & 0o777).toBe(0o600);
    expect(admission.acquisition).toHaveBeenCalledWith(
      join(session, "acquisition", "owner.json"),
      runtime,
    );
    expect(runtime.close).toHaveBeenCalledTimes(
      admission.borrow.mock.calls.length,
    );

    const unsafe = join(state, "unsafe");
    writeFileSync(unsafe, "PrivateValue!", { mode: 0o644 });
    expect(() => readBrowserState("unsafe")).toThrow(
      /^browser private state is unreadable$/,
    );
    expect(() => writeBrowserState("unsafe", "replacement")).toThrow(
      /^browser private state publication failed$/,
    );
    expect(readFileSync(unsafe, "utf8")).toBe("PrivateValue!");
    symlinkSync(join(state, "secret"), join(state, "alias"));
    expect(() => readBrowserState("alias")).toThrow("unreadable");
    expect(() => writeBrowserState("alias", "overwrite")).toThrow(
      "publication failed",
    );
    expect(readBrowserState("secret")).toBe("new-private-value");

    const lock = join(state, "global-setup.lock");
    const primary = new Error("setup failed");
    await expect(
      withBrowserSetupLock(async () => {
        expect(statSync(lock).mode & 0o777).toBe(0o700);
        throw primary;
      }),
    ).rejects.toBe(primary);
    expect(existsSync(lock)).toBe(false);
    expect(await withBrowserSetupLock(async () => "complete")).toBe("complete");
    expect(existsSync(lock)).toBe(false);
    let release!: () => void;
    const held = new Promise<void>((resolve) => {
      release = resolve;
    });
    const first = withBrowserSetupLock(async () => {
      await held;
    });
    let entered = false;
    const second = withBrowserSetupLock(async () => {
      entered = true;
    });
    expect(entered).toBe(false);
    release();
    await Promise.all([first, second]);
    expect(entered).toBe(true);
    expect(existsSync(lock)).toBe(false);
    expect(readBrowserState("secret")).toBe("new-private-value");
    expect(existsSync(session)).toBe(true);
    symlinkSync(session, lock);
    await expect(
      withBrowserSetupLock(async () => {
        throw new Error("must not enter");
      }),
    ).rejects.toThrow("lock is unsafe");
    expect(existsSync(session)).toBe(true);
  });
});
