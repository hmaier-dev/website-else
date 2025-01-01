VERSION 0.8

hugo:
    FROM alpine:3.20
    RUN apk add --no-cache hugo

    # Hugo cannot work in root (/)
    WORKDIR tmp
    COPY --dir content static layouts ./
    COPY hugo.toml hugo.toml

    RUN hugo
    RUN ls -la public
    SAVE ARTIFACT ./public AS LOCAL ./public

css:
    FROM alpine:3.20
    RUN apk add --no-cache npm
    RUN npm install -D tailwindcss
    RUN npx tailwindcss init
    COPY tailwind.config.js tailwind.config.js
    RUN npx tailwindcss -i ./src/input.css -o ./src/output.css --watch


build:
    BUILD +hugo
