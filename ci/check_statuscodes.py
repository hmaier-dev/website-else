#!/usr/bin/env python3

import requests
import sys
from dataclasses import dataclass
from typing import Callable


@dataclass
class Response:
    uri: str
    need: Callable
    want: str


def check_response(uri, need, want):
    try:
        response = requests.get(uri)
        actual = need(response)
        if actual == want:
            print(f"({actual}) matches expected ({want})")
        else:
            print(f"Expected {want}, got {actual}")
            sys.exit(2)
    except requests.exceptions.RequestException as e:
        print(f"An error occurred while checking {uri}: {e}")


uris = [
    Response(uri="https://www.elisa-adam.com/preview", need=lambda r: r.status_code, want=200),
    Response(uri="https://elisa-adam.com/preview/about-me", need=lambda r: r.status_code, want=200),
    Response(uri="https://www.elisa-adam.com/preview/workshops/",need=lambda r: r.status_code, want=200),
    Response(uri="https://www.elisa-adam.com/missing", need=lambda r: r.status_code, want=404),
    Response(uri="https://elisa-adam.com/preview/media/spass-bei-der-arbeit.webp", need=lambda r: r.headers.get("Content-Type"), want="image/webp")
]


for x in uris:
    check_response(x.uri, x.need, x.want)
