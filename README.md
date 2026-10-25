# RCJV Soccer Stats

The RCJV Soccer Stats project is an open source web application designed
to show the upcoming matches and current standings from the [Catigoal](https://catigoal.com/)
application. 
There is a total of three different sites available:

```
Sites/
 ├─ https://example.com/#/referees               (Referee overview)
 ├─ https://example.com/#/matches?leagues=       (Match overview)
 └─ https://example.com/#/standings              (Standings overview)
```

The matches site have a league argument, where you can specify the 
leagues with the league abbreviation (e.g. `et,lwl,open`).

For the standings and matches site you can also specify the cycle time.
The parameter is called `time` and is specified in seconds (default: 20).

Matches and Standings sites cycle through the different leagues and stages
every 20 seconds, that's also the interval new data gets fetched from the backend, which fetches the data from Catigoal every 10 seconds.

## Game Master screens (v2)

The v2 screens implement the public API in
[`docs/rcgamemaster_restapi_v1_openapi.yaml`](docs/rcgamemaster_restapi_v1_openapi.yaml).
Point the app at your Game Master REST base URL and set its tournament abbreviation:

```bash
TOURNAMENT_API_URL=https://gamemaster.example/rest/v1/ \
TOURNAMENT_ABBREVIATION=RCJ2027 \
go run ./backend/main.go
```

`TOURNAMENT_API_URL` applies to both API versions and defaults to
`https://catigoal.com/rest/v1/`. It is the REST base, without the tournament
abbreviation. No API key or login is required. For Docker, pass both settings
with `-e` before the image name.

Open either screen and choose a main league and an optional sub-league:

```text
http://localhost:3030/v2/matches
http://localhost:3030/v2/matches?main_league=soccer
http://localhost:3030/v2/matches?main_league=soccer&sub_league=ir
http://localhost:3030/v2/standings?main_league=soccer
http://localhost:3030/v2/standings?main_league=soccer&sub_league=ir&league_stage=0
```

These links redirect to the existing hash router (`/#/v2/matches` and
`/#/v2/standings`), preserving the filters. The home page also links to both views.
Selections are saved in the URL and retained when switching views.

- No main league selected shows every main league.
- No sub-league selected shows every child in the selected scope, plus the main
  league's own matches and standings when it has direct stages. Standings are
  fetched separately for each exact league and displayed as separate tables.
  Empty sub-leagues remain available in the selector.
- Select one league to enable the stage selector. `league_stage` is the zero-based
  index in that league's complete stage list, including unpublished stages.
  With no stage selected, matches include every stage and standings show the
  cumulative Seeding table. The exact `league` alias is also supported, including
  `league=soccer` for only the root's own stages.
- The screens refresh every 20 seconds and provide a manual refresh button.
  Only approved results display scores and points. Schedule and publication
  times stay in the API's displayed tournament timezone.

The backend exposes `/api/v2/leagues`, `/api/v2/matches`, and
`/api/v2/standings` as read-only proxies of the configured tournament. They retain
the API's default JSON envelopes, hierarchy, match status, and upstream error
responses. The original screen routes and `/api/v1` endpoints remain available.

## Development
To run the project locally, you need to have [Node.js](https://nodejs.org/) and [Golang](https://go.dev) installed.
Then, follow these steps:
```bash
# First build the frontend
cd frontend
npm install
npm run build

# Now start the backend
go run ./backend/main.go (TOURNAMENT_ABBREVIATION)

# You can now access all the sites, e.g. http://localhost:3030/#/standings
# If you change something in the frontend code, you need to rebuild it
cd frontend
npm run build
```

If you have any questions, feel free to open an issue or contact me directly.

## Deployment
You can either run it directly or use the docker setup provided below.
```bash
docker run -d -p 3030:3030 ghcr.io/eliasb-nu/rcjvsoccerstats:latest --env TOURNAMENT_ABBREVIATION=(LeagueAbbreviation)
```
