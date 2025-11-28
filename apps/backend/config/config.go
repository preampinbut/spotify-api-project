// Package config provides server configuration loading and credential caching.
package config

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/oauth2"
)

const (
	CallbackPath = "/api/callback"
	mongoTimeout = 5 * time.Second
)

// Cache stores the serialized OAuth2 token in MongoDB.
type Cache struct {
	Data []byte `bson:"data"`
}

// ServerConfig holds server settings from environment variables.
type ServerConfig struct {
	ClientID         string
	BaseURL          string
	Port             int
	ConnectionString string
}

// LoadConfig reads and validates required environment variables.
func LoadConfig() (*ServerConfig, error) {
	connectionString := os.Getenv("MONGO_CONNECTION_STRING")
	if strings.TrimSpace(connectionString) == "" {
		logrus.Fatal("FATAL: MONGO_CONNECTION_STRING must be set")
	}

	clientID := os.Getenv("CLIENT_ID")
	baseURL := os.Getenv("BASE_URL")
	portStr := os.Getenv("PORT")
	if strings.TrimSpace(clientID) == "" || strings.TrimSpace(baseURL) == "" || strings.TrimSpace(portStr) == "" {
		return nil, fmt.Errorf("CLIENT_ID, BASE_URL, and PORT environment variables must be set")
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid port value '%s': %w", portStr, err)
	}

	return &ServerConfig{
		ClientID:         clientID,
		BaseURL:          baseURL,
		Port:             port,
		ConnectionString: connectionString,
	}, nil
}

// SaveCredentials serializes and upserts an oauth2.Token in MongoDB.
func SaveCredentials(collection *mongo.Collection, token *oauth2.Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(token); err != nil {
		return fmt.Errorf("failed to encode OAuth2 token: %w", err)
	}

	update := bson.M{"$set": bson.M{"data": buf.Bytes()}}
	opts := options.FindOneAndUpdate().SetUpsert(false).SetReturnDocument(options.After)

	var cache Cache
	err := collection.FindOneAndUpdate(ctx, bson.M{}, update, opts).Decode(&cache)
	if err == mongo.ErrNoDocuments {
		cache.Data = buf.Bytes()
		if _, insertErr := collection.InsertOne(ctx, cache); insertErr != nil {
			logrus.WithError(insertErr).Error("failed to insert token cache")
			return insertErr
		}
		return nil
	} else if err != nil {
		logrus.WithError(err).Error("failed to update token cache")
		return err
	}
	return nil
}

// LoadCredentials retrieves and deserializes the OAuth2 token from MongoDB.
func LoadCredentials(collection *mongo.Collection) *oauth2.Token {
	ctx, cancel := context.WithTimeout(context.Background(), mongoTimeout)
	defer cancel()

	var cache Cache
	err := collection.FindOne(ctx, bson.D{}).Decode(&cache)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		logrus.WithError(err).Fatal("FATAL: failed to load token")
	}

	var token oauth2.Token
	if err := gob.NewDecoder(bytes.NewReader(cache.Data)).Decode(&token); err != nil {
		logrus.WithError(err).Fatal("FATAL: failed to decode token")
	}
	return &token
}
