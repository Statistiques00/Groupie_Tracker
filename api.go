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

// apiBase est la racine de l'API distante.
const apiBase = "https://groupietrackers.herokuapp.com/api"

// variables du cache
var (
	cacheMu      sync.RWMutex // allow concurrent readers
	artistsCache []Artist
	cacheUpdated time.Time
	cacheTTL     = 5 * time.Minute
	httpClient   = &http.Client{Timeout: 8 * time.Second}
)

// getArtistsCached retourne une copie de la liste des artistes.
// Comportement :
// - Si le cache est encore valide, le retourne immédiatement.
// - Sinon, tente de récupérer des données fraîches depuis l'API.
// - Si la récupération échoue et qu'un cache précédent existe, retourne le cache périmé.
// - Si la récupération échoue et qu'il n'y a pas de cache, retourne une erreur.
func getArtistsCached() ([]Artist, error) {
	// Chemin rapide : cache valide -> retourne une copie sous verrou en lecture.
	if cached := getCacheIfValid(); cached != nil {
		return cached, nil
	}

	// Récupère des données fraîches depuis l'API distante.
	artists, err := fetchArtists()
	if err != nil {
		log.Println("fetchArtists failed:", err)
		// Si nous avons un cache précédent, le retourner en mode dégradé mais fonctionnel.
		if cached := getCacheCopy(); cached != nil {
			return cached, nil
		}
		// No cache to fall back to -> propagate error.
		return nil, err
	}

	// Update cache with fresh data.
	setCache(artists)
	return artists, nil
}

// getCacheIfValid retourne une copie du cache si celui-ci est encore dans la TTL, sinon nil.
func getCacheIfValid() []Artist {
	cacheMu.RLock()
	valid := time.Since(cacheUpdated) < cacheTTL && len(artistsCache) > 0
	if !valid {
		cacheMu.RUnlock()
		return nil
	}
	copied := make([]Artist, len(artistsCache))
	copy(copied, artistsCache)
	cacheMu.RUnlock()
	return copied
}

// getCacheCopy retourne une copie du cache actuel indépendamment de la TTL, ou nil si vide.
func getCacheCopy() []Artist {
	cacheMu.RLock()
	if len(artistsCache) == 0 {
		cacheMu.RUnlock()
		return nil
	}
	copied := make([]Artist, len(artistsCache))
	copy(copied, artistsCache)
	cacheMu.RUnlock()
	return copied
}

// setCache remplace la tranche en cache et met à jour l'horodatage.
func setCache(a []Artist) {
	cacheMu.Lock()
	artistsCache = make([]Artist, len(a))
	copy(artistsCache, a)
	cacheUpdated = time.Now()
	cacheMu.Unlock()
}

// fetchArtists effectue une requête HTTP GET vers l'API distante et décode le JSON.
// Elle retourne une erreur si la requête échoue ou si la réponse ne peut pas être décodée.
func fetchArtists() ([]Artist, error) {
	resp, err := httpClient.Get(apiBase + "/artists")
	if err != nil {
		log.Println("fetchArtists error get:", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Lit une petite portion du corps pour le logging (évite les très grosses réponses).
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		log.Printf("fetchArtists bad status: %d - %s\n", resp.StatusCode, string(b))
		return nil, errors.New("bad status from API")
	}

	var artists []Artist
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&artists); err != nil {
		log.Println("fetchArtists decode error:", err)
		return nil, err
	}
	return artists, nil
}
