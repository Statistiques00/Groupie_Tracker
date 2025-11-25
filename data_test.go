package main

import (
	"testing"
	"time"
)

func TestParseAPIDate(t *testing.T) {
	tests := []struct {
		input string
		year  int
		day   int
	}{
		{"*23-08-2019", 2019, 23},
		{"2024-01-02", 2024, 2},
		{"07-02-2020", 2020, 7},
	}
	for _, tt := range tests {
		got, err := parseAPIDate(tt.input)
		if err != nil {
			t.Fatalf("parseAPIDate(%q) error: %v", tt.input, err)
		}
		if got.Year() != tt.year || got.Day() != tt.day {
			t.Fatalf("parseAPIDate(%q) = %v, want year %d day %d", tt.input, got, tt.year, tt.day)
		}
	}
}

func TestSplitLocationSlug(t *testing.T) {
	ln := splitLocationSlug("los_angeles-usa")
	if ln.City != "Los Angeles" || ln.Country != "Usa" {
		t.Fatalf("unexpected location %+v", ln)
	}
	ln = splitLocationSlug("paris-france")
	if ln.City != "Paris" || ln.Country != "France" {
		t.Fatalf("unexpected location %+v", ln)
	}
}

func TestBuildEvents(t *testing.T) {
	artists := []Artist{
		{ID: 1, Name: "Test Artist"},
	}
	relations := []Relation{
		{
			ID: 1,
			DatesLocations: map[string][]string{
				"paris-france": {"01-01-2020", "02-01-2020"},
				"london-uk":    {"03-01-2020"},
			},
		},
	}
	events := buildEvents(artists, relations)
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
	if events[0].ArtistName != "Test Artist" {
		t.Fatalf("expected artist name propagated")
	}
	if events[0].Date.After(events[1].Date) {
		t.Fatalf("events not sorted chronologically: %v then %v", events[0].Date, events[1].Date)
	}
	for _, ev := range events {
		if _, err := time.Parse("2006-01-02", ev.DateISO); err != nil {
			t.Fatalf("event date is not ISO formatted: %v", ev.DateISO)
		}
	}
}
