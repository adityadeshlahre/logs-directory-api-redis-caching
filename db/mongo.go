package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/adityadeshlahre/logs-directory-api/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	Client     *mongo.Client
	Collection *mongo.Collection
}

func NewMongoStore(uri, dbName, collectionName string) (*MongoStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	collection := client.Database(dbName).Collection(collectionName)
	return &MongoStore{Client: client, Collection: collection}, nil
}

func (m *MongoStore) SaveLog(log models.LogEntry) error {
	_, err := m.Collection.InsertOne(context.Background(), log)
	return err
}

func (m *MongoStore) GetLogsByUser(userID string, limit int64) ([]models.LogEntry, error) {
	filter := map[string]any{"userid": userID}
	cur, err := m.Collection.Find(context.Background(), filter, options.Find().SetLimit(limit))
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	var results []models.LogEntry
	for cur.Next(context.Background()) {
		var log models.LogEntry
		err := cur.Decode(&log)
		if err != nil {
			continue
		}
		results = append(results, log)
	}
	return results, nil
}

func (m *MongoStore) SearchLogs(userID string, query string) ([]models.LogEntry, error) {
	fmt.Println("Searching logs for user:", userID, "with query:", query)
	filter := map[string]any{
		"userid": userID,
		"level": map[string]any{
			"$regex":   query,
			"$options": "i",
		},
	}
	cur, err := m.Collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	fmt.Println("cur:", cur)
	defer cur.Close(context.Background())

	var logs []models.LogEntry
	for cur.Next(context.Background()) {
		var logEntry models.LogEntry
		if err := cur.Decode(&logEntry); err != nil {
			continue
		}
		logs = append(logs, logEntry)
	}
	if err := cur.Err(); err != nil {
		log.Printf("Error with cursor: %v", err)
		return nil, err
	}
	return logs, nil
}

func (m *MongoStore) GetLogByID(userID, logID string) (*models.LogEntry, error) {
	filter := map[string]any{
		"userid": userID,
		"logid":  logID,
	}
	var logEntry models.LogEntry
	err := m.Collection.FindOne(context.Background(), filter).Decode(&logEntry)
	if err != nil {
		return nil, err
	}
	return &logEntry, nil
}
