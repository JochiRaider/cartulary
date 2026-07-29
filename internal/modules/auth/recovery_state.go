package auth

import recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"

func RecoveryStateContribution() recoverystate.Contribution {
	tables := recoverystate.AuthoritativeTables(
		"account_preferences",
		"bootstrap_tokens",
		"enterprise_auth_bindings",
		"enterprise_auth_providers",
		"users",
	)
	tables = append(tables,
		recoverystate.SecurityStateTables(
			"auth.invalidate_enterprise_transactions.v1",
			"enterprise_auth_transactions",
		)...,
	)
	tables = append(tables,
		recoverystate.SecurityStateTables(
			"auth.invalidate_pending_totp.v1",
			"pending_totp_enrollments",
		)...,
	)
	tables = append(tables,
		recoverystate.SecurityStateTables(
			"auth.invalidate_route_idempotency.v1",
			"route_idempotency",
		)...,
	)
	tables = append(tables,
		recoverystate.SecurityStateTables(
			"auth.invalidate_sessions.v1",
			"user_sessions",
		)...,
	)
	return recoverystate.NewContribution("module.auth", tables)
}
