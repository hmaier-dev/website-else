VERSION 0.8

hugo:
    FROM alpine:3.20
    RUN apk add --no-cache hugo

    # Hugo cannot work in root (/)
    WORKDIR tmp
    COPY --dir content static layouts ./
    COPY hugo.toml hugo.toml

    COPY +css/tailwindcss ./tailwindcss

    RUN hugo
    RUN ls -la public
    SAVE ARTIFACT ./public AS LOCAL ./public

css:
    FROM alpine:3.20
    RUN apk add --no-cache curl
    RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64
    RUN chmod +x tailwindcss-linux-x64
    RUN mv tailwindcss-linux-x64 tailwindcss
    SAVE ARTIFACT tailwindcss

build:
    BUILD +hugo
