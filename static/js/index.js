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
  return `
    <div class="artist-placeholder gradient-primary">
      <span>${initial}</span>
    </div>
  `;
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
    let membersTags = '';
    if (isSpotify) {
      const genres = Array.isArray(artist.genres) ? artist.genres.slice(0, 3) : [];
      if (genres.length > 0) {
        membersTags = genres.map((g) => `<span class="artist-member-tag">${g}</span>`).join('');
      } else {
        membersTags = '<span class="artist-member-tag">Sur Spotify</span>';
      }
    } else if (Array.isArray(artist.members)) {
      const tags = artist.members.slice(0, 3).map((m) => `<span class="artist-member-tag">${m}</span>`);
      membersTags = tags.join('');
      if (membersCount > 3) {
        membersTags += `<span class="artist-member-tag">+${membersCount - 3}</span>`;
      }
    }
    const imageMarkup = imageURL
      ? `<img src="${imageURL}" alt="${artist.name}" />`
      : placeholderAvatar(artist.name);
    const leftDetail = isSpotify
      ? (Array.isArray(artist.genres) && artist.genres.length > 0 ? artist.genres[0] : 'Artiste Spotify')
      : `${membersCount} membres`;
    const rightDetail = isSpotify
      ? (artist.popularity ? `Popularité ${artist.popularity}` : '')
      : (albumYear(artist.firstAlbum) || artist.creationDate || '');
    card.innerHTML = `
      <div class="artist-media">
        ${imageMarkup}
        <span class="artist-source ${badgeClass}">${badgeLabel}</span>
        <div class="artist-meta">${metaText}</div>
      </div>
      <div class="artist-card-content">
        <h3>${artist.name}</h3>
        <div class="artist-details-list">
          <span>${leftDetail}</span>
          <span>${rightDetail}</span>
        </div>
        <div class="artist-members">${membersTags}</div>
        <button class="btn btn-primary" style="width: 100%; margin-top: 0.5rem;">${isSpotify ? 'Voir sur Spotify' : 'Voir les détails'}</button>
      </div>
    `;
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
