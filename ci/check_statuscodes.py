#!/usr/bin/env python3

import requests
import os
from dataclasses import dataclass


# Define a dataclass to hold the URIs and their expected status codes
@dataclass
class URI:
    uri: str
    want: int


# Function to check the status of a URL
def check_status(uri, want):
    try:
        response = requests.get(uri)
        print(f"Checking {uri}: Status code {response.status_code}")
        if response.status_code == want:
            print(f"Status matches expected ({want})")
        else:
            print(f"Status does not match! Expected {want}, got {response.status_code}")
            os.exit(2)
    except requests.exceptions.RequestException as e:
        print(f"An error occurred while checking {uri}: {e}")


# List of URIs to check
uris = [
    URI(uri="https://www.elisa-adam.com/preview", want=200),
    URI(uri="https://elisa-adam.com/preview/about-me", want=200),
    URI(uri="https://www.elisa-adam.com/preview/workshops/", want=200),
    URI(uri="https://www.elisa-adam.com/missing", want=404),
]

# Iterate through the list and check each URI
for x in uris:
    check_status(x.uri, x.want)
