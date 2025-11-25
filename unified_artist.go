package main

import (
	"strconv"
	"strings"
)

const (
	sourceGroupie = "groupie"
	sourceSpotify = "spotify"
)

// UnifiedArtist represents the harmonised shape returned by the search endpoint,
// regardless of whether data comes from the Groupie Tracker API or Spotify.
type UnifiedArtist struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	ImageURL     string   `json:"image_url"`
	Source       string   `json:"source"`
	CreationDate int      `json:"creationDate,omitempty"`
	FirstAlbum   string   `json:"firstAlbum,omitempty"`
	Members      []string `json:"members,omitempty"`
	Genres       []string `json:"genres,omitempty"`
	Popularity   int      `json:"popularity,omitempty"`
}

func toUnifiedGroupie(a ArtistWithMeta) UnifiedArtist {
	return UnifiedArtist{
		ID:           strconv.Itoa(a.ID),
		Name:         a.Name,
		ImageURL:     a.Image,
		Source:       sourceGroupie,
		CreationDate: a.CreationDate,
		FirstAlbum:   a.FirstAlbum,
		Members:      append([]string(nil), a.Members...),
	}
}

func toUnifiedSpotify(a SpotifyArtist) UnifiedArtist {
	return UnifiedArtist{
		ID:         a.ID,
		Name:       a.Name,
		ImageURL:   pickBestImage(a.Images),
		Source:     sourceSpotify,
		Genres:     append([]string(nil), a.Genres...),
		Popularity: a.Popularity,
	}
}

func pickBestImage(images []SpotifyImage) string {
	if len(images) == 0 {
		return ""
	}
	// Spotify already returns images from largest to smallest; keep first.
	if images[0].URL != "" {
		return images[0].URL
	}
	for _, img := range images {
		if img.URL != "" {
			return img.URL
		}
	}
	return ""
}

func mergeUnifiedArtists(groupie []UnifiedArtist, spotify []UnifiedArtist) []UnifiedArtist {
	merged := make([]UnifiedArtist, 0, len(groupie)+len(spotify))
	seen := make(map[string]bool, len(groupie)+len(spotify))

	for _, a := range groupie {
		key := strings.ToLower(strings.TrimSpace(a.Name))
		if key == "" {
			continue
		}
		seen[key] = true
		merged = append(merged, a)
	}
	for _, a := range spotify {
		key := strings.ToLower(strings.TrimSpace(a.Name))
		if key == "" {
			continue
		}
		if seen[key] {
			// Prefer Groupie Tracker entries when names collide.
			continue
		}
		seen[key] = true
		merged = append(merged, a)
	}
	return merged
}
