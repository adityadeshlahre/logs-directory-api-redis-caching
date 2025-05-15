package api

import (
	"log"
	"net/http"

	"github.com/adityadeshlahre/logs-directory-api/cache"
	"github.com/adityadeshlahre/logs-directory-api/db"
	"github.com/adityadeshlahre/logs-directory-api/utils"
	"github.com/gin-gonic/gin"
)

func Routes(r *gin.Engine, redisCache *cache.RedisCache, mongoStore *db.MongoStore) {
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

		logs, err := redisCache.SearchLogs(userId, query)
		if err == nil && len(logs) > 0 {
			log.Println("[CACHE HIT] Search logs from Redis")
			c.JSON(http.StatusOK, gin.H{
				"logs": logs,
			})
			return
		}

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

		logEntry, err := redisCache.GetLogByID(userId, logId)
		if err == nil && logEntry != nil {
			log.Println("[CACHE HIT] Log fetched from Redis")
			c.JSON(http.StatusOK, gin.H{
				"log": logEntry,
			})
			return
		}

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
}
