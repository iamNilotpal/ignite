package engine

import (
	"github.com/iamNilotpal/ignite/internal/compaction"
	"github.com/iamNilotpal/ignite/internal/index"
	"github.com/iamNilotpal/ignite/internal/storage"
	"github.com/iamNilotpal/ignite/pkg/options"
	"go.uber.org/zap"
)

type Engine struct {
	options *options.Options
	log     *zap.SugaredLogger

	index      *index.Index
	storage    *storage.Storage
	compaction *compaction.Compaction
}

type Config struct {
	Options *options.Options
	Logger  *zap.SugaredLogger
}

func New(config *Config) *Engine {
	index := index.New()
	storage := storage.New()
	compaction := compaction.New()

	return &Engine{
		options:    config.Options,
		log:        config.Logger,
		index:      index,
		storage:    storage,
		compaction: compaction,
	}
}
