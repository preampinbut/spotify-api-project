package config

import (
	"backend/db"
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	CallbackPath = "/api/callback"
)

type ConfigType struct {
	ClientId string `yaml:"client_id"`
	BaseURL  string `yaml:"base_url"`
	Port     int    `yaml:"port"`
}

func LoadConfig() (*ConfigType, error) {
	clientId := os.Getenv("CLIENT_ID")
	baseURL := os.Getenv("BASE_URL")
	port := os.Getenv("PORT")

	if len(strings.TrimSpace(clientId)) == 0 || len(strings.TrimSpace(baseURL)) == 0 || len(strings.TrimSpace(port)) == 0 {
		return nil, errors.New(fmt.Sprintf("required environment CLIENT_ID, BASE_URL, PORT"))
	}

	iPort, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	config := ConfigType{
		ClientId: clientId,
		BaseURL:  baseURL,
		Port:     iPort,
	}

	return &config, nil
}

func SaveCredentials(dbClient *db.PrismaClient, token *oauth2.Token) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	if err := encoder.Encode(token); err != nil {
		return err
	}

	config, _ := dbClient.Config.FindFirst().Exec(ctx)
	if config == nil {
		dbClient.Config.CreateOne(
			db.Config.Data.Set(buffer.Bytes()),
		).Exec(ctx)
		return nil
	}

	dbClient.Config.FindUnique(
		db.Config.ID.Equals(config.ID),
	).Update(
		db.Config.Data.Set(buffer.Bytes()),
	).Exec(ctx)
	return nil
}

func LoadCredentials(dbClient *db.PrismaClient) *oauth2.Token {
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	config, err := dbClient.Config.FindFirst().Exec(ctx)
	if err != nil {
		return nil
	}

	var token oauth2.Token
	decoder := gob.NewDecoder(bytes.NewReader(config.Data))
	if err = decoder.Decode(&token); err != nil {
		logrus.WithError(err).Fatalf("failed to decode token")
	}

	return &token
}
