package web

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

var tournamentClient = &http.Client{Timeout: 10 * time.Second}

// Forward the public object envelopes without losing hierarchy, stages or status.
// Only the configured tournament and the three read-only resources are exposed.
func (a *API) getTournamentData(c fiber.Ctx) error {
	resource := c.Params("resource")
	if resource != "leagues" && resource != "matches" && resource != "standings" {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Unknown tournament resource"})
	}
	if a.SourceURL == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Tournament API is not configured"})
	}
	query := url.Values{}
	for _, key := range []string{"main_league", "sub_league", "league", "league_stage"} {
		if value := c.Query(key); value != "" {
			query.Set(key, value)
		}
	}
	endpoint := strings.TrimRight(a.SourceURL, "/") + "/" + resource
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(c.Context(), http.MethodGet, endpoint, nil)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Invalid tournament API URL"})
	}
	req.Header.Set("Accept", "application/json")
	response, err := tournamentClient.Do(req)
	if err != nil {
		log.Printf("Error fetching tournament %s: %v", resource, err)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Could not reach the tournament API"})
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil || !json.Valid(body) {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "Tournament API returned an invalid JSON response"})
	}
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	c.Set(fiber.HeaderCacheControl, "no-cache")
	return c.Status(response.StatusCode).Send(body)
}
