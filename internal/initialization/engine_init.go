package initialization

import (
	"errors"

	configuration "github.com/amiosamu/inmemory-kv-database/internal/config"
	"github.com/amiosamu/inmemory-kv-database/internal/storage/engine/in_memory"
	"go.uber.org/zap"
)

func CreateEngine(cfg *configuration.EngineConfig, logger *zap.Logger) (*in_memory.Engine, error) {
	if logger == nil {
		return nil, errors.New("logger is invalid")
	}
	if cfg == nil {
		return in_memory.NewEngine(logger)
	}
	if cfg.Type != "" {
		supportedTypes := map[string]struct{}{
			"in_memory": {},
		}
		if _, found := supportedTypes[cfg.Type]; !found {
			return nil, errors.New("engine type is incorrect")
		}
	}

	var options []in_memory.EngineOption
	if cfg.PartitionsNumber != 0 {
		options = append(options, in_memory.WithPartitions(cfg.PartitionsNumber))
	}
	return in_memory.NewEngine(logger, options...)
}
