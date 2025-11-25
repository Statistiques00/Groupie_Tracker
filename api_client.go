package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// APIClient wraps access to the upstream Groupie Tracker API.
type APIClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

func newAPIClient(baseURL string) *APIClient {
	base := strings.TrimRight(baseURL, "/")
	if base == "" {
		base = defaultAPIBase
	}
	return &APIClient{
		BaseURL: base,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *APIClient) buildURL(path string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(c.BaseURL, "/"), strings.TrimLeft(path, "/"))
}

func (c *APIClient) fetch(ctx context.Context, path string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.buildURL(path), nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("upstream %s returned %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

func (c *APIClient) FetchArtists(ctx context.Context) ([]Artist, error) {
	var artists []Artist
	if err := c.fetch(ctx, "/artists", &artists); err != nil {
		return nil, err
	}
	return artists, nil
}

func (c *APIClient) FetchLocations(ctx context.Context) ([]LocationIndex, error) {
	var payload struct {
		Index []LocationIndex `json:"index"`
	}
	if err := c.fetch(ctx, "/locations", &payload); err != nil {
		return nil, err
	}
	return payload.Index, nil
}

func (c *APIClient) FetchDates(ctx context.Context) ([]DatesIndex, error) {
	var payload struct {
		Index []DatesIndex `json:"index"`
	}
	if err := c.fetch(ctx, "/dates", &payload); err != nil {
		return nil, err
	}
	return payload.Index, nil
}

func (c *APIClient) FetchRelations(ctx context.Context) ([]Relation, error) {
	var payload struct {
		Index []Relation `json:"index"`
	}
	if err := c.fetch(ctx, "/relation", &payload); err != nil {
		return nil, err
	}
	return payload.Index, nil
}

// FetchAll pulls all datasets concurrently.
func (c *APIClient) FetchAll(ctx context.Context) (DataBundle, error) {
	var (
		artists   []Artist
		locations []LocationIndex
		dates     []DatesIndex
		relations []Relation
		errA      error
		errL      error
		errD      error
		errR      error
		wg        sync.WaitGroup
	)

	wg.Add(4)
	go func() {
		defer wg.Done()
		artists, errA = c.FetchArtists(ctx)
	}()
	go func() {
		defer wg.Done()
		locations, errL = c.FetchLocations(ctx)
	}()
	go func() {
		defer wg.Done()
		dates, errD = c.FetchDates(ctx)
	}()
	go func() {
		defer wg.Done()
		relations, errR = c.FetchRelations(ctx)
	}()
	wg.Wait()

	switch {
	case errA != nil:
		return DataBundle{}, errA
	case errL != nil:
		return DataBundle{}, errL
	case errD != nil:
		return DataBundle{}, errD
	case errR != nil:
		return DataBundle{}, errR
	}

	return DataBundle{
		Artists:   artists,
		Locations: locations,
		Dates:     dates,
		Relations: relations,
	}, nil
}
