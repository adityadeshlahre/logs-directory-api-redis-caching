package cache

import (
	"context"
	"strings"
	"time"

	"github.com/adityadeshlahre/logs-directory-api/models"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisAddr string) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr: redisAddr, // e.g., "localhost:6379"
	})

	return &RedisCache{client: client}
}

func (c *RedisCache) AddLog(log models.LogEntry) error {
	logKey := "logs:" + log.UserID

	logStr := log.Timestamp.Format(time.RFC3339) + " " + log.Level + " " + log.Component + " " + log.Message

	_, err := c.client.LPush(ctx, logKey, logStr).Result()
	if err != nil {
		return err
	}

	c.client.Expire(ctx, logKey, time.Hour)

	c.client.LTrim(ctx, logKey, 0, 999)

	return nil
}

func (c *RedisCache) GetLogs(userID string, offset, limit int) ([]models.LogEntry, int, error) {
	logKey := "logs:" + userID

	logs, err := c.client.LRange(ctx, logKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, 0, err
	}

	var result []models.LogEntry
	for _, logStr := range logs {
		result = append(result, models.LogEntry{
			Timestamp: time.Now(),
			Level:     "INFO",
			Component: "example",
			Message:   logStr,
			UserID:    userID,
		})
	}

	totalLogs, err := c.client.LLen(ctx, logKey).Result()
	if err != nil {
		return nil, 0, err
	}

	return result, int(totalLogs), nil
}

func (c *RedisCache) SearchLogs(userID, query string) ([]models.LogEntry, error) {
	logKey := "logs:" + userID

	logs, err := c.client.LRange(ctx, logKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var result []models.LogEntry
	for _, logStr := range logs {
		if strings.Contains(strings.ToLower(logStr), strings.ToLower(query)) {
			result = append(result, models.LogEntry{
				Timestamp: time.Now(),
				Level:     "INFO",
				Component: "example",
				Message:   logStr,
				UserID:    userID,
			})
		}
	}

	return result, nil
}
