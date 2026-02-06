package logic

import (
	"errors"
	"hash/fnv"
	"os"
	"sync"
	"time"
)

const (
	snowflakeNodeBits  = 10
	snowflakeStepBits  = 12
	snowflakeNodeMax   = -1 ^ (-1 << snowflakeNodeBits)
	snowflakeStepMask  = -1 ^ (-1 << snowflakeStepBits)
	snowflakeTimeShift = snowflakeNodeBits + snowflakeStepBits
	snowflakeNodeShift = snowflakeStepBits
	snowflakeEpochMs   = int64(1704067200000) // 2024-01-01T00:00:00Z
)

type snowflake struct {
	mu            sync.Mutex
	lastTimestamp int64
	sequence      int64
	node          int64
}

func newSnowflake(node int64) (*snowflake, error) {
	if node < 0 || node > snowflakeNodeMax {
		return nil, errors.New("snowflake node id out of range")
	}
	return &snowflake{node: node}, nil
}

func (s *snowflake) nextID() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	ts := time.Now().UnixMilli()
	if ts < s.lastTimestamp {
		ts = s.waitUntil(s.lastTimestamp)
	}

	if ts == s.lastTimestamp {
		s.sequence = (s.sequence + 1) & snowflakeStepMask
		if s.sequence == 0 {
			ts = s.waitUntil(s.lastTimestamp + 1)
		}
	} else {
		s.sequence = 0
	}

	s.lastTimestamp = ts

	return ((ts - snowflakeEpochMs) << snowflakeTimeShift) |
		(s.node << snowflakeNodeShift) |
		s.sequence
}

func (s *snowflake) waitUntil(target int64) int64 {
	ts := time.Now().UnixMilli()
	for ts < target {
		time.Sleep(time.Millisecond)
		ts = time.Now().UnixMilli()
	}
	return ts
}

var (
	idGenerator *snowflake
	idOnce      sync.Once
)

func generateUserID() int64 {
	idOnce.Do(func() {
		node := machineNodeID()
		gen, err := newSnowflake(node)
		if err != nil {
			gen, _ = newSnowflake(0)
		}
		idGenerator = gen
	})

	if idGenerator == nil {
		return time.Now().UnixNano()
	}
	return idGenerator.nextID()
}

func machineNodeID() int64 {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return 0
	}
	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(host))
	return int64(hasher.Sum32() % uint32(snowflakeNodeMax+1))
}
