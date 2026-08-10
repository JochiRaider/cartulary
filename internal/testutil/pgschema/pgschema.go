package pgschema

import (
	"strings"

	dbmigrations "github.com/JochiRaider/cartulary/db/migrations"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
)

const (
	runnerIdentity = "cartulary-postgres-migrate/goose/v3.27.0"
)

// Hash returns a stable hash of the migration inputs that define the current
// test schema template.
func Hash() (string, error) {
	source, err := dbmigrations.Source()
	if err != nil {
		return "", err
	}
	return database_migrations.SchemaHash(source, runnerIdentity)
}

func MustHash() string {
	hash, err := Hash()
	if err != nil {
		panic(err)
	}
	return hash
}

func ShortHash(hash string) string {
	hash = strings.TrimSpace(hash)
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
