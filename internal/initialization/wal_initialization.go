package initialization

import (
	"errors"
	"time"

	"github.com/amiosamu/inmemory-kv-database/internal/common"
	configuration "github.com/amiosamu/inmemory-kv-database/internal/config"
	"github.com/amiosamu/inmemory-kv-database/internal/database/filesystem"
	"github.com/amiosamu/inmemory-kv-database/internal/database/storage/wal"
	"go.uber.org/zap"
)

const (
	defaultFlushingBatchSize    = 100
	defaultFlushingBatchTimeout = time.Millisecond * 10
	defaultMaxSegmentSize       = 10 << 20
	defaultWALDataDirectory     = "./data/spider/wal"
)

func CreateWAL(cfg *configuration.WALConfig, logger *zap.Logger) (*wal.WAL, error) {
	if logger == nil {
		return nil, errors.New("logger is invalid")
	} else if cfg == nil {
		return nil, nil
	}

	flushingBatchSize := defaultFlushingBatchSize
	flushingBatchTimeout := defaultFlushingBatchTimeout
	maxSegmentSize := defaultMaxSegmentSize
	dataDirectory := defaultWALDataDirectory

	if cfg.FlushingBatchLength != 0 {
		flushingBatchSize = cfg.FlushingBatchLength
	}

	if cfg.FlushingBatchTimeout != 0 {
		flushingBatchTimeout = cfg.FlushingBatchTimeout
	}

	if cfg.MaxSegmentSize != "" {
		size, err := common.ParseSize(cfg.MaxSegmentSize)
		if err != nil {
			return nil, errors.New("max segment size is incorrect")
		}

		maxSegmentSize = size
	}

	if cfg.DataDirectory != "" {
		// TODO: need to create a directory,
		// if it is missing
		dataDirectory = cfg.DataDirectory
	}

	segmentsDirectory := filesystem.NewSegmentsDirectory(dataDirectory)
	reader, err := wal.NewLogsReader(segmentsDirectory)
	if err != nil {
		return nil, err
	}

	segment := filesystem.NewSegment(dataDirectory, maxSegmentSize)
	writer, err := wal.NewLogsWriter(segment, logger)
	if err != nil {
		return nil, err
	}

	return wal.NewWAL(writer, reader, flushingBatchTimeout, flushingBatchSize)
}
