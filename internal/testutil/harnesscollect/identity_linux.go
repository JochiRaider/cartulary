//go:build linux

package harnesscollect

// ProcessIdentity captures the collector's physical identity without starting
// collection. Callers transport this private proof; retained records contain
// only its opaque digest. The adapter remains the single counter/identity owner.
func ProcessIdentity(pid int) (map[string]any, error) {
	a, err := newAdapter()
	if err != nil {
		return nil, err
	}
	p, err := a.proof(pid)
	if err != nil {
		return nil, err
	}
	return map[string]any{"pid": p.PID, "start": p.Start, "identity": p.Identity, "stat": p.Stat}, nil
}
