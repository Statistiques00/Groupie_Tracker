// Locations page: render locations grouped by artist with filters.

const locationsList = document.getElementById('locations-list');
const locationsEmpty = document.getElementById('locations-empty');
const countryFilter = document.getElementById('loc-country');
const artistFilter = document.getElementById('loc-artist');
const cityFilter = document.getElementById('loc-search');

let allLocations = [];

async function fetchLocations() {
  const res = await fetch('/api/locations');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

function populateFilters(locations) {
  const countries = new Set();
  const artists = new Set();
  locations.forEach((loc) => {
    if (loc.country) countries.add(loc.country);
    if (loc.artistName) artists.add(loc.artistName);
  });

  Array.from(countries).sort((a, b) => a.localeCompare(b)).forEach((country) => {
    const opt = document.createElement('option');
    opt.value = country;
    opt.textContent = country;
    countryFilter.appendChild(opt);
  });

  Array.from(artists).sort((a, b) => a.localeCompare(b)).forEach((artist) => {
    const opt = document.createElement('option');
    opt.value = artist;
    opt.textContent = artist;
    artistFilter.appendChild(opt);
  });
}

function renderLocations(locations) {
  locationsList.innerHTML = '';
  if (locations.length === 0) {
    locationsEmpty.style.display = 'block';
    return;
  }
  locationsEmpty.style.display = 'none';

  locations.forEach((loc) => {
    const card = document.createElement('div');
    card.className = 'info-card hover-lift location-card';

    const title = document.createElement('h4');
    title.textContent = `${loc.city}, ${loc.country}`;

    const artist = document.createElement('p');
    artist.className = 'muted';
    artist.textContent = loc.artistName || 'Artiste inconnu';

    const meta = document.createElement('div');
    meta.className = 'location-meta';
    meta.innerHTML = `
      <span class="badge">${loc.eventCount || 0} concerts</span>
      <span class="badge">Artiste #${loc.artistId}</span>
    `;

    card.appendChild(title);
    card.appendChild(artist);
    card.appendChild(meta);
    locationsList.appendChild(card);
  });
}

function applyLocationFilters() {
  const country = countryFilter.value.toLowerCase();
  const artist = artistFilter.value.toLowerCase();
  const city = cityFilter.value.toLowerCase().trim();

  const filtered = allLocations.filter((loc) => {
    if (country && !loc.country.toLowerCase().includes(country)) return false;
    if (artist && !loc.artistName.toLowerCase().includes(artist)) return false;
    if (city && !loc.city.toLowerCase().includes(city)) return false;
    return true;
  });
  renderLocations(filtered);
}

function attachLocationFilters() {
  countryFilter.addEventListener('change', applyLocationFilters);
  artistFilter.addEventListener('change', applyLocationFilters);
  cityFilter.addEventListener('input', applyLocationFilters);
}

document.addEventListener('DOMContentLoaded', async () => {
  try {
    allLocations = await fetchLocations();
    populateFilters(allLocations);
    renderLocations(allLocations);
    attachLocationFilters();
  } catch (err) {
    console.error('Failed to load locations', err);
    locationsEmpty.textContent = 'Échec du chargement des lieux depuis le serveur.';
    locationsEmpty.style.display = 'block';
  }
});
