package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

const apiBase = "https://groupietrackers.herokuapp.com/api"

var (
	cacheMu      sync.Mutex
	artistsCache []Artist
	cacheUpdated time.Time
	cacheTTL     = 5 * time.Minute
)

// getArtistsCached récupère la liste d'artistes en utilisant un cache en mémoire.
// Si l'API est indisponible, retourne les données en cache si présentes.
func getArtistsCached() ([]Artist, error) {
	cacheMu.Lock()
	if time.Since(cacheUpdated) < cacheTTL && len(artistsCache) > 0 {
		cached := make([]Artist, len(artistsCache))
		copy(cached, artistsCache)
		cacheMu.Unlock()
		return cached, nil
	}
	cacheMu.Unlock()

	artists, err := fetchArtists()
	if err != nil {
		// si on a un cache précédent, on le retourne (stale) au lieu d'échouer
		cacheMu.Lock()
		if len(artistsCache) > 0 {
			cached := make([]Artist, len(artistsCache))
			copy(cached, artistsCache)
			cacheMu.Unlock()
			return cached, nil
		}
		cacheMu.Unlock()
		return nil, err
	}

	cacheMu.Lock()
	artistsCache = make([]Artist, len(artists))
	copy(artistsCache, artists)
	cacheUpdated = time.Now()
	cacheMu.Unlock()
	return artists, nil
}

// fetchArtists appelle directement l'API externe et parse la réponse JSON.
func fetchArtists() ([]Artist, error) {
	client := &http.Client{
		Timeout: 8 * time.Second,
	}
	resp, err := client.Get(apiBase + "/artists")
	if err != nil {
		log.Println("fetchArtists error get:", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Println("fetchArtists bad status:", resp.StatusCode, string(b))
		return nil, errors.New("bad status from API")
	}

	var artists []Artist
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&artists); err != nil {
		return nil, err
	}
	return artists, nil
}
