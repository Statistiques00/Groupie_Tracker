package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	spotifyTokenURL  = "https://accounts.spotify.com/api/token"
	spotifySearchURL = "https://api.spotify.com/v1/search"
	spotifyArtistURL = "https://api.spotify.com/v1/artists/"
)

var (
	ErrSpotifyNotFound = errors.New("spotify artist not found")
	ErrSpotifyUpstream = errors.New("spotify upstream error")
)

// SpotifyArtist mirrors the subset of fields we need from the Spotify API.
type SpotifyArtist struct {
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Genres     []string         `json:"genres"`
	Popularity int              `json:"popularity"`
	Images     []SpotifyImage   `json:"images"`
	Followers  SpotifyFollowers `json:"followers"`
}

type SpotifyFollowers struct {
	Total int `json:"total"`
}

type SpotifyImage struct {
	URL    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

// SpotifyClient handles authentication and artist lookups.
type SpotifyClient struct {
	ClientID     string
	ClientSecret string
	HTTPClient   *http.Client

	tokenMu     sync.Mutex
	accessToken string
	expiresAt   time.Time
}

func newSpotifyClient(clientID, clientSecret string) *SpotifyClient {
	return &SpotifyClient{
		ClientID:     strings.TrimSpace(clientID),
		ClientSecret: strings.TrimSpace(clientSecret),
		HTTPClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}
}

// token retrieves (or reuses) a client credentials token.
func (c *SpotifyClient) token(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()

	if c.ClientID == "" || c.ClientSecret == "" {
		return "", errors.New("spotify credentials are missing")
	}

	// Reuse token when it is still valid.
	if c.accessToken != "" && time.Now().Before(c.expiresAt.Add(-30*time.Second)) {
		return c.accessToken, nil
	}

	body := strings.NewReader("grant_type=client_credentials")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, spotifyTokenURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	basic := base64.StdEncoding.EncodeToString([]byte(c.ClientID + ":" + c.ClientSecret))
	req.Header.Set("Authorization", "Basic "+basic)

	resp, err := c.http().Do(req)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrSpotifyUpstream, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("%w: token request failed: %d", ErrSpotifyUpstream, resp.StatusCode)
	}
	var payload struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	if payload.AccessToken == "" {
		return "", errors.New("spotify returned empty access token")
	}
	c.accessToken = payload.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(payload.ExpiresIn) * time.Second)
	return c.accessToken, nil
}

// SearchArtists calls the Spotify Search API for artists only and filters out
// incomplete or low-signal results (no name, no image).
func (c *SpotifyClient) SearchArtists(ctx context.Context, query string, limit int) ([]SpotifyArtist, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return nil, nil
	}
	token, err := c.token(ctx)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 20 {
		limit = 8
	}

	u, err := url.Parse(spotifySearchURL)
	if err != nil {
		return nil, err
	}
	params := u.Query()
	params.Set("q", q)
	params.Set("type", "artist")
	params.Set("limit", strconv.Itoa(limit))
	u.RawQuery = params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http().Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSpotifyUpstream, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: search failed: %d", ErrSpotifyUpstream, resp.StatusCode)
	}

	var payload struct {
		Artists struct {
			Items []SpotifyArtist `json:"items"`
		} `json:"artists"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	results := make([]SpotifyArtist, 0, len(payload.Artists.Items))
	for _, artist := range payload.Artists.Items {
		if strings.TrimSpace(artist.Name) == "" {
			continue
		}
		if len(artist.Images) == 0 {
			continue
		}
		// Light popularity gate to avoid noisy matches.
		if artist.Popularity > 0 && artist.Popularity < 5 {
			continue
		}
		results = append(results, artist)
	}
	return results, nil
}

// GetArtist fetches full details for a specific Spotify artist.
func (c *SpotifyClient) GetArtist(ctx context.Context, id string) (SpotifyArtist, error) {
	token, err := c.token(ctx)
	if err != nil {
		return SpotifyArtist{}, err
	}
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return SpotifyArtist{}, errors.New("spotify artist id is required")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, spotifyArtistURL+url.PathEscape(trimmed), nil)
	if err != nil {
		return SpotifyArtist{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.http().Do(req)
	if err != nil {
		return SpotifyArtist{}, fmt.Errorf("%w: %v", ErrSpotifyUpstream, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return SpotifyArtist{}, fmt.Errorf("%w: %s", ErrSpotifyNotFound, trimmed)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return SpotifyArtist{}, fmt.Errorf("%w: artist fetch failed: %d", ErrSpotifyUpstream, resp.StatusCode)
	}

	var artist SpotifyArtist
	if err := json.NewDecoder(resp.Body).Decode(&artist); err != nil {
		return SpotifyArtist{}, err
	}
	if strings.TrimSpace(artist.Name) == "" || len(artist.Images) == 0 {
		return SpotifyArtist{}, fmt.Errorf("%w: incomplete artist data", ErrSpotifyUpstream)
	}
	return artist, nil
}

func (c *SpotifyClient) http() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}
