package generator

import (
	"math/rand"
	"time"

	"github.com/adityadeshlahre/logs-directory-api/models"
)

var logLevels = []string{"INFO", "WARN", "ERROR"}
var components = []string{"auth-service", "payment-service", "user-service", "inventory-service"}
var messages = []string{
	"Token expired for user",
	"Payment failed due to timeout",
	"User not found in database",
	"Inventory updated successfully",
	"Session invalidated manually",
}

var userIDs = []string{"user-123", "user-456", "user-789"}

func StartLogGenerator(out chan<- models.LogEntry, interval time.Duration) {
	go func() {
		for {
			log := models.LogEntry{
				Timestamp: time.Now().UTC(),
				Level:     randChoice(logLevels),
				Component: randChoice(components),
				Message:   randChoice(messages),
				UserID:    randChoice(userIDs),
			}
			out <- log
			time.Sleep(interval)
		}
	}()
}

func randChoice(choices []string) string {
	return choices[rand.Intn(len(choices))]
}
