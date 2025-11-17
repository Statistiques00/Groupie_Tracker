package main

import (
	"html/template"
	"log"
	"net/http"
)

var tpl = template.Must(template.ParseFiles("templates/index.html"))

func main() {
	// Fichiers statiques (CSS, JS, images…)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Route principale
	http.HandleFunc("/", homeHandler)

	log.Println("Server running on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Plus tard tu passeras ici les données API
	err := tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
