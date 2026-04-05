document.addEventListener('DOMContentLoaded', () => {

  // const nextEvent = 'events/load?sort=next&head=4'
  const nextMock = '[{"Name":"Queer Slam","Date":"2026-04-08T00:00:00Z","Doors":"2026-04-08T00:00:00Z","Begin":"2026-04-08T00:00:00Z","Place":"tba","City":"Bochum","Coordinates":"","Link":null,"Hash":"2JWVQZTcqJBBFRNR5bFEzDV8VP6nRrfJi9Q1z3EUAzVj","Separator":null,"Upcomming":null},{"Name":"Poetry Slam","Date":"2026-04-09T00:00:00Z","Doors":"2026-04-09T00:00:00Z","Begin":"2026-04-09T00:00:00Z","Place":"Aaltra","City":"Chemnitz","Coordinates":"","Link":null,"Hash":"5npXGmQ8EAmQrPPYCwMGySyoxgRwBcTw3Mpkx5fnkvxT","Separator":null,"Upcomming":null},{"Name":"PowerPoint Karaoke","Date":"2026-04-14T00:00:00Z","Doors":"2026-04-14T00:00:00Z","Begin":"2026-04-14T00:00:00Z","Place":"tba","City":"Würzburg","Coordinates":"","Link":"https://ppkfun.com/","Hash":"3zMhybn3hbAFBQY5iijXg8uYkqwW84KeYTqn3V7r1rgY","Separator":null,"Upcomming":null},{"Name":"Queer Slam","Date":"2026-04-15T00:00:00Z","Doors":"2026-04-15T00:00:00Z","Begin":"2026-04-15T00:00:00Z","Place":"SO36","City":"Berlin","Coordinates":"","Link":null,"Hash":"CveakfHLDuwNdJ9cJAZ9dEDLybhaGFLoQLK6aQdqmHRN","Separator":null,"Upcomming":null}]'
  // const prevEvent = 'events/load?sort=prev&head=3'
  const prevMock = '[{"Name":"„LE TALK\" (Eintritt for free)","Date":"2026-04-01T00:00:00Z","Doors":"2026-04-01T00:00:00Z","Begin":"2026-04-01T00:00:00Z","Place":"Ilses Erika","City":"Leipzig","Coordinates":"","Link":null,"Hash":"7fvMfv8AhdfsQ8HERjRk2TGJjgdapjqNmuShjJEdSXoj","Separator":null,"Upcomming":null},{"Name":"„Weniger ist mehr“ Erotik, Buslesque, Drag und Poesie","Date":"2026-03-31T00:00:00Z","Doors":"2026-03-31T00:00:00Z","Begin":"2026-03-31T00:00:00Z","Place":"Werk 2","City":"Leipzig","Coordinates":"","Link":null,"Hash":"51RJrL3DYXucRvtJNvwMgPg9nq5B9hXefmyWjxuXpMX4","Separator":null,"Upcomming":null},{"Name":"Schreibworkshop","Date":"2026-03-25T00:00:00Z","Doors":"2026-03-25T00:00:00Z","Begin":"2026-03-25T00:00:00Z","Place":"tba","City":"Zwickau","Coordinates":"","Link":null,"Hash":"9uGHPjDrWoHX5MwT9t4Ec77yDaG7HyK6UEAkkJWhJHD4","Separator":null,"Upcomming":null}'

  generateHero()

});


function generateHero(){
  const eventName = document.getElementById('event-name-hero')
  eventName.innerHTML = "Die Else Late Night Show"
  console.log("Hello")
}
