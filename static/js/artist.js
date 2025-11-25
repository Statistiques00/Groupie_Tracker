// JavaScript for the Groupie Tracker artist detail page.
// Fetches artist info and relations, then renders list and timeline views.

async function fetchArtists() {
  const res = await fetch('/api/artists');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

async function fetchRelations() {
  const res = await fetch('/api/relation');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  if (Array.isArray(data)) return data;
  if (data && Array.isArray(data.index)) return data.index;
  if (data && typeof data === 'object' && 'id' in data && 'datesLocations' in data) {
    return [data];
  }
  return [];
}

function parseDate(dateStr) {
  const cleaned = dateStr.replace(/^[*]/, '');
  const parts = cleaned.split('-');
  if (parts[0].length === 4) {
    return new Date(cleaned);
  }
  const [day, month, year] = parts;
  return new Date(`${year}-${month}-${day}`);
}

function parseLocationSlug(slug) {
  const parts = slug.split('-');
  const countrySlug = parts.pop() || '';
  const citySlug = parts.join('-');
  const toTitle = (str) => str.replace(/_/g, ' ').split(' ').map((w) => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
  return {
    city: toTitle(citySlug),
    country: toTitle(countrySlug),
  };
}

function buildConcerts(artistId, relations) {
  const relation = relations.find((rel) => Number(rel.id) === Number(artistId));
  if (!relation || !relation.datesLocations) return [];
  const concerts = [];
  for (const [slug, dates] of Object.entries(relation.datesLocations)) {
    const { city, country } = parseLocationSlug(slug);
    (Array.isArray(dates) ? dates : []).forEach((dateStr) => {
      const dt = parseDate(dateStr);
      concerts.push({ location: city, country, date: dt });
    });
  }
  concerts.sort((a, b) => a.date - b.date);
  return concerts.map((c) => ({
    location: c.location,
    country: c.country,
    date: c.date.toISOString().split('T')[0],
  }));
}

function renderArtistHeader(artist) {
  const header = document.getElementById('artist-header');
  header.innerHTML = '';
  const img = document.createElement('img');
  img.src = artist.image || '';
  img.alt = artist.name;
  header.appendChild(img);

  const overlay = document.createElement('div');
  overlay.className = 'artist-header-overlay';
  header.appendChild(overlay);

  const content = document.createElement('div');
  content.className = 'artist-header-content';

  const backBtn = document.createElement('button');
  backBtn.className = 'btn btn-outline';
  backBtn.textContent = 'Retour aux artistes';
  backBtn.addEventListener('click', () => {
    window.location.href = '/';
  });
  backBtn.style.marginBottom = '1rem';

  const title = document.createElement('h1');
  title.className = 'gradient-text font-display';
  title.style.fontSize = '3rem';
  title.style.fontWeight = '700';
  title.textContent = artist.name;

  const info = document.createElement('div');
  info.className = 'artist-info';
  const yearItem = document.createElement('span');
  yearItem.textContent = `Formé en ${artist.creationDate || ''}`;
  const albumItem = document.createElement('span');
  const parsedAlbum = artist.firstAlbum ? parseDate(artist.firstAlbum) : null;
  const albumYear = parsedAlbum && !Number.isNaN(parsedAlbum.getTime()) ? parsedAlbum.getFullYear() : '';
  albumItem.textContent = albumYear ? `Premier album : ${albumYear}` : 'Premier album : n/a';
  const membersItem = document.createElement('span');
  membersItem.textContent = `${Array.isArray(artist.members) ? artist.members.length : 0} membres`;
  info.appendChild(yearItem);
  info.appendChild(albumItem);
  info.appendChild(membersItem);

  content.appendChild(backBtn);
  content.appendChild(title);
  content.appendChild(info);
  header.appendChild(content);
}

function renderMembers(members) {
  const grid = document.getElementById('members-grid');
  grid.innerHTML = '';
  members.forEach((member) => {
    const card = document.createElement('div');
    card.className = 'member-card hover-lift';
    const avatar = document.createElement('div');
    avatar.className = 'member-avatar';
    avatar.textContent = member.charAt(0);
    const name = document.createElement('p');
    name.textContent = member;
    name.style.fontWeight = '600';
    card.appendChild(avatar);
    card.appendChild(name);
    grid.appendChild(card);
  });
}

function renderConcertList(concerts) {
  const listContainer = document.getElementById('concerts-list');
  listContainer.innerHTML = '';
  concerts.forEach((concert) => {
    const item = document.createElement('div');
    item.className = 'concert-item hover-lift';
    const left = document.createElement('div');
    const location = document.createElement('div');
    location.className = 'location';
    location.textContent = concert.location;
    const country = document.createElement('div');
    country.className = 'country';
    country.textContent = concert.country;
    left.appendChild(location);
    left.appendChild(country);
    const date = document.createElement('div');
    date.className = 'date';
    const d = new Date(concert.date);
    const options = { month: 'short', day: 'numeric', year: 'numeric' };
    date.textContent = d.toLocaleDateString(undefined, options);
    item.appendChild(left);
    item.appendChild(date);
    listContainer.appendChild(item);
  });
}

function renderConcertTimeline(concerts) {
  const timelineContainer = document.getElementById('concerts-timeline');
  timelineContainer.innerHTML = '';
  concerts.forEach((concert) => {
    const wrapper = document.createElement('div');
    wrapper.className = 'timeline-item';
    const content = document.createElement('div');
    content.className = 'timeline-content hover-lift';
    const location = document.createElement('div');
    location.className = 'timeline-location';
    location.textContent = concert.location;
    const country = document.createElement('div');
    country.className = 'timeline-country';
    country.textContent = concert.country;
    const date = document.createElement('div');
    date.className = 'timeline-date';
    const d = new Date(concert.date);
    const options = { month: 'long', day: 'numeric', year: 'numeric' };
    date.textContent = d.toLocaleDateString(undefined, options);
    content.appendChild(location);
    content.appendChild(country);
    content.appendChild(date);
    wrapper.appendChild(content);
    timelineContainer.appendChild(wrapper);
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
    const [artists, relations] = await Promise.all([fetchArtists(), fetchRelations()]);
    const artist = artists.find((a) => Number(a.id) === Number(id));
    if (!artist) {
      window.location.href = '/404';
      return;
    }
    renderArtistHeader(artist);
    renderMembers(Array.isArray(artist.members) ? artist.members : []);
    const concerts = buildConcerts(id, relations);
    renderConcertList(concerts);
    renderConcertTimeline(concerts);

    const btnList = document.getElementById('btn-list');
    const btnTimeline = document.getElementById('btn-timeline');
    const listContainer = document.getElementById('concerts-list');
    const timelineContainer = document.getElementById('concerts-timeline');
    btnList.addEventListener('click', () => {
      btnList.classList.add('active');
      btnTimeline.classList.remove('active');
      listContainer.style.display = '';
      timelineContainer.style.display = 'none';
    });
    btnTimeline.addEventListener('click', () => {
      btnTimeline.classList.add('active');
      btnList.classList.remove('active');
      listContainer.style.display = 'none';
      timelineContainer.style.display = '';
    });
  } catch (err) {
    console.error('Error loading artist page:', err);
    window.location.href = '/500';
  }
});
