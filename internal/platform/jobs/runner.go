package jobs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrRunnerClosed = errors.New("jobs: runner closed")
var ErrHandlerNotRegistered = errors.New("jobs: handler not registered")
var ErrHandlerAlreadyRegistered = errors.New("jobs: handler already registered")
var ErrDequeueGateClosed = errors.New("jobs: dequeue gate closed")

type HandlerFunc func(context.Context, uuid.UUID) error

type DequeueGate interface {
	AdmissionOpen() bool
}

type Runner struct {
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.Mutex
	closed          bool
	manager         *Manager
	workerID        string
	leaseDuration   time.Duration
	recoverLimit    int
	handlers        map[string]HandlerFunc
	gate            DequeueGate
	pendingRecovery map[string]struct{}
}

func NewRunner() *Runner {
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		ctx:             ctx,
		cancel:          cancel,
		workerID:        uuid.NewString(),
		leaseDuration:   30 * time.Second,
		recoverLimit:    100,
		handlers:        map[string]HandlerFunc{},
		pendingRecovery: map[string]struct{}{},
	}
}

func (r *Runner) ConfigureDequeueGate(gate DequeueGate) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gate = gate
}

func (r *Runner) Configure(manager *Manager) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.manager = manager
}

func (r *Runner) RegisterHandler(name string, handler HandlerFunc) error {
	if r == nil || name == "" || handler == nil {
		return ErrInvalidJobDefinition
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.handlers[name]; exists {
		return ErrHandlerAlreadyRegistered
	}
	r.handlers[name] = handler
	return nil
}

func (r *Runner) Dispatch(work func(context.Context)) error {
	if r == nil || work == nil {
		return nil
	}
	if err := r.addWork(); err != nil {
		return err
	}
	go func() {
		defer r.wg.Done()
		work(r.ctx)
	}()
	return nil
}

func (r *Runner) DispatchJob(handlerName string, jobID string) error {
	parsed, err := uuid.Parse(jobID)
	if err != nil {
		return err
	}
	return r.DispatchJobID(handlerName, parsed)
}

func (r *Runner) DispatchJobID(handlerName string, jobID uuid.UUID) error {
	handler, manager, err := r.namedWork(handlerName)
	if err != nil {
		return err
	}
	if err := r.addWork(); err != nil {
		return err
	}
	go func() {
		defer r.wg.Done()
		r.runNamedJob(handlerName, jobID, manager, handler)
	}()
	return nil
}

func (r *Runner) RecoverHandler(ctx context.Context, handlerName string) error {
	_, manager, err := r.namedWork(handlerName)
	if err != nil {
		return err
	}
	r.mu.Lock()
	if r.gate != nil && !r.gate.AdmissionOpen() {
		r.pendingRecovery[handlerName] = struct{}{}
		r.mu.Unlock()
		return nil
	}
	r.mu.Unlock()
	return r.recoverHandler(ctx, handlerName, manager)
}

func (r *Runner) Activate(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ErrRunnerClosed
	}
	if r.gate != nil && !r.gate.AdmissionOpen() {
		r.mu.Unlock()
		return ErrDequeueGateClosed
	}
	handlerNames := make([]string, 0, len(r.pendingRecovery))
	for handlerName := range r.pendingRecovery {
		handlerNames = append(handlerNames, handlerName)
	}
	r.mu.Unlock()
	sort.Strings(handlerNames)
	for _, handlerName := range handlerNames {
		_, manager, err := r.namedWork(handlerName)
		if err != nil {
			return err
		}
		if err := r.recoverHandler(ctx, handlerName, manager); err != nil {
			return err
		}
		r.mu.Lock()
		delete(r.pendingRecovery, handlerName)
		r.mu.Unlock()
	}
	return nil
}

func (r *Runner) recoverHandler(ctx context.Context, handlerName string, manager *Manager) error {
	jobIDs, err := manager.RecoverableHandlerJobs(ctx, handlerName, r.recoverLimit)
	if err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		if err := r.DispatchJobID(handlerName, jobID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Runner) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		r.cancel()
	}
	r.mu.Unlock()

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *Runner) addWork() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRunnerClosed
	}
	if r.gate != nil && !r.gate.AdmissionOpen() {
		return ErrDequeueGateClosed
	}
	r.wg.Add(1)
	return nil
}

func (r *Runner) namedWork(handlerName string) (HandlerFunc, *Manager, error) {
	if r == nil || handlerName == "" {
		return nil, nil, ErrInvalidJobDefinition
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	handler := r.handlers[handlerName]
	if handler == nil {
		return nil, nil, ErrHandlerNotRegistered
	}
	if r.manager == nil {
		return nil, nil, ErrNotConfigured
	}
	return handler, r.manager, nil
}

func (r *Runner) runNamedJob(handlerName string, jobID uuid.UUID, manager *Manager, handler HandlerFunc) {
	claimed, err := manager.ClaimHandlerJob(r.ctx, jobID, handlerName, r.workerID, r.leaseDuration)
	if err != nil || !claimed {
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = manager.RecordHandlerError(context.Background(), jobID, r.workerID, fmt.Errorf("job handler panic: %v", recovered))
		}
	}()
	err = handler(r.ctx, jobID)
	if err == nil {
		_ = manager.ReleaseHandlerLease(context.Background(), jobID, r.workerID)
		return
	}
	if r.ctx.Err() != nil && errors.Is(err, context.Canceled) {
		_ = manager.ReleaseHandlerLease(context.Background(), jobID, r.workerID)
		return
	}
	_ = manager.RecordHandlerError(context.Background(), jobID, r.workerID, err)
}
