// @vitest-environment node

import {
  mkdtempSync,
  readdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { afterEach, describe, expect, it } from "vitest";

import {
  parseWorkerAdminManifest,
  workerAdminManifestSchemaID,
} from "../auth/workerAdmin";
import { atomicWritePrivateFile } from "./privateState";

const roots: string[] = [];

afterEach(() => {
  for (const root of roots.splice(0)) {
    rmSync(root, { force: true, recursive: true });
  }
});

function scratchFile() {
  const root = mkdtempSync(join(tmpdir(), "cartulary-private-state-"));
  roots.push(root);
  return { file: join(root, "state.json"), root };
}

function manifest(overrides: Record<string, unknown> = {}) {
  return {
    schema_id: workerAdminManifestSchemaID,
    worker_admins: [
      {
        parallel_index: 0,
        user_id: "user-0",
        email: "worker-0@example.test",
        password: "PrivateWorker0Pass!",
      },
    ],
    ...overrides,
  };
}

describe("atomic private harness state", () => {
  it("publishes complete old-or-new contents with owner-only permissions", () => {
    const { file } = scratchFile();
    writeFileSync(file, "old\n", { encoding: "utf8", mode: 0o600 });

    atomicWritePrivateFile(file, "new\n", {
      beforeRename: (temporaryPath) => {
        expect(readFileSync(file, "utf8")).toBe("old\n");
        expect(readFileSync(temporaryPath, "utf8")).toBe("new\n");
        expect(statSync(temporaryPath).mode & 0o777).toBe(0o600);
      },
    });

    expect(readFileSync(file, "utf8")).toBe("new\n");
    expect(statSync(file).mode & 0o777).toBe(0o600);
  });

  it("preserves the prior value and removes temporary files after publication failure", () => {
    const { file, root } = scratchFile();
    writeFileSync(file, "old\n", { encoding: "utf8", mode: 0o600 });

    expect(() =>
      atomicWritePrivateFile(file, "new\n", {
        beforeRename: () => {
          throw new Error("synthetic publication failure");
        },
      }),
    ).toThrow("synthetic publication failure");

    expect(readFileSync(file, "utf8")).toBe("old\n");
    expect(readdirSync(root)).toEqual(["state.json"]);
  });

  it("validates worker manifests without exposing secret values", () => {
    expect(parseWorkerAdminManifest(JSON.stringify(manifest()))).toEqual(
      manifest(),
    );
    expect(() =>
      parseWorkerAdminManifest(
        JSON.stringify(manifest({ schema_id: "cartulary.unknown.v1" })),
      ),
    ).toThrow(/unsupported schema/);
    expect(() =>
      parseWorkerAdminManifest(
        JSON.stringify({
          ...manifest(),
          worker_admins: [
            manifest().worker_admins[0],
            {
              ...manifest().worker_admins[0],
              parallel_index: 1,
              password: "DoNotExposeThisPassword!",
            },
          ],
        }),
      ),
    ).toThrow(/duplicate user_id/);
    let validationError: unknown;
    try {
      parseWorkerAdminManifest(
        JSON.stringify({
          ...manifest(),
          worker_admins: [{ ...manifest().worker_admins[0], password: 42 }],
        }),
      );
    } catch (error) {
      validationError = error;
    }
    expect(validationError).toBeInstanceOf(Error);
    expect(String(validationError)).not.toContain("PrivateWorker0Pass!");
    expect(String(validationError)).not.toContain("DoNotExposeThisPassword!");
  });
});
