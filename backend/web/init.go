package web

import (
	"RCJV-SoccerStats/backend/config"
	"RCJV-SoccerStats/backend/data"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
	"github.com/gofiber/fiber/v3/middleware/static"
)

type API struct {
	leaguesMu sync.RWMutex
	Leagues   *data.League
	Referees  *[]data.RefereeConverted
	SourceURL string
}

func Init(cfg *config.Config, url string) {
	var (
		addr = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		a    = API{SourceURL: url}
	)

	// Get Data
	ticker := time.NewTicker(10 * time.Second)
	a.Leagues = data.GetLeagues(url)
	if a.Leagues == nil {
		a.Leagues = &data.League{}
	}
	a.Referees = data.GetReferees(url)
	go func() {
		for {
			<-ticker.C
			leagues := data.GetLeagues(url)
			if leagues != nil {
				leagues = data.GetStandings(url, leagues)
				leagues = data.GetMatches(url, leagues)

				a.leaguesMu.Lock()
				a.Leagues = leagues
				a.leaguesMu.Unlock()
			}

			//for _, l := range leagues.Leagues {
			//	log.Printf("Updated standings for league: %s\n", l.LeagueName)
			//	log.Println(l.LatestStandings.Standings)
			//}
			//for _, l := range leagues.Leagues {
			//	for _, stage := range l.Stages {
			//		log.Printf("Updated matches for league: %s, stage: %s\n", l.LeagueName, stage.Name)
			//		log.Println(stage.Matches)
			//	}
			//}

			a.Referees = data.GetReferees(url)
			// log.Printf("Updated referees: %v\n", referees)
		}
	}()

	// Start Server
	err := newApp(&a, "./frontend/dist").Listen(addr)
	if err != nil {
		log.Fatalf("Error starting webserver: %v\n", err)
	}
}

// newApp configures routes independently of data polling and the network listener.
func newApp(a *API, frontendDir string) *fiber.App {
	app := fiber.New(fiber.Config{
		ServerHeader: "rcjv_soccerStats:fiber",
		AppName:      "rcjv_soccerStats",
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		AllowMethods:     []string{fiber.MethodGet, fiber.MethodPost, fiber.MethodDelete},
		AllowCredentials: false,
	}))
	app.Get("/livez", healthcheck.New())
	app.Get("/readyz", healthcheck.New())
	app.Get("/healthcheck", getHealthCheck)

	// API Endpoints
	apiV1 := fiber.New()
	app.Use("/api/v1", apiV1)
	apiV1.Get("/leagues", a.getLeagues)
	apiV1.Get("/standings/:league", a.getStandings)
	apiV1.Get("/matches/:league", a.getMatches)
	apiV1.Get("/referees", a.getReferees)
	apiV1.Get("/referees/:referee/matches", a.getRefereeMatches)
	app.Get("/api/v2/:resource", a.getTournamentData)

	// Keep public v2 links short while retaining the existing hash router.
	for _, view := range []string{"matches", "standings"} {
		app.Get("/v2/"+view, func(c fiber.Ctx) error {
			target := "/#/v2/" + view
			if query := string(c.Request().URI().QueryString()); query != "" {
				target += "?" + query
			}
			return c.Redirect().Status(fiber.StatusFound).To(target)
		})
	}

	// Website
	app.Use("/", static.New(frontendDir))

	return app
}
