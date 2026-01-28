While the main part of the website is static html, for receiving booking-request we need an HTTP-endpoint.
This is the purpose this small go-program serves.

Start the binary:
```bash
go build -x -v -p 4 -o ./bin/backend main.go && ./bin/backend -db=sqlite.db
```
And test the endpoints:

## Mailbox
```bash
## POST to /contact
## This inserts a new entry into the database
curl -d 'email=test@mail.xyz&name=Mensch&message=Hello to Golang&approval=true' localhost:3000/contact
## GET to /mailbox/all
## gets all messages from database
curl localhost:3000/mailbox/all
```

## Events
```bash
## POST to /events/manage/create
## This inserts a new entry into the database
curl -d 'datum=2026-02-22&event=LateNightShow&ort=Moritzbastei&stadt=Leipzig' localhost:3000/events/manage/api/create
## POST to /events/manage/api/delete
curl -d 'datum=22.02.2026&event=LateNightShow&ort=Moritzbastei&stadt=Leipzig' localhost:3000/events/manage/api/delete
## GET to /events/manage/load
## gets all messages from database
curl localhost:3000/events/load?sort=desc&head=3&json=true
## Highlight the next comming event desc/asc
curl "localhost:3000/events/load?sort=asc&upcomming=true" | jq '.'
curl "localhost:3000/events/load?sort=desc&upcomming=true" | jq '.'
## Set a seperator for the year. Used in the html template
curl "localhost:3000/events/load?sort=desc&sep=true&out=html"
## POST to /events/manage/create/bulk
## uploads a yaml-file with multiple events
curl -X POST -F "yaml=@events.yml" http://localhost:3000/events/manage/api/create/bulk
```
