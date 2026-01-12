// JavaScript for the Spotify artist detail page.
// Fetches artist info from the backend proxy and renders stats + genres.

function placeholderAvatar(name) {
  const initial = (name || '?').trim().charAt(0).toUpperCase() || '?';
  const wrapper = document.createElement('div');
  wrapper.className = 'artist-placeholder gradient-primary';
  const span = document.createElement('span');
  span.textContent = initial;
  wrapper.appendChild(span);
  return wrapper;
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
  if (!header) return;
  header.innerHTML = '';

  if (artist.image_url) {
    const img = document.createElement('img');
    img.src = artist.image_url;
    img.alt = artist.name || '';
    header.appendChild(img);
  } else {
    header.appendChild(placeholderAvatar(artist.name));
  }

  const overlay = document.createElement('div');
  overlay.className = 'artist-header-overlay';
  header.appendChild(overlay);

  const content = document.createElement('div');
  content.id = 'spotify-hero-content';
  content.className = 'artist-header-content';

  const badge = document.createElement('span');
  badge.className = 'artist-source badge-spotify';
  badge.textContent = 'Spotify';

  const title = document.createElement('h1');
  title.className = 'gradient-text font-display';
  title.style.fontSize = '3rem';
  title.style.fontWeight = '700';
  title.textContent = artist.name || '';

  const followerText = artist.followers ? `${Number(artist.followers).toLocaleString()} abonnés` : 'Artiste Spotify';
  const info = document.createElement('div');
  info.className = 'artist-info';
  const followers = document.createElement('span');
  followers.textContent = followerText;
  const popularity = document.createElement('span');
  popularity.textContent = artist.popularity ? `Popularité ${artist.popularity}` : '';
  info.appendChild(followers);
  info.appendChild(popularity);

  const pillRow = document.createElement('div');
  pillRow.className = 'pill-row';
  pillRow.id = 'hero-genres';

  const actionWrap = document.createElement('div');
  actionWrap.style.marginTop = '1rem';
  const link = document.createElement('a');
  link.className = 'btn btn-primary';
  link.href = `https://open.spotify.com/artist/${artist.id}`;
  link.target = '_blank';
  link.rel = 'noreferrer';
  link.textContent = 'Ouvrir sur Spotify';
  actionWrap.appendChild(link);

  content.appendChild(badge);
  content.appendChild(title);
  content.appendChild(info);
  content.appendChild(pillRow);
  content.appendChild(actionWrap);

  header.appendChild(content);

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
    const label = document.createElement('div');
    label.className = 'stat-label';
    label.textContent = card.label;
    const value = document.createElement('div');
    value.className = 'stat-value';
    value.textContent = card.value;
    const helper = document.createElement('div');
    helper.className = 'stat-label';
    helper.textContent = card.helper;
    el.appendChild(label);
    el.appendChild(value);
    el.appendChild(helper);
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





