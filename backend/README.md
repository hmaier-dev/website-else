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
curl -d 'datum=2026-01-23&event=LateNightShow&ort=Moritzbastei&stadt=Leipzig&link=http%3A%2F%2Fwww.ort.de' localhost:3000/events/manage/create
## GET to /events/manage/load
## gets all messages from database
curl localhost/events/load
```
