#!/usr/bin/env python3

import requests
import sys
from dataclasses import dataclass
from typing import Callable


@dataclass
class Response:
    uri: str
    resp_type: str  # just needed for print-statement
    # use lambda here to get the needed requests-Object
    need: Callable
    want: str


def check_response(uri, resp_type, need, want):
    try:
        response = requests.get(uri)
        actual = need(response)
        print(f"Testing {uri} for ({resp_type}): ({want}) ")
        if actual == want:
            print(f"Success: ({actual}) matches expected ({want})")
        else:
            print(f"Failure: Expected ({want}), got ({actual})")
            sys.exit(2)
    except requests.exceptions.RequestException as e:
        print(f"An error occurred while checking {uri}: {e}")


uris = [
    Response(
        uri="https://www.elisa-adam.com/preview",
        resp_type="Status Code",
        need=lambda r: r.status_code,
        want=200),
    Response(
        uri="https://elisa-adam.com/preview/about-me",
        resp_type="Status Code",
        need=lambda r: r.status_code,
        want=200),
    Response(
        uri="https://www.elisa-adam.com/preview/workshops/",
        resp_type="Status Code",
        need=lambda r: r.status_code,
        want=200),
    Response(
        uri="https://www.elisa-adam.com/missing",
        resp_type="Status Code",
        need=lambda r: r.status_code,
        want=404),
    Response(
        uri="https://elisa-adam.com/preview/media/spass-bei-der-arbeit.webp",
        resp_type="Content-Type",
        need=lambda r: r.headers.get("Content-Type", ""),
        want="image/webp")
]


for x in uris:
    check_response(x.uri, x.resp_type, x.need, x.want)
