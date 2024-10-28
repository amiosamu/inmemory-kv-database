package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/amiosamu/inmemory-kv-database/internal/database/compute"
	"go.uber.org/zap"
)

type computerLayer interface {
	Parse(string) (compute.Query, error)
}

type storageLayer interface {
	Set(context.Context, string, string) error
	Get(context.Context, string) (string, error)
	Del(context.Context, string) error
}

type Database struct {
	computeLayer computerLayer
	storageLayer storageLayer
	logger       *zap.Logger
}

func NewDatabase(computerLayer computerLayer, storageLayer storageLayer, logger *zap.Logger) (*Database, error) {
	if computerLayer == nil {
		return nil, errors.New("compute is invalid")
	}
	if storageLayer == nil {
		return nil, errors.New("storage is invalid")
	}
	if logger == nil {
		return nil, errors.New("logger is invalid")
	}
	return &Database{
		computeLayer: computerLayer,
		storageLayer: storageLayer,
		logger:       logger,
	}, nil
}

func (d *Database) HandleQuery(ctx context.Context, queryStr string) string {
	d.logger.Debug("handling query", zap.String("query", queryStr))
	query, err := d.computeLayer.Parse(queryStr)
	if err != nil {
		return fmt.Sprintf("[error] %s", err.Error())
	}
	switch query.CommandID() {
	case compute.SetCommandID:
		return d.handleSetQuery(ctx, query)
	case compute.GetCommandID:
		return d.handleGetQuery(ctx, query)
	case compute.DelCommandID:
		return d.handleDelQuery(ctx, query)
	}

	d.logger.Error(
		"computer layer is incorrect",
		zap.Int("command_id", query.CommandID()),
	)
	return "[error] internal error"
}

func (d *Database) handleSetQuery(ctx context.Context, query compute.Query) string {
	panic("implement me")
}

func (d *Database) handleGetQuery(ctx context.Context, query compute.Query) string {
	panic("implement me")
}

func (d *Database) handleDelQuery(ctx context.Context, query compute.Query) string {
	panic("implement me")
}
