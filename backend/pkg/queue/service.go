package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/registryx/registryx/backend/pkg/config"
)

const ScanQueueKey = "registryx:scan_queue"

type Job struct {
	ManifestID uuid.UUID `json:"manifest_id"`
	Repository string    `json:"repository"`
	Reference  string    `json:"reference"`
}

type Service struct {
	Client *redis.Client
}

func NewService(cfg *config.Config) (*Service, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddr,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Service{Client: rdb}, nil
}

func (s *Service) EnqueueScan(ctx context.Context, manifestID uuid.UUID, repoName, reference string) error {
	job := Job{ManifestID: manifestID, Repository: repoName, Reference: reference}
	bytes, _ := json.Marshal(job)
	
	return s.Client.RPush(ctx, ScanQueueKey, bytes).Err()
}

func (s *Service) DequeueScan(ctx context.Context) (*Job, error) {
	// Block for 0 seconds (infinite) until a job is available
	result, err := s.Client.BLPop(ctx, 0*time.Second, ScanQueueKey).Result()
	if err != nil {
		return nil, err
	}

	// result[0] is the key, result[1] is the value
	var job Job
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}

	return &job, nil
}
