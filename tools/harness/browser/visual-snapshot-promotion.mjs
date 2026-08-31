import { existsSync, renameSync, rmSync } from "node:fs";

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
