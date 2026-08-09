package storage

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// Store is the private physical projection-storage boundary. Provider SQL is
// moved behind this type one provider at a time in WS-16 through WS-25.
type Store struct {
	db postgres.DB
}

func New(db postgres.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("projection storage database is required")
	}
	return &Store{db: db}, nil
}

func (store *Store) Ready() bool {
	return store != nil && store.db != nil
}
