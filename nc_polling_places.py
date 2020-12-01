#!/usr/bin/env python

import csv
import json
import time

import requests

from bs4 import BeautifulSoup

PRECINCT_LIST_URL = "https://vt.ncsbe.gov/PPLkup/LoadPrecincts/"
PRECINCT_ADDR_URL = "https://vt.ncsbe.gov/PPLkup/PollingPlaceResult/?CountyID=%d&PollingPlaceID=%d"

with open('nc_polling_places.csv', 'w', newline='') as csvfile:
    fil = csv.writer(csvfile, delimiter=',',
                            quotechar='"', quoting=csv.QUOTE_MINIMAL)

    for i in range(1, 101):
        precinct_locs = requests.post(PRECINCT_LIST_URL, {'CountyId': i})
        precinct_json = precinct_locs.json()
        for precinct in precinct_json:
            if precinct['Description'] == "<Select a Precinct>":
                continue
            # time to look up the specific polling place
            place_name = precinct['PollingPlaceName']
            full_page = requests.get(PRECINCT_ADDR_URL % (precinct['CountyID'], precinct['PollingPlaceID']))
            soup = BeautifulSoup(full_page.content, 'html.parser')
            ppdiv = soup.find(id="divPollingPlace")
            # print(ppdiv)
            place_addr = ', '.join(soup.find(id="divPollingPlace").p.a.stripped_strings)
            fil.writerow([precinct['CountyID'], precinct['Label'], precinct['Description'], place_name, place_addr])
            time.sleep(1.0)