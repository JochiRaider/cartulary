# Reference Pack development cutover

The PostgreSQL 18.6 v3 baseline creates final Reference Pack and Reporting state
directly in `00026_reference_data.sql` and `00023_reporting.sql`. It retains
immutable dependency evidence, validation summaries, explicit operator attribution,
Base fallback attribution, portable-content inventories and private trigger ACLs.
The previous per-owner preflight categories and migrations 30–95 are retired;
there is no replay or conversion path. See the complete
[consolidation mapping](../handoffs/postgresql-18.6-ddl-consolidation.md#complete-old-to-new-mapping).

This procedure describes the pre-production Reference Pack cutover. Current implementation and evidence status is tracked in the [remediation handoff](../handoffs/testing-assurance-remediation.md). This document does not authorize a production reset or establish completed validation or formal production conformance.

## Preserve the historical deployment

Keep the previous application, its configuration, encryption keys, database lineage, object-store configuration and retained filesystem roots available together. With that matching application, create and inspect a backup and verify restoration into an isolated destination using the [operator recovery procedure](cartulary-dev-guide.md#125-backup-restore-and-failure-modes). Export incident artifacts that need independent retention while the historical application can still render and export them. Record the application version and the successful backup/restore-verification result with those artifacts.

A historical backup remains a recovery artifact for its matching historical application. Resetting the development database does not make that backup, an old incident bundle, a checksum-labelled pack, or an unbound historical snapshot importable by the new format. The cutover provides no converter that manufactures signatures, successful envelopes, snapshot bindings or provenance. Preserve the historical environment when its incident identities must remain usable.

## Inspect the preflight

Use the normal migration entry point, `make db-migrate`, against a deliberately selected development configuration. The current catalog is lineage `cartulary.prod_ddl_rebaseline.v3`, versions 1–29. After cryptographic-state admission it rejects v2 and all other previous lineages before applying SQL and reports the exact reset-only remedy. Retained state without compatible crypto identity fails earlier under the cryptographic-state gate. The retired per-owner cutover inventories and migration steps 30–95 are no longer runnable. The new baseline creates immutable Reference Pack state directly; it neither converts historical trust nor invents successful provenance.

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

Earlier documentation reported two disposable rehearsals through migration 63, but the referenced ledger is unavailable in this checkout. Treat that statement as unverified historical context, not current cutover evidence. For each new rehearsal, retain the selected configuration identity, migration/preflight results, reset preview, explicit reset invocation and startup result. Restoration of new-format backups performs historical trust validation, index rebuilding, exact-set pin validation and required-content readiness checks. Those checks remain necessary for the actual backup being restored; a successful database reset is not restoration evidence.

## Operator import and unavailable Base content

With the server lifecycle worker running, place a canonical bundle under the configured Reference Pack root's `incoming` directory and invoke `operator reference-pack import <bundle_name>`. The command admits through the same coordinator as HTTP and observes the durable Job. The filename is 1–128 ASCII characters beginning with a letter or digit, followed by letters, digits, dot, underscore, plus or hyphen. The command never accepts an arbitrary path. Its closed canonical JSON result includes the operation and Job identities, with explicit nulls where no successful verification exists. Stopping observation does not cancel or repeat an admitted Job; a stopped worker leaves it queued.

If a required active registry and its Base fallback become unavailable, the runtime retains the invalidation, clears the unusable current selection and reports not ready. Historical sets and provenance are preserved. Restore the retained content using the matching recovery procedure and reconcile Base before serving traffic again. Never repair readiness by relabelling failed content or selecting a partial registry set. A healthy imported replacement can remain current after loss of an inactive historical Base; an administrative disablement that would remove the last healthy registry is rejected.


## Removed content and backup retention

The Reporting baseline (`00023_reporting.sql`) retains imported Reporting artifacts as immutable source evidence; it does not create native snapshots, Jobs or approvals from imported history. The current incident-bundle format carries the exact source pack catalog and optional artifact files. Sensitive reveal maps are excluded under the adopted Reporting portability contract.

Removal releases consumer availability immediately. Byte collection runs after terminal-Job reconciliation at startup and once per minute thereafter. It reclaims removed extracted copies, completed-operation scratch publications and abandoned immutable publications only after rechecking retention. Successful envelopes keep their exact signed containers for historical verification, so removal does not promise to reclaim all storage associated with that logical version. Backup capture holds retention through database and byte streaming; a backup already in progress may preserve copies that later become collectible. Exact reimport writes new physical objects and retains removal history. Do not manually delete retained containers or alter object availability flags to reclaim space.
