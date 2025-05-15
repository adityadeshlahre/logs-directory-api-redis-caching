package models

import "time"

type LogEntry struct {
	LogID     string    `json:"logid"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
	UserID    string    `json:"userid"`
}
