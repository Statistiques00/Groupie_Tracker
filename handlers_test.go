package main

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestApp() *App {
	tpl := template.Must(template.New("index.html").Parse("index"))
	template.Must(tpl.New("artist.html").Parse("artist"))
	template.Must(tpl.New("artist_spotify.html").Parse("artist_spotify"))
	template.Must(tpl.New("dates.html").Parse("dates"))
	template.Must(tpl.New("locations.html").Parse("locations"))
	template.Must(tpl.New("relations.html").Parse("relations"))
	template.Must(tpl.New("404.html").Parse("404"))
	template.Must(tpl.New("500.html").Parse("500"))

	return &App{
		cache:     newCache(),
		api:       nil,
		templates: tpl,
		staticDir: defaultStaticDir,
	}
}

func TestHandleAPIArtistsFilters(t *testing.T) {
	app := newTestApp()
	app.cache.Set(DataBundle{
		Artists: []Artist{
			{ID: 1, Name: "Alpha", Members: []string{"A"}, CreationDate: 2000},
			{ID: 2, Name: "Beta", Members: []string{"B"}, CreationDate: 2010},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/artists?name=alp", nil)
	rr := httptest.NewRecorder()
	app.handleAPIArtists(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rr.Code)
	}
	var artists []ArtistWithMeta
	if err := json.NewDecoder(rr.Body).Decode(&artists); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(artists) != 1 || artists[0].Name != "Alpha" {
		t.Fatalf("filtering failed, got %+v", artists)
	}
}

func TestHandleAPIEvents(t *testing.T) {
	app := newTestApp()
	app.cache.Set(DataBundle{
		Artists: []Artist{{ID: 1, Name: "Gamma"}},
		Relations: []Relation{
			{ID: 1, DatesLocations: map[string][]string{"london-uk": {"01-01-2020"}}},
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/events?country=uk", nil)
	rr := httptest.NewRecorder()
	app.handleAPIEvents(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("unexpected status %d", rr.Code)
	}
	var events []Event
	if err := json.NewDecoder(rr.Body).Decode(&events); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(events) != 1 || events[0].Country != "Uk" || events[0].City != "London" {
		t.Fatalf("unexpected events %+v", events)
	}
}

func TestHandleRootNotFound(t *testing.T) {
	app := newTestApp()
	req := httptest.NewRequest(http.MethodGet, "/unknown", nil)
	rr := httptest.NewRecorder()
	app.handleRoot(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}
