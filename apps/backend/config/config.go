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

const (
	CallbackPath = "/api/callback"
)

type Cache struct {
	Data []byte `json:"data"`
}

type ServerConfig struct {
	ClientId string
	BaseURL  string
	Port     int

	ConnectionString string
	Collection       string
}

func LoadConfig() (*ServerConfig, error) {

	connectionString := os.Getenv("MONGO_CONNECTION_STRING")
	connectionStringT := len(strings.TrimSpace(connectionString)) == 0
	collection := os.Getenv("MONGO_COLLECTION")
	collectionT := len(strings.TrimSpace(collection)) == 0

	if connectionStringT || collectionT {
		logrus.Fatalf("DB_CONNECTION_STRING | DB_COLLECTION did not set")
	}

	clientId := os.Getenv("CLIENT_ID")
	clientIdT := len(strings.TrimSpace(clientId)) == 0
	baseURL := os.Getenv("BASE_URL")
	baseURLT := len(strings.TrimSpace(baseURL)) == 0
	port := os.Getenv("PORT")
	portT := len(strings.TrimSpace(port)) == 0

	if clientIdT || baseURLT || portT {
		return nil, fmt.Errorf("CLIENT_ID | BASE_URL | PORT did not set")
	}

	iPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	config := ServerConfig{
		ClientId: clientId,
		BaseURL:  baseURL,
		Port:     iPort,

		ConnectionString: connectionString,
		Collection:       collection,
	}

	return &config, nil
}

func SaveCredentials(collection *mongo.Collection, token *oauth2.Token) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(token); err != nil {
		return err
	}

	var cache Cache
	update := bson.M{
		"$set": bson.M{"data": buffer.Bytes()},
	}
	opts := options.FindOneAndUpdate().SetUpsert(false).SetReturnDocument(options.After)

	err := collection.FindOneAndUpdate(ctx, bson.M{}, update, opts).Decode(&cache)
	if err == mongo.ErrNoDocuments {
		cache.Data = buffer.Bytes()
		_, insertErr := collection.InsertOne(ctx, cache)
		if insertErr != nil {
			logrus.WithError(insertErr).Error("failed to insert cache into collection")
			return insertErr
		}
		return nil
	} else if err != nil {
		logrus.WithError(err).Error("failed during FindOneAndUpdate")
		return err
	}

	return nil
}

func LoadCredentials(collection *mongo.Collection) *oauth2.Token {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	var cache Cache

	err := collection.FindOne(ctx, bson.D{}).Decode(&cache)
	if err == mongo.ErrNoDocuments {
		return nil
	} else if err != nil {
		logrus.WithError(err).Fatalf("failed during FindOne")
	}

	var token oauth2.Token
	decoder := gob.NewDecoder(bytes.NewReader(cache.Data))
	if err = decoder.Decode(&token); err != nil {
		logrus.WithError(err).Fatalf("failed to decode token")
	}

	return &token
}
