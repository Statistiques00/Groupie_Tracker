package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultAddr      = ":8080"
	defaultStaticDir = "static"
	defaultTplGlob   = "templates/*.html"
)

// App bundles the HTTP handlers, template set and data cache.
type App struct {
	cache     *Cache
	api       *APIClient
	spotify   *SpotifyClient
	templates *template.Template
	staticDir string
}

func newApp(apiBase, staticDir, tplGlob, spotifyID, spotifySecret string) (*App, error) {
	tpls, err := template.ParseGlob(tplGlob)
	if err != nil {
		return nil, fmt.Errorf("parse templates: %w", err)
	}
	var spotifyClient *SpotifyClient
	if spotifyID != "" && spotifySecret != "" {
		spotifyClient = newSpotifyClient(spotifyID, spotifySecret)
		log.Printf("spotify client enabled")
	}
	return &App{
		cache:     newCache(),
		api:       newAPIClient(apiBase),
		spotify:   spotifyClient,
		templates: tpls,
		staticDir: staticDir,
	}, nil
}

func (a *App) refreshData(ctx context.Context) error {
	data, err := a.api.FetchAll(ctx)
	if err != nil {
		return err
	}
	a.cache.Set(data)
	return nil
}

func (a *App) routes() http.Handler {
	mux := http.NewServeMux()

	// Static assets
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir(filepath.Join(a.staticDir, "css")))))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir(filepath.Join(a.staticDir, "js")))))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(a.staticDir))))

	// Favicon: serve favicon.ico from project root if present, otherwise fall back to common static locations
	mux.HandleFunc("/favicon.ico", a.handleFavicon)

	// API endpoints
	mux.HandleFunc("/api/artists", a.handleAPIArtists)
	mux.HandleFunc("/api/artists/", a.handleAPIArtistByID)
	mux.HandleFunc("/api/locations", a.handleAPILocations)
	mux.HandleFunc("/api/dates", a.handleAPIDates)
	mux.HandleFunc("/api/relation", a.handleAPIRelation)
	mux.HandleFunc("/api/events", a.handleAPIEvents)
	mux.HandleFunc("/api/spotify/artist", a.handleAPISpotifyArtist)
	mux.HandleFunc("/healthz", a.handleHealth)

	// HTML pages
	mux.HandleFunc("/artist", a.handleArtistPage)
	mux.HandleFunc("/artist.html", a.handleArtistPage)
	mux.HandleFunc("/artist-spotify", a.handleSpotifyArtistPage)
	mux.HandleFunc("/artist-spotify.html", a.handleSpotifyArtistPage)
	mux.HandleFunc("/dates", a.handleDatesPage)
	mux.HandleFunc("/dates.html", a.handleDatesPage)
	mux.HandleFunc("/locations", a.handleLocationsPage)
	mux.HandleFunc("/locations.html", a.handleLocationsPage)
	mux.HandleFunc("/relations", a.handleRelationsPage)
	mux.HandleFunc("/relations.html", a.handleRelationsPage)
	mux.HandleFunc("/404", a.handle404Page)
	mux.HandleFunc("/404.html", a.handle404Page)
	mux.HandleFunc("/500", a.handle500Page)
	mux.HandleFunc("/500.html", a.handle500Page)
	mux.HandleFunc("/index.html", a.handleRoot)
	mux.HandleFunc("/", a.handleRoot)

	return recoverMiddleware(loggingMiddleware(mux))
}

func main() {
	// allow overriding defaults via environment variables
	addrDefault := defaultAddr
	if v := os.Getenv("ADDR"); v != "" {
		addrDefault = v
	}
	apiDefault := defaultAPIBase
	if v := os.Getenv("API"); v != "" {
		apiDefault = v
	}
	staticDefault := defaultStaticDir
	if v := os.Getenv("STATIC"); v != "" {
		staticDefault = v
	}
	tplDefault := defaultTplGlob
	if v := os.Getenv("TEMPLATES"); v != "" {
		tplDefault = v
	}

	addr := flag.String("addr", addrDefault, "HTTP address to listen on (e.g. :8080)")
	apiBase := flag.String("api", apiDefault, "Upstream Groupie Tracker API base URL")
	staticDir := flag.String("static", staticDefault, "Directory that hosts static assets")
	tplGlob := flag.String("templates", tplDefault, "Glob pattern for HTML templates")
	spotifyID := flag.String("spotify-client-id", os.Getenv("SPOTIFY_CLIENT_ID"), "Spotify Client ID (defaults to SPOTIFY_CLIENT_ID env)")
	spotifySecret := flag.String("spotify-client-secret", os.Getenv("SPOTIFY_CLIENT_SECRET"), "Spotify Client Secret (defaults to SPOTIFY_CLIENT_SECRET env)")
	flag.Parse()

	app, err := newApp(*apiBase, *staticDir, *tplGlob, *spotifyID, *spotifySecret)
	if err != nil {
		log.Fatalf("initialise app: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := app.refreshData(ctx); err != nil {
		log.Printf("warning: failed to prefetch data: %v", err)
	}

	srv := &http.Server{
		Addr:              *addr,
		Handler:           app.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Groupie Tracker backend running at http://localhost%s", *addr)
	if err := srv.ListenAndServe(); err != nil && !strings.Contains(err.Error(), "Server closed") {
		log.Fatalf("server error: %v", err)
	}
}

// handleFavicon tries to serve a favicon from several locations.
// Priority:
// 1. ./favicon.ico (project root)
// 2. static/image/grouper_tracke.ico
// 3. static/image/favicon.ico
// If none exist it returns 404.
func (a *App) handleFavicon(w http.ResponseWriter, r *http.Request) {
	// candidate paths
	candidates := []string{
		"favicon.ico",
		filepath.Join(a.staticDir, "image", "grouper_tracke.ico"),
		filepath.Join(a.staticDir, "image", "favicon.ico"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			http.ServeFile(w, r, p)
			return
		}
	}
	http.NotFound(w, r)
}
