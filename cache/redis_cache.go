package cache

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/adityadeshlahre/logs-directory-api/models"
	"github.com/go-redis/redis/v8"
)

var ctx = context.Background()

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(redisAddr string) (*RedisCache, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) AddLog(log models.LogEntry) error {
	logKey := log.UserID

	logBytes, err := json.Marshal(log)
	if err != nil {
		return err
	}
	logStr := string(logBytes)

	_, err = c.client.LPush(ctx, logKey, logStr).Result()
	if err != nil {
		return err
	}

	c.client.Expire(ctx, logKey, time.Minute)

	c.client.LTrim(ctx, logKey, 0, 99)

	return nil
}

func (c *RedisCache) GetLogs(userID string, offset, limit int) ([]models.LogEntry, int, error) {
	logKey := userID

	logs, err := c.client.LRange(ctx, logKey, int64(offset), int64(offset+limit-1)).Result()
	if err != nil {
		return nil, 0, err
	}

	var result []models.LogEntry
	for _, logStr := range logs {
		var logEntry models.LogEntry
		err := json.Unmarshal([]byte(logStr), &logEntry)
		if err != nil {
			continue
		}
		result = append(result, logEntry)
	}

	totalLogs, err := c.client.LLen(ctx, logKey).Result()
	if err != nil {
		return nil, 0, err
	}

	return result, int(totalLogs), nil
}

func (c *RedisCache) SearchLogs(userID, query string) ([]models.LogEntry, error) {
	logKey := userID

	logs, err := c.client.LRange(ctx, logKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	var result []models.LogEntry
	for _, logStr := range logs {
		if strings.Contains(strings.ToLower(logStr), strings.ToLower(query)) {
			var logEntry models.LogEntry
			err := json.Unmarshal([]byte(logStr), &logEntry)
			if err != nil {
				continue
			}
			result = append(result, logEntry)
		}
	}

	return result, nil
}

func (c *RedisCache) GetLogByID(userID, logId string) (*models.LogEntry, error) {
	logKey := userID

	logs, err := c.client.LRange(ctx, logKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	for _, logStr := range logs {
		var logEntry models.LogEntry
		err := json.Unmarshal([]byte(logStr), &logEntry)
		if err != nil {
			continue
		}
		if logEntry.LogID == logId {
			return &logEntry, nil
		}
	}

	return nil, nil
}
