package utils

import (
	"sync"
	"time"
)

type Snowflake struct {
	mu        sync.Mutex
	startTime int64
	workerID  int64
	sequence  int64
}

const (
	workerBits     = 10
	sequenceBits   = 12
	workerMax      = -1 ^ (-1 << workerBits)
	sequenceMask   = -1 ^ (-1 << sequenceBits)
	workerShift    = sequenceBits
	timestampShift = sequenceBits + workerBits
)

func NewSnowflake(workerID int64) *Snowflake {
	if workerID < 0 || workerID > workerMax {
		panic("worker ID out of range")
	}
	return &Snowflake{
		startTime: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
		workerID:  workerID,
	}
}

func (s *Snowflake) NextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < s.startTime {
		panic("clock moved backwards")
	}

	if now == s.startTime {
		s.sequence = (s.sequence + 1) & sequenceMask
		if s.sequence == 0 {
			for now <= s.startTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.startTime = now
	return (now << timestampShift) | (s.workerID << workerShift) | s.sequence
}
