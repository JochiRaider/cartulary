import assert from "node:assert/strict";
import test from "node:test";
import { sqlRelationAccesses } from "../sql-relations.mjs";

export function registerBoundariesTests() {
  const cases = [
    ['Go comment delimiter data is not SQL', 'package p\nvar forbidden = []string{"/*", "--", "/* partial"}', []],
    ['Go nested leading SQL comments', 'package p\nconst query = `/* outer /* inner */ comment */ UPDATE records SET id=1`', ['write:records']],
    ['Go commented prose is not SQL', 'package p\nconst note = `/* note */ saved record`', []],
    ['Go labels are not SQL', 'package p\nconst label = "table"', []],
    ['Go update comments and qualification', 'package p\nconst sql = `UPDATE /* comment */ "public"."records" SET row_version=2`', ['write:records']],
    ['Go insert comments between keywords', 'package p\nconst sql = `INSERT /* comment */ INTO records(id) VALUES(1)`', ['write:records']],
    ['Go delete line comment', 'package p\nconst sql = `DELETE -- comment\nFROM records`', ['write:records']],
    ['CTE relation and its physical body', 'WITH events AS (SELECT * FROM change_sets) SELECT * FROM events', ['read:change_sets']],
    ['recursive scope', 'WITH RECURSIVE events AS (SELECT * FROM records UNION ALL SELECT r.* FROM events e JOIN records r ON true) SELECT * FROM events', ['read:records', 'read:records']],
    ['CTE never hides qualified physical relation', 'WITH records AS (VALUES (1)) SELECT * FROM records JOIN public.records r ON true', ['read:records']],
    ['nested scope does not escape', 'SELECT * FROM (WITH records AS (VALUES (1)) SELECT * FROM records) c JOIN records r ON true', ['read:records']],
    ['CTE preceding name shadows only following body', 'WITH records AS (SELECT * FROM records), other AS (SELECT * FROM records) SELECT * FROM other', ['read:records']],
    ['set returning functions and lateral subquery', 'SELECT * FROM unnest($1::uuid[]) x LEFT JOIN LATERAL (SELECT * FROM records) r ON true', ['read:records']],
    ['subqueries inside function expressions', 'SELECT coalesce((SELECT id FROM records), 0) FROM unnest($1::int[])', ['read:records']],
    ['comma relations', 'SELECT * FROM records r, jobs j WHERE r.id = j.id', ['read:records', 'read:jobs']],
    ['quoted and qualified names', 'SELECT * FROM "public"."records" r JOIN "MixedCase" m ON true', ['read:records', 'read:MixedCase']],
    ['quoted keywords remain identifiers', 'SELECT * FROM "order" AS "from" FOR UPDATE OF "from"', ['read:order', 'lock:order']],
    ['comments and values are not relations', "SELECT 'FROM hidden', $$JOIN concealed$$ FROM records /* JOIN hidden */ -- FROM hidden\n", ['read:records']],
    ['expression FROM is not a relation', 'SELECT extract(epoch FROM created_at) FROM records WHERE id IS NOT DISTINCT FROM $1', ['read:records']],
    ['write target and read source', 'INSERT INTO records(id) SELECT id FROM jobs ON CONFLICT(id) DO UPDATE SET id = excluded.id', ['write:records', 'read:jobs']],
    ['write CTE body', 'WITH changed AS (UPDATE records SET id = 1 RETURNING *) DELETE FROM jobs USING changed WHERE true', ['write:records', 'write:jobs']],
    ['lock alias association', 'SELECT * FROM jobs j JOIN records r ON true FOR NO KEY UPDATE OF j', ['read:jobs', 'read:records', 'lock:jobs']],
    ['all row lock modes', 'SELECT * FROM jobs FOR KEY SHARE', ['read:jobs', 'lock:jobs']],
    ['statement scope prevents lock leakage', 'SELECT * FROM jobs; SELECT * FROM records FOR UPDATE', ['read:jobs', 'read:records', 'lock:records']],
    ['statement scope prevents CTE leakage', 'WITH records AS (VALUES (1)) SELECT * FROM records; SELECT * FROM records', ['read:records']],
    ['dynamic target remains explicit', 'SELECT * FROM %s JOIN records r ON true', ['read:<dynamic>', 'read:records']],
    ['Go comments and error copy are not SQL', 'package fixture\n// SELECT * FROM hidden\nconst query = `SELECT * FROM records`\nconst message = "update saved view"', ['read:records']],
    ['Go escapes do not break SQL extraction', 'package fixture\nconst alert = "\\a"\nconst query = "SELECT * FROM \\"records\\""', ['read:records']],
    ['Go constant composition retains relations', 'package fixture\nconst body = `SELECT * FROM records`\nvar query = `SELECT * FROM (` + body + `) r`', ['read:records', 'read:records']],
    ['Go expression composition retains following relations', 'package fixture\nvar query = `SELECT * FROM records WHERE id IN (` + strings.Join(ids, ",") + `) AND EXISTS (SELECT * FROM jobs)`', ['read:records', 'read:jobs']],
    ['explicit table lock', 'LOCK TABLE public.jobs, records IN SHARE MODE', ['lock:jobs', 'lock:records']],
    ['truncate targets', 'TRUNCATE TABLE jobs, records', ['write:jobs', 'write:records']],
    ['SELECT INTO is a physical write', 'SELECT * INTO records FROM jobs', ['write:records', 'read:jobs']],
  ];
  for (const [label, sql, expected] of cases) {
    test(`SQL relations: ${label}`, () => {
      assert.deepEqual(sqlRelationAccesses(sql).map(({ operation, table }) => `${operation}:${table}`), expected);
    });
  }
  for (const sql of ['SELECT * FROM', 'WITH events AS SELECT * FROM records', 'SELECT * FROM (records JOIN jobs ON true)', 'SELECT * FROM records FOR UPDATE OF missing', 'MERGE INTO records USING jobs ON true']) {
    test(`SQL relations reject unclassified access: ${sql}`, () => assert.throws(() => sqlRelationAccesses(sql)));
  }
}
