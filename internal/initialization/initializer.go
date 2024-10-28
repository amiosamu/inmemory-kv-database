package initialization

import (
	"context"
	"errors"
	"fmt"

	configuration "github.com/amiosamu/inmemory-kv-database/internal/config"
	"github.com/amiosamu/inmemory-kv-database/internal/database"
	"github.com/amiosamu/inmemory-kv-database/internal/database/compute"
	"github.com/amiosamu/inmemory-kv-database/internal/database/storage"
	replication "github.com/amiosamu/inmemory-kv-database/internal/database/storage/replicaiton"
	"github.com/amiosamu/inmemory-kv-database/internal/database/storage/wal"
	"github.com/amiosamu/inmemory-kv-database/internal/network"
	"github.com/amiosamu/inmemory-kv-database/internal/storage/engine/in_memory"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Initializer struct {
	wal    *wal.WAL
	engine *in_memory.Engine
	server *network.TCPServer
	slave  *replication.Slave
	master *replication.Master
	logger *zap.Logger
}

func NewInitialzer(cfg *configuration.Config) (*Initializer, error) {
	if cfg == nil {
		return nil, errors.New("failed to init: config is empty or invalid")
	}
	logger, err := CreateLogger(cfg.Logging)
	if err != nil {
		return nil, fmt.Errorf("failed to init logger: %w", err)
	}
	wal, err := CreateWAL(cfg.WAL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init WAL: %w", err)

	}

	engine, err := CreateEngine(cfg.Engine, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init Engine: %w", err)
	}

	server, err := CreateNetwork(cfg.Network, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init network: %w", err)
	}

	replica, err := CreateReplica(cfg.Replication, cfg.WAL, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to init replication: %w", err)
	}

	init := &Initializer{
		wal:    wal,
		engine: engine,
		server: server,
		logger: logger,
	}

	switch v := replica.(type) {
	case *replication.Slave:
		init.slave = v

	case *replication.Master:
		init.master = v
	}

	return init, nil
}

func (i *Initializer) StartDatabase(ctx context.Context) error {
	compute, err := compute.NewCompute(i.logger)

	if err != nil {
		return err
	}

	var options []storage.StorageOption
	if i.wal != nil {
		options = append(options, storage.WithWAL(i.wal))
	}

	if i.master != nil {
		options = append(options, storage.WithWAL(i.wal))
	} else if i.slave != nil {
		options = append(options, storage.WithReplication(i.slave))
		options = append(options, storage.WithReplicationStream(i.slave.ReplicationStream()))
	}

	storage, err := storage.NewStorage(i.engine, i.logger, options...)
	if err != nil {
		return err
	}

	database, err := database.NewDatabase(compute, storage, i.logger)

	if err != nil {
		return err
	}

	group, groupCtx := errgroup.WithContext(ctx)
	if i.wal != nil {
		if i.slave != nil {
			group.Go(func() error {
				i.slave.Start(groupCtx)
				return nil
			})
		} else {
			group.Go(func() error {
				i.wal.Start(groupCtx)
				return nil
			})
		}
		if i.master != nil {
			group.Go(func() error {
				i.master.Start(groupCtx)
				return nil
			})
		}
	}
	group.Go(func() error {
		i.server.HandleQueries(groupCtx, func(ctx context.Context, query []byte) []byte {
			response := database.HandleQuery(ctx, string(query))
			return []byte(response)
		})
		return nil
	})
	return group.Wait()
}
