package config

import (
	"log"
	"os"
	"strings"
)

type Config struct {
	URL                    string
	TournamentAbbreviation string
	Host                   string
	Port                   int
}

func GetConfig() *Config {
	var cfg Config
	cfg.Port = 3030
	cfg.Host = "0.0.0.0"
	cfg.URL = "https://catigoal.com/rest/v1/"
	if baseURL := os.Getenv("TOURNAMENT_API_URL"); baseURL != "" {
		cfg.URL = strings.TrimRight(baseURL, "/") + "/"
	}
	// If command line arguments are present, use these, else use environment variables
	if len(os.Args) >= 2 {
		cfg.TournamentAbbreviation = os.Args[1]
	} else if len(os.Getenv("TOURNAMENT_ABBREVIATION")) > 0 {
		cfg.TournamentAbbreviation = os.Getenv("TOURNAMENT_ABBREVIATION")
	} else {
		log.Fatalf("\n	TOURNAMENT_ABBREVIATION not set!\n	Please set it either via an environment variable or via command line argument.")
	}

	return &cfg
}
