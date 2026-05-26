package incidentbundles

import "sync"

var incidentBundleWorkerHook = struct {
	sync.Mutex
	fn func(string)
}{}

func SetIncidentBundleWorkerStartHookForTesting(fn func(string)) func() {
	incidentBundleWorkerHook.Lock()
	previous := incidentBundleWorkerHook.fn
	incidentBundleWorkerHook.fn = fn
	incidentBundleWorkerHook.Unlock()
	return func() {
		incidentBundleWorkerHook.Lock()
		incidentBundleWorkerHook.fn = previous
		incidentBundleWorkerHook.Unlock()
	}
}

func runIncidentBundleWorkerStartHook(jobKind string) {
	incidentBundleWorkerHook.Lock()
	fn := incidentBundleWorkerHook.fn
	incidentBundleWorkerHook.Unlock()
	if fn != nil {
		fn(jobKind)
	}
}
