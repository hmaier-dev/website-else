VERSION 0.8

download:
  FROM debian:bullseye
  RUN apt-get update && apt-get install -y curl

download-hugo:
  FROM +download
  RUN curl -SLO https://github.com/gohugoio/hugo/releases/download/v0.140.2/hugo_0.140.2_linux-amd64.tar.gz
  RUN tar -xvzf hugo_0.140.2_linux-amd64.tar.gz
  RUN chmod +x hugo
  SAVE ARTIFACT hugo

download-tailwindcss:
  FROM +download
  RUN curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.0.0-beta.8/tailwindcss-linux-x64
  RUN chmod +x tailwindcss-linux-x64
  RUN mv tailwindcss-linux-x64 tailwindcss
  RUN ./tailwindcss --help
  SAVE ARTIFACT tailwindcss

build:
  FROM hmaierdev/hugo-plus-tailwind
  # COPY +download-hugo/hugo /usr/local/bin/hugo
  # COPY +download-tailwindcss/tailwindcss /usr/local/bin/tailwindcss

  # Hugo cannot work in root (/)
  WORKDIR tmp
  COPY --dir content assets layouts ./
  COPY hugo.toml hugo.toml
  COPY tailwind.config.js tailwind.config.js

  RUN hugo
  RUN ls -la public
  SAVE ARTIFACT ./public AS LOCAL ./public
  SAVE ARTIFACT ./public 

build-image:
  FROM nginx:1.27.4
  LABEL org.opencontainers.image.source = "https://github.com/hmaier-dev/website-else"
  COPY +build-html/public /usr/share/nginx/html
  EXPOSE 8080
  SAVE IMAGE --push ghcr.io/hmaier-dev/website-else/public-html

setup-ssh:
  FROM debian:bullseye
  RUN apt-get update && apt-get install -y openssh-client rsync
  RUN mkdir -p ~/.ssh
  RUN --secret key echo "$key" > /root/.ssh/id_ed25519
  RUN chmod 600 /root/.ssh/id_ed25519
  RUN --secret known_hosts echo "$known_hosts" > /root/.ssh/known_hosts

rsync:
  FROM +setup-ssh
  COPY +build/public ./public
  RUN --no-cache --secret port \
      --secret username \
      --secret host \
      --secret dest \
      rsync --port=$port -rav ./public $username@$host:$dest

deploy-test:
  FROM +rsync
  RUN --secret host --secret username ssh $username@$host "~/update-test.sh"

deploy-prod:
  FROM +rsync
  RUN --secret host --secret username ssh $username@$host "~/update-prod.sh"


test:
  FROM python:3.12
  COPY --dir ci ./
  RUN python -m pip install --upgrade pip 
  RUN pip install -r ci/requirements.txt
  RUN --no-cache python ci/check_response.py

