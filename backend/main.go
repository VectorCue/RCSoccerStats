package main

import (
	"RCJV-SoccerStats/backend/config"
	"RCJV-SoccerStats/backend/web"
	"fmt"
	"log"
	"net/url"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting server...")

	cfg := config.GetConfig()
	sourceURL := fmt.Sprintf("%s%s/", cfg.URL, url.PathEscape(cfg.TournamentAbbreviation))

	// Start Webserver
	web.Init(cfg, sourceURL)
}
