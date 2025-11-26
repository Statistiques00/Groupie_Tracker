// JavaScript for the Spotify artist detail page.
// Fetches artist info from the backend proxy and renders stats + genres.

function placeholderAvatar(name) {
  const initial = (name || '?').trim().charAt(0).toUpperCase() || '?';
  return `
    <div class="artist-placeholder gradient-primary">
      <span>${initial}</span>
    </div>
  `;
}

async function fetchSpotifyArtist(id) {
  const res = await fetch(`/api/spotify/artist?id=${encodeURIComponent(id)}`);
  if (res.status === 404) {
    window.location.href = '/404';
    return null;
  }
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  return res.json();
}

function renderHeader(artist) {
  const header = document.getElementById('spotify-header');
  const hero = document.getElementById('spotify-hero-content');
  if (!header || !hero) return;
  const imageMarkup = artist.image_url
    ? `<img src="${artist.image_url}" alt="${artist.name}" />`
    : placeholderAvatar(artist.name);
  const followerText = artist.followers ? `${Number(artist.followers).toLocaleString()} abonnés` : 'Artiste Spotify';

  header.innerHTML = `
    ${imageMarkup}
    <div class="artist-header-overlay"></div>
    <div id="spotify-hero-content" class="artist-header-content">
      <span class="artist-source badge-spotify">Spotify</span>
      <h1 class="gradient-text font-display" style="font-size: 3rem; font-weight: 700;">${artist.name}</h1>
      <div class="artist-info">
        <span>${followerText}</span>
        <span>${artist.popularity ? `Popularité ${artist.popularity}` : ''}</span>
      </div>
      <div class="pill-row" id="hero-genres"></div>
      <div style="margin-top: 1rem;">
        <a class="btn btn-primary" href="https://open.spotify.com/artist/${artist.id}" target="_blank" rel="noreferrer">Ouvrir sur Spotify</a>
      </div>
    </div>
  `;

  const heroGenres = document.getElementById('hero-genres');
  if (heroGenres) {
    heroGenres.innerHTML = '';
    const genres = Array.isArray(artist.genres) ? artist.genres.slice(0, 4) : [];
    if (genres.length === 0) {
      const span = document.createElement('span');
      span.className = 'pill';
      span.textContent = 'Aucun genre indiqué';
      heroGenres.appendChild(span);
    } else {
      genres.forEach((g) => {
        const pill = document.createElement('span');
        pill.className = 'pill';
        pill.textContent = g;
        heroGenres.appendChild(pill);
      });
    }
  }
}

function renderStats(artist) {
  const container = document.getElementById('spotify-stats');
  if (!container) return;
  container.innerHTML = '';
  const cards = [
    {
      label: 'Abonnés',
      value: artist.followers ? Number(artist.followers).toLocaleString() : 'N/A',
      helper: "Nombre d'abonnés en direct depuis Spotify",
    },
    {
      label: 'Popularité',
      value: artist.popularity || 'N/A',
      helper: 'Popularité Spotify (0-100)',
    },
    {
      label: 'Genres',
      value: Array.isArray(artist.genres) && artist.genres.length > 0 ? artist.genres.slice(0, 2).join(' • ') : 'Aucun genre',
      helper: 'Genres principaux détectés par Spotify',
    },
  ];
  cards.forEach((card) => {
    const el = document.createElement('div');
    el.className = 'stat-card hover-lift';
    el.innerHTML = `
      <div class="stat-label">${card.label}</div>
      <div class="stat-value">${card.value}</div>
      <div class="stat-label">${card.helper}</div>
    `;
    container.appendChild(el);
  });
}

function renderGenres(artist) {
  const container = document.getElementById('spotify-genres');
  if (!container) return;
  container.innerHTML = '';
  const genres = Array.isArray(artist.genres) ? artist.genres : [];
  if (genres.length === 0) {
    const empty = document.createElement('span');
    empty.className = 'muted';
    empty.textContent = 'Aucun genre disponible pour cet artiste.';
    container.appendChild(empty);
    return;
  }
  genres.forEach((genre) => {
    const pill = document.createElement('span');
    pill.className = 'pill';
    pill.textContent = genre;
    container.appendChild(pill);
  });
}

document.addEventListener('DOMContentLoaded', async () => {
  const params = new URLSearchParams(window.location.search);
  const id = params.get('id');
  if (!id) {
    window.location.href = '/404';
    return;
  }
  try {
    const artist = await fetchSpotifyArtist(id);
    if (!artist) return;
    const openLink = document.getElementById('open-spotify-link');
    if (openLink) {
      openLink.href = `https://open.spotify.com/artist/${artist.id}`;
    }
    renderHeader(artist);
    renderStats(artist);
    renderGenres(artist);
  } catch (err) {
    console.error('Error loading Spotify artist page:', err);
    window.location.href = '/500';
  }
});
