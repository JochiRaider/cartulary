package stream

import (
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// PostgresStream is the single stateful durable-stream component. Cohesive
// sources implement replay, sequencing, tailing, and retention over this
// shared database handle.
type PostgresStream struct {
	db dbport.DB
}

func NewPostgresStream(db dbport.DB) (*PostgresStream, error) {
	if db == nil {
		return nil, errors.New("collaboration PostgreSQL stream dependency is required")
	}
	return &PostgresStream{db: db}, nil
}
