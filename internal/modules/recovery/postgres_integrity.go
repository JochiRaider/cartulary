package recovery

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/maintenancequery"
	"github.com/JochiRaider/cartulary/internal/platform/sqlshape"
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
	type check struct {
		name, child, parent, parentSchema, match                string
		childColumns, parentColumns, operatorSchemas, operators []string
	}
	var checks []check
	for rows.Next() {
		var c check
		if err := rows.Scan(&c.name, &c.child, &c.parent, &c.parentSchema, &c.match, &c.childColumns, &c.parentColumns, &c.operatorSchemas, &c.operators); err != nil {
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
		if c.parentSchema != "public" || len(c.childColumns) == 0 || len(c.childColumns) != len(c.parentColumns) || len(c.childColumns) != len(c.operatorSchemas) || len(c.childColumns) != len(c.operators) {
			return fmt.Errorf("unsupported restored foreign key %s", c.name)
		}
		child, err := maintenancequery.From(c.child)
		if err != nil {
			return err
		}
		parent, err := maintenancequery.From(c.parent)
		if err != nil {
			return err
		}
		present, missing, equalities := []sqlshape.Fragment{}, []sqlshape.Fragment{}, []sqlshape.Fragment{}
		for i, column := range c.childColumns {
			childColumn := sqlshape.Format("c.%s", sqlshape.Identifier(column))
			parentColumn := sqlshape.Format("p.%s", sqlshape.Identifier(c.parentColumns[i]))
			operator, err := sqlshape.Operator(c.operatorSchemas[i], c.operators[i])
			if err != nil {
				return err
			}
			present = append(present, sqlshape.Format("%s IS NOT NULL", childColumn))
			missing = append(missing, sqlshape.Format("%s IS NULL", childColumn))
			equalities = append(equalities, sqlshape.Format("%s %s %s", parentColumn, operator, childColumn))
		}
		predicate := sqlshape.Format("(%s) AND NOT EXISTS (SELECT 1 %s p WHERE %s)", sqlshape.Join(present, sqlshape.Literal("AND")), parent, sqlshape.Join(equalities, sqlshape.Literal("AND")))
		switch c.match {
		case "f":
			predicate = sqlshape.Format("((%s) AND NOT (%s)) OR (%s)", sqlshape.Join(missing, sqlshape.Literal("OR")), sqlshape.Join(missing, sqlshape.Literal("AND")), predicate)
		case "s":
		default:
			return fmt.Errorf("unsupported restored foreign key match %s", c.name)
		}
		query := sqlshape.Format("SELECT EXISTS (SELECT 1 %s c WHERE %s)", child, predicate)
		var invalid bool
		if err := tx.QueryRow(ctx, sqlshape.Text(query)).Scan(&invalid); err != nil {
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
 AND public.indicator_active_identities_are_valid()
 AND public.network_flow_orientation_state_is_valid_v1()`).Scan(&valid); err != nil {
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
SELECT k.conname::text,child.relname::text,parent.relname::text,pns.nspname::text,k.confmatchtype::text,
 array_agg(ca.attname::text ORDER BY key.ordinality),array_agg(pa.attname::text ORDER BY key.ordinality),
 array_agg(ons.nspname::text ORDER BY key.ordinality),array_agg(op.oprname::text ORDER BY key.ordinality)
FROM pg_catalog.pg_constraint k
JOIN pg_catalog.pg_class child ON child.oid=k.conrelid
JOIN pg_catalog.pg_namespace ns ON ns.oid=child.relnamespace
JOIN pg_catalog.pg_class parent ON parent.oid=k.confrelid
JOIN pg_catalog.pg_namespace pns ON pns.oid=parent.relnamespace
CROSS JOIN LATERAL unnest(k.conkey,k.confkey,k.conpfeqop) WITH ORDINALITY key(cnum,pnum,equality,ordinality)
JOIN pg_catalog.pg_attribute ca ON ca.attrelid=child.oid AND ca.attnum=key.cnum
JOIN pg_catalog.pg_attribute pa ON pa.attrelid=parent.oid AND pa.attnum=key.pnum
JOIN pg_catalog.pg_operator op ON op.oid=key.equality
JOIN pg_catalog.pg_namespace ons ON ons.oid=op.oprnamespace
WHERE ns.nspname='public' AND k.contype='f'
GROUP BY k.oid,k.conname,child.relname,parent.relname,pns.nspname,k.confmatchtype
ORDER BY child.relname,k.conname`
