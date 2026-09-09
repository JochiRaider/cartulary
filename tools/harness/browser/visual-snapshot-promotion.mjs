import { existsSync, lstatSync, mkdirSync, mkdtempSync, readFileSync, readdirSync, renameSync, rmSync, writeFileSync } from "node:fs";
import path from "node:path";

export function stageVisualSnapshotCandidate(source, destination) {
  const validateExisting = () => {
    const info = lstatSync(destination);
    if (!info.isDirectory() || info.isSymbolicLink() || info.uid !== process.getuid() || (info.mode & 0o777) !== 0o700) throw new Error("visual snapshot candidate must be an owned private directory");
  };
  if (existsSync(destination)) { validateExisting(); return; }
  const staging = mkdtempSync(path.join(path.dirname(destination), ".snapshot-staging-"));
  const copy = (from, to) => {
    const info = lstatSync(from);
    if (!info.isDirectory() || info.isSymbolicLink() || info.uid !== process.getuid()) throw new Error("visual snapshot source must be an owned directory");
    for (const name of readdirSync(from).sort()) {
      const input = path.join(from, name);
      const output = path.join(to, name);
      const entry = lstatSync(input);
      if (entry.isSymbolicLink() || entry.uid !== process.getuid()) throw new Error("visual snapshot source contains an unowned entry or symlink");
      if (entry.isDirectory()) {
        mkdirSync(output, { mode: 0o700 });
        copy(input, output);
      } else if (entry.isFile()) {
        writeFileSync(output, readFileSync(input), { mode: 0o600, flag: "wx" });
      } else throw new Error("visual snapshot source contains a non-regular entry");
    }
  };
  try {
    copy(source, staging);
    try { renameSync(staging, destination); }
    catch (error) {
      if (!["EEXIST", "ENOTEMPTY"].includes(error.code) || !existsSync(destination)) throw error;
      validateExisting();
    }
  } finally { rmSync(staging, { recursive: true, force: true }); }
}

const defaultFileSystem = Object.freeze({ existsSync, renameSync, rmSync });

function rollbackRename(fileSystem, from, to, errors) {
  if (!fileSystem.existsSync(from)) return;
  try {
    fileSystem.renameSync(from, to);
  } catch (error) {
    errors.push(error);
  }
}

export function promoteVisualSnapshotCandidate(
  {
    sourceSnapshots,
    candidateSnapshots,
    sourceManifest,
    candidateManifest,
    snapshotBackup,
    manifestBackup,
  },
  fileSystem = defaultFileSystem,
) {
  if (
    !fileSystem.existsSync(candidateSnapshots) ||
    !fileSystem.existsSync(candidateManifest) ||
    fileSystem.existsSync(snapshotBackup) ||
    fileSystem.existsSync(manifestBackup)
  ) {
    throw new Error("visual snapshot promotion candidate or backup state is invalid");
  }
  let sourceSnapshotsMoved = false;
  let candidateSnapshotsMoved = false;
  let sourceManifestMoved = false;
  let candidateManifestMoved = false;
  try {
    fileSystem.renameSync(sourceSnapshots, snapshotBackup);
    sourceSnapshotsMoved = true;
    fileSystem.renameSync(candidateSnapshots, sourceSnapshots);
    candidateSnapshotsMoved = true;
    if (fileSystem.existsSync(sourceManifest)) {
      fileSystem.renameSync(sourceManifest, manifestBackup);
      sourceManifestMoved = true;
    }
    fileSystem.renameSync(candidateManifest, sourceManifest);
    candidateManifestMoved = true;
  } catch (error) {
    const rollbackErrors = [];
    if (candidateManifestMoved) {
      rollbackRename(fileSystem, sourceManifest, candidateManifest, rollbackErrors);
    }
    if (sourceManifestMoved) {
      rollbackRename(fileSystem, manifestBackup, sourceManifest, rollbackErrors);
    }
    if (candidateSnapshotsMoved) {
      rollbackRename(fileSystem, sourceSnapshots, candidateSnapshots, rollbackErrors);
    }
    if (sourceSnapshotsMoved) {
      rollbackRename(fileSystem, snapshotBackup, sourceSnapshots, rollbackErrors);
    }
    if (rollbackErrors.length > 0) {
      throw new AggregateError(
        [error, ...rollbackErrors],
        "visual snapshot promotion and rollback failed",
      );
    }
    throw error;
  }

  // Promotion is already complete. Backup cleanup is best-effort so a retained
  // run-root cleanup problem cannot turn a successful atomic swap into a failed
  // update with changed tracked files.
  try {
    fileSystem.rmSync(snapshotBackup, { recursive: true, force: true });
    fileSystem.rmSync(manifestBackup, { force: true });
  } catch {
    // Backups live only under the retained run root and contain no page data.
  }
}
