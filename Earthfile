VERSION 0.8

hugo:
    FROM debian:bullseye
    RUN apt-get update && apt-get install -y curl 
    RUN curl -SLO https://github.com/gohugoio/hugo/releases/download/v0.140.2/hugo_0.140.2_linux-amd64.tar.gz
    RUN tar -xvzf hugo_0.140.2_linux-amd64.tar.gz
    RUN chmod +x hugo
    RUN mv hugo /usr/local/bin/hugo

    COPY +css/tailwindcss ./tailwindcss
    RUN chmod +x tailwindcss
    RUN mv tailwindcss /usr/local/bin/tailwindcss

    # Hugo cannot work in root (/)
    WORKDIR tmp
    COPY --dir content assets layouts ./
    COPY hugo.toml hugo.toml
    COPY tailwind.config.js tailwind.config.js

    RUN hugo
    RUN ls -la public
    SAVE ARTIFACT ./public AS LOCAL ./public

css:
    FROM debian:bullseye
    RUN apt-get update && apt-get install -y curl
    RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64
    RUN chmod +x tailwindcss-linux-x64
    RUN mv tailwindcss-linux-x64 tailwindcss
    RUN ./tailwindcss --help
    SAVE ARTIFACT tailwindcss

build:
    BUILD +hugo
