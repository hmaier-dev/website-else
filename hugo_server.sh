docker build --tag 'hugo_server' .

docker run --rm -it --name hugo_server \
    --volume $(pwd)/hugo.toml:/tmp/hugo.toml \
    --volume $(pwd)/layouts:/tmp/layouts \
    --volume $(pwd)/contents:/tmp/contents \
    --volume $(pwd)/assets:/tmp/assets \
    hugo_server \
    /bin/bash -c 'hugo server'
