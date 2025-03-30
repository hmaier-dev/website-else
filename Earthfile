VERSION 0.8

build-html:
  FROM hmaierdev/hugo-plus-tailwind

  # Hugo cannot work in root (/)
  WORKDIR tmp
  COPY --dir hugo/content hugo/assets hugo/layouts ./
  COPY hugo/hugo.toml hugo.toml
  COPY hugo/tailwind.config.js tailwind.config.js

  RUN hugo
  RUN ls -la public
  # SAVE ARTIFACT ./public AS LOCAL ./public
  SAVE ARTIFACT ./public 

build-image:
  FROM nginx:1.27.4
  COPY +build-html/public /usr/share/nginx/html
  EXPOSE 8080
  SAVE IMAGE ghcr.io/hmaier-dev/website-else/public-html

setup-ssh:
  FROM debian:bullseye
  RUN apt-get update && apt-get install -y openssh-client rsync
  RUN mkdir -p ~/.ssh
  RUN --secret key echo "$key" > /root/.ssh/id_ed25519
  RUN chmod 600 /root/.ssh/id_ed25519
  RUN --secret known_hosts echo "$known_hosts" > /root/.ssh/known_hosts

deploy:
  FROM +setup-ssh
