import requests
import zipfile
import csv
import datetime
import redis

from django.conf import settings


def daily_bhavcopy():
    date = datetime.date.today().strftime("%d%m%y")
    hdr = {
        "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
        "Accept-Charset": "ISO-8859-1,utf-8;q=0.7,*;q=0.3",
        "Accept-Encoding": "none",
        "Accept-Language": "en-US,en;q=0.8",
        "Connection": "keep-alive",
    }
    temp = "tmp.zip"
    url = f"https://www.bseindia.com/download/BhavCopy/Equity/EQ{date}_CSV.ZIP"

    with requests.get(url, headers=hdr) as r:
        with open(temp, "wb") as fd:
            for chunk in r.iter_content(chunk_size=128):
                fd.write(chunk)

    with zipfile.ZipFile(temp, "r") as zf:
        zf.extractall()
        zf.close()

    with open(f"EQ{date}.CSV", "r") as csv_file:
        csv_reader = csv.reader(csv_file)
        count = 0
        r = redis.StrictRedis(host=settings.REDIS_HOST, port=settings.REDIS_PORT, db=0)

        for row in csv_reader:
            if count!=0:
                r.lpush("code", row[0].strip())
                r.lpush("name", row[1].strip())
                r.lpush("open", row[4].strip())
                r.lpush("high", row[5].strip())
                r.lpush("low", row[6].strip())
                r.lpush("close", row[7].strip())
            
            count+=1

        r.lpush("daily_len", count)