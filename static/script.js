// Client-side search: interroge /search?q= et met à jour la liste d'artistes
(function() {
	// Theme handling
	const themeToggle = document.getElementById('theme-toggle');
	const root = document.documentElement;
	const saved = localStorage.getItem('theme');
	if (saved === 'dark') root.classList.add('dark');

	if (themeToggle) {
		themeToggle.addEventListener('click', () => {
			const isDark = root.classList.toggle('dark');
			localStorage.setItem('theme', isDark ? 'dark' : 'light');
			themeToggle.textContent = isDark ? '☀️' : '🌙';
		});
		// set initial icon
		themeToggle.textContent = root.classList.contains('dark') ? '☀️' : '🌙';
	}

	// Search
	const input = document.getElementById('search');
	const form = document.getElementById('search-form');
	const container = document.getElementById('artists');
	if (!input || !container) return;

	let timer = null;
	function render(list) {
		if (!Array.isArray(list)) return;
		if (list.length === 0) {
			container.innerHTML = '<p>Aucun artiste trouvé.</p>';
			return;
		}
		const html = list.map(a => `
			<article class="card">
				<a href="/artist?id=${encodeURIComponent(a.id)}">
					<img src="${a.image}" alt="${escapeHtml(a.name)}">
					<h3>${escapeHtml(a.name)}</h3>
				</a>
			</article>
		`).join('');
		container.innerHTML = html;
	}

	function escapeHtml(s) {
		return String(s)
			.replace(/&/g, '&amp;')
			.replace(/</g, '&lt;')
			.replace(/>/g, '&gt;')
			.replace(/"/g, '&quot;')
			.replace(/'/g, '&#39;');
	}

	function doSearch() {
		const q = input.value.trim();
		fetch('/search?q=' + encodeURIComponent(q))
			.then(r => {
				if (!r.ok) throw new Error('network');
				return r.json();
			})
			.then(render)
			.catch(err => {
				console.error('search error', err);
			});
	}

	input.addEventListener('input', function() {
		clearTimeout(timer);
		timer = setTimeout(doSearch, 250);
	});

	if (form) {
		form.addEventListener('submit', function(e) {
			e.preventDefault();
			clearTimeout(timer);
			doSearch();
		});
	}

})();
