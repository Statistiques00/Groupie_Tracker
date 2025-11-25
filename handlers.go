package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type viewLocation struct {
	ArtistID   int    `json:"artistId"`
	ArtistName string `json:"artistName"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Raw        string `json:"raw"`
	EventCount int    `json:"eventCount"`
}

type spotifyArtistDetail struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	ImageURL   string   `json:"image_url"`
	Genres     []string `json:"genres"`
	Popularity int      `json:"popularity"`
	Followers  int      `json:"followers"`
	Source     string   `json:"source"`
}

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		a.renderError(w, http.StatusNotFound)
		return
	}
	a.ensureCache(r.Context())
	// Pass simple stats to the template in case they are used.
	snapshot := a.cache.Snapshot()
	data := map[string]interface{}{
		"ArtistCount": len(snapshot.Artists),
		"LastUpdated": time.Now().Format(time.RFC3339),
	}
	if err := a.renderTemplate(w, "index.html", data); err != nil {
		log.Printf("render index: %v", err)
		a.renderError(w, http.StatusInternalServerError)
	}
}

func (a *App) handleArtistPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/artist" && r.URL.Path != "/artist.html" {
		a.renderError(w, http.StatusNotFound)
		return
	}
	if err := a.renderTemplate(w, "artist.html", nil); err != nil {
		log.Printf("render artist: %v", err)
		a.renderError(w, http.StatusInternalServerError)
	}
}

func (a *App) handleSpotifyArtistPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/artist-spotify" && r.URL.Path != "/artist-spotify.html" {
		a.renderError(w, http.StatusNotFound)
		return
	}
	if err := a.renderTemplate(w, "artist_spotify.html", nil); err != nil {
		log.Printf("render spotify artist: %v", err)
		a.renderError(w, http.StatusInternalServerError)
	}
}

func (a *App) handleDatesPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/dates" && r.URL.Path != "/dates.html" {
		a.renderError(w, http.StatusNotFound)
		return
	}
	if err := a.renderTemplate(w, "dates.html", nil); err != nil {
		log.Printf("render dates: %v", err)
		a.renderError(w, http.StatusInternalServerError)
	}
}

func (a *App) handleLocationsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/locations" && r.URL.Path != "/locations.html" {
		a.renderError(w, http.StatusNotFound)
		return
	}
	if err := a.renderTemplate(w, "locations.html", nil); err != nil {
		log.Printf("render locations: %v", err)
		a.renderError(w, http.StatusInternalServerError)
	}
}

func (a *App) handleRelationsPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/relations" && r.URL.Path != "/relations.html" {
		a.renderError(w, http.StatusNotFound)
		return
	}
	if err := a.renderTemplate(w, "relations.html", nil); err != nil {
		log.Printf("render relations: %v", err)
		a.renderError(w, http.StatusInternalServerError)
	}
}

func (a *App) handle404Page(w http.ResponseWriter, _ *http.Request) {
	a.renderError(w, http.StatusNotFound)
}

func (a *App) handle500Page(w http.ResponseWriter, _ *http.Request) {
	a.renderError(w, http.StatusInternalServerError)
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *App) handleAPIArtists(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}
	a.ensureCache(r.Context())
	q := r.URL.Query()
	nameFilter := strings.ToLower(strings.TrimSpace(q.Get("name")))
	yearFilter, _ := strconv.Atoi(q.Get("year"))
	memberFilter := strings.TrimSpace(q.Get("member"))

	sourceParam := strings.ToLower(strings.TrimSpace(q.Get("source")))
	externalParam := strings.ToLower(strings.TrimSpace(q.Get("external")))
	includeSpotify := sourceParam == "spotify" || sourceParam == "all" || externalParam == "spotify"
	includeGroupie := sourceParam == "" || sourceParam == "groupie" || sourceParam == "all"
	if externalParam == "spotify" && sourceParam == "" {
		includeGroupie = true
	}
	unifiedResponse := includeSpotify || sourceParam == "groupie"
	limitParam, _ := strconv.Atoi(q.Get("limit"))
	spotifyLimit := limitParam
	if spotifyLimit <= 0 {
		spotifyLimit = 8
	}

	artists := a.cache.ArtistsWithMeta()
	filtered := make([]ArtistWithMeta, 0, len(artists))
	for _, art := range artists {
		if nameFilter != "" && !strings.Contains(strings.ToLower(art.Name), nameFilter) {
			continue
		}
		if yearFilter > 0 && art.CreationDate != yearFilter {
			continue
		}
		if memberFilter != "" && !containsMember(art.Members, memberFilter) {
			continue
		}
		filtered = append(filtered, art)
	}

	// Legacy behaviour: only return Groupie Tracker data unless a unified response is requested.
	if !unifiedResponse {
		writeJSON(w, http.StatusOK, filtered)
		return
	}

	if a.spotify == nil && !includeGroupie {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "l'intégration Spotify n'est pas configurée"})
		return
	}

	groupieUnified := make([]UnifiedArtist, 0, len(filtered))
	if includeGroupie {
		for _, art := range filtered {
			groupieUnified = append(groupieUnified, toUnifiedGroupie(art))
		}
	}

	spotifyUnified := make([]UnifiedArtist, 0)
	if includeSpotify && a.spotify != nil && nameFilter != "" {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		results, err := a.spotify.SearchArtists(ctx, nameFilter, spotifyLimit)
		if err != nil {
			log.Printf("spotify search failed: %v", err)
		} else {
			for _, sa := range results {
				u := toUnifiedSpotify(sa)
				if strings.TrimSpace(u.ImageURL) == "" || strings.TrimSpace(u.Name) == "" {
					continue
				}
				spotifyUnified = append(spotifyUnified, u)
			}
		}
	}

	merged := mergeUnifiedArtists(groupieUnified, spotifyUnified)
	writeJSON(w, http.StatusOK, merged)
}

func (a *App) handleAPIArtistByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}
	if !strings.HasPrefix(r.URL.Path, "/api/artists/") {
		a.renderError(w, http.StatusNotFound)
		return
	}
	idStr := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/artists/"), "/")
	if idStr == "" {
		a.handleAPIArtists(w, r)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "identifiant d'artiste invalide"})
		return
	}
	for _, art := range a.cache.ArtistsWithMeta() {
		if art.ID == id {
			writeJSON(w, http.StatusOK, art)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "artiste introuvable"})
}

func (a *App) handleAPILocations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}
	a.ensureCache(r.Context())
	snap := a.cache.Snapshot()
	names := make(map[int]string, len(snap.Artists))
	for _, a := range snap.Artists {
		names[a.ID] = a.Name
	}
	relCounts := make(map[int]map[string]int, len(snap.Relations))
	for _, rel := range snap.Relations {
		counts := make(map[string]int, len(rel.DatesLocations))
		for slug, dates := range rel.DatesLocations {
			counts[slug] = len(dates)
		}
		relCounts[rel.ID] = counts
	}
	q := r.URL.Query()
	countryFilter := strings.ToLower(strings.TrimSpace(q.Get("country")))
	cityFilter := strings.ToLower(strings.TrimSpace(q.Get("city")))
	artistFilter := strings.ToLower(strings.TrimSpace(q.Get("artist")))

	views := make([]viewLocation, 0)
	for _, loc := range snap.Locations {
		for _, slug := range loc.Locations {
			name := splitLocationSlug(slug)
			view := viewLocation{
				ArtistID:   loc.ID,
				ArtistName: names[loc.ID],
				City:       name.City,
				Country:    name.Country,
				Raw:        name.Raw,
				EventCount: relCounts[loc.ID][slug],
			}
			if countryFilter != "" && !strings.Contains(strings.ToLower(view.Country), countryFilter) {
				continue
			}
			if cityFilter != "" && !strings.Contains(strings.ToLower(view.City), cityFilter) {
				continue
			}
			if artistFilter != "" && !strings.Contains(strings.ToLower(view.ArtistName), artistFilter) {
				continue
			}
			views = append(views, view)
		}
	}
	writeJSON(w, http.StatusOK, views)
}

func (a *App) handleAPIDates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}
	a.ensureCache(r.Context())
	q := r.URL.Query()
	yearFilter, _ := strconv.Atoi(q.Get("year"))

	var filtered []DatesIndex
	for _, entry := range a.cache.Snapshot().Dates {
		if yearFilter == 0 {
			filtered = append(filtered, entry)
			continue
		}
		matching := make([]string, 0, len(entry.Dates))
		for _, d := range entry.Dates {
			ts, err := parseAPIDate(d)
			if err != nil {
				continue
			}
			if ts.Year() == yearFilter {
				matching = append(matching, d)
			}
		}
		if len(matching) > 0 {
			filtered = append(filtered, DatesIndex{
				ID:    entry.ID,
				Dates: matching,
			})
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (a *App) handleAPIRelation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}
	a.ensureCache(r.Context())
	q := r.URL.Query()
	artistFilter, _ := strconv.Atoi(q.Get("id"))
	relations := a.cache.Snapshot().Relations
	if artistFilter > 0 {
		out := make([]Relation, 0, 1)
		for _, rel := range relations {
			if rel.ID == artistFilter {
				out = append(out, rel)
				break
			}
		}
		writeJSON(w, http.StatusOK, out)
		return
	}
	writeJSON(w, http.StatusOK, relations)
}

func (a *App) handleAPIEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}
	a.ensureCache(r.Context())
	events := a.cache.Events()
	q := r.URL.Query()
	countryFilter := strings.ToLower(strings.TrimSpace(q.Get("country")))
	cityFilter := strings.ToLower(strings.TrimSpace(q.Get("city")))
	artistFilter := strings.ToLower(strings.TrimSpace(q.Get("artist")))
	yearFilter, _ := strconv.Atoi(q.Get("year"))

	filtered := make([]Event, 0, len(events))
	for _, ev := range events {
		if countryFilter != "" && !strings.Contains(strings.ToLower(ev.Country), countryFilter) {
			continue
		}
		if cityFilter != "" && !strings.Contains(strings.ToLower(ev.City), cityFilter) {
			continue
		}
		if artistFilter != "" && !strings.Contains(strings.ToLower(ev.ArtistName), artistFilter) {
			continue
		}
		if yearFilter > 0 && ev.Date.Year() != yearFilter {
			continue
		}
		filtered = append(filtered, ev)
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (a *App) handleAPISpotifyArtist(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}
	if a.spotify == nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "l'intégration Spotify n'est pas configurée"})
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "l'identifiant est requis"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
	defer cancel()
	artist, err := a.spotify.GetArtist(ctx, id)
	if err != nil {
		log.Printf("spotify artist lookup failed: %v", err)
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "artiste introuvable"})
		return
	}
	view := spotifyArtistDetail{
		ID:         artist.ID,
		Name:       artist.Name,
		ImageURL:   pickBestImage(artist.Images),
		Genres:     artist.Genres,
		Popularity: artist.Popularity,
		Followers:  artist.Followers.Total,
		Source:     sourceSpotify,
	}
	writeJSON(w, http.StatusOK, view)
}

func (a *App) renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return a.templates.ExecuteTemplate(w, name, data)
}

func (a *App) renderError(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	name := "500.html"
	if status == http.StatusNotFound {
		name = "404.html"
	}
	if err := a.templates.ExecuteTemplate(w, name, nil); err != nil {
		http.Error(w, http.StatusText(status), status)
	}
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", http.MethodGet)
	http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("encode json: %v", err)
	}
}

func containsMember(members []string, needle string) bool {
	needle = strings.ToLower(needle)
	for _, m := range members {
		if strings.Contains(strings.ToLower(m), needle) {
			return true
		}
	}
	return false
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Truncate(time.Millisecond))
	})
}

// reloadData fetches data in the background; useful for tests or hooks.
func (a *App) reloadData(ctx context.Context) error {
	return a.refreshData(ctx)
}

// templatePath is exposed for tests to ensure template lookup works as expected.
func templatePath(base, name string) string {
	return filepath.Join(base, name)
}

// parseYear helper to safely parse a year value.
func parseYear(value string) (int, error) {
	if strings.TrimSpace(value) == "" {
		return 0, nil
	}
	year, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid year %q", value)
	}
	if year < 0 {
		return 0, errors.New("year must be positive")
	}
	return year, nil
}

// ensureCache triggers a refresh if the cache is empty.
func (a *App) ensureCache(ctx context.Context) {
	if a.api == nil {
		return
	}
	if len(a.cache.Snapshot().Artists) == 0 {
		if err := a.refreshData(ctx); err != nil {
			log.Printf("refresh data: %v", err)
		}
	}
}
