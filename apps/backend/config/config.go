// Package config provides functionality to load server configuration
package config

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/oauth2"
)

// CallbackPath defines the expected path for the OAuth2 redirect/callback endpoint.
const (
	CallbackPath = "/api/callback"
)

// Cache is the structure used to store and retrieve the serialized OAuth2 token
// within the MongoDB collection.
type Cache struct {
	Data []byte `bson:"data"` // GOB-encoded byte slice of the oauth2.Token.
}

// ServerConfig holds all necessary configuration settings for the server,
// primarily sourced from environment variables.
type ServerConfig struct {
	ClientID string
	BaseURL  string
	Port     int

	ConnectionString string
}

// LoadConfig reads and validates all required environment variables,
// returning a populated ServerConfig struct.
func LoadConfig() (*ServerConfig, error) {

	// --- MongoDB Configuration ---

	connectionString := os.Getenv("MONGO_CONNECTION_STRING")

	// Check for required MongoDB environment variables.
	if len(strings.TrimSpace(connectionString)) == 0 {
		logrus.Fatalf("FATAL: Required environment variables MONGO_CONNECTION_STRING must be set.")
	}

	// --- Application/OAuth Configuration ---

	clientID := os.Getenv("CLIENT_ID")
	baseURL := os.Getenv("BASE_URL")
	port := os.Getenv("PORT")

	if len(strings.TrimSpace(clientID)) == 0 || len(strings.TrimSpace(baseURL)) == 0 || len(strings.TrimSpace(port)) == 0 {
		return nil, fmt.Errorf("CLIENT_ID, BASE_URL, and PORT environment variables must be set")
	}

	iPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, fmt.Errorf("invalid port value '%s': %w", port, err)
	}

	// Create and populate the configuration struct.
	config := ServerConfig{
		ClientID: clientID,
		BaseURL:  baseURL,
		Port:     iPort,

		ConnectionString: connectionString,
	}

	return &config, nil
}

// SaveCredentials serializes an oauth2.Token using GOB and persists it
// to the specified MongoDB collection. It handles upsert logic (insert or update).
func SaveCredentials(collection *mongo.Collection, token *oauth2.Token) error {
	// Use a context with cancellation for the database operation.
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	// Serialize the oauth2.Token struct into a byte buffer using GOB encoding.
	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(token); err != nil {
		return fmt.Errorf("failed to encode OAuth2 token: %w", err)
	}

	var cache Cache
	// $set operator updates the 'data' field with the new serialized token.
	update := bson.M{
		"$set": bson.M{"data": buffer.Bytes()},
	}
	// Use FindOneAndUpdate options to ensure we get the updated document (if found).
	// SetUpsert(false) is used here because we manage the insert manually below if ErrNoDocuments occurs.
	opts := options.FindOneAndUpdate().SetUpsert(false).SetReturnDocument(options.After)

	// Attempt to update an existing document.
	err := collection.FindOneAndUpdate(ctx, bson.M{}, update, opts).Decode(&cache)

	if err == mongo.ErrNoDocuments {
		// If no document was found, insert the new token data.
		cache.Data = buffer.Bytes()
		_, insertErr := collection.InsertOne(ctx, cache)
		if insertErr != nil {
			logrus.WithError(insertErr).Error("failed to insert initial token cache into collection")
			return insertErr
		}
		return nil
	} else if err != nil {
		// Handle other MongoDB errors.
		logrus.WithError(err).Error("failed during FindOneAndUpdate to save token")
		return err
	}

	return nil // Update succeeded
}

// LoadCredentials retrieves the serialized token data from MongoDB,
// deserializes it using GOB, and returns the oauth2.Token.
func LoadCredentials(collection *mongo.Collection) *oauth2.Token {
	// Use a context with cancellation for the database operation.
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	var cache Cache

	// Find the stored credentials document (using an empty filter since only one is stored).
	err := collection.FindOne(ctx, bson.D{}).Decode(&cache)

	if err == mongo.ErrNoDocuments {
		// Credentials not found, return nil.
		return nil
	} else if err != nil {
		// Log and fail on unexpected database error.
		logrus.WithError(err).Fatalf("FATAL: failed during FindOne to load token")
	}

	var token oauth2.Token
	// Deserialize the GOB data back into an oauth2.Token.
	decoder := gob.NewDecoder(bytes.NewReader(cache.Data))
	if err = decoder.Decode(&token); err != nil {
		// Log and fail if deserialization fails (data corruption).
		logrus.WithError(err).Fatalf("FATAL: failed to decode token from GOB data")
	}

	return &token
}
