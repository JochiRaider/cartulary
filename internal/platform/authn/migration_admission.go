package authn

import (
	"context"
	"database/sql"
)

// RetainedRouteStatePresent is the Auth-owned pre-migration presence probe.
// It includes every retained receipt, regardless of age or source availability.
func RetainedRouteStatePresent(ctx context.Context, conn *sql.Conn, routes []string) (bool, error) {
	var exists bool
	if err := conn.QueryRowContext(ctx, `SELECT to_regclass('public.route_idempotency') IS NOT NULL`).Scan(&exists); err != nil || !exists {
		return false, err
	}
	for _, route := range routes {
		var retained bool
		if err := conn.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM public.route_idempotency WHERE route_key=$1)`, route).Scan(&retained); err != nil {
			return false, err
		}
		if retained {
			return true, nil
		}
	}
	return false, nil
}
