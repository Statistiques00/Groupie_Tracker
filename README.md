# Groupie Tracker
Minimal Go web app that consumes the Groupie Tracker API and renders artists, dates, locations, and relations.

## Features
- Server-rendered pages with a JS-driven UI
- Unified artist search with `source=groupie|spotify|all`
- Timeline, locations grid, and relations accordion views
- Optional Spotify enrichment (search + artist detail)
- JSON API endpoints for frontend fetch calls
- Go standard library only (no external deps)

## Quickstart
Prereqs: Go 1.20+

PowerShell:
```powershell
go run .
```

bash:
```bash
go run .
```

Open: http://localhost:8080

## Configuration
Environment variables:

| Name | Description | Default |
| --- | --- | --- |
| `SPOTIFY_CLIENT_ID` | Spotify Client ID (optional) | (none) |
| `SPOTIFY_CLIENT_SECRET` | Spotify Client Secret (optional) | (none) |

Flags:

| Flag | Description | Default |
| --- | --- | --- |
| `-addr` | HTTP address to listen on | `:8080` |
| `-api` | Upstream Groupie Tracker API base URL | `https://groupietrackers.herokuapp.com/api` |
| `-static` | Static assets directory | `static` |
| `-templates` | HTML templates glob | `templates/*.html` |
| `-spotify-client-id` | Spotify Client ID | from env |
| `-spotify-client-secret` | Spotify Client Secret | from env |

## API
- `GET /api/artists` (filters: `name`, `year`, `member`, `source=groupie|spotify|all`, `external=spotify`, `limit`)
- `GET /api/artists/{id}`
- `GET /api/locations`
- `GET /api/dates`
- `GET /api/relation`
- `GET /api/events`
- `GET /api/spotify/artist?id=...`

## Project structure
```
.
├─ api_client.go
├─ cache.go
├─ data.go
├─ handlers.go
├─ server.go
├─ spotify_client.go
├─ templates/
├─ static/
└─ docs/
```

## Troubleshooting
- Spotify endpoints return 503: set `SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET`.
- Port already in use: change `-addr` (e.g. `-addr :8081`).
- Upstream API down: `/api/*` may return 503 or empty data.
- Spotify search with empty `name`: no Spotify results are returned by design.
- Wrong static/template paths: verify `-static` and `-templates`.
