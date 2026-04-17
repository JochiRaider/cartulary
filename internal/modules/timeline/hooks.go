package timeline

import (
	"sync"

	"github.com/google/uuid"
)

type StoreHooks struct {
	BeforeCommit func(routeKey string, recordID uuid.UUID) error
}

var (
	storeHooksMu sync.RWMutex
	storeHooks   StoreHooks
)

func SetStoreHooksForTesting(hooks StoreHooks) func() {
	storeHooksMu.Lock()
	previous := storeHooks
	storeHooks = hooks
	storeHooksMu.Unlock()

	return func() {
		storeHooksMu.Lock()
		storeHooks = previous
		storeHooksMu.Unlock()
	}
}

func currentStoreHooks() StoreHooks {
	storeHooksMu.RLock()
	defer storeHooksMu.RUnlock()
	return storeHooks
}
