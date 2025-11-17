package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	addr = flag.String("addr", ":8080", "server address")
)

func main() {
	flag.Parse()

	// Fichiers statiques (CSS, JS, images…)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/artist", artistHandler)
	http.HandleFunc("/search", searchHandler)

	srvAddr := *addr
	log.Printf("Server running on %s\n", srvAddr)
	srv := &http.Server{
		Addr:         srvAddr,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

// homeHandler affiche la liste des artistes (rendu côté serveur)
func homeHandler(w http.ResponseWriter, r *http.Request) {
	artists, err := getArtistsCached()
	if err != nil {
		// On continue avec une slice vide pour ne pas planter
		log.Println("warning: could not fetch artists:", err)
		artists = []Artist{}
	}

	data := struct {
		Artists []Artist
	}{Artists: artists}

	// Execute into buffer to avoid partial writes which lead to
	// "superfluous response.WriteHeader" when template execution fails
	// Parse only layout + index to avoid template name collisions between
	// different page templates (index and artist define their own content blocks).
	t, err := template.ParseFiles("templates/layout.html", "templates/index.html")
	if err != nil {
		log.Println("template parse error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout.html", data); err != nil {
		log.Println("template execute error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// On success, copy buffer to ResponseWriter
	if _, err := buf.WriteTo(w); err != nil {
		log.Println("write response error:", err)
	}
}

// artistHandler affiche la page détail d'un artiste
func artistHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	artists, _ := getArtistsCached()
	var found *Artist
	for _, a := range artists {
		if a.ID == id {
			// copy to avoid referencing loop variable
			tmp := a
			found = &tmp
			break
		}
	}
	if found == nil {
		http.NotFound(w, r)
		return
	}

	t, err := template.ParseFiles("templates/layout.html", "templates/artist.html")
	if err != nil {
		log.Println("template parse error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "layout.html", found); err != nil {
		log.Println("template execute error:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if _, err := buf.WriteTo(w); err != nil {
		log.Println("write response error:", err)
	}
}

// searchHandler sert d'endpoint pour la recherche côté client
// retourne JSON des artistes filtrés (par nom)
func searchHandler(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	artists, _ := getArtistsCached()

	var res []Artist
	if q == "" {
		res = artists
	} else {
		ql := strings.ToLower(q)
		for _, a := range artists {
			if strings.Contains(strings.ToLower(a.Name), ql) {
				res = append(res, a)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	if err := enc.Encode(res); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
