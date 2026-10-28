package recovery

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// validateRestoredPostgresRows scans the actual restore transaction. A VALID
// catalog constraint is not evidence about rows loaded with referential triggers
// disabled. Identifiers and comparison operators come from the admitted catalog,
// never from backup payloads. The stopped target excludes concurrent writers.
func validateRestoredPostgresRows(ctx context.Context, tx pgx.Tx) error {
	rows, err := tx.Query(ctx, restoredForeignKeyQueries)
	if err != nil {
		return fmt.Errorf("inspect restore foreign keys: %w", err)
	}
	type check struct{ name, query string }
	var checks []check
	for rows.Next() {
		var c check
		if err := rows.Scan(&c.name, &c.query); err != nil {
			rows.Close()
			return err
		}
		checks = append(checks, c)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, c := range checks {
		var invalid bool
		if err := tx.QueryRow(ctx, c.query).Scan(&invalid); err != nil {
			return fmt.Errorf("scan restored foreign key %s: %w", c.name, err)
		}
		if invalid {
			return fmt.Errorf("%w: restored foreign key %s", ErrVNextBackup, c.name)
		}
	}
	// These owner predicates cover source/claim correspondence and the JSON
	// support references that ordinary relational foreign keys cannot express.
	var valid bool
	if err := tx.QueryRow(ctx, `SELECT
 public.entities_active_identifier_claims_are_valid_v1()
 AND public.parties_active_key_claims_are_valid_v1()
 AND public.indicator_active_identities_are_valid()`).Scan(&valid); err != nil {
		return fmt.Errorf("validate restored owner invariants: %w", err)
	}
	if !valid {
		return fmt.Errorf("%w: restored owner invariants", ErrVNextBackup)
	}
	return nil
}

// MATCH SIMPLE exempts any-null keys; MATCH FULL exempts only all-null keys.
// Use each FK's own parent/child equality operator, including extension types.
const restoredForeignKeyQueries = `
SELECT k.conname::text,
 format('SELECT EXISTS (SELECT 1 FROM %I.%I c WHERE %s (%s) AND NOT EXISTS (SELECT 1 FROM %I.%I p WHERE %s))',
   ns.nspname,child.relname,
   CASE k.confmatchtype WHEN 'f' THEN format('(%s) AND NOT (%s) OR',parts.any_null,parts.all_null) ELSE '' END,
   parts.all_present,pns.nspname,parent.relname,parts.equal)
FROM pg_catalog.pg_constraint k
JOIN pg_catalog.pg_class child ON child.oid=k.conrelid
JOIN pg_catalog.pg_namespace ns ON ns.oid=child.relnamespace
JOIN pg_catalog.pg_class parent ON parent.oid=k.confrelid
JOIN pg_catalog.pg_namespace pns ON pns.oid=parent.relnamespace
CROSS JOIN LATERAL (
 SELECT string_agg(format('c.%I IS NULL',ca.attname),' OR ' ORDER BY key.ordinality) AS any_null,
        string_agg(format('c.%I IS NULL',ca.attname),' AND ' ORDER BY key.ordinality) AS all_null,
        string_agg(format('c.%I IS NOT NULL',ca.attname),' AND ' ORDER BY key.ordinality) AS all_present,
        string_agg(format('p.%I OPERATOR(%I.%s) c.%I',pa.attname,ons.nspname,op.oprname,ca.attname),' AND ' ORDER BY key.ordinality) AS equal
 FROM unnest(k.conkey,k.confkey,k.conpfeqop) WITH ORDINALITY key(cnum,pnum,equality,ordinality)
 JOIN pg_catalog.pg_attribute ca ON ca.attrelid=child.oid AND ca.attnum=key.cnum
 JOIN pg_catalog.pg_attribute pa ON pa.attrelid=parent.oid AND pa.attnum=key.pnum
 JOIN pg_catalog.pg_operator op ON op.oid=key.equality
 JOIN pg_catalog.pg_namespace ons ON ons.oid=op.oprnamespace
) parts
WHERE ns.nspname='public' AND k.contype='f'
ORDER BY child.relname,k.conname`
