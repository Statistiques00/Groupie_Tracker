package main

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Relations est un type personnalisé pour gérer plusieurs formats
// retournés par l'API : parfois un objet map[string][]string, parfois
// une chaîne ou des tableaux mixtes. Nous normalisons en map[string][]string.
type Relations map[string][]string

func (r *Relations) UnmarshalJSON(b []byte) error {
	// cas : string ("" ou message) -> ignorer
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		// treat string as empty relations
		*r = nil
		return nil
	}

	// cas habituel : map[string][]string
	var m map[string][]string
	if err := json.Unmarshal(b, &m); err == nil {
		*r = m
		return nil
	}

	// cas mixte : map[string]interface{} où les valeurs peuvent être []interface{} ou autres
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err == nil {
		out := make(map[string][]string)
		for k, v := range raw {
			switch vv := v.(type) {
			case []interface{}:
				for _, x := range vv {
					out[k] = append(out[k], fmt.Sprint(x))
				}
			default:
				out[k] = []string{fmt.Sprint(vv)}
			}
		}
		*r = out
		return nil
	}

	return errors.New("unsupported relations format")
}

// Artist représente la structure de l'API Groupie Tracker.
// Les tags JSON sont basés sur la structure publique habituelle.
type Artist struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Image        string    `json:"image"`
	Members      []string  `json:"members"`
	CreationDate int       `json:"creationDate"`
	FirstAlbum   string    `json:"firstAlbum"`
	Relations    Relations `json:"relations"`
}
