import requests
import zipfile
import csv
import datetime
import redis
import xml.etree.ElementTree as ET

from django.conf import settings


class DownloadCSV:
    def __init__(self):
        # Redis connection to access DB
        self.r = redis.StrictRedis(
            host=settings.REDIS_HOST, port=settings.REDIS_PORT, db=0
        )
        # List of date when holiday
        self.holidays = []

        # Set the list of market holidays
        with requests.get(
            "https://zerodha.com/marketintel/holiday-calendar/?format=xml"
        ) as days:
            root = ET.fromstring(days.content)
            for holiday in root[0][5:]:
                self.holidays.append(
                    datetime.datetime.strptime(holiday[4].text[5:16], "%d %b %Y").date()
                )

    def push(self, key, value):
        self.r.lpush(key, value)

    def daily_bhavcopy(self):
        date = datetime.date.today()
        # Exit without downloading if today is a holiday or Sat/Sunday
        if date in self.holidays:
            return
        elif date.weekday() > 4:
            return

        hdr = {
            "User-Agent": "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11",
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
            "Accept-Charset": "ISO-8859-1,utf-8;q=0.7,*;q=0.3",
            "Accept-Encoding": "none",
            "Accept-Language": "en-US,en;q=0.8",
            "Connection": "keep-alive",
        }
        temp = "tmp.zip"
        date_str = date.strftime("%d%m%y")
        url = f"https://www.bseindia.com/download/BhavCopy/Equity/EQ{date_str}_CSV.ZIP"

        with requests.get(url, headers=hdr) as r:
            with open(temp, "wb") as fd:
                for chunk in r.iter_content(chunk_size=128):
                    fd.write(chunk)

        with zipfile.ZipFile(temp, "r") as zf:
            zf.extractall()
            zf.close()

        with open(f"EQ{date_str}.CSV", "r") as csv_file:
            csv_reader = csv.reader(csv_file)
            first = True

            for row in csv_reader:
                if not first:
                    self.push("code", row[0].strip())
                    self.push("name", row[1].strip())
                    self.push("open", row[4].strip())
                    self.push("high", row[5].strip())
                    self.push("low", row[6].strip())
                    self.push("close", row[7].strip())
                else:
                    first = False
