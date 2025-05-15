package main

import (
	"log"
	"time"

	"github.com/adityadeshlahre/logs-directory-api/api"
	"github.com/adityadeshlahre/logs-directory-api/cache"
	"github.com/adityadeshlahre/logs-directory-api/db"
	"github.com/adityadeshlahre/logs-directory-api/generator"
	"github.com/adityadeshlahre/logs-directory-api/models"
	"github.com/gin-gonic/gin"
)

func main() {
	redisCache, redisErr := cache.NewRedisCache("localhost:6379")
	if redisErr != nil {
		log.Fatalf("Failed to connect to Redis Instance: %v", redisErr)
	}

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

	api.Routes(r, redisCache, mongoStore)

	err := r.Run(":8080")
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
