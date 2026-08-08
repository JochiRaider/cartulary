package postgres

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestTransactionRunnerLifecycle(t *testing.T) {
	ctx := context.Background()
	operationErr := errors.New("operation failed")
	beginErr := errors.New("begin failed")
	commitErr := errors.New("commit failed")

	tests := []struct {
		name      string
		db        DB
		operation func(pgx.Tx) error
		wantErr   error
		wantCalls []string
	}{
		{
			name:      "unavailable database",
			db:        nil,
			operation: func(pgx.Tx) error { return nil },
			wantCalls: nil,
		},
		{
			name:      "begin failure",
			db:        &transactionLifecycleDB{beginErr: beginErr},
			operation: func(pgx.Tx) error { return nil },
			wantErr:   beginErr,
			wantCalls: []string{"begin"},
		},
		{
			name: "operation failure rolls back",
			db:   &transactionLifecycleDB{tx: &transactionLifecycleTx{}},
			operation: func(tx pgx.Tx) error {
				tx.(*transactionLifecycleTx).calls = append(tx.(*transactionLifecycleTx).calls, "operation")
				return operationErr
			},
			wantErr:   operationErr,
			wantCalls: []string{"begin", "operation", "rollback"},
		},
		{
			name: "success commits then performs defensive rollback",
			db:   &transactionLifecycleDB{tx: &transactionLifecycleTx{}},
			operation: func(tx pgx.Tx) error {
				tx.(*transactionLifecycleTx).calls = append(tx.(*transactionLifecycleTx).calls, "operation")
				return nil
			},
			wantCalls: []string{"begin", "operation", "commit", "rollback"},
		},
		{
			name: "commit failure is returned",
			db:   &transactionLifecycleDB{tx: &transactionLifecycleTx{commitErr: commitErr}},
			operation: func(tx pgx.Tx) error {
				tx.(*transactionLifecycleTx).calls = append(tx.(*transactionLifecycleTx).calls, "operation")
				return nil
			},
			wantErr:   commitErr,
			wantCalls: []string{"begin", "operation", "commit", "rollback"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runner := NewTransactionRunner(test.db)
			err := runner.WithinTx(ctx, pgx.TxOptions{}, test.operation)
			if test.db == nil {
				if err == nil {
					t.Fatal("expected unavailable-runner error")
				}
				return
			}
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("WithinTx error = %v, want %v", err, test.wantErr)
			}
			db := test.db.(*transactionLifecycleDB)
			gotCalls := append([]string{}, db.calls...)
			if db.tx != nil {
				gotCalls = append(gotCalls, db.tx.(*transactionLifecycleTx).calls...)
			}
			if !reflect.DeepEqual(gotCalls, test.wantCalls) {
				t.Fatalf("lifecycle calls = %#v, want %#v", gotCalls, test.wantCalls)
			}
		})
	}
}

type transactionLifecycleDB struct {
	DB
	tx       pgx.Tx
	beginErr error
	calls    []string
}

func (db *transactionLifecycleDB) BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error) {
	db.calls = append(db.calls, "begin")
	return db.tx, db.beginErr
}

type transactionLifecycleTx struct {
	pgx.Tx
	calls     []string
	commitErr error
}

func (tx *transactionLifecycleTx) Commit(context.Context) error {
	tx.calls = append(tx.calls, "commit")
	return tx.commitErr
}

func (tx *transactionLifecycleTx) Rollback(context.Context) error {
	tx.calls = append(tx.calls, "rollback")
	return nil
}
