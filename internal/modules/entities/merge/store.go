package merge

import (
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool      postgres.DB
	authStore *authn.Store
	ports     entityStorePorts
}

func NewStore(pool postgres.DB) *Store {
	return &Store{
		pool:      pool,
		authStore: authn.NewStore(pool),
		ports:     newEntityStorePorts(pool),
	}
}
