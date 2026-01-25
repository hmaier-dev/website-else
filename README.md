## Develop locally
If you want to develop the website further, do it locally. You can start a server with hugo:
```bash
hugo server --bind 0.0.0.0 --baseURL http://$(ip addr | grep -w "inet 192.168" | awk '{print $2}' | cut -d/ -f1):1313/
```
### Run the backend
To receive contact-request and to load the events, starting the backed is required.
```bash
cd ./backend
go build -x -v -p 4 -o ./bin/backend main.go && ./bin/backend -db=sqlite.db
```

## Building a new image
If you do any changes to the hugo project, run the following command:
```bash
earthly --push +run
```
If you have a valid auth-token for the container registry, the `--push`-flag will do push the image upstream.
Beware: no tag is set and the new image will overwrite the old on.

## Backend
Even though the website is static html, there is a endpoint at `/mailbox/contact` which receives user input.

