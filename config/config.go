package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	TelegramToken string
	GrokAPIKey    string
	SystemPrompt  string
	AllowedUsers  []int64
	ErrorLogUser  int64
}

func LoadConfig() (Config, error) {
	var allowedUsers []int64
	if users := os.Getenv("ALLOWED_USERS"); users != "" {
		for _, u := range strings.Split(users, ",") {
			id, err := strconv.ParseInt(u, 10, 64)
			if err != nil {
				return Config{}, fmt.Errorf("invalid ALLOWED_USERS: %w", err)
			}
			allowedUsers = append(allowedUsers, id)
		}
	}

	errorLogUser, err := strconv.ParseInt(os.Getenv("ERROR_LOG_USER"), 10, 64)
	if err != nil && os.Getenv("ERROR_LOG_USER") != "" {
		return Config{}, fmt.Errorf("invalid ERROR_LOG_USER: %w", err)
	}

	return Config{
		TelegramToken: os.Getenv("TELEGRAM_TOKEN"),
		GrokAPIKey:    os.Getenv("GROK_API_TOKEN"),
		SystemPrompt:  os.Getenv("SYS"),
		AllowedUsers:  allowedUsers,
		ErrorLogUser:  errorLogUser,
	}, nil
}
