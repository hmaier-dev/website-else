document.addEventListener('DOMContentLoaded', () => {
	const nextEvent = '/events/load?out=json&sort=next&head=4'
	const prevEvent = '/events/load?out=json&sort=prev&head=3'

	fetch(nextEvent)
		.then(response => response.json())
		.then(data => {
			if (!Array.isArray(data) || data.length === 0) {
				generateEmptyState()
				return
			}
			generateHero(data[0])
			generateUpcommingEvents(data)
		})
		.catch(error => {
			console.error('Error:', error)
			generateEmptyState()
		})

	fetch(prevEvent)
		.then(response => response.json())
		.then(data => generatePreviousEvents(Array.isArray(data) ? data : []))
		.catch(error => console.error('Error:', error))
})

// Werte aus der Datenbank nie ungeprüft als HTML einsetzen
function esc(value) {
	return String(value ?? '').replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]))
}

function generateEmptyState() {
	document.getElementById('event-name-hero').textContent = 'Bald geht’s weiter!'
	document.getElementById('next-events-cards-wrapper').innerHTML = `
		<p class="card-pop col-span-full p-8 text-lg">Gerade sind keine Termine eingetragen. Schau bald wieder vorbei oder folge mir auf
		<a class="font-semibold underline" href="https://www.instagram.com/einfach_else/">Instagram</a>.</p>`
}

function generateHero(next) {
	document.getElementById('event-name-hero').textContent = next['Name']
	document.getElementById('hero-date').textContent = next['FDate']
	document.getElementById('hero-place').textContent = next['Place']
	document.getElementById('hero-city').textContent = next['City']
	const link = document.getElementById('hero-link')
	if (next['Link']) {
		link.href = next['Link']
	} else {
		link.remove()
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
		const link = event['Link']
		html += `
			<article class="card-pop group flex flex-col p-6 hover:-translate-y-1 hover:shadow-pop-lg">
				<div class="mb-8 flex items-start justify-between">
					<div class="${colors[i % colors.length]} rounded-2xl border-2 border-ink px-4 py-2 text-center">
						<p class="font-headline text-3xl font-extrabold leading-none">${day}</p>
						<p class="mt-1 text-xs font-bold uppercase tracking-widest">${esc(abbr)}</p>
					</div>
					<span class="text-3xl text-lilac-500 transition-transform group-hover:rotate-90" aria-hidden="true">✦</span>
				</div>
				<h3 class="font-headline text-2xl font-extrabold leading-tight">${esc(event['Name'])}</h3>
				<p class="mt-3 flex-1 text-ink/70">${esc(event['Place'])}, ${esc(event['City'])}</p>
				${link ? `<a href="${esc(link)}" class="btn-lilac mt-6 w-full">Details →</a>` : ''}
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
					<h3 class="font-headline text-lg font-bold">${esc(event['Name'])}</h3>
					<p class="text-sm uppercase tracking-widest text-ink/60">${esc(event['Place'])} • ${esc(event['City'])}</p>
				</div>
			</div>`
	}
	document.getElementById('prev-events-cards-wrapper').innerHTML = html
}
