// Texte aus content/events/overview.md (Front Matter)
const texts = Object.assign(
	{ noLink: 'Infos folgen bald', last: 'Zuletzt auf der Bühne' },
	JSON.parse(document.getElementById('overview-texts')?.textContent || '{}'),
)

const loadJSON = url => fetch(url).then(r => {
	if (!r.ok) throw new Error(`${url}: HTTP ${r.status}`)
	return r.json()
}).then(d => (Array.isArray(d) ? d : []))

document.addEventListener('DOMContentLoaded', async () => {
	const [next, prev] = await Promise.allSettled([
		loadJSON('/events/load?out=json&sort=next&head=4'),
		loadJSON('/events/load?out=json&sort=prev&head=3'),
	])
	if (next.status === 'rejected') {
		console.error(next.reason)
		generateError()
		return
	}
	const upcoming = next.value
	const past = prev.status === 'fulfilled' ? prev.value : []

	if (upcoming.length > 0) {
		generateHero(upcoming[0])
		generateUpcommingEvents(upcoming)
	} else if (past.length > 0) {
		// Kein Termin in Sicht: den letzten Termin oben zeigen statt einer leeren Seite
		document.getElementById('hero-kicker').textContent = texts.last
		generateHero(past[0])
		generateEmptyUpcoming()
	} else {
		generateEmptyUpcoming()
		document.getElementById('event-name-hero').textContent = 'Neue Termine folgen bald!'
	}
	generatePreviousEvents(past)
})

// Werte aus der Datenbank nie ungeprüft als HTML einsetzen
function esc(value) {
	return String(value ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]))
}

// Nur echte http(s)-Links verwenden; null, "" oder "null" gelten als "kein Link"
function validLink(link) {
	if (!link || link === 'null') return null
	try {
		const u = new URL(link, location.href)
		return u.protocol === 'http:' || u.protocol === 'https:' ? u.href : null
	} catch (_) {
		return null
	}
}

function generateError() {
	document.getElementById('event-name-hero').textContent = 'Termine konnten gerade nicht geladen werden.'
	document.getElementById('next-events-cards-wrapper').innerHTML = `
		<p class="card-pop col-span-full p-8 text-lg">Bitte versuch es gleich nochmal oder schau auf
		<a class="font-semibold underline" href="https://www.instagram.com/einfach_else/">Instagram</a> vorbei.</p>`
}

function generateEmptyUpcoming() {
	document.getElementById('next-events-cards-wrapper').innerHTML = `
		<p class="card-pop col-span-full p-8 text-lg">Gerade sind keine weiteren Termine eingetragen. Schau bald wieder vorbei oder folge mir auf
		<a class="font-semibold underline" href="https://www.instagram.com/einfach_else/">Instagram</a>.</p>`
}

function generateHero(event) {
	document.getElementById('event-name-hero').textContent = event['Name']
	document.getElementById('hero-date').textContent = event['FDate']
	document.getElementById('hero-place').textContent = event['Place']
	document.getElementById('hero-city').textContent = event['City']
	const link = validLink(event['Link'])
	const a = document.getElementById('hero-link')
	if (link) {
		a.href = link
		a.hidden = false
	} else {
		const nolink = document.getElementById('hero-nolink')
		if (event['Note']) nolink.textContent = event['Note'] // z. B. "Eintritt: Spendenbasis"
		nolink.hidden = false
	}
	const details = document.getElementById('hero-details')
	details.classList.remove('hidden')
	details.classList.add('flex')
}

function generateUpcommingEvents(next) {
	const colors = ['bg-lime', 'bg-lilac-300', 'bg-white']
	let html = ''
	next.slice(1).forEach((event, i) => {
		const date = new Date(event['Date'])
		const abbr = date.toLocaleString('de-DE', { month: 'short' }).replace('.', '')
		const day = date.getDate().toString().padStart(2, '0')
		const link = validLink(event['Link'])
		html += `
			<article class="card-pop group flex flex-col p-6 hover:-translate-y-1 hover:shadow-pop-lg">
				<div class="mb-8 flex items-start justify-between">
					<div class="${colors[i % colors.length]} rounded-2xl border-2 border-ink px-4 py-2 text-center">
						<p class="font-headline text-3xl font-extrabold leading-none">${day}</p>
						<p class="mt-1 text-xs font-bold uppercase tracking-widest">${esc(abbr)}</p>
					</div>
					<span class="text-3xl text-lilac-500 transition-transform group-hover:rotate-90" aria-hidden="true">✦</span>
				</div>
				<h3 lang="de" class="hyphens-auto break-words font-headline text-2xl font-extrabold leading-tight">${esc(event['Name'])}</h3>
				<p class="mt-3 flex-1 text-ink/70">${esc(event['Place'])}, ${esc(event['City'])}</p>
				${link
					? `<a href="${esc(link)}" class="btn-lilac mt-6 w-full">Mehr Infos →</a>`
					: `<p class="mt-6 rounded-full border-2 border-dashed border-ink/40 px-4 py-2.5 text-center font-headline font-bold text-ink/60">${esc(event['Note'] || texts.noLink)}</p>`}
			</article>`
	})
	document.getElementById('next-events-cards-wrapper').innerHTML = html
}

function generatePreviousEvents(prev) {
	let html = ''
	for (const event of prev) {
		const date = new Date(event['Date'])
		const abbr = date.toLocaleString('de-DE', { month: 'short' }).replace('.', '')
		const day = date.getDate().toString().padStart(2, '0')
		html += `
			<div class="flex items-center gap-6 rounded-2xl border-2 border-ink/15 bg-white p-5 opacity-70 transition-all hover:border-ink hover:opacity-100">
				<div class="w-14 shrink-0 text-center">
					<p class="font-headline text-2xl font-extrabold leading-none">${day}</p>
					<p class="text-xs font-bold uppercase text-ink/60">${esc(abbr)}</p>
				</div>
				<div>
					<h3 lang="de" class="hyphens-auto break-words font-headline text-lg font-bold">${esc(event['Name'])}</h3>
					<p class="text-sm uppercase tracking-widest text-ink/60">${esc(event['Place'])} • ${esc(event['City'])}</p>
				</div>
			</div>`
	}
	document.getElementById('prev-events-cards-wrapper').innerHTML = html
}
