package postgresstore

import (
	"context"
	"fmt"
	"time"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool        postgres.DB
	cursorCodec *graphCursorCodec
	now         func() time.Time
	hooks       Hooks
}

type Hooks struct {
	BeforePublication func(context.Context, graphprojection.ProjectionRun) error
}

type Options struct {
	DB        postgres.DB
	Now       func() time.Time
	CursorKey []byte
	Hooks     Hooks
}

func New(options Options) (*Store, error) {
	if options.DB == nil {
		return nil, fmt.Errorf("graph projection store database is required")
	}
	codec, err := newGraphCursorCodec(options.CursorKey)
	if err != nil {
		return nil, err
	}
	now := options.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Store{pool: options.DB, cursorCodec: codec, now: now, hooks: options.Hooks}, nil
}
