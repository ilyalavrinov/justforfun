#!/usr/bin/python2
import time
import sys
import zipfile
import json
import urllib
import urllib2


def extract(f, url):
    archive = zipfile.ZipFile(f, 'r')
    for part in archive.namelist():
        content = archive.read(part)
        data = json.loads(content)
        for rec in data["accounts"]:
            req = urllib2.Request(url + "/accounts/new", json.dumps(rec))
            try:
                urllib2.urlopen(req)
            except urllib2.URLError as e:
                print e.read()


if __name__ == "__main__":
    time.sleep(3)
    extract(sys.argv[1], sys.argv[2])
