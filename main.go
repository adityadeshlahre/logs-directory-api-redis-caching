package main

import (
	"log"
	"net/http"
	"time"

	"github.com/adityadeshlahre/logs-directory-api/cache"
	"github.com/adityadeshlahre/logs-directory-api/db"
	"github.com/adityadeshlahre/logs-directory-api/generator"
	"github.com/adityadeshlahre/logs-directory-api/models"
	"github.com/adityadeshlahre/logs-directory-api/utils"
	"github.com/gin-gonic/gin"
)

func main() {
	redisCache := cache.NewRedisCache("localhost:6379")

	mongoStore, mongoErr := db.NewMongoStore(
		"mongodb://go:og@localhost:27017/",
		"logsdb",
		"logs",
	)
	if mongoErr != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", mongoErr)
	}

	logsChannel := make(chan models.LogEntry)
	go generator.StartLogGenerator(logsChannel, 10*time.Second)
	go func() {
		for logEntry := range logsChannel {
			err := mongoStore.SaveLog(logEntry)
			if err != nil {
				log.Printf("Error saving log to MongoDB: %v", err)
			}
		}
	}()

	r := gin.Default()

	r.GET("/:userId/logs", func(c *gin.Context) {
		page := c.DefaultQuery("page", "1")
		limit := c.DefaultQuery("limit", "5")

		pagination, err := utils.GetPagination(page, limit)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid pagination parameters"})
			return
		}

		offset := pagination.Skip
		userId := c.Param("userId")
		// 1. Try Redis
		logs, totalLogs, err := redisCache.GetLogs(userId, offset, pagination.Limit)
		if err == nil && len(logs) > 0 {
			log.Println("[CACHE HIT] Logs fetched from Redis")
			c.JSON(http.StatusOK, gin.H{
				"logs":     logs,
				"total":    totalLogs,
				"page":     page,
				"limit":    pagination.Limit,
				"nextPage": offset+pagination.Limit < totalLogs,
			})
			return
		}

		// 2. Cache miss: Fetch from Mongo
		log.Println("[CACHE MISS] Fetching logs from MongoDB")
		mongoLogs, err := mongoStore.GetLogsByUser(userId, int64(pagination.Limit))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch from MongoDB"})
			return
		}

		for _, logEntry := range mongoLogs {
			_ = redisCache.AddLog(logEntry)
		}

		c.JSON(http.StatusOK, gin.H{
			"logs":     mongoLogs,
			"total":    len(mongoLogs),
			"page":     page,
			"limit":    pagination.Limit,
			"nextPage": false,
		})
	})

	r.GET("/:userId/logs/search", func(c *gin.Context) {
		query := c.DefaultQuery("q", "")
		userId := c.Param("userId")
		// 1. Try Redis cache first
		logs, err := redisCache.SearchLogs(userId, query)
		if err == nil && len(logs) > 0 {
			log.Println("[CACHE HIT] Search logs from Redis")
			c.JSON(http.StatusOK, gin.H{
				"logs": logs,
			})
			return
		}

		// 2. Cache miss, try MongoDB
		log.Println("[CACHE MISS] Search logs from MongoDB")
		mongoLogs, err := mongoStore.SearchLogs(userId, query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "MongoDB search failed"})
			return
		}

		for _, logEntry := range mongoLogs {
			_ = redisCache.AddLog(logEntry)
		}

		c.JSON(http.StatusOK, gin.H{
			"logs": mongoLogs,
		})
	})

	r.GET("/:userId/:logId", func(c *gin.Context) {
		logId := c.Param("logId")

		userId := c.Param("userId")
		// 1. Try Redis cache first
		logEntry, err := redisCache.GetLogByID(userId, logId)
		if err == nil && logEntry != nil {
			log.Println("[CACHE HIT] Log fetched from Redis")
			c.JSON(http.StatusOK, gin.H{
				"log": logEntry,
			})
			return
		}
		// 2. Cache miss, try MongoDB
		log.Println("[CACHE MISS] Fetching log from MongoDB")
		mongoLog, err := mongoStore.GetLogByID(userId, logId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch from MongoDB"})
			return
		}
		if mongoLog == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Log not found"})
			return
		}

		_ = redisCache.AddLog(*mongoLog)

		c.JSON(http.StatusOK, gin.H{
			"log": mongoLog,
		})
	})

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
