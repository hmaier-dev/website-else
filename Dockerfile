## This is the base for hmaierdev/hugo-plus-tailwind/
FROM debian:bullseye
RUN apt-get update && apt-get install -y curl 
RUN curl -SLO https://github.com/gohugoio/hugo/releases/download/v0.154.5/hugo_0.154.5_linux-amd64.tar.gz
RUN tar -xvzf hugo_0.154.5_linux-amd64.tar.gz
RUN chmod +x hugo
RUN mv hugo /usr/local/bin/hugo

RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64
RUN chmod +x tailwindcss-linux-x64
RUN mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss
