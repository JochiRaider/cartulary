package actorfixture

import (
	"sync"
	"testing"
)

func TestConcurrentActorPreparationSharesOneImmutableInput(t *testing.T) {
	const readers = 32
	values := make([]string, readers)
	errors := make([]error, readers)
	var workers sync.WaitGroup
	for index := range readers {
		workers.Go(func() { values[index], errors[index] = Credential() })
	}
	workers.Wait()
	for index := range readers {
		if errors[index] != nil || values[index] == "" || values[index] != values[0] {
			t.Fatal("concurrent preparation did not return one valid immutable input")
		}
	}
}
