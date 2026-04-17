package jobs

// Manager is a placeholder background-job shell.
// TODO: replace this bootstrap shell with the canonical job resource and runner.
type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}
