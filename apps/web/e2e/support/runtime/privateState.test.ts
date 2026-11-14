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
import { atomicLocalFile } from "../../../../../tools/harness/runtime/secure-local-files.mjs";
import {
  parseWorkerAdminManifest,
  workerAdminManifestSchemaID,
} from "../auth/workerAdmin";

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
    const { file, root } = scratchFile();
    writeFileSync(file, "old\n", { encoding: "utf8", mode: 0o600 });

    function* contents() {
      yield "new\n";
      expect(readFileSync(file, "utf8")).toBe("old\n");
      const temporary = readdirSync(root).find((name) =>
        name.startsWith(".publishing-"),
      );
      expect(temporary).toBeDefined();
      if (temporary === undefined)
        throw new Error("missing temporary publication");
      const temporaryPath = join(root, temporary);
      expect(readFileSync(temporaryPath, "utf8")).toBe("new\n");
      expect(statSync(temporaryPath).mode & 0o777).toBe(0o600);
    }
    atomicLocalFile(file, contents(), { replace: true });

    expect(readFileSync(file, "utf8")).toBe("new\n");
    expect(statSync(file).mode & 0o777).toBe(0o600);
  });

  it("preserves the prior value and removes temporary files after publication failure", () => {
    const { file, root } = scratchFile();
    writeFileSync(file, "old\n", { encoding: "utf8", mode: 0o600 });

    function* failedContents() {
      yield "new\n";
      throw new Error("synthetic publication failure");
    }
    expect(() =>
      atomicLocalFile(file, failedContents(), { replace: true }),
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
          worker_admins: [{ ...manifest().worker_admins[0], password: 42 }],
        }),
      ),
    ).toThrow(/password must be a non-empty string/);

    for (const { contents, reason, secrets } of [
      {
        contents: `${JSON.stringify(manifest())} trailing`,
        reason: /invalid JSON/,
        secrets: ["PrivateWorker0Pass!"],
      },
      {
        contents: JSON.stringify({
          ...manifest(),
          worker_admins: [
            { ...manifest().worker_admins[0], parallel_index: -1 },
          ],
        }),
        reason: /parallel_index must be a non-negative integer/,
        secrets: ["PrivateWorker0Pass!"],
      },
      {
        contents: JSON.stringify({
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
        reason: /duplicate user_id/,
        secrets: ["PrivateWorker0Pass!", "DoNotExposeThisPassword!"],
      },
    ]) {
      let validationError: unknown;
      try {
        parseWorkerAdminManifest(contents);
      } catch (error) {
        validationError = error;
      }
      expect(validationError).toBeInstanceOf(Error);
      expect(String(validationError)).toMatch(reason);
      for (const secret of secrets) {
        expect(contents).toContain(secret);
        expect(String(validationError)).not.toContain(secret);
      }
    }
  });
});
