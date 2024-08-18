package config

import (
	"encoding/json"
	"os"

	"golang.org/x/oauth2"
	"gopkg.in/yaml.v3"
)

const (
	RedirectURI     = "http://localhost:8888/api/callback"
	ConfigPath      = "config.yml"
	CredentialsPath = "credentials.json"
)

type ConfigType struct {
	ClientId string `yaml:"client_id"`
}

type Credentials struct {
	Data []byte `json:"data"`
}

func LoadConfig() (*ConfigType, error) {
	data, err := os.ReadFile(ConfigPath)

	if err != nil {
		return nil, err
	}

	var config ConfigType
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func SaveCredentials(token *oauth2.Token) error {
	f, err := os.Create(CredentialsPath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	tokenData, err := json.Marshal(token)
	if err != nil {
		return err
	}

	cred := Credentials{
		Data: tokenData,
	}

	return json.NewEncoder(f).Encode(cred)
}

func LoadCredentials() (*oauth2.Token, error) {
	f, err := os.Open(CredentialsPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var creds Credentials
	if err := json.NewDecoder(f).Decode(&creds); err != nil {
		return nil, err
	}

	var token oauth2.Token
	if err := json.Unmarshal(creds.Data, &token); err != nil {
		return nil, err
	}

	return &token, nil
}
