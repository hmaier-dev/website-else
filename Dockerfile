## This is the base for hmaierdev/hugo-plus-tailwind/
FROM debian:bookworm
RUN apt-get update && apt-get install -y curl
RUN curl -SLO https://github.com/gohugoio/hugo/releases/download/v0.160.1/hugo_0.160.1_linux-amd64.tar.gz
RUN tar -xvzf hugo_0.160.1_linux-amd64.tar.gz
RUN chmod +x hugo
RUN mv hugo /usr/local/bin/hugo

RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.3.5/tailwindcss-linux-x64
RUN chmod +x tailwindcss-linux-x64
RUN mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss
