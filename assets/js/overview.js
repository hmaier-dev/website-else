document.addEventListener('DOMContentLoaded', () => {

	const nextEvent = '/events/load?out=json&sort=next&head=4'
	const prevEvent = '/events/load?out=json&sort=prev&head=3'

	fetch(nextEvent)
		.then(response => response.json())
		.then(data => {
			generateHero(data[0])
			generateUpcommingEvents(data)
		})
		.catch(error => {
			console.error('Error:', error);
		});

	fetch(prevEvent)
		.then(response => response.json())
		.then(data => {
			generatePreviousEvents(data)
		})
		.catch(error => {
			console.error('Error:', error);
		});

});


function generateHero(next) {

	const name = document.getElementById('event-name-hero')
	const date = document.getElementById('hero-date')
	const place = document.getElementById('hero-place')
	const city = document.getElementById('hero-city')
	const link = document.getElementById('hero-link')

	name.innerHTML = next["Name"]
	date.innerHTML = next["FDate"]
	place.innerHTML = next["Place"]
	city.innerHTML = next["City"]
	link.href = next["Link"]
}

function generateUpcommingEvents(next) {
	html = ""
	const arr = next.slice(1)
	for (const event of arr) {
		const name = event["Name"]
		const date = new Date(event["Date"])
		// Get the three-letter month abbreviation
		const abbr = date.toLocaleString('de-DE', { month: 'short' });
		// Get the day as two digits (e.g., "08")
		const day = date.getDate().toString().padStart(2, '0');

		const place = event["Place"]
		const city = event["City"]
		add = `
			<div class="group bg-white p-8 rounded-[2rem] border border-violet-100 hover:border-primary/30 hover:shadow-2xl hover:shadow-violet-200/50 shadow-md transition-all duration-500">
				<div class="flex justify-between items-start mb-8">
					<div class="bg-violet-50 px-4 py-2 rounded-2xl border border-violet-100">
						<p class="text-xl font-black text-primary leading-none">${day}</p>
						<p class="text-[10px] font-bold text-primary/60 uppercase tracking-widest text-center mt-1">${abbr}</p>
					</div>
				</div>
				<div>
					<p class="text-[10px] font-black text-primary uppercase tracking-[0.2em] mb-3"></p>
					<h4 class="text-2xl font-black text-on-surface leading-tight mb-6 group-hover:text-primary transition-colors">${name}</h4>
					<div class="space-y-2 mb-8">
						<div class="flex items-center gap-3 text-on-surface-variant text-sm">
							<span></span>
							<span></span>
						</div>
						<div class="flex items-center gap-3 text-on-surface-variant text-sm">
							<span></span>
							<span>${place}, ${city}</span>
						</div>
					</div>
					<button class="w-full py-4 bg-violet-50 text-primary rounded-2xl text-xs font-black uppercase tracking-widest hover:bg-primary hover:text-white transition-all">Details</button>
				</div>
			</div>
    `
		html += add
	}
	document.getElementById('next-events-cards-wrapper').innerHTML = html
}

function generatePreviousEvents(prev) {
	html = ""
	for (const event of prev) {
		const name = event["Name"]
		const date = new Date(event["Date"])
		// Get the three-letter month abbreviation
		const abbr = date.toLocaleString('en-US', { month: 'short' });
		// Get the day as two digits (e.g., "08")
		const day = date.getDate().toString().padStart(2, '0');
		const place = event["Place"]
		const city = event["City"]
		add = `
			<div class="flex flex-col md:flex-row items-center justify-between p-6 bg-surface-container-lowest rounded-2xl border border-outline-variant/30 opacity-70 grayscale hover:grayscale-0 hover:opacity-100 transition-all group">
				<div class="flex flex-col md:flex-row items-center gap-8 text-center md:text-left">
					<div class="w-16 text-center">
						<p class="text-lg font-black text-on-surface-variant leading-none">${day}</p>
						<p class="text-[10px] font-bold uppercase text-outline">${abbr}</p>
					</div>
					<div class="h-10 w-[1px] bg-outline-variant hidden md:block"></div>
					<div>
						<h5 class="text-lg font-bold text-on-surface group-hover:text-primary">${name}</h5>
						<p class="text-xs text-on-surface-variant uppercase tracking-widest">${place} • ${city}</p>
					</div>
				</div>
				<div class="mt-4 md:mt-0">
					<a href="/events/all" class="text-[10px] font-black uppercase tracking-widest text-primary flex items-center gap-2">Alle Events anzeigen</a>
				</div>
			</div>
    `
		html += add
	}
	document.getElementById('prev-events-cards-wrapper').innerHTML = html

}


