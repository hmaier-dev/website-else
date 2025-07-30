VERSION 0.8

build:
  FROM hmaierdev/hugo-plus-tailwind

  # Hugo cannot work in root (/)
  WORKDIR tmp
  COPY --dir content assets layouts ./
  COPY hugo.toml hugo.toml
  COPY tailwind.config.js tailwind.config.js

  RUN hugo
  RUN ls -la public
  # SAVE ARTIFACT ./public AS LOCAL ./public
  SAVE ARTIFACT ./public 

push:
  FROM nginx:1.27.4
  LABEL org.opencontainers.image.source='https://github.com/hmaier-dev/website-else'
  COPY +build-html/public /usr/share/nginx/html
  EXPOSE 8080
  SAVE IMAGE --push ghcr.io/hmaier-dev/website-else


## 
# setup-ssh:
#   FROM debian:bullseye
#   RUN apt-get update && apt-get install -y openssh-client rsync
#   RUN mkdir -p ~/.ssh
#   RUN --secret key echo "$key" > /root/.ssh/id_ed25519
#   RUN chmod 600 /root/.ssh/id_ed25519
#   RUN --secret known_hosts echo "$known_hosts" > /root/.ssh/known_hosts

# deploy:
#   BUILD +build-image
#   FROM +setup-ssh
#   RUN --secret username --secret host --secret dir ssh $username@$host "cd $dir; docker compose pull; docker compose up -d"
