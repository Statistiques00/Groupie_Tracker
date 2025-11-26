# Explication complète – Groupie Tracker

## 1. Introduction générale
Groupie Tracker s’appuie sur l’API publique fournie (endpoints `/artists`, `/locations`, `/dates`, `/relation`) pour afficher des artistes, leurs membres, leurs concerts (dates/lieux) et les relations entre ces données. L’application ajoute une recherche moderne côté client, des pages dédiées et peut, si configurée, enrichir les résultats avec Spotify.

Contraintes respectées :
- Backend 100 % Go (librairie standard), templates HTML, manipulation JSON.
- Aucune dépendance à un framework Go ou JS lourd : serveur HTTP standard, fetch côté navigateur, DOM vanilla.
- Séparation nette client/serveur : HTML généré par templates, données servies en JSON, rendu dynamique en JS.

## 2. Vue d’ensemble de l’architecture
Arborescence (principaux éléments) :
```
./                 fichiers Go (serveur, handlers, clients API, modèles)
templates/         pages HTML (index, artist, artist_spotify, dates, locations, relations, 404, 500)
static/css/        styles globaux (look sombre, gradients, cartes)
static/js/         logique front (fetch, rendu DOM, filtres)
docs/              documentation projet
```
Backend : exécute un serveur HTTP, sert les templates et expose des endpoints JSON.  
Frontend : pages statiques (HTML+CSS) et scripts qui interrogent les endpoints (`/api/...`) pour afficher cartes, timelines, accordéons, etc.

## 3. Backend Go : architecture et logique
- `server.go`  
  - `App` regroupe cache, clients API (Groupie Tracker, Spotify), templates, config statique.  
  - `newApp` parse les templates, initialise le client API principal et éventuellement `SpotifyClient`.  
  - `routes` enregistre les handlers (pages HTML + endpoints API + assets).  
  - `main` lit les flags (`-addr`, `-api`, `-static`, `-templates`, `-spotify-client-id/secret`), crée l’`App`, précharge le cache via `refreshData`, lance le serveur HTTP.

- `api_client.go`  
  - `APIClient` wrappe l’API Groupie Tracker (`/artists`, `/locations`, `/dates`, `/relation`) avec `fetch` générique.  
  - `FetchAll` lance en parallèle la récupération des 4 datasets et renvoie un `DataBundle`.

- `spotify_client.go` (optionnel, activé si ID/secret fournis)  
  - Gère le token client credentials (POST `accounts.spotify.com/api/token`), le met en cache avec expiration.  
  - `SearchArtists` interroge `/v1/search?type=artist` avec filtres (nom non vide, images obligatoires, popularité minimale).  
  - `GetArtist` interroge `/v1/artists/{id}` pour le détail.  
  - Requêtes signées via le token Bearer, timeouts courts.

- `data.go`  
  - Modèles : `Artist`, `LocationIndex`, `DatesIndex`, `Relation`, `ArtistWithMeta`, `LocationName`, `Event`.  
  - Helpers : `parseAPIDate` (dates `DD-MM-YYYY` ou `YYYY-MM-DD`), `splitLocationSlug` (slug vers ville/pays), `titleCase`.  
  - `mergeArtists` assemble données `Artists` avec `Locations`/`Dates`/`Relations` en `ArtistWithMeta`.  
  - `buildEvents` aplatit `relations` en liste chronologique d’`Event` (ville/pays nommés, date ISO).

- `unified_artist.go`  
  - Définit `UnifiedArtist` (shape harmonisée).  
  - `toUnifiedGroupie`, `toUnifiedSpotify`, `pickBestImage`, `mergeUnifiedArtists` (dedoublonnage par nom, priorité Groupie sur Spotify).

- `cache.go`  
  - `Cache` stocke le dernier `DataBundle` et un timestamp.  
  - `Set`, `Snapshot` (copies défensives), `ArtistsWithMeta`, `Events` (recalcule les événements).

- `handlers.go`  
  - Pages HTML : `handleRoot`, `handleArtistPage`, `handleSpotifyArtistPage`, `handleDatesPage`, `handleLocationsPage`, `handleRelationsPage`, `handle404Page`, `handle500Page` → rendent les templates.  
  - Endpoints JSON :  
    - `/api/artists` : filtres `name`, `year`, `member`, `source` (`groupie|spotify|all`), `external=spotify`, `limit`. Renvoie `ArtistWithMeta` (Groupie) ou fusion Groupie/Spotify (`UnifiedArtist`).  
    - `/api/artists/{id}` : détail groupie par ID.  
    - `/api/locations` : liste des lieux flatten (ville/pays, artiste, nombre d’événements) avec filtres `country`, `city`, `artist`.  
    - `/api/dates` : dates filtrables par `year`.  
    - `/api/relation` : toutes les relations ou une seule via `id`.  
    - `/api/events` : événements aplanis, filtres `country`, `city`, `artist`, `year`.  
    - `/api/spotify/artist` : proxy détail Spotify (404/503/400 gérés).  
  - Utilitaires : `methodNotAllowed`, `writeJSON`, `ensureCache` (rafraîchit si cache vide), `renderError` (404/500).

- `handlers_test.go` / `data_test.go`  
  - Vérifient filtres `/api/artists`, normalisation d’événements, 404 root, parsing dates et slugs, tri chrono.

## 4. Frontend : pages et comportement
Templates (HTML) :
- `index.html` : page d’accueil, hero recherche, toggle source (Groupie/Tous/Spotify), stats, sections “Artistes mis en avant” et cartes de navigation.  
- `artist.html` : détail d’un artiste Groupie (hero image, membres, dates de tournée avec toggle liste/chronologie).  
- `artist_spotify.html` : détail d’un artiste Spotify (stats followers/popularité, genres, lien “Ouvrir sur Spotify”).  
- `dates.html` : chronologie des concerts avec filtres année/pays/texte.  
- `locations.html` : grille des lieux par artiste avec filtres pays/artiste/ville.  
- `relations.html` : accordéon relations artiste → lieux/dates avec filtres artiste/lieu.  
- `404.html` / `500.html` : pages d’erreur stylées.

Scripts (JS) :
- `static/js/index.js` :  
  - Appelle `/api/artists` avec filtres nom/source/limit.  
  - Gère le toggle de source, la recherche, le fallback cache si résultat vide.  
  - Rend des cartes artistes (badge source, popularité/genres ou membres, CTA “Voir sur Spotify”/“Voir les détails”).  
  - Messages d’erreur utilisateur si fetch échoue ou si recherche Spotify sans terme.
- `static/js/artist.js` :  
  - Charge `/api/artists` + `/api/relation`, trouve l’artiste par `id` dans l’URL.  
  - Construit les concerts (ville/pays/dates) et rend vue liste + timeline avec toggle.  
  - Redirige vers 404/500 en cas d’erreur.
- `static/js/artist_spotify.js` :  
  - Appelle `/api/spotify/artist?id=...`, rend hero (badge, followers, popularité), stats cartes, genres.  
  - Bouton externe vers Spotify.
- `static/js/dates.js` :  
  - Appelle `/api/events`, peuple les filtres (années, pays), applique filtres live, rend timeline (ville – pays, artiste, date locale).  
  - Message d’erreur si fetch échoue.
- `static/js/locations.js` :  
  - Appelle `/api/locations`, peuple filtres (pays, artiste), filtre par ville, rend cartes (ville, pays, artiste, nombre de concerts).  
  - Message d’erreur si fetch échoue.
- `static/js/relations.js` :  
  - Appelle `/api/artists` + `/api/relation`, construit une map ID→nom, rend accordéon avec badges (lieux/dates), filtres artiste/lieu.  
  - Message d’erreur si fetch échoue.

## 5. Flux de données complet (scénarios)
- Chargement de la page d’accueil  
  1) Navigateur GET `/` → handler Go rend `index.html`.  
  2) JS (`index.js`) se lance, appelle `/api/artists?source=all&limit=8`.  
  3) Backend filtre cache Groupie (et Spotify si configuré et demandé), fusionne, renvoie JSON.  
  4) JS rend les cartes artistes, met à jour le compteur.

- Recherche d’un artiste  
  1) L’utilisateur saisit un nom, submit.  
  2) JS envoie `/api/artists?name=...&source=...`.  
  3) Backend filtre Groupie (et Spotify si `source=spotify|all` + nom non vide), merge résultats.  
  4) JS affiche la liste ; si vide, affiche “Aucun artiste trouvé…” ; si Spotify sans terme, avertit l’utilisateur.

- Affichage d’un artiste Groupie Tracker  
  1) Clic sur une carte Groupie → `/artist?id=...`.  
  2) Handler rend `artist.html`.  
  3) JS charge `/api/artists` (cache) + `/api/relation`, trouve l’artiste par ID.  
  4) Affiche hero (image, nom, année, premier album, membres), construit concerts (liste + timeline), toggle vue.

- Timeline des concerts  
  1) Visite `/dates` → template `dates.html`.  
  2) JS appelle `/api/events`, peuple filtres année/pays, rend timeline.  
  3) Filtrer déclenche un rerendu local.

- Relations entre artistes/dates/lieux  
  1) Visite `/relations` → template `relations.html`.  
  2) JS appelle `/api/artists` + `/api/relation`, construit vues par artiste.  
  3) Accordéon interactif ; filtres artiste/lieu filtrent côté client.

- Lieux de tournée  
  1) Visite `/locations` → template `locations.html`.  
  2) JS appelle `/api/locations`, remplit filtres (pays, artiste), applique filtres ville/texte, rend cartes.

## 6. Intégration Spotify
- Authentification : client credentials flow, token stocké en mémoire avec expiration.  
- Recherche : `/v1/search?type=artist&q=...&limit=...` ; résultats rejetés si nom vide, image absente, popularité trop faible.  
- Détail : `/v1/artists/{id}` renvoie images, genres, popularité, followers.  
- Fusion : `/api/artists` combine Groupie (toujours) et Spotify (si `source=spotify|all` et nom fourni) via `UnifiedArtist`, avec dédoublonnage par nom (priorité Groupie).  
- Front : cartes Spotify portent un badge et un CTA “Voir sur Spotify” ; page `artist_spotify.html` affiche stats/genres et bouton externe.

## 7. Gestion des erreurs et robustesse
- Backend :  
  - 404/500 via `renderError` (templates dédiés).  
  - `/api/artists/{id}` : 400 si ID invalide, 404 si introuvable.  
  - `/api/spotify/artist` : 400 sans id, 503 si Spotify non configuré, 404 si introuvable, 5xx si erreurs amont.  
  - `methodNotAllowed` pour les méthodes ≠ GET.  
  - Logs serveur pour diagnostics (sans polluer la réponse utilisateur).
- Frontend :  
  - Messages utilisateur traduits : “Aucun artiste trouvé…”, “Échec du chargement des événements…”, “La recherche Spotify est indisponible…”, etc.  
  - Redirection vers `/404` ou `/500` dans les scripts en cas d’absence d’ID ou d’erreur critique.
- Pages 404 / 500 : messages stylés, CTA retour / rechargement.

## 8. Tests unitaires
- `data_test.go` : vérifie `parseAPIDate`, `splitLocationSlug`, `buildEvents` (tri chrono, format ISO, propagation du nom d’artiste).  
- `handlers_test.go` : filtre `/api/artists`, normalisation `/api/events`, 404 sur route inconnue.  
- Exécution : `go test ./...` à la racine.

## 9. Pistes d’extensions possibles
- Nouveaux filtres : par popularité Spotify, par nombre de membres, par intervalle de dates.  
- Intégrer d’autres APIs musicales (Deezer, MusicBrainz) via un nouveau client et unifier dans `UnifiedArtist`.  
- UX : lazy-loading, pagination, favoris, tri dynamique, cartes avec audio preview Spotify.  
- i18n : externaliser les chaînes (fichiers JSON/PO) et injecter via templates/JS ; ajouter un sélecteur de langue.  
- Visualisations : cartes géographiques (Leaflet/Mapbox), graphes de relations, histogrammes de tournées.  
- Robustesse : stockage cache persistant, circuit breaker pour Spotify, métriques/observabilité (Prometheus).  
- Sécurité : rate limiting léger, validation stricte des paramètres de requête.
