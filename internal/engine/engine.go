package engine

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/iamNilotpal/ignite/internal/compaction"
	"github.com/iamNilotpal/ignite/internal/index"
	"github.com/iamNilotpal/ignite/internal/storage"
	"github.com/iamNilotpal/ignite/pkg/options"
	"go.uber.org/zap"
)

var (
	ErrEngineClosed = errors.New("operation failed: cannot access closed engine")
)

type Engine struct {
	options *options.Options
	log     *zap.SugaredLogger

	closed     atomic.Bool
	index      *index.Index
	storage    *storage.Storage
	compaction *compaction.Compaction
}

type Config struct {
	Options *options.Options
	Logger  *zap.SugaredLogger
}

func New(ctx context.Context, config *Config) (*Engine, error) {
	index := index.New()
	compaction := compaction.New()

	storage, err := storage.New(ctx, &storage.Config{
		Logger:  config.Logger,
		Options: config.Options,
	})
	if err != nil {
		return nil, err
	}

	return &Engine{
		options:    config.Options,
		log:        config.Logger,
		index:      index,
		storage:    storage,
		compaction: compaction,
	}, nil
}

func (e *Engine) Close() error {
	if !e.closed.CompareAndSwap(false, true) {
		return ErrEngineClosed
	}
	return e.storage.Close()
}
