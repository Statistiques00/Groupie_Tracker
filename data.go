package main

import (
	"errors"
	"sort"
	"strings"
	"time"
)

const defaultAPIBase = "https://groupietrackers.herokuapp.com/api"

// Artist describes the payload returned by /artists.
type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	LocationsURL string   `json:"locations"`
	DatesURL     string   `json:"concertDates"`
	RelationsURL string   `json:"relations"`
}

// LocationIndex represents one entry from /locations.
type LocationIndex struct {
	ID        int      `json:"id"`
	Locations []string `json:"locations"`
	DatesURL  string   `json:"dates"`
}

// DatesIndex represents one entry from /dates.
type DatesIndex struct {
	ID    int      `json:"id"`
	Dates []string `json:"dates"`
}

// Relation represents one entry from /relation.
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

// DataBundle stores the complete dataset pulled from the upstream API.
type DataBundle struct {
	Artists   []Artist
	Locations []LocationIndex
	Dates     []DatesIndex
	Relations []Relation
}

// ArtistWithMeta enriches an Artist with resolved locations, dates and relations.
type ArtistWithMeta struct {
	Artist
	LocationList   []string            `json:"locations,omitempty"`
	DateList       []string            `json:"dates,omitempty"`
	DatesLocations map[string][]string `json:"datesLocations,omitempty"`
}

// LocationName contains a human readable location for a slug.
type LocationName struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Raw     string `json:"raw"`
}

// Event represents a single concert event with human readable details.
type Event struct {
	ArtistID   int       `json:"artistId"`
	ArtistName string    `json:"artistName"`
	City       string    `json:"city"`
	Country    string    `json:"country"`
	Date       time.Time `json:"-"`
	DateISO    string    `json:"date"`
}

// parseAPIDate handles the different date formats returned by the upstream API.
// Dates may be prefixed with an asterisk and are commonly provided as DD-MM-YYYY.
func parseAPIDate(value string) (time.Time, error) {
	cleaned := strings.TrimSpace(strings.TrimPrefix(value, "*"))
	if cleaned == "" {
		return time.Time{}, errors.New("empty date")
	}
	// If the date starts with a 4 digit year, treat it as YYYY-MM-DD.
	if len(cleaned) >= 10 && cleaned[4] == '-' {
		if parsed, err := time.Parse("2006-01-02", cleaned); err == nil {
			return parsed, nil
		}
	}
	return time.Parse("02-01-2006", cleaned)
}

// splitLocationSlug converts a slug like "los_angeles-usa" into readable city/country names.
func splitLocationSlug(slug string) LocationName {
	parts := strings.Split(slug, "-")
	if len(parts) == 1 {
		return LocationName{
			City:    titleCase(strings.ReplaceAll(slug, "_", " ")),
			Country: "",
			Raw:     slug,
		}
	}
	country := parts[len(parts)-1]
	city := strings.Join(parts[:len(parts)-1], "-")
	return LocationName{
		City:    titleCase(strings.ReplaceAll(city, "_", " ")),
		Country: titleCase(strings.ReplaceAll(country, "_", " ")),
		Raw:     slug,
	}
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) == 0 {
			continue
		}
		words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}

// mergeArtists combines base artist data with related locations, dates and relations.
func mergeArtists(bundle DataBundle) []ArtistWithMeta {
	locMap := make(map[int][]string, len(bundle.Locations))
	for _, loc := range bundle.Locations {
		locMap[loc.ID] = append([]string(nil), loc.Locations...)
	}

	dateMap := make(map[int][]string, len(bundle.Dates))
	for _, d := range bundle.Dates {
		dateMap[d.ID] = append([]string(nil), d.Dates...)
	}

	relMap := make(map[int]map[string][]string, len(bundle.Relations))
	for _, rel := range bundle.Relations {
		copyMap := make(map[string][]string, len(rel.DatesLocations))
		for k, v := range rel.DatesLocations {
			copyMap[k] = append([]string(nil), v...)
		}
		relMap[rel.ID] = copyMap
	}

	enriched := make([]ArtistWithMeta, 0, len(bundle.Artists))
	for _, art := range bundle.Artists {
		enriched = append(enriched, ArtistWithMeta{
			Artist:         art,
			LocationList:   locMap[art.ID],
			DateList:       dateMap[art.ID],
			DatesLocations: relMap[art.ID],
		})
	}
	return enriched
}

// buildEvents flattens relations into a chronological list of events with artist names.
func buildEvents(artists []Artist, relations []Relation) []Event {
	nameByID := make(map[int]string, len(artists))
	for _, a := range artists {
		nameByID[a.ID] = a.Name
	}

	events := make([]Event, 0, len(relations)*4)
	for _, rel := range relations {
		for slug, dates := range rel.DatesLocations {
			loc := splitLocationSlug(slug)
			for _, d := range dates {
				ts, err := parseAPIDate(d)
				if err != nil {
					continue
				}
				events = append(events, Event{
					ArtistID:   rel.ID,
					ArtistName: nameByID[rel.ID],
					City:       loc.City,
					Country:    loc.Country,
					Date:       ts,
				})
			}
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Date.Before(events[j].Date)
	})

	for i := range events {
		events[i].DateISO = events[i].Date.Format("2006-01-02")
	}

	return events
}
