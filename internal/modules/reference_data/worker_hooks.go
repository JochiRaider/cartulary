package reference_data

import "sync"

var referencePackWorkerHook = struct {
	sync.Mutex
	fn func(string)
}{}

func SetReferencePackWorkerStartHookForTesting(fn func(string)) func() {
	referencePackWorkerHook.Lock()
	previous := referencePackWorkerHook.fn
	referencePackWorkerHook.fn = fn
	referencePackWorkerHook.Unlock()
	return func() {
		referencePackWorkerHook.Lock()
		referencePackWorkerHook.fn = previous
		referencePackWorkerHook.Unlock()
	}
}

func runReferencePackWorkerStartHook(jobKind string) {
	referencePackWorkerHook.Lock()
	fn := referencePackWorkerHook.fn
	referencePackWorkerHook.Unlock()
	if fn != nil {
		fn(jobKind)
	}
}
