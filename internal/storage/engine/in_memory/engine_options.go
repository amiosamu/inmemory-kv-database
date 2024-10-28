package in_memory

type EngineOption func(*Engine)

func WithPartitions(partitionsNum uint) EngineOption {
	return func(e *Engine) {
		e.partitions = make([]*HashTable, partitionsNum)
		for i := 0; i < int(partitionsNum); i++ {
			e.partitions[i] = NewHashTable()
		}
	}
}
