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
    RUN apk add --no-cache curl
    RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.17/tailwindcss-linux-x64
    RUN chmod +x tailwindcss-linux-x64
    RUN mv tailwindcss-linux-x64 tailwindcss
    COPY --dir static layouts ./
    COPY tailwind.config.js tailwind.config.js

    # Compile and minify your CSS for production
    RUN ./tailwindcss -i static/css/input.css -o output.css --minify
    RUN cat output.css


build:
    BUILD +hugo
