// Timeline page logic: fetch events from /api/events and provide filters.

const timelineContainer = document.getElementById('timeline-events');
const eventsEmpty = document.getElementById('events-empty');
const yearSelect = document.getElementById('filter-year');
const countrySelect = document.getElementById('filter-country');
const searchInput = document.getElementById('filter-search');

let allEvents = [];

async function fetchEvents() {
  const res = await fetch('/api/events');
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  const data = await res.json();
  return Array.isArray(data) ? data : [];
}

function populateFilters(events) {
  const years = new Set();
  const countries = new Set();
  events.forEach((ev) => {
    const year = new Date(ev.date).getFullYear();
    if (!Number.isNaN(year)) years.add(year);
    if (ev.country) countries.add(ev.country);
  });

  const yearOptions = Array.from(years).sort((a, b) => a - b);
  yearOptions.forEach((year) => {
    const opt = document.createElement('option');
    opt.value = String(year);
    opt.textContent = String(year);
    yearSelect.appendChild(opt);
  });

  const countryOptions = Array.from(countries).sort((a, b) => a.localeCompare(b));
  countryOptions.forEach((country) => {
    const opt = document.createElement('option');
    opt.value = country;
    opt.textContent = country;
    countrySelect.appendChild(opt);
  });
}

function renderEvents(events) {
  timelineContainer.innerHTML = '';
  if (events.length === 0) {
    eventsEmpty.style.display = 'block';
    return;
  }
  eventsEmpty.style.display = 'none';
  events.forEach((ev) => {
    const item = document.createElement('div');
    item.className = 'timeline-item';
    const content = document.createElement('div');
    content.className = 'timeline-content hover-lift';

    const heading = document.createElement('div');
    heading.className = 'timeline-location';
    heading.textContent = `${ev.city} - ${ev.country}`;

    const artist = document.createElement('div');
    artist.className = 'timeline-country';
    artist.textContent = ev.artistName;

    const date = document.createElement('div');
    date.className = 'timeline-date';
    const d = new Date(ev.date);
    const options = { month: 'long', day: 'numeric', year: 'numeric' };
    date.textContent = d.toLocaleDateString(undefined, options);

    content.appendChild(heading);
    content.appendChild(artist);
    content.appendChild(date);
    item.appendChild(content);
    timelineContainer.appendChild(item);
  });
}

function applyFilters() {
  const year = yearSelect.value ? Number(yearSelect.value) : 0;
  const country = countrySelect.value.toLowerCase();
  const query = searchInput.value.toLowerCase().trim();

  const filtered = allEvents.filter((ev) => {
    if (year && new Date(ev.date).getFullYear() !== year) return false;
    if (country && !ev.country.toLowerCase().includes(country)) return false;
    if (query && !`${ev.artistName} ${ev.city}`.toLowerCase().includes(query)) return false;
    return true;
  });
  renderEvents(filtered);
}

function attachFilters() {
  yearSelect.addEventListener('change', applyFilters);
  countrySelect.addEventListener('change', applyFilters);
  searchInput.addEventListener('input', applyFilters);
}

document.addEventListener('DOMContentLoaded', async () => {
  try {
    allEvents = await fetchEvents();
    populateFilters(allEvents);
    renderEvents(allEvents);
    attachFilters();
  } catch (err) {
    console.error('Failed to load events', err);
    eventsEmpty.textContent = 'Échec du chargement des événements depuis le serveur.';
    eventsEmpty.style.display = 'block';
  }
});

