package web

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTournamentProxy(t *testing.T) {
	for _, tt := range []struct {
		name       string
		path       string
		upstream   string
		body       string
		status     int
		wantStatus int
	}{
		{"catalog including empty children", "/api/v2/leagues", "/rest/v1/RCJ2027/leagues", `{"main_leagues":[{"abbrev":"soccer","league_stages":[],"sub_leagues":[{"abbrev":"vision","league_stages":[]}]}]}`, 200, 200},
		{"all matches", "/api/v2/matches", "/rest/v1/RCJ2027/matches", `{"matches":[]}`, 200, 200},
		{"main aggregate", "/api/v2/matches?main_league=soccer", "/rest/v1/RCJ2027/matches?main_league=soccer", `{"matches":[]}`, 200, 200},
		{"child and stage zero", "/api/v2/matches?main_league=soccer&sub_league=ir&league_stage=0", "/rest/v1/RCJ2027/matches?league_stage=0&main_league=soccer&sub_league=ir", `{"matches":[{"status":"submitted","is_placeholder":false,"is_bye":false,"team2":null,"goals1":0}]}`, 200, 200},
		{"exact root seeding", "/api/v2/standings?league=soccer", "/rest/v1/RCJ2027/standings?league=soccer", `{"league_stage_name":"Seeding","standings":[]}`, 200, 200},
		{"empty selectors omitted", "/api/v2/standings?main_league=soccer&sub_league=ir&league_stage=", "/rest/v1/RCJ2027/standings?main_league=soccer&sub_league=ir", `{"standings":[]}`, 200, 200},
		{"selector encoding", "/api/v2/matches?sub_league=ir%20%26%20vision", "/rest/v1/RCJ2027/matches?sub_league=ir+%26+vision", `{"matches":[]}`, 200, 200},
		{"upstream selection error", "/api/v2/standings?main_league=soccer", "/rest/v1/RCJ2027/standings?main_league=soccer", `{"error":"standings require an exact league or sub_league"}`, 404, 404},
		{"object format only", "/api/v2/leagues?format=openapi&url=https://elsewhere.invalid", "/rest/v1/RCJ2027/leagues", `{"main_leagues":[]}`, 200, 200},
		{"upstream outage", "/api/v2/matches", "/rest/v1/RCJ2027/matches", `{"error":"temporarily unavailable"}`, 503, 503},
		{"invalid upstream JSON", "/api/v2/leagues", "/rest/v1/RCJ2027/leagues", `<html>Unavailable</html>`, 502, 502},
	} {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.RequestURI != tt.upstream {
					t.Errorf("upstream request = %s %s, want GET %s", r.Method, r.RequestURI, tt.upstream)
				}
				if r.Header.Get("Accept") != "application/json" {
					t.Error("upstream request must ask for JSON")
				}
				w.WriteHeader(tt.status)
				fmt.Fprint(w, tt.body)
			}))
			defer server.Close()
			app := newApp(&API{SourceURL: server.URL + "/rest/v1/RCJ2027/"}, t.TempDir())
			t.Cleanup(func() { _ = app.Shutdown() })
			response, err := app.Test(httptest.NewRequest(http.MethodGet, tt.path, nil))
			if err != nil {
				t.Fatal(err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if response.StatusCode != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.StatusCode, tt.wantStatus, body)
			}
			if strings.HasPrefix(tt.body, "{") {
				if string(body) != tt.body {
					t.Errorf("response lost upstream fields: got %s, want %s", body, tt.body)
				}
				if response.Header.Get("Cache-Control") != "no-cache" {
					t.Error("tournament responses must not be cached")
				}
			} else if !strings.Contains(string(body), "invalid JSON") {
				t.Errorf("expected a JSON error, got %s", body)
			}
		})
	}
}

func TestTournamentProxyUnavailable(t *testing.T) {
	app := newApp(&API{}, t.TempDir())
	t.Cleanup(func() { _ = app.Shutdown() })
	for _, tt := range []struct {
		path   string
		status int
	}{
		{"/api/v2/leagues", 503},
		{"/api/v2/private", 404},
	} {
		response, err := app.Test(httptest.NewRequest(http.MethodGet, tt.path, nil))
		if err != nil {
			t.Fatal(err)
		}
		response.Body.Close()
		if response.StatusCode != tt.status {
			t.Errorf("%s: status = %d, want %d", tt.path, response.StatusCode, tt.status)
		}
	}
}

func TestV2PageLinks(t *testing.T) {
	app := newApp(&API{}, t.TempDir())
	t.Cleanup(func() { _ = app.Shutdown() })
	for _, path := range []string{"/v2/matches", "/v2/standings"} {
		for _, query := range []string{"", "?main_league=soccer&sub_league=ir&league_stage=0"} {
			response, err := app.Test(httptest.NewRequest(http.MethodGet, path+query, nil))
			if err != nil {
				t.Fatal(err)
			}
			response.Body.Close()
			if response.StatusCode != http.StatusFound || response.Header.Get("Location") != "/#"+path+query {
				t.Errorf("%s: status = %d, location = %q", path+query, response.StatusCode, response.Header.Get("Location"))
			}
		}
	}
}
