# Reference Pack development cutover

Migration 00063 separates snapshot materialization from release/preview identity. Reporting preflight runs before schema mutation and reports `historical_snapshots`, `retained_reporting_jobs`, `imported_reporting_artifacts`, and `retained_composition_previews` as category counts. Any retained affected history rejects this cutover; it is never assigned guessed release IDs or admission times. Preserve the matching application and backup before using the deliberate disposable-development reset below.

Migration 00060 removes direct execution grants on Reporting's private artifact immutability trigger; ordinary writes still execute the trigger. Migration 00061 adds current Base fallback attribution. Both application preflight and migration 00061 reject intermediate state containing `unclassified_safety_fallbacks` rather than infer current selection causes by wall-clock event ordering. Preserve a matching backup before deliberately resetting disposable development data. Migration 00062 replaces the fallback foreign key's partial index with the full index required by the adopted DDL profile; it does not change retained data.

Migration 00055 rejects retained portability operations lacking the closed embedded-content inventory and exact required-member classification, reporting `unclassified_portable_content`. It adds a database shape constraint only after the rejection check passes. It never infers that an old imported artifact had no required members. Preserve incompatible backups with the matching historical application; resetting does not convert them.

Migration 00054 rejects nonterminal verification operations without frozen dependency reads (`pending_dependency_captures`) and retained logical versions with unproven dependency history (`unproven_dependency_history`). It reports category counts before schema mutation and never upgrades that state into verified dependency evidence. Its new immutable dependency-key records participate in backup and recovery.

Migration 00053 rejects retained import Jobs that predate explicit human/local-operator attribution, reporting `retired_import_attribution`. It does not invent a submitting identity. The new local operator Job retains its operation UUID even after Job compaction.

Migration 00052 also rejects retained content-rejection attempts that predate canonical validation summaries. The preflight reports `historical_validation_summaries` as a category count. A reason code cannot reconstruct the original findings, their count, or their primary issue. Preserve that database and its matching historical application before deliberately resetting disposable development data; the migration never manufactures diagnostic history.

This procedure supports the validated pre-production Reference Pack cutover. The [implementation ledger](../handoffs/reference-pack-remediation.md) records the completed implementation, disposable reset rehearsal and passing final release gate. This document does not authorize a production reset or establish formal production conformance.

## Preserve the historical deployment

Keep the previous application, its configuration, encryption keys, database lineage, object-store configuration and retained filesystem roots available together. With that matching application, create and inspect a backup and verify restoration into an isolated destination using the [operator recovery procedure](cartulary-dev-guide.md#125-backup-restore-and-failure-modes). Export incident artifacts that need independent retention while the historical application can still render and export them. Record the application version and the successful backup/restore-verification result with those artifacts.

A historical backup remains a recovery artifact for its matching historical application. Resetting the development database does not make that backup, an old incident bundle, a checksum-labelled pack, or an unbound historical snapshot importable by the new format. The cutover provides no converter that manufactures signatures, successful envelopes, snapshot bindings or provenance. Preserve the historical environment when its incident identities must remain usable.

## Inspect the preflight

Use the normal migration entry point, `make db-migrate`, against a deliberately selected development configuration. The application preflight runs before migration changes, and migrations 00046, 00048, 00051, 00054, 00055 and 00058 repeat the incompatible-state checks under locks. The application also inventories retained Indicator identities when upgrading an intermediate schema through migration 00048 and imported incidents whose source reference catalog was not retained before migration 00051; an existing immutable pack schema alone does not bypass those checks. It reports counts for retained legacy packs, pack payloads and Jobs, indicator identities, observations and historical artifacts requiring unavailable bindings. A nonzero incompatible category stops the cutover; it does not delete or rewrite that state. Even an apparently empty historical reference catalog must not be inferred from an old imported incident. Migration 00058 retires the four mutable legacy tables only when all are empty; legacy activation and audit rows also block it, including on an otherwise current schema. There is no cascade deletion or compatibility store.

Do not bypass the check by clearing individual tables, relabelling old signatures, editing migration history, or copying old snapshots into new tables. A successful fresh-state migration is evidence only for initialization, not a migration of historical trust or incident meaning.

## Deliberate development reset

Stop all processes using the selected development database. Verify the effective database and storage destinations against the retained historical inventory. Preserve backup and historical object roots independently of any disposable development paths.

Preview the repository-owned database reset:

```sh
make db-reset CARTULARY_CLEANUP_DRY_RUN=1
```

After checking the preview and confirming that this specific development database may be discarded, the explicit destructive invocation is:

```sh
make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset
```

The reset recreates the development database; it does not reset object storage. Use deliberately provisioned storage destinations for the new deployment. Keep historical backup and content roots intact, and do not run an object-store reset as an implicit part of this procedure. Consult `make help` for the current configuration and local setup surface rather than reusing commands from another deployment.

The development configuration starts with `reference_pack.claimed=false`; startup still establishes the required Base registries. For explicit pre-production administration testing, set `reference_pack.claimed=true` and configure `reference_packs.trust_bootstrap_path` with a canonical offline trust bootstrap for the repositories the operator actually trusts. Production profile-conformance claims remain subject to owner adoption and the target deployment's cryptographic compliance decision. No separate recreation or interchangeability report is required for this pre-production cutover. The bootstrap binds exact Ed25519 roots; the test fixture keys are test data, not deployment trust. Set `reference_packs.clock_trusted=true` only as an explicit operator assertion that the deployment clock is trustworthy. The default is false. Base registries initialize without this assertion; fresh operator TUF verification requires it.

## Validate and retain the result

Fresh startup must establish the three current-release Base registries before readiness. Verify that old operator packs were not implicitly promoted, incident identities were not copied or rekeyed, and historical backup roots remain untouched. A subsequent unchanged startup must reuse immutable content, envelopes and provenance, and preserve configuration revision identity.

The procedure has been rehearsed twice against an isolated disposable database, reaching migration 63 and verifying cleanup; exact results are in the ledger. No real deployment was migrated or reset. For each new rehearsal, retain the selected configuration identity, migration/preflight results, reset preview, explicit reset invocation and startup result. Restoration of new-format backups performs historical trust validation, index rebuilding, exact-set pin validation and required-content readiness checks. Those checks remain necessary for the actual backup being restored; a successful database reset is not restoration evidence.

## Operator import and unavailable Base content

With the server lifecycle worker running, place a canonical bundle under the configured Reference Pack root's `incoming` directory and invoke `operator reference-pack import <bundle_name>`. The command admits through the same coordinator as HTTP and observes the durable Job. The filename is 1–128 ASCII characters beginning with a letter or digit, followed by letters, digits, dot, underscore, plus or hyphen. The command never accepts an arbitrary path. Its closed canonical JSON result includes the operation and Job identities, with explicit nulls where no successful verification exists. Stopping observation does not cancel or repeat an admitted Job; a stopped worker leaves it queued.

If a required active registry and its Base fallback become unavailable, the runtime retains the invalidation, clears the unusable current selection and reports not ready. Historical sets and provenance are preserved. Restore the retained content using the matching recovery procedure and reconcile Base before serving traffic again. Never repair readiness by relabelling failed content or selecting a partial registry set. A healthy imported replacement can remain current after loss of an inactive historical Base; an administrative disablement that would remove the last healthy registry is rejected.


## Removed content and backup retention

Migration 00059 retains imported Reporting artifacts as immutable source evidence; it does not create native snapshots, Jobs or approvals from imported history. Incident-bundle format 5 carries the exact source pack catalog and optional artifact files. Reveal-map bytes remain omitted, with their source manifest entries preserved and explicit omission records.

Removal releases consumer availability immediately. Byte collection runs after terminal-Job reconciliation at startup and once per minute thereafter. It reclaims removed extracted copies, completed-operation scratch publications and abandoned immutable publications only after rechecking retention. Successful envelopes keep their exact signed containers for historical verification, so removal does not promise to reclaim all storage associated with that logical version. Backup capture holds retention through database and byte streaming; a backup already in progress may preserve copies that later become collectible. Exact reimport writes new physical objects and retains removal history. Do not manually delete retained containers or alter object availability flags to reclaim space.
