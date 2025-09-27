While the main part of the website is static html, for receiving booking-request we need an HTTP-endpoint.
This is the purpose this small go-program serves.

Start the binary:
```bash
go build -o ./bin/contact main.go && ./bin/contact
```
And test the endpoints:
```bash
## POST to /contact
## This inserts a new entry into the database
curl -d 'email=test@mail.xyz&name=Mensch&message=Hello to Golang&approval=true' localhost:3000/contact
## GET to /mailbox/all
## gets all messages from database
curl localhost:3000/mailbox/all
```

