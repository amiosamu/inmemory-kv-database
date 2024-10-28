package wal

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
)

type segmentsDirectory interface {
	ForEach(func([]byte) error) error
}

type LogsReader struct {
	segmentsDirectory segmentsDirectory
}

func NewLogsReader(segmentsDirectory segmentsDirectory) (*LogsReader, error) {
	if segmentsDirectory == nil {
		return nil, errors.New("segments directory is invalid")
	}

	return &LogsReader{
		segmentsDirectory: segmentsDirectory,
	}, nil
}

func (r *LogsReader) Read() ([]Log, error) {
	var logs []Log

	err := r.segmentsDirectory.ForEach(func(b []byte) error {
		var err error
		logs, err = r.readSegment(logs, b)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to read segments: %w", err)
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].LSN < logs[j].LSN
	})
	return logs, nil
}

func (r *LogsReader) readSegment(logs []Log, data []byte) ([]Log, error) {
	buffer := bytes.NewBuffer(data)
	for buffer.Len() > 0 {
		var log Log
		if err := log.Decode(buffer); err != nil {
			return nil, fmt.Errorf("failed to parse logs data: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, nil
}
