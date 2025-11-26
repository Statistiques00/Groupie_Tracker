package main

import (
	"sync"
	"time"
)

// Cache stores the latest dataset fetched from the upstream API.
type Cache struct {
	mu        sync.RWMutex
	data      DataBundle
	fetchedAt time.Time
}

func newCache() *Cache {
	return &Cache{}
}

// Set replaces the cached data with a fresh copy.
func (c *Cache) Set(bundle DataBundle) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = bundle
	c.fetchedAt = time.Now()
}

// Snapshot returns a copy of the cached data to prevent callers from mutating it.
func (c *Cache) Snapshot() DataBundle {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return DataBundle{
		Artists:   cloneArtists(c.data.Artists),
		Locations: cloneLocations(c.data.Locations),
		Dates:     cloneDates(c.data.Dates),
		Relations: cloneRelations(c.data.Relations),
	}
}

func (c *Cache) ArtistsWithMeta() []ArtistWithMeta {
	return mergeArtists(c.Snapshot())
}

func (c *Cache) Events() []Event {
	snap := c.Snapshot()
	return buildEvents(snap.Artists, snap.Relations)
}

func cloneArtists(src []Artist) []Artist {
	out := make([]Artist, len(src))
	copy(out, src)
	return out
}

func cloneLocations(src []LocationIndex) []LocationIndex {
	out := make([]LocationIndex, len(src))
	for i, v := range src {
		out[i] = LocationIndex{
			ID:        v.ID,
			Locations: append([]string(nil), v.Locations...),
			DatesURL:  v.DatesURL,
		}
	}
	return out
}

func cloneDates(src []DatesIndex) []DatesIndex {
	out := make([]DatesIndex, len(src))
	for i, v := range src {
		out[i] = DatesIndex{
			ID:    v.ID,
			Dates: append([]string(nil), v.Dates...),
		}
	}
	return out
}

func cloneRelations(src []Relation) []Relation {
	out := make([]Relation, len(src))
	for i, v := range src {
		copyMap := make(map[string][]string, len(v.DatesLocations))
		for key, dates := range v.DatesLocations {
			copyMap[key] = append([]string(nil), dates...)
		}
		out[i] = Relation{
			ID:             v.ID,
			DatesLocations: copyMap,
		}
	}
	return out
}
