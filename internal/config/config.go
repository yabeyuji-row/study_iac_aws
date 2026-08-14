package config

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
}

func Load() (cfg Config, err error) {
	cfg.HTTPAddr = getenv("HTTP_ADDR", ":8080")
	cfg.DatabaseURL, err = loadDatabaseURL()
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

type databaseSecret struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DBName   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"sslmode"`
}

func loadDatabaseURL() (databaseURL string, err error) {
	databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		secretJSON := os.Getenv("DATABASE_SECRET_JSON")
		if secretJSON == "" {
			return "", fmt.Errorf("DATABASE_URL or DATABASE_SECRET_JSON is required")
		}

		databaseURL, err = databaseURLFromSecret(secretJSON)
		if err != nil {
			return "", fmt.Errorf("build database url from secret: %w", err)
		}
	}

	return databaseURL, nil
}

func databaseURLFromSecret(secretJSON string) (databaseURL string, err error) {
	var secret databaseSecret
	if err := json.Unmarshal([]byte(secretJSON), &secret); err != nil {
		return "", fmt.Errorf("parse database secret json: %w", err)
	}

	if secret.Host == "" {
		return "", fmt.Errorf("host is required")
	}
	if secret.Port == 0 {
		return "", fmt.Errorf("port is required")
	}
	if secret.DBName == "" {
		return "", fmt.Errorf("dbname is required")
	}
	if secret.Username == "" {
		return "", fmt.Errorf("username is required")
	}
	if secret.Password == "" {
		return "", fmt.Errorf("password is required")
	}
	if secret.SSLMode == "" {
		secret.SSLMode = "require"
	}

	values := url.Values{}
	values.Set("sslmode", secret.SSLMode)

	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(secret.Username, secret.Password),
		Host:     net.JoinHostPort(secret.Host, strconv.Itoa(secret.Port)),
		Path:     "/" + secret.DBName,
		RawQuery: values.Encode(),
	}

	return dbURL.String(), nil
}

func getenv(key string, fallback string) (value string) {
	value = os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
