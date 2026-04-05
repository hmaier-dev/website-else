document.addEventListener('DOMContentLoaded', () => {

  // const nextEvent = 'events/load?sort=next&head=4'
  const nextMock = [{ "Name": "Queer Slam", "Date": "2026-04-08T00:00:00Z", "Doors": "2026-04-08T00:00:00Z", "Begin": "2026-04-08T00:00:00Z", "Place": "tba", "City": "Bochum", "Coordinates": "", "Link": null, "Hash": "2JWVQZTcqJBBFRNR5bFEzDV8VP6nRrfJi9Q1z3EUAzVj", "Separator": null, "Upcomming": null }, { "Name": "Poetry Slam", "Date": "2026-04-09T00:00:00Z", "Doors": "2026-04-09T00:00:00Z", "Begin": "2026-04-09T00:00:00Z", "Place": "Aaltra", "City": "Chemnitz", "Coordinates": "", "Link": null, "Hash": "5npXGmQ8EAmQrPPYCwMGySyoxgRwBcTw3Mpkx5fnkvxT", "Separator": null, "Upcomming": null }, { "Name": "PowerPoint Karaoke", "Date": "2026-04-14T00:00:00Z", "Doors": "2026-04-14T00:00:00Z", "Begin": "2026-04-14T00:00:00Z", "Place": "tba", "City": "Würzburg", "Coordinates": "", "Link": "https://ppkfun.com/", "Hash": "3zMhybn3hbAFBQY5iijXg8uYkqwW84KeYTqn3V7r1rgY", "Separator": null, "Upcomming": null }, { "Name": "Queer Slam", "Date": "2026-04-15T00:00:00Z", "Doors": "2026-04-15T00:00:00Z", "Begin": "2026-04-15T00:00:00Z", "Place": "SO36", "City": "Berlin", "Coordinates": "", "Link": null, "Hash": "CveakfHLDuwNdJ9cJAZ9dEDLybhaGFLoQLK6aQdqmHRN", "Separator": null, "Upcomming": null }]
  // const prevEvent = 'events/load?sort=prev&head=3'
  const prevMock = [{"Name":"„LE TALK\" (Eintritt for free)","Date":"2026-04-01T00:00:00Z","Doors":"2026-04-01T00:00:00Z","Begin":"2026-04-01T00:00:00Z","Place":"Ilses Erika","City":"Leipzig","Coordinates":"","Link":null,"Hash":"7fvMfv8AhdfsQ8HERjRk2TGJjgdapjqNmuShjJEdSXoj","Separator":null,"Upcomming":null},{"Name":"„Weniger ist mehr“ Erotik, Buslesque, Drag und Poesie","Date":"2026-03-31T00:00:00Z","Doors":"2026-03-31T00:00:00Z","Begin":"2026-03-31T00:00:00Z","Place":"Werk 2","City":"Leipzig","Coordinates":"","Link":null,"Hash":"51RJrL3DYXucRvtJNvwMgPg9nq5B9hXefmyWjxuXpMX4","Separator":null,"Upcomming":null},{"Name":"Schreibworkshop","Date":"2026-03-25T00:00:00Z","Doors":"2026-03-25T00:00:00Z","Begin":"2026-03-25T00:00:00Z","Place":"tba","City":"Zwickau","Coordinates":"","Link":null,"Hash":"9uGHPjDrWoHX5MwT9t4Ec77yDaG7HyK6UEAkkJWhJHD4","Separator":null,"Upcomming":null}]

  generateHero(nextMock[0])
  generateUpcommingEvents(nextMock)
  generatePreviousEvents(prevMock)

});


function generateHero(next) {

  const name = document.getElementById('event-name-hero')
  const date = document.getElementById('hero-date')
  const place = document.getElementById('hero-place')
  const city = document.getElementById('hero-city')

  name.innerHTML = next["Name"]
  date.innerHTML = next["Date"]
  place.innerHTML = next["Place"]
  city.innerHTML = next["City"]
}

function generateUpcommingEvents(next) {
  html = ""
  const arr = next.slice(1)
  for (const event of arr) {
    const name = event["Name"]
    const date = new Date(event["Date"])
		// Get the three-letter month abbreviation
		const abbr = date.toLocaleString('en-US', { month: 'short' });
		// Get the day as two digits (e.g., "08")
		const day = date.getDate().toString().padStart(2, '0');

    const place = event["Place"]
    const city = event["City"]
    add = `
			<div class="group bg-white p-8 rounded-[2rem] border border-violet-100 hover:border-primary/30 hover:shadow-2xl hover:shadow-violet-200/50 transition-all duration-500">
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

function generatePreviousEvents(prev){
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
					<button class="text-[10px] font-black uppercase tracking-widest text-primary flex items-center gap-2">Alle Events anzeigen</button>
				</div>
			</div>
    `
    html += add
  }
  document.getElementById('prev-events-cards-wrapper').innerHTML = html
	
}


