// Relations page logic: fetch artists and relations, then display them with filters.

const relationsList = document.getElementById('relations-list');
const relationsEmpty = document.getElementById('relations-empty');
const relArtistSelect = document.getElementById('rel-artist');
const relLocationInput = document.getElementById('rel-location');

let relationsData = [];
let artistNames = new Map();

async function fetchArtists() {
  const res = await fetch('/api/artists');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  if (!Array.isArray(data)) return [];
  return data.map((a) => ({ id: a.id, name: a.name }));
}

async function fetchRelations() {
  const res = await fetch('/api/relation');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  if (Array.isArray(data)) return data;
  if (data && Array.isArray(data.index)) return data.index;
  return [];
}

function parseSlug(slug) {
  const parts = slug.split('-');
  const countrySlug = parts.pop() || '';
  const citySlug = parts.join('-');
  const toTitle = (str) => str.replace(/_/g, ' ').split(' ').map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
  return { city: toTitle(citySlug), country: toTitle(countrySlug) };
}

function populateArtistFilter(artists) {
  artists.forEach((artist) => {
    const opt = document.createElement('option');
    opt.value = artist.id;
    opt.textContent = artist.name;
    relArtistSelect.appendChild(opt);
  });
}

function buildViews() {
  const locationQuery = relLocationInput.value.toLowerCase().trim();
  const artistFilter = Number(relArtistSelect.value || 0);

  const views = relationsData
    .filter((rel) => (artistFilter ? rel.id === artistFilter : true))
    .map((rel) => {
      const entries = Object.entries(rel.datesLocations || {});
      const expanded = entries.map(([slug, dates]) => ({
        slug,
        name: parseSlug(slug),
        dates: Array.isArray(dates) ? dates : [],
      }));
      return {
        id: rel.id,
        name: artistNames.get(rel.id) || `Artist #${rel.id}`,
        entries: expanded,
      };
    })
    .filter((view) => {
      if (!locationQuery) return true;
      return view.entries.some((entry) => `${entry.name.city} ${entry.name.country}`.toLowerCase().includes(locationQuery));
    });

  return views;
}

function renderRelations(views) {
  relationsList.innerHTML = '';
  if (views.length === 0) {
    relationsEmpty.style.display = 'block';
    return;
  }
  relationsEmpty.style.display = 'none';

  views.forEach((view, index) => {
    const item = document.createElement('div');
    item.className = 'accordion-item glass';

    const header = document.createElement('button');
    header.className = 'accordion-header';
    header.setAttribute('aria-expanded', index === 0 ? 'true' : 'false');
    const headerMain = document.createElement('div');
    const eyebrow = document.createElement('p');
    eyebrow.className = 'eyebrow';
    eyebrow.textContent = `Artiste #${view.id}`;
    const title = document.createElement('h4');
    title.textContent = view.name;
    headerMain.appendChild(eyebrow);
    headerMain.appendChild(title);

    const badgeRow = document.createElement('div');
    badgeRow.className = 'badge-row';
    const locationsBadge = document.createElement('span');
    locationsBadge.className = 'badge';
    locationsBadge.textContent = `${view.entries.length} lieux`;
    const datesBadge = document.createElement('span');
    datesBadge.className = 'badge';
    datesBadge.textContent = `${view.entries.reduce((acc, e) => acc + e.dates.length, 0)} dates`;
    badgeRow.appendChild(locationsBadge);
    badgeRow.appendChild(datesBadge);

    header.appendChild(headerMain);
    header.appendChild(badgeRow);

    const body = document.createElement('div');
    body.className = 'accordion-body';
    body.style.display = index === 0 ? 'block' : 'none';

    const list = document.createElement('div');
    list.className = 'relation-grid';
    view.entries.forEach((entry) => {
      const card = document.createElement('div');
      card.className = 'relation-card';
      const title = document.createElement('div');
      title.className = 'relation-title';
      title.textContent = `${entry.name.city}, ${entry.name.country}`;

      const dates = document.createElement('div');
      dates.className = 'relation-dates';
      entry.dates.forEach((d) => {
        const badge = document.createElement('span');
        badge.className = 'badge badge-soft';
        badge.textContent = d;
        dates.appendChild(badge);
      });

      card.appendChild(title);
      card.appendChild(dates);
      list.appendChild(card);
    });

    body.appendChild(list);

    header.addEventListener('click', () => {
      const expanded = header.getAttribute('aria-expanded') === 'true';
      header.setAttribute('aria-expanded', expanded ? 'false' : 'true');
      body.style.display = expanded ? 'none' : 'block';
    });

    item.appendChild(header);
    item.appendChild(body);
    relationsList.appendChild(item);
  });
}

function attachRelationFilters() {
  relArtistSelect.addEventListener('change', () => {
    renderRelations(buildViews());
  });
  relLocationInput.addEventListener('input', () => {
    renderRelations(buildViews());
  });
}

document.addEventListener('DOMContentLoaded', async () => {
  try {
    const [artists, relations] = await Promise.all([fetchArtists(), fetchRelations()]);
    artistNames = new Map(artists.map((a) => [a.id, a.name]));
    relationsData = relations;
    populateArtistFilter(artists);
    renderRelations(buildViews());
    attachRelationFilters();
  } catch (err) {
    console.error('Failed to load relations', err);
    relationsEmpty.textContent = 'Échec du chargement des relations depuis le serveur.';
    relationsEmpty.style.display = 'block';
  }
});

