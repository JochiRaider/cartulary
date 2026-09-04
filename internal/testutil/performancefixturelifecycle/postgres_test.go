package performancefixturelifecycle

import (
	"context"
	"strings"
	"testing"
)

func TestWaitForNoTemplateConnectionsAllowsOnlyBoundedOwnedDrain(t *testing.T) {
	t.Parallel()
	owned := &ownedConnectionPIDs{}
	owned.Add(101)
	observations := [][]templateConnection{
		{{pid: 101, backendType: "client backend"}},
		{{pid: 202, leaderPID: 101, backendType: "parallel worker"}},
		{{pid: 303, backendType: "autovacuum worker"}},
		{},
	}
	index := 0
	err := waitForNoTemplateConnections(context.Background(), owned, func(context.Context) ([]templateConnection, error) {
		observation := observations[index]
		index++
		return observation, nil
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if index != len(observations) {
		t.Fatalf("observed %d connection states, want %d", index, len(observations))
	}
}

func TestWaitForNoTemplateConnectionsRejectsUnownedConnectionImmediately(t *testing.T) {
	t.Parallel()
	owned := &ownedConnectionPIDs{}
	owned.Add(101)
	calls := 0
	err := waitForNoTemplateConnections(context.Background(), owned, func(context.Context) ([]templateConnection, error) {
		calls++
		return []templateConnection{
			{pid: 101, backendType: "client backend"},
			{pid: 202, backendType: "client backend"},
		}, nil
	}, 0)
	if err == nil || !strings.Contains(err.Error(), "1 unowned open connection") {
		t.Fatalf("unexpected unowned connection result: %v", err)
	}
	if calls != 1 {
		t.Fatalf("unowned connection was retried %d times", calls)
	}
}

func TestWaitForNoTemplateConnectionsHonorsDrainDeadline(t *testing.T) {
	t.Parallel()
	owned := &ownedConnectionPIDs{}
	owned.Add(101)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := waitForNoTemplateConnections(ctx, owned, func(context.Context) ([]templateConnection, error) {
		return []templateConnection{{pid: 101, backendType: "client backend"}}, nil
	}, 0)
	if err == nil || !strings.Contains(err.Error(), "owned connections did not drain") {
		t.Fatalf("unexpected drain deadline result: %v", err)
	}
}
