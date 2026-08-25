package stream

import (
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// PostgresStream is the single stateful durable-stream component. Cohesive
// sources implement replay, sequencing, tailing, and retention over this
// shared database handle and clock.
type PostgresStream struct {
	db  postgres.DB
	now func() time.Time
}

func NewPostgresStream(db postgres.DB, now func() time.Time) *PostgresStream {
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &PostgresStream{db: db, now: now}
}
