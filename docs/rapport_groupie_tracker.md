# Rapport integration Spotify

## Configuration
- Authentification Spotify via Client Credentials Flow. Renseigner `SPOTIFY_CLIENT_ID` et `SPOTIFY_CLIENT_SECRET` (ou flags `-spotify-client-id` / `-spotify-client-secret`). Le backend met en cache le token d acces avec un timeout et le renouvellera automatiquement avant expiration.

## API backend
- Endpoint historique `/api/artists` conserve le format Groupie Tracker par defaut (pas de parametre `source`).  
- Nouvelle reponse unifiee quand `source=groupie` ou `source=all` (ou `external=spotify`) est passe : `[{id, name, image_url, source, creationDate, firstAlbum, members?, genres?, popularity?}]`.
- `source=all` ou `source=spotify` interroge aussi Spotify `/v1/search?type=artist` (limit par defaut 8, parametre `limit` possible). Filtres Spotify appliques: type=artist, nom non vide, au moins une image, popularite faible eliminee, dedoublonnage par nom avec priorite Groupie en cas de collision.
- Endpoint detail `/api/spotify/artist?id=...` : proxie `/v1/artists/{id}` et renvoie `{id,name,image_url,genres,popularity,followers,source}`. Si Spotify n est pas configure, retour 503.

## Frontend
- Page d accueil : toggle source (Groupie, All, Spotify), memes cartes modernes, badge de provenance, placeholder gradient si image absente. Recherche appelle `/api/artists?source=...` et fusionne les resultats.
- Cartes Spotify redirigent vers `/artist-spotify?id=...` qui charge les infos via l endpoint detail. Hero avec badge Spotify, followers, popularite et genres (chips), bouton "Open on Spotify".

## Anti faux resultats
- Requete Spotify force `type=artist` et rejette les entrees sans nom ou sans image. Popularite trop faible ignoree. Aucune requete Spotify n est lancee si le champ `name` est vide (evite les resultats bruit).
