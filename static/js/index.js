// JavaScript for the Groupie Tracker home page.
// Fetches artists from the Go backend and renders the cards. Includes
// client-side filtering as a fallback when the backend search returns
// an empty array.

const errorMessage = document.getElementById('error-message');
const artistCount = document.getElementById('artist-count');
const sourceInput = document.getElementById('source-input');
const sourceButtons = Array.from(document.querySelectorAll('.source-btn'));
let cachedArtists = [];
let currentSource = sourceInput ? sourceInput.value : 'all';

function parseAPIDate(value) {
  if (!value) return null;
  const cleaned = value.replace(/^\*/, '').trim();
  const parts = cleaned.split('-');
  if (parts.length === 3 && parts[0].length === 4) {
    const date = new Date(cleaned);
    return Number.isNaN(date.getTime()) ? null : date;
  }
  const [day, month, year] = parts;
  const iso = `${year}-${month}-${day}`;
  const date = new Date(iso);
  return Number.isNaN(date.getTime()) ? null : date;
}

function albumYear(value) {
  const parsed = parseAPIDate(value);
  return parsed ? parsed.getFullYear() : '';
}

async function fetchArtists(term) {
  const params = new URLSearchParams();
  if (term) params.set('name', term);
  if (currentSource) params.set('source', currentSource);
  params.set('limit', '8');
  const url = `/api/artists?${params.toString()}`;
  const res = await fetch(url);
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`);
  }
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

function placeholderAvatar(name) {
  const initial = (name || '?').trim().charAt(0).toUpperCase() || '?';
  const wrapper = document.createElement('div');
  wrapper.className = 'artist-placeholder gradient-primary';
  const span = document.createElement('span');
  span.textContent = initial;
  wrapper.appendChild(span);
  return wrapper;
}

function renderArtists(artists) {
  const container = document.getElementById('artists-container');
  container.innerHTML = '';
  artists.forEach((artist) => {
    const imageURL = artist.image_url || artist.image || '';
    const isSpotify = (artist.source || '').toLowerCase() === 'spotify';
    const target = isSpotify ? `/artist-spotify?id=${artist.id}` : `artist?id=${artist.id}`;
    const badgeLabel = isSpotify ? 'Spotify' : 'Groupie Tracker';
    const badgeClass = isSpotify ? 'badge-spotify' : 'badge-groupie';
    const metaText = isSpotify
      ? (artist.popularity ? `Pop. ${artist.popularity}` : 'Spotify')
      : (artist.creationDate || '');

    const card = document.createElement('div');
    card.className = 'artist-card hover-lift';
    card.addEventListener('click', () => {
      window.location.href = target;
    });

    const membersCount = Array.isArray(artist.members) ? artist.members.length : 0;
    const leftDetail = isSpotify
      ? (Array.isArray(artist.genres) && artist.genres.length > 0 ? artist.genres[0] : 'Artiste Spotify')
      : `${membersCount} membres`;
    const rightDetail = isSpotify
      ? (artist.popularity ? `Popularité ${artist.popularity}` : '')
      : (albumYear(artist.firstAlbum) || artist.creationDate || '');

    const media = document.createElement('div');
    media.className = 'artist-media';
    if (imageURL) {
      const img = document.createElement('img');
      img.src = imageURL;
      img.alt = artist.name || '';
      media.appendChild(img);
    } else {
      media.appendChild(placeholderAvatar(artist.name));
    }
    const badge = document.createElement('span');
    badge.className = `artist-source ${badgeClass}`;
    badge.textContent = badgeLabel;
    media.appendChild(badge);
    const meta = document.createElement('div');
    meta.className = 'artist-meta';
    meta.textContent = String(metaText);
    media.appendChild(meta);

    const content = document.createElement('div');
    content.className = 'artist-card-content';
    const title = document.createElement('h3');
    title.textContent = artist.name || '';

    const details = document.createElement('div');
    details.className = 'artist-details-list';
    const leftSpan = document.createElement('span');
    leftSpan.textContent = leftDetail;
    const rightSpan = document.createElement('span');
    rightSpan.textContent = rightDetail;
    details.appendChild(leftSpan);
    details.appendChild(rightSpan);

    const members = document.createElement('div');
    members.className = 'artist-members';
    if (isSpotify) {
      const genres = Array.isArray(artist.genres) ? artist.genres.slice(0, 3) : [];
      if (genres.length > 0) {
        genres.forEach((g) => {
          const tag = document.createElement('span');
          tag.className = 'artist-member-tag';
          tag.textContent = g;
          members.appendChild(tag);
        });
      } else {
        const tag = document.createElement('span');
        tag.className = 'artist-member-tag';
        tag.textContent = 'Sur Spotify';
        members.appendChild(tag);
      }
    } else if (Array.isArray(artist.members)) {
      artist.members.slice(0, 3).forEach((m) => {
        const tag = document.createElement('span');
        tag.className = 'artist-member-tag';
        tag.textContent = m;
        members.appendChild(tag);
      });
      if (membersCount > 3) {
        const tag = document.createElement('span');
        tag.className = 'artist-member-tag';
        tag.textContent = `+${membersCount - 3}`;
        members.appendChild(tag);
      }
    }

    const button = document.createElement('button');
    button.className = 'btn btn-primary';
    button.style.width = '100%';
    button.style.marginTop = '0.5rem';
    button.textContent = isSpotify ? 'Voir sur Spotify' : 'Voir les détails';

    content.appendChild(title);
    content.appendChild(details);
    content.appendChild(members);
    content.appendChild(button);

    card.appendChild(media);
    card.appendChild(content);
    container.appendChild(card);
  });
}

function setArtistCount(count) {
  if (artistCount) {
    artistCount.textContent = String(count);
  }
}

function showError(text) {
  if (!errorMessage) return;
  if (text) {
    errorMessage.style.display = 'block';
    errorMessage.textContent = text;
  } else {
    errorMessage.style.display = 'none';
    errorMessage.textContent = '';
  }
}

document.addEventListener('DOMContentLoaded', async () => {
  const form = document.getElementById('search-form');
  const input = document.getElementById('search-input');
  sourceButtons.forEach((btn) => {
    btn.addEventListener('click', () => {
      const src = btn.getAttribute('data-source');
      currentSource = src;
      showError('');
      if (sourceInput) {
        sourceInput.value = src;
      }
      sourceButtons.forEach((b) => b.classList.toggle('active', b === btn));
      const term = input.value.trim();
      form.dispatchEvent(new Event('submit', { cancelable: true, bubbles: false }));
    });
  });
  try {
    cachedArtists = await fetchArtists();
    renderArtists(cachedArtists);
    setArtistCount(cachedArtists.length);
  } catch (err) {
    console.error(err);
    showError('Échec du chargement des artistes. Veuillez réessayer plus tard.');
  }

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    showError('');
    const term = input.value.trim();
    if (term === '' && currentSource === 'spotify') {
      showError('Ajoutez un terme de recherche pour interroger Spotify.');
      return;
    }
    try {
      const searched = await fetchArtists(term);
      if (searched.length > 0 || term === '') {
        cachedArtists = searched;
      }
      const fallback = cachedArtists.filter((a) => a.name.toLowerCase().includes(term.toLowerCase()));
      const results = searched.length === 0 ? fallback : searched;
      renderArtists(results);
      setArtistCount(results.length);
      if (results.length === 0) {
        showError('Aucun artiste trouvé pour cette recherche.');
      }
    } catch (err) {
      console.error(err);
      const message = currentSource === 'spotify'
        ? 'La recherche Spotify est indisponible. Vérifiez les identifiants.'
        : 'Échec du chargement des artistes. Veuillez réessayer.';
      showError(message);
    }
  });
});


