VERSION 0.8

hugo:
    FROM alpine:3.20
    RUN apk add --no-cache hugo

    # Hugo cannot work in root (/)
    WORKDIR tmp
    COPY content content
    COPY static static
    COPY hugo.toml hugo.toml
    COPY layouts layouts

    # generate meta-data
    COPY build/count-lines.sh .
    RUN ./count-lines.sh

    RUN mv content/index.md content/_index.md
    RUN hugo
    RUN ls -la public
    SAVE ARTIFACT ./public AS LOCAL ./public

build:
    BUILD +hugo
    # BUILD +pandoc --source ./content
